package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	client           *http.Client
	baseURL          *url.URL
	token            string
	userAgent        string
	rateLimitRetry   bool
	rateLimitHandler func(*http.Response) error
	retryMax         int
	retryWaitMin     time.Duration
	retryWaitMax     time.Duration
	requestHook      func(*http.Request)
	responseHook     func(*http.Response)

	User         *UsersService
	Repositories *RepositoriesService
	Issues       *IssuesService
	PullRequests *PullRequestsService
	Search       *SearchService
	RateLimit    *RateLimitService
}

const (
	defaultBaseURL  = "https://api.github.com"
	defaultWaitMin  = 5 * time.Second
	defaultWaitMax  = 60 * time.Second
	defaultRetryMax = 10

	userAgent = "go-github-client/1.0"
)

func NewClient(opts ...option) (*Client, error) {
	baseURL, _ := url.Parse(defaultBaseURL)
	client := &Client{
		client:       http.DefaultClient,
		baseURL:      baseURL,
		userAgent:    userAgent,
		retryMax:     defaultRetryMax,
		retryWaitMin: defaultWaitMin,
		retryWaitMax: defaultWaitMax,
	}

	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	client.User = &UsersService{client}
	client.Repositories = &RepositoriesService{client}
	client.Issues = &IssuesService{client}
	client.PullRequests = &PullRequestsService{client}
	client.Search = &SearchService{client}
	client.RateLimit = &RateLimitService{client}

	return client, nil
}

func (c *Client) NewRequest(method, path string, body any) (*http.Request, error) {
	var payload io.ReadWriter
	if body != nil {
		payload = &bytes.Buffer{}

		err := json.NewEncoder(payload).Encode(body)
		if err != nil {
			return nil, fmt.Errorf("failed to encode JSON: %w", err)
		}
	}

	url, err := c.baseURL.Parse(path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url.String(), payload)
	if err != nil {
		return nil, fmt.Errorf("failed create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("X-Github-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", c.userAgent)

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return req, nil
}

func (c *Client) Do(ctx context.Context, req *http.Request, v any) (*Response, error) {
	if c.requestHook != nil {
		c.requestHook(req)
	}

	var res *http.Response
	var err error
	var rateLim *RateLimit

	req = req.WithContext(ctx)

	maxAtm := max(c.retryMax, 1)
	for atm := range maxAtm {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		res, err = c.client.Do(req)
		rateLim = getRateLimit(res)
		if err != nil {
			return nil, err
		}

		if c.responseHook != nil {
			c.responseHook(res)
		}

		if (res.StatusCode == 403 || res.StatusCode == 429) && rateLim.Remaining == 0 {
			if !c.rateLimitRetry {
				return buildResponse(res, rateLim), newApiError(res)
			}

			if c.rateLimitHandler != nil {
				if err := c.rateLimitHandler(res); err != nil {
					return buildResponse(res, rateLim), err
				}
				continue
			}
			_ = res.Body.Close()

			c.waitRateLimit(rateLim, atm)
			continue
		}
		break
	}

	response := buildResponse(res, rateLim)

	if res.StatusCode >= 400 {
		return response, newApiError(res)
	}

	if v != nil {
		if err := json.NewDecoder(res.Body).Decode(v); err != nil {
			return response, err
		}
	}

	_ = res.Body.Close()

	return response, nil
}

func (c *Client) waitRateLimit(rl *RateLimit, attempt int) {
	var waitTime time.Duration
	if rl.Reset != 0 {
		resetTime := time.Unix(rl.Reset, 0)
		waitTime = time.Until(resetTime)

		if waitTime < 0 {
			waitTime = time.Second
		}
	} else {
		waitTime = c.calculateBackoff(attempt)
	}

	time.Sleep(waitTime)
}

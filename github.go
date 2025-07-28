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

const (
	defaultBaseURL  = "https://api.github.com"
	defaultWaitMin  = time.Second
	defaultWaitMax  = 60 * time.Second
	defaultRetryMax = 5

	userAgent = "go-github"
)

// Client manages communication with the API.
// It provides methods for accessing various API endpoints through
// specialized services and handles common functionality such as
// authentication, rate limiting, request/response processing,
// and error handling.
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
	responseHook     func(*Response)

	// User service for user-related operations
	User *UsersService

	// Repositories service for repository-related operations
	Repositories *RepositoriesService

	// Issues service for issue-related operations
	Issues *IssuesService

	// PullRequests service for pull request-related operations
	PullRequests *PullRequestsService

	// Search service for search operations
	Search *SearchService

	// RateLimit service for rate limiting operations
	RateLimit *RateLimitService
}

// NewClient creates a new API client with optional configuration.
// This function initializes a new Client instance with default settings
// and applies any provided functional options to customize the client's
// behavior.
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
			return nil, fmt.Errorf("failed to apply client option: %w", err)
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

// NewRequest creates an API request with the specified HTTP method, path, and body.
// This method constructs an HTTP request with proper headers including authentication,
// content type, accept headers, and user agent.
func (c *Client) NewRequest(method, path string, body any) (*http.Request, error) {
	var payload io.ReadWriter
	if body != nil {
		payload = &bytes.Buffer{}

		err := json.NewEncoder(payload).Encode(body)
		if err != nil {
			return nil, fmt.Errorf("failed to encode request body: %w", err)
		}
	}

	url, err := c.baseURL.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL path %s: %w", path, err)
	}

	req, err := http.NewRequest(method, url.String(), payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
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

// Do sends an API request and returns the API response.
// This method executes the provided HTTP request and handles the response,
// including automatic retry logic for rate limiting, error handling, and
// JSON decoding of the response body into the provided target value.
func (c *Client) Do(ctx context.Context, req *http.Request, v any) (*Response, error) {
	req = req.WithContext(ctx)

	var resp *http.Response
	var err error
	var rateLim *RateLimit

	maxAtm := max(c.retryMax, 1)
	for atm := range maxAtm {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if c.requestHook != nil {
			c.requestHook(req)
		}

		resp, err = c.client.Do(req)
		if err != nil {
			return nil, err
		}

		rateLim = getRateLimit(resp)
		if c.responseHook != nil {
			c.responseHook(buildResponse(resp, rateLim))
		}

		if (resp.StatusCode == 403 || resp.StatusCode == 429) && rateLim.Remaining == 0 {
			if !c.rateLimitRetry {
				return buildResponse(resp, rateLim), newAPIError(resp)
			}

			if c.rateLimitHandler != nil {
				if err := c.rateLimitHandler(resp); err != nil {
					return buildResponse(resp, rateLim), err
				}
				continue
			}
			_ = resp.Body.Close()

			c.waitRateLimit(rateLim, atm)
			continue
		}
		break
	}

	response := buildResponse(resp, rateLim)

	if resp.StatusCode >= 400 {
		return response, newAPIError(resp)
	}

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return response, err
		}
	}

	_ = resp.Body.Close()

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

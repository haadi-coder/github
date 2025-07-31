package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"slices"
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

	var httpresp *http.Response
	var err error
	var resp *Response

	maxAtm := max(c.retryMax, 1)
	for attempt := range maxAtm {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if c.requestHook != nil {
			c.requestHook(req)
		}

		httpresp, err = c.client.Do(req)
		if err != nil {
			return nil, err
		}

		resp, err = newResponse(httpresp)
		if err != nil {
			return resp, err
		}

		if c.responseHook != nil {
			c.responseHook(resp)
		}

		if !checkRetry(resp) {
			break
		}

		if !c.rateLimitRetry {
			return resp, newAPIError(httpresp)
		}

		if c.rateLimitHandler != nil {
			err = c.rateLimitHandler(httpresp)
			if err != nil {
				return resp, err
			}
		}

		_ = httpresp.Body.Close()

		if attempt >= maxAtm-1 {
			return resp, fmt.Errorf("max retry attempts %d exceeded", maxAtm)
		}

		waitTime := calcBackoff(c.retryWaitMin, c.retryWaitMax, attempt, resp)
		select {
		case <-ctx.Done():
			return resp, ctx.Err()
		case <-time.After(waitTime):
			continue
		default:
		}
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return resp, newAPIError(httpresp)
	}

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
		if err != nil {
			return resp, err
		}
	}

	_ = resp.Body.Close()

	return resp, nil
}

func checkRetry(resp *Response) bool {
	serviceShutted := []int{
		http.StatusForbidden,
		http.StatusTooManyRequests,
	}
	if slices.Contains(serviceShutted, resp.StatusCode) && resp.Remaining == 0 {
		return true
	}

	serviceUnavailable := []int{
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
	}

	return slices.Contains(serviceUnavailable, resp.StatusCode)
}

func calcBackoff(minD time.Duration, maxD time.Duration, attempt int, resp *Response) time.Duration {
	if resp.Reset != 0 {
		resetTime := time.Unix(resp.Reset, 0)

		return time.Until(resetTime)
	}

	const binBase = 2

	backoff := float64(minD) * math.Pow(binBase, float64(attempt))
	wait := time.Duration(backoff)

	return min(wait, maxD)
}

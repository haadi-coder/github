package github

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"time"
)

// RateLimitService provides access to rate limit API methods.
type RateLimitService struct {
	client *Client
}

// RateLimit represents a GitHub rate limit information for a specific resource.
// It contains details about how many requests you can make, how many
// remaining, and when the limit will reset.
// GitHub API docs: https://docs.github.com/en/rest/rate-limit/rate-limit
type RateLimit struct {
	Limit     int   `json:"limit"`
	Remaining int   `json:"remaining"`
	Used      int   `json:"used"`
	Reset     int64 `json:"reset"`
}

// RateLimitResponse represents the complete rate limit information
// returned by the GitHub API
type RateLimitResponse struct {
	Resources *RateLimitResources
	Rate      *RateLimit
}

// RateLimitResources represents rate limits for different API resources.
// Each field corresponds to a different category of GitHub API endpoints
// with their own separate rate limits.
type RateLimitResources struct {
	Core                      *RateLimit
	Search                    *RateLimit
	Graphql                   *RateLimit
	IntegrationManifest       *RateLimit
	SourceImport              *RateLimit
	CodeScanningUpload        *RateLimit
	ActionsRunnerRegistration *RateLimit
	Scim                      *RateLimit
	DependencySnapshots       *RateLimit
	CodeSearch                *RateLimit
	CodeScanningAutofix       *RateLimit
}

const (
	rateLimitHeader    = "X-RateLimit-Limit"
	rateRemainigHeader = "X-RateLimit-Remaining"
	rateResetHeader    = "X-RateLimit-Reset"
	rateUsedHeader     = "X-RateLimit-Used"
)

// Get retrieves the current rate limit status for the authenticated user.
// This method returns detailed information about rate limits for all
// API resources, including how many requests have been made, how many
// are remaining, and when the limits will reset.
func (s *RateLimitService) Get(ctx context.Context) (*RateLimitResponse, error) {
	path := "rate_limit"

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	rl := new(RateLimitResponse)
	if _, err := s.client.Do(ctx, req, rl); err != nil {
		return nil, err
	}

	return rl, nil
}

func getRateLimit(resp *http.Response) *RateLimit {
	rl := RateLimit{}

	if lim := resp.Header.Get(rateLimitHeader); lim != "" {
		if intL, err := strconv.Atoi(lim); err == nil {
			rl.Limit = intL
		}
	}

	if rem := resp.Header.Get(rateRemainigHeader); rem != "" {
		if intRm, err := strconv.Atoi(rem); err == nil {
			rl.Remaining = intRm
		}
	}

	if res := resp.Header.Get(rateResetHeader); res != "" {
		if intRes, err := strconv.ParseInt(res, 10, 64); err == nil {
			rl.Reset = intRes
		}
	}

	if used := resp.Header.Get(rateUsedHeader); used != "" {
		if intUsed, err := strconv.Atoi(used); err == nil {
			rl.Used = intUsed
		}
	}

	return &rl
}

func (c *Client) calculateBackoff(attempt int) time.Duration {
	if c.retryWaitMin.Seconds() == 0 {
		c.retryWaitMin = 5 * time.Second
	}
	if c.retryWaitMax.Seconds() == 0 {
		c.retryWaitMax = 60 * time.Second
	}

	wait := c.retryWaitMin.Seconds() * math.Pow(2, float64(attempt))
	if wait > c.retryWaitMax.Seconds() {
		wait = c.retryWaitMax.Seconds()
	}

	return time.Duration(wait * float64(time.Second))
}

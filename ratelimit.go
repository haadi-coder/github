package github

import (
	"context"
	"net/http"
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

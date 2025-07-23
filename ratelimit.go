package main

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"time"
)

type RateLimitService struct {
	client *Client
}

type RateLimit struct {
	Limit     int   `json:"limit"`
	Remaining int   `json:"remaining"`
	Used      int   `json:"used"`
	Reset     int64 `json:"reset"`
}

type RateLimitResponse struct {
	Resources struct {
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
	Rate *RateLimit
}

func (s *RateLimitService) Get(ctx context.Context) (*RateLimitResponse, error) {
	url := s.client.baseUrl.JoinPath("rate_limit").String()

	req, err := s.client.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	rl := new(RateLimitResponse)
	if _, err := s.client.Do(ctx, req, rl); err != nil {
		return nil, err
	}

	return rl, nil
}

func getRateLimit(res *http.Response) *RateLimit {
	rl := RateLimit{}

	if lim := res.Header.Get(rateLimitHeader); lim != "" {
		if intL, err := strconv.Atoi(lim); err == nil {
			rl.Limit = intL
		}
	}

	if rem := res.Header.Get(rateRemainigHeader); rem != "" {
		if intRm, err := strconv.Atoi(rem); err == nil {
			rl.Remaining = intRm
		}
	}

	if res := res.Header.Get(rateResetHeader); res != "" {
		if intRes, err := strconv.ParseInt(res, 10, 64); err == nil {
			rl.Reset = intRes
		}
	}

	rl.Used = rl.Limit - rl.Remaining

	return &rl
}

func (c *Client) calculateBackoff(attempt int) time.Duration {
	if c.retryWaitMin == 0 {
		c.retryWaitMin = 5
	}
	if c.retryWaitMax == 0 {
		c.retryWaitMax = 60
	}

	wait := c.retryWaitMin * math.Pow(2, float64(attempt))
	if wait > c.retryWaitMax {
		wait = c.retryWaitMax
	}

	return time.Duration(wait * float64(time.Second))
}

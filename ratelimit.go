package main

import (
	"math"
	"net/http"
	"strconv"
	"time"
)

type RateLimit struct {
	Limit     int
	Remaining int
	Reset     int64
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

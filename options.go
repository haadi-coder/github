package github

import (
	"net/http"
	"net/url"
)

type option func(*Client)

func WithToken(t string) option {
	return func(c *Client) {
		c.token = t
	}
}

func WithHTTPClient(hc *http.Client) option {
	return func(c *Client) {
		c.hc = hc
	}
}

func WithBaseURl(u string) option {
	return func(c *Client) {
		parsed, err := url.Parse(u)
		if err == nil {
			c.baseUrl = parsed
		}
	}
}

func WithUserAgent(ua string) option {
	return func(c *Client) {
		c.userAgent = ua
	}
}

func WithRateLimitRetry(r bool) option {
	return func(c *Client) {
		c.rateLimitRetry = r
	}
}

func WithRetryMax(rc int) option {
	return func(c *Client) {
		c.retryMax = rc
	}
}

func WithRetryWaitMin(rwMin float64) option {
	return func(c *Client) {
		c.retryWaitMin = rwMin
	}
}

func WithRetryWaitMax(rwMax float64) option {
	return func(c *Client) {
		c.retryWaitMax = rwMax
	}
}

func WithRequestHook(rqh func(*http.Request)) option {
	return func(c *Client) {
		c.requestHook = rqh
	}
}

func WithResposneHook(rsh func(*http.Response)) option {
	return func(c *Client) {
		c.responseHook = rsh
	}
}

func WithRateLimitHandler(rlh func(*http.Response) error) option {
	return func(c *Client) {
		c.rateLimitHandler = rlh
	}
}

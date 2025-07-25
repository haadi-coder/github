package github

import (
	"net/http"
	"net/url"
	"time"
)

type option func(*Client) error

func WithToken(t string) option {
	return func(c *Client) error {
		c.token = t
		return nil
	}
}

func WithHTTPClient(hc *http.Client) option {
	return func(c *Client) error {
		c.client = hc
		return nil
	}
}

func WithBaseURl(u string) option {
	return func(c *Client) error {
		parsed, err := url.Parse(u)
		if err != nil {
			return err
		}
		c.baseURL = parsed
		return nil
	}
}

func WithUserAgent(ua string) option {
	return func(c *Client) error {
		c.userAgent = ua
		return nil
	}
}

func WithRateLimitRetry(r bool) option {
	return func(c *Client) error {
		c.rateLimitRetry = r
		return nil
	}
}

func WithRetryMax(rc int) option {
	return func(c *Client) error {
		c.retryMax = rc
		return nil
	}
}

func WithRetryWaitMin(rwMin time.Duration) option {
	return func(c *Client) error {
		c.retryWaitMin = rwMin
		return nil
	}
}

func WithRetryWaitMax(rwMax time.Duration) option {
	return func(c *Client) error {
		c.retryWaitMax = rwMax
		return nil
	}
}

func WithRequestHook(rqh func(*http.Request)) option {
	return func(c *Client) error {
		c.requestHook = rqh
		return nil
	}
}

func WithResposneHook(rsh func(*http.Response)) option {
	return func(c *Client) error {
		c.responseHook = rsh
		return nil
	}
}

func WithRateLimitHandler(rlh func(*http.Response) error) option {
	return func(c *Client) error {
		c.rateLimitHandler = rlh
		return nil
	}
}

package github

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type option func(*Client) error

// WithToken configures the client to use the specified authentication token.
// This token will be used for authenticating API requests, typically as a
// Bearer token in the Authorization header.
func WithToken(token string) option {
	return func(c *Client) error {
		c.token = token
		return nil
	}
}

// WithHTTPClient configures the client to use the specified HTTP client
// for making requests. This allows customization of the underlying HTTP
// transport, timeouts, and other HTTP-related settings.
func WithHTTPClient(client *http.Client) option {
	return func(c *Client) error {
		c.client = client
		return nil
	}
}

// WithBaseURL configures the client to use the specified base URL for
// all API requests. This is useful for testing against different API
// endpoints or enterprise GitHub instances.
func WithBaseURL(baseUrl string) option {
	return func(c *Client) error {
		parsed, err := url.Parse(baseUrl)
		if err != nil {
			return fmt.Errorf("failed to parse base URL %s: %w", baseUrl, err)
		}
		c.baseURL = parsed
		return nil
	}
}

// WithUserAgent configures the client to use the specified User-Agent
// header value for all requests. This helps API operators identify
// requests made by this client.
func WithUserAgent(agent string) option {
	return func(c *Client) error {
		c.userAgent = agent
		return nil
	}
}

// WithRateLimitRetry configures whether the client should automatically
// retry requests that hit rate limits. When enabled, the client will
// wait for the rate limit to reset before retrying the request.
func WithRateLimitRetry(retry bool) option {
	return func(c *Client) error {
		c.rateLimitRetry = retry
		return nil
	}
}

// WithRateLimitHandler configures a custom handler function for rate
// limit responses. This function will be called when a rate limit
// is encountered, allowing for custom rate limit handling logic.
func WithRateLimitHandler(handler func(*http.Response) error) option {
	return func(c *Client) error {
		c.rateLimitHandler = handler
		return nil
	}
}

// WithRetryMax configures the maximum number of retry attempts for
// failed requests. This setting applies to retryable errors such as
// network issues or server errors.
func WithRetryMax(retryCount int) option {
	return func(c *Client) error {
		c.retryMax = retryCount
		return nil
	}
}

// WithRetryWaitMin configures the minimum wait time between retry attempts.
// The actual wait time may be longer due to exponential backoff or
// rate limit reset times.
func WithRetryWaitMin(wait time.Duration) option {
	return func(c *Client) error {
		c.retryWaitMin = wait
		return nil
	}
}

// WithRetryWaitMax configures the maximum wait time between retry attempts.
// This caps the exponential backoff algorithm to prevent excessively
// long waits between retries.
func WithRetryWaitMax(wait time.Duration) option {
	return func(c *Client) error {
		c.retryWaitMax = wait
		return nil
	}
}

// WithRequestHook configures a hook function that will be called before
// each HTTP request is sent. This allows for request inspection,
// logging, or modification before the request is executed.
func WithRequestHook(hook func(*http.Request)) option {
	return func(c *Client) error {
		c.requestHook = hook
		return nil
	}
}

// WithResponseHook configures a hook function that will be called after
// each HTTP response is received. This allows for response inspection,
// logging, or custom processing of API responses.
func WithResponseHook(hook func(*Response)) option {
	return func(c *Client) error {
		c.responseHook = hook
		return nil
	}
}

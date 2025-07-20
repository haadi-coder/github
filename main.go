package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	hc               *http.Client
	baseUrl          *url.URL
	token            string
	userAgent        string
	rateLimitRetry   bool
	retryMax         int
	retryWaitMin     float64
	retryWaitMax     float64
	requestHook      func(*http.Request)
	responseHook     func(*http.Response)
	rateLimitHandler func(*http.Response) error

	User UsersService
}

const (
	defaultBaseUrl  = "https://api.github.com"
	defaultWaitMin  = 5
	defaultWaitMax  = 60
	defaultRetryMax = 10

	userAgentHeader    = "go-github-client/1.0"
	acceptHeader       = "application/vnd.github.v3+json"
	versionHeader      = "2022-11-28"
	rateLimitHeader    = "X-RateLimit-Limit"
	rateRemainigHeader = "X-RateLimit-Remaining"
	rateResetHeader    = "X-RateLimit-Reset"

	linkPrev  = "prev"
	linkNext  = "next"
	linkFirst = "first"
	linkLast  = "last"
)

func main() {
	gc := NewClient()
	ctx := context.Background()

	user, err := gc.User.Get(ctx, "haadi-coder")
	// user, err := gc.User.GetAuthenticated(ctx)

	// fmt.Printf("ID: %d\nLogin: %s\nName: %s\n", user.Id, user.Login, user.Name)
	fmt.Print(user, err)

}

func NewClient(opts ...option) *Client {
	parsed, _ := url.Parse(defaultBaseUrl)
	client := &Client{
		hc:           http.DefaultClient,
		baseUrl:      parsed,
		userAgent:    userAgentHeader,
		retryMax:     defaultRetryMax,
		retryWaitMin: defaultWaitMin,
		retryWaitMax: defaultWaitMax,
	}

	client.User = UsersService{client: client}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (c *Client) NewRequest(method, path string, body any) (*http.Request, error) {
	url := c.baseUrl.JoinPath(path)

	var payload io.ReadWriter
	if body != nil {
		payload = &bytes.Buffer{}

		err := json.NewEncoder(payload).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url.String(), payload)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", acceptHeader)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("X-Github-Api-Version", versionHeader)

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return req, nil
}

type Response struct {
	*http.Response

	PreviousPage int
	NextPage     int
	FirstPage    int
	LastPage     int
}

type ErrorResponse struct {
	*http.Response
	Message          string `json:"message"`
	DocumentationUrl string `json:"documentation_url,omitempty"`

	Errors []struct {
		Code     string `json:"code"`
		Resource string `json:"resource"`
		Field    string `json:"field"`
	}
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("API Error: %d - %s\n", e.StatusCode, e.Message)
}

type ListOptions struct {
	Page    int
	PerPage int
}

type RateLimit struct {
	Limit     int
	Remaining int
	Reset     int64
}

func (c *Client) Do(ctx context.Context, req *http.Request) (*Response, error) {
	if c.requestHook != nil {
		c.requestHook(req)
	}

	var res *http.Response
	var err error

	maxAtm := max(c.retryMax, 1)

	for atm := range maxAtm {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		res, err = c.hc.Do(req)
		if err != nil {
			return nil, err
		}

		if c.responseHook != nil {
			c.responseHook(res)
		}

		rateLim := getRateLimit(res)
		fmt.Print(rateLim)
		if (res.StatusCode == 403 || res.StatusCode == 429) && rateLim.Remaining == 0 {
			_ = res.Body.Close()

			if !c.rateLimitRetry {
				return buildResponse(res), buildErrorResponse(res)
			}

			if c.rateLimitHandler != nil {
				if err := c.rateLimitHandler(res); err != nil {
					return buildResponse(res), err
				}
				continue
			}

			var waitTime time.Duration
			if rateLim.Reset != 0 {
				resetTime := time.Unix(rateLim.Reset, 0)
				waitTime = time.Until(resetTime)

				if waitTime < 0 {
					waitTime = time.Second
				}
			} else {
				waitTime = c.calculateBackoff(atm)
			}

			time.Sleep(waitTime)
			continue
		}

		break
	}

	response := buildResponse(res)

	if res.StatusCode >= 400 {
		_ = res.Body.Close()
		return response, buildErrorResponse(res)
	}

	return response, nil
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

func buildResponse(hr *http.Response) *Response {
	if hr == nil {
		return &Response{}
	}

	res := &Response{
		Response: hr,
	}

	if l := hr.Header.Get("Link"); l != "" {
		err := parseLinkHeader(res, l)
		if err != nil {
			return res
		}
	}

	return res
}

func buildErrorResponse(hr *http.Response) error {
	if hr == nil {
		return errors.New("received nil response")
	}

	errResponse := &ErrorResponse{
		Response: hr,
	}

	if err := json.NewDecoder(hr.Body).Decode(errResponse); err != nil {
		errResponse.Message = fmt.Sprintf("Request failed with status %d", hr.StatusCode)
	}

	return errResponse
}

func parseLinkHeader(res *Response, link string) error {
	if link == "" {
		return errors.New("invalid Link Header")
	}

	for pair := range strings.SplitSeq(link, ",") {
		parts := strings.Split(pair, ";")
		if len(parts) != 2 {
			continue
		}

		rawUrl := strings.Trim(parts[0], "< >")
		rel := strings.Trim(parts[1], " rel=")

		url, err := url.Parse(rawUrl)
		if err != nil {
			return err
		}

		queries := url.Query()
		pageCount, err := strconv.Atoi(queries.Get("page"))
		if err != nil {
			return err
		}

		if pageCount == 0 {
			continue
		}

		switch rel {
		case linkPrev:
			res.PreviousPage = pageCount
		case linkNext:
			res.NextPage = pageCount
		case linkFirst:
			res.FirstPage = pageCount
		case linkLast:
			res.LastPage = pageCount
		default:
			continue
		}
	}

	return nil
}

package github

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/tomnomnom/linkheader"
)

// Response wraps the standard http.Response and provides additional
// convenience methods for working with API responses. It includes
// rate limit information and pagination helpers extracted from
// response headers.
type Response struct {
	*http.Response

	// RateLimit contains the rate limit information extracted from
	// the response headers for the current request
	*RateLimit

	// PreviousPage contains the page number of the previous page
	// of results, if available
	PreviousPage int

	// NextPage contains the page number of the next page of results,
	// if available
	NextPage int

	// FirstPage contains the page number of the first page of results
	FirstPage int

	// LastPage contains the page number of the last page of results,
	// if available
	LastPage int
}

func newResponse(httpresp *http.Response) (*Response, error) {
	resp := &Response{
		Response:  httpresp,
		RateLimit: &RateLimit{},
	}

	if err := populateRateLimit(resp); err != nil {
		return resp, err
	}

	if err := populatePagination(resp); err != nil {
		return resp, err
	}

	return resp, nil
}

const (
	rateLimitHeader    = "X-RateLimit-Limit"
	rateRemainigHeader = "X-RateLimit-Remaining"
	rateResetHeader    = "X-RateLimit-Reset"
	rateUsedHeader     = "X-RateLimit-Used"
)

func populateRateLimit(resp *Response) error {
	rawLimit := resp.Header.Get(rateLimitHeader)
	if rawLimit != "" {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil {
			return fmt.Errorf("failed to parse rate limit header: %w", err)
		}

		resp.Limit = limit
	}

	rawRemaining := resp.Header.Get(rateRemainigHeader)
	if rawRemaining != "" {
		remaining, err := strconv.Atoi(rawRemaining)
		if err != nil {
			return fmt.Errorf("failed to parse rate remaining header: %w", err)
		}

		resp.Remaining = remaining
	}

	rawReset := resp.Header.Get(rateResetHeader)
	if rawReset != "" {
		reset, err := strconv.ParseInt(rawReset, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse rate reset header: %w", err)
		}

		resp.Reset = reset
	}

	rawUsed := resp.Header.Get(rateUsedHeader)
	if rawUsed != "" {
		used, err := strconv.Atoi(rawUsed)
		if err != nil {
			return fmt.Errorf("failed to parse rate used header: %w", err)
		}

		resp.Used = used
	}

	return nil
}

const (
	linkPrev  = "prev"
	linkNext  = "next"
	linkFirst = "first"
	linkLast  = "last"
)

func populatePagination(resp *Response) error {
	header := resp.Header.Get("Link")
	if header == "" {
		return nil
	}

	links := linkheader.Parse(header)

	for _, link := range links {
		url, err := url.Parse(link.URL)
		if err != nil {
			return err
		}

		page, err := strconv.Atoi(url.Query().Get("page"))
		if err != nil {
			return err
		}

		if page == 0 {
			continue
		}

		switch link.Rel {
		case linkPrev:
			resp.PreviousPage = page
		case linkNext:
			resp.NextPage = page
		case linkFirst:
			resp.FirstPage = page
		case linkLast:
			resp.LastPage = page
		}
	}

	return nil
}

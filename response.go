package github

import (
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

func newResponse(httpresp *http.Response) *Response {
	resp := &Response{
		Response:  httpresp,
		RateLimit: &RateLimit{},
	}

	populateRateLimit(resp)

	err := populateLinkHeader(resp)
	if err != nil {
		return resp
	}

	return resp
}

const (
	rateLimitHeader    = "X-RateLimit-Limit"
	rateRemainigHeader = "X-RateLimit-Remaining"
	rateResetHeader    = "X-RateLimit-Reset"
	rateUsedHeader     = "X-RateLimit-Used"
)

func populateRateLimit(resp *Response) {
	if rawLimit := resp.Header.Get(rateLimitHeader); rawLimit != "" {
		if limit, err := strconv.Atoi(rawLimit); err == nil {
			resp.Limit = limit
		}
	}
	if rawRemaining := resp.Header.Get(rateRemainigHeader); rawRemaining != "" {
		if remaining, err := strconv.Atoi(rawRemaining); err == nil {
			resp.Remaining = remaining
		}
	}
	if rawReset := resp.Header.Get(rateResetHeader); rawReset != "" {
		if reset, err := strconv.ParseInt(rawReset, 10, 64); err == nil {
			resp.Reset = reset
		}
	}
	if rawUsed := resp.Header.Get(rateUsedHeader); rawUsed != "" {
		if used, err := strconv.Atoi(rawUsed); err == nil {
			resp.Used = used
		}
	}
}

const (
	linkPrev  = "prev"
	linkNext  = "next"
	linkFirst = "first"
	linkLast  = "last"
)

func populateLinkHeader(resp *Response) error {
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

		pageCount, err := strconv.Atoi(url.Query().Get("page"))
		if err != nil {
			return err
		}

		if pageCount == 0 {
			continue
		}

		switch link.Rel {
		case linkPrev:
			resp.PreviousPage = pageCount
		case linkNext:
			resp.NextPage = pageCount
		case linkFirst:
			resp.FirstPage = pageCount
		case linkLast:
			resp.LastPage = pageCount
		default:
			continue
		}
	}

	return nil
}

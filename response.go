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

const (
	rateLimitHeader    = "X-RateLimit-Limit"
	rateRemainigHeader = "X-RateLimit-Remaining"
	rateResetHeader    = "X-RateLimit-Reset"
	rateUsedHeader     = "X-RateLimit-Used"
)

func buildResponse(httpresp *http.Response) *Response {
	resp := &Response{
		Response:  httpresp,
		RateLimit: &RateLimit{},
	}

	if lim := resp.Header.Get(rateLimitHeader); lim != "" {
		if intL, err := strconv.Atoi(lim); err == nil {
			resp.Limit = intL
		}
	}
	if rem := resp.Header.Get(rateRemainigHeader); rem != "" {
		if intRm, err := strconv.Atoi(rem); err == nil {
			resp.Remaining = intRm
		}
	}
	if res := resp.Header.Get(rateResetHeader); res != "" {
		if intRes, err := strconv.ParseInt(res, 10, 64); err == nil {
			resp.Reset = intRes
		}
	}
	if used := resp.Header.Get(rateUsedHeader); used != "" {
		if intUsed, err := strconv.Atoi(used); err == nil {
			resp.Used = intUsed
		}
	}

	err := parseLinkHeader(resp)
	if err != nil {
		return resp
	}

	return resp
}

const (
	linkPrev  = "prev"
	linkNext  = "next"
	linkFirst = "first"
	linkLast  = "last"
)

func parseLinkHeader(resp *Response) error {
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

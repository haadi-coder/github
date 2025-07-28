package github

import (
	"errors"
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

func buildResponse(hr *http.Response, rl *RateLimit) *Response {
	resp := &Response{
		Response:  hr,
		RateLimit: rl,
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
		return errors.New("invalid Link Header")
	}

	links := linkheader.Parse(header)
	for _, link := range links {
		rel := link.Rel

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

		switch rel {
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

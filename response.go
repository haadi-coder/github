package github

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/tomnomnom/linkheader"
)

type Response struct {
	*http.Response
	*RateLimit

	PreviousPage int
	NextPage     int
	FirstPage    int
	LastPage     int
}

func buildResponse(hr *http.Response, rl *RateLimit) *Response {
	res := &Response{
		Response:  hr,
		RateLimit: rl,
	}

	err := parseLinkHeader(res)
	if err != nil {
		return res
	}

	return res
}

const (
	linkPrev  = "prev"
	linkNext  = "next"
	linkFirst = "first"
	linkLast  = "last"
)

func parseLinkHeader(res *Response) error {
	header := res.Header.Get("Link")
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

package github

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

type SearchService struct {
	client *Client
}

type Search[T Repository | User] struct {
	TotalCount        int  `json:"total_count"`
	IncompleteResults bool `json:"incomplete_results"`
	Items             []*T `json:"items"`
}

type SearchOptions struct {
	*ListOptions
	Sort  *string
	Order *string
}

func (s *SearchService) Repositories(ctx context.Context, sq string, opts *SearchOptions) (*Search[Repository], error) {
	path := "search/repositories"

	q := url.Values{}
	if opts != nil {
		if opts.ListOptions != nil {
			opts.paginateQuery(q)
		}
		if opts.Order != nil {
			q.Set("order", *opts.Order)
		}
		if opts.Sort != nil {
			q.Set("sort", *opts.Sort)
		}

		if len(q) != 0 {
			path += "?" + q.Encode()
		}
	}

	if len(q) != 0 {
		path += "&" + buildSearchParams(sq)
	} else {
		path += "?" + buildSearchParams(sq)
	}

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	search := new(Search[Repository])
	if _, err := s.client.Do(ctx, req, search); err != nil {
		return nil, err
	}

	return search, nil
}

func (s *SearchService) Users(ctx context.Context, sq string, opts *SearchOptions) (*Search[User], error) {
	path := "search/users"

	q := url.Values{}
	if opts != nil {

		if opts.ListOptions != nil {
			opts.paginateQuery(q)
		}
		if opts.Order != nil {
			q.Set("order", *opts.Order)
		}
		if opts.Sort != nil {
			q.Set("sort", *opts.Sort)
		}

		if len(q) != 0 {
			path += "?" + q.Encode()
		}
	}

	if len(q) != 0 {
		path += "&" + buildSearchParams(sq)
	} else {
		path += "?" + buildSearchParams(sq)
	}

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	search := new(Search[User])
	if _, err := s.client.Do(ctx, req, search); err != nil {
		return nil, err
	}

	return search, nil
}

func buildSearchParams(s string) string {
	trimmed := strings.TrimSpace(s)
	chars := strings.Split(trimmed, " ")
	encodedQuery := strings.Join(chars, "+")

	return "q=" + encodedQuery
}

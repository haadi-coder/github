package main

import (
	"context"
	"net/http"
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
	rawUrl := s.client.baseUrl.JoinPath("search", "repositories")
	if opts != nil {
		q := rawUrl.Query()
		opts.paginateQuery(q)
		if opts.Order != nil {
			q.Set("order", *opts.Order)
		}
		if opts.Sort != nil {
			q.Set("sort", *opts.Sort)
		}

		rawUrl.RawQuery = q.Encode()
	}

	rawUrl.RawQuery += "&" + buildSearchParams(sq)

	url := rawUrl.String()
	req, err := s.client.NewRequest(http.MethodGet, url, nil)
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
	rawUrl := s.client.baseUrl.JoinPath("search", "users")
	if opts != nil {
		q := rawUrl.Query()
		opts.paginateQuery(q)
		if opts.Order != nil {
			q.Set("order", *opts.Order)
		}
		if opts.Sort != nil {
			q.Set("sort", *opts.Sort)
		}

		rawUrl.RawQuery = q.Encode()
	}

	rawUrl.RawQuery += "&" + buildSearchParams(sq)

	url := rawUrl.String()
	req, err := s.client.NewRequest(http.MethodGet, url, nil)
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

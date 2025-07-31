package github

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

// SearchService provides access to search API methods.
type SearchService struct {
	client *Client
}

// Search represents the response from a search GitHub API request.
// The type parameter T allows this struct to be used with different
// resource types like Repository or User.
// GitHub API docs: https://docs.github.com/en/rest/search/search
type Search[T Repository | User] struct {
	TotalCount        int  `json:"total_count"`
	IncompleteResults bool `json:"incomplete_results"`
	Items             []*T `json:"items"`
}

// SearchOptions specifies the optional parameters for search operations.
// GitHub API docs: https://docs.github.com/en/rest/search/search
type SearchOptions struct {
	*ListOptions
	Sort  *string
	Order *string
}

// Repositories searches for repositories based on the provided query.
// This method allows you to search repositories using GitHub's code search
// syntax. You can filter by various criteria such as language, stars,
// forks, and more. The results can be sorted and paginated using
// the SearchOptions parameter.
func (s *SearchService) Repositories(ctx context.Context, sq string, opts *SearchOptions) (*Search[Repository], *Response, error) {
	path := "search/repositories"

	v := url.Values{}

	if opts != nil {
		if opts.ListOptions != nil {
			opts.Apply(v)
		}

		if opts.Order != nil {
			v.Set("order", *opts.Order)
		}

		if opts.Sort != nil {
			v.Set("sort", *opts.Sort)
		}

		if len(v) != 0 {
			path += "?" + v.Encode()
		}
	}

	if len(v) != 0 {
		path += "&" + buildSearchParams(sq)
	} else {
		path += "?" + buildSearchParams(sq)
	}

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	search := new(Search[Repository])
	
	resp, err := s.client.Do(ctx, req, search)
	if err != nil {
		return nil, resp, err
	}

	return search, resp, nil
}

// Users searches for users based on the provided query.
// This method allows you to search for GitHub users using various
// search criteria such as username, full name, location, and followers.
// The results can be sorted by different fields and paginated using
// the SearchOptions parameter.
func (s *SearchService) Users(ctx context.Context, sq string, opts *SearchOptions) (*Search[User], *Response, error) {
	path := "search/users"

	v := url.Values{}

	if opts != nil {
		if opts.ListOptions != nil {
			opts.Apply(v)
		}

		if opts.Order != nil {
			v.Set("order", *opts.Order)
		}

		if opts.Sort != nil {
			v.Set("sort", *opts.Sort)
		}

		if len(v) != 0 {
			path += "?" + v.Encode()
		}
	}

	if len(v) != 0 {
		path += "&" + buildSearchParams(sq)
	} else {
		path += "?" + buildSearchParams(sq)
	}

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	search := new(Search[User])

	resp, err := s.client.Do(ctx, req, search)
	if err != nil {
		return nil, resp, err
	}

	return search, resp, nil
}

func buildSearchParams(s string) string {
	trimmed := strings.TrimSpace(s)
	chars := strings.Split(trimmed, " ")
	encodedQuery := strings.Join(chars, "+")

	return "q=" + encodedQuery
}

package github

import (
	"net/url"
	"strconv"
)

// ListOptions specifies the optional parameters for pagination.
// This struct is used to control the pagination behavior of list operations
// across various API endpoints. It allows you to specify which page of
// results to retrieve and how many items per page to return.
type ListOptions struct {
	// Page specifies the page number of results to retrieve
	Page int

	// PerPage specifies the number of items per page.
	PerPage int
}

func (lo *ListOptions) Apply(v url.Values) {
	if lo.Page != 0 {
		v.Set("page", strconv.Itoa(lo.Page))
	}
	if lo.PerPage != 0 {
		v.Set("per_page", strconv.Itoa(lo.PerPage))
	}
}

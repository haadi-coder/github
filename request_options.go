package github

import (
	"net/url"
	"strconv"
)

type ListOptions struct {
	Page    int
	PerPage int
}

func (lo *ListOptions) paginateQuery(q url.Values) {
	if lo.Page != 0 {
		q.Set("page", strconv.Itoa(lo.Page))
	}
	if lo.PerPage != 0 {
		q.Set("per_page", strconv.Itoa(lo.PerPage))
	}
}

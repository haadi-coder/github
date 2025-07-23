package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSearchParams(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"go", "q=go"},
		{"go lang", "q=go+lang"},
		{"  go lang  ", "q=go+lang"},
		{"", "q="},
	}

	for _, tt := range tests {
		result := buildSearchParams(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestSearch_Repositories(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/search/repositories", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.RawQuery, "q=go+lang")
		assert.Contains(t, r.URL.RawQuery, "sort=stars")
		assert.Contains(t, r.URL.RawQuery, "order=desc")
		assert.Contains(t, r.URL.RawQuery, "page=2")
		assert.Contains(t, r.URL.RawQuery, "per_page=50")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"total_count":100,"incomplete_results":false,"items":[{"id":1,"name":"go-lang"}]}`))
	}))
	defer ts.Close()

	sort := "stars"
	order := "desc"
	opts := &SearchOptions{
		ListOptions: &ListOptions{
			Page:    2,
			PerPage: 50,
		},
		Sort:  &sort,
		Order: &order,
	}
	client := NewClient(WithBaseURl(ts.URL))

	result, err := client.Search.Repositories(context.Background(), "go lang", opts)
	require.NoError(t, err)
	assert.Equal(t, 100, result.TotalCount)
	assert.False(t, result.IncompleteResults)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, "go-lang", result.Items[0].Name)
}

func TestSearch_Users(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/search/users", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.RawQuery, "q=john")
		assert.Contains(t, r.URL.RawQuery, "page=1")
		assert.Contains(t, r.URL.RawQuery, "per_page=30")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"total_count":50,"incomplete_results":false,"items":[{"id":1,"login":"john_doe"}]}`))
	}))
	defer ts.Close()

	opts := &SearchOptions{
		ListOptions: &ListOptions{
			Page:    1,
			PerPage: 30,
		},
	}
	client := NewClient(WithBaseURl(ts.URL))

	result, err := client.Search.Users(context.Background(), "john", opts)
	require.NoError(t, err)
	assert.Equal(t, 50, result.TotalCount)
	assert.False(t, result.IncompleteResults)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, "john_doe", result.Items[0].Login)
}

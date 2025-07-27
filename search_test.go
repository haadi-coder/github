package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearch_Repositories(t *testing.T) {
	sort := "stars"
	order := "desc"
	tests := []struct {
		name         string
		searchQuery  string
		opts         *SearchOptions
		expectedURL  string
		responseBody string
		expectError  bool
		expected     *Search[Repository]
	}{
		{
			name:        "Полный поиск репозиториев",
			searchQuery: "go lang",
			opts: &SearchOptions{
				ListOptions: &ListOptions{Page: 2, PerPage: 50},
				Sort:        &sort,
				Order:       &order,
			},
			expectedURL: "/search/repositories?order=desc&page=2&per_page=50&sort=stars&q=go+lang",
			responseBody: `{
                "total_count":100,
                "incomplete_results":false,
                "items":[{"id":1,"name":"go-lang"}]
            }`,
			expected: &Search[Repository]{
				TotalCount:        100,
				IncompleteResults: false,
				Items:             []*Repository{{ID: 1, Name: "go-lang"}},
			},
		},
		{
			name:        "Пустой поиск",
			searchQuery: "",
			opts:        nil,
			expectedURL: "/search/repositories?q=",
			responseBody: `{
                "total_count":0,
                "incomplete_results":false,
                "items":[]
            }`,
			expected: &Search[Repository]{
				TotalCount:        0,
				IncompleteResults: false,
				Items:             []*Repository{},
			},
		},
		{
			name:        "Ошибка сервера",
			searchQuery: "go",
			opts:        nil,
			expectedURL: "/search/repositories?q=go",
			responseBody: `{
                "message":"Internal Server Error"
            }`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedURL, r.URL.String())
				assert.Equal(t, "GET", r.Method)

				w.Header().Set("Content-Type", "application/json")
				if tt.expectError {
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					w.WriteHeader(http.StatusOK)
				}
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client, err := NewClient(WithBaseURL(ts.URL))
			require.NoError(t, err)

			result, err := client.Search.Repositories(context.Background(), tt.searchQuery, tt.opts)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Equal(t, tt.expected.TotalCount, result.TotalCount)
			assert.Equal(t, tt.expected.IncompleteResults, result.IncompleteResults)
			assert.Len(t, result.Items, len(tt.expected.Items))
			if len(tt.expected.Items) > 0 {
				assert.Equal(t, tt.expected.Items, result.Items)
			}
		})
	}
}

func TestSearch_Users(t *testing.T) {
	tests := []struct {
		name         string
		searchQuery  string
		opts         *SearchOptions
		expectedURL  string
		responseBody string
		expectError  bool
		expected     *Search[User]
	}{
		{
			name:        "Users search",
			searchQuery: "john",
			opts: &SearchOptions{
				ListOptions: &ListOptions{Page: 1, PerPage: 30},
			},
			expectedURL: "/search/users?page=1&per_page=30&q=john",
			responseBody: `{
                "total_count":50,
                "incomplete_results":false,
                "items":[{"id":1,"login":"john_doe"}]
            }`,
			expected: &Search[User]{
				TotalCount:        50,
				IncompleteResults: false,
				Items:             []*User{{ID: 1, Login: "john_doe"}},
			},
		},
		{
			name:        "Empty query",
			searchQuery: "",
			opts:        nil,
			expectedURL: "/search/users?q=",
			responseBody: `{
                "total_count":0,
                "incomplete_results":false,
                "items":[]
            }`,
			expected: &Search[User]{
				TotalCount:        0,
				IncompleteResults: false,
				Items:             []*User{},
			},
		},
		{
			name:        "Server error",
			searchQuery: "john",
			opts:        nil,
			expectedURL: "/search/users?q=john",
			responseBody: `{
                "message":"Internal Server Error"
            }`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedURL, r.URL.String())
				assert.Equal(t, "GET", r.Method)

				w.Header().Set("Content-Type", "application/json")
				if tt.expectError {
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					w.WriteHeader(http.StatusOK)
				}
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client, err := NewClient(WithBaseURL(ts.URL))
			require.NoError(t, err)

			result, err := client.Search.Users(context.Background(), tt.searchQuery, tt.opts)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Equal(t, tt.expected.TotalCount, result.TotalCount)
			assert.Equal(t, tt.expected.IncompleteResults, result.IncompleteResults)
			assert.Len(t, result.Items, len(tt.expected.Items))
			if len(tt.expected.Items) > 0 {
				assert.Equal(t, tt.expected.Items, result.Items)
			}
		})
	}
}

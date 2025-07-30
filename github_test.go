package github

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLinkHeader(t *testing.T) {
	tests := []struct {
		name          string
		linkHeader    string
		expectedPrev  int
		expectedNext  int
		expectedFirst int
		expectedLast  int
		expectError   bool
	}{
		{
			name: "first page (GitHub example)",
			linkHeader: `<https://api.github.com/repositories/123/issues?page=2&per_page=30>; rel="next", ` +
				`<https://api.github.com/repositories/123/issues?page=1&per_page=30>; rel="first", ` +
				`<https://api.github.com/repositories/123/issues?page=5&per_page=30>; rel="last"`,

			expectedNext:  2,
			expectedFirst: 1,
			expectedLast:  5,
		},
		{
			name: "last page (GitHub example)",
			linkHeader: `<https://api.github.com/repositories/123/issues?page=4&per_page=30>; rel="prev", ` +
				`<https://api.github.com/repositories/123/issues?page=1&per_page=30>; rel="first", ` +
				`<https://api.github.com/repositories/123/issues?page=5&per_page=30>; rel="last"`,
			expectedPrev:  4,
			expectedFirst: 1,
			expectedLast:  5,
		},
		{
			name: "middle page (GitHub example)",
			linkHeader: `<https://api.github.com/repositories/123/issues?page=3&per_page=30>; rel="next", ` +
				`<https://api.github.com/repositories/123/issues?page=1&per_page=30>; rel="prev", ` +
				`<https://api.github.com/repositories/123/issues?page=1&per_page=30>; rel="first", ` +
				`<https://api.github.com/repositories/123/issues?page=5&per_page=30>; rel="last"`,
			expectedNext:  3,
			expectedPrev:  1,
			expectedFirst: 1,
			expectedLast:  5,
		},
		{
			name: "single page (no pagination)",
			linkHeader: `<https://api.github.com/repositories/123/issues?page=1&per_page=30>; rel="first", ` +
				`<https://api.github.com/repositories/123/issues?page=1&per_page=30>; rel="last"`,
			expectedFirst: 1,
			expectedLast:  1,
		},
		{
			name:        "invalid page number",
			linkHeader:  `<https://api.github.com/repositories/123/issues?page=invalid>; rel="next"`,
			expectError: true,
		},
		{
			name:         "page zero should be skipped",
			linkHeader:   `<https://api.github.com/repositories/123/issues?page=0>; rel="next"`,
			expectedNext: 0,
		},
		{
			name:        "empty link header",
			linkHeader:  "",
			expectError: false,
		},
		{
			name:        "malformed URL",
			linkHeader:  `<invalid-url>; rel="next"`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &Response{
				Response: &http.Response{
					Header: make(http.Header),
				},
			}
			if tt.linkHeader != "" {
				resp.Header.Set("Link", tt.linkHeader)
			}

			err := parseLinkHeader(resp)

			if (err != nil) != tt.expectError {
				t.Fatalf("expected error: %v, got: %v", tt.expectError, err)
			}

			if !tt.expectError {
				assert.Equal(t, resp.PreviousPage, tt.expectedPrev)
				assert.Equal(t, resp.NextPage, tt.expectedNext)
				assert.Equal(t, resp.FirstPage, tt.expectedFirst)
				assert.Equal(t, resp.LastPage, tt.expectedLast)
			}
		})
	}
}

func TestBuildErrorResponse(t *testing.T) {
	cases := []struct {
		name             string
		statusCode       int
		body             string
		expectedMsg      string
		expectedDocURL   string
		expectedErrors   []APIErrorDetail
		expectedErrIsNil bool
	}{
		{
			name:       "Valid JSON with all fields",
			statusCode: http.StatusUnauthorized,
			body: `{
                "message": "Unauthorized access",
                "documentation_url": "https://example.com/docs ",
                "errors": [
                    {
                        "code": "401",
                        "resource": "auth",
                        "field": "token"
                    }
                ]
            }`,
			expectedMsg:    "Unauthorized access",
			expectedDocURL: "https://example.com/docs ",
			expectedErrors: []APIErrorDetail{
				{Code: "401", Resource: "auth", Field: "token"},
			},
			expectedErrIsNil: false,
		},
		{
			name:       "Valid JSON with message and doc_url",
			statusCode: http.StatusBadRequest,
			body: `{
                "message": "Validation failed",
                "documentation_url": "https://example.com/validation "
            }`,
			expectedMsg:      "Validation failed",
			expectedDocURL:   "https://example.com/validation ",
			expectedErrIsNil: false,
		},
		{
			name:       "Valid JSON with message only",
			statusCode: http.StatusNotFound,
			body: `{
                "message": "Not found"
            }`,
			expectedMsg:      "Not found",
			expectedErrIsNil: false,
		},
		{
			name:             "Invalid JSON",
			statusCode:       http.StatusBadRequest,
			body:             `invalid-json`,
			expectedMsg:      "request failed with status 400",
			expectedErrIsNil: false,
		},
		{
			name:             "Empty body",
			statusCode:       http.StatusInternalServerError,
			body:             "",
			expectedMsg:      "request failed with status 500",
			expectedErrIsNil: false,
		},
		{
			name:             "Non-JSON body",
			statusCode:       http.StatusNotFound,
			body:             "Not Found",
			expectedMsg:      "request failed with status 404",
			expectedErrIsNil: false,
		},
		{
			name:       "JSON with empty errors array",
			statusCode: http.StatusConflict,
			body: `{
                "message": "Conflict detected",
                "errors": []
            }`,
			expectedMsg:      "Conflict detected",
			expectedErrors:   []APIErrorDetail{},
			expectedErrIsNil: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var resp *http.Response
			if tt.statusCode != 0 || tt.body != "" {
				body := io.NopCloser(bytes.NewBufferString(tt.body))
				resp = &http.Response{
					StatusCode: tt.statusCode,
					Body:       body,
				}
			}

			err := newAPIError(resp)

			if !tt.expectedErrIsNil {
				assert.Error(t, err)
				e := err.(*APIError)
				assert.Equal(t, tt.expectedMsg, e.Message)
				assert.Equal(t, tt.expectedDocURL, e.DocumentationURL)

				if len(tt.expectedErrors) > 0 {
					assert.Equal(t, len(tt.expectedErrors), len(e.Errors))
					assert.Equal(t, tt.expectedErrors, e.Errors)
				} else {
					assert.Empty(t, e.Errors)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDo_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "60")
		w.Header().Set("X-RateLimit-Remaining", "59")
		w.Header().Set("X-RateLimit-Used", "1")
		w.Header().Set("X-RateLimit-Reset", "1717029203")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"key": "value"})
	}))
	defer ts.Close()

	client, _ := NewClient()
	client.baseURL, _ = url.Parse(ts.URL)

	req, err := client.NewRequest("GET", ts.URL, nil)
	require.NoError(t, err)

	ctx := context.Background()
	resp, err := client.Do(ctx, req, &map[string]string{})

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	result := &map[string]string{}
	_ = json.NewDecoder(resp.Body).Decode(result)

	assert.Equal(t, 60, resp.Limit)
	assert.Equal(t, 59, resp.Remaining)
	assert.Equal(t, 1, resp.Used)
	assert.Equal(t, int64(1717029203), resp.Reset)
}

func TestDo_RateLimitExceeded_NoRetry(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "1")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Used", "1")
		w.Header().Set("X-RateLimit-Reset", "1717029203")
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"message": "Too Many Requests"}`, http.StatusTooManyRequests)
	}))
	defer ts.Close()

	client, _ := NewClient()
	client.baseURL, _ = url.Parse(ts.URL)
	client.rateLimitRetry = false

	req, err := client.NewRequest("GET", ts.URL, nil)
	require.NoError(t, err)

	ctx := context.Background()
	resp, err := client.Do(ctx, req, nil)

	require.Error(t, err)
	errorResp, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusTooManyRequests, errorResp.StatusCode)
	assert.Equal(t, "Too Many Requests", errorResp.Message)
	assert.Equal(t, 1, resp.Limit)
	assert.Equal(t, 0, resp.Remaining)
}

func TestDo_ContextTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "should not be called", http.StatusInternalServerError)
	}))
	defer ts.Close()

	client, _ := NewClient()
	client.baseURL, _ = url.Parse(ts.URL)

	req, err := client.NewRequest("GET", ts.URL, nil)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	resp, err := client.Do(ctx, req, nil)

	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Nil(t, resp)
}

func TestDo_LinkHeaderParsing(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		link := `<https://api.github.com/resource?page=2>; rel="next", <https://api.github.com/resource?page=1>; rel="prev"`
		w.Header().Set("Link", link)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client, _ := NewClient()
	client.baseURL, _ = url.Parse(ts.URL)

	req, err := client.NewRequest("GET", ts.URL, nil)
	require.NoError(t, err)

	ctx := context.Background()
	resp, err := client.Do(ctx, req, nil)

	require.NoError(t, err)
	assert.Equal(t, 2, resp.NextPage)
	assert.Equal(t, 1, resp.PreviousPage)
}

func TestDo_HooksCalled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client, _ := NewClient()
	client.baseURL, _ = url.Parse(ts.URL)

	req, err := client.NewRequest("GET", ts.URL, nil)
	require.NoError(t, err)

	ctx := context.Background()
	resp, err := client.Do(ctx, req, nil)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDo_RetryOnRateLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "60")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", "1717029203")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	client, _ := NewClient(WithBaseURL(ts.URL), WithRateLimitRetry(true), WithRetryMax(3))

	req, err := client.NewRequest("GET", ts.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(context.Background(), req, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "max retry attempts")
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
}

func TestDo_InvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer ts.Close()

	client, _ := NewClient(WithBaseURL(ts.URL))

	req, err := client.NewRequest("GET", ts.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(context.Background(), req, &map[string]string{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid character")
	assert.NotNil(t, resp)
}

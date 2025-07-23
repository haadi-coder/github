package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWaitRateLimit(t *testing.T) {
	cases := []struct {
		name         string
		retryWaitMin int
		retryWaitMax int
		rl           *RateLimit
		attempt      int
		expectedWait time.Duration
	}{
		{
			name:         "Reset в будущем",
			retryWaitMin: 5,
			retryWaitMax: 60,
			rl: &RateLimit{
				Reset: time.Now().Add(10 * time.Second).Unix(),
			},
			attempt:      2,
			expectedWait: 10 * time.Second,
		},
		{
			name:         "Reset в прошлом",
			retryWaitMin: 5,
			retryWaitMax: 60,
			rl: &RateLimit{
				Reset: time.Now().Add(-10 * time.Second).Unix(),
			},
			attempt:      3,
			expectedWait: time.Second,
		},
		{
			name:         "Reset == 0, attempt=0",
			retryWaitMin: 5,
			retryWaitMax: 60,
			rl:           &RateLimit{Reset: 0},
			attempt:      0,
			expectedWait: 5 * time.Second, // 5 * 2^0 = 5
		},
		{
			name:         "Reset == 0, attempt=1",
			retryWaitMin: 5,
			retryWaitMax: 60,
			rl:           &RateLimit{Reset: 0},
			attempt:      1,
			expectedWait: 10 * time.Second, // 5 * 2^1 = 10
		},
		{
			name:         "Reset == 0, attempt=100 (ограничение retryWaitMax)",
			retryWaitMin: 5,
			retryWaitMax: 60,
			rl:           &RateLimit{Reset: 0},
			attempt:      100,
			expectedWait: 60 * time.Second, // max = 60
		},
		{
			name:         "Reset == 0, retryWaitMin по умолчанию",
			retryWaitMin: 0, // default = 5
			retryWaitMax: 0, // default = 60
			rl:           &RateLimit{Reset: 0},
			attempt:      2,
			expectedWait: 20 * time.Second, // 5 * 2^2 = 20
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := NewClient(
				WithRetryWaitMin(float64(tc.retryWaitMin)),
				WithRetryWaitMax(float64(tc.retryWaitMax)),
			)

			start := time.Now()
			client.waitRateLimit(tc.rl, tc.attempt)
			end := time.Now()

			elapsed := end.Sub(start)

			assert.InDelta(t, tc.expectedWait.Seconds(), elapsed.Seconds(), 1)
		})
	}
}

func TestParseLinkHeader(t *testing.T) {
	cases := []struct {
		name       string
		linkHeader string
		want       Response
		wantErr    bool
	}{
		{
			name:       "Пустая строка",
			linkHeader: "",
			wantErr:    true,
		},
		{
			name:       "Одна корректная ссылка next",
			linkHeader: `<https://api.example.com?page=2>; rel="next"`,
			want: Response{
				NextPage: 2,
			},
			wantErr: false,
		},
		{
			name:       "Одна корректная ссылка prev",
			linkHeader: `<https://api.example.com?page=1>; rel="prev"`,
			want: Response{
				PreviousPage: 1,
			},
			wantErr: false,
		},
		{
			name:       "Одна корректная ссылка first",
			linkHeader: `<https://api.example.com?page=1>; rel="first"`,
			want: Response{
				FirstPage: 1,
			},
			wantErr: false,
		},
		{
			name:       "Одна корректная ссылка last",
			linkHeader: `<https://api.example.com?page=5>; rel="last"`,
			want: Response{
				LastPage: 5,
			},
			wantErr: false,
		},
		{
			name:       "Несколько ссылок",
			linkHeader: `<https://api.example.com?page=1>; rel="first", <https://api.example.com?page=3>; rel="next", <https://api.example.com?page=1>; rel="prev", <https://api.example.com?page=5>; rel="last"`,
			want: Response{
				FirstPage:    1,
				NextPage:     3,
				PreviousPage: 1,
				LastPage:     5,
			},
			wantErr: false,
		},
		{
			name:       "Некорректный URL",
			linkHeader: `<https://api.example.com?invalid>; rel="next"`,
			wantErr:    true,
		},
		{
			name:       "Отсутствует параметр page",
			linkHeader: `<https://api.example.com?limit=10>; rel="next"`,
			wantErr:    true,
		},
		{
			name:       "Некорректное значение page",
			linkHeader: `<https://api.example.com?page=abc>; rel="next"`,
			wantErr:    true,
		},
		{
			name:       "Пустое значение page",
			linkHeader: `<https://api.example.com?page=>; rel="next"`,
			wantErr:    true,
		},
		{
			name:       "Дублирующийся rel",
			linkHeader: `<https://api.example.com?page=2>; rel="next", <https://api.example.com?page=3>; rel="next"`,
			want: Response{
				NextPage: 3, // последнее значение должно перезаписать предыдущее
			},
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			resp := &Response{}
			err := parseLinkHeader(resp, tt.linkHeader)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, *resp)
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
		expectedDocUrl   string
		expectedErrors   []struct{ Code, Resource, Field string }
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
			expectedDocUrl: "https://example.com/docs ",
			expectedErrors: []struct{ Code, Resource, Field string }{
				{"401", "auth", "token"},
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
			expectedDocUrl:   "https://example.com/validation ",
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
			expectedMsg:      "Request failed with status 400",
			expectedErrIsNil: false,
		},
		{
			name:             "Empty body",
			statusCode:       http.StatusInternalServerError,
			body:             "",
			expectedMsg:      "Request failed with status 500",
			expectedErrIsNil: false,
		},
		{
			name:             "Non-JSON body",
			statusCode:       http.StatusNotFound,
			body:             "Not Found",
			expectedMsg:      "Request failed with status 404",
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
			expectedErrors:   []struct{ Code, Resource, Field string }{},
			expectedErrIsNil: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var hr *http.Response
			if tt.statusCode != 0 || tt.body != "" {
				body := io.NopCloser(bytes.NewBufferString(tt.body))
				hr = &http.Response{
					StatusCode: tt.statusCode,
					Body:       body,
				}
			}

			err := buildErrorResponse(hr)

			if !tt.expectedErrIsNil {
				assert.Error(t, err)
				e := err.(*ErrorResponse)
				assert.Equal(t, tt.expectedMsg, e.Message)
				assert.Equal(t, tt.expectedDocUrl, e.DocumentationUrl)

				if len(tt.expectedErrors) > 0 {
					assert.Equal(t, len(tt.expectedErrors), len(e.Errors))
					for i := range tt.expectedErrors {
						assert.Equal(t, tt.expectedErrors[i].Code, e.Errors[i].Code)
						assert.Equal(t, tt.expectedErrors[i].Resource, e.Errors[i].Resource)
						assert.Equal(t, tt.expectedErrors[i].Field, e.Errors[i].Field)
					}
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

	client := NewClient()
	client.baseUrl, _ = url.Parse(ts.URL)

	req, err := client.NewRequest("GET", ts.URL, nil)
	require.NoError(t, err)

	var result map[string]string
	resp, err := client.Do(context.Background(), req, &result)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "value", result["key"])
	assert.Equal(t, 60, resp.RateLimit.Limit)
	assert.Equal(t, 59, resp.RateLimit.Remaining)
	assert.Equal(t, 1, resp.RateLimit.Used)
	assert.Equal(t, int64(1717029203), resp.RateLimit.Reset)
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

	client := NewClient()
	client.baseUrl, _ = url.Parse(ts.URL)
	client.rateLimitRetry = false

	req, err := client.NewRequest("GET", ts.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(context.Background(), req, nil)
	require.Error(t, err)

	errorResp, ok := err.(*ErrorResponse)
	require.True(t, ok)
	assert.Equal(t, http.StatusTooManyRequests, errorResp.StatusCode)
	assert.Equal(t, "Too Many Requests", errorResp.Message)
	assert.Equal(t, 1, resp.RateLimit.Limit)
	assert.Equal(t, 0, resp.RateLimit.Remaining)
}

func TestDo_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "should not be called", http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := NewClient()
	client.baseUrl, _ = url.Parse(ts.URL)

	req, err := client.NewRequest("GET", ts.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(ctx, req, nil)
	require.Error(t, err)
	assert.Equal(t, ctx.Err(), err)
	assert.Nil(t, resp)
}

func TestDo_LinkHeaderParsing(t *testing.T) {
	link := `<https://api.github.com/resource?page=2>; rel="next", < https://api.github.com/resource?page=1>; rel="prev"`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Link", link)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient()
	client.baseUrl, _ = url.Parse(ts.URL)

	req, err := client.NewRequest("GET", ts.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(context.Background(), req, nil)
	require.NoError(t, err)

	assert.Equal(t, 2, resp.NextPage)
	assert.Equal(t, 1, resp.PreviousPage)
}

func TestDo_HooksCalled(t *testing.T) {
	calledRequest := false
	calledResponse := false

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient()
	client.baseUrl, _ = url.Parse(ts.URL)
	client.requestHook = func(r *http.Request) {
		calledRequest = true
	}
	client.responseHook = func(r *http.Response) {
		calledResponse = true
	}

	req, err := client.NewRequest("GET", ts.URL, nil)
	require.NoError(t, err)

	_, err = client.Do(context.Background(), req, nil)
	require.NoError(t, err)

	assert.True(t, calledRequest)
	assert.True(t, calledResponse)
}

func TestDo_RetryOnRateLimit(t *testing.T) {
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts == 0 {
			w.Header().Set("X-RateLimit-Limit", "60")
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", "1717029203")
			w.WriteHeader(http.StatusTooManyRequests)
			attempts++
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer ts.Close()

	url, _ := url.Parse(ts.URL)
	client := NewClient(WithBaseURl(url.String()), WithRateLimitRetry(true))

	req, err := client.NewRequest("GET", ts.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(context.Background(), req, nil)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 1, attempts)
}

func TestDo_InvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer ts.Close()

	client := NewClient()
	client.baseUrl, _ = url.Parse(ts.URL)

	req, err := client.NewRequest("GET", ts.URL, nil)
	require.NoError(t, err)

	var result map[string]string
	resp, err := client.Do(context.Background(), req, &result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid character")
	assert.NotNil(t, resp)
}

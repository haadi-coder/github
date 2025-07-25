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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWaitRateLimit(t *testing.T) {
	cases := []struct {
		name         string
		retryWaitMin time.Duration
		retryWaitMax time.Duration
		rl           *RateLimit
		attempt      int
		expectedWait time.Duration
	}{
		{
			name:         "Reset в будущем",
			retryWaitMin: 5 * time.Second,
			retryWaitMax: 60 * time.Second,
			rl: &RateLimit{
				Reset: time.Now().Add(10 * time.Second).Unix(),
			},
			attempt:      2,
			expectedWait: 10 * time.Second,
		},
		{
			name:         "Reset в прошлом",
			retryWaitMin: 5 * time.Second,
			retryWaitMax: 60 * time.Second,
			rl: &RateLimit{
				Reset: time.Now().Add(-10 * time.Second).Unix(),
			},
			attempt:      3,
			expectedWait: time.Second,
		},
		{
			name:         "Reset == 0, attempt=0",
			retryWaitMin: 5 * time.Second,
			retryWaitMax: 60 * time.Second,
			rl:           &RateLimit{Reset: 0},
			attempt:      0,
			expectedWait: 5 * time.Second, // 5 * 2^0 = 5
		},
		{
			name:         "Reset == 0, attempt=1",
			retryWaitMin: 5 * time.Second,
			retryWaitMax: 60 * time.Second,
			rl:           &RateLimit{Reset: 0},
			attempt:      1,
			expectedWait: 10 * time.Second, // 5 * 2^1 = 10
		},
		{
			name:         "Reset == 0, attempt=100 (ограничение retryWaitMax)",
			retryWaitMin: 5 * time.Second,
			retryWaitMax: 20 * time.Second,
			rl:           &RateLimit{Reset: 0},
			attempt:      100,
			expectedWait: 20 * time.Second, // max = 10
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
				WithRetryWaitMin(tc.retryWaitMin),
				WithRetryWaitMax(tc.retryWaitMax),
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

			err := ApiError(hr)

			if !tt.expectedErrIsNil {
				assert.Error(t, err)
				e := err.(*APIError)
				assert.Equal(t, tt.expectedMsg, e.Message)
				assert.Equal(t, tt.expectedDocUrl, e.DocumentationUrl)

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

func TestDo(t *testing.T) {
	tests := []struct {
		name           string
		setupServer    func() *httptest.Server
		setupClient    func(client *Client)
		setupContext   func() context.Context
		target         any
		wantErr        bool
		wantStatusCode int
		validate       func(t *testing.T, resp *Response, err error, target any)
	}{
		{
			name: "Success",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("X-RateLimit-Limit", "60")
					w.Header().Set("X-RateLimit-Remaining", "59")
					w.Header().Set("X-RateLimit-Used", "1")
					w.Header().Set("X-RateLimit-Reset", "1717029203")
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(map[string]string{"key": "value"})
				}))
			},
			setupClient: func(client *Client) {},
			setupContext: func() context.Context {
				return context.Background()
			},
			target:         &map[string]string{},
			wantErr:        false,
			wantStatusCode: http.StatusOK,
			validate: func(t *testing.T, resp *Response, err error, target interface{}) {
				require.NoError(t, err)
				result := target.(*map[string]string)
				assert.Equal(t, "value", (*result)["key"])
				assert.Equal(t, 60, resp.Limit)
				assert.Equal(t, 59, resp.Remaining)
				assert.Equal(t, 1, resp.Used)
				assert.Equal(t, int64(1717029203), resp.Reset)
			},
		},
		{
			name: "RateLimitExceeded_NoRetry",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("X-RateLimit-Limit", "1")
					w.Header().Set("X-RateLimit-Remaining", "0")
					w.Header().Set("X-RateLimit-Used", "1")
					w.Header().Set("X-RateLimit-Reset", "1717029203")
					w.Header().Set("Content-Type", "application/json")
					http.Error(w, `{"message": "Too Many Requests"}`, http.StatusTooManyRequests)
				}))
			},
			setupClient: func(client *Client) {
				client.rateLimitRetry = false
			},
			setupContext: func() context.Context {
				return context.Background()
			},
			target:         nil,
			wantErr:        true,
			wantStatusCode: http.StatusTooManyRequests,
			validate: func(t *testing.T, resp *Response, err error, target interface{}) {
				require.Error(t, err)
				errorResp, ok := err.(*APIError)
				require.True(t, ok)
				assert.Equal(t, http.StatusTooManyRequests, errorResp.StatusCode)
				assert.Equal(t, "Too Many Requests", errorResp.Message)
				assert.Equal(t, 1, resp.Limit)
				assert.Equal(t, 0, resp.Remaining)
			},
		},
		{
			name: "ContextTimeout",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "should not be called", http.StatusInternalServerError)
				}))
			},
			setupClient: func(client *Client) {},
			setupContext: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			target:         nil,
			wantErr:        true,
			wantStatusCode: 0,
			validate: func(t *testing.T, resp *Response, err error, target interface{}) {
				require.Error(t, err)
				assert.Equal(t, context.Canceled, err)
				assert.Nil(t, resp)
			},
		},
		{
			name: "LinkHeaderParsing",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					link := `<https://api.github.com/resource?page=2>; rel="next", < https://api.github.com/resource?page=1>; rel="prev"`
					w.Header().Set("Link", link)
					w.WriteHeader(http.StatusOK)
				}))
			},
			setupClient: func(client *Client) {},
			setupContext: func() context.Context {
				return context.Background()
			},
			target:         nil,
			wantErr:        false,
			wantStatusCode: http.StatusOK,
			validate: func(t *testing.T, resp *Response, err error, target interface{}) {
				require.NoError(t, err)
				assert.Equal(t, 2, resp.NextPage)
				assert.Equal(t, 1, resp.PreviousPage)
			},
		},
		{
			name: "HooksCalled",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
			},
			setupClient: func(client *Client) {},
			setupContext: func() context.Context {
				return context.Background()
			},
			target:         nil,
			wantErr:        false,
			wantStatusCode: http.StatusOK,
			validate: func(t *testing.T, resp *Response, err error, target interface{}) {
				require.NoError(t, err)
			},
		},
		{
			name: "RetryOnRateLimit",
			setupServer: func() *httptest.Server {
				attempts := 0
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			},
			setupClient: func(client *Client) {
				client.rateLimitRetry = true
				client.retryMax = 10
			},
			setupContext: func() context.Context {
				return context.Background()
			},
			target:         nil,
			wantErr:        false,
			wantStatusCode: http.StatusOK,
			validate: func(t *testing.T, resp *Response, err error, target interface{}) {
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			},
		},
		{
			name: "InvalidJSON",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("invalid json"))
				}))
			},
			setupClient: func(client *Client) {},
			setupContext: func() context.Context {
				return context.Background()
			},
			target:         &map[string]string{},
			wantErr:        true,
			wantStatusCode: http.StatusOK,
			validate: func(t *testing.T, resp *Response, err error, target interface{}) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid character")
				assert.NotNil(t, resp)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := tt.setupServer()
			defer ts.Close()

			client := NewClient()
			client.baseURL, _ = url.Parse(ts.URL)
			tt.setupClient(client)

			req, err := client.NewRequest("GET", ts.URL, nil)
			require.NoError(t, err)

			ctx := tt.setupContext()
			resp, err := client.Do(ctx, req, tt.target)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.wantStatusCode != 0 && resp != nil {
				assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			}

			if tt.validate != nil {
				tt.validate(t, resp, err, tt.target)
			}
		})
	}
}

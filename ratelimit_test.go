package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildResponseRateLimit(t *testing.T) {
	tests := []struct {
		name              string
		headers           map[string]string
		expectedRateLimit *RateLimit
	}{
		{
			name: "All headers present",
			headers: map[string]string{
				rateLimitHeader:    "100",
				rateRemainigHeader: "50",
				rateUsedHeader:     "50",
				rateResetHeader:    "1717029203",
			},
			expectedRateLimit: &RateLimit{
				Limit:     100,
				Remaining: 50,
				Used:      50,
				Reset:     1717029203,
			},
		},
		{
			name:    "Headers missing",
			headers: map[string]string{},
			expectedRateLimit: &RateLimit{
				Limit:     0,
				Remaining: 0,
				Used:      0,
				Reset:     0,
			},
		},
		{
			name: "Invalid header values",
			headers: map[string]string{
				rateLimitHeader:    "abc",
				rateRemainigHeader: "def",
				rateResetHeader:    "ghi",
			},
			expectedRateLimit: &RateLimit{
				Limit:     0,
				Remaining: 0,
				Used:      0,
				Reset:     0,
			},
		},
		{
			name: "Only Limit",
			headers: map[string]string{
				rateLimitHeader: "200",
			},
			expectedRateLimit: &RateLimit{
				Limit:     200,
				Remaining: 0,
				Used:      0,
				Reset:     0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				for k, v := range tt.headers {
					w.Header().Set(k, v)
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			resp, err := http.Get(ts.URL)
			require.NoError(t, err)
			defer resp.Body.Close()

			response := buildResponse(resp)
			assert.Equal(t, tt.expectedRateLimit, response.RateLimit)
		})
	}
}

func TestRateLimitService_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rate_limit", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
            "resources": {
                "core": {"limit": 100, "remaining": 50, "used": 50, "reset": 1717029203},
                "search": {"limit": 200, "remaining": 100, "used": 100, "reset": 1717029203}
            },
            "rate": {"limit": 100, "remaining": 50, "used": 50, "reset": 1717029203}
        }`))
	}))
	defer ts.Close()

	client, err := NewClient(WithBaseURL(ts.URL))
	require.NoError(t, err)

	result, err := client.RateLimit.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result)

	expectedResult := &RateLimitResponse{
		Resources: &RateLimitResources{
			Core:   &RateLimit{Limit: 100, Remaining: 50, Used: 50, Reset: 1717029203},
			Search: &RateLimit{Limit: 200, Remaining: 100, Used: 100, Reset: 1717029203},
		},
		Rate: &RateLimit{Limit: 100, Remaining: 50, Used: 50, Reset: 1717029203},
	}

	assert.Equal(t, expectedResult, result)
}

func TestCalcBackoff(t *testing.T) {
	tests := []struct {
		name     string
		attempt  int
		waitMin  time.Duration
		waitMax  time.Duration
		reset    int64
		expected time.Duration
	}{

		{
			name:     "reset zero, first attempt",
			attempt:  0,
			waitMin:  1 * time.Second,
			waitMax:  30 * time.Second,
			reset:    0,
			expected: 1 * time.Second,
		},
		{
			name:     "reset zero, second attempt",
			attempt:  1,
			waitMin:  1 * time.Second,
			waitMax:  30 * time.Second,
			reset:    0,
			expected: 2 * time.Second,
		},
		{
			name:     "reset zero, capped at max",
			attempt:  10,
			waitMin:  1 * time.Second,
			waitMax:  30 * time.Second,
			reset:    0,
			expected: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calcBackoff(tt.waitMin, tt.waitMax, tt.attempt, &Response{RateLimit: &RateLimit{Reset: tt.reset}})
			assert.InDelta(t, tt.expected, result, 0.5)
		})
	}
}

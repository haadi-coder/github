package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRateLimit(t *testing.T) {
	tests := []struct {
		name              string
		headers           map[string]string
		expectedRateLimit *RateLimit
	}{
		{
			name: "Все заголовки присутствуют",
			headers: map[string]string{
				rateLimitHeader:    "100",
				rateRemainigHeader: "50",
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
			name:    "Отсутствуют заголовки",
			headers: map[string]string{},
			expectedRateLimit: &RateLimit{
				Limit:     0,
				Remaining: 0,
				Used:      0,
				Reset:     0,
			},
		},
		{
			name: "Некорректные значения в заголовках",
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
			name: "Только Limit",
			headers: map[string]string{
				rateLimitHeader: "200",
			},
			expectedRateLimit: &RateLimit{
				Limit:     200,
				Remaining: 0,
				Used:      200,
				Reset:     0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			rl := getRateLimit(resp)
			assert.Equal(t, tt.expectedRateLimit.Limit, rl.Limit)
			assert.Equal(t, tt.expectedRateLimit.Remaining, rl.Remaining)
			assert.Equal(t, tt.expectedRateLimit.Used, rl.Used)
			assert.Equal(t, tt.expectedRateLimit.Reset, rl.Reset)
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

	client := NewClient(WithBaseURl(ts.URL))

	result, err := client.RateLimit.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result)

	expectedResult := &RateLimitResponse{
		Resources: &Resources{
			Core:   &RateLimit{Limit: 100, Remaining: 50, Used: 50, Reset: 1717029203},
			Search: &RateLimit{Limit: 200, Remaining: 100, Used: 100, Reset: 1717029203},
		},
		Rate: &RateLimit{Limit: 100, Remaining: 50, Used: 50, Reset: 1717029203},
	}

	assert.Equal(t, expectedResult.Resources.Core.Limit, result.Resources.Core.Limit)
	assert.Equal(t, expectedResult.Resources.Core.Remaining, result.Resources.Core.Remaining)
	assert.Equal(t, expectedResult.Resources.Core.Used, result.Resources.Core.Used)
	assert.Equal(t, expectedResult.Resources.Core.Reset, result.Resources.Core.Reset)

	assert.Equal(t, expectedResult.Rate.Limit, result.Rate.Limit)
	assert.Equal(t, expectedResult.Rate.Remaining, result.Rate.Remaining)
	assert.Equal(t, expectedResult.Rate.Used, result.Rate.Used)
	assert.Equal(t, expectedResult.Rate.Reset, result.Rate.Reset)

}

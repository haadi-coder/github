package github

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimestamp_MarshalJSON(t *testing.T) {
	cases := []struct {
		name         string
		timestamp    Timestamp
		expectedJSON string
		expectError  bool
	}{
		{
			name:         "Zero time should return null",
			timestamp:    Timestamp{},
			expectedJSON: "null",
			expectError:  false,
		},
		{
			name:         "Valid time should be quoted",
			timestamp:    Timestamp{time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)},
			expectedJSON: `"2023-01-01T12:00:00Z"`,
			expectError:  false,
		},
		{
			name:         "Time with timezone",
			timestamp:    Timestamp{time.Date(2023, 6, 15, 14, 30, 45, 0, time.FixedZone("UTC+3", 3*60*60))},
			expectedJSON: `"2023-06-15T14:30:45+03:00"`,
			expectError:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := tc.timestamp.MarshalJSON()

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.JSONEq(t, tc.expectedJSON, string(result))
		})
	}
}

func TestTimestamp_UnmarshalJSON(t *testing.T) {
	cases := []struct {
		name         string
		jsonData     string
		expectedTime time.Time
		expectError  bool
	}{
		{
			name:         "Valid RFC3339 time",
			jsonData:     `"2023-01-01T12:00:00Z"`,
			expectedTime: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			expectError:  false,
		},
		{
			name:         "null value",
			jsonData:     "null",
			expectedTime: time.Time{},
			expectError:  false,
		},
		{
			name:         "Time with timezone offset",
			jsonData:     `"2023-06-15T14:30:45+03:00"`,
			expectedTime: time.Date(2023, 6, 15, 11, 30, 45, 0, time.UTC), // Конвертируется в UTC
			expectError:  false,
		},
		{
			name:         "Time with nanoseconds",
			jsonData:     `"2023-12-31T23:59:59.123456789Z"`,
			expectedTime: time.Date(2023, 12, 31, 23, 59, 59, 123456789, time.UTC),
			expectError:  false,
		},
		{
			name:        "Invalid time format",
			jsonData:    `"2023-01-01 12:00:00"`,
			expectError: true,
		},
		{
			name:        "Empty string",
			jsonData:    `""`,
			expectError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			
			var ts Timestamp

			err := ts.UnmarshalJSON([]byte(tc.jsonData))

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.True(t, tc.expectedTime.Equal(ts.Time))
		})
	}
}

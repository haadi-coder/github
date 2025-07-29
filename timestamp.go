package github

import (
	"time"
)

// Timestamp represents a time.Time value that can be marshaled and unmarshaled
// to and from JSON in RFC3339 format. This type is used for GitHub API timestamp
// fields that need proper JSON serialization handling.
type Timestamp struct {
	time.Time
}

func (t *Timestamp) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}

	return []byte(`"` + t.Format(time.RFC3339) + `"`), nil
}

func (t *Timestamp) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		t.Time = time.Time{}
		return nil
	}

	s := string(data[1 : len(data)-1])
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	t.Time = parsed
	return nil
}

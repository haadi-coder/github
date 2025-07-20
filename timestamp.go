package main

import (
	"strings"
	"time"
)

type Timestamp time.Time

func (t Timestamp) MarshalJSON() ([]byte, error) {
	tm := time.Time(t)
	if tm.IsZero() {
		return []byte{}, nil
	}

	return []byte(tm.Format(time.RFC3339)), nil
}

func (t *Timestamp) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), "\"")
	if s == "null" {
		*t = Timestamp(time.Time{})
		return nil
	}

	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	*t = Timestamp(parsed)
	return nil
}

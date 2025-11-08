package types

import (
	"encoding/json"
	"errors"
	"time"
)

// Time is a wrapper around time.Time that implements json.Marshaler and json.Unmarshaler.
// It parses string values in RFC3339 format (with RFC3339Nano as fallback)
type Time struct {
	time.Time `json:"-"`
}

func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Time.Format(time.RFC3339))
}

func (t *Time) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return nil
	}

	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case nil:
		return nil
	case string:
		if value == "" {
			return nil
		}

		var err error
		t.Time, err = time.Parse(time.RFC3339, value)
		if err != nil {
			// Try RFC3339Nano as fallback
			t.Time, err = time.Parse(time.RFC3339Nano, value)
			if err != nil {
				return err
			}
		}
		return nil
	default:
		return errors.New("invalid time: expected string in RFC3339 format")
	}
}

func (t Time) UnmarshalKind() string {
	return "string"
}

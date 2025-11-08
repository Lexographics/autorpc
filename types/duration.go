package types

import (
	"encoding/json"
	"time"
)

// Duration is a wrapper around time.Duration that implements json.Marshaler and json.Unmarshaler.
// It parses strings with time.ParseDuration(value) and numbers with time.Duration(value)
type Duration struct {
	time.Duration `json:"-"`
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
	case string:
		d.Duration, _ = time.ParseDuration(value)
	}
	return nil
}

func (d Duration) UnmarshalKind() string {
	return "string"
}

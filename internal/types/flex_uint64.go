package types

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// FlexUint64 is a uint64 that can be unmarshaled from either a JSON number or a JSON string.
type FlexUint64 uint64

// UnmarshalJSON implements the json.Unmarshaler interface.
func (f *FlexUint64) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	// Try unmarshaling as a number first
	var n uint64
	if err := json.Unmarshal(data, &n); err == nil {
		*f = FlexUint64(n)
		return nil
	}

	// Try unmarshaling as a string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		val, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return fmt.Errorf("FlexUint64: invalid uint64 string %q: %w", s, err)
		}
		*f = FlexUint64(val)
		return nil
	}

	return fmt.Errorf("FlexUint64: unexpected type, expected number or string")
}

// MarshalJSON implements the json.Marshaler interface.
func (f FlexUint64) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint64(f))
}

// Uint64 converts FlexUint64 back to uint64.
func (f FlexUint64) Uint64() uint64 {
	return uint64(f)
}

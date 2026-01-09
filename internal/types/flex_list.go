package types

import (
	"encoding/json"
)

// FlexList is a slice that can be unmarshaled from either a single JSON object or a JSON array.
type FlexList[T any] []T

// UnmarshalJSON implements the json.Unmarshaler interface.
func (f *FlexList[T]) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	// If it starts with '[', treat it as a normal array
	if data[0] == '[' {
		var slice []T
		if err := json.Unmarshal(data, &slice); err != nil {
			return err
		}
		*f = FlexList[T](slice)
		return nil
	}

	// Otherwise, try to unmarshal as a single item and wrap it in a slice
	var item T
	if err := json.Unmarshal(data, &item); err != nil {
		return err
	}
	*f = FlexList[T]{item}
	return nil
}

// Slice converts FlexList[T] back to []T.
func (f FlexList[T]) Slice() []T {
	return []T(f)
}

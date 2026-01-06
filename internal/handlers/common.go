package handlers

import "reflect"

// hasContent checks if the result map contains any non-empty properties
// ignoring metadata like "__version"
func hasContent(result map[string]interface{}) bool {
	if result == nil {
		return false
	}

	for key, value := range result {
		// Ignore metadata
		if key == "__version" {
			continue
		}

		// Check if value is non-empty
		if value != nil {
			// If it's a map (nested collection/doc), check recursively
			if vMap, ok := value.(map[string]interface{}); ok {
				if hasContent(vMap) {
					return true
				}
			} else {
				// It's a property value (leaf node)
				// Check for empty nil values or empty structures if needed
				// But typically a property value exists if it's in the map
				if !isEmptyValue(reflect.ValueOf(value)) {
					return true
				}
			}
		}
	}

	return false
}

// isEmptyValue checks if a value is empty (zero value)
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

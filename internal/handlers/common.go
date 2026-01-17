// common.go
//
// A scalable, high performance drop-in replacement for the jam-build nodejs data service
// Copyright (c) 2026 Alex Grant <info@localnerve.com> (https://www.localnerve.com), LocalNerve LLC
//
// This file is part of jam-build-propsdb.
// jam-build-propsdb is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License as published by the Free Software
// Foundation, either version 3 of the License, or (at your option) any later version.
// jam-build-propsdb is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
// without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU Affero General Public License for more details.
// You should have received a copy of the GNU Affero General Public License along with jam-build-propsdb.
// If not, see <https://www.gnu.org/licenses/>.
// Additional terms under GNU AGPL version 3 section 7:
// a) The reasonable legal notice of original copyright and author attribution must be preserved
//    by including the string: "Copyright (c) 2026 Alex Grant <info@localnerve.com> (https://www.localnerve.com), LocalNerve LLC"
//    in this material, copies, or source code of derived works.

package handlers

import (
	"reflect"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// parseCollections extracts collections from query parameters,
// supporting both multiple 'collections' keys and comma-separated values.
func parseCollections(c *fiber.Ctx) []string {
	collectionMap := make(map[string]struct{})

	// Visit all query arguments to collect multiple 'collections' parameters
	args := c.Context().QueryArgs()
	for key, value := range args.All() {
		if string(key) == "collections" {
			// Split by comma in case the value itself is comma-separated
			vals := strings.Split(string(value), ",")
			for _, v := range vals {
				v = strings.TrimSpace(v)
				if v != "" {
					collectionMap[v] = struct{}{}
				}
			}
		}
	}

	if len(collectionMap) == 0 {
		return nil
	}

	collections := make([]string, 0, len(collectionMap))
	for k := range collectionMap {
		collections = append(collections, k)
	}

	return collections
}

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

// flex_list.go
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

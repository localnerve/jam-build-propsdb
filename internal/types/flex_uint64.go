// flex_uint64.go
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

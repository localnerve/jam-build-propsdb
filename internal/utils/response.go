// response.go
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

package utils

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

// SuccessResponse sends a standard success response
func SuccessResponse(c *fiber.Ctx, data interface{}, status int) error {
	return c.Status(status).JSON(data)
}

// ErrorResponse sends a standard error response matching Node.js format
func ErrorResponse(c *fiber.Ctx, message string, status int, errorType string) error {
	return c.Status(status).JSON(fiber.Map{
		"status":    status,
		"message":   message,
		"ok":        false,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"url":       c.OriginalURL(),
		"type":      errorType,
	})
}

// VersionErrorResponse sends a version conflict error (409)
func VersionErrorResponse(c *fiber.Ctx) error {
	return c.Status(fiber.StatusConflict).JSON(fiber.Map{
		"status":       fiber.StatusConflict,
		"message":      "E_VERSION - Refresh and reconcile with current version and retry.",
		"ok":           false,
		"versionError": true,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"url":          c.OriginalURL(),
		"type":         "version",
	})
}

// NotFoundResponse sends a 404 not found response
func NotFoundResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"status":    fiber.StatusNotFound,
		"message":   message,
		"ok":        false,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"url":       c.OriginalURL(),
	})
}

// MutationSuccessResponse sends a success response for mutations (POST/DELETE)
func MutationSuccessResponse(c *fiber.Ctx, newVersion uint64, affectedRows int64) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "Success",
		"ok":           true,
		"newVersion":   fmt.Sprintf("%d", newVersion),
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"affectedRows": affectedRows,
	})
}

// ErrorResponseStruct defines the schema for error responses
type ErrorResponseStruct struct {
	Status       int    `json:"status"`
	Message      string `json:"message"`
	Ok           bool   `json:"ok"`
	Timestamp    string `json:"timestamp"`
	URL          string `json:"url"`
	Type         string `json:"type,omitempty"`
	VersionError bool   `json:"versionError,omitempty"`
}

// SuccessResponseStruct defines the schema for mutation success responses
type SuccessResponseStruct struct {
	Message      string `json:"message"`
	Ok           bool   `json:"ok"`
	NewVersion   string `json:"newVersion"`
	Timestamp    string `json:"timestamp"`
	AffectedRows int64  `json:"affectedRows"`
}

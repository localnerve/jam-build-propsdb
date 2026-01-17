// auth.go
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

package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/localnerve/jam-build-propsdb/internal/config"
	"github.com/localnerve/jam-build-propsdb/internal/services"
	"github.com/localnerve/jam-build-propsdb/internal/types"
)

// AuthAdmin validates that the request has admin role authorization
func AuthAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return authorize(c, []string{"admin"}, "data.authorization.admin")
	}
}

// AuthUser validates that the request has user role authorization
func AuthUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return authorize(c, []string{"user"}, "data.authorization.user")
	}
}

// authorize performs the authorization check
func authorize(c *fiber.Ctx, roles []string, errorType string) error {
	// Lazy initialization of Authorizer client
	if !services.IsAuthorizerInitialized() {
		cfg, err := config.Load()
		if err != nil {
			return &types.CustomError{
				Code:    fiber.StatusInternalServerError,
				Message: fmt.Sprintf("Failed to load config for authorizer: %v", err),
				Type:    errorType,
			}
		}
		if err := services.InitAuthorizer(cfg, c.Protocol(), c.Hostname()); err != nil {
			return &types.CustomError{
				Code:    fiber.StatusInternalServerError,
				Message: fmt.Sprintf("Failed to initialize authorizer: %v", err),
				Type:    errorType,
			}
		}
	}

	// Get session cookie
	session := c.Cookies("cookie_session")
	if session == "" {
		return &types.CustomError{
			Code:    fiber.StatusForbidden,
			Message: "Authorizer cookie \"cookie_session\" not found",
			Type:    errorType,
		}
	}

	// Validate session
	data, err := services.ValidateSession(session, roles)
	if err != nil {
		return &types.CustomError{
			Code:    fiber.StatusForbidden,
			Message: fmt.Sprintf("Invalid session: %v", err),
			Type:    errorType,
		}
	}

	// Set user data in context
	if user, ok := data["user"]; ok {
		c.Locals("user", user)
	}

	return c.Next()
}

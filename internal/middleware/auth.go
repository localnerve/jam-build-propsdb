package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/localnerve/propsdb/internal/services"
	"github.com/localnerve/propsdb/internal/types"
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

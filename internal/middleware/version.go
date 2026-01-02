package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// VersionMiddleware parses the X-Api-Version header and stores it in context
func VersionMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		version := c.Get("X-Api-Version", "1.0.0")

		// Support version aliases
		if version == "1.0" {
			version = "1.0.0"
		}

		// Store version in context
		c.Locals("apiVersion", version)

		return c.Next()
	}
}

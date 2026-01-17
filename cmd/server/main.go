// main.go
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

package main

import (
	"log"
	"os"
	"os/signal"
	"runtime/coverage"
	"syscall"
	"time"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	swagger "github.com/gofiber/swagger"
	"github.com/localnerve/jam-build-propsdb/internal/config"
	"github.com/localnerve/jam-build-propsdb/internal/database"
	"github.com/localnerve/jam-build-propsdb/internal/handlers"
	"github.com/localnerve/jam-build-propsdb/internal/middleware"
	"github.com/localnerve/jam-build-propsdb/internal/types"

	_ "github.com/localnerve/jam-build-propsdb/docs/api" // Swagger docs
)

// @title Jam-Build-PropsDB API
// @version 1.0.0
// @description Go Fiber data service with multi-database support
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/localnerve/jam-build-propsdb
// @contact.email info@localnerve.com

// @license.name AGPL-3.0
// @license.url https://www.gnu.org/licenses/agpl-3.0.html

// @host localhost:3000
// @BasePath /api
// @schemes http https

// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name cookie_session

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database (app pool)
	appDB, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to app database at startup: %v", err)
	}
	defer database.Close(appDB)

	// Connect to database (user pool)
	userDB, err := database.ConnectUser(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to user database at startup: %v", err)
	}
	defer database.Close(userDB)

	// Run auto-migrations
	if err := database.AutoMigrate(appDB); err != nil {
		log.Fatalf("Failed to run migrations at startup: %v", err)
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
		// Disable startup message for cleaner logs
		DisableStartupMessage: false,
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(compress.New())

	// Prometheus metrics
	prometheus := fiberprometheus.New("propsdb")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)

	// API routes under /api
	api := app.Group("/api")

	// Version middleware
	api.Use(middleware.VersionMiddleware())

	// Data routes
	data := api.Group("/data")

	// Create handlers
	appHandler := &handlers.AppDataHandler{DB: appDB}
	userHandler := &handlers.UserDataHandler{DB: userDB}

	// Application data routes (public GET, admin POST/DELETE)
	appRoutes := data.Group("/app")
	// Middleware for app mutations
	appRoutes.Use(func(c *fiber.Ctx) error {
		if c.Method() == fiber.MethodGet {
			return c.Next()
		}
		return middleware.AuthAdmin()(c)
	})
	appRoutes.Get("/:document/:collection", appHandler.GetAppProperties)
	appRoutes.Get("/:document", appHandler.GetAppCollectionsAndProperties)
	appRoutes.Get("/", appHandler.GetAppDocumentsCollectionsAndProperties)
	appRoutes.Post("/:document", appHandler.SetAppProperties)
	appRoutes.Delete("/:document/:collection", appHandler.DeleteAppCollection)
	appRoutes.Delete("/:document", appHandler.DeleteAppProperties)

	// User data routes (all require user authentication)
	userRoutes := data.Group("/user", middleware.AuthUser())
	userRoutes.Get("/:document/:collection", userHandler.GetUserProperties)
	userRoutes.Get("/:document", userHandler.GetUserCollectionsAndProperties)
	userRoutes.Get("/", userHandler.GetUserDocumentsCollectionsAndProperties)
	userRoutes.Post("/:document", userHandler.SetUserProperties)
	userRoutes.Delete("/:document/:collection", userHandler.DeleteUserCollection)
	userRoutes.Delete("/:document", userHandler.DeleteUserProperties)

	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":    fiber.StatusNotFound,
			"message":   "[404] Resource Not Found",
			"ok":        false,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"url":       c.OriginalURL(),
		})
	})

	// Initialize Authorizer (will be done on first auth request)
	// This is a placeholder - actual initialization happens in middleware
	log.Printf("Authorizer will be initialized on first authenticated request")

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Gracefully shutting down...")

		// Explicitly write coverage data before exiting
		// This ensures data is flushed even when we catch the signal
		if coverDir := os.Getenv("GOCOVERDIR"); coverDir != "" {
			log.Printf("Flushing coverage data to %s...", coverDir)
			// Ensure the directory exists (it should, but just in case)
			if err := os.MkdirAll(coverDir, 0755); err != nil {
				log.Printf("Warning: failed to create coverage directory: %v", err)
			}
			if err := coverage.WriteCountersDir(coverDir); err != nil {
				log.Printf("Warning: failed to write coverage counters: %v", err)
			}
			log.Println("Coverage flush complete. Waiting for orchestrator to extract...")
			time.Sleep(5 * time.Second) // Give the host time to extract files
		}

		_ = app.Shutdown()
	}()

	// Start server
	port := cfg.Port
	log.Printf("Starting server on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Println("Server stopped")
}

// customErrorHandler handles errors globally
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := err.Error()
	errorType := "unknown"

	// Check if it's a Fiber error
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	// Check for custom errors
	if e, ok := err.(*types.CustomError); ok {
		code = e.Code
		message = e.Message
		errorType = e.Type
	}

	// Check for version errors
	versionError := false
	if code == fiber.StatusConflict || (message != "" && len(message) >= 9 && message[:9] == "E_VERSION") {
		versionError = true
		errorType = "version"
		code = fiber.StatusConflict
	}

	return c.Status(code).JSON(fiber.Map{
		"status":       code,
		"message":      message,
		"ok":           false,
		"versionError": versionError,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"url":          c.OriginalURL(),
		"type":         errorType,
	})
}

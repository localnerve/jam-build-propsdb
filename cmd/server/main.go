package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	swagger "github.com/gofiber/swagger"
	"github.com/localnerve/propsdb/internal/config"
	"github.com/localnerve/propsdb/internal/database"
	"github.com/localnerve/propsdb/internal/handlers"
	"github.com/localnerve/propsdb/internal/middleware"

	_ "github.com/localnerve/propsdb/docs/api" // Swagger docs
)

// @title PropsDB API
// @version 1.0.0
// @description Go Fiber data service with multi-database support
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/localnerve/propsdb
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
		log.Fatalf("Failed to connect to app database: %v", err)
	}
	defer database.Close(appDB)

	// Connect to database (user pool)
	userDB, err := database.ConnectUser(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to user database: %v", err)
	}
	defer database.Close(userDB)

	// Run auto-migrations
	if err := database.AutoMigrate(appDB); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
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
	data.Get("/app/:document/:collection", appHandler.GetAppProperties)
	data.Get("/app/:document", appHandler.GetAppCollectionsAndProperties)
	data.Get("/app", appHandler.GetAppDocumentsCollectionsAndProperties)

	// Admin-only application routes
	data.Post("/app/:document", middleware.AuthAdmin(), appHandler.SetAppProperties)
	data.Delete("/app/:document/:collection", middleware.AuthAdmin(), appHandler.DeleteAppCollection)
	data.Delete("/app/:document", middleware.AuthAdmin(), appHandler.DeleteAppProperties)

	// User data routes (all require user authentication)
	data.Get("/user/:document/:collection", middleware.AuthUser(), userHandler.GetUserProperties)
	data.Get("/user/:document", middleware.AuthUser(), userHandler.GetUserCollectionsAndProperties)
	data.Get("/user", middleware.AuthUser(), userHandler.GetUserDocumentsCollectionsAndProperties)
	data.Post("/user/:document", middleware.AuthUser(), userHandler.SetUserProperties)
	data.Delete("/user/:document/:collection", middleware.AuthUser(), userHandler.DeleteUserCollection)
	data.Delete("/user/:document", middleware.AuthUser(), userHandler.DeleteUserProperties)

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

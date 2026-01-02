package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/localnerve/propsdb/internal/config"
	"github.com/localnerve/propsdb/internal/database"
	"github.com/localnerve/propsdb/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database (app pool)
	appDB, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close(appDB)

	// Perform health check
	result := services.HealthCheck(cfg, appDB)

	// Output result as JSON
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal health check result: %v", err)
	}

	fmt.Println(string(output))

	// Exit with appropriate code
	if result.Status != "healthy" {
		os.Exit(1)
	}
	os.Exit(0)
}

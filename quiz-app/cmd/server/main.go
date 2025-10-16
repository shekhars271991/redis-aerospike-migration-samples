package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"redis-migration/internal/config"
	"redis-migration/internal/router"
	"redis-migration/internal/service"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database service based on configuration
	var dbService service.DBService

	if cfg.Database.Type != "redis" {
		log.Fatalf("Unsupported database type: %s (only redis is supported)", cfg.Database.Type)
	}
	
	log.Println("Initializing Redis service...")
	dbService = service.NewRedisService(
		cfg.Redis.Addr,
		cfg.Redis.Password,
		cfg.Redis.DB,
	)

	// Test database connection
	log.Println("Testing database connection...")
	if err := dbService.HealthCheck(); err != nil {
		log.Fatalf("Database health check failed: %v", err)
	}
	log.Println("Database connection successful!")

	// Setup router
	r := router.SetupRouter(dbService)

	// Create server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s:%s using %s database", 
			cfg.Server.Host, cfg.Server.Port, cfg.Database.Type)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Print API endpoints
	printAPIEndpoints(cfg.Server.Port)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	log.Println("Server shutdown complete")
}

func printAPIEndpoints(port string) {
	baseURL := fmt.Sprintf("http://localhost:%s", port)
	
	fmt.Println("\n=== API Endpoints ===")
	fmt.Printf("Health Check: GET %s/health\n", baseURL)
	fmt.Printf("\nUser Management:\n")
	fmt.Printf("  Create User: POST %s/api/v1/users\n", baseURL)
	fmt.Printf("  Get User:    GET %s/api/v1/users/{id}\n", baseURL)
	fmt.Printf("\nGame Management:\n")
	fmt.Printf("  Start Game:  POST %s/api/v1/game/start\n", baseURL)
	fmt.Printf("  Finish Game: POST %s/api/v1/game/finish\n", baseURL)
	fmt.Printf("  Get Session: GET %s/api/v1/game/session/{id}\n", baseURL)
	fmt.Printf("\nLeaderboard:\n")
	fmt.Printf("  Get Leaderboard: GET %s/api/v1/leaderboard?top=10\n", baseURL)
	fmt.Printf("\nLegacy Endpoints (backward compatibility):\n")
	fmt.Printf("  POST %s/users\n", baseURL)
	fmt.Printf("  POST %s/game/start\n", baseURL)
	fmt.Printf("  POST %s/game/finish\n", baseURL)
	fmt.Printf("  GET %s/leaderboard\n", baseURL)
	fmt.Println("\n=== Example Usage ===")
	fmt.Printf("curl -X POST %s/api/v1/users -H \"Content-Type: application/json\" -d '{\"user_id\":\"user1\",\"name\":\"John Doe\",\"email\":\"john@example.com\"}'\n", baseURL)
	fmt.Printf("curl -X POST %s/api/v1/game/start -H \"Content-Type: application/json\" -d '{\"user_id\":\"user1\"}'\n", baseURL)
	fmt.Printf("curl -X GET %s/api/v1/leaderboard?top=5\n", baseURL)
	fmt.Println("========================\n")
}

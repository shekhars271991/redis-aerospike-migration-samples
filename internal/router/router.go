package router

import (
	"github.com/gin-gonic/gin"
	"redis-migration/internal/handlers"
	"redis-migration/internal/service"
)

// SetupRouter sets up the HTTP router with all routes
func SetupRouter(dbService service.DBService) *gin.Engine {
	// Create Gin router
	r := gin.Default()

	// Create handler instance
	handler := handlers.NewHandler(dbService)

	// Add middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	// Health check endpoint
	r.GET("/health", handler.HealthCheck)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// User routes
		users := v1.Group("/users")
		{
			users.POST("", handler.CreateUser)           // POST /api/v1/users
			users.GET("", handler.GetUsers)              // GET /api/v1/users
			users.GET("/:id", handler.GetUser)           // GET /api/v1/users/:id
		}

		// Game routes
		game := v1.Group("/game")
		{
			game.POST("/start", handler.StartGame)       // POST /api/v1/game/start
			game.POST("/finish", handler.FinishGame)     // POST /api/v1/game/finish
			game.GET("/session/:id", handler.GetGameSession) // GET /api/v1/game/session/:id
		}

		// Leaderboard route
		v1.GET("/leaderboard", handler.GetLeaderboard)   // GET /api/v1/leaderboard
	}

	// Legacy routes for backward compatibility (as specified in requirements)
	r.POST("/users", handler.CreateUser)
	r.GET("/users/:id", handler.GetUser)
	r.POST("/game/start", handler.StartGame)
	r.POST("/game/finish", handler.FinishGame)
	r.GET("/leaderboard", handler.GetLeaderboard)

	return r
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

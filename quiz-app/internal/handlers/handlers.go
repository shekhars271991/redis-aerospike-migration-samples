package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"redis-migration/internal/service"
)

// Handler contains the HTTP handlers and dependencies
type Handler struct {
	dbService service.DBService
}

// NewHandler creates a new handler instance
func NewHandler(dbService service.DBService) *Handler {
	return &Handler{
		dbService: dbService,
	}
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
}

// StartGameRequest represents the request body for starting a game
type StartGameRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// FinishGameRequest represents the request body for finishing a game
type FinishGameRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	Score     int    `json:"score" binding:"min=0"`
}

// CreateUser creates a new user
func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userData := map[string]interface{}{
		"name":  req.Name,
		"email": req.Email,
	}

	err := h.dbService.CreateUser(req.UserID, userData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user_id": req.UserID,
	})
}

// GetUser retrieves a user by ID
func (h *Handler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	userData, err := h.dbService.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": userData,
	})
}

// GetUsers lists all users (for demonstration - in production you'd want pagination)
func (h *Handler) GetUsers(c *gin.Context) {
	// This is a simplified implementation
	// In a real application, you'd want to implement proper user listing with pagination
	c.JSON(http.StatusOK, gin.H{
		"message": "Use GET /users/{id} to get a specific user",
	})
}

// StartGame starts a new game session
func (h *Handler) StartGame(c *gin.Context) {
	var req StartGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessionID, err := h.dbService.StartGameSession(req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"user_id":    req.UserID,
		"message":    "Game session started",
	})
}

// FinishGame finishes a game session and updates the score
func (h *Handler) FinishGame(c *gin.Context) {
	var req FinishGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.dbService.EndGameSession(req.SessionID, req.Score)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Game finished successfully",
		"session_id": req.SessionID,
		"score":      req.Score,
	})
}

// GetGameSession retrieves a game session
func (h *Handler) GetGameSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	session, err := h.dbService.GetGameSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session": session,
	})
}

// GetLeaderboard retrieves the leaderboard
func (h *Handler) GetLeaderboard(c *gin.Context) {
	topNStr := c.DefaultQuery("top", "10")
	topN, err := strconv.Atoi(topNStr)
	if err != nil || topN <= 0 {
		topN = 10
	}

	// Limit to reasonable maximum
	if topN > 100 {
		topN = 100
	}

	leaderboard, err := h.dbService.GetLeaderboard(topN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": leaderboard,
		"count":       len(leaderboard),
	})
}

// HealthCheck checks the health of the service
func (h *Handler) HealthCheck(c *gin.Context) {
	err := h.dbService.HealthCheck()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

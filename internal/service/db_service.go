package service

import "redis-migration/internal/models"

// DBService defines the interface for database operations
type DBService interface {
	// CreateUser creates a new user with the provided data
	CreateUser(userID string, data map[string]interface{}) error
	
	// GetUser retrieves user data by userID
	GetUser(userID string) (map[string]interface{}, error)
	
	// UpdateScore updates a user's score and game statistics
	UpdateScore(userID string, score int) error
	
	// GetLeaderboard retrieves the top N users by score
	GetLeaderboard(topN int) ([]models.UserScore, error)
	
	// Additional methods for game session management
	StartGameSession(userID string) (string, error)
	EndGameSession(sessionID string, score int) error
	GetGameSession(sessionID string) (*models.GameSession, error)
	
	// Health check method
	HealthCheck() error
}

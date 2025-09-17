package models

// UserScore represents a user's score in the leaderboard
type UserScore struct {
	UserID string  `json:"user_id"`
	Name   string  `json:"name"`
	Score  int     `json:"score"`
}

// User represents a user in the system
type User struct {
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	GamesPlayed int    `json:"games_played"`
	LastScore   int    `json:"last_score"`
}

// GameSession represents an active game session
type GameSession struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
	StartTime int64  `json:"start_time"`
	Active    bool   `json:"active"`
}

// GameResult represents the result of a completed game
type GameResult struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
	Score     int    `json:"score"`
	EndTime   int64  `json:"end_time"`
}

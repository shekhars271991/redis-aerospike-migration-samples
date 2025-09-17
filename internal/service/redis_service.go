package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"redis-migration/internal/models"
)

const (
	// Redis key prefixes
	userPrefix    = "user:"
	sessionPrefix = "session:"
	leaderboard   = "leaderboard"
)

// RedisService implements DBService using Redis
type RedisService struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisService creates a new Redis service instance
func NewRedisService(addr, password string, db int) *RedisService {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisService{
		client: rdb,
		ctx:    context.Background(),
	}
}

// CreateUser creates a new user with the provided data
func (r *RedisService) CreateUser(userID string, data map[string]interface{}) error {
	userKey := userPrefix + userID
	
	// Check if user already exists
	exists, err := r.client.Exists(r.ctx, userKey).Result()
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists > 0 {
		return fmt.Errorf("user %s already exists", userID)
	}

	// Set default values
	userData := map[string]interface{}{
		"user_id":      userID,
		"name":         "",
		"email":        "",
		"games_played": 0,
		"last_score":   0,
	}

	// Override with provided data
	for k, v := range data {
		userData[k] = v
	}

	// Store user data in hash
	err = r.client.HMSet(r.ctx, userKey, userData).Err()
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Initialize user in leaderboard with score 0
	err = r.client.ZAdd(r.ctx, leaderboard, redis.Z{
		Score:  0,
		Member: userID,
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to add user to leaderboard: %w", err)
	}

	return nil
}

// GetUser retrieves user data by userID
func (r *RedisService) GetUser(userID string) (map[string]interface{}, error) {
	userKey := userPrefix + userID
	
	result, err := r.client.HGetAll(r.ctx, userKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	if len(result) == 0 {
		return nil, fmt.Errorf("user %s not found", userID)
	}

	// Convert string values to appropriate types
	userData := make(map[string]interface{})
	for k, v := range result {
		switch k {
		case "games_played", "last_score":
			if intVal, err := strconv.Atoi(v); err == nil {
				userData[k] = intVal
			} else {
				userData[k] = 0
			}
		default:
			userData[k] = v
		}
	}

	return userData, nil
}

// UpdateScore updates a user's score and game statistics
func (r *RedisService) UpdateScore(userID string, score int) error {
	userKey := userPrefix + userID
	
	// Check if user exists
	exists, err := r.client.Exists(r.ctx, userKey).Result()
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("user %s not found", userID)
	}

	// Use pipeline for atomic operations
	pipe := r.client.Pipeline()
	
	// Update user stats
	pipe.HIncrBy(r.ctx, userKey, "games_played", 1)
	pipe.HSet(r.ctx, userKey, "last_score", score)
	
	// Update leaderboard - Redis sorted sets automatically handle duplicates
	pipe.ZAdd(r.ctx, leaderboard, redis.Z{
		Score:  float64(score),
		Member: userID,
	})
	
	_, err = pipe.Exec(r.ctx)
	if err != nil {
		return fmt.Errorf("failed to update score: %w", err)
	}

	return nil
}

// GetLeaderboard retrieves the top N users by score
func (r *RedisService) GetLeaderboard(topN int) ([]models.UserScore, error) {
	// Get top N users from sorted set (highest scores first)
	results, err := r.client.ZRevRangeWithScores(r.ctx, leaderboard, 0, int64(topN-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	leaderboardData := make([]models.UserScore, 0, len(results))
	
	for _, result := range results {
		userID := result.Member.(string)
		score := int(result.Score)
		
		// Get user name
		userName, err := r.client.HGet(r.ctx, userPrefix+userID, "name").Result()
		if err != nil {
			// If we can't get the name, use userID as fallback
			userName = userID
		}
		
		leaderboardData = append(leaderboardData, models.UserScore{
			UserID: userID,
			Name:   userName,
			Score:  score,
		})
	}

	return leaderboardData, nil
}

// StartGameSession creates a new game session
func (r *RedisService) StartGameSession(userID string) (string, error) {
	// Check if user exists
	userKey := userPrefix + userID
	exists, err := r.client.Exists(r.ctx, userKey).Result()
	if err != nil {
		return "", fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists == 0 {
		return "", fmt.Errorf("user %s not found", userID)
	}

	sessionID := generateSessionID()
	sessionKey := sessionPrefix + sessionID
	
	session := models.GameSession{
		SessionID: sessionID,
		UserID:    userID,
		StartTime: time.Now().Unix(),
		Active:    true,
	}
	
	sessionData, err := json.Marshal(session)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session: %w", err)
	}
	
	// Store session with 1 hour expiration
	err = r.client.Set(r.ctx, sessionKey, sessionData, time.Hour).Err()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return sessionID, nil
}

// EndGameSession ends a game session and updates the score
func (r *RedisService) EndGameSession(sessionID string, score int) error {
	sessionKey := sessionPrefix + sessionID
	
	// Get session data
	sessionData, err := r.client.Get(r.ctx, sessionKey).Result()
	if err == redis.Nil {
		return fmt.Errorf("session %s not found or expired", sessionID)
	} else if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	
	var session models.GameSession
	err = json.Unmarshal([]byte(sessionData), &session)
	if err != nil {
		return fmt.Errorf("failed to unmarshal session: %w", err)
	}
	
	if !session.Active {
		return fmt.Errorf("session %s is not active", sessionID)
	}
	
	// Update user score
	err = r.UpdateScore(session.UserID, score)
	if err != nil {
		return fmt.Errorf("failed to update user score: %w", err)
	}
	
	// Mark session as inactive and delete it
	err = r.client.Del(r.ctx, sessionKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// GetGameSession retrieves a game session by sessionID
func (r *RedisService) GetGameSession(sessionID string) (*models.GameSession, error) {
	sessionKey := sessionPrefix + sessionID
	
	sessionData, err := r.client.Get(r.ctx, sessionKey).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("session %s not found or expired", sessionID)
	} else if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	var session models.GameSession
	err = json.Unmarshal([]byte(sessionData), &session)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// HealthCheck checks if Redis connection is healthy
func (r *RedisService) HealthCheck() error {
	return r.client.Ping(r.ctx).Err()
}

// generateSessionID generates a simple session ID without external dependencies
func generateSessionID() string {
	return fmt.Sprintf("session_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

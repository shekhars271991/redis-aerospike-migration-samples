package service

import (
	"fmt"
	"time"

	"github.com/aerospike/aerospike-client-go/v8"
	"redis-migration/internal/models"
)

const (
	// Aerospike namespaces and sets
	namespace      = "quiz"
	usersSet       = "users"
	sessionsSet    = "sessions"
	leaderboardKey = "leaderboard:top"
	maxLeaderboard = 1000 // Maximum entries to keep in materialized leaderboard
)

// AerospikeService implements DBService using Aerospike
type AerospikeService struct {
	client *aerospike.Client
}

// NewAerospikeService creates a new Aerospike service instance
func NewAerospikeService(hosts []*aerospike.Host) (*AerospikeService, error) {
	client, err := aerospike.NewClientWithPolicyAndHost(nil, hosts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Aerospike: %w", err)
	}

	service := &AerospikeService{
		client: client,
	}

	// Initialize empty leaderboard if it doesn't exist
	if err := service.initializeLeaderboard(); err != nil {
		return nil, fmt.Errorf("failed to initialize leaderboard: %w", err)
	}

	return service, nil
}

// initializeLeaderboard creates an empty leaderboard record if it doesn't exist
func (a *AerospikeService) initializeLeaderboard() error {
	key, err := aerospike.NewKey(namespace, usersSet, leaderboardKey)
	if err != nil {
		return err
	}

	// Check if leaderboard exists
	exists, err := a.client.Exists(nil, key)
	if err != nil {
		return err
	}

	if !exists {
		// Create empty leaderboard
		bins := aerospike.BinMap{
			"top_users": []interface{}{},
			"updated":   time.Now().Unix(),
		}
		
		err = a.client.Put(nil, key, bins)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateUser creates a new user with the provided data
func (a *AerospikeService) CreateUser(userID string, data map[string]interface{}) error {
	key, err := aerospike.NewKey(namespace, usersSet, userID)
	if err != nil {
		return fmt.Errorf("failed to create key: %w", err)
	}

	// Check if user already exists
	exists, err := a.client.Exists(nil, key)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return fmt.Errorf("user %s already exists", userID)
	}

	// Set default values
	bins := aerospike.BinMap{
		"user_id":      userID,
		"name":         "",
		"email":        "",
		"games_played": 0,
		"last_score":   0,
		"created_at":   time.Now().Unix(),
	}

	// Override with provided data
	for k, v := range data {
		bins[k] = v
	}

	err = a.client.Put(nil, key, bins)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUser retrieves user data by userID
func (a *AerospikeService) GetUser(userID string) (map[string]interface{}, error) {
	key, err := aerospike.NewKey(namespace, usersSet, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create key: %w", err)
	}

	record, err := a.client.Get(nil, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if record == nil {
		return nil, fmt.Errorf("user %s not found", userID)
	}

	userData := make(map[string]interface{})
	for k, v := range record.Bins {
		userData[k] = v
	}

	return userData, nil
}

// UpdateScore updates a user's score and game statistics
func (a *AerospikeService) UpdateScore(userID string, score int) error {
	key, err := aerospike.NewKey(namespace, usersSet, userID)
	if err != nil {
		return fmt.Errorf("failed to create key: %w", err)
	}

	// Check if user exists
	exists, err := a.client.Exists(nil, key)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("user %s not found", userID)
	}

	// Update user record
	bins := aerospike.BinMap{
		"last_score": score,
		"updated_at": time.Now().Unix(),
	}

	// Increment games_played
	policy := aerospike.NewWritePolicy(0, 0)
	err = a.client.Put(policy, key, bins)
	if err != nil {
		return fmt.Errorf("failed to update user score: %w", err)
	}

	// Increment games_played separately
	incBins := aerospike.BinMap{
		"games_played": 1,
	}
	err = a.client.Add(policy, key, incBins)
	if err != nil {
		return fmt.Errorf("failed to increment games_played: %w", err)
	}

	// Update leaderboard using ordered list
	if err := a.updateLeaderboardOrderedList(userID, score); err != nil {
		return fmt.Errorf("failed to update leaderboard: %w", err)
	}

	return nil
}

// updateLeaderboardOrderedList updates the leaderboard using Aerospike ordered lists
func (a *AerospikeService) updateLeaderboardOrderedList(userID string, score int) error {
	leaderboardKey, err := aerospike.NewKey(namespace, "leaderboard", "global")
	if err != nil {
		return fmt.Errorf("failed to create leaderboard key: %w", err)
	}

	// Get user name for the leaderboard entry
	userKey, err := aerospike.NewKey(namespace, usersSet, userID)
	if err != nil {
		return fmt.Errorf("failed to create user key: %w", err)
	}
	
	userRecord, err := a.client.Get(nil, userKey)
	if err != nil {
		return fmt.Errorf("failed to get user record: %w", err)
	}
	
	userName := userID // fallback
	if userRecord != nil && userRecord.Bins["name"] != nil {
		if name, ok := userRecord.Bins["name"].(string); ok && name != "" {
			userName = name
		}
	}

	// Create composite leaderboard entry: [negative_score, user_id, name]
	// We use negative score for descending order (highest scores first)
	leaderboardEntry := []interface{}{-score, userID, userName}

	// Use list operations to maintain ordered leaderboard
	policy := aerospike.NewWritePolicy(0, 0)
	listPolicy := aerospike.NewListPolicy(aerospike.ListOrderOrdered, aerospike.ListWriteFlagsDefault)
	
	// Simply insert the new entry - Aerospike will maintain order automatically
	_, err = a.client.Operate(policy, leaderboardKey,
		aerospike.ListAppendWithPolicyOp(listPolicy, "scores", leaderboardEntry),
	)
	if err != nil {
		return fmt.Errorf("failed to append to leaderboard: %w", err)
	}

	return nil
}

// GetLeaderboard retrieves the top N users by score from ordered list
func (a *AerospikeService) GetLeaderboard(topN int) ([]models.UserScore, error) {
	key, err := aerospike.NewKey(namespace, "leaderboard", "global")
	if err != nil {
		return nil, fmt.Errorf("failed to create leaderboard key: %w", err)
	}

	// Get the leaderboard record
	record, err := a.client.Get(nil, key)
	if err != nil {
		// If the leaderboard doesn't exist yet, return empty slice
		return []models.UserScore{}, nil
	}

	if record == nil || record.Bins["scores"] == nil {
		return []models.UserScore{}, nil
	}

	var leaderboardData []models.UserScore
	if scoresData, ok := record.Bins["scores"].([]interface{}); ok {
		// Limit to requested number of entries
		limit := topN
		if len(scoresData) < limit {
			limit = len(scoresData)
		}

		// Process entries (already sorted by Aerospike with highest scores first)
		for i := 0; i < limit; i++ {
			if entrySlice, ok := scoresData[i].([]interface{}); ok && len(entrySlice) >= 3 {
				// Entry format: [negative_score, user_id, name]
				negativeScore, _ := entrySlice[0].(int)
				userID, _ := entrySlice[1].(string)
				name, _ := entrySlice[2].(string)

				// Convert back to positive score
				score := -negativeScore

				leaderboardData = append(leaderboardData, models.UserScore{
					UserID: userID,
					Name:   name,
					Score:  score,
				})
			}
		}
	}

	return leaderboardData, nil
}

// StartGameSession creates a new game session
func (a *AerospikeService) StartGameSession(userID string) (string, error) {
	// Check if user exists
	userKey, err := aerospike.NewKey(namespace, usersSet, userID)
	if err != nil {
		return "", fmt.Errorf("failed to create user key: %w", err)
	}

	exists, err := a.client.Exists(nil, userKey)
	if err != nil {
		return "", fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return "", fmt.Errorf("user %s not found", userID)
	}

	sessionID := generateSessionID()
	sessionKey, err := aerospike.NewKey(namespace, usersSet, "session:"+sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to create session key: %w", err)
	}

	bins := aerospike.BinMap{
		"session_id": sessionID,
		"user_id":    userID,
		"start_time": time.Now().Unix(),
		"active":     true,
	}

	// Create session without TTL for now (Aerospike may not allow TTL on this storage config)
	// We'll handle expiration at the application level if needed
	policy := aerospike.NewWritePolicy(0, 0) // No TTL for now
	err = a.client.Put(policy, sessionKey, bins)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return sessionID, nil
}

// EndGameSession ends a game session and updates the score
func (a *AerospikeService) EndGameSession(sessionID string, score int) error {
	sessionKey, err := aerospike.NewKey(namespace, usersSet, "session:"+sessionID)
	if err != nil {
		return fmt.Errorf("failed to create session key: %w", err)
	}

	// Get session data
	record, err := a.client.Get(nil, sessionKey)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	if record == nil {
		return fmt.Errorf("session %s not found or expired", sessionID)
	}

	userID, ok := record.Bins["user_id"].(string)
	if !ok {
		return fmt.Errorf("invalid session data")
	}

	// Check if session has expired (1 hour = 3600 seconds)
	if startTime, ok := record.Bins["start_time"].(int); ok {
		if time.Now().Unix()-int64(startTime) > 3600 {
			return fmt.Errorf("session %s has expired", sessionID)
		}
	}

	active, ok := record.Bins["active"].(bool)
	if !ok || !active {
		return fmt.Errorf("session %s is not active", sessionID)
	}

	// Update user score
	if err := a.UpdateScore(userID, score); err != nil {
		return fmt.Errorf("failed to update user score: %w", err)
	}

	// Delete session
	_, err = a.client.Delete(nil, sessionKey)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// GetGameSession retrieves a game session by sessionID
func (a *AerospikeService) GetGameSession(sessionID string) (*models.GameSession, error) {
	sessionKey, err := aerospike.NewKey(namespace, usersSet, "session:"+sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to create session key: %w", err)
	}

	record, err := a.client.Get(nil, sessionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	if record == nil {
		return nil, fmt.Errorf("session %s not found or expired", sessionID)
	}

	session := &models.GameSession{
		SessionID: sessionID,
	}

	if userID, ok := record.Bins["user_id"].(string); ok {
		session.UserID = userID
	}
	if startTime, ok := record.Bins["start_time"].(int); ok {
		session.StartTime = int64(startTime)
		
		// Check if session has expired (1 hour = 3600 seconds)
		if time.Now().Unix()-int64(startTime) > 3600 {
			// Session has expired, mark as inactive and return error
			return nil, fmt.Errorf("session %s has expired", sessionID)
		}
	}
	if active, ok := record.Bins["active"].(bool); ok {
		session.Active = active
	}

	return session, nil
}

// HealthCheck checks if Aerospike connection is healthy
func (a *AerospikeService) HealthCheck() error {
	// Try to get cluster info
	nodes := a.client.GetNodes()
	if len(nodes) == 0 {
		return fmt.Errorf("no Aerospike nodes available")
	}
	return nil
}

// Close closes the Aerospike client connection
func (a *AerospikeService) Close() {
	if a.client != nil {
		a.client.Close()
	}
}


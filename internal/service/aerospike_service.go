package service

import (
	"fmt"
	"sort"
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
	client, err := aerospike.NewClient(hosts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Aerospike: %w", err)
	}

	service := &AerospikeService{
		client: client,
	}

	// Initialize empty leaderboard if it doesn't exist
	err = service.initializeLeaderboard()
	if err != nil {
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
	incBin := aerospike.NewBin("games_played", 1)
	err = a.client.Add(policy, key, incBin)
	if err != nil {
		return fmt.Errorf("failed to increment games_played: %w", err)
	}

	// Update materialized leaderboard
	err = a.updateMaterializedLeaderboard(userID, score)
	if err != nil {
		return fmt.Errorf("failed to update leaderboard: %w", err)
	}

	return nil
}

// updateMaterializedLeaderboard updates the materialized leaderboard with a new score
func (a *AerospikeService) updateMaterializedLeaderboard(userID string, score int) error {
	leaderboardKey, err := aerospike.NewKey(namespace, usersSet, leaderboardKey)
	if err != nil {
		return err
	}

	// Get current leaderboard
	record, err := a.client.Get(nil, leaderboardKey)
	if err != nil {
		return err
	}

	var topUsers []map[string]interface{}
	if record != nil && record.Bins["top_users"] != nil {
		if usersData, ok := record.Bins["top_users"].([]interface{}); ok {
			for _, userData := range usersData {
				if userMap, ok := userData.(map[string]interface{}); ok {
					topUsers = append(topUsers, userMap)
				}
			}
		}
	}

	// Get user name for the leaderboard
	userKey, err := aerospike.NewKey(namespace, usersSet, userID)
	if err != nil {
		return err
	}
	
	userRecord, err := a.client.Get(nil, userKey)
	if err != nil {
		return err
	}
	
	userName := userID // fallback
	if userRecord != nil && userRecord.Bins["name"] != nil {
		if name, ok := userRecord.Bins["name"].(string); ok && name != "" {
			userName = name
		}
	}

	// Find existing user in leaderboard or add new entry
	found := false
	for i, user := range topUsers {
		if userIDVal, ok := user["user_id"].(string); ok && userIDVal == userID {
			// Update existing entry
			topUsers[i]["score"] = score
			topUsers[i]["name"] = userName
			found = true
			break
		}
	}

	if !found {
		// Add new entry
		topUsers = append(topUsers, map[string]interface{}{
			"user_id": userID,
			"name":    userName,
			"score":   score,
		})
	}

	// Sort by score (descending)
	sort.Slice(topUsers, func(i, j int) bool {
		scoreI, okI := topUsers[i]["score"].(int)
		scoreJ, okJ := topUsers[j]["score"].(int)
		if !okI || !okJ {
			return false
		}
		return scoreI > scoreJ
	})

	// Keep only top maxLeaderboard entries
	if len(topUsers) > maxLeaderboard {
		topUsers = topUsers[:maxLeaderboard]
	}

	// Convert to []interface{} for Aerospike
	topUsersInterface := make([]interface{}, len(topUsers))
	for i, user := range topUsers {
		topUsersInterface[i] = user
	}

	// Update leaderboard record
	bins := aerospike.BinMap{
		"top_users": topUsersInterface,
		"updated":   time.Now().Unix(),
	}

	err = a.client.Put(nil, leaderboardKey, bins)
	if err != nil {
		return err
	}

	return nil
}

// GetLeaderboard retrieves the top N users by score
func (a *AerospikeService) GetLeaderboard(topN int) ([]models.UserScore, error) {
	key, err := aerospike.NewKey(namespace, usersSet, leaderboardKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create key: %w", err)
	}

	record, err := a.client.Get(nil, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}
	if record == nil {
		return []models.UserScore{}, nil
	}

	var leaderboardData []models.UserScore
	if topUsersData, ok := record.Bins["top_users"].([]interface{}); ok {
		limit := topN
		if len(topUsersData) < limit {
			limit = len(topUsersData)
		}

		for i := 0; i < limit; i++ {
			if userMap, ok := topUsersData[i].(map[string]interface{}); ok {
				userID, _ := userMap["user_id"].(string)
				name, _ := userMap["name"].(string)
				score, _ := userMap["score"].(int)

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
	sessionKey, err := aerospike.NewKey(namespace, sessionsSet, sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to create session key: %w", err)
	}

	bins := aerospike.BinMap{
		"session_id": sessionID,
		"user_id":    userID,
		"start_time": time.Now().Unix(),
		"active":     true,
	}

	// Set TTL for session (1 hour)
	policy := aerospike.NewWritePolicy(0, 3600) // 3600 seconds = 1 hour
	err = a.client.Put(policy, sessionKey, bins)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return sessionID, nil
}

// EndGameSession ends a game session and updates the score
func (a *AerospikeService) EndGameSession(sessionID string, score int) error {
	sessionKey, err := aerospike.NewKey(namespace, sessionsSet, sessionID)
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

	active, ok := record.Bins["active"].(bool)
	if !ok || !active {
		return fmt.Errorf("session %s is not active", sessionID)
	}

	// Update user score
	err = a.UpdateScore(userID, score)
	if err != nil {
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
	sessionKey, err := aerospike.NewKey(namespace, sessionsSet, sessionID)
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
	}
	if active, ok := record.Bins["active"].(bool); ok {
		session.Active = active
	}

	return session, nil
}

// HealthCheck checks if Aerospike connection is healthy
func (a *AerospikeService) HealthCheck() error {
	// Try to get cluster info
	_, err := a.client.GetNodes()
	return err
}

// Close closes the Aerospike client connection
func (a *AerospikeService) Close() {
	if a.client != nil {
		a.client.Close()
	}
}

// generateSessionID generates a simple session ID without external dependencies
func generateSessionID() string {
	return fmt.Sprintf("session_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

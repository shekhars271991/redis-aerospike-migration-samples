package generator

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"
)

// DataGenerator generates test data for API requests
type DataGenerator struct {
	mu             sync.RWMutex
	userIDs        []string
	sessionIDs     []string
	names          []string
	emails         []string
	rand           *rand.Rand
	userIDCounter  int64
	sessionCounter int64
}

// NewDataGenerator creates a new data generator
func NewDataGenerator() *DataGenerator {
	return &DataGenerator{
		userIDs:    make([]string, 0, 10000),
		sessionIDs: make([]string, 0, 1000),
		names: []string{
			"Alice Johnson", "Bob Smith", "Charlie Brown", "Diana Prince", "Edward Norton",
			"Fiona Apple", "George Lucas", "Hannah Montana", "Ivan Drago", "Julia Roberts",
			"Kevin Hart", "Laura Croft", "Michael Jordan", "Nancy Drew", "Oliver Twist",
			"Penelope Cruz", "Quentin Tarantino", "Rachel Green", "Steve Jobs", "Tina Turner",
			"Uma Thurman", "Vincent Vega", "Will Smith", "Xena Warrior", "Yoda Master", "Zoe Saldana",
		},
		emails: []string{
			"user@example.com", "test@loadtest.com", "demo@quiz.com", "player@game.com",
			"gamer@test.org", "user@benchmark.net", "load@test.io", "perf@example.org",
		},
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateUserID generates a unique user ID
func (g *DataGenerator) GenerateUserID() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	g.userIDCounter++
	userID := fmt.Sprintf("user_%d_%d", g.userIDCounter, time.Now().UnixNano()%1000000)
	g.userIDs = append(g.userIDs, userID)
	
	// Keep only the last 10000 user IDs to prevent memory growth
	if len(g.userIDs) > 10000 {
		g.userIDs = g.userIDs[len(g.userIDs)-10000:]
	}
	
	return userID
}

// GetRandomUserID returns a random existing user ID
func (g *DataGenerator) GetRandomUserID() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	if len(g.userIDs) == 0 {
		return g.GenerateUserID()
	}
	
	return g.userIDs[g.rand.Intn(len(g.userIDs))]
}

// GenerateSessionID generates a session ID
func (g *DataGenerator) GenerateSessionID() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	g.sessionCounter++
	sessionID := fmt.Sprintf("session_%d_%d", g.sessionCounter, time.Now().UnixNano()%1000000)
	g.sessionIDs = append(g.sessionIDs, sessionID)
	
	// Keep only the last 1000 session IDs
	if len(g.sessionIDs) > 1000 {
		g.sessionIDs = g.sessionIDs[len(g.sessionIDs)-1000:]
	}
	
	return sessionID
}

// GetRandomSessionID returns a random existing session ID
func (g *DataGenerator) GetRandomSessionID() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	if len(g.sessionIDs) == 0 {
		return g.GenerateSessionID()
	}
	
	return g.sessionIDs[g.rand.Intn(len(g.sessionIDs))]
}

// GenerateName returns a random name
func (g *DataGenerator) GenerateName() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	return g.names[g.rand.Intn(len(g.names))]
}

// GenerateEmail returns a random email
func (g *DataGenerator) GenerateEmail() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	baseEmail := g.emails[g.rand.Intn(len(g.emails))]
	parts := strings.Split(baseEmail, "@")
	if len(parts) != 2 {
		return baseEmail
	}
	
	return fmt.Sprintf("%s_%d@%s", parts[0], g.rand.Intn(10000), parts[1])
}

// GenerateScore returns a random game score
func (g *DataGenerator) GenerateScore() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	// Generate scores with realistic distribution
	scores := []int{
		g.rand.Intn(500),      // 0-499 (low scores)
		500 + g.rand.Intn(500), // 500-999 (medium scores)
		1000 + g.rand.Intn(1000), // 1000-1999 (good scores)
		2000 + g.rand.Intn(1000), // 2000-2999 (high scores)
		3000 + g.rand.Intn(2000), // 3000-4999 (excellent scores)
	}
	
	// Weight distribution: 30% low, 30% medium, 25% good, 10% high, 5% excellent
	weights := []float64{0.30, 0.30, 0.25, 0.10, 0.05}
	r := g.rand.Float64()
	
	cumulative := 0.0
	for i, weight := range weights {
		cumulative += weight
		if r <= cumulative {
			return scores[i]
		}
	}
	
	return scores[0]
}

// GenerateTopN returns a random top N value for leaderboard queries
func (g *DataGenerator) GenerateTopN() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	options := []int{5, 10, 20, 50, 100}
	return options[g.rand.Intn(len(options))]
}

// TemplateData represents data available for template substitution
type TemplateData struct {
	UserID    string
	SessionID string
	Name      string
	Email     string
	Score     int
	TopN      int
}

// GenerateTemplateData generates a complete set of template data
func (g *DataGenerator) GenerateTemplateData() *TemplateData {
	return &TemplateData{
		UserID:    g.GetRandomUserID(),
		SessionID: g.GetRandomSessionID(),
		Name:      g.GenerateName(),
		Email:     g.GenerateEmail(),
		Score:     g.GenerateScore(),
		TopN:      g.GenerateTopN(),
	}
}

// ProcessTemplate processes a template string with generated data
func (g *DataGenerator) ProcessTemplate(templateStr string, data *TemplateData) (string, error) {
	if data == nil {
		data = g.GenerateTemplateData()
	}
	
	// Create template
	tmpl, err := template.New("request").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}
	
	// Execute template
	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}
	
	return buf.String(), nil
}

// ProcessPath processes a URL path template
func (g *DataGenerator) ProcessPath(pathTemplate string, data *TemplateData) string {
	if data == nil {
		data = g.GenerateTemplateData()
	}
	
	result := pathTemplate
	result = strings.ReplaceAll(result, "{{.UserID}}", data.UserID)
	result = strings.ReplaceAll(result, "{{.SessionID}}", data.SessionID)
	result = strings.ReplaceAll(result, "{{.TopN}}", strconv.Itoa(data.TopN))
	
	return result
}

// GetStats returns generator statistics
func (g *DataGenerator) GetStats() map[string]interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	return map[string]interface{}{
		"total_users":    len(g.userIDs),
		"total_sessions": len(g.sessionIDs),
		"user_counter":   g.userIDCounter,
		"session_counter": g.sessionCounter,
	}
}

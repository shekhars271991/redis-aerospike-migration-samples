package generator

import (
	"fmt"
	"math/rand"
	"time"

	"loadtest-client/internal/client"
	"loadtest-client/internal/config"
)

// RequestGenerator generates HTTP requests based on configuration
type RequestGenerator struct {
	config        *config.Config
	dataGenerator *DataGenerator
	apiWeights    []float64
	cumulativeWeights []float64
	rand          *rand.Rand
}

// NewRequestGenerator creates a new request generator
func NewRequestGenerator(cfg *config.Config) *RequestGenerator {
	generator := &RequestGenerator{
		config:        cfg,
		dataGenerator: NewDataGenerator(),
		apiWeights:    make([]float64, len(cfg.APIs)),
		cumulativeWeights: make([]float64, len(cfg.APIs)),
		rand:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	
	// Calculate cumulative weights for weighted selection
	cumulative := 0.0
	for i, api := range cfg.APIs {
		generator.apiWeights[i] = api.Weight
		cumulative += api.Weight
		generator.cumulativeWeights[i] = cumulative
	}
	
	return generator
}

// GenerateRequest generates a random request based on API weights
func (g *RequestGenerator) GenerateRequest() (*client.Request, error) {
	// Select API based on weights
	apiIndex := g.selectWeightedAPI()
	if apiIndex < 0 || apiIndex >= len(g.config.APIs) {
		return nil, fmt.Errorf("invalid API index: %d", apiIndex)
	}
	
	apiConfig := g.config.APIs[apiIndex]
	templateData := g.dataGenerator.GenerateTemplateData()
	
	// Process path template
	path := g.dataGenerator.ProcessPath(apiConfig.Path, templateData)
	
	// Process body template if present
	var body []byte
	if apiConfig.Body != "" {
		bodyStr, err := g.dataGenerator.ProcessTemplate(apiConfig.Body, templateData)
		if err != nil {
			return nil, fmt.Errorf("body template processing failed: %w", err)
		}
		body = []byte(bodyStr)
	}
	
	// Copy headers
	headers := make(map[string]string)
	for k, v := range apiConfig.Headers {
		headers[k] = v
	}
	
	return &client.Request{
		Method:  apiConfig.Method,
		Path:    path,
		Body:    body,
		Headers: headers,
	}, nil
}

// GenerateSpecificRequest generates a request for a specific API
func (g *RequestGenerator) GenerateSpecificRequest(apiName string) (*client.Request, error) {
	var apiConfig *config.APIConfig
	for _, api := range g.config.APIs {
		if api.Name == apiName {
			apiConfig = &api
			break
		}
	}
	
	if apiConfig == nil {
		return nil, fmt.Errorf("API not found: %s", apiName)
	}
	
	templateData := g.dataGenerator.GenerateTemplateData()
	
	// Process path template
	path := g.dataGenerator.ProcessPath(apiConfig.Path, templateData)
	
	// Process body template if present
	var body []byte
	if apiConfig.Body != "" {
		bodyStr, err := g.dataGenerator.ProcessTemplate(apiConfig.Body, templateData)
		if err != nil {
			return nil, fmt.Errorf("body template processing failed: %w", err)
		}
		body = []byte(bodyStr)
	}
	
	// Copy headers
	headers := make(map[string]string)
	for k, v := range apiConfig.Headers {
		headers[k] = v
	}
	
	return &client.Request{
		Method:  apiConfig.Method,
		Path:    path,
		Body:    body,
		Headers: headers,
	}, nil
}

// GenerateBatch generates a batch of requests
func (g *RequestGenerator) GenerateBatch(count int) ([]*client.Request, error) {
	requests := make([]*client.Request, count)
	
	for i := 0; i < count; i++ {
		req, err := g.GenerateRequest()
		if err != nil {
			return nil, fmt.Errorf("failed to generate request %d: %w", i, err)
		}
		requests[i] = req
	}
	
	return requests, nil
}

// selectWeightedAPI selects an API based on configured weights
func (g *RequestGenerator) selectWeightedAPI() int {
	r := g.rand.Float64()
	
	for i, weight := range g.cumulativeWeights {
		if r <= weight {
			return i
		}
	}
	
	// Fallback to first API
	return 0
}

// GetAPIDistribution returns the current API distribution
func (g *RequestGenerator) GetAPIDistribution() map[string]float64 {
	distribution := make(map[string]float64)
	for i, api := range g.config.APIs {
		distribution[api.Name] = g.apiWeights[i]
	}
	return distribution
}

// GetDataGenerator returns the data generator for external use
func (g *RequestGenerator) GetDataGenerator() *DataGenerator {
	return g.dataGenerator
}

// CreateUserRequest creates a specific user creation request
func (g *RequestGenerator) CreateUserRequest() (*client.Request, error) {
	return g.GenerateSpecificRequest("create_user")
}

// GetUserRequest creates a specific get user request
func (g *RequestGenerator) GetUserRequest() (*client.Request, error) {
	return g.GenerateSpecificRequest("get_user")
}

// StartGameRequest creates a specific start game request
func (g *RequestGenerator) StartGameRequest() (*client.Request, error) {
	return g.GenerateSpecificRequest("start_game")
}

// FinishGameRequest creates a specific finish game request
func (g *RequestGenerator) FinishGameRequest() (*client.Request, error) {
	return g.GenerateSpecificRequest("finish_game")
}

// GetLeaderboardRequest creates a specific leaderboard request
func (g *RequestGenerator) GetLeaderboardRequest() (*client.Request, error) {
	return g.GenerateSpecificRequest("get_leaderboard")
}

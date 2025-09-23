package config

import (
	"fmt"
	"time"
)

// Config represents the load test configuration
type Config struct {
	Target     TargetConfig     `yaml:"target" mapstructure:"target"`
	Load       LoadConfig       `yaml:"load" mapstructure:"load"`
	APIs       []APIConfig      `yaml:"apis" mapstructure:"apis"`
	Reporting  ReportingConfig  `yaml:"reporting" mapstructure:"reporting"`
	HTTP       HTTPConfig       `yaml:"http" mapstructure:"http"`
}

// TargetConfig defines the target service configuration
type TargetConfig struct {
	BaseURL string `yaml:"base_url" mapstructure:"base_url"`
	Timeout string `yaml:"timeout" mapstructure:"timeout"`
}

// LoadConfig defines the load generation parameters
type LoadConfig struct {
	Duration        string `yaml:"duration" mapstructure:"duration"`
	Workers         int    `yaml:"workers" mapstructure:"workers"`
	RampUpDuration  string `yaml:"ramp_up_duration" mapstructure:"ramp_up_duration"`
	TargetQPS       int    `yaml:"target_qps" mapstructure:"target_qps"`
	MaxConcurrency  int    `yaml:"max_concurrency" mapstructure:"max_concurrency"`
}

// APIConfig defines configuration for each API endpoint
type APIConfig struct {
	Name        string  `yaml:"name" mapstructure:"name"`
	Method      string  `yaml:"method" mapstructure:"method"`
	Path        string  `yaml:"path" mapstructure:"path"`
	Weight      float64 `yaml:"weight" mapstructure:"weight"`
	Body        string  `yaml:"body" mapstructure:"body"`
	Headers     map[string]string `yaml:"headers" mapstructure:"headers"`
	Variables   map[string]interface{} `yaml:"variables" mapstructure:"variables"`
}

// ReportingConfig defines reporting and metrics configuration
type ReportingConfig struct {
	Interval      string `yaml:"interval" mapstructure:"interval"`
	OutputFile    string `yaml:"output_file" mapstructure:"output_file"`
	DetailedStats bool   `yaml:"detailed_stats" mapstructure:"detailed_stats"`
}

// HTTPConfig defines HTTP client configuration
type HTTPConfig struct {
	MaxIdleConns        int    `yaml:"max_idle_conns" mapstructure:"max_idle_conns"`
	MaxConnsPerHost     int    `yaml:"max_conns_per_host" mapstructure:"max_conns_per_host"`
	IdleConnTimeout     string `yaml:"idle_conn_timeout" mapstructure:"idle_conn_timeout"`
	DisableCompression  bool   `yaml:"disable_compression" mapstructure:"disable_compression"`
	DisableKeepAlives   bool   `yaml:"disable_keep_alives" mapstructure:"disable_keep_alives"`
}

// GetDuration parses duration string and returns time.Duration
func (c *Config) GetDuration() (time.Duration, error) {
	return time.ParseDuration(c.Load.Duration)
}

// GetRampUpDuration parses ramp up duration string and returns time.Duration
func (c *Config) GetRampUpDuration() (time.Duration, error) {
	return time.ParseDuration(c.Load.RampUpDuration)
}

// GetTargetTimeout parses timeout string and returns time.Duration
func (c *Config) GetTargetTimeout() (time.Duration, error) {
	return time.ParseDuration(c.Target.Timeout)
}

// GetReportingInterval parses reporting interval string and returns time.Duration
func (c *Config) GetReportingInterval() (time.Duration, error) {
	return time.ParseDuration(c.Reporting.Interval)
}

// GetIdleConnTimeout parses idle connection timeout string and returns time.Duration
func (c *Config) GetIdleConnTimeout() (time.Duration, error) {
	return time.ParseDuration(c.HTTP.IdleConnTimeout)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Target.BaseURL == "" {
		return fmt.Errorf("target base_url is required")
	}
	
	if c.Load.Duration == "" {
		return fmt.Errorf("load duration is required")
	}
	
	if c.Load.Workers <= 0 {
		return fmt.Errorf("load workers must be greater than 0")
	}
	
	if c.Load.TargetQPS <= 0 {
		return fmt.Errorf("load target_qps must be greater than 0")
	}
	
	if len(c.APIs) == 0 {
		return fmt.Errorf("at least one API configuration is required")
	}
	
	// Validate API weights sum to 1.0
	totalWeight := 0.0
	for _, api := range c.APIs {
		if api.Weight <= 0 {
			return fmt.Errorf("API %s weight must be greater than 0", api.Name)
		}
		totalWeight += api.Weight
	}
	
	if totalWeight < 0.99 || totalWeight > 1.01 {
		return fmt.Errorf("API weights must sum to 1.0, got %.2f", totalWeight)
	}
	
	return nil
}

// GetDefaultConfig returns a default configuration
func GetDefaultConfig() *Config {
	return &Config{
		Target: TargetConfig{
			BaseURL: "http://localhost:8080",
			Timeout: "30s",
		},
		Load: LoadConfig{
			Duration:        "60s",
			Workers:         100,
			RampUpDuration:  "10s",
			TargetQPS:       1000,
			MaxConcurrency:  1000,
		},
		APIs: []APIConfig{
			{
				Name:   "create_user",
				Method: "POST",
				Path:   "/api/v1/users",
				Weight: 0.1,
				Body:   `{"user_id":"{{.UserID}}","name":"{{.Name}}","email":"{{.Email}}"}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			{
				Name:   "get_user",
				Method: "GET",
				Path:   "/api/v1/users/{{.UserID}}",
				Weight: 0.2,
			},
			{
				Name:   "start_game",
				Method: "POST",
				Path:   "/api/v1/game/start",
				Weight: 0.3,
				Body:   `{"user_id":"{{.UserID}}"}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			{
				Name:   "finish_game",
				Method: "POST",
				Path:   "/api/v1/game/finish",
				Weight: 0.3,
				Body:   `{"session_id":"{{.SessionID}}","score":{{.Score}}}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			{
				Name:   "get_leaderboard",
				Method: "GET",
				Path:   "/api/v1/leaderboard?top={{.TopN}}",
				Weight: 0.1,
			},
		},
		Reporting: ReportingConfig{
			Interval:      "10s",
			OutputFile:    "loadtest_results.json",
			DetailedStats: true,
		},
		HTTP: HTTPConfig{
			MaxIdleConns:       1000,
			MaxConnsPerHost:    1000,
			IdleConnTimeout:    "90s",
			DisableCompression: false,
			DisableKeepAlives:  false,
		},
	}
}

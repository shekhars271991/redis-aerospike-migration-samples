package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/aerospike/aerospike-client-go/v8"
)

// Config holds the application configuration
type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	Aerospike  AerospikeConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
	Host string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type string // "redis" or "aerospike"
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// AerospikeConfig holds Aerospike configuration
type AerospikeConfig struct {
	Hosts []string
	Port  int
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	config := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Type: getEnv("DB_TYPE", "redis"), // default to redis
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Aerospike: AerospikeConfig{
			Hosts: getEnvAsSlice("AEROSPIKE_HOSTS", []string{"localhost"}),
			Port:  getEnvAsInt("AEROSPIKE_PORT", 3000),
		},
	}

	return config
}

// GetAerospikeHosts returns Aerospike hosts in the format expected by the client
func (c *Config) GetAerospikeHosts() []*aerospike.Host {
	var hosts []*aerospike.Host
	for _, hostname := range c.Aerospike.Hosts {
		hosts = append(hosts, aerospike.NewHost(hostname, c.Aerospike.Port))
	}
	return hosts
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
// getEnvAsSlice gets an environment variable as slice or returns a default value
func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}


// Package config handles application configuration
package config

import (
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	Server ServerConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	GRPCPort string
	HTTPPort string
}

// AuthConfig holds authentication settings
type AuthConfig struct {
	JWTSecret string
}

// OpenSearchConfig holds OpenSearch configuration
type OpenSearchConfig struct {
	Host               string
	Port               int
	Username           string
	Password           string
	InsecureSkipVerify bool
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			GRPCPort: getEnv("GRPC_PORT", "50051"),
			HTTPPort: getEnv("HTTP_PORT", "8080"),
		},
	}
}

// LoadAuthConfig loads authentication configuration
func LoadAuthConfig() *AuthConfig {
	return &AuthConfig{JWTSecret: getEnv("JWT_SECRET", "dev-secret-change-me")}
}

// LoadOpenSearchConfig loads OpenSearch configuration
func LoadOpenSearchConfig() *OpenSearchConfig {
	port := 9200
	if portStr := os.Getenv("OPENSEARCH_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	insecureSkipVerify := os.Getenv("OPENSEARCH_INSECURE_SKIP_VERIFY") == "true"

	return &OpenSearchConfig{
		Host:               getEnv("OPENSEARCH_HOST", "localhost"),
		Port:               port,
		Username:           getEnv("OPENSEARCH_USERNAME", "admin"),
		Password:           getEnv("OPENSEARCH_PASSWORD", "admin"),
		InsecureSkipVerify: insecureSkipVerify,
	}
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

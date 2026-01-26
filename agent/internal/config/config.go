// Package config handles agent configuration
package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds agent configuration
type Config struct {
	// Backend connection
	BackendAddr string
	BackendTLS  bool

	// Agent identity
	AgentVersion string
	Hostname     string
	IPAddress    string

	// Monitoring settings
	MetricsInterval time.Duration
	BatchSize       int

	// Storage
	TokenFile  string
	ConfigFile string
	LogFile    string
	CacheDir   string

	// Retry settings
	MaxRetries     int
	RetryInterval  time.Duration
	ReconnectDelay time.Duration

	// Metadata
	Metadata map[string]string
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown-host"
	}

	return &Config{
		BackendAddr:     getEnv("BACKEND_ADDR", "localhost:50051"),
		BackendTLS:      getEnvBool("BACKEND_TLS", false),
		AgentVersion:    "2.0.0",
		Hostname:        hostname,
		MetricsInterval: time.Duration(getEnvInt("METRICS_INTERVAL", 5)) * time.Second,
		BatchSize:       getEnvInt("BATCH_SIZE", 10),
		TokenFile:       getEnv("TOKEN_FILE", ".agent_token"),
		ConfigFile:      getEnv("CONFIG_FILE", "agent.yaml"),
		LogFile:         getEnv("LOG_FILE", "agent.log"),
		CacheDir:        getEnv("CACHE_DIR", ".cache"),
		MaxRetries:      getEnvInt("MAX_RETRIES", 3),
		RetryInterval:   time.Duration(getEnvInt("RETRY_INTERVAL", 5)) * time.Second,
		ReconnectDelay:  time.Duration(getEnvInt("RECONNECT_DELAY", 10)) * time.Second,
		Metadata: map[string]string{
			"environment": getEnv("ENVIRONMENT", "production"),
			"location":    getEnv("LOCATION", "default"),
			"datacenter":  getEnv("DATACENTER", "dc-01"),
		},
	}
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets environment variable as int with default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool gets environment variable as bool with default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}

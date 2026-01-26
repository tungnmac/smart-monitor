// Package identity handles agent identification and credentials
package identity

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

// Credentials holds agent authentication information
type Credentials struct {
	AgentID     string `json:"agent_id"`
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
	Hostname    string `json:"hostname"`
	IPAddress   string `json:"ip_address"`
}

// Manager handles agent identity
type Manager struct {
	tokenFile string
}

// NewManager creates a new identity manager
func NewManager(tokenFile string) *Manager {
	return &Manager{
		tokenFile: tokenFile,
	}
}

// GenerateAgentID creates a unique identifier for this agent
func (m *Manager) GenerateAgentID(hostname, ipAddress string) string {
	data := fmt.Sprintf("%s-%s-%d", hostname, ipAddress, time.Now().Unix())
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("agent-%x", hash[:8])
}

// GetLocalIP returns the local IP address
func (m *Manager) GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknown"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "unknown"
}

// SaveCredentials saves agent credentials to file
func (m *Manager) SaveCredentials(creds *Credentials) error {
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	if err := os.WriteFile(m.tokenFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials: %w", err)
	}

	return nil
}

// LoadCredentials loads agent credentials from file
func (m *Manager) LoadCredentials() (*Credentials, error) {
	data, err := os.ReadFile(m.tokenFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials: %w", err)
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	return &creds, nil
}

// IsTokenExpired checks if the token is expired
func (m *Manager) IsTokenExpired(creds *Credentials) bool {
	return time.Now().Unix() > creds.ExpiresAt
}

// HasValidCredentials checks if valid credentials exist
func (m *Manager) HasValidCredentials() bool {
	creds, err := m.LoadCredentials()
	if err != nil {
		return false
	}
	return !m.IsTokenExpired(creds)
}

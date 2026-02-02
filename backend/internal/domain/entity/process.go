// Package entity defines core business entities
package entity

// Process represents a running process on an agent
type Process struct {
	PID     int32
	Name    string
	CPU     float64
	Memory  float64
	AgentID string
	Host    string
}

// NewProcess creates a new process entity
func NewProcess(pid int32, name string, cpu, memory float64, agentID, host string) *Process {
	return &Process{
		PID:     pid,
		Name:    name,
		CPU:     cpu,
		Memory:  memory,
		AgentID: agentID,
		Host:    host,
	}
}

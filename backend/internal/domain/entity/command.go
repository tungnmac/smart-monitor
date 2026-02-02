package entity

// CommandType represents the type of command to send to an agent
type CommandType string

const (
	CommandKillProcess  CommandType = "KILL_PROCESS"
	CommandRestartAgent CommandType = "RESTART_AGENT"
	CommandUpdateConfig CommandType = "UPDATE_CONFIG"
	CommandRefreshStats CommandType = "REFRESH_STATS"
)

// AgentCommand represents a command to be sent to a monitoring agent
type AgentCommand struct {
	Type CommandType
	Args map[string]string
}

// NewKillProcessCommand creates a command to kill a process
func NewKillProcessCommand(pid string) *AgentCommand {
	return &AgentCommand{
		Type: CommandKillProcess,
		Args: map[string]string{
			"pid": pid,
		},
	}
}

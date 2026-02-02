// Package grpc implements gRPC handlers
package grpc

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"smart-monitor/backend/internal/application/dto"
	"smart-monitor/backend/internal/application/usecase"
	"smart-monitor/backend/internal/domain/entity"
	"smart-monitor/backend/internal/domain/service"
	pb "smart-monitor/pbtypes/monitor"
)

// MonitorServiceServer implements the gRPC MonitorService
type MonitorServiceServer struct {
	pb.UnimplementedMonitorServiceServer
	monitorUseCase *usecase.MonitorUseCase
	authService    *service.AuthService
	controlService *service.AgentControlService
	policyService  *service.PolicyService
}

// NewMonitorServiceServer creates a new gRPC server
func NewMonitorServiceServer(
	monitorUseCase *usecase.MonitorUseCase,
	authService *service.AuthService,
	controlService *service.AgentControlService,
	policyService *service.PolicyService,
) *MonitorServiceServer {
	return &MonitorServiceServer{
		monitorUseCase: monitorUseCase,
		authService:    authService,
		controlService: controlService,
		policyService:  policyService,
	}
}

// ListAgents returns all registered agents with their status and metrics
func (s *MonitorServiceServer) ListAgents(ctx context.Context, req *pb.ListAgentsRequest) (*pb.ListAgentsResponse, error) {
	log.Printf("ListAgents called")

	// Get all agents from the registry
	agents, err := s.monitorUseCase.ListAllAgents(ctx)
	if err != nil {
		log.Printf("Failed to list agents: %v", err)
		return &pb.ListAgentsResponse{
			Agents: []*pb.AgentInfo{},
			Total:  0,
		}, nil
	}

	// Convert to protobuf format
	var pbAgents []*pb.AgentInfo
	for _, agent := range agents {
		// Determine status based on last_auth_at
		status := "offline"
		if time.Since(agent.LastAuthAt) < 2*time.Minute {
			status = "online"
		} else if time.Since(agent.LastAuthAt) < 10*time.Minute {
			status = "degraded"
		}

		// Get latest metrics for this agent
		stats, _ := s.monitorUseCase.GetStats(ctx, agent.Hostname)
		cpu, ram, disk := 0.0, 0.0, 0.0
		if stats != nil {
			cpu = stats.CPU
			ram = stats.RAM
			disk = stats.Disk
		}

		pbAgents = append(pbAgents, &pb.AgentInfo{
			Id:           agent.AgentID,
			Host:         agent.Hostname,
			Env:          agent.Metadata["environment"],
			Region:       agent.Metadata["location"],
			Status:       status,
			Version:      agent.AgentVersion,
			Cpu:          cpu,
			Ram:          ram,
			Disk:         disk,
			IpAddress:    agent.IPAddress,
			LastSeen:     agent.LastAuthAt.Unix(),
			RegisteredAt: agent.RegisteredAt.Unix(),
		})
	}

	log.Printf("✓ Returning %d agents", len(pbAgents))

	return &pb.ListAgentsResponse{
		Agents: pbAgents,
		Total:  int32(len(pbAgents)),
	}, nil
}

// RegisterAgent handles agent registration
func (s *MonitorServiceServer) RegisterAgent(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	log.Printf("Registration request from hostname: %s, IP: %s", req.Hostname, req.IpAddress)

	// Validate request
	if req.Hostname == "" {
		return &pb.RegisterResponse{
			Success: false,
			Message: "Hostname is required",
		}, nil
	}

	// Register agent through auth service
	agent, err := s.authService.RegisterAgent(ctx, req.Hostname, req.IpAddress, req.AgentVersion, req.Metadata)
	if err != nil {
		log.Printf("Failed to register agent: %v", err)
		return &pb.RegisterResponse{
			Success: false,
			Message: fmt.Sprintf("Registration failed: %v", err),
		}, nil
	}

	log.Printf("✓ Agent registered successfully: %s (Token: %s...)", agent.AgentID, agent.AccessToken[:16])

	return &pb.RegisterResponse{
		Success:     true,
		Message:     "Agent registered successfully",
		AgentId:     agent.AgentID,
		AccessToken: agent.AccessToken,
		ExpiresAt:   agent.TokenExpiry.Unix(),
	}, nil
}

// StreamStats handles bidirectional streaming from agents
func (s *MonitorServiceServer) StreamStats(stream pb.MonitorService_StreamStatsServer) error {
	var agentID string
	commandCh := make(chan *entity.AgentCommand, 10)
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	// Goroutine to send commands back to agent
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case cmd := <-commandCh:
				pbCmd := &pb.Command{
					Type: pb.Command_NONE,
					Args: cmd.Args,
				}

				switch cmd.Type {
				case entity.CommandKillProcess:
					pbCmd.Type = pb.Command_KILL_PROCESS
				case entity.CommandRestartAgent:
					pbCmd.Type = pb.Command_RESTART_AGENT
				case entity.CommandUpdateConfig:
					pbCmd.Type = pb.Command_UPDATE_CONFIG
				}

				err := stream.Send(&pb.StatsResponse{
					Message:   "Executing command",
					Timestamp: time.Now().Unix(),
					Command:   pbCmd,
				})
				if err != nil {
					log.Printf("Error sending command to agent %s: %v", agentID, err)
					return
				}
			}
		}
	}()

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			log.Printf("Stream closed by client (agent %s)", agentID)
			s.controlService.UnregisterCommandChan(agentID)
			return nil
		}
		if err != nil {
			log.Printf("Error receiving stats from agent %s: %v", agentID, err)
			s.controlService.UnregisterCommandChan(agentID)
			return fmt.Errorf("failed to receive stats: %w", err)
		}

		// Set agentID if not already set
		if agentID == "" {
			agentID = req.AgentId
			s.controlService.RegisterCommandChan(agentID, commandCh)
		}

		// Authenticate agent
		if req.AccessToken == "" {
			log.Printf("Missing access token from agent %s", req.AgentId)
			continue
		}

		if err := s.authService.ValidateToken(context.Background(), req.AgentId, req.AccessToken); err != nil {
			log.Printf("Authentication failed for agent %s: %v", req.AgentId, err)
			return status.Errorf(codes.Unauthenticated, "authentication failed: %v", err)
		}

		// Convert protobuf to DTO
		var pbProcesses []*dto.ProcessDTO
		for _, p := range req.GetProcesses() {
			pbProcesses = append(pbProcesses, &dto.ProcessDTO{
				PID:     p.Pid,
				Name:    p.Name,
				CPU:     p.Cpu,
				Memory:  p.Memory,
				Command: p.Command,
				Port:    p.Port,
			})
		}

		statsReq := &dto.StatsRequest{
			Hostname:     req.Hostname,
			AgentID:      req.AgentId,
			IPAddress:    req.IpAddress,
			AgentVersion: req.AgentVersion,
			CPU:          req.Cpu,
			RAM:          req.Ram,
			Disk:         req.Disk,
			Metadata:     req.Metadata,
			Processes:    pbProcesses,
		}

		// Process through use case
		if err := s.monitorUseCase.RecordStats(context.Background(), statsReq); err != nil {
			log.Printf("Failed to record stats from agent %s: %v", req.AgentId, err)
			continue
		}

		// Log received stats with agent info
		log.Printf("[Agent:%s | Host:%s | IP:%s] CPU: %.2f%% | RAM: %.2f%% | Disk: %.2f%%",
			req.AgentId, req.Hostname, req.IpAddress, req.Cpu, req.Ram, req.Disk)

		// Send heartbeat response
		err = stream.Send(&pb.StatsResponse{
			Message:   "Stats received",
			Timestamp: time.Now().Unix(),
		})
		if err != nil {
			log.Printf("Error sending response to agent %s: %v", agentID, err)
			s.controlService.UnregisterCommandChan(agentID)
			return err
		}
	}
}

// GetStats returns stats for a specific hostname
func (s *MonitorServiceServer) GetStats(ctx context.Context, req *pb.StatsRequest) (*pb.StatsResponse, error) {
	hostname := req.Hostname
	if hostname == "" {
		return nil, fmt.Errorf("hostname is required")
	}

	log.Printf("GetStats called for hostname: %s", hostname)

	// Get stats through use case
	stats, err := s.monitorUseCase.GetStats(ctx, hostname)
	if err != nil {
		return &pb.StatsResponse{
			Message:   fmt.Sprintf("No stats available for hostname: %s", hostname),
			Timestamp: time.Now().Unix(),
		}, nil
	}

	message := fmt.Sprintf("Stats for %s: CPU=%.2f%%, RAM=%.2f%%, Disk=%.2f%% (Last received: %s)",
		hostname, stats.CPU, stats.RAM, stats.Disk, stats.LastReceived.Format(time.RFC3339))

	return &pb.StatsResponse{
		Message:   message,
		Timestamp: time.Now().Unix(),
	}, nil
}

// ControlAgent handles agent control operations
func (s *MonitorServiceServer) ControlAgent(ctx context.Context, req *pb.ControlAgentRequest) (*pb.ControlAgentResponse, error) {
	log.Printf("Control request for agent %s: action=%s, reason=%s", req.AgentId, req.Action, req.Reason)

	if req.AgentId == "" {
		return &pb.ControlAgentResponse{
			Success: false,
			Message: "Agent ID is required",
		}, nil
	}

	// Validate action
	var action entity.AgentControlAction
	switch req.Action {
	case "start":
		action = entity.AgentActionStart
	case "shutdown":
		action = entity.AgentActionShutdown
	case "restart":
		action = entity.AgentActionRestart
	default:
		return &pb.ControlAgentResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid action: %s (must be: start, shutdown, restart)", req.Action),
		}, nil
	}

	// Execute control action
	if err := s.controlService.ControlAgent(req.AgentId, action, req.Reason); err != nil {
		log.Printf("Control action failed: %v", err)
		return &pb.ControlAgentResponse{
			Success: false,
			Message: fmt.Sprintf("Control action failed: %v", err),
		}, nil
	}

	return &pb.ControlAgentResponse{
		Success:   true,
		Message:   fmt.Sprintf("Control action '%s' sent successfully", req.Action),
		AgentId:   req.AgentId,
		Action:    req.Action,
		Timestamp: time.Now().Unix(),
	}, nil
}

// BlockAgent handles agent blocking
func (s *MonitorServiceServer) BlockAgent(ctx context.Context, req *pb.BlockAgentRequest) (*pb.BlockAgentResponse, error) {
	log.Printf("Block request for agent %s: blocked=%v, reason=%s", req.AgentId, req.Blocked, req.Reason)

	if req.AgentId == "" {
		return &pb.BlockAgentResponse{
			Success: false,
			Message: "Agent ID is required",
		}, nil
	}

	var err error
	if req.Blocked {
		err = s.controlService.BlockAgent(req.AgentId, req.Reason)
	} else {
		err = s.controlService.UnblockAgent(req.AgentId)
	}

	if err != nil {
		return &pb.BlockAgentResponse{
			Success: false,
			Message: fmt.Sprintf("Block operation failed: %v", err),
		}, nil
	}

	action := "unblocked"
	if req.Blocked {
		action = "blocked"
	}

	return &pb.BlockAgentResponse{
		Success: true,
		Message: fmt.Sprintf("Agent %s successfully", action),
		AgentId: req.AgentId,
		Blocked: req.Blocked,
	}, nil
}

// AddPolicy handles policy creation
func (s *MonitorServiceServer) AddPolicy(ctx context.Context, req *pb.PolicyRequest) (*pb.PolicyResponse, error) {
	log.Printf("Add policy request: name=%s", req.Name)

	if req.Name == "" {
		return &pb.PolicyResponse{
			Success: false,
			Message: "Policy name is required",
		}, nil
	}

	policy, err := s.policyService.CreatePolicy(req.Name, req.Description, req.Thresholds, req.Actions, req.Metadata)
	if err != nil {
		return &pb.PolicyResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to create policy: %v", err),
		}, nil
	}

	return &pb.PolicyResponse{
		Success:   true,
		Message:   "Policy created successfully",
		PolicyId:  policy.PolicyID,
		Timestamp: time.Now().Unix(),
	}, nil
}

// UpdatePolicy handles policy updates
func (s *MonitorServiceServer) UpdatePolicy(ctx context.Context, req *pb.PolicyRequest) (*pb.PolicyResponse, error) {
	log.Printf("Update policy request: policy_id=%s", req.PolicyId)

	if req.PolicyId == "" {
		return &pb.PolicyResponse{
			Success: false,
			Message: "Policy ID is required",
		}, nil
	}

	policy, err := s.policyService.UpdatePolicy(req.PolicyId, req.Name, req.Description, req.Thresholds, req.Actions, req.Metadata)
	if err != nil {
		return &pb.PolicyResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to update policy: %v", err),
		}, nil
	}

	return &pb.PolicyResponse{
		Success:   true,
		Message:   "Policy updated successfully",
		PolicyId:  policy.PolicyID,
		Timestamp: time.Now().Unix(),
	}, nil
}

// RemovePolicy handles policy deletion
func (s *MonitorServiceServer) RemovePolicy(ctx context.Context, req *pb.RemovePolicyRequest) (*pb.PolicyResponse, error) {
	log.Printf("Remove policy request: policy_id=%s", req.PolicyId)

	if req.PolicyId == "" {
		return &pb.PolicyResponse{
			Success: false,
			Message: "Policy ID is required",
		}, nil
	}

	if err := s.policyService.RemovePolicy(req.PolicyId); err != nil {
		return &pb.PolicyResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to remove policy: %v", err),
		}, nil
	}

	return &pb.PolicyResponse{
		Success:   true,
		Message:   "Policy removed successfully",
		PolicyId:  req.PolicyId,
		Timestamp: time.Now().Unix(),
	}, nil
}

// ListPolicies handles policy listing
func (s *MonitorServiceServer) ListPolicies(ctx context.Context, req *pb.ListPoliciesRequest) (*pb.ListPoliciesResponse, error) {
	log.Printf("List policies request: page=%d, page_size=%d", req.Page, req.PageSize)

	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	policies, total, err := s.policyService.ListPolicies(page, pageSize)
	if err != nil {
		return &pb.ListPoliciesResponse{
			Policies: []*pb.Policy{},
			Total:    0,
		}, nil
	}

	var pbPolicies []*pb.Policy
	for _, p := range policies {
		pbPolicies = append(pbPolicies, &pb.Policy{
			PolicyId:      p.PolicyID,
			Name:          p.Name,
			Description:   p.Description,
			Thresholds:    p.Thresholds,
			Actions:       p.Actions,
			Metadata:      p.Metadata,
			Enabled:       p.Enabled,
			CreatedAt:     p.CreatedAt.Unix(),
			UpdatedAt:     p.UpdatedAt.Unix(),
			AppliedAgents: p.AppliedAgents,
		})
	}

	return &pb.ListPoliciesResponse{
		Policies: pbPolicies,
		Total:    int32(total),
	}, nil
}

// ApplyPolicy applies policy to agent
func (s *MonitorServiceServer) ApplyPolicy(ctx context.Context, req *pb.ApplyPolicyRequest) (*pb.PolicyResponse, error) {
	log.Printf("Apply policy request: agent_id=%s, policy_id=%s", req.AgentId, req.PolicyId)

	if req.AgentId == "" || req.PolicyId == "" {
		return &pb.PolicyResponse{
			Success: false,
			Message: "Agent ID and Policy ID are required",
		}, nil
	}

	if err := s.policyService.ApplyPolicyToAgent(req.PolicyId, req.AgentId); err != nil {
		return &pb.PolicyResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to apply policy: %v", err),
		}, nil
	}

	return &pb.PolicyResponse{
		Success:   true,
		Message:   "Policy applied successfully",
		PolicyId:  req.PolicyId,
		Timestamp: time.Now().Unix(),
	}, nil
}

// UnapplyPolicy removes policy from agent
func (s *MonitorServiceServer) UnapplyPolicy(ctx context.Context, req *pb.UnapplyPolicyRequest) (*pb.PolicyResponse, error) {
	log.Printf("Unapply policy request: agent_id=%s, policy_id=%s", req.AgentId, req.PolicyId)

	if req.AgentId == "" || req.PolicyId == "" {
		return &pb.PolicyResponse{
			Success: false,
			Message: "Agent ID and Policy ID are required",
		}, nil
	}

	if err := s.policyService.UnapplyPolicyFromAgent(req.PolicyId, req.AgentId); err != nil {
		return &pb.PolicyResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to unapply policy: %v", err),
		}, nil
	}

	return &pb.PolicyResponse{
		Success:   true,
		Message:   "Policy unapplied successfully",
		PolicyId:  req.PolicyId,
		Timestamp: time.Now().Unix(),
	}, nil
}

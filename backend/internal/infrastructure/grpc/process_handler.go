// Package grpc implements gRPC handlers for process operations
package grpc

import (
	"context"
	"fmt"
	"log"
	"time"

	"smart-monitor/backend/internal/domain/service"
	pb "smart-monitor/pbtypes/process"
)

// ProcessServiceServer implements the gRPC ProcessService
type ProcessServiceServer struct {
	pb.UnimplementedProcessServiceServer
	controlService *service.AgentControlService
}

// NewProcessServiceServer creates a new gRPC process server
func NewProcessServiceServer(controlService *service.AgentControlService) *ProcessServiceServer {
	return &ProcessServiceServer{
		controlService: controlService,
	}
}

// GetProcesses retrieves list of processes from an agent
func (s *ProcessServiceServer) GetProcesses(ctx context.Context, req *pb.GetProcessesRequest) (*pb.GetProcessesResponse, error) {
	log.Printf("GetProcesses request for hostname: %s", req.Hostname)

	if req.Hostname == "" {
		return nil, fmt.Errorf("hostname is required")
	}

	// TODO: Map hostname to agentID properly
	// For now, using hostname as agentID
	processes, err := s.controlService.GetProcessList(req.Hostname)
	if err != nil {
		log.Printf("Failed to get processes: %v", err)
		return &pb.GetProcessesResponse{
			Processes: []*pb.Process{},
			Timestamp: time.Now().Unix(),
		}, nil
	}

	// Convert to protobuf format
	var pbProcesses []*pb.Process
	for _, proc := range processes {
		pbProcesses = append(pbProcesses, &pb.Process{
			Pid:    proc.PID,
			Name:   proc.Name,
			Cpu:    proc.CPU,
			Memory: proc.Memory,
		})
	}

	log.Printf("✓ Returning %d processes for %s", len(pbProcesses), req.Hostname)

	return &pb.GetProcessesResponse{
		Processes: pbProcesses,
		Timestamp: time.Now().Unix(),
	}, nil
}

// KillProcess sends a kill signal to a process on an agent
func (s *ProcessServiceServer) KillProcess(ctx context.Context, req *pb.KillProcessRequest) (*pb.KillProcessResponse, error) {
	log.Printf("KillProcess request for hostname: %s, PID: %d", req.Hostname, req.Pid)

	if req.Hostname == "" {
		return nil, fmt.Errorf("hostname is required")
	}

	if req.Pid <= 0 {
		return nil, fmt.Errorf("invalid PID")
	}

	// TODO: Map hostname to agentID properly
	err := s.controlService.KillProcess(req.Hostname, req.Pid)
	if err != nil {
		log.Printf("Failed to kill process: %v", err)
		return &pb.KillProcessResponse{
			Message:   fmt.Sprintf("Failed to kill process: %v", err),
			Timestamp: time.Now().Unix(),
		}, nil
	}

	log.Printf("✓ Process %d killed successfully on %s", req.Pid, req.Hostname)

	return &pb.KillProcessResponse{
		Message:   fmt.Sprintf("Process %d killed successfully", req.Pid),
		Timestamp: time.Now().Unix(),
	}, nil
}

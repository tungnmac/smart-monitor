// Package grpc implements gRPC handlers for process operations
package grpc

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
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
	log.Printf("GetProcesses request for hostname: %s (sort: %s %s)", req.Hostname, req.SortBy, req.Order)

	if req.Hostname == "" {
		return nil, fmt.Errorf("hostname is required")
	}

	processes, err := s.controlService.GetProcessList(req.Hostname)
	if err != nil {
		log.Printf("Failed to get processes: %v", err)
		return &pb.GetProcessesResponse{
			Processes: []*pb.Process{},
			Timestamp: time.Now().Unix(),
		}, nil
	}

	// Apply sorting if requested
	if req.SortBy != "" {
		sort.Slice(processes, func(i, j int) bool {
			var less bool
			switch strings.ToLower(req.SortBy) {
			case "name":
				less = processes[i].Name < processes[j].Name
			case "cpu":
				less = processes[i].CPU < processes[j].CPU
			case "memory", "mem":
				less = processes[i].Memory < processes[j].Memory
			case "pid":
				less = processes[i].PID < processes[j].PID
			default:
				less = processes[i].CPU > processes[j].CPU // Default to CPU desc
			}

			if strings.ToLower(req.Order) == "desc" {
				return !less
			}
			return less
		})
	}

	// Convert to protobuf format
	var pbProcesses []*pb.Process
	for _, proc := range processes {
		pbProcesses = append(pbProcesses, &pb.Process{
			Pid:     proc.PID,
			Name:    proc.Name,
			Cpu:     proc.CPU,
			Memory:  proc.Memory,
			Command: proc.Command,
			Port:    proc.Port,
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

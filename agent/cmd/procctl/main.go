package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "smart-monitor/pbtypes/process"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	backendDefault := getEnv("BACKEND_ADDR", "localhost:50051")

	backend := flag.String("backend", backendDefault, "backend gRPC address")
	hostname := flag.String("hostname", "", "target hostname (required)")
	action := flag.String("action", "list", "action: list|detail|restart")
	pid := flag.Int("pid", 0, "process PID (for detail/restart)")

	flag.Parse()

	if *hostname == "" {
		log.Fatal("hostname is required (use --hostname)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, *backend, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("failed to connect to backend %s: %v", *backend, err)
	}
	defer conn.Close()

	client := pb.NewProcessServiceClient(conn)

	switch *action {
	case "list":
		if err := listProcesses(ctx, client, *hostname); err != nil {
			log.Fatalf("list error: %v", err)
		}
	case "detail":
		if *pid <= 0 {
			log.Fatal("detail requires --pid")
		}
		if err := showProcessDetail(ctx, client, *hostname, int32(*pid)); err != nil {
			log.Fatalf("detail error: %v", err)
		}
	case "restart":
		if *pid <= 0 {
			log.Fatal("restart requires --pid")
		}
		if err := restartProcess(ctx, client, *hostname, int32(*pid)); err != nil {
			log.Fatalf("restart error: %v", err)
		}
	default:
		log.Fatalf("unknown action: %s", *action)
	}
}

func listProcesses(ctx context.Context, client pb.ProcessServiceClient, hostname string) error {
	resp, err := client.GetProcesses(ctx, &pb.GetProcessesRequest{Hostname: hostname})
	if err != nil {
		return fmt.Errorf("GetProcesses RPC failed: %w", err)
	}

	fmt.Printf("Processes on %s (count=%d)\n", hostname, len(resp.Processes))
	fmt.Printf("%-8s %-32s %-10s %-10s\n", "PID", "NAME", "CPU%", "MEM%")
	fmt.Println("--------------------------------------------------------------------")
	for _, p := range resp.Processes {
		fmt.Printf("%-8d %-32s %-10.2f %-10.2f\n", p.Pid, truncate(p.Name, 32), p.Cpu, p.Memory)
	}
	return nil
}

func showProcessDetail(ctx context.Context, client pb.ProcessServiceClient, hostname string, pid int32) error {
	resp, err := client.GetProcesses(ctx, &pb.GetProcessesRequest{Hostname: hostname})
	if err != nil {
		return fmt.Errorf("GetProcesses RPC failed: %w", err)
	}

	for _, p := range resp.Processes {
		if p.Pid == pid {
			fmt.Printf("Process detail on %s (pid=%d)\n", hostname, pid)
			fmt.Printf("Name    : %s\n", p.Name)
			fmt.Printf("CPU %%   : %.2f\n", p.Cpu)
			fmt.Printf("Memory %%: %.2f\n", p.Memory)
			fmt.Printf("Timestamp: %d\n", resp.Timestamp)
			return nil
		}
	}

	return fmt.Errorf("pid %d not found on host %s", pid, hostname)
}

// restartProcess currently issues KillProcess; it assumes an external supervisor restarts the process.
func restartProcess(ctx context.Context, client pb.ProcessServiceClient, hostname string, pid int32) error {
	resp, err := client.KillProcess(ctx, &pb.KillProcessRequest{Hostname: hostname, Pid: pid})
	if err != nil {
		return fmt.Errorf("KillProcess RPC failed: %w", err)
	}

	fmt.Printf("Restart requested for pid=%d on %s\n", pid, hostname)
	fmt.Printf("Backend response: %s (ts=%d)\n", resp.Message, resp.Timestamp)
	fmt.Println("Note: restart assumes a supervisor will relaunch the process after kill.")
	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

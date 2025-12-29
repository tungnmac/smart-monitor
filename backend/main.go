package main

import (
	"io"
	"log"
	"net"

	pb "smart-monitor/proto" // Thay bằng path của bạn

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedMonitorServiceServer
}

func (s *server) StreamStats(stream pb.MonitorService_StreamStatsServer) error {
	for {
		// Nhận dữ liệu liên tục từ stream
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.StatsResponse{Message: "Stream closed"})
		}
		if err != nil {
			return err
		}

		// Tại đây bạn có thể đẩy dữ liệu qua WebSocket cho Frontend
		log.Printf("[%s] CPU: %.2f%% | RAM: %.2f%%", req.Hostname, req.Cpu, req.Ram)
	}
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterMonitorServiceServer(s, &server{})

	log.Println("gRPC Server đang chạy tại port :50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

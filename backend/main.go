package main

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"

	pb "smart-monitor/pbtypes/monitor" // Thay bằng path của bạn

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

func (s *server) GetStats(ctx context.Context, req *pb.StatsRequest) (*pb.StatsResponse, error) {
	log.Printf("GetStats called for %s", req.Hostname)
	return &pb.StatsResponse{Message: "Stats received"}, nil
}

func main() {
	// Start gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterMonitorServiceServer(s, &server{})

	go func() {
		log.Println("gRPC Server đang chạy tại port :50051...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Start HTTP gateway
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Use http.ServeMux as main mux
	httpMux := http.NewServeMux()

	// Create gateway mux
	gwMux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err = pb.RegisterMonitorServiceHandlerFromEndpoint(ctx, gwMux, "localhost:50051", opts)
	if err != nil {
		log.Fatalf("failed to register gateway: %v", err)
	}

	// Mount gateway under /v1/
	httpMux.Handle("/v1/", gwMux)

	// Serve Swagger
	httpMux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../pbtypes/combined.swagger.json")
	})
	// Serve Swagger UI
	httpMux.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", http.FileServer(http.Dir("./static/"))))

	log.Println("HTTP Gateway và Swagger đang chạy tại port :8080...")
	if err := http.ListenAndServe(":8080", httpMux); err != nil {
		log.Fatalf("failed to serve HTTP: %v", err)
	}
}

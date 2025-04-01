package main

import (
	"log/slog"
	"net"

	"grpc-file-server/internal/limiter"
	"grpc-file-server/internal/logger"
	"grpc-file-server/internal/server"
	pb "grpc-file-server/proto"

	"google.golang.org/grpc"
)

func main() {
	log := logger.NewLogger()
	slog.SetDefault(log)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		slog.Error("failed to listen", slog.String("error", err.Error()))
	}
	
	grpcServer := grpc.NewServer()

	srv := &server.FileServer{
		UploadLimiter: limiter.NewSemafore(10),
		DownloadLimiter: limiter.NewSemafore(10),
		ListLimiter: limiter.NewSemafore(100),
	}

	pb.RegisterFileServiceServer(grpcServer, srv)

	slog.Info("gRPC server is running", slog.String("port", ":50051"))


	if err := grpcServer.Serve(lis); err != nil {
		slog.Error("failed to serve", slog.String("error", err.Error()))
	}
}


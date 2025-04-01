package main

import (
	"log"
	"net"


	"google.golang.org/grpc"
	pb "grpc-file-server/proto"
	"grpc-file-server/internal/server"
	"grpc-file-server/internal/limiter"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Can't listen tcp server: %s", err)
	}
	
	grpcServer := grpc.NewServer()

	srv := &server.FileServer{
		UploadLimiter: limiter.NewSemafore(10),
		DownloadLimiter: limiter.NewSemafore(10),
		ListLimiter: limiter.NewSemafore(100),
	}

	pb.RegisterFileServiceServer(grpcServer, srv)

	log.Println("gRPC server is running on port: 50051")


	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}


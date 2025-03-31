package main

import (
	"log"
	"net"


	"google.golang.org/grpc"
	pb "grpc-file-server/proto"
)

type server struct {
	pb.UnimplementedFileServiceServer
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Can't listen tcp server: %s", err)
	}
	
	grpcServer := grpc.NewServer()

	pb.RegisterFileServiceServer(grpcServer, &server{})

	log.Println("gRPC server is running on port: 50051")


	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}


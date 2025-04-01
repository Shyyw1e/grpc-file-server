/*
go build -o server ./cmd/server
go build -o client ./client
*/

package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	pb "grpc-file-server/proto"

	"google.golang.org/grpc"
)

func main () {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		slog.Error("Failed to connect", slog.String("error", err.Error()))
	}
	defer conn.Close()

	client := pb.NewFileServiceClient(conn)

	err = uploadFile(client, "assets/test.txt")
	if err != nil {
		slog.Error("Failed to upload file", slog.String("error", err.Error()))
	}
}

func uploadFile(client pb.FileServiceClient, path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 10)
	defer cancel()

	stream, err := client.UploadFile(ctx)
	if err != nil {
		slog.Error("Failed to create upload stream", slog.String("error", err.Error()))
		return err
	}

	file, err := os.Open(path)
	if err != nil {
		slog.Error("Failed to open file", slog.String("error", err.Error()))
		return err
	}
	defer file.Close()

	filename := filepath.Base(path)
	slog.Info("Starting file upload", slog.String("filename", filename))

	buf := make([]byte, 32 * 1024)

	chunkCount := 0

	/*
	TODO infinite for {} 
	
	*/
	return nil
}
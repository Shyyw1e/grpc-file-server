package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"grpc-file-server/internal/logger"
	pb "grpc-file-server/proto"
)

func main() {
	log := logger.NewLogger()
	slog.SetDefault(log)

	uploadPath := flag.String("upload", "", "Path to file to upload")
	downloadFile := flag.String("download", "", "Filename to download")
	outPath := flag.String("out", "", "Where to save downloaded file")
	listFiles := flag.Bool("list", false, "List all files on server")

	flag.Parse()

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("connection error", slog.String("error", err.Error()))
		return
	}
	defer conn.Close()

	client := pb.NewFileServiceClient(conn)

	switch {
	case *uploadPath != "":
		if err := uploadFile(client, *uploadPath); err != nil {
			slog.Error("upload failed", slog.String("error", err.Error()))
		}

	case *downloadFile != "":
		out := *outPath
		if out == "" {
			out = filepath.Join("downloads", *downloadFile)
		}
		if err := downloadFileFunc(client, *downloadFile, out); err != nil {
			slog.Error("download failed", slog.String("error", err.Error()))
		}

	case *listFiles:
		if err := listFilesFunc(client); err != nil {
			slog.Error("list files failed", slog.String("error", err.Error()))
		}

	default:
		fmt.Println("Usage:")
		flag.PrintDefaults()
	}
}

func uploadFile(client pb.FileServiceClient, path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.UploadFile(ctx)
	if err != nil {
		slog.Error("failed to create upload stream", slog.String("error", err.Error()))
		return err
	}

	file, err := os.Open(path)
	if err != nil {
		slog.Error("failed to open file", slog.String("path", path), slog.String("error", err.Error()))
		return err
	}
	defer file.Close()

	filename := filepath.Base(path)
	slog.Info("starting file upload", slog.String("filename", filename))

	buf := make([]byte, 32*1024)
	chunkCount := 0

	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Error("failed to read chunk", slog.String("error", err.Error()))
			return err
		}

		err = stream.Send(&pb.FileChunk{
			Filename: filename,
			Content:  buf[:n],
		})
		if err != nil {
			slog.Error("failed to send chunk", slog.Int("chunk", chunkCount), slog.String("error", err.Error()))
			return err
		}
		chunkCount++
		slog.Debug("chunk sent", slog.Int("chunk_size", n), slog.Int("chunk_num", chunkCount))
	}

	status, err := stream.CloseAndRecv()
	if err != nil {
		slog.Error("failed to receive upload status", slog.String("error", err.Error()))
		return err
	}

	slog.Info("upload completed", slog.String("message", status.Message), slog.Bool("success", status.Success))
	return nil
}

func downloadFileFunc(client pb.FileServiceClient, filename string, savePath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req := &pb.FileRequest{Filename: filename}
	stream, err := client.DownloadFile(ctx, req)
	if err != nil {
		slog.Error("failed to start DownloadFile stream", slog.String("error", err.Error()))
		return err
	}

	f, err := os.Create(savePath)
	if err != nil {
		slog.Error("failed to create file", slog.String("path", savePath), slog.String("error", err.Error()))
		return err
	}
	defer f.Close()

	slog.Info("starting download", slog.String("filename", filename))

	totalBytes := 0
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			slog.Info("download completed", slog.Int("bytes", totalBytes), slog.String("saved_to", savePath))
			break
		}
		if err != nil {
			slog.Error("failed to receive chunk", slog.String("error", err.Error()))
			return err
		}

		n, err := f.Write(chunk.Content)
		if err != nil {
			slog.Error("failed to write chunk", slog.String("error", err.Error()))
			return err
		}

		totalBytes += n
		slog.Debug("chunk written", slog.Int("size", n))
	}

	return nil
}

func listFilesFunc(client pb.FileServiceClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.ListFiles(ctx, &pb.Empty{})
	if err != nil {
		slog.Error("failed to start ListFiles stream", slog.String("error", err.Error()))
		return err
	}

	slog.Info("retrieving file list...")

	count := 0
	for {
		info, err := stream.Recv()
		if err == io.EOF {
			slog.Info("file list completed", slog.Int("total_files", count))
			break
		}
		if err != nil {
			slog.Error("failed to receive file info", slog.String("error", err.Error()))
			return err
		}

		slog.Info("file info",
			slog.String("name", info.Filename),
			slog.String("created_at", info.CreatedAt),
			slog.String("updated_at", info.UpdatedAt),
		)
		count++
	}
	return nil
}

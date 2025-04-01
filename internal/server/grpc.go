package server

import (
	
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"grpc-file-server/internal/limiter"
	pb "grpc-file-server/proto"
)

type FileServer struct {
	pb.UnimplementedFileServiceServer
	UploadLimiter limiter.Semafore
	DownloadLimiter limiter.Semafore
	ListLimiter limiter.Semafore
	
}

func (s *FileServer)UploadFile(stream pb.FileService_UploadFileServer) error {
	s.UploadLimiter.Acquire()
	defer s.UploadLimiter.Release()

	var file *os.File
	var filename string

	slog.Info("Upload started")


	for  {
		chunk, err := stream.Recv()
		if err == io.EOF {
			slog.Info("Upload finished", slog.String("filename", filename))
			return stream.SendAndClose(&pb.UploadStatus{
				Message: "File uploaded succesfully",
				Success: true,
			})
		}
		if err != nil {
			slog.Error("Failed to receive chunk", slog.String("error", err.Error()))
			return err
		}
		if file == nil {
			filename = filepath.Base(chunk.GetFilename())
			path := filepath.Join("storage", filename)
			slog.Info("Creating file", slog.String("path", path))
			file, err = os.Create(path)
			if err != nil {
				slog.Error("Failed to create file", slog.String("error", err.Error()))
				return err
			}
			defer file.Close()
		}

		if _, err := file.Write(chunk.GetContent()); err != nil {
			slog.Error("Failed to write to file", slog.String("error", err.Error()))
			return err
		}
	}
}

func (s *FileServer)DownloadFile(req *pb.FileRequest,stream pb.FileService_DownloadFileServer) error {
	s.DownloadLimiter.Acquire()
	defer s.DownloadLimiter.Release()

	filename := filepath.Base(req.GetFilename())
	path := filepath.Join("storage", filename)

	file, err := os.Open(path)
	if err != nil {
		slog.Error("Failed to open file", slog.String("error", err.Error()))
			return err
	}
	defer file.Close()

	buf := make([]byte, 1024 * 32)

	slog.Info("Download started", slog.String("filename", filename))

	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		
		if err != nil {
			slog.Error("Failed to read file", slog.String("error", err.Error()))
			return err
		}

		err = stream.Send(&pb.FileChunk{
			Filename: filename,
			Content: buf[:n],
		})
		if err != nil {
			slog.Error("Failed to send chunk", slog.String("error", err.Error()))
			return err
		}


	
	}

	slog.Info("Download completed", slog.String("filename", filename))
	return nil
}


func (s *FileServer)ListFiles(_ *pb.Empty, stream pb.FileService_ListFilesServer) error {
	s.ListLimiter.Acquire()
	defer s.ListLimiter.Release()

	dir := "storage"

	entries, err := os.ReadDir(dir)
	if err != nil {
		 return fmt.Errorf("failed to read storage directory: %s", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}
		modtime := info.ModTime().Format(time.RFC3339)
		
		err = stream.Send(&pb.FileInfo{
			Filename: entry.Name(),
			CreatedAt: modtime,
			UpdatedAt: modtime,
		})
		if err != nil {
			return fmt.Errorf("failed to send file info: %s", err)
		}
	}
	
	return nil
}
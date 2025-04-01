package server

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	pb "grpc-file-server/proto"
)

type FileServer struct {
	pb.UnimplementedFileServiceServer
}

func (s *FileServer)UploadFile(stream pb.FileService_UploadFileServer) error {
	var file *os.File
	var filename string

	for  {
		chunk, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.UploadStatus{
				Message: "File uploaded succesfully",
				Success: true,
			})
		}
		if err != nil {
			return err
		}
		if file == nil {
			filename = filepath.Base(chunk.GetFilename())
			path := filepath.Join("storage", filename)
			file, err := os.Create(path)
			if err != nil {
				return nil
			}
			defer file.Close()
		}

		if _, err := file.Write(chunk.GetContent()); err != nil {
			return err
		}
	}
}

func (s *FileServer)DownloadFile(req *pb.FileRequest,stream pb.FileService_DownloadFileServer) error {
	filename := filepath.Base(req.GetFilename())
	path := filepath.Join("storage", filename)

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %s", err)
	}
	defer file.Close()

	buf := make([]byte, 1024 * 32)

	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		
		if err != nil {
			return fmt.Errorf("failed to read file: %s",err)
		}

		err = stream.Send(&pb.FileChunk{
			Filename: filename,
			Content: buf[:n],
		})
		if err != nil {
			return fmt.Errorf("failed to send chunk: %s", err)
		}


	
	}

	return nil
}


func (s *FileServer)ListFiles(_ *pb.Empty, stream pb.FileService_ListFilesServer) error {
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
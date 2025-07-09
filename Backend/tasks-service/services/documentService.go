package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"tasks-service/model"
	"tasks-service/repositories"
	"time"

	"github.com/colinmarc/hdfs"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DocumentService struct {
	repo       *repositories.DocumentRepo
	hdfsClient *hdfs.Client
}

func NewDocumentService(repo *repositories.DocumentRepo, hdfsClient *hdfs.Client) *DocumentService {
	return &DocumentService{
		repo:       repo,
		hdfsClient: hdfsClient,
	}
}

func (service *DocumentService) UploadDocument(ctx context.Context, taskID string, userID string, fileName string, fileContent []byte) (*model.Document, error) {
	hdfsDir := "/tasks/" + taskID
	hdfsPath := path.Join(hdfsDir, fileName)

	// Ensure the directory exists
	err := service.ensureHDFSDirectoryExists(hdfsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure HDFS directory exists: %v", err)
	}

	// Upload to HDFS
	file, err := service.hdfsClient.Create(hdfsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file in HDFS: %v", err)
	}
	defer file.Close()

	_, err = file.Write(fileContent)
	if err != nil {
		return nil, fmt.Errorf("failed to write file content to HDFS: %v", err)
	}

	// Convert IDs
	taskObjectID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return nil, fmt.Errorf("invalid taskID format: %v", err)
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid userID format: %v", err)
	}

	// Save metadata to MongoDB
	document := &model.Document{
		TaskID:     taskObjectID,
		FileName:   fileName,
		FilePath:   hdfsPath,
		UploadedBy: userObjectID,
		UploadedAt: time.Now().Unix(),
	}

	result, err := service.repo.InsertDocument(ctx, document)
	if err != nil {
		return nil, fmt.Errorf("failed to save document in MongoDB: %v", err)
	}
	document.ID = result.InsertedID.(primitive.ObjectID)

	return document, nil
}

func (service *DocumentService) ensureHDFSDirectoryExists(dir string) error {
	_, err := service.hdfsClient.Stat(dir)
	if err != nil {
		// If the error is "not exist", try to create the directory
		if os.IsNotExist(err) {
			return service.hdfsClient.MkdirAll(dir, os.ModePerm)
		}
		// Return any other error
		return err
	}
	return nil // Directory exists
}

func (service *DocumentService) GetDocumentsByTaskID(ctx context.Context, taskID string) ([]model.Document, error) {
	documents, err := service.repo.FindDocumentsByTaskID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve documents: %v", err)
	}

	if documents == nil {
		// Ensure empty slice, not nil
		documents = []model.Document{}
	}

	return documents, nil
}

func (service *DocumentService) DownloadDocument(filePath string) ([]byte, error) {
	file, err := service.hdfsClient.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file in HDFS: %v", err)
	}
	defer file.Close()

	var content []byte
	buffer := make([]byte, 4096)

	for {
		n, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read file content: %v", err)
		}
		content = append(content, buffer[:n]...)
	}

	return content, nil
}

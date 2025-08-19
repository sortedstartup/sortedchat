package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sortedstartup/inferenceservice/dao"
	"time"
)

type InferenceService struct {
	dao dao.DAO
}

func NewInferenceService(daoFactory dao.DAOFactory) *InferenceService {
	dao, err := daoFactory.CreateDAO()
	if err != nil {
		log.Fatalf("Failed to create DAO: %v", err)
	}
	return &InferenceService{
		dao: dao,
	}
}

func (s *InferenceService) DownloadModel(ctx context.Context, userID string, modelName string) error {
	// Get model metadata from database
	model, err := s.dao.GetModelByName(modelName)
	if err != nil {
		return fmt.Errorf("model not found: %w", err)
	}

	if !model.IsDownloadable {
		return fmt.Errorf("model %s is not downloadable", modelName)
	}

	if model.Status == dao.StatusDownloading {
		return fmt.Errorf("model %s is already downloading", modelName)
	}

	if model.Status == dao.StatusCompleted {
		return fmt.Errorf("model %s is already downloaded", modelName)
	}

	// Start download in goroutine
	go func() {
		if err := s.downloadModelFromURL(model.ID, model.Name, model.URL); err != nil {
			log.Printf("Failed to download model %s: %v", model.Name, err)
			// Update status to failed
			failedProgress := &dao.DownloadProgress{
				FileSize: 0,
				Status:   dao.StatusFailed,
				Progress: 0,
				Speed:    0,
			}
			s.dao.UpdateModelProgress(model.ID, failedProgress)
		}
	}()

	return nil
}

func (s *InferenceService) downloadModelFromURL(modelID string, modelName string, url string) error {
	log.Printf("Starting download of model %s from %s", modelName, url)

	// Update status to downloading
	downloadingProgress := &dao.DownloadProgress{
		FileSize: 0,
		Status:   dao.StatusDownloading,
		Progress: 0,
		Speed:    0,
	}
	s.dao.UpdateModelProgress(modelID, downloadingProgress)

	// Create HTTP client
	client := &http.Client{}

	// Make HTTP request
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	// Get file size from Content-Length header
	fileSize := resp.ContentLength
	if fileSize <= 0 {
		fileSize = 0 // Unknown file size
	}

	modelDir := filepath.Join("filestore", "models")
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return fmt.Errorf("failed to create model directory: %w", err)
	}

	filename := fmt.Sprintf("%s.model", modelName)
	filePath := filepath.Join(modelDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Create progress tracking writer
	progressWriter := &ProgressWriter{
		modelID:    modelID,
		fileSize:   fileSize,
		writer:     file,
		downloaded: 0,
		startTime:  time.Now(),
		lastUpdate: time.Now(),
		onProgress: func(progress *dao.DownloadProgress) {
			s.dao.UpdateModelProgress(modelID, progress)
		},
	}

	// Copy with progress tracking
	_, err = io.Copy(progressWriter, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	// Mark as completed
	completedProgress := &dao.DownloadProgress{
		FileSize: fileSize,
		Status:   dao.StatusCompleted,
		Progress: 100,
		Speed:    0,
	}
	s.dao.UpdateModelProgress(modelID, completedProgress)

	// Update the filestore_id in the database to point to the downloaded file
	if err := s.dao.UpdateModelFileStoreID(modelID, filePath); err != nil {
		log.Printf("Warning: Failed to update filestore_id for model %s: %v", modelName, err)
	}

	log.Printf("Successfully downloaded model %s to %s", modelName, filePath)
	return nil
}

// ProgressWriter wraps an io.Writer to track download progress
type ProgressWriter struct {
	modelID    string
	fileSize   int64
	writer     io.Writer
	downloaded int64
	startTime  time.Time
	lastUpdate time.Time
	onProgress func(*dao.DownloadProgress)
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	if err != nil {
		return n, err
	}

	pw.downloaded += int64(n)
	now := time.Now()

	if now.Sub(pw.lastUpdate) >= 3*time.Second {
		pw.updateProgress()
		pw.lastUpdate = now
	}

	return n, nil
}

func (pw *ProgressWriter) updateProgress() {
	var progress int
	var speed int64

	if pw.fileSize > 0 {
		progress = int((pw.downloaded * 100) / pw.fileSize)
	} else {
		progress = 0
	}

	elapsed := time.Since(pw.startTime).Seconds()
	if elapsed > 0 {
		speed = int64(float64(pw.downloaded) / elapsed / 1024) // kilobytes per second
	}

	progressData := &dao.DownloadProgress{
		FileSize: pw.fileSize,
		Status:   dao.StatusDownloading,
		Progress: progress,
		Speed:    speed,
	}

	if pw.onProgress != nil {
		pw.onProgress(progressData)
	}
}

// ListModels returns all models and streams updates for downloading models
func (s *InferenceService) ListLLMModels(ctx context.Context, sendModels func([]*dao.ModelMetadata) error) error {
	// Get all models initially
	models, err := s.dao.GetAllModels()
	if err != nil {
		return fmt.Errorf("failed to get models: %w", err)
	}

	// Send initial list
	if err := sendModels(models); err != nil {
		return err
	}

	// Check if any models are downloading
	hasDownloadingModels := false
	for _, model := range models {
		if model.Status == dao.StatusDownloading {
			hasDownloadingModels = true
			break
		}
	}

	// If no downloading models, we're done
	if !hasDownloadingModels {
		return nil
	}

	// Stream updates every 3 seconds for downloading models
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Get updated models
			updatedModels, err := s.dao.GetAllModels()
			if err != nil {
				log.Printf("Error getting updated models: %v", err)
				continue
			}

			// Check if any models are still downloading
			stillDownloading := false
			for _, model := range updatedModels {
				if model.Status == dao.StatusDownloading {
					stillDownloading = true
					break
				}
			}

			// Send updated models
			if err := sendModels(updatedModels); err != nil {
				return err
			}

			// If no models are downloading anymore, stop streaming
			if !stillDownloading {
				return nil
			}
		}
	}
}

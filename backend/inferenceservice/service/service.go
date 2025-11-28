package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sortedstartup/inferenceservice/dao"
	"sync"
	"time"
)

const UPDATE_INTERVAL = 3 * time.Second

type InferenceService struct {
	dao                dao.DAO
	downloadingCancels map[string]context.CancelFunc
	mu                 sync.Mutex
}

func NewInferenceService(daoFactory dao.DAOFactory) *InferenceService {
	slog.Info("inferenceservice:service:NewInferenceService")
	dao, err := daoFactory.CreateDAO()
	if err != nil {
		log.Fatalf("Failed to create DAO: %v", err)
	}
	return &InferenceService{
		dao:                dao,
		downloadingCancels: make(map[string]context.CancelFunc),
		mu:                 sync.Mutex{},
	}
}

// Initialize run on startup to reset models that were left in a pending state
func (s *InferenceService) Initialize() error {
	slog.Info("inferenceservice:service:Initialize")
	models, err := s.dao.GetAllModels()
	if err != nil {
		slog.Error("inferenceservice:service:Initialize", "message", "failed to get all models", "error", err)
		return fmt.Errorf("failed to get all models: %w", err)
	}
	for _, model := range models {
		if model.Status == dao.StatusDownloading {
			err := s.dao.UpdateModelProgress(model.ID, &dao.DownloadProgress{
				FileSize: 0,
				Status:   dao.StatusFailed,
				Progress: 0,
				Speed:    0,
			})
			if err != nil {
				return fmt.Errorf("failed to update model progress: %w", err)
			}
			err = s.deleteFilestoreObject(*model.FileStoreID)
			if err != nil {
				return fmt.Errorf("failed to delete filestore object: %w", err)
			}
		}
	}
	return nil
}

func (s *InferenceService) DownloadModel(ctx context.Context, userID string, modelName string) error {
	slog.Info("inferenceservice:service:DownloadModel")
	// Get model metadata from database
	model, err := s.dao.GetModelByName(modelName)
	if err != nil {
		slog.Error("inferenceservice:service:DownloadModel", "error", err)
		return fmt.Errorf("model not found")
	}

	if !model.IsDownloadable {
		slog.Error("inferenceservice:service:DownloadModel", "message", "model is not downloadable", "modelName", modelName)
		return fmt.Errorf("model %s is not downloadable", modelName)
	}

	if model.Status == dao.StatusDownloading {
		slog.Error("inferenceservice:service:DownloadModel", "message", "model is already downloading", "modelName", modelName)
		return fmt.Errorf("model %s is already downloading", modelName)
	}

	if model.Status == dao.StatusCompleted {
		slog.Error("inferenceservice:service:DownloadModel", "message", "model is already downloaded", "modelName", modelName)
		return fmt.Errorf("model %s is already downloaded", modelName)
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	s.downloadingCancels[modelName] = cancel
	s.mu.Unlock()

	// Starting downloading in a new goroutine (background task)
	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.downloadingCancels, modelName)
			s.mu.Unlock()
		}()

		if err := s.downloadModelFromURL(ctx, model.ID, model.Name, model.URL); err != nil {
			if errors.Is(err, context.Canceled) {
				slog.Error("inferenceservice:service:DownloadModel", "message", "context canceled", "modelName", modelName)
				if resetErr := s.dao.ResetModelToInitialState(model.ID); resetErr != nil {
					slog.Error("inferenceservice:service:DownloadModel", "message", "failed to reset model to initial state", "modelName", modelName, "error", resetErr)
				} else {
					slog.Info("inferenceservice:service:DownloadModel", "message", "successfully reset model to initial state", "modelName", modelName)
				}
			} else {
				// Update status to failed
				failedProgress := &dao.DownloadProgress{
					FileSize: 0,
					Status:   dao.StatusFailed,
					Progress: 0,
					Speed:    0,
				}
				if updateErr := s.dao.UpdateModelProgress(model.ID, failedProgress); updateErr != nil {
					slog.Error("inferenceservice:service:DownloadModel", "message", "failed to update failed status", "modelName", modelName, "error", updateErr)
				}
				err := s.deleteFilestoreObject(*model.FileStoreID)
				if err != nil {
					slog.Error("inferenceservice:service:DownloadModel", "message", "failed to delete filestore object", "modelName", modelName, "error", err)
				}
			}
		}
	}()

	return nil
}

func (s *InferenceService) downloadModelFromURL(ctx context.Context, modelID string, modelName string, url string) error {
	slog.Info("inferenceservice:service:downloadModelFromURL", "modelID", modelID, "modelName", modelName, "url", url)
	// Update status to downloading
	downloadingProgress := &dao.DownloadProgress{
		FileSize: 0,
		Status:   dao.StatusDownloading,
		Progress: 0,
		Speed:    0,
	}
	s.dao.UpdateModelProgress(modelID, downloadingProgress)
	slog.Info("inferenceservice:service:downloadModelFromURL", "message", "Updated model progress to downloading", "modelID", modelID, "modelName", modelName, "url", url)

	// Create HTTP client
	client := &http.Client{}

	// Make HTTP request
	resp, err := client.Get(url)
	if err != nil {
		slog.Error("inferenceservice:service:downloadModelFromURL", "message", "failed to make HTTP request", "modelName", modelName, "error", err)
		return fmt.Errorf("failed to make HTTP request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("inferenceservice:service:downloadModelFromURL", "message", "HTTP request failed with status", "modelName", modelName, "statusCode", resp.StatusCode)
		return fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	// Get file size from Content-Length header
	fileSize := resp.ContentLength
	if fileSize <= 0 {
		fileSize = 0 // Unknown file size
	}

	modelDir := filepath.Join("filestore", "models")
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		slog.Error("inferenceservice:service:downloadModelFromURL", "message", "failed to create model directory", "modelName", modelName, "error", err)
		return fmt.Errorf("failed to create model directory")
	}

	filename := fmt.Sprintf("%s.model", modelName)
	filePath := filepath.Join(modelDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		slog.Error("inferenceservice:service:downloadModelFromURL", "message", "failed to create file", "modelName", modelName, "error", err)
		return fmt.Errorf("failed to create file")
	}
	defer file.Close()

	// Update the filestore_id in the database to point to the downloaded file
	if err := s.dao.UpdateModelFileStoreID(modelID, filePath); err != nil {
		slog.Error("inferenceservice:service:downloadModelFromURL", "message", "failed to update filestore_id", "modelName", modelName, "error", err)
	}

	// Create progress tracking writer
	progressWriter := &ProgressWriter{
		modelID:    modelID,
		fileSize:   fileSize,
		writer:     file,
		downloaded: 0,
		startTime:  time.Now(),
		lastUpdate: time.Now().Add(-UPDATE_INTERVAL), // Force immediate update
		onProgress: func(progress *dao.DownloadProgress) {
			s.dao.UpdateModelProgress(modelID, progress)
		},
	}

	// Context-aware reader
	ctxReader := &ctxReader{ctx: ctx, r: resp.Body}

	// Copy with progress tracking
	_, err = io.Copy(progressWriter, ctxReader)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			file.Close()
			s.deleteFilestoreObject(filePath)
			return ctx.Err()
		}
		slog.Error("inferenceservice:service:downloadModelFromURL", "message", "failed to copy file", "modelName", modelName, "error", err)
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
	slog.Info("inferenceservice:service:downloadModelFromURL", "message", "Updated model progress to completed", "modelID", modelID, "modelName", modelName, "url", url)
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
		slog.Error("inferenceservice:service:ProgressWriter:Write", "message", "failed to write file", "modelName", pw.modelID, "error", err)
		return n, err
	}

	pw.downloaded += int64(n)
	now := time.Now()

	//TODO: This calculation may have high CPU cost
	if now.Sub(pw.lastUpdate) >= UPDATE_INTERVAL {
		pw.updateProgress()
		pw.lastUpdate = now
	}

	return n, nil
}

func (pw *ProgressWriter) updateProgress() {
	slog.Info("inferenceservice:service:ProgressWriter:updateProgress")
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

type ctxReader struct {
	ctx context.Context
	r   io.Reader
}

func (cr *ctxReader) Read(p []byte) (int, error) {
	select {
	case <-cr.ctx.Done():
		return 0, cr.ctx.Err()
	default:
		return cr.r.Read(p)
	}
}

// ListModels returns all models and streams updates for downloading models
func (s *InferenceService) GetLLMModels(ctx context.Context, sendModels func([]*dao.ModelMetadata) error) error {
	slog.Info("inferenceservice:service:GetLLMModels")
	// Get all models initially
	models, err := s.dao.GetAllModels()
	if err != nil {
		slog.Error("inferenceservice:service:GetLLMModels", "message", "failed to get models", "error", err)
		return fmt.Errorf("failed to get models")
	}

	// Send initial list
	if err := sendModels(models); err != nil {
		slog.Error("inferenceservice:service:GetLLMModels", "message", "failed to send models", "error", err)
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
		slog.Info("inferenceservice:service:GetLLMModels", "message", "no downloading models")
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
				slog.Error("inferenceservice:service:GetLLMModels", "message", "failed to get updated models", "error", err)
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
				slog.Error("inferenceservice:service:GetLLMModels", "message", "failed to send updated models", "error", err)
				return err
			}

			// If no models are downloading anymore, stop streaming
			if !stillDownloading {
				slog.Info("inferenceservice:service:GetLLMModels", "message", "no downloading models anymore")
				return nil
			}
		}
	}
}

func (s *InferenceService) CancelDownload(ctx context.Context, userID string, modelName string) error {
	slog.Info("inferenceservice:service:CancelDownload")
	// Get model metadata from database
	model, err := s.dao.GetModelByName(modelName)
	if err != nil {
		slog.Error("inferenceservice:service:CancelDownload", "message", "failed to get model", "error", err)
		return fmt.Errorf("model not found")
	}

	if model.Status != dao.StatusDownloading {
		slog.Error("inferenceservice:service:CancelDownload", "message", "model is not downloading", "modelName", modelName)
		return fmt.Errorf("model %s is not downloading", modelName)
	}

	s.mu.Lock()
	if cancel, ok := s.downloadingCancels[modelName]; ok {
		cancel()
	}
	s.mu.Unlock()

	// This was leading to a unexpected situation some times
	// the UI ws stuck in cancelling state if there was a databas error
	// the db error was happening in sqlite without WAL due to read & write locks
	// err = s.dao.UpdateModelProgress(model.ID, &dao.DownloadProgress{
	// 	FileSize: 0,
	// 	Status:   dao.StatusCancelling,
	// 	Progress: 0,
	// 	Speed:    0,
	// })
	// if err != nil {
	// 	return fmt.Errorf("failed to update model progress: %w", err)
	// }

	slog.Info("inferenceservice:service:CancelDownload", "message", "Download cancellation initiated for model", "modelName", modelName)
	return nil
}

func (s *InferenceService) DeleteModel(ctx context.Context, userID string, modelName string) error {
	slog.Info("inferenceservice:service:DeleteModel")
	// Get model metadata from database
	model, err := s.dao.GetModelByName(modelName)
	if err != nil {
		slog.Error("inferenceservice:service:DeleteModel", "message", "failed to get model", "error", err)
		return fmt.Errorf("model not found")
	}

	if !model.IsDownloadable {
		slog.Error("inferenceservice:service:DeleteModel", "message", "model is not downloadable", "modelName", modelName)
		return fmt.Errorf("model %s is not downloadable", modelName)
	}

	// Check if model is currently downloading and cancel if needed
	if model.Status != dao.StatusCompleted {
		slog.Error("inferenceservice:service:DeleteModel", "message", "model is not downloaded", "modelName", modelName)
		return fmt.Errorf("model %s is not downloaded", modelName)
	}

	// Delete the physical file if it exists
	if model.FileStoreID != nil {
		err := s.deleteFilestoreObject(*model.FileStoreID)
		if err != nil {
			slog.Error("inferenceservice:service:DeleteModel", "message", "failed to delete filestore object", "modelName", modelName, "error", err)
			return fmt.Errorf("failed to delete filestore object")
		}
		// Don't return error here - we still want to delete from database
	}

	if err := s.dao.ResetModelToInitialState(model.ID); err != nil {
		slog.Error("inferenceservice:service:DeleteModel", "message", "failed to reset model to initial state", "modelName", modelName, "error", err)
		return fmt.Errorf("failed to reset model to initial state")
	}

	return nil
}

// deleteFilestoreObject safely deletes a file from the filestore and logs any errors
func (s *InferenceService) deleteFilestoreObject(filePath string) error {
	slog.Info("inferenceservice:service:deleteFilestoreObject", "filePath", filePath)
	if filePath == "" {

		return nil // Nothing to delete
	}

	if err := os.Remove(filePath); err != nil {
		slog.Error("inferenceservice:service:deleteFilestoreObject", "message", "failed to delete filestore object", "filePath", filePath, "error", err)
		return fmt.Errorf("failed to delete filestore object")
	}

	slog.Info("inferenceservice:service:deleteFilestoreObject", "message", "Successfully deleted filestore object", "filePath", filePath)
	return nil
}

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"sortedstartup/chatservice/queue"
	"sortedstartup/inferenceservice/dao"
	"sortedstartup/inferenceservice/llama"
	"strings"
	"sync"
	"time"
)

const UPDATE_INTERVAL = 3 * time.Second

type InferenceService struct {
	dao                dao.DAO
	downloadingCancels map[string]context.CancelFunc
	mu                 sync.Mutex
	queue              queue.Queue
	proxyAddr          string
}

func NewInferenceService(daoFactory dao.DAOFactory, queue queue.Queue) *InferenceService {
	slog.Info("inferenceservice:service:NewInferenceService")
	dao, err := daoFactory.CreateDAO()
	if err != nil {
		log.Fatalf("Failed to create DAO: %v", err)
	}
	return &InferenceService{
		dao:                dao,
		downloadingCancels: make(map[string]context.CancelFunc),
		mu:                 sync.Mutex{},
		queue:              queue,
		proxyAddr:          ":8081",
	}
}

/*
TODO:

// 1. On initialize we have to load all downloaded models from DB and update the ModelRegistry

// 2. When a model is downloaded successfully, send out a event to the queue
s.queue.Publish(ctx, "model.downloaded", msgBytes)
the queue should be stored in InferenceService like - 	queue queue.Queue, look at SettingsService for details

// 3. this should be registered in intialize(), may be create a new function registerListeners
// we should update the ModelRegistry with model and paths, thats it !
s.queue.Subscribe(ctx, "model.downloaded", handler)

// 4. Write code for startLLamaServerProxy based on cmd/main.go
    this startLLamaServerProxy should be done in the Initialize function
	in a seperate go routine
   the server and port of this server should saved in the InferenceService struct
*/

func (s *InferenceService) startLLamaServerProxy() {
	slog.Info("inferenceservice:service:startLLamaServerProxy", "addr", s.proxyAddr)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// We need to read the body to extract the model, but also keep it for the proxy
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		slog.Debug("inferenceservice:service:startLLamaServerProxy", "body", string(bodyBytes))

		// Restore the io.ReadCloser to its original state
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Parse model from body
		var reqBody struct {
			Model string `json:"model"`
		}
		if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
			// If we can't parse JSON, maybe it's not a chat completion request or similar.
			// But for this specific requirement, we expect "model" in request.
			// Let's proceed with caution or return error.
			// Given the prompt implies routing based on this, we should probably error if missing.
			http.Error(w, "Invalid JSON or missing 'model' field", http.StatusBadRequest)
			return
		}

		if reqBody.Model == "" {
			http.Error(w, "Model field is required", http.StatusBadRequest)
			return
		}

		//TODO: remove llama- prefix
		reqBody.Model = strings.TrimPrefix(reqBody.Model, "llama-")

		// Get or start server for the model

		// If the url contains /embedding, then it is an embedding request,
		// the llama-server needs an extra flag -embedding to treat it specially and create a /embedding endpoint
		isEmbeddingModel := false
		if strings.Contains(r.RequestURI, "/embedding") {
			slog.Debug("inferenceservice:service:startLLamaServerProxy", "isEmbeddingModel", true)
			isEmbeddingModel = true
		}
		socketPath, err := llama.GetOrStartServer(reqBody.Model, isEmbeddingModel)
		if err != nil {
			slog.Error("Error getting server for model", "model", reqBody.Model, "error", err)
			http.Error(w, fmt.Sprintf("Failed to load model: %v", err), http.StatusInternalServerError)
			return
		}

		// Proxy the request to the unix socket
		s.proxyToSocket(w, r, socketPath)
	})

	slog.Info("Starting proxy server", "addr", s.proxyAddr)
	if err := http.ListenAndServe(s.proxyAddr, nil); err != nil {
		slog.Error("Server failed", "error", err)
	}
}

func (s *InferenceService) proxyToSocket(w http.ResponseWriter, r *http.Request, socketPath string) {
	// Define the dialer for Unix socket
	dialer := func(network, addr string) (net.Conn, error) {
		return net.Dial("unix", socketPath)
	}

	// Create a reverse proxy
	// The target URL doesn't matter much for Unix sockets, but we need a valid URL struct
	targetUrl, _ := url.Parse("http://unix")

	proxy := httputil.NewSingleHostReverseProxy(targetUrl)

	// Override the transport to use our Unix socket dialer
	proxy.Transport = &http.Transport{
		Dial: dialer,
	}

	// Update the request URL scheme and host to match what the transport expects (though transport ignores host for unix)
	r.URL.Scheme = "http"
	r.URL.Host = "unix"

	// Serve
	proxy.ServeHTTP(w, r)
}

// Initialize run on startup to reset models that were left in a pending state
func (s *InferenceService) Initialize() error {
	slog.Info("inferenceservice:service:Initialize")
	models, err := s.dao.GetAllModels()
	if err != nil {
		slog.Error("inferenceservice:service:Initialize", "message", "failed to get all models", "error", err)
		return fmt.Errorf("failed to get all models: %w", err)
	}

	// 1. On initialize we have to load all downloaded models from DB and update the ModelRegistry
	for _, model := range models {
		if model.IsDownloaded && model.FileStoreID != nil {
			// Update ModelRegistry (assuming llama package has a way to register or we just rely on it checking disk)
			// The llama package currently has a hardcoded registry, but also checks disk.
			// We might need to update the llama package to allow dynamic registration if the hardcoded map isn't enough.
			// For now, based on the prompt "we should update the ModelRegistry with model and paths",
			// I'll assume we can directly modify the map or add a function.
			// Since ModelRegistry is a public map in llama package:
			absPath, err := getModelAbsolutePath(model.Name)
			if err != nil {
				slog.Error("Failed to get model absolute path", "model", model.Name, "error", err)
				continue
			}
			llama.ModelRegistry[model.Name] = absPath
		}
	}

	// 3. Register listeners
	go s.registerListeners()

	// 4. Start Proxy
	go s.startLLamaServerProxy()

	// 5. Reset models that were left in a pending state
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

func (s *InferenceService) registerListeners() {
	ctx := context.Background()
	ch, err := s.queue.Subscribe(ctx, "model.downloaded")
	if err != nil {
		slog.Error("Failed to subscribe to model.downloaded", "error", err)
		return
	}

	for msg := range ch {
		slog.Info("Received model.downloaded event", "data", string(msg.Data))
		// we should update the ModelRegistry with model and paths
		var eventData struct {
			ModelName string `json:"modelName"`
			FilePath  string `json:"filePath"`
		}
		if err := json.Unmarshal(msg.Data, &eventData); err != nil {
			slog.Error("Failed to unmarshal event data", "error", err)
			continue
		}

		llama.ModelRegistry[eventData.ModelName] = eventData.FilePath
		slog.Info("Updated ModelRegistry", "model", eventData.ModelName, "path", eventData.FilePath)
	}
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

	filePath, err := getModelAbsolutePath(modelName)
	if err != nil {
		return err
	}

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

	// 2. When a model is downloaded successfully, send out a event to the queue
	eventData := struct {
		ModelName string `json:"modelName"`
		FilePath  string `json:"filePath"`
	}{
		ModelName: modelName,
		FilePath:  filePath,
	}
	msgBytes, _ := json.Marshal(eventData)
	if err := s.queue.Publish(ctx, "model.downloaded", msgBytes); err != nil {
		slog.Error("Failed to publish model.downloaded event", "error", err)
	}

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

func getModelAbsolutePath(modelName string) (string, error) {
	modelDir := filepath.Join("filestore", "models")
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		slog.Error("inferenceservice:service:getModelAbsolutePath", "message", "failed to create model directory", "modelName", modelName, "error", err)
		return "", fmt.Errorf("failed to create model directory")
	}

	filename := fmt.Sprintf("%s.model", modelName)
	return filepath.Join(modelDir, filename), nil
}

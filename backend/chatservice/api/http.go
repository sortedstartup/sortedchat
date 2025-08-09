package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sortedstartup/chatservice/events"
	"strings"

	"github.com/google/uuid"
)

const (
	MaxFileSize          = 50 * 1024 * 1024  // 50MB
	MaxProjectUploadSize = 500 * 1024 * 1024 // 500MB
)

type GenerateEmbeddingMessage struct {
	DocsID string `json:"docs_id"`
}

// registerRoutes binds HTTP routes to the Server
func (s *ChatService) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/upload", s.handleUpload)
	mux.HandleFunc("/documents/", s.handleDownload)
}

func (s *ChatService) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	projectID := r.FormValue("project_id")
	if projectID == "" {
		http.Error(w, "Missing project_id", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File not provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileSize := header.Size
	if fileSize > MaxFileSize {
		http.Error(w, "File exceeds 50MB limit", http.StatusRequestEntityTooLarge)
		return
	}

	totalUsed, err := s.dao.TotalUsedSize(projectID)
	if err != nil {
		http.Error(w, "Failed to fetch usage: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if totalUsed+fileSize > MaxProjectUploadSize {
		http.Error(w, "Project storage exceeds 500MB", http.StatusRequestEntityTooLarge)
		return
	}

	objectID := uuid.New().String()
	if err := s.store.StoreObject(r.Context(), objectID, file); err != nil {
		http.Error(w, "Failed to store file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.dao.FileSave(projectID, objectID, header.Filename, fileSize); err != nil {
		http.Error(w, "Failed to save metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	msg := GenerateEmbeddingMessage{DocsID: objectID}
	msgBytes, _ := json.Marshal(msg)
	err = s.queue.Publish(r.Context(), events.GENERATE_EMBEDDINGS, msgBytes)
	if err != nil {
		err := fmt.Errorf("failed publish %v", err)
		http.Error(w, "Failed to publish event: "+err.Error(), http.StatusInternalServerError)

	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "File uploaded successfully", "id": "%s"}`, objectID)
}

func (s *ChatService) handleDownload(w http.ResponseWriter, r *http.Request) {
	docsId := strings.TrimPrefix(r.URL.Path, "/documents/")
	if docsId == "" {
		http.Error(w, "Missing document ID", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join("filestore", "objects", docsId)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found on disk", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, filePath)
}

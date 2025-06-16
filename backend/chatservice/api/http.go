package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"sortedstartup/chatservice/dao"
	"sortedstartup/chatservice/store"

	"github.com/google/uuid"
)

const (
	MaxFileSize          = 50 * 1024 * 1024  // 50MB
	MaxProjectUploadSize = 500 * 1024 * 1024 // 500MB
)

// HTTPHandler holds dependencies and route registration logic
type HTTPHandler struct {
	db    *dao.SQLiteDAO
	store *store.DiskObjectStore
}

// NewHTTPHandler constructs the handler with dependencies
func NewHTTPHandler(db *dao.SQLiteDAO, store *store.DiskObjectStore) *HTTPHandler {
	return &HTTPHandler{
		db:    db,
		store: store,
	}
}

// RegisterRoutes adds /upload and /documents/ handlers to mux
func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/upload", h.handleUpload)
	mux.HandleFunc("/documents/", h.handleDownload)
}

func (h *HTTPHandler) handleUpload(w http.ResponseWriter, r *http.Request) {
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

	totalUsed, err := h.db.TotalUsedSize(projectID)
	if err != nil {
		http.Error(w, "Failed to fetch usage: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if totalUsed+fileSize > MaxProjectUploadSize {
		http.Error(w, "Project storage exceeds 500MB", http.StatusRequestEntityTooLarge)
		return
	}

	objectID := uuid.New().String()
	if err := h.store.StoreObject(r.Context(), objectID, file); err != nil {
		http.Error(w, "Failed to store file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.db.FileSave(projectID, objectID, header.Filename, fileSize); err != nil {
		http.Error(w, "Failed to save metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "File uploaded successfully", "id": "%s"}`, objectID)
}

func (h *HTTPHandler) handleDownload(w http.ResponseWriter, r *http.Request) {
	docsId := strings.TrimPrefix(r.URL.Path, "/documents/")
	if docsId == "" {
		http.Error(w, "Missing document ID", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join("filestore", "objects", docsId)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("File does not exist at path: %s\n", filePath)
		http.Error(w, "File not found on disk", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, filePath)
}

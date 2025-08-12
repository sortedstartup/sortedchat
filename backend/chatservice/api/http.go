package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	MaxFileSize          = 50 * 1024 * 1024  // 50MB
	MaxProjectUploadSize = 500 * 1024 * 1024 // 500MB
)

// registerRoutes binds HTTP routes to the Server
func (s *ChatServiceAPI) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/upload", s.handleUpload)
	mux.HandleFunc("/documents/", s.handleDownload)
}

func (s *ChatServiceAPI) handleUpload(w http.ResponseWriter, r *http.Request) {
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

	// Use service layer to handle file upload with hardcoded user ID
	objectID, err := s.service.UploadFile(HARDCODED_USER_ID, projectID, file, header, MaxFileSize, MaxProjectUploadSize)
	if err != nil {
		http.Error(w, "Failed to upload file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "File uploaded successfully", "id": "%s"}`, objectID)
}

func (s *ChatServiceAPI) handleDownload(w http.ResponseWriter, r *http.Request) {
	docsId := strings.TrimPrefix(r.URL.Path, "/documents/")
	if docsId == "" {
		http.Error(w, "Missing document ID", http.StatusBadRequest)
		return
	}

	// TODO: This should also validate user access to the document
	// and use service layer for access control

	filePath := filepath.Join("filestore", "objects", docsId)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found on disk", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, filePath)
}

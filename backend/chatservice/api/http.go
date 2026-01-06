package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sortedstartup/common/auth"
	"strings"
)

const (
	MaxFileSize          = 50 * 1024 * 1024  // 50MB
	MaxProjectUploadSize = 500 * 1024 * 1024 // 500MB
)

// registerRoutes binds HTTP routes to the Server
func (s *ChatServiceAPI) registerRoutes(mux *http.ServeMux) {
	slog.Info("api:registerRoutes")
	mux.HandleFunc("/upload", s.handleUpload)
	mux.HandleFunc("/documents/", s.handleDownload)
	mux.HandleFunc("/agents/upload", s.handleAgentUpload)
	mux.HandleFunc("/agents/files/", s.handleAgentFileDownload)
	mux.HandleFunc("/agents/files-list/", s.handleAgentFilesList)
	mux.HandleFunc("/agents/files/delete", s.handleAgentFileDelete)
	mux.HandleFunc("/agents/files/update", s.handleAgentFileUpdate)
}

func (s *ChatServiceAPI) handleUpload(w http.ResponseWriter, r *http.Request) {
	slog.Info("handling upload request", "method", r.Method, "path", r.URL.Path)
	if r.Method != http.MethodPost {
		slog.Error("Method not allowed", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	projectID := r.FormValue("project_id")
	if projectID == "" {
		slog.Error("Missing project_id")
		http.Error(w, "Missing project_id", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		slog.Error("File not provided", "error", err)
		http.Error(w, "File not provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Use service layer to handle file upload with
	userID, err := auth.GetUserIDFromContext_WithError(r.Context())
	if err != nil {
		slog.Error("User ID not found", "error", err)
		http.Error(w, "User ID not found", http.StatusInternalServerError)
		return
	}

	objectID, err := s.service.UploadFile(r.Context(), userID, projectID, file, header, MaxFileSize, MaxProjectUploadSize)
	if err != nil {
		slog.Error("Failed to upload file", "error", err)
		http.Error(w, "Failed to upload file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "File uploaded successfully", "id": "%s"}`, objectID)
}

func (s *ChatServiceAPI) handleDownload(w http.ResponseWriter, r *http.Request) {
	slog.Info("api:handleDownload", "method", r.Method, "path", r.URL.Path)
	docsId := strings.TrimPrefix(r.URL.Path, "/documents/")
	if docsId == "" {
		slog.Error("Missing document ID")
		http.Error(w, "Missing document ID", http.StatusBadRequest)
		return
	}

	// TODO: This should also validate user access to the document
	// and use service layer for access control

	filePath := filepath.Join("filestore", "objects", docsId)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		slog.Error("File not found on disk", "error", err)
		http.Error(w, "File not found on disk", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, filePath)
}

func (s *ChatServiceAPI) handleAgentUpload(w http.ResponseWriter, r *http.Request) {
	slog.Info("api:handleAgentUpload", "method", r.Method, "path", r.URL.Path)
	if r.Method != http.MethodPost {
		slog.Error("Method not allowed", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agentID := r.FormValue("agent_id")
	if agentID == "" {
		slog.Error("Missing agent_id")
		http.Error(w, "Missing agent_id", http.StatusBadRequest)
		return
	}

	// Get file path from form data (sent by FileUploader with folder structure)
	filePath := r.FormValue("file_path")
	if filePath == "" {
		// Fallback to just filename if no path provided
		filePath = r.FormValue("file_name")
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		slog.Error("File not provided", "error", err)
		http.Error(w, "File not provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get user ID from context
	userID, err := auth.GetUserIDFromContext_WithError(r.Context())
	if err != nil {
		slog.Error("User ID not found", "error", err)
		http.Error(w, "User ID not found", http.StatusUnauthorized)
		return
	}

	// Upload file using service layer
	objectID, err := s.service.UploadAgentFile(r.Context(), userID, agentID, file, header, filePath, MaxFileSize)
	if err != nil {
		slog.Error("Failed to upload agent file", "error", err)
		http.Error(w, "Failed to upload file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "File uploaded successfully", "id": "%s", "path": "%s"}`, objectID, filePath)
}

func (s *ChatServiceAPI) handleAgentFileDownload(w http.ResponseWriter, r *http.Request) {
	slog.Info("api:handleAgentFileDownload", "method", r.Method, "path", r.URL.Path)
	docsID := strings.TrimPrefix(r.URL.Path, "/agents/files/")
	if docsID == "" {
		slog.Error("Missing document ID")
		http.Error(w, "Missing document ID", http.StatusBadRequest)
		return
	}

	// TODO: Validate user has access to this agent's files

	// Reuse same filestore location
	filePath := filepath.Join("filestore", "objects", docsID)
	absPath, _ := filepath.Abs(filePath)
	slog.Info("Loading file", "path", filePath, "abs_path", absPath)

	if stat, err := os.Stat(filePath); os.IsNotExist(err) {
		slog.Error("File not found on disk", "error", err, "path", filePath)
		http.Error(w, "File not found on disk", http.StatusNotFound)
		return
	} else {
		slog.Info("File found", "size", stat.Size(), "modified", stat.ModTime())
	}

	// Prevent caching to ensure we always get the latest version
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	http.ServeFile(w, r, filePath)
}

func (s *ChatServiceAPI) handleAgentFilesList(w http.ResponseWriter, r *http.Request) {
	slog.Info("api:handleAgentFilesList", "method", r.Method, "path", r.URL.Path)
	agentID := strings.TrimPrefix(r.URL.Path, "/agents/files-list/")
	if agentID == "" {
		slog.Error("Missing agent ID")
		http.Error(w, "Missing agent ID", http.StatusBadRequest)
		return
	}

	// TODO: Validate user has access to this agent

	files, err := s.service.GetAgentFiles(r.Context(), agentID)
	if err != nil {
		slog.Error("Failed to get agent files", "error", err)
		http.Error(w, "Failed to get agent files: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"files": files,
	})
}

func (s *ChatServiceAPI) handleAgentFileDelete(w http.ResponseWriter, r *http.Request) {
	slog.Info("api:handleAgentFileDelete", "method", r.Method, "path", r.URL.Path)
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AgentID string `json:"agent_id"`
		DocsID  string `json:"docs_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AgentID == "" || req.DocsID == "" {
		http.Error(w, "Missing agent_id or docs_id", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, err := auth.GetUserIDFromContext_WithError(r.Context())
	if err != nil {
		slog.Error("User ID not found", "error", err)
		http.Error(w, "User ID not found", http.StatusUnauthorized)
		return
	}
	_ = userID // Will use for validation later

	// TODO: Validate user has access to this agent

	if err := s.service.DeleteAgentFile(r.Context(), req.AgentID, req.DocsID); err != nil {
		slog.Error("Failed to delete agent file", "error", err)
		http.Error(w, "Failed to delete file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Also delete the physical file from storage
	filePath := filepath.Join("filestore", "objects", req.DocsID)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		slog.Warn("Failed to delete physical file", "error", err, "path", filePath)
		// Don't fail the request if physical deletion fails
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "File deleted successfully",
	})
}

func (s *ChatServiceAPI) handleAgentFileUpdate(w http.ResponseWriter, r *http.Request) {
	slog.Info("api:handleAgentFileUpdate", "method", r.Method, "path", r.URL.Path)
	if r.Method != http.MethodPut {
		slog.Error("Method not allowed", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AgentID string `json:"agent_id"`
		DocsID  string `json:"docs_id"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	slog.Info("api:handleAgentFileUpdate decoded request", "agent_id", req.AgentID, "docs_id", req.DocsID, "content_length", len(req.Content))

	if req.AgentID == "" || req.DocsID == "" {
		slog.Error("Missing required fields", "agent_id", req.AgentID, "docs_id", req.DocsID)
		http.Error(w, "Missing agent_id or docs_id", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, err := auth.GetUserIDFromContext_WithError(r.Context())
	if err != nil {
		slog.Error("User ID not found", "error", err)
		http.Error(w, "User ID not found", http.StatusUnauthorized)
		return
	}
	_ = userID // Will use for validation later

	// TODO: Validate user has access to this agent

	// Write the updated content to the physical file
	filePath := filepath.Join("filestore", "objects", req.DocsID)
	absPath, _ := filepath.Abs(filePath)
	slog.Info("Writing file", "path", filePath, "abs_path", absPath, "size", len(req.Content))

	// Verify file exists before writing
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		slog.Warn("File doesn't exist, will create new", "path", filePath)
	}

	if err := os.WriteFile(filePath, []byte(req.Content), 0644); err != nil {
		slog.Error("Failed to write file", "error", err, "path", filePath, "abs_path", absPath)
		http.Error(w, "Failed to update file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Verify the write was successful
	if stat, err := os.Stat(filePath); err == nil {
		slog.Info("File updated successfully", "path", filePath, "abs_path", absPath, "new_size", stat.Size())
	} else {
		slog.Error("File stat failed after write", "error", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "File updated successfully",
	})
}

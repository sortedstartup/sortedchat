package dao

import (
	pb "sortedstartup/inferenceservice/proto"
)

type ModelMetadata struct {
	ID               string  `db:"id"`
	Name             string  `db:"name"`
	URL              string  `db:"url"`
	Provider         string  `db:"provider"`
	InputTokenCost   float64 `db:"input_token_cost"`
	OutputTokenCost  float64 `db:"output_token_cost"`
	Progress         string  `db:"progress"`
	IsDownloaded     bool    `db:"is_downloaded"`
	IsDownloadable   bool    `db:"is_downloadable"`
	Status           int     `db:"status"`
	FileStoreID      *string `db:"filestore_id"`
	IsEmbeddingModel bool    `db:"is_embedding_model"`
	CachedTokenCost  float64 `db:"cached_token_cost"`
	IsEnabled        bool    `db:"is_enabled"`
	ModelInfo        string  `db:"model_info"`
}

// Status constants - using proto enum values
const (
	StatusNone        = int(pb.DownloadStatus_NONE)        // For non-downloadable models
	StatusPending     = int(pb.DownloadStatus_PENDING)     // Ready to download
	StatusDownloading = int(pb.DownloadStatus_DOWNLOADING) // Currently downloading
	StatusCompleted   = int(pb.DownloadStatus_COMPLETED)   // Download completed
	StatusFailed      = int(pb.DownloadStatus_FAILED)      // Download failed
	StatusCancelling  = int(pb.DownloadStatus_CANCELLING)  // Download cancelled
)

// DownloadProgress represents the progress of a model download
type DownloadProgress struct {
	FileSize int64 `json:"filesize"` // Total file size in bytes
	Status   int   `json:"status"`   // Status constant (0-4)
	Progress int   `json:"progress"` // Progress percentage (0-100)
	Speed    int64 `json:"speed"`    // Download speed in kilobytes per second
}

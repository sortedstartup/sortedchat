package dao

import (
	"encoding/json"
)

// Status constants
const (
	StatusNone        = 0 // For non-downloadable models
	StatusPending     = 1 // Ready to download
	StatusDownloading = 2 // Currently downloading
	StatusCompleted   = 3 // Download completed
	StatusFailed      = 4 // Download failed
)

// DownloadProgress represents the progress of a model download
type DownloadProgress struct {
	FileSize int64 `json:"filesize"` // Total file size in bytes
	Status   int   `json:"status"`   // Status constant (0-4)
	Progress int   `json:"progress"` // Progress percentage (0-100)
	Speed    int64 `json:"speed"`    // Download speed in kilobytes per second
}

// ToJSON converts DownloadProgress to JSON string
func (dp *DownloadProgress) ToJSON() string {
	data, err := json.Marshal(dp)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// FromJSON creates DownloadProgress from JSON string
func FromJSON(jsonStr string) (*DownloadProgress, error) {
	var dp DownloadProgress
	err := json.Unmarshal([]byte(jsonStr), &dp)
	if err != nil {
		return nil, err
	}
	return &dp, nil
}

type DAO interface {
	Infer(dummy string) error
	DownloadModel(userID string, modelName string, url string) error
	GetModelByName(modelName string) (*ModelMetadata, error)
	GetAllModels() ([]*ModelMetadata, error)
	UpdateModelProgress(id string, progress *DownloadProgress) error
}

// ModelMetadata represents a model in the database
type ModelMetadata struct {
	ID              string  `db:"id"`
	Name            string  `db:"name"`
	URL             string  `db:"url"`
	Provider        string  `db:"provider"`
	InputTokenCost  float64 `db:"input_token_cost"`
	OutputTokenCost float64 `db:"output_token_cost"`
	Progress        string  `db:"progress"`
	IsDownloaded    bool    `db:"is_downloaded"`
	IsDownloadable  bool    `db:"is_downloadable"`
	Status          int     `db:"status"`
}

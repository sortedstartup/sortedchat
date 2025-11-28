package dao

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

// ToJSON converts DownloadProgress to JSON string
func (dp *DownloadProgress) ToJSON() (string, error) {
	slog.Info("inferenceservice:dao:ToJSON")
	data, err := json.Marshal(dp)
	if err != nil {
		slog.Error("inferenceservice:dao:ToJSON", "message", "failed to marshal download progress", "error", err)
		return "", fmt.Errorf("failed to marshal download progress")
	}
	return string(data), nil
}

// FromJSON creates DownloadProgress from JSON string
func FromJSON(jsonStr string) (*DownloadProgress, error) {
	slog.Info("inferenceservice:dao:FromJSON")
	var dp DownloadProgress
	err := json.Unmarshal([]byte(jsonStr), &dp)
	if err != nil {
		slog.Error("inferenceservice:dao:FromJSON", "message", "failed to unmarshal download progress", "error", err)
		return nil, fmt.Errorf("failed to unmarshal download progress")
	}
	return &dp, nil
}

type DAO interface {
	Infer(dummy string) error
	GetModelByName(modelName string) (*ModelMetadata, error)
	GetAllModels() ([]*ModelMetadata, error)
	UpdateModelProgress(id string, progress *DownloadProgress) error
	UpdateModelFileStoreID(id string, filestoreID string) error
	ResetModelToInitialState(id string) error
}

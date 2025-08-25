package dao

import (
	"encoding/json"
)

// ToJSON converts DownloadProgress to JSON string
func (dp *DownloadProgress) ToJSON() (string, error) {
	data, err := json.Marshal(dp)
	if err != nil {
		return "", err
	}
	return string(data), nil
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
	GetModelByName(modelName string) (*ModelMetadata, error)
	GetAllModels() ([]*ModelMetadata, error)
	UpdateModelProgress(id string, progress *DownloadProgress) error
	UpdateModelFileStoreID(id string, filestoreID string) error
}

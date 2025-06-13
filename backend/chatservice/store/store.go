package store

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type ObjectStore interface {
	GetObject(ctx context.Context, objectID string) (name string, object io.Reader, err error)
	StoreObject(ctx context.Context, objectID string, object io.Reader) error
}

// DiskObjectStore implements ObjectStore interface by storing objects on disk
type DiskObjectStore struct {
	basePath string
}

// NewDiskObjectStore creates a new disk-based object store
func NewDiskObjectStore(basePath string) (*DiskObjectStore, error) {
	// Create the base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	// Create subdirectory for objects
	objectsDir := filepath.Join(basePath, "objects")

	if err := os.MkdirAll(objectsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create objects directory: %w", err)
	}

	return &DiskObjectStore{basePath: basePath}, nil
}

// StoreObject stores an object on disk with the given object ID
func (d *DiskObjectStore) StoreObject(ctx context.Context, objectID string, object io.Reader) error {
	// Validate objectID
	if objectID == "" {
		return fmt.Errorf("objectID cannot be empty")
	}

	// Create object path
	objectPath := filepath.Join(d.basePath, "objects", objectID)

	// Create and write the object
	outFile, err := os.Create(objectPath)
	if err != nil {
		return fmt.Errorf("failed to create object file: %w", err)
	}
	defer outFile.Close()

	// Copy object content
	if _, err := io.Copy(outFile, object); err != nil {
		// Clean up the file if copy failed
		os.Remove(objectPath)
		return fmt.Errorf("failed to write object content: %w", err)
	}

	return nil
}

// GetObject retrieves an object from disk by object ID
func (d *DiskObjectStore) GetObject(ctx context.Context, objectID string) (string, io.Reader, error) {
	// Validate objectID
	if objectID == "" {
		return "", nil, fmt.Errorf("objectID cannot be empty")
	}

	// Open the object file
	objectPath := filepath.Join(d.basePath, "objects", objectID)
	file, err := os.Open(objectPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil, fmt.Errorf("object not found: %s", objectID)
		}
		return "", nil, fmt.Errorf("failed to open object file: %w", err)
	}

	// Return objectID as name since we don't store original names
	return objectID, file, nil
}

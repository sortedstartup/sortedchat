package dao

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
	FileStoreID     *string `db:"filestore_id"`
}

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

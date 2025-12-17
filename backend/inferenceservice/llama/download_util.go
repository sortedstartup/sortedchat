package llama

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// DownloadStatus represents the current state of the download.
type DownloadStatus string

const (
	StatusDownloading DownloadStatus = "downloading"
	StatusExtracting  DownloadStatus = "extracting"
	StatusCompleted   DownloadStatus = "completed"
	StatusError       DownloadStatus = "error"
	StatusCancelled   DownloadStatus = "cancelled"
)

// DownloadProgress represents the progress of the download and extraction.
type DownloadProgress struct {
	Status          DownloadStatus
	TotalBytes      int64
	DownloadedBytes int64
	Percent         float64
	Error           error
}

// downloadFile downloads a file from the given URL to the destination directory.
// If unzip is true, it extracts the file (supporting zip and tar.gz) and removes the archive.
// It returns a channel that streams progress updates.
func downloadFile(ctx context.Context, url string, destDir string, unzip bool) (<-chan DownloadProgress, error) {
	progressChan := make(chan DownloadProgress, 10)

	go func() {
		defer close(progressChan)

		// Create destination directory
		if err := os.MkdirAll(destDir, 0755); err != nil {
			progressChan <- DownloadProgress{Status: StatusError, Error: fmt.Errorf("failed to create directory: %w", err)}
			return
		}

		// Download
		progressChan <- DownloadProgress{Status: StatusDownloading, TotalBytes: 0, DownloadedBytes: 0, Percent: 0}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			progressChan <- DownloadProgress{Status: StatusError, Error: fmt.Errorf("failed to create request: %w", err)}
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				progressChan <- DownloadProgress{Status: StatusCancelled, Error: ctx.Err()}
			} else {
				progressChan <- DownloadProgress{Status: StatusError, Error: fmt.Errorf("failed to download: %w", err)}
			}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			progressChan <- DownloadProgress{Status: StatusError, Error: fmt.Errorf("bad status: %s", resp.Status)}
			return
		}

		totalBytes := resp.ContentLength

		// Determine output file
		var outFile *os.File
		var outPath string
		if unzip {
			// Create a temporary file for the download
			outFile, err = os.CreateTemp(destDir, "download-*.tmp")
		} else {
			// Use filename from URL
			filename := filepath.Base(url)
			outPath = filepath.Join(destDir, filename)
			outFile, err = os.Create(outPath)
		}

		if err != nil {
			progressChan <- DownloadProgress{Status: StatusError, Error: fmt.Errorf("failed to create output file: %w", err)}
			return
		}

		// If it's a temp file, we want to remove it later if we are extracting,
		// or if there is an error.
		// If it's not a temp file (unzip=false), we keep it unless error.

		// We'll handle cleanup manually based on success/failure

		defer outFile.Close()

		// Copy with progress
		buf := make([]byte, 32*1024)
		var downloadedBytes int64
		for {
			select {
			case <-ctx.Done():
				outFile.Close()
				os.Remove(outFile.Name()) // Clean up partial file
				progressChan <- DownloadProgress{Status: StatusCancelled, Error: ctx.Err()}
				return
			default:
				n, err := resp.Body.Read(buf)
				if n > 0 {
					_, wErr := outFile.Write(buf[:n])
					if wErr != nil {
						progressChan <- DownloadProgress{Status: StatusError, Error: fmt.Errorf("failed to write to file: %w", wErr)}
						return
					}
					downloadedBytes += int64(n)
					if totalBytes > 0 {
						percent := float64(downloadedBytes) / float64(totalBytes) * 100
						progressChan <- DownloadProgress{Status: StatusDownloading, TotalBytes: totalBytes, DownloadedBytes: downloadedBytes, Percent: percent}
					}
				}
				if err == io.EOF {
					goto DownloadComplete
				}
				if err != nil {
					progressChan <- DownloadProgress{Status: StatusError, Error: fmt.Errorf("error downloading: %w", err)}
					return
				}
			}
		}

	DownloadComplete:
		if !unzip {
			progressChan <- DownloadProgress{Status: StatusCompleted, Percent: 100}
			return
		}

		// Extract
		progressChan <- DownloadProgress{Status: StatusExtracting, TotalBytes: totalBytes, DownloadedBytes: downloadedBytes, Percent: 100}

		// Reset file pointer to beginning for reading
		if _, err := outFile.Seek(0, 0); err != nil {
			progressChan <- DownloadProgress{Status: StatusError, Error: fmt.Errorf("failed to seek temp file: %w", err)}
			return
		}

		var extractErr error
		if strings.HasSuffix(url, ".zip") {
			extractErr = unzipFunc(outFile, destDir)
		} else if strings.HasSuffix(url, ".tar.gz") || strings.HasSuffix(url, ".tgz") {
			extractErr = untarFunc(outFile, destDir)
		} else {
			extractErr = fmt.Errorf("unknown archive format: %s", url)
		}

		// Close file before removing
		outFile.Close()
		os.Remove(outFile.Name())

		if extractErr != nil {
			progressChan <- DownloadProgress{Status: StatusError, Error: extractErr}
			return
		}

		progressChan <- DownloadProgress{Status: StatusCompleted, Percent: 100}
	}()

	return progressChan, nil
}

func unzipFunc(f *os.File, destDir string) error {
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	zipReader, err := zip.NewReader(f, stat.Size())
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}

	for _, f := range zipReader.File {
		fpath := filepath.Join(destDir, f.Name)

		// Check for ZipSlip
		if !strings.HasPrefix(fpath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			continue // illegal file path
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		// Handle Symlinks
		if f.FileInfo().Mode()&os.ModeSymlink != 0 {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			linkTarget, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return err
			}

			// Remove existing file/symlink if it exists
			if err := os.Remove(fpath); err != nil && !os.IsNotExist(err) {
				return err
			}

			if err := os.Symlink(string(linkTarget), fpath); err != nil {
				return err
			}
			continue
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func untarFunc(f *os.File, destDir string) error {
	gzr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar reading error: %w", err)
		}

		target := filepath.Join(destDir, header.Name)

		// Check for ZipSlip
		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) {
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			// Remove existing file/symlink if it exists
			if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
				return err
			}
			if err := os.Symlink(header.Linkname, target); err != nil {
				return err
			}
		}
	}
	return nil
}

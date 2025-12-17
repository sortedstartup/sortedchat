package llama

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestUnzipFunc_Symlink(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "unzip_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a zip file with a symlink
	zipPath := filepath.Join(tempDir, "test.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}

	w := zip.NewWriter(zipFile)

	// Add a regular file
	f, err := w.Create("target.txt")
	if err != nil {
		t.Fatalf("failed to create file in zip: %v", err)
	}
	_, err = f.Write([]byte("hello world"))
	if err != nil {
		t.Fatalf("failed to write to file in zip: %v", err)
	}

	// Add a symlink
	// In zip, symlinks are stored as files with specific mode bits and the content is the target path
	header := &zip.FileHeader{
		Name:   "link.txt",
		Method: zip.Deflate,
	}
	header.SetMode(os.ModeSymlink | 0755)
	sl, err := w.CreateHeader(header)
	if err != nil {
		t.Fatalf("failed to create symlink header: %v", err)
	}
	_, err = sl.Write([]byte("target.txt"))
	if err != nil {
		t.Fatalf("failed to write symlink target: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
	zipFile.Close()

	// Re-open the zip file for reading
	zipFileRead, err := os.Open(zipPath)
	if err != nil {
		t.Fatalf("failed to open zip file: %v", err)
	}
	defer zipFileRead.Close()

	// Extract
	extractDir := filepath.Join(tempDir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		t.Fatalf("failed to create extract dir: %v", err)
	}

	if err := unzipFunc(zipFileRead, extractDir); err != nil {
		t.Fatalf("unzipFunc failed: %v", err)
	}

	// Verify symlink
	linkPath := filepath.Join(extractDir, "link.txt")
	info, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("failed to stat extracted link: %v", err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		t.Errorf("extracted file is not a symlink")
	}

	target, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("failed to read link: %v", err)
	}

	if target != "target.txt" {
		t.Errorf("symlink target mismatch: got %s, want target.txt", target)
	}
}

func TestUntarFunc_Symlink(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "untar_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a tar.gz file with a symlink
	tarPath := filepath.Join(tempDir, "test.tar.gz")
	tarFile, err := os.Create(tarPath)
	if err != nil {
		t.Fatalf("failed to create tar file: %v", err)
	}

	gw := gzip.NewWriter(tarFile)
	tw := tar.NewWriter(gw)

	// Add a regular file
	content := []byte("hello world")
	if err := tw.WriteHeader(&tar.Header{
		Name: "target.txt",
		Mode: 0644,
		Size: int64(len(content)),
	}); err != nil {
		t.Fatalf("failed to write tar header: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("failed to write tar content: %v", err)
	}

	// Add a symlink
	if err := tw.WriteHeader(&tar.Header{
		Name:     "link.txt",
		Mode:     0777,
		Typeflag: tar.TypeSymlink,
		Linkname: "target.txt",
	}); err != nil {
		t.Fatalf("failed to write symlink header: %v", err)
	}

	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}
	tarFile.Close()

	// Re-open the tar file for reading
	tarFileRead, err := os.Open(tarPath)
	if err != nil {
		t.Fatalf("failed to open tar file: %v", err)
	}
	defer tarFileRead.Close()

	// Extract
	extractDir := filepath.Join(tempDir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		t.Fatalf("failed to create extract dir: %v", err)
	}

	if err := untarFunc(tarFileRead, extractDir); err != nil {
		t.Fatalf("untarFunc failed: %v", err)
	}

	// Verify symlink
	linkPath := filepath.Join(extractDir, "link.txt")
	info, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("failed to stat extracted link: %v", err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		t.Errorf("extracted file is not a symlink")
	}

	target, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("failed to read link: %v", err)
	}

	if target != "target.txt" {
		t.Errorf("symlink target mismatch: got %s, want target.txt", target)
	}
}

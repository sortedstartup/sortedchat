package llama

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const LLAMASERVER_DIR = "llamaserver"
const LLAMASERVER_VERSION = "b7388"

var releases = map[string]string{
	"linux-amd64-cpu":      "https://github.com/ggml-org/llama.cpp/releases/download/b7388/llama-b7388-bin-ubuntu-x64.tar.gz",
	"linux-amd64-vulkan":   "https://github.com/ggml-org/llama.cpp/releases/download/b7388/llama-b7388-bin-ubuntu-vulkan-x64.tar.gz",
	"darwin-arm64":         "https://github.com/ggml-org/llama.cpp/releases/download/b7388/llama-b7388-bin-macos-arm64.tar.gz",
	"windows-amd64-cpu":    "https://github.com/ggml-org/llama.cpp/releases/download/b7388/llama-b7388-bin-win-cpu-x64.zip",
	"windows-amd64-cuda":   "https://github.com/ggml-org/llama.cpp/releases/download/b7388/llama-b7388-bin-win-cuda-13.1-x64.zip",
	"windows-amd64-vulkan": "https://github.com/ggml-org/llama.cpp/releases/download/b7388/llama-b7388-bin-win-vulkan-x64.zip",
}

/*
TODO
- need to downlaod llama-server/ from github ? or sortedserver and unzip in a configurable location
  - filestore ?
  - configurable in settings ?
- Show embedding models with "embedding label" in the list
- during onboarding, ask user to start download embedding+local model
- RAG uses a hardcoded model "nomic-embed-text"
- llama-proxy is started at a fixed port 8082
   - desktop app: it should start at a unix socket (what on windows ?)
   - even on server binary it can use unix socket, but in microsvce mode it should start at a TCP port ?

- problem: Redirecting llama-server output directly to the parent process's stdout/stderr can:
	Clutter application logs with model server output
	Cause issues if the parent file descriptors are closed or redirected
	Make debugging difficult by mixing output from multiple servers
*/

// Model represents a LLM model with its name and file path.
type Model struct {
	Name string
	Path string
}

// ModelRegistry is a hardcoded map of model names to their file paths.
// modelname -> absolutepath
var ModelRegistry = map[string]string{}

// activeServers keeps track of running llama-server instances.
// Key: Model Name, Value: Unix Socket Path
var activeServers = make(map[string]string)
var serverMutex sync.Mutex

// isModelDownloaded checks if the model exists in the registry and if the file exists on disk.
func isModelDownloaded(name string) (bool, Model) {
	slog.Debug("llama:llama:isModelDownloaded", "model", name)
	slog.Debug("llama:llama:isModelDownloaded", "registry", ModelRegistry)
	path, ok := ModelRegistry[name]
	if !ok {
		slog.Info("Model not found in registry", "model", name)
		return false, Model{}
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		slog.Info("Model not found on disk", "model", name)
		return false, Model{}
	}

	return true, Model{Name: name, Path: path}
}

// GetOrStartServer returns the Unix socket path for a given model, starting a server if necessary.
func GetOrStartServer(modelName string, isEmbeddingModel bool) (string, error) {
	serverMutex.Lock()
	defer serverMutex.Unlock()

	// Check if server is already running
	if socketPath, ok := activeServers[modelName]; ok {
		// Verify if the socket file actually exists (simple health check)
		if _, err := os.Stat(socketPath); err == nil {
			return socketPath, nil
		}
		// If socket missing, assume server died and cleanup
		delete(activeServers, modelName)
	}

	// Check if model is available
	downloaded, model := isModelDownloaded(modelName)
	slog.Debug("llama:llama:GetOrStartServer", "model", modelName, "downloaded", downloaded, "model", model)
	if !downloaded {
		return "", fmt.Errorf("model '%s' not found or not downloaded", modelName)
	}

	// Prepare socket path
	// Using a hash or sanitized name for the socket file would be safer, but for now simple replacement
	// safeName := "llama-" + modelName // In real app, sanitize this string
	// For simplicity, let's just use a temp file pattern or similar.
	// But the requirement says "unix .sock".
	// Let's put sockets in /tmp/
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("llama-%s-%d.sock", "server", time.Now().UnixNano()))

	// Prepare command
	// llama-server -m /path/to/model.gguf --host HOST --unix .sock --port PORT
	// Note: --unix expects the socket path.
	// We'll bind to a random port or just ignore port if unix socket is primary.
	// However, llama-server might require --port. Let's give it 0 or a random unused one if needed,
	// but usually --unix overrides or works alongside.
	// The user prompt said: llama-server -m ... --host HOST - unix .sock --port PORT
	// It seems there might be a typo in user prompt " - unix .sock". I assume it means "--unix <path>".

	var cmd *exec.Cmd
	var err error

	llamaServerPath, err := filepath.Abs(LLAMASERVER_DIR)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for llama-server: %w", err)
	}

	llamaServerPath = path.Join(llamaServerPath, "llama-"+LLAMASERVER_VERSION, "llama-server")
	slog.Info("Starting llama-server at", "path", llamaServerPath)
	if isEmbeddingModel {
		cmd = exec.Command(llamaServerPath,
			"--embeddings",
			"-m", model.Path,
			"--no-webui",
			"--host", socketPath,
			"--port", "0", // Let it pick a port or ignore if unix is used exclusively for our proxy
		)
	} else {
		cmd = exec.Command(llamaServerPath,
			"-m", model.Path,
			"--no-webui",
			"--host", socketPath,
			"--port", "0", // Let it pick a port or ignore if unix is used exclusively for our proxy
		)
	}

	// Detach process or just start it.
	// For a long running server, we should probably not wait for it.
	// But we need to know if it started successfully.
	// A simple way is to start it and wait for the socket to appear.

	// Set SysProcAttr to detach if needed, but for now let's keep it simple as a child process.
	// If we want it to survive proxy restarts, we'd need more complex process management.
	setSysProcAttr(cmd)

	// Redirect output for debugging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start llama-server: %w", err)
	}

	// Wait for socket to be created
	// This is a naive wait. In production, we'd parse logs or use a better health check.
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			// Kill process if it timed out
			_ = cmd.Process.Kill()
			return "", fmt.Errorf("timed out waiting for llama-server to start")
		case <-ticker.C:
			if _, err := os.Stat(socketPath); err == nil {
				// Socket exists!
				activeServers[modelName] = socketPath

				// Optional: Monitor process exit in a goroutine to cleanup map
				go func() {
					_ = cmd.Wait()
					serverMutex.Lock()
					if activeServers[modelName] == socketPath {
						delete(activeServers, modelName)
					}
					serverMutex.Unlock()
				}()

				return socketPath, nil
			}
		}
	}
}

// GetSystemInfo returns the OS and Architecture of the current system.
func GetSystemInfo() (string, string) {
	return runtime.GOOS, runtime.GOARCH
}

// DetectPlatform returns the platform key for the releases map.
// It defaults to "cpu" variant for Linux and Windows if no specific variant is detected/requested.
func DetectPlatform() (string, error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	if osName == "darwin" {
		if arch == "arm64" {
			return "darwin-arm64", nil
		}
		return "", fmt.Errorf("unsupported macos architecture: %s", arch)
	}

	if osName == "linux" {
		if arch == "amd64" {
			// Default to CPU for now
			return "linux-amd64-cpu", nil
		}
		return "", fmt.Errorf("unsupported linux architecture: %s", arch)
	}

	if osName == "windows" {
		if arch == "amd64" {
			// Default to CPU for now
			return "windows-amd64-cpu", nil
		}
		return "", fmt.Errorf("unsupported windows architecture: %s", arch)
	}

	return "", fmt.Errorf("unsupported os: %s", osName)
}

// DownloadLlamaServer downloads and extracts the llama-server binary for the current OS.
// It returns a channel that streams progress updates.
func DownloadLlamaServer(ctx context.Context, destDir string) (<-chan DownloadProgress, error) {
	platform, err := DetectPlatform()
	if err != nil {
		return nil, err
	}

	url, ok := releases[platform]
	if !ok {
		return nil, fmt.Errorf("no release found for platform: %s", platform)
	}

	return downloadFile(ctx, url, destDir, true)
}

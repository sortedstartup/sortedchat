package llama

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

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
func GetOrStartServer(modelName string) (string, error) {
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

	cmd := exec.Command("llama-server",
		"-m", model.Path,
		"--no-webui",
		"--host", socketPath,
		"--port", "0", // Let it pick a port or ignore if unix is used exclusively for our proxy
	)

	// Detach process or just start it.
	// For a long running server, we should probably not wait for it.
	// But we need to know if it started successfully.
	// A simple way is to start it and wait for the socket to appear.

	// Set SysProcAttr to detach if needed, but for now let's keep it simple as a child process.
	// If we want it to survive proxy restarts, we'd need more complex process management.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

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

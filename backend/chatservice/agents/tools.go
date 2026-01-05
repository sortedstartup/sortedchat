package agents

// File system tools for agents - provides sandboxed filesystem access
// Each agent gets isolated access to their own directory, preventing path traversal attacks
//
// Usage Example:
//   fs, err := NewFileSystemTools("agent-id-123", "/path/to/agent/workspace")
//   if err != nil {
//       return err
//   }
//
//   // Get tools for use with Gemini/ADK agent
//   tools, err := fs.GetTools()
//   if err != nil {
//       return err
//   }
//
//   // Create agent with file system tools
//   llmAgent, err := llmagent.New(llmagent.Config{
//       Name:  "MyAgent",
//       Model: model,
//       Tools: tools,  // File system tools ready to use
//   })

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// FileSystemTools defines the interface for sandboxed file system operations
type FileSystemTools interface {
	ReadFile(path string, showLineNumbers bool) (string, error)
	WriteFile(path string, content string) (string, error)
	ListDir(path string) ([]FileInfo, error)
	CreateDir(path string) (string, error)
	FileExists(path string) (bool, error)
	MoveFile(sourcePath string, destPath string) (string, error)
	AppendToFile(path string, content string) (string, error)
	ReadLines(path string, start int, end int) (string, error)
	DeleteLines(path string, start int, end int) (string, error)
	ReplaceLines(path string, start int, end int, newContent string) (string, error)
	SearchRegex(path string, pattern string) ([]RegexMatch, error)
	RegexReplaceAll(path string, pattern string, replacement string) (string, error)
	GetTools() ([]tool.Tool, error)
}

// FileInfo represents file/directory information
type FileInfo struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
}

// RegexMatch represents a regex search match with line number
type RegexMatch struct {
	LineNumber int    `json:"line_number"`
	Line       string `json:"line"`
	Match      string `json:"match"`
}

// sandboxedFileSystem implements FileSystemTools with path sandboxing
type sandboxedFileSystem struct {
	agentID  string
	basePath string
}

const agentsSubPath = "agents"

// NewFileSystemTools creates a new sandboxed file system tools instance
// All operations are restricted to basePath - no path traversal allowed
func NewFileSystemTools(agentID string, basePath string) (FileSystemTools, error) {
	// Clean and make absolute
	absPath, err := filepath.Abs(filepath.Join(basePath, agentsSubPath))
	if err != nil {
		return nil, fmt.Errorf("invalid base path: %w", err)
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &sandboxedFileSystem{
		agentID:  agentID,
		basePath: absPath,
	}, nil
}

// validatePath ensures the path is within basePath (prevents path traversal)
func (fs *sandboxedFileSystem) validatePath(path string) (string, error) {
	// Clean the path
	cleanPath := filepath.Clean(path)

	// Make it absolute relative to basePath
	var fullPath string
	if filepath.IsAbs(cleanPath) {
		fullPath = cleanPath
	} else {
		fullPath = filepath.Join(fs.basePath, cleanPath)
	}

	// Get absolute path and clean it
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Ensure it's within basePath
	if !strings.HasPrefix(absPath, fs.basePath) {
		return "", fmt.Errorf("path outside allowed directory: %s", path)
	}

	return absPath, nil
}

// ReadFile reads a file from the sandboxed directory
func (fs *sandboxedFileSystem) ReadFile(path string, showLineNumbers bool) (string, error) {
	validPath, err := fs.validatePath(path)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(validPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	if !showLineNumbers {
		return string(content), nil
	}

	// Add line numbers
	lines := strings.Split(string(content), "\n")
	var result strings.Builder
	for i, line := range lines {
		result.WriteString(fmt.Sprintf("%4d | %s\n", i+1, line))
	}
	return result.String(), nil
}

// WriteFile writes content to a file in the sandboxed directory
func (fs *sandboxedFileSystem) WriteFile(path string, content string) (string, error) {
	validPath, err := fs.validatePath(path)
	if err != nil {
		return "", err
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(validPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create parent directory: %w", err)
	}

	if err := os.WriteFile(validPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path), nil
}

// ListDir lists files and directories in the sandboxed directory
func (fs *sandboxedFileSystem) ListDir(path string) ([]FileInfo, error) {
	validPath, err := fs.validatePath(path)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(validPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, FileInfo{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
			Size:  info.Size(),
		})
	}

	return files, nil
}

// CreateDir creates a directory in the sandboxed directory
func (fs *sandboxedFileSystem) CreateDir(path string) (string, error) {
	validPath, err := fs.validatePath(path)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(validPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	return fmt.Sprintf("Successfully created directory %s", path), nil
}

// FileExists checks if a file exists in the sandboxed directory
func (fs *sandboxedFileSystem) FileExists(path string) (bool, error) {
	validPath, err := fs.validatePath(path)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(validPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// MoveFile moves/renames a file within the sandboxed directory
func (fs *sandboxedFileSystem) MoveFile(sourcePath string, destPath string) (string, error) {
	validSrc, err := fs.validatePath(sourcePath)
	if err != nil {
		return "", fmt.Errorf("invalid source path: %w", err)
	}

	validDest, err := fs.validatePath(destPath)
	if err != nil {
		return "", fmt.Errorf("invalid destination path: %w", err)
	}

	// Ensure destination parent directory exists
	if err := os.MkdirAll(filepath.Dir(validDest), 0755); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	if err := os.Rename(validSrc, validDest); err != nil {
		return "", fmt.Errorf("failed to move file: %w", err)
	}

	return fmt.Sprintf("Successfully moved %s to %s", sourcePath, destPath), nil
}

// AppendToFile appends content to a file in the sandboxed directory
func (fs *sandboxedFileSystem) AppendToFile(path string, content string) (string, error) {
	validPath, err := fs.validatePath(path)
	if err != nil {
		return "", err
	}

	// Open file for appending, create if doesn't exist
	file, err := os.OpenFile(validPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return "", fmt.Errorf("failed to append to file: %w", err)
	}

	return fmt.Sprintf("Successfully appended %d bytes to %s", len(content), path), nil
}

// ReadLines reads specific lines from a file (1-indexed, inclusive)
func (fs *sandboxedFileSystem) ReadLines(path string, start int, end int) (string, error) {
	validPath, err := fs.validatePath(path)
	if err != nil {
		return "", err
	}

	file, err := os.Open(validPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var result strings.Builder
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		if lineNum >= start && lineNum <= end {
			result.WriteString(fmt.Sprintf("%4d | %s\n", lineNum, scanner.Text()))
		}
		if lineNum > end {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return result.String(), nil
}

// DeleteLines deletes specific lines from a file (1-indexed, inclusive)
func (fs *sandboxedFileSystem) DeleteLines(path string, start int, end int) (string, error) {
	validPath, err := fs.validatePath(path)
	if err != nil {
		return "", err
	}

	// Read all lines
	content, err := os.ReadFile(validPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// Validate range
	if start < 1 || end > len(lines) || start > end {
		return "", fmt.Errorf("invalid line range: %d-%d (file has %d lines)", start, end, len(lines))
	}

	// Build new content without deleted lines (convert to 0-indexed)
	var newLines []string
	newLines = append(newLines, lines[:start-1]...)
	newLines = append(newLines, lines[end:]...)

	// Write back
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(validPath, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	deletedCount := end - start + 1
	return fmt.Sprintf("Successfully deleted %d lines (%d-%d) from %s", deletedCount, start, end, path), nil
}

// ReplaceLines replaces specific lines in a file with new content (1-indexed, inclusive)
func (fs *sandboxedFileSystem) ReplaceLines(path string, start int, end int, newContent string) (string, error) {
	validPath, err := fs.validatePath(path)
	if err != nil {
		return "", err
	}

	// Read all lines
	content, err := os.ReadFile(validPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// Validate range
	if start < 1 || end > len(lines) || start > end {
		return "", fmt.Errorf("invalid line range: %d-%d (file has %d lines)", start, end, len(lines))
	}

	// Build new content with replaced lines (convert to 0-indexed)
	var newLines []string
	newLines = append(newLines, lines[:start-1]...)
	newLines = append(newLines, newContent)
	newLines = append(newLines, lines[end:]...)

	// Write back
	finalContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(validPath, []byte(finalContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	replacedCount := end - start + 1
	return fmt.Sprintf("Successfully replaced %d lines (%d-%d) in %s", replacedCount, start, end, path), nil
}

// SearchRegex searches for a regex pattern in a file and returns matches with line numbers
func (fs *sandboxedFileSystem) SearchRegex(path string, pattern string) ([]RegexMatch, error) {
	validPath, err := fs.validatePath(path)
	if err != nil {
		return nil, err
	}

	// Compile regex
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Read file
	file, err := os.Open(validPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var matches []RegexMatch
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if re.MatchString(line) {
			match := re.FindString(line)
			matches = append(matches, RegexMatch{
				LineNumber: lineNum,
				Line:       line,
				Match:      match,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return matches, nil
}

// RegexReplaceAll replaces all regex matches in a file with replacement text
func (fs *sandboxedFileSystem) RegexReplaceAll(path string, pattern string, replacement string) (string, error) {
	validPath, err := fs.validatePath(path)
	if err != nil {
		return "", err
	}

	// Compile regex
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Read file
	content, err := os.ReadFile(validPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Replace all matches
	originalContent := string(content)
	newContent := re.ReplaceAllString(originalContent, replacement)

	// Count replacements
	matchCount := len(re.FindAllString(originalContent, -1))

	// Write back
	if err := os.WriteFile(validPath, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully replaced %d matches in %s", matchCount, path), nil
}

// Tool argument types for Gemini function calling

type ReadFileArgs struct {
	Path            string `json:"path" jsonschema:"Path to the file to read (relative to agent workspace)"`
	ShowLineNumbers bool   `json:"show_line_numbers" jsonschema:"Whether to show line numbers in the output (default: false)"`
}

type WriteFileArgs struct {
	Path    string `json:"path" jsonschema:"Path to the file to write (relative to agent workspace)"`
	Content string `json:"content" jsonschema:"Content to write to the file"`
}

type ListDirArgs struct {
	Path string `json:"path" jsonschema:"Path to the directory to list (relative to agent workspace, use '.' for root)"`
}

type CreateDirArgs struct {
	Path string `json:"path" jsonschema:"Path to the directory to create (relative to agent workspace)"`
}

type FileExistsArgs struct {
	Path string `json:"path" jsonschema:"Path to check for existence (relative to agent workspace)"`
}

type MoveFileArgs struct {
	SourcePath string `json:"source_path" jsonschema:"Source file path (relative to agent workspace)"`
	DestPath   string `json:"dest_path" jsonschema:"Destination file path (relative to agent workspace)"`
}

type AppendToFileArgs struct {
	Path    string `json:"path" jsonschema:"Path to the file to append to (relative to agent workspace)"`
	Content string `json:"content" jsonschema:"Content to append to the file"`
}

type ReadLinesArgs struct {
	Path  string `json:"path" jsonschema:"Path to the file to read (relative to agent workspace)"`
	Start int    `json:"start" jsonschema:"Starting line number (1-indexed, inclusive)"`
	End   int    `json:"end" jsonschema:"Ending line number (1-indexed, inclusive)"`
}

type DeleteLinesArgs struct {
	Path  string `json:"path" jsonschema:"Path to the file to modify (relative to agent workspace)"`
	Start int    `json:"start" jsonschema:"Starting line number to delete (1-indexed, inclusive)"`
	End   int    `json:"end" jsonschema:"Ending line number to delete (1-indexed, inclusive)"`
}

type ReplaceLinesArgs struct {
	Path       string `json:"path" jsonschema:"Path to the file to modify (relative to agent workspace)"`
	Start      int    `json:"start" jsonschema:"Starting line number to replace (1-indexed, inclusive)"`
	End        int    `json:"end" jsonschema:"Ending line number to replace (1-indexed, inclusive)"`
	NewContent string `json:"new_content" jsonschema:"New content to replace the lines with"`
}

type SearchRegexArgs struct {
	Path    string `json:"path" jsonschema:"Path to the file to search (relative to agent workspace)"`
	Pattern string `json:"pattern" jsonschema:"Regular expression pattern to search for"`
}

type RegexReplaceAllArgs struct {
	Path        string `json:"path" jsonschema:"Path to the file to modify (relative to agent workspace)"`
	Pattern     string `json:"pattern" jsonschema:"Regular expression pattern to match"`
	Replacement string `json:"replacement" jsonschema:"Text to replace matches with"`
}

// GetTools returns all file system tools as a slice of tool.Tool for use with Gemini
func (fs *sandboxedFileSystem) GetTools() ([]tool.Tool, error) {
	var tools []tool.Tool

	// Read File Tool
	readFileTool, err := functiontool.New(functiontool.Config{
		Name:        "read_file",
		Description: "Reads the contents of a file from the agent's workspace. Optionally show line numbers. Returns the file content as a string.",
	}, func(ctx tool.Context, args *ReadFileArgs) (string, error) {
		return fs.ReadFile(args.Path, args.ShowLineNumbers)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create read_file tool: %w", err)
	}
	tools = append(tools, readFileTool)

	// Write File Tool
	writeFileTool, err := functiontool.New(functiontool.Config{
		Name:        "write_file",
		Description: "Writes content to a file in the agent's workspace. Creates parent directories if needed. Returns success message.",
	}, func(ctx tool.Context, args *WriteFileArgs) (string, error) {
		return fs.WriteFile(args.Path, args.Content)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create write_file tool: %w", err)
	}
	tools = append(tools, writeFileTool)

	// List Directory Tool
	listDirTool, err := functiontool.New(functiontool.Config{
		Name:        "list_dir",
		Description: "Lists files and directories in the specified path within the agent's workspace. Returns array of file information including name, type (file/dir), and size.",
	}, func(ctx tool.Context, args *ListDirArgs) ([]FileInfo, error) {
		return fs.ListDir(args.Path)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create list_dir tool: %w", err)
	}
	tools = append(tools, listDirTool)

	// Create Directory Tool
	createDirTool, err := functiontool.New(functiontool.Config{
		Name:        "create_dir",
		Description: "Creates a directory in the agent's workspace. Creates parent directories if needed. Returns success message.",
	}, func(ctx tool.Context, args *CreateDirArgs) (string, error) {
		return fs.CreateDir(args.Path)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create create_dir tool: %w", err)
	}
	tools = append(tools, createDirTool)

	// File Exists Tool
	fileExistsTool, err := functiontool.New(functiontool.Config{
		Name:        "file_exists",
		Description: "Checks if a file or directory exists in the agent's workspace. Returns true if exists, false otherwise.",
	}, func(ctx tool.Context, args *FileExistsArgs) (bool, error) {
		return fs.FileExists(args.Path)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create file_exists tool: %w", err)
	}
	tools = append(tools, fileExistsTool)

	// Move File Tool
	moveFileTool, err := functiontool.New(functiontool.Config{
		Name:        "move_file",
		Description: "Moves or renames a file within the agent's workspace. Can move files between directories. Returns success message.",
	}, func(ctx tool.Context, args *MoveFileArgs) (string, error) {
		return fs.MoveFile(args.SourcePath, args.DestPath)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create move_file tool: %w", err)
	}
	tools = append(tools, moveFileTool)

	// Append To File Tool
	appendToFileTool, err := functiontool.New(functiontool.Config{
		Name:        "append_to_file",
		Description: "Appends content to the end of a file in the agent's workspace. Creates the file if it doesn't exist. Returns success message.",
	}, func(ctx tool.Context, args *AppendToFileArgs) (string, error) {
		return fs.AppendToFile(args.Path, args.Content)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create append_to_file tool: %w", err)
	}
	tools = append(tools, appendToFileTool)

	// Read Lines Tool
	readLinesTool, err := functiontool.New(functiontool.Config{
		Name:        "read_lines",
		Description: "Reads specific lines from a file (1-indexed, inclusive). Useful for reading large files partially. Returns lines with line numbers.",
	}, func(ctx tool.Context, args *ReadLinesArgs) (string, error) {
		return fs.ReadLines(args.Path, args.Start, args.End)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create read_lines tool: %w", err)
	}
	tools = append(tools, readLinesTool)

	// Delete Lines Tool
	deleteLinesTool, err := functiontool.New(functiontool.Config{
		Name:        "delete_lines",
		Description: "Deletes specific lines from a file (1-indexed, inclusive). Rewrites the file without the deleted lines. Returns success message.",
	}, func(ctx tool.Context, args *DeleteLinesArgs) (string, error) {
		return fs.DeleteLines(args.Path, args.Start, args.End)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create delete_lines tool: %w", err)
	}
	tools = append(tools, deleteLinesTool)

	// Replace Lines Tool
	replaceLinesTool, err := functiontool.New(functiontool.Config{
		Name:        "replace_lines",
		Description: "Replaces specific lines in a file with new content (1-indexed, inclusive). Useful for precise file editing. Returns success message.",
	}, func(ctx tool.Context, args *ReplaceLinesArgs) (string, error) {
		return fs.ReplaceLines(args.Path, args.Start, args.End, args.NewContent)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create replace_lines tool: %w", err)
	}
	tools = append(tools, replaceLinesTool)

	// Search Regex Tool
	searchRegexTool, err := functiontool.New(functiontool.Config{
		Name:        "search_regex",
		Description: "Searches for a regex pattern in a file. Returns array of matches with line numbers, line content, and matched text.",
	}, func(ctx tool.Context, args *SearchRegexArgs) ([]RegexMatch, error) {
		return fs.SearchRegex(args.Path, args.Pattern)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create search_regex tool: %w", err)
	}
	tools = append(tools, searchRegexTool)

	// Regex Replace All Tool
	regexReplaceAllTool, err := functiontool.New(functiontool.Config{
		Name:        "regex_replace_all",
		Description: "Replaces all occurrences of a regex pattern in a file with replacement text. Supports capture groups. Returns success message with count.",
	}, func(ctx tool.Context, args *RegexReplaceAllArgs) (string, error) {
		return fs.RegexReplaceAll(args.Path, args.Pattern, args.Replacement)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create regex_replace_all tool: %w", err)
	}
	tools = append(tools, regexReplaceAllTool)

	return tools, nil
}

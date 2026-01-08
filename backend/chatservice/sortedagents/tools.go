package sortedagents

import (
	"context"
	"fmt"
)

// WeatherTool is a simple example tool that provides weather information
type WeatherTool struct{}

// NewWeatherTool creates a new WeatherTool instance
func NewWeatherTool() *WeatherTool {
	return &WeatherTool{}
}

// Name returns the tool's name
func (t *WeatherTool) Name() string {
	return "get_weather"
}

// Description returns the tool's description
func (t *WeatherTool) Description() string {
	return "Get current weather information for a location"
}

func BoolPtr(b bool) *bool {
	return &b
}

// Parameters returns the tool's parameter schema
func (t *WeatherTool) Parameters() *JSONSchema {
	return &JSONSchema{
		Type: "object",
		Properties: map[string]JSONSchema{
			"location": {
				Type:        "string",
				Description: "The city and state, e.g. San Francisco, CA",
			},
		},
		Required:             []string{"location"},
		AdditionalProperties: BoolPtr(false),
	}
}

// Execute runs the weather tool with the given arguments
func (t *WeatherTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	location, ok := args["location"].(string)
	if !ok {
		return nil, fmt.Errorf("location parameter is required and must be a string")
	}

	// This is a placeholder implementation
	// In a real implementation, you would call a weather API
	result := map[string]interface{}{
		"location":    location,
		"temperature": "22°C",
		"condition":   "Sunny",
		"humidity":    "65%",
	}

	return result, nil
}

// ListFilesTool lists files in a directory
type ListFilesTool struct{}

func NewListFilesTool() *ListFilesTool {
	return &ListFilesTool{}
}

func (t *ListFilesTool) Name() string {
	return "list_files"
}

func (t *ListFilesTool) Description() string {
	return "List files in a directory with optional filter"
}

func (t *ListFilesTool) Parameters() *JSONSchema {
	return &JSONSchema{
		Type: "object",
		Properties: map[string]JSONSchema{
			"directory": {
				Type:        "string",
				Description: "The directory to list files from",
			},
			"filter": {
				Type:        "string",
				Description: "Optional filter to apply to the file list",
			},
		},
		Required:             []string{"directory", "filter"},
		AdditionalProperties: BoolPtr(false),
	}
}

func (t *ListFilesTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	directory, ok := args["directory"].(string)
	if !ok || directory == "" {
		directory = "/Documents" // default directory
	}
	filter, _ := args["filter"].(string)

	// Fake file list
	files := []string{"document.pdf", "image.jpg", "script.py", "data.csv", "notes.txt"}
	if filter != "" {
		files = []string{"document.pdf", "notes.txt"} // fake filtered results
	}

	return map[string]interface{}{
		"directory": directory,
		"files":     files,
		"count":     len(files),
	}, nil
}

// AnalyzeFileTypeTool analyzes file type
type AnalyzeFileTypeTool struct{}

func NewAnalyzeFileTypeTool() *AnalyzeFileTypeTool {
	return &AnalyzeFileTypeTool{}
}

func (t *AnalyzeFileTypeTool) Name() string {
	return "analyze_file_type"
}

func (t *AnalyzeFileTypeTool) Description() string {
	return "Analyze and determine the type of a file"
}

func (t *AnalyzeFileTypeTool) Parameters() *JSONSchema {
	return &JSONSchema{
		Type: "object",
		Properties: map[string]JSONSchema{
			"file_path": {
				Type:        "string",
				Description: "The path to the file to analyze",
			},
		},
		Required:             []string{"file_path"},
		AdditionalProperties: BoolPtr(false),
	}
}

func (t *AnalyzeFileTypeTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("file_path parameter is required")
	}

	// Simple fake file type detection based on extension
	var fileType string
	if len(filePath) > 4 {
		ext := filePath[len(filePath)-4:]
		switch ext {
		case ".pdf":
			fileType = "document"
		case ".jpg", ".png":
			fileType = "image"
		case ".txt":
			fileType = "text"
		case ".csv":
			fileType = "data"
		default:
			fileType = "unknown"
		}
	}

	return map[string]interface{}{
		"file_path": filePath,
		"type":      fileType,
		"category":  fileType,
	}, nil
}

// MoveFileTool moves files
type MoveFileTool struct{}

func NewMoveFileTool() *MoveFileTool {
	return &MoveFileTool{}
}

func (t *MoveFileTool) Name() string {
	return "move_file"
}

func (t *MoveFileTool) Description() string {
	return "Move a file from source to destination"
}

func (t *MoveFileTool) Parameters() *JSONSchema {
	return &JSONSchema{
		Type: "object",
		Properties: map[string]JSONSchema{
			"src": {
				Type:        "string",
				Description: "The source file path",
			},
			"dest": {
				Type:        "string",
				Description: "The destination file path",
			},
		},
		Required:             []string{"src", "dest"},
		AdditionalProperties: BoolPtr(false),
	}
}

func (t *MoveFileTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	src, ok := args["src"].(string)
	if !ok || src == "" {
		return nil, fmt.Errorf("src parameter is required")
	}
	dest, ok := args["dest"].(string)
	if !ok || dest == "" {
		return nil, fmt.Errorf("dest parameter is required")
	}

	return map[string]interface{}{
		"success": true,
		"src":     src,
		"dest":    dest,
		"message": fmt.Sprintf("Moved %s to %s", src, dest),
	}, nil
}

// CreateFolderTool creates folders
type CreateFolderTool struct{}

func NewCreateFolderTool() *CreateFolderTool {
	return &CreateFolderTool{}
}

func (t *CreateFolderTool) Name() string {
	return "create_folder"
}

func (t *CreateFolderTool) Description() string {
	return "Create a new folder at the specified path"
}

func (t *CreateFolderTool) Parameters() *JSONSchema {
	return &JSONSchema{
		Type: "object",
		Properties: map[string]JSONSchema{
			"location": {
				Type:        "string",
				Description: "The path where the folder should be created",
			},
		},
		Required:             []string{"location"},
		AdditionalProperties: BoolPtr(false),
	}
}

func (t *CreateFolderTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	location, ok := args["location"].(string)
	if !ok || location == "" {
		return nil, fmt.Errorf("location parameter is required")
	}

	return map[string]interface{}{
		"success":  true,
		"location": location,
		"message":  fmt.Sprintf("Created folder at %s", location),
	}, nil
}
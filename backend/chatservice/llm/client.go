package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"

	"sortedstartup/chatservice/dao"
	pb "sortedstartup/chatservice/proto"
	"sortedstartup/chatservice/settings"
)

type Client struct {
	settingsManager *settings.SettingsManager
	httpClient      *http.Client
}

func NewClient(settingsManager *settings.SettingsManager) *Client {
	return &Client{
		settingsManager: settingsManager,
		httpClient:      &http.Client{},
	}
}

func (c *Client) Call(ctx context.Context, model string, history []dao.ChatMessageRow, req *pb.ChatRequest, userMessage string, enhancedPrompt string) (*http.Response, error) {
	jsonData, err := c.generateRequestBody(model, history, req, userMessage, enhancedPrompt)
	if err != nil {
		slog.Error("llm:Call", "error", err)
		return nil, fmt.Errorf("failed to generate request body: %v", err)
	}

	var url string
	var apiKey string

	if strings.HasPrefix(model, "gemini") {
		url = "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions"
		apiKey = c.settingsManager.GetSettings().GeminiAPIKey
	} else if strings.HasPrefix(model, "claude") {
		url = "https://api.anthropic.com/v1/chat/completions"
		apiKey = c.settingsManager.GetSettings().ClaudeAPIKey
	} else {
		url = "https://api.openai.com/v1/chat/completions"
		apiKey = c.settingsManager.GetSettings().OpenAIAPIKey
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		slog.Error("llm:Call", "error", err)
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	return c.httpClient.Do(httpReq)
}

func (c *Client) generateRequestBody(model string, history []dao.ChatMessageRow, req *pb.ChatRequest, userMessage string, enhancedPrompt string) ([]byte, error) {
	// Convert history to CustomChatRequest format
	var customMessages []Message
	for _, msg := range history {
		var content interface{}

		// Let's simplify the reconstruction logic to match the new structs
		var parts []ContentPart

		// 1. Text Content (from content column)
		if strings.HasPrefix(msg.Content, "[") && strings.HasSuffix(msg.Content, "]") {
			var textContents []*pb.MessageContent
			if err := json.Unmarshal([]byte(msg.Content), &textContents); err == nil {
				for _, tc := range textContents {
					part := ContentPart{Type: tc.Type, Text: tc.Text}
					if tc.Type == "image_url" && tc.ImageUrl != nil {
						part.ImageURL = &ImageURL{URL: tc.ImageUrl.Url}
					}
					parts = append(parts, part)
				}
			}
		} else if msg.Content != "" {
			parts = append(parts, ContentPart{Type: "text", Text: msg.Content})
		}

		// 2. Image Content (from content_image column)
		if msg.ContentImage != "" {
			var imageContents []*pb.MessageContent
			if err := json.Unmarshal([]byte(msg.ContentImage), &imageContents); err == nil {
				for _, ic := range imageContents {
					part := ContentPart{Type: ic.Type}
					if ic.ImageUrl != nil {
						part.ImageURL = &ImageURL{URL: ic.ImageUrl.Url}
					}
					parts = append(parts, part)
				}
			}
		}

		// Determine final Content format
		if len(parts) == 1 && parts[0].Type == "text" {
			content = parts[0].Text
		} else if len(parts) > 0 {
			content = parts
		} else {
			continue // Skip empty messages
		}

		customMessages = append(customMessages, Message{
			Role:    msg.Role,
			Content: content,
		})
	}

	// Add current user message
	var currentMessageContent interface{}
	if enhancedPrompt != "" && len(req.GetContents()) == 0 {
		// Text-only with RAG
		currentMessageContent = enhancedPrompt
	} else if enhancedPrompt != "" && len(req.GetContents()) > 0 {
		// Multi-modal with RAG
		var parts []ContentPart
		parts = append(parts, ContentPart{Type: "text", Text: enhancedPrompt})
		for _, content := range req.GetContents() {
			if content.Type == "image_url" {
				part := ContentPart{Type: "image_url"}
				if content.ImageUrl != nil {
					part.ImageURL = &ImageURL{URL: content.ImageUrl.Url}
				}
				parts = append(parts, part)
			}
		}
		currentMessageContent = parts
	} else if len(req.GetContents()) > 0 {
		// Multi-modal without RAG
		var parts []ContentPart
		for _, content := range req.GetContents() {
			part := ContentPart{Type: content.Type, Text: content.Text}
			if content.Type == "image_url" && content.ImageUrl != nil {
				part.ImageURL = &ImageURL{URL: content.ImageUrl.Url}
			}
			parts = append(parts, part)
		}
		currentMessageContent = parts
	} else {
		// Plain text
		currentMessageContent = userMessage
	}

	customMessages = append(customMessages, Message{
		Role:    "user",
		Content: currentMessageContent,
	})

	// Create CustomChatRequest
	customReq := CustomChatRequest{
		ModelName: model,
		Messages:  customMessages,
		Stream:    true,
		StreamOptions: &StreamOptions{
			IncludeUsage: true,
		},
	}

	templateFile := "chatservice/llm/templates/openai.txt"
	if strings.HasPrefix(model, "gemini") {
		templateFile = "chatservice/llm/templates/gemini.txt"
	}
	if strings.HasPrefix(model, "claude") {
		templateFile = "chatservice/llm/templates/claude.txt"
	}

	// Parse and execute template
	funcMap := template.FuncMap{
		"toJson": func(v interface{}) (string, error) {
			b, err := json.Marshal(v)
			return string(b), err
		},
	}

	tmpl, err := template.New(filepath.Base(templateFile)).Funcs(funcMap).ParseFiles(templateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %v", err)
	}

	var bodyBuffer bytes.Buffer
	if err := tmpl.Execute(&bodyBuffer, customReq); err != nil {
		return nil, fmt.Errorf("failed to execute template: %v", err)
	}

	return bodyBuffer.Bytes(), nil
}

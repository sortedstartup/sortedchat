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
	"time"

	"sortedstartup/chatservice/settings"
	"sortedstartup/chatservice/types"
)

type Client struct {
	settingsManager *settings.SettingsManager
	httpClient      *http.Client
}

func NewClient(settingsManager *settings.SettingsManager) *Client {
	return &Client{
		settingsManager: settingsManager,
		httpClient:      &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) Call(ctx context.Context, req types.ChatCompletionRequest) (*http.Response, error) {
	jsonData, err := c.generateRequestBody(req)
	if err != nil {
		slog.Error("llm:Call", "error", err)
		return nil, fmt.Errorf("failed to generate request body: %v", err)
	}

	var url string
	var apiKey string

	model := req.Model

	if strings.HasPrefix(model, "gemini") {
		url = c.settingsManager.GetSettings().GeminiAPIUrl
		apiKey = c.settingsManager.GetSettings().GeminiAPIKey
	} else if strings.HasPrefix(model, "claude") {
		url = c.settingsManager.GetSettings().ClaudeAPIUrl
		apiKey = c.settingsManager.GetSettings().ClaudeAPIKey
	} else {
		url = c.settingsManager.GetSettings().OpenaiAPIUrl
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

func (c *Client) generateRequestBody(req types.ChatCompletionRequest) ([]byte, error) {
	// Convert ChatCompletionRequest to CustomChatRequest
	// The structure is very similar, mainly mapping Model -> ModelName
	customReq := types.CustomChatRequest{
		ModelName:     req.Model,
		Messages:      req.Messages,
		Stream:        req.Stream,
		StreamOptions: req.StreamOptions,
	}

	templateFile := "chatservice/llm/templates/openai.txt"
	if strings.HasPrefix(req.Model, "gemini") {
		templateFile = "chatservice/llm/templates/gemini.txt"
	}
	if strings.HasPrefix(req.Model, "claude") {
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

package llm

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"text/template"
	"time"

	"sortedstartup/chatservice/settings"
	"sortedstartup/chatservice/types"
)

//go:embed templates/*.txt
var templateFS embed.FS

type Client struct {
	settingsManager *settings.SettingsManager
	httpClient      *http.Client
	templates       map[string]*template.Template
}

const LOCAL_PROVIDER = "local"
const OPENAI_PROVIDER = "openai"
const GEMINI_PROVIDER = "gemini"
const CLAUDE_PROVIDER = "claude"

func NewClient(settingsManager *settings.SettingsManager) *Client {
	// Parse templates once during initialization
	funcMap := template.FuncMap{
		"toJson": func(v interface{}) (string, error) {
			b, err := json.Marshal(v)
			return string(b), err
		},
	}

	templates := make(map[string]*template.Template)
	templateFiles := map[string]string{
		"openai": "templates/openai.txt",
		"gemini": "templates/gemini.txt",
		"claude": "templates/claude.txt",
	}

	for provider, templateFile := range templateFiles {
		content, err := templateFS.ReadFile(templateFile)
		if err != nil {
			slog.Error("llm:NewClient", "provider", provider, "error", err)
			continue
		}
		tmpl, err := template.New(provider).Funcs(funcMap).Parse(string(content))
		if err != nil {
			slog.Error("llm:NewClient", "provider", provider, "error", err)
			continue
		}
		templates[provider] = tmpl
	}

	return &Client{
		settingsManager: settingsManager,
		httpClient:      &http.Client{Timeout: 60 * time.Second},
		templates:       templates,
	}
}

func (c *Client) Call(ctx context.Context, req types.ChatCompletionRequest, provider string) (*http.Response, error) {
	jsonData, err := c.generateRequestBody(req)
	if err != nil {
		slog.Error("llm:Call", "error", err)
		return nil, fmt.Errorf("failed to generate request body: %v", err)
	}

	var url string
	var apiKey string

	model := req.Model
	slog.Info("llm:Call", "model", model)

	switch provider {
	case GEMINI_PROVIDER:
		url = c.settingsManager.GetSettings().GeminiAPIUrl
		apiKey = c.settingsManager.GetSettings().GeminiAPIKey
	case CLAUDE_PROVIDER:
		url = c.settingsManager.GetSettings().ClaudeAPIUrl
		apiKey = c.settingsManager.GetSettings().ClaudeAPIKey
	case LOCAL_PROVIDER:
		//TODO: should not be hard coded
		url = "http://localhost:8081/v1/chat/completions"
		apiKey = "x"
	default:
		url = c.settingsManager.GetSettings().OpenaiAPIUrl
		apiKey = c.settingsManager.GetSettings().OpenAIAPIKey
	}

	// if provider == GEMINI_PROVIDER {
	// 	url = c.settingsManager.GetSettings().GeminiAPIUrl
	// 	apiKey = c.settingsManager.GetSettings().GeminiAPIKey
	// } else if provider == CLAUDE_PROVIDER {
	// 	url = c.settingsManager.GetSettings().ClaudeAPIUrl
	// 	apiKey = c.settingsManager.GetSettings().ClaudeAPIKey
	// } else if provider == LOCAL_PROVIDER {
	// 	//TODO: should not be hard coded
	// 	url = "http://localhost:8081/v1/chat/completions"
	// 	apiKey = "x"
	// } else {
	// 	url = c.settingsManager.GetSettings().OpenaiAPIUrl
	// 	apiKey = c.settingsManager.GetSettings().OpenAIAPIKey
	// }
	// if apiKey == "" {
	// 	return nil, fmt.Errorf("API key not configured for model: %s", model)
	// }

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

	// Determine which template to use based on model prefix
	var provider string
	if strings.HasPrefix(req.Model, "gemini") {
		provider = "gemini"
	} else if strings.HasPrefix(req.Model, "claude") {
		provider = "claude"
	} else {
		provider = "openai"
	}

	// Get the cached template
	tmpl, ok := c.templates[provider]
	if !ok {
		return nil, fmt.Errorf("template not found for provider: %s", provider)
	}

	var bodyBuffer bytes.Buffer
	if err := tmpl.Execute(&bodyBuffer, customReq); err != nil {
		return nil, fmt.Errorf("failed to execute template: %v", err)
	}

	return bodyBuffer.Bytes(), nil
}

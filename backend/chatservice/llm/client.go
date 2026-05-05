package llm

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"text/template"
	"time"

	constants "sortedstartup/chatservice/constants"
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
	jsonData, err := c.generateRequestBody(req, provider)
	if err != nil {
		slog.Error("llm:Call", "error", err)
		return nil, fmt.Errorf("failed to generate request body: %v", err)
	}

	var url string
	var apiKey string

	providerSettings, err := c.settingsManager.GetProviderSetting(provider)
	if err != nil {
		slog.Error("llm:Call", "error", err)
		return nil, fmt.Errorf("failed to get provider setting: %v", err)
	}
	// For local provider, if settings don't exist yet, we might want to fail or handle it.
	// But since we are seeding it, it should exist.
	if providerSettings == nil {
		slog.Error("llm:Call", "error", "provider settings not found", "provider", provider)
		return nil, fmt.Errorf("provider settings not found for: %s", provider)
	}
	url = providerSettings.ApiUrl
	apiKey = providerSettings.ApiKey

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		slog.Error("llm:Call", "error", err)
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	return c.httpClient.Do(httpReq)
}

func (c *Client) generateRequestBody(req types.ChatCompletionRequest, provider string) ([]byte, error) {
	// Convert ChatCompletionRequest to CustomChatRequest
	// The structure is very similar, mainly mapping Model -> ModelName
	customReq := types.CustomChatRequest{
		ModelName:     req.Model,
		Messages:      req.Messages,
		Stream:        req.Stream,
		StreamOptions: req.StreamOptions,
	}

	var provider_rest_api_format string
	if provider == constants.LOCAL_PROVIDER {
		provider_rest_api_format = constants.OPENAI_PROVIDER
	} else {
		provider_rest_api_format = provider
	}

	// Get the cached template
	tmpl, ok := c.templates[provider_rest_api_format]
	if !ok {
		tmpl, ok = c.templates[constants.OPENAI_PROVIDER]
		if !ok {
			return nil, fmt.Errorf("template not found for provider: %s", provider_rest_api_format)
		}
	}

	var bodyBuffer bytes.Buffer
	if err := tmpl.Execute(&bodyBuffer, customReq); err != nil {
		return nil, fmt.Errorf("failed to execute template: %v", err)
	}

	return bodyBuffer.Bytes(), nil
}

package ai

import (
	"context"
)

// ContentType represents the type of content in a message
type ContentType string

const (
	ContentTypeText  ContentType = "text"
	ContentTypeImage ContentType = "image"
)

// MessageContent represents a piece of content in a message
type MessageContent struct {
	Type     ContentType `json:"type"`
	Text     string      `json:"text,omitempty"`
	ImageURL string      `json:"image_url,omitempty"`
}

// ChatMessage represents a message in the conversation
type ChatMessage struct {
	Role    string           `json:"role"`
	Content []MessageContent `json:"content"`
}

// StreamingResponse represents a streaming response chunk
type StreamingResponse struct {
	Type         string `json:"type"`
	Text         string `json:"text,omitempty"`
	Delta        string `json:"delta,omitempty"`
	IsComplete   bool   `json:"is_complete"`
	InputTokens  int    `json:"input_tokens,omitempty"`
	OutputTokens int    `json:"output_tokens,omitempty"`
	Error        string `json:"error,omitempty"`
}

// ChatRequest represents a request to generate chat response
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Stream      bool          `json:"stream"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

// ModelProvider interface defines the contract for AI model providers
type ModelProvider interface {
	Name() string
	Chat(ctx context.Context, req ChatRequest) (<-chan StreamingResponse, error)
	SupportedModels() []string
	SupportsImages() bool
}

// ModelManager manages multiple AI providers
type ModelManager struct {
	providers map[string]ModelProvider
}

func NewModelManager() *ModelManager {
	return &ModelManager{
		providers: make(map[string]ModelProvider),
	}
}

func (m *ModelManager) RegisterProvider(provider ModelProvider) {
	m.providers[provider.Name()] = provider
}

func (m *ModelManager) GetProvider(name string) (ModelProvider, bool) {
	provider, exists := m.providers[name]
	return provider, exists
}

func (m *ModelManager) ListProviders() []string {
	providers := make([]string, 0, len(m.providers))
	for name := range m.providers {
		providers = append(providers, name)
	}
	return providers
}

func (m *ModelManager) GetSupportedModels() map[string][]string {
	models := make(map[string][]string)
	for name, provider := range m.providers {
		models[name] = provider.SupportedModels()
	}
	return models
}

// GetAllProviders returns a map of all registered providers
func (m *ModelManager) GetAllProviders() map[string]ModelProvider {
	providers := make(map[string]ModelProvider)
	for name, provider := range m.providers {
		providers[name] = provider
	}
	return providers
}

// Helper function to create a text message
func NewTextMessage(role, text string) ChatMessage {
	return ChatMessage{
		Role: role,
		Content: []MessageContent{
			{Type: ContentTypeText, Text: text},
		},
	}
}

// Helper function to create a multimodal message
func NewMultimodalMessage(role string, contents []MessageContent) ChatMessage {
	return ChatMessage{
		Role:    role,
		Content: contents,
	}
}

// Helper function to create text content
func NewTextContent(text string) MessageContent {
	return MessageContent{
		Type: ContentTypeText,
		Text: text,
	}
}

// Helper function to create image content
func NewImageContent(imageURL string) MessageContent {
	return MessageContent{
		Type:     ContentTypeImage,
		ImageURL: imageURL,
	}
}

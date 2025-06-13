package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ClaudeProvider implements ModelProvider for Anthropic Claude
type ClaudeProvider struct {
	APIKey  string
	BaseURL string
}

func NewClaudeProvider(apiKey string) *ClaudeProvider {
	return &ClaudeProvider{
		APIKey:  apiKey,
		BaseURL: "https://api.anthropic.com/v1",
	}
}

func (p *ClaudeProvider) Name() string {
	return "claude"
}

func (p *ClaudeProvider) SupportedModels() []string {
	return []string{"claude-3-5-sonnet-20241022", "claude-3-5-haiku-20241022", "claude-3-opus-20240229"}
}

func (p *ClaudeProvider) SupportsImages() bool {
	return true
}

func (p *ClaudeProvider) Chat(ctx context.Context, req ChatRequest) (<-chan StreamingResponse, error) {
	// Convert messages to Claude format
	messages := make([]map[string]interface{}, 0, len(req.Messages))

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			continue // Handle system messages separately in Claude
		}

		if len(msg.Content) == 1 && msg.Content[0].Type == ContentTypeText {
			// Simple text message
			messages = append(messages, map[string]interface{}{
				"role":    msg.Role,
				"content": msg.Content[0].Text,
			})
		} else {
			// Multimodal message
			content := make([]map[string]interface{}, 0, len(msg.Content))
			for _, c := range msg.Content {
				switch c.Type {
				case ContentTypeText:
					content = append(content, map[string]interface{}{
						"type": "text",
						"text": c.Text,
					})
				case ContentTypeImage:
					// Claude expects base64 encoded images
					content = append(content, map[string]interface{}{
						"type": "image",
						"source": map[string]interface{}{
							"type":       "base64",
							"media_type": "image/jpeg",
							"data":       c.ImageURL, // Assuming this is base64 data
						},
					})
				}
			}
			messages = append(messages, map[string]interface{}{
				"role":    msg.Role,
				"content": content,
			})
		}
	}

	requestBody := map[string]interface{}{
		"model":      req.Model,
		"max_tokens": 4096,
		"messages":   messages,
		"stream":     req.Stream,
	}

	if req.Temperature > 0 {
		requestBody["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		requestBody["max_tokens"] = req.MaxTokens
	}

	// Add system message if present
	for _, msg := range req.Messages {
		if msg.Role == "system" && len(msg.Content) > 0 {
			requestBody["system"] = msg.Content[0].Text
			break
		}
	}

	return p.makeStreamingRequest(ctx, requestBody)
}

func (p *ClaudeProvider) makeStreamingRequest(ctx context.Context, requestBody map[string]interface{}) (<-chan StreamingResponse, error) {
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL+"/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	responseChan := make(chan StreamingResponse, 10)

	go func() {
		defer close(responseChan)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		var assistantText strings.Builder

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var chunk map[string]interface{}
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			eventType, _ := chunk["type"].(string)
			switch eventType {
			case "content_block_delta":
				if delta, ok := chunk["delta"].(map[string]interface{}); ok {
					if text, ok := delta["text"].(string); ok {
						assistantText.WriteString(text)
						responseChan <- StreamingResponse{
							Type:  "text_delta",
							Delta: text,
						}
					}
				}
			case "message_stop":
				responseChan <- StreamingResponse{
					Type:       "completion",
					Text:       assistantText.String(),
					IsComplete: true,
				}
			}
		}

		if err := scanner.Err(); err != nil {
			responseChan <- StreamingResponse{
				Type:  "error",
				Error: fmt.Sprintf("Error reading stream: %v", err),
			}
		}
	}()

	return responseChan, nil
}

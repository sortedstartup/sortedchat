package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// GeminiProvider implements ModelProvider for Google Gemini
type GeminiProvider struct {
	APIKey  string
	BaseURL string
}

func NewGeminiProvider(apiKey string) *GeminiProvider {
	return &GeminiProvider{
		APIKey:  apiKey,
		BaseURL: "https://generativelanguage.googleapis.com/v1beta",
	}
}

func (p *GeminiProvider) Name() string {
	return "gemini"
}

func (p *GeminiProvider) SupportedModels() []string {
	return []string{"gemini-1.5-pro", "gemini-1.5-flash", "gemini-1.0-pro"}
}

func (p *GeminiProvider) SupportsImages() bool {
	return true
}

func (p *GeminiProvider) Chat(ctx context.Context, req ChatRequest) (<-chan StreamingResponse, error) {
	// Convert messages to Gemini format
	contents := make([]map[string]interface{}, 0, len(req.Messages))

	for _, msg := range req.Messages {
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}

		if len(msg.Content) == 1 && msg.Content[0].Type == ContentTypeText {
			// Simple text message
			contents = append(contents, map[string]interface{}{
				"role": role,
				"parts": []map[string]interface{}{
					{"text": msg.Content[0].Text},
				},
			})
		} else {
			// Multimodal message
			parts := make([]map[string]interface{}, 0, len(msg.Content))
			for _, c := range msg.Content {
				switch c.Type {
				case ContentTypeText:
					parts = append(parts, map[string]interface{}{
						"text": c.Text,
					})
				case ContentTypeImage:
					// Gemini expects inline_data format
					parts = append(parts, map[string]interface{}{
						"inline_data": map[string]interface{}{
							"mime_type": "image/jpeg",
							"data":      c.ImageURL, // Assuming this is base64 data
						},
					})
				}
			}
			contents = append(contents, map[string]interface{}{
				"role":  role,
				"parts": parts,
			})
		}
	}

	requestBody := map[string]interface{}{
		"contents": contents,
	}

	if req.Temperature > 0 {
		requestBody["generationConfig"] = map[string]interface{}{
			"temperature": req.Temperature,
		}
		if req.MaxTokens > 0 {
			requestBody["generationConfig"].(map[string]interface{})["maxOutputTokens"] = req.MaxTokens
		}
	}

	return p.makeStreamingRequest(ctx, requestBody, req.Model)
}

func (p *GeminiProvider) makeStreamingRequest(ctx context.Context, requestBody map[string]interface{}, model string) (<-chan StreamingResponse, error) {
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	url := fmt.Sprintf("%s/models/%s:streamGenerateContent?key=%s", p.BaseURL, model, p.APIKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	responseChan := make(chan StreamingResponse, 10)

	go func() {
		defer close(responseChan)
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		var assistantText strings.Builder

		for {
			var chunk map[string]interface{}
			if err := decoder.Decode(&chunk); err != nil {
				if err == io.EOF {
					break
				}
				responseChan <- StreamingResponse{
					Type:  "error",
					Error: fmt.Sprintf("Failed to decode chunk: %v", err),
				}
				continue
			}

			if candidates, ok := chunk["candidates"].([]interface{}); ok && len(candidates) > 0 {
				if candidate, ok := candidates[0].(map[string]interface{}); ok {
					if content, ok := candidate["content"].(map[string]interface{}); ok {
						if parts, ok := content["parts"].([]interface{}); ok && len(parts) > 0 {
							if part, ok := parts[0].(map[string]interface{}); ok {
								if text, ok := part["text"].(string); ok {
									assistantText.WriteString(text)
									responseChan <- StreamingResponse{
										Type:  "text_delta",
										Delta: text,
									}
								}
							}
						}
					}
				}
			}
		}

		responseChan <- StreamingResponse{
			Type:       "completion",
			Text:       assistantText.String(),
			IsComplete: true,
		}
	}()

	return responseChan, nil
}

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

// OpenAIProvider implements ModelProvider for OpenAI
type OpenAIProvider struct {
	APIKey  string
	BaseURL string
}

func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	return &OpenAIProvider{
		APIKey:  apiKey,
		BaseURL: "https://api.openai.com/v1",
	}
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

func (p *OpenAIProvider) SupportedModels() []string {
	return []string{"gpt-4o", "gpt-4o-mini", "gpt-4", "gpt-3.5-turbo"}
}

func (p *OpenAIProvider) SupportsImages() bool {
	return true
}

func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (<-chan StreamingResponse, error) {
	// Convert our format to OpenAI's Responses API format
	input := make([]map[string]interface{}, 0, len(req.Messages))

	for _, msg := range req.Messages {
		if len(msg.Content) == 1 && msg.Content[0].Type == ContentTypeText {
			// Simple text message
			input = append(input, map[string]interface{}{
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
						"type": "input_text",
						"text": c.Text,
					})
				case ContentTypeImage:
					content = append(content, map[string]interface{}{
						"type":      "input_image",
						"image_url": c.ImageURL,
					})
				}
			}
			input = append(input, map[string]interface{}{
				"role":    msg.Role,
				"content": content,
			})
		}
	}

	requestBody := map[string]interface{}{
		"model":        req.Model,
		"instructions": "You are a helpful assistant",
		"input":        input,
		"stream":       req.Stream,
	}

	if req.Temperature > 0 {
		requestBody["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		requestBody["max_output_tokens"] = req.MaxTokens
	}

	return p.makeStreamingRequest(ctx, requestBody)
}

func (p *OpenAIProvider) makeStreamingRequest(ctx context.Context, requestBody map[string]interface{}) (<-chan StreamingResponse, error) {
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL+"/responses", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)

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
				responseChan <- StreamingResponse{
					Type:  "error",
					Error: fmt.Sprintf("Failed to parse chunk: %v", err),
				}
				continue
			}

			switch chunk["type"] {
			case "response.output_text.delta":
				if text, ok := chunk["delta"].(string); ok {
					assistantText.WriteString(text)
					responseChan <- StreamingResponse{
						Type:  "text_delta",
						Delta: text,
					}
				}
			case "response.completed":
				response, ok := chunk["response"].(map[string]interface{})
				if !ok {
					continue
				}

				// Extract full text from response
				var fullText string
				if outputArr, ok := response["output"].([]interface{}); ok && len(outputArr) > 0 {
					if outputObj, ok := outputArr[0].(map[string]interface{}); ok {
						if contentArr, ok := outputObj["content"].([]interface{}); ok && len(contentArr) > 0 {
							if contentObj, ok := contentArr[0].(map[string]interface{}); ok {
								fullText, _ = contentObj["text"].(string)
							}
						}
					}
				}

				// Extract token usage
				inputTokens := 0
				outputTokens := 0
				if usage, ok := response["usage"].(map[string]interface{}); ok {
					if val, ok := usage["input_tokens"].(float64); ok {
						inputTokens = int(val)
					}
					if val, ok := usage["output_tokens"].(float64); ok {
						outputTokens = int(val)
					}
				}

				responseChan <- StreamingResponse{
					Type:         "completion",
					Text:         fullText,
					IsComplete:   true,
					InputTokens:  inputTokens,
					OutputTokens: outputTokens,
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

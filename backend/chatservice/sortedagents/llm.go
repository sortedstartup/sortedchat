package sortedagents

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// LLM interface for language model interactions
type LLM interface {
	Call(ctx context.Context, messages []Message, tools []ToolDefinition) (*ChatResponse, error)
	CallStream(ctx context.Context, messages []Message, tools []ToolDefinition) (<-chan StreamChunk, <-chan error)
}

// Message represents a chat message

type Message struct {
	Role string `json:"role"`

	Content string `json:"content"`

	ToolCallID string `json:"tool_call_id,omitempty"`

	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	ExtraContent *ExtraContent `json:"extra_content,omitempty"`
}

// ToolCall represents a tool call from the assistant

type ToolCall struct {
	ID string `json:"id"`

	Type string `json:"type"`

	Function Function `json:"function"`

	ThoughtSignature string `json:"thought_signature,omitempty"`

	ExtraContent *ExtraContent `json:"extra_content,omitempty"`
}

// ExtraContent represents non-standard fields for Gemini/Vertex

type ExtraContent struct {
	Google *GoogleExtra `json:"google,omitempty"`
}

// GoogleExtra contains Google-specific fields like thought_signature

type GoogleExtra struct {
	ThoughtSignature string `json:"thought_signature,omitempty"`
}

// Function represents a function call

type Function struct {
	Name string `json:"name"`

	Arguments string `json:"arguments"`

	ThoughtSignature string `json:"thought_signature,omitempty"`
}

// ToolDefinition represents a tool definition for OpenAI

type ToolDefinition struct {
	Type string `json:"type"`

	Function struct {
		Name string `json:"name"`

		Description string `json:"description"`

		Parameters *JSONSchema `json:"parameters"`

		Strict bool `json:"strict"`
	} `json:"function"`
}

// ChatResponse represents the response from OpenAI

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`

		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// StreamChunk represents a chunk from the streaming API

type StreamChunk struct {
	Usage *struct {
		PromptTokens int `json:"prompt_tokens"`

		CompletionTokens int `json:"completion_tokens"`

		TotalTokens int `json:"total_tokens"`

		PromptTokensDetails struct {
			CachedTokens int `json:"cached_tokens"`
		} `json:"prompt_tokens_details"`
	} `json:"usage,omitempty"`

	Choices []struct {
		Delta struct {
			Role string `json:"role,omitempty"`

			Content string `json:"content,omitempty"`

			ExtraContent *ExtraContent `json:"extra_content,omitempty"`

			ToolCalls []struct {
				Index int `json:"index"`

				ID string `json:"id,omitempty"`

				Type string `json:"type,omitempty"`

				Function struct {
					Name string `json:"name,omitempty"`

					Arguments string `json:"arguments,omitempty"`

					ThoughtSignature string `json:"thought_signature,omitempty"`
				} `json:"function,omitempty"`

				ThoughtSignature string `json:"thought_signature,omitempty"`

				ExtraContent *ExtraContent `json:"extra_content,omitempty"`
			} `json:"tool_calls,omitempty"`
		} `json:"delta"`

		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

// OpenAILLM implements the LLM interface using OpenAI's API
type OpenAILLM struct {
	apiKey  string
	baseURL string
	model   string
}

// NewOpenAILLM creates a new OpenAI LLM instance
func NewOpenAILLM() *OpenAILLM {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// Just warn or allow empty if it's going to be set later?
		// For backward compatibility, we panic if env is missing, or we can change this behavior.
		// Given the user wants to move away from env vars, let's just use empty string or panic.
		// But existing code might rely on panic.
		panic("OPENAI_API_KEY environment variable is required")
	}

	return &OpenAILLM{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1",
		model:   "gpt-4o-mini", // Default model
	}
}

// NewOpenAILLMWithKey creates a new OpenAI LLM instance with a specific API key and model
func NewOpenAILLMWithKey(apiKey string, model string) *OpenAILLM {
	return &OpenAILLM{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1",
		model:   model,
	}
}

// NewOpenAILLMWithConfig creates a new OpenAI LLM instance with a specific API key, base URL and model
func NewOpenAILLMWithConfig(apiKey string, baseURL string, model string) *OpenAILLM {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	// Normalize Base URL: remove trailing slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	// Remove "/chat/completions" if present to avoid duplication
	baseURL = strings.TrimSuffix(baseURL, "/chat/completions")

	// Remove trailing slash again in case it was ".../chat/completions/"
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &OpenAILLM{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
	}
}

// NewOpenAILLMWithModel creates a new OpenAI LLM instance with a specific model
func NewOpenAILLMWithModel(model string) *OpenAILLM {
	llm := NewOpenAILLM()
	llm.model = model
	return llm
}

// Call makes a request to OpenAI's chat completions API
func (llm *OpenAILLM) Call(ctx context.Context, messages []Message, tools []ToolDefinition) (*ChatResponse, error) {
	requestBody := map[string]interface{}{
		"model":    llm.model,
		"messages": messages,
	}

	if len(tools) > 0 {
		requestBody["tools"] = tools
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", llm.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+llm.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, _ := io.ReadAll(resp.Body)

	var chatResponse ChatResponse
	if err := json.NewDecoder(bytes.NewBuffer(bodyBytes)).Decode(&chatResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Synchronize ExtraContent to ThoughtSignature for easier replay
	for i := range chatResponse.Choices {
		msg := &chatResponse.Choices[i].Message
		if msg.ExtraContent != nil && msg.ExtraContent.Google != nil && msg.ExtraContent.Google.ThoughtSignature != "" {
			// If we ever need message-level signature
		}
		for j := range msg.ToolCalls {
			tc := &msg.ToolCalls[j]
			if tc.ExtraContent != nil && tc.ExtraContent.Google != nil && tc.ExtraContent.Google.ThoughtSignature != "" {
				tc.ThoughtSignature = tc.ExtraContent.Google.ThoughtSignature
				tc.Function.ThoughtSignature = tc.ExtraContent.Google.ThoughtSignature
			}
		}
	}

	return &chatResponse, nil
}

// CallStream makes a streaming request to OpenAI's chat completions API
func (llm *OpenAILLM) CallStream(ctx context.Context, messages []Message, tools []ToolDefinition) (<-chan StreamChunk, <-chan error) {
	chunkChan := make(chan StreamChunk, 10)
	errChan := make(chan error, 1)

	go func() {
		defer close(chunkChan)
		defer close(errChan)

		requestBody := map[string]interface{}{
			"model":    llm.model,
			"messages": messages,
			"stream":   true,
			"stream_options": map[string]bool{
				"include_usage": true,
			},
		}

		if len(tools) > 0 {
			requestBody["tools"] = tools
		}

		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			errChan <- fmt.Errorf("failed to marshal request: %v", err)
			return
		}

		req, err := http.NewRequestWithContext(ctx, "POST", llm.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
		if err != nil {
			errChan <- fmt.Errorf("failed to create request: %v", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+llm.apiKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			errChan <- fmt.Errorf("failed to make request: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			errChan <- fmt.Errorf("API request failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			// Skip empty lines
			if line == "" {
				continue
			}

			// SSE format: "data: {...}"
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			// Check for end of stream
			if data == "[DONE]" {
				return
			}

			var chunk StreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				errChan <- fmt.Errorf("failed to decode chunk: %v", err)
				return
			}

			select {
			case chunkChan <- chunk:
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			}
		}

		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("error reading stream: %v", err)
		}
	}()

	return chunkChan, errChan
}

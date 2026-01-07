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
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall represents a tool call from the assistant
type ToolCall struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// Function represents a function call
type Function struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolDefinition represents a tool definition for OpenAI
type ToolDefinition struct {
	Type     string `json:"type"`
	Function struct {
		Name        string      `json:"name"`
		Description string      `json:"description"`
		Parameters  *JSONSchema `json:"parameters"`
		Strict      bool        `json:"strict"`
	} `json:"function"`
}

// ChatResponse represents the response from OpenAI
type ChatResponse struct {
	Choices []struct {
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
}

// StreamChunk represents a chunk from the streaming API
type StreamChunk struct {
	Choices []struct {
		Delta struct {
			Role      string `json:"role,omitempty"`
			Content   string `json:"content,omitempty"`
			ToolCalls []struct {
				Index    int    `json:"index"`
				ID       string `json:"id,omitempty"`
				Type     string `json:"type,omitempty"`
				Function struct {
					Name      string `json:"name,omitempty"`
					Arguments string `json:"arguments,omitempty"`
				} `json:"function,omitempty"`
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
		panic("OPENAI_API_KEY environment variable is required")
	}

	return &OpenAILLM{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1",
		model:   "gpt-4o-mini", // Default model
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

	var chatResponse ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
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

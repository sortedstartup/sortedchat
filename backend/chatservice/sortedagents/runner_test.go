package sortedagents

import (
	"context"
	"testing"
)

type MockLLM struct {
	LastMessages []Message
}

func (m *MockLLM) Call(ctx context.Context, messages []Message, tools []ToolDefinition) (*ChatResponse, error) {
	m.LastMessages = messages
	return &ChatResponse{
		Choices: []struct {
			Message      Message `json:"message"`
			FinishReason string  `json:"finish_reason"`
		}{
			{
				Message: Message{Role: "assistant", Content: TextContent("I understand.")},
			},
		},
	}, nil
}

func (m *MockLLM) CallStream(ctx context.Context, messages []Message, tools []ToolDefinition) (<-chan StreamChunk, <-chan error) {
	ch := make(chan StreamChunk, 1)
	errCh := make(chan error, 1)
	m.LastMessages = messages
	chunk := StreamChunk{}
	chunk.Choices = append(chunk.Choices, struct {
		Delta struct {
			Role         string        `json:"role,omitempty"`
			Content      string        `json:"content,omitempty"`
			ExtraContent *ExtraContent `json:"extra_content,omitempty"`
			ToolCalls    []struct {
				Index    int    `json:"index"`
				ID       string `json:"id,omitempty"`
				Type     string `json:"type,omitempty"`
				Function struct {
					Name             string `json:"name,omitempty"`
					Arguments        string `json:"arguments,omitempty"`
					ThoughtSignature string `json:"thought_signature,omitempty"`
				} `json:"function,omitempty"`
				ThoughtSignature string        `json:"thought_signature,omitempty"`
				ExtraContent     *ExtraContent `json:"extra_content,omitempty"`
			} `json:"tool_calls,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	}{})
	chunk.Choices[0].Delta.Role = "assistant"
	chunk.Choices[0].Delta.Content = "I understand."
	ch <- chunk
	close(ch)
	close(errCh)
	return ch, errCh
}

func TestSystemPromptPersistence(t *testing.T) {
	mockLLM := &MockLLM{}
	runner := &BasicRunner{llm: mockLLM}

	agent := NewAgent("Test", "SYSTEM_INSTRUCTIONS", "model", nil)
	session := NewSession()

	ctx := context.Background()

	// First run
	_, err := runner.Run(ctx, agent, Message{Role: "user", Content: TextContent("Hello")}, 1, session)
	if err != nil {
		t.Fatalf("First run failed: %v", err)
	}

	// Check if system prompt was sent
	hasSystem := false
	for _, msg := range mockLLM.LastMessages {
		if msg.Role == "system" && contentToFlatString(msg.Content) == "SYSTEM_INSTRUCTIONS" {
			hasSystem = true
		}
	}
	if !hasSystem {
		t.Errorf("System prompt missing in first run")
	}

	// Second run with same session
	_, err = runner.Run(ctx, agent, Message{Role: "user", Content: TextContent("How are you?")}, 1, session)
	if err != nil {
		t.Fatalf("Second run failed: %v", err)
	}

	// Check if system prompt was sent in second run
	hasSystem = false
	for _, msg := range mockLLM.LastMessages {
		if msg.Role == "system" && contentToFlatString(msg.Content) == "SYSTEM_INSTRUCTIONS" {
			hasSystem = true
		}
	}
	if !hasSystem {
		t.Errorf("System prompt missing in second run!")
	}
}

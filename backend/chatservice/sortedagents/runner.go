package sortedagents

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
)

// StreamEvent is the base interface for all streaming events
type StreamEvent interface {
	EventType() string
}

// TextChunkEvent represents a chunk of text from the LLM
type TextChunkEvent struct {
	Chunk string
}

func (e *TextChunkEvent) EventType() string { return "text_chunk" }

// ToolCallStartEvent represents the start of a tool call
type ToolCallStartEvent struct {
	ToolName string
	Args     map[string]interface{}
}

func (e *ToolCallStartEvent) EventType() string { return "tool_call_start" }

// ToolCallEndEvent represents the completion of a tool call
type ToolCallEndEvent struct {
	ToolName string
	Result   interface{}
	Error    error
}

func (e *ToolCallEndEvent) EventType() string { return "tool_call_end" }

// CompleteEvent represents the completion of the agent run
type CompleteEvent struct {
	FinalMessage string
}

func (e *CompleteEvent) EventType() string { return "complete" }

// ErrorEvent represents an error during execution
type ErrorEvent struct {
	Error error
}

func (e *ErrorEvent) EventType() string { return "error" }

// Runner executes an agent with a conversation loop
type Runner interface {
	Run(ctx context.Context, agent Agent, input string, maxTurns int, session Session) (string, error)
	RunStream(ctx context.Context, agent Agent, input string, maxTurns int, session Session) <-chan StreamEvent
}

// BasicRunner is a simple implementation of the Runner interface
type BasicRunner struct {
	llm LLM
}

// NewRunner creates a new BasicRunner
func NewRunner() *BasicRunner {
	return &BasicRunner{
		llm: NewOpenAILLM(),
	}
}

// NewRunnerWithLLM creates a new BasicRunner with a specific LLM implementation
func NewRunnerWithLLM(llm LLM) *BasicRunner {
	return &BasicRunner{
		llm: llm,
	}
}

// NewBasicRunnerWithModel creates a new BasicRunner with a specific model
func NewBasicRunnerWithModel(model string) *BasicRunner {
	return &BasicRunner{
		llm: NewOpenAILLMWithModel(model),
	}
}

// formatToolOutput formats tool output as key=value pairs in parentheses
func formatToolOutput(result interface{}) string {
	if result == nil {
		return "()"
	}

	// Try to convert to map
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		// If not a map, just return the JSON representation
		jsonBytes, err := json.Marshal(result)
		if err != nil {
			return fmt.Sprintf("(%v)", result)
		}
		return string(jsonBytes)
	}

	var pairs []string
	for key, value := range resultMap {
		// Skip certain internal fields for cleaner output
		if key == "success" && value == true {
			continue
		}
		pairs = append(pairs, fmt.Sprintf("%s=%v", key, value))
	}

	if len(pairs) == 0 {
		return "(success)"
	}

	return fmt.Sprintf("(%s)", strings.Join(pairs, ", "))
}

// convertToolsToDefinitions converts Tool interfaces to OpenAI tool definitions
func convertToolsToDefinitions(tools []Tool) []ToolDefinition {
	definitions := make([]ToolDefinition, len(tools))
	for i, tool := range tools {
		params := tool.Parameters()
		
		// Auto-detect Strict Mode compatibility
		// If properties are present, we enforce strict mode.
		// If properties are empty (e.g. no args), we disable strict mode.
		isStrict := false
		if params != nil && params.Properties != nil && len(params.Properties) > 0 {
			isStrict = true
		}

		definitions[i] = ToolDefinition{
			Type: "function",
			Function: struct {
				Name        string      `json:"name"`
				Description string      `json:"description"`
				Parameters  *JSONSchema `json:"parameters"`
				Strict      bool        `json:"strict"`
			}{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters:  params,
				Strict:      isStrict,
			},
		}
	}
	return definitions
}

// Run executes the agent with a provided session for conversation history
func (r *BasicRunner) Run(ctx context.Context, agent Agent, input string, maxTurns int, session Session) (string, error) {
	// Initialize conversation from session history
	messages := make([]Message, 0, len(session.GetMessages())+2)

	// If session is empty, add system prompt
	if session.MessageCount() == 0 {
		systemMsg := Message{Role: "system", Content: agent.Instructions()}
		messages = append(messages, systemMsg)
		session.AddMessage(systemMsg)
	} else {
		// Use existing session messages
		messages = append(messages, session.GetMessages()...)
	}

	// Add new user input
	userMessage := Message{Role: "user", Content: input}
	messages = append(messages, userMessage)
	session.AddMessage(userMessage)

	tools := convertToolsToDefinitions(agent.Tools())
	toolMap := make(map[string]Tool)
	for _, tool := range agent.Tools() {
		toolMap[tool.Name()] = tool
	}

	for turn := 0; turn < maxTurns; turn++ {
		// Call LLM
		response, err := r.llm.Call(ctx, messages, tools)
		if err != nil {
			return "", fmt.Errorf("LLM call failed: %v", err)
		}

		if len(response.Choices) == 0 {
			return "", fmt.Errorf("no response from LLM")
		}

		choice := response.Choices[0]
		assistantMessage := choice.Message

		// Add assistant message to conversation and session
		messages = append(messages, assistantMessage)
		session.AddMessage(assistantMessage)

		// If no tool calls, we're done
		if len(assistantMessage.ToolCalls) == 0 {
			return assistantMessage.Content, nil
		}

		// Log what the LLM wants to execute
		slog.Debug("%s -> %d tools", agent.Model(), len(assistantMessage.ToolCalls))
		for _, toolCall := range assistantMessage.ToolCalls {
			// Parse args to format as func_name(arg1=value1, arg2=value2)
			var args map[string]interface{}
			json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
			var argPairs []string
			for k, v := range args {
				argPairs = append(argPairs, fmt.Sprintf("%s=%v", k, v))
			}
			slog.Debug("  %s(%s)", toolCall.Function.Name, strings.Join(argPairs, ", "))
		}

		// Execute tool calls
		for _, toolCall := range assistantMessage.ToolCalls {
			tool, exists := toolMap[toolCall.Function.Name]
			if !exists {
				return "", fmt.Errorf("unknown tool: %s", toolCall.Function.Name)
			}

			// Parse tool arguments
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				return "", fmt.Errorf("failed to parse tool arguments: %v", err)
			}

			// Execute tool
			result, err := tool.Execute(ctx, args)
			if err != nil {
				slog.Debug("  <- %s [FAIL] %v", toolCall.Function.Name, err)
				return "", fmt.Errorf("tool execution failed: %v", err)
			}

			// Convert result to JSON string and log it
			resultJSON, err := json.Marshal(result)
			if err != nil {
				slog.Debug("  <- %s [OK] (marshal error: %v)", toolCall.Function.Name, err)
				return "", fmt.Errorf("failed to marshal tool result: %v", err)
			}

			// Format output as key=value pairs in parentheses
			formattedOutput := formatToolOutput(result)
			slog.Debug("  <- %s %s", toolCall.Function.Name, formattedOutput)

			// Add tool result message
			toolMessage := Message{
				Role:       "tool",
				Content:    string(resultJSON),
				ToolCallID: toolCall.ID,
			}
			messages = append(messages, toolMessage)
			session.AddMessage(toolMessage)
		}
	}

	return "", fmt.Errorf("max turns (%d) reached without completion", maxTurns)
}

// RunStream executes the agent with streaming events and a provided session
func (r *BasicRunner) RunStream(ctx context.Context, agent Agent, input string, maxTurns int, session Session) <-chan StreamEvent {
	eventChan := make(chan StreamEvent, 10)

	go func() {
		defer close(eventChan)

		// Initialize conversation from session history
		messages := make([]Message, 0, len(session.GetMessages())+2)

		// If session is empty, add system prompt
		if session.MessageCount() == 0 {
			systemMsg := Message{Role: "system", Content: agent.Instructions()}
			messages = append(messages, systemMsg)
			session.AddMessage(systemMsg)
		} else {
			// Use existing session messages
			messages = append(messages, session.GetMessages()...)
		}

		// Add new user input
		userMessage := Message{Role: "user", Content: input}
		messages = append(messages, userMessage)
		session.AddMessage(userMessage)

		tools := convertToolsToDefinitions(agent.Tools())
		toolMap := make(map[string]Tool)
		for _, tool := range agent.Tools() {
			toolMap[tool.Name()] = tool
		}

		for turn := 0; turn < maxTurns; turn++ {
			// Call LLM with streaming
			chunkChan, errChan := r.llm.CallStream(ctx, messages, tools)

			// Accumulate the response
			var contentBuilder strings.Builder
			var role string
			toolCallsMap := make(map[int]*ToolCall) // index -> partial tool call

			// Process stream chunks
		streamLoop:
			for {
				select {
				case chunk, ok := <-chunkChan:
					if !ok {
						break streamLoop
					}

					if len(chunk.Choices) == 0 {
						continue
					}

					choice := chunk.Choices[0]
					delta := choice.Delta

					// Capture role
					if delta.Role != "" {
						role = delta.Role
					}

					// Handle text content
					if delta.Content != "" {
						contentBuilder.WriteString(delta.Content)
						eventChan <- &TextChunkEvent{Chunk: delta.Content}
					}

					// Handle tool calls (accumulate deltas)
					for _, tcDelta := range delta.ToolCalls {
						tc, exists := toolCallsMap[tcDelta.Index]
						if !exists {
							tc = &ToolCall{}
							toolCallsMap[tcDelta.Index] = tc
						}

						if tcDelta.ID != "" {
							tc.ID = tcDelta.ID
						}
						if tcDelta.Type != "" {
							tc.Type = tcDelta.Type
						}
						if tcDelta.Function.Name != "" {
							tc.Function.Name = tcDelta.Function.Name
						}
						if tcDelta.Function.Arguments != "" {
							tc.Function.Arguments += tcDelta.Function.Arguments
						}
					}

				case err := <-errChan:
					if err != nil {
						eventChan <- &ErrorEvent{Error: err}
						return
					}
				}
			}

			// Build the complete assistant message
			assistantMessage := Message{
				Role:    role,
				Content: contentBuilder.String(),
			}

			// Convert toolCallsMap to slice
			if len(toolCallsMap) > 0 {
				toolCalls := make([]ToolCall, 0, len(toolCallsMap))
				for i := 0; i < len(toolCallsMap); i++ {
					if tc, exists := toolCallsMap[i]; exists {
						toolCalls = append(toolCalls, *tc)
					}
				}
				assistantMessage.ToolCalls = toolCalls
			}

			// Add assistant message to conversation and session
			messages = append(messages, assistantMessage)
			session.AddMessage(assistantMessage)

			// If no tool calls, we're done
			if len(assistantMessage.ToolCalls) == 0 {
				eventChan <- &CompleteEvent{FinalMessage: assistantMessage.Content}
				return
			}

			// Log and execute tool calls
			slog.Debug("%s -> %d tools", agent.Model(), len(assistantMessage.ToolCalls))

			for _, toolCall := range assistantMessage.ToolCalls {
				tool, exists := toolMap[toolCall.Function.Name]
				if !exists {
					eventChan <- &ErrorEvent{Error: fmt.Errorf("unknown tool: %s", toolCall.Function.Name)}
					return
				}

				// Parse tool arguments
				var args map[string]interface{}
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
					eventChan <- &ErrorEvent{Error: fmt.Errorf("failed to parse tool arguments: %v", err)}
					return
				}

				// Log call details
				var argPairs []string
				for k, v := range args {
					argPairs = append(argPairs, fmt.Sprintf("%s=%v", k, v))
				}
				slog.Debug("  %s(%s)", toolCall.Function.Name, strings.Join(argPairs, ", "))

				// Emit tool call start event
				eventChan <- &ToolCallStartEvent{
					ToolName: toolCall.Function.Name,
					Args:     args,
				}

				// Execute tool
				result, err := tool.Execute(ctx, args)

				// Emit tool call end event
				eventChan <- &ToolCallEndEvent{
					ToolName: toolCall.Function.Name,
					Result:   result,
					Error:    err,
				}

				if err != nil {
					slog.Debug("  <- %s [FAIL] %v", toolCall.Function.Name, err)
					eventChan <- &ErrorEvent{Error: fmt.Errorf("tool execution failed: %v", err)}
					return
				}

				// Convert result to JSON string and log it
				resultJSON, err := json.Marshal(result)
				if err != nil {
					slog.Debug("  <- %s [OK] (marshal error: %v)", toolCall.Function.Name, err)
					eventChan <- &ErrorEvent{Error: fmt.Errorf("failed to marshal tool result: %v", err)}
					return
				}

				// Format output as key=value pairs in parentheses
				formattedOutput := formatToolOutput(result)
				slog.Debug("  <- %s %s", toolCall.Function.Name, formattedOutput)

				// Add tool result message
				toolMessage := Message{
					Role:       "tool",
					Content:    string(resultJSON),
					ToolCallID: toolCall.ID,
				}
				messages = append(messages, toolMessage)
				session.AddMessage(toolMessage)
			}
		}

		eventChan <- &ErrorEvent{Error: fmt.Errorf("max turns (%d) reached without completion", maxTurns)}
	}()

	return eventChan
}

package sortedagents

import (
	"context"
	"testing"
)

func TestNewAgent(t *testing.T) {
	weatherTool := NewWeatherTool()
	tools := []Tool{weatherTool}

	agent := NewAgent(
		"Test Agent",
		"Always be helpful",
		"gpt-5-mini",
		tools,
	)

	if agent.Name() != "Test Agent" {
		t.Errorf("Expected name 'Test Agent', got %s", agent.Name())
	}

	if agent.Instructions() != "Always be helpful" {
		t.Errorf("Expected instructions 'Always be helpful', got %s", agent.Instructions())
	}

	if agent.Model() != "gpt-5-mini" {
		t.Errorf("Expected model 'gpt-5-mini', got %s", agent.Model())
	}

	if len(agent.Tools()) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(agent.Tools()))
	}
}

func TestAgentRun(t *testing.T) {
	mockLLM := &MockLLM{}
	runner := NewRunnerWithLLM(mockLLM)
	weatherTool := NewWeatherTool()
	tools := []Tool{weatherTool}

	agent := NewAgent(
		"Haiku Agent",
		"Always respond in haiku form",
		"gpt-5-mini",
		tools,
	)

	ctx := context.Background()
	maxTurns := 10
	session := NewSession()
	result, err := runner.Run(ctx, agent, Message{Role: "user", Content: TextContent("What's the weather like?")}, maxTurns, session)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result")
	}

	t.Logf("Agent response: %s", result)
}

func TestWeatherTool(t *testing.T) {
	tool := NewWeatherTool()

	if tool.Name() != "get_weather" {
		t.Errorf("Expected name 'get_weather', got %s", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("Expected non-empty description")
	}

	ctx := context.Background()
	args := map[string]any{
		"location": "New York",
	}

	result, err := tool.Execute(ctx, args)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil result")
	}

	t.Logf("Weather tool result: %v", result)
}

func TestWeatherToolMissingLocation(t *testing.T) {
	tool := NewWeatherTool()
	ctx := context.Background()
	args := map[string]any{} // No location parameter

	_, err := tool.Execute(ctx, args)
	if err == nil {
		t.Error("Expected error for missing location parameter")
	}
}

func TestBasicRunner(t *testing.T) {
	mockLLM := &MockLLM{}
	runner := NewRunnerWithLLM(mockLLM)
	weatherTool := NewWeatherTool()
	tools := []Tool{weatherTool}

	agent := NewAgent(
		"Test Agent",
		"Be helpful",
		"gpt-4",
		tools,
	)

	ctx := context.Background()
	maxTurns := 5
	session := NewSession()
	result, err := runner.Run(ctx, agent, Message{Role: "user", Content: TextContent("Hello")}, maxTurns, session)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result")
	}
}

func TestRunnerWithDifferentMaxTurns(t *testing.T) {
	mockLLM := &MockLLM{}
	runner := NewRunnerWithLLM(mockLLM)
	weatherTool := NewWeatherTool()
	tools := []Tool{weatherTool}

	agent := NewAgent(
		"Test Agent",
		"Be helpful",
		"gpt-4",
		tools,
	)

	ctx := context.Background()

	// Test with different max turns - same agent, different execution parameters
	session1 := NewSession()
	result1, err := runner.Run(ctx, agent, Message{Role: "user", Content: TextContent("Hello")}, 3, session1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	session2 := NewSession()
	result2, err := runner.Run(ctx, agent, Message{Role: "user", Content: TextContent("Hello")}, 10, session2)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Both should succeed but may have different behavior based on max turns
	if result1 == "" || result2 == "" {
		t.Error("Expected non-empty results")
	}

	t.Logf("Result with 3 turns: %s", result1)
	t.Logf("Result with 10 turns: %s", result2)
}

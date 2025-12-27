package dao

import (
	"testing"
)

func TestAgentDAO(t *testing.T) {
	daoInstance := SetupSQLiteInMemoryTestDB(t)

	// Test Agent CRUD
	t.Run("Create and Get Agent", func(t *testing.T) {
		agent := AgentRow{
			ID:           "agent-123",
			Name:         "Test Agent",
			Description:  "A test agent",
			SystemPrompt: "You are a test bot",
			Provider:     "openai",
			Model:        "gpt-4",
			LocalTools:   `["tool1", "tool2"]`,
		}

		err := daoInstance.CreateAgent(agent)
		if err != nil {
			t.Fatalf("failed to create agent: %v", err)
		}

		agents, err := daoInstance.GetAgents()
		if err != nil {
			t.Fatalf("failed to get agents: %v", err)
		}
		if len(agents) == 0 {
			t.Errorf("expected at least one agent")
		}
		if agents[0].ID != agent.ID {
			t.Errorf("expected agent ID %s, got %s", agent.ID, agents[0].ID)
		}

		fetchedAgent, err := daoInstance.GetAgent(agent.ID)
		if err != nil {
			t.Fatalf("failed to get specific agent: %v", err)
		}
		if fetchedAgent.Name != agent.Name {
			t.Errorf("expected agent name %s, got %s", agent.Name, fetchedAgent.Name)
		}
	})

	// Test Session CRUD
	t.Run("Create and Get Session", func(t *testing.T) {
		userId := "user-1"
		title := "Test Session"
		session := AgentSessionRow{
			ID:      "session-1",
			AgentID: "agent-123",
			UserID:  userId,
			Status:  "active",
			Title:   &title,
		}

		err := daoInstance.CreateSession(session)
		if err != nil {
			t.Fatalf("failed to create session: %v", err)
		}

		fetchedSession, err := daoInstance.GetSession(session.ID)
		if err != nil {
			t.Fatalf("failed to get session: %v", err)
		}
		if fetchedSession.AgentID != session.AgentID {
			t.Errorf("expected agent ID %s, got %s", session.AgentID, fetchedSession.AgentID)
		}

		sessions, err := daoInstance.GetAgentSessions("agent-123")
		if err != nil {
			t.Fatalf("failed to get agent sessions: %v", err)
		}
		if len(sessions) == 0 {
			t.Errorf("expected at least one session")
		}
	})

	// Test Message CRUD
	t.Run("Add and Get Agent Message", func(t *testing.T) {
		toolName := "calculator"
		toolCallID := "call_123"
		toolArgs := `{"a": 1, "b": 2}`

		msg := AgentMessageRow{
			ID:             "msg-1",
			SessionID:      "session-1",
			SequenceNumber: 1,
			Role:           "tool",
			Type:           "tool_call",
			Content:        "Calling calculator",
			ToolName:       &toolName,
			ToolCallID:     &toolCallID,
			ToolArgs:       &toolArgs,
		}

		err := daoInstance.AddAgentMessage(msg)
		if err != nil {
			t.Fatalf("failed to add agent message: %v", err)
		}

		messages, err := daoInstance.GetAgentMessages("session-1")
		if err != nil {
			t.Fatalf("failed to get agent messages: %v", err)
		}
		if len(messages) != 1 {
			t.Errorf("expected 1 message, got %d", len(messages))
		}
		if messages[0].Type != "tool_call" {
			t.Errorf("expected message type tool_call, got %s", messages[0].Type)
		}
		if *messages[0].ToolName != toolName {
			t.Errorf("expected tool name %s, got %s", toolName, *messages[0].ToolName)
		}
	})
}

package sortedagents

import "testing"

func TestNewSession(t *testing.T) {
	session := NewSession()
	if session == nil {
		t.Fatal("NewSession() returned nil")
	}
	if session.ID() == "" {
		t.Error("Session ID should not be empty")
	}
	if session.MessageCount() != 0 {
		t.Errorf("New session should have 0 messages, got %d", session.MessageCount())
	}
}

func TestNewSessionWithID(t *testing.T) {
	customID := "test-session-123"
	session := NewSessionWithID(customID)
	if session.ID() != customID {
		t.Errorf("Expected ID %s, got %s", customID, session.ID())
	}
}

func TestSessionAddMessage(t *testing.T) {
	session := NewSession()
	msg := Message{Role: "user", Content: "Hello"}
	session.AddMessage(msg)

	if session.MessageCount() != 1 {
		t.Errorf("Expected 1 message, got %d", session.MessageCount())
	}

	messages := session.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message in GetMessages(), got %d", len(messages))
	}
	if messages[0].Content != "Hello" {
		t.Errorf("Expected message content 'Hello', got '%s'", messages[0].Content)
	}
}

func TestSessionClone(t *testing.T) {
	original := NewSession()
	original.AddMessage(Message{Role: "user", Content: "Test"})

	clone := original.Clone()

	// IDs should be different
	if clone.ID() == original.ID() {
		t.Error("Clone should have a different ID")
	}

	// Message count should be the same
	if clone.MessageCount() != original.MessageCount() {
		t.Errorf("Clone should have same message count as original")
	}

	// Modifying clone shouldn't affect original
	clone.AddMessage(Message{Role: "assistant", Content: "Response"})
	if original.MessageCount() == clone.MessageCount() {
		t.Error("Modifying clone affected original")
	}
}

func TestSessionClear(t *testing.T) {
	session := NewSession()
	session.AddMessage(Message{Role: "user", Content: "Test"})
	session.Clear()

	if session.MessageCount() != 0 {
		t.Errorf("Expected 0 messages after Clear(), got %d", session.MessageCount())
	}
}

func TestNewSessionFromMessages(t *testing.T) {
	messages := []Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there"},
	}
	id := "restored-session"
	session := NewSessionFromMessages(id, messages)

	if session.ID() != id {
		t.Errorf("Expected ID %s, got %s", id, session.ID())
	}
	if session.MessageCount() != 2 {
		t.Errorf("Expected 2 messages, got %d", session.MessageCount())
	}
}

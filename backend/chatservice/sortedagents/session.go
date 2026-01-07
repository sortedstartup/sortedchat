package sortedagents

import (
	"github.com/google/uuid"
)

// Session manages conversation history across multiple agent runs
type Session interface {
	ID() string

	AddMessage(msg Message)
	GetMessages() []Message
	AddMessages(msgs []Message)

	Clear()
	Clone() Session
	MessageCount() int
}

// InMemorySession is a basic in-memory implementation of Session
type InMemorySession struct {
	messages []Message
	id       string
}

// NewSession creates a new in-memory session with a generated UUID
func NewSession() *InMemorySession {
	return &InMemorySession{
		messages: make([]Message, 0),
		id:       uuid.New().String(),
	}
}

// NewSessionWithID creates a new in-memory session with a specific ID
func NewSessionWithID(id string) *InMemorySession {
	return &InMemorySession{
		messages: make([]Message, 0),
		id:       id,
	}
}

// NewSessionFromMessages creates a session initialized with existing messages
func NewSessionFromMessages(id string, messages []Message) *InMemorySession {
	return &InMemorySession{
		messages: append([]Message{}, messages...), // deep copy
		id:       id,
	}
}

// GetMessages returns the current message history
func (s *InMemorySession) GetMessages() []Message {
	return s.messages
}

// AddMessage appends a single message to history
func (s *InMemorySession) AddMessage(msg Message) {
	s.messages = append(s.messages, msg)
}

// AddMessages appends multiple messages to history
func (s *InMemorySession) AddMessages(msgs []Message) {
	s.messages = append(s.messages, msgs...)
}

// Clear removes all messages from history
func (s *InMemorySession) Clear() {
	s.messages = make([]Message, 0)
}

// Clone creates a deep copy of the session
func (s *InMemorySession) Clone() Session {
	clone := &InMemorySession{
		messages: make([]Message, len(s.messages)),
		id:       uuid.New().String(), // new ID for clone
	}
	copy(clone.messages, s.messages)
	return clone
}

// MessageCount returns the number of messages
func (s *InMemorySession) MessageCount() int {
	return len(s.messages)
}

// ID returns unique session identifier
func (s *InMemorySession) ID() string {
	return s.id
}

// Verify interface implementation at compile time
var _ Session = (*InMemorySession)(nil)

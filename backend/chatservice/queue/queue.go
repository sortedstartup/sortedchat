package queue

import (
	"context"
	"fmt"
	"sync"
)

// Message represents a message in the queue
type Message struct {
	ID      string
	Subject string
	Data    []byte
}

// NOTE: this is a simple queue for getting started, we want to explore these options
// - embedding NATS in sortedchat

// Queue defines the interface for queue operations
// Designed to work with both in-memory and NATS implementations
type Queue interface {
	// Publish sends a message to the queue
	Publish(ctx context.Context, subject string, data []byte) error

	// Subscribe returns a channel that receives messages for the given subject
	Subscribe(ctx context.Context, subject string) (<-chan Message, error)

	// Close closes the queue and cleans up resources
	Close() error
}

// InMemoryQueue is a simple in-memory implementation of Queue
type InMemoryQueue struct {
	mu          sync.RWMutex
	subscribers map[string][]chan Message
	closed      bool
	msgID       int
}

// NewInMemoryQueue creates a new in-memory queue
func NewInMemoryQueue() *InMemoryQueue {
	return &InMemoryQueue{
		subscribers: make(map[string][]chan Message),
	}
}

// Publish sends a message to all subscribers of the subject
func (q *InMemoryQueue) Publish(ctx context.Context, subject string, data []byte) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return ErrQueueClosed
	}

	q.msgID++
	msg := Message{
		ID:      fmt.Sprintf("msg-%d", q.msgID),
		Subject: subject,
		Data:    data,
	}

	subscribers, exists := q.subscribers[subject]
	if !exists {
		return nil // No subscribers, message is dropped
	}

	// Send to all subscribers (non-blocking)
	for _, ch := range subscribers {
		select {
		case ch <- msg:
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Channel is full, skip this subscriber
		}
	}

	return nil
}

// Subscribe creates a subscription to receive messages for the given subject
func (q *InMemoryQueue) Subscribe(ctx context.Context, subject string) (<-chan Message, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return nil, ErrQueueClosed
	}

	// Create a buffered channel to avoid blocking publishers
	ch := make(chan Message, 100)

	if q.subscribers[subject] == nil {
		q.subscribers[subject] = make([]chan Message, 0)
	}
	q.subscribers[subject] = append(q.subscribers[subject], ch)

	// Handle context cancellation
	go func() {
		<-ctx.Done()
		q.unsubscribe(subject, ch)
		close(ch)
	}()

	return ch, nil
}

// Close closes the queue and all active subscriptions
func (q *InMemoryQueue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return nil
	}

	q.closed = true

	// Close all subscriber channels
	for subject, channels := range q.subscribers {
		for _, ch := range channels {
			close(ch)
		}
		delete(q.subscribers, subject)
	}

	return nil
}

// unsubscribe removes a channel from the subscribers list
func (q *InMemoryQueue) unsubscribe(subject string, ch chan Message) {
	q.mu.Lock()
	defer q.mu.Unlock()

	channels := q.subscribers[subject]
	for i, c := range channels {
		if c == ch {
			// Remove the channel from the slice
			q.subscribers[subject] = append(channels[:i], channels[i+1:]...)
			break
		}
	}

	// Clean up empty subject entries
	if len(q.subscribers[subject]) == 0 {
		delete(q.subscribers, subject)
	}
}

// Common errors
var (
	ErrQueueClosed = fmt.Errorf("queue is closed")
)

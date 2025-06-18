package queue

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestInMemoryQueue_BasicPubSub(t *testing.T) {
	queue := NewInMemoryQueue()
	defer queue.Close()

	ctx := context.Background()

	// Subscribe to a subject
	messages, err := queue.Subscribe(ctx, "test.subject")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Publish a message
	testData := []byte("hello world")
	err = queue.Publish(ctx, "test.subject", testData)
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	// Receive the message
	select {
	case msg := <-messages:
		if msg.Subject != "test.subject" {
			t.Errorf("Expected subject 'test.subject', got '%s'", msg.Subject)
		}
		if string(msg.Data) != string(testData) {
			t.Errorf("Expected data '%s', got '%s'", testData, msg.Data)
		}
		if msg.ID == "" {
			t.Error("Message ID should not be empty")
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for message")
	}
}

func TestInMemoryQueue_MultipleSubscribers(t *testing.T) {
	queue := NewInMemoryQueue()
	defer queue.Close()

	ctx := context.Background()
	subject := "test.multiple"

	// Create multiple subscribers
	sub1, err := queue.Subscribe(ctx, subject)
	if err != nil {
		t.Fatalf("Failed to create subscriber 1: %v", err)
	}

	sub2, err := queue.Subscribe(ctx, subject)
	if err != nil {
		t.Fatalf("Failed to create subscriber 2: %v", err)
	}

	// Publish a message
	testData := []byte("broadcast message")
	err = queue.Publish(ctx, subject, testData)
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	// Both subscribers should receive the message
	var wg sync.WaitGroup
	wg.Add(2)

	checkMessage := func(messages <-chan Message, subscriberName string) {
		defer wg.Done()
		select {
		case msg := <-messages:
			if string(msg.Data) != string(testData) {
				t.Errorf("Subscriber %s: Expected data '%s', got '%s'", subscriberName, testData, msg.Data)
			}
		case <-time.After(1 * time.Second):
			t.Errorf("Subscriber %s: Timeout waiting for message", subscriberName)
		}
	}

	go checkMessage(sub1, "1")
	go checkMessage(sub2, "2")

	wg.Wait()
}

func TestInMemoryQueue_NoSubscribers(t *testing.T) {
	queue := NewInMemoryQueue()
	defer queue.Close()

	ctx := context.Background()

	// Publish to a subject with no subscribers (should not error)
	err := queue.Publish(ctx, "nonexistent.subject", []byte("lost message"))
	if err != nil {
		t.Errorf("Publishing to subject with no subscribers should not error: %v", err)
	}
}

func TestInMemoryQueue_ContextCancellation(t *testing.T) {
	queue := NewInMemoryQueue()
	defer queue.Close()

	ctx, cancel := context.WithCancel(context.Background())

	// Subscribe with cancellable context
	messages, err := queue.Subscribe(ctx, "test.cancel")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Cancel the context
	cancel()

	// Give some time for the goroutine to process cancellation
	time.Sleep(100 * time.Millisecond)

	// Channel should be closed
	select {
	case msg, ok := <-messages:
		if ok {
			t.Errorf("Expected channel to be closed, but received message: %v", msg)
		}
	case <-time.After(1 * time.Second):
		t.Error("Expected channel to be closed")
	}
}

func TestInMemoryQueue_PublishWithCancelledContext(t *testing.T) {
	queue := NewInMemoryQueue()
	defer queue.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Subscribe first
	subCtx := context.Background()
	_, err := queue.Subscribe(subCtx, "test.cancelled")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Publishing with cancelled context should return context error
	err = queue.Publish(ctx, "test.cancelled", []byte("test"))
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got: %v", err)
	}
}

func TestInMemoryQueue_CloseQueue(t *testing.T) {
	queue := NewInMemoryQueue()

	ctx := context.Background()

	// Subscribe to a subject
	messages, err := queue.Subscribe(ctx, "test.close")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Close the queue
	err = queue.Close()
	if err != nil {
		t.Errorf("Failed to close queue: %v", err)
	}

	// Publishing to closed queue should return error
	err = queue.Publish(ctx, "test.close", []byte("should fail"))
	if err != ErrQueueClosed {
		t.Errorf("Expected ErrQueueClosed, got: %v", err)
	}

	// Subscribing to closed queue should return error
	_, err = queue.Subscribe(ctx, "test.new")
	if err != ErrQueueClosed {
		t.Errorf("Expected ErrQueueClosed, got: %v", err)
	}

	// Channel should be closed
	select {
	case _, ok := <-messages:
		if ok {
			t.Error("Expected channel to be closed")
		}
	case <-time.After(1 * time.Second):
		t.Error("Expected channel to be closed")
	}

	// Closing again should not error
	err = queue.Close()
	if err != nil {
		t.Errorf("Closing queue twice should not error: %v", err)
	}
}

func TestInMemoryQueue_ConcurrentPublishSubscribe(t *testing.T) {
	queue := NewInMemoryQueue()
	defer queue.Close()

	ctx := context.Background()
	subject := "test.concurrent"
	numMessages := 100
	numSubscribers := 3

	// Create multiple subscribers
	var subscribers []<-chan Message
	for i := 0; i < numSubscribers; i++ {
		sub, err := queue.Subscribe(ctx, subject)
		if err != nil {
			t.Fatalf("Failed to create subscriber %d: %v", i, err)
		}
		subscribers = append(subscribers, sub)
	}

	// Track received messages per subscriber
	receivedCounts := make([]int, numSubscribers)
	var wg sync.WaitGroup

	// Start subscribers
	for i, sub := range subscribers {
		wg.Add(1)
		go func(subscriberIndex int, messages <-chan Message) {
			defer wg.Done()
			for range messages {
				receivedCounts[subscriberIndex]++
				if receivedCounts[subscriberIndex] >= numMessages {
					return
				}
			}
		}(i, sub)
	}

	// Publish messages concurrently
	var publishWg sync.WaitGroup
	for i := 0; i < numMessages; i++ {
		publishWg.Add(1)
		go func(msgIndex int) {
			defer publishWg.Done()
			data := []byte("message " + string(rune('0'+msgIndex%10)))
			err := queue.Publish(ctx, subject, data)
			if err != nil {
				t.Errorf("Failed to publish message %d: %v", msgIndex, err)
			}
		}(i)
	}

	publishWg.Wait()

	// Wait for all messages to be received (with timeout)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Check that all subscribers received all messages
		for i, count := range receivedCounts {
			if count != numMessages {
				t.Errorf("Subscriber %d received %d messages, expected %d", i, count, numMessages)
			}
		}
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for all messages to be received")
		for i, count := range receivedCounts {
			t.Logf("Subscriber %d received %d messages", i, count)
		}
	}
}

func TestInMemoryQueue_MessageIDs(t *testing.T) {
	queue := NewInMemoryQueue()
	defer queue.Close()

	ctx := context.Background()
	subject := "test.ids"

	messages, err := queue.Subscribe(ctx, subject)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Publish multiple messages and check IDs are unique and sequential
	numMessages := 5
	var receivedIDs []string

	for i := 0; i < numMessages; i++ {
		err := queue.Publish(ctx, subject, []byte("test"))
		if err != nil {
			t.Fatalf("Failed to publish message %d: %v", i, err)
		}
	}

	// Collect all message IDs
	for i := 0; i < numMessages; i++ {
		select {
		case msg := <-messages:
			receivedIDs = append(receivedIDs, msg.ID)
		case <-time.After(1 * time.Second):
			t.Fatalf("Timeout waiting for message %d", i)
		}
	}

	// Check IDs are unique
	idSet := make(map[string]bool)
	for _, id := range receivedIDs {
		if idSet[id] {
			t.Errorf("Duplicate message ID found: %s", id)
		}
		idSet[id] = true
	}

	// Check IDs follow expected pattern
	for i, id := range receivedIDs {
		expectedID := "msg-" + string(rune('1'+i))
		if id != expectedID {
			t.Errorf("Expected message ID '%s', got '%s'", expectedID, id)
		}
	}
}

package api

import (
	"context"
	"testing"

	pb "sortedstartup.com/chatservice/proto"
)

func TestChat(t *testing.T) {

	server := &Server{}

	req := &pb.ChatRequest{Text: "Hello"}
	resp, err := server.Chat(context.Background(), req)
	if err != nil {
		t.Fatalf("Error calling Chat: %v", err)
	}

	expectedResponse := "Hello: Hello"
	if resp.Text != expectedResponse {
		t.Errorf("Expected response %q, got %q", expectedResponse, resp.Text)
	}
}

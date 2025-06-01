package api

import (
	"testing"

	pb "sortedstartup/chatservice/proto"

	"google.golang.org/grpc"
)

func TestChat(t *testing.T) {

	server := &Server{}

	req := &pb.ChatRequest{Text: "Hello"}
	stream := &grpc.GenericServerStream[pb.ChatRequest, pb.ChatResponse]{ServerStream: stream}
	err := server.Chat(req, stream)
	if err != nil {
		t.Fatalf("Error calling Chat: %v", err)
	}
}

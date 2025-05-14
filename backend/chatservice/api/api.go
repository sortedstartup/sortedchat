package api

import (
	"context"
	"fmt"

	pb "sortedstartup.com/chatservice/proto"
)

type Server struct {
	pb.UnimplementedSortedChatServer
}

func (s *Server) Chat(ctx context.Context, req *pb.ChatRequest) (*pb.ChatResponse, error) {
	fmt.Println("Received message:", req.Text)
	return &pb.ChatResponse{
		Text: "Hello: " + req.Text,
	}, nil
}

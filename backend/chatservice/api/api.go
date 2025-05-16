package api

import (
	"context"
	"fmt"
	"time"

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

func (s *Server) LotsOfReplies(req *pb.HelloRequest, stream pb.SortedChat_LotsOfRepliesServer) error {
	words := []string{
		"hello", "this", "is", "a", "test", "of", "streaming", "responses", "over", "gRPC",
		"each", "word", "is", "sent", "with", "a", "delay", "to", "simulate", "live",
		"data", "coming", "in", "real", "time", "we", "hope", "this", "makes", "sense",
		"and", "shows", "how", "server", "side", "streaming", "works", "in", "practice",
		"thank", "you", "for", "trying", "this", "example", "with", "your", "project", "Sanskar",
	}

	for _, word := range words {
		resp := &pb.HelloResponse{
			Text: word,
		}

		if err := stream.Send(resp); err != nil {
			return err
		}

		time.Sleep(500 * time.Millisecond)
	}

	return nil

}

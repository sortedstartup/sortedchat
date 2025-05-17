package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

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

	// apiKey := ""
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OpenAI API key not set")
	}

	requestBody := map[string]interface{}{
		"model":        "gpt-4.1",
		"instructions": "You are a Marine Engineer",
		"input":        req.Text,
		"stream":       true,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.openai.com/v1/responses", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("OpenAI request failed: %v", err)
	}
	defer resp.Body.Close()

	bufferSize := 15
	tokenBuffer := ""

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			var chunk map[string]interface{}
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				log.Printf("Failed to parse chunk: %v", err)
				log.Println("Offending line:", data)
				continue
			}

			if chunk["type"] == "response.output_text.delta" {
				delta, ok := chunk["delta"].(string)
				if ok {
					log.Println("Received token:", delta)

					tokenBuffer += delta

					if len(tokenBuffer) >= bufferSize {
						if err := stream.Send(&pb.HelloResponse{Text: tokenBuffer}); err != nil {
							return fmt.Errorf("failed to send stream response: %v", err)
						}
						tokenBuffer = ""
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %v", err)
	}

	return nil
}

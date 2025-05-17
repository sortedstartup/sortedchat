package api

import (
	"bufio"
	"bytes"
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

func (s *Server) Chat(req *pb.ChatRequest, stream pb.SortedChat_ChatServer) error {

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

	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		var chunk map[string]interface{}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			log.Printf("Failed to parse chunk: %v", err)
			continue
		}

		if chunk["type"] == "response.output_text.delta" {
			if text, ok := chunk["delta"].(string); ok {
				fmt.Println("Text Delta:", text)
				stream.Send(&pb.ChatResponse{Text: text})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %v", err)
	}
	return nil
}

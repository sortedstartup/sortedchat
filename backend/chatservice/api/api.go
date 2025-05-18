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

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

var chatHistory = make(map[string][]ChatMessage)

func (s *Server) Chat(req *pb.ChatRequest, stream pb.SortedChat_ChatServer) error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OpenAI API key not set")
	}
	threadID := req.ThreadId

	if threadID == "" {
		return fmt.Errorf(" Thread ID is required to maintain context")
	}

	history := chatHistory[threadID]

	if len(history) == 0 {
		history = append(history, ChatMessage{
			Role:    "system",
			Content: "You are a Marine Engineer.",
		})
	}

	history = append(history, ChatMessage{
		Role:    "user",
		Content: req.Text,
	})
	chatHistory[threadID] = history

	requestBody := map[string]interface{}{
		"model":        "gpt-4.1",
		"instructions": "You are a Marine Engineer",
		"input":        history,
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

		switch chunk["type"] {
		case "response.output_text.delta":
			if text, ok := chunk["delta"].(string); ok {
				stream.Send(&pb.ChatResponse{Text: text})
			}
		case "response.output_item.done":
			item, ok := chunk["item"].(map[string]interface{})
			if !ok {
				continue
			}

			contentArr, ok := item["content"].([]interface{})
			if !ok || len(contentArr) == 0 {
				continue
			}

			contentObj, ok := contentArr[0].(map[string]interface{})
			if !ok {
				continue
			}

			text, ok := contentObj["text"].(string)
			if !ok {
				continue
			}

			history = append(history, ChatMessage{
				Role:    "assistant",
				Content: text,
			})
			chatHistory[threadID] = history
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %v", err)
	}

	return nil
}

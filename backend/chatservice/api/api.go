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

	"github.com/google/uuid"
	db "sortedstartup.com/chatservice/dao"
	pb "sortedstartup.com/chatservice/proto"
)

type Server struct {
	pb.UnimplementedSortedChatServer
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (s *Server) Chat(req *pb.ChatRequest, stream pb.SortedChat_ChatServer) error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OpenAI API key not set")
	}

	chatId := req.ChatId
	if chatId == "" {
		return fmt.Errorf(" Chat ID is required to maintain context")
	}

	model := req.Model
	if model == "" {
		return fmt.Errorf(" Chat ID is required to maintain context")
	}

	fmt.Println(model)

	var history []ChatMessage
	err := db.DB.Select(&history, `
        SELECT role, content FROM chat_messages 
        WHERE chat_id = ? ORDER BY id`, chatId)
	if err != nil {
		return fmt.Errorf("failed to fetch history: %v", err)
	}

	if len(history) == 0 {
		history = append(history, ChatMessage{
			Role:    "system",
			Content: "You are a helpful assistant",
		})
	}

	_, err = db.DB.Exec(`
        INSERT INTO chat_messages (chat_id, role, content,model) 
        VALUES (?, ?, ?, ?)`, chatId, "user", req.Text, req.Model)
	if err != nil {
		return fmt.Errorf("failed to insert user message: %v", err)
	}

	history = append(history, ChatMessage{Role: "user", Content: req.Text})

	requestBody := map[string]interface{}{
		// "model":        "gpt-4.1",
		"model":        model,
		"instructions": "You are a helpful assistant",
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

	resp, err := http.DefaultClient.Do(httpReq)
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

			_, err := db.DB.Exec(`
                INSERT INTO chat_messages (chat_id, role, content,model) 
                VALUES (?, ?, ?, ?)`, chatId, "assistant", text, req.Model)
			if err != nil {
				log.Printf("failed to insert assistant message: %v", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %v", err)
	}

	return nil
}

func (s *Server) GetHistory(ctx context.Context, req *pb.GetHistoryRequest) (*pb.GetHistoryResponse, error) {
	chatId := req.ChatId
	if chatId == "" {
		return nil, fmt.Errorf("chat ID is required")
	}

	var messages []struct {
		Role    string `db:"role"`
		Content string `db:"content"`
	}

	err := db.DB.Select(&messages, `
		SELECT role, content FROM chat_messages
		WHERE chat_id = ? ORDER BY id`, chatId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch history: %v", err)
	}

	var pbMessages []*pb.ChatMessage
	for _, m := range messages {
		pbMessages = append(pbMessages, &pb.ChatMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	return &pb.GetHistoryResponse{
		History: pbMessages,
	}, nil
}

type ChatRow struct {
	ChatID string `db:"chat_id"`
	Name   string `db:"name"`
}

func (s *Server) GetChatList(ctx context.Context, req *pb.GetChatListRequest) (*pb.GetChatListResponse, error) {
	var rows []ChatRow

	err := db.DB.Select(&rows, `
		SELECT chat_id, name FROM chat_list ORDER BY chat_id
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch chat list: %v", err)
	}

	var chats []*pb.ChatInfo
	for _, row := range rows {
		chats = append(chats, &pb.ChatInfo{
			ChatId: row.ChatID,
			Name:   row.Name,
		})
	}

	return &pb.GetChatListResponse{Chats: chats}, nil
}

func (s *Server) CreateChat(ctx context.Context, req *pb.CreateChatRequest) (*pb.CreateChatResponse, error) {
	name := req.Name
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	chatId := uuid.New().String()

	_, err := db.DB.Exec(`
        INSERT INTO chat_list (chat_id, name) 
        VALUES (?, ?)`, chatId, name)
	if err != nil {
		return nil, fmt.Errorf("failed to insert chat record: %w", err)
	}

	return &pb.CreateChatResponse{
		Message: "Chat created successfully",
		ChatId:  chatId, // return chatId so the frontend can use it for messages
	}, nil
}

func (s *Server) ListModel(ctx context.Context, req *pb.ListModelsRequest) (*pb.ListModelsResponse, error) {
	type Model struct {
		ID   string `db:"id"`
		Name string `db:"name"`
	}

	var models []Model
	err := db.DB.Select(&models, "SELECT id, name FROM model_metadata")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models: %v", err)
	}

	pbModels := make([]*pb.ModelListInfo, 0, len(models))
	for _, m := range models {
		pbModels = append(pbModels, &pb.ModelListInfo{
			Id:    m.ID,
			Label: m.Name,
		})
	}

	return &pb.ListModelsResponse{Models: pbModels}, nil
}

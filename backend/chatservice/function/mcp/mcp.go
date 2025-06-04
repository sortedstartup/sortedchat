package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	apiKey := ""
	mcpURL := "https://test-mcp-server.fly.dev/mcp"

	input := []map[string]interface{}{
		{
			"role":    "user",
			"content": "What is 98564 + 123124 - 32234 + 78878?",
		},
	}

	reqBody := map[string]interface{}{
		"model": "gpt-4.1",
		"input": input,
		"tools": []map[string]interface{}{
			{
				"type":             "mcp",
				"server_label":     "SortedChat",
				"server_url":       mcpURL,
				"require_approval": "never",
			},
		},
		"store": true,
	}

	bodyJson, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatalf("Failed %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/responses", bytes.NewBuffer(bodyJson))
	if err != nil {
		log.Fatalf("Failed %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("failed: %v", err)
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed %v", err)
	}

	// fmt.Println("Responsedata", string(responseData))

	var response struct {
		Output []map[string]interface{} `json:"output"`
	}
	if err := json.Unmarshal(responseData, &response); err != nil {
		log.Fatalf("Failed %v", err)
	}

	for _, item := range response.Output {
		if item["type"] == "message" {
			contentList := item["content"].([]interface{})
			for _, content := range contentList {
				contentMap := content.(map[string]interface{})
				if contentMap["type"] == "output_text" {
					fmt.Println(contentMap["text"].(string))
					return
				}
			}
		}
	}
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func add(a, b int) int {
	return a + b
}

func subtract(a, b int) int {
	return a - b
}

func main() {
	tools := []map[string]interface{}{
		{
			"type":        "function",
			"name":        "add",
			"description": "Add two numbers",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"a": map[string]string{"type": "number"},
					"b": map[string]string{"type": "number"},
				},
				"required": []string{"a", "b"},
			},
		},
		{
			"type":        "function",
			"name":        "subtract",
			"description": "Subtract two numbers",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"a": map[string]string{"type": "number"},
					"b": map[string]string{"type": "number"},
				},
				"required": []string{"a", "b"},
			},
		},
	}

	apiKey := ""

	input := []map[string]interface{}{
		{
			"role":    "user",
			"content": "Find the accurate result for 98564 + 123124 - 32234 + 78878",
		},
	}

	reqBody := map[string]interface{}{
		"model": "gpt-4.1",
		"input": input,
		"tools": tools,
	}

	bodyJson, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatalf("Failed  %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/responses", bytes.NewBuffer(bodyJson))
	if err != nil {
		log.Fatalf("Failed  %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Failed %v", err)
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed  %v", err)
	}

	fmt.Println(string(responseData))

}

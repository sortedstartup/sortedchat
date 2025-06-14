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
	for i := 0; i < 10; i++ {
		reqBody := map[string]interface{}{
			"model": "gpt-4.1",
			"input": input,
			"tools": tools,
			"store": true,
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

		var fullResp struct {
			Output []struct {
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
			} `json:"output"`
		}

		if err := json.Unmarshal(responseData, &fullResp); err != nil {
			log.Fatalf("Failed %v", err)
		}

		for _, msg := range fullResp.Output {
			for _, c := range msg.Content {
				if c.Type == "output_text" {
					fmt.Println(c.Text)
					return
				}
			}
		}

		type responseStructure struct {
			Output []struct {
				Name   string `json:"name"`
				Args   string `json:"arguments"`
				CallID string `json:"call_id"`
			} `json:"output"`
		}
		var apiResponse responseStructure
		if err := json.Unmarshal(responseData, &apiResponse); err != nil {
			log.Fatalf("Failed %v", err)
		}

		for _, output := range apiResponse.Output {
			var argsMap map[string]float64
			err := json.Unmarshal([]byte(output.Args), &argsMap)
			if err != nil {
				log.Fatalf("Failed %v", err)
			}

			a := int(argsMap["a"])
			b := int(argsMap["b"])

			var result int
			switch output.Name {
			case "add":
				result = add(a, b)
			case "subtract":
				result = subtract(a, b)
			default:
				log.Fatalf("function %s", output.Name)
			}

			fmt.Printf("Function: %s(%d, %d) = %d\n", output.Name, a, b, result)

			input = append(input, map[string]interface{}{
				"type":      "function_call",
				"call_id":   output.CallID,
				"name":      output.Name,
				"arguments": output.Args,
			})

			input = append(input, map[string]interface{}{
				"type":    "function_call_output",
				"call_id": output.CallID,
				"output":  fmt.Sprintf("%d", result),
			})

		}
	}

}

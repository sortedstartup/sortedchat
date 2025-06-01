package rag

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type DocumentChunk struct {
	ID         string
	Content    string
	Source     string
	StartBytes int
	EndBytes   int
}

func LoadAndSplitDocs(directory string) ([]DocumentChunk, error) {
	var chunks []DocumentChunk

	files, error := os.ReadDir(directory)
	if error != nil {
		return nil, error
	}

	for _, file := range files {

		contentBytes, err := os.ReadFile(fmt.Sprintf("%s/%s", directory, file.Name()))
		if err != nil {
			return nil, err
		}

		content := string(contentBytes)
		parts := splitText(content, 3)

		for i, part := range parts {
			chunks = append(chunks, DocumentChunk{
				ID:         fmt.Sprintf("%s_%d", file.Name(), i),
				Content:    part.Content,
				StartBytes: part.StartBytes,
				EndBytes:   part.EndBytes,
				Source:     file.Name(),
			})
		}

	}

	return chunks, nil
}

type SplittedChunk struct {
	Content    string
	StartBytes int
	EndBytes   int
}

func splitText(text string, part int) []SplittedChunk {
	var result []SplittedChunk
	length := len(text)

	if length == 0 || part <= 0 {
		return result
	}

	chunkSize := length / part

	for i := 0; i < part; i++ {
		start := i * chunkSize
		end := start + chunkSize

		if i == part-1 {
			end = length
		}

		chunk := text[start:end]
		if chunk != "" {
			result = append(result, SplittedChunk{
				Content:    chunk,
				StartBytes: start,
				EndBytes:   end,
			})
		}
	}

	return result
}

func GenerateEmbeddings(text string) ([]float64, error) {

	payload := map[string]string{
		"model":  "nomic-embed-text:v1.5",
		"prompt": text,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshaling payload:", err)
		return nil, fmt.Errorf("failed %w", err)
	}

	httpReq, error := http.NewRequest("POST", "http://localhost:11434/api/embeddings", bytes.NewBuffer(jsonData))
	if error != nil {
		fmt.Println("Error creating HTTP request:", err)
		return nil, fmt.Errorf("failed %w", error)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed %w", err)
	}

	type Response struct {
		Embedding []float64 `json:"embedding"`
	}

	var result Response
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("failed  %w", err)
	}

	return result.Embedding, nil
}

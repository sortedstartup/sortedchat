package rag

import (
	"fmt"
	"os"
	"strings"
)

type DocumentChunk struct {
	ID      string
	Content string
	Source  string
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
				ID:      fmt.Sprintf("%s_%d", file.Name(), i),
				Content: part,
				Source:  file.Name(),
			})
		}

	}

	return chunks, nil
}

func splitText(text string, part int) []string {
	var result []string
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

		chunk := strings.TrimSpace(text[start:end])
		if chunk != "" {
			result = append(result, chunk)
		}
	}

	return result

}

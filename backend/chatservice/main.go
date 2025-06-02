package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	db "sortedstartup.com/chatservice/dao"
	"sortedstartup.com/chatservice/rag"
)

func main() {

	db.InitDB()

	chunks, err := rag.LoadAndSplitDocs("rag/directory")
	if err != nil {
		panic(err)
	}

	for _, chunk := range chunks {
		embedding, error := rag.GenerateEmbeddings(chunk.Content)
		if error != nil {
			fmt.Printf("failed %w", error)
		}
		embeddingJSON, err := json.Marshal(embedding)
		if err != nil {
			log.Printf("Failed to marshal embedding: %v", err)
			continue
		}

		var insertedID int64
		err = db.DB.QueryRow(`
			INSERT INTO rag_chunks (chunk_id, source, start_byte, end_byte)
			VALUES (?, ?, ?, ?)
			RETURNING id
		`, chunk.ID, chunk.Source, chunk.StartBytes, chunk.EndBytes).Scan(&insertedID)
		if err != nil {
			log.Printf("Failed to insert rag_chunks row: %v", err)
			continue
		}

		_, err = db.DB.Exec(`
			INSERT INTO rag_chunks_vec (id, embedding)
			VALUES (?, ?)`,
			insertedID, string(embeddingJSON),
		)
		if err != nil {
			log.Printf("Failed to insert into rag_chunks_vec: %v", err)
			continue
		}
	}

	userInput := "when do we broadcasts the block to all nodes ?"
	userEmbedding, err := rag.GenerateEmbeddings(userInput)
	if err != nil {
		log.Fatalf("Failed to generate embedding: %v", err)
	}

	embed, err := json.Marshal(userEmbedding)
	if err != nil {
		log.Fatalf("Failed %v", err)
	}

	rows, err := db.DB.Query(`
	SELECT *
	FROM rag_chunks
	WHERE id IN (
  		SELECT id
  		FROM rag_chunks_vec
  		WHERE embedding MATCH ?
  		ORDER BY distance
  		LIMIT 2
	)
	`, string(embed))

	var context string
	for rows.Next() {
		var id int
		var chunkID string
		var source string
		var startByte int
		var endByte int

		if err := rows.Scan(&id, &chunkID, &source, &startByte, &endByte); err != nil {
			log.Printf("Failed to scan row: %v", err)
		}

		fmt.Printf("Chunk ID: %s | Source: %s | Bytes: %d-%d\n",
			chunkID, source, startByte, endByte)

		daata, err := os.ReadFile(fmt.Sprintf("./rag/directory/%s", source))
		if err != nil {
			log.Printf("Failed to read file %v", err)
		}

		chunk := string(daata[startByte:endByte])
		context += chunk
	}

	prompt := fmt.Sprintf(`Use the following context to answer the following question: Context: %s, Question : %s`, context, userInput)
	fmt.Println(prompt)

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Errorf("OpenAI API key not set")
	}

	reqBody := map[string]interface{}{
		"model": "gpt-4.1",
		"input": prompt,
	}

	bodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatalf("Failed  %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/responses", bytes.NewBuffer(bodyJSON))
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

	// fmt.Println(string(responseData))

	type ResponseStructure struct {
		Output []struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}

	var apiResp ResponseStructure
	if err := json.Unmarshal(responseData, &apiResp); err != nil {
		log.Fatalf("Failed %v", err)
	}

	fmt.Println(apiResp.Output[0].Content[0].Text)
}

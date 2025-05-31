package main

import (
	"encoding/json"
	"fmt"
	"log"

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

}

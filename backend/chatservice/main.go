package main

import (
	"fmt"

	db "sortedstartup.com/chatservice/dao"
	"sortedstartup.com/chatservice/rag"
)

func main() {

	db.InitDB()

	chunks, err := rag.LoadAndSplitDocs("rag/directory")
	if err != nil {
		panic(err)
	}

	// for _, chunk := range chunks {
	// 	fmt.Printf("[%s]\n%s\n\n", chunk.ID, chunk.Content)
	// }

	embedding, err := rag.GenerateEmbeddings(chunks[0].Content)
	if err != nil {
		fmt.Errorf("error generating embeddings: %v", err)
	}
	fmt.Println(embedding)

}

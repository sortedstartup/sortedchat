package main

import (
	"sortedstartup.com/chatservice/rag"
)

func main() {
	chunks, err := rag.LoadAndSplitDocs("rag/directory")
	if err != nil {
		panic(err)
	}

	// for _, chunk := range chunks {
	// 	fmt.Printf("[%s]\n%s\n\n", chunk.ID, chunk.Content)
	// }

	rag.GenerateEmbeddings(chunks[0].Content)

}

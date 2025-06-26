package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type defaultPipeline struct {
	ex Extractor
	ch Chunker
	em Embedder
}

func NewPipeline(ex Extractor, ch Chunker, em Embedder) Pipeline {
	return &defaultPipeline{ex: ex, ch: ch, em: em}
}

func (p *defaultPipeline) Run(ctx context.Context, r io.Reader, mime string) ([]Embedding, error) {
	docs, err := p.ex.Extract(ctx, r, mime)
	if err != nil {
		return nil, err
	}

	chunks, err := p.ch.Chunk(ctx, docs)
	if err != nil {
		return nil, err
	}

	return p.em.Embed(ctx, chunks)
}

// RunWithChunks returns both chunks and embeddings, and allows passing metadata
func (p *defaultPipeline) RunWithChunks(ctx context.Context, r io.Reader, mime string, metadata map[string]string) (PipelineResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return PipelineResult{}, err
	}
	docs := Document{
		ID:       metadata["docs_id"],
		MIME:     mime,
		Text:     string(data),
		Metadata: metadata,
	}
	chunks, err := p.ch.Chunk(ctx, docs)
	if err != nil {
		return PipelineResult{}, err
	}
	embs, err := p.em.Embed(ctx, chunks)
	if err != nil {
		return PipelineResult{}, err
	}
	return PipelineResult{Chunks: chunks, Embeddings: embs}, nil
}

// ------

// -- Future uses --
// Apache Tika is a great java based library for extracting text from any file, it can be hosted as a server
// TikaExtractor uses Apache Tika-server for any MIME
type TikaExtractor struct{ Endpoint string }

type TextExtractor struct{}

func (e *TextExtractor) Extract(ctx context.Context, r io.Reader, mime string) (Document, error) {
	data, err := io.ReadAll(r)
	// fmt.Println(string(data))
	if err != nil {
		return Document{}, err
	}

	return Document{
		ID:       "text", //document id form subscriber
		MIME:     mime,
		Text:     string(data),
		Metadata: map[string]string{}, //idk
	}, nil
}

// EqualSizeChunker splits the text into chunks of equal size
// Now generates UUID, sets ProjectID, DocsID, and chunk tracking fields
// projectID and docsID should be passed in Document.Metadata

type EqualSizeChunker struct{ ChunkSize int }

func (e *EqualSizeChunker) Chunk(ctx context.Context, docs Document) ([]Chunk, error) {
	text := docs.Text
	wordsArr := strings.Fields(text)
	chunkSize := e.ChunkSize

	projectID := docs.Metadata["project_id"]
	docsID := docs.Metadata["docs_id"]

	var chunks []Chunk
	byteIdx := 0
	for i := 0; i < len(wordsArr); i += chunkSize {
		end := i + chunkSize
		if end > len(wordsArr) {
			end = len(wordsArr)
		}

		chunkText := strings.Join(wordsArr[i:end], " ")
		startByte := byteIdx
		endByte := byteIdx + len(chunkText)
		chunkID := uuid.New().String()

		chunks = append(chunks, Chunk{
			ID:        chunkID,
			ProjectID: projectID,
			DocsID:    docsID,
			StartByte: startByte,
			EndByte:   endByte,
			Text:      chunkText,
		})
		byteIdx = endByte + 1 // +1 for space
	}
	return chunks, nil
}

// ParagraphChunker splits on double newlines , if paragaph exceeds token limit, it will be split into multiple chunkss
type ParagraphChunker struct{ TokenLimit int }

func (e *ParagraphChunker) Chunk(ctx context.Context, docs Document) ([]Chunk, error) {
	return nil, nil
}

// OLLamaEmbedder hits /v1/embeddings with batching
// Now returns Embedding with ChunkID and Vector

type OLLamaEmbedder struct {
	Model     string
	APIKey    string
	AccountID string
}

func (e *OLLamaEmbedder) Embed(ctx context.Context, chunks []Chunk) ([]Embedding, error) {
	fmt.Println("hii from embedding function")
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ai/run/@cf/baai/bge-m3", e.AccountID)

	var texts []string
	for _, chunk := range chunks {
		texts = append(texts, chunk.Text)
	}

	reqBody := map[string]interface{}{
		"text": texts,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+e.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return nil, err
	}

	var embeddings []Embedding
	result, ok := respData["result"].(map[string]interface{})
	if ok {
		if data, ok := result["data"].([]interface{}); ok {
			for i, emb := range data {
				vec := []float64{}
				if arr, ok := emb.([]interface{}); ok {
					fmt.Printf("Raw embedding %d length: %d\n", i, len(arr))
					for _, v := range arr {
						if f, ok := v.(float64); ok {
							vec = append(vec, f)
						}
					}
				}

				fmt.Printf("Processed vector %d length: %d\n", i, len(vec))

				embeddings = append(embeddings, Embedding{
					ChunkID:  chunks[i].ID,
					Vector:   vec,
					Provider: e.Model,
				})
			}

			fmt.Printf("Total embeddings created: %d\n", len(embeddings))
			for i, emb := range embeddings {
				fmt.Printf("Final embedding %d: ChunkID=%s, Vector length=%d\n", i, emb.ChunkID, len(emb.Vector))
			}
		}
	}
	fmt.Println("sanskar136", embeddings)

	return embeddings, nil
}

func sampleCode() {
	// Sample Usage
	pipeline := NewPipeline(
		&TextExtractor{},
		&EqualSizeChunker{ChunkSize: 512},
		&OLLamaEmbedder{Model: "@cf/baai/bge-m3", APIKey: "ewtTCyj2LDi1W-YefyI8wl_GlUVLMj25mGj5UGNF", AccountID: "0b1342921c6940c378a8bf50d24de341"},
	)

	// Sample code for how to run a pipeline ->
	stringReader := strings.NewReader("Hello, world!")
	pipeline.Run(context.Background(), stringReader, "text/plain")
}

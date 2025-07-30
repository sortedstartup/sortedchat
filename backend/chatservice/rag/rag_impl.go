package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
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
	if err != nil {
		return Document{}, err
	}

	return Document{
		ID:       "text",
		MIME:     mime,
		Text:     string(data),
		Metadata: map[string]string{},
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
	url := os.Getenv("EMBEDDING_API_URL")

	var embeddings []Embedding
	
	for _, chunk := range chunks {
		reqBody := map[string]interface{}{
			"model":  "nomic-embed-text",
			"prompt": chunk.Text,
		}
		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		var respData map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		var vec []float64
		if embeddingData, ok := respData["embedding"].([]interface{}); ok {
			vec = make([]float64, len(embeddingData))
			for i, v := range embeddingData {
				if f, ok := v.(float64); ok {
					vec[i] = f
				}
			}
		}

		embeddings = append(embeddings, Embedding{
			ChunkID:  chunk.ID,
			Vector:   vec,
			Provider: e.Model,
		})
	}

	return embeddings, nil
}
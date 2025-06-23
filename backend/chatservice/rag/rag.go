package rag

import (
	"context"
	"io"
)

// Abstraction to create a extraction and embedding generation pipeline
type Document struct {
	ID       string
	MIME     string            // "application/pdf", "text/markdown", …
	Text     string            // full plain-text body, need to careful, huge RAM usage
	Metadata map[string]string // extra fields (author, title, etc.)
}

type Chunk struct {
	DocID string // parent document
	//TODO: need to store which piece of text was used, may be a byte range ? from the original docs or parsed text ? obviously we need to store the original text then
	Text string // chunk body
}

type Embedding struct {
	ChunkID  string    // Need to decide
	Vector   []float64 // dense vector
	Provider string    // "openai", "bge-base-en", …
}

// 1. file/stream → Document, this will be used for extracting text from for e.g. PDFs
type Extractor interface {
	Extract(ctx context.Context, r io.Reader, mime string) (Document, error)
}

// 2. Chunker — Document → []Chunk
type Chunker interface {
	Chunk(ctx context.Context, docs Document) ([]Chunk, error)
}

// 3. Embedder — []Chunk → []Embedding
type Embedder interface {
	Embed(ctx context.Context, chunks []Chunk) ([]Embedding, error)
}

// 4. Pipeline — convenience wrapper
// This signature may change, since we may want to different document types
type Pipeline interface {
	Run(ctx context.Context, r io.Reader, mime string) ([]Embedding, error)
}

/*
TODO: RAG Retreival Pipeline

- find relevant documents
- extract relevant information
- create prompt

- send prompt to LLMProvider

- return answer
*/

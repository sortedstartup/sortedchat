package rag

import (
	"context"
	"fmt"
	"io"
	"strings"
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

// ------

// -- Future uses --
// Apache Tika is a great java based library for extracting text from any file, it can be hosted as a server
// TikaExtractor uses Apache Tika-server for any MIME
type TikaExtractor struct{ Endpoint string }

type TextExtractor struct{}

func (e *TextExtractor) Extract(ctx context.Context, r io.Reader, mime string) (Document, error) {
	data, err := io.ReadAll(r)
	fmt.Println(string(data))
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
type EqualSizeChunker struct{ ChunkSize int }

func (e *EqualSizeChunker) Chunk(ctx context.Context, docs Document) ([]Chunk, error) {
	text := docs.Text
	length := len(text)

	partSize := length / 3
	chunks := make([]Chunk, 0, 3)

	chunks = append(chunks, Chunk{
		DocID: docs.ID,
		Text:  text[0:partSize],
	})

	chunks = append(chunks, Chunk{
		DocID: docs.ID,
		Text:  text[partSize : 2*partSize],
	})

	chunks = append(chunks, Chunk{
		DocID: docs.ID,
		Text:  text[2*partSize:],
	})

	return chunks, nil
}

// ParagraphChunker splits on double newlines , if paragaph exceeds token limit, it will be split into multiple chunkss
type ParagraphChunker struct{ TokenLimit int }

func (e *ParagraphChunker) Chunk(ctx context.Context, docs Document) ([]Chunk, error) {
	return nil, nil
}

// OLLamaEmbedder hits /v1/embeddings with batching
type OLLamaEmbedder struct{ Model, APIKey string }

func (e *OLLamaEmbedder) Embed(ctx context.Context, chunks []Chunk) ([]Embedding, error) {
	return nil, nil
}

func sampleCode() {
	// Sample Usage
	pipeline := NewPipeline(
		&TextExtractor{},
		&EqualSizeChunker{ChunkSize: 512},
		&OLLamaEmbedder{Model: "bge-base-en", APIKey: "ollama"},
	)

	// Sample code for how to run a pipeline ->
	stringReader := strings.NewReader("Hello, world!")
	pipeline.Run(context.Background(), stringReader, "text/plain")
}

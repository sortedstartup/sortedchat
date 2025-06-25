package rag

import (
	"context"
	"fmt"
	"strings"
)

// BasicRetrieve performs simple cosine similarity search
func BasicRetrieve(ctx context.Context, chunks []Chunk, embeddings []Embedding, embedding []float64, params SearchParams) ([]Result, error) {
	var results []Result
	// TODO: Implement cosine similarity search
	return results, nil
}

// BasicPromptBuilder creates a simple RAG prompt
func BasicPromptBuilder(ctx context.Context, query string, results []Result) (string, error) {
	if len(results) == 0 {
		return fmt.Sprintf("Answer the following question: %s", query), nil
	}

	var contextParts []string
	for _, result := range results {
		contextParts = append(contextParts, fmt.Sprintf("- %s", result.Chunk.Text))
	}

	prompt := fmt.Sprintf(`Use the following context to answer the question:

Context:
%s

Question: %s
Answer:`, strings.Join(contextParts, "\n"), query)

	return prompt, nil
}

// BasicPipeline combines retrieval and prompt building
func BasicPipeline(ctx context.Context, chunks []Chunk, embeddings []Embedding, embedding []float64, query string, params SearchParams) (*Response, error) {
	results, err := BasicRetrieve(ctx, chunks, embeddings, embedding, params)
	if err != nil {
		return nil, err
	}

	prompt, err := BasicPromptBuilder(ctx, query, results)
	if err != nil {
		return nil, err
	}

	return &Response{
		Results: results,
		Prompt:  prompt,
	}, nil
}

// RunSamplePipeline demonstrates how to use the RAG pipeline with sample data
func RunSamplePipeline(ctx context.Context) (*Response, error) {
	// Sample data
	chunks := []Chunk{
		{ID: "chunk-1", ProjectID: "project-123", Text: "Go is a programming language developed by Google."},
		{ID: "chunk-2", ProjectID: "project-123", Text: "React is a JavaScript library for building user interfaces."},
		{ID: "chunk-3", ProjectID: "project-123", Text: "Tailwind CSS is a utility-first CSS framework."},
	}

	embeddings := []Embedding{
		{ChunkID: "chunk-1", Vector: []float64{0.1, 0.2, 0.3, 0.4}, Provider: "openai"},
		{ChunkID: "chunk-2", Vector: []float64{0.5, 0.6, 0.7, 0.8}, Provider: "openai"},
		{ChunkID: "chunk-3", Vector: []float64{0.2, 0.4, 0.6, 0.8}, Provider: "openai"},
	}

	// Run pipeline directly
	return BasicPipeline(ctx, chunks, embeddings,
		[]float64{0.15, 0.25, 0.35, 0.45}, // query embedding
		"What is Go programming language?",
		SearchParams{TopK: 2, ProjectID: "project-123"},
	)
}

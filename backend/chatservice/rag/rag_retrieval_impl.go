package rag

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

func BasicRetrieve(ctx context.Context, embedding []float64, params SearchParams) ([]Result, error) {
	// needs dao
	return nil, nil
}

// BasicPromptBuilder creates a simple RAG prompt
func BasicPromptBuilder(ctx context.Context, query string, results []Result) (string, error) {
	slog.Info("rag_retrieval_impl:BasicPromptBuilder", "query", query, "results", results)
	if len(results) == 0 {
		return fmt.Sprintf("Answer the following question: %s", query), nil
	}

	var contextParts []string
	for _, result := range results {
		slog.Info("rag_retrieval_impl:BasicPromptBuilder", "result", result)
		contextParts = append(contextParts, fmt.Sprintf("- %s", result.Chunk.Text))
	}

	prompt := fmt.Sprintf(`Use the following context to answer the question:

Context:
%s

Question: %s
Answer:`, strings.Join(contextParts, "\n"), query)

	return prompt, nil
}

func BasicRetrievePipeline(ctx context.Context, retriever Retrieve, promptBuilder BuildPrompt, embedding []float64, query string, params SearchParams) (*Response, error) {
	slog.Info("rag_retrieval_impl:BasicRetrievePipeline", "embedding_dim", len(embedding), "query", query, "params", params)
	results, err := retriever(ctx, embedding, params)
	if err != nil {
		slog.Error("rag_retrieval_impl:BasicRetrievePipeline", "step", "failed to retrieve", "error", err, "embedding_dim", len(embedding), "query", query, "params", params)
		return nil, err
	}

	prompt, err := promptBuilder(ctx, query, results)
	if err != nil {
		slog.Error("rag_retrieval_impl:BasicRetrievePipeline", "step", "failed to build prompt", "error", err, "embedding_dim", len(embedding), "query", query, "params", params)
		return nil, err
	}

	return &Response{
		Results: results,
		Prompt:  prompt,
	}, nil
}

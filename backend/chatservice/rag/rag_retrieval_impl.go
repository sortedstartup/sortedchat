package rag

import (
	"context"
	"fmt"
	"strings"
)

func BasicRetrieve(ctx context.Context, embedding []float64, params SearchParams) ([]Result, error) {
	// needs dao
	return nil, nil
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

func BasicRetrievePipeline(ctx context.Context, retriever Retrieve, promptBuilder BuildPrompt, embedding []float64, query string, params SearchParams) (*Response, error) {
	results, err := retriever(ctx, embedding, params)
	if err != nil {
		return nil, err
	}

	prompt, err := promptBuilder(ctx, query, results)
	if err != nil {
		return nil, err
	}

	return &Response{
		Results: results,
		Prompt:  prompt,
	}, nil
}

func testPipeline() {

	var retrievalPipeline RetrievePipeline = BasicRetrievePipeline

	x, err := retrievalPipeline(context.Background(), BasicRetrieve, BasicPromptBuilder, []float64{0.15, 0.25, 0.35, 0.45}, "What is Go programming language?", SearchParams{TopK: 2, ProjectID: "project-123"})
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println("Response:", x)

}

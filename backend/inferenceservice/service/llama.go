package service

import (
	"context"
	"fmt"
	l "sortedstartup/inferenceservice/llamacpp"
)

func (s *InferenceService) Predict(ctx context.Context, modelName string, prompt string) (<-chan string, error) {
	// Load model by name (uses caching for faster subsequent loads)

	fmt.Println("predict %s, %s", modelName, prompt)
	model, err := l.LoadModelByName(modelName)
	if err != nil {
		return nil, err
	}
	// Note: Don't defer model.Close() here since we're caching the model

	// Generate predictions
	tokens, err := l.Predict(model, prompt)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

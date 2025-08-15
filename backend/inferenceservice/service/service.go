package service

import "context"

type InferenceService struct {
}

func NewInferenceService() *InferenceService {
	return &InferenceService{}
}

func (s *InferenceService) Infer(ctx context.Context) error {
	return nil
}

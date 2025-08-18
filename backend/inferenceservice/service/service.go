package service

import (
	"context"
	"log"
	"sortedstartup/inferenceservice/dao"
)

type InferenceService struct {
	dao dao.DAO
}

func NewInferenceService(daoFactory dao.DAOFactory) *InferenceService {
	dao, err := daoFactory.CreateDAO()
	if err != nil {
		log.Fatalf("Failed to create DAO: %v", err)
	}
	return &InferenceService{
		dao: dao,
	}
}

func (s *InferenceService) Infer(ctx context.Context, dummy string) error {
	return s.dao.Infer(dummy)
}

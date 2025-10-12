package service

import (
	"context"
	"log"
	"sortedstartup/paymentservice/dao"
)

type PaymentService struct {
	dao dao.DAO
}

func NewPaymentService(daoFactory dao.DAOFactory) *PaymentService {
	dao, err := daoFactory.CreateDAO()
	if err != nil {
		log.Fatalf("Failed to create DAO: %v", err)
	}
	return &PaymentService{
		dao: dao,
	}
}

func (s *PaymentService) Infer(ctx context.Context, dummy string) error {
	return s.dao.Infer(dummy)
}

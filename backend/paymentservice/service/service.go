package service

import (
	"context"
	"log/slog"
	"sortedstartup/paymentservice/dao"
)

type PaymentService struct {
	dao dao.DAO
}

func NewPaymentService(daoFactory dao.DAOFactory) *PaymentService {
	dao, err := daoFactory.CreateDAO()
	if err != nil {
		slog.Error("paymentservice:service:NewPaymentService", "error", err)
		return nil
	}
	return &PaymentService{
		dao: dao,
	}
}

func (s *PaymentService) Infer(ctx context.Context, dummy string) error {
	return s.dao.Infer(dummy)
}

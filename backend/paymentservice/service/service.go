package service

import (
	"context"
	"fmt"
	"log/slog"
	"sortedstartup/paymentservice/dao"

	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/product"
)

type PaymentService struct {
	dao dao.DAO
}

func NewPaymentService(daoFactory dao.DAOFactory) (*PaymentService, error) {
	dao, err := daoFactory.CreateDAO()
	if err != nil {
		slog.Error("paymentservice:service:NewPaymentService", "error", err)
		return nil, err
	}
	return &PaymentService{
		dao: dao,
	}, nil
}

func (s *PaymentService) Infer(ctx context.Context, dummy string) error {
	return s.dao.Infer(dummy)
}

func (s *PaymentService) CreateProduct(ctx context.Context, userID string, name string, description string, amountInCents int64, currency string) (string, error) {

	slog.Info("paymentservice:service:CreateProduct", "userID", userID, "name", name, "amountInCents", amountInCents, "currency", currency)

	// add MIN and MAX lenght validation for name and description
	if name == "" || description == "" || currency == "" || amountInCents <= 0 {
		slog.Error("paymentservice:service:CreateProduct", "error", "invalid request, please try again with valid parameters")
		return "", fmt.Errorf("invalid request, please try again with valid parameters")
	}

	// Create the product first (amount is already in cents)
	productParams := &stripe.ProductParams{
		Name:        stripe.String(name),
		Description: stripe.String(description),
		DefaultPriceData: &stripe.ProductDefaultPriceDataParams{
			Currency:   stripe.String(currency),
			UnitAmount: stripe.Int64(amountInCents),
		},
	}

	stripeProduct, err := product.New(productParams)
	if err != nil {
		slog.Error("paymentservice:service:CreateProduct", "error", err)
		return "", fmt.Errorf("failed to process the request")
	}

	slog.Info("paymentservice:service:CreateProduct", "id", stripeProduct.ID)

	_, err = s.dao.CreateProduct(stripeProduct.ID, userID, name, description, amountInCents, currency)
	if err != nil {
		slog.Error("paymentservice:service:CreateProduct", "error", err)
		return "", fmt.Errorf("failed to process the request")
	}

	// Return the product ID (you might want to return both product ID and price ID)
	return stripeProduct.ID, nil
}

func (s *PaymentService) ListProducts(ctx context.Context) ([]*dao.Product, error) {
	slog.Info("paymentservice:service:ListProducts")
	products, err := s.dao.ListProducts()
	if err != nil {
		slog.Error("paymentservice:service:ListProducts", "error", err)
		return nil, fmt.Errorf("failed to process the request")
	}
	slog.Info("paymentservice:service:ListProducts", "products", products)
	return products, nil
}

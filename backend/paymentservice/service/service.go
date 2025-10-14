package service

import (
	"context"
	"fmt"
	"log/slog"
	"sortedstartup/paymentservice/dao"
	"strconv"

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

func (s *PaymentService) CreateProduct(ctx context.Context, userID string, name string, description string, cost string, currency string) (string, error) {
	slog.Info("paymentservice:service:CreateProduct", "userID", userID, "name", name)

	// Convert price string to int64 (Stripe expects amount in smallest currency unit)
	priceAmount, err := strconv.ParseInt(cost, 10, 64)
	if err != nil {
		slog.Error("paymentservice:service:CreateProduct", "error", err)
		return "", fmt.Errorf("failed to process the request")
	}
	slog.Info("paymentservice:service:CreateProduct", "priceAmount", priceAmount)

	// Create the product first
	productParams := &stripe.ProductParams{
		Name:        stripe.String(name),
		Description: stripe.String(description),
		DefaultPriceData: &stripe.ProductDefaultPriceDataParams{
			Currency:   stripe.String(currency),
			UnitAmount: stripe.Int64(priceAmount * 100),
		},
	}

	product, err := product.New(productParams)
	if err != nil {
		slog.Error("paymentservice:service:CreateProduct", "error", err)
		return "", fmt.Errorf("failed to process the request")
	}

	slog.Info("paymentservice:service:CreateProduct", "id", product.ID)

	_, err = s.dao.CreateProduct(product.ID, userID, name, description, cost, currency)
	if err != nil {
		slog.Error("paymentservice:service:CreateProduct", "error", err)
		return "", fmt.Errorf("failed to process the request")
	}

	slog.Info("paymentservice:service:CreateProduct", "defaultPrice", product.DefaultPrice.ID)

	// Return the product ID (you might want to return both product ID and price ID)
	return product.ID, nil
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

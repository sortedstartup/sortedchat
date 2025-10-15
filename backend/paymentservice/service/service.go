package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sortedstartup/paymentservice/dao"

	razorpay "github.com/razorpay/razorpay-go"
)

type PaymentService struct {
	dao            dao.DAO
	razorpayClient *razorpay.Client
}

func NewPaymentService(daoFactory dao.DAOFactory) (*PaymentService, error) {
	dao, err := daoFactory.CreateDAO()
	if err != nil {
		slog.Error("paymentservice:service:NewPaymentService", "error", err)
		return nil, err
	}
	return &PaymentService{
		dao:            dao,
		razorpayClient: razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET")),
	}, nil
}

func (s *PaymentService) Infer(ctx context.Context, dummy string) error {
	return s.dao.Infer(dummy)
}

func (s *PaymentService) CreateProduct(ctx context.Context, userID string, name string, description string, amountInSmallestUnit int64, currency string) (string, error) {
	slog.Info("paymentservice:service:CreateProduct", "userID", userID, "name", name)

	// Create product on Stripe
	stripeProductID, err := s.CreateProductStripe(ctx, name, description, amountInSmallestUnit, currency)
	if err != nil {
		slog.Error("paymentservice:service:CreateProduct", "error", "failed to create Stripe product", "details", err)
		return "", fmt.Errorf("failed to create Stripe product")
	}
	slog.Info("paymentservice:service:CreateProduct", "stripeProductID", stripeProductID)

	// Create product on Razorpay
	razorpayProductID, err := s.CreateProductRazorpay(ctx, name, description, amountInSmallestUnit, currency)
	if err != nil {
		slog.Error("paymentservice:service:CreateProduct", "error", "failed to create Razorpay product", "details", err)
		return "", fmt.Errorf("failed to create Razorpay product")
	}
	slog.Info("paymentservice:service:CreateProduct", "razorpayProductID", razorpayProductID)

	// Save to database with both provider IDs
	productID, err := s.dao.CreateProduct(stripeProductID, razorpayProductID, userID, name, description, amountInSmallestUnit, currency)
	if err != nil {
		slog.Error("paymentservice:service:CreateProduct", "error", "failed to save product to database", "details", err)
		return "", fmt.Errorf("failed to save product to database")
	}

	slog.Info("paymentservice:service:CreateProduct", "productID", productID, "stripeProductID", stripeProductID, "razorpayProductID", razorpayProductID)
	return productID, nil
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

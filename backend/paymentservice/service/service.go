package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sortedstartup/paymentservice/dao"
	"strconv"

	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/checkout/session"
	"github.com/stripe/stripe-go/v83/price"
	"github.com/stripe/stripe-go/v83/product"
	"github.com/stripe/stripe-go/v83/webhook"
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

func (s *PaymentService) ListProducts(ctx context.Context, userID string) ([]*dao.Product, error) {
	slog.Info("paymentservice:service:ListProducts", "userID", userID)
	products, err := s.dao.ListProducts(userID)
	if err != nil {
		slog.Error("paymentservice:service:ListProducts", "error", err)
		return nil, fmt.Errorf("failed to process the request")
	}
	slog.Info("paymentservice:service:ListProducts", "products", products)
	return products, nil
}

func (s *PaymentService) CreateCheckoutSession(ctx context.Context, userID string, productID string) (string, error) {
	slog.Info("paymentservice:service:CreateCheckoutSession", "userID", userID, "productID", productID)

	var priceID string
	//lets get price id from product id
	params := &stripe.PriceListParams{
		Product: stripe.String(productID),
		Active:  stripe.Bool(true),
	}

	i := price.List(params)
	if i.Next() {
		priceID = i.Price().ID
		slog.Info("paymentservice:service:CreateCheckoutSession", "priceFound", priceID)
	} else {
		slog.Error("paymentservice:service:CreateCheckoutSession", "error", "no active prices found for product")
		return "", fmt.Errorf("no active prices found for product")
	}

	if err := i.Err(); err != nil {
		slog.Error("paymentservice:service:CreateCheckoutSession", "error", err)
		return "", fmt.Errorf("failed to list prices")
	}

	//lets create session
	sessionParams := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String("http://localhost:5173/success"),
		CancelURL:  stripe.String("http://localhost:5173/cancel"),
	}

	session, err := session.New(sessionParams)
	if err != nil {
		slog.Error("paymentservice:service:CreateCheckoutSession", "error", err)
		return "", fmt.Errorf("failed to process the request")
	}

	slog.Info("paymentservice:service:CreateCheckoutSession", "product", priceID)
	slog.Info("paymentservice:service:CreateCheckoutSession", "session", session.ID)
	slog.Info("paymentservice:service:CreateCheckoutSession", "session", session.URL)

	return session.URL, nil
}

func (s *PaymentService) HandleWebhook(ctx context.Context, r *http.Request) error {
	slog.Info("paymentservice:service:HandleWebhook")

	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(nil, r.Body, MaxBodyBytes)

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("paymentservice:service:HandleWebhook", "error", "failed to read request body", "details", err)
		return fmt.Errorf("error reading request body")
	}

	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	sigHeader := r.Header.Get("Stripe-Signature")

	slog.Info("paymentservice:service:HandleWebhook", "webhook", "received")
	slog.Info("paymentservice:service:HandleWebhook", "signature", sigHeader)
	slog.Info("paymentservice:service:HandleWebhook", "payload", string(payload))

	event, err := webhook.ConstructEventWithOptions(payload, sigHeader, endpointSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		slog.Error("paymentservice:service:HandleWebhook", "error", "signature verification failed", "details", err)
		return fmt.Errorf("webhook signature verification failed: %v", err)
	}

	switch event.Type {
	case "checkout.session.completed":
		err := s.handleCheckoutSessionCompleted(ctx, event)
		if err != nil {
			slog.Error("paymentservice:service:HandleWebhook", "error", "failed to handle checkout session completed", "details", err)
			return fmt.Errorf("failed to handle checkout session completed: %v", err)
		}

	case "invoice.payment_failed":
		slog.Info("paymentservice:service:HandleWebhook", "event", "payment failed for subscription")
		err := s.handleInvoicePaymentFailed(ctx, event)
		if err != nil {
			slog.Error("paymentservice:service:HandleWebhook", "error", "failed to handle invoice payment failed", "details", err)
			return fmt.Errorf("failed to handle invoice payment failed: %v", err)
		}

	default:
		slog.Info("paymentservice:service:HandleWebhook", "event", "unhandled event type", "type", event.Type)
	}

	slog.Info("paymentservice:service:HandleWebhook", "status", "webhook processed successfully")
	return nil
}

func (s *PaymentService) handleCheckoutSessionCompleted(ctx context.Context, event stripe.Event) error {
	var session stripe.CheckoutSession
	err := json.Unmarshal(event.Data.Raw, &session)
	if err != nil {
		slog.Error("paymentservice:service:handleCheckoutSessionCompleted", "error", "failed to parse webhook JSON", "details", err)
		return fmt.Errorf("error parsing webhook JSON: %v", err)
	}

	email := session.CustomerDetails.Email
	if email == "" {
		// Fallback if CustomerDetails is nil
		email = session.CustomerEmail
	}

	// isSub := session.Mode == stripe.CheckoutSessionModeSubscription

	slog.Info("paymentservice:service:handleCheckoutSessionCompleted", "email", email, "mode", session.Mode, "sessionID", session.ID)

	// Save user after payment using your DAO
	// err = s.dao.SaveUserAfterPayment(ctx, email, isSub, session.ID)
	// if err != nil {
	// 	slog.Error("paymentservice:service:handleCheckoutSessionCompleted", "error", "failed to save user to database", "details", err)
	// 	return fmt.Errorf("error saving user to DB: %v", err)
	// }

	slog.Info("paymentservice:service:handleCheckoutSessionCompleted", "status", "user saved successfully")
	return nil
}

func (s *PaymentService) handleInvoicePaymentFailed(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	err := json.Unmarshal(event.Data.Raw, &invoice)
	if err != nil {
		slog.Error("paymentservice:service:handleInvoicePaymentFailed", "error", "failed to parse webhook JSON", "details", err)
		return fmt.Errorf("error parsing webhook JSON: %v", err)
	}

	slog.Info("paymentservice:service:handleInvoicePaymentFailed", "invoiceID", invoice.ID, "customerID", invoice.Customer.ID)

	// Handle payment failure logic here
	// You might want to update user status, send notification, etc.
	// err = s.dao.HandlePaymentFailure(ctx, invoice.Customer.ID, invoice.ID)
	// if err != nil {
	// 	slog.Error("paymentservice:service:handleInvoicePaymentFailed", "error", "failed to handle payment failure", "details", err)
	// 	return fmt.Errorf("error handling payment failure: %v", err)
	// }

	return nil
}

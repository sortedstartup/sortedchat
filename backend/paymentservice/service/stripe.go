package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/checkout/session"
	"github.com/stripe/stripe-go/v83/price"
	"github.com/stripe/stripe-go/v83/product"
	"github.com/stripe/stripe-go/v83/webhook"
)

func (s *PaymentService) CreateProductStripe(ctx context.Context, name string, description string, amountInSmallestUnit int64, currency string) (string, error) {
	slog.Info("paymentservice:service:CreateProductStripe", "name", name)

	// Create the product
	productParams := &stripe.ProductParams{
		Name:        stripe.String(name),
		Description: stripe.String(description),
		DefaultPriceData: &stripe.ProductDefaultPriceDataParams{
			Currency:   stripe.String(currency),
			UnitAmount: stripe.Int64(amountInSmallestUnit),
		},
	}

	stripeProduct, err := product.New(productParams)
	if err != nil {
		slog.Error("paymentservice:service:CreateProductStripe", "error", err)
		return "", fmt.Errorf("failed to process the request")
	}

	slog.Info("paymentservice:service:CreateProductStripe", "id", stripeProduct.ID)
	return stripeProduct.ID, nil
}

func (s *PaymentService) CreateStripeCheckoutSession(ctx context.Context, userID string, productID string) (string, error) {
	slog.Info("paymentservice:service:CreateCheckoutSession", "userID", userID, "productID", productID)

	// Get product by product ID to get the Stripe product ID
	product, err := s.dao.GetProductById(productID)
	if err != nil {
		slog.Error("paymentservice:service:CreateStripeCheckoutSession", "error", err)
		return "", fmt.Errorf("failed to create Stripe checkout session")
	}

	var priceID string
	//lets get price id from stripe product id
	params := &stripe.PriceListParams{
		Product: stripe.String(product.StripeProductID),
		Active:  stripe.Bool(true),
	}

	i := price.List(params)
	if i.Next() {
		priceID = i.Price().ID
		slog.Info("paymentservice:service:CreateCheckoutSession", "priceFound", priceID)
	} else {
		slog.Error("paymentservice:service:CreateCheckoutSession", "error", "no active prices found for product")
		return "", fmt.Errorf("failed to process the request")
	}

	if err := i.Err(); err != nil {
		slog.Error("paymentservice:service:CreateCheckoutSession", "error", err)
		return "", fmt.Errorf("failed to process the request")
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
		SuccessURL: stripe.String(os.Getenv("FRONTEND_URL") + "/success"),
		CancelURL:  stripe.String(os.Getenv("FRONTEND_URL") + "/cancel"),
		Metadata:   map[string]string{"user_id": userID, "product_id": productID},
	}

	session, err := session.New(sessionParams)
	if err != nil {
		slog.Error("paymentservice:service:CreateCheckoutSession", "error", err)
		return "", fmt.Errorf("failed to process the request")
	}

	return session.URL, nil
}

func (s *PaymentService) HandleStripeWebhook(ctx context.Context, r *http.Request) error {
	slog.Info("paymentservice:service:HandleWebhook")

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("paymentservice:service:HandleWebhook", "error", "failed to read request body", "details", err)
		return fmt.Errorf("error reading request body")
	}

	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	sigHeader := r.Header.Get("Stripe-Signature")

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

	case "checkout.session.expired":
		err := s.handlePaymentFailed(ctx, event)
		if err != nil {
			slog.Error("paymentservice:service:HandleWebhook", "error", "failed to handle checkout session expired", "details", err)
			return fmt.Errorf("failed to handle checkout session expired: %v", err)
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

	userID, exists := session.Metadata["user_id"]
	if !exists {
		return fmt.Errorf("user_id not found in session metadata")
	}

	productID, exists := session.Metadata["product_id"]
	if !exists {
		return fmt.Errorf("product_id not found in session metadata")
	}

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		slog.Error("paymentservice:service:handleCheckoutSessionCompleted", "error", "failed to marshal session to JSON", "details", err)
		return fmt.Errorf("failed to marshal session to JSON: %v", err)
	}

	_, err = s.dao.CreateUserPurchase(userID, productID, string(sessionJSON), true, "stripe")
	if err != nil {
		slog.Error("paymentservice:service:handleCheckoutSessionCompleted", "error", "failed to create user purchase", "details", err)
		return fmt.Errorf("failed to create user purchase: %v", err)
	}

	slog.Info("paymentservice:service:handleCheckoutSessionCompleted", "userID", userID, "productID", productID, "data", "saved in transaction in database")

	return nil
}

func (s *PaymentService) handlePaymentFailed(ctx context.Context, event stripe.Event) error {
	var session stripe.CheckoutSession
	err := json.Unmarshal(event.Data.Raw, &session)
	if err != nil {
		slog.Error("paymentservice:service:handlePaymentFailed", "error", "failed to parse webhook JSON", "details", err)
		return fmt.Errorf("error parsing webhook JSON: %v", err)
	}

	userID, exists := session.Metadata["user_id"]
	if !exists {
		return fmt.Errorf("user_id not found in session metadata")
	}

	productID, exists := session.Metadata["product_id"]
	if !exists {
		return fmt.Errorf("product_id not found in session metadata")
	}

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		slog.Error("paymentservice:service:handlePaymentFailed", "error", "failed to marshal session to JSON", "details", err)
		return fmt.Errorf("failed to marshal session to JSON: %v", err)
	}

	_, err = s.dao.CreateUserPurchase(userID, productID, string(sessionJSON), false, "stripe")
	if err != nil {
		slog.Error("paymentservice:service:handlePaymentFailed", "error", "failed to create user purchase", "details", err)
		return fmt.Errorf("failed to create user purchase: %v", err)
	}

	return nil
}

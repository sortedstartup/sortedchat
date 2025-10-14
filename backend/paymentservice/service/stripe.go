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
	"github.com/stripe/stripe-go/v83/webhook"
)

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

func (s *PaymentService) HandleWebhook(ctx context.Context, r *http.Request) error {
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

	userID, exists := session.Metadata["user_id"]
	if !exists {
		return fmt.Errorf("user_id not found in session metadata")
	}

	productID, exists := session.Metadata["product_id"]
	if !exists {
		return fmt.Errorf("plan_id not found in session metadata")
	}

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		slog.Error("paymentservice:service:handleCheckoutSessionCompleted", "error", "failed to marshal session to JSON", "details", err)
		return fmt.Errorf("failed to marshal session to JSON: %v", err)
	}

	_, err = s.dao.CreateUserPurchase(userID, productID, string(sessionJSON), true)
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
		return fmt.Errorf("plan_id not found in session metadata")
	}

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		slog.Error("paymentservice:service:handlePaymentFailed", "error", "failed to marshal session to JSON", "details", err)
		return fmt.Errorf("failed to marshal session to JSON: %v", err)
	}

	_, err = s.dao.CreateUserPurchase(userID, productID, string(sessionJSON), false)
	if err != nil {
		slog.Error("paymentservice:service:handlePaymentFailed", "error", "failed to create user purchase", "details", err)
		return fmt.Errorf("failed to create user purchase: %v", err)
	}

	return nil
}

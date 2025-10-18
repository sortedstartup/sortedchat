package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/checkout/session"
	"github.com/stripe/stripe-go/v83/price"
	"github.com/stripe/stripe-go/v83/product"
	"github.com/stripe/stripe-go/v83/webhook"
)

func (s *PaymentService) CreateProductStripe(ctx context.Context, name string, description string, amountInSmallestUnit int64, currency string, isRecurring bool, intervalCount int64, interval string) (string, error) {
	slog.Info("paymentservice:service:CreateProductStripe", "name", name, "isRecurring", isRecurring, "intervalCount", intervalCount, "interval", interval)

	// Create the product with appropriate pricing
	productParams := &stripe.ProductParams{
		Name:        stripe.String(name),
		Description: stripe.String(description),
	}

	// Handle recurring vs one-time payments
	if isRecurring {
		productParams.DefaultPriceData = &stripe.ProductDefaultPriceDataParams{
			Currency:   stripe.String(currency),
			UnitAmount: stripe.Int64(amountInSmallestUnit),
			Recurring: &stripe.ProductDefaultPriceDataRecurringParams{
				Interval:      stripe.String(interval),     // "day", "week", "month", or "year"
				IntervalCount: stripe.Int64(intervalCount), // e.g., 1 for every month, 3 for every 3 months
			},
		}
	} else {
		productParams.DefaultPriceData = &stripe.ProductDefaultPriceDataParams{
			Currency:   stripe.String(currency),
			UnitAmount: stripe.Int64(amountInSmallestUnit),
		}
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

	frontendURL := os.Getenv("FRONTEND_URL")
	if strings.TrimSpace(frontendURL) == "" {
		slog.Error("paymentservice:service:CreateCheckoutSession", "error", "FRONTEND_URL is not set")
		return "", fmt.Errorf("configuration error")
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
		SuccessURL: stripe.String(frontendURL + "/success"),
		CancelURL:  stripe.String(frontendURL + "/cancel"),
		Metadata:   map[string]string{"user_id": userID, "product_id": productID},
	}

	session, err := session.New(sessionParams)
	if err != nil {
		slog.Error("paymentservice:service:CreateCheckoutSession", "error", err)
		return "", fmt.Errorf("failed to process the request")
	}

	return session.URL, nil
}

func (s *PaymentService) CreateStripeSubscriptionCheckoutSession(ctx context.Context, userID string, productID string) (string, error) {
	slog.Info("paymentservice:service:CreateStripeSubscriptionCheckoutSession", "userID", userID, "productID", productID)

	// Get product by product ID to get the Stripe product ID
	product, err := s.dao.GetProductById(productID)
	if err != nil {
		slog.Error("paymentservice:service:CreateStripeSubscriptionCheckoutSession", "error", err)
		return "", fmt.Errorf("failed to create Stripe subscription checkout session")
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
		slog.Info("paymentservice:service:CreateStripeSubscriptionCheckoutSession", "priceFound", priceID)
	} else {
		slog.Error("paymentservice:service:CreateStripeSubscriptionCheckoutSession", "error", "no active prices found for product")
		return "", fmt.Errorf("failed to process the request")
	}

	if err := i.Err(); err != nil {
		slog.Error("paymentservice:service:CreateStripeSubscriptionCheckoutSession", "error", err)
		return "", fmt.Errorf("failed to process the request")
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if strings.TrimSpace(frontendURL) == "" {
		slog.Error("paymentservice:service:CreateStripeSubscriptionCheckoutSession", "error", "FRONTEND_URL is not set")
		return "", fmt.Errorf("configuration error")
	}

	sessionParams := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID), // Use price_id, not product_id
				Quantity: stripe.Int64(1),
			},
		},
		Mode:              stripe.String("subscription"), // Key difference: subscription mode
		SuccessURL:        stripe.String(frontendURL + "/success"),
		CancelURL:         stripe.String(frontendURL + "/cancel"),
		ClientReferenceID: stripe.String(userID),
		Metadata:          map[string]string{"user_id": userID, "product_id": productID},
	}

	session, err := session.New(sessionParams)
	if err != nil {
		slog.Error("paymentservice:service:CreateStripeSubscriptionCheckoutSession", "error", err)
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

	// Subscription events
	case "customer.subscription.created":
		slog.Info("paymentservice:service:HandleWebhooksanskar", "event", "customer.subscription.created", "data", event.Data.Raw)
		err := s.handleSubscriptionCreated(ctx, event)
		if err != nil {
			slog.Error("paymentservice:service:HandleWebhook", "error", "failed to handle subscription created", "details", err)
			return fmt.Errorf("failed to handle subscription created: %v", err)
		}

	case "customer.subscription.updated":
		err := s.handleSubscriptionUpdated(ctx, event)
		if err != nil {
			slog.Error("paymentservice:service:HandleWebhook", "error", "failed to handle subscription updated", "details", err)
			return fmt.Errorf("failed to handle subscription updated: %v", err)
		}

	case "customer.subscription.deleted":
		err := s.handleSubscriptionDeleted(ctx, event)
		if err != nil {
			slog.Error("paymentservice:service:HandleWebhook", "error", "failed to handle subscription deleted", "details", err)
			return fmt.Errorf("failed to handle subscription deleted: %v", err)
		}

	case "invoice.paid":
		err := s.handleInvoicePaid(ctx, event)
		if err != nil {
			slog.Error("paymentservice:service:HandleWebhook", "error", "failed to handle invoice paid", "details", err)
			return fmt.Errorf("failed to handle invoice paid: %v", err)
		}

	case "invoice.payment_failed":
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

	_, err = s.dao.CreateUserPurchase(session.ID, userID, productID, string(sessionJSON), true, "stripe")
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

	_, err = s.dao.CreateUserPurchase(session.ID, userID, productID, string(sessionJSON), false, "stripe")
	if err != nil {
		slog.Error("paymentservice:service:handlePaymentFailed", "error", "failed to create user purchase", "details", err)
		return fmt.Errorf("failed to create user purchase: %v", err)
	}

	return nil
}

// New subscription webhook handlers
func (s *PaymentService) handleSubscriptionCreated(ctx context.Context, event stripe.Event) error {
	var subscription stripe.Subscription
	err := json.Unmarshal(event.Data.Raw, &subscription)
	if err != nil {
		slog.Error("paymentservice:service:handleSubscriptionCreated", "error", "failed to parse webhook JSON", "details", err)
		return fmt.Errorf("error parsing webhook JSON: %v", err)
	}

	// For now, we'll log the subscription creation
	// In a real implementation, you might want to find the subscription by metadata or other means
	slog.Info("paymentservice:service:handleSubscriptionCreated", "subscriptionID", subscription.ID, "status", subscription.Status)

	// TODO: Update subscription record in database with provider_subscription_id and status
	// This would require finding the subscription by some correlation (e.g., customer metadata)

	return nil
}

func (s *PaymentService) handleSubscriptionUpdated(ctx context.Context, event stripe.Event) error {
	var subscription stripe.Subscription
	err := json.Unmarshal(event.Data.Raw, &subscription)
	if err != nil {
		slog.Error("paymentservice:service:handleSubscriptionUpdated", "error", "failed to parse webhook JSON", "details", err)
		return fmt.Errorf("error parsing webhook JSON: %v", err)
	}

	slog.Info("paymentservice:service:handleSubscriptionUpdated", "subscriptionID", subscription.ID, "status", subscription.Status)

	// TODO: Update subscription status in database

	return nil
}

func (s *PaymentService) handleSubscriptionDeleted(ctx context.Context, event stripe.Event) error {
	var subscription stripe.Subscription
	err := json.Unmarshal(event.Data.Raw, &subscription)
	if err != nil {
		slog.Error("paymentservice:service:handleSubscriptionDeleted", "error", "failed to parse webhook JSON", "details", err)
		return fmt.Errorf("error parsing webhook JSON: %v", err)
	}

	slog.Info("paymentservice:service:handleSubscriptionDeleted", "subscriptionID", subscription.ID)

	// TODO: Mark subscription as cancelled in database

	return nil
}

func (s *PaymentService) handleInvoicePaid(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	err := json.Unmarshal(event.Data.Raw, &invoice)
	if err != nil {
		slog.Error("paymentservice:service:handleInvoicePaid", "error", "failed to parse webhook JSON", "details", err)
		return fmt.Errorf("error parsing webhook JSON: %v", err)
	}

	slog.Info("paymentservice:service:handleInvoicePaid", "invoiceID", invoice.ID, "amount", invoice.AmountPaid)

	// TODO: Create user_payment record for this invoice payment

	return nil
}

func (s *PaymentService) handleInvoicePaymentFailed(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	err := json.Unmarshal(event.Data.Raw, &invoice)
	if err != nil {
		slog.Error("paymentservice:service:handleInvoicePaymentFailed", "error", "failed to parse webhook JSON", "details", err)
		return fmt.Errorf("error parsing webhook JSON: %v", err)
	}

	slog.Info("paymentservice:service:handleInvoicePaymentFailed", "invoiceID", invoice.ID)

	// TODO: Handle failed payment (maybe update subscription status, send notification, etc.)

	return nil
}

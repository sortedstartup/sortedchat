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
		Mode:       stripe.String("payment"),
		SuccessURL: stripe.String(frontendURL + "/success"),
		CancelURL:  stripe.String(frontendURL + "/cancel"),
		Metadata:   map[string]string{"user_id": userID, "product_id": productID},
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: map[string]string{
				"user_id":    userID,
				"product_id": productID,
			},
		},
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
		// Set customer metadata when customer is created
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{"user_id": userID, "product_id": productID},
		},
	}

	session, err := session.New(sessionParams)
	if err != nil {
		slog.Error("paymentservice:service:CreateStripeSubscriptionCheckoutSession", "error", err)
		return "", fmt.Errorf("failed to process the request")
	}

	return session.URL, nil
}

func (s *PaymentService) HandleStripeWebhook(ctx context.Context, r *http.Request) error {
	slog.Info("paymentservice:service:HandleStripeWebhook")

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("paymentservice:service:HandleStripeWebhook", "error", "failed to read request body", "details", err)
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

	slog.Info("paymentservice:service:HandleStripeWebhook", "event", event.Type)

	switch event.Type {

	case "charge.succeeded":
		err := s.handleChargeSucceeded(ctx, event)
		if err != nil {
			slog.Error("paymentservice:service:HandleStripeWebhook", "error", "failed to handle charge succeeded", "details", err)
			return fmt.Errorf("failed to handle charge succeeded: %v", err)
		}

	case "checkout.session.completed":
		err := s.handleCheckoutSessionCompleted(ctx, event)
		if err != nil {
			slog.Error("paymentservice:service:HandleStripeWebhook", "error", "failed to handle checkout session completed", "details", err)
			return fmt.Errorf("failed to handle checkout session completed: %v", err)
		}

	case "checkout.session.expired":
		err := s.handlePaymentFailed(ctx, event)
		if err != nil {
			slog.Error("paymentservice:service:HandleStripeWebhook", "error", "failed to handle checkout session expired", "details", err)
			return fmt.Errorf("failed to handle checkout session expired: %v", err)
		}

	// Subscription events
	case "customer.subscription.created":
		slog.Info("paymentservice:service:HandleStripeWebhook", "event", "customer.subscription.created", "data", event.Data.Raw)
		err := s.handleSubscriptionCreated(ctx, event)
		if err != nil {
			slog.Error("paymentservice:service:HandleWebhook", "error", "failed to handle subscription created", "details", err)
			return fmt.Errorf("failed to handle subscription created: %v", err)
		}

	case "invoice.paid":
		err := s.handleInvoicePaid(ctx, event)
		if err != nil {
			slog.Error("paymentservice:service:HandleStripeWebhook", "error", "failed to handle invoice paid", "details", err)
			return fmt.Errorf("failed to handle invoice paid: %v", err)
		}

	case "invoice.payment_failed":
		err := s.handleInvoicePaymentFailed(ctx, event)
		if err != nil {
			slog.Error("paymentservice:service:HandleStripeWebhook", "error", "failed to handle invoice payment failed", "details", err)
			return fmt.Errorf("failed to handle invoice payment failed: %v", err)
		}

	default:
		slog.Info("paymentservice:service:HandleStripeWebhook", "event", "unhandled event type", "type", event.Type)
	}

	slog.Info("paymentservice:service:HandleStripeWebhook", "status", "webhook processed successfully")
	return nil
}

// this is for one-time payments
func (s *PaymentService) handleChargeSucceeded(ctx context.Context, event stripe.Event) error {
	var charge stripe.Charge
	err := json.Unmarshal(event.Data.Raw, &charge)
	if err != nil {
		slog.Error("paymentservice:service:handleChargeSucceeded", "error", "failed to parse webhook JSON", "details", err)
		return fmt.Errorf("error parsing webhook JSON: %v", err)
	}

	userID, exists := charge.Metadata["user_id"]
	if !exists {
		return fmt.Errorf("user_id not found in charge metadata")
	}

	productID, exists := charge.Metadata["product_id"]
	if !exists {
		return fmt.Errorf("product_id not found in charge metadata")
	}

	// For one-time payments, create subscription with period end < period start to indicate it's expired/one-time
	currentTime := charge.Created // Use charge creation time
	periodStart := currentTime
	periodEnd := currentTime - 1 // Set end time to be less than start time to indicate one-time payment

	// Create subscription record for one-time payment
	subscriptionID, err := s.dao.CreateSubscription(
		userID,
		productID,
		"stripe",    // provider
		"",          // provider_subscription_id - empty for one-time payments
		"",          // provider_customer_id - might be empty if no customer
		"",          // provider_subscription_status - empty for one-time payments
		"active",    // status - active since payment succeeded
		periodStart, // current_period_start
		periodEnd,   // current_period_end - less than start to indicate one-time
		false,       // cancel_at_period_end - false for one-time payments
	)
	if err != nil {
		slog.Error("paymentservice:service:handleChargeSucceeded", "error", "failed to create subscription", "details", err)
		return fmt.Errorf("failed to create subscription: %v", err)
	}

	// Create user_payment record for the one-time payment
	chargeJSON, err := json.Marshal(charge)
	if err != nil {
		slog.Error("paymentservice:service:handleChargeSucceeded", "error", "failed to marshal charge to JSON", "details", err)
		return fmt.Errorf("failed to marshal charge to JSON: %v", err)
	}

	_, err = s.dao.CreateUserPayment(userID, productID, subscriptionID, charge.ID, string(chargeJSON))
	if err != nil {
		slog.Error("paymentservice:service:handleChargeSucceeded", "error", "failed to create user payment", "details", err)
		return fmt.Errorf("failed to create user payment: %v", err)
	}

	slog.Info("paymentservice:service:handleChargeSucceeded", "subscriptionID", subscriptionID, "chargeID", charge.ID, "amount", charge.Amount, "userID", userID, "productID", productID)

	return nil
}

func (s *PaymentService) handleCheckoutSessionCompleted(ctx context.Context, event stripe.Event) error {
	var session stripe.CheckoutSession
	err := json.Unmarshal(event.Data.Raw, &session)
	if err != nil {
		slog.Error("paymentservice:service:handleCheckoutSessionCompleted", "error", "failed to parse webhook JSON", "details", err)
		return fmt.Errorf("error parsing webhook JSON: %v", err)
	}

	slog.Info("stripe:service:handleCheckoutSessionCompleted", "data", "checkout session completed")

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

// New subscription webhook handlers for recurring payments
func (s *PaymentService) handleSubscriptionCreated(ctx context.Context, event stripe.Event) error {
	var subscription stripe.Subscription
	err := json.Unmarshal(event.Data.Raw, &subscription)
	if err != nil {
		slog.Error("paymentservice:service:handleSubscriptionCreated", "error", "failed to parse webhook JSON", "details", err)
		return fmt.Errorf("error parsing webhook JSON: %v", err)
	}

	userID, userExists := subscription.Metadata["user_id"]
	if !userExists {
		slog.Error("paymentservice:service:handleSubscriptionCreated", "error", "user_id not found in subscription metadata")
		return fmt.Errorf("user_id not found in subscription metadata")
	}

	productID, productExists := subscription.Metadata["product_id"]
	if !productExists {
		slog.Error("paymentservice:service:handleSubscriptionCreated", "error", "product_id not found in subscription metadata")
		return fmt.Errorf("product_id not found in subscription metadata")
	}

	// Extract period information from subscription items
	var currentPeriodStart, currentPeriodEnd int64
	if len(subscription.Items.Data) > 0 {
		// Use Unix timestamps directly
		currentPeriodStart = subscription.Items.Data[0].CurrentPeriodStart
		currentPeriodEnd = subscription.Items.Data[0].CurrentPeriodEnd

		slog.Info("paymentservice:service:handleSubscriptionCreated", "currentPeriodStart", currentPeriodStart, "currentPeriodEnd", currentPeriodEnd)
	}

	// Create subscription record in our database with all details
	subscriptionID, err := s.dao.CreateSubscription(
		userID,
		productID,
		"stripe",                    // provider
		subscription.ID,             // provider_subscription_id
		subscription.Customer.ID,    // provider_customer_id
		string(subscription.Status), // provider_subscription_status
		"active",                    // status - set to active since subscription is created
		currentPeriodStart,          // current_period_start from subscription items
		currentPeriodEnd,            // current_period_end from subscription items
		subscription.CancelAtPeriodEnd,
	)
	if err != nil {
		slog.Error("paymentservice:service:handleSubscriptionCreated", "error", "failed to create subscription", "details", err)
		return fmt.Errorf("failed to create subscription: %v", err)
	}

	slog.Info("paymentservice:service:handleSubscriptionCreated", "subscriptionID", subscriptionID, "userID", userID, "productID", productID, "stripeSubscriptionID", subscription.ID, "currentPeriodStart", currentPeriodStart, "currentPeriodEnd", currentPeriodEnd, "CustomerID", subscription.Customer.ID)
	return nil
}

// this is for recurring payments
func (s *PaymentService) handleInvoicePaid(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	err := json.Unmarshal(event.Data.Raw, &invoice)
	if err != nil {
		slog.Error("paymentservice:service:handleInvoicePaid", "error", "failed to parse webhook JSON", "details", err)
		return fmt.Errorf("error parsing webhook JSON: %v", err)
	}

	slog.Info("paymentservice:service:handleInvoicePaid", "invoiceID", invoice.ID, "amount", invoice.AmountPaid, "customerID", invoice.Customer.ID)

	// If this invoice is for a subscription, update the subscription period dates
	if invoice.Customer != nil {

		providerCustomerID := invoice.Customer.ID
		if providerCustomerID == "" {
			slog.Error("paymentservice:service:handleInvoicePaid", "error", "customerID not found in invoice")
			return fmt.Errorf("customerID not found in invoice")
		}
		// Find our internal subscription by provider_subscription_id
		subscription, err := s.dao.GetSubscriptionByProviderID(providerCustomerID)
		if err != nil {
			slog.Error("paymentservice:service:handleInvoicePaid", "error", "failed to find subscription", "details", err)
			return fmt.Errorf("failed to find subscription: %v", err)
		}

		// Use invoice period timestamps directly
		currentPeriodStart := invoice.PeriodStart
		currentPeriodEnd := invoice.PeriodEnd

		// Update subscription with period dates and set status to active
		err = s.dao.UpdateSubscription(
			subscription.ID,
			subscription.ProviderSubscriptionID,     // keep existing provider_subscription_id
			subscription.ProviderCustomerID,         // keep existing provider_customer_id
			subscription.ProviderSubscriptionStatus, // keep existing provider_subscription_status
			"active",                                // set status to active since invoice is paid
			currentPeriodStart,
			currentPeriodEnd,
			subscription.CancelAtPeriodEnd, // keep existing cancel_at_period_end
		)
		if err != nil {
			slog.Error("paymentservice:service:handleInvoicePaid", "error", "failed to update subscription", "details", err)
			return fmt.Errorf("failed to update subscription: %v", err)
		}

		slog.Info("paymentservice:service:handleInvoicePaid", "status", "subscription updated", "subscriptionID", subscription.ID, "currentPeriodStart", currentPeriodStart, "currentPeriodEnd", currentPeriodEnd)

		// Create user_payment record for this invoice payment
		invoiceJSON, err := json.Marshal(invoice)
		if err != nil {
			slog.Error("paymentservice:service:handleInvoicePaid", "error", "failed to marshal invoice to JSON", "details", err)
			return fmt.Errorf("failed to marshal invoice to JSON: %v", err)
		}

		_, err = s.dao.CreateUserPayment(subscription.UserID, subscription.ProductID, subscription.ID, invoice.ID, string(invoiceJSON))
		if err != nil {
			slog.Error("paymentservice:service:handleInvoicePaid", "error", "failed to create user payment", "details", err)
			return fmt.Errorf("failed to create user payment: %v", err)
		}
	}

	return nil
}

// this is for recurring payments
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

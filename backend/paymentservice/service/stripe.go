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
	slog.Info("paymentservice:stripe:CreateProductStripe", "name", name, "isRecurring", isRecurring, "intervalCount", intervalCount, "interval", interval)

	// Create the product with appropriate pricing
	productParams := &stripe.ProductParams{
		Name:        stripe.String(name),
		Description: stripe.String(description),
	}

	if isRecurring {
		// Normalize unsupported Stripe interval "quarter" -> month*3
		normalizedInterval := interval
		normalizedCount := intervalCount
		if strings.EqualFold(interval, "quarter") {
			normalizedInterval = "month"
			normalizedCount = intervalCount * 3
		}
		interval = normalizedInterval
		intervalCount = normalizedCount
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
		slog.Error("paymentservice:stripe:CreateProductStripe", "error", err)
		return "", fmt.Errorf("failed to process the request")
	}

	slog.Info("paymentservice:stripe:CreateProductStripe", "id", stripeProduct.ID)
	return stripeProduct.ID, nil
}

func (s *PaymentService) CreateStripeCheckoutSession(ctx context.Context, userID string, productID string) (string, error) {
	slog.Info("paymentservice:stripe:CreateStripeCheckoutSession", "userID", userID, "productID", productID)

	hasAccess, err := s.dao.CheckUserProductAccess(userID, productID)
	if err != nil {
		slog.Error("paymentservice:stripe:CreateStripeCheckoutSession", "error", err)
		return "", fmt.Errorf("failed to process the request")
	}

	if hasAccess {
		slog.Error("paymentservice:stripe:CreateStripeCheckoutSession", "error", "user already has access to this product")
		return "", fmt.Errorf("already have access to this product")
	}

	// Get product by product ID to get the Stripe product ID
	product, err := s.dao.GetProductById(productID)
	if err != nil {
		slog.Error("paymentservice:stripe:CreateStripeCheckoutSession", "error", err)
		return "", fmt.Errorf("failed to create Stripe checkout session")
	}

	if product.IsRecurring {
		slog.Error("paymentservice:stripe:CreateStripeCheckoutSession", "error", "cannot create checkout session for a one-time product")
		return "", fmt.Errorf("cannot create checkout session for a subscription product")
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
		slog.Info("paymentservice:stripe:CreateStripeCheckoutSession", "priceFound", priceID)
	} else {
		slog.Error("paymentservice:stripe:CreateStripeCheckoutSession", "error", "no active prices found for product")
		return "", fmt.Errorf("failed to process the request")
	}

	if err := i.Err(); err != nil {
		slog.Error("paymentservice:stripe:CreateStripeCheckoutSession", "error", err)
		return "", fmt.Errorf("failed to process the request")
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if strings.TrimSpace(frontendURL) == "" {
		slog.Error("paymentservice:stripe:CreateStripeCheckoutSession", "error", "FRONTEND_URL is not set")
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
		slog.Error("paymentservice:stripe:CreateStripeCheckoutSession", "error", err)
		return "", fmt.Errorf("failed to process the request")
	}

	return session.URL, nil
}

func (s *PaymentService) CreateStripeSubscriptionCheckoutSession(ctx context.Context, userID string, productID string) (string, error) {
	slog.Info("paymentservice:stripe:CreateStripeSubscriptionCheckoutSession", "userID", userID, "productID", productID)

	hasAccess, err := s.dao.CheckUserProductAccess(userID, productID)
	if err != nil {
		slog.Error("paymentservice:stripe:CreateStripeCheckoutSession", "error", err)
		return "", fmt.Errorf("failed to process the request")
	}

	if hasAccess {
		slog.Error("paymentservice:stripe:CreateStripeSubscriptionCheckoutSession", "error", "user already has access to this product")
		return "", fmt.Errorf("already have access to this product")
	}

	// Get product by product ID to get the Stripe product ID
	product, err := s.dao.GetProductById(productID)
	if err != nil {
		slog.Error("paymentservice:stripe:CreateStripeSubscriptionCheckoutSession", "error", err)
		return "", fmt.Errorf("failed to create Stripe subscription checkout session")
	}

	if !product.IsRecurring {
		slog.Error("paymentservice:stripe:CreateStripeSubscriptionCheckoutSession", "error", "cannot create subscription session for a one-time product")
		return "", fmt.Errorf("cannot create subscription session for a one-time product")
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
		slog.Info("paymentservice:stripe:CreateStripeSubscriptionCheckoutSession", "priceFound", priceID)
	} else {
		slog.Error("paymentservice:stripe:CreateStripeSubscriptionCheckoutSession", "error", "no active prices found for product")
		return "", fmt.Errorf("failed to process the request")
	}

	if err := i.Err(); err != nil {
		slog.Error("paymentservice:stripe:CreateStripeSubscriptionCheckoutSession", "error", err)
		return "", fmt.Errorf("failed to process the request")
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if strings.TrimSpace(frontendURL) == "" {
		slog.Error("paymentservice:stripe:CreateStripeSubscriptionCheckoutSession", "error", "FRONTEND_URL is not set")
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
		slog.Error("paymentservice:stripe:CreateStripeSubscriptionCheckoutSession", "error", err)
		return "", fmt.Errorf("failed to process the request")
	}

	return session.URL, nil
}

func (s *PaymentService) HandleStripeWebhook(ctx context.Context, r *http.Request) error {
	slog.Info("paymentservice:stripe:HandleStripeWebhook")

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("paymentservice:stripe:HandleStripeWebhook", "error", "failed to read request body", "details", err)
		return fmt.Errorf("error reading request body")
	}

	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	sigHeader := r.Header.Get("Stripe-Signature")

	event, err := webhook.ConstructEventWithOptions(payload, sigHeader, endpointSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		slog.Error("paymentservice:stripe:HandleWebhook", "error", "signature verification failed", "details", err)
		return fmt.Errorf("webhook signature verification failed: %v", err)
	}

	slog.Info("paymentservice:stripe:HandleStripeWebhook", "event", event.Type)

	switch event.Type {

	case "charge.succeeded":
		err := s.handleChargeSucceeded(ctx, event)
		if err != nil {
			slog.Error("paymentservice:stripe:HandleStripeWebhook", "error", "failed to handle charge succeeded", "details", err)
			return fmt.Errorf("failed to handle charge succeeded: %v", err)
		}
	case "charge.failed":
		err := s.handleChargeFailed(ctx, event)
		if err != nil {
			slog.Error("paymentservice:stripe:HandleStripeWebhook", "error", "failed to handle charge failed", "details", err)
			return fmt.Errorf("failed to handle charge failed: %v", err)
		}

	// Subscription events
	case "customer.subscription.created":
		slog.Info("paymentservice:stripe:HandleStripeWebhook", "event", "customer.subscription.created")
		err := s.handleSubscriptionCreated(ctx, event)
		if err != nil {
			slog.Error("paymentservice:stripe:HandleWebhook", "error", "failed to handle subscription created", "details", err)
			return fmt.Errorf("failed to handle subscription created: %v", err)
		}

	case "customer.subscription.updated":
		err := s.handleSubscriptionUpdated(ctx, event)
		if err != nil {
			slog.Error("paymentservice:stripe:HandleStripeWebhook", "error", "failed to handle subscription updated", "details", err)
			return fmt.Errorf("failed to handle subscription updated: %v", err)
		}

	case "invoice.paid":
		err := s.handleInvoicePaid(ctx, event)
		if err != nil {
			slog.Error("paymentservice:stripe:HandleStripeWebhook", "error", "failed to handle invoice paid", "details", err)
			return fmt.Errorf("failed to handle invoice paid: %v", err)
		}

	case "invoice.payment_failed":
		err := s.handleInvoicePaymentFailed(ctx, event)
		if err != nil {
			slog.Error("paymentservice:stripe:HandleStripeWebhook", "error", "failed to handle invoice payment failed", "details", err)
			return fmt.Errorf("failed to handle invoice payment failed: %v", err)
		}

	default:
		slog.Info("paymentservice:stripe:HandleStripeWebhook", "event", "unhandled event type", "type", event.Type)
	}

	slog.Info("paymentservice:stripe:HandleStripeWebhook", "status", "webhook processed successfully")
	return nil
}

// this is for one-time payments
func (s *PaymentService) handleChargeSucceeded(ctx context.Context, event stripe.Event) error {
	var charge stripe.Charge
	err := json.Unmarshal(event.Data.Raw, &charge)
	if err != nil {
		slog.Error("paymentservice:stripe:handleChargeSucceeded", "error", "failed to parse webhook JSON", "details", err)
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
	periodEnd := currentTime // Set end time to be less than start time to indicate one-time payment

	eventID := charge.ID // use charge ID as event ID to avoid duplicate subscriptions

	product, err := s.dao.GetProductById(productID)
	if err != nil {
		slog.Error("paymentservice:stripe:handleChargeSucceeded", "error", "failed to get product", "details", err)
		return fmt.Errorf("failed to get product: %v", err)
	}

	// Create subscription record for one-time payment
	subscriptionID, err := s.dao.CreateSubscription(
		eventID,
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
		product.IsRecurring,
	)
	if err != nil {
		slog.Error("paymentservice:stripe:handleChargeSucceeded", "error", "failed to create subscription", "details", err)
		return fmt.Errorf("failed to create subscription: %v", err)
	}

	// Create user_payment record for the one-time payment
	chargeJSON, err := json.Marshal(charge)
	if err != nil {
		slog.Error("paymentservice:stripe:handleChargeSucceeded", "error", "failed to marshal charge to JSON", "details", err)
		return fmt.Errorf("failed to marshal charge to JSON: %v", err)
	}

	_, err = s.dao.CreateUserPayment(userID, product.ID, subscriptionID, charge.ID, string(chargeJSON), true)
	if err != nil {
		slog.Error("paymentservice:stripe:handleChargeSucceeded", "error", "failed to create user payment", "details", err)
		return fmt.Errorf("failed to create user payment: %v", err)
	}

	slog.Info("paymentservice:stripe:handleChargeSucceeded", "subscriptionID", subscriptionID, "chargeID", charge.ID, "amount", charge.Amount, "userID", userID, "productID", productID)

	return nil
}

func (s *PaymentService) handleChargeFailed(ctx context.Context, event stripe.Event) error {
	var charge stripe.Charge
	err := json.Unmarshal(event.Data.Raw, &charge)
	if err != nil {
		slog.Error("paymentservice:stripe:handleChargeFailed", "error", "failed to parse webhook JSON", "details", err)
		return fmt.Errorf("error parsing webhook JSON: %v", err)
	}
	chargeJSON, err := json.Marshal(charge)
	if err != nil {
		slog.Error("paymentservice:stripe:handleChargeFailed", "error", "failed to marshal charge to JSON", "details", err)
		return fmt.Errorf("failed to marshal charge to JSON: %v", err)
	}

	userID, exists := charge.Metadata["user_id"]
	if !exists {
		return fmt.Errorf("user_id not found in charge metadata")
	}

	productID, exists := charge.Metadata["product_id"]
	if !exists {
		return fmt.Errorf("product_id not found in charge metadata")
	}

	if charge.Customer != nil {

		providerCustomerID := charge.Customer.ID
		if providerCustomerID == "" {
			slog.Error("paymentservice:stripe:handleChargeFailed", "error", "provider_customer_id not found in charge metadata")
			return fmt.Errorf("provider_customer_id not found in charge metadata")
		}

		subscription, err := s.dao.GetSubscriptionByProviderCustomerID(providerCustomerID)
		if err != nil {
			slog.Warn("paymentservice:service:handleChargeFailed", "warning", "subscription not found, creating payment without subscription reference", "details", err)
			// for one-time payments, create user payment without subscription reference
			_, err = s.dao.CreateUserPayment(userID, productID, "", charge.ID, string(chargeJSON), false)
			if err != nil {
				slog.Error("paymentservice:service:handleChargeFailed", "error", "failed to create user payment", "details", err)
				return fmt.Errorf("failed to create user payment: %v", err)
			}
			return nil
		}

		_, err = s.dao.CreateUserPayment(userID, productID, subscription.ID, charge.ID, string(chargeJSON), false)
		if err != nil {
			slog.Error("paymentservice:stripe:handleChargeFailed", "error", "failed to create user payment", "details", err)
			return fmt.Errorf("failed to create user payment: %v", err)
		}
	}

	return nil
}

// New subscription webhook handlers for recurring payments
func (s *PaymentService) handleSubscriptionCreated(ctx context.Context, event stripe.Event) error {
	var subscription stripe.Subscription
	err := json.Unmarshal(event.Data.Raw, &subscription)
	if err != nil {
		slog.Error("paymentservice:stripe:handleSubscriptionCreated", "error", "failed to parse webhook JSON", "details", err)
		return fmt.Errorf("error parsing webhook JSON: %v", err)
	}

	userID, userExists := subscription.Metadata["user_id"]
	if !userExists {
		slog.Error("paymentservice:stripe:handleSubscriptionCreated", "error", "user_id not found in subscription metadata")
		return fmt.Errorf("user_id not found in subscription metadata")
	}

	productID, productExists := subscription.Metadata["product_id"]
	if !productExists {
		slog.Error("paymentservice:stripe:handleSubscriptionCreated", "error", "product_id not found in subscription metadata")
		return fmt.Errorf("product_id not found in subscription metadata")
	}

	product, err := s.dao.GetProductById(productID)
	if err != nil {
		slog.Error("paymentservice:stripe:handleSubscriptionCreated", "error", "failed to get product", "details", err)
		return fmt.Errorf("failed to get product: %v", err)
	}

	// Extract period information from subscription items
	var currentPeriodStart, currentPeriodEnd int64
	if len(subscription.Items.Data) > 0 {
		// Use Unix timestamps directly
		currentPeriodStart = subscription.Items.Data[0].CurrentPeriodStart
		currentPeriodEnd = subscription.Items.Data[0].CurrentPeriodEnd

		slog.Info("paymentservice:stripe:handleSubscriptionCreated", "currentPeriodStart", currentPeriodStart, "currentPeriodEnd", currentPeriodEnd)
	}

	eventID := subscription.ID // use subscription ID as event ID to avoid duplicate subscriptions

	// Create subscription record in our database with all details
	subscriptionID, err := s.dao.CreateSubscription(
		eventID,
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
		product.IsRecurring,
	)
	if err != nil {
		slog.Error("paymentservice:stripe:handleSubscriptionCreated", "error", "failed to create subscription", "details", err)
		return fmt.Errorf("failed to create subscription: %v", err)
	}

	slog.Info("paymentservice:stripe:handleSubscriptionCreated", "subscriptionID", subscriptionID, "userID", userID, "productID", productID, "stripeSubscriptionID", subscription.ID, "currentPeriodStart", currentPeriodStart, "currentPeriodEnd", currentPeriodEnd, "CustomerID", subscription.Customer.ID)
	return nil
}

// this is for recurring payments
func (s *PaymentService) handleSubscriptionUpdated(ctx context.Context, event stripe.Event) error {
	var subscription stripe.Subscription
	err := json.Unmarshal(event.Data.Raw, &subscription)
	if err != nil {
		slog.Error("paymentservice:stripe:handleSubscriptionUpdated", "error", "failed to parse webhook JSON", "details", err)
		return fmt.Errorf("error parsing webhook JSON: %v", err)
	}

	userID, userExists := subscription.Metadata["user_id"]
	if !userExists {
		slog.Error("paymentservice:stripe:handleSubscriptionUpdated", "error", "user_id not found in subscription metadata")
		return fmt.Errorf("user_id not found in subscription metadata")
	}

	productID, productExists := subscription.Metadata["product_id"]
	if !productExists {
		slog.Error("paymentservice:stripe:handleSubscriptionUpdated", "error", "product_id not found in subscription metadata")
		return fmt.Errorf("product_id not found in subscription metadata")
	}

	// Extract period information from subscription items
	var currentPeriodStart, currentPeriodEnd int64
	if len(subscription.Items.Data) > 0 {
		// Use Unix timestamps directly
		currentPeriodStart = subscription.Items.Data[0].CurrentPeriodStart
		currentPeriodEnd = subscription.Items.Data[0].CurrentPeriodEnd

		slog.Info("paymentservice:stripe:handleSubscriptionUpdated", "currentPeriodStart", currentPeriodStart, "currentPeriodEnd", currentPeriodEnd)
	}

	providerCustomerID := subscription.Customer.ID
	if providerCustomerID == "" {
		slog.Error("paymentservice:stripe:handleSubscriptionUpdated", "error", "provider_customer_id not found in subscription metadata")
		return fmt.Errorf("provider_customer_id not found in subscription metadata")
	}

	subscriptionRecord, err := s.dao.GetSubscriptionByProviderCustomerID(providerCustomerID)
	if err != nil {
		slog.Error("paymentservice:stripe:handleSubscriptionUpdated", "error", "failed to find subscription", "details", err)
		return fmt.Errorf("failed to find subscription: %v", err)
	}

	// Update subscription record in our database with all details
	err = s.dao.UpdateSubscription(
		subscriptionRecord.ID,
		subscription.ID,             // keep existing provider_subscription_id
		subscription.Customer.ID,    // keep existing provider_customer_id
		string(subscription.Status), // keep existing provider_subscription_status
		"active",                    // set status to active since subscription is updated
		currentPeriodStart,
		currentPeriodEnd,
		subscription.CancelAtPeriodEnd,
	)
	if err != nil {
		slog.Error("paymentservice:stripe:handleSubscriptionUpdated", "error", "failed to update subscription", "details", err)
		return fmt.Errorf("failed to update subscription: %v", err)
	}

	slog.Info("paymentservice:stripe:handleSubscriptionUpdated", "subscriptionID", subscription.ID, "userID", userID, "productID", productID, "stripeSubscriptionID", subscription.ID, "currentPeriodStart", currentPeriodStart, "currentPeriodEnd", currentPeriodEnd, "CustomerID", subscription.Customer.ID)
	return nil
}

// this is for recurring payments
func (s *PaymentService) handleInvoicePaid(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	err := json.Unmarshal(event.Data.Raw, &invoice)
	if err != nil {
		slog.Error("paymentservice:stripe:handleInvoicePaid", "error", "failed to parse webhook JSON", "details", err)
		return fmt.Errorf("error parsing webhook JSON: %v", err)
	}

	slog.Info("paymentservice:stripe:handleInvoicePaid", "invoiceID", invoice.ID, "amount", invoice.AmountPaid, "customerID", invoice.Customer.ID)

	invoiceJSON, err := json.Marshal(invoice)
	if err != nil {
		slog.Error("paymentservice:stripe:handleInvoicePaid", "error", "failed to marshal invoice to JSON", "details", err)
		return fmt.Errorf("failed to marshal invoice to JSON: %v", err)
	}

	// Get subscription by customer ID to get user and product info
	providerCustomerID := invoice.Customer.ID
	if providerCustomerID == "" {
		slog.Error("paymentservice:stripe:handleInvoicePaid", "error", "provider_customer_id not found in invoice")
		return fmt.Errorf("provider_customer_id not found in invoice")
	}

	subscription, err := s.dao.GetSubscriptionByProviderCustomerID(providerCustomerID)
	if err != nil {
		slog.Error("paymentservice:stripe:handleInvoicePaid", "error", "failed to find subscription", "details", err)
		return fmt.Errorf("failed to find subscription: %v", err)
	}

	_, err = s.dao.CreateUserPayment(subscription.UserID, subscription.ProductID, subscription.ID, invoice.ID, string(invoiceJSON), true)
	if err != nil {
		slog.Error("paymentservice:stripe:handleInvoicePaid", "error", "failed to create user payment", "details", err)
		return fmt.Errorf("failed to create user payment: %v", err)
	}

	return nil
}

// this is for recurring payments
func (s *PaymentService) handleInvoicePaymentFailed(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	err := json.Unmarshal(event.Data.Raw, &invoice)
	if err != nil {
		slog.Error("paymentservice:stripe:handleInvoicePaymentFailed", "error", "failed to parse webhook JSON", "details", err)
		return fmt.Errorf("error parsing webhook JSON: %v", err)
	}

	slog.Info("paymentservice:stripe:handleInvoicePaymentFailed", "invoiceID", invoice.ID)

	var userID, productID, subscriptionID string

	if invoice.Customer != nil {
		providerCustomerID := invoice.Customer.ID
		if providerCustomerID == "" {
			slog.Error("paymentservice:stripe:handleInvoicePaymentFailed", "error", "customerID not found in invoice")
			return fmt.Errorf("customerID not found in invoice")
		}

		subscription, err := s.dao.GetSubscriptionByProviderCustomerID(providerCustomerID)
		if err != nil {
			slog.Error("paymentservice:stripe:handleInvoicePaymentFailed", "error", "failed to find subscription", "details", err)
			return fmt.Errorf("failed to find subscription: %v", err)
		}

		userID = subscription.UserID
		productID = subscription.ProductID
		subscriptionID = subscription.ID

		if userID == "" {
			slog.Error("paymentservice:stripe:handleInvoicePaymentFailed", "error", "userID not found in subscription")
			return fmt.Errorf("userID not found in subscription")
		}

		if productID == "" {
			slog.Error("paymentservice:stripe:handleInvoicePaymentFailed", "error", "productID not found in subscription")
			return fmt.Errorf("productID not found in subscription")
		}
	}

	// Create user_payment record for this invoice payment
	invoiceJSON, err := json.Marshal(invoice)
	if err != nil {
		slog.Error("paymentservice:stripe:handleInvoicePaymentFailed", "error", "failed to marshal invoice to JSON", "details", err)
		return fmt.Errorf("failed to marshal invoice to JSON: %v", err)
	}

	_, err = s.dao.CreateUserPayment(userID, productID, subscriptionID, invoice.ID, string(invoiceJSON), false)
	if err != nil {
		slog.Error("paymentservice:stripe:handleInvoicePaymentFailed", "error", "failed to create user payment", "details", err)
		return fmt.Errorf("failed to create user payment: %v", err)
	}

	return nil
}

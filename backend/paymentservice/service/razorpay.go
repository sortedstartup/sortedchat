package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
)

func (s *PaymentService) CreateProductRazorpay(ctx context.Context, name string, description string, amountInSmallestUnit int64, currency string, isRecurring bool, intervalCount int64, period string) (string, error) {
	slog.Info("paymentservice:razorpay:CreateProductRazorpay", "name", name, "isRecurring", isRecurring, "intervalCount", intervalCount, "period", period, "description", description, "amountInSmallestUnit", amountInSmallestUnit, "currency", currency)

	if isRecurring {

		actualIntervalCount := intervalCount

		planData := map[string]interface{}{
			"period":   period,              // "weekly", "monthly", "quarterly", "yearly"
			"interval": actualIntervalCount, // e.g., 1 for every month, 3 for every 3 months
			"item": map[string]interface{}{
				"name":        name,
				"amount":      amountInSmallestUnit,
				"currency":    currency, // Required field according to API docs
				"description": description,
			},
			"notes": map[string]interface{}{
				"created_by": "sortedstartup",
			},
		}

		slog.Info("paymentservice:service:CreateProductRazorpay", "planData", planData)

		// Create plan in Razorpay
		plan, err := s.razorpayClient.Plan.Create(planData, nil)
		if err != nil {
			slog.Error("paymentservice:service:CreateProductRazorpay", "error", err, "errorType", fmt.Sprintf("%T", err), "planData", planData)
			// Return more specific error information for debugging
			return "", fmt.Errorf("failed to create Razorpay plan: %v", err)
		}

		// Extract plan ID from response
		planID, ok := plan["id"].(string)
		if !ok {
			slog.Error("paymentservice:service:CreateProductRazorpay", "error", "failed to get plan ID", "response", plan)
			return "", fmt.Errorf("failed to process the request")
		}

		slog.Info("paymentservice:service:CreateProductRazorpay", "planID", planID)
		return planID, nil // Return plan ID for recurring payments
	} else {
		// Create item data for one-time payments
		itemData := map[string]interface{}{
			"name":        name,
			"description": description,
			"amount":      amountInSmallestUnit,
			"currency":    currency,
		}

		// Create item in Razorpay
		item, err := s.razorpayClient.Item.Create(itemData, nil)
		if err != nil {
			slog.Error("paymentservice:service:CreateProductRazorpay", "error", err)
			return "", fmt.Errorf("failed to process the request")
		}

		// Extract item ID from response
		itemID, ok := item["id"].(string)
		if !ok {
			slog.Error("paymentservice:service:CreateProductRazorpay", "error", "failed to get item ID")
			return "", fmt.Errorf("failed to process the request")
		}

		slog.Info("paymentservice:service:CreateProductRazorpay", "itemID", itemID)
		return itemID, nil
	}
}

func (s *PaymentService) CreateRazorpayCheckoutSession(ctx context.Context, userID string, productID string) (string, int64, string, error) {
	slog.Info("paymentservice:service:CreateRazorpayCheckoutSession", "userID", userID, "productID", productID)

	hasAccess, err := s.dao.CheckUserProductAccess(userID, productID)
	if err != nil {
		slog.Error("paymentservice:razorpay:CreateRazorpayCheckoutSession", "error", err)
		return "", 0, "", fmt.Errorf("failed to process the request")
	}

	if hasAccess {
		slog.Error("paymentservice:razorpay:CreateRazorpayCheckoutSession", "error", "user already has access to this product")
		return "", 0, "", fmt.Errorf("already have access to this product")
	}

	// Get product by product ID
	product, err := s.dao.GetProductById(productID)
	if err != nil {
		slog.Error("paymentservice:razorpay:CreateRazorpayCheckoutSession", "error", err)
		return "", 0, "", fmt.Errorf("failed to create Razorpay checkout session")
	}

	if product.IsRecurring {
		slog.Error("paymentservice:service:CreateRazorpayCheckoutSession", "error", "cannot create checkout session for a subscription product")
		return "", 0, "", fmt.Errorf("cannot create checkout session for a subscription product")
	}

	slog.Info("paymentservice:razorpay:CreateRazorpayCheckoutSession", "product", product.Price, "type", reflect.TypeOf(product.Price))

	// Razorpay expects amount in the smallest currency unit (paise for INR, cents for USD, etc.)
	orderParams := map[string]interface{}{
		"amount":   product.Price,
		"currency": product.Currency,
		"receipt":  product.RazorpayProductID,
		"notes": map[string]interface{}{
			"user_id":    userID,
			"product_id": product.ID,
		},
		"payment_capture": 1, // Auto-capture payment after authorization
	}

	order, err := s.razorpayClient.Order.Create(orderParams, nil)
	if err != nil {
		slog.Error("paymentservice:razorpay:CreateRazorpayCheckoutSession", "error", err)
		return "", 0, "", fmt.Errorf("failed to create Razorpay checkout session")
	}

	// Extract Order ID
	orderID, ok := order["id"].(string)
	if !ok {
		slog.Error("paymentservice:razorpay:CreateRazorpayCheckoutSession", "error", "Order ID not found in response")
		return "", 0, "", fmt.Errorf("failed to create Razorpay checkout session")
	}

	slog.Info("paymentservice:razorpay:CreateRazorpayCheckoutSession", "orderID", orderID)
	return orderID, product.Price, product.Currency, nil
}

func (s *PaymentService) CreateRazorpaySubscriptionCheckoutSession(ctx context.Context, userID string, productID string) (string, int64, string, error) {
	slog.Info("paymentservice:razorpay:CreateRazorpaySubscriptionCheckoutSession", "userID", userID, "productID", productID)

	hasAccess, err := s.dao.CheckUserProductAccess(userID, productID)
	if err != nil {
		slog.Error("paymentservice:razorpay:CreateRazorpaySubscriptionCheckoutSession", "error", err)
		return "", 0, "", fmt.Errorf("failed to process the request")
	}

	if hasAccess {
		slog.Error("paymentservice:razorpay:CreateRazorpaySubscriptionCheckoutSession", "error", "user already has access to this product")
		return "", 0, "", fmt.Errorf("already have access to this product")
	}

	// Get product by product ID
	product, err := s.dao.GetProductById(productID)
	if err != nil {
		slog.Error("paymentservice:razorpay:CreateRazorpaySubscriptionCheckoutSession", "error", err)
		return "", 0, "", fmt.Errorf("failed to create Razorpay subscription checkout session")
	}

	if !product.IsRecurring {
		slog.Error("paymentservice:service:CreateRazorpaySubscriptionCheckoutSession", "error", "cannot create checkout session for a one-time product")
		return "", 0, "", fmt.Errorf("cannot create checkout session for a one-time product")
	}

	subscriptionData := map[string]interface{}{
		"plan_id":         product.RazorpayProductID,
		"total_count":     999, //taking high value for longer period, stripe doesn't ask for it
		"quantity":        1,
		"customer_notify": true,
		"notes": map[string]interface{}{
			"user_id":    userID,
			"product_id": product.ID,
		},
	}

	subscription, err := s.razorpayClient.Subscription.Create(subscriptionData, nil)
	if err != nil {
		slog.Error("paymentservice:razorpay:CreateRazorpaySubscriptionCheckoutSession", "error", err)
		return "", 0, "", fmt.Errorf("failed to create Razorpay subscription checkout session")
	}

	subscriptionID, ok := subscription["id"].(string)
	if !ok {
		slog.Error("paymentservice:razorpay:CreateRazorpaySubscriptionCheckoutSession", "error", "Subscription ID not found in response")
		return "", 0, "", fmt.Errorf("failed to create Razorpay subscription checkout session")
	}

	slog.Info("paymentservice:razorpay:CreateRazorpaySubscriptionCheckoutSession", "subscriptionID", subscriptionID)
	return subscriptionID, product.Price, product.Currency, nil
}

func (s *PaymentService) HandleRazorpayWebhook(ctx context.Context, r *http.Request) error {
	slog.Info("paymentservice:razorpay:HandleRazorpayWebhook")

	// Read the raw webhook payload
	defer r.Body.Close()
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("paymentservice:razorpay:HandleRazorpayWebhook", "error", "failed to read request body", "details", err)
		return fmt.Errorf("error reading request body: %w", err)
	}

	// Get signature from header
	signature := r.Header.Get("X-Razorpay-Signature")
	if signature == "" {
		slog.Error("paymentservice:razorpay:HandleRazorpayWebhook", "error", "missing signature header")
		return fmt.Errorf("missing signature header")
	}

	secret := os.Getenv("RAZORPAY_WEBHOOK_SECRET")
	if strings.TrimSpace(secret) == "" {
		slog.Error("paymentservice:razorpay:HandleRazorpayWebhook", "error", "RAZORPAY_WEBHOOK_SECRET is not set")
		return fmt.Errorf("configuration error")
	}

	// Verify signature using raw payload
	if !s.verifySignature(payload, signature, secret) {
		slog.Error("paymentservice:razorpay:HandleRazorpayWebhook", "error", "invalid signature")
		return fmt.Errorf("invalid signature")
	}

	// Get event ID for idempotency check
	eventID := r.Header.Get("x-razorpay-event-id")
	slog.Info("paymentservice:razorpay:HandleRazorpayWebhook", "event_id", eventID)

	// Parse JSON payload
	var webhookData map[string]interface{}
	if err := json.Unmarshal(payload, &webhookData); err != nil {
		slog.Error("paymentservice:razorpay:HandleRazorpayWebhook", "error", "failed to parse JSON", "details", err)
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Extract event type
	event, ok := webhookData["event"].(string)
	if !ok {
		slog.Error("paymentservice:razorpay:HandleRazorpayWebhook", "error", "event type not found")
		return fmt.Errorf("event type not found")
	}

	switch event {
	case "payment.captured":
		err := s.handleRazorpayPaymentCaptured(ctx, webhookData)
		if err != nil {
			slog.Error("paymentservice:razorpay:HandleRazorpayWebhook", "error", "failed to handle payment captured", "details", err)
			return fmt.Errorf("failed to handle payment captured: %v", err)
		}
	case "payment.failed":
		err := s.handleRazorpayPaymentFailed(ctx, webhookData)
		if err != nil {
			slog.Error("paymentservice:razorpay:HandleRazorpayWebhook", "error", "failed to handle payment failed", "details", err)
			return fmt.Errorf("failed to handle payment failed: %v", err)
		}

	// Subscription events
	case "subscription.authenticated":
		err := s.handleRazorpaySubscriptionAuthenticated(ctx, webhookData)
		if err != nil {
			slog.Error("paymentservice:razorpay:HandleRazorpayWebhook", "error", "failed to handle subscription authenticated", "details", err)
			return fmt.Errorf("failed to handle subscription authenticated: %v", err)
		}
	case "subscription.charged":
		err := s.handleRazorpaySubscriptionCharged(ctx, webhookData)
		if err != nil {
			slog.Error("paymentservice:razorpay:HandleRazorpayWebhook", "error", "failed to handle subscription charged", "details", err)
			return fmt.Errorf("failed to handle subscription charged: %v", err)
		}
	case "subscription.pending":
		err := s.handleRazorpaySubscriptionPaymentFailed(ctx, webhookData)
		if err != nil {
			slog.Error("paymentservice:razorpay:HandleRazorpayWebhook", "error", "failed to handle subscription cancelled", "details", err)
			return fmt.Errorf("failed to handle subscription cancelled: %v", err)
		}

	default:
		slog.Info("paymentservice:razorpay:HandleRazorpayWebhook", "event", "unhandled event type", "type", event)
	}

	slog.Info("paymentservice:razorpay:HandleRazorpayWebhook", "status", "webhook processed successfully")
	return nil
}

// this is for one-time payments
func (s *PaymentService) handleRazorpayPaymentCaptured(ctx context.Context, webhookData map[string]interface{}) error {
	slog.Info("paymentservice:razorpay:handleRazorpayPaymentCaptured")
	// Extract payment data from webhook
	payloadData, ok := webhookData["payload"].(map[string]interface{})
	if !ok {
		slog.Error("paymentservice:razorpay:handleRazorpayPaymentCaptured", "error", "payload not found in webhook data")
		return fmt.Errorf("payload not found in webhook data")
	}

	paymentData, ok := payloadData["payment"].(map[string]interface{})
	if !ok {
		slog.Error("paymentservice:razorpay:handleRazorpayPaymentCaptured", "error", "payment data not found in payload")
		return fmt.Errorf("payment data not found in payload")
	}

	entityData, ok := paymentData["entity"].(map[string]interface{})
	if !ok {
		slog.Error("paymentservice:razorpay:handleRazorpayPaymentCaptured", "error", "entity data not found in payment")
		return fmt.Errorf("entity data not found in payment")
	}

	// Extract payment ID for session tracking
	paymentID, ok := entityData["id"].(string)
	if !ok {
		slog.Error("paymentservice:razorpay:handleRazorpayPaymentCaptured", "error", "payment id not found in entity")
		return fmt.Errorf("payment id not found in entity")
	}

	// Extract user_id and product_id from notes
	notes, ok := entityData["notes"].(map[string]interface{})
	if !ok {
		slog.Error("paymentservice:razorpay:handleRazorpayPaymentCaptured", "error", "notes not found in payment entity")
		return fmt.Errorf("notes not found in payment entity")
	}

	userID, ok := notes["user_id"].(string)
	if !ok {
		slog.Error("paymentservice:razorpay:handleRazorpayPaymentCaptured", "error", "user_id not found in notes")
		return fmt.Errorf("user_id not found in notes")
	}

	productID, ok := notes["product_id"].(string)
	if !ok {
		slog.Error("paymentservice:razorpay:handleRazorpayPaymentCaptured", "error", "product_id not found in notes")
		return fmt.Errorf("product_id not found in notes")
	}

	// Extract created_at timestamp from payment entity
	var currentTime int64
	if createdAt, exists := entityData["created_at"]; exists {
		if createdAtFloat, ok := createdAt.(float64); ok {
			currentTime = int64(createdAtFloat)
		}
	}
	if currentTime == 0 {
		// Fallback to current Unix timestamp if not available
		currentTime = time.Now().Unix()
	}

	// For one-time payments, create subscription with period end = period start to indicate it's expired/one-time
	periodStart := currentTime
	periodEnd := currentTime // Set end time to be the same as start time to indicate one-time payment

	product, err := s.dao.GetProductById(productID)
	if err != nil {
		slog.Error("paymentservice:razorpay:handleRazorpayPaymentCaptured", "error", "failed to get product", "details", err)
		return fmt.Errorf("failed to get product: %v", err)
	}

	eventID := paymentID // use payment ID as event ID to avoid duplicate subscriptions
	// Create subscription record for one-time payment
	subscriptionID, err := s.dao.CreateSubscription(
		eventID,
		userID,
		productID,
		"razorpay",  // provider
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
		slog.Error("paymentservice:razorpay:handleRazorpayPaymentCaptured", "error", "failed to create subscription", "details", err)
		return fmt.Errorf("failed to create subscription: %v", err)
	}

	// Marshal the entire webhook data to JSON for storage
	webhookJSON, err := json.Marshal(webhookData)
	if err != nil {
		slog.Error("paymentservice:razorpay:handleRazorpayPaymentCaptured", "error", "failed to marshal webhook to JSON", "details", err)
		return fmt.Errorf("failed to marshal webhook to JSON: %v", err)
	}

	// Create user_payment record for the one-time payment
	_, err = s.dao.CreateUserPayment(userID, productID, subscriptionID, paymentID, string(webhookJSON), true)
	if err != nil {
		slog.Error("paymentservice:razorpay:handleRazorpayPaymentCaptured", "error", "failed to create user payment", "details", err)
		return fmt.Errorf("failed to create user payment: %v", err)
	}

	slog.Info("paymentservice:razorpay:handleRazorpayPaymentCaptured", "subscriptionID", subscriptionID, "paymentID", paymentID, "userID", userID, "productID", productID)

	return nil
}

// this is for one-time payments
func (s *PaymentService) handleRazorpayPaymentFailed(ctx context.Context, webhookData map[string]interface{}) error {
	// Extract payment data from webhook
	payloadData, ok := webhookData["payload"].(map[string]interface{})
	if !ok {
		slog.Error("paymentservice:razorpay:handleRazorpayPaymentFailed", "error", "payload not found in webhook data")
		return fmt.Errorf("payload not found in webhook data")
	}

	paymentData, ok := payloadData["payment"].(map[string]interface{})
	if !ok {
		slog.Error("paymentservice:razorpay:handleRazorpayPaymentFailed", "error", "payment data not found in payload")
		return fmt.Errorf("payment data not found in payload")
	}

	entityData, ok := paymentData["entity"].(map[string]interface{})
	if !ok {
		slog.Error("paymentservice:razorpay:handleRazorpayPaymentFailed", "error", "entity data not found in payment")
		return fmt.Errorf("entity data not found in payment")
	}

	// Extract payment ID for session tracking
	paymentID, ok := entityData["id"].(string)
	if !ok {
		slog.Error("paymentservice:razorpay:handleRazorpayPaymentFailed", "error", "payment id not found in entity")
		return fmt.Errorf("payment id not found in entity")
	}

	// Extract user_id and product_id from notes
	notes, ok := entityData["notes"].(map[string]interface{})
	if !ok {
		slog.Error("paymentservice:razorpay:handleRazorpayPaymentFailed", "error", "notes not found in payment entity")
		return fmt.Errorf("notes not found in payment entity")
	}

	userID, ok := notes["user_id"].(string)
	if !ok {
		slog.Error("paymentservice:razorpay:handleRazorpayPaymentFailed", "error", "user_id not found in notes")
		return fmt.Errorf("user_id not found in notes")
	}

	productID, ok := notes["product_id"].(string)
	if !ok {
		slog.Error("paymentservice:razorpay:handleRazorpayPaymentFailed", "error", "product_id not found in notes")
		return fmt.Errorf("product_id not found in notes")
	}

	// Marshal the entire webhook data to JSON for storage
	webhookJSON, err := json.Marshal(webhookData)
	if err != nil {
		slog.Error("paymentservice:razorpay:handleRazorpayPaymentFailed", "error", "failed to marshal webhook to JSON", "details", err)
		return fmt.Errorf("failed to marshal webhook to JSON: %v", err)
	}

	subscription, err := s.dao.GetSubscriptionByUserIDAndProductID(userID, productID)
	if err != nil {
		slog.Error("paymentservice:razorpay:handleRazorpayPaymentFailed", "error", "failed to get subscription", "details", err)
		// for one-time payments, create user payment without subscription reference
		_, err = s.dao.CreateUserPayment(userID, productID, "", paymentID, string(webhookJSON), false)
		if err != nil {
			slog.Error("paymentservice:razorpay:handleRazorpayPaymentFailed", "error", "failed to create user payment", "details", err)
			return fmt.Errorf("failed to create user payment: %v", err)
		}
		return nil
	}

	// Save to database with is_success = false
	_, err = s.dao.CreateUserPayment(userID, productID, subscription.ID, paymentID, string(webhookJSON), false)
	if err != nil {
		slog.Error("paymentservice:razorpay:handleRazorpayPaymentFailed", "error", "failed to create user payment", "details", err)
		return fmt.Errorf("failed to create user payment: %v", err)
	}

	slog.Info("paymentservice:razorpay:handleRazorpayPaymentFailed", "userID", userID, "productID", productID, "data", "saved failed transaction in database")

	return nil
}

// New Razorpay subscription webhook handlers- this is for recurring payments(subscription created)
func (s *PaymentService) handleRazorpaySubscriptionAuthenticated(ctx context.Context, webhookData map[string]interface{}) error {
	slog.Info("paymentservice:razorpay:handleRazorpaySubscriptionAuthenticated", "webhookData", webhookData)

	// Extract subscription data from webhook - try different possible structures
	var subscriptionEntity map[string]interface{}

	// First try: payload.subscription.entity structure (like payment webhooks)
	if payloadData, ok := webhookData["payload"].(map[string]interface{}); ok {
		if subscriptionData, ok := payloadData["subscription"].(map[string]interface{}); ok {
			if entity, ok := subscriptionData["entity"].(map[string]interface{}); ok {
				subscriptionEntity = entity
				slog.Info("paymentservice:razorpay:handleRazorpaySubscriptionAuthenticated", "structure", "payload.subscription.entity")
			} else {
				// Second try: payload.subscription structure (direct)
				subscriptionEntity = subscriptionData
				slog.Info("paymentservice:razorpay:handleRazorpaySubscriptionAuthenticated", "structure", "payload.subscription")
			}
		}
	}

	// Third try: direct subscription in webhook data
	if subscriptionEntity == nil {
		if subscription, ok := webhookData["subscription"].(map[string]interface{}); ok {
			subscriptionEntity = subscription
			slog.Info("paymentservice:razorpay:handleRazorpaySubscriptionAuthenticated", "structure", "direct subscription")
		}
	}

	if subscriptionEntity == nil {
		slog.Error("paymentservice:razorpay:handleRazorpaySubscriptionAuthenticated", "error", "subscription entity not found in any expected structure")
		return fmt.Errorf("subscription entity not found in webhook data")
	}

	// Extract subscription ID
	razorpaySubscriptionID, ok := subscriptionEntity["id"].(string)
	if !ok {
		slog.Error("paymentservice:razorpay:handleRazorpaySubscriptionAuthenticated", "error", "subscription ID not found", "subscriptionEntity", subscriptionEntity)
		return fmt.Errorf("subscription ID not found")
	}

	razorpayCustomerID, ok := subscriptionEntity["customer_id"].(string)
	if !ok {
		slog.Warn("paymentservice:razorpay:handleRazorpaySubscriptionAuthenticated", "warning", "customer ID not found in subscription entity")
		razorpayCustomerID = ""
	}

	// Extract user_id and product_id from notes
	notes, ok := subscriptionEntity["notes"].(map[string]interface{})
	if !ok {
		slog.Error("paymentservice:razorpay:handleRazorpaySubscriptionAuthenticated", "error", "notes not found in subscription entity")
		return fmt.Errorf("notes not found in subscription entity")
	}

	userID, userExists := notes["user_id"].(string)
	if !userExists {
		slog.Error("paymentservice:razorpay:handleRazorpaySubscriptionAuthenticated", "error", "user_id not found in subscription notes")
		return fmt.Errorf("user_id not found in subscription notes")
	}

	productID, productExists := notes["product_id"].(string)
	if !productExists {
		slog.Error("paymentservice:razorpay:handleRazorpaySubscriptionAuthenticated", "error", "product_id not found in subscription notes")
		return fmt.Errorf("product_id not found in subscription notes")
	}

	// Extract period information from subscription
	var currentPeriodStart, currentPeriodEnd int64

	// Extract start_at timestamp (when subscription starts)
	if startAt, exists := subscriptionEntity["start_at"]; exists {
		if startAtFloat, ok := startAt.(float64); ok {
			currentPeriodStart = int64(startAtFloat)
		}
	}

	// Extract end_at timestamp (when subscription ends)
	if endAt, exists := subscriptionEntity["end_at"]; exists {
		if endAtFloat, ok := endAt.(float64); ok {
			currentPeriodEnd = int64(endAtFloat)
		}
	}

	product, err := s.dao.GetProductById(productID)
	if err != nil {
		slog.Error("paymentservice:razorpay:handleRazorpaySubscriptionAuthenticated", "error", "failed to get product", "details", err)
		return fmt.Errorf("failed to get product: %v", err)
	}

	// Extract subscription status
	status, ok := subscriptionEntity["status"].(string)
	if !ok {
		status = "created" // Default status
	}

	eventID := razorpaySubscriptionID // use subscription ID as event ID to avoid duplicate subscriptions

	// Create subscription record in our database with all details
	subscriptionID, err := s.dao.CreateSubscription(
		eventID,
		userID,
		productID,
		"razorpay",             // provider
		razorpaySubscriptionID, // provider_subscription_id
		razorpayCustomerID,     // provider_customer_id - Razorpay doesn't have explicit customer ID in this webhook
		status,                 // provider_subscription_status
		"active",               // status - set to active since subscription is authenticated
		currentPeriodStart,     // current_period_start from subscription
		currentPeriodEnd,       // current_period_end from subscription
		false,                  // cancel_at_period_end - default to false
		product.IsRecurring,
	)
	if err != nil {
		slog.Error("paymentservice:razorpay:handleRazorpaySubscriptionAuthenticated", "error", "failed to create subscription", "details", err)
		return fmt.Errorf("failed to create subscription: %v", err)
	}

	slog.Info("paymentservice:razorpay:handleRazorpaySubscriptionAuthenticated", "subscriptionID", subscriptionID, "userID", userID, "productID", productID, "razorpaySubscriptionID", razorpaySubscriptionID, "currentPeriodStart", currentPeriodStart, "currentPeriodEnd", currentPeriodEnd, "status", status)
	return nil
}

// this is for recurring payments(subscription charged)
func (s *PaymentService) handleRazorpaySubscriptionCharged(ctx context.Context, webhookData map[string]interface{}) error {
	slog.Info("paymentservice:razorpay:handleRazorpaySubscriptionCharged")

	// Extract payment and subscription data from webhook
	payloadData, ok := webhookData["payload"].(map[string]interface{})
	if !ok {
		slog.Error("paymentservice:razorpay:handleRazorpaySubscriptionCharged", "error", "payload not found in webhook data")
		return fmt.Errorf("payload not found in webhook data")
	}

	// Extract payment data
	paymentData, ok := payloadData["payment"].(map[string]interface{})
	if !ok {
		slog.Error("paymentservice:razorpay:handleRazorpaySubscriptionCharged", "error", "payment data not found in payload")
		return fmt.Errorf("payment data not found in payload")
	}

	paymentEntity, ok := paymentData["entity"].(map[string]interface{})
	if !ok {
		slog.Error("paymentservice:razorpay:handleRazorpaySubscriptionCharged", "error", "payment entity not found")
		return fmt.Errorf("payment entity not found")
	}

	paymentID, ok := paymentEntity["id"].(string)
	if !ok {
		slog.Error("paymentservice:razorpay:handleRazorpaySubscriptionCharged", "error", "payment ID not found")
		return fmt.Errorf("payment ID not found")
	}

	// Extract subscription data
	subscriptionData, ok := payloadData["subscription"].(map[string]interface{})
	if !ok {
		slog.Error("paymentservice:razorpay:handleRazorpaySubscriptionCharged", "error", "subscription data not found in payload")
		return fmt.Errorf("subscription data not found in payload")
	}

	subscriptionEntity, ok := subscriptionData["entity"].(map[string]interface{})
	if !ok {
		slog.Error("paymentservice:razorpay:handleRazorpaySubscriptionCharged", "error", "subscription entity not found")
		return fmt.Errorf("subscription entity not found")
	}

	razorpaySubscriptionID, ok := subscriptionEntity["id"].(string)
	if !ok {
		slog.Error("paymentservice:razorpay:handleRazorpaySubscriptionCharged", "error", "subscription ID not found")
		return fmt.Errorf("subscription ID not found")
	}

	slog.Info("paymentservice:razorpay:handleRazorpaySubscriptionCharged", "paymentID", paymentID, "subscriptionID", razorpaySubscriptionID)

	// Extract customer ID from subscription entity
	razorpayCustomerID, ok := subscriptionEntity["customer_id"].(string)
	if !ok {
		slog.Error("paymentservice:razorpay:handleRazorpaySubscriptionCharged", "error", "customer ID not found in subscription entity")
		return fmt.Errorf("customer ID not found in subscription entity")
	}

	// Find our internal subscription by provider_customer_id
	subscription, err := s.dao.GetSubscriptionByProviderCustomerID(razorpayCustomerID)
	if err != nil {
		slog.Error("paymentservice:razorpay:handleRazorpaySubscriptionCharged", "error", "failed to find subscription", "details", err)
		return fmt.Errorf("failed to find subscription: %v", err)
	}

	// Extract period information from subscription entity
	var currentPeriodStart, currentPeriodEnd int64

	// Extract current_start timestamp
	if currentStart, exists := subscriptionEntity["current_start"]; exists {
		if currentStartFloat, ok := currentStart.(float64); ok {
			currentPeriodStart = int64(currentStartFloat)
		}
	}

	// Extract current_end timestamp
	if currentEnd, exists := subscriptionEntity["current_end"]; exists {
		if currentEndFloat, ok := currentEnd.(float64); ok {
			currentPeriodEnd = int64(currentEndFloat)
		}
	}

	// Extract subscription status
	status, ok := subscriptionEntity["status"].(string)
	if !ok {
		status = subscription.ProviderSubscriptionStatus // Keep existing status if not provided
	}

	// Update subscription with period dates and set status to active
	err = s.dao.UpdateSubscription(
		subscription.ID,
		subscription.ProviderSubscriptionID, // keep existing provider_subscription_id
		subscription.ProviderCustomerID,     // keep existing provider_customer_id
		status,                              // update provider_subscription_status
		"active",                            // set status to active since payment is charged
		currentPeriodStart,
		currentPeriodEnd,
		subscription.CancelAtPeriodEnd, // keep existing cancel_at_period_end
	)
	if err != nil {
		slog.Error("paymentservice:razorpay:handleRazorpaySubscriptionCharged", "error", "failed to update subscription", "details", err)
		return fmt.Errorf("failed to update subscription: %v", err)
	}

	slog.Info("paymentservice:razorpay:handleRazorpaySubscriptionCharged", "status", "subscription updated", "subscriptionID", subscription.ID, "currentPeriodStart", currentPeriodStart, "currentPeriodEnd", currentPeriodEnd)

	// Create user_payment record for this subscription charge
	webhookJSON, err := json.Marshal(webhookData)
	if err != nil {
		slog.Error("paymentservice:razorpay:handleRazorpaySubscriptionCharged", "error", "failed to marshal webhook to JSON", "details", err)
		return fmt.Errorf("failed to marshal webhook to JSON: %v", err)
	}

	_, err = s.dao.CreateUserPayment(subscription.UserID, subscription.ProductID, subscription.ID, paymentID, string(webhookJSON), true)
	if err != nil {
		slog.Error("paymentservice:razorpay:handleRazorpaySubscriptionCharged", "error", "failed to create user payment", "details", err)
		return fmt.Errorf("failed to create user payment: %v", err)
	}

	slog.Info("paymentservice:razorpay:handleRazorpaySubscriptionCharged", "status", "user payment created", "paymentID", paymentID, "userID", subscription.UserID, "productID", subscription.ProductID)

	return nil
}

func (s *PaymentService) handleRazorpaySubscriptionPaymentFailed(ctx context.Context, webhookData map[string]interface{}) error {
	slog.Info("paymentservice:razorpay:handleRazorpaySubscriptionPaymentFailed")

	webhookJSON, err := json.Marshal(webhookData)
	if err != nil {
		slog.Error("paymentservice:razorpay:handleRazorpaySubscriptionPaymentFailed", "error", "failed to marshal webhook to JSON", "details", err)
		return fmt.Errorf("failed to marshal webhook to JSON: %v", err)
	}

	payloadData, ok := webhookData["payload"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("payload not found in webhook data")
	}

	subscriptionEntity, ok := payloadData["subscription"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("subscription entity not found in payload")
	}

	subscriptionID, ok := subscriptionEntity["id"].(string)
	if !ok {
		return fmt.Errorf("subscription ID not found")
	}

	customerID, ok := subscriptionEntity["customer_id"].(string)
	if !ok {
		return fmt.Errorf("customer ID not found in subscription entity")
	}

	subscription, err := s.dao.GetSubscriptionByProviderCustomerID(customerID)
	if err != nil {
		slog.Error("paymentservice:razorpay:handleRazorpaySubscriptionPaymentFailed", "error", "failed to find subscription", "details", err)
		return fmt.Errorf("failed to find subscription: %v", err)
	}

	slog.Info("paymentservice:razorpay:handleRazorpaySubscriptionPaymentFailed", "subscriptionID", subscriptionID)

	failedPaymentID := fmt.Sprintf("failed_%s_%d", subscription.ID, time.Now().UnixNano())
	// Create user_payment record for this subscription payment failed
	_, err = s.dao.CreateUserPayment(subscription.UserID, subscription.ProductID, subscription.ID, failedPaymentID, string(webhookJSON), false)
	if err != nil {
		slog.Error("paymentservice:razorpay:handleRazorpaySubscriptionPaymentFailed", "error", "failed to create user payment", "details", err)
		return fmt.Errorf("failed to create user payment: %v", err)
	}

	return nil
}

// Helper function to verify HMAC SHA256 signature
func (s *PaymentService) verifySignature(payload []byte, signature, secret string) bool {
	// Create HMAC with SHA256
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures using constant time comparison
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

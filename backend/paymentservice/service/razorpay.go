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
)

func (s *PaymentService) CreateProductRazorpay(ctx context.Context, name string, description string, amountInSmallestUnit int64, currency string, isRecurring bool, intervalCount int64, period string) (string, error) {
	slog.Info("paymentservice:service:CreateProductRazorpay", "name", name, "isRecurring", isRecurring, "intervalCount", intervalCount, "period", period, "description", description, "amountInSmallestUnit", amountInSmallestUnit, "currency", currency)

	if isRecurring {

		// Validate interval count for daily plans (minimum 7 according to Razorpay docs)
		actualIntervalCount := intervalCount
		if period == "daily" && intervalCount < 7 {
			slog.Warn("paymentservice:service:CreateProductRazorpay", "warning", "Daily plans require minimum interval of 7, adjusting", "original", intervalCount, "adjusted", 7)
			actualIntervalCount = 7
		}

		planData := map[string]interface{}{
			"period":   period,              // "daily", "weekly", "monthly", "quarterly", "yearly"
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

	// Get product by product ID
	product, err := s.dao.GetProductById(productID)
	if err != nil {
		slog.Error("paymentservice:service:CreateRazorpayCheckoutSession", "error", err)
		return "", 0, "", fmt.Errorf("failed to create Razorpay checkout session")
	}

	slog.Info("paymentservice:service:CreateRazorpayCheckoutSession", "product", product.Price, "type", reflect.TypeOf(product.Price))

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
		slog.Error("paymentservice:service:CreateRazorpayCheckoutSession", "error", err)
		return "", 0, "", fmt.Errorf("failed to create Razorpay checkout session")
	}

	// Extract Order ID
	orderID, ok := order["id"].(string)
	if !ok {
		slog.Error("paymentservice:service:CreateRazorpayCheckoutSession", "error", "Order ID not found in response")
		return "", 0, "", fmt.Errorf("failed to create Razorpay checkout session")
	}

	slog.Info("paymentservice:service:CreateRazorpayCheckoutSession", "orderID", orderID)
	return orderID, product.Price, product.Currency, nil
}

func (s *PaymentService) CreateRazorpaySubscriptionCheckoutSession(ctx context.Context, userID string, productID string) (string, int64, string, error) {
	slog.Info("paymentservice:service:CreateRazorpaySubscriptionCheckoutSession", "userID", userID, "productID", productID)

	// Get product by product ID
	product, err := s.dao.GetProductById(productID)
	if err != nil {
		slog.Error("paymentservice:service:CreateRazorpaySubscriptionCheckoutSession", "error", err)
		return "", 0, "", fmt.Errorf("failed to create Razorpay subscription checkout session")
	}

	subscriptionData := map[string]interface{}{
		"plan_id":         product.RazorpayProductID,
		"total_count":     12, //maybe take it from ui
		"quantity":        1,
		"customer_notify": true,
		"addons": []interface{}{
			map[string]interface{}{
				"item": map[string]interface{}{
					"name":     product.Name,
					"amount":   product.Price,
					"currency": product.Currency,
				},
			},
		},
		"notes": map[string]interface{}{
			"user_id":    userID,
			"product_id": product.ID,
		},
	}

	subscription, err := s.razorpayClient.Subscription.Create(subscriptionData, nil)
	if err != nil {
		slog.Error("paymentservice:service:CreateRazorpaySubscriptionCheckoutSession", "error", err)
		return "", 0, "", fmt.Errorf("failed to create Razorpay subscription checkout session")
	}

	subscriptionID, ok := subscription["id"].(string)
	if !ok {
		slog.Error("paymentservice:service:CreateRazorpaySubscriptionCheckoutSession", "error", "Subscription ID not found in response")
		return "", 0, "", fmt.Errorf("failed to create Razorpay subscription checkout session")
	}

	slog.Info("paymentservice:service:CreateRazorpaySubscriptionCheckoutSession", "subscriptionID", subscriptionID)
	return subscriptionID, product.Price, product.Currency, nil
}

func (s *PaymentService) HandleRazorpayWebhook(ctx context.Context, r *http.Request) error {
	slog.Info("paymentservice:service:HandleRazorpayWebhook")

	// Read the raw webhook payload
	defer r.Body.Close()
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("paymentservice:service:HandleRazorpayWebhook", "error", "failed to read request body", "details", err)
		return fmt.Errorf("error reading request body: %w", err)
	}

	// Get signature from header
	signature := r.Header.Get("X-Razorpay-Signature")
	if signature == "" {
		slog.Error("paymentservice:service:HandleRazorpayWebhook", "error", "missing signature header")
		return fmt.Errorf("missing signature header")
	}

	secret := os.Getenv("RAZORPAY_WEBHOOK_SECRET")
	if strings.TrimSpace(secret) == "" {
		slog.Error("paymentservice:service:HandleRazorpayWebhook", "error", "RAZORPAY_WEBHOOK_SECRET is not set")
		return fmt.Errorf("configuration error")
	}

	// Verify signature using raw payload
	if !s.verifySignature(payload, signature, secret) {
		slog.Error("paymentservice:service:HandleRazorpayWebhook", "error", "invalid signature")
		return fmt.Errorf("invalid signature")
	}

	// Get event ID for idempotency check
	eventID := r.Header.Get("x-razorpay-event-id")
	slog.Info("paymentservice:service:HandleRazorpayWebhook", "event_id", eventID)

	// Parse JSON payload
	var webhookData map[string]interface{}
	if err := json.Unmarshal(payload, &webhookData); err != nil {
		slog.Error("paymentservice:service:HandleRazorpayWebhook", "error", "failed to parse JSON", "details", err)
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Extract event type
	event, ok := webhookData["event"].(string)
	if !ok {
		slog.Error("paymentservice:service:HandleRazorpayWebhook", "error", "event type not found")
		return fmt.Errorf("event type not found")
	}

	switch event {
	case "payment.captured":
		err := s.handleRazorpayPaymentCaptured(ctx, webhookData)
		if err != nil {
			slog.Error("paymentservice:service:HandleRazorpayWebhook", "error", "failed to handle payment captured", "details", err)
			return fmt.Errorf("failed to handle payment captured: %v", err)
		}
	case "payment.failed":
		err := s.handleRazorpayPaymentFailed(ctx, webhookData)
		if err != nil {
			slog.Error("paymentservice:service:HandleRazorpayWebhook", "error", "failed to handle payment failed", "details", err)
			return fmt.Errorf("failed to handle payment failed: %v", err)
		}

	case "invoice.paid":

	// Subscription events
	case "subscription.authenticated":
		err := s.handleRazorpaySubscriptionAuthenticated(ctx, webhookData)
		if err != nil {
			slog.Error("paymentservice:service:HandleRazorpayWebhook", "error", "failed to handle subscription authenticated", "details", err)
			return fmt.Errorf("failed to handle subscription authenticated: %v", err)
		}
	case "subscription.activated":
		err := s.handleRazorpaySubscriptionActivated(ctx, webhookData)
		if err != nil {
			slog.Error("paymentservice:service:HandleRazorpayWebhook", "error", "failed to handle subscription activated", "details", err)
			return fmt.Errorf("failed to handle subscription activated: %v", err)
		}
	case "subscription.charged":
		err := s.handleRazorpaySubscriptionCharged(ctx, webhookData)
		if err != nil {
			slog.Error("paymentservice:service:HandleRazorpayWebhook", "error", "failed to handle subscription charged", "details", err)
			return fmt.Errorf("failed to handle subscription charged: %v", err)
		}
	case "subscription.cancelled":
		err := s.handleRazorpaySubscriptionCancelled(ctx, webhookData)
		if err != nil {
			slog.Error("paymentservice:service:HandleRazorpayWebhook", "error", "failed to handle subscription cancelled", "details", err)
			return fmt.Errorf("failed to handle subscription cancelled: %v", err)
		}

	default:
		slog.Info("paymentservice:service:HandleRazorpayWebhook", "event", "unhandled event type", "type", event)
	}

	slog.Info("paymentservice:service:HandleRazorpayWebhook", "status", "webhook processed successfully")
	return nil
}

func (s *PaymentService) handleRazorpayPaymentCaptured(ctx context.Context, webhookData map[string]interface{}) error {
	slog.Info("paymentservice:service:handleRazorpayPaymentCaptured")
	// Extract payment data from webhook
	payloadData, ok := webhookData["payload"].(map[string]interface{})
	if !ok {
		slog.Error("paymentservice:service:handleRazorpayPaymentCaptured", "error", "payload not found in webhook data")
		return fmt.Errorf("payload not found in webhook data")
	}

	paymentData, ok := payloadData["payment"].(map[string]interface{})
	if !ok {
		slog.Error("paymentservice:service:handleRazorpayPaymentCaptured", "error", "payment data not found in payload")
		return fmt.Errorf("payment data not found in payload")
	}

	entityData, ok := paymentData["entity"].(map[string]interface{})
	if !ok {
		slog.Error("paymentservice:service:handleRazorpayPaymentCaptured", "error", "entity data not found in payment")
		return fmt.Errorf("entity data not found in payment")
	}

	// Extract payment ID for session tracking
	paymentID, ok := entityData["id"].(string)
	if !ok {
		slog.Error("paymentservice:service:handleRazorpayPaymentCaptured", "error", "payment id not found in entity")
		return fmt.Errorf("payment id not found in entity")
	}

	// Extract user_id and product_id from notes
	notes, ok := entityData["notes"].(map[string]interface{})
	if !ok {
		slog.Error("paymentservice:service:handleRazorpayPaymentCaptured", "error", "notes not found in payment entity")
		return fmt.Errorf("notes not found in payment entity")
	}

	userID, ok := notes["user_id"].(string)
	if !ok {
		slog.Error("paymentservice:service:handleRazorpayPaymentCaptured", "error", "user_id not found in notes")
		return fmt.Errorf("user_id not found in notes")
	}

	productID, ok := notes["product_id"].(string)
	if !ok {
		slog.Error("paymentservice:service:handleRazorpayPaymentCaptured", "error", "product_id not found in notes")
		return fmt.Errorf("product_id not found in notes")
	}

	// Marshal the entire webhook data to JSON for storage
	webhookJSON, err := json.Marshal(webhookData)
	if err != nil {
		slog.Error("paymentservice:service:handleRazorpayPaymentCaptured", "error", "failed to marshal webhook to JSON", "details", err)
		return fmt.Errorf("failed to marshal webhook to JSON: %v", err)
	}

	// Save to database
	_, err = s.dao.CreateUserPurchase(paymentID, userID, productID, string(webhookJSON), true, "razorpay")
	if err != nil {
		slog.Error("paymentservice:service:handleRazorpayPaymentCaptured", "error", "failed to create user purchase", "details", err)
		return fmt.Errorf("failed to create user purchase: %v", err)
	}

	slog.Info("paymentservice:service:handleRazorpayPaymentCaptured", "userID", userID, "productID", productID, "data", "saved in transaction in database")

	return nil
}

func (s *PaymentService) handleRazorpayPaymentFailed(ctx context.Context, webhookData map[string]interface{}) error {
	// Extract payment data from webhook
	payloadData, ok := webhookData["payload"].(map[string]interface{})
	if !ok {
		slog.Error("paymentservice:service:handleRazorpayPaymentFailed", "error", "payload not found in webhook data")
		return fmt.Errorf("payload not found in webhook data")
	}

	paymentData, ok := payloadData["payment"].(map[string]interface{})
	if !ok {
		slog.Error("paymentservice:service:handleRazorpayPaymentFailed", "error", "payment data not found in payload")
		return fmt.Errorf("payment data not found in payload")
	}

	entityData, ok := paymentData["entity"].(map[string]interface{})
	if !ok {
		slog.Error("paymentservice:service:handleRazorpayPaymentFailed", "error", "entity data not found in payment")
		return fmt.Errorf("entity data not found in payment")
	}

	// Extract payment ID for session tracking
	paymentID, ok := entityData["id"].(string)
	if !ok {
		slog.Error("paymentservice:service:handleRazorpayPaymentFailed", "error", "payment id not found in entity")
		return fmt.Errorf("payment id not found in entity")
	}

	// Extract user_id and product_id from notes
	notes, ok := entityData["notes"].(map[string]interface{})
	if !ok {
		slog.Error("paymentservice:service:handleRazorpayPaymentFailed", "error", "notes not found in payment entity")
		return fmt.Errorf("notes not found in payment entity")
	}

	userID, ok := notes["user_id"].(string)
	if !ok {
		slog.Error("paymentservice:service:handleRazorpayPaymentFailed", "error", "user_id not found in notes")
		return fmt.Errorf("user_id not found in notes")
	}

	productID, ok := notes["product_id"].(string)
	if !ok {
		slog.Error("paymentservice:service:handleRazorpayPaymentFailed", "error", "product_id not found in notes")
		return fmt.Errorf("product_id not found in notes")
	}

	// Marshal the entire webhook data to JSON for storage
	webhookJSON, err := json.Marshal(webhookData)
	if err != nil {
		slog.Error("paymentservice:service:handleRazorpayPaymentFailed", "error", "failed to marshal webhook to JSON", "details", err)
		return fmt.Errorf("failed to marshal webhook to JSON: %v", err)
	}

	// Save to database with is_success = false
	_, err = s.dao.CreateUserPurchase(paymentID, userID, productID, string(webhookJSON), false, "razorpay")
	if err != nil {
		slog.Error("paymentservice:service:handleRazorpayPaymentFailed", "error", "failed to create user purchase", "details", err)
		return fmt.Errorf("failed to create user purchase: %v", err)
	}

	slog.Info("paymentservice:service:handleRazorpayPaymentFailed", "userID", userID, "productID", productID, "data", "saved failed transaction in database")

	return nil
}

// New Razorpay subscription webhook handlers
func (s *PaymentService) handleRazorpaySubscriptionAuthenticated(ctx context.Context, webhookData map[string]interface{}) error {
	slog.Info("paymentservice:service:handleRazorpaySubscriptionAuthenticated")

	// Extract subscription data from webhook
	payloadData, ok := webhookData["payload"].(map[string]interface{})
	if !ok {
		slog.Error("paymentservice:service:handleRazorpaySubscriptionAuthenticated", "error", "payload not found in webhook data")
		return fmt.Errorf("payload not found in webhook data")
	}

	subscriptionEntity, ok := payloadData["subscription"].(map[string]interface{})
	if !ok {
		slog.Error("paymentservice:service:handleRazorpaySubscriptionAuthenticated", "error", "subscription entity not found in payload")
		return fmt.Errorf("subscription entity not found in payload")
	}

	subscriptionID, ok := subscriptionEntity["id"].(string)
	if !ok {
		slog.Error("paymentservice:service:handleRazorpaySubscriptionAuthenticated", "error", "subscription ID not found")
		return fmt.Errorf("subscription ID not found")
	}

	slog.Info("paymentservice:service:handleRazorpaySubscriptionAuthenticated", "subscriptionID", subscriptionID)

	// TODO: Update subscription record in database with provider_subscription_id and status

	return nil
}

func (s *PaymentService) handleRazorpaySubscriptionActivated(ctx context.Context, webhookData map[string]interface{}) error {
	slog.Info("paymentservice:service:handleRazorpaySubscriptionActivated")

	// Similar implementation to authenticated
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

	slog.Info("paymentservice:service:handleRazorpaySubscriptionActivated", "subscriptionID", subscriptionID)

	// TODO: Update subscription status to active

	return nil
}

func (s *PaymentService) handleRazorpaySubscriptionCharged(ctx context.Context, webhookData map[string]interface{}) error {
	slog.Info("paymentservice:service:handleRazorpaySubscriptionCharged")

	// Extract payment data for subscription charge
	payloadData, ok := webhookData["payload"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("payload not found in webhook data")
	}

	paymentEntity, ok := payloadData["payment"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("payment entity not found in payload")
	}

	paymentID, ok := paymentEntity["id"].(string)
	if !ok {
		return fmt.Errorf("payment ID not found")
	}

	slog.Info("paymentservice:service:handleRazorpaySubscriptionCharged", "paymentID", paymentID)

	// TODO: Create user_payment record for this subscription charge

	return nil
}

func (s *PaymentService) handleRazorpaySubscriptionCancelled(ctx context.Context, webhookData map[string]interface{}) error {
	slog.Info("paymentservice:service:handleRazorpaySubscriptionCancelled")

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

	slog.Info("paymentservice:service:handleRazorpaySubscriptionCancelled", "subscriptionID", subscriptionID)

	// TODO: Mark subscription as cancelled in database

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

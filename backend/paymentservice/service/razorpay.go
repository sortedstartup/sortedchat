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
)

func (s *PaymentService) CreateProductRazorpay(ctx context.Context, name string, description string, amountInSmallestUnit int64, currency string) (string, error) {
	slog.Info("paymentservice:service:CreateProductRazorpay", "name", name)

	// Create item data for Razorpay
	itemData := map[string]interface{}{
		"name":        name,
		"description": description,
		"amount":      amountInSmallestUnit, // Convert to smallest currency unit (paise)
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

func (s *PaymentService) CreateRazorpayCheckoutSession(ctx context.Context, userID string, razorpayItemID string) (string, int64, string, error) {
	slog.Info("paymentservice:service:CreateRazorpayCheckoutSession", "userID", userID, "itemID", razorpayItemID)

	product, err := s.dao.GetRazorpayProductById(razorpayItemID)
	if err != nil {
		slog.Error("paymentservice:service:CreateRazorpayCheckoutSession", "error", err)
		return "", 0, "", fmt.Errorf("failed to create Razorpay checkout session")
	}

	slog.Info("paymentservice:service:CreateRazorpayCheckoutSession", "product", product.Price, "type", reflect.TypeOf(product.Price))

	// Razorpay expects amount in the smallest currency unit (paise for INR, cents for USD, etc.)
	orderParams := map[string]interface{}{
		"amount":   product.Price,
		"currency": product.Currency,
		"receipt":  razorpayItemID,
		"notes": map[string]interface{}{
			"user_id": userID,
			"item_id": razorpayItemID,
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

func (s *PaymentService) HandleRazorpayWebhook(ctx context.Context, r *http.Request) error {
	slog.Info("paymentservice:service:HandleRazorpayWebhook")

	// Read the raw webhook payload
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("paymentservice:service:HandleWebhook", "error", "failed to read request body", "details", err)
		return fmt.Errorf("error reading request body: %w", err)
	}

	// Get signature from header
	signature := r.Header.Get("X-Razorpay-Signature")
	if signature == "" {
		slog.Error("paymentservice:service:HandleWebhook", "error", "missing signature header")
		return fmt.Errorf("missing signature header")
	}

	// Verify signature using raw payload
	if !s.verifySignature(payload, signature, os.Getenv("RAZORPAY_WEBHOOK_SECRET")) {
		slog.Error("paymentservice:service:HandleWebhook", "error", "invalid signature")
		return fmt.Errorf("invalid signature")
	}

	// Get event ID for idempotency check
	eventID := r.Header.Get("x-razorpay-event-id")
	slog.Info("paymentservice:service:HandleWebhook", "event_id", eventID)

	// Parse JSON payload
	var webhookData map[string]interface{}
	if err := json.Unmarshal(payload, &webhookData); err != nil {
		slog.Error("paymentservice:service:HandleWebhook", "error", "failed to parse JSON", "details", err)
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Extract event type
	event, ok := webhookData["event"].(string)
	if !ok {
		slog.Error("paymentservice:service:HandleWebhook", "error", "event type not found")
		return fmt.Errorf("event type not found")
	}

	switch event {
	case "payment.captured":
		err := s.handleRazorpayPaymentCaptured(ctx, webhookData)
		if err != nil {
			slog.Error("paymentservice:service:HandleWebhook", "error", "failed to handle payment captured", "details", err)
			return fmt.Errorf("failed to handle payment captured: %v", err)
		}
	case "payment.failed":
		err := s.handleRazorpayPaymentFailed(ctx, webhookData)
		if err != nil {
			slog.Error("paymentservice:service:HandleWebhook", "error", "failed to handle payment failed", "details", err)
			return fmt.Errorf("failed to handle payment failed: %v", err)
		}

	default:
		slog.Info("paymentservice:service:HandleWebhook", "event", "unhandled event type", "type", event)
	}

	slog.Info("paymentservice:service:HandleWebhook", "status", "webhook processed successfully")
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

	// Extract user_id and item_id from notes
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

	productID, ok := notes["item_id"].(string)
	if !ok {
		slog.Error("paymentservice:service:handleRazorpayPaymentCaptured", "error", "item_id not found in notes")
		return fmt.Errorf("item_id not found in notes")
	}

	// Marshal the entire webhook data to JSON for storage
	webhookJSON, err := json.Marshal(webhookData)
	if err != nil {
		slog.Error("paymentservice:service:handleRazorpayPaymentCaptured", "error", "failed to marshal webhook to JSON", "details", err)
		return fmt.Errorf("failed to marshal webhook to JSON: %v", err)
	}

	// Save to database
	_, err = s.dao.CreateUserPurchase(userID, productID, string(webhookJSON), true)
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

	// Extract user_id and item_id from notes
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

	productID, ok := notes["item_id"].(string)
	if !ok {
		slog.Error("paymentservice:service:handleRazorpayPaymentFailed", "error", "item_id not found in notes")
		return fmt.Errorf("item_id not found in notes")
	}

	// Marshal the entire webhook data to JSON for storage
	webhookJSON, err := json.Marshal(webhookData)
	if err != nil {
		slog.Error("paymentservice:service:handleRazorpayPaymentFailed", "error", "failed to marshal webhook to JSON", "details", err)
		return fmt.Errorf("failed to marshal webhook to JSON: %v", err)
	}

	// Save to database with is_success = false
	_, err = s.dao.CreateUserPurchase(userID, productID, string(webhookJSON), false)
	if err != nil {
		slog.Error("paymentservice:service:handleRazorpayPaymentFailed", "error", "failed to create user purchase", "details", err)
		return fmt.Errorf("failed to create user purchase: %v", err)
	}

	slog.Info("paymentservice:service:handleRazorpayPaymentFailed", "userID", userID, "productID", productID, "data", "saved failed transaction in database")

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

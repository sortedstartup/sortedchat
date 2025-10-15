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
	"strconv"
)

func (s *PaymentService) CreateProductRazorpay(ctx context.Context, name string, description string, cost string, currency string) (string, error) {
	slog.Info("paymentservice:service:CreateProductRazorpay", "name", name)

	// Convert price string to int64 (Razorpay expects amount in smallest currency unit)
	priceAmount, err := strconv.ParseInt(cost, 10, 64)
	if err != nil {
		slog.Error("paymentservice:service:CreateProductRazorpay", "error", err)
		return "", fmt.Errorf("failed to process the request")
	}
	slog.Info("paymentservice:service:CreateProductRazorpay", "priceAmount", priceAmount)

	// Create item data for Razorpay
	itemData := map[string]interface{}{
		"name":        name,
		"description": description,
		"amount":      priceAmount * 100, // Convert to smallest currency unit (paise)
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

func (s *PaymentService) CreateRazorpayCheckoutSession(ctx context.Context, userID string, razorpayItemID string) (string, string, string, error) {
	slog.Info("paymentservice:service:CreateRazorpayCheckoutSession", "userID", userID, "itemID", razorpayItemID)

	product, err := s.dao.GetRazorpayProductById(razorpayItemID)
	if err != nil {
		slog.Error("paymentservice:service:CreateRazorpayCheckoutSession", "error", err)
		return "", "", "", fmt.Errorf("failed to create Razorpay checkout session")
	}

	priceAmount, err := strconv.ParseInt(product.Price, 10, 64)
	if err != nil {
		slog.Error("paymentservice:service:CreateRazorpayCheckoutOrder", "error", err)
		return "", "", "", fmt.Errorf("failed to parse price: %w", err)
	}

	// Razorpay expects amount in the smallest currency unit (paise for INR, cents for USD, etc.)
	orderParams := map[string]interface{}{
		"amount":   priceAmount * 100,
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
		return "", "", "", fmt.Errorf("failed to create Razorpay checkout session")
	}

	// Extract Order ID
	orderID, ok := order["id"].(string)
	if !ok {
		slog.Error("paymentservice:service:CreateRazorpayCheckoutSession", "error", "Order ID not found in response")
		return "", "", "", fmt.Errorf("failed to create Razorpay checkout session")
	}

	slog.Info("paymentservice:service:CreateRazorpayCheckoutSession", "orderID", orderID)
	return orderID, strconv.FormatInt(priceAmount*100, 10), product.Currency, nil
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
	if !s.verifySignature(payload, signature, "Test@123") { // TODO: Replace with actual webhook secret
		slog.Error("paymentservice:service:HandleWebhook", "error", "invalid signature")
		return fmt.Errorf("invalid signature")
	}

	// Get event ID for idempotency check
	eventID := r.Header.Get("x-razorpay-event-id")
	slog.Info("paymentservice:service:HandleWebhook", "event_id", eventID)

	// Parse JSON payload for logging
	var webhookData map[string]interface{}
	if err := json.Unmarshal(payload, &webhookData); err != nil {
		slog.Error("paymentservice:service:HandleWebhook", "error", "failed to parse JSON", "details", err)
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Log the complete transaction data as JSON
	transactionJSON, _ := json.MarshalIndent(webhookData, "", "  ")
	slog.Info("paymentservice:service:HandleWebhook", "transaction_data", string(transactionJSON))

	// Extract and log specific event details
	if event, ok := webhookData["event"].(string); ok {
		slog.Info("paymentservice:service:HandleWebhook", "event_type", event)

		// Log payment details if available
		if payload, ok := webhookData["payload"].(map[string]interface{}); ok {
			if payment, ok := payload["payment"].(map[string]interface{}); ok {
				if entity, ok := payment["entity"].(map[string]interface{}); ok {
					slog.Info("paymentservice:service:HandleWebhook",
						"payment_id", entity["id"],
						"amount", entity["amount"],
						"currency", entity["currency"],
						"status", entity["status"],
						"method", entity["method"])
				}
			}
		}
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

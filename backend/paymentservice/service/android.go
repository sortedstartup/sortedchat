package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/option"
)

func (s *PaymentService) HandleInAppPurchaseProduct(ctx context.Context, r *http.Request) error {
	slog.Info("paymentservice:android:HandleInAppPurchaseProduct")

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("paymentservice:android:HandleInAppPurchaseProduct", "error", "failed to read request body", "details", err)
		return fmt.Errorf("error reading request body: %w", err)
	}

	// Parse JSON payload
	var webhookData map[string]interface{}
	if err := json.Unmarshal(payload, &webhookData); err != nil {
		slog.Error("paymentservice:android:HandleInAppPurchaseProduct", "error", "failed to parse JSON", "details", err)
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	slog.Info("paymentservice:android:HandleInAppPurchaseProduct", "webhook data", webhookData)

	purchaseToken, ok := webhookData["purchaseToken"].(string)
	if !ok {
		slog.Error("paymentservice:android:HandleInAppPurchaseProduct", "error", "purchase data not found")
		return fmt.Errorf("purchase data not found")
	}

	packageName, ok := webhookData["packageName"].(string)
	if !ok {
		slog.Error("paymentservice:android:HandleInAppPurchaseProduct", "error", "package name not found")
		return fmt.Errorf("package name not found")
	}

	productID, ok := webhookData["productId"].(string)
	if !ok {
		slog.Error("paymentservice:android:HandleInAppPurchaseProduct", "error", "product id not found")
		return fmt.Errorf("product id not found")
	}

	accountId, ok := webhookData["accountId"].(string)
	if !ok {
		slog.Error("paymentservice:android:HandleInAppPurchaseProduct", "error", "account id not found")
		return fmt.Errorf("account id not found")
	}

	customProductId, ok := webhookData["customProductId"].(string)
	if !ok {
		slog.Error("paymentservice:android:HandleInAppPurchaseProduct", "error", "custom product id not found")
		return fmt.Errorf("custom product id not found")
	}

	transactionId, ok := webhookData["transactionId"].(string)
	if !ok {
		slog.Error("paymentservice:android:HandleInAppPurchaseProduct", "error", "transaction id not found")
		return fmt.Errorf("transaction id not found")
	}

	purchaseState, ok := webhookData["purchaseState"].(float64)
	if !ok {
		slog.Error("paymentservice:android:HandleInAppPurchaseProduct", "error", "purchase state not found")
		return fmt.Errorf("purchase state not found")
	}

	purchaseTime, ok := webhookData["purchaseTime"].(float64)
	if !ok {
		slog.Error("paymentservice:android:HandleInAppPurchaseProduct", "error", "purchase time not found")
		return fmt.Errorf("purchase time not found")
	}

	slog.Info("paymentservice:android:HandleInAppPurchaseProduct", "purchase token", purchaseToken)
	slog.Info("paymentservice:android:HandleInAppPurchaseProduct", "product id", productID)
	slog.Info("paymentservice:android:HandleInAppPurchaseProduct", "package name", packageName)
	slog.Info("paymentservice:android:HandleInAppPurchaseProduct", "account id", accountId)
	slog.Info("paymentservice:android:HandleInAppPurchaseProduct", "custom product id", customProductId)
	slog.Info("paymentservice:android:HandleInAppPurchaseProduct", "transaction id", transactionId)
	slog.Info("paymentservice:android:HandleInAppPurchaseProduct", "purchase state", purchaseState)

	//make an api call to the purchase product endpoint

	androidPublisherClient, err := androidpublisher.NewService(ctx, option.WithCredentialsFile("/home/sanskar/Documents/work/sortedchat/backend/paymentservice/service_account.json"))
	if err != nil {
		slog.Error("paymentservice:android:HandleInAppPurchaseProduct", "error", "failed to create android publisher client", "details", err)
		return fmt.Errorf("failed to create android publisher client: %w", err)
	}

	response, err := androidPublisherClient.Purchases.Products.Get(packageName, productID, purchaseToken).Do()
	if err != nil {
		slog.Error("paymentservice:android:HandleInAppPurchaseProduct", "error", "failed to get purchase product", "details", err)
		return fmt.Errorf("failed to get purchase product: %w", err)
	}
	slog.Info("paymentservice:android:HandleInAppPurchaseProduct", "purchase product response", response)

	//save it in db
	subscriptionID, err := s.dao.CreateSubscription(
		transactionId,
		accountId,
		customProductId,
		"android",
		purchaseToken,
		"providerCustomerId", //think of it
		"purchaseState",      //in android comig as int
		"active",
		int64(purchaseTime),
		int64(purchaseTime),
		false,
		false,
	)
	if err != nil {
		slog.Error("paymentservice:android:HandleInAppPurchaseProduct", "error", "failed to create subscription", "details", err)
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	slog.Info("paymentservice:android:HandleInAppPurchaseProduct", "developer payload", response.DeveloperPayload)

	webhookBytes, err := json.Marshal(webhookData)
	if err != nil {
		slog.Error("paymentservice:android:HandleInAppPurchaseProduct", "error", "failed to marshal webhook data", "details", err)
		return fmt.Errorf("failed to marshal webhook data: %w", err)
	}

	webhookMetadata := string(webhookBytes)

	paymentId, err := s.dao.CreateUserPayment(
		accountId,
		customProductId,
		subscriptionID,
		transactionId,
		webhookMetadata,
		true,
	)
	if err != nil {
		slog.Error("paymentservice:android:HandleInAppPurchaseProduct", "error", "failed to create user payment", "details", err)
		return fmt.Errorf("failed to create user payment: %w", err)
	}

	acknowledgeErr := androidPublisherClient.Purchases.Products.Acknowledge(packageName, productID, purchaseToken, &androidpublisher.ProductPurchasesAcknowledgeRequest{
		DeveloperPayload: response.DeveloperPayload,
	}).Do()
	if acknowledgeErr != nil {
		slog.Error("paymentservice:android:HandleInAppPurchaseProduct", "error", "failed to acknowledge purchase product", "details", acknowledgeErr)
		return fmt.Errorf("failed to acknowledge purchase product: %w", acknowledgeErr)
	}

	slog.Info("paymentservice:android:HandleInAppPurchaseProduct", "package name", packageName, "subscriptionId", subscriptionID, "paymentId", paymentId)
	return nil
}

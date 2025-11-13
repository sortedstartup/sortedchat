package api

import (
	"log/slog"
	"net/http"
)

func (s *PaymentServiceAPI) registerRoutes(mux *http.ServeMux) {
	slog.Info("api:registerRoutes")
	mux.HandleFunc("/stripe-webhook", s.handleWebhook)
	mux.HandleFunc("/razorpay-webhook", s.handleRazorpayWebhook)
	mux.HandleFunc("/inapp-purchase-product", s.handleInAppPurchaseProduct)
}

func (s *PaymentServiceAPI) handleWebhook(w http.ResponseWriter, r *http.Request) {
	slog.Info("api:handleWebhook")

	err := s.service.HandleStripeWebhook(r.Context(), r)
	if err != nil {
		slog.Error("api:handleWebhook", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook received"))

}

func (s *PaymentServiceAPI) handleRazorpayWebhook(w http.ResponseWriter, r *http.Request) {
	slog.Info("api:handleRazorpayWebhook")
	err := s.service.HandleRazorpayWebhook(r.Context(), r)
	if err != nil {
		slog.Error("api:handleRazorpayWebhook", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *PaymentServiceAPI) handleInAppPurchaseProduct(w http.ResponseWriter, r *http.Request) {
	slog.Info("api:handleInAppPurchaseProduct")
	err := s.service.HandleInAppPurchaseProduct(r.Context(), r)
	if err != nil {
		slog.Error("api:handleInAppPurchaseProduct", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

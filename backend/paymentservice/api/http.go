package api

import (
	"log/slog"
	"net/http"
)

func (s *PaymentServiceAPI) registerRoutes(mux *http.ServeMux) {
	slog.Info("api:registerRoutes")
	mux.HandleFunc("/webhook", s.handleWebhook)
}

func (s *PaymentServiceAPI) handleWebhook(w http.ResponseWriter, r *http.Request) {
	slog.Info("api:handleWebhook")

	err := s.service.HandleWebhook(r.Context(), r)
	if err != nil {
		slog.Error("api:handleWebhook", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook received"))

}

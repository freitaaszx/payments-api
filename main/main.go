package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/seu-usuario/payments-api/internal/handler"
	"github.com/seu-usuario/payments-api/internal/middleware"
	"github.com/seu-usuario/payments-api/internal/service"
	"github.com/seu-usuario/payments-api/internal/store"
)

func main() {
	port   := envOr("PORT", "8080")
	apiKey := envOr("API_KEY", "sk_test_mock_secret") // vazio = sem auth

	// Wiring
	st  := store.New()
	svc := service.New(st)
	h   := handler.New(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Middleware stack: Logger → CORS → APIKeyAuth → handler
	stack := middleware.Chain(
		mux,
		middleware.Logger,
		middleware.CORS,
		func(next http.Handler) http.Handler {
			return middleware.APIKeyAuth(apiKey, next)
		},
	)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("🚀  payments-api (mock) rodando em http://localhost%s", addr)
	log.Printf("🔑  API Key: %s", apiKey)
	log.Printf("📋  Rotas disponíveis:")
	log.Printf("    GET    /health")
	log.Printf("    POST   /v1/payments")
	log.Printf("    GET    /v1/payments")
	log.Printf("    GET    /v1/payments/{id}")
	log.Printf("    POST   /v1/payments/{id}/capture")
	log.Printf("    POST   /v1/payments/{id}/cancel")
	log.Printf("    POST   /v1/payments/{id}/refunds")
	log.Printf("    GET    /v1/payments/{id}/refunds")
	log.Printf("    GET    /v1/refunds/{id}")

	if err := http.ListenAndServe(addr, stack); err != nil {
		log.Fatalf("erro ao iniciar servidor: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

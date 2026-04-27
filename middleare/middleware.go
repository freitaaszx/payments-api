package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"
)

// Logger loga método, path, status e duração de cada request.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		log.Printf("[%s] %s %s → %d (%s)",
			time.Now().Format("15:04:05"),
			r.Method, r.URL.Path,
			rw.status,
			time.Since(start).Round(time.Microsecond),
		)
	})
}

// CORS adiciona os headers necessários para chamadas cross-origin.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Api-Key")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// APIKeyAuth exige o header X-Api-Key com o valor configurado.
// Passe uma apiKey vazia para desabilitar (útil em dev).
func APIKeyAuth(apiKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Health check nunca exige auth
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}
		if apiKey == "" {
			next.ServeHTTP(w, r)
			return
		}

		key := r.Header.Get("X-Api-Key")
		if key == "" {
			// Tenta extrair do header Authorization: Bearer <key>
			auth := r.Header.Get("Authorization")
			key = strings.TrimPrefix(auth, "Bearer ")
		}

		if key != apiKey {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"code":"UNAUTHORIZED","message":"invalid or missing API key"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Chain encadeia middlewares da esquerda para a direita.
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// responseWriter captura o status code para o logger.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

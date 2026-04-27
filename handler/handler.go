package handler

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/seu-usuario/payments-api/internal/models"
	"github.com/seu-usuario/payments-api/internal/service"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

// ─── Router ───────────────────────────────────────────────────────────────────

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Payments
	mux.HandleFunc("/v1/payments", h.handlePayments)
	mux.HandleFunc("/v1/payments/", h.handlePaymentByID)

	// Refunds
	mux.HandleFunc("/v1/refunds/", h.handleRefundByID)

	// Health
	mux.HandleFunc("/health", h.handleHealth)
}

// ─── /v1/payments ─────────────────────────────────────────────────────────────

func (h *Handler) handlePayments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listPayments(w, r)
	case http.MethodPost:
		h.createPayment(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
	}
}

func (h *Handler) createPayment(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	payment, err := h.svc.CreatePayment(req)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "PAYMENT_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, payment)
}

func (h *Handler) listPayments(w http.ResponseWriter, r *http.Request) {
	q := models.ListPaymentsQuery{
		Status:   models.PaymentStatus(r.URL.Query().Get("status")),
		Method:   models.PaymentMethod(r.URL.Query().Get("method")),
		Page:     queryInt(r, "page", 1),
		PageSize: queryInt(r, "page_size", 20),
	}

	payments, total := h.svc.ListPayments(q)

	resp := models.PaginatedResponse[*models.Payment]{
		Data:       payments,
		Total:      total,
		Page:       q.Page,
		PageSize:   q.PageSize,
		TotalPages: int(math.Ceil(float64(total) / float64(q.PageSize))),
	}
	writeJSON(w, http.StatusOK, resp)
}

// ─── /v1/payments/{id}[/action] ──────────────────────────────────────────────

func (h *Handler) handlePaymentByID(w http.ResponseWriter, r *http.Request) {
	// Parse: /v1/payments/{id}[/capture|cancel|refunds]
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/payments/"), "/")
	id := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch {
	case action == "" && r.Method == http.MethodGet:
		h.getPayment(w, r, id)

	case action == "capture" && r.Method == http.MethodPost:
		h.capturePayment(w, r, id)

	case action == "cancel" && r.Method == http.MethodPost:
		h.cancelPayment(w, r, id)

	case action == "refunds" && r.Method == http.MethodPost:
		h.refundPayment(w, r, id)

	case action == "refunds" && r.Method == http.MethodGet:
		h.listRefunds(w, r, id)

	default:
		writeError(w, http.StatusNotFound, "NOT_FOUND", "route not found")
	}
}

func (h *Handler) getPayment(w http.ResponseWriter, r *http.Request, id string) {
	payment, err := h.svc.GetPayment(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "PAYMENT_NOT_FOUND", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, payment)
}

func (h *Handler) capturePayment(w http.ResponseWriter, r *http.Request, id string) {
	payment, err := h.svc.CapturePayment(id)
	if err != nil {
		code, errCode := classifyError(err)
		writeError(w, code, errCode, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, payment)
}

func (h *Handler) cancelPayment(w http.ResponseWriter, r *http.Request, id string) {
	payment, err := h.svc.CancelPayment(id)
	if err != nil {
		code, errCode := classifyError(err)
		writeError(w, code, errCode, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, payment)
}

func (h *Handler) refundPayment(w http.ResponseWriter, r *http.Request, id string) {
	var req models.RefundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	refund, err := h.svc.RefundPayment(id, req)
	if err != nil {
		code, errCode := classifyError(err)
		writeError(w, code, errCode, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, refund)
}

func (h *Handler) listRefunds(w http.ResponseWriter, r *http.Request, paymentID string) {
	refunds, err := h.svc.ListRefunds(paymentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "PAYMENT_NOT_FOUND", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":  refunds,
		"total": len(refunds),
	})
}

// ─── /v1/refunds/{id} ─────────────────────────────────────────────────────────

func (h *Handler) handleRefundByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/v1/refunds/")
	refund, err := h.svc.GetRefund(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "REFUND_NOT_FOUND", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, refund)
}

// ─── /health ──────────────────────────────────────────────────────────────────

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "payments-api",
		"version": "1.0.0",
	})
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, models.ErrorResponse{Code: code, Message: message})
}

func queryInt(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func classifyError(err error) (int, string) {
	msg := err.Error()
	switch {
	case errors.Is(err, errors.New("not found")) || strings.Contains(msg, "not found"):
		return http.StatusNotFound, "NOT_FOUND"
	case strings.Contains(msg, "cannot be") || strings.Contains(msg, "only"):
		return http.StatusUnprocessableEntity, "INVALID_STATE"
	default:
		return http.StatusBadRequest, "BAD_REQUEST"
	}
}

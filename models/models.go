package models

import (
	"time"
)

// ─── Enums ────────────────────────────────────────────────────────────────────

type PaymentStatus string
type PaymentMethod string
type Currency string

const (
	StatusPending    PaymentStatus = "pending"
	StatusProcessing PaymentStatus = "processing"
	StatusApproved   PaymentStatus = "approved"
	StatusDeclined   PaymentStatus = "declined"
	StatusRefunded   PaymentStatus = "refunded"
	StatusCancelled  PaymentStatus = "cancelled"

	MethodCreditCard  PaymentMethod = "credit_card"
	MethodDebitCard   PaymentMethod = "debit_card"
	MethodPix         PaymentMethod = "pix"
	MethodBoleto      PaymentMethod = "boleto"
	MethodBankTransfer PaymentMethod = "bank_transfer"

	CurrencyBRL Currency = "BRL"
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
)

// ─── Core Models ──────────────────────────────────────────────────────────────

type Payment struct {
	ID          string        `json:"id"`
	Amount      float64       `json:"amount"`
	Currency    Currency      `json:"currency"`
	Status      PaymentStatus `json:"status"`
	Method      PaymentMethod `json:"method"`
	Description string        `json:"description,omitempty"`
	Customer    Customer      `json:"customer"`
	Card        *CardInfo     `json:"card,omitempty"`
	Pix         *PixInfo      `json:"pix,omitempty"`
	Boleto      *BoletoInfo   `json:"boleto,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	PaidAt      *time.Time    `json:"paid_at,omitempty"`
	RefundedAt  *time.Time    `json:"refunded_at,omitempty"`
}

type Customer struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Document string `json:"document"` // CPF ou CNPJ
	Phone    string `json:"phone,omitempty"`
}

type CardInfo struct {
	Last4      string `json:"last4"`
	Brand      string `json:"brand"`       // visa, mastercard, elo, etc.
	HolderName string `json:"holder_name"`
	ExpiryMonth int   `json:"expiry_month"`
	ExpiryYear  int   `json:"expiry_year"`
}

type PixInfo struct {
	Key        string    `json:"key"`
	QRCode     string    `json:"qr_code"`
	QRCodeURL  string    `json:"qr_code_url"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type BoletoInfo struct {
	Code      string    `json:"code"`
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Refund struct {
	ID        string    `json:"id"`
	PaymentID string    `json:"payment_id"`
	Amount    float64   `json:"amount"`
	Reason    string    `json:"reason,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// ─── Request / Response DTOs ──────────────────────────────────────────────────

type CreatePaymentRequest struct {
	Amount      float64           `json:"amount"`
	Currency    Currency          `json:"currency"`
	Method      PaymentMethod     `json:"method"`
	Description string            `json:"description,omitempty"`
	Customer    Customer          `json:"customer"`
	Card        *CardRequest      `json:"card,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type CardRequest struct {
	Number      string `json:"number"`
	HolderName  string `json:"holder_name"`
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
	CVV         string `json:"cvv"`
}

type RefundRequest struct {
	Amount float64 `json:"amount,omitempty"` // 0 = reembolso total
	Reason string  `json:"reason,omitempty"`
}

type ListPaymentsQuery struct {
	Status   PaymentStatus `json:"status,omitempty"`
	Method   PaymentMethod `json:"method,omitempty"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

type PaginatedResponse[T any] struct {
	Data       []T   `json:"data"`
	Total      int   `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

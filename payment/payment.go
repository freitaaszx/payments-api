package service

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/seu-usuario/payments-api/internal/models"
	"github.com/seu-usuario/payments-api/internal/store"
	"github.com/seu-usuario/payments-api/pkg/idgen"
)

// Cartões especiais que simulam comportamentos específicos (como Stripe test cards).
var mockCardBehaviors = map[string]models.PaymentStatus{
	"4000000000000002": models.StatusDeclined,   // sempre recusado
	"4000000000009995": models.StatusDeclined,   // saldo insuficiente
	"4111111111111111": models.StatusApproved,   // sempre aprovado (Visa)
	"5500000000000004": models.StatusApproved,   // sempre aprovado (MC)
}

type Service struct {
	store *store.Store
}

func New(s *store.Store) *Service {
	return &Service{store: s}
}

// ─── Create ───────────────────────────────────────────────────────────────────

func (svc *Service) CreatePayment(req models.CreatePaymentRequest) (*models.Payment, error) {
	if err := validateCreateRequest(req); err != nil {
		return nil, err
	}

	now := time.Now()
	payment := &models.Payment{
		ID:          idgen.New("pay"),
		Amount:      req.Amount,
		Currency:    req.Currency,
		Method:      req.Method,
		Description: req.Description,
		Customer:    req.Customer,
		Metadata:    req.Metadata,
		Status:      models.StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Processa de acordo com o método de pagamento
	switch req.Method {
	case models.MethodCreditCard, models.MethodDebitCard:
		if req.Card == nil {
			return nil, errors.New("card data is required for card payments")
		}
		payment.Card = maskCard(req.Card)
		payment.Status = simulateCardApproval(req.Card.Number)

	case models.MethodPix:
		payment.Pix = generatePixInfo()
		payment.Status = models.StatusPending // aguarda pagamento

	case models.MethodBoleto:
		payment.Boleto = generateBoletoInfo()
		payment.Status = models.StatusPending // aguarda pagamento

	case models.MethodBankTransfer:
		payment.Status = models.StatusPending
	}

	if payment.Status == models.StatusApproved {
		t := now
		payment.PaidAt = &t
	}

	svc.store.SavePayment(payment)
	return payment, nil
}

// ─── Get ──────────────────────────────────────────────────────────────────────

func (svc *Service) GetPayment(id string) (*models.Payment, error) {
	return svc.store.GetPayment(id)
}

// ─── List ─────────────────────────────────────────────────────────────────────

func (svc *Service) ListPayments(q models.ListPaymentsQuery) ([]*models.Payment, int) {
	return svc.store.ListPayments(q.Status, q.Method, q.Page, q.PageSize)
}

// ─── Capture (para pagamentos pre-autorizados) ────────────────────────────────

func (svc *Service) CapturePayment(id string) (*models.Payment, error) {
	payment, err := svc.store.GetPayment(id)
	if err != nil {
		return nil, err
	}
	if payment.Status != models.StatusProcessing {
		return nil, fmt.Errorf("payment cannot be captured: current status is %s", payment.Status)
	}

	now := time.Now()
	payment.Status = models.StatusApproved
	payment.PaidAt = &now
	payment.UpdatedAt = now
	svc.store.SavePayment(payment)
	return payment, nil
}

// ─── Cancel ───────────────────────────────────────────────────────────────────

func (svc *Service) CancelPayment(id string) (*models.Payment, error) {
	payment, err := svc.store.GetPayment(id)
	if err != nil {
		return nil, err
	}
	if payment.Status != models.StatusPending && payment.Status != models.StatusProcessing {
		return nil, fmt.Errorf("only pending/processing payments can be cancelled, current status: %s", payment.Status)
	}

	payment.Status = models.StatusCancelled
	payment.UpdatedAt = time.Now()
	svc.store.SavePayment(payment)
	return payment, nil
}

// ─── Refund ───────────────────────────────────────────────────────────────────

func (svc *Service) RefundPayment(paymentID string, req models.RefundRequest) (*models.Refund, error) {
	payment, err := svc.store.GetPayment(paymentID)
	if err != nil {
		return nil, err
	}
	if payment.Status != models.StatusApproved {
		return nil, fmt.Errorf("only approved payments can be refunded, current status: %s", payment.Status)
	}

	refundAmount := req.Amount
	if refundAmount == 0 {
		refundAmount = payment.Amount // reembolso total
	}
	if refundAmount > payment.Amount {
		return nil, fmt.Errorf("refund amount (%.2f) exceeds payment amount (%.2f)", refundAmount, payment.Amount)
	}

	now := time.Now()
	refund := &models.Refund{
		ID:        idgen.New("ref"),
		PaymentID: paymentID,
		Amount:    refundAmount,
		Reason:    req.Reason,
		Status:    "approved",
		CreatedAt: now,
	}

	payment.Status = models.StatusRefunded
	payment.RefundedAt = &now
	payment.UpdatedAt = now

	svc.store.SaveRefund(refund)
	svc.store.SavePayment(payment)
	return refund, nil
}

// ─── Get Refund ───────────────────────────────────────────────────────────────

func (svc *Service) GetRefund(id string) (*models.Refund, error) {
	return svc.store.GetRefund(id)
}

func (svc *Service) ListRefunds(paymentID string) ([]*models.Refund, error) {
	// Valida que o pagamento existe
	if _, err := svc.store.GetPayment(paymentID); err != nil {
		return nil, err
	}
	return svc.store.ListRefundsByPayment(paymentID), nil
}

// ─── Mock helpers ─────────────────────────────────────────────────────────────

func simulateCardApproval(cardNumber string) models.PaymentStatus {
	// Remove espaços do número
	number := strings.ReplaceAll(cardNumber, " ", "")

	if status, ok := mockCardBehaviors[number]; ok {
		return status
	}

	// 90% de aprovação aleatória para outros cartões
	if rand.Float32() < 0.9 {
		return models.StatusApproved
	}
	return models.StatusDeclined
}

func maskCard(req *models.CardRequest) *models.CardInfo {
	number := strings.ReplaceAll(req.Number, " ", "")
	last4 := number
	if len(number) >= 4 {
		last4 = number[len(number)-4:]
	}
	return &models.CardInfo{
		Last4:       last4,
		Brand:       detectBrand(number),
		HolderName:  req.HolderName,
		ExpiryMonth: req.ExpiryMonth,
		ExpiryYear:  req.ExpiryYear,
	}
}

func detectBrand(number string) string {
	switch {
	case strings.HasPrefix(number, "4"):
		return "visa"
	case strings.HasPrefix(number, "5"):
		return "mastercard"
	case strings.HasPrefix(number, "34") || strings.HasPrefix(number, "37"):
		return "amex"
	case strings.HasPrefix(number, "6011"):
		return "discover"
	case strings.HasPrefix(number, "636"):
		return "elo"
	default:
		return "unknown"
	}
}

func generatePixInfo() *models.PixInfo {
	key := fmt.Sprintf("mock-pix-%s@loja.com", idgen.Short())
	qr := fmt.Sprintf("00020126580014br.gov.bcb.pix0136%s5204000053039865802BR5915LOJA MOVEIS6009SAO PAULO62140510%s6304ABCD",
		key, idgen.Short())
	return &models.PixInfo{
		Key:       key,
		QRCode:    qr,
		QRCodeURL: fmt.Sprintf("https://mock-pix.example.com/qr/%s", idgen.Short()),
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
}

func generateBoletoInfo() *models.BoletoInfo {
	code := fmt.Sprintf("34191.09008 61207.727308 71140.063308 1 %d%016d",
		time.Now().Year(), rand.Int63n(9999999999999999))
	return &models.BoletoInfo{
		Code:      code,
		URL:       fmt.Sprintf("https://mock-boleto.example.com/%s", idgen.Short()),
		ExpiresAt: time.Now().Add(3 * 24 * time.Hour),
	}
}

// ─── Validation ───────────────────────────────────────────────────────────────

func validateCreateRequest(req models.CreatePaymentRequest) error {
	if req.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	if req.Currency == "" {
		return errors.New("currency is required")
	}
	if req.Method == "" {
		return errors.New("payment method is required")
	}
	if req.Customer.Name == "" || req.Customer.Email == "" {
		return errors.New("customer name and email are required")
	}
	validMethods := map[models.PaymentMethod]bool{
		models.MethodCreditCard: true, models.MethodDebitCard: true,
		models.MethodPix: true, models.MethodBoleto: true,
		models.MethodBankTransfer: true,
	}
	if !validMethods[req.Method] {
		return fmt.Errorf("invalid payment method: %s", req.Method)
	}
	return nil
}

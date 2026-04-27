package store

import (
	"fmt"
	"sync"

	"github.com/seu-usuario/payments-api/internal/models"
)

// Store é um repositório em memória thread-safe para pagamentos e reembolsos.
type Store struct {
	mu       sync.RWMutex
	payments map[string]*models.Payment
	refunds  map[string]*models.Refund
}

func New() *Store {
	return &Store{
		payments: make(map[string]*models.Payment),
		refunds:  make(map[string]*models.Refund),
	}
}

// ─── Payments ─────────────────────────────────────────────────────────────────

func (s *Store) SavePayment(p *models.Payment) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.payments[p.ID] = p
}

func (s *Store) GetPayment(id string) (*models.Payment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.payments[id]
	if !ok {
		return nil, fmt.Errorf("payment not found: %s", id)
	}
	return p, nil
}

func (s *Store) ListPayments(status models.PaymentStatus, method models.PaymentMethod, page, pageSize int) ([]*models.Payment, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []*models.Payment
	for _, p := range s.payments {
		if status != "" && p.Status != status {
			continue
		}
		if method != "" && p.Method != method {
			continue
		}
		filtered = append(filtered, p)
	}

	total := len(filtered)
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}

	start := (page - 1) * pageSize
	if start >= total {
		return []*models.Payment{}, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	return filtered[start:end], total
}

// ─── Refunds ──────────────────────────────────────────────────────────────────

func (s *Store) SaveRefund(r *models.Refund) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refunds[r.ID] = r
}

func (s *Store) GetRefund(id string) (*models.Refund, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	r, ok := s.refunds[id]
	if !ok {
		return nil, fmt.Errorf("refund not found: %s", id)
	}
	return r, nil
}

func (s *Store) ListRefundsByPayment(paymentID string) []*models.Refund {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*models.Refund
	for _, r := range s.refunds {
		if r.PaymentID == paymentID {
			result = append(result, r)
		}
	}
	return result
}

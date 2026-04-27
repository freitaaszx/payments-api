
# 💳 Payments API Mock — Go

API REST de pagamentos mock com suporte a cartão, Pix, boleto e reembolso.
Sem dependências externas — apenas a stdlib do Go.

## Estrutura

```
payments-api/
├── cmd/api/main.go              # Entrypoint
├── internal/
│   ├── handler/handler.go       # HTTP handlers
│   ├── middleware/middleware.go  # Logger, CORS, Auth
│   ├── models/models.go         # Domínio / DTOs
│   ├── service/payment.go       # Lógica de negócio
│   └── store/store.go           # Repositório em memória
├── pkg/idgen/idgen.go           # Gerador de IDs
└── go.mod
```

## Executar

```bash
go run ./cmd/api

# Variáveis de ambiente opcionais
PORT=9000 API_KEY=minha-chave go run ./cmd/api
```

## Autenticação

Passe a chave em qualquer header:
```
X-Api-Key: sk_test_mock_secret
# ou
Authorization: Bearer sk_test_mock_secret
```

---

## 🧪 Cartões de teste

| Número             | Comportamento         |
|--------------------|-----------------------|
| 4111111111111111   | ✅ Sempre aprovado (Visa) |
| 5500000000000004   | ✅ Sempre aprovado (MC)   |
| 4000000000000002   | ❌ Sempre recusado        |
| 4000000000009995   | ❌ Saldo insuficiente      |
| Qualquer outro     | 90% aprovado (aleatório)  |

---

## 📋 Exemplos com cURL

### Health check
```bash
curl http://localhost:8080/health
```

---

### Criar pagamento — Cartão de Crédito
```bash
curl -X POST http://localhost:8080/v1/payments \
  -H "Content-Type: application/json" \
  -H "X-Api-Key: sk_test_mock_secret" \
  -d '{
    "amount": 299.90,
    "currency": "BRL",
    "method": "credit_card",
    "description": "Cadeira Estofada Premium",
    "customer": {
      "id": "cus_001",
      "name": "João Silva",
      "email": "joao@email.com",
      "document": "123.456.789-00"
    },
    "card": {
      "number": "4111111111111111",
      "holder_name": "JOAO SILVA",
      "expiry_month": 12,
      "expiry_year": 2027,
      "cvv": "123"
    },
    "metadata": {
      "order_id": "ORD-555"
    }
  }'
```

---

### Criar pagamento — Pix
```bash
curl -X POST http://localhost:8080/v1/payments \
  -H "Content-Type: application/json" \
  -H "X-Api-Key: sk_test_mock_secret" \
  -d '{
    "amount": 150.00,
    "currency": "BRL",
    "method": "pix",
    "customer": {
      "id": "cus_002",
      "name": "Maria Souza",
      "email": "maria@email.com",
      "document": "987.654.321-00"
    }
  }'
```

---

### Criar pagamento — Boleto
```bash
curl -X POST http://localhost:8080/v1/payments \
  -H "Content-Type: application/json" \
  -H "X-Api-Key: sk_test_mock_secret" \
  -d '{
    "amount": 899.00,
    "currency": "BRL",
    "method": "boleto",
    "customer": {
      "id": "cus_003",
      "name": "Carlos Lima",
      "email": "carlos@email.com",
      "document": "111.222.333-44"
    }
  }'
```

---

### Buscar pagamento
```bash
curl http://localhost:8080/v1/payments/pay_1712345678_a3f9b2c1 \
  -H "X-Api-Key: sk_test_mock_secret"
```

---

### Listar pagamentos (com filtros)
```bash
# Todos
curl "http://localhost:8080/v1/payments" \
  -H "X-Api-Key: sk_test_mock_secret"

# Apenas aprovados, página 1
curl "http://localhost:8080/v1/payments?status=approved&page=1&page_size=10" \
  -H "X-Api-Key: sk_test_mock_secret"

# Apenas Pix
curl "http://localhost:8080/v1/payments?method=pix" \
  -H "X-Api-Key: sk_test_mock_secret"
```

---

### Cancelar pagamento
```bash
curl -X POST http://localhost:8080/v1/payments/pay_ID/cancel \
  -H "X-Api-Key: sk_test_mock_secret"
```

---

### Capturar pagamento (pre-autorizado)
```bash
curl -X POST http://localhost:8080/v1/payments/pay_ID/capture \
  -H "X-Api-Key: sk_test_mock_secret"
```

---

### Reembolso total
```bash
curl -X POST http://localhost:8080/v1/payments/pay_ID/refunds \
  -H "Content-Type: application/json" \
  -H "X-Api-Key: sk_test_mock_secret" \
  -d '{"reason": "cliente desistiu"}'
```

### Reembolso parcial
```bash
curl -X POST http://localhost:8080/v1/payments/pay_ID/refunds \
  -H "Content-Type: application/json" \
  -H "X-Api-Key: sk_test_mock_secret" \
  -d '{"amount": 50.00, "reason": "item avariado"}'
```

---

### Listar reembolsos de um pagamento
```bash
curl http://localhost:8080/v1/payments/pay_ID/refunds \
  -H "X-Api-Key: sk_test_mock_secret"
```

### Buscar reembolso por ID
```bash
curl http://localhost:8080/v1/refunds/ref_ID \
  -H "X-Api-Key: sk_test_mock_secret"
```

---

## Status de pagamento

| Status       | Descrição                            |
|--------------|--------------------------------------|
| `pending`    | Aguardando pagamento (Pix/Boleto)    |
| `processing` | Em processamento                     |
| `approved`   | Aprovado e pago                      |
| `declined`   | Recusado (cartão)                    |
| `refunded`   | Reembolsado                          |
| `cancelled`  | Cancelado                            |

## Métodos suportados

`credit_card` · `debit_card` · `pix` · `boleto` · `bank_transfer`

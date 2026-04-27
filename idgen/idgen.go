package idgen

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// New gera um ID com prefixo no formato: pay_1712345678_a3f9b2c1
func New(prefix string) string {
	return fmt.Sprintf("%s_%d_%s", prefix, time.Now().Unix(), Short())
}

// Short gera um token aleatório curto de 8 caracteres hex.
func Short() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

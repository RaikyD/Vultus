package stream

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateKey returns a cryptographically random 32-character hex stream key.
func GenerateKey() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

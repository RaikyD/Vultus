package stream

import (
	"encoding/hex"
	"testing"
)

func TestGenerateKey_Format(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey error: %v", err)
	}
	if len(key) != 32 {
		t.Fatalf("expected 32 hex chars, got %d: %q", len(key), key)
	}
	if _, err := hex.DecodeString(key); err != nil {
		t.Fatalf("key is not valid hex: %v", err)
	}
}

func TestGenerateKey_Unique(t *testing.T) {
	const n = 100
	seen := make(map[string]bool, n)
	for i := range n {
		key, err := GenerateKey()
		if err != nil {
			t.Fatalf("GenerateKey error at iteration %d: %v", i, err)
		}
		if seen[key] {
			t.Fatalf("duplicate key generated at iteration %d: %q", i, key)
		}
		seen[key] = true
	}
}

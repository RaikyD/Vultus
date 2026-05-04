package auth

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("supersecret")
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	if hash == "supersecret" {
		t.Fatal("hash must not equal plaintext")
	}
	if !strings.HasPrefix(hash, "$2a$") {
		t.Fatalf("expected bcrypt hash, got: %s", hash)
	}
}

func TestCheckPassword(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	if !CheckPassword(hash, "correct-horse-battery") {
		t.Error("CheckPassword returned false for correct password")
	}
	if CheckPassword(hash, "wrong-password") {
		t.Error("CheckPassword returned true for wrong password")
	}
	if CheckPassword(hash, "") {
		t.Error("CheckPassword returned true for empty password")
	}
}

func TestHashPassword_DifferentSalts(t *testing.T) {
	h1, _ := HashPassword("same-password")
	h2, _ := HashPassword("same-password")
	if h1 == h2 {
		t.Error("two hashes of the same password should differ (bcrypt uses random salt)")
	}
}

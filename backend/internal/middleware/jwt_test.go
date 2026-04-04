package middleware

import (
	"testing"
	"time"
)

func TestGenerateAndValidateToken(t *testing.T) {
	secret := []byte("test-secret-key-for-jwt-signing!")

	token, err := GenerateToken(secret, "user-123", 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := ValidateToken(secret, token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.UserID != "user-123" {
		t.Errorf("expected user-123, got %s", claims.UserID)
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	secret := []byte("test-secret-key-for-jwt-signing!")

	token, err := GenerateToken(secret, "user-123", -1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = ValidateToken(secret, token)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	secret1 := []byte("secret-key-one-for-signing-test!")
	secret2 := []byte("secret-key-two-for-signing-test!")

	token, err := GenerateToken(secret1, "user-123", 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = ValidateToken(secret2, token)
	if err == nil {
		t.Error("expected error for invalid signature")
	}
}

func TestValidateToken_InvalidFormat(t *testing.T) {
	secret := []byte("test-secret-key-for-jwt-signing!")

	_, err := ValidateToken(secret, "not-a-valid-token")
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestValidateToken_TamperedPayload(t *testing.T) {
	secret := []byte("test-secret-key-for-jwt-signing!")

	token, err := GenerateToken(secret, "user-123", 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Tamper with the payload
	parts := token
	tampered := "dGFtcGVyZWQ" + parts[10:]

	_, err = ValidateToken(secret, tampered)
	if err == nil {
		t.Error("expected error for tampered token")
	}
}

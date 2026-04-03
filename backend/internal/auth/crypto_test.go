package auth

import (
	"crypto/rand"
	"testing"
)

func TestTokenEncryptor_RoundTrip(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	enc, err := NewTokenEncryptor(key)
	if err != nil {
		t.Fatalf("failed to create encryptor: %v", err)
	}

	tests := []struct {
		name      string
		plaintext string
	}{
		{"simple token", "ya29.a0AfH6SMBx..."},
		{"empty string", ""},
		{"unicode", "テストトークン"},
		{"long token", "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := enc.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("encrypt failed: %v", err)
			}

			if encrypted == tt.plaintext && tt.plaintext != "" {
				t.Error("encrypted text should differ from plaintext")
			}

			decrypted, err := enc.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("decrypt failed: %v", err)
			}

			if decrypted != tt.plaintext {
				t.Errorf("expected %q, got %q", tt.plaintext, decrypted)
			}
		})
	}
}

func TestTokenEncryptor_DifferentCiphertexts(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	enc, err := NewTokenEncryptor(key)
	if err != nil {
		t.Fatalf("failed to create encryptor: %v", err)
	}

	plaintext := "same-token"
	enc1, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	enc2, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	if enc1 == enc2 {
		t.Error("encrypting same plaintext should produce different ciphertexts (random nonce)")
	}
}

func TestNewTokenEncryptor_InvalidKeySize(t *testing.T) {
	_, err := NewTokenEncryptor([]byte("short"))
	if err == nil {
		t.Error("expected error for invalid key size")
	}
}

func TestTokenEncryptor_DecryptInvalid(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	enc, err := NewTokenEncryptor(key)
	if err != nil {
		t.Fatalf("failed to create encryptor: %v", err)
	}

	_, err = enc.Decrypt("not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}

	_, err = enc.Decrypt("aGVsbG8=")
	if err == nil {
		t.Error("expected error for invalid ciphertext")
	}
}

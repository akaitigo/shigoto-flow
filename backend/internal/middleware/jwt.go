package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type TokenClaims struct {
	UserID    string `json:"sub"`
	ExpiresAt int64  `json:"exp"`
}

func GenerateToken(secret []byte, userID string, duration time.Duration) (string, error) {
	claims := TokenClaims{
		UserID:    userID,
		ExpiresAt: time.Now().Add(duration).Unix(),
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}

	payload := base64.RawURLEncoding.EncodeToString(claimsJSON)
	sig := sign(secret, payload)

	return payload + "." + sig, nil
}

func ValidateToken(secret []byte, token string) (*TokenClaims, error) {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid token format")
	}

	payload, signature := parts[0], parts[1]

	expectedSig := sign(secret, payload)
	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		return nil, fmt.Errorf("invalid token signature")
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	var claims TokenClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	if claims.ExpiresAt < time.Now().Unix() {
		return nil, fmt.Errorf("token expired")
	}

	if claims.UserID == "" {
		return nil, fmt.Errorf("token missing user ID")
	}

	return &claims, nil
}

func sign(secret []byte, payload string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

package handler

import (
	"net/http"
	"strings"
	"testing"
)

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  string
	}{
		{name: "typical address", email: "user@example.com", want: "***@example.com"},
		{name: "subdomain", email: "a.b@mail.corp.example", want: "***@mail.corp.example"},
		{name: "empty", email: "", want: "***"},
		{name: "no at sign", email: "notanemail", want: "***"},
		{name: "at start", email: "@example.com", want: "***"},
		{name: "at end", email: "user@", want: "***"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskEmail(tt.email)
			if got != tt.want {
				t.Errorf("maskEmail(%q) = %q, want %q", tt.email, got, tt.want)
			}
			// The masked value must never contain the local part of the address.
			if local, _, ok := strings.Cut(tt.email, "@"); ok && local != "" {
				if strings.Contains(got, local) {
					t.Errorf("masked value %q leaked local part %q", got, local)
				}
			}
		})
	}
}

func TestParseSameSite(t *testing.T) {
	tests := []struct {
		value string
		want  http.SameSite
	}{
		{"lax", http.SameSiteLaxMode},
		{"Lax", http.SameSiteLaxMode},
		{"strict", http.SameSiteStrictMode},
		{"STRICT", http.SameSiteStrictMode},
		{"none", http.SameSiteNoneMode},
		{" none ", http.SameSiteNoneMode},
		{"", http.SameSiteLaxMode},
		{"garbage", http.SameSiteLaxMode},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			if got := parseSameSite(tt.value); got != tt.want {
				t.Errorf("parseSameSite(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

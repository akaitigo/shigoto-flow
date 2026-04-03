package sender

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSlackSender_Send(t *testing.T) {
	var receivedPayload map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&receivedPayload); err != nil {
			t.Fatalf("failed to decode payload: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewSlack(server.URL, server.Client())
	err := sender.Send(context.Background(), "#general", "日報", "テスト内容")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedPayload["channel"] != "#general" {
		t.Errorf("expected channel #general, got %s", receivedPayload["channel"])
	}

	if receivedPayload["text"] == "" {
		t.Error("expected non-empty text")
	}
}

func TestSlackSender_Send_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("error")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	sender := NewSlack(server.URL, server.Client())
	err := sender.Send(context.Background(), "#general", "日報", "テスト")
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestService_Send(t *testing.T) {
	called := false
	mock := &mockSender{
		typeName: "test",
		sendFn: func(_ context.Context, _, _, _ string) error {
			called = true
			return nil
		},
	}

	svc := NewService(mock)
	err := svc.Send(context.Background(), "test", "to", "subject", "body")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected sender to be called")
	}
}

func TestService_Send_UnknownType(t *testing.T) {
	svc := NewService()
	err := svc.Send(context.Background(), "unknown", "to", "subject", "body")
	if err == nil {
		t.Error("expected error for unknown sender type")
	}
}

type mockSender struct {
	typeName string
	sendFn   func(ctx context.Context, to, subject, body string) error
}

func (m *mockSender) Type() string { return m.typeName }
func (m *mockSender) Send(ctx context.Context, to, subject, body string) error {
	return m.sendFn(ctx, to, subject, body)
}

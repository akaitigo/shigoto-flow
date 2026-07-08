package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/akaitigo/shigoto-flow/backend/internal/middleware"
	"github.com/akaitigo/shigoto-flow/backend/internal/summary"
)

type fakeSummaryService struct {
	weekly    string
	monthly   string
	weeklyErr error
	monthErr  error
	called    string
	gotUserID string
	gotDate   time.Time
}

func (f *fakeSummaryService) GenerateWeeklySummary(_ context.Context, userID string, weekStart time.Time) (string, error) {
	f.called = "weekly"
	f.gotUserID = userID
	f.gotDate = weekStart
	return f.weekly, f.weeklyErr
}

func (f *fakeSummaryService) GenerateMonthlySummary(_ context.Context, userID string, month time.Time) (string, error) {
	f.called = "monthly"
	f.gotUserID = userID
	f.gotDate = month
	return f.monthly, f.monthErr
}

func doSummarize(h *Handler, body string, authed bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/reports/summarize", strings.NewReader(body))
	if authed {
		req = req.WithContext(middleware.WithUserID(req.Context(), "user-1"))
	}
	rec := httptest.NewRecorder()
	h.SummarizeReport(rec, req)
	return rec
}

func TestSummarizeReport_Weekly(t *testing.T) {
	fake := &fakeSummaryService{weekly: "今週のまとめ"}
	h := &Handler{summarySvc: fake}

	rec := doSummarize(h, `{"type":"weekly","date":"2026-04-06"}`, true)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	var resp summarizeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Content != "今週のまとめ" {
		t.Errorf("expected weekly content, got %q", resp.Content)
	}
	if fake.called != "weekly" {
		t.Errorf("expected weekly generator called, got %q", fake.called)
	}
	if fake.gotUserID != "user-1" {
		t.Errorf("expected user-1, got %q", fake.gotUserID)
	}
	if want := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC); !fake.gotDate.Equal(want) {
		t.Errorf("expected date %v, got %v", want, fake.gotDate)
	}
}

func TestSummarizeReport_Monthly(t *testing.T) {
	fake := &fakeSummaryService{monthly: "今月のまとめ"}
	h := &Handler{summarySvc: fake}

	rec := doSummarize(h, `{"type":"monthly","date":"2026-04-01"}`, true)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if fake.called != "monthly" {
		t.Errorf("expected monthly generator called, got %q", fake.called)
	}
}

func TestSummarizeReport_Errors(t *testing.T) {
	tests := []struct {
		name     string
		svc      summaryService
		body     string
		authed   bool
		wantCode int
	}{
		{
			name:     "unauthenticated",
			svc:      &fakeSummaryService{},
			body:     `{"type":"weekly","date":"2026-04-06"}`,
			authed:   false,
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "invalid type daily",
			svc:      &fakeSummaryService{},
			body:     `{"type":"daily","date":"2026-04-06"}`,
			authed:   true,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "invalid date",
			svc:      &fakeSummaryService{},
			body:     `{"type":"weekly","date":"nope"}`,
			authed:   true,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "malformed json",
			svc:      &fakeSummaryService{},
			body:     `{`,
			authed:   true,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "no reports",
			svc:      &fakeSummaryService{weeklyErr: summary.ErrNoReports},
			body:     `{"type":"weekly","date":"2026-04-06"}`,
			authed:   true,
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name:     "generation failure",
			svc:      &fakeSummaryService{weeklyErr: errors.New("claude down")},
			body:     `{"type":"weekly","date":"2026-04-06"}`,
			authed:   true,
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{summarySvc: tt.svc}
			rec := doSummarize(h, tt.body, tt.authed)
			if rec.Code != tt.wantCode {
				t.Errorf("expected %d, got %d (%s)", tt.wantCode, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestSummarizeReport_ServiceNotConfigured(t *testing.T) {
	h := &Handler{} // summarySvc is nil
	rec := doSummarize(h, `{"type":"weekly","date":"2026-04-06"}`, true)
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 when summary service is unset, got %d", rec.Code)
	}
}

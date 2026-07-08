package collector

import (
	"strconv"
	"testing"
	"time"
)

func TestParseInternalDate(t *testing.T) {
	want := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		internalDate string
		wantFallback bool
		want         time.Time
	}{
		{
			name:         "valid unix millis",
			internalDate: strconv.FormatInt(want.UnixMilli(), 10),
			want:         want,
		},
		{
			name:         "empty falls back to now",
			internalDate: "",
			wantFallback: true,
		},
		{
			name:         "non-numeric falls back to now",
			internalDate: "not-a-number",
			wantFallback: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseInternalDate(tt.internalDate)

			if tt.wantFallback {
				// Fallback should be close to now, never the zero value.
				if got.IsZero() {
					t.Error("expected fallback to current time, got zero value")
				}
				if time.Since(got) > time.Minute {
					t.Errorf("expected fallback near now, got %v", got)
				}
				return
			}

			if !got.UTC().Equal(tt.want) {
				t.Errorf("expected %v, got %v", tt.want, got.UTC())
			}
		})
	}
}

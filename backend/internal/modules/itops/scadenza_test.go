package itops

import "testing"

func TestComputeSeverity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		daysLeft int
		want     ScadenzaSeverity
	}{
		{"expired", -5, SeverityBad},
		{"today", 0, SeverityBad},
		{"within_30", 15, SeverityBad},
		{"exactly_30", 30, SeverityBad},
		{"31_days", 31, SeverityWarn},
		{"within_90", 60, SeverityWarn},
		{"exactly_90", 90, SeverityWarn},
		{"91_days", 91, SeverityOK},
		{"far_future", 365, SeverityOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ComputeSeverity(tt.daysLeft)
			if got != tt.want {
				t.Errorf("ComputeSeverity(%d) = %q, want %q", tt.daysLeft, got, tt.want)
			}
		})
	}
}

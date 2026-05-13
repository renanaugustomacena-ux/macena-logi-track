package itops

import "testing"

func validIncident() *Incident {
	return &Incident{
		Title:    "Stampante piano 2 offline",
		Category: CatPrinter,
		Priority: PriorityMedium,
	}
}

func TestIncidentValidate(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		if err := validIncident().Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing title rejected", func(t *testing.T) {
		t.Parallel()
		inc := validIncident()
		inc.Title = ""
		if err := inc.Validate(); err == nil {
			t.Fatal("expected error for empty title")
		}
	})

	t.Run("invalid category rejected", func(t *testing.T) {
		t.Parallel()
		inc := validIncident()
		inc.Category = "magic"
		if err := inc.Validate(); err == nil {
			t.Fatal("expected error for invalid category")
		}
	})

	t.Run("invalid priority rejected", func(t *testing.T) {
		t.Parallel()
		inc := validIncident()
		inc.Priority = "apocalyptic"
		if err := inc.Validate(); err == nil {
			t.Fatal("expected error for invalid priority")
		}
	})
}

func TestSLAHours(t *testing.T) {
	t.Parallel()

	cases := []struct {
		priority IncidentPriority
		want     int
	}{
		{PriorityCritical, 4},
		{PriorityHigh, 8},
		{PriorityMedium, 24},
		{PriorityLow, 72},
		{"unknown", 72},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(string(tc.priority), func(t *testing.T) {
			t.Parallel()
			if got := tc.priority.SLAHours(); got != tc.want {
				t.Fatalf("SLAHours(%q) = %d, want %d", tc.priority, got, tc.want)
			}
		})
	}
}

func TestCanTransition(t *testing.T) {
	t.Parallel()

	allowed := []struct {
		from, to IncidentStatus
	}{
		{IncidentOpen, IncidentAssigned},
		{IncidentOpen, IncidentClosed},
		{IncidentAssigned, IncidentInProgress},
		{IncidentAssigned, IncidentClosed},
		{IncidentInProgress, IncidentPending},
		{IncidentInProgress, IncidentResolved},
		{IncidentInProgress, IncidentClosed},
		{IncidentPending, IncidentInProgress},
		{IncidentPending, IncidentClosed},
		{IncidentResolved, IncidentInProgress},
		{IncidentResolved, IncidentClosed},
	}
	for _, tc := range allowed {
		tc := tc
		t.Run(string(tc.from)+"_to_"+string(tc.to), func(t *testing.T) {
			t.Parallel()
			if !CanTransition(tc.from, tc.to) {
				t.Fatalf("expected %s -> %s to be allowed", tc.from, tc.to)
			}
		})
	}

	blocked := []struct {
		from, to IncidentStatus
	}{
		{IncidentOpen, IncidentInProgress},
		{IncidentOpen, IncidentResolved},
		{IncidentClosed, IncidentOpen},
		{IncidentClosed, IncidentAssigned},
		{IncidentResolved, IncidentOpen},
		{IncidentAssigned, IncidentPending},
	}
	for _, tc := range blocked {
		tc := tc
		t.Run("blocked_"+string(tc.from)+"_to_"+string(tc.to), func(t *testing.T) {
			t.Parallel()
			if CanTransition(tc.from, tc.to) {
				t.Fatalf("expected %s -> %s to be blocked", tc.from, tc.to)
			}
		})
	}
}

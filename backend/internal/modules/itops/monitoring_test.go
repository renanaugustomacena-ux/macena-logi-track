package itops

import "testing"

func TestValidateCheckStatus(t *testing.T) {
	t.Parallel()

	valid := []CheckStatus{CheckOK, CheckWarning, CheckCritical, CheckUnknown}
	for _, s := range valid {
		s := s
		t.Run("valid_"+string(s), func(t *testing.T) {
			t.Parallel()
			if err := ValidateCheckStatus(s); err != nil {
				t.Fatalf("expected no error for %q, got %v", s, err)
			}
		})
	}

	t.Run("unknown status", func(t *testing.T) {
		t.Parallel()
		if err := ValidateCheckStatus("exploding"); err == nil {
			t.Fatal("expected error for unknown check status")
		}
	})
}

func TestMonitoringCheckValidate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		check   MonitoringCheck
		wantErr bool
	}{
		{
			name:    "valid",
			check:   MonitoringCheck{TenantID: "t1", AssetID: "a1", CheckName: "ping"},
			wantErr: false,
		},
		{
			name:    "missing tenant",
			check:   MonitoringCheck{AssetID: "a1", CheckName: "ping"},
			wantErr: true,
		},
		{
			name:    "missing asset",
			check:   MonitoringCheck{TenantID: "t1", CheckName: "ping"},
			wantErr: true,
		},
		{
			name:    "missing check_name",
			check:   MonitoringCheck{TenantID: "t1", AssetID: "a1"},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.check.Validate()
			if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestMonitoringAlertValidate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		alert   MonitoringAlert
		wantErr bool
	}{
		{
			name:    "valid",
			alert:   MonitoringAlert{TenantID: "t1", Title: "CPU high"},
			wantErr: false,
		},
		{
			name:    "missing tenant",
			alert:   MonitoringAlert{Title: "CPU high"},
			wantErr: true,
		},
		{
			name:    "missing title",
			alert:   MonitoringAlert{TenantID: "t1"},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.alert.Validate()
			if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

package itops

import "testing"

func TestLicenseValidate(t *testing.T) {
	t.Parallel()
	valid := License{
		Software:    "Microsoft 365",
		Vendor:      "Microsoft",
		LicenseType: LicenseSubscription,
		Seats:       50,
		SeatsUsed:   42,
	}
	tests := []struct {
		name    string
		mod     func(*License)
		wantErr bool
	}{
		{"valid", func(_ *License) {}, false},
		{"missing_software", func(l *License) { l.Software = "" }, true},
		{"missing_vendor", func(l *License) { l.Vendor = "" }, true},
		{"bad_license_type", func(l *License) { l.LicenseType = "pirated" }, true},
		{"negative_seats", func(l *License) { l.Seats = -1 }, true},
		{"negative_seats_used", func(l *License) { l.SeatsUsed = -1 }, true},
		{"zero_seats_ok", func(l *License) { l.Seats = 0 }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			lic := valid
			tt.mod(&lic)
			err := lic.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLicenseIsOverDeployed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		seats int
		used  int
		want  bool
	}{
		{"under", 50, 30, false},
		{"exact", 50, 50, false},
		{"over", 50, 51, true},
		{"unlimited_zero_seats", 0, 100, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := License{Seats: tt.seats, SeatsUsed: tt.used}
			if got := l.IsOverDeployed(); got != tt.want {
				t.Errorf("IsOverDeployed() = %v, want %v", got, tt.want)
			}
		})
	}
}

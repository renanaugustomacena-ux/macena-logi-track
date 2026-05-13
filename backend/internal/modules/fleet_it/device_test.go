package fleet_it

import "testing"

func validDevice() *DriverDevice {
	return &DriverDevice{
		SerialNumber: "DEV-001",
		DeviceType:   "tablet",
		Status:       DeviceActive,
	}
}

func TestDriverDeviceValidate(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		if err := validDevice().Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing serial rejected", func(t *testing.T) {
		t.Parallel()
		d := validDevice()
		d.SerialNumber = ""
		if err := d.Validate(); err == nil {
			t.Fatal("expected error for empty serial number")
		}
	})

	t.Run("missing device type rejected", func(t *testing.T) {
		t.Parallel()
		d := validDevice()
		d.DeviceType = ""
		if err := d.Validate(); err == nil {
			t.Fatal("expected error for empty device type")
		}
	})

	t.Run("invalid status rejected", func(t *testing.T) {
		t.Parallel()
		d := validDevice()
		d.Status = "stolen"
		if err := d.Validate(); err == nil {
			t.Fatal("expected error for invalid status")
		}
	})

	t.Run("all valid statuses accepted", func(t *testing.T) {
		t.Parallel()
		for _, s := range []DeviceStatus{DeviceActive, DeviceLost, DeviceDamaged, DeviceSpare} {
			d := validDevice()
			d.Status = s
			if err := d.Validate(); err != nil {
				t.Fatalf("expected no error for status %q, got %v", s, err)
			}
		}
	})
}

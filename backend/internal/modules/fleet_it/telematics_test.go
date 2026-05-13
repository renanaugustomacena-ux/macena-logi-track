package fleet_it

import "testing"

func validUnit() *TelematicsUnit {
	return &TelematicsUnit{
		VehicleID:    "v-1",
		Provider:     "ecofleet",
		DeviceSerial: "ECO-9001",
	}
}

func TestTelematicsUnitValidate(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		if err := validUnit().Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing vehicle ID rejected", func(t *testing.T) {
		t.Parallel()
		u := validUnit()
		u.VehicleID = ""
		if err := u.Validate(); err == nil {
			t.Fatal("expected error for empty vehicle ID")
		}
	})

	t.Run("missing provider rejected", func(t *testing.T) {
		t.Parallel()
		u := validUnit()
		u.Provider = ""
		if err := u.Validate(); err == nil {
			t.Fatal("expected error for empty provider")
		}
	})

	t.Run("missing serial rejected", func(t *testing.T) {
		t.Parallel()
		u := validUnit()
		u.DeviceSerial = ""
		if err := u.Validate(); err == nil {
			t.Fatal("expected error for empty device serial")
		}
	})
}

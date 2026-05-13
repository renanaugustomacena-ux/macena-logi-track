package itops

import "testing"

func validAsset() *ITAsset {
	return &ITAsset{
		Name:         "DC01",
		Kind:         AssetServer,
		Status:       StatusActive,
		SerialNumber: "SN-001",
	}
}

func TestITAssetValidate(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		if err := validAsset().Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing name rejected", func(t *testing.T) {
		t.Parallel()
		a := validAsset()
		a.Name = ""
		if err := a.Validate(); err == nil {
			t.Fatal("expected error for empty name")
		}
	})

	t.Run("missing serial rejected", func(t *testing.T) {
		t.Parallel()
		a := validAsset()
		a.SerialNumber = ""
		if err := a.Validate(); err == nil {
			t.Fatal("expected error for empty serial number")
		}
	})

	t.Run("invalid kind rejected", func(t *testing.T) {
		t.Parallel()
		a := validAsset()
		a.Kind = "toaster"
		if err := a.Validate(); err == nil {
			t.Fatal("expected error for invalid kind")
		}
	})

	t.Run("invalid status rejected", func(t *testing.T) {
		t.Parallel()
		a := validAsset()
		a.Status = "exploded"
		if err := a.Validate(); err == nil {
			t.Fatal("expected error for invalid status")
		}
	})
}

func TestValidateAssetKind(t *testing.T) {
	t.Parallel()

	valid := []AssetKind{
		AssetServer, AssetVM, AssetWorkstation, AssetLaptop,
		AssetSwitch, AssetRouter, AssetFirewall, AssetAP,
		AssetUPS, AssetPrinter, AssetNAS, AssetPhone,
		AssetTablet, AssetOther,
	}
	for _, k := range valid {
		k := k
		t.Run("valid_"+string(k), func(t *testing.T) {
			t.Parallel()
			if err := ValidateAssetKind(k); err != nil {
				t.Fatalf("expected no error for %q, got %v", k, err)
			}
		})
	}

	t.Run("unknown kind", func(t *testing.T) {
		t.Parallel()
		if err := ValidateAssetKind("fridge"); err == nil {
			t.Fatal("expected error for unknown kind")
		}
	})
}

func TestValidateAssetStatus(t *testing.T) {
	t.Parallel()

	valid := []AssetStatus{
		StatusActive, StatusMaintenance, StatusDecommissioned,
		StatusSpare, StatusOrdered,
	}
	for _, s := range valid {
		s := s
		t.Run("valid_"+string(s), func(t *testing.T) {
			t.Parallel()
			if err := ValidateAssetStatus(s); err != nil {
				t.Fatalf("expected no error for %q, got %v", s, err)
			}
		})
	}

	t.Run("unknown status", func(t *testing.T) {
		t.Parallel()
		if err := ValidateAssetStatus("on_fire"); err == nil {
			t.Fatal("expected error for unknown status")
		}
	})
}

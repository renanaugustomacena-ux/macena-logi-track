package itops

import (
	"fmt"
	"time"
)

// LicenseType classifies the licensing model.
type LicenseType string

const (
	LicensePerpetual    LicenseType = "perpetual"
	LicenseSubscription LicenseType = "subscription"
	LicenseOEM          LicenseType = "oem"
	LicenseVolume       LicenseType = "volume"
	LicenseOpenSource   LicenseType = "open_source"
)

var validLicenseTypes = map[LicenseType]bool{
	LicensePerpetual: true, LicenseSubscription: true,
	LicenseOEM: true, LicenseVolume: true, LicenseOpenSource: true,
}

// RenewalType indicates how the license renews.
type RenewalType string

const (
	RenewalAuto   RenewalType = "auto"
	RenewalManual RenewalType = "manual"
	RenewalNA     RenewalType = "n/a"
)

// License tracks a software license and its seat utilization.
type License struct {
	ID           string      `bson:"_id,omitempty" json:"id"`
	TenantID     string      `bson:"tenant_id" json:"tenantId"`
	Software     string      `bson:"software" json:"software"`
	Vendor       string      `bson:"vendor" json:"vendor"`
	LicenseType  LicenseType `bson:"license_type" json:"licenseType"`
	LicenseKey   string      `bson:"license_key,omitempty" json:"licenseKey,omitempty"`
	Seats        int         `bson:"seats" json:"seats"`
	SeatsUsed    int         `bson:"seats_used" json:"seatsUsed"`
	PurchaseDate *time.Time  `bson:"purchase_date,omitempty" json:"purchaseDate,omitempty"`
	ExpiryDate   *time.Time  `bson:"expiry_date,omitempty" json:"expiryDate,omitempty"`
	CostCents    int         `bson:"cost_cents,omitempty" json:"costCents,omitempty"`
	Currency     string      `bson:"currency,omitempty" json:"currency,omitempty"`
	Renewal      RenewalType `bson:"renewal" json:"renewal"`
	LinkedAssets []string    `bson:"linked_assets,omitempty" json:"linkedAssets,omitempty"`
	Notes        string      `bson:"notes,omitempty" json:"notes,omitempty"`
	CreatedAt    time.Time   `bson:"created_at" json:"createdAt"`
	UpdatedAt    time.Time   `bson:"updated_at" json:"updatedAt"`
}

// Validate checks required fields.
func (l *License) Validate() error {
	if l.Software == "" {
		return fmt.Errorf("itops: software name is required")
	}
	if l.Vendor == "" {
		return fmt.Errorf("itops: vendor is required")
	}
	if !validLicenseTypes[l.LicenseType] {
		return fmt.Errorf("itops: unknown license type %q", l.LicenseType)
	}
	if l.Seats < 0 {
		return fmt.Errorf("itops: seats cannot be negative")
	}
	if l.SeatsUsed < 0 {
		return fmt.Errorf("itops: seats_used cannot be negative")
	}
	return nil
}

// IsOverDeployed returns true when more seats are used than purchased.
func (l *License) IsOverDeployed() bool {
	return l.Seats > 0 && l.SeatsUsed > l.Seats
}

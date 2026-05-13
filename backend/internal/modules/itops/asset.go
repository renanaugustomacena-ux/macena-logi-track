package itops

import (
	"fmt"
	"time"
)

// AssetKind distinguishes the basic categories of IT assets.
type AssetKind string

const (
	AssetServer      AssetKind = "server"
	AssetVM          AssetKind = "vm"
	AssetWorkstation AssetKind = "workstation"
	AssetLaptop      AssetKind = "laptop"
	AssetSwitch      AssetKind = "switch"
	AssetRouter      AssetKind = "router"
	AssetFirewall    AssetKind = "firewall"
	AssetAP          AssetKind = "access_point"
	AssetUPS         AssetKind = "ups"
	AssetPrinter     AssetKind = "printer"
	AssetNAS         AssetKind = "nas"
	AssetPhone       AssetKind = "phone"
	AssetTablet      AssetKind = "tablet"
	AssetOther       AssetKind = "other"
)

var validAssetKinds = map[AssetKind]bool{
	AssetServer: true, AssetVM: true, AssetWorkstation: true,
	AssetLaptop: true, AssetSwitch: true, AssetRouter: true,
	AssetFirewall: true, AssetAP: true, AssetUPS: true,
	AssetPrinter: true, AssetNAS: true, AssetPhone: true,
	AssetTablet: true, AssetOther: true,
}

// ValidateAssetKind returns an error if k is not a recognized kind.
func ValidateAssetKind(k AssetKind) error {
	if !validAssetKinds[k] {
		return fmt.Errorf("itops: unknown asset kind %q", k)
	}
	return nil
}

// AssetStatus tracks the lifecycle state of an IT asset.
type AssetStatus string

const (
	StatusActive         AssetStatus = "active"
	StatusMaintenance    AssetStatus = "maintenance"
	StatusDecommissioned AssetStatus = "decommissioned"
	StatusSpare          AssetStatus = "spare"
	StatusOrdered        AssetStatus = "ordered"
)

var validAssetStatuses = map[AssetStatus]bool{
	StatusActive: true, StatusMaintenance: true,
	StatusDecommissioned: true, StatusSpare: true,
	StatusOrdered: true,
}

// ValidateAssetStatus returns an error if s is not a recognized status.
func ValidateAssetStatus(s AssetStatus) error {
	if !validAssetStatuses[s] {
		return fmt.Errorf("itops: unknown asset status %q", s)
	}
	return nil
}

// ITAsset represents a managed IT infrastructure component: servers,
// workstations, network devices, peripherals, mobile devices, etc.
type ITAsset struct {
	ID             string      `bson:"_id,omitempty" json:"id"`
	TenantID       string      `bson:"tenant_id" json:"tenantId"`
	Name           string      `bson:"name" json:"name"`
	Kind           AssetKind   `bson:"kind" json:"kind"`
	Status         AssetStatus `bson:"status" json:"status"`
	Location       string      `bson:"location" json:"location"`
	Rack           string      `bson:"rack,omitempty" json:"rack,omitempty"`
	RackUnit       int         `bson:"rack_unit,omitempty" json:"rackUnit,omitempty"`
	SerialNumber   string      `bson:"serial_number" json:"serialNumber"`
	AssetTag       string      `bson:"asset_tag,omitempty" json:"assetTag,omitempty"`
	Manufacturer   string      `bson:"manufacturer" json:"manufacturer"`
	Model          string      `bson:"model" json:"model"`
	PurchaseDate   *time.Time  `bson:"purchase_date,omitempty" json:"purchaseDate,omitempty"`
	WarrantyExpiry *time.Time  `bson:"warranty_expiry,omitempty" json:"warrantyExpiry,omitempty"`
	IPAddress      string      `bson:"ip_address,omitempty" json:"ipAddress,omitempty"`
	MACAddress     string      `bson:"mac_address,omitempty" json:"macAddress,omitempty"`
	OS             string      `bson:"os,omitempty" json:"os,omitempty"`
	OSVersion      string      `bson:"os_version,omitempty" json:"osVersion,omitempty"`
	CPU            string      `bson:"cpu,omitempty" json:"cpu,omitempty"`
	RAMGb          int         `bson:"ram_gb,omitempty" json:"ramGb,omitempty"`
	StorageGb      int         `bson:"storage_gb,omitempty" json:"storageGb,omitempty"`
	AssignedTo     string      `bson:"assigned_to,omitempty" json:"assignedTo,omitempty"`
	Department     string      `bson:"department,omitempty" json:"department,omitempty"`
	LastSeen       *time.Time  `bson:"last_seen,omitempty" json:"lastSeen,omitempty"`
	Notes          string      `bson:"notes,omitempty" json:"notes,omitempty"`
	Tags           []string    `bson:"tags,omitempty" json:"tags,omitempty"`
	ParentAssetID  string      `bson:"parent_asset_id,omitempty" json:"parentAssetId,omitempty"`
	CreatedAt      time.Time   `bson:"created_at" json:"createdAt"`
	UpdatedAt      time.Time   `bson:"updated_at" json:"updatedAt"`
}

// Validate checks required fields and enum values.
func (a *ITAsset) Validate() error {
	if a.Name == "" {
		return fmt.Errorf("itops: asset name is required")
	}
	if err := ValidateAssetKind(a.Kind); err != nil {
		return err
	}
	if err := ValidateAssetStatus(a.Status); err != nil {
		return err
	}
	if a.SerialNumber == "" {
		return fmt.Errorf("itops: serial number is required")
	}
	return nil
}

package fleet_it

import (
	"fmt"
	"time"
)

// DeviceStatus tracks the lifecycle of a driver-assigned device.
type DeviceStatus string

const (
	DeviceActive  DeviceStatus = "active"
	DeviceLost    DeviceStatus = "lost"
	DeviceDamaged DeviceStatus = "damaged"
	DeviceSpare   DeviceStatus = "spare"
)

var validDeviceStatuses = map[DeviceStatus]bool{
	DeviceActive: true, DeviceLost: true,
	DeviceDamaged: true, DeviceSpare: true,
}

// DriverDevice represents a tablet, phone, or other mobile device
// assigned to a driver for fleet operations (trip management, CMR,
// communication).
type DriverDevice struct {
	ID           string       `bson:"_id,omitempty" json:"id"`
	TenantID     string       `bson:"tenant_id" json:"tenantId"`
	DriverID     string       `bson:"driver_id,omitempty" json:"driverId,omitempty"`
	DeviceType   string       `bson:"device_type" json:"deviceType"`
	Manufacturer string       `bson:"manufacturer" json:"manufacturer"`
	Model        string       `bson:"model" json:"model"`
	SerialNumber string       `bson:"serial_number" json:"serialNumber"`
	IMEI         string       `bson:"imei,omitempty" json:"imei,omitempty"`
	PhoneNumber  string       `bson:"phone_number,omitempty" json:"phoneNumber,omitempty"`
	SIMProvider  string       `bson:"sim_provider,omitempty" json:"simProvider,omitempty"`
	ICCID        string       `bson:"iccid,omitempty" json:"iccid,omitempty"`
	OSVersion    string       `bson:"os_version,omitempty" json:"osVersion,omitempty"`
	MDMEnrolled  bool         `bson:"mdm_enrolled" json:"mdmEnrolled"`
	MDMCompliant bool         `bson:"mdm_compliant" json:"mdmCompliant"`
	AssignedDate *time.Time   `bson:"assigned_date,omitempty" json:"assignedDate,omitempty"`
	LastCheckIn  *time.Time   `bson:"last_check_in,omitempty" json:"lastCheckIn,omitempty"`
	BatteryLevel int          `bson:"battery_level,omitempty" json:"batteryLevel,omitempty"`
	Status       DeviceStatus `bson:"status" json:"status"`
	CreatedAt    time.Time    `bson:"created_at" json:"createdAt"`
	UpdatedAt    time.Time    `bson:"updated_at" json:"updatedAt"`
}

// Validate checks required fields.
func (d *DriverDevice) Validate() error {
	if d.SerialNumber == "" {
		return fmt.Errorf("fleet_it: device serial number is required")
	}
	if d.DeviceType == "" {
		return fmt.Errorf("fleet_it: device type is required")
	}
	if !validDeviceStatuses[d.Status] {
		return fmt.Errorf("fleet_it: unknown device status %q", d.Status)
	}
	return nil
}

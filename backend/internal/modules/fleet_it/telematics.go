package fleet_it

import (
	"fmt"
	"time"
)

// TelematicsStatus represents the communication state of a unit.
type TelematicsStatus string

const (
	TelematicsOnline  TelematicsStatus = "online"
	TelematicsOffline TelematicsStatus = "offline"
	TelematicsFault   TelematicsStatus = "fault"
)

// TelematicsUnit links a vehicle to its telematics hardware
// (Ecofleet aftermarket unit, Iveco Connectivity Box, etc.).
type TelematicsUnit struct {
	ID              string           `bson:"_id,omitempty" json:"id"`
	TenantID        string           `bson:"tenant_id" json:"tenantId"`
	VehicleID       string           `bson:"vehicle_id" json:"vehicleId"`
	Provider        string           `bson:"provider" json:"provider"`
	DeviceSerial    string           `bson:"device_serial" json:"deviceSerial"`
	SIMICCID        string           `bson:"sim_iccid,omitempty" json:"simIccid,omitempty"`
	FirmwareVersion string           `bson:"firmware_version,omitempty" json:"firmwareVersion,omitempty"`
	InstallDate     *time.Time       `bson:"install_date,omitempty" json:"installDate,omitempty"`
	LastComm        *time.Time       `bson:"last_comm,omitempty" json:"lastComm,omitempty"`
	CommIntervalSec int              `bson:"comm_interval_sec" json:"commIntervalSec"`
	Status          TelematicsStatus `bson:"status" json:"status"`
	Features        []string         `bson:"features,omitempty" json:"features,omitempty"`
	CreatedAt       time.Time        `bson:"created_at" json:"createdAt"`
	UpdatedAt       time.Time        `bson:"updated_at" json:"updatedAt"`
}

// Validate checks required fields.
func (t *TelematicsUnit) Validate() error {
	if t.VehicleID == "" {
		return fmt.Errorf("fleet_it: vehicle ID is required")
	}
	if t.Provider == "" {
		return fmt.Errorf("fleet_it: telematics provider is required")
	}
	if t.DeviceSerial == "" {
		return fmt.Errorf("fleet_it: device serial is required")
	}
	return nil
}

package models

import "time"

// VehicleKind distinguishes the basic categories of fleet assets.
type VehicleKind string

const (
	VehicleTruck        VehicleKind = "truck"
	VehicleSemiTrailer  VehicleKind = "semi_trailer"
	VehicleVan          VehicleKind = "van"
	VehicleTrain        VehicleKind = "train_wagon"
	VehicleContainer    VehicleKind = "container"
	VehicleReefer       VehicleKind = "reefer"
)

// Vehicle is a fleet asset registered in the Italian motorisation
// register (MCTC). The Plate field matches the Italian format
// (e.g. "EB123XY") and is unique within a carrier.
type Vehicle struct {
	ID            string      `bson:"_id,omitempty" json:"id"`
	TenantID      string      `bson:"tenant_id" json:"tenantId"`
	Plate         string      `bson:"plate" json:"plate"`
	Carrier       string      `bson:"carrier" json:"carrier"`
	Kind          VehicleKind `bson:"kind" json:"kind"`
	CapacityKG    float64     `bson:"capacity_kg" json:"capacityKg"`
	CapacityM3    float64     `bson:"capacity_m3" json:"capacityM3"`
	Axles         int         `bson:"axles" json:"axles"`
	EuroClass     string      `bson:"euro_class,omitempty" json:"euroClass,omitempty"`
	TelematicsID  string      `bson:"telematics_id,omitempty" json:"telematicsId,omitempty"`
	ProviderName  string      `bson:"provider_name,omitempty" json:"providerName,omitempty"`
	TelepassID    string      `bson:"telepass_id,omitempty" json:"telepassId,omitempty"`
	ADRCertified  bool        `bson:"adr_certified" json:"adrCertified"`
	ATPCertified  bool        `bson:"atp_certified" json:"atpCertified"`
	Active        bool        `bson:"active" json:"active"`
	CreatedAt     time.Time   `bson:"created_at" json:"createdAt"`
	UpdatedAt     time.Time   `bson:"updated_at" json:"updatedAt"`
}

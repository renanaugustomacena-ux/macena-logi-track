package models

import "time"

// LicenceCategory enumerates the Italian driving licence categories
// relevant to professional road freight. CQC (Carta di Qualificazione
// del Conducente) certification is required on top of the licence for
// commercial transport, tracked separately.
type LicenceCategory string

const (
	LicenceB  LicenceCategory = "B"
	LicenceC  LicenceCategory = "C"
	LicenceCE LicenceCategory = "CE"
	LicenceD  LicenceCategory = "D"
	LicenceDE LicenceCategory = "DE"
)

// Driver is a professional road-freight driver associated with a
// carrier. The Codice Fiscale (fiscal code) is not stored by default:
// LogiTrack treats it as "sensitive by default" per GDPR minimisation
// guidance and only ingests it for operators requiring CCNL payroll
// integration (opt-in flag on the tenant settings).
type Driver struct {
	ID             string            `bson:"_id,omitempty" json:"id"`
	TenantID       string            `bson:"tenant_id" json:"tenantId"`
	Name           string            `bson:"name" json:"name"`
	Carrier        string            `bson:"carrier" json:"carrier"`
	Licence        LicenceCategory   `bson:"licence" json:"licence"`
	LicenceExpiry  time.Time         `bson:"licence_expiry" json:"licenceExpiry"`
	CQCExpiry      time.Time         `bson:"cqc_expiry,omitempty" json:"cqcExpiry,omitempty"`
	ADRQualified   bool              `bson:"adr_qualified" json:"adrQualified"`
	AlboRegistered bool              `bson:"albo_registered" json:"alboRegistered"`
	CurrentVehicle string            `bson:"current_vehicle,omitempty" json:"currentVehicle,omitempty"`
	ContactEmail   string            `bson:"contact_email,omitempty" json:"contactEmail,omitempty"`
	ContactPhone   string            `bson:"contact_phone,omitempty" json:"contactPhone,omitempty"`
	DrivingHours   DrivingHoursToday `bson:"driving_hours" json:"drivingHours"`
	Active         bool              `bson:"active" json:"active"`
	CreatedAt      time.Time         `bson:"created_at" json:"createdAt"`
	UpdatedAt      time.Time         `bson:"updated_at" json:"updatedAt"`
}

// DrivingHoursToday mirrors the daily counters required by
// Regulation (EC) 561/2006 on driving times. The data originates from
// the digital tachograph and is cached here for rapid compliance
// checks (e.g. dispatch a new load only if 9h cap has not been
// reached).
type DrivingHoursToday struct {
	DrivenSecs   int64     `bson:"driven_secs" json:"drivenSecs"`
	RestedSecs   int64     `bson:"rested_secs" json:"restedSecs"`
	PeriodStart  time.Time `bson:"period_start" json:"periodStart"`
	LastSyncedAt time.Time `bson:"last_synced_at" json:"lastSyncedAt"`
}

package models

import "time"

// GeofenceType enumerates the business categories of geofenced
// polygons relevant to the Verona logistics corridor.
type GeofenceType string

const (
	GeofenceWarehouse   GeofenceType = "warehouse"
	GeofenceCustoms     GeofenceType = "customs"
	GeofencePort        GeofenceType = "port"
	GeofenceDepot       GeofenceType = "depot"
	GeofenceTerminal    GeofenceType = "terminal"
	GeofenceRestricted  GeofenceType = "restricted"
	GeofenceLoadingBay  GeofenceType = "loading_bay"
	GeofenceRestArea    GeofenceType = "rest_area"
)

// GeoPolygon is a GeoJSON Polygon. The first linear ring is the outer
// boundary; subsequent rings describe holes (rare in logistics use).
type GeoPolygon struct {
	Type        string        `bson:"type" json:"type"`
	Coordinates [][][]float64 `bson:"coordinates" json:"coordinates"`
}

// Geofence is a named spatial area tracked by the platform. Entering
// and exiting a geofence emits a TrackingEvent.
type Geofence struct {
	ID          string       `bson:"_id,omitempty" json:"id"`
	TenantID    string       `bson:"tenant_id" json:"tenantId"`
	Name        string       `bson:"name" json:"name"`
	Type        GeofenceType `bson:"type" json:"type"`
	Polygon     GeoPolygon   `bson:"polygon" json:"polygon"`
	AreaRef     string       `bson:"area_ref,omitempty" json:"areaRef,omitempty"`
	DwellAlertMinutes int    `bson:"dwell_alert_minutes,omitempty" json:"dwellAlertMinutes,omitempty"`
	BusinessHours string     `bson:"business_hours,omitempty" json:"businessHours,omitempty"`
	Active      bool         `bson:"active" json:"active"`
	CreatedAt   time.Time    `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time    `bson:"updated_at" json:"updatedAt"`
}

// Package models defines the persistent domain entities used by the
// LogiTrack platform. The types are consumed by the MongoDB repository
// layer, the Gin HTTP handlers, and the WebSocket fan-out hub.
//
// BSON tags follow the LogiTrack naming convention: lowercase_snake on
// disk, camelCase on the JSON wire. This lets the carrier-integration
// team ingest webhooks with minimal translation while keeping database
// documents compact and grep-friendly.
package models

import (
	"strings"
	"time"
)

// ShipmentMode captures the logistics mode of the consignment.
// Values align with the UN/CEFACT mode codes used in customs
// declarations (AIDA) and in the Consorzio ZAI intermodal booking
// system at Quadrante Europa.
type ShipmentMode string

const (
	ModeRoad       ShipmentMode = "road"
	ModeRail       ShipmentMode = "rail"
	ModeMultimodal ShipmentMode = "multimodal"
	ModeSea        ShipmentMode = "sea"
	ModeAir        ShipmentMode = "air"
)

// ShipmentStatus follows a narrow state machine. Transitions are
// validated by the service layer so external webhooks cannot drive the
// entity into a contradictory state.
type ShipmentStatus string

const (
	StatusDraft     ShipmentStatus = "draft"
	StatusBooked    ShipmentStatus = "booked"
	StatusPickedUp  ShipmentStatus = "picked_up"
	StatusInTransit ShipmentStatus = "in_transit"
	StatusDelayed   ShipmentStatus = "delayed"
	StatusAtCustoms ShipmentStatus = "at_customs"
	StatusDelivered ShipmentStatus = "delivered"
	StatusCancelled ShipmentStatus = "cancelled"
)

// GeoPoint is a GeoJSON Point — the canonical representation MongoDB
// expects for $near and $geoWithin queries.
type GeoPoint struct {
	Type        string    `bson:"type" json:"type"`
	Coordinates []float64 `bson:"coordinates" json:"coordinates"`
}

// NewGeoPoint returns a GeoJSON point from longitude/latitude pair.
// Note the order: MongoDB stores (lon, lat) not the (lat, lon) form
// often used in Leaflet. The helper exists to make the ordering
// explicit at every call site.
func NewGeoPoint(lon, lat float64) GeoPoint {
	return GeoPoint{Type: "Point", Coordinates: []float64{lon, lat}}
}

// Party is a generic reference to a consignor, consignee, notify party
// or carrier. The VATNumber (partita IVA) is kept as the stable
// identifier because it is regulatorily required on every Italian
// commercial document.
type Party struct {
	Name         string `bson:"name" json:"name"`
	VATNumber    string `bson:"vat_number" json:"vatNumber"`
	Address      string `bson:"address" json:"address"`
	PostalCode   string `bson:"postal_code" json:"postalCode"`
	City         string `bson:"city" json:"city"`
	Province     string `bson:"province" json:"province"`
	Country      string `bson:"country" json:"country"`
	ContactEmail string `bson:"contact_email,omitempty" json:"contactEmail,omitempty"`
	ContactPhone string `bson:"contact_phone,omitempty" json:"contactPhone,omitempty"`
}

// Waypoint represents a discrete position report ingested from a
// telematics provider (Viasat, Octo, Geotab) or an EDI partner.
type Waypoint struct {
	RecordedAt time.Time `bson:"recorded_at" json:"recordedAt"`
	Position   GeoPoint  `bson:"position" json:"position"`
	SpeedKPH   float64   `bson:"speed_kph" json:"speedKph"`
	HeadingDeg float64   `bson:"heading_deg" json:"headingDeg"`
	Source     string    `bson:"source" json:"source"`
	RawEventID string    `bson:"raw_event_id,omitempty" json:"rawEventId,omitempty"`
}

// CustomsStatus records the Agenzia delle Dogane (AIDA) state of a
// cross-border consignment. Values align with the T1/T2 transit
// document lifecycle.
type CustomsStatus struct {
	Declared       bool      `bson:"declared" json:"declared"`
	DocumentType   string    `bson:"document_type,omitempty" json:"documentType,omitempty"`
	MRN            string    `bson:"mrn,omitempty" json:"mrn,omitempty"`
	ClearedAt      time.Time `bson:"cleared_at,omitempty" json:"clearedAt,omitempty"`
	ClearancePoint string    `bson:"clearance_point,omitempty" json:"clearancePoint,omitempty"`
	HSCode         string    `bson:"hs_code,omitempty" json:"hsCode,omitempty"`
}

// Document attaches EDI, CMR or commercial invoices to a shipment.
// Large payloads are stored in object storage and referenced by URL.
type Document struct {
	ID          string    `bson:"id" json:"id"`
	Kind        string    `bson:"kind" json:"kind"`
	URL         string    `bson:"url" json:"url"`
	ContentHash string    `bson:"content_hash" json:"contentHash"`
	UploadedBy  string    `bson:"uploaded_by" json:"uploadedBy"`
	UploadedAt  time.Time `bson:"uploaded_at" json:"uploadedAt"`
}

// Shipment is the aggregate root for the consignment domain. The
// document is intentionally denormalised: copies of consignor and
// consignee are captured at booking time so later edits to the master
// data do not rewrite history.
type Shipment struct {
	ID              string             `bson:"_id,omitempty" json:"id"`
	TenantID        string             `bson:"tenant_id" json:"tenantId"`
	Reference       string             `bson:"reference" json:"reference"`
	Carrier         string             `bson:"carrier" json:"carrier"`
	Mode            ShipmentMode       `bson:"mode" json:"mode"`
	Status          ShipmentStatus     `bson:"status" json:"status"`
	Consignor       Party              `bson:"consignor" json:"consignor"`
	Consignee       Party              `bson:"consignee" json:"consignee"`
	Origin          GeoPoint           `bson:"origin" json:"origin"`
	Destination     GeoPoint           `bson:"destination" json:"destination"`
	Waypoints       []Waypoint         `bson:"waypoints" json:"waypoints"`
	CurrentPosition *GeoPoint          `bson:"current_position,omitempty" json:"currentPosition,omitempty"`
	ETD             time.Time          `bson:"etd" json:"etd"`
	ETA             time.Time          `bson:"eta" json:"eta"`
	ActualDeparture time.Time          `bson:"actual_departure,omitempty" json:"actualDeparture,omitempty"`
	ActualArrival   time.Time          `bson:"actual_arrival,omitempty" json:"actualArrival,omitempty"`
	Customs         CustomsStatus      `bson:"customs_status" json:"customsStatus"`
	Documents       []Document         `bson:"docs" json:"docs"`
	VehiclePlate    string             `bson:"vehicle_plate,omitempty" json:"vehiclePlate,omitempty"`
	DriverID        string             `bson:"driver_id,omitempty" json:"driverId,omitempty"`
	ADRClass        ADRClass           `bson:"adr_class,omitempty" json:"adrClass,omitempty"`
	ATPClass        ATPClass           `bson:"atp_class,omitempty" json:"atpClass,omitempty"`
	TelepassCodes   []TelepassTollCode `bson:"telepass_codes,omitempty" json:"telepassCodes,omitempty"`
	RoutePolyline   string             `bson:"route_polyline,omitempty" json:"routePolyline,omitempty"`
	CreatedAt       time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updatedAt"`
	Tags            []string           `bson:"tags,omitempty" json:"tags,omitempty"`
}

// Validate performs lightweight invariant checks before persistence.
// Deeper validation (VAT-number checksum, HS-code existence) lives in
// the service layer so this method stays free of external I/O.
func (s *Shipment) Validate() error {
	if s.Reference == "" {
		return ErrMissingReference
	}
	if s.Carrier == "" {
		return ErrMissingCarrier
	}
	if s.Mode == "" {
		s.Mode = ModeRoad
	}
	if s.Status == "" {
		s.Status = StatusDraft
	}
	if len(s.Origin.Coordinates) != 2 || len(s.Destination.Coordinates) != 2 {
		return ErrInvalidGeoPoint
	}
	if s.VehiclePlate != "" {
		if err := ValidatePlate(s.VehiclePlate); err != nil {
			return err
		}
		// Store the plate in the compact (post-1994) form for uniqueness
		// and simpler downstream queries; the historical separator, if
		// any, is kept in the optional `label` of future extensions.
		compact := NormalisePlate(s.VehiclePlate)
		compact = strings.ReplaceAll(compact, " ", "")
		compact = strings.ReplaceAll(compact, "-", "")
		s.VehiclePlate = compact
	}
	return nil
}

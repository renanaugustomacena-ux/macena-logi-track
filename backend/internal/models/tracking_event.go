package models

import (
	"errors"
	"time"
)

// ErrMissingReference is returned when a shipment lacks the
// carrier-facing reference (CMR number, house BL, or internal WO).
var ErrMissingReference = errors.New("shipment reference is required")

// ErrMissingCarrier is returned when the shipment is submitted without
// a nominated carrier. Downstream telematics ingestion keys on the
// carrier identifier so it cannot be blank.
var ErrMissingCarrier = errors.New("shipment carrier is required")

// ErrInvalidGeoPoint indicates a malformed GeoJSON coordinate pair.
var ErrInvalidGeoPoint = errors.New("geo point must be [longitude, latitude]")

// TrackingEventType enumerates the categories of events the platform
// ingests. The list is open-ended by design (carriers publish bespoke
// event catalogues) but the values below are the stable taxonomy used
// for dashboards, alerting and SLA calculations.
type TrackingEventType string

const (
	EventPositionUpdate   TrackingEventType = "position_update"
	EventDeparture        TrackingEventType = "departure"
	EventArrival          TrackingEventType = "arrival"
	EventGeofenceEnter    TrackingEventType = "geofence_enter"
	EventGeofenceExit     TrackingEventType = "geofence_exit"
	EventDelayDetected    TrackingEventType = "delay_detected"
	EventCustomsHold      TrackingEventType = "customs_hold"
	EventCustomsCleared   TrackingEventType = "customs_cleared"
	EventDocumentAttached TrackingEventType = "document_attached"
	EventDriverAssigned   TrackingEventType = "driver_assigned"
	EventVehicleAssigned  TrackingEventType = "vehicle_assigned"
	EventSealBroken       TrackingEventType = "seal_broken"
	EventTemperatureAlarm TrackingEventType = "temperature_alarm"
	EventHandover         TrackingEventType = "handover"
	EventDelivery         TrackingEventType = "delivery"
)

// TrackingEvent is the unit of information that flows through the
// Redis pub/sub channel and, in the future, through Kafka. Events are
// idempotent: consumers must be prepared to receive duplicates and
// deduplicate on (ShipmentID, Sequence).
type TrackingEvent struct {
	ID            string            `bson:"_id,omitempty" json:"id"`
	TenantID      string            `bson:"tenant_id" json:"tenantId"`
	ShipmentID    string            `bson:"shipment_id" json:"shipmentId"`
	Type          TrackingEventType `bson:"type" json:"type"`
	Sequence      int64             `bson:"sequence" json:"sequence"`
	OccurredAt    time.Time         `bson:"occurred_at" json:"occurredAt"`
	RecordedAt    time.Time         `bson:"recorded_at" json:"recordedAt"`
	Position      *GeoPoint         `bson:"position,omitempty" json:"position,omitempty"`
	GeofenceID    string            `bson:"geofence_id,omitempty" json:"geofenceId,omitempty"`
	DelaySecs     int64             `bson:"delay_secs,omitempty" json:"delaySecs,omitempty"`
	Source        string            `bson:"source" json:"source"`
	Actor         string            `bson:"actor,omitempty" json:"actor,omitempty"`
	Metadata      map[string]any    `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CorrelationID string            `bson:"correlation_id,omitempty" json:"correlationId,omitempty"`
}

// IsGeofence reports whether the event represents geofence traversal.
func (e *TrackingEvent) IsGeofence() bool {
	return e.Type == EventGeofenceEnter || e.Type == EventGeofenceExit
}

// IsTerminal reports whether the event completes the shipment
// lifecycle.
func (e *TrackingEvent) IsTerminal() bool {
	return e.Type == EventDelivery || e.Type == EventArrival
}

package logistics

import "time"

// CustodyAction enumerates the recognised custody transitions.
type CustodyAction string

const (
	CustodyCreated    CustodyAction = "created"
	CustodyLoaded     CustodyAction = "loaded"
	CustodySealed     CustodyAction = "sealed"
	CustodyPickedUp   CustodyAction = "picked_up"
	CustodyHandover   CustodyAction = "handover"
	CustodyUnsealed   CustodyAction = "unsealed"
	CustodyUnloaded   CustodyAction = "unloaded"
	CustodyDelivered  CustodyAction = "delivered"
	CustodyInspection CustodyAction = "inspection"
	CustodyException  CustodyAction = "exception"
)

// CustodyRecord is an append-only entry in the chain-of-custody log.
// Records are linked to form a tamper-evident hash chain: each record
// embeds the SHA-256 of the previous record's canonical form, so a
// later modification to any earlier record is detectable.
//
// The platform never deletes or mutates custody records. Correction
// requires appending a compensating CustodyException entry.
type CustodyRecord struct {
	ID           string        `bson:"_id,omitempty" json:"id"`
	TenantID     string        `bson:"tenant_id" json:"tenantId"`
	ShipmentID   string        `bson:"shipment_id" json:"shipmentId"`
	Sequence     int64         `bson:"sequence" json:"sequence"`
	Action       CustodyAction `bson:"action" json:"action"`
	OccurredAt   time.Time     `bson:"occurred_at" json:"occurredAt"`
	RecordedAt   time.Time     `bson:"recorded_at" json:"recordedAt"`
	Actor        CustodyActor  `bson:"actor" json:"actor"`
	Location     *GeoPoint     `bson:"location,omitempty" json:"location,omitempty"`
	LocationName string        `bson:"location_name,omitempty" json:"locationName,omitempty"`
	SealNumber   string        `bson:"seal_number,omitempty" json:"sealNumber,omitempty"`
	Signature    string        `bson:"signature,omitempty" json:"signature,omitempty"`
	Evidence     []string      `bson:"evidence,omitempty" json:"evidence,omitempty"`
	PrevHash     string        `bson:"prev_hash" json:"prevHash"`
	Hash         string        `bson:"hash" json:"hash"`
	Notes        string        `bson:"notes,omitempty" json:"notes,omitempty"`
}

// CustodyActor describes the human or organisation responsible for
// the custody transition. Free-form Name is retained for display; the
// ID enables joins to the Driver / user directory.
type CustodyActor struct {
	ID           string `bson:"id,omitempty" json:"id,omitempty"`
	Name         string `bson:"name" json:"name"`
	Role         string `bson:"role" json:"role"`
	Organisation string `bson:"organisation" json:"organisation"`
	VATNumber    string `bson:"vat_number,omitempty" json:"vatNumber,omitempty"`
}

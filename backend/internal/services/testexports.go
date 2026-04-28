package services

import (
	"context"

	"github.com/logitrack/backend/internal/modules/logistics"
)

// ComputeHashForTest exposes the internal computeHash helper to
// external test packages so they can build valid chain-of-custody
// fixtures without duplicating the algorithm. It is deliberately
// documented as "for tests" — production callers should use the
// ShipmentService.AppendCustody path which chains the hash correctly.
//
// computeHash returns (string, error); a marshal error here would
// indicate a programmer mistake (introducing an un-marshalable field
// onto CustodyRecord), so we panic for fast feedback during tests.
func ComputeHashForTest(r *logistics.CustodyRecord) string {
	h, err := computeHash(r)
	if err != nil {
		panic(err)
	}
	return h
}

// ComputeFromForTest exposes the pure-math ETA computation path so
// external test packages can exercise the smoother without standing up
// a real Mongo repository.
func (e *ETAService) ComputeFromForTest(ctx context.Context, shp *logistics.Shipment, current logistics.GeoPoint) (*ETAResult, error) {
	return e.computeFrom(ctx, shp, current)
}

package services

import (
	"context"

	"github.com/logitrack/backend/internal/models"
)

// ComputeHashForTest exposes the internal computeHash helper to
// external test packages so they can build valid chain-of-custody
// fixtures without duplicating the algorithm. It is deliberately
// documented as "for tests" — production callers should use the
// ShipmentService.AppendCustody path which chains the hash correctly.
func ComputeHashForTest(r *models.CustodyRecord) string { return computeHash(r) }

// ComputeFromForTest exposes the pure-math ETA computation path so
// external test packages can exercise the smoother without standing up
// a real Mongo repository.
func (e *ETAService) ComputeFromForTest(ctx context.Context, shp *models.Shipment, current models.GeoPoint) (*ETAResult, error) {
	return e.computeFrom(ctx, shp, current)
}

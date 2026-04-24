package demo

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/logitrack/backend/internal/models"
	"github.com/logitrack/backend/internal/repository"
	"github.com/logitrack/backend/internal/services"
)

// Seed creates the three demo shipments for the given tenant. It is
// idempotent: shipments whose deterministic IDs already exist are
// skipped without error so a developer can restart the compose stack
// without duplicating data.
func Seed(ctx context.Context, mongo *repository.MongoRepository, shipments *services.ShipmentService, _ services.RouteOptimizer, tenantID string, log *zap.Logger) error {
	if tenantID == "" {
		tenantID = "demo-tenant"
	}
	created := 0
	for _, r := range AllRoutes() {
		if _, err := mongo.FindShipment(ctx, tenantID, r.ID); err == nil {
			continue
		} else if !errors.Is(err, repository.ErrNotFound) {
			// Non-404 error — surface it so the operator can investigate.
			return err
		}
		s := r.ShipmentTemplate(tenantID)
		s.Waypoints = []models.Waypoint{}
		if err := shipments.CreateShipment(ctx, s); err != nil {
			log.Warn("demo seed create failed", zap.String("id", r.ID), zap.Error(err))
			continue
		}
		created++
	}
	log.Info("demo seed complete", zap.Int("created", created), zap.String("tenant", tenantID))
	return nil
}

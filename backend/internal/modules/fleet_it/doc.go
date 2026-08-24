// Package fleet_it bridges fleet operations with IT infrastructure
// management. It owns domain types for driver devices (tablets,
// phones, MDM status), telematics hardware (GPS units, data SIMs),
// and vehicle-IT equipment that the sistemista must track alongside
// traditional IT assets.
//
// Module contract (formal version in docs/DOMAIN-MODULES.md):
//
//   - Owns its persistent entities: DriverDevice, TelematicsUnit.
//
//   - Exposes domain validators but no I/O. Persistence and HTTP
//     wiring live in the platform layers (repository, handlers).
//
//   - Imports the cross-vertical primitives and may reference
//     logistics types (Vehicle, Driver) as foreign-key strings,
//     but does NOT import other domain modules directly.
//
// This module is toggled via MODULE_FLEET_IT (default false, see
// internal/config/config.go). When disabled, fleet-IT routes and
// collections are skipped.
package fleet_it

// Package logistics is the first vertical module on the LogiTrack
// kit. It owns the entities, validators and Italian regulatory
// primitives needed to track freight shipments end-to-end.
//
// Module contract (informal — see docs/DOMAIN-MODULES.md for the
// formal version):
//
//   - Owns its persistent entities: Shipment, Waypoint, TrackingEvent,
//     CustodyRecord, Vehicle, Driver, Geofence, plus the Italian
//     regulatory typed enums (ADRClass, ATPClass, LicenceCategory,
//     TelepassTollCode) that ride alongside them.
//   - Exposes domain validators (ValidatePlate, NewGeoPoint, the
//     Shipment.Validate state-machine guard) but no I/O. Persistence
//     and HTTP wiring live in the platform layers (repository, handlers).
//   - Imports the cross-vertical primitives (errors, time helpers) but
//     does NOT import any other domain module. Modules are leaf nodes
//     in the dependency graph.
//
// Subsequent vertical modules (wine, oil, cheese, etc.) follow the
// same contract and reuse the platform layers without reaching back
// into this package.
package logistics

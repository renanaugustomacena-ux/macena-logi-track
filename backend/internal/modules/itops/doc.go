// Package itops is the IT-operations vertical module on the LogiTrack
// kit. It owns the entities, validators and domain primitives needed
// for a sistemista to manage infrastructure, track assets, handle
// incidents, and monitor service health from a single dashboard.
//
// Module contract (formal version in docs/DOMAIN-MODULES.md):
//
//   - Owns its persistent entities: ITAsset, MonitoringCheck,
//     MonitoringAlert, Incident, WorkLogEntry, Certificate,
//     BackupJob, License.
//
//   - Exposes domain validators (ValidateAssetKind, ValidateAssetStatus,
//     Incident state-machine guard, priority-to-SLA mapping) but no I/O.
//     Persistence and HTTP wiring live in the platform layers
//     (repository, handlers).
//
//   - Imports the cross-vertical primitives (errors, time helpers) but
//     does NOT import any other domain module. Modules are leaf nodes
//     in the dependency graph.
//
// This module is toggled via MODULE_ITOPS (default true). When disabled,
// no itops routes are registered and no itops collections are created.
package itops

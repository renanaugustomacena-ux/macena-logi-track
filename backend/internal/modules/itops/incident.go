package itops

import (
	"fmt"
	"time"
)

// IncidentPriority maps to SLA targets.
type IncidentPriority string

const (
	PriorityCritical IncidentPriority = "critical" // 4 business hours
	PriorityHigh     IncidentPriority = "high"     // 8 business hours
	PriorityMedium   IncidentPriority = "medium"   // 24 business hours
	PriorityLow      IncidentPriority = "low"      // 72 business hours
)

var validPriorities = map[IncidentPriority]bool{
	PriorityCritical: true, PriorityHigh: true,
	PriorityMedium: true, PriorityLow: true,
}

// SLAHours returns the target resolution hours for the priority.
func (p IncidentPriority) SLAHours() int {
	switch p {
	case PriorityCritical:
		return 4
	case PriorityHigh:
		return 8
	case PriorityMedium:
		return 24
	case PriorityLow:
		return 72
	default:
		return 72
	}
}

// IncidentStatus tracks lifecycle state.
type IncidentStatus string

const (
	IncidentOpen       IncidentStatus = "open"
	IncidentAssigned   IncidentStatus = "assigned"
	IncidentInProgress IncidentStatus = "in_progress"
	IncidentPending    IncidentStatus = "pending"
	IncidentResolved   IncidentStatus = "resolved"
	IncidentClosed     IncidentStatus = "closed"
)

// validTransitions defines the incident state machine.
var validTransitions = map[IncidentStatus][]IncidentStatus{
	IncidentOpen:       {IncidentAssigned, IncidentClosed},
	IncidentAssigned:   {IncidentInProgress, IncidentClosed},
	IncidentInProgress: {IncidentPending, IncidentResolved, IncidentClosed},
	IncidentPending:    {IncidentInProgress, IncidentClosed},
	IncidentResolved:   {IncidentInProgress, IncidentClosed},
}

// CanTransition reports whether moving from the current status to
// target is valid per the incident state machine.
func CanTransition(from, to IncidentStatus) bool {
	targets, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == to {
			return true
		}
	}
	return false
}

// IncidentCategory classifies the type of issue.
type IncidentCategory string

const (
	CatHardware   IncidentCategory = "hardware"
	CatSoftware   IncidentCategory = "software"
	CatNetwork    IncidentCategory = "network"
	CatSecurity   IncidentCategory = "security"
	CatAccess     IncidentCategory = "access"
	CatPrinter    IncidentCategory = "printer"
	CatEmail      IncidentCategory = "email"
	CatTelephony  IncidentCategory = "telephony"
	CatTelematics IncidentCategory = "telematics"
	CatOther      IncidentCategory = "other"
)

var validCategories = map[IncidentCategory]bool{
	CatHardware: true, CatSoftware: true, CatNetwork: true,
	CatSecurity: true, CatAccess: true, CatPrinter: true,
	CatEmail: true, CatTelephony: true, CatTelematics: true,
	CatOther: true,
}

// Incident is an IT incident or service request tracked by the
// sistemista. The state machine mirrors a simplified ITIL v4
// incident lifecycle.
type Incident struct {
	ID            string           `bson:"_id,omitempty" json:"id"`
	TenantID      string           `bson:"tenant_id" json:"tenantId"`
	Reference     string           `bson:"reference" json:"reference"`
	Title         string           `bson:"title" json:"title"`
	Description   string           `bson:"description" json:"description"`
	Category      IncidentCategory `bson:"category" json:"category"`
	Priority      IncidentPriority `bson:"priority" json:"priority"`
	Status        IncidentStatus   `bson:"status" json:"status"`
	Reporter      string           `bson:"reporter" json:"reporter"`
	AssignedTo    string           `bson:"assigned_to,omitempty" json:"assignedTo,omitempty"`
	AffectedAsset string           `bson:"affected_asset,omitempty" json:"affectedAsset,omitempty"`
	Resolution    string           `bson:"resolution,omitempty" json:"resolution,omitempty"`
	WorkLog       []WorkLogEntry   `bson:"work_log" json:"workLog"`
	OpenedAt      time.Time        `bson:"opened_at" json:"openedAt"`
	ResolvedAt    *time.Time       `bson:"resolved_at,omitempty" json:"resolvedAt,omitempty"`
	ClosedAt      *time.Time       `bson:"closed_at,omitempty" json:"closedAt,omitempty"`
	SLATarget     time.Time        `bson:"sla_target" json:"slaTarget"`
	CreatedAt     time.Time        `bson:"created_at" json:"createdAt"`
	UpdatedAt     time.Time        `bson:"updated_at" json:"updatedAt"`
}

// WorkLogEntry records a timestamped action on an incident.
type WorkLogEntry struct {
	Author    string    `bson:"author" json:"author"`
	Action    string    `bson:"action" json:"action"`
	Note      string    `bson:"note" json:"note"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}

// Validate checks required fields and enum values.
func (inc *Incident) Validate() error {
	if inc.Title == "" {
		return fmt.Errorf("itops: incident title is required")
	}
	if !validCategories[inc.Category] {
		return fmt.Errorf("itops: unknown incident category %q", inc.Category)
	}
	if !validPriorities[inc.Priority] {
		return fmt.Errorf("itops: unknown incident priority %q", inc.Priority)
	}
	return nil
}

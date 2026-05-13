package itops

import (
	"fmt"
	"time"
)

// CheckStatus represents the health state of a monitoring check.
type CheckStatus string

const (
	CheckOK       CheckStatus = "ok"
	CheckWarning  CheckStatus = "warning"
	CheckCritical CheckStatus = "critical"
	CheckUnknown  CheckStatus = "unknown"
)

var validCheckStatuses = map[CheckStatus]bool{
	CheckOK: true, CheckWarning: true,
	CheckCritical: true, CheckUnknown: true,
}

// ValidateCheckStatus returns an error if s is not recognized.
func ValidateCheckStatus(s CheckStatus) error {
	if !validCheckStatuses[s] {
		return fmt.Errorf("itops: unknown check status %q", s)
	}
	return nil
}

// MonitoringCheck stores the latest result of a health check
// against an IT asset (ping, HTTP, disk, CPU, service, port, SNMP).
type MonitoringCheck struct {
	ID          string      `bson:"_id,omitempty" json:"id"`
	TenantID    string      `bson:"tenant_id" json:"tenantId"`
	AssetID     string      `bson:"asset_id" json:"assetId"`
	CheckName   string      `bson:"check_name" json:"checkName"`
	CheckType   string      `bson:"check_type" json:"checkType"`
	Status      CheckStatus `bson:"status" json:"status"`
	Output      string      `bson:"output,omitempty" json:"output,omitempty"`
	Metric      float64     `bson:"metric,omitempty" json:"metric,omitempty"`
	MetricUnit  string      `bson:"metric_unit,omitempty" json:"metricUnit,omitempty"`
	Threshold   float64     `bson:"threshold,omitempty" json:"threshold,omitempty"`
	LastChecked time.Time   `bson:"last_checked" json:"lastChecked"`
	NextCheck   time.Time   `bson:"next_check" json:"nextCheck"`
	CreatedAt   time.Time   `bson:"created_at" json:"createdAt"`
}

// Validate checks required fields on a MonitoringCheck.
func (c *MonitoringCheck) Validate() error {
	if c.TenantID == "" {
		return fmt.Errorf("itops: tenant_id is required")
	}
	if c.AssetID == "" {
		return fmt.Errorf("itops: asset_id is required")
	}
	if c.CheckName == "" {
		return fmt.Errorf("itops: check_name is required")
	}
	return nil
}

// Validate checks required fields on a MonitoringAlert.
func (a *MonitoringAlert) Validate() error {
	if a.TenantID == "" {
		return fmt.Errorf("itops: tenant_id is required")
	}
	if a.Title == "" {
		return fmt.Errorf("itops: title is required")
	}
	return nil
}

// AlertSeverity classifies alert urgency.
type AlertSeverity string

const (
	SeverityCritical AlertSeverity = "critical"
	SeverityHigh     AlertSeverity = "high"
	SeverityMedium   AlertSeverity = "medium"
	SeverityLow      AlertSeverity = "low"
)

// AlertStatus tracks the lifecycle of a monitoring alert.
type AlertStatus string

const (
	AlertOpen         AlertStatus = "open"
	AlertAcknowledged AlertStatus = "acknowledged"
	AlertResolved     AlertStatus = "resolved"
)

// MonitoringAlert records a monitoring event that requires attention.
type MonitoringAlert struct {
	ID          string        `bson:"_id,omitempty" json:"id"`
	TenantID    string        `bson:"tenant_id" json:"tenantId"`
	AssetID     string        `bson:"asset_id" json:"assetId"`
	CheckID     string        `bson:"check_id,omitempty" json:"checkId,omitempty"`
	Severity    AlertSeverity `bson:"severity" json:"severity"`
	Status      AlertStatus   `bson:"status" json:"status"`
	Title       string        `bson:"title" json:"title"`
	Description string        `bson:"description" json:"description"`
	OccurredAt  time.Time     `bson:"occurred_at" json:"occurredAt"`
	AckedAt     *time.Time    `bson:"acked_at,omitempty" json:"ackedAt,omitempty"`
	AckedBy     string        `bson:"acked_by,omitempty" json:"ackedBy,omitempty"`
	ResolvedAt  *time.Time    `bson:"resolved_at,omitempty" json:"resolvedAt,omitempty"`
	ResolvedBy  string        `bson:"resolved_by,omitempty" json:"resolvedBy,omitempty"`
	CreatedAt   time.Time     `bson:"created_at" json:"createdAt"`
}

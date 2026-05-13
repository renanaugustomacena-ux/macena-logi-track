package itops

import (
	"fmt"
	"time"
)

// Runbook is a stored operational procedure document.
type Runbook struct {
	ID           string    `bson:"_id,omitempty" json:"id"`
	TenantID     string    `bson:"tenant_id" json:"tenantId"`
	Title        string    `bson:"title" json:"title"`
	Category     string    `bson:"category" json:"category"`
	Content      string    `bson:"content" json:"content"`
	RelatedAssets []string  `bson:"related_assets,omitempty" json:"relatedAssets,omitempty"`
	LastReviewed *time.Time `bson:"last_reviewed,omitempty" json:"lastReviewed,omitempty"`
	Version      int       `bson:"version" json:"version"`
	CreatedAt    time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt    time.Time `bson:"updated_at" json:"updatedAt"`
}

// Validate checks required fields.
func (r *Runbook) Validate() error {
	if r.Title == "" {
		return fmt.Errorf("itops: runbook title is required")
	}
	if r.Category == "" {
		return fmt.Errorf("itops: runbook category is required")
	}
	if r.Content == "" {
		return fmt.Errorf("itops: runbook content is required")
	}
	return nil
}

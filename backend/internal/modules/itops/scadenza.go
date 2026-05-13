package itops

import "time"

// ScadenzaSeverity indicates urgency based on days remaining.
type ScadenzaSeverity string

const (
	SeverityOK   ScadenzaSeverity = "ok"
	SeverityWarn ScadenzaSeverity = "warn"
	SeverityBad  ScadenzaSeverity = "bad"
)

// Scadenza is a deadline item surfaced on the dashboard.
type Scadenza struct {
	ID       string           `json:"id"`
	Kind     string           `json:"kind"`
	Severity ScadenzaSeverity `json:"severity"`
	Title    string           `json:"title"`
	Detail   string           `json:"detail"`
	DueDate  time.Time        `json:"dueDate"`
	DaysLeft int              `json:"daysLeft"`
}

// ComputeSeverity returns severity based on days until deadline.
func ComputeSeverity(daysLeft int) ScadenzaSeverity {
	switch {
	case daysLeft < 0:
		return SeverityBad
	case daysLeft <= 30:
		return SeverityBad
	case daysLeft <= 90:
		return SeverityWarn
	default:
		return SeverityOK
	}
}

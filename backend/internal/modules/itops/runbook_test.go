package itops

import "testing"

func TestRunbookValidate(t *testing.T) {
	t.Parallel()
	valid := Runbook{
		Title:    "Server Down Response",
		Category: "incident",
		Content:  "## Steps\n1. Check ping\n2. Check iDRAC",
	}
	tests := []struct {
		name    string
		mod     func(*Runbook)
		wantErr bool
	}{
		{"valid", func(_ *Runbook) {}, false},
		{"missing_title", func(r *Runbook) { r.Title = "" }, true},
		{"missing_category", func(r *Runbook) { r.Category = "" }, true},
		{"missing_content", func(r *Runbook) { r.Content = "" }, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rb := valid
			tt.mod(&rb)
			err := rb.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

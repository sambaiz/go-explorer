package report

import (
	"encoding/json"
	"golang.org/x/xerrors"
	"io"
)

// Summary of report
type Summary struct {
	UpdatedAt           int64 `json:"updatedAt"` // Unix sec
	ActiveRepositoryNum int   `json:"activeRepositoryNum"`
}

// Module of report
type Module struct {
	Path        string `json:"path"`
	Description string `json:"description"`
	// Dependency count from active repositories
	ActiveDepCount int `json:"activeDepCount"`
}

// Report to write
type Report struct {
	Summary Summary  `json:"summary"`
	Modules []Module `json:"modules"`
}

// Write the report
func (r *Report) Write(writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	if err := encoder.Encode(r); err != nil {
		return xerrors.Errorf("failed to encode report: %w", err)
	}
	return nil
}

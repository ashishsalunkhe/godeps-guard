package types

import "time"

// HistoryRecord represents a single point in time of dependency data.
type HistoryRecord struct {
	Date          time.Time `json:"date"`
	Commit        string    `json:"commit"`
	DirectDeps    int       `json:"direct_deps"`
	TotalPackages int       `json:"total_packages"`
	BinarySize    int64     `json:"binary_size"`
}

// History representations a collection of records over time.
type History struct {
	Target  string          `json:"target"`
	Records []HistoryRecord `json:"records"`
}

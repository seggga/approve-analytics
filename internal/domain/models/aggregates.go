package models

import "time"

// Totals represents cumulative statistics on tasks
type Totals struct {
	Finished uint64 `json:"finished"`
	Declined uint64 `json:"declined"`
}

// Delay is a time lag on particular task ID
type Delay struct {
	ID  uint64        `json:"id"`
	Lag time.Duration `json:"lag"`
	// Lag uint64
}

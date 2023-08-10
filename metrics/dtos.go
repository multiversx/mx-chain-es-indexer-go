package metrics

import "time"

// ArgsAddIndexingData holds all the data needed for indexing metrics
type ArgsAddIndexingData struct {
	StatusCode int
	GotError   bool
	MessageLen uint64
	Topic      string
	Duration   time.Duration
}

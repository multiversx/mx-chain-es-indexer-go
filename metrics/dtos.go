package metrics

import "time"

// ArgsAddIndexingData holds all the data needed for indexing metrics
type ArgsAddIndexingData struct {
	GotError   bool
	MessageLen uint64
	Topic      string
	Duration   time.Duration
}

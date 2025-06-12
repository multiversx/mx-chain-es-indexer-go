package converters

// MillisecondsToSeconds will convert the provided milliseconds in seconds
func MillisecondsToSeconds(ms uint64) uint64 {
	return ms / 1000
}

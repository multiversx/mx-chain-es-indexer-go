package utils

import (
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
)

func LogExecutionTime(log logger.Logger, start time.Time, message string) {
	log.Debug(message, "duration in seconds", time.Since(start).Seconds())
}

package postgres

import (
	"context"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

var log = logger.GetOrCreate("indexer/postgres")

type logLevel int

type postgresLogger struct {
}

func newPostgresLogger() (gormLogger.Interface, error) {
	return &postgresLogger{}, nil
}

// LogMode log mode
func (pl *postgresLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	newlogger := *pl
	return &newlogger
}

// Info print info
func (pl *postgresLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	log.Info("was hereeeee")
	log.Info(msg, data)
}

// Warn print warn messages
func (pl *postgresLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	log.Info("was hereeeee")
	log.Warn(msg, data)
}

// Error print error messages
func (pl *postgresLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	log.Info("was hereeeee")
	log.Error(msg, data)
}

// Trace print sql message
func (pl *postgresLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	if rows == -1 {
		log.Info("PostgreSQL",
			"path", utils.FileWithLineNum(),
			"time", float64(elapsed.Nanoseconds())/1e6,
			"rows", "-",
			"\nquery", sql,
		)
	} else {
		log.Info("PostgreSQL",
			"path", utils.FileWithLineNum(),
			"time", float64(elapsed.Nanoseconds())/1e6,
			"rows", rows,
			"\nquery", sql,
		)
	}
}

package data

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm/logger"
)

var (
	_ logger.Interface = Logger{}

	// SlowThresholdMilliseconds is the threshold for slow database operations
	SlowThresholdMilliseconds = 200
)

// Logger is a custom logger for GORM
type Logger struct {
	log *log.Helper
}

// NewLogger creates a new Logger instance
func NewLogger(logger log.Logger) Logger {
	l := log.NewHelper(log.With(logger, "module", "data"))
	return Logger{log: l}
}

// LogMode implements the logger.Interface interface
func (l Logger) LogMode(lvl logger.LogLevel) logger.Interface {
	return l
}

// Info implements the logger.Interface interface
func (l Logger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.log.Info(fmt.Sprintf(msg, data...))
}

// Warn implements the logger.Interface interface
func (l Logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.log.Warn(fmt.Sprintf(msg, data...))
}

// Error implements the logger.Interface interface
func (l Logger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.log.Error(fmt.Sprintf(msg, data...))
}

// Trace implements the logger.Interface interface
func (l Logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsedMs := time.Since(begin).Milliseconds()

	// omit any values for batch inserts as they can be very long
	sql, rows := fc()
	if i := strings.Index(strings.ToLower(sql), "values"); i > 0 {
		sql = fmt.Sprintf("%sVALUES (...)", sql[:i])
	}

	if elapsedMs < 200 {
		l.log.Debug("database operation", "duration_ms", elapsedMs, "rows_affected", rows, "sql", sql)
	} else {
		l.log.Warn("database operation", "duration_ms", elapsedMs, "rows_affected", rows, "sql", sql)
	}
}

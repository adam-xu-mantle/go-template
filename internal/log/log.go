package log

import (
	"go-template/internal/conf"
	"os"

	kzap "github.com/go-kratos/kratos/contrib/log/zap/v2"
	"github.com/go-kratos/kratos/v2/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// convertLogLevel converts conf.LogLevel to log.Level
func convertLogLevel(level conf.LogLevel) log.Level {
	switch level {
	case conf.LogLevel_DEBUG:
		return log.LevelDebug
	case conf.LogLevel_INFO:
		return log.LevelInfo
	case conf.LogLevel_WARN:
		return log.LevelWarn
	case conf.LogLevel_ERROR:
		return log.LevelError
	case conf.LogLevel_FATAL:
		return log.LevelFatal
	default:
		return log.LevelInfo // default to info level
	}
}

// NewLogger creates a new logger.
func NewLogger(c *conf.Log) log.Logger {
	var filter *log.Filter
	logLevel := convertLogLevel(c.Level)

	if c.Format == conf.FormatType_JSON {
		encoderConfig := zapcore.EncoderConfig{
			MessageKey:  "msg",
			LineEnding:  zapcore.DefaultLineEnding,
			EncodeLevel: zapcore.CapitalLevelEncoder,
		}
		logEncoder := zapcore.NewJSONEncoder(encoderConfig)
		core := zapcore.NewCore(logEncoder, zapcore.AddSync(os.Stdout), zap.DebugLevel)
		zlogger := zap.New(core).WithOptions()
		logger := kzap.NewLogger(zlogger)
		filter = log.NewFilter(logger, log.FilterLevel(logLevel))
	} else {
		filter = log.NewFilter(log.NewStdLogger(os.Stdout), log.FilterLevel(logLevel))
	}

	return filter
}

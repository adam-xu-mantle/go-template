package log

import (
	"go-template/internal/conf"
	"os"

	kzap "github.com/go-kratos/kratos/contrib/log/zap/v2"
	"github.com/go-kratos/kratos/v2/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(c *conf.Log) log.Logger {

	var filter *log.Filter
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
		filter = log.NewFilter(logger, log.FilterLevel(log.Level(c.Level)))
	} else {
		filter = log.NewFilter(log.NewStdLogger(os.Stdout), log.FilterLevel(log.Level(c.Level)))
	}

	return filter
}

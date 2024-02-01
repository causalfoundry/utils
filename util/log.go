package util

import (
	"os"
	"runtime"
	"time"

	"github.com/rs/zerolog"
)

var DefaultLoggingLevel = zerolog.InfoLevel
var LOG = NewLogger("default")

func init() {
	zerolog.TimestampFieldName = "ts"
	zerolog.DurationFieldUnit = time.Second
}

func NewLoggerWithLevel(from string, minLevel zerolog.Level) zerolog.Logger {
	return zerolog.New(os.Stdout).Level(minLevel).With().Timestamp().Logger().Hook(
		zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
			if level == zerolog.ErrorLevel {
				_, file, line, _ := runtime.Caller(4)
				e.Str("file", file).Int("line", line)
			}
		}),
	).With().Str("from", from).Logger()
}

func NewLogger(from string) zerolog.Logger {
	return NewLoggerWithLevel(from, DefaultLoggingLevel)
}

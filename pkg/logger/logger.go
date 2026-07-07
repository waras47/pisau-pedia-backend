package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func New(debug bool) zerolog.Logger {
	level := zerolog.InfoLevel
	if debug {
		level = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = time.RFC3339

	return zerolog.New(os.Stdout).With().Timestamp().Logger()
}

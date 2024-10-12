package utils

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitializeLogger sets up the global logger configuration.
func InitializeLogger() {
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	})
}

// Log exposes the global logger for use in other packages.
func Log() *zerolog.Logger {
	return &log.Logger
}

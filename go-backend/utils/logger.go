package utils

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger initializes the zerolog logger with configuration
func InitLogger(logLevel string) {
	// Set up console writer for human-readable logs in development
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Pretty logging for development
	if logLevel == "debug" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Set log level
	switch logLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Info().Str("level", logLevel).Msg("Logger initialized")
}

// LogRequestMiddleware returns a Gin middleware for request logging
func LogRequestMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("ip", c.ClientIP()).
			Msg("Incoming request")

		c.Next()

		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Msg("Request completed")
	}
}

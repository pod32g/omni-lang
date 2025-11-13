// Package logging provides a centralized logger for the Omni compiler.
package logging

import (
	"os"
	"strings"
	"sync"

	slogger "github.com/pod32g/simple-logger"
)

var (
	initOnce sync.Once
	global   *slogger.Logger
)

// Logger returns the process-wide logger instance configured from environment variables.
func Logger() *slogger.Logger {
	initOnce.Do(func() {
		cfg := slogger.LoadConfigFromEnv()
		if _, ok := os.LookupEnv("LOG_OUTPUT"); !ok {
			cfg.Output = "stderr"
		}
		if _, ok := os.LookupEnv("LOG_COLORIZE"); !ok {
			cfg.Colorize = true
		}
		cfg.EnableCaller = false
		cfg.SyncWrites = true
		global = slogger.ApplyConfig(cfg)
	})
	return global
}

// SetLevel overrides the active log level for the shared logger.
func SetLevel(level slogger.LogLevel) {
	Logger().SetLevel(level)
}

// SetLevelByName adjusts the log level using a string such as "debug", "info", etc.
// Returns true when the level name is recognised.
func SetLevelByName(name string) bool {
	switch strings.ToUpper(name) {
	case "DEBUG":
		SetLevel(LevelDebug)
	case "INFO":
		SetLevel(LevelInfo)
	case "WARN", "WARNING":
		SetLevel(LevelWarn)
	case "ERROR", "ERR":
		SetLevel(LevelError)
	default:
		return false
	}
	return true
}

// Level aliases simplify call sites without importing simple-logger directly.
const (
	LevelDebug = slogger.DEBUG
	LevelInfo  = slogger.INFO
	LevelWarn  = slogger.WARN
	LevelError = slogger.ERROR
)

// Field exposes the structured field type from simple-logger.
type Field = slogger.Field

// Field constructors mirror simple-logger helpers for convenience.
var (
	String = slogger.String
	Int    = slogger.Int
	Bool   = slogger.Bool
	Error  = slogger.Error
)

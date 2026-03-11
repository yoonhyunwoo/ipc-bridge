package main

import (
	"log"
	"os"
	"strings"
)

type level int

const (
	levelError level = iota
	levelInfo
	levelDebug
)

type leveledLogger struct {
	level  level
	logger *log.Logger
}

func newLogger(raw string) *leveledLogger {
	lvl := levelInfo

	switch strings.ToLower(raw) {
	case "error":
		lvl = levelError
	case "debug":
		lvl = levelDebug
	}

	return &leveledLogger{
		level:  lvl,
		logger: log.New(os.Stderr, "[ipc-bridge] ", log.LstdFlags),
	}
}

func (l *leveledLogger) error(format string, args ...any) {
	if l.level >= levelError {
		l.logger.Printf("ERROR "+format, args...)
	}
}

func (l *leveledLogger) info(format string, args ...any) {
	if l.level >= levelInfo {
		l.logger.Printf("INFO "+format, args...)
	}
}

func (l *leveledLogger) debug(format string, args ...any) {
	if l.level >= levelDebug {
		l.logger.Printf("DEBUG "+format, args...)
	}
}

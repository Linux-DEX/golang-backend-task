package logger

import (
	"log"
	"os"
)

// Simple logger wrapper for backward compatibility
type Logger struct {
	logger *log.Logger
}

func New() *Logger {
	return &Logger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

func (l *Logger) Info(args ...interface{}) {
	l.logger.Println(args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.logger.Println(args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}

func (l *Logger) Warn(args ...interface{}) {
	l.logger.Println(args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}

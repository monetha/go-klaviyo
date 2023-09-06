package log

import "go.uber.org/zap"

// LeveledZapLogger is a wrapper around zap.SugaredLogger that implements the LeveledLogger interface.
// The LeveledLogger interface provides leveled logging with methods for logging messages at different levels (Error, Info, Debug, Warn).
// The methods accept a message string and a variadic number of key-value pairs.
type LeveledZapLogger struct {
	sl *zap.SugaredLogger
}

// Error logs an error message with the given key-value pairs.
func (l *LeveledZapLogger) Error(msg string, keysAndValues ...interface{}) {
	l.sl.Errorw(msg, keysAndValues...)
}

// Info logs an info message with the given key-value pairs.
func (l *LeveledZapLogger) Info(msg string, keysAndValues ...interface{}) {
	l.sl.Infow(msg, keysAndValues...)
}

// Debug logs a debug message with the given key-value pairs.
func (l *LeveledZapLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.sl.Debugw(msg, keysAndValues...)
}

// Warn logs a warning message with the given key-value pairs.
func (l *LeveledZapLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.sl.Warnw(msg, keysAndValues...)
}

// NewLeveledLogger returns a new instance of LeveledZapLogger by wrapping provided zap.Logger.
func NewLeveledLogger(logger *zap.Logger) *LeveledZapLogger {
	return &LeveledZapLogger{sl: logger.WithOptions(zap.AddCallerSkip(1)).Sugar()}
}

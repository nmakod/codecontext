package parser

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

// NopLogger is a no-op logger that discards all log messages
// This is the safe default for library code
type NopLogger struct{}

func (n NopLogger) Debug(msg string, fields ...LogField) {}
func (n NopLogger) Info(msg string, fields ...LogField)  {}
func (n NopLogger) Warn(msg string, fields ...LogField)  {}
func (n NopLogger) Error(msg string, err error, fields ...LogField) {}
func (n NopLogger) With(fields ...LogField) Logger { return n }

// StdLogger is a simple logger that writes to stderr for development/testing
// Production code should use a proper structured logger like logrus, zap, etc.
type StdLogger struct {
	output io.Writer
	prefix string
	level  LogLevel
}

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// NewStdLogger creates a new standard logger
func NewStdLogger(output io.Writer, level LogLevel) *StdLogger {
	if output == nil {
		output = os.Stderr // Never write to stdout in library code
	}
	
	return &StdLogger{
		output: output,
		prefix: "[parser] ",
		level:  level,
	}
}

// NewDevLogger creates a logger suitable for development (writes to stderr)
func NewDevLogger() *StdLogger {
	return NewStdLogger(os.Stderr, LogLevelInfo)
}

func (s *StdLogger) shouldLog(level LogLevel) bool {
	return level >= s.level
}

func (s *StdLogger) formatMessage(level LogLevel, msg string, fields []LogField) string {
	var parts []string
	
	// Add timestamp
	parts = append(parts, time.Now().Format("2006-01-02 15:04:05"))
	
	// Add level
	parts = append(parts, level.String())
	
	// Add message
	parts = append(parts, msg)
	
	// Add fields
	if len(fields) > 0 {
		var fieldStrs []string
		for _, field := range fields {
			fieldStrs = append(fieldStrs, fmt.Sprintf("%s=%v", field.Key, field.Value))
		}
		if len(fieldStrs) > 0 {
			parts = append(parts, fmt.Sprintf("[%s]", strings.Join(fieldStrs, " ")))
		}
	}
	
	return s.prefix + strings.Join(parts, " ")
}

func (s *StdLogger) log(level LogLevel, msg string, fields []LogField) {
	if !s.shouldLog(level) {
		return
	}
	
	formatted := s.formatMessage(level, msg, fields)
	fmt.Fprintln(s.output, formatted)
}

func (s *StdLogger) Debug(msg string, fields ...LogField) {
	s.log(LogLevelDebug, msg, fields)
}

func (s *StdLogger) Info(msg string, fields ...LogField) {
	s.log(LogLevelInfo, msg, fields)
}

func (s *StdLogger) Warn(msg string, fields ...LogField) {
	s.log(LogLevelWarn, msg, fields)
}

func (s *StdLogger) Error(msg string, err error, fields ...LogField) {
	// Add error to fields if provided
	errorFields := make([]LogField, len(fields))
	copy(errorFields, fields)
	
	if err != nil {
		errorFields = append(errorFields, LogField{Key: "error", Value: err.Error()})
		
		// Add additional context for ParseError
		if parseErr, ok := err.(*ParseError); ok {
			if parseErr.Path != "" {
				errorFields = append(errorFields, LogField{Key: "file_path", Value: parseErr.Path})
			}
			if parseErr.Language != "" {
				errorFields = append(errorFields, LogField{Key: "language", Value: parseErr.Language})
			}
			if parseErr.IsRecoveredPanic() {
				errorFields = append(errorFields, LogField{Key: "panic_recovered", Value: true})
				if len(parseErr.Stack) > 0 {
					errorFields = append(errorFields, LogField{Key: "stack_trace", Value: "available"})
				}
			}
		}
	}
	
	s.log(LogLevelError, msg, errorFields)
}

func (s *StdLogger) With(fields ...LogField) Logger {
	// For simplicity, we'll just return the same logger
	// A full implementation would create a new logger with persistent fields
	return s
}

// GoLogger wraps Go's standard logger for compatibility
type GoLogger struct {
	logger *log.Logger
	level  LogLevel
}

// NewGoLogger creates a logger that uses Go's standard log package
func NewGoLogger(logger *log.Logger, level LogLevel) *GoLogger {
	if logger == nil {
		// Use stderr, never stdout for library logging
		logger = log.New(os.Stderr, "[parser] ", log.LstdFlags)
	}
	
	return &GoLogger{
		logger: logger,
		level:  level,
	}
}

func (g *GoLogger) shouldLog(level LogLevel) bool {
	return level >= g.level
}

func (g *GoLogger) formatFields(fields []LogField) string {
	if len(fields) == 0 {
		return ""
	}
	
	var parts []string
	for _, field := range fields {
		parts = append(parts, fmt.Sprintf("%s=%v", field.Key, field.Value))
	}
	return " [" + strings.Join(parts, " ") + "]"
}

func (g *GoLogger) Debug(msg string, fields ...LogField) {
	if g.shouldLog(LogLevelDebug) {
		g.logger.Printf("DEBUG %s%s", msg, g.formatFields(fields))
	}
}

func (g *GoLogger) Info(msg string, fields ...LogField) {
	if g.shouldLog(LogLevelInfo) {
		g.logger.Printf("INFO %s%s", msg, g.formatFields(fields))
	}
}

func (g *GoLogger) Warn(msg string, fields ...LogField) {
	if g.shouldLog(LogLevelWarn) {
		g.logger.Printf("WARN %s%s", msg, g.formatFields(fields))
	}
}

func (g *GoLogger) Error(msg string, err error, fields ...LogField) {
	if g.shouldLog(LogLevelError) {
		errorFields := make([]LogField, len(fields))
		copy(errorFields, fields)
		
		if err != nil {
			errorFields = append(errorFields, LogField{Key: "error", Value: err.Error()})
		}
		
		g.logger.Printf("ERROR %s%s", msg, g.formatFields(errorFields))
	}
}

func (g *GoLogger) With(fields ...LogField) Logger {
	// For simplicity, return the same logger
	return g
}
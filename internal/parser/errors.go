package parser

import (
	"fmt"
	"path/filepath"
	"runtime/debug"
	"strings"
)

// Domain-specific error types
var (
	ErrEmptyContent       = fmt.Errorf("empty content provided")
	ErrUnsupportedLanguage = fmt.Errorf("unsupported language")
	ErrInvalidFilePath    = fmt.Errorf("invalid file path")
	ErrCacheFailure       = fmt.Errorf("cache operation failed")
	ErrParseTimeout       = fmt.Errorf("parsing operation timed out")
)

// ParseError represents a parsing error with context
type ParseError struct {
	Op       string // The operation that failed
	Path     string // File path being parsed
	Language string // Language being parsed
	Err      error  // Underlying error
	Recovery any    // Panic value if recovered from panic
	Stack    []byte // Stack trace if from panic
}

func (e *ParseError) Error() string {
	if e.Recovery != nil {
		return fmt.Sprintf("%s %s (%s): panic recovered: %v", e.Op, e.Path, e.Language, e.Recovery)
	}
	if e.Path != "" {
		return fmt.Sprintf("%s %s (%s): %v", e.Op, e.Path, e.Language, e.Err)
	}
	return fmt.Sprintf("%s (%s): %v", e.Op, e.Language, e.Err)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// IsRecoveredPanic returns true if this error was recovered from a panic
func (e *ParseError) IsRecoveredPanic() bool {
	return e.Recovery != nil
}

// GetStack returns the stack trace if available
func (e *ParseError) GetStack() []byte {
	return e.Stack
}

// NewParseError creates a new parse error
func NewParseError(op, path, language string, err error) *ParseError {
	return &ParseError{
		Op:       op,
		Path:     path,
		Language: language,
		Err:      err,
	}
}

// NewPanicError creates a parse error from a recovered panic
func NewPanicError(op, path, language string, recovery any) *ParseError {
	return &ParseError{
		Op:       op,
		Path:     path,
		Language: language,
		Recovery: recovery,
		Stack:    debug.Stack(),
	}
}

// CacheError represents cache-related errors
type CacheError struct {
	Op    string // Operation that failed (get, set, invalidate, etc.)
	Key   string // Cache key
	Err   error  // Underlying error
}

func (e *CacheError) Error() string {
	return fmt.Sprintf("cache %s %s: %v", e.Op, e.Key, e.Err)
}

func (e *CacheError) Unwrap() error {
	return e.Err
}

// ValidationError represents configuration validation errors
type ValidationError struct {
	Field string // Configuration field that failed validation
	Value any    // Invalid value
	Err   error  // Validation error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s=%v: %v", e.Field, e.Value, e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// validateFilePath performs input sanitization on file paths
func validateFilePath(filePath string) error {
	if filePath == "" {
		return nil // Empty path is allowed
	}
	
	// Check for null bytes (security risk)
	if strings.Contains(filePath, "\x00") {
		return fmt.Errorf("file path contains null bytes")
	}
	
	// Check for excessively long paths (DoS prevention)
	const maxPathLength = 4096
	if len(filePath) > maxPathLength {
		return fmt.Errorf("file path too long: %d > %d", len(filePath), maxPathLength)
	}
	
	// Check for directory traversal attempts
	cleanPath := filepath.Clean(filePath)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path traversal detected in: %s", filePath)
	}
	
	return nil
}
package parser

import (
	"context"
	"fmt"
	"runtime/debug"
)

// PanicHandler handles panic recovery with proper logging and context
type PanicHandler struct {
	logger Logger
}

// NewPanicHandler creates a new panic handler with the given logger
func NewPanicHandler(logger Logger) *PanicHandler {
	return &PanicHandler{
		logger: logger,
	}
}

// Recover recovers from panics and returns a proper error
// This should be called as: defer func() { err = h.Recover(ctx, "operation_name", err) }()
func (h *PanicHandler) Recover(ctx context.Context, op string, existingErr error) error {
	if r := recover(); r != nil {
		// Create panic error
		panicErr := &ParseError{
			Op:       op,
			Recovery: r,
			Stack:    debug.Stack(),
		}
		
		// Extract context information if available
		if ctx != nil {
			if reqID := RequestIDFromContext(ctx); reqID != "" {
				// Add request ID to operation for better tracking
				panicErr.Op = fmt.Sprintf("%s[%s]", op, reqID)
			}
			
			if filePath := FilePathFromContext(ctx); filePath != "" {
				panicErr.Path = filePath
			}
			
			if language := LanguageFromContext(ctx); language != "" {
				panicErr.Language = language
			}
		}
		
		// Log the panic with structured logging
		h.logger.Error("panic recovered", panicErr, 
			LogField{Key: "operation", Value: op},
			LogField{Key: "panic_value", Value: r},
			LogField{Key: "has_stack", Value: true},
		)
		
		return panicErr
	}
	
	return existingErr
}

// WithOperation wraps a function call with panic recovery
func (h *PanicHandler) WithOperation(ctx context.Context, op string, fn func() error) (err error) {
	defer func() {
		err = h.Recover(ctx, op, err)
	}()
	
	return fn()
}

// WithOperationReturn wraps a function call with panic recovery that returns a value
// Returns interface{} to avoid generics - callers should type assert
func (h *PanicHandler) WithOperationReturn(ctx context.Context, op string, fn func() (any, error)) (result any, err error) {
	defer func() {
		err = h.Recover(ctx, op, err)
	}()
	
	return fn()
}

// Context helpers for extracting information
type contextKey string

const (
	requestIDKey contextKey = "request_id"
	filePathKey  contextKey = "file_path" 
	languageKey  contextKey = "language"
)

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDFromContext extracts request ID from context
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqID, ok := ctx.Value(requestIDKey).(string); ok {
		return reqID
	}
	return ""
}

// WithFilePath adds a file path to the context  
func WithFilePath(ctx context.Context, filePath string) context.Context {
	return context.WithValue(ctx, filePathKey, filePath)
}

// FilePathFromContext extracts file path from context
func FilePathFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if path, ok := ctx.Value(filePathKey).(string); ok {
		return path
	}
	return ""
}

// WithLanguage adds a language to the context
func WithLanguage(ctx context.Context, language string) context.Context {
	return context.WithValue(ctx, languageKey, language)
}

// LanguageFromContext extracts language from context
func LanguageFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if lang, ok := ctx.Value(languageKey).(string); ok {
		return lang
	}
	return ""
}

// NopPanicHandler is a no-op panic handler for testing
type NopPanicHandler struct{}

func (n *NopPanicHandler) Recover(ctx context.Context, op string, existingErr error) error {
	if r := recover(); r != nil {
		return NewPanicError(op, "", "", r)
	}
	return existingErr
}

func (n *NopPanicHandler) WithOperation(ctx context.Context, op string, fn func() error) error {
	return fn()
}

func (n *NopPanicHandler) WithOperationReturn(ctx context.Context, op string, fn func() (any, error)) (any, error) {
	return fn()
}
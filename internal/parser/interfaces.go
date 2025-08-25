package parser

import (
	"time"

	"github.com/nuthan-ms/codecontext/pkg/types"
)

// Parser defines the interface for language parsers
type Parser interface {
	// Parse parses source code content and returns an AST
	Parse(content, filePath string) (*types.AST, error)
	
	// ExtractSymbols extracts symbols from a parsed AST
	ExtractSymbols(ast *types.AST) ([]*types.Symbol, error)
	
	// ExtractImports extracts import statements from a parsed AST
	ExtractImports(ast *types.AST) ([]*types.Import, error)
	
	// GetSupportedLanguages returns the list of languages this parser supports
	GetSupportedLanguages() []string
}

// Cache defines the interface for AST caching
type Cache interface {
	// Get retrieves an AST from the cache
	Get(key string, version ...string) (*types.VersionedAST, error)
	
	// Set stores an AST in the cache
	Set(key string, ast *types.VersionedAST) error
	
	// Invalidate removes an entry from the cache
	Invalidate(key string) error
	
	// Clear removes all entries from the cache
	Clear() error
	
	// Size returns the current number of cached entries
	Size() int
	
	// Stats returns cache statistics
	Stats() map[string]any
	
	// SetMaxSize configures the maximum cache size
	SetMaxSize(size int)
	
	// SetTTL configures the cache entry lifetime
	SetTTL(ttl time.Duration)
}

// IFrameworkDetector defines the interface for framework detection
type IFrameworkDetector interface {
	// DetectFramework analyzes content and returns framework information
	DetectFramework(content string) FrameworkInfo
	
	// GetSupportedFrameworks returns the list of frameworks this detector supports
	GetSupportedFrameworks() []string
}

// ExtractionStrategy defines the interface for different parsing strategies
type ExtractionStrategy interface {
	// Extract extracts AST nodes from content using a specific strategy
	Extract(content string, lines []string) []*types.ASTNode
	
	// SupportsFileSize returns true if this strategy can handle the given file size
	SupportsFileSize(sizeBytes int) bool
	
	// GetStrategyName returns a human-readable name for this strategy
	GetStrategyName() string
}

// Metrics defines the interface for performance metrics collection
type Metrics interface {
	// RecordParseTime records the time taken to parse a file
	RecordParseTime(language string, fileSize int, duration time.Duration)
	
	// RecordCacheHit records a cache hit
	RecordCacheHit(language string)
	
	// RecordCacheMiss records a cache miss
	RecordCacheMiss(language string)
	
	// RecordError records an error during parsing
	RecordError(language string, errorType string)
	
	// GetMetrics returns current metrics data
	GetMetrics() map[string]any
}

// Logger defines the interface for structured logging
type Logger interface {
	// Debug logs a debug message
	Debug(msg string, fields ...LogField)
	
	// Info logs an info message
	Info(msg string, fields ...LogField)
	
	// Warn logs a warning message
	Warn(msg string, fields ...LogField)
	
	// Error logs an error message
	Error(msg string, err error, fields ...LogField)
	
	// With returns a logger with additional context fields
	With(fields ...LogField) Logger
}

// LogField represents a structured logging field
type LogField struct {
	Key   string
	Value any
}

// FrameworkInfo contains information about detected frameworks
type FrameworkInfo struct {
	Name        string            `json:"name"`
	Version     string            `json:"version,omitempty"`
	IsDetected  bool              `json:"is_detected"`
	Confidence  float64           `json:"confidence"`
	Features    []string          `json:"features,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// ParserManager defines the main interface for the parser manager
type ParserManager interface {
	Parser
	
	// GetParser returns a parser for the specified language
	GetParser(language string) (Parser, error)
	
	// RegisterParser registers a new parser for a language
	RegisterParser(language string, parser Parser) error
	
	// SetCache configures the cache implementation
	SetCache(cache Cache)
	
	// SetLogger configures the logger implementation
	SetLogger(logger Logger)
	
	// SetMetrics configures the metrics implementation
	SetMetrics(metrics Metrics)
	
	// SetConfig updates the parser configuration
	SetConfig(config *ParserConfig) error
	
	// Close performs cleanup when shutting down
	Close() error
}

// Ensure our concrete types implement the interfaces
var (
	_ Cache = (*ASTCache)(nil)
	_ ParserManager = (*Manager)(nil)
)
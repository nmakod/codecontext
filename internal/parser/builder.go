package parser

import (
	"context"
	"fmt"
	"time"
	
	sitter "github.com/tree-sitter/go-tree-sitter"
)

// ManagerBuilder provides a clean way to construct Manager instances with dependency injection
type ManagerBuilder struct {
	logger       Logger
	cache        Cache
	config       *ParserConfig
	panicHandler *PanicHandler
	projectRoot  string
}

// NewManagerBuilder creates a new manager builder with safe defaults
func NewManagerBuilder() *ManagerBuilder {
	logger := NopLogger{} // Safe default - no output
	
	return &ManagerBuilder{
		logger:       logger,
		cache:        NewASTCache(),
		config:       DefaultConfig(),
		panicHandler: NewPanicHandler(logger),
		projectRoot:  ".",
	}
}

// WithLogger sets the logger for the manager
func (b *ManagerBuilder) WithLogger(logger Logger) *ManagerBuilder {
	b.logger = logger
	// Update panic handler to use the new logger
	b.panicHandler = NewPanicHandler(logger)
	return b
}

// WithCache sets the cache implementation
func (b *ManagerBuilder) WithCache(cache Cache) *ManagerBuilder {
	b.cache = cache
	return b
}

// WithConfig sets the configuration
func (b *ManagerBuilder) WithConfig(config *ParserConfig) *ManagerBuilder {
	b.config = config
	return b
}

// WithProjectRoot sets the project root directory
func (b *ManagerBuilder) WithProjectRoot(root string) *ManagerBuilder {
	b.projectRoot = root
	return b
}

// WithDevLogger sets up a development logger that writes to stderr
func (b *ManagerBuilder) WithDevLogger() *ManagerBuilder {
	devLogger := NewDevLogger()
	return b.WithLogger(devLogger)
}

// WithGoLogger sets up logging using Go's standard log package
func (b *ManagerBuilder) WithGoLogger() *ManagerBuilder {
	goLogger := NewGoLogger(nil, LogLevelInfo)
	return b.WithLogger(goLogger)
}

// Build creates and validates the Manager instance
func (b *ManagerBuilder) Build() (*Manager, error) {
	// Validate configuration
	if err := b.config.Validate(); err != nil {
		return nil, &ValidationError{
			Field: "config",
			Value: b.config,
			Err:   err,
		}
	}
	
	// Create manager with injected dependencies
	manager := &Manager{
		parsers:           make(map[string]*sitter.Parser),
		languages:         make(map[string]*sitter.Language),
		cache:             b.cache,
		frameworkDetector: NewFrameworkDetector(b.projectRoot),
		logger:            b.logger,
		panicHandler:     b.panicHandler,
		config:           b.config,
	}
	
	// Apply cache configuration
	if astCache, ok := b.cache.(*ASTCache); ok {
		astCache.SetMaxSize(b.config.Cache.MaxSize)
		astCache.SetTTL(b.config.Cache.TTL)
	}
	
	// Initialize languages
	manager.initLanguages()
	
	// Log successful initialization
	b.logger.Info("parser manager initialized",
		LogField{Key: "languages_count", Value: len(manager.languages)},
		LogField{Key: "cache_enabled", Value: b.config.Cache.Enabled},
		LogField{Key: "project_root", Value: b.projectRoot},
	)
	
	return manager, nil
}

// BuildWithContext creates a Manager with context for better error reporting
func (b *ManagerBuilder) BuildWithContext(ctx context.Context) (*Manager, error) {
	result, err := b.panicHandler.WithOperationReturn(ctx, "build_manager", func() (any, error) {
		manager, err := b.Build()
		return manager, err
	})
	
	if err != nil {
		return nil, err
	}
	
	if manager, ok := result.(*Manager); ok {
		return manager, nil
	}
	
	return nil, fmt.Errorf("internal error: build returned unexpected type")
}

// Common builder configurations for different use cases

// ForProduction creates a builder configured for production use
func ForProduction() *ManagerBuilder {
	return NewManagerBuilder().
		WithConfig(&ParserConfig{
			Cache: struct {
				MaxSize int           `yaml:"max_size" json:"max_size"`
				TTL     time.Duration `yaml:"ttl" json:"ttl"`
				Enabled bool          `yaml:"enabled" json:"enabled"`
			}{
				MaxSize: 5000,           // Larger cache for production
				TTL:     2 * time.Hour,  // Longer TTL
				Enabled: true,
			},
			Performance: struct {
				StreamingThreshold int  `yaml:"streaming_threshold" json:"streaming_threshold"`
				LimitedThreshold   int  `yaml:"limited_threshold" json:"limited_threshold"`
				MaxSymbols        int  `yaml:"max_symbols" json:"max_symbols"`
				EnableCaching     bool `yaml:"enable_caching" json:"enable_caching"`
			}{
				StreamingThreshold: StreamingThresholdBytes,
				LimitedThreshold:   LimitedThresholdBytes,
				MaxSymbols:        MaxSymbolsPerFile,
				EnableCaching:     true,
			},
			Dart: struct {
				EnableFlutterDetection bool `yaml:"enable_flutter_detection" json:"enable_flutter_detection"`
				MaxFileSize           int  `yaml:"max_file_size" json:"max_file_size"`
				EnableAsyncAnalysis   bool `yaml:"enable_async_analysis" json:"enable_async_analysis"`
			}{
				EnableFlutterDetection: true,
				MaxFileSize:           MaxFileSize,
				EnableAsyncAnalysis:   true,
			},
			Logging: struct {
				Level          string `yaml:"level" json:"level"`
				EnableMetrics  bool   `yaml:"enable_metrics" json:"enable_metrics"`
				EnableProfiling bool  `yaml:"enable_profiling" json:"enable_profiling"`
			}{
				Level:          "warn", // Less verbose for production
				EnableMetrics:  true,
				EnableProfiling: false,
			},
		})
}

// ForDevelopment creates a builder configured for development use
func ForDevelopment() *ManagerBuilder {
	return NewManagerBuilder().
		WithDevLogger(). // Enables stderr logging
		WithConfig(&ParserConfig{
			Cache: struct {
				MaxSize int           `yaml:"max_size" json:"max_size"`
				TTL     time.Duration `yaml:"ttl" json:"ttl"`
				Enabled bool          `yaml:"enabled" json:"enabled"`
			}{
				MaxSize: 1000,          // Smaller cache for development
				TTL:     30 * time.Minute, // Shorter TTL for faster iteration
				Enabled: true,
			},
			Performance: struct {
				StreamingThreshold int  `yaml:"streaming_threshold" json:"streaming_threshold"`
				LimitedThreshold   int  `yaml:"limited_threshold" json:"limited_threshold"`
				MaxSymbols        int  `yaml:"max_symbols" json:"max_symbols"`
				EnableCaching     bool `yaml:"enable_caching" json:"enable_caching"`
			}{
				StreamingThreshold: StreamingThresholdBytes,
				LimitedThreshold:   LimitedThresholdBytes,
				MaxSymbols:        MaxSymbolsPerFile,
				EnableCaching:     true,
			},
			Dart: struct {
				EnableFlutterDetection bool `yaml:"enable_flutter_detection" json:"enable_flutter_detection"`
				MaxFileSize           int  `yaml:"max_file_size" json:"max_file_size"`
				EnableAsyncAnalysis   bool `yaml:"enable_async_analysis" json:"enable_async_analysis"`
			}{
				EnableFlutterDetection: true,
				MaxFileSize:           MaxFileSize,
				EnableAsyncAnalysis:   true,
			},
			Logging: struct {
				Level          string `yaml:"level" json:"level"`
				EnableMetrics  bool   `yaml:"enable_metrics" json:"enable_metrics"`
				EnableProfiling bool  `yaml:"enable_profiling" json:"enable_profiling"`
			}{
				Level:          "debug", // Verbose for development
				EnableMetrics:  true,
				EnableProfiling: true,
			},
		})
}

// ForTesting creates a builder configured for testing
func ForTesting() *ManagerBuilder {
	return NewManagerBuilder(). // Uses NopLogger by default
		WithConfig(&ParserConfig{
			Cache: struct {
				MaxSize int           `yaml:"max_size" json:"max_size"`
				TTL     time.Duration `yaml:"ttl" json:"ttl"`
				Enabled bool          `yaml:"enabled" json:"enabled"`
			}{
				MaxSize: 100,                // Small cache for tests
				TTL:     1 * time.Minute,    // Very short TTL
				Enabled: false,              // Disable cache for predictable tests
			},
			Performance: struct {
				StreamingThreshold int  `yaml:"streaming_threshold" json:"streaming_threshold"`
				LimitedThreshold   int  `yaml:"limited_threshold" json:"limited_threshold"`
				MaxSymbols        int  `yaml:"max_symbols" json:"max_symbols"`
				EnableCaching     bool `yaml:"enable_caching" json:"enable_caching"`
			}{
				StreamingThreshold: StreamingThresholdBytes,
				LimitedThreshold:   LimitedThresholdBytes,
				MaxSymbols:        1000, // Lower limits for faster tests
				EnableCaching:     false,
			},
			Dart: struct {
				EnableFlutterDetection bool `yaml:"enable_flutter_detection" json:"enable_flutter_detection"`
				MaxFileSize           int  `yaml:"max_file_size" json:"max_file_size"`
				EnableAsyncAnalysis   bool `yaml:"enable_async_analysis" json:"enable_async_analysis"`
			}{
				EnableFlutterDetection: true,
				MaxFileSize:           MaxFileSize,
				EnableAsyncAnalysis:   false, // Simpler for tests
			},
			Logging: struct {
				Level          string `yaml:"level" json:"level"`
				EnableMetrics  bool   `yaml:"enable_metrics" json:"enable_metrics"`
				EnableProfiling bool  `yaml:"enable_profiling" json:"enable_profiling"`
			}{
				Level:          "error", // Minimal logging in tests
				EnableMetrics:  false,
				EnableProfiling: false,
			},
		})
}
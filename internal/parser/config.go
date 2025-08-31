package parser

import "time"

// ParserConstants defines all configuration constants to replace magic numbers
const (
	// File size thresholds for parser strategy selection
	StreamingThresholdBytes = 200 * 1024 // Files larger than 200KB use streaming parser
	LimitedThresholdBytes   = 50 * 1024  // Files larger than 50KB use limited extraction
	
	// Processing limits to prevent resource exhaustion
	MaxSymbolsPerFile      = 10000       // Maximum symbols to extract per file
	MaxNestingDepth        = 100         // Maximum nesting depth for classes/methods
	MaxLineLength          = 100000      // Maximum line length to process
	MaxFileSize            = 10 * 1024 * 1024 // Maximum file size (10MB)
	
	// Cache configuration
	DefaultCacheMaxSize    = 1000        // Default maximum cache entries
	DefaultCacheTTL        = time.Hour   // Default cache entry lifetime
	
	// Performance tuning
	ChunkSize              = 64 * 1024   // Size of chunks for streaming parser
	RegexTimeout           = 5 * time.Second // Timeout for regex operations
	
	// Symbol extraction limits (to prevent excessive processing)
	MaxClassesPerFile      = 1000        // Maximum classes to extract
	MaxMethodsPerClass     = 500         // Maximum methods per class
	MaxVariablesPerClass   = 1000        // Maximum variables per class
)

// ParserConfig holds runtime configuration options
type ParserConfig struct {
	Cache struct {
		MaxSize    int           `yaml:"max_size" json:"max_size"`
		TTL        time.Duration `yaml:"ttl" json:"ttl"`
		Enabled    bool          `yaml:"enabled" json:"enabled"`
	} `yaml:"cache" json:"cache"`
	
	Performance struct {
		StreamingThreshold int  `yaml:"streaming_threshold" json:"streaming_threshold"`
		LimitedThreshold   int  `yaml:"limited_threshold" json:"limited_threshold"`
		MaxSymbols        int  `yaml:"max_symbols" json:"max_symbols"`
		EnableCaching     bool `yaml:"enable_caching" json:"enable_caching"`
	} `yaml:"performance" json:"performance"`
	
	Dart struct {
		EnableFlutterDetection bool `yaml:"enable_flutter_detection" json:"enable_flutter_detection"`
		MaxFileSize           int  `yaml:"max_file_size" json:"max_file_size"`
		EnableAsyncAnalysis   bool `yaml:"enable_async_analysis" json:"enable_async_analysis"`
	} `yaml:"dart" json:"dart"`
	
	Cpp struct {
		MaxNestingDepth       int  `yaml:"max_nesting_depth" json:"max_nesting_depth"`
		MaxTemplateDepth      int  `yaml:"max_template_depth" json:"max_template_depth"`
		MaxClassesPerFile     int  `yaml:"max_classes_per_file" json:"max_classes_per_file"`
		MaxMethodsPerClass    int  `yaml:"max_methods_per_class" json:"max_methods_per_class"`
		MaxFileSize           int  `yaml:"max_file_size" json:"max_file_size"`
		EnableVirtualDetection bool `yaml:"enable_virtual_detection" json:"enable_virtual_detection"`
		ParseTimeout          time.Duration `yaml:"parse_timeout" json:"parse_timeout"`
		StrictTimeoutEnforcement bool `yaml:"strict_timeout_enforcement" json:"strict_timeout_enforcement"`
	} `yaml:"cpp" json:"cpp"`
	
	Logging struct {
		Level          string `yaml:"level" json:"level"`
		EnableMetrics  bool   `yaml:"enable_metrics" json:"enable_metrics"`
		EnableProfiling bool  `yaml:"enable_profiling" json:"enable_profiling"`
	} `yaml:"logging" json:"logging"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *ParserConfig {
	config := &ParserConfig{}
	
	// Cache defaults
	config.Cache.MaxSize = DefaultCacheMaxSize
	config.Cache.TTL = DefaultCacheTTL
	config.Cache.Enabled = true
	
	// Performance defaults
	config.Performance.StreamingThreshold = StreamingThresholdBytes
	config.Performance.LimitedThreshold = LimitedThresholdBytes
	config.Performance.MaxSymbols = MaxSymbolsPerFile
	config.Performance.EnableCaching = true
	
	// Dart-specific defaults
	config.Dart.EnableFlutterDetection = true
	config.Dart.MaxFileSize = MaxFileSize
	config.Dart.EnableAsyncAnalysis = true
	
	// C++ specific defaults
	config.Cpp.MaxNestingDepth = MaxNestingDepth
	config.Cpp.MaxTemplateDepth = 20 // Reasonable template depth
	config.Cpp.MaxClassesPerFile = MaxClassesPerFile
	config.Cpp.MaxMethodsPerClass = MaxMethodsPerClass
	config.Cpp.MaxFileSize = MaxFileSize
	config.Cpp.EnableVirtualDetection = true
	config.Cpp.ParseTimeout = 30 * time.Second
	config.Cpp.StrictTimeoutEnforcement = false // Default to lenient mode
	
	// Logging defaults
	config.Logging.Level = "info"
	config.Logging.EnableMetrics = true
	config.Logging.EnableProfiling = false
	
	return config
}

// Validate ensures the configuration values are valid
func (c *ParserConfig) Validate() error {
	if c.Cache.MaxSize <= 0 {
		c.Cache.MaxSize = DefaultCacheMaxSize
	}
	
	if c.Cache.TTL <= 0 {
		c.Cache.TTL = DefaultCacheTTL
	}
	
	if c.Performance.StreamingThreshold <= c.Performance.LimitedThreshold {
		c.Performance.StreamingThreshold = StreamingThresholdBytes
		c.Performance.LimitedThreshold = LimitedThresholdBytes
	}
	
	if c.Performance.MaxSymbols <= 0 {
		c.Performance.MaxSymbols = MaxSymbolsPerFile
	}
	
	if c.Dart.MaxFileSize <= 0 {
		c.Dart.MaxFileSize = MaxFileSize
	}
	
	return nil
}
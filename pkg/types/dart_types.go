package types

// Dart-specific symbol types extending the base SymbolType
const (
	// Dart 3.0+ specific symbol types
	SymbolTypeMixin     SymbolType = "mixin"
	SymbolTypeExtension SymbolType = "extension"
	SymbolTypeEnum      SymbolType = "enum"
	SymbolTypeTypedef   SymbolType = "typedef"
	
	// Flutter-specific symbol types
	SymbolTypeWidget           SymbolType = "widget"
	SymbolTypeBuildMethod      SymbolType = "build_method"
	SymbolTypeLifecycleMethod  SymbolType = "lifecycle_method"
	SymbolTypeStateClass       SymbolType = "state_class"
)

// DartSymbolMetadata contains Dart-specific metadata for symbols
type DartSymbolMetadata struct {
	// Basic Dart features
	IsAsync     bool     `json:"is_async,omitempty"`
	IsGenerator bool     `json:"is_generator,omitempty"`
	IsAbstract  bool     `json:"is_abstract,omitempty"`
	
	// Flutter-specific
	FlutterType      string `json:"flutter_type,omitempty"`      // "widget", "state", etc.
	WidgetType       string `json:"widget_type,omitempty"`       // "stateless", "stateful"
	HasBuildMethod   bool   `json:"has_build_method,omitempty"`
	HasOverride      bool   `json:"has_override,omitempty"`
	
	// File relationships (for part files)
	IsPartFile       bool     `json:"is_part_file,omitempty"`
	PartOfFile       string   `json:"part_of_file,omitempty"`
	PartFiles        []string `json:"part_files,omitempty"`
}

// DartParseMetadata contains information about how a Dart file was parsed
type DartParseMetadata struct {
	Parser       string `json:"parser"`        // "tree-sitter", "regex", "mock"
	ParseQuality string `json:"parse_quality"` // "complete", "partial", "basic"
	HasFlutter   bool   `json:"has_flutter"`
	HasErrors    bool   `json:"has_errors"`
	ErrorCount   int    `json:"error_count,omitempty"`
}
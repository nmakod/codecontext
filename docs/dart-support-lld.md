# Dart Language Support - Low-Level Design

## Executive Summary

This document outlines the complete Low-Level Design (LLD) for adding comprehensive Dart language support to CodeContext, including full Flutter framework intelligence. The implementation follows core software engineering principles (SOLID, DRY, KISS, YAGNI) and employs a test-driven development (TDD) approach.

## Goals

- **Primary**: 100% Dart language support including modern Dart 3.0+ features
- **Secondary**: Deep Flutter framework intelligence and state management pattern recognition
- **Performance**: Parse Dart files with comparable speed to existing language parsers
- **Reliability**: Graceful handling of malformed code and edge cases
- **Maintainability**: Clean, extensible architecture following established patterns

## Architecture Overview

### High-Level Components

```
CodeContext
├── Parser Layer
│   ├── manager.go (existing) - Extended with Dart support
│   ├── dart.go (new) - Dart-specific parsing logic
│   └── dart_symbols.go (new) - Dart symbol extraction
├── Framework Layer
│   └── flutter.go (new) - Flutter pattern detection
├── Analysis Layer
│   └── dart_analyzer.go (new) - Dart-specific analysis
└── Types Layer
    └── dart_types.go (new) - Dart-specific type definitions
```

### Design Principles Applied

- **Single Responsibility Principle (SRP)**: Each component has one clear responsibility
- **Open/Closed Principle**: Extend existing interfaces without modifying core code
- **Interface Segregation**: Small, focused interfaces over monolithic ones
- **Dependency Inversion**: Depend on abstractions, not implementations
- **DRY**: Reuse existing parser patterns and infrastructure
- **KISS**: Start simple, add complexity only when needed
- **YAGNI**: Build features incrementally based on actual needs

## Detailed Component Design

### 1. Tree-sitter Integration Strategy

#### Option Analysis

**Selected Approach**: Direct integration with custom Go bindings

```go
// internal/parser/dart.go
package parser

import (
    "fmt"
    "strings"
    "time"
    
    "github.com/nuthan-ms/codecontext/pkg/types"
    sitter "github.com/tree-sitter/go-tree-sitter"
    dart "github.com/UserNobody14/tree-sitter-dart/bindings/go" // Custom bindings
)

type DartParser struct {
    parser   *sitter.Parser
    language *sitter.Language
    timeout  time.Duration
}

func NewDartParser() (*DartParser, error) {
    parser := sitter.NewParser()
    language := sitter.NewLanguage(dart.Language())
    
    parser.SetLanguage(language)
    parser.SetTimeout(30 * time.Second)
    
    return &DartParser{
        parser:   parser,
        language: language,
        timeout:  30 * time.Second,
    }, nil
}
```

**Rationale**: 
- Consistent with existing codebase patterns
- Minimal changes to core architecture
- Leverages existing tree-sitter infrastructure
- Fast implementation and testing

### 2. Symbol Extraction Architecture

#### Dart-Specific Symbol Types

```go
// pkg/types/dart_types.go
const (
    // Dart 3.0+ specific
    SymbolTypeSealedClass   SymbolType = "sealed_class"
    SymbolTypeRecord        SymbolType = "record"
    SymbolTypePattern       SymbolType = "pattern"
    SymbolTypeMixin         SymbolType = "mixin"
    SymbolTypeExtension     SymbolType = "extension"
    SymbolTypeEnum          SymbolType = "enum"
    SymbolTypeTypedef       SymbolType = "typedef"
    
    // Flutter-specific
    SymbolTypeWidget       SymbolType = "widget"
    SymbolTypeState        SymbolType = "state"
    SymbolTypeProvider     SymbolType = "provider"
    SymbolTypeBuildMethod  SymbolType = "build_method"
    SymbolTypeHook         SymbolType = "hook"
)

// Dart-specific metadata
type DartSymbolMetadata struct {
    IsAsync        bool     `json:"is_async"`
    IsGenerator    bool     `json:"is_generator"`
    ClassModifiers []string `json:"class_modifiers,omitempty"`
    FlutterType    string   `json:"flutter_type,omitempty"`
    StateManagement string  `json:"state_management,omitempty"`
}
```

#### Symbol Extraction Strategy

```go
// internal/parser/dart_symbols.go
func (m *Manager) nodeToSymbolDart(node *types.ASTNode, filePath, language string) *types.Symbol {
    // Phase 1: Basic Dart constructs
    switch node.Type {
    case "class_declaration":
        return m.extractDartClass(node, filePath, language)
    case "function_declaration", "method_declaration":
        return m.extractDartFunction(node, filePath, language)
    case "variable_declaration":
        return m.extractDartVariable(node, filePath, language)
    case "import_directive":
        return m.extractDartImport(node, filePath, language)
    
    // Phase 2: Advanced Dart features
    case "mixin_declaration":
        return m.extractDartMixin(node, filePath, language)
    case "extension_declaration":
        return m.extractDartExtension(node, filePath, language)
    case "enum_declaration":
        return m.extractDartEnum(node, filePath, language)
    
    // Phase 3: Dart 3.0+ features
    case "sealed_class_declaration":
        return m.extractSealedClass(node, filePath, language)
    case "record_type_annotation", "record_literal":
        return m.extractRecord(node, filePath, language)
    case "pattern_variable_declaration":
        return m.extractPattern(node, filePath, language)
    
    // Phase 4: Part files
    case "part_directive":
        return m.extractPartDirective(node, filePath, language)
    case "part_of_directive":
        return m.extractPartOfDirective(node, filePath, language)
    
    default:
        // Check for Flutter-specific patterns
        return m.extractFlutterSymbol(node, filePath, language)
    }
}

// Example implementation for class extraction
func (m *Manager) extractDartClass(node *types.ASTNode, filePath, language string) *types.Symbol {
    name := m.extractSymbolName(node)
    symbolType := types.SymbolTypeClass
    
    // Flutter widget detection
    if m.isFlutterWidget(node) {
        symbolType = SymbolTypeWidget
    }
    
    // Extract class modifiers (Dart 3.0)
    modifiers := m.extractClassModifiers(node)
    
    symbol := &types.Symbol{
        Id:           types.SymbolId(fmt.Sprintf("class-%s-%d", filePath, node.Location.Line)),
        Name:         name,
        Type:         symbolType,
        Location:     convertLocation(node.Location),
        Language:     language,
        Hash:         calculateHash(node.Value),
        LastModified: time.Now(),
        Metadata: map[string]interface{}{
            "modifiers": modifiers,
        },
    }
    
    return symbol
}
```

### 3. Flutter Framework Intelligence

#### Flutter Pattern Detection

```go
// internal/framework/flutter.go
package framework

type FlutterDetector struct {
    patterns map[string]*regexp.Regexp
}

func NewFlutterDetector() *FlutterDetector {
    return &FlutterDetector{
        patterns: map[string]*regexp.Regexp{
            "stateless_widget": regexp.MustCompile(`extends\s+StatelessWidget`),
            "stateful_widget":  regexp.MustCompile(`extends\s+StatefulWidget`),
            "hook_widget":      regexp.MustCompile(`extends\s+HookWidget`),
            "provider":         regexp.MustCompile(`Provider<[^>]+>`),
            "riverpod":         regexp.MustCompile(`(StateProvider|FutureProvider|StreamProvider)`),
            "bloc":            regexp.MustCompile(`extends\s+(Bloc|Cubit)`),
        },
    }
}

func (d *FlutterDetector) DetectFramework(filePath, language, content string) string {
    if language != "dart" {
        return ""
    }
    
    // Check for Flutter imports
    if !strings.Contains(content, "package:flutter/") {
        return ""
    }
    
    // Detect widget types
    for pattern, regex := range d.patterns {
        if regex.MatchString(content) {
            return mapPatternToFramework(pattern)
        }
    }
    
    return "dart" // Plain Dart, not Flutter
}

// State management detection
func (d *FlutterDetector) DetectStateManagement(content string) string {
    patterns := map[string]string{
        "riverpod":  `package:flutter_riverpod/`,
        "bloc":      `package:flutter_bloc/`,
        "provider":  `package:provider/`,
        "getx":      `package:get/`,
    }
    
    for name, pattern := range patterns {
        if strings.Contains(content, pattern) {
            return name
        }
    }
    
    return ""
}
```

#### Widget Analysis

```go
// Extract Flutter-specific symbols
func (m *Manager) extractFlutterSymbol(node *types.ASTNode, filePath, language string) *types.Symbol {
    // Build method detection
    if node.Type == "method_declaration" {
        methodName := m.extractSymbolName(node)
        if methodName == "build" && m.hasOverrideAnnotation(node) {
            return &types.Symbol{
                Id:       types.SymbolId(fmt.Sprintf("build-%s-%d", filePath, node.Location.Line)),
                Name:     "build",
                Type:     SymbolTypeBuildMethod,
                Location: convertLocation(node.Location),
                Language: language,
                Signature: "Widget build(BuildContext context)",
                Metadata: map[string]interface{}{
                    "flutter":  true,
                    "override": true,
                },
            }
        }
    }
    
    // Provider pattern detection
    if node.Type == "variable_declaration" && m.isProvider(node) {
        return &types.Symbol{
            Id:       types.SymbolId(fmt.Sprintf("provider-%s-%d", filePath, node.Location.Line)),
            Name:     m.extractSymbolName(node),
            Type:     SymbolTypeProvider,
            Location: convertLocation(node.Location),
            Language: language,
            Metadata: map[string]interface{}{
                "provider_type": m.extractProviderType(node),
            },
        }
    }
    
    return nil
}
```

### 4. Advanced Dart Features

#### Part Files System

```go
// Handle Dart part files
type DartCompilationUnit struct {
    MainFile  string
    PartFiles []string
    Symbols   map[string][]*types.Symbol
    mutex     sync.RWMutex
}

func (m *Manager) ParseDartCompilationUnit(mainFile string) (*DartCompilationUnit, error) {
    unit := &DartCompilationUnit{
        MainFile: mainFile,
        Symbols:  make(map[string][]*types.Symbol),
    }
    
    // Parse main file
    ast, err := m.ParseFile(mainFile, types.Language{Name: "dart"})
    if err != nil {
        return nil, err
    }
    
    // Extract part directives
    parts := m.extractPartFiles(ast)
    unit.PartFiles = parts
    
    // Parse all part files
    for _, partFile := range parts {
        partPath := filepath.Join(filepath.Dir(mainFile), partFile)
        if _, err := os.Stat(partPath); os.IsNotExist(err) {
            continue // Part file might not exist yet
        }
        
        partAST, err := m.ParseFile(partPath, types.Language{Name: "dart"})
        if err != nil {
            continue
        }
        
        symbols, _ := m.ExtractSymbols(partAST)
        unit.Symbols[partPath] = symbols
    }
    
    return unit, nil
}
```

#### Async Pattern Analysis

```go
// Detect async patterns
func (m *Manager) analyzeAsyncPatterns(node *types.ASTNode) map[string]interface{} {
    metadata := make(map[string]interface{})
    
    // Check for async keywords
    if strings.Contains(node.Value, "async") {
        metadata["async"] = true
        
        // Check for generators
        if strings.Contains(node.Value, "async*") {
            metadata["generator"] = "async"
        } else if strings.Contains(node.Value, "sync*") {
            metadata["generator"] = "sync"
        }
    }
    
    // Check for Future/Stream patterns
    if strings.Contains(node.Value, "Future<") {
        metadata["returns_future"] = true
    }
    if strings.Contains(node.Value, "Stream<") {
        metadata["returns_stream"] = true
    }
    
    return metadata
}
```

## Performance Considerations

### 1. Parsing Optimization

```go
// Streaming parser for large files
func (p *DartParser) ParseLargeFile(filePath string) (*types.AST, error) {
    stat, err := os.Stat(filePath)
    if err != nil {
        return nil, err
    }
    
    // Use streaming for files > 1MB
    if stat.Size() > 1024*1024 {
        return p.parseStreaming(filePath)
    }
    
    return p.parseStandard(filePath)
}

func (p *DartParser) parseStreaming(filePath string) (*types.AST, error) {
    // Implement chunked parsing for very large files
    // This prevents memory exhaustion on generated Dart files
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    
    // Parse in chunks, build incremental AST
    // Implementation details...
    return nil, nil
}
```

### 2. Caching Strategy

```go
// Enhanced caching for Dart compilation units
type DartCache struct {
    compilationUnits map[string]*DartCompilationUnit
    lastModified     map[string]time.Time
    mutex            sync.RWMutex
}

func (c *DartCache) GetCompilationUnit(mainFile string) (*DartCompilationUnit, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    unit, exists := c.compilationUnits[mainFile]
    if !exists {
        return nil, false
    }
    
    // Check if any part files have been modified
    if c.isStale(mainFile) {
        return nil, false
    }
    
    return unit, true
}
```

### 3. Memory Management

```go
// Memory-efficient symbol extraction
var dartSymbolPool = sync.Pool{
    New: func() interface{} {
        return &types.Symbol{}
    },
}

func getDartSymbol() *types.Symbol {
    symbol := dartSymbolPool.Get().(*types.Symbol)
    // Reset fields
    *symbol = types.Symbol{}
    return symbol
}

func putDartSymbol(symbol *types.Symbol) {
    dartSymbolPool.Put(symbol)
}
```

## Error Handling Strategy

### 1. Graceful Degradation

```go
// Fallback parser for when tree-sitter fails
type DartFallbackParser struct {
    regexPatterns map[string]*regexp.Regexp
}

func (p *DartFallbackParser) ParseBasicSymbols(content string) ([]*types.Symbol, error) {
    var symbols []*types.Symbol
    
    // Class detection
    classPattern := regexp.MustCompile(`(?m)^(?:abstract\s+)?(?:final\s+)?class\s+(\w+)`)
    for _, match := range classPattern.FindAllStringSubmatch(content, -1) {
        symbols = append(symbols, &types.Symbol{
            Name: match[1],
            Type: types.SymbolTypeClass,
        })
    }
    
    // Function detection
    funcPattern := regexp.MustCompile(`(?m)^(?:\w+\s+)*(\w+)\s*\([^)]*\)\s*{`)
    for _, match := range funcPattern.FindAllStringSubmatch(content, -1) {
        symbols = append(symbols, &types.Symbol{
            Name: match[1],
            Type: types.SymbolTypeFunction,
        })
    }
    
    return symbols, nil
}

// Main parsing with fallback
func (m *Manager) ParseDartWithFallback(filePath string) (*types.AST, error) {
    // Try tree-sitter first
    if ast, err := m.parseDartTreeSitter(filePath); err == nil {
        return ast, nil
    }
    
    // Fallback to regex-based parsing
    return m.parseDartFallback(filePath)
}
```

### 2. Custom Error Types

```go
// Dart-specific error types
type DartParseError struct {
    FilePath string
    Line     int
    Column   int
    Message  string
    Code     string
}

func (e DartParseError) Error() string {
    return fmt.Sprintf("dart parse error at %s:%d:%d: %s", 
        e.FilePath, e.Line, e.Column, e.Message)
}

type DartCompilationError struct {
    MainFile  string
    PartFiles []string
    Errors    []error
}

func (e DartCompilationError) Error() string {
    return fmt.Sprintf("dart compilation unit error in %s: %d errors", 
        e.MainFile, len(e.Errors))
}
```

## Testing Strategy

### 1. Unit Tests

```go
// Test-driven development approach
func TestDartBasicSymbolExtraction(t *testing.T) {
    tests := []struct {
        name     string
        dartCode string
        expected []expectedSymbol
    }{
        {
            name: "basic class",
            dartCode: `
                class MyClass {
                    void method() {}
                    int variable = 0;
                }
            `,
            expected: []expectedSymbol{
                {name: "MyClass", symbolType: types.SymbolTypeClass},
                {name: "method", symbolType: types.SymbolTypeMethod},
                {name: "variable", symbolType: types.SymbolTypeVariable},
            },
        },
        {
            name: "flutter widget",
            dartCode: `
                class MyWidget extends StatelessWidget {
                    @override
                    Widget build(BuildContext context) {
                        return Container();
                    }
                }
            `,
            expected: []expectedSymbol{
                {name: "MyWidget", symbolType: SymbolTypeWidget},
                {name: "build", symbolType: SymbolTypeBuildMethod},
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
            manager := NewTestManager()
            symbols := manager.extractDartSymbols(tt.dartCode)
            
            assert.Equal(t, len(tt.expected), len(symbols))
            for i, expected := range tt.expected {
                assert.Equal(t, expected.name, symbols[i].Name)
                assert.Equal(t, expected.symbolType, symbols[i].Type)
            }
        })
    }
}
```

### 2. Integration Tests

```go
// Test with real Flutter projects
func TestFlutterProjectAnalysis(t *testing.T) {
    testCases := []struct {
        name        string
        projectPath string
        expectation projectExpectation
    }{
        {
            name:        "Flutter Gallery",
            projectPath: "testdata/flutter_gallery",
            expectation: projectExpectation{
                minWidgets:  50,
                hasProvider: true,
                hasBloc:     true,
            },
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            analyzer := NewAnalyzer()
            graph, err := analyzer.AnalyzeProject(tc.projectPath)
            
            assert.NoError(t, err)
            assert.GreaterOrEqual(t, countWidgets(graph), tc.expectation.minWidgets)
        })
    }
}
```

### 3. Performance Tests

```go
// Performance benchmarks
func BenchmarkDartParsing(b *testing.B) {
    largeDartFile := generateLargeDartFile(10000) // 10k lines
    manager := NewManager()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := manager.ParseDartContent(largeDartFile)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkFlutterWidgetDetection(b *testing.B) {
    flutterCode := loadFlutterTestCode()
    detector := NewFlutterDetector()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        detector.DetectFramework("test.dart", "dart", flutterCode)
    }
}
```

## Configuration Integration

### 1. Enhanced Language Configuration

```yaml
# Enhanced configuration for Dart support
languages:
  dart:
    extensions: [".dart"]
    parser: "tree-sitter-dart"
    parser_timeout: 30s
    features:
      version: "3.0"
      flutter: true
      null_safety: true
      macros: false
      patterns: true
      records: true
      sealed_classes: true
    compilation_units:
      enabled: true
      follow_parts: true
      cache_timeout: 5m
    performance:
      large_file_threshold: 1048576  # 1MB
      max_parse_time: 30s
      streaming_enabled: true
    exclude_patterns:
      - ".dart_tool/**"
      - "build/**"
      - "*.g.dart"
      - "*.freezed.dart"
      - "*.mocks.dart"
    framework_detection:
      flutter:
        enabled: true
        patterns:
          - "StatelessWidget"
          - "StatefulWidget"
          - "HookWidget"
      state_management:
        enabled: true
        frameworks:
          - "riverpod"
          - "bloc"
          - "provider"
          - "getx"
```

### 2. Runtime Configuration

```go
// Configuration management
type DartConfig struct {
    Version          string            `yaml:"version"`
    Flutter          bool              `yaml:"flutter"`
    NullSafety       bool              `yaml:"null_safety"`
    Macros           bool              `yaml:"macros"`
    Patterns         bool              `yaml:"patterns"`
    Records          bool              `yaml:"records"`
    SealedClasses    bool              `yaml:"sealed_classes"`
    CompilationUnits CompilationConfig `yaml:"compilation_units"`
    Performance      PerformanceConfig `yaml:"performance"`
    ExcludePatterns  []string          `yaml:"exclude_patterns"`
    FrameworkDetection FrameworkConfig `yaml:"framework_detection"`
}

type CompilationConfig struct {
    Enabled      bool          `yaml:"enabled"`
    FollowParts  bool          `yaml:"follow_parts"`
    CacheTimeout time.Duration `yaml:"cache_timeout"`
}

type PerformanceConfig struct {
    LargeFileThreshold int           `yaml:"large_file_threshold"`
    MaxParseTime       time.Duration `yaml:"max_parse_time"`
    StreamingEnabled   bool          `yaml:"streaming_enabled"`
}
```

## Monitoring and Metrics

### 1. Performance Metrics

```go
// Dart parsing metrics
type DartMetrics struct {
    FilesProcessed        int           `json:"files_processed"`
    SymbolsExtracted      int           `json:"symbols_extracted"`
    ParseErrors           int           `json:"parse_errors"`
    AverageParseTime      time.Duration `json:"average_parse_time"`
    FlutterFilesFound     int           `json:"flutter_files_found"`
    WidgetsDetected       int           `json:"widgets_detected"`
    CompilationUnits      int           `json:"compilation_units"`
    PartFilesProcessed    int           `json:"part_files_processed"`
    LargestFileSize       int64         `json:"largest_file_size"`
    CacheHitRate          float64       `json:"cache_hit_rate"`
    FallbackParseCount    int           `json:"fallback_parse_count"`
}

// Collect metrics during parsing
func (m *Manager) collectDartMetrics(filePath string, parseTime time.Duration, symbols int) {
    m.metrics.FilesProcessed++
    m.metrics.SymbolsExtracted += symbols
    m.metrics.AverageParseTime = updateAverage(m.metrics.AverageParseTime, parseTime, m.metrics.FilesProcessed)
    
    // Check file size
    if stat, err := os.Stat(filePath); err == nil {
        if stat.Size() > m.metrics.LargestFileSize {
            m.metrics.LargestFileSize = stat.Size()
        }
    }
}
```

### 2. Health Checks

```go
// Health monitoring for Dart parsing
func (m *Manager) DartHealthCheck() error {
    // Test basic parsing capability
    testCode := `class TestClass { void testMethod() {} }`
    
    start := time.Now()
    ast, err := m.parseDartContent(testCode, "test.dart")
    duration := time.Since(start)
    
    if err != nil {
        return fmt.Errorf("dart parser health check failed: %w", err)
    }
    
    if duration > 100*time.Millisecond {
        return fmt.Errorf("dart parser too slow: %v", duration)
    }
    
    symbols, err := m.ExtractSymbols(ast)
    if err != nil {
        return fmt.Errorf("dart symbol extraction failed: %w", err)
    }
    
    if len(symbols) == 0 {
        return fmt.Errorf("dart symbol extraction returned no symbols")
    }
    
    return nil
}
```

## Migration and Deployment

### 1. Feature Flags

```go
// Feature flag management
type DartFeatureFlags struct {
    Enabled          bool `json:"enabled"`
    FlutterDetection bool `json:"flutter_detection"`
    Dart3Features    bool `json:"dart3_features"`
    PartFiles        bool `json:"part_files"`
    AsyncAnalysis    bool `json:"async_analysis"`
    StreamingParser  bool `json:"streaming_parser"`
}

func (m *Manager) isDartFeatureEnabled(feature string) bool {
    flags := m.config.DartFeatures
    
    switch feature {
    case "flutter":
        return flags.Enabled && flags.FlutterDetection
    case "dart3":
        return flags.Enabled && flags.Dart3Features
    case "parts":
        return flags.Enabled && flags.PartFiles
    case "async":
        return flags.Enabled && flags.AsyncAnalysis
    case "streaming":
        return flags.Enabled && flags.StreamingParser
    default:
        return flags.Enabled
    }
}
```

### 2. Backward Compatibility

```go
// Ensure backward compatibility
func (m *Manager) GetSupportedLanguages() []types.Language {
    languages := m.getSupportedLanguagesV1() // Existing implementation
    
    // Add Dart only if enabled
    if m.isDartFeatureEnabled("") {
        dartLang := types.Language{
            Name:       "dart",
            Extensions: []string{".dart"},
            Parser:     "tree-sitter-dart",
            Enabled:    true,
        }
        languages = append(languages, dartLang)
    }
    
    return languages
}
```

## Conclusion

This Low-Level Design provides a comprehensive, production-ready approach to adding Dart support to CodeContext. The design emphasizes:

1. **Incremental Implementation**: Phased approach from basic syntax to advanced features
2. **Performance**: Optimized parsing with caching and streaming capabilities
3. **Reliability**: Robust error handling with graceful degradation
4. **Maintainability**: Clean architecture following SOLID principles
5. **Testability**: Comprehensive testing strategy with TDD approach
6. **Extensibility**: Flutter framework intelligence and state management detection

The implementation follows established patterns in the codebase while introducing necessary extensions for Dart-specific features. The design supports both current Dart language features and provides a foundation for future enhancements.
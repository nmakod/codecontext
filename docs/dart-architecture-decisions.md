# Dart Support Architecture Decisions

## Overview

This document captures the key architectural decisions made during the design of Dart language support for CodeContext. Each decision includes the context, options considered, rationale, and consequences.

## Decision Log

### ADR-001: Tree-sitter Integration Strategy

**Date**: 2024-08-21  
**Status**: Proposed  
**Context**: Need to choose how to integrate tree-sitter-dart parser with Go codebase

#### Options Considered

**Option A: Direct Integration with Official Bindings**
```go
import dart "github.com/tree-sitter/tree-sitter-dart/bindings/go"
```

**Pros:**
- Consistent with existing language integrations
- Officially maintained
- Minimal complexity

**Cons:**
- May not exist or be outdated
- Limited control over binding quality
- Dependency on external maintenance

**Option B: UserNobody14 Grammar with Custom Go Bindings**
```go
// Generate custom bindings from UserNobody14/tree-sitter-dart
import dart "github.com/nuthan-ms/codecontext/internal/bindings/dart"
```

**Pros:**
- Most up-to-date grammar
- Full control over implementation
- Can optimize for CodeContext needs

**Cons:**
- Maintenance burden
- CGO compilation complexity
- Platform-specific build requirements

**Option C: Runtime Plugin Loading**
```go
// Load pre-compiled parser at runtime
lib, err := plugin.Open("tree-sitter-dart.so")
```

**Pros:**
- No compilation dependencies
- Can distribute pre-built binaries
- Flexible deployment

**Cons:**
- Runtime complexity
- Platform-specific binaries needed
- Plugin loading overhead

**Option D: Hybrid Approach with Fallback**
```go
func ParseDart(content string) (*AST, error) {
    if ast, err := parseTreeSitter(content); err == nil {
        return ast, nil
    }
    return parseRegexFallback(content)
}
```

**Pros:**
- Always works
- Graceful degradation
- Best user experience

**Cons:**
- Dual implementation complexity
- Potential inconsistencies
- Higher maintenance cost

#### Decision

**Selected: Option B (Custom Go Bindings) with Option D (Fallback) principles**

**Rationale:**
1. **Control**: Custom bindings give us full control over quality and updates
2. **Performance**: Can optimize for CodeContext-specific use cases
3. **Reliability**: Fallback ensures parsing always works
4. **Maintenance**: Team has Go expertise to maintain bindings

**Implementation:**
```go
// Primary: Custom tree-sitter bindings
// Fallback: Regex-based parsing for reliability
func (m *Manager) ParseDart(content string) (*types.AST, error) {
    if m.treeSitterEnabled {
        if ast, err := m.parseWithTreeSitter(content); err == nil {
            return ast, nil
        }
        log.Warn("Tree-sitter parsing failed, falling back to regex")
    }
    return m.parseWithRegex(content)
}
```

**Consequences:**
- Need to maintain Go bindings for tree-sitter-dart
- Requires CGO build pipeline  
- Provides reliable parsing with graceful degradation
- Enables optimization for CodeContext use cases

---

### ADR-002: Symbol Extraction Architecture

**Date**: 2024-08-21  
**Status**: Proposed  
**Context**: Choose approach for extracting symbols from Dart AST

#### Options Considered

**Option A: Pattern Matching (Simple Switch Statements)**
```go
func extractDartSymbol(node *ASTNode) *Symbol {
    switch node.Type {
    case "class_declaration":
        return extractClass(node)
    case "function_declaration":
        return extractFunction(node)
    }
}
```

**Pros:**
- Simple to implement and understand
- Fast execution
- Easy to debug

**Cons:**
- Not easily extensible
- Hard-coded logic
- Difficult to add complex patterns

**Option B: Visitor Pattern**
```go
type DartVisitor interface {
    VisitClass(*ASTNode) *Symbol
    VisitFunction(*ASTNode) *Symbol
}

func (v *DartSymbolVisitor) Visit(node *ASTNode) {
    // Dynamic dispatch based on node type
}
```

**Pros:**
- Extensible architecture
- Clean separation of concerns
- Easy to add new symbol types

**Cons:**
- More complex implementation
- Slight performance overhead
- Over-engineering for initial needs

**Option C: Query-Based (Tree-sitter Queries)**
```scheme
(class_declaration
  name: (identifier) @class-name
  body: (class_body) @class-body)
```

**Pros:**
- Leverages tree-sitter's query engine
- Very precise pattern matching
- Declarative syntax

**Cons:**
- Additional learning curve
- Query language dependency
- Less flexibility for custom logic

**Option D: Hybrid Approach**
```go
// Simple patterns for basic symbols
// Visitor pattern for complex analysis
func (m *Manager) extractDartSymbol(node *ASTNode) *Symbol {
    if symbol := m.extractBasicSymbol(node); symbol != nil {
        return symbol
    }
    return m.complexAnalyzer.Visit(node)
}
```

#### Decision

**Selected: Option A (Pattern Matching) for Phase 1, evolve to Option D (Hybrid)**

**Rationale:**
1. **KISS Principle**: Start with simplest approach that works
2. **TDD Compatibility**: Easy to write tests for simple patterns
3. **Performance**: Direct pattern matching is fastest
4. **Evolution Path**: Can refactor to visitor pattern when complexity grows

**Implementation:**
```go
// Phase 1: Simple pattern matching
func (m *Manager) nodeToSymbolDart(node *types.ASTNode, filePath, language string) *types.Symbol {
    switch node.Type {
    case "class_declaration":
        return m.extractDartClass(node, filePath, language)
    case "function_declaration":
        return m.extractDartFunction(node, filePath, language)
    // ... other basic patterns
    }
}

// Phase 2+: Add visitor pattern for complex cases
func (m *Manager) extractComplexDartSymbol(node *types.ASTNode) *types.Symbol {
    return m.dartVisitor.Visit(node)
}
```

**Consequences:**
- Fast initial implementation
- Easy to test and debug
- Will need refactoring as complexity grows
- Provides clear evolution path

---

### ADR-003: Flutter Framework Detection Strategy

**Date**: 2024-08-21  
**Status**: Proposed  
**Context**: Determine how to detect and analyze Flutter-specific patterns

#### Options Considered

**Option A: Import-Based Detection**
```go
func isFlutterFile(imports []string) bool {
    for _, imp := range imports {
        if strings.Contains(imp, "package:flutter/") {
            return true
        }
    }
    return false
}
```

**Pros:**
- Simple and reliable
- Fast detection
- Clear criteria

**Cons:**
- Only detects files with Flutter imports
- Misses indirect Flutter usage
- Binary classification (Flutter or not)

**Option B: Pattern-Based Detection**
```go
var flutterPatterns = map[string]*regexp.Regexp{
    "widget":     regexp.MustCompile(`extends\s+(Stateless|Stateful)Widget`),
    "build":      regexp.MustCompile(`Widget\s+build\s*\(`),
    "state":      regexp.MustCompile(`extends\s+State<`),
}
```

**Pros:**
- Detects Flutter usage patterns
- More comprehensive detection
- Can classify different Flutter constructs

**Cons:**
- More complex implementation
- Potential false positives
- Requires pattern maintenance

**Option C: AST-Based Analysis**
```go
func (d *FlutterDetector) AnalyzeClass(node *ASTNode) FlutterClassInfo {
    info := FlutterClassInfo{}
    
    // Analyze inheritance chain
    if superclass := extractSuperclass(node); superclass != "" {
        info.WidgetType = classifyWidget(superclass)
    }
    
    // Analyze methods
    for _, method := range extractMethods(node) {
        if method.Name == "build" && method.HasOverride {
            info.HasBuildMethod = true
        }
    }
    
    return info
}
```

**Pros:**
- Deep semantic analysis
- Accurate classification
- Rich information extraction

**Cons:**
- Complex implementation
- Performance overhead
- Requires complete AST parsing

**Option D: Multi-Layer Detection**
```go
type FlutterDetector struct {
    importDetector  ImportDetector
    patternDetector PatternDetector
    astAnalyzer     ASTAnalyzer
}

func (d *FlutterDetector) Analyze(file *DartFile) FlutterInfo {
    info := FlutterInfo{}
    
    // Layer 1: Quick import check
    if !d.importDetector.HasFlutterImports(file) {
        return info
    }
    
    // Layer 2: Pattern matching
    patterns := d.patternDetector.FindPatterns(file.Content)
    info.Patterns = patterns
    
    // Layer 3: Deep AST analysis (if needed)
    if requiresDeepAnalysis(patterns) {
        info.DeepAnalysis = d.astAnalyzer.Analyze(file.AST)
    }
    
    return info
}
```

#### Decision

**Selected: Option D (Multi-Layer Detection)**

**Rationale:**
1. **Performance**: Quick rejection of non-Flutter files
2. **Accuracy**: Deep analysis when needed
3. **Extensibility**: Can add new detection layers
4. **Optimization**: Expensive analysis only when necessary

**Implementation:**
```go
// Layer 1: Import-based quick check
func (d *FlutterDetector) IsFlutterFile(imports []string) bool {
    for _, imp := range imports {
        if strings.HasPrefix(imp, "package:flutter/") {
            return true
        }
    }
    return false
}

// Layer 2: Pattern detection
func (d *FlutterDetector) DetectPatterns(content string) map[string]bool {
    patterns := make(map[string]bool)
    patterns["has_widgets"] = d.widgetPattern.MatchString(content)
    patterns["has_state"] = d.statePattern.MatchString(content)
    patterns["has_build"] = d.buildPattern.MatchString(content)
    return patterns
}

// Layer 3: AST analysis (selective)
func (d *FlutterDetector) AnalyzeAST(ast *types.AST) FlutterAnalysis {
    if ast == nil {
        return FlutterAnalysis{}
    }
    
    // Deep widget hierarchy analysis
    // State management pattern detection
    // Lifecycle method identification
}
```

**Consequences:**
- Optimal performance with layered approach
- Comprehensive Flutter detection
- Complex implementation requiring careful testing
- Clear separation of detection strategies

---

### ADR-004: Error Handling and Fallback Strategy

**Date**: 2024-08-21  
**Status**: Proposed  
**Context**: Handle parsing failures and malformed Dart code gracefully

#### Options Considered

**Option A: Fail Fast**
```go
func (p *DartParser) Parse(content string) (*AST, error) {
    ast, err := p.treeSitter.Parse(content)
    if err != nil {
        return nil, fmt.Errorf("dart parsing failed: %w", err)
    }
    return ast, nil
}
```

**Pros:**
- Simple implementation
- Clear error conditions
- No hidden failures

**Cons:**
- Poor user experience
- No partial results
- Analysis stops on any error

**Option B: Silent Fallback**
```go
func (p *DartParser) Parse(content string) (*AST, error) {
    if ast, err := p.treeSitter.Parse(content); err == nil {
        return ast, nil
    }
    
    // Silent fallback to regex parsing
    return p.regexParser.Parse(content)
}
```

**Pros:**
- Always returns results
- Transparent to caller
- Good user experience

**Cons:**
- Hidden failures
- Inconsistent results
- Debugging difficulties

**Option C: Explicit Fallback with Metadata**
```go
type ParseResult struct {
    AST        *types.AST
    ParsedWith string  // "tree-sitter" | "regex" | "partial"
    Errors     []error
    Warnings   []string
}

func (p *DartParser) Parse(content string) (*ParseResult, error) {
    result := &ParseResult{}
    
    // Try tree-sitter first
    if ast, err := p.treeSitter.Parse(content); err == nil {
        result.AST = ast
        result.ParsedWith = "tree-sitter"
        return result, nil
    } else {
        result.Errors = append(result.Errors, err)
        result.Warnings = append(result.Warnings, "Tree-sitter parsing failed")
    }
    
    // Fallback to regex
    if ast, err := p.regexParser.Parse(content); err == nil {
        result.AST = ast
        result.ParsedWith = "regex"
        return result, nil
    }
    
    return nil, fmt.Errorf("all parsing strategies failed")
}
```

**Pros:**
- Transparent error reporting
- Metadata about parsing method
- Comprehensive error information

**Cons:**
- Complex return type
- Changes existing interfaces
- Additional complexity

**Option D: Graceful Degradation with Recovery**
```go
func (p *DartParser) Parse(content string) (*types.AST, error) {
    // Try full parsing first
    if ast, err := p.parseComplete(content); err == nil {
        return ast, nil
    }
    
    // Try partial parsing
    if ast, err := p.parsePartial(content); err == nil {
        ast.Metadata["parse_quality"] = "partial"
        return ast, nil
    }
    
    // Basic structure parsing
    if ast, err := p.parseBasic(content); err == nil {
        ast.Metadata["parse_quality"] = "basic"
        return ast, nil
    }
    
    return nil, fmt.Errorf("parsing failed completely")
}
```

#### Decision

**Selected: Option D (Graceful Degradation) with Option C (Metadata) principles**

**Rationale:**
1. **User Experience**: Always provide useful results when possible
2. **Transparency**: Include metadata about parse quality
3. **Debugging**: Preserve error information for troubleshooting
4. **Compatibility**: Maintain existing interface contracts

**Implementation:**
```go
func (m *Manager) ParseDartFile(filePath string) (*types.AST, error) {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }
    
    ast := &types.AST{
        FilePath: filePath,
        Language: "dart",
        Metadata: make(map[string]interface{}),
    }
    
    // Strategy 1: Tree-sitter parsing
    if tsAST, err := m.parseWithTreeSitter(string(content)); err == nil {
        ast.Root = tsAST.Root
        ast.Metadata["parser"] = "tree-sitter"
        ast.Metadata["parse_quality"] = "complete"
        return ast, nil
    } else {
        ast.Metadata["tree_sitter_error"] = err.Error()
    }
    
    // Strategy 2: Regex-based parsing
    if regexAST, err := m.parseWithRegex(string(content)); err == nil {
        ast.Root = regexAST.Root
        ast.Metadata["parser"] = "regex"
        ast.Metadata["parse_quality"] = "basic"
        ast.Metadata["warning"] = "Fallback parsing used"
        return ast, nil
    }
    
    // Strategy 3: Minimal structure parsing
    ast.Root = m.createMinimalAST(string(content))
    ast.Metadata["parser"] = "minimal"
    ast.Metadata["parse_quality"] = "structural"
    return ast, nil
}
```

**Consequences:**
- Always returns usable results
- Transparent about parsing quality
- Complex implementation requiring thorough testing
- Enables analysis even with malformed code

---

### ADR-005: Performance Optimization Strategy

**Date**: 2024-08-21  
**Status**: Proposed  
**Context**: Ensure Dart parsing performance is comparable to other language parsers

#### Options Considered

**Option A: No Special Optimization**
- Use same approach as other languages
- Rely on tree-sitter performance
- Simple implementation

**Option B: File Size-Based Strategies**
```go
func (p *DartParser) Parse(filePath string) (*AST, error) {
    stat, _ := os.Stat(filePath)
    
    if stat.Size() > 1024*1024 { // 1MB
        return p.parseStreaming(filePath)
    }
    
    return p.parseStandard(filePath)
}
```

**Option C: Content-Based Optimization**
```go
func (p *DartParser) Parse(content string) (*AST, error) {
    if p.isGenerated(content) {
        return p.parseGenerated(content)
    }
    
    if p.hasComplexPatterns(content) {
        return p.parseComplex(content)
    }
    
    return p.parseStandard(content)
}
```

**Option D: Caching and Memoization**
```go
type DartCache struct {
    asts map[string]*CachedAST
    mu   sync.RWMutex
}

type CachedAST struct {
    AST          *types.AST
    Hash         string
    LastModified time.Time
}

func (c *DartCache) Get(filePath string) (*types.AST, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    cached, exists := c.asts[filePath]
    if !exists {
        return nil, false
    }
    
    // Check if file has been modified
    if c.isStale(filePath, cached.LastModified) {
        return nil, false
    }
    
    return cached.AST, true
}
```

#### Decision

**Selected: Combination of Options B, C, and D**

**Rationale:**
1. **Dart Specifics**: Large generated files are common in Dart/Flutter
2. **Performance**: Streaming prevents memory issues
3. **Efficiency**: Caching reduces repeated parsing
4. **Flexibility**: Multiple strategies for different content types

**Implementation:**
```go
type DartPerformanceOptimizer struct {
    cache           *DartCache
    sizeLimitStream int64
    sizeLimitSkip   int64
}

func (p *DartPerformanceOptimizer) ParseOptimized(filePath string) (*types.AST, error) {
    // Check cache first
    if ast, found := p.cache.Get(filePath); found {
        return ast, nil
    }
    
    stat, err := os.Stat(filePath)
    if err != nil {
        return nil, err
    }
    
    // Skip very large files (likely generated)
    if stat.Size() > p.sizeLimitSkip {
        return p.createSkippedAST(filePath, "file_too_large"), nil
    }
    
    // Use streaming for large files
    if stat.Size() > p.sizeLimitStream {
        ast, err := p.parseStreaming(filePath)
        if err == nil {
            p.cache.Set(filePath, ast)
        }
        return ast, err
    }
    
    // Standard parsing
    ast, err := p.parseStandard(filePath)
    if err == nil {
        p.cache.Set(filePath, ast)
    }
    
    return ast, err
}

// Streaming parser for large files
func (p *DartPerformanceOptimizer) parseStreaming(filePath string) (*types.AST, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    
    // Parse in chunks to prevent memory exhaustion
    scanner := bufio.NewScanner(file)
    scanner.Buffer(make([]byte, 64*1024), 1024*1024) // 64KB buffer, 1MB max
    
    ast := &types.AST{
        FilePath: filePath,
        Language: "dart",
    }
    
    // Extract high-level structure without full parsing
    ast.Root = p.extractStructure(scanner)
    
    return ast, nil
}
```

**Consequences:**
- Optimal performance for different file sizes
- Memory-efficient parsing of large generated files
- Cache reduces repeated parsing overhead
- Complex implementation requiring performance testing

---

### ADR-006: Testing Strategy Architecture

**Date**: 2024-08-21  
**Status**: Proposed  
**Context**: Ensure comprehensive testing coverage for Dart support

#### Options Considered

**Option A: Unit Tests Only**
- Test individual functions in isolation
- Fast execution
- Simple to implement

**Option B: Integration Tests Only**
- Test complete workflows
- Real-world validation
- Comprehensive coverage

**Option C: Property-Based Testing**
```go
func TestDartParsingProperties(t *testing.T) {
    quick.Check(func(code string) bool {
        ast, err := parser.Parse(code)
        if err != nil {
            return true // Invalid code is acceptable
        }
        
        // Property: All symbols should have valid locations
        symbols, _ := parser.ExtractSymbols(ast)
        for _, symbol := range symbols {
            if symbol.Location.Line < 1 {
                return false
            }
        }
        return true
    }, nil)
}
```

**Option D: Multi-Layer Testing**
```go
// Layer 1: Unit tests for individual components
func TestDartSymbolExtraction(t *testing.T) { ... }

// Layer 2: Integration tests for complete workflows
func TestDartProjectAnalysis(t *testing.T) { ... }

// Layer 3: Performance benchmarks
func BenchmarkDartParsing(b *testing.B) { ... }

// Layer 4: Real-world validation
func TestFlutterSDKCompatibility(t *testing.T) { ... }
```

#### Decision

**Selected: Option D (Multi-Layer Testing)**

**Rationale:**
1. **Comprehensive Coverage**: Different test types catch different issues
2. **TDD Compatibility**: Unit tests enable test-driven development
3. **Performance Validation**: Benchmarks ensure performance goals
4. **Real-world Validation**: Integration tests with actual projects

**Implementation:**
```go
// test/dart/unit_test.go - Unit tests
func TestDartBasicSymbols(t *testing.T) {
    tests := []struct {
        name     string
        code     string
        expected []Symbol
    }{
        {
            name: "simple class",
            code: "class MyClass {}",
            expected: []Symbol{{Name: "MyClass", Type: ClassSymbol}},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            symbols := extractSymbols(tt.code)
            assert.Equal(t, tt.expected, symbols)
        })
    }
}

// test/dart/integration_test.go - Integration tests  
func TestFlutterProjectAnalysis(t *testing.T) {
    analyzer := NewAnalyzer()
    graph, err := analyzer.AnalyzeProject("testdata/flutter_counter")
    
    assert.NoError(t, err)
    assert.Contains(t, graph.GetSymbols(), "MyApp")
    assert.Contains(t, graph.GetSymbols(), "MyHomePage")
}

// test/dart/benchmark_test.go - Performance tests
func BenchmarkDartParsing(b *testing.B) {
    content := loadLargeDartFile()
    parser := NewDartParser()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := parser.Parse(content)
        if err != nil {
            b.Fatal(err)
        }
    }
}

// test/dart/compatibility_test.go - Real-world validation
func TestFlutterSDKCompatibility(t *testing.T) {
    sdkPath := os.Getenv("FLUTTER_SDK_PATH")
    if sdkPath == "" {
        t.Skip("FLUTTER_SDK_PATH not set")
    }
    
    // Test parsing of Flutter framework files
    frameworkFiles := findDartFiles(filepath.Join(sdkPath, "packages/flutter/lib"))
    
    parser := NewDartParser()
    for _, file := range frameworkFiles {
        t.Run(filepath.Base(file), func(t *testing.T) {
            _, err := parser.ParseFile(file)
            assert.NoError(t, err, "Failed to parse %s", file)
        })
    }
}
```

**Consequences:**
- Comprehensive test coverage at all levels
- Enables confident refactoring and optimization
- Catches issues early in development
- Requires significant testing infrastructure

---

## Implementation Guidelines

### Code Quality Standards

1. **Test Coverage**: Maintain >90% test coverage for all Dart-related code
2. **Performance**: Dart parsing should be within 20% of other language parsers
3. **Error Handling**: All parsing functions must handle malformed input gracefully
4. **Documentation**: Public APIs must have comprehensive documentation
5. **Backwards Compatibility**: Changes must not break existing functionality

### Review Criteria

Each architectural decision should be validated against:

1. **SOLID Principles**: Does it follow good OOP practices?
2. **Performance**: Does it meet performance requirements?  
3. **Maintainability**: Is it easy to understand and modify?
4. **Testability**: Can it be thoroughly tested?
5. **Extensibility**: Does it allow for future enhancements?

### Evolution Process

1. **Weekly Reviews**: Evaluate decisions against actual implementation experience
2. **Performance Monitoring**: Track metrics to validate optimization decisions
3. **User Feedback**: Incorporate feedback from real-world usage
4. **Refactoring**: Plan refactoring when decisions prove suboptimal

## Risk Assessment

### High-Risk Decisions

1. **ADR-001 (Tree-sitter Integration)**: Custom bindings require ongoing maintenance
2. **ADR-004 (Error Handling)**: Complex fallback logic may introduce bugs
3. **ADR-005 (Performance)**: Multiple optimization strategies increase complexity

### Mitigation Strategies

1. **Documentation**: Comprehensive documentation for all architectural decisions
2. **Testing**: Extensive testing to validate decision outcomes
3. **Monitoring**: Performance and error monitoring to detect issues early
4. **Rollback Plans**: Ability to revert decisions if they prove problematic

## Conclusion

These architectural decisions provide a solid foundation for implementing comprehensive Dart support in CodeContext. The decisions prioritize:

1. **Reliability**: Graceful error handling and fallback strategies
2. **Performance**: Optimized parsing for different content types
3. **Maintainability**: Clean architecture following established patterns
4. **Extensibility**: Ability to add new features incrementally
5. **Quality**: Comprehensive testing at all levels

The decisions will be continuously evaluated and refined based on implementation experience and user feedback.
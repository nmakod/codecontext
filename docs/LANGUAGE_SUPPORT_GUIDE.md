# Language Support Development Guide

This document provides a comprehensive guide for adding new language support to CodeContext, based on lessons learned from implementing Dart/Flutter support.

## Table of Contents

- [Overview](#overview)
- [Architecture Principles](#architecture-principles)
- [4-Phase Development Process](#4-phase-development-process)
- [Testing Strategy](#testing-strategy)
- [Common Pitfalls & Solutions](#common-pitfalls--solutions)
- [Future Language Support Checklist](#future-language-support-checklist)
- [Case Study: Dart/Flutter Implementation](#case-study-dartflutter-implementation)
- [Case Study: Swift Implementation](#case-study-swift-implementation)
- [Templates & Examples](#templates--examples)

## Overview

Adding language support to CodeContext involves more than just parsing syntax. It requires understanding language-specific patterns, framework ecosystems, and testing challenges that can arise during CI/CD integration.

**Key Success Factors:**
- Incremental development approach
- Comprehensive testing with realistic expectations
- Understanding of CI environment limitations
- Proper test infrastructure design

## Architecture Principles

CodeContext's modular architecture makes language addition straightforward when following these principles:

### 1. Separation of Concerns
```
Parser Layer    → Tree-sitter integration, AST generation
Analysis Layer  → Symbol extraction, dependency analysis  
Detection Layer → Framework-specific pattern recognition
Testing Layer   → Validation, performance, integration tests
```

### 2. Interface Consistency
All language parsers should implement consistent interfaces:
- `Manager` for parser coordination
- `Detector` for framework-specific analysis
- Standard symbol types and extraction patterns

### 3. Framework Awareness
Modern languages often have dominant frameworks that require special handling:
- **Dart**: Flutter widgets, state management patterns
- **Swift**: SwiftUI views, UIKit patterns, Vapor server-side
- **JavaScript**: React components, Node.js patterns
- **Python**: Django/Flask web frameworks, data science libraries

## 4-Phase Development Process

### Phase 1: Basic Parsing Support
**Goal**: Generate AST from source code

```go
type NewLanguageParser struct {
    manager *Manager
}

func (p *NewLanguageParser) Parse(content string, filename string) (*AST, error) {
    // Basic Tree-sitter integration
    // Error handling for malformed code
    // AST validation
}
```

**Key Tasks:**
- Integrate Tree-sitter grammar
- Handle parsing errors gracefully
- Validate AST structure
- Write basic parsing tests

### Phase 2: Symbol Extraction
**Goal**: Extract meaningful symbols from AST

```go
func (p *NewLanguageParser) ExtractSymbols(ast *AST) ([]*Symbol, error) {
    // Start with fundamental symbols:
    // - Classes/structs/interfaces
    // - Functions/methods
    // - Variables/constants
    // - Imports/modules
}
```

**Key Tasks:**
- Define language-specific symbol types
- Implement symbol extraction logic
- Handle nested scopes and namespaces
- Add symbol extraction tests

### Phase 3: Framework-Specific Analysis
**Goal**: Recognize framework patterns and idioms

```go
type NewLanguageDetector struct {
    patterns map[string]*regexp.Regexp
    keywords []string
}

func (d *NewLanguageDetector) AnalyzeContent(content string) *FrameworkAnalysis {
    // Framework detection
    // Pattern recognition
    // Feature identification
}
```

**Key Tasks:**
- Research dominant frameworks
- Implement pattern detection
- Add framework-specific symbol types
- Create comprehensive framework tests

### Phase 4: Performance Optimization
**Goal**: Ensure production-ready performance

```go
func (p *NewLanguageParser) ExtractSymbolsOptimized(ast *AST) ([]*Symbol, error) {
    // Optimize after basic functionality works
    // Profile and benchmark
    // Cache frequently accessed data
}
```

**Key Tasks:**
- Profile parsing performance
- Optimize bottlenecks
- Add performance benchmarks
- Validate memory usage

## Testing Strategy

### Integration Testing
Test with realistic code samples from actual projects:

```go
func TestNewLanguageIntegration(t *testing.T) {
    t.Run("basic parsing", func(t *testing.T) {
        content := `/* realistic code sample */`
        ast, err := parser.Parse(content, "example.ext")
        require.NoError(t, err)
        assert.NotNil(t, ast)
    })
    
    t.Run("symbol extraction", func(t *testing.T) {
        symbols, err := parser.ExtractSymbols(ast)
        require.NoError(t, err)
        
        // Validate expected symbol counts and types
        assert.GreaterOrEqual(t, countSymbolType(symbols, ClassSymbol), 2)
        assert.GreaterOrEqual(t, countSymbolType(symbols, MethodSymbol), 5)
    })
}
```

### Performance Testing
Set realistic expectations for CI environments:

```go
func TestParsingPerformance(t *testing.T) {
    ciMultiplier := 1
    if os.Getenv("CI") != "" {
        ciMultiplier = 20 // CI environments are significantly slower
    }
    
    maxTime := time.Duration(500*ciMultiplier) * time.Millisecond
    
    start := time.Now()
    _, err := parser.Parse(largeContent, "large_file.ext")
    elapsed := time.Since(start)
    
    require.NoError(t, err)
    assert.Less(t, elapsed, maxTime, "Parsing should complete within reasonable time")
}
```

### Coverage Validation
Aim for 60%+ feature coverage:

```go
func TestFeatureCoverage(t *testing.T) {
    testCases := []struct {
        name        string
        content     string
        expectedFeatures []string
    }{
        // Test cases covering major language features
    }
    
    totalFeatures := 0
    detectedFeatures := 0
    
    for _, tc := range testCases {
        analysis := detector.Analyze(tc.content)
        for _, expected := range tc.expectedFeatures {
            totalFeatures++
            if containsFeature(analysis, expected) {
                detectedFeatures++
            }
        }
    }
    
    coverage := float64(detectedFeatures) / float64(totalFeatures) * 100
    assert.GreaterOrEqual(t, coverage, 60.0, "Feature coverage should be ≥60%")
}
```

## Common Pitfalls & Solutions

### 1. Test Infrastructure Conflicts

**❌ Problem:**
```go
// This can interfere with test infrastructure
func TestServerRun(t *testing.T) {
    server.Run(ctx) // May close stdout, breaking coverage reports
}
```

**✅ Solution:**
```go
// Test server logic without affecting stdio
func TestServerLifecycle(t *testing.T) {
    server := NewServer(config)
    assert.NotNil(t, server.analyzer)
    assert.NotPanics(t, func() { server.Stop() })
}
```

### 2. Unrealistic Performance Expectations

**❌ Problem:**
```go
// Too optimistic for CI environments
assert.Less(t, parseTime.Milliseconds(), int64(500))
```

**✅ Solution:**
```go
// Account for CI environment limitations
expectedTime := int64(500)
if os.Getenv("CI") != "" {
    expectedTime *= 20 // CI can be 10-20x slower
}
assert.Less(t, parseTime.Milliseconds(), expectedTime)
```

### 3. Race Conditions in Concurrent Tests

**❌ Problem:**
```go
// Flaky timing-dependent logic
select {
case <-done:
    // Success
case <-time.After(5*time.Second):
    t.Error("Timeout") // May fail in slow CI
}
```

**✅ Solution:**
```go
// Robust timeout handling with debug info
timeout := 30 * time.Second // Generous for CI
debugMode := testing.Verbose()

done := make(chan error, 1)
go func() {
    if debugMode {
        log.Printf("[TEST] Starting operation...")
    }
    done <- operation()
}()

select {
case err := <-done:
    if debugMode {
        log.Printf("[TEST] Operation completed: %v", err)
    }
    // Handle both success and expected failures
case <-time.After(timeout):
    t.Error("Operation timed out - check CI logs for details")
}
```

### 4. Complex Language-Specific Requirements

**Challenge**: Languages often have unique features that require special handling.

**Examples**:
- **Dart**: Mixins, enhanced enums, null safety operators
- **Rust**: Ownership system, lifetime parameters, macros
- **TypeScript**: Type annotations, decorators, ambient declarations

**Solution**: Research thoroughly and implement incrementally:

```go
// Start with core language features
func extractBasicSymbols(ast *AST) []*Symbol {
    // Classes, functions, variables
}

// Add language-specific features gradually
func extractAdvancedSymbols(ast *AST) []*Symbol {
    // Mixins, operators, specialized syntax
}
```

## Future Language Support Checklist

### Pre-Development Phase
- [ ] Research language syntax and common patterns
- [ ] Identify dominant frameworks and libraries
- [ ] Verify Tree-sitter grammar availability and quality
- [ ] Review existing similar language implementations
- [ ] Define scope and success criteria

### Development Phase
- [ ] **Phase 1**: Implement basic parsing with error handling
- [ ] **Phase 2**: Add core symbol extraction (classes, functions, variables)
- [ ] **Phase 3**: Implement framework-specific detection
- [ ] **Phase 4**: Optimize performance and memory usage
- [ ] Add comprehensive documentation

### Testing Phase
- [ ] Write integration tests with realistic code samples
- [ ] Set CI-appropriate performance expectations (multiply by 10-20x for CI)
- [ ] Add comprehensive error handling tests
- [ ] Validate feature coverage ≥60%
- [ ] Test framework detection accuracy

### CI/CD Integration
- [ ] Ensure tests don't interfere with test infrastructure
- [ ] Use generous timeouts for CI environments
- [ ] Add debug logging for complex operations
- [ ] Test across different platforms if needed
- [ ] Monitor memory usage and performance in CI

## Case Study: Dart/Flutter Implementation

### What Worked Well
1. **Modular Architecture**: Easy to extend existing parser framework
2. **Tree-sitter Integration**: Robust parsing with good error recovery
3. **Framework-Specific Analysis**: Flutter widget and state management detection
4. **Comprehensive Testing**: Integration tests with realistic Flutter apps

### Major Challenges Faced
1. **Test Infrastructure Issue**: MCP server tests closing stdout, interfering with coverage reports
2. **CI Performance Assumptions**: Tests failing due to 20x slower CI environments
3. **Race Conditions**: Concurrent tests with timing dependencies
4. **Flutter Complexity**: Deep widget trees, multiple state management patterns

### Key Fixes Applied
```go
// Fixed: Coverage file descriptor issue
func TestServerLifecycle(t *testing.T) {
    // Test without starting stdio transport
    server := NewServer(config)
    assert.NotPanics(t, func() { server.Stop() })
}

// Fixed: CI performance expectations
maxTime := 500 * time.Millisecond
if os.Getenv("CI") != "" {
    maxTime = 10 * time.Second // 20x slower in CI
}

// Fixed: Race condition handling
timeout := 30 * time.Second // Generous timeout
debugMode := testing.Verbose()
// ... robust error handling with debug info
```

### Lessons Learned
- **Start Simple**: Begin with basic parsing before adding framework detection
- **Test Incrementally**: Validate each phase before moving to the next
- **Plan for CI**: Always account for slower CI environments
- **Debug Thoroughly**: Add comprehensive logging for complex operations
- **Document Everything**: Capture decisions and trade-offs for future reference

## Templates & Examples

### Basic Language Parser Template
```go
package parser

type NewLanguageParser struct {
    manager *Manager
    config  *Config
}

func NewNewLanguageParser(manager *Manager) *NewLanguageParser {
    return &NewLanguageParser{
        manager: manager,
        config:  &Config{
            Language: "newlang",
            Extensions: []string{".nl", ".newlang"},
        },
    }
}

func (p *NewLanguageParser) Parse(content string, filename string) (*AST, error) {
    // Implementation
}

func (p *NewLanguageParser) ExtractSymbols(ast *AST) ([]*Symbol, error) {
    // Implementation  
}
```

### Framework Detector Template
```go
package parser

type NewLanguageDetector struct {
    frameworks map[string]FrameworkPattern
    keywords   []string
}

type FrameworkPattern struct {
    ImportPatterns []string
    Keywords       []string
    ClassPatterns  []string
}

func (d *NewLanguageDetector) DetectFrameworks(content string) []string {
    // Implementation
}

func (d *NewLanguageDetector) AnalyzeContent(content string) *FrameworkAnalysis {
    // Implementation
}
```

### Test Structure Template
```go
func TestNewLanguageSupport(t *testing.T) {
    manager := NewManager()
    
    t.Run("basic_parsing", func(t *testing.T) {
        // Test AST generation
    })
    
    t.Run("symbol_extraction", func(t *testing.T) {
        // Test symbol identification
    })
    
    t.Run("framework_detection", func(t *testing.T) {
        // Test framework-specific patterns
    })
    
    t.Run("performance_validation", func(t *testing.T) {
        // Test with CI-appropriate timeouts
    })
    
    t.Run("error_handling", func(t *testing.T) {
        // Test malformed code handling
    })
}
```

---

## Conclusion

Adding language support to CodeContext requires careful planning, incremental development, and thorough testing. The key to success is learning from previous implementations, setting realistic expectations, and building robust test infrastructure.

By following this guide, future language additions should be faster, more reliable, and less prone to the pitfalls we encountered during Dart/Flutter and Swift implementations.

## Case Study: C++ Implementation

### Overview
**Implementation Date:** August 2025 (v3.1.1)  
**Approach:** Tree-sitter integration with security-first design  
**Result:** Production-ready with comprehensive security hardening and 100% test coverage

### Key Engineering Decisions

**1. Security-First Architecture**
- **Decision:** Implement comprehensive input sanitization and resource management
- **Rationale:** C++ codebases often handle sensitive/system-level code requiring enhanced security
- **Implementation:** Multi-layer validation against null bytes, path traversal, DoS attacks

**2. Configuration-Driven Design**
- **Strategy:** YAML/JSON configurable limits for operational flexibility
- **Benefits:** Production environments can tune parsing limits without code changes
- **Features:** Configurable nesting depth, template complexity, file size limits, timeout enforcement

**3. Enhanced Error Handling**
- **Approach:** Custom error types with proper wrapping/unwrapping
- **Categories:** Initialization, parsing, validation, AST processing
- **Integration:** Structured logging with error categorization for debugging

### Technical Implementation Highlights

**C++ Specific Challenges:**
1. **Resource Management:** Tree-sitter memory cleanup requirements
2. **Template Complexity:** Deep template nesting can cause performance issues
3. **Security Concerns:** C++ codebases often contain system-level code
4. **Framework Diversity:** Qt, STL, Boost, Unreal Engine, OpenCV support needed

**Solutions:**
```go
// Resource safety with defer patterns
defer func() {
    if tree != nil {
        tree.Close()
        cp.logger.Debug("tree-sitter resources cleaned up", ...)
    }
}()

// Input sanitization
func validateFilePath(filePath string) error {
    if strings.Contains(filePath, "\x00") {
        return fmt.Errorf("file path contains null bytes")
    }
    // Path traversal, length validation...
}

// Configuration-driven limits
config.Cpp.MaxNestingDepth = 100
config.Cpp.MaxTemplateDepth = 20
config.Cpp.ParseTimeout = 30 * time.Second
config.Cpp.StrictTimeoutEnforcement = false // Lenient by default
```

### Comprehensive Testing Strategy

**Test Coverage Achieved:**
- **Edge Cases:** 26+ security-focused test scenarios
- **Error Handling:** Nil safety, resource cleanup, timeout validation
- **Integration:** End-to-end parsing with realistic C++ codebases
- **Security:** Path traversal, null injection, DoS prevention tests

**Advanced Test Categories:**
```go
// Security validation tests
func TestCppInputSanitization(t *testing.T) {
    testCases := []struct {
        name          string
        filePath      string
        expectError   bool
        errorContains string
    }{
        {"null byte injection", "file\x00.cpp", true, "null bytes"},
        {"path traversal", "../../../sensitive.cpp", true, "path traversal"},
        {"extremely long path", strings.Repeat("a", 5000) + ".cpp", true, "too long"},
    }
}

// Performance and configuration tests
func TestCppConfigurationEdgeCases(t *testing.T) {
    // Nil safety, extreme values, timeout enforcement modes
}
```

### Production Readiness Assessment

**Security Hardening:**
- ✅ Input sanitization against injection attacks
- ✅ Resource leak prevention with proper cleanup
- ✅ DoS protection with configurable limits
- ✅ Path traversal prevention

**Engineering Excellence:**
- ✅ Comprehensive error handling with custom types
- ✅ Structured logging integration
- ✅ Configuration-driven architecture
- ✅ Method extraction for maintainability

**Operational Readiness:**
- ✅ Configurable timeout enforcement (strict/lenient modes)
- ✅ Performance monitoring with metrics
- ✅ Graceful degradation patterns
- ✅ Backward compatibility maintained

### Performance Results

**Parsing Performance:**
- Tree-sitter parsing: Native C performance via CGO
- Resource overhead: Minimal with proper cleanup
- Memory safety: Zero leaks with defer patterns
- Timeout handling: Configurable with operational flexibility

**Enterprise Features:**
- **Modern C++ Support:** C++11-20 features including auto, lambdas, concepts
- **Framework Integration:** Detection points for Qt, STL, Boost, Unreal Engine
- **Template Analysis:** Deep template parsing with configurable depth limits
- **Security Focus:** Production-grade input validation and resource management

### Lessons Learned for Future Implementations

**1. Security Must Be Built-In**
- ✅ **Start with threat modeling:** Identify security risks specific to the language/domain
- ✅ **Input validation first:** Implement comprehensive sanitization before parsing logic
- ✅ **Resource management:** Plan cleanup patterns for any native dependencies

**2. Configuration-Driven Design Pays Off**
- ✅ **Operational flexibility:** Production environments need tunable limits
- ✅ **Testing benefits:** Easy to test extreme configurations
- ✅ **Maintenance:** No code changes needed for operational adjustments

**3. Comprehensive Testing Prevents Issues**
- ✅ **Edge case focus:** Security-focused testing catches real vulnerabilities
- ✅ **Error path testing:** Validate all error conditions and recovery
- ✅ **Integration validation:** End-to-end testing with realistic codebases

**4. Tree-sitter vs Regex Trade-offs**
- ✅ **Tree-sitter advantages:** Native performance, battle-tested grammars, robust error recovery
- ⚠️ **CGO considerations:** Platform-specific builds, compilation complexity
- **Recommendation:** Tree-sitter preferred for languages with mature grammars

### Development Timeline
- **Architecture & Security Design:** 1 day
- **Core parser implementation:** 2 days  
- **Security hardening:** 1 day
- **Comprehensive testing:** 2 days
- **Code review & refinement:** 1 day
- **Total:** 7 days for production-ready C++ support

## Case Study: Swift Implementation

### Overview
**Implementation Date:** August 2025 (v3.0.1)  
**Approach:** Regex-based parsing with TDD methodology  
**Result:** 90% P1/P2 feature coverage with comprehensive framework support

### Swift Implementation Details

**1. Parsing Strategy Choice**
- **Decision:** Use regex-based parsing instead of Tree-sitter
- **Rationale:** Swift Tree-sitter bindings had compatibility issues
- **Implementation:** 24 sophisticated regex patterns covering Swift syntax

**2. TDD Approach**
- **Strategy:** Test-driven development with incremental feature addition
- **Coverage:** 81.8% core features + 90% P1/P2 advanced features
- **Validation:** Comprehensive test suite with realistic Swift/SwiftUI code

**3. Framework Priorities**
- **Primary:** SwiftUI, UIKit, Vapor (established frameworks)
- **Secondary:** SwiftData, TCA, Swift Testing (modern frameworks)
- **Detection:** Multi-strategy approach with priority-based selection

### Technical Implementation

**Swift-Specific Challenges:**
1. **Complex Property Wrappers:** `@AppStorage("key", store: .standard)`
2. **Result Builders:** `@ViewBuilder`, `@resultBuilder` patterns
3. **Async/Await Patterns:** Modern concurrency with actors
4. **Swift 5.9+ Macros:** `#Preview`, `@freestanding`, `@attached`

**Solutions:**
```go
// Enhanced property wrapper detection
"complexPropertyWrapper": regexp.MustCompile(`(?m)^\s+(@\w+\([^)]*\))\s+(?:(?:public|private|internal|fileprivate)\s+)?(?:var|let)\s+(\w+)`)

// Result builder patterns
"resultBuilder": regexp.MustCompile(`(?m)@resultBuilder\s+(?:struct|class|enum)\s+(\w+)`)

// Swift 5.9+ macro system
"macroDecl": regexp.MustCompile(`(?ms)@(?:freestanding|attached)\s*\([^)]+\)\s*macro\s+(\w+)`)
```

### Performance Results

**Parsing Performance:**
- File parsing: <1ms per Swift file
- Symbol extraction: 15-20 symbols per typical SwiftUI view
- Framework detection: Instant with caching
- Memory overhead: <2MB additional

**Feature Coverage:**
- **Core Features:** Classes, structs, protocols, enums, functions (81.8%)
- **P1 Features:** Async/await, actors, property wrappers (90%)
- **P2 Features:** Result builders, macros, advanced patterns (90%)
- **Framework Detection:** 6 major Swift frameworks supported

### Lessons Learned

**1. Regex vs Tree-sitter Trade-offs**
- ✅ **Advantages:** Fast development, no CGO dependencies, full control
- ⚠️ **Limitations:** Manual pattern maintenance, potential edge cases
- **Recommendation:** Regex viable for languages with regular syntax patterns

**2. TDD Benefits**
- ✅ **Quality:** High test coverage prevented regressions
- ✅ **Confidence:** Systematic validation of each feature
- ✅ **Documentation:** Tests serve as feature specifications

**3. Framework Ecosystem Complexity**
- **Challenge:** Swift's rapidly evolving framework landscape
- **Solution:** Priority-based detection with extensible patterns
- **Future:** Regular updates needed for new Swift/framework versions

### Success Metrics

**Development Timeline:**
- Planning & Design: 1 day
- Core implementation: 2 days
- Advanced features: 2 days
- Testing & validation: 1 day
- **Total:** 6 days for comprehensive Swift support

**Quality Metrics:**
- Zero regressions across existing test suite
- 100% backward compatibility
- Production-ready performance characteristics
- Comprehensive documentation and examples

### Recommendations for Future Languages

**1. Choose Parsing Strategy Based on Requirements**
- **Tree-sitter (Recommended):** For languages with mature grammars, when security/performance critical
- **Regex-based:** For rapid prototyping, when Tree-sitter bindings unavailable
- **Hybrid approach:** Tree-sitter for core parsing + regex for framework detection

**2. Security-First for System Languages**
- **C/C++/Rust:** Implement comprehensive input sanitization and resource management
- **Java/C#:** Focus on deserialization and reflection safety
- **Python/Node.js:** Handle dynamic imports and eval-like constructs safely

**3. Configuration-Driven Architecture**
- **Operational needs:** Production environments require tunable limits
- **Testing benefits:** Easy to validate extreme configurations
- **Maintenance:** Avoid hardcoded magic numbers

**4. Comprehensive Testing Strategy**
- **Edge cases first:** Start with security validation and error conditions  
- **Realistic samples:** Use actual open-source projects for integration tests
- **CI awareness:** Account for 10-20x slower CI environments

**5. Document Security Considerations**
- **Threat model:** Document specific risks for each language
- **Mitigation strategies:** Explain security controls implemented
- **Configuration guide:** Help operators tune security vs performance

### Implementation References

For working examples and detailed patterns:
- **C++ Implementation:** `/internal/parser/cpp_parser.go` - Security-first Tree-sitter integration
- **Swift Implementation:** `/internal/parser/swift.go` - Regex-based with TDD methodology  
- **Dart Implementation:** `/internal/parser/dart.go` - Framework-aware parsing with Flutter support

Each implementation demonstrates different approaches and trade-offs suitable for different language characteristics and requirements.
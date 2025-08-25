package parser

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
	
	"github.com/nuthan-ms/codecontext/pkg/types"
)

// Dart language patterns for regex-based parsing (fallback approach)
var dartPatterns = map[string]*regexp.Regexp{
	// Class patterns - updated to support Dart 3.0+ modifiers
	"class":      regexp.MustCompile(`(?m)^(?:(?:sealed|final|base|interface|mixin)\s+)?(?:abstract\s+)?class\s+(\w+)(?:<[\w\s,<>]+>)?(?:\s+extends\s+[\w<>]+)?(?:\s+with\s+[\w\s,<>]+)?(?:\s+implements\s+[\w\s,<>]+)?\s*{`),
	"sealedClass": regexp.MustCompile(`(?m)^sealed\s+class\s+(\w+)(?:<[\w\s,<>]+>)?(?:\s+extends\s+[\w<>]+)?(?:\s+with\s+[\w\s,<>]+)?(?:\s+implements\s+[\w\s,<>]+)?\s*{`),
	"finalClass": regexp.MustCompile(`(?m)^final\s+class\s+(\w+)(?:<[\w\s,<>]+>)?(?:\s+extends\s+[\w<>]+)?(?:\s+with\s+[\w\s,<>]+)?(?:\s+implements\s+[\w\s,<>]+)?\s*{`),
	"baseClass":  regexp.MustCompile(`(?m)^base\s+class\s+(\w+)(?:<[\w\s,<>]+>)?(?:\s+extends\s+[\w<>]+)?(?:\s+with\s+[\w\s,<>]+)?(?:\s+implements\s+[\w\s,<>]+)?\s*{`),
	"interfaceClass": regexp.MustCompile(`(?m)^interface\s+class\s+(\w+)(?:<[\w\s,<>]+>)?(?:\s+extends\s+[\w<>]+)?(?:\s+with\s+[\w\s,<>]+)?(?:\s+implements\s+[\w\s,<>]+)?\s*{`),
	"mixinClassModifier": regexp.MustCompile(`(?m)^mixin\s+class\s+(\w+)(?:<[\w\s,<>]+>)?(?:\s+extends\s+[\w<>]+)?(?:\s+with\s+[\w\s,<>]+)?(?:\s+implements\s+[\w\s,<>]+)?\s*{`),
	"stateClass": regexp.MustCompile(`(?m)^(?:abstract\s+)?class\s+(\w+)\s+extends\s+State<[\w<>]+>`),
	"mixinClass": regexp.MustCompile(`(?m)^(?:abstract\s+)?class\s+(\w+)(?:\s+extends\s+[\w<>]+)?\s+with\s+([\w\s,<>]+)(?:\s+implements\s+[\w\s,<>]+)?\s*{`),
	
	// Mixin and extension patterns
	"mixin":      regexp.MustCompile(`(?m)^mixin\s+(\w+(?:<[\w\s,<>]+>)?)(?:\s+on\s+[\w\s,<>]+)?\s*{`),
	"extension":  regexp.MustCompile(`(?m)^extension\s+(\w*(?:<[\w,\s]+>)?)\s*on\s+([\w<>\[\],\s]+)\s*{`),
	
	// Enum patterns - enhanced for Dart 2.17+
	"enum":       regexp.MustCompile(`(?m)^enum\s+(\w+)(?:<[\w\s,<>]+>)?(?:\s+implements\s+[\w\s,<>]+)?(?:\s+with\s+[\w\s,<>]+)?\s*{`),
	"enumValue":  regexp.MustCompile(`(?m)^\s*(\w+)(?:\([^)]*\))?(?:\s*,|\s*;|\s*})`),
	
	// Type patterns - updated for records
	"typedef":    regexp.MustCompile(`(?m)^typedef\s+(\w+)(?:<[\w\s,<>]+>)?\s*=\s*([\w<>\[\],\s\(\){}]+);`),
	"recordType": regexp.MustCompile(`(?m)^\([\w\s,<>?]+(?:,\s*[\w\s,<>?]+)*\)`),
	"namedRecord": regexp.MustCompile(`(?m)^\({[\w\s,<>?:]+}\)`),
	
	// Function and method patterns - updated for records
	"function":   regexp.MustCompile(`(?m)^[\w<>\[\],\s\(\){}]*?\b(\w+)\s*\([^)]*\)\s*(?:async\s*)?\s*(?:\{|=>)`),
	"method":     regexp.MustCompile(`(?m)^\s+(?:@override\s+)?(?:static\s+)?[\w<>\[\],\s\(\){}]*?\b(\w+)\s*\([^)]*\)\s*(?:async\s*)?\s*(?:\{|=>)`),
	"privateMethod": regexp.MustCompile(`(?m)^\s+[\w<>\[\],\s]*?\b(_\w+)\s*\([^)]*\)\s*(?:async\s*)?\s*(?:\{|=>)`),
	
	// Variable patterns - updated for records and late
	"variable":   regexp.MustCompile(`(?m)^\s*(?:late\s+)?(?:final\s+|const\s+|var\s+|static\s+)?(?:[\w<>\[\],\s?\(\){}]+\s+)?(\w+)\s*=`),
	"lateVariable": regexp.MustCompile(`(?m)^\s*late\s+(?:final\s+)?(?:[\w<>\[\],\s?]+\s+)?(\w+)(?:\s*=|\s*;)`),
	
	// Import and part patterns
	"import":     regexp.MustCompile(`(?m)^\s*import\s+['"]([^'"]+)['"](?:\s+as\s+\w+)?;`),
	"buildMethod": regexp.MustCompile(`(?m)^\s+(?:@override\s+)?Widget\s+build\s*\(\s*BuildContext\s+\w+\s*\)`),
	"lifecycleMethod": regexp.MustCompile(`(?m)^\s+@override\s+void\s+(initState|dispose|didUpdateWidget|didChangeDependencies)\s*\(`),
	"partDirective":   regexp.MustCompile(`(?m)^part\s+['"]([^'"]+)['"];`),
	"partOfDirective": regexp.MustCompile(`(?m)^part\s+of\s+(?:['"]([^'"]+)['"]|(\w+(?:\.\w+)*));`),
	
	// Pattern matching patterns
	"switchExpression": regexp.MustCompile(`(?m)switch\s*\([^)]+\)\s*{`),
	"switchExpressionNew": regexp.MustCompile(`(?m)=>\s*switch\s*\([^)]+\)\s*{`),
	"patternCase": regexp.MustCompile(`(?m)case\s+[\w\s\(\),<>{}:]+(?:when\s+[^:]+)?:`),
	
	// Async patterns - Week 6 additions
	"asyncGenerator": regexp.MustCompile(`(?m)^\s*Stream<[\w\s<>,]+>\s+(\w+)\s*\([^)]*\)\s*async\s*\*\s*{`),
	"asyncMethod": regexp.MustCompile(`(?m)^\s+(?:Future<[\w\s<>,]+>\s+)?(\w+)\s*\([^)]*\)\s+async\s*{`),
	"asyncFunction": regexp.MustCompile(`(?m)^(?:Future<[\w\s<>,]+>\s+)?(\w+)\s*\([^)]*\)\s+async\s*{`),
	"yieldKeyword": regexp.MustCompile(`(?m)\byield\s+`),
	"awaitKeyword": regexp.MustCompile(`(?m)\bawait\s+`),
	"streamController": regexp.MustCompile(`(?m)StreamController<[\w\s<>,]+>\s+(\w+)`),
	"futureBuilder": regexp.MustCompile(`(?m)FutureBuilder<[\w\s<>,]+>\s*\(`),
	"streamBuilder": regexp.MustCompile(`(?m)StreamBuilder<[\w\s<>,]+>\s*\(`),
	"streamSubscription": regexp.MustCompile(`(?m)StreamSubscription<[\w\s<>,]+>\s+(\w+)`),
	
	// Error handling patterns
	"tryBlock": regexp.MustCompile(`(?m)\btry\s*{`),
	"catchBlock": regexp.MustCompile(`(?m)\bcatch\s*\([^)]+\)\s*{`),
	"finallyBlock": regexp.MustCompile(`(?m)\bfinally\s*{`),
	"throwStatement": regexp.MustCompile(`(?m)\bthrow\s+`),
	"rethrowStatement": regexp.MustCompile(`(?m)\brethrow\s*;`),
	
	// Functional programming patterns
	"higherOrderFunction": regexp.MustCompile(`(?m)^\s*(?:static\s+)?[\w<>\[\],\s\(\){}]*?\b(map|filter|reduce|compose|curry|memoize|pipe|asyncMap|asyncFilter)\s*(?:<[^>]*>)?\s*\(`),
	"callbackFunction": regexp.MustCompile(`(?m)Function\s*\([^)]*\)\s+(\w+)`),
	"closureFactory": regexp.MustCompile(`(?m)^\s*(?:static\s+)?(?:[\w\s<>\(\)]*)?Function(?:\(\))?\s+(create\w+)\s*\(`),
	"functionTypedef": regexp.MustCompile(`(?m)^typedef\s+(\w+)\s*=\s*[\w\s<>\[\],\(\){}?]+Function\s*\([^)]*\)`),
}

// Flutter-specific patterns
var flutterPatterns = map[string]*regexp.Regexp{
	"flutterImport":    regexp.MustCompile(`package:flutter/`),
	"statelessWidget": regexp.MustCompile(`extends\s+StatelessWidget`),
	"statefulWidget":  regexp.MustCompile(`extends\s+StatefulWidget`),
	"stateClass":      regexp.MustCompile(`extends\s+State<`),
	"overrideAnnotation": regexp.MustCompile(`@override`),
}

// parseDartContent parses Dart content using regex-based approach
// This is our fallback implementation that will be replaced with tree-sitter when available
func (m *Manager) parseDartContent(content, filePath string) (*types.AST, error) {
	return m.parseDartContentWithContext(context.Background(), content, filePath)
}

// parseDartContentWithContext parses Dart content with context for better error reporting
func (m *Manager) parseDartContentWithContext(ctx context.Context, content, filePath string) (*types.AST, error) {
	// Calculate content hash for caching
	contentHash := calculateHash(content)
	cacheKey := filePath
	version := "1.0"
	
	// Check cache first for performance optimization
	if cachedAST, err := m.cache.Get(cacheKey, version); err == nil {
		if cachedAST.Hash == contentHash {
			// Cache hit - return cached AST
			return cachedAST.AST, nil
		} else {
			// Content changed - invalidate old cache entry
			m.cache.Invalidate(cacheKey)
		}
	}
	
	ast := &types.AST{
		Language:  "dart",
		Content:   content,
		FilePath:  filePath,
		Hash:      contentHash,
		Version:   version,
		ParsedAt:  time.Now(),
	}
	
	// Enhanced Flutter analysis with proper error handling
	flutterDetector := NewFlutterDetector()
	flutterAnalysis := m.safeAnalyzeFlutter(flutterDetector, content)
	
	// Create root AST node with parse metadata
	parseMetadata := map[string]any{
		"parser":         "regex", // Will be "tree-sitter" when we have real bindings
		"parse_quality":  "basic",
		"has_flutter":    flutterAnalysis.IsFlutter,
		"has_errors":     false,
		"error_count":    0,
	}
	
	// Extract nodes with proper error handling
	nodes := m.safeExtractDartNodes(content, cacheKey)
	
	ast.Root = &types.ASTNode{
		Id:    "root",
		Type:  "compilation_unit",
		Value: content,
		Location: types.FileLocation{
			FilePath:  filePath,
			Line:      1,
			Column:    1,
			EndLine:   len(strings.Split(content, "\n")),
			EndColumn: 1,
		},
		Children: nodes,
		Metadata: parseMetadata,
	}
	
	// Integrate Flutter analysis with AST (with error handling)
	if err := m.safeIntegrateFlutterAnalysis(ast, flutterAnalysis); err != nil {
		// Don't fail the entire parse for Flutter integration errors
		ast.Root.Metadata["flutter_integration_error"] = err.Error()
	}
	
	// Cache the parsed AST for future use (with error handling)
	versionedAST := &types.VersionedAST{
		AST:     ast,
		Version: version,
		Hash:    contentHash,
	}
	if err := m.cache.Set(cacheKey, versionedAST); err != nil {
		// Don't fail parse for caching errors, just log
		ast.Root.Metadata["cache_error"] = err.Error()
	}
	
	return ast, nil
}

// safeAnalyzeFlutter safely analyzes Flutter content with proper panic recovery
func (m *Manager) safeAnalyzeFlutter(detector *FlutterDetector, content string) *FlutterAnalysis {
	defer func() {
		if r := recover(); r != nil {
			// Proper structured logging instead of fmt.Printf
			panicErr := NewPanicError("analyze_flutter", "", "dart", r)
			m.logger.Error("Flutter analysis panic recovered", panicErr,
				LogField{Key: "operation", Value: "analyze_flutter"},
				LogField{Key: "content_length", Value: len(content)},
			)
		}
	}()
	
	if analysis := detector.AnalyzeFlutterContent(content); analysis != nil {
		return analysis
	}
	
	// Return safe fallback
	return &FlutterAnalysis{
		IsFlutter: false,
		Framework: "unknown",
	}
}

// safeExtractDartNodes safely extracts Dart nodes with proper panic recovery
func (m *Manager) safeExtractDartNodes(content, cacheKey string) []*types.ASTNode {
	defer func() {
		if r := recover(); r != nil {
			// Proper structured logging and cleanup on panic
			panicErr := NewPanicError("extract_dart_nodes", "", "dart", r)
			m.logger.Error("Dart node extraction panic recovered", panicErr,
				LogField{Key: "operation", Value: "extract_dart_nodes"},
				LogField{Key: "cache_key", Value: cacheKey},
				LogField{Key: "content_length", Value: len(content)},
			)
			
			// Clean up any partial cache entries on panic
			m.cache.Invalidate(cacheKey)
		}
	}()
	
	return m.extractDartNodes(content)
}

// safeIntegrateFlutterAnalysis safely integrates Flutter analysis with error recovery
func (m *Manager) safeIntegrateFlutterAnalysis(ast *types.AST, analysis *FlutterAnalysis) error {
	defer func() {
		if r := recover(); r != nil {
			// Proper structured logging instead of fmt.Printf
			panicErr := NewPanicError("integrate_flutter_analysis", ast.FilePath, "dart", r)
			m.logger.Error("Flutter integration panic recovered", panicErr,
				LogField{Key: "operation", Value: "integrate_flutter_analysis"},
				LogField{Key: "file_path", Value: ast.FilePath},
				LogField{Key: "is_flutter", Value: analysis.IsFlutter},
			)
		}
	}()
	
	m.IntegrateFlutterAnalysis(ast, analysis)
	return nil
}

// extractDartNodes extracts AST nodes from Dart content using optimized regex patterns
func (m *Manager) extractDartNodes(content string) []*types.ASTNode {
	nodes, _ := m.extractDartNodesWithError(content)
	return nodes
}

// extractDartNodesWithError extracts AST nodes and returns any errors encountered
func (m *Manager) extractDartNodesWithError(content string) ([]*types.ASTNode, error) {
	// Validate input
	if len(content) == 0 {
		return nil, nil // Empty content is not an error
	}
	
	if len(content) > MaxFileSize {
		return nil, NewParseError("extract_nodes", "", "dart", 
			fmt.Errorf("file too large: %d bytes (max: %d)", len(content), MaxFileSize))
	}
	
	// Strategy selection based on file size
	strategy := m.selectExtractionStrategy(len(content))
	return strategy.extractNodesWithError(content)
}

// ExtractionStrategy defines different parsing strategies
type DartExtractionStrategy struct {
	manager   *Manager
	threshold int
	name      string
}

// selectExtractionStrategy selects appropriate extraction strategy based on content size
func (m *Manager) selectExtractionStrategy(contentSize int) *DartExtractionStrategy {
	if contentSize > StreamingThresholdBytes {
		return &DartExtractionStrategy{
			manager:   m,
			threshold: StreamingThresholdBytes,
			name:      "streaming",
		}
	}
	
	if contentSize > LimitedThresholdBytes {
		return &DartExtractionStrategy{
			manager:   m,
			threshold: LimitedThresholdBytes,
			name:      "limited",
		}
	}
	
	return &DartExtractionStrategy{
		manager:   m,
		threshold: 0,
		name:      "full",
	}
}

// extractNodes extracts nodes using the appropriate strategy
func (s *DartExtractionStrategy) extractNodes(content string) []*types.ASTNode {
	nodes, _ := s.extractNodesWithError(content)
	return nodes
}

// extractNodesWithError extracts nodes using the appropriate strategy and returns errors
func (s *DartExtractionStrategy) extractNodesWithError(content string) ([]*types.ASTNode, error) {
	lines := strings.Split(content, "\n")
	
	switch s.name {
	case "streaming":
		return s.manager.extractDartNodesStreamingWithError(content, lines)
	case "limited":
		return s.manager.extractDartNodesLimitedWithError(content, lines)
	default:
		return s.manager.extractDartNodesFullWithError(content, lines)
	}
}

// extractDartNodesFull performs full extraction for smaller files
func (m *Manager) extractDartNodesFull(content string, lines []string) []*types.ASTNode {
	nodes, _ := m.extractDartNodesFullWithError(content, lines)
	return nodes
}

// extractDartNodesFullWithError performs full extraction for smaller files with error handling
func (m *Manager) extractDartNodesFullWithError(content string, lines []string) ([]*types.ASTNode, error) {
	defer func() {
		if r := recover(); r != nil {
			panicErr := NewPanicError("extract_dart_nodes_full", "", "dart", r)
			m.logger.Error("Full Dart extraction panic recovered", panicErr,
				LogField{Key: "operation", Value: "extract_dart_nodes_full"},
				LogField{Key: "content_length", Value: len(content)},
				LogField{Key: "lines_count", Value: len(lines)},
			)
		}
	}()
	
	extractor := &DartNodeExtractor{
		manager: m,
		content: content,
		lines:   lines,
		nodes:   make([]*types.ASTNode, 0),
	}
	
	// Extract all types of nodes with error recovery
	try := func(operation string, extractFunc func()) error {
		defer func() {
			if r := recover(); r != nil {
				panicErr := NewPanicError(operation, "", "dart", r)
				m.logger.Error("Node extraction step failed", panicErr,
					LogField{Key: "extraction_step", Value: operation},
				)
			}
		}()
		extractFunc()
		return nil
	}
	
	// Extract each type with individual error recovery
	try("extract_imports", extractor.extractImports)
	try("extract_classes", extractor.extractClasses)
	try("extract_mixins", extractor.extractMixins)
	try("extract_extensions", extractor.extractExtensions)
	try("extract_enums", extractor.extractEnums)
	try("extract_functions", extractor.extractFunctions)
	try("extract_variables", extractor.extractVariables)
	try("extract_part_directives", extractor.extractPartDirectives)
	
	return extractor.nodes, nil
}

// DartNodeExtractor handles the extraction of different node types
type DartNodeExtractor struct {
	manager *Manager
	content string
	lines   []string
	nodes   []*types.ASTNode
}

// extractImports extracts import statements
func (e *DartNodeExtractor) extractImports() {
	if matches := dartPatterns["import"].FindAllStringSubmatch(e.content, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				lineNum := e.manager.findLineNumber(e.content, match[0])
				e.nodes = append(e.nodes, &types.ASTNode{
					Id:   fmt.Sprintf("import-%d", lineNum),
					Type: "import_statement",
					Value: match[0],
					Location: types.FileLocation{
						Line:      lineNum,
						Column:    1,
						EndLine:   lineNum,
						EndColumn: len(match[0]) + 1,
					},
					Children: []*types.ASTNode{
						{
							Id:    fmt.Sprintf("import-path-%d", lineNum),
							Type:  "string_literal",
							Value: match[1],
						},
					},
				})
			}
		}
	}
}

// extractClasses extracts class declarations
func (e *DartNodeExtractor) extractClasses() {
	if matches := dartPatterns["class"].FindAllStringSubmatch(e.content, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				lineNum := e.manager.findLineNumber(e.content, match[0])
				
				// Check if this is a State class
				isStateClass := dartPatterns["stateClass"].MatchString(match[0])
				classType := "class_declaration"
				if isStateClass {
					classType = "state_class_declaration"
				}
				
				// Check for class modifiers (Dart 3.0+)
				var classModifier string
				if dartPatterns["sealedClass"].MatchString(match[0]) {
					classModifier = "sealed"
				} else if dartPatterns["finalClass"].MatchString(match[0]) {
					classModifier = "final"
				} else if dartPatterns["baseClass"].MatchString(match[0]) {
					classModifier = "base"
				} else if dartPatterns["interfaceClass"].MatchString(match[0]) {
					classModifier = "interface"
				} else if dartPatterns["mixinClassModifier"].MatchString(match[0]) {
					classModifier = "mixin"
				}
				
				classNode := &types.ASTNode{
					Id:   fmt.Sprintf("class-%s-%d", match[1], lineNum),
					Type: classType,
					Value: match[0],
					Location: types.FileLocation{
						Line:    lineNum,
						Column:  1,
						EndLine: lineNum,
						EndColumn: len(match[0]) + 1,
					},
					Metadata: map[string]any{},
					Children: []*types.ASTNode{
						{
							Id:    fmt.Sprintf("class-name-%s", match[1]),
							Type:  "identifier",
							Value: match[1],
						},
					},
				}
				
				// Add class modifier to metadata if present
				if classModifier != "" {
					classNode.Metadata["modifier"] = classModifier
				}
				
				// Extract methods within the class
				classContent := e.manager.extractClassContent(e.content, match[0], lineNum)
				classNode.Children = append(classNode.Children, e.manager.extractClassMethods(classContent, lineNum, match[1])...)
				
				e.nodes = append(e.nodes, classNode)
			}
		}
	}
}

// extractMixins extracts mixin declarations
func (e *DartNodeExtractor) extractMixins() {
	if matches := dartPatterns["mixin"].FindAllStringSubmatch(e.content, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				lineNum := e.manager.findLineNumber(e.content, match[0])
				mixinName := match[1]
				
				// For generic mixins like "FormMixin<T extends StatefulWidget>", extract just the base name
				if strings.Contains(mixinName, "<") {
					mixinName = strings.Split(mixinName, "<")[0]
				}
				
				mixinNode := &types.ASTNode{
					Id:   fmt.Sprintf("mixin-%s-%d", mixinName, lineNum),
					Type: "mixin_declaration",
					Value: match[0],
					Location: types.FileLocation{
						Line:    lineNum,
						Column:  1,
						EndLine: lineNum,
						EndColumn: len(match[0]) + 1,
					},
					Children: []*types.ASTNode{
						{
							Id:    fmt.Sprintf("mixin-name-%s", mixinName),
							Type:  "identifier",
							Value: mixinName,
						},
					},
				}
				
				// Extract methods within the mixin
				mixinContent := e.manager.extractClassContent(e.content, match[0], lineNum)
				mixinNode.Children = append(mixinNode.Children, e.manager.extractClassMethods(mixinContent, lineNum, mixinName)...)
				
				e.nodes = append(e.nodes, mixinNode)
			}
		}
	}
}

// extractExtensions extracts extension declarations
func (e *DartNodeExtractor) extractExtensions() {
	if matches := dartPatterns["extension"].FindAllStringSubmatch(e.content, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 2 {
				lineNum := e.manager.findLineNumber(e.content, match[0])
				extensionName := match[1]
				if extensionName == "" {
					// Unnamed extension, generate a name
					extensionName = fmt.Sprintf("Extension%d", lineNum)
				} else {
					// For generic extensions like "ListExtensions<T>", extract just the base name
					if strings.Contains(extensionName, "<") {
						extensionName = strings.Split(extensionName, "<")[0]
					}
				}
				
				extensionNode := &types.ASTNode{
					Id:   fmt.Sprintf("extension-%s-%d", extensionName, lineNum),
					Type: "extension_declaration",
					Value: match[0],
					Location: types.FileLocation{
						Line:    lineNum,
						Column:  1,
						EndLine: lineNum,
						EndColumn: len(match[0]) + 1,
					},
					Children: []*types.ASTNode{
						{
							Id:    fmt.Sprintf("extension-name-%s", extensionName),
							Type:  "identifier",
							Value: extensionName,
						},
						{
							Id:    fmt.Sprintf("extension-target-%s", match[2]),
							Type:  "type_identifier",
							Value: match[2],
						},
					},
				}
				
				// Extract methods within the extension
				extensionContent := e.manager.extractClassContent(e.content, match[0], lineNum)
				extensionNode.Children = append(extensionNode.Children, e.manager.extractClassMethods(extensionContent, lineNum, extensionName)...)
				
				e.nodes = append(e.nodes, extensionNode)
			}
		}
	}
}

// extractEnums extracts enum declarations
func (e *DartNodeExtractor) extractEnums() {
	if matches := dartPatterns["enum"].FindAllStringSubmatch(e.content, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				lineNum := e.manager.findLineNumber(e.content, match[0])
				enumName := match[1]
				
				// For generic enums like "Result<T>", extract just the base name
				if strings.Contains(enumName, "<") {
					enumName = strings.Split(enumName, "<")[0]
				}
				
				enumNode := &types.ASTNode{
					Id:   fmt.Sprintf("enum-%s-%d", enumName, lineNum),
					Type: "enum_declaration",
					Value: match[0],
					Location: types.FileLocation{
						Line:    lineNum,
						Column:  1,
						EndLine: lineNum,
						EndColumn: len(match[0]) + 1,
					},
					Children: []*types.ASTNode{
						{
							Id:    fmt.Sprintf("enum-name-%s", enumName),
							Type:  "identifier",
							Value: enumName,
						},
					},
				}
				
				// Extract enum values
				enumContent := e.manager.extractClassContent(e.content, match[0], lineNum)
				enumNode.Children = append(enumNode.Children, e.manager.extractEnumValues(enumContent, lineNum)...)
				
				e.nodes = append(e.nodes, enumNode)
			}
		}
	}
}

// extractFunctions extracts function declarations
func (e *DartNodeExtractor) extractFunctions() {
	if matches := dartPatterns["function"].FindAllStringSubmatch(e.content, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				lineNum := e.manager.findLineNumber(e.content, match[0])
				// Skip if this is inside a class (crude check)
				if !e.manager.isInsideClass(e.lines, lineNum-1) {
					e.nodes = append(e.nodes, &types.ASTNode{
						Id:   fmt.Sprintf("function-%s-%d", match[1], lineNum),
						Type: "function_declaration",
						Value: match[0],
						Location: types.FileLocation{
							Line:    lineNum,
							Column:  1,
							EndLine: lineNum,
							EndColumn: len(match[0]) + 1,
						},
						Children: []*types.ASTNode{
							{
								Id:    fmt.Sprintf("function-name-%s", match[1]),
								Type:  "identifier",
								Value: match[1],
							},
						},
					})
				}
			}
		}
	}
}

// extractVariables extracts variable declarations
func (e *DartNodeExtractor) extractVariables() {
	if matches := dartPatterns["variable"].FindAllStringSubmatch(e.content, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				lineNum := e.manager.findLineNumber(e.content, match[0])
				// Skip if this is inside a class or function (crude check)
				if !e.manager.isInsideClass(e.lines, lineNum-1) && !e.manager.isInsideFunction(e.lines, lineNum-1) {
					e.nodes = append(e.nodes, &types.ASTNode{
						Id:   fmt.Sprintf("variable-%s-%d", match[1], lineNum),
						Type: "variable_declaration",
						Value: match[0],
						Location: types.FileLocation{
							Line:    lineNum,
							Column:  1,
							EndLine: lineNum,
							EndColumn: len(match[0]) + 1,
						},
						Children: []*types.ASTNode{
							{
								Id:    fmt.Sprintf("variable-name-%s", match[1]),
								Type:  "identifier",
								Value: match[1],
							},
						},
					})
				}
			}
		}
	}
}

// extractPartDirectives extracts part directives
func (e *DartNodeExtractor) extractPartDirectives() {
	// Extract part directives
	if matches := dartPatterns["partDirective"].FindAllStringSubmatch(e.content, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				lineNum := e.manager.findLineNumber(e.content, match[0])
				partFile := match[1]
				
				e.nodes = append(e.nodes, &types.ASTNode{
					Id:   fmt.Sprintf("part-%s-%d", partFile, lineNum),
					Type: "part_directive",
					Value: match[0],
					Location: types.FileLocation{
						Line:      lineNum,
						Column:    1,
						EndLine:   lineNum,
						EndColumn: len(match[0]) + 1,
					},
					Children: []*types.ASTNode{
						{
							Id:    fmt.Sprintf("part-file-%s", partFile),
							Type:  "string_literal",
							Value: partFile,
						},
					},
				})
			}
		}
	}
	
	// Extract part of directives
	if matches := dartPatterns["partOfDirective"].FindAllStringSubmatch(e.content, -1); matches != nil {
		for _, match := range matches {
			lineNum := e.manager.findLineNumber(e.content, match[0])
			var partOfTarget string
			
			// Check if it's a file path (match[1]) or library name (match[2])
			if len(match) > 1 && match[1] != "" {
				partOfTarget = match[1] // File path
			} else if len(match) > 2 && match[2] != "" {
				partOfTarget = match[2] // Library name
			}
			
			if partOfTarget != "" {
				e.nodes = append(e.nodes, &types.ASTNode{
					Id:   fmt.Sprintf("part-of-%s-%d", partOfTarget, lineNum),
					Type: "part_of_directive",
					Value: match[0],
					Location: types.FileLocation{
						Line:      lineNum,
						Column:    1,
						EndLine:   lineNum,
						EndColumn: len(match[0]) + 1,
					},
					Children: []*types.ASTNode{
						{
							Id:    fmt.Sprintf("part-of-target-%s", partOfTarget),
							Type:  "identifier",
							Value: partOfTarget,
						},
					},
				})
			}
		}
	}
}

// extractDartNodesLimited processes medium-sized files with limited extraction for performance
func (m *Manager) extractDartNodesLimited(content string, lines []string) []*types.ASTNode {
	nodes, _ := m.extractDartNodesLimitedWithError(content, lines)
	return nodes
}

// extractDartNodesLimitedWithError processes medium-sized files with limited extraction and error handling
func (m *Manager) extractDartNodesLimitedWithError(content string, lines []string) ([]*types.ASTNode, error) {
	defer func() {
		if r := recover(); r != nil {
			panicErr := NewPanicError("extract_dart_nodes_limited", "", "dart", r)
			m.logger.Error("Limited Dart extraction panic recovered", panicErr,
				LogField{Key: "operation", Value: "extract_dart_nodes_limited"},
				LogField{Key: "content_length", Value: len(content)},
				LogField{Key: "lines_count", Value: len(lines)},
			)
		}
	}()
	
	var nodes []*types.ASTNode
	
	// Performance optimization: limit the number of patterns we process
	const maxSymbols = 5000
	symbolCount := 0
	
	// Priority patterns - only process the most important ones for medium files
	priorityExtractions := map[string]int{
		"import":        50,  // Limit imports
		"class":         1000, // Limit classes  
		"function":      500,  // Limit functions
		"mixin":         100,  // Limit mixins
		"extension":     100,  // Limit extensions
		"enum":          100,  // Limit enums
		"typedef":       100,  // Limit typedefs
		"asyncGenerator": 100, // Limit async generators
		"asyncFunction": 200,  // Limit async functions
	}
	
	for patternName, limit := range priorityExtractions {
		if symbolCount >= maxSymbols {
			break
		}
		
		pattern, exists := dartPatterns[patternName]
		if !exists {
			continue
		}
		
		// Safely extract matches with error recovery
		var matches [][]string
		func() {
			defer func() {
				if r := recover(); r != nil {
					panicErr := NewPanicError("pattern_matching", "", "dart", r)
					m.logger.Error("Pattern matching panic recovered", panicErr,
						LogField{Key: "pattern_name", Value: patternName},
						LogField{Key: "limit", Value: limit},
					)
					matches = nil
				}
			}()
			matches = pattern.FindAllStringSubmatch(content, limit) // Limit matches
		}()
		
		if matches == nil {
			continue
		}
		
		for _, match := range matches {
			if symbolCount >= maxSymbols {
				break
			}
			
			if len(match) > 1 {
				// Safely extract node information
				func() {
					defer func() {
						if r := recover(); r != nil {
							panicErr := NewPanicError("node_creation", "", "dart", r)
							m.logger.Error("Node creation panic recovered", panicErr,
								LogField{Key: "pattern_name", Value: patternName},
								LogField{Key: "match_text", Value: match[0]},
							)
						}
					}()
					
					name := match[1]
					lineNum := m.findLineNumber(content, match[0])
					nodeType := m.getNodeTypeForPattern(patternName)
					
					node := &types.ASTNode{
						Id:   fmt.Sprintf("%s-%s-%d", patternName, name, lineNum),
						Type: nodeType,
						Value: match[0],
						Location: types.FileLocation{
							Line:    lineNum,
							Column:  1,
							EndLine: lineNum,
							EndColumn: len(match[0]) + 1,
						},
						Children: []*types.ASTNode{
							{
								Id:    fmt.Sprintf("%s-name-%s", patternName, name),
								Type:  "identifier",
								Value: name,
							},
						},
					}
					
					// Add metadata for special patterns
					if patternName == "asyncGenerator" || patternName == "asyncFunction" {
						node.Metadata = map[string]any{
							"async_type": strings.TrimPrefix(patternName, "async"),
						}
					}
					
					nodes = append(nodes, node)
					symbolCount++
				}()
			}
		}
	}
	
	return nodes, nil
}

// extractDartNodesStreaming processes large files in chunks for better performance
func (m *Manager) extractDartNodesStreaming(content string, lines []string) []*types.ASTNode {
	nodes, _ := m.extractDartNodesStreamingWithError(content, lines)
	return nodes
}

// extractDartNodesStreamingWithError processes large files in chunks with error handling
func (m *Manager) extractDartNodesStreamingWithError(content string, lines []string) ([]*types.ASTNode, error) {
	defer func() {
		if r := recover(); r != nil {
			panicErr := NewPanicError("extract_dart_nodes_streaming", "", "dart", r)
			m.logger.Error("Streaming Dart extraction panic recovered", panicErr,
				LogField{Key: "operation", Value: "extract_dart_nodes_streaming"},
				LogField{Key: "content_length", Value: len(content)},
				LogField{Key: "lines_count", Value: len(lines)},
			)
		}
	}()
	
	var nodes []*types.ASTNode
	
	// Performance optimization: process in chunks to reduce memory pressure
	const chunkSize = 100 * 1024 // 100KB chunks
	contentLen := len(content)
	
	// For very large files, limit the number of symbols we extract to prevent excessive processing
	const maxSymbols = 10000
	symbolCount := 0
	
	for offset := 0; offset < contentLen && symbolCount < maxSymbols; offset += chunkSize {
		end := offset + chunkSize
		if end > contentLen {
			end = contentLen
		}
		
		chunk := content[offset:end]
		
		// Ensure we don't break in the middle of a class or function
		// Find the last complete construct in this chunk
		if end < contentLen {
			lastBrace := strings.LastIndex(chunk, "}")
			if lastBrace > 0 && lastBrace < len(chunk)-1 {
				chunk = chunk[:lastBrace+1]
				end = offset + lastBrace + 1
			}
		}
		
		// Extract patterns from this chunk with error recovery - focus on the most important ones first
		var chunkNodes []*types.ASTNode
		func() {
			defer func() {
				if r := recover(); r != nil {
					panicErr := NewPanicError("chunk_extraction", "", "dart", r)
					m.logger.Error("Chunk extraction panic recovered", panicErr,
						LogField{Key: "chunk_offset", Value: offset},
						LogField{Key: "chunk_size", Value: len(chunk)},
					)
					chunkNodes = nil
				}
			}()
			chunkNodes = m.extractDartNodesFromChunk(chunk, offset)
		}()
		
		if chunkNodes != nil {
			nodes = append(nodes, chunkNodes...)
			symbolCount += len(chunkNodes)
		}
		
		// Performance optimization: if we found enough symbols, stop processing
		if symbolCount >= maxSymbols {
			m.logger.Debug("Streaming extraction reached symbol limit",
				LogField{Key: "symbols_extracted", Value: symbolCount},
				LogField{Key: "max_symbols", Value: maxSymbols},
			)
			break
		}
		
		// Adjust offset to avoid duplicates
		if end < contentLen {
			offset = end - 1
		}
	}
	
	return nodes, nil
}

// extractDartNodesFromChunk extracts nodes from a content chunk with offset adjustment
func (m *Manager) extractDartNodesFromChunk(chunk string, baseOffset int) []*types.ASTNode {
	var nodes []*types.ASTNode
	
	// Priority patterns - extract most important constructs first
	priorityPatterns := []string{
		"class", "mixin", "extension", "enum", 
		"function", "typedef", "import",
		"asyncGenerator", "asyncFunction",
	}
	
	for _, patternName := range priorityPatterns {
		pattern, exists := dartPatterns[patternName]
		if !exists {
			continue
		}
		
		matches := pattern.FindAllStringSubmatchIndex(chunk, -1)
		if matches == nil {
			continue
		}
		
		for _, match := range matches {
			if len(match) >= 4 { // Ensure we have start/end positions and at least one capture group
				matchText := chunk[match[0]:match[1]]
				capturedName := chunk[match[2]:match[3]]
				
				// Calculate actual line number considering base offset
				lineNum := m.findLineNumberInChunk(chunk[:match[0]]) + m.findLineNumber(chunk[:match[0]], "")
				
				// Create appropriate node type
				nodeType := m.getNodeTypeForPattern(patternName)
				nodeId := fmt.Sprintf("%s-%s-%d", patternName, capturedName, baseOffset+match[0])
				
				node := &types.ASTNode{
					Id:   nodeId,
					Type: nodeType,
					Value: matchText,
					Location: types.FileLocation{
						Line:    lineNum,
						Column:  1,
						EndLine: lineNum,
						EndColumn: len(matchText) + 1,
					},
					Children: []*types.ASTNode{
						{
							Id:    fmt.Sprintf("%s-name-%s", patternName, capturedName),
							Type:  "identifier",
							Value: capturedName,
						},
					},
				}
				
				// Add metadata for async patterns
				if patternName == "asyncGenerator" || patternName == "asyncFunction" {
					node.Metadata = map[string]any{
						"async_type": strings.TrimPrefix(patternName, "async"),
					}
				}
				
				nodes = append(nodes, node)
			}
		}
	}
	
	return nodes
}

// findLineNumberInChunk finds line number within a chunk
func (m *Manager) findLineNumberInChunk(content string) int {
	return strings.Count(content, "\n") + 1
}

// getNodeTypeForPattern maps pattern names to AST node types
func (m *Manager) getNodeTypeForPattern(patternName string) string {
	switch patternName {
	case "class":
		return "class_declaration"
	case "mixin":
		return "mixin_declaration"
	case "extension":
		return "extension_declaration"
	case "enum":
		return "enum_declaration"
	case "typedef":
		return "typedef_declaration"
	case "function":
		return "function_declaration"
	case "import":
		return "import_statement"
	case "asyncGenerator":
		return "async_generator"
	case "asyncFunction":
		return "async_function"
	default:
		return "unknown_declaration"
	}
}

// extractClassMethods extracts methods from within a class
func (m *Manager) extractClassMethods(classContent string, startLine int, className string) []*types.ASTNode {
	var methods []*types.ASTNode
	
	// Safety check for empty class content
	if classContent == "" {
		return methods
	}
	
	// Extract regular methods
	if matches := dartPatterns["method"].FindAllStringSubmatch(classContent, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				methodName := match[1]
				
				// Skip constructors (methods with same name as class)
				if methodName == className {
					continue
				}
				
				// Skip control flow keywords that might be matched
				controlFlowKeywords := []string{"if", "else", "for", "while", "do", "switch", "case", "break", "continue", "return", "throw", "try", "catch", "finally"}
				isControlFlow := false
				for _, keyword := range controlFlowKeywords {
					if methodName == keyword {
						isControlFlow = true
						break
					}
				}
				if isControlFlow {
					continue
				}
				
				// Skip if this looks like a class name used in pattern matching
				if len(methodName) > 0 && methodName[0] >= 'A' && methodName[0] <= 'Z' && strings.Contains(match[0], methodName+"(") {
					continue
				}
				
				methodLineNum := startLine + m.findLineNumber(classContent, match[0]) - 1
				methodType := "method_declaration"
				
				// Check if this is a build method
				if methodName == "build" {
					if buildPattern, exists := dartPatterns["buildMethod"]; exists && buildPattern != nil {
						if buildPattern.MatchString(match[0]) {
							methodType = "build_method"
						}
					}
				}
				
				methods = append(methods, &types.ASTNode{
					Id:   fmt.Sprintf("method-%s-%d", methodName, methodLineNum),
					Type: methodType,
					Value: match[0],
					Location: types.FileLocation{
						Line:    methodLineNum,
						Column:  1,
						EndLine: methodLineNum,
						EndColumn: len(match[0]) + 1,
					},
					Children: []*types.ASTNode{
						{
							Id:    fmt.Sprintf("method-name-%s", methodName),
							Type:  "identifier",
							Value: methodName,
						},
					},
				})
			}
		}
	}
	
	// Extract lifecycle methods specifically
	if matches := dartPatterns["lifecycleMethod"].FindAllStringSubmatch(classContent, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				methodLineNum := startLine + m.findLineNumber(classContent, match[0]) - 1
				
				methods = append(methods, &types.ASTNode{
					Id:   fmt.Sprintf("lifecycle-%s-%d", match[1], methodLineNum),
					Type: "lifecycle_method",
					Value: match[0],
					Location: types.FileLocation{
						Line:    methodLineNum,
						Column:  1,
						EndLine: methodLineNum,
						EndColumn: len(match[0]) + 1,
					},
					Children: []*types.ASTNode{
						{
							Id:    fmt.Sprintf("lifecycle-name-%s", match[1]),
							Type:  "identifier",
							Value: match[1],
						},
					},
				})
			}
		}
	}
	
	// Extract async methods - Week 6 async patterns
	if matches := dartPatterns["asyncMethod"].FindAllStringSubmatch(classContent, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				methodName := match[1]
				methodLineNum := startLine + m.findLineNumber(classContent, match[0]) - 1
				
				// Skip constructors (methods with same name as class)
				if methodName == className {
					continue
				}
				
				methods = append(methods, &types.ASTNode{
					Id:   fmt.Sprintf("async-method-%s-%d", methodName, methodLineNum),
					Type: "async_method",
					Value: match[0],
					Location: types.FileLocation{
						Line:    methodLineNum,
						Column:  1,
						EndLine: methodLineNum,
						EndColumn: len(match[0]) + 1,
					},
					Metadata: map[string]any{
						"async_type": "method",
						"return_type": "Future",
					},
					Children: []*types.ASTNode{
						{
							Id:    fmt.Sprintf("async-method-name-%s", methodName),
							Type:  "identifier",
							Value: methodName,
						},
					},
				})
			}
		}
	}
	
	// Extract higher-order methods - Week 6 functional patterns
	if matches := dartPatterns["higherOrderFunction"].FindAllStringSubmatch(classContent, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				methodName := match[1]
				methodLineNum := startLine + m.findLineNumber(classContent, match[0]) - 1
				
				methods = append(methods, &types.ASTNode{
					Id:   fmt.Sprintf("higher-order-method-%s-%d", methodName, methodLineNum),
					Type: "higher_order_method",
					Value: match[0],
					Location: types.FileLocation{
						Line:    methodLineNum,
						Column:  1,
						EndLine: methodLineNum,
						EndColumn: len(match[0]) + 1,
					},
					Metadata: map[string]any{
						"functional_type": "higher_order",
						"pattern_name": methodName,
					},
					Children: []*types.ASTNode{
						{
							Id:    fmt.Sprintf("higher-order-method-name-%s", methodName),
							Type:  "identifier",
							Value: methodName,
						},
					},
				})
			}
		}
	}
	
	// Extract class member variables
	if matches := dartPatterns["variable"].FindAllStringSubmatch(classContent, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				variableLineNum := startLine + m.findLineNumber(classContent, match[0]) - 1
				
				methods = append(methods, &types.ASTNode{
					Id:   fmt.Sprintf("class-variable-%s-%d", match[1], variableLineNum),
					Type: "variable_declaration",
					Value: match[0],
					Location: types.FileLocation{
						Line:    variableLineNum,
						Column:  1,
						EndLine: variableLineNum,
						EndColumn: len(match[0]) + 1,
					},
					Children: []*types.ASTNode{
						{
							Id:    fmt.Sprintf("class-variable-name-%s", match[1]),
							Type:  "identifier",
							Value: match[1],
						},
					},
				})
			}
		}
	}
	
	return methods
}

// extractEnumValues extracts enum values from within an enum declaration
func (m *Manager) extractEnumValues(enumContent string, startLine int) []*types.ASTNode {
	var enumValues []*types.ASTNode
	
	// Safety check for empty enum content
	if enumContent == "" {
		return enumValues
	}
	
	// Extract enum values using the enumValue pattern
	if matches := dartPatterns["enumValue"].FindAllStringSubmatch(enumContent, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				valueName := match[1]
				valueLineNum := startLine + m.findLineNumber(enumContent, match[0]) - 1
				
				enumValues = append(enumValues, &types.ASTNode{
					Id:   fmt.Sprintf("enum-value-%s-%d", valueName, valueLineNum),
					Type: "enum_value",
					Value: match[0],
					Location: types.FileLocation{
						Line:    valueLineNum,
						Column:  1,
						EndLine: valueLineNum,
						EndColumn: len(match[0]) + 1,
					},
					Children: []*types.ASTNode{
						{
							Id:    fmt.Sprintf("enum-value-name-%s", valueName),
							Type:  "identifier",
							Value: valueName,
						},
					},
				})
			}
		}
	}
	
	return enumValues
}

// Helper methods for parsing
func (m *Manager) findLineNumber(content, pattern string) int {
	if content == "" || pattern == "" {
		return 1
	}
	index := strings.Index(content, pattern)
	if index == -1 {
		return 1
	}
	lines := strings.Split(content[:index], "\n")
	return len(lines)
}

func (m *Manager) extractClassContent(content, classDeclaration string, startLine int) string {
	// Simple extraction of class body - this is a crude implementation
	// In a real tree-sitter implementation, this would be much more accurate
	classIndex := strings.Index(content, classDeclaration)
	if classIndex == -1 {
		return ""
	}
	
	remaining := content[classIndex:]
	braceIndex := strings.Index(remaining, "{")
	if braceIndex == -1 {
		return ""
	}
	
	// Find matching closing brace (simplified)
	braceCount := 1
	start := classIndex + braceIndex + 1
	for i := start; i < len(content) && braceCount > 0; i++ {
		if content[i] == '{' {
			braceCount++
		} else if content[i] == '}' {
			braceCount--
		}
		if braceCount == 0 {
			return content[start:i]
		}
	}
	
	return content[start:]
}

func (m *Manager) isInsideClass(lines []string, lineIndex int) bool {
	if lineIndex < 0 || lineIndex >= len(lines) || len(lines) == 0 {
		return false
	}
	
	// More accurate check: count braces to determine if we're inside a class
	braceCount := 0
	classFound := false
	
	// Look backwards from current line
	for i := 0; i <= lineIndex && i < len(lines); i++ {
		line := lines[i]
		// Check if this line has a class declaration
		if dartPatterns["class"].MatchString(line) || 
		   dartPatterns["mixin"].MatchString(line) ||
		   dartPatterns["extension"].MatchString(line) ||
		   dartPatterns["enum"].MatchString(line) {
			classFound = true
		}
		
		// Count braces
		for _, ch := range line {
			if ch == '{' {
				braceCount++
			} else if ch == '}' {
				braceCount--
			}
		}
	}
	
	// We're inside a class if we found a class and have unclosed braces
	return classFound && braceCount > 0
}

func (m *Manager) isInsideFunction(lines []string, lineIndex int) bool {
	// Similar crude check - this would be much better with real AST parsing
	return m.isInsideClass(lines, lineIndex)
}

// detectFlutterInContent checks if content contains Flutter imports or patterns
// Deprecated: Use FlutterDetector.AnalyzeFlutterContent for comprehensive analysis
func (m *Manager) detectFlutterInContent(content string) bool {
	return flutterPatterns["flutterImport"].MatchString(content)
}

// nodeToSymbolDart converts Dart AST nodes to symbols
func (m *Manager) nodeToSymbolDart(node *types.ASTNode, filePath, language string) *types.Symbol {
	if node == nil {
		return nil
	}
	
	switch node.Type {
	case "class_declaration":
		return m.extractDartClassSymbol(node, filePath, language)
		
	case "state_class_declaration":
		return m.extractDartStateClassSymbol(node, filePath, language)
		
	case "mixin_declaration":
		return m.extractDartMixinSymbol(node, filePath, language)
		
	case "extension_declaration":
		return m.extractDartExtensionSymbol(node, filePath, language)
		
	case "enum_declaration":
		return m.extractDartEnumSymbol(node, filePath, language)
		
	case "typedef_declaration":
		return m.extractDartTypedefSymbol(node, filePath, language)
		
	case "function_declaration":
		return m.extractDartFunctionSymbol(node, filePath, language)
		
	case "method_declaration":
		return m.extractDartMethodSymbol(node, filePath, language)
		
	case "build_method":
		return m.extractDartBuildMethodSymbol(node, filePath, language)
		
	case "lifecycle_method":
		return m.extractDartLifecycleMethodSymbol(node, filePath, language)
		
	case "variable_declaration":
		return m.extractDartVariableSymbol(node, filePath, language)
		
	case "import_statement":
		return m.extractDartImportSymbol(node, filePath, language)
		
	case "part_directive":
		return m.extractDartPartDirectiveSymbol(node, filePath, language)
		
	case "part_of_directive":
		return m.extractDartPartOfDirectiveSymbol(node, filePath, language)
	
	// Week 6 async patterns
	case "async_generator":
		return m.extractDartAsyncGeneratorSymbol(node, filePath, language)
		
	case "async_function":
		return m.extractDartAsyncFunctionSymbol(node, filePath, language)
		
	case "async_method":
		return m.extractDartAsyncMethodSymbol(node, filePath, language)
		
	// Week 6 functional patterns
	case "higher_order_function":
		return m.extractDartHigherOrderFunctionSymbol(node, filePath, language)
		
	case "higher_order_method":
		return m.extractDartHigherOrderMethodSymbol(node, filePath, language)
		
	case "closure_factory":
		return m.extractDartClosureFactorySymbol(node, filePath, language)
		
	case "function_typedef":
		return m.extractDartFunctionTypedefSymbol(node, filePath, language)
		
	default:
		return nil
	}
}

// extractDartClassSymbol extracts class symbols with Flutter detection
func (m *Manager) extractDartClassSymbol(node *types.ASTNode, filePath, language string) *types.Symbol {
	name := m.extractSymbolName(node)
	symbolType := types.SymbolTypeClass
	
	// Check if this is a Flutter widget
	if m.isFlutterWidget(node, name) {
		symbolType = types.SymbolTypeWidget
	}
	
	symbol := &types.Symbol{
		Id:           types.SymbolId(fmt.Sprintf("class-%s-%d", filePath, node.Location.Line)),
		Name:         name,
		Type:         symbolType,
		Location:     convertLocation(node.Location),
		Language:     language,
		Hash:         calculateHash(node.Value),
		LastModified: time.Now(),
	}
	
	// For now, we'll store Dart metadata in the node's metadata instead
	// since Symbol doesn't have a metadata field
	if symbolType == types.SymbolTypeWidget && node.Metadata == nil {
		node.Metadata = make(map[string]any)
		node.Metadata["flutter_type"] = "widget"
		node.Metadata["widget_type"] = m.detectWidgetType(node.Value)
		node.Metadata["has_build_method"] = m.hasBuildMethod(node)
	}
	
	return symbol
}

// extractDartStateClassSymbol extracts State class symbols (Flutter StatefulWidget state classes)
func (m *Manager) extractDartStateClassSymbol(node *types.ASTNode, filePath, language string) *types.Symbol {
	name := m.extractSymbolName(node)
	
	symbol := &types.Symbol{
		Id:           types.SymbolId(fmt.Sprintf("state-class-%s-%d", filePath, node.Location.Line)),
		Name:         name,
		Type:         types.SymbolTypeStateClass, // Use the specific state class type
		Location:     convertLocation(node.Location),
		Language:     language,
		Hash:         calculateHash(node.Value),
		LastModified: time.Now(),
	}
	
	// Add Flutter-specific metadata to the AST node for context
	if node.Metadata == nil {
		node.Metadata = make(map[string]any)
	}
	node.Metadata["flutter_type"] = "state_class"
	node.Metadata["extends"] = "State"
	node.Metadata["has_lifecycle_methods"] = m.hasLifecycleMethods(node)
	
	return symbol
}

// extractDartMixinSymbol extracts mixin symbols
func (m *Manager) extractDartMixinSymbol(node *types.ASTNode, filePath, language string) *types.Symbol {
	name := m.extractSymbolName(node)
	
	symbol := &types.Symbol{
		Id:           types.SymbolId(fmt.Sprintf("mixin-%s-%d", filePath, node.Location.Line)),
		Name:         name,
		Type:         types.SymbolTypeMixin,
		Location:     convertLocation(node.Location),
		Language:     language,
		Hash:         calculateHash(node.Value),
		LastModified: time.Now(),
	}
	
	// Add mixin-specific metadata
	if node.Metadata == nil {
		node.Metadata = make(map[string]any)
	}
	node.Metadata["dart_type"] = "mixin"
	node.Metadata["has_constraint"] = strings.Contains(node.Value, " on ")
	node.Metadata["constraint_type"] = m.extractMixinConstraint(node.Value)
	
	return symbol
}

// extractDartExtensionSymbol extracts extension symbols
func (m *Manager) extractDartExtensionSymbol(node *types.ASTNode, filePath, language string) *types.Symbol {
	name := m.extractSymbolName(node)
	
	symbol := &types.Symbol{
		Id:           types.SymbolId(fmt.Sprintf("extension-%s-%d", filePath, node.Location.Line)),
		Name:         name,
		Type:         types.SymbolTypeExtension,
		Location:     convertLocation(node.Location),
		Language:     language,
		Hash:         calculateHash(node.Value),
		LastModified: time.Now(),
	}
	
	// Add extension-specific metadata
	if node.Metadata == nil {
		node.Metadata = make(map[string]any)
	}
	node.Metadata["dart_type"] = "extension"
	node.Metadata["extends_type"] = m.extractExtensionTarget(node)
	node.Metadata["is_unnamed"] = name == "" || strings.HasPrefix(name, "Extension")
	
	return symbol
}

// extractDartEnumSymbol extracts enum symbols
func (m *Manager) extractDartEnumSymbol(node *types.ASTNode, filePath, language string) *types.Symbol {
	name := m.extractSymbolName(node)
	
	symbol := &types.Symbol{
		Id:           types.SymbolId(fmt.Sprintf("enum-%s-%d", filePath, node.Location.Line)),
		Name:         name,
		Type:         types.SymbolTypeEnum,
		Location:     convertLocation(node.Location),
		Language:     language,
		Hash:         calculateHash(node.Value),
		LastModified: time.Now(),
	}
	
	// Add enum-specific metadata
	if node.Metadata == nil {
		node.Metadata = make(map[string]any)
	}
	node.Metadata["dart_type"] = "enum"
	node.Metadata["is_enhanced"] = m.isEnhancedEnum(node)
	node.Metadata["value_count"] = m.countEnumValues(node)
	node.Metadata["has_methods"] = m.enumHasMethods(node)
	
	return symbol
}

// extractDartTypedefSymbol extracts typedef symbols  
func (m *Manager) extractDartTypedefSymbol(node *types.ASTNode, filePath, language string) *types.Symbol {
	name := m.extractSymbolName(node)
	
	symbol := &types.Symbol{
		Id:           types.SymbolId(fmt.Sprintf("typedef-%s-%d", filePath, node.Location.Line)),
		Name:         name,
		Type:         types.SymbolTypeTypedef,
		Location:     convertLocation(node.Location),
		Language:     language,
		Hash:         calculateHash(node.Value),
		LastModified: time.Now(),
		Signature:    m.extractTypedefSignature(node),
	}
	
	// Add typedef-specific metadata
	if node.Metadata == nil {
		node.Metadata = make(map[string]any)
	}
	node.Metadata["dart_type"] = "typedef"
	node.Metadata["target_type"] = m.extractTypedefTargetType(node)
	node.Metadata["is_function_type"] = m.isFunctionTypedef(node)
	node.Metadata["is_generic"] = strings.Contains(node.Value, "<")
	
	return symbol
}

// extractDartLifecycleMethodSymbol extracts Flutter lifecycle method symbols
func (m *Manager) extractDartLifecycleMethodSymbol(node *types.ASTNode, filePath, language string) *types.Symbol {
	name := m.extractSymbolName(node)
	
	symbol := &types.Symbol{
		Id:           types.SymbolId(fmt.Sprintf("lifecycle-%s-%d", filePath, node.Location.Line)),
		Name:         name,
		Type:         types.SymbolTypeLifecycleMethod, // Use the specific lifecycle method type
		Location:     convertLocation(node.Location),
		Language:     language,
		Hash:         calculateHash(node.Value),
		LastModified: time.Now(),
		Signature:    fmt.Sprintf("void %s()", name), // Most lifecycle methods are void with no params
	}
	
	// Add Flutter-specific metadata
	if node.Metadata == nil {
		node.Metadata = make(map[string]any)
	}
	node.Metadata["flutter_type"] = "lifecycle_method"
	node.Metadata["lifecycle_stage"] = name
	node.Metadata["has_override"] = strings.Contains(node.Value, "@override")
	node.Metadata["widget_lifecycle"] = m.getLifecycleStage(name)
	
	return symbol
}

// extractDartFunctionSymbol extracts function symbols
func (m *Manager) extractDartFunctionSymbol(node *types.ASTNode, filePath, language string) *types.Symbol {
	return &types.Symbol{
		Id:           types.SymbolId(fmt.Sprintf("function-%s-%d", filePath, node.Location.Line)),
		Name:         m.extractSymbolName(node),
		Type:         types.SymbolTypeFunction,
		Location:     convertLocation(node.Location),
		Language:     language,
		Hash:         calculateHash(node.Value),
		LastModified: time.Now(),
		Signature:    m.extractFunctionSignature(node),
	}
}

// extractDartMethodSymbol extracts method symbols
func (m *Manager) extractDartMethodSymbol(node *types.ASTNode, filePath, language string) *types.Symbol {
	return &types.Symbol{
		Id:           types.SymbolId(fmt.Sprintf("method-%s-%d", filePath, node.Location.Line)),
		Name:         m.extractSymbolName(node),
		Type:         types.SymbolTypeMethod,
		Location:     convertLocation(node.Location),
		Language:     language,
		Hash:         calculateHash(node.Value),
		LastModified: time.Now(),
		Signature:    m.extractFunctionSignature(node),
	}
}

// extractDartBuildMethodSymbol extracts Flutter build method symbols
func (m *Manager) extractDartBuildMethodSymbol(node *types.ASTNode, filePath, language string) *types.Symbol {
	symbol := &types.Symbol{
		Id:           types.SymbolId(fmt.Sprintf("build-%s-%d", filePath, node.Location.Line)),
		Name:         "build",
		Type:         types.SymbolTypeBuildMethod,
		Location:     convertLocation(node.Location),
		Language:     language,
		Hash:         calculateHash(node.Value),
		LastModified: time.Now(),
		Signature:    "Widget build(BuildContext context)",
	}
	
	// Store Flutter metadata in the AST node
	if node.Metadata == nil {
		node.Metadata = make(map[string]any)
	}
	node.Metadata["flutter_type"] = "build_method"
	node.Metadata["has_override"] = strings.Contains(node.Value, "@override")
	
	return symbol
}

// extractDartVariableSymbol extracts variable symbols
func (m *Manager) extractDartVariableSymbol(node *types.ASTNode, filePath, language string) *types.Symbol {
	return &types.Symbol{
		Id:           types.SymbolId(fmt.Sprintf("variable-%s-%d", filePath, node.Location.Line)),
		Name:         m.extractSymbolName(node),
		Type:         types.SymbolTypeVariable,
		Location:     convertLocation(node.Location),
		Language:     language,
		Hash:         calculateHash(node.Value),
		LastModified: time.Now(),
	}
}

// extractDartImportSymbol extracts import symbols
func (m *Manager) extractDartImportSymbol(node *types.ASTNode, filePath, language string) *types.Symbol {
	importPath := m.extractImportPath(node)
	return &types.Symbol{
		Id:           types.SymbolId(fmt.Sprintf("import-%s-%d", filePath, node.Location.Line)),
		Name:         importPath,
		Type:         types.SymbolTypeImport,
		Location:     convertLocation(node.Location),
		Language:     language,
		Hash:         calculateHash(node.Value),
		LastModified: time.Now(),
	}
}

// Helper methods for Flutter detection
func (m *Manager) isFlutterWidget(node *types.ASTNode, className string) bool {
	return flutterPatterns["statelessWidget"].MatchString(node.Value) ||
		   flutterPatterns["statefulWidget"].MatchString(node.Value)
}

func (m *Manager) detectWidgetType(nodeValue string) string {
	if flutterPatterns["statelessWidget"].MatchString(nodeValue) {
		return "stateless"
	}
	if flutterPatterns["statefulWidget"].MatchString(nodeValue) {
		return "stateful"
	}
	return ""
}

func (m *Manager) hasBuildMethod(node *types.ASTNode) bool {
	for _, child := range node.Children {
		if child.Type == "build_method" {
			return true
		}
	}
	return false
}

func (m *Manager) extractImportPath(node *types.ASTNode) string {
	for _, child := range node.Children {
		if child.Type == "string_literal" {
			return child.Value
		}
	}
	return "unknown"
}

// hasLifecycleMethods checks if a State class contains lifecycle methods
func (m *Manager) hasLifecycleMethods(node *types.ASTNode) bool {
	for _, child := range node.Children {
		if child.Type == "lifecycle_method" {
			return true
		}
	}
	return false
}

// getLifecycleStage returns the lifecycle stage category for a lifecycle method
func (m *Manager) getLifecycleStage(methodName string) string {
	switch methodName {
	case "initState":
		return "initialization"
	case "didChangeDependencies":
		return "initialization"
	case "build":
		return "rendering"
	case "didUpdateWidget":
		return "update"
	case "setState":
		return "update"
	case "deactivate":
		return "disposal"
	case "dispose":
		return "disposal"
	default:
		return "unknown"
	}
}

// extractMixinConstraint extracts the constraint type from a mixin declaration
func (m *Manager) extractMixinConstraint(mixinDeclaration string) string {
	onPattern := regexp.MustCompile(`\son\s+([\w<>,\s]+)\s*\{`)
	if matches := onPattern.FindStringSubmatch(mixinDeclaration); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// extractExtensionTarget extracts the target type from an extension node
func (m *Manager) extractExtensionTarget(node *types.ASTNode) string {
	for _, child := range node.Children {
		if child.Type == "type_identifier" {
			return child.Value
		}
	}
	return "unknown"
}

// isEnhancedEnum checks if an enum uses Dart 3.0+ enhanced enum features
func (m *Manager) isEnhancedEnum(node *types.ASTNode) bool {
	// Enhanced enums have constructors, methods, or implements clauses
	return strings.Contains(node.Value, "const ") || 
		   strings.Contains(node.Value, "implements ") ||
		   strings.Contains(node.Value, "{") && strings.Contains(node.Value, "(")
}

// countEnumValues counts the number of enum values in an enum
func (m *Manager) countEnumValues(node *types.ASTNode) int {
	count := 0
	for _, child := range node.Children {
		if child.Type == "enum_value" {
			count++
		}
	}
	return count
}

// enumHasMethods checks if an enum has custom methods (enhanced enum)
func (m *Manager) enumHasMethods(node *types.ASTNode) bool {
	// Look for method-like patterns in the enum body
	for _, child := range node.Children {
		if child.Type == "method_declaration" {
			return true
		}
	}
	// Also check for methods in the enum content
	return strings.Contains(node.Value, "() {") || strings.Contains(node.Value, "get ")
}

// extractTypedefSignature extracts the full signature of a typedef
func (m *Manager) extractTypedefSignature(node *types.ASTNode) string {
	// Extract the part after the typedef name
	parts := strings.SplitN(node.Value, "=", 2)
	if len(parts) > 1 {
		return strings.TrimSpace(strings.TrimSuffix(parts[1], ";"))
	}
	return ""
}

// extractTypedefTargetType extracts the target type from a typedef node
func (m *Manager) extractTypedefTargetType(node *types.ASTNode) string {
	for _, child := range node.Children {
		if child.Type == "type_identifier" {
			return child.Value
		}
	}
	return ""
}

// isFunctionTypedef checks if a typedef defines a function type
func (m *Manager) isFunctionTypedef(node *types.ASTNode) bool {
	return strings.Contains(node.Value, "(") && strings.Contains(node.Value, ")")
}

// extractDartPartDirectiveSymbol extracts part directive symbols
func (m *Manager) extractDartPartDirectiveSymbol(node *types.ASTNode, filePath, language string) *types.Symbol {
	name := m.extractSymbolName(node)
	
	symbol := &types.Symbol{
		Id:           types.SymbolId(fmt.Sprintf("part-directive-%s-%d", filePath, node.Location.Line)),
		Name:         name,
		Type:         types.SymbolTypeDirective,
		Location:     convertLocation(node.Location),
		Language:     language,
		Hash:         calculateHash(node.Value),
		LastModified: time.Now(),
	}
	
	return symbol
}

// extractDartPartOfDirectiveSymbol extracts part of directive symbols
func (m *Manager) extractDartPartOfDirectiveSymbol(node *types.ASTNode, filePath, language string) *types.Symbol {
	name := m.extractSymbolName(node)
	
	symbol := &types.Symbol{
		Id:           types.SymbolId(fmt.Sprintf("part-of-directive-%s-%d", filePath, node.Location.Line)),
		Name:         name,
		Type:         types.SymbolTypeDirective,
		Location:     convertLocation(node.Location),
		Language:     language,
		Hash:         calculateHash(node.Value),
		LastModified: time.Now(),
	}
	
	return symbol
}

// Week 6 Async Pattern Symbol Extractors

// extractDartAsyncGeneratorSymbol extracts async generator function symbols
func (m *Manager) extractDartAsyncGeneratorSymbol(node *types.ASTNode, filePath, language string) *types.Symbol {
	name := m.extractSymbolName(node)
	
	symbol := &types.Symbol{
		Id:           types.SymbolId(fmt.Sprintf("async-generator-%s-%d", filePath, node.Location.Line)),
		Name:         name,
		Type:         types.SymbolTypeFunction,
		Location:     convertLocation(node.Location),
		Language:     language,
		Hash:         calculateHash(node.Value),
		LastModified: time.Now(),
	}
	
	return symbol
}

// extractDartAsyncFunctionSymbol extracts async function symbols
func (m *Manager) extractDartAsyncFunctionSymbol(node *types.ASTNode, filePath, language string) *types.Symbol {
	name := m.extractSymbolName(node)
	
	symbol := &types.Symbol{
		Id:           types.SymbolId(fmt.Sprintf("async-function-%s-%d", filePath, node.Location.Line)),
		Name:         name,
		Type:         types.SymbolTypeFunction,
		Location:     convertLocation(node.Location),
		Language:     language,
		Hash:         calculateHash(node.Value),
		LastModified: time.Now(),
	}
	
	return symbol
}

// extractDartAsyncMethodSymbol extracts async method symbols
func (m *Manager) extractDartAsyncMethodSymbol(node *types.ASTNode, filePath, language string) *types.Symbol {
	name := m.extractSymbolName(node)
	
	symbol := &types.Symbol{
		Id:           types.SymbolId(fmt.Sprintf("async-method-%s-%d", filePath, node.Location.Line)),
		Name:         name,
		Type:         types.SymbolTypeMethod,
		Location:     convertLocation(node.Location),
		Language:     language,
		Hash:         calculateHash(node.Value),
		LastModified: time.Now(),
	}
	
	return symbol
}

// Week 6 Functional Pattern Symbol Extractors

// extractDartHigherOrderFunctionSymbol extracts higher-order function symbols
func (m *Manager) extractDartHigherOrderFunctionSymbol(node *types.ASTNode, filePath, language string) *types.Symbol {
	name := m.extractSymbolName(node)
	
	symbol := &types.Symbol{
		Id:           types.SymbolId(fmt.Sprintf("higher-order-%s-%d", filePath, node.Location.Line)),
		Name:         name,
		Type:         types.SymbolTypeMethod,  // Higher-order functions are treated as methods
		Location:     convertLocation(node.Location),
		Language:     language,
		Hash:         calculateHash(node.Value),
		LastModified: time.Now(),
	}
	
	return symbol
}

// extractDartHigherOrderMethodSymbol extracts higher-order method symbols (methods inside classes)
func (m *Manager) extractDartHigherOrderMethodSymbol(node *types.ASTNode, filePath, language string) *types.Symbol {
	name := m.extractSymbolName(node)
	
	symbol := &types.Symbol{
		Id:           types.SymbolId(fmt.Sprintf("higher-order-method-%s-%d", filePath, node.Location.Line)),
		Name:         name,
		Type:         types.SymbolTypeMethod,  // Higher-order methods are treated as methods
		Location:     convertLocation(node.Location),
		Language:     language,
		Hash:         calculateHash(node.Value),
		LastModified: time.Now(),
	}
	
	return symbol
}

// extractDartClosureFactorySymbol extracts closure factory function symbols
func (m *Manager) extractDartClosureFactorySymbol(node *types.ASTNode, filePath, language string) *types.Symbol {
	name := m.extractSymbolName(node)
	
	symbol := &types.Symbol{
		Id:           types.SymbolId(fmt.Sprintf("closure-factory-%s-%d", filePath, node.Location.Line)),
		Name:         name,
		Type:         types.SymbolTypeMethod,  // Closure factories are treated as methods
		Location:     convertLocation(node.Location),
		Language:     language,
		Hash:         calculateHash(node.Value),
		LastModified: time.Now(),
	}
	
	return symbol
}

// extractDartFunctionTypedefSymbol extracts function typedef symbols
func (m *Manager) extractDartFunctionTypedefSymbol(node *types.ASTNode, filePath, language string) *types.Symbol {
	name := m.extractSymbolName(node)
	
	symbol := &types.Symbol{
		Id:           types.SymbolId(fmt.Sprintf("function-typedef-%s-%d", filePath, node.Location.Line)),
		Name:         name,
		Type:         types.SymbolTypeTypedef,
		Location:     convertLocation(node.Location),
		Language:     language,
		Hash:         calculateHash(node.Value),
		LastModified: time.Now(),
	}
	
	return symbol
}
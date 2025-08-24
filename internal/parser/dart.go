package parser

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	
	"github.com/nuthan-ms/codecontext/pkg/types"
)

// Dart language patterns for regex-based parsing (fallback approach)
var dartPatterns = map[string]*regexp.Regexp{
	"class":      regexp.MustCompile(`(?m)^(?:abstract\s+)?class\s+(\w+)(?:<[\w\s,<>]+>)?(?:\s+extends\s+[\w<>]+)?(?:\s+with\s+[\w\s,<>]+)?(?:\s+implements\s+[\w\s,<>]+)?\s*{`),
	"stateClass": regexp.MustCompile(`(?m)^(?:abstract\s+)?class\s+(\w+)\s+extends\s+State<[\w<>]+>`),
	"mixinClass": regexp.MustCompile(`(?m)^(?:abstract\s+)?class\s+(\w+)(?:\s+extends\s+[\w<>]+)?\s+with\s+([\w\s,<>]+)(?:\s+implements\s+[\w\s,<>]+)?\s*{`),
	"mixin":      regexp.MustCompile(`(?m)^mixin\s+(\w+(?:<[\w\s,<>]+>)?)(?:\s+on\s+[\w\s,<>]+)?\s*{`),
	"extension":  regexp.MustCompile(`(?m)^extension\s+(\w*(?:<[\w,\s]+>)?)\s*on\s+([\w<>\[\],\s]+)\s*{`),
	"enum":       regexp.MustCompile(`(?m)^enum\s+(\w+)(?:<[\w\s,<>]+>)?(?:\s+implements\s+[\w\s,<>]+)?\s*{`),
	"enumValue":  regexp.MustCompile(`(?m)^\s*(\w+)(?:\([^)]*\))?(?:\s*,|\s*;|\s*})`),
	"typedef":    regexp.MustCompile(`(?m)^typedef\s+(\w+)(?:<[\w\s,<>]+>)?\s*=\s*([\w<>\[\],\s\(\)]+);`),
	"function":   regexp.MustCompile(`(?m)^(?:[\w<>\[\],\s]+\s+)?(\w+)\s*\([^{]*\)\s*(?:async\s*)?{`),
	"method":     regexp.MustCompile(`(?m)^\s+(?:@override\s+)?(?:[\w<>\[\],\s]+\s+)(\w+)\s*\([^{}]*?\)\s*(?:async\s*)?\s*(?:{|=>)`),
	"privateMethod": regexp.MustCompile(`(?m)^\s+(?:[\w<>\[\],\s]+\s+)?(_\w+)\s*\([^{]*\)\s*(?:async\s*)?\s*{`),
	"variable":   regexp.MustCompile(`(?m)^\s*(?:final\s+|const\s+|var\s+|static\s+)?(?:[\w<>\[\],\s?]+\s+)?(\w+)\s*=`),
	"import":     regexp.MustCompile(`(?m)^\s*import\s+['"]([^'"]+)['"](?:\s+as\s+\w+)?;`),
	"buildMethod": regexp.MustCompile(`(?m)^\s+(?:@override\s+)?Widget\s+build\s*\(\s*BuildContext\s+\w+\s*\)`),
	"lifecycleMethod": regexp.MustCompile(`(?m)^\s+@override\s+void\s+(initState|dispose|didUpdateWidget|didChangeDependencies)\s*\(`),
	"partDirective":   regexp.MustCompile(`(?m)^part\s+['"]([^'"]+)['"];`),
	"partOfDirective": regexp.MustCompile(`(?m)^part\s+of\s+(?:['"]([^'"]+)['"]|(\w+(?:\.\w+)*));`),
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
	ast := &types.AST{
		Language:  "dart",
		Content:   content,
		FilePath:  filePath,
		Hash:      calculateHash(content),
		Version:   "1.0",
		ParsedAt:  time.Now(),
	}
	
	// Enhanced Flutter analysis
	flutterDetector := NewFlutterDetector()
	flutterAnalysis := flutterDetector.AnalyzeFlutterContent(content)
	
	// Create root AST node with parse metadata
	parseMetadata := map[string]interface{}{
		"parser":         "regex", // Will be "tree-sitter" when we have real bindings
		"parse_quality":  "basic",
		"has_flutter":    flutterAnalysis.IsFlutter,
		"has_errors":     false,
		"error_count":    0,
	}
	
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
		Children: m.extractDartNodes(content),
		Metadata: parseMetadata,
	}
	
	// Integrate Flutter analysis with AST
	m.IntegrateFlutterAnalysis(ast, flutterAnalysis)
	
	return ast, nil
}

// extractDartNodes extracts AST nodes from Dart content using regex patterns
func (m *Manager) extractDartNodes(content string) []*types.ASTNode {
	var nodes []*types.ASTNode
	lines := strings.Split(content, "\n")
	
	// Extract imports
	if matches := dartPatterns["import"].FindAllStringSubmatch(content, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				lineNum := m.findLineNumber(content, match[0])
				nodes = append(nodes, &types.ASTNode{
					Id:   fmt.Sprintf("import-%d", lineNum),
					Type: "import_statement",
					Value: match[0],
					Location: types.FileLocation{
						Line:    lineNum,
						Column:  1,
						EndLine: lineNum,
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
	
	// Extract mixins
	if matches := dartPatterns["mixin"].FindAllStringSubmatch(content, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				lineNum := m.findLineNumber(content, match[0])
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
				mixinContent := m.extractClassContent(content, match[0], lineNum)
				mixinNode.Children = append(mixinNode.Children, m.extractClassMethods(mixinContent, lineNum, mixinName)...)
				
				nodes = append(nodes, mixinNode)
			}
		}
	}
	
	// Extract extensions
	if matches := dartPatterns["extension"].FindAllStringSubmatch(content, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 2 {
				lineNum := m.findLineNumber(content, match[0])
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
				extensionContent := m.extractClassContent(content, match[0], lineNum)
				extensionNode.Children = append(extensionNode.Children, m.extractClassMethods(extensionContent, lineNum, extensionName)...)
				
				nodes = append(nodes, extensionNode)
			}
		}
	}
	
	// Extract enums
	if matches := dartPatterns["enum"].FindAllStringSubmatch(content, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				lineNum := m.findLineNumber(content, match[0])
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
				enumContent := m.extractClassContent(content, match[0], lineNum)
				enumNode.Children = append(enumNode.Children, m.extractEnumValues(enumContent, lineNum)...)
				
				nodes = append(nodes, enumNode)
			}
		}
	}
	
	// Extract typedefs
	if matches := dartPatterns["typedef"].FindAllStringSubmatch(content, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 2 {
				lineNum := m.findLineNumber(content, match[0])
				typedefName := match[1]
				typedefType := match[2]
				
				// For generic typedefs like "Callback<T>", extract just the base name
				if strings.Contains(typedefName, "<") {
					typedefName = strings.Split(typedefName, "<")[0]
				}
				
				typedefNode := &types.ASTNode{
					Id:   fmt.Sprintf("typedef-%s-%d", typedefName, lineNum),
					Type: "typedef_declaration",
					Value: match[0],
					Location: types.FileLocation{
						Line:    lineNum,
						Column:  1,
						EndLine: lineNum,
						EndColumn: len(match[0]) + 1,
					},
					Children: []*types.ASTNode{
						{
							Id:    fmt.Sprintf("typedef-name-%s", typedefName),
							Type:  "identifier",
							Value: typedefName,
						},
						{
							Id:    fmt.Sprintf("typedef-type-%s", typedefType),
							Type:  "type_identifier",
							Value: strings.TrimSpace(typedefType),
						},
					},
				}
				
				nodes = append(nodes, typedefNode)
			}
		}
	}
	
	// Extract classes (regular and state classes)
	if matches := dartPatterns["class"].FindAllStringSubmatch(content, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				lineNum := m.findLineNumber(content, match[0])
				
				// Check if this is a State class
				isStateClass := dartPatterns["stateClass"].MatchString(match[0])
				classType := "class_declaration"
				if isStateClass {
					classType = "state_class_declaration"
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
					Children: []*types.ASTNode{
						{
							Id:    fmt.Sprintf("class-name-%s", match[1]),
							Type:  "identifier",
							Value: match[1],
						},
					},
				}
				
				// Extract methods within the class
				classContent := m.extractClassContent(content, match[0], lineNum)
				classNode.Children = append(classNode.Children, m.extractClassMethods(classContent, lineNum, match[1])...)
				
				nodes = append(nodes, classNode)
			}
		}
	}
	
	// Extract top-level functions
	if matches := dartPatterns["function"].FindAllStringSubmatch(content, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				lineNum := m.findLineNumber(content, match[0])
				// Skip if this is inside a class (crude check)
				if !m.isInsideClass(lines, lineNum-1) {
					nodes = append(nodes, &types.ASTNode{
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
	
	// Extract global variables
	if matches := dartPatterns["variable"].FindAllStringSubmatch(content, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				lineNum := m.findLineNumber(content, match[0])
				// Skip if this is inside a class or function (crude check)
				if !m.isInsideClass(lines, lineNum-1) && !m.isInsideFunction(lines, lineNum-1) {
					nodes = append(nodes, &types.ASTNode{
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
	
	// Extract part directives
	if matches := dartPatterns["partDirective"].FindAllStringSubmatch(content, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				lineNum := m.findLineNumber(content, match[0])
				partFile := match[1]
				
				nodes = append(nodes, &types.ASTNode{
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
	if matches := dartPatterns["partOfDirective"].FindAllStringSubmatch(content, -1); matches != nil {
		for _, match := range matches {
			lineNum := m.findLineNumber(content, match[0])
			var partOfTarget string
			
			// Check if it's a file path (match[1]) or library name (match[2])
			if len(match) > 1 && match[1] != "" {
				partOfTarget = match[1] // File path
			} else if len(match) > 2 && match[2] != "" {
				partOfTarget = match[2] // Library name
			}
			
			if partOfTarget != "" {
				nodes = append(nodes, &types.ASTNode{
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
	
	return nodes
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
	
	// Crude check for indentation - if line starts with spaces, assume it's inside something
	line := lines[lineIndex]
	return len(line) > 0 && (line[0] == ' ' || line[0] == '\t')
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
		node.Metadata = make(map[string]interface{})
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
		node.Metadata = make(map[string]interface{})
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
		node.Metadata = make(map[string]interface{})
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
		node.Metadata = make(map[string]interface{})
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
		node.Metadata = make(map[string]interface{})
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
		node.Metadata = make(map[string]interface{})
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
		node.Metadata = make(map[string]interface{})
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
		node.Metadata = make(map[string]interface{})
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
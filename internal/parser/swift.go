package parser

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
	
	"github.com/nuthan-ms/codecontext/pkg/types"
)

// Swift language patterns for regex-based parsing (fallback approach)
var swiftPatterns = map[string]*regexp.Regexp{
	// Class patterns
	"class":      regexp.MustCompile(`(?m)^(?:(?:public|private|internal|fileprivate|open)\s+)?(?:final\s+)?class\s+(\w+)(?:<[\w\s,<>:]+>)?(?:\s*:\s*[\w\s,<>]+)?\s*{`),
	"finalClass": regexp.MustCompile(`(?m)^(?:(?:public|private|internal|fileprivate|open)\s+)?final\s+class\s+(\w+)(?:<[\w\s,<>:]+>)?(?:\s*:\s*[\w\s,<>]+)?\s*{`),
	
	// Struct patterns
	"struct": regexp.MustCompile(`(?m)^(?:(?:public|private|internal|fileprivate)\s+)?struct\s+(\w+)(?:<[\w\s,<>:]+>)?(?:\s*:\s*[\w\s,<>]+)?\s*{`),
	
	// Protocol patterns
	"protocol": regexp.MustCompile(`(?m)^(?:(?:public|private|internal|fileprivate)\s+)?protocol\s+(\w+)(?:<[\w\s,<>:]+>)?(?:\s*:\s*[\w\s,<>]+)?\s*{`),
	
	// Enum patterns
	"enum": regexp.MustCompile(`(?m)^(?:(?:public|private|internal|fileprivate)\s+)?enum\s+(\w+)(?:<[\w\s,<>:]+>)?(?:\s*:\s*[\w\s,<>]+)?\s*{`),
	
	// Actor patterns (Swift concurrency)
	"actor": regexp.MustCompile(`(?m)^(?:(?:public|private|internal|fileprivate|open)\s+)?actor\s+(\w+)(?:<[\w\s,<>:]+>)?(?:\s*:\s*[\w\s,<>]+)?\s*{`),
	
	// Typealias patterns
	"typealias": regexp.MustCompile(`(?m)^(?:(?:public|private|internal|fileprivate)\s+)?typealias\s+(\w+)(?:<[\w\s,<>:]+>)?\s*=\s*[^\n]+`),
	
	// Function patterns (both top-level and methods)
	"function": regexp.MustCompile(`(?m)(?:^|\s+)(?:(?:public|private|internal|fileprivate|open)\s+)?(?:static\s+)?(?:override\s+)?func\s+(\w+)(?:<[\w\s,<>:]+>)?\s*\([^)]*\)(?:\s*(?:async\s+)?(?:throws\s+)?(?:->\s*[\w<>\[\],\s\?\!]+)?)?\s*{`),
	
	// Initializer patterns
	"init": regexp.MustCompile(`(?m)^\s+(?:(?:public|private|internal|fileprivate)\s+)?(?:convenience\s+)?init(?:\?)?\s*\([^)]*\)(?:\s*(?:async\s+)?(?:throws\s+)?)?\s*{`),
	
	// Deinitializer patterns
	"deinit": regexp.MustCompile(`(?m)^\s+deinit\s*{`),
	
	// Property patterns - both stored and computed
	"storedProperty": regexp.MustCompile(`(?m)^\s+(?:(?:public|private|internal|fileprivate|open)\s+)?(?:static\s+)?(?:let|var)\s+(\w+)\s*:\s*[\w<>\[\],\s\?\!]+\s*=`),
	"computedProperty": regexp.MustCompile(`(?m)^\s+(?:(?:public|private|internal|fileprivate|open)\s+)?(?:static\s+)?var\s+(\w+)\s*:\s*[\w<>\[\],\s\?\!]+\s*\{\s*(?:get|set)`),
	"propertyDeclaration": regexp.MustCompile(`(?m)^\s+(?:(?:public|private|internal|fileprivate|open)\s+)?(?:static\s+)?(?:let|var)\s+(\w+)\s*:\s*[\w<>\[\],\s\?\!]+`),
	
	// Property wrapper patterns
	"propertyWrapper": regexp.MustCompile(`(?m)^\s+(@\w+(?:\([^)]*\))?)\s+(?:(?:public|private|internal|fileprivate)\s+)?(?:var|let)\s+(\w+)`),
	
	// Closure patterns
	"closure": regexp.MustCompile(`(?m)\{\s*\(?[^}]*\)?\s*in[^}]*\}`),
	"closureParameter": regexp.MustCompile(`(?m):\s*@escaping\s*\([^)]*\)\s*->\s*[\w\?\!]+`),
	"trailingClosure": regexp.MustCompile(`(?m)\w+\s*\{\s*[^}]*\}`),
	
	// Async/await patterns
	"asyncFunction": regexp.MustCompile(`(?m)func\s+(\w+)\([^)]*\)\s*async(?:\s+throws)?\s*(?:->\s*[\w<>\?\!\[\]]+)?\s*\{`),
	"awaitCall": regexp.MustCompile(`(?m)\bawait\s+\w+`),
	"asyncProperty": regexp.MustCompile(`(?m)var\s+(\w+)\s*:\s*[\w<>\?\!\[\]]+\s*\{\s*get\s+async`),
	
	// Optional patterns
	"optionalChaining": regexp.MustCompile(`(?m)\w+\?\.\w+`),
	"optionalBinding": regexp.MustCompile(`(?m)(?:if|guard)\s+let\s+(\w+)\s*=`),
	"nilCoalescing": regexp.MustCompile(`(?m)\w+\s*\?\?\s*\w+`),
	"forceUnwrap": regexp.MustCompile(`(?m)\w+\!`),
	
	// Control flow patterns
	"guardStatement": regexp.MustCompile(`(?m)^\s*guard\s+[^{]+\s+else\s*\{`),
	"deferStatement": regexp.MustCompile(`(?m)^\s*defer\s*\{`),
	
	// Associated types in protocols
	"associatedType": regexp.MustCompile(`(?m)^\s+associatedtype\s+(\w+)(?:\s*:\s*[\w\s,<>]+)?`),
	
	// Subscript patterns
	"subscript": regexp.MustCompile(`(?m)^\s+(?:(?:public|private|internal|fileprivate)\s+)?subscript\s*\([^)]+\)\s*->\s*[\w<>\[\],\s\?\!]+\s*\{`),
	
	// Operator overloading patterns
	"operatorFunc": regexp.MustCompile(`(?m)^\s*(?:(?:public|private|internal|fileprivate)\s+)?static\s+func\s+([\+\-\*\/\%\=\!\<\>\&\|\^\~]+)\s*\([^)]*\)(?:\s*->\s*[\w<>\?\!]+)?\s*\{`),
	"operatorDecl": regexp.MustCompile(`(?m)^(?:prefix|postfix|infix)\s+operator\s+([\+\-\*\/\%\=\!\<\>\&\|\^\~]+)`),
	
	// Async sequence patterns
	"asyncSequence": regexp.MustCompile(`(?m)for\s+await\s+\w+\s+in\s+\w+`),
	"asyncIterator": regexp.MustCompile(`(?m):\s*AsyncSequence`),
	
	// Result builder patterns
	"resultBuilder": regexp.MustCompile(`(?m)@resultBuilder\s+(?:struct|class|enum)\s+(\w+)`),
	"viewBuilder": regexp.MustCompile(`(?m)@ViewBuilder\s+(?:var|func)\s+(\w+)`),
	"functionBuilder": regexp.MustCompile(`(?m)@_functionBuilder\s+(?:struct|class|enum)\s+(\w+)`),
	
	// Macro patterns (Swift 5.9+)
	"macroDecl": regexp.MustCompile(`(?ms)@(?:freestanding|attached)\s*\([^)]+\)\s*macro\s+(\w+)`),
	"macroUsage": regexp.MustCompile(`(?m)#(\w+)\s*\(`),
	
	// Enhanced property wrapper patterns
	"complexPropertyWrapper": regexp.MustCompile(`(?m)^\s+(@\w+\([^)]*\))\s+(?:(?:public|private|internal|fileprivate)\s+)?(?:var|let)\s+(\w+)`),
	
	// Extension patterns
	"extension": regexp.MustCompile(`(?m)^(?:(?:public|private|internal|fileprivate)\s+)?extension\s+(\w+)(?:<[\w\s,<>:]+>)?(?:\s*:\s*[\w\s,<>]+)?\s*{`),
	
	// Import patterns
	"import": regexp.MustCompile(`(?m)^import\s+(?:(?:public|private|internal|fileprivate)\s+)?(\w+)(?:\.[\w\.]+)?`),
}

// parseSwiftContentWithContext parses Swift content using regex patterns
func (m *Manager) parseSwiftContentWithContext(ctx context.Context, content, filePath string) (*types.AST, error) {
	ast := &types.AST{
		Language:       "swift",
		Content:        content,
		FilePath:       filePath,
		Hash:           calculateHash(content),
		Version:        "1.0",
		ParsedAt:       time.Now(),
		TreeSitterTree: nil,
	}

	// Create root node
	root := &types.ASTNode{
		Id:   "swift-root",
		Type: "compilation_unit",
		Location: types.FileLocation{
			FilePath: filePath,
			Line:     1,
			Column:   1,
		},
		Value:    content,
		Children: []*types.ASTNode{},
		Metadata: make(map[string]interface{}),
	}

	// Parse Swift constructs using regex patterns
	m.parseSwiftClasses(content, root)
	m.parseSwiftStructs(content, root)
	m.parseSwiftProtocols(content, root)
	m.parseSwiftEnums(content, root)
	m.parseSwiftActors(content, root)
	m.parseSwiftTypealias(content, root)
	m.parseSwiftFunctions(content, root)
	m.parseSwiftProperties(content, root)
	m.parseSwiftExtensions(content, root)
	m.parseSwiftImports(content, root)
	m.parseSwiftAssociatedTypes(content, root)
	
	// Parse advanced Swift patterns
	m.parseSwiftClosures(content, root)
	m.parseSwiftAsyncAwait(content, root)
	m.parseSwiftOptionals(content, root)
	m.parseSwiftControlFlow(content, root)
	
	// Parse P1/P2 Swift patterns
	m.parseSwiftSubscripts(content, root)
	m.parseSwiftOperators(content, root)
	m.parseSwiftAsyncSequences(content, root)
	m.parseSwiftResultBuilders(content, root)
	m.parseSwiftMacros(content, root)

	// Detect framework usage
	m.detectSwiftFrameworks(content, root)

	ast.Root = root
	return ast, nil
}

// parseSwiftClasses extracts class declarations
func (m *Manager) parseSwiftClasses(content string, root *types.ASTNode) {
	matches := swiftPatterns["class"].FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			className := match[1]
			lineNum := strings.Count(content[:strings.Index(content, match[0])], "\n") + 1
			
			classNode := &types.ASTNode{
				Id:   fmt.Sprintf("class-%s-%d", className, lineNum),
				Type: "class_declaration",
				Location: types.FileLocation{
					FilePath: root.Location.FilePath,
					Line:     lineNum,
					Column:   1,
				},
				Value: match[0],
				Children: []*types.ASTNode{
					{
						Id:   fmt.Sprintf("class-name-%s", className),
						Type: "identifier",
						Value: className,
						Location: types.FileLocation{
							FilePath: root.Location.FilePath,
							Line:     lineNum,
							Column:   strings.Index(match[0], className) + 1,
						},
					},
				},
			}
			root.Children = append(root.Children, classNode)
		}
	}
}

// parseSwiftStructs extracts struct declarations
func (m *Manager) parseSwiftStructs(content string, root *types.ASTNode) {
	matches := swiftPatterns["struct"].FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			structName := match[1]
			lineNum := strings.Count(content[:strings.Index(content, match[0])], "\n") + 1
			
			structNode := &types.ASTNode{
				Id:   fmt.Sprintf("struct-%s-%d", structName, lineNum),
				Type: "struct_declaration",
				Location: types.FileLocation{
					FilePath: root.Location.FilePath,
					Line:     lineNum,
					Column:   1,
				},
				Value: match[0],
				Children: []*types.ASTNode{
					{
						Id:   fmt.Sprintf("struct-name-%s", structName),
						Type: "identifier",
						Value: structName,
						Location: types.FileLocation{
							FilePath: root.Location.FilePath,
							Line:     lineNum,
							Column:   strings.Index(match[0], structName) + 1,
						},
					},
				},
			}
			root.Children = append(root.Children, structNode)
		}
	}
}

// parseSwiftProtocols extracts protocol declarations
func (m *Manager) parseSwiftProtocols(content string, root *types.ASTNode) {
	matches := swiftPatterns["protocol"].FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			protocolName := match[1]
			lineNum := strings.Count(content[:strings.Index(content, match[0])], "\n") + 1
			
			protocolNode := &types.ASTNode{
				Id:   fmt.Sprintf("protocol-%s-%d", protocolName, lineNum),
				Type: "protocol_declaration",
				Location: types.FileLocation{
					FilePath: root.Location.FilePath,
					Line:     lineNum,
					Column:   1,
				},
				Value: match[0],
				Children: []*types.ASTNode{
					{
						Id:   fmt.Sprintf("protocol-name-%s", protocolName),
						Type: "identifier",
						Value: protocolName,
						Location: types.FileLocation{
							FilePath: root.Location.FilePath,
							Line:     lineNum,
							Column:   strings.Index(match[0], protocolName) + 1,
						},
					},
				},
			}
			root.Children = append(root.Children, protocolNode)
		}
	}
}

// parseSwiftEnums extracts enum declarations
func (m *Manager) parseSwiftEnums(content string, root *types.ASTNode) {
	matches := swiftPatterns["enum"].FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			enumName := match[1]
			lineNum := strings.Count(content[:strings.Index(content, match[0])], "\n") + 1
			
			enumNode := &types.ASTNode{
				Id:   fmt.Sprintf("enum-%s-%d", enumName, lineNum),
				Type: "enum_declaration",
				Location: types.FileLocation{
					FilePath: root.Location.FilePath,
					Line:     lineNum,
					Column:   1,
				},
				Value: match[0],
			}
			root.Children = append(root.Children, enumNode)
		}
	}
}

// parseSwiftFunctions extracts function declarations
func (m *Manager) parseSwiftFunctions(content string, root *types.ASTNode) {
	// All functions and methods (both top-level and inside classes/structs)
	matches := swiftPatterns["function"].FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			funcName := match[1]
			lineNum := strings.Count(content[:strings.Index(content, match[0])], "\n") + 1
			
			funcNode := &types.ASTNode{
				Id:   fmt.Sprintf("func-%s-%d", funcName, lineNum),
				Type: "function_declaration",
				Location: types.FileLocation{
					FilePath: root.Location.FilePath,
					Line:     lineNum,
					Column:   1,
				},
				Value: match[0],
				Children: []*types.ASTNode{
					{
						Id:   fmt.Sprintf("func-name-%s", funcName),
						Type: "identifier",
						Value: funcName,
						Location: types.FileLocation{
							FilePath: root.Location.FilePath,
							Line:     lineNum,
							Column:   strings.Index(match[0], funcName) + 1,
						},
					},
				},
			}
			root.Children = append(root.Children, funcNode)
		}
	}
	
	// Initializers
	initMatches := swiftPatterns["init"].FindAllStringSubmatch(content, -1)
	for _, match := range initMatches {
		lineNum := strings.Count(content[:strings.Index(content, match[0])], "\n") + 1
		
		initNode := &types.ASTNode{
			Id:   fmt.Sprintf("init-%d", lineNum),
			Type: "init_declaration",
			Location: types.FileLocation{
				FilePath: root.Location.FilePath,
				Line:     lineNum,
				Column:   1,
			},
			Value: match[0],
		}
		root.Children = append(root.Children, initNode)
	}
	
	// Deinitializers
	deinitMatches := swiftPatterns["deinit"].FindAllStringSubmatch(content, -1)
	for _, match := range deinitMatches {
		lineNum := strings.Count(content[:strings.Index(content, match[0])], "\n") + 1
		
		deinitNode := &types.ASTNode{
			Id:   fmt.Sprintf("deinit-%d", lineNum),
			Type: "deinit_declaration",
			Location: types.FileLocation{
				FilePath: root.Location.FilePath,
				Line:     lineNum,
				Column:   1,
			},
			Value: match[0],
		}
		root.Children = append(root.Children, deinitNode)
	}
}

// parseSwiftProperties extracts property declarations (stored, computed, and wrapped)
func (m *Manager) parseSwiftProperties(content string, root *types.ASTNode) {
	// Parse complex property wrappers first (highest priority)
	complexWrapperMatches := swiftPatterns["complexPropertyWrapper"].FindAllStringSubmatch(content, -1)
	for _, match := range complexWrapperMatches {
		if len(match) > 2 {
			wrapper := match[1]
			propName := match[2]
			lineNum := strings.Count(content[:strings.Index(content, match[0])], "\n") + 1
			
			propNode := &types.ASTNode{
				Id:   fmt.Sprintf("property-%s-%d", propName, lineNum),
				Type: "property_declaration",
				Location: types.FileLocation{
					FilePath: root.Location.FilePath,
					Line:     lineNum,
					Column:   1,
				},
				Value: match[0],
				Metadata: map[string]interface{}{
					"wrapper": wrapper,
					"is_wrapped": true,
					"has_wrapper_args": true,
				},
				Children: []*types.ASTNode{
					{
						Id:   fmt.Sprintf("property-name-%s", propName),
						Type: "identifier",
						Value: propName,
						Location: types.FileLocation{
							FilePath: root.Location.FilePath,
							Line:     lineNum,
							Column:   strings.Index(match[0], propName) + 1,
						},
					},
				},
			}
			root.Children = append(root.Children, propNode)
		}
	}
	
	// Parse simple property wrappers (without arguments)
	wrapperMatches := swiftPatterns["propertyWrapper"].FindAllStringSubmatch(content, -1)
	for _, match := range wrapperMatches {
		if len(match) > 2 {
			wrapper := match[1]
			propName := match[2]
			lineNum := strings.Count(content[:strings.Index(content, match[0])], "\n") + 1
			
			// Skip if already processed by complex wrapper pattern
			if strings.Contains(wrapper, "(") {
				continue
			}
			
			propNode := &types.ASTNode{
				Id:   fmt.Sprintf("property-%s-%d", propName, lineNum),
				Type: "property_declaration",
				Location: types.FileLocation{
					FilePath: root.Location.FilePath,
					Line:     lineNum,
					Column:   1,
				},
				Value: match[0],
				Metadata: map[string]interface{}{
					"wrapper": wrapper,
					"is_wrapped": true,
				},
				Children: []*types.ASTNode{
					{
						Id:   fmt.Sprintf("property-name-%s", propName),
						Type: "identifier",
						Value: propName,
						Location: types.FileLocation{
							FilePath: root.Location.FilePath,
							Line:     lineNum,
							Column:   strings.Index(match[0], propName) + 1,
						},
					},
				},
			}
			root.Children = append(root.Children, propNode)
		}
	}
	
	// Parse computed properties
	computedMatches := swiftPatterns["computedProperty"].FindAllStringSubmatch(content, -1)
	for _, match := range computedMatches {
		if len(match) > 1 {
			propName := match[1]
			lineNum := strings.Count(content[:strings.Index(content, match[0])], "\n") + 1
			
			propNode := &types.ASTNode{
				Id:   fmt.Sprintf("property-%s-%d", propName, lineNum),
				Type: "property_declaration",
				Location: types.FileLocation{
					FilePath: root.Location.FilePath,
					Line:     lineNum,
					Column:   1,
				},
				Value: match[0],
				Metadata: map[string]interface{}{
					"is_computed": true,
				},
				Children: []*types.ASTNode{
					{
						Id:   fmt.Sprintf("property-name-%s", propName),
						Type: "identifier",
						Value: propName,
						Location: types.FileLocation{
							FilePath: root.Location.FilePath,
							Line:     lineNum,
							Column:   strings.Index(match[0], propName) + 1,
						},
					},
				},
			}
			root.Children = append(root.Children, propNode)
		}
	}
	
	// Parse stored properties
	storedMatches := swiftPatterns["storedProperty"].FindAllStringSubmatch(content, -1)
	for _, match := range storedMatches {
		if len(match) > 1 {
			propName := match[1]
			lineNum := strings.Count(content[:strings.Index(content, match[0])], "\n") + 1
			
			propNode := &types.ASTNode{
				Id:   fmt.Sprintf("property-%s-%d", propName, lineNum),
				Type: "property_declaration",
				Location: types.FileLocation{
					FilePath: root.Location.FilePath,
					Line:     lineNum,
					Column:   1,
				},
				Value: match[0],
				Metadata: map[string]interface{}{
					"is_stored": true,
				},
				Children: []*types.ASTNode{
					{
						Id:   fmt.Sprintf("property-name-%s", propName),
						Type: "identifier",
						Value: propName,
						Location: types.FileLocation{
							FilePath: root.Location.FilePath,
							Line:     lineNum,
							Column:   strings.Index(match[0], propName) + 1,
						},
					},
				},
			}
			root.Children = append(root.Children, propNode)
		}
	}
	
	// Parse basic property declarations (fallback)
	propMatches := swiftPatterns["propertyDeclaration"].FindAllStringSubmatch(content, -1)
	for _, match := range propMatches {
		if len(match) > 1 {
			propName := match[1]
			lineNum := strings.Count(content[:strings.Index(content, match[0])], "\n") + 1
			
			// Skip if already processed by more specific patterns
			fullMatch := match[0]
			if strings.Contains(fullMatch, "@") || 
			   strings.Contains(fullMatch, "=") ||
			   strings.Contains(fullMatch, "{") {
				continue
			}
			
			propNode := &types.ASTNode{
				Id:   fmt.Sprintf("property-%s-%d", propName, lineNum),
				Type: "property_declaration",
				Location: types.FileLocation{
					FilePath: root.Location.FilePath,
					Line:     lineNum,
					Column:   1,
				},
				Value: match[0],
				Children: []*types.ASTNode{
					{
						Id:   fmt.Sprintf("property-name-%s", propName),
						Type: "identifier",
						Value: propName,
						Location: types.FileLocation{
							FilePath: root.Location.FilePath,
							Line:     lineNum,
							Column:   strings.Index(match[0], propName) + 1,
						},
					},
				},
			}
			root.Children = append(root.Children, propNode)
		}
	}
}

// parseSwiftExtensions extracts extension declarations
func (m *Manager) parseSwiftExtensions(content string, root *types.ASTNode) {
	matches := swiftPatterns["extension"].FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			extensionName := match[1]
			lineNum := strings.Count(content[:strings.Index(content, match[0])], "\n") + 1
			
			extensionNode := &types.ASTNode{
				Id:   fmt.Sprintf("extension-%s-%d", extensionName, lineNum),
				Type: "extension_declaration",
				Location: types.FileLocation{
					FilePath: root.Location.FilePath,
					Line:     lineNum,
					Column:   1,
				},
				Value: match[0],
			}
			root.Children = append(root.Children, extensionNode)
		}
	}
}

// parseSwiftImports extracts import statements
func (m *Manager) parseSwiftImports(content string, root *types.ASTNode) {
	matches := swiftPatterns["import"].FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			importName := match[1]
			lineNum := strings.Count(content[:strings.Index(content, match[0])], "\n") + 1
			
			importNode := &types.ASTNode{
				Id:   fmt.Sprintf("import-%s-%d", importName, lineNum),
				Type: "import_declaration",
				Location: types.FileLocation{
					FilePath: root.Location.FilePath,
					Line:     lineNum,
					Column:   1,
				},
				Value: match[0],
				Children: []*types.ASTNode{
					{
						Id:   fmt.Sprintf("import-name-%s", importName),
						Type: "identifier",
						Value: importName,
						Location: types.FileLocation{
							FilePath: root.Location.FilePath,
							Line:     lineNum,
							Column:   8, // After "import "
						},
					},
				},
			}
			root.Children = append(root.Children, importNode)
		}
	}
}

// detectSwiftFrameworks adds framework detection metadata
func (m *Manager) detectSwiftFrameworks(content string, root *types.ASTNode) {
	hasSwiftUI := strings.Contains(content, "import SwiftUI")
	hasUIKit := strings.Contains(content, "import UIKit")
	hasVapor := strings.Contains(content, "import Vapor")
	hasCombine := strings.Contains(content, "import Combine")
	hasSwiftData := strings.Contains(content, "import SwiftData")
	hasSwiftTesting := strings.Contains(content, "import Testing")
	hasTCA := strings.Contains(content, "import ComposableArchitecture") || strings.Contains(content, "import TCA")
	hasFoundation := strings.Contains(content, "import Foundation") || hasSwiftUI || hasSwiftData // SwiftUI and SwiftData implicitly import Foundation
	
	root.Metadata["has_swiftui"] = hasSwiftUI
	root.Metadata["has_uikit"] = hasUIKit
	root.Metadata["has_vapor"] = hasVapor
	root.Metadata["has_combine"] = hasCombine
	root.Metadata["has_swiftdata"] = hasSwiftData
	root.Metadata["has_swift_testing"] = hasSwiftTesting
	root.Metadata["has_tca"] = hasTCA
	root.Metadata["has_foundation"] = hasFoundation
}

// parseSwiftActors extracts actor declarations
func (m *Manager) parseSwiftActors(content string, root *types.ASTNode) {
	matches := swiftPatterns["actor"].FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			actorName := match[1]
			lineNum := strings.Count(content[:strings.Index(content, match[0])], "\n") + 1
			
			actorNode := &types.ASTNode{
				Id:   fmt.Sprintf("actor-%s-%d", actorName, lineNum),
				Type: "actor_declaration",
				Location: types.FileLocation{
					FilePath: root.Location.FilePath,
					Line:     lineNum,
					Column:   1,
				},
				Value: match[0],
				Metadata: map[string]interface{}{
					"is_actor": true,
				},
				Children: []*types.ASTNode{
					{
						Id:   fmt.Sprintf("actor-name-%s", actorName),
						Type: "identifier",
						Value: actorName,
						Location: types.FileLocation{
							FilePath: root.Location.FilePath,
							Line:     lineNum,
							Column:   strings.Index(match[0], actorName) + 1,
						},
					},
				},
			}
			root.Children = append(root.Children, actorNode)
		}
	}
}

// parseSwiftTypealias extracts typealias declarations
func (m *Manager) parseSwiftTypealias(content string, root *types.ASTNode) {
	matches := swiftPatterns["typealias"].FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			typealiasName := match[1]
			lineNum := strings.Count(content[:strings.Index(content, match[0])], "\n") + 1
			
			typealiasNode := &types.ASTNode{
				Id:   fmt.Sprintf("typealias-%s-%d", typealiasName, lineNum),
				Type: "typealias_declaration",
				Location: types.FileLocation{
					FilePath: root.Location.FilePath,
					Line:     lineNum,
					Column:   1,
				},
				Value: match[0],
				Children: []*types.ASTNode{
					{
						Id:   fmt.Sprintf("typealias-name-%s", typealiasName),
						Type: "identifier",
						Value: typealiasName,
						Location: types.FileLocation{
							FilePath: root.Location.FilePath,
							Line:     lineNum,
							Column:   strings.Index(match[0], typealiasName) + 1,
						},
					},
				},
			}
			root.Children = append(root.Children, typealiasNode)
		}
	}
}

// parseSwiftClosures adds closure detection metadata
func (m *Manager) parseSwiftClosures(content string, root *types.ASTNode) {
	closureCount := len(swiftPatterns["closure"].FindAllString(content, -1))
	trailingClosureCount := len(swiftPatterns["trailingClosure"].FindAllString(content, -1))
	escapingClosureCount := len(swiftPatterns["closureParameter"].FindAllString(content, -1))
	
	if closureCount > 0 || trailingClosureCount > 0 || escapingClosureCount > 0 {
		root.Metadata["has_closures"] = true
		root.Metadata["closure_count"] = closureCount
		root.Metadata["trailing_closure_count"] = trailingClosureCount
		root.Metadata["escaping_closure_count"] = escapingClosureCount
	}
}

// parseSwiftAsyncAwait adds async/await pattern detection metadata
func (m *Manager) parseSwiftAsyncAwait(content string, root *types.ASTNode) {
	// Parse async functions
	asyncFuncs := swiftPatterns["asyncFunction"].FindAllStringSubmatch(content, -1)
	asyncFuncCount := len(asyncFuncs)
	
	// Parse async properties
	asyncProps := swiftPatterns["asyncProperty"].FindAllStringSubmatch(content, -1)
	asyncPropCount := len(asyncProps)
	
	// Parse await calls
	awaitCalls := swiftPatterns["awaitCall"].FindAllString(content, -1)
	awaitCallCount := len(awaitCalls)
	
	if asyncFuncCount > 0 || asyncPropCount > 0 || awaitCallCount > 0 {
		root.Metadata["has_async_await"] = true
		root.Metadata["async_function_count"] = asyncFuncCount
		root.Metadata["async_property_count"] = asyncPropCount
		root.Metadata["await_call_count"] = awaitCallCount
	}
	
	// Add async function nodes
	for _, match := range asyncFuncs {
		if len(match) > 1 {
			funcName := match[1]
			lineNum := strings.Count(content[:strings.Index(content, match[0])], "\n") + 1
			
			asyncFuncNode := &types.ASTNode{
				Id:   fmt.Sprintf("async-func-%s-%d", funcName, lineNum),
				Type: "async_function_declaration",
				Location: types.FileLocation{
					FilePath: root.Location.FilePath,
					Line:     lineNum,
					Column:   1,
				},
				Value: match[0],
				Metadata: map[string]interface{}{
					"is_async": true,
				},
			}
			root.Children = append(root.Children, asyncFuncNode)
		}
	}
}

// parseSwiftOptionals adds optional pattern detection metadata
func (m *Manager) parseSwiftOptionals(content string, root *types.ASTNode) {
	optionalChainCount := len(swiftPatterns["optionalChaining"].FindAllString(content, -1))
	optionalBindCount := len(swiftPatterns["optionalBinding"].FindAllString(content, -1))
	nilCoalescingCount := len(swiftPatterns["nilCoalescing"].FindAllString(content, -1))
	forceUnwrapCount := len(swiftPatterns["forceUnwrap"].FindAllString(content, -1))
	
	if optionalChainCount > 0 || optionalBindCount > 0 || nilCoalescingCount > 0 || forceUnwrapCount > 0 {
		root.Metadata["has_optionals"] = true
		root.Metadata["optional_chaining_count"] = optionalChainCount
		root.Metadata["optional_binding_count"] = optionalBindCount
		root.Metadata["nil_coalescing_count"] = nilCoalescingCount
		root.Metadata["force_unwrap_count"] = forceUnwrapCount
	}
}

// parseSwiftControlFlow adds control flow pattern detection metadata
func (m *Manager) parseSwiftControlFlow(content string, root *types.ASTNode) {
	guardStatements := swiftPatterns["guardStatement"].FindAllString(content, -1)
	deferStatements := swiftPatterns["deferStatement"].FindAllString(content, -1)
	
	guardCount := len(guardStatements)
	deferCount := len(deferStatements)
	
	if guardCount > 0 || deferCount > 0 {
		root.Metadata["has_control_flow"] = true
		root.Metadata["guard_statement_count"] = guardCount
		root.Metadata["defer_statement_count"] = deferCount
	}
	
	// Add guard statement nodes
	for i, guardStmt := range guardStatements {
		lineNum := strings.Count(content[:strings.Index(content, guardStmt)], "\n") + 1
		
		guardNode := &types.ASTNode{
			Id:   fmt.Sprintf("guard-%d-%d", i, lineNum),
			Type: "guard_statement",
			Location: types.FileLocation{
				FilePath: root.Location.FilePath,
				Line:     lineNum,
				Column:   1,
			},
			Value: guardStmt,
		}
		root.Children = append(root.Children, guardNode)
	}
	
	// Add defer statement nodes
	for i, deferStmt := range deferStatements {
		lineNum := strings.Count(content[:strings.Index(content, deferStmt)], "\n") + 1
		
		deferNode := &types.ASTNode{
			Id:   fmt.Sprintf("defer-%d-%d", i, lineNum),
			Type: "defer_statement",
			Location: types.FileLocation{
				FilePath: root.Location.FilePath,
				Line:     lineNum,
				Column:   1,
			},
			Value: deferStmt,
		}
		root.Children = append(root.Children, deferNode)
	}
}

// parseSwiftAssociatedTypes extracts associated type declarations from protocols
func (m *Manager) parseSwiftAssociatedTypes(content string, root *types.ASTNode) {
	matches := swiftPatterns["associatedType"].FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			typeName := match[1]
			lineNum := strings.Count(content[:strings.Index(content, match[0])], "\n") + 1
			
			associatedTypeNode := &types.ASTNode{
				Id:   fmt.Sprintf("associatedtype-%s-%d", typeName, lineNum),
				Type: "associatedtype_declaration",
				Location: types.FileLocation{
					FilePath: root.Location.FilePath,
					Line:     lineNum,
					Column:   1,
				},
				Value: match[0],
				Children: []*types.ASTNode{
					{
						Id:   fmt.Sprintf("associatedtype-name-%s", typeName),
						Type: "identifier",
						Value: typeName,
						Location: types.FileLocation{
							FilePath: root.Location.FilePath,
							Line:     lineNum,
							Column:   strings.Index(match[0], typeName) + 1,
						},
					},
				},
			}
			root.Children = append(root.Children, associatedTypeNode)
		}
	}
}

// parseSwiftSubscripts extracts subscript declarations
func (m *Manager) parseSwiftSubscripts(content string, root *types.ASTNode) {
	matches := swiftPatterns["subscript"].FindAllStringSubmatch(content, -1)
	subscriptCount := len(matches)
	
	if subscriptCount > 0 {
		root.Metadata["has_subscripts"] = true
		root.Metadata["subscript_count"] = subscriptCount
	}
	
	for i, match := range matches {
		lineNum := strings.Count(content[:strings.Index(content, match[0])], "\n") + 1
		
		subscriptNode := &types.ASTNode{
			Id:   fmt.Sprintf("subscript-%d-%d", i, lineNum),
			Type: "subscript_declaration",
			Location: types.FileLocation{
				FilePath: root.Location.FilePath,
				Line:     lineNum,
				Column:   1,
			},
			Value: match[0],
		}
		root.Children = append(root.Children, subscriptNode)
	}
}

// parseSwiftOperators extracts operator overloading
func (m *Manager) parseSwiftOperators(content string, root *types.ASTNode) {
	operatorFuncs := swiftPatterns["operatorFunc"].FindAllStringSubmatch(content, -1)
	operatorDecls := swiftPatterns["operatorDecl"].FindAllStringSubmatch(content, -1)
	
	operatorFuncCount := len(operatorFuncs)
	operatorDeclCount := len(operatorDecls)
	
	if operatorFuncCount > 0 || operatorDeclCount > 0 {
		root.Metadata["has_operators"] = true
		root.Metadata["operator_function_count"] = operatorFuncCount
		root.Metadata["operator_declaration_count"] = operatorDeclCount
	}
	
	// Add operator function nodes
	for _, match := range operatorFuncs {
		if len(match) > 1 {
			operatorName := match[1]
			lineNum := strings.Count(content[:strings.Index(content, match[0])], "\n") + 1
			
			operatorNode := &types.ASTNode{
				Id:   fmt.Sprintf("operator-func-%s-%d", operatorName, lineNum),
				Type: "operator_function_declaration",
				Location: types.FileLocation{
					FilePath: root.Location.FilePath,
					Line:     lineNum,
					Column:   1,
				},
				Value: match[0],
				Metadata: map[string]interface{}{
					"operator_symbol": operatorName,
				},
			}
			root.Children = append(root.Children, operatorNode)
		}
	}
}

// parseSwiftAsyncSequences adds async sequence pattern detection
func (m *Manager) parseSwiftAsyncSequences(content string, root *types.ASTNode) {
	asyncSeqCount := len(swiftPatterns["asyncSequence"].FindAllString(content, -1))
	asyncIterCount := len(swiftPatterns["asyncIterator"].FindAllString(content, -1))
	
	if asyncSeqCount > 0 || asyncIterCount > 0 {
		root.Metadata["has_async_sequences"] = true
		root.Metadata["async_sequence_count"] = asyncSeqCount
		root.Metadata["async_iterator_count"] = asyncIterCount
	}
}

// parseSwiftResultBuilders adds result builder pattern detection
func (m *Manager) parseSwiftResultBuilders(content string, root *types.ASTNode) {
	resultBuilders := swiftPatterns["resultBuilder"].FindAllStringSubmatch(content, -1)
	viewBuilders := swiftPatterns["viewBuilder"].FindAllStringSubmatch(content, -1)
	functionBuilders := swiftPatterns["functionBuilder"].FindAllStringSubmatch(content, -1)
	
	resultBuilderCount := len(resultBuilders)
	viewBuilderCount := len(viewBuilders)
	functionBuilderCount := len(functionBuilders)
	
	if resultBuilderCount > 0 || viewBuilderCount > 0 || functionBuilderCount > 0 {
		root.Metadata["has_result_builders"] = true
		root.Metadata["result_builder_count"] = resultBuilderCount
		root.Metadata["view_builder_count"] = viewBuilderCount
		root.Metadata["function_builder_count"] = functionBuilderCount
	}
	
	// Add result builder nodes
	for _, match := range resultBuilders {
		if len(match) > 1 {
			builderName := match[1]
			lineNum := strings.Count(content[:strings.Index(content, match[0])], "\n") + 1
			
			builderNode := &types.ASTNode{
				Id:   fmt.Sprintf("result-builder-%s-%d", builderName, lineNum),
				Type: "result_builder_declaration",
				Location: types.FileLocation{
					FilePath: root.Location.FilePath,
					Line:     lineNum,
					Column:   1,
				},
				Value: match[0],
			}
			root.Children = append(root.Children, builderNode)
		}
	}
}

// parseSwiftMacros adds macro system detection (Swift 5.9+)
func (m *Manager) parseSwiftMacros(content string, root *types.ASTNode) {
	macroDecls := swiftPatterns["macroDecl"].FindAllStringSubmatch(content, -1)
	macroUsages := swiftPatterns["macroUsage"].FindAllStringSubmatch(content, -1)
	
	macroDeclCount := len(macroDecls)
	macroUsageCount := len(macroUsages)
	
	if macroDeclCount > 0 || macroUsageCount > 0 {
		root.Metadata["has_macros"] = true
		root.Metadata["macro_declaration_count"] = macroDeclCount
		root.Metadata["macro_usage_count"] = macroUsageCount
	}
	
	// Add macro declaration nodes
	for _, match := range macroDecls {
		if len(match) > 1 {
			macroName := match[1]
			lineNum := strings.Count(content[:strings.Index(content, match[0])], "\n") + 1
			
			macroNode := &types.ASTNode{
				Id:   fmt.Sprintf("macro-%s-%d", macroName, lineNum),
				Type: "macro_declaration",
				Location: types.FileLocation{
					FilePath: root.Location.FilePath,
					Line:     lineNum,
					Column:   1,
				},
				Value: match[0],
			}
			root.Children = append(root.Children, macroNode)
		}
	}
}
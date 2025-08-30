package parser

import (
	"testing"
	
	"github.com/nuthan-ms/codecontext/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSwiftSimpleParsing(t *testing.T) {
	manager := NewManager()
	
	// Test simple class parsing
	t.Run("simple class", func(t *testing.T) {
		swiftCode := `class MyClass {
    func myMethod() -> String {
        return "hello"
    }
}`
		
		ast, err := manager.parseContent(swiftCode, types.Language{
			Name: "swift",
			Extensions: []string{".swift"},
			Parser: "tree-sitter-swift",
			Enabled: true,
		}, "test.swift")
		require.NoError(t, err)
		require.NotNil(t, ast)
		assert.Equal(t, "swift", ast.Language)
		assert.Equal(t, "test.swift", ast.FilePath)
		
		// Extract symbols
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		
		// Should have at least 2 symbols (class and method)
		assert.GreaterOrEqual(t, len(symbols), 2)
		
		// Find the class symbol
		var classSymbol *types.Symbol
		var methodSymbol *types.Symbol
		for _, symbol := range symbols {
			if symbol.Name == "MyClass" {
				classSymbol = symbol
			}
			if symbol.Name == "myMethod" {
				methodSymbol = symbol
			}
		}
		
		require.NotNil(t, classSymbol, "Should find MyClass symbol")
		assert.Equal(t, "MyClass", classSymbol.Name)
		assert.Equal(t, types.SymbolTypeClass, classSymbol.Type)
		
		require.NotNil(t, methodSymbol, "Should find myMethod symbol")
		assert.Equal(t, "myMethod", methodSymbol.Name)
		assert.Equal(t, types.SymbolTypeMethod, methodSymbol.Type)
	})
	
	// Test struct parsing
	t.Run("simple struct", func(t *testing.T) {
		swiftCode := `struct Person {
    let name: String
    var age: Int
    
    init(name: String, age: Int) {
        self.name = name
        self.age = age
    }
}`
		
		ast, err := manager.parseContent(swiftCode, types.Language{
			Name: "swift",
			Extensions: []string{".swift"},
			Parser: "tree-sitter-swift",
			Enabled: true,
		}, "person.swift")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		
		// Should find struct, properties, and initializer
		assert.GreaterOrEqual(t, len(symbols), 3)
		
		var structSymbol *types.Symbol
		for _, symbol := range symbols {
			if symbol.Name == "Person" {
				structSymbol = symbol
				break
			}
		}
		
		require.NotNil(t, structSymbol, "Should find Person struct")
		assert.Equal(t, "Person", structSymbol.Name)
		assert.Equal(t, types.SymbolTypeClass, structSymbol.Type) // Structs map to class type
	})
	
	// Test protocol parsing
	t.Run("simple protocol", func(t *testing.T) {
		swiftCode := `protocol Drawable {
    func draw()
    var color: String { get }
}`
		
		ast, err := manager.parseContent(swiftCode, types.Language{
			Name: "swift",
			Extensions: []string{".swift"},
			Parser: "tree-sitter-swift",
			Enabled: true,
		}, "drawable.swift")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find protocol and method declarations
		assert.GreaterOrEqual(t, len(symbols), 1)
		
		var protocolSymbol *types.Symbol
		for _, symbol := range symbols {
			if symbol.Name == "Drawable" {
				protocolSymbol = symbol
				break
			}
		}
		
		require.NotNil(t, protocolSymbol, "Should find Drawable protocol")
		assert.Equal(t, "Drawable", protocolSymbol.Name)
		assert.Equal(t, types.SymbolTypeInterface, protocolSymbol.Type)
	})
	
	// Test import parsing
	t.Run("imports", func(t *testing.T) {
		swiftCode := `import Foundation
import UIKit
import SwiftUI

class ViewController: UIViewController {}`
		
		ast, err := manager.parseContent(swiftCode, types.Language{
			Name: "swift",
			Extensions: []string{".swift"},
			Parser: "tree-sitter-swift",
			Enabled: true,
		}, "controller.swift")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		
		// Should find imports and class
		assert.GreaterOrEqual(t, len(symbols), 4)
		
		var foundImports []string
		for _, symbol := range symbols {
			if symbol.Type == types.SymbolTypeImport {
				foundImports = append(foundImports, symbol.Name)
			}
		}
		
		assert.Contains(t, foundImports, "Foundation")
		assert.Contains(t, foundImports, "UIKit")
		assert.Contains(t, foundImports, "SwiftUI")
	})
}

func TestSwiftLanguageDetection(t *testing.T) {
	manager := NewManager()
	
	t.Run("swift file extension", func(t *testing.T) {
		lang := manager.detectLanguage("test.swift")
		require.NotNil(t, lang, "Should detect Swift language")
		assert.Equal(t, "swift", lang.Name)
		assert.Contains(t, lang.Extensions, ".swift")
		assert.Equal(t, "tree-sitter-swift", lang.Parser)
		assert.True(t, lang.Enabled)
	})
	
	t.Run("non-swift file", func(t *testing.T) {
		lang := manager.detectLanguage("test.py")
		require.NotNil(t, lang, "Should detect Python, not Swift")
		assert.NotEqual(t, "swift", lang.Name)
	})
}
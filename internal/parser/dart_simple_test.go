package parser

import (
	"testing"
	
	"github.com/nuthan-ms/codecontext/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDartSimpleParsing(t *testing.T) {
	manager := NewManager()
	
	// Test simple class parsing
	t.Run("simple class only", func(t *testing.T) {
		dartCode := `class MyClass {}`
		
		ast, err := manager.parseDartContent(dartCode, "test.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		assert.Equal(t, "dart", ast.Language)
		assert.Equal(t, "test.dart", ast.FilePath)
		
		// Check root node
		require.NotNil(t, ast.Root)
		assert.Equal(t, "compilation_unit", ast.Root.Type)
		assert.NotNil(t, ast.Root.Metadata)
		
		// Extract symbols
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		t.Logf("Found %d symbols", len(symbols))
		for i, symbol := range symbols {
			t.Logf("Symbol %d: Name=%s, Type=%s", i, symbol.Name, symbol.Type)
		}
		
		// Should have at least one symbol (the class)
		assert.GreaterOrEqual(t, len(symbols), 1)
		
		// Find the class symbol
		var classSymbol *types.Symbol
		for _, symbol := range symbols {
			if symbol.Name == "MyClass" {
				classSymbol = symbol
				break
			}
		}
		
		require.NotNil(t, classSymbol, "Should find MyClass symbol")
		assert.Equal(t, "MyClass", classSymbol.Name)
		assert.Equal(t, types.SymbolTypeClass, classSymbol.Type)
	})
}

func TestDartSimpleFlutterDetection(t *testing.T) {
	manager := NewManager()
	
	t.Run("flutter import detection", func(t *testing.T) {
		dartCode := `import 'package:flutter/material.dart';`
		
		ast, err := manager.parseDartContent(dartCode, "test.dart")
		require.NoError(t, err)
		require.NotNil(t, ast.Root)
		require.NotNil(t, ast.Root.Metadata)
		
		hasFlutter, exists := ast.Root.Metadata["has_flutter"]
		require.True(t, exists, "Should have flutter detection metadata")
		assert.True(t, hasFlutter.(bool), "Should detect Flutter import")
	})
	
	t.Run("non-flutter code", func(t *testing.T) {
		dartCode := `class MyClass { void method() {} }`
		
		ast, err := manager.parseDartContent(dartCode, "test.dart")
		require.NoError(t, err)
		require.NotNil(t, ast.Root)
		require.NotNil(t, ast.Root.Metadata)
		
		hasFlutter, exists := ast.Root.Metadata["has_flutter"]
		require.True(t, exists, "Should have flutter detection metadata")
		assert.False(t, hasFlutter.(bool), "Should not detect Flutter for plain Dart")
	})
}
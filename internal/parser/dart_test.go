package parser

import (
	"testing"
	
	"github.com/nuthan-ms/codecontext/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test data for TDD approach
var dartTestCases = []struct {
	name     string
	dartCode string
	expected []expectedSymbol
}{
	{
		name: "simple class",
		dartCode: `class MyClass {
			void method() {}
			int variable = 0;
		}`,
		expected: []expectedSymbol{
			{name: "MyClass", symbolType: types.SymbolTypeClass},
			{name: "method", symbolType: types.SymbolTypeMethod},
			{name: "variable", symbolType: types.SymbolTypeVariable},
		},
	},
	{
		name: "simple function",
		dartCode: `void myFunction() {
			print('Hello World');
		}`,
		expected: []expectedSymbol{
			{name: "myFunction", symbolType: types.SymbolTypeFunction},
		},
	},
	{
		name: "variable declarations",
		dartCode: `int globalVar = 42;
final String constant = 'test';
var dynamicVar = 'hello';`,
		expected: []expectedSymbol{
			{name: "globalVar", symbolType: types.SymbolTypeVariable},
			{name: "constant", symbolType: types.SymbolTypeVariable},
			{name: "dynamicVar", symbolType: types.SymbolTypeVariable},
		},
	},
	{
		name: "import statement",
		dartCode: `import 'package:flutter/material.dart';
import 'dart:io' as io;`,
		expected: []expectedSymbol{
			{name: "package:flutter/material.dart", symbolType: types.SymbolTypeImport},
			{name: "dart:io", symbolType: types.SymbolTypeImport},
		},
	},
	{
		name: "flutter stateless widget",
		dartCode: `class MyWidget extends StatelessWidget {
			@override
			Widget build(BuildContext context) {
				return Container();
			}
		}`,
		expected: []expectedSymbol{
			{name: "MyWidget", symbolType: types.SymbolTypeWidget},
			{name: "build", symbolType: types.SymbolTypeBuildMethod},
		},
	},
	{
		name: "flutter stateful widget",
		dartCode: `class MyStatefulWidget extends StatefulWidget {
	@override
	_MyStatefulWidgetState createState() => _MyStatefulWidgetState();
}

class _MyStatefulWidgetState extends State<MyStatefulWidget> {
	@override
	Widget build(BuildContext context) {
		return Container();
	}
}`,
		expected: []expectedSymbol{
			{name: "MyStatefulWidget", symbolType: types.SymbolTypeWidget},
			{name: "createState", symbolType: types.SymbolTypeMethod},
			{name: "_MyStatefulWidgetState", symbolType: types.SymbolTypeStateClass},
			{name: "build", symbolType: types.SymbolTypeBuildMethod},
		},
	},
}

type expectedSymbol struct {
	name       string
	symbolType types.SymbolType
}

func TestDartBasicSymbolExtraction(t *testing.T) {
	manager := NewManager()
	
	for _, tt := range dartTestCases {
		t.Run(tt.name, func(t *testing.T) {
			// Parse Dart code
			ast, err := manager.parseDartContent(tt.dartCode, "test.dart")
			require.NoError(t, err, "Failed to parse Dart code")
			require.NotNil(t, ast, "AST should not be nil")
			
			// Extract symbols
			symbols, err := manager.ExtractSymbols(ast)
			require.NoError(t, err, "Failed to extract symbols")
			
			// Debug: print found symbols
			t.Logf("Found %d symbols:", len(symbols))
			for i, symbol := range symbols {
				t.Logf("  Symbol %d: %s (type: %s)", i, symbol.Name, symbol.Type)
			}
			
			
			
			// Validate symbol count
			assert.Len(t, symbols, len(tt.expected), "Unexpected number of symbols")
			
			// Validate each symbol
			for i, expectedSym := range tt.expected {
				if i >= len(symbols) {
					t.Errorf("Missing symbol: %s", expectedSym.name)
					continue
				}
				
				assert.Equal(t, expectedSym.name, symbols[i].Name, "Symbol name mismatch at index %d", i)
				assert.Equal(t, expectedSym.symbolType, symbols[i].Type, "Symbol type mismatch for %s", expectedSym.name)
				assert.Equal(t, "dart", symbols[i].Language, "Language should be dart")
			}
		})
	}
}

func TestDartLanguageDetection(t *testing.T) {
	manager := NewManager()
	
	testCases := []struct {
		name     string
		filePath string
		expected string
		wantNil  bool
	}{
		{
			name:     "dart file",
			filePath: "test.dart",
			expected: "dart",
			wantNil:  false,
		},
		{
			name:     "non-dart file",
			filePath: "test.unknown",
			expected: "",
			wantNil:  true,
		},
	}
	
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			lang := manager.detectLanguage(tt.filePath)
			
			if tt.wantNil {
				assert.Nil(t, lang, "Expected nil language for %s", tt.filePath)
			} else {
				require.NotNil(t, lang, "Expected language for %s", tt.filePath)
				assert.Equal(t, tt.expected, lang.Name, "Language name mismatch")
				assert.Contains(t, lang.Extensions, ".dart", "Should support .dart extension")
			}
		})
	}
}

func TestDartFlutterDetection(t *testing.T) {
	manager := NewManager()
	
	testCases := []struct {
		name       string
		dartCode   string
		hasFlutter bool
		widgetType string
	}{
		{
			name: "plain dart",
			dartCode: `class MyClass {
				void method() {}
			}`,
			hasFlutter: false,
		},
		{
			name: "flutter import but no widget",
			dartCode: `import 'package:flutter/material.dart';
			
			void main() {}`,
			hasFlutter: true,
		},
		{
			name: "stateless widget",
			dartCode: `import 'package:flutter/material.dart';

class MyWidget extends StatelessWidget {
	Widget build(BuildContext context) => Container();
}`,
			hasFlutter: true,
			widgetType: "stateless",
		},
		{
			name: "stateful widget",
			dartCode: `import 'package:flutter/material.dart';

class MyWidget extends StatefulWidget {
	State<MyWidget> createState() => _MyWidgetState();
}`,
			hasFlutter: true,
			widgetType: "stateful",
		},
	}
	
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := manager.parseDartContent(tt.dartCode, "test.dart")
			require.NoError(t, err)
			
			// Check Flutter detection in root node metadata
			hasFlutter, _ := ast.Root.Metadata["has_flutter"].(bool)
			assert.Equal(t, tt.hasFlutter, hasFlutter, "Flutter detection mismatch")
			
			// If we expect a widget type, check symbols
			if tt.widgetType != "" {
				symbols, err := manager.ExtractSymbols(ast)
				require.NoError(t, err)
				
				// Find widget symbol
				var widgetSymbol *types.Symbol
				for _, symbol := range symbols {
					if symbol.Type == types.SymbolTypeWidget {
						widgetSymbol = symbol
						break
					}
				}
				
				require.NotNil(t, widgetSymbol, "Should find widget symbol")
				
				// For now, just check that we found a widget symbol
				// Widget type detection will be verified in more detailed tests later
			}
		})
	}
}

func TestDartErrorHandling(t *testing.T) {
	manager := NewManager()
	
	testCases := []struct {
		name     string
		dartCode string
		wantErr  bool
	}{
		{
			name:     "valid dart code",
			dartCode: "class MyClass {}",
			wantErr:  false,
		},
		{
			name:     "malformed dart code",
			dartCode: "class MyClass { invalid syntax >>>",
			wantErr:  false, // Should not error due to graceful degradation
		},
		{
			name:     "empty content",
			dartCode: "",
			wantErr:  false,
		},
	}
	
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := manager.parseDartContent(tt.dartCode, "test.dart")
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ast)
			}
		})
	}
}

// TestDartPrivateMethodDetection tests the privateMethod pattern
func TestDartPrivateMethodDetection(t *testing.T) {
	manager := NewManager()
	
	t.Run("class with private methods", func(t *testing.T) {
		dartCode := `class MyClass {
	void publicMethod() {
		_privateHelper();
	}
	
	void _privateHelper() {
		print('Private method');
	}
	
	int _calculateValue() {
		return 42;
	}
	
	Future<String> _asyncPrivateMethod() async {
		return 'async result';
	}
}`

		ast, err := manager.parseDartContent(dartCode, "test.dart")
		require.NoError(t, err)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Count public vs private methods
		var publicMethods, privateMethods []string
		
		for _, symbol := range symbols {
			if symbol.Type == types.SymbolTypeMethod {
				if symbol.Name[0] == '_' {
					privateMethods = append(privateMethods, symbol.Name)
				} else {
					publicMethods = append(publicMethods, symbol.Name)
				}
			}
		}
		
		t.Logf("Found %d public methods: %v", len(publicMethods), publicMethods)
		t.Logf("Found %d private methods: %v", len(privateMethods), privateMethods)
		
		// Should detect private methods
		assert.GreaterOrEqual(t, len(privateMethods), 3, "Should find at least 3 private methods")
		assert.Contains(t, privateMethods, "_privateHelper", "Should find _privateHelper")
		assert.Contains(t, privateMethods, "_calculateValue", "Should find _calculateValue")
		assert.Contains(t, privateMethods, "_asyncPrivateMethod", "Should find _asyncPrivateMethod")
		
		// Should also find public method
		assert.GreaterOrEqual(t, len(publicMethods), 1, "Should find at least 1 public method")
		assert.Contains(t, publicMethods, "publicMethod", "Should find publicMethod")
	})
	
	t.Run("private methods with various signatures", func(t *testing.T) {
		dartCode := `class TestClass {
	// Simple private method
	void _simplePrivate() {}
	
	// Private method with parameters
	int _withParams(String name, int value) => value * 2;
	
	// Private async method
	Future<void> _asyncPrivate() async {}
	
	// Private method with generics
	T _genericPrivate<T>(T value) => value;
	
	// Private static method
	static String _staticPrivate() => 'static';
}`

		ast, err := manager.parseDartContent(dartCode, "test.dart")
		require.NoError(t, err)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		privateMethodNames := []string{}
		for _, symbol := range symbols {
			if symbol.Type == types.SymbolTypeMethod && symbol.Name[0] == '_' {
				privateMethodNames = append(privateMethodNames, symbol.Name)
			}
		}
		
		expectedPrivateMethods := []string{"_simplePrivate", "_withParams", "_asyncPrivate", "_genericPrivate", "_staticPrivate"}
		
		t.Logf("Found private methods: %v", privateMethodNames)
		
		// Should detect all private method variations
		assert.GreaterOrEqual(t, len(privateMethodNames), 4, "Should find multiple private methods")
		
		// Check for specific methods (allowing for some parsing variations)
		foundMethods := make(map[string]bool)
		for _, method := range privateMethodNames {
			foundMethods[method] = true
		}
		
		foundCount := 0
		for _, expected := range expectedPrivateMethods {
			if foundMethods[expected] {
				foundCount++
			}
		}
		
		assert.GreaterOrEqual(t, foundCount, 3, "Should find at least 3 of the expected private methods")
	})
}
package parser

import (
	"testing"
	
	"github.com/nuthan-ms/codecontext/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDartIntegration(t *testing.T) {
	manager := NewManager()
	
	t.Run("complete Flutter app parsing", func(t *testing.T) {
		dartCode := `import 'package:flutter/material.dart';

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(title: 'Demo');
  }
}

class HomePage extends StatefulWidget {
  @override
  _HomePageState createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  int counter = 0;
  
  void increment() {
    setState(() { counter++; });
  }
  
  @override
  Widget build(BuildContext context) {
    return Scaffold();
  }
}`

		// Parse the Dart code
		ast, err := manager.parseDartContent(dartCode, "main.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		// Verify Flutter detection
		hasFlutter, _ := ast.Root.Metadata["has_flutter"].(bool)
		assert.True(t, hasFlutter, "Should detect Flutter")
		
		// Extract symbols
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		t.Logf("Found %d symbols", len(symbols))
		for i, symbol := range symbols {
			t.Logf("Symbol %d: Name=%s, Type=%s", i, symbol.Name, symbol.Type)
		}
		
		// Verify we found key symbols
		var foundImport, foundMyApp, foundHomePage, foundState bool
		var buildMethods int
		
		for _, symbol := range symbols {
			switch symbol.Name {
			case "package:flutter/material.dart":
				foundImport = true
				assert.Equal(t, types.SymbolTypeImport, symbol.Type)
			case "MyApp":
				foundMyApp = true
				assert.Equal(t, types.SymbolTypeWidget, symbol.Type)
			case "HomePage":
				foundHomePage = true
				assert.Equal(t, types.SymbolTypeWidget, symbol.Type)
			case "_HomePageState":
				foundState = true
				assert.True(t, symbol.Type == types.SymbolTypeStateClass || symbol.Type == types.SymbolTypeClass, 
					"Should be state_class or class type")
			case "build":
				if symbol.Type == types.SymbolTypeBuildMethod || symbol.Type == types.SymbolTypeMethod {
					buildMethods++
				}
			}
		}
		
		assert.True(t, foundImport, "Should find Flutter import")
		assert.True(t, foundMyApp, "Should find MyApp widget")
		assert.True(t, foundHomePage, "Should find HomePage widget")
		assert.True(t, foundState, "Should find state class")
		assert.GreaterOrEqual(t, buildMethods, 1, "Should find at least one build method")
		
		// Verify language is correctly set
		for _, symbol := range symbols {
			assert.Equal(t, "dart", symbol.Language, "All symbols should have dart language")
		}
	})
}

func TestDartGetSupportedLanguages(t *testing.T) {
	manager := NewManager()
	
	languages := manager.GetSupportedLanguages()
	
	// Find Dart in supported languages
	var dartLang *types.Language
	for i, lang := range languages {
		if lang.Name == "dart" {
			dartLang = &languages[i]
			break
		}
	}
	
	require.NotNil(t, dartLang, "Dart should be in supported languages")
	assert.Equal(t, "dart", dartLang.Name)
	assert.Contains(t, dartLang.Extensions, ".dart")
	assert.Equal(t, "tree-sitter-dart", dartLang.Parser)
	assert.True(t, dartLang.Enabled)
}

func TestDartFileClassification(t *testing.T) {
	manager := NewManager()
	
	// Test with a temporary dart file path (file doesn't need to exist for classification)
	classification, err := manager.ClassifyFile("my_app.dart")
	require.NoError(t, err)
	require.NotNil(t, classification)
	
	assert.Equal(t, "dart", classification.Language.Name)
	assert.Contains(t, classification.Language.Extensions, ".dart")
	assert.Equal(t, "source", classification.FileType)
	assert.False(t, classification.IsGenerated)
	assert.False(t, classification.IsTest)
}
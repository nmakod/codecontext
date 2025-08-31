package parser

import (
	"strings"
	"testing"
	
	"github.com/nuthan-ms/codecontext/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TDD Red Phase - These tests will fail initially
func TestCppBasicParsing(t *testing.T) {
	manager := NewManager()
	
	// Test simple class parsing
	t.Run("simple class", func(t *testing.T) {
		cppCode := `class Calculator {
public:
    int add(int a, int b) {
        return a + b;
    }
private:
    int value;
};`
		
		ast, err := manager.parseContent(cppCode, types.Language{
			Name: "cpp",
			Extensions: []string{".cpp"},
			Parser: "tree-sitter-cpp",
			Enabled: true,
		}, "test.cpp")
		require.NoError(t, err)
		require.NotNil(t, ast)
		assert.Equal(t, "cpp", ast.Language)
		assert.Equal(t, "test.cpp", ast.FilePath)
		
		// Extract symbols
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Validate symbol extraction
		t.Logf("Found %d symbols", len(symbols))
		
		// Should have at least 3 symbols (class, method, variable)
		assert.GreaterOrEqual(t, len(symbols), 3)
		
		// Find the class symbol
		var classSymbol *types.Symbol
		var methodSymbol *types.Symbol
		var variableSymbol *types.Symbol
		for _, symbol := range symbols {
			if symbol.Name == "Calculator" {
				classSymbol = symbol
			}
			if symbol.Name == "add" {
				methodSymbol = symbol
			}
			if symbol.Name == "value" {
				variableSymbol = symbol
			}
		}
		
		require.NotNil(t, classSymbol, "Should find Calculator class")
		assert.Equal(t, "Calculator", classSymbol.Name)
		assert.Equal(t, types.SymbolTypeClass, classSymbol.Type)
		
		require.NotNil(t, methodSymbol, "Should find add method")
		assert.Equal(t, "add", methodSymbol.Name)
		assert.Equal(t, types.SymbolTypeMethod, methodSymbol.Type)
		
		require.NotNil(t, variableSymbol, "Should find value variable")
		assert.Equal(t, "value", variableSymbol.Name)
		assert.Equal(t, types.SymbolTypeVariable, variableSymbol.Type)
	})
	
	// Test namespace parsing
	t.Run("namespace", func(t *testing.T) {
		cppCode := `namespace math {
    class Vector {
    public:
        double x, y;
        Vector(double x, double y) : x(x), y(y) {}
    };
}`
		
		ast, err := manager.parseContent(cppCode, types.Language{
			Name: "cpp",
			Extensions: []string{".cpp"},
			Parser: "tree-sitter-cpp",
			Enabled: true,
		}, "vector.cpp")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Validate symbol extraction
		t.Logf("Found %d symbols", len(symbols))
		
		// Should find namespace and class symbols
		assert.GreaterOrEqual(t, len(symbols), 2)
		
		// Enhanced parser provides better symbol classification
		// including detecting both classes and constructors separately
		
		var namespaceSymbol *types.Symbol
		var classSymbol *types.Symbol
		for _, symbol := range symbols {
			if symbol.Name == "math" {
				namespaceSymbol = symbol
			}
			if symbol.Name == "Vector" && symbol.Type == types.SymbolTypeClass {
				classSymbol = symbol
			}
		}
		
		require.NotNil(t, namespaceSymbol, "Should find math namespace")
		assert.Equal(t, types.SymbolTypeNamespace, namespaceSymbol.Type)
		
		require.NotNil(t, classSymbol, "Should find Vector class")
		assert.Equal(t, types.SymbolTypeClass, classSymbol.Type)
	})
}

// Feature coverage calculation for Phase 1
func TestCppCoreFeatureCoverage(t *testing.T) {
	manager := NewManager()
	
	// Comprehensive C++ code sample
	cppCode := `#include <iostream>
#include <vector>

namespace utils {
    class Logger {
    public:
        Logger() = default;
        ~Logger() = default;
        
        void log(const std::string& message) {
            std::cout << message << std::endl;
        }
        
    private:
        std::vector<std::string> buffer;
    };
    
    struct Config {
        int max_size;
        bool enabled;
    };
}

int main() {
    utils::Logger logger;
    logger.log("Hello World");
    return 0;
}`
	
	ast, err := manager.parseContent(cppCode, types.Language{
		Name: "cpp",
		Extensions: []string{".cpp"},
		Parser: "tree-sitter-cpp",
		Enabled: true,
	}, "main.cpp")
	require.NoError(t, err)
	require.NotNil(t, ast)
	
	// Core features to detect
	coreFeatures := map[string]bool{
		"has_classes":      false,
		"has_structs":      false,
		"has_functions":    false,
		"has_namespaces":   false,
		"has_constructors": false,
		"has_destructors":  false,
		"has_inheritance":  false,
		"has_includes":     false,
	}
	
	// Check feature detection against AST metadata
	for feature := range coreFeatures {
		if val, exists := ast.Root.Metadata[feature]; exists && val.(bool) {
			coreFeatures[feature] = true
		}
	}
	
	// Calculate coverage
	detected := 0
	total := len(coreFeatures)
	for _, isDetected := range coreFeatures {
		if isDetected {
			detected++
		}
	}
	
	coverage := float64(detected) / float64(total) * 100
	t.Logf("Core C++ Feature Coverage: %.1f%% (%d/%d)", coverage, detected, total)
	
	// Phase 1 target: 85% core feature coverage
	assert.GreaterOrEqual(t, coverage, 85.0, "Should achieve 85%+ core feature coverage")
}

// Helper function to debug AST structure
func debugPrintASTNodes(t *testing.T, node *types.ASTNode, depth int) {
	if node == nil {
		return
	}
	
	indent := strings.Repeat("  ", depth)
	t.Logf("%sNode: %s (Type: %s) Value: %q", indent, node.Id, node.Type, 
		truncateString(node.Value, 50))
	
	for _, child := range node.Children {
		debugPrintASTNodes(t, child, depth+1)
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
package parser

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/nuthan-ms/codecontext/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCppEdgeCases covers edge cases that might not be handled properly
func TestCppEdgeCases(t *testing.T) {
	// Create a parser with test configuration
	logger := NopLogger{}
	config := DefaultConfig()
	config.Cpp.ParseTimeout = 1 * time.Second
	config.Cpp.MaxNestingDepth = 5
	config.Cpp.MaxFileSize = 1024 // Small limit for testing

	parser, err := NewCppParserWithConfig(logger, config)
	require.NoError(t, err)

	t.Run("empty content", func(t *testing.T) {
		ctx := context.Background()
		ast, err := parser.ParseContent(ctx, "", "empty.cpp")
		assert.Error(t, err)
		assert.Nil(t, ast)
		assert.Contains(t, err.Error(), "content is empty")
	})

	t.Run("nil parser", func(t *testing.T) {
		var nilParser *CppParser
		ctx := context.Background()
		ast, err := nilParser.ParseContent(ctx, "test", "nil.cpp")
		assert.Error(t, err)
		assert.Nil(t, ast)
		assert.Contains(t, err.Error(), "CppParser is nil")
	})

	t.Run("file too large", func(t *testing.T) {
		ctx := context.Background()
		largeContent := strings.Repeat("// Large file content\n", 100) // Create content larger than 1024 bytes
		ast, err := parser.ParseContent(ctx, largeContent, "large.cpp")
		assert.Error(t, err)
		assert.Nil(t, ast)
		assert.Contains(t, err.Error(), "file too large")
	})

	t.Run("malformed file path", func(t *testing.T) {
		ctx := context.Background()
		// Test with null bytes in path (security edge case)
		malformedPath := "test\x00.cpp"
		manager := NewManager()
		ast, err := manager.parseContentWithContext(ctx, "class Test{};", types.Language{
			Name:       "cpp",
			Extensions: []string{".cpp"},
			Parser:     "tree-sitter-cpp",
			Enabled:    true,
		}, malformedPath)
		assert.Error(t, err)
		assert.Nil(t, ast)
	})

	t.Run("extremely long file path", func(t *testing.T) {
		ctx := context.Background()
		// Test with excessively long path (DoS prevention)
		longPath := strings.Repeat("a", 5000) + ".cpp"
		manager := NewManager()
		ast, err := manager.parseContentWithContext(ctx, "class Test{};", types.Language{
			Name:       "cpp",
			Extensions: []string{".cpp"},
			Parser:     "tree-sitter-cpp",
			Enabled:    true,
		}, longPath)
		assert.Error(t, err)
		assert.Nil(t, ast)
		assert.Contains(t, err.Error(), "file path too long")
	})

	t.Run("directory traversal attempt", func(t *testing.T) {
		ctx := context.Background()
		// Test with directory traversal attempt
		traversalPath := "../../../etc/passwd"
		manager := NewManager()
		ast, err := manager.parseContentWithContext(ctx, "class Test{};", types.Language{
			Name:       "cpp",
			Extensions: []string{".cpp"},
			Parser:     "tree-sitter-cpp",
			Enabled:    true,
		}, traversalPath)
		assert.Error(t, err)
		assert.Nil(t, ast)
		assert.Contains(t, err.Error(), "path traversal detected")
	})

	t.Run("cancelled context", func(t *testing.T) {
		// Test parsing with already cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		ast, err := parser.ParseContent(ctx, "class Test{};", "cancelled.cpp")
		assert.Error(t, err)
		assert.Nil(t, ast)
		assert.Contains(t, err.Error(), "parsing cancelled before start")
	})

	t.Run("malformed cpp syntax", func(t *testing.T) {
		ctx := context.Background()
		malformedCode := `class Broken {
			public unclosed method(
			private: invalid syntax here }{[
		`
		ast, err := parser.ParseContent(ctx, malformedCode, "malformed.cpp")
		// Should not error (tree-sitter handles malformed syntax gracefully)
		assert.NoError(t, err)
		assert.NotNil(t, ast)
		// But should produce a tree with potential error nodes
		assert.NotNil(t, ast.TreeSitterTree)
	})

	t.Run("deeply nested structures", func(t *testing.T) {
		ctx := context.Background()
		// Create deeply nested class structure
		deepCode := "class A {\n"
		for i := 0; i < 10; i++ {
			deepCode += "class Nested" + string(rune('0'+i)) + " {\n"
		}
		deepCode += "int value;\n"
		for i := 0; i < 10; i++ {
			deepCode += "};\n"
		}
		deepCode += "};"

		ast, err := parser.ParseContent(ctx, deepCode, "deep.cpp")
		// Should handle deep nesting gracefully
		assert.NoError(t, err)
		assert.NotNil(t, ast)
	})

	t.Run("unicode and special characters", func(t *testing.T) {
		ctx := context.Background()
		unicodeCode := `// Unicode comments: 测试 🚀 ñáéíóú
class UnicodeTest {
public:
    // Method with unicode in comments
    void método_test() { /* 测试方法 */ }
    std::string unicode_string = "Hello 🌍";
};`

		ast, err := parser.ParseContent(ctx, unicodeCode, "unicode.cpp")
		assert.NoError(t, err)
		assert.NotNil(t, ast)
		assert.Equal(t, "cpp", ast.Language)
	})

	t.Run("very long lines", func(t *testing.T) {
		ctx := context.Background()
		// Create content with very long line
		longLine := "// " + strings.Repeat("This is a very long comment ", 10) // Small enough to stay under 1024 byte limit
		longLineCode := longLine + "\nclass Test {};"

		ast, err := parser.ParseContent(ctx, longLineCode, "longline.cpp")
		assert.NoError(t, err)
		assert.NotNil(t, ast)
	})
}

// TestCppParserErrorHandling tests error handling in various parser methods
func TestCppParserErrorHandling(t *testing.T) {
	logger := NopLogger{}
	parser, err := NewCppParser(logger)
	require.NoError(t, err)

	t.Run("NodeToSymbol with nil node", func(t *testing.T) {
		symbol := parser.NodeToSymbol(nil, "test.cpp", "cpp", "content", nil)
		assert.Nil(t, symbol)
	})

	t.Run("NodeToSymbolWithContext with nil context", func(t *testing.T) {
		node := &types.ASTNode{Type: "class_specifier"}
		symbol := parser.NodeToSymbolWithContext(node, nil)
		assert.Nil(t, symbol)
	})

	t.Run("NodeToSymbolWithContext with nil node", func(t *testing.T) {
		ctx := &SymbolExtractionContext{
			FilePath: "test.cpp",
			Language: "cpp",
			Content:  "content",
		}
		symbol := parser.NodeToSymbolWithContext(nil, ctx)
		assert.Nil(t, symbol)
	})
}

// TestCppConfigurationEdgeCases tests configuration validation and edge cases
func TestCppConfigurationEdgeCases(t *testing.T) {
	t.Run("nil config uses defaults", func(t *testing.T) {
		logger := NopLogger{}
		parser, err := NewCppParserWithConfig(logger, nil)
		require.NoError(t, err)
		require.NotNil(t, parser)
		require.NotNil(t, parser.config)
		assert.Equal(t, DefaultConfig().Cpp.MaxNestingDepth, parser.config.Cpp.MaxNestingDepth)
	})

	t.Run("nil logger uses NopLogger", func(t *testing.T) {
		parser, err := NewCppParser(nil)
		require.NoError(t, err)
		require.NotNil(t, parser)
		require.NotNil(t, parser.logger)
		// Should be able to call logger methods without panic
		parser.logger.Info("test message")
		parser.logger.Error("test error", nil)
	})

	t.Run("extreme configuration values", func(t *testing.T) {
		logger := NopLogger{}
		config := DefaultConfig()
		// Set extreme values
		config.Cpp.MaxNestingDepth = 0
		config.Cpp.ParseTimeout = 0
		config.Cpp.MaxFileSize = 0

		parser, err := NewCppParserWithConfig(logger, config)
		require.NoError(t, err)
		require.NotNil(t, parser)
		
		// Parser should handle extreme configs gracefully
		ctx := context.Background()
		ast, err := parser.ParseContent(ctx, "class Test{};", "extreme.cpp")
		// Depending on implementation, this might fail due to size limit
		if err != nil {
			assert.Contains(t, err.Error(), "file too large")
		} else {
			assert.NotNil(t, ast)
		}
	})

	t.Run("strict timeout enforcement enabled", func(t *testing.T) {
		logger := NopLogger{}
		config := DefaultConfig()
		config.Cpp.ParseTimeout = 30 * time.Second     // Normal timeout
		config.Cpp.StrictTimeoutEnforcement = true     // Enable strict enforcement
		config.Cpp.MaxFileSize = 10 * 1024 * 1024     // Large enough to not trigger size limit

		parser, err := NewCppParserWithConfig(logger, config)
		require.NoError(t, err)
		require.NotNil(t, parser)
		
		// Test that the configuration is set correctly
		assert.True(t, parser.config.Cpp.StrictTimeoutEnforcement)
		
		// With normal content and timeout, parsing should succeed
		ctx := context.Background()
		ast, err := parser.ParseContent(ctx, "class Test { void method() {} };", "normal_test.cpp")
		assert.NoError(t, err)
		assert.NotNil(t, ast)
	})

	t.Run("lenient timeout mode is default", func(t *testing.T) {
		logger := NopLogger{}
		config := DefaultConfig()
		// Don't set StrictTimeoutEnforcement - should default to false

		parser, err := NewCppParserWithConfig(logger, config)
		require.NoError(t, err)
		require.NotNil(t, parser)
		
		// Test that the default configuration is lenient
		assert.False(t, parser.config.Cpp.StrictTimeoutEnforcement)
		
		// Parsing should work normally in default mode
		ctx := context.Background()
		ast, err := parser.ParseContent(ctx, "class Test { void method() {} };", "default_test.cpp")
		assert.NoError(t, err)
		assert.NotNil(t, ast)
	})
}

// TestCppInputSanitization tests various input sanitization edge cases
func TestCppInputSanitization(t *testing.T) {
	manager := NewManager()
	ctx := context.Background()

	testCases := []struct {
		name        string
		filePath    string
		expectError bool
		errorContains string
	}{
		{
			name:        "normal path",
			filePath:    "normal/path/file.cpp",
			expectError: false,
		},
		{
			name:          "null byte injection",
			filePath:      "file\x00.cpp",
			expectError:   true,
			errorContains: "null bytes",
		},
		{
			name:          "path traversal double dot",
			filePath:      "../../../sensitive.cpp",
			expectError:   true,
			errorContains: "path traversal",
		},
		{
			name:          "path traversal encoded",
			filePath:      "dir/..%2F..%2Fsensitive.cpp",
			expectError:   true,
			errorContains: "path traversal",
		},
		{
			name:          "extremely long path",
			filePath:      strings.Repeat("very_long_directory_name/", 200) + "file.cpp",
			expectError:   true,
			errorContains: "too long",
		},
		{
			name:        "empty path",
			filePath:    "",
			expectError: false, // Empty path should be allowed
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := manager.parseContentWithContext(ctx, "class Test{};", types.Language{
				Name:       "cpp",
				Extensions: []string{".cpp"},
				Parser:     "tree-sitter-cpp",
				Enabled:    true,
			}, tc.filePath)

			if tc.expectError {
				assert.Error(t, err, "Expected error for path: %s", tc.filePath)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains, "Error should contain: %s", tc.errorContains)
				}
				assert.Nil(t, ast)
			} else {
				assert.NoError(t, err, "Expected no error for path: %s", tc.filePath)
				assert.NotNil(t, ast)
			}
		})
	}
}
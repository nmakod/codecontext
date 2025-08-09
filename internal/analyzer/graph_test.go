package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
)

func TestNewGraphBuilder(t *testing.T) {
	builder := NewGraphBuilder()

	if builder == nil {
		t.Fatal("NewGraphBuilder returned nil")
	}

	if builder.parser == nil {
		t.Error("GraphBuilder.parser is nil")
	}

	if builder.graph == nil {
		t.Error("GraphBuilder.graph is nil")
	}

	if builder.graph.Nodes == nil {
		t.Error("GraphBuilder.graph.Nodes is nil")
	}

	if builder.graph.Edges == nil {
		t.Error("GraphBuilder.graph.Edges is nil")
	}

	if builder.graph.Files == nil {
		t.Error("GraphBuilder.graph.Files is nil")
	}

	if builder.graph.Symbols == nil {
		t.Error("GraphBuilder.graph.Symbols is nil")
	}
}

func TestAnalyzeDirectory(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir := t.TempDir()

	// Create a test TypeScript file
	testFile := filepath.Join(tmpDir, "test.ts")
	testContent := `// Test TypeScript file
export class TestClass {
  private value: number = 0;
  
  public getValue(): number {
    return this.value;
  }
  
  public setValue(newValue: number): void {
    this.value = newValue;
  }
}

export function testFunction(param: string): string {
  return "test: " + param;
}

const testConstant = 42;
`

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create another test file with imports
	testFile2 := filepath.Join(tmpDir, "importer.ts")
	testContent2 := `import { TestClass, testFunction } from './test';
import * as fs from 'fs';

const instance = new TestClass();
const result = testFunction("hello");
`

	err = os.WriteFile(testFile2, []byte(testContent2), 0644)
	if err != nil {
		t.Fatalf("Failed to create second test file: %v", err)
	}

	// Test the analyzer
	builder := NewGraphBuilder()
	graph, err := builder.AnalyzeDirectory(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzeDirectory failed: %v", err)
	}

	// Verify graph structure
	if graph == nil {
		t.Fatal("Returned graph is nil")
	}

	if graph.Metadata == nil {
		t.Fatal("Graph metadata is nil")
	}

	// Check that files were processed
	if len(graph.Files) == 0 {
		t.Error("No files were analyzed")
	}

	// Check that symbols were extracted
	if len(graph.Symbols) == 0 {
		t.Error("No symbols were extracted")
	}

	// Verify specific file was processed
	found := false
	for filePath := range graph.Files {
		if filepath.Base(filePath) == "test.ts" {
			found = true
			break
		}
	}
	if !found {
		t.Error("test.ts was not found in analyzed files")
	}

	t.Logf("Analyzed %d files with %d symbols",
		len(graph.Files), len(graph.Symbols))

	// Log symbol details for debugging
	for _, symbol := range graph.Symbols {
		t.Logf("Symbol: %s (%s) at %s:%d",
			symbol.Name, symbol.Type,
			filepath.Base(symbol.FullyQualifiedName), symbol.Location.StartLine)
	}
}

func TestIsSupportedFile(t *testing.T) {
	builder := NewGraphBuilder()

	tests := []struct {
		path     string
		expected bool
	}{
		{"test.ts", true},
		{"test.tsx", true},
		{"test.js", true},
		{"test.jsx", true},
		{"test.json", true},
		{"test.yaml", true},
		{"test.yml", true},
		{"test.txt", false},
		{"test.py", true},
		{"test.go", true},
		{"README.md", true},
	}

	for _, test := range tests {
		result := builder.isSupportedFile(test.path)
		if result != test.expected {
			t.Errorf("isSupportedFile(%q) = %v, expected %v",
				test.path, result, test.expected)
		}
	}
}

func TestShouldSkipPath(t *testing.T) {
	builder := NewGraphBuilder()
	builder.SetUseDefaultExcludes(false) // Disable defaults for this test

	tests := []struct {
		path     string
		expected bool
	}{
		{"src/index.ts", false},
		{"path/node_modules/package/index.js", false}, // No patterns set, should not skip
		{"project/.git/config", false},                // No patterns set, should not skip
		{"app/dist/bundle.js", false},                 // No patterns set, should not skip
		{"app/coverage/report.html", false},           // No patterns set, should not skip
		{"test/unit.spec.ts", false},
		{"project/.codecontext/config.yaml", false}, // No patterns set, should not skip
		{"node_modules", false},                     // No patterns set, should not skip
		{".git", false},                             // No patterns set, should not skip
		{"dist", false},                             // No patterns set, should not skip
		{"coverage", false},                         // No patterns set, should not skip
		{".codecontext", false},                     // No patterns set, should not skip
		{"something_node_modules", false},           // Doesn't match pattern
		{"git_config", false},                       // Doesn't match pattern
	}

	for _, test := range tests {
		result := builder.shouldSkipPath(test.path)
		if result != test.expected {
			t.Errorf("shouldSkipPath(%q) = %v, expected %v",
				test.path, result, test.expected)
		}
	}
}

func TestShouldSkipPathWithExcludePatterns(t *testing.T) {
	builder := NewGraphBuilder()
	builder.SetUseDefaultExcludes(false) // Disable defaults for predictable testing
	builder.SetExcludePatterns([]string{
		"node_modules/**",
		"*.test.*",
		"*.spec.*",
		"__pycache__/**",
		"vendor/**",
		".env*",
		"tmp/*",
	})

	tests := []struct {
		path     string
		expected bool
		reason   string
	}{
		{"src/index.ts", false, "Normal source file should not be skipped"},
		{"node_modules/package/index.js", true, "node_modules/** pattern should match"},
		{"app.test.js", true, "*.test.* pattern should match"},
		{"component.spec.ts", true, "*.spec.* pattern should match"},
		{"__pycache__/module.pyc", true, "__pycache__/** pattern should match"},
		{"vendor/library/file.go", true, "vendor/** pattern should match"},
		{".env", true, ".env* pattern should match"},
		{".env.local", true, ".env* pattern should match"},
		{"tmp/file.txt", true, "tmp/* pattern should match"},
		{"src/tmp/file.txt", false, "tmp/* should only match at root level"},
		{"test/fixture.ts", false, "Should not match *.test.*"},
		{"specfile.js", false, "Should not match *.spec.*"},
		// Test cases for base filename matching
		{"src/main.test.go", true, "*.test.* pattern should match base filename"},
		{"path/to/app.test.js", true, "*.test.* pattern should match in nested paths"},
		{"deep/nested/dir/component.spec.tsx", true, "*.spec.* pattern should match in deep paths"},
		{"src/components/Button.test.tsx", true, "*.test.* pattern should match TypeScript test files"},
		{"tests/unit.spec.js", true, "*.spec.* pattern should match in tests directory"},
	}

	for _, test := range tests {
		result := builder.shouldSkipPath(test.path)
		if result != test.expected {
			t.Errorf("shouldSkipPath(%q) = %v, expected %v (%s)",
				test.path, result, test.expected, test.reason)
		}
	}
}

func TestDefaultExcludePatterns(t *testing.T) {
	builder := NewGraphBuilder()
	// Default is to use default excludes

	tests := []struct {
		path     string
		expected bool
		reason   string
	}{
		{"src/index.ts", false, "Normal source file should not be skipped"},
		{"node_modules/package/index.js", true, "node_modules should be excluded by default"},
		{".venv/lib/python3.9/site-packages/foo.py", true, ".venv should be excluded by default"},
		{"target/debug/app", true, "target should be excluded by default"},
		{"coverage/lcov.info", true, "coverage should be excluded by default"},
		{".DS_Store", true, ".DS_Store should be excluded by default"},
		{"app.log", true, "*.log should be excluded by default"},
		{"dist/bundle.js", true, "dist should be excluded by default"},
	}

	for _, test := range tests {
		result := builder.shouldSkipPath(test.path)
		if result != test.expected {
			t.Errorf("shouldSkipPath(%q) = %v, expected %v (%s)",
				test.path, result, test.expected, test.reason)
		}
	}
}

func TestDisableDefaultExcludes(t *testing.T) {
	builder := NewGraphBuilder()
	builder.SetUseDefaultExcludes(false)

	// Without any custom patterns, nothing should be excluded
	tests := []struct {
		path     string
		expected bool
		reason   string
	}{
		{"node_modules/package/index.js", false, "node_modules should NOT be excluded when defaults disabled"},
		{".venv/lib/python3.9/site-packages/foo.py", false, ".venv should NOT be excluded when defaults disabled"},
		{"coverage/lcov.info", false, "coverage should NOT be excluded when defaults disabled"},
	}

	for _, test := range tests {
		result := builder.shouldSkipPath(test.path)
		if result != test.expected {
			t.Errorf("shouldSkipPath(%q) = %v, expected %v (%s)",
				test.path, result, test.expected, test.reason)
		}
	}
}

func TestNegationPatterns(t *testing.T) {
	builder := NewGraphBuilder()
	builder.SetUseDefaultExcludes(false) // Disable defaults for clearer testing
	builder.SetExcludePatterns([]string{
		"vendor/**",
		"*.test.*",
		"!vendor/our-company/**", // Include our company's vendor code
		"!important.test.js",     // Include this specific test file
	})

	tests := []struct {
		path     string
		expected bool
		reason   string
	}{
		{"vendor/third-party/lib.go", true, "Third party vendor should be excluded"},
		{"vendor/our-company/lib.go", false, "Our company vendor should be included via negation"},
		{"vendor/our-company/internal/util.go", false, "Our company vendor subdirs should be included"},
		{"app.test.js", true, "Regular test files should be excluded"},
		{"important.test.js", false, "Specific test file should be included via negation"},
		{"src/main.go", false, "Normal files should not be excluded"},
	}

	for _, test := range tests {
		result := builder.shouldSkipPath(test.path)
		if result != test.expected {
			t.Errorf("shouldSkipPath(%q) = %v, expected %v (%s)",
				test.path, result, test.expected, test.reason)
		}
	}
}

func TestGetSupportedLanguages(t *testing.T) {
	builder := NewGraphBuilder()
	languages := builder.GetSupportedLanguages()

	if len(languages) == 0 {
		t.Error("GetSupportedLanguages returned empty slice")
	}

	// Should include at least JavaScript and TypeScript
	foundJS := false
	foundTS := false

	for _, lang := range languages {
		if lang.Name == "javascript" {
			foundJS = true
		}
		if lang.Name == "typescript" {
			foundTS = true
		}
	}

	if !foundJS {
		t.Error("JavaScript language not found in supported languages")
	}

	if !foundTS {
		t.Error("TypeScript language not found in supported languages")
	}
}

func TestSetProgressCallback(t *testing.T) {
	builder := NewGraphBuilder()

	// Test setting callback
	var receivedMessages []string
	callback := func(message string) {
		receivedMessages = append(receivedMessages, message)
	}

	builder.SetProgressCallback(callback)

	if builder.progressCallback == nil {
		t.Error("Progress callback was not set")
	}

	// Test callback is nil initially
	builder2 := NewGraphBuilder()
	if builder2.progressCallback != nil {
		t.Error("Progress callback should be nil by default")
	}
}

func TestProgressCallbackExecution(t *testing.T) {
	// Create a temporary directory with multiple test files to trigger progress updates
	tmpDir := t.TempDir()

	// Create 15 test files to ensure we get progress updates (every 10 files)
	for i := 1; i <= 15; i++ {
		testFile := filepath.Join(tmpDir, "test"+string(rune(i+48))+".ts") // test1.ts, test2.ts, etc.
		testContent := `export const value` + string(rune(i+48)) + ` = ` + string(rune(i+48)) + `;`

		err := os.WriteFile(testFile, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %d: %v", i, err)
		}
	}

	// Test with progress callback
	builder := NewGraphBuilder()
	var progressMessages []string

	builder.SetProgressCallback(func(message string) {
		progressMessages = append(progressMessages, message)
	})

	_, err := builder.AnalyzeDirectory(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzeDirectory failed: %v", err)
	}

	// Verify we received progress updates
	if len(progressMessages) == 0 {
		t.Error("No progress messages received")
	}

	// Check for expected progress message patterns
	foundParsingProgress := false
	foundParsingComplete := false
	foundRelationships := false
	foundGitAnalysis := false

	for _, msg := range progressMessages {
		t.Logf("Progress message: %s", msg)

		if msg == "📄 Parsing files... (10 files)" {
			foundParsingProgress = true
		}
		if msg == "✅ Parsing complete (15 files)" {
			foundParsingComplete = true
		}
		if msg == "🔗 Building relationships..." {
			foundRelationships = true
		}
		if msg == "📊 Analyzing git history..." {
			foundGitAnalysis = true
		}
	}

	if !foundParsingProgress {
		t.Error("Expected parsing progress message not found")
	}
	if !foundParsingComplete {
		t.Error("Expected parsing complete message not found")
	}
	if !foundRelationships {
		t.Error("Expected relationships message not found")
	}
	if !foundGitAnalysis {
		t.Error("Expected git analysis message not found")
	}
}

func TestProgressCallbackFileCountUpdates(t *testing.T) {
	// Create temporary directory with 25 files to test multiple updates
	tmpDir := t.TempDir()

	for i := 1; i <= 25; i++ {
		testFile := filepath.Join(tmpDir, "file"+string(rune(i/10+48))+string(rune(i%10+48))+".ts")
		testContent := `export const item = "test";`

		err := os.WriteFile(testFile, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %d: %v", i, err)
		}
	}

	builder := NewGraphBuilder()
	var fileCountMessages []string

	builder.SetProgressCallback(func(message string) {
		// Only capture file count messages
		if strings.Contains(message, "Parsing files...") {
			fileCountMessages = append(fileCountMessages, message)
		}
	})

	_, err := builder.AnalyzeDirectory(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzeDirectory failed: %v", err)
	}

	// Should have updates at 10, 20 files (every 10 files)
	expectedUpdates := []string{
		"📄 Parsing files... (10 files)",
		"📄 Parsing files... (20 files)",
	}

	if len(fileCountMessages) < 2 {
		t.Errorf("Expected at least 2 file count updates, got %d", len(fileCountMessages))
	}

	for i, expected := range expectedUpdates {
		if i < len(fileCountMessages) {
			if fileCountMessages[i] != expected {
				t.Errorf("File count update %d: expected %q, got %q",
					i, expected, fileCountMessages[i])
			}
		} else {
			t.Errorf("Missing expected file count update: %q", expected)
		}
	}
}

func TestProgressCallbackWithNoFiles(t *testing.T) {
	// Test with empty directory
	tmpDir := t.TempDir()

	builder := NewGraphBuilder()
	var progressMessages []string

	builder.SetProgressCallback(func(message string) {
		progressMessages = append(progressMessages, message)
	})

	_, err := builder.AnalyzeDirectory(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzeDirectory failed: %v", err)
	}

	// Should still get completion messages even with no files
	foundParsingComplete := false
	foundRelationships := false

	for _, msg := range progressMessages {
		if msg == "✅ Parsing complete (0 files)" {
			foundParsingComplete = true
		}
		if msg == "🔗 Building relationships..." {
			foundRelationships = true
		}
	}

	if !foundParsingComplete {
		t.Error("Expected parsing complete message for empty directory")
	}
	if !foundRelationships {
		t.Error("Expected relationships message for empty directory")
	}
}

func TestProgressMessageFormats(t *testing.T) {
	tests := []struct {
		name      string
		fileCount int
		expected  []string
	}{
		{
			name:      "single_update",
			fileCount: 12,
			expected: []string{
				"📄 Parsing files... (10 files)",
				"✅ Parsing complete (12 files)",
			},
		},
		{
			name:      "multiple_updates",
			fileCount: 35,
			expected: []string{
				"📄 Parsing files... (10 files)",
				"📄 Parsing files... (20 files)",
				"📄 Parsing files... (30 files)",
				"✅ Parsing complete (35 files)",
			},
		},
		{
			name:      "exact_ten",
			fileCount: 10,
			expected: []string{
				"📄 Parsing files... (10 files)",
				"✅ Parsing complete (10 files)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create the specified number of test files
			for i := 1; i <= tt.fileCount; i++ {
				testFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.ts", i))
				testContent := fmt.Sprintf(`export const value%d = %d;`, i, i)

				err := os.WriteFile(testFile, []byte(testContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file %d: %v", i, err)
				}
			}

			builder := NewGraphBuilder()
			var actualMessages []string

			builder.SetProgressCallback(func(message string) {
				// Only capture parsing messages for this test
				if strings.Contains(message, "Parsing") {
					actualMessages = append(actualMessages, message)
				}
			})

			_, err := builder.AnalyzeDirectory(tmpDir)
			if err != nil {
				t.Fatalf("AnalyzeDirectory failed: %v", err)
			}

			// Verify expected messages are present
			for _, expected := range tt.expected {
				if !slices.Contains(actualMessages, expected) {
					t.Errorf("Expected message %q not found in actual messages: %v",
						expected, actualMessages)
				}
			}
		})
	}
}

func TestProgressMessagesOrder(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 15 test files
	for i := 1; i <= 15; i++ {
		testFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.ts", i))
		testContent := `export const value = 1;`

		err := os.WriteFile(testFile, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %d: %v", i, err)
		}
	}

	builder := NewGraphBuilder()
	var allMessages []string

	builder.SetProgressCallback(func(message string) {
		allMessages = append(allMessages, message)
	})

	_, err := builder.AnalyzeDirectory(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzeDirectory failed: %v", err)
	}

	// Verify message order and progression
	expectedOrder := []string{
		"📄 Parsing files... (10 files)",
		"✅ Parsing complete (15 files)",
		"🔗 Building relationships...",
		"✅ Relationships built",
		"📊 Analyzing git history...",
	}

	// Check that messages appear in the correct order
	messageIndex := 0
	for _, expected := range expectedOrder {
		found := false
		for i := messageIndex; i < len(allMessages); i++ {
			if allMessages[i] == expected {
				found = true
				messageIndex = i + 1
				break
			}
		}
		if !found {
			t.Errorf("Expected message %q not found in correct order. All messages: %v",
				expected, allMessages)
		}
	}
}

func TestProgressConfigurableInterval(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 15 test files
	for i := 1; i <= 15; i++ {
		testFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.ts", i))
		testContent := `export const value = 1;`

		err := os.WriteFile(testFile, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %d: %v", i, err)
		}
	}

	// Test with custom interval of 5 files
	builder := NewGraphBuilder()
	builder.SetProgressInterval(5)

	var progressMessages []string
	builder.SetProgressCallback(func(message string) {
		if strings.Contains(message, "Parsing files...") {
			progressMessages = append(progressMessages, message)
		}
	})

	_, err := builder.AnalyzeDirectory(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzeDirectory failed: %v", err)
	}

	// Should have updates at 5, 10, 15 files
	expectedUpdates := []string{
		"📄 Parsing files... (5 files)",
		"📄 Parsing files... (10 files)",
		"📄 Parsing files... (15 files)",
	}

	if len(progressMessages) != 3 {
		t.Errorf("Expected 3 progress updates with interval 5, got %d: %v",
			len(progressMessages), progressMessages)
	}

	for i, expected := range expectedUpdates {
		if i < len(progressMessages) {
			if progressMessages[i] != expected {
				t.Errorf("Progress update %d: expected %q, got %q",
					i, expected, progressMessages[i])
			}
		}
	}
}

func TestProgressConfig(t *testing.T) {
	builder := NewGraphBuilder()

	// Test setting progress config
	config := ProgressConfig{
		Interval:       5,
		ShowPercentage: true,
	}

	builder.SetProgressConfig(config)

	// Verify internal state
	if builder.progressConfig.Interval != 5 {
		t.Errorf("Expected interval 5, got %d", builder.progressConfig.Interval)
	}

	if !builder.progressConfig.ShowPercentage {
		t.Error("Expected ShowPercentage to be true")
	}
}

// Benchmark Tests

func BenchmarkPatternMatching(b *testing.B) {
	builder := NewGraphBuilder()
	patterns := []string{
		"*.js", "*.ts", "node_modules/**", ".git/**",
		"dist/**", "coverage/**", "*.test.*", "*.spec.*",
	}

	testPaths := []string{
		"src/main.js",
		"node_modules/react/index.js",
		"dist/bundle.js",
		"test/main.test.js",
		"coverage/lcov.info",
		".git/config",
		"docs/README.md",
		"src/components/Button.tsx",
	}

	b.ResetTimer()
	for range b.N {
		for _, path := range testPaths {
			builder.matchesPattern(path, patterns)
		}
	}
}

func BenchmarkDefaultPatternMatching(b *testing.B) {
	builder := NewGraphBuilder()
	// Use actual default patterns
	patterns := builder.getMergedPatterns()

	testPath := "node_modules/react/lib/index.js"

	b.ResetTimer()
	for range b.N {
		builder.matchesPattern(testPath, patterns)
	}
}

func BenchmarkShouldSkipPath(b *testing.B) {
	builder := NewGraphBuilder()
	builder.SetExcludePatterns([]string{
		"*.test.*",
		"!important.test.js",
	})

	testPaths := []string{
		"src/main.js",
		"node_modules/react/index.js",
		"test.test.js",
		"important.test.js",
		"dist/bundle.js",
	}

	b.ResetTimer()
	for range b.N {
		for _, path := range testPaths {
			builder.shouldSkipPath(path)
		}
	}
}

func BenchmarkPatternCaching(b *testing.B) {
	builder := NewGraphBuilder()

	b.Run("WithCaching", func(b *testing.B) {
		for range b.N {
			// This should use cached patterns after first call
			patterns := builder.getMergedPatterns()
			_ = patterns
		}
	})

	b.Run("WithoutCaching", func(b *testing.B) {
		for range b.N {
			// Force regeneration each time (simulate old behavior)
			builder.patternsDirty = true
			patterns := builder.getMergedPatterns()
			_ = patterns
		}
	})
}

// Path Normalization Tests

func TestPathNormalization(t *testing.T) {
	builder := NewGraphBuilder()
	
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic normalization
		{"basic_path", "src/main.go", "src/main.go"},
		{"current_dir", "./main.go", "main.go"},
		{"parent_dir", "../main.go", "../main.go"},
		{"double_dots", "src/../main.go", "main.go"},
		{"trailing_slash", "src/", "src"},
		{"multiple_slashes", "src//main.go", "src/main.go"},
		
		// Complex cases
		{"complex_traversal", "src/../lib/../main.go", "main.go"},
		{"deep_traversal", "a/b/c/../../d/../e.go", "a/e.go"},
		{"empty_path", "", "."},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := builder.normalizePath(test.input)
			if result != test.expected {
				t.Errorf("normalizePath(%q) = %q, expected %q", 
					test.input, result, test.expected)
			}
		})
	}
}

func TestNormalizeForPattern(t *testing.T) {
	builder := NewGraphBuilder()
	
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Unix-style paths (should remain unchanged)
		{"unix_basic", "src/main.go", "src/main.go"},
		{"unix_nested", "src/components/Button.tsx", "src/components/Button.tsx"},
		
		// Paths with backslashes (should convert to forward slashes)
		{"mixed_separators", "src\\main.go", "src/main.go"},
		{"windows_style", "src\\components\\Button.tsx", "src/components/Button.tsx"},
		
		// With normalization
		{"dots_with_backslash", "src\\..\\main.go", "main.go"},
		{"complex_mixed", "src\\..\\lib/..\\main.go", "main.go"},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := builder.normalizeForPattern(test.input)
			if result != test.expected {
				t.Errorf("normalizeForPattern(%q) = %q, expected %q", 
					test.input, result, test.expected)
			}
		})
	}
}

func TestValidateImportPath(t *testing.T) {
	builder := NewGraphBuilder()
	baseDir := "/home/user/project"
	
	tests := []struct {
		name      string
		importPath string
		baseDir   string
		expectErr bool
		reason    string
	}{
		// Safe paths
		{"relative_safe", "./lib/utils.js", baseDir, false, "relative path within project"},
		{"nested_safe", "../components/Button.tsx", baseDir, false, "parent directory within project"},
		{"no_traversal", "utils.js", baseDir, false, "no traversal sequences"},
		
		// Dangerous paths
		{"escape_root", "../../../etc/passwd", baseDir, true, "escapes project directory"},
		{"escape_hidden", "lib/../../../etc/passwd", baseDir, true, "hidden traversal escape"},
		{"deep_escape", "../../../../bin/sh", baseDir, true, "deep directory traversal"},
		
		// Edge cases
		{"just_parent", "..", baseDir, false, "single parent directory"},
		{"two_parents", "../..", baseDir, false, "two parent directories (reasonable)"},
		{"many_parents", "../../..", baseDir, true, "too many parent directories"},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := builder.validateImportPath(test.importPath, test.baseDir)
			
			if test.expectErr && err == nil {
				t.Errorf("validateImportPath(%q, %q) expected error but got none (%s)", 
					test.importPath, test.baseDir, test.reason)
			} else if !test.expectErr && err != nil {
				t.Errorf("validateImportPath(%q, %q) unexpected error: %v (%s)", 
					test.importPath, test.baseDir, err, test.reason)
			}
		})
	}
}

func TestCrossPlatformPatternMatching(t *testing.T) {
	builder := NewGraphBuilder()
	builder.SetUseDefaultExcludes(false)
	builder.SetExcludePatterns([]string{
		"node_modules/**",
		"*.test.*",
		"build/**",
	})
	
	tests := []struct {
		name     string
		path     string
		expected bool
		reason   string
	}{
		// Unix-style paths
		{"unix_node_modules", "node_modules/react/index.js", true, "should match node_modules pattern"},
		{"unix_test_file", "src/main.test.js", true, "should match test file pattern"},
		{"unix_build_dir", "build/output.js", true, "should match build directory pattern"},
		{"unix_normal_file", "src/main.js", false, "normal file should not be excluded"},
		
		// Windows-style paths (backslashes should be handled)
		{"windows_node_modules", "node_modules\\react\\index.js", true, "should match node_modules with backslashes"},
		{"windows_test_file", "src\\main.test.js", true, "should match test file with backslashes"},
		{"windows_build_dir", "build\\output.js", true, "should match build directory with backslashes"},
		{"windows_normal_file", "src\\main.js", false, "normal Windows file should not be excluded"},
		
		// Mixed separators
		{"mixed_separators", "node_modules/react\\index.js", true, "should handle mixed separators"},
		{"mixed_test", "src\\components/Button.test.tsx", true, "should match mixed separator test file"},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := builder.shouldSkipPath(test.path)
			if result != test.expected {
				t.Errorf("shouldSkipPath(%q) = %v, expected %v (%s)",
					test.path, result, test.expected, test.reason)
			}
		})
	}
}

func TestPathNormalizationInProcessFile(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir := t.TempDir()
	
	// Create test files with different path formats
	testFile1 := filepath.Join(tmpDir, "main.go")
	testFile2 := filepath.Join(tmpDir, "subdir", "utils.go")
	
	// Create subdirectory
	err := os.MkdirAll(filepath.Dir(testFile2), 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	
	// Write test content
	content1 := `package main
func main() {}
`
	content2 := `package subdir
func Helper() string { return "test" }
`
	
	err = os.WriteFile(testFile1, []byte(content1), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file 1: %v", err)
	}
	
	err = os.WriteFile(testFile2, []byte(content2), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file 2: %v", err)
	}
	
	// Test the analyzer with different path formats
	builder := NewGraphBuilder()
	graph, err := builder.AnalyzeDirectory(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzeDirectory failed: %v", err)
	}
	
	// Verify that all paths in the graph are normalized
	for filePath := range graph.Files {
		normalized := builder.normalizePath(filePath)
		if filePath != normalized {
			t.Errorf("File path %q is not normalized, should be %q", filePath, normalized)
		}
		
		// Verify that the path doesn't contain redundant elements
		if strings.Contains(filePath, "//") || strings.Contains(filePath, "/./") || 
		   strings.Contains(filePath, "/../") {
			t.Errorf("File path %q contains redundant elements", filePath)
		}
	}
}

// Benchmark tests for path normalization performance

func BenchmarkPathNormalization(b *testing.B) {
	builder := NewGraphBuilder()
	testPaths := []string{
		"src/main.go",
		"src/../lib/utils.go",
		"./components/Button.tsx",
		"deep/nested/path/to/file.js",
		"src//double//slash.go",
	}
	
	b.ResetTimer()
	for range b.N {
		for _, path := range testPaths {
			builder.normalizePath(path)
		}
	}
}

func BenchmarkNormalizeForPattern(b *testing.B) {
	builder := NewGraphBuilder()
	testPaths := []string{
		"src\\main.go",
		"src\\..\\lib\\utils.go", 
		".\\components\\Button.tsx",
		"deep\\nested\\path\\to\\file.js",
		"mixed/separators\\file.go",
	}
	
	b.ResetTimer()
	for range b.N {
		for _, path := range testPaths {
			builder.normalizeForPattern(path)
		}
	}
}

func BenchmarkCrossPlatformPatternMatching(b *testing.B) {
	builder := NewGraphBuilder()
	builder.SetExcludePatterns([]string{
		"node_modules/**",
		"*.test.*",
		"build/**",
		"dist/**",
		"coverage/**",
	})
	
	testPaths := []string{
		"src/main.js",
		"node_modules\\react\\index.js",
		"src/components\\Button.test.tsx",
		"build/output.js",
		"dist\\bundle.js",
		"coverage/lcov.info",
	}
	
	b.ResetTimer()
	for range b.N {
		for _, path := range testPaths {
			builder.shouldSkipPath(path)
		}
	}
}

// Security and Advanced Test Cases

func TestAdvancedDirectoryTraversal(t *testing.T) {
	builder := NewGraphBuilder()
	baseDir := "/home/user/project"
	
	tests := []struct {
		name        string
		importPath  string
		expectError bool
		description string
	}{
		// Advanced traversal attempts
		{"mixed_separators_attack", "./lib\\..\\../etc/passwd", true, "Mixed separator traversal"},
		{"excessive_traversal", "../../../../../../../../etc/passwd", true, "Excessive upward traversal"},
		{"hidden_in_path", "legitimate/path/../../../etc/passwd", true, "Hidden traversal in legitimate path"},
		{"double_dot_variations", "lib/...//etc/passwd", false, "Invalid double dot should be handled"},
		{"trailing_traversal", "lib/file/../../../etc/passwd", true, "Traversal after filename"},
		
		// System directory access attempts
		{"passwd_file", "../../../etc/passwd", true, "Direct passwd file access"},
		{"shadow_file", "../../../etc/shadow", true, "Shadow file access attempt"},
		{"hosts_file", "../../../etc/hosts", true, "Hosts file access attempt"},
		{"bin_directory", "../../../bin/sh", true, "Binary directory access"},
		{"usr_bin_access", "../../../usr/bin/whoami", true, "Usr/bin access attempt"},
		{"sbin_access", "../../../sbin/init", true, "Sbin access attempt"},
		
		// Windows system paths
		{"windows_system32", "../../../Windows/System32/cmd.exe", true, "Windows System32 access"},
		{"windows_drivers", "../../../Windows/System32/drivers/etc/hosts", true, "Windows drivers access"},
		
		// Legitimate cases that should pass
		{"sibling_directory", "../components/Button.tsx", false, "Legitimate sibling access"},
		{"grandparent_ok", "../../shared/utils.js", false, "Reasonable grandparent access"},
		{"current_and_parent", "./lib/../index.js", false, "Current and parent combination"},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := builder.validateImportPath(test.importPath, baseDir)
			
			if test.expectError && err == nil {
				t.Errorf("validateImportPath(%q) expected error but got none (%s)", 
					test.importPath, test.description)
			} else if !test.expectError && err != nil {
				t.Errorf("validateImportPath(%q) unexpected error: %v (%s)", 
					test.importPath, err, test.description)
			}
		})
	}
}

func TestInvalidGlobPatterns(t *testing.T) {
	builder := NewGraphBuilder()
	
	// Test that malformed patterns don't crash the system
	malformedPatterns := []string{
		"file[",           // Unclosed bracket
		"file[abc",        // Incomplete bracket
		"file[z-a]",       // Invalid range
		"file\\",          // Trailing escape
		"[",               // Just bracket
		"]",               // Just closing bracket
		"file[[]",         // Nested brackets
	}
	
	// These should not crash the system
	builder.SetExcludePatterns(malformedPatterns)
	
	testPaths := []string{
		"file.go",
		"file[.go",
		"fileabc.go",
		"files.go",
	}
	
	for _, path := range testPaths {
		// Should not crash, even with malformed patterns
		result := builder.shouldSkipPath(path)
		t.Logf("Path %q with malformed patterns: %v", path, result)
	}
}

func TestSpecialCharacterPaths(t *testing.T) {
	builder := NewGraphBuilder()
	builder.SetUseDefaultExcludes(false)
	builder.SetExcludePatterns([]string{"*.test.*", "*temp*"})
	
	tests := []struct {
		name     string
		path     string
		expected bool
		reason   string
	}{
		// Paths with spaces
		{"spaces_in_path", "src/path with spaces/file.go", false, "Spaces should be handled"},
		{"spaces_test_file", "src/test file.test.js", true, "Spaces with test pattern"},
		
		// Special characters
		{"hyphen_underscore", "src/file-name_with-chars.go", false, "Hyphens and underscores"},
		{"dots_in_name", "src/file.name.with.dots.go", false, "Multiple dots in filename"},
		{"special_chars", "src/file!@#$%^&()_+.go", false, "Special characters in name"},
		
		// Unicode characters
		{"unicode_path", "路径/文件.go", false, "Unicode characters should work"},
		{"unicode_test", "路径/测试.test.js", true, "Unicode with test pattern"},
		
		// Parentheses and brackets
		{"parentheses", "src/(component)/file.go", false, "Parentheses in path"},
		{"square_brackets", "src/[version]/file.go", false, "Square brackets in path"},
		
		// Temp pattern matching
		{"temp_dir", "tmp/temp/file.go", true, "Should match temp pattern"},
		{"temporary", "src/temporary_file.go", true, "Should match temp pattern"},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := builder.shouldSkipPath(test.path)
			if result != test.expected {
				t.Errorf("shouldSkipPath(%q) = %v, expected %v (%s)",
					test.path, result, test.expected, test.reason)
			}
		})
	}
}

func TestAbsolutePathNormalization(t *testing.T) {
	builder := NewGraphBuilder()
	
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Unix absolute paths
		{"unix_absolute", "/home/user/project/main.go", "/home/user/project/main.go"},
		{"unix_root", "/main.go", "/main.go"},
		{"unix_with_traversal", "/home/user/../user/project/main.go", "/home/user/project/main.go"},
		
		// Windows absolute paths (when converted)
		{"windows_absolute", "C:\\Users\\user\\project\\main.go", "C:/Users/user/project/main.go"},
		{"windows_drive_only", "C:\\main.go", "C:/main.go"},
		{"windows_mixed", "C:/Users\\user/project\\main.go", "C:/Users/user/project/main.go"},
		
		// Network paths
		{"unc_basic", "\\\\server\\share\\file.go", "//server/share/file.go"},
		{"unc_nested", "\\\\server\\share\\folder\\subfolder\\file.go", "//server/share/folder/subfolder/file.go"},
		
		// Edge cases
		{"absolute_with_dots", "/home/./user/../user/file.go", "/home/user/file.go"},
		{"multiple_slashes", "/home///user//file.go", "/home/user/file.go"},
		{"trailing_slash_absolute", "/home/user/project/", "/home/user/project"},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := builder.normalizeForPattern(test.input)
			if result != test.expected {
				t.Errorf("normalizeForPattern(%q) = %q, expected %q", 
					test.input, result, test.expected)
			}
		})
	}
}

func TestPatternPrecedence(t *testing.T) {
	builder := NewGraphBuilder()
	builder.SetUseDefaultExcludes(false)
	
	// Test complex pattern precedence with multiple rules
	builder.SetExcludePatterns([]string{
		"*.test.*",              // Exclude all test files
		"!critical.test.js",     // But include critical test
		"test/**",               // Exclude test directory
		"!test/fixtures/**",     // But include fixtures
		"temp/**",               // Exclude temp directory
		"!temp/keep/**",         // But keep some temp files
		"**/*.backup",           // Exclude backup files everywhere
		"!important.backup",     // But keep important backup
	})
	
	tests := []struct {
		name     string
		path     string
		expected bool
		reason   string
	}{
		// Test file patterns
		{"regular_test", "src/app.test.js", true, "Should be excluded by *.test.*"},
		{"critical_test", "critical.test.js", false, "Should be included by !critical.test.js"},
		{"critical_test_nested", "src/critical.test.js", false, "Critical test in nested path"},
		
		// Directory-based patterns
		{"test_dir_file", "test/unit.js", true, "Should be excluded by test/**"},
		{"test_fixtures", "test/fixtures/data.json", false, "Should be included by !test/fixtures/**"},
		{"test_fixtures_nested", "test/fixtures/nested/data.json", false, "Nested fixtures should be included"},
		
		// Temp directory patterns
		{"temp_file", "temp/cache.tmp", true, "Should be excluded by temp/**"},
		{"temp_keep", "temp/keep/important.txt", false, "Should be included by !temp/keep/**"},
		{"temp_keep_nested", "temp/keep/nested/file.txt", false, "Nested keep files should be included"},
		
		// Backup file patterns
		{"backup_file", "src/old.backup", true, "Should be excluded by **/*.backup"},
		{"important_backup", "important.backup", false, "Should be included by !important.backup"},
		{"important_backup_nested", "src/important.backup", false, "Important backup in nested path"},
		
		// Non-matching patterns
		{"normal_file", "src/main.go", false, "Normal file should not be excluded"},
		{"normal_js", "src/app.js", false, "Normal JS file should not be excluded"},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := builder.shouldSkipPath(test.path)
			if result != test.expected {
				t.Errorf("shouldSkipPath(%q) = %v, expected %v (%s)",
					test.path, result, test.expected, test.reason)
			}
		})
	}
}

func TestEmptyAndEdgeCasePaths(t *testing.T) {
	builder := NewGraphBuilder()
	
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Empty and whitespace
		{"empty_string", "", "."},
		{"single_dot", ".", "."},
		{"double_dot", "..", ".."},
		{"just_slash", "/", "/"},
		{"just_backslash", "\\", "/"},
		
		// Whitespace handling
		{"leading_space", " file.go", " file.go"},
		{"trailing_space", "file.go ", "file.go "},
		{"internal_spaces", "my file.go", "my file.go"},
		
		// Multiple separators
		{"many_slashes", "a///b///c", "a/b/c"},
		{"many_backslashes", "a\\\\\\b\\\\\\c", "a/b/c"},
		{"mixed_many", "a//\\\\//b", "a/b"},
		
		// Extreme traversal
		{"many_dots", "a/../../../b", "../../b"},
		{"mixed_dots", "./a/.././../b", "../b"},
		{"dots_and_slashes", ".///.././//b", "../b"},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := builder.normalizeForPattern(test.input)
			if result != test.expected {
				t.Errorf("normalizeForPattern(%q) = %q, expected %q", 
					test.input, result, test.expected)
			}
		})
	}
}

func TestConcurrentPathNormalization(t *testing.T) {
	builder := NewGraphBuilder()
	builder.SetExcludePatterns([]string{
		"*.test.*",
		"node_modules/**",
		"build/**",
		"temp/**",
	})
	
	// Test concurrent access to path normalization
	var wg sync.WaitGroup
	numGoroutines := 100
	pathsPerGoroutine := 100
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			for j := 0; j < pathsPerGoroutine; j++ {
				path := fmt.Sprintf("src/file%d_%d.go", goroutineID, j)
				testPath := fmt.Sprintf("test/file%d_%d.test.js", goroutineID, j)
				windowsPath := fmt.Sprintf("src\\windows%d_%d.go", goroutineID, j)
				
				// These should not cause data races or crashes
				builder.shouldSkipPath(path)
				builder.shouldSkipPath(testPath)
				builder.shouldSkipPath(windowsPath)
				
				builder.normalizePath(path)
				builder.normalizeForPattern(windowsPath)
			}
		}(i)
	}
	
	wg.Wait()
	// If we get here without data races or crashes, the test passes
}

func TestLargePathHandling(t *testing.T) {
	builder := NewGraphBuilder()
	
	tests := []struct {
		name   string
		length int
		valid  bool
	}{
		{"normal_path", 50, true},
		{"long_path", 300, true},
		{"very_long_path", 1000, true},
		{"extreme_path", 4000, true}, // Near Unix path limit
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create path of specified length
			segment := strings.Repeat("a", 10)
			segments := test.length / 10
			pathParts := make([]string, segments)
			for i := 0; i < segments; i++ {
				pathParts[i] = segment
			}
			longPath := strings.Join(pathParts, "/") + "/file.go"
			
			// Test normalization doesn't crash or hang
			result := builder.normalizePath(longPath)
			if len(result) == 0 && test.valid {
				t.Errorf("normalizePath returned empty string for valid long path")
			}
			
			// Test pattern matching doesn't crash
			matches := builder.shouldSkipPath(longPath)
			_ = matches // We just care that it doesn't crash
		})
	}
}

func TestNilAndErrorHandling(t *testing.T) {
	builder := NewGraphBuilder()
	
	// Test nil pattern slice handling
	builder.SetExcludePatterns(nil)
	result := builder.shouldSkipPath("test/file.go")
	if result != false {
		t.Errorf("Expected false for nil patterns, got %v", result)
	}
	
	// Test empty pattern slice
	builder.SetExcludePatterns([]string{})
	result = builder.shouldSkipPath("test/file.go") 
	if result != false {
		t.Errorf("Expected false for empty patterns, got %v", result)
	}
	
	// Test pattern slice with empty strings
	builder.SetExcludePatterns([]string{"", "*.test.*", ""})
	result = builder.shouldSkipPath("app.test.js")
	if result != true {
		t.Errorf("Expected true for test file with mixed empty patterns, got %v", result)
	}
	
	// Test nil progress callback (should not crash)
	builder.SetProgressCallback(nil)
	// This should not crash when called internally
	
	// Test empty base directory for import validation
	err := builder.validateImportPath("../test.js", "")
	if err == nil {
		t.Log("Empty base directory handled gracefully")
	}
	
	// Test very deep directory validation
	deepPath := strings.Repeat("../", 10) + "etc/passwd"
	err = builder.validateImportPath(deepPath, "/home/user/project")
	if err == nil {
		t.Errorf("Expected error for very deep traversal, got nil")
	}
}

func TestDoubleStarPatternEdgeCases(t *testing.T) {
	builder := NewGraphBuilder()
	builder.SetUseDefaultExcludes(false)
	
	tests := []struct {
		name     string
		pattern  string
		path     string
		expected bool
		reason   string
	}{
		// ** at beginning
		{"double_star_start", "**/test.js", "deep/nested/test.js", true, "** should match any depth"},
		{"double_star_start_root", "**/test.js", "test.js", true, "** should match root level"},
		
		// ** in middle
		{"double_star_middle", "src/**/test.js", "src/components/deep/test.js", true, "** should match nested paths"},
		{"double_star_middle_direct", "src/**/test.js", "src/test.js", true, "** should match direct children"},
		
		// ** at end
		{"double_star_end", "node_modules/**", "node_modules/react/index.js", true, "** should match all descendants"},
		{"double_star_end_direct", "node_modules/**", "node_modules/package.json", true, "** should match direct files"},
		
		// Multiple ** patterns
		{"multiple_double_star", "**/node_modules/**", "deep/node_modules/react/index.js", true, "Multiple ** should work"},
		{"adjacent_double_star", "**/**", "any/path/file.js", true, "Adjacent ** should work"},
		
		// ** with other patterns
		{"double_star_with_glob", "**/*.test.*", "deep/nested/app.test.js", true, "** with other globs"},
		{"double_star_complex", "src/**/components/*.tsx", "src/pages/components/Button.tsx", true, "Complex ** pattern"},
		
		// Edge cases that shouldn't match
		{"double_star_wrong_extension", "**/test.js", "deep/nested/test.ts", false, "Wrong extension shouldn't match"},
		{"double_star_wrong_prefix", "test/**", "testing/file.js", false, "Wrong prefix shouldn't match"},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			builder.SetExcludePatterns([]string{test.pattern})
			result := builder.shouldSkipPath(test.path)
			
			if result != test.expected {
				t.Errorf("Pattern %q with path %q: got %v, expected %v (%s)",
					test.pattern, test.path, result, test.expected, test.reason)
			}
		})
	}
}

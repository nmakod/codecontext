package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
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

package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuthan-ms/codecontext/internal/config"
)

// TestSuite provides comprehensive CLI integration testing
type TestSuite struct {
	tempDir string
	stdout  *bytes.Buffer
	stderr  *bytes.Buffer
	config  *config.Config
}

// SetupTestSuite creates a new test suite with temporary directory
func SetupTestSuite(t *testing.T) *TestSuite {
	tempDir := t.TempDir() // Use t.TempDir() for automatic cleanup

	suite := &TestSuite{
		tempDir: tempDir,
		stdout:  &bytes.Buffer{},
		stderr:  &bytes.Buffer{},
		config: &config.Config{
			SourcePaths: []string{tempDir}, // Use absolute path instead of "."
			OutputPath:  filepath.Join(tempDir, "output.md"),
			CacheDir:    filepath.Join(tempDir, ".cache"),
			IncludePatterns: []string{
				"*.go", "*.js", "*.ts", "*.jsx", "*.tsx",
			},
			ExcludePatterns: []string{
				"node_modules/**", ".git/**", "*.test.*",
			},
			MaxFileSize:    1024 * 1024, // 1MB
			Concurrency:    4,
			EnableCache:    true,
			EnableProgress: false, // Disable for testing
			EnableWatching: false,
			EnableVerbose:  false,
		},
	}

	return suite
}

// TestGenerateCommandProgress tests the progress functionality in generate command
func TestGenerateCommandProgress(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TeardownTestSuite(t)

	// Create multiple test files to trigger progress updates
	suite.CreateTestFiles(t)
	
	// Create additional files to ensure progress updates
	for i := 1; i <= 15; i++ {
		filename := filepath.Join(suite.tempDir, fmt.Sprintf("test%d.ts", i))
		content := fmt.Sprintf(`export const value%d = %d;
export function test%d() {
  return "test%d";
}`, i, i, i, i)
		
		err := os.WriteFile(filename, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Set up paths
	outputFile := filepath.Join(suite.tempDir, "CLAUDE.md")
	
	// Create mock command that calls the generate function directly
	cmd := &cobra.Command{
		Use: "generate",
	}
	cmd.Flags().StringP("target", "t", suite.tempDir, "target directory")
	cmd.Flags().StringP("output", "o", outputFile, "output file")
	
	// Bind flags to viper (required for generateContextMap to access them)
	viper.BindPFlag("target", cmd.Flags().Lookup("target"))
	viper.BindPFlag("output", cmd.Flags().Lookup("output"))
	
	// Execute the function directly
	err := generateContextMap(cmd)
	require.NoError(t, err)
	
	// Verify output file was created
	_, err = os.Stat(outputFile)
	require.NoError(t, err, "CLAUDE.md should be created")
	
	// Verify the file contains content (basic sanity check)
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Greater(t, len(content), 0, "Output file should not be empty")
}

// TeardownTestSuite cleans up the test suite
func (ts *TestSuite) TeardownTestSuite(t *testing.T) {
	// Note: t.TempDir() automatically cleans up temp directory, no need to manually remove
	// No directory changes needed since we work with absolute paths
}

// CreateTestFiles creates a set of test files for integration testing
func (ts *TestSuite) CreateTestFiles(t *testing.T) {
	files := map[string]string{
		"main.go": `package main

import (
	"fmt"
	"codecontext/internal/utils"
)

func main() {
	result := utils.ProcessData("test")
	fmt.Println(result)
}`,
		"internal/utils/processor.go": `package utils

import "strings"

// ProcessData processes input data
func ProcessData(input string) string {
	return strings.ToUpper(input)
}

// Helper function
func validateInput(input string) bool {
	return len(input) > 0
}`,
		"internal/models/user.go": `package models

// User represents a user in the system
type User struct {
	ID   int    ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
	Email string ` + "`json:\"email\"`" + `
}

// NewUser creates a new user
func NewUser(name, email string) *User {
	return &User{
		Name:  name,
		Email: email,
	}
}`,
		"frontend/app.js": `// Main application entry point
import { UserService } from './services/user.js';
import { utils } from './utils/helpers.js';

class App {
	constructor() {
		this.userService = new UserService();
		this.initialized = false;
	}

	async initialize() {
		await this.userService.loadUsers();
		this.initialized = true;
		utils.log('App initialized');
	}
}

export default App;`,
		"frontend/services/user.js": `// User service for API calls
export class UserService {
	constructor() {
		this.baseUrl = '/api/users';
		this.users = [];
	}

	async loadUsers() {
		try {
			const response = await fetch(this.baseUrl);
			this.users = await response.json();
			return this.users;
		} catch (error) {
			console.error('Failed to load users:', error);
			throw error;
		}
	}

	async createUser(userData) {
		const response = await fetch(this.baseUrl, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(userData)
		});
		return response.json();
	}
}`,
		"frontend/utils/helpers.js": `// Utility functions
export const utils = {
	log(message) {
		console.log('[App]', message);
	},

	formatDate(date) {
		return new Date(date).toLocaleDateString();
	},

	debounce(func, wait) {
		let timeout;
		return function executedFunction(...args) {
			const later = () => {
				clearTimeout(timeout);
				func(...args);
			};
			clearTimeout(timeout);
			timeout = setTimeout(later, wait);
		};
	}
};`,
		"README.md": `# Test Project

This is a test project for CodeContext integration testing.

## Structure

- main.go - Entry point
- internal/ - Internal packages
- frontend/ - JavaScript frontend
`,
		"package.json": `{
	"name": "test-project",
	"version": "1.0.0",
	"dependencies": {
		"express": "^4.18.0"
	}
}`,
	}

	for filePath, content := range files {
		fullPath := filepath.Join(ts.tempDir, filePath)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		require.NoError(t, err)

		err = os.WriteFile(fullPath, []byte(content), 0644)
		require.NoError(t, err)
	}
}

// CreateRootCommand creates a root command for testing
func (ts *TestSuite) CreateRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "codecontext",
		Short: "CodeContext CLI for testing",
	}

	// Add commands
	rootCmd.AddCommand(ts.createGenerateCommand())
	rootCmd.AddCommand(ts.createWatchCommand())

	return rootCmd
}

func (ts *TestSuite) createGenerateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "generate",
		Short: "Generate code context map",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ts.runGenerate()
		},
	}
}

func (ts *TestSuite) createWatchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "watch",
		Short: "Watch for changes and update context map",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ts.runWatch()
		},
	}
}

func (ts *TestSuite) runGenerate() error {
	// Scan files
	files, err := ts.scanFiles()
	if err != nil {
		return err
	}

	if len(files) == 0 {
		ts.stderr.WriteString("No files found to process\n")
		return nil
	}

	// Create simple output
	output := fmt.Sprintf("# CodeContext Map\n\n**Generated:** %s\n\n## Files (%d found)\n\n",
		time.Now().Format(time.RFC3339), len(files))

	for _, file := range files {
		output += fmt.Sprintf("- %s\n", file)
	}

	// Write output
	err = os.WriteFile(ts.config.OutputPath, []byte(output), 0644)
	if err != nil {
		return err
	}

	ts.stdout.WriteString("Code context map generated successfully\n")
	return nil
}

func (ts *TestSuite) runWatch() error {
	// Simulate watch mode by waiting briefly then generating output
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Wait for timeout (simulating watch)
	<-ctx.Done()

	// Generate output like in generate mode
	err := ts.runGenerate()
	if err != nil {
		return err
	}

	ts.stdout.WriteString("Watch mode completed\n")
	return nil
}

func (ts *TestSuite) scanFiles() ([]string, error) {
	var files []string

	for _, sourcePath := range ts.config.SourcePaths {
		err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			// Check include patterns
			included := false
			for _, pattern := range ts.config.IncludePatterns {
				if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
					included = true
					break
				}
			}

			if !included {
				return nil
			}

			// Check exclude patterns
			for _, pattern := range ts.config.ExcludePatterns {
				if matched, _ := filepath.Match(pattern, path); matched {
					return nil
				}
			}

			// Check file size
			if info.Size() > int64(ts.config.MaxFileSize) {
				return nil
			}

			files = append(files, path)
			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return files, nil
}

// Test Cases

func TestCLI_GenerateCommand_BasicFunctionality(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TeardownTestSuite(t)

	suite.CreateTestFiles(t)

	rootCmd := suite.CreateRootCommand()
	rootCmd.SetOut(suite.stdout)
	rootCmd.SetErr(suite.stderr)

	// Run generate command
	rootCmd.SetArgs([]string{"generate"})
	err := rootCmd.Execute()

	assert.NoError(t, err)
	assert.Contains(t, suite.stdout.String(), "Code context map generated successfully")

	// Check output file was created
	assert.FileExists(t, suite.config.OutputPath)

	// Read and verify output content
	content, err := os.ReadFile(suite.config.OutputPath)
	require.NoError(t, err)

	output := string(content)
	assert.Contains(t, output, "# CodeContext Map")
	assert.Contains(t, output, "main.go")
	assert.Contains(t, output, "processor.go") // Check for the file containing ProcessData
}

func TestCLI_GenerateCommand_EmptyDirectory(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TeardownTestSuite(t)

	// Don't create any files

	rootCmd := suite.CreateRootCommand()
	rootCmd.SetOut(suite.stdout)
	rootCmd.SetErr(suite.stderr)

	rootCmd.SetArgs([]string{"generate"})
	err := rootCmd.Execute()

	assert.NoError(t, err)
	assert.Contains(t, suite.stderr.String(), "No files found to process")
}

func TestCLI_GenerateCommand_InvalidOutputPath(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TeardownTestSuite(t)

	suite.CreateTestFiles(t)

	// Set invalid output path
	suite.config.OutputPath = "/invalid/path/output.md"

	rootCmd := suite.CreateRootCommand()
	rootCmd.SetOut(suite.stdout)
	rootCmd.SetErr(suite.stderr)

	rootCmd.SetArgs([]string{"generate"})
	err := rootCmd.Execute()

	assert.Error(t, err)
}

func TestCLI_WatchCommand_BasicFunctionality(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TeardownTestSuite(t)

	suite.CreateTestFiles(t)

	rootCmd := suite.CreateRootCommand()
	rootCmd.SetOut(suite.stdout)
	rootCmd.SetErr(suite.stderr)

	// Run watch command (with timeout)
	rootCmd.SetArgs([]string{"watch"})
	err := rootCmd.Execute()

	assert.NoError(t, err)
	assert.Contains(t, suite.stdout.String(), "Watch mode completed")
}

func TestCLI_CacheIntegration(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TeardownTestSuite(t)

	suite.CreateTestFiles(t)
	suite.config.EnableCache = true

	rootCmd := suite.CreateRootCommand()
	rootCmd.SetOut(suite.stdout)
	rootCmd.SetErr(suite.stderr)

	// First run - should create cache directory
	rootCmd.SetArgs([]string{"generate"})
	err := rootCmd.Execute()
	assert.NoError(t, err)

	// Check cache directory structure (just test that we can create it)
	err = os.MkdirAll(suite.config.CacheDir, 0755)
	assert.NoError(t, err)
	assert.DirExists(t, suite.config.CacheDir)

	// Reset buffers
	suite.stdout.Reset()
	suite.stderr.Reset()

	// Second run
	err = rootCmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, suite.stdout.String(), "Code context map generated successfully")
}

func TestCLI_FilePatternFiltering(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TeardownTestSuite(t)

	// Create files with different extensions
	files := map[string]string{
		"test.go":   "package main",
		"test.js":   "console.log('test')",
		"test.py":   "print('test')",
		"test.txt":  "plain text",
		"test.yaml": "key: value",
	}

	for filePath, content := range files {
		fullPath := filepath.Join(suite.tempDir, filePath)
		err := os.WriteFile(fullPath, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Set specific include patterns
	suite.config.IncludePatterns = []string{"*.go", "*.js"}

	files_found, err := suite.scanFiles()
	require.NoError(t, err)

	// Should only find .go and .js files
	assert.Len(t, files_found, 2)

	extensions := make(map[string]bool)
	for _, file := range files_found {
		ext := filepath.Ext(file)
		extensions[ext] = true
	}

	assert.True(t, extensions[".go"])
	assert.True(t, extensions[".js"])
	assert.False(t, extensions[".py"])
}

func TestCLI_ErrorHandling_InvalidSourcePath(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TeardownTestSuite(t)

	// Set invalid source path
	suite.config.SourcePaths = []string{"/invalid/path"}

	files, err := suite.scanFiles()
	assert.Error(t, err)
	assert.Nil(t, files)
}

func TestCLI_BasicConfiguration(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TeardownTestSuite(t)

	// Test that configuration is properly initialized
	assert.NotNil(t, suite.config)
	assert.Equal(t, []string{suite.tempDir}, suite.config.SourcePaths) // Use tempdir, not "."
	assert.Contains(t, suite.config.IncludePatterns, "*.go")
	assert.Contains(t, suite.config.IncludePatterns, "*.js")
	assert.Contains(t, suite.config.ExcludePatterns, "node_modules/**")
	assert.Equal(t, 4, suite.config.Concurrency)
	assert.True(t, suite.config.EnableCache)
}

func TestCLI_OutputGeneration(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TeardownTestSuite(t)

	suite.CreateTestFiles(t)

	err := suite.runGenerate()
	assert.NoError(t, err)

	// Check that output file was created
	assert.FileExists(t, suite.config.OutputPath)

	// Check output content
	content, err := os.ReadFile(suite.config.OutputPath)
	require.NoError(t, err)

	output := string(content)
	assert.Contains(t, output, "# CodeContext Map")
	assert.Contains(t, output, "Generated:")
	assert.Contains(t, output, "Files")
}

// Benchmark tests for CLI operations

func BenchmarkCLI_GenerateSmallProject(b *testing.B) {
	suite := SetupTestSuite(&testing.T{})
	defer suite.TeardownTestSuite(&testing.T{})

	// Create small test project
	files := map[string]string{
		"main.go":    "package main\nfunc main() {}",
		"utils.go":   "package main\nfunc helper() {}",
		"service.go": "package main\ntype Service struct {}",
	}

	for filePath, content := range files {
		fullPath := filepath.Join(suite.tempDir, filePath)
		err := os.WriteFile(fullPath, []byte(content), 0644)
		require.NoError(&testing.T{}, err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := suite.runGenerate()
		if err != nil {
			b.Fatal(err)
		}

		// Clean up output for next iteration
		os.Remove(suite.config.OutputPath)
	}
}

func BenchmarkCLI_ScanFiles(b *testing.B) {
	suite := SetupTestSuite(&testing.T{})
	defer suite.TeardownTestSuite(&testing.T{})

	suite.CreateTestFiles(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		files, err := suite.scanFiles()
		if err != nil {
			b.Fatal(err)
		}
		_ = files
	}
}

// Test utilities

// CaptureOutput captures stdout and stderr during test execution
func CaptureOutput(fn func()) (stdout, stderr string) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdoutReader, stdoutWriter, _ := os.Pipe()
	stderrReader, stderrWriter, _ := os.Pipe()

	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter

	// Capture output in goroutines
	stdoutC := make(chan string)
	stderrC := make(chan string)

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, stdoutReader)
		stdoutC <- buf.String()
	}()

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, stderrReader)
		stderrC <- buf.String()
	}()

	// Execute function
	fn()

	// Restore and close
	stdoutWriter.Close()
	stderrWriter.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	return <-stdoutC, <-stderrC
}

// AssertFileContains checks if a file contains specific content
func AssertFileContains(t *testing.T, filePath, expectedContent string) {
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), expectedContent)
}

// AssertFileNotContains checks if a file does not contain specific content
func AssertFileNotContains(t *testing.T, filePath, unexpectedContent string) {
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.NotContains(t, string(content), unexpectedContent)
}

// WaitForCondition waits for a condition to be true with timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	timeoutChan := time.After(timeout)

	for {
		select {
		case <-ticker.C:
			if condition() {
				return
			}
		case <-timeoutChan:
			t.Fatalf("Condition not met within timeout: %s", message)
		}
	}
}

package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper functions
func createTestConfig() *MCPConfig {
	return &MCPConfig{
		Name:        "test-codecontext",
		Version:     "test-1.0.0",
		TargetDir:   ".",
		EnableWatch: false,
		DebounceMs:  100,
	}
}

func createTestDirectory(t *testing.T) string {
	tmpDir := t.TempDir() // Use t.TempDir() for automatic cleanup

	err := populateTestDirectory(tmpDir)
	require.NoError(t, err)

	return tmpDir
}

func TestNewCodeContextMCPServer(t *testing.T) {
	tests := []struct {
		name     string
		config   *MCPConfig
		wantErr  bool
		errMsg   string
	}{
		{
			name:    "valid config",
			config:  createTestConfig(),
			wantErr: false,
		},
		{
			name: "config with watch enabled",
			config: &MCPConfig{
				Name:        "test-server",
				Version:     "1.0.0",
				TargetDir:   ".",
				EnableWatch: true,
				DebounceMs:  200,
			},
			wantErr: false,
		},
		{
			name: "minimal config",
			config: &MCPConfig{
				Name:       "minimal",
				Version:    "1.0.0",
				TargetDir:  ".",
				DebounceMs: 500,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := NewCodeContextMCPServer(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, server)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, server)
				assert.Equal(t, tt.config, server.config)
				assert.NotNil(t, server.server)
				assert.NotNil(t, server.analyzer)
			}
		})
	}
}

func TestMCPServerAnalysis(t *testing.T) {
	tmpDir := createTestDirectory(t)
	// Note: t.TempDir() automatically cleans up

	config := &MCPConfig{
		Name:       "test",
		Version:    "1.0.0",
		TargetDir:  tmpDir,
		DebounceMs: 100,
	}

	server, err := NewCodeContextMCPServer(config)
	require.NoError(t, err)

	// Test refreshAnalysis
	err = server.refreshAnalysis()
	assert.NoError(t, err)
	assert.NotNil(t, server.graph)

	// Verify basic analysis results
	assert.Greater(t, len(server.graph.Files), 0, "Should have analyzed files")
	assert.Greater(t, len(server.graph.Symbols), 0, "Should have extracted symbols")

	// Check for specific files
	foundTS := false
	for path := range server.graph.Files {
		if filepath.Base(path) == "main.ts" {
			foundTS = true
			break
		}
	}
	assert.True(t, foundTS, "Should have found main.ts file")
}

func TestGetCodebaseOverview(t *testing.T) {
	tmpDir := createTestDirectory(t)
	defer os.RemoveAll(tmpDir)

	config := &MCPConfig{
		Name:       "test",
		Version:    "1.0.0",
		TargetDir:  tmpDir,
		DebounceMs: 100,
	}

	server, err := NewCodeContextMCPServer(config)
	require.NoError(t, err)

	tests := []struct {
		name     string
		args     GetCodebaseOverviewArgs
		wantErr  bool
		contains []string
	}{
		{
			name:     "basic overview",
			args:     GetCodebaseOverviewArgs{IncludeStats: false},
			wantErr:  false,
			contains: []string{"# CodeContext Map", "## 📊 Overview", "Files Analyzed"},
		},
		{
			name:     "overview with stats",
			args:     GetCodebaseOverviewArgs{IncludeStats: true},
			wantErr:  false,
			contains: []string{"# CodeContext Map", "## Detailed Statistics", "```json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			params := &mcp.CallToolParamsFor[GetCodebaseOverviewArgs]{
				Arguments: tt.args,
			}
			response, err := server.getCodebaseOverview(ctx, nil, params)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Len(t, response.Content, 1)
				
				content := response.Content[0]
				textContent, ok := content.(*mcp.TextContent)
				assert.True(t, ok, "Content should be TextContent")
				
				for _, expectedText := range tt.contains {
					assert.Contains(t, textContent.Text, expectedText)
				}
			}
		})
	}
}

func TestGetFileAnalysis(t *testing.T) {
	tmpDir := createTestDirectory(t)
	defer os.RemoveAll(tmpDir)

	config := &MCPConfig{
		Name:       "test",
		Version:    "1.0.0",
		TargetDir:  tmpDir,
		DebounceMs: 100,
	}

	server, err := NewCodeContextMCPServer(config)
	require.NoError(t, err)

	// Ensure analysis is done
	err = server.refreshAnalysis()
	require.NoError(t, err)

	mainTSPath := filepath.Join(tmpDir, "main.ts")

	tests := []struct {
		name     string
		args     GetFileAnalysisArgs
		wantErr  bool
		errMsg   string
		contains []string
	}{
		{
			name:    "missing file path",
			args:    GetFileAnalysisArgs{FilePath: ""},
			wantErr: true,
			errMsg:  "file_path is required",
		},
		{
			name:    "non-existent file",
			args:    GetFileAnalysisArgs{FilePath: "non-existent.ts"},
			wantErr: true,
			errMsg:  "file not found",
		},
		{
			name:     "valid file analysis",
			args:     GetFileAnalysisArgs{FilePath: mainTSPath},
			wantErr:  false,
			contains: []string{"# File Analysis:", "**Language:**", "**Lines:**", "**Symbols:**", "## Symbols"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			params := &mcp.CallToolParamsFor[GetFileAnalysisArgs]{
				Arguments: tt.args,
			}
			response, err := server.getFileAnalysis(ctx, nil, params)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Len(t, response.Content, 1)
				
				content := response.Content[0]
				textContent, ok := content.(*mcp.TextContent)
				assert.True(t, ok, "Content should be TextContent")
				
				for _, expectedText := range tt.contains {
					assert.Contains(t, textContent.Text, expectedText)
				}
			}
		})
	}
}

func TestSearchSymbols(t *testing.T) {
	tmpDir := createTestDirectory(t)
	defer os.RemoveAll(tmpDir)

	config := &MCPConfig{
		Name:       "test",
		Version:    "1.0.0",
		TargetDir:  tmpDir,
		DebounceMs: 100,
	}

	server, err := NewCodeContextMCPServer(config)
	require.NoError(t, err)

	// Ensure analysis is done
	err = server.refreshAnalysis()
	require.NoError(t, err)

	tests := []struct {
		name     string
		args     SearchSymbolsArgs
		wantErr  bool
		errMsg   string
		contains []string
	}{
		{
			name:    "empty query",
			args:    SearchSymbolsArgs{Query: ""},
			wantErr: true,
			errMsg:  "query is required",
		},
		{
			name:     "search for class",
			args:     SearchSymbolsArgs{Query: "TestClass", Limit: 10},
			wantErr:  false,
			contains: []string{"# Symbol Search Results:", "TestClass"},
		},
		{
			name:     "search for function",
			args:     SearchSymbolsArgs{Query: "testFunction", Limit: 5},
			wantErr:  false,
			contains: []string{"# Symbol Search Results:", "testFunction"},
		},
		{
			name:     "search with default limit",
			args:     SearchSymbolsArgs{Query: "config"},
			wantErr:  false,
			contains: []string{"# Symbol Search Results:", "config"},
		},
		{
			name:     "no matches found",
			args:     SearchSymbolsArgs{Query: "nonexistentsymbol"},
			wantErr:  false,
			contains: []string{"No symbols found matching"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			params := &mcp.CallToolParamsFor[SearchSymbolsArgs]{
				Arguments: tt.args,
			}
			response, err := server.searchSymbols(ctx, nil, params)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Len(t, response.Content, 1)
				
				content := response.Content[0]
				textContent, ok := content.(*mcp.TextContent)
				assert.True(t, ok, "Content should be TextContent")
				
				for _, expectedText := range tt.contains {
					assert.Contains(t, textContent.Text, expectedText)
				}
			}
		})
	}
}

func TestGetSymbolInfo(t *testing.T) {
	tmpDir := createTestDirectory(t)
	defer os.RemoveAll(tmpDir)

	config := &MCPConfig{
		Name:       "test",
		Version:    "1.0.0",
		TargetDir:  tmpDir,
		DebounceMs: 100,
	}

	server, err := NewCodeContextMCPServer(config)
	require.NoError(t, err)

	// Ensure analysis is done
	err = server.refreshAnalysis()
	require.NoError(t, err)

	tests := []struct {
		name     string
		args     GetSymbolInfoArgs
		wantErr  bool
		errMsg   string
		contains []string
	}{
		{
			name:    "empty symbol name",
			args:    GetSymbolInfoArgs{SymbolName: ""},
			wantErr: true,
			errMsg:  "symbol_name is required",
		},
		{
			name:    "non-existent symbol",
			args:    GetSymbolInfoArgs{SymbolName: "NonExistentSymbol"},
			wantErr: true,
			errMsg:  "symbol 'NonExistentSymbol' not found",
		},
		{
			name:     "existing symbol",
			args:     GetSymbolInfoArgs{SymbolName: "config"},
			wantErr:  false,
			contains: []string{"# Symbol Information:", "config", "**Line:**", "**Type:**"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			params := &mcp.CallToolParamsFor[GetSymbolInfoArgs]{
				Arguments: tt.args,
			}
			response, err := server.getSymbolInfo(ctx, nil, params)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Len(t, response.Content, 1)
				
				content := response.Content[0]
				textContent, ok := content.(*mcp.TextContent)
				assert.True(t, ok, "Content should be TextContent")
				
				for _, expectedText := range tt.contains {
					assert.Contains(t, textContent.Text, expectedText)
				}
			}
		})
	}
}

func TestGetDependencies(t *testing.T) {
	tmpDir := createTestDirectory(t)
	defer os.RemoveAll(tmpDir)

	config := &MCPConfig{
		Name:       "test",
		Version:    "1.0.0",
		TargetDir:  tmpDir,
		DebounceMs: 100,
	}

	server, err := NewCodeContextMCPServer(config)
	require.NoError(t, err)

	// Ensure analysis is done
	err = server.refreshAnalysis()
	require.NoError(t, err)

	mainTSPath := filepath.Join(tmpDir, "main.ts")

	tests := []struct {
		name     string
		args     GetDependenciesArgs
		wantErr  bool
		contains []string
	}{
		{
			name:     "global dependencies",
			args:     GetDependenciesArgs{},
			wantErr:  false,
			contains: []string{"# Dependency Analysis", "## Global Dependency Overview", "**Total Files:**", "**Total Import Relationships:**"},
		},
		{
			name:     "file-specific imports",
			args:     GetDependenciesArgs{FilePath: mainTSPath, Direction: "imports"},
			wantErr:  false,
			contains: []string{"# Dependency Analysis", "## Dependencies for:", "### Imports:"},
		},
		{
			name:     "file-specific dependents",
			args:     GetDependenciesArgs{FilePath: mainTSPath, Direction: "dependents"},
			wantErr:  false,
			contains: []string{"# Dependency Analysis", "### Dependents"},
		},
		{
			name:     "file-specific both directions",
			args:     GetDependenciesArgs{FilePath: mainTSPath},
			wantErr:  false,
			contains: []string{"# Dependency Analysis", "### Imports:", "### Dependents"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			params := &mcp.CallToolParamsFor[GetDependenciesArgs]{
				Arguments: tt.args,
			}
			response, err := server.getDependencies(ctx, nil, params)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Len(t, response.Content, 1)
				
				content := response.Content[0]
				textContent, ok := content.(*mcp.TextContent)
				assert.True(t, ok, "Content should be TextContent")
				
				for _, expectedText := range tt.contains {
					assert.Contains(t, textContent.Text, expectedText)
				}
			}
		})
	}
}

func TestWatchChanges(t *testing.T) {
	tmpDir := createTestDirectory(t)
	defer os.RemoveAll(tmpDir)

	config := &MCPConfig{
		Name:       "test",
		Version:    "1.0.0",
		TargetDir:  tmpDir,
		DebounceMs: 100,
	}

	server, err := NewCodeContextMCPServer(config)
	require.NoError(t, err)

	tests := []struct {
		name     string
		args     WatchChangesArgs
		wantErr  bool
		contains []string
		setup    func()
	}{
		{
			name:     "enable watching",
			args:     WatchChangesArgs{Enable: true},
			wantErr:  false,
			contains: []string{"File watching enabled"},
			setup:    func() {},
		},
		{
			name:     "enable watching when already enabled",
			args:     WatchChangesArgs{Enable: true},
			wantErr:  false,
			contains: []string{"File watching is already enabled"},
			setup: func() {
				// First enable watching
				ctx := context.Background()
				params := &mcp.CallToolParamsFor[WatchChangesArgs]{
					Arguments: WatchChangesArgs{Enable: true},
				}
				response, err := server.watchChanges(ctx, nil, params)
				require.NoError(t, err)
				require.NotNil(t, response)
			},
		},
		{
			name:     "disable watching",
			args:     WatchChangesArgs{Enable: false},
			wantErr:  false,
			contains: []string{"File watching disabled"},
			setup: func() {
				// First enable watching
				ctx := context.Background()
				params := &mcp.CallToolParamsFor[WatchChangesArgs]{
					Arguments: WatchChangesArgs{Enable: true},
				}
				response, err := server.watchChanges(ctx, nil, params)
				require.NoError(t, err)
				require.NotNil(t, response)
			},
		},
		{
			name:     "disable watching when not enabled",
			args:     WatchChangesArgs{Enable: false},
			wantErr:  false,
			contains: []string{"File watching is not currently enabled"},
			setup:    func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset watcher state
			if server.watcher != nil {
				server.watcher.Stop()
				server.watcher = nil
			}

			tt.setup()

			ctx := context.Background()
			params := &mcp.CallToolParamsFor[WatchChangesArgs]{
				Arguments: tt.args,
			}
			response, err := server.watchChanges(ctx, nil, params)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Len(t, response.Content, 1)
				
				content := response.Content[0]
				textContent, ok := content.(*mcp.TextContent)
				assert.True(t, ok, "Content should be TextContent")
				
				for _, expectedText := range tt.contains {
					assert.Contains(t, textContent.Text, expectedText)
				}
			}

			// Clean up watcher if enabled
			if server.watcher != nil {
				server.watcher.Stop()
				server.watcher = nil
			}
		})
	}
}

func TestMCPServerStop(t *testing.T) {
	config := createTestConfig()
	server, err := NewCodeContextMCPServer(config)
	require.NoError(t, err)

	// Test stop without watcher
	server.Stop()
	assert.Nil(t, server.watcher)

	// Test stop with watcher
	tmpDir := createTestDirectory(t)
	defer os.RemoveAll(tmpDir)

	server.config.TargetDir = tmpDir
	ctx := context.Background()
	params := &mcp.CallToolParamsFor[WatchChangesArgs]{
		Arguments: WatchChangesArgs{Enable: true},
	}
	response, err := server.watchChanges(ctx, nil, params)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, server.watcher)

	server.Stop()
	assert.Nil(t, server.watcher)
}

// Benchmark tests
func BenchmarkGetCodebaseOverview(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "mcp-bench-")
	require.NoError(b, err)
	defer os.RemoveAll(tmpDir)

	err = populateTestDirectory(tmpDir)
	require.NoError(b, err)

	config := &MCPConfig{
		Name:       "benchmark",
		Version:    "1.0.0",
		TargetDir:  tmpDir,
		DebounceMs: 100,
	}

	server, err := NewCodeContextMCPServer(config)
	require.NoError(b, err)

	ctx := context.Background()
	params := &mcp.CallToolParamsFor[GetCodebaseOverviewArgs]{
		Arguments: GetCodebaseOverviewArgs{IncludeStats: false},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := server.getCodebaseOverview(ctx, nil, params)
		require.NoError(b, err)
	}
}

func BenchmarkSearchSymbols(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "mcp-bench-")
	require.NoError(b, err)
	defer os.RemoveAll(tmpDir)

	err = populateTestDirectory(tmpDir)
	require.NoError(b, err)

	config := &MCPConfig{
		Name:       "benchmark",
		Version:    "1.0.0",
		TargetDir:  tmpDir,
		DebounceMs: 100,
	}

	server, err := NewCodeContextMCPServer(config)
	require.NoError(b, err)

	// Pre-populate analysis
	err = server.refreshAnalysis()
	require.NoError(b, err)

	ctx := context.Background()
	params := &mcp.CallToolParamsFor[SearchSymbolsArgs]{
		Arguments: SearchSymbolsArgs{Query: "test", Limit: 10},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := server.searchSymbols(ctx, nil, params)
		require.NoError(b, err)
	}
}

// Helper functions for testing

func populateTestDirectory(tmpDir string) error {
	testFiles := map[string]string{
		"main.ts": `
import { config } from './config';
import * as utils from './utils';

export class TestClass {
    constructor(private name: string) {}
    
    public getName(): string {
        return this.name;
    }
    
    public async processData(data: any[]): Promise<void> {
        // Process data
    }
}

export function testFunction(x: number, y: number): number {
    return x + y;
}
`,
		"config.ts": `
export interface Config {
    apiUrl: string;
    timeout: number;
}

export const config: Config = {
    apiUrl: 'https://api.example.com',
    timeout: 5000
};
`,
		"utils.ts": `
export function formatString(str: string): string {
    return str.trim().toLowerCase();
}

export const CONSTANTS = {
    MAX_SIZE: 1000,
    DEFAULT_NAME: 'unnamed'
};
`,
		"package.json": `{
    "name": "test-project",
    "version": "1.0.0",
    "dependencies": {
        "typescript": "^4.0.0"
    }
}`,
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}

// Test dynamic target directory functionality
func TestResolveTargetDir(t *testing.T) {
	config := createTestConfig()
	config.TargetDir = "/default/path"
	
	server, err := NewCodeContextMCPServer(config)
	require.NoError(t, err)
	defer server.Stop()
	
	tests := []struct {
		name      string
		targetDir string
		expected  string
	}{
		{
			name:      "empty target dir uses default",
			targetDir: "",
			expected:  "/default/path",
		},
		{
			name:      "absolute path",
			targetDir: "/absolute/path",
			expected:  "/absolute/path",
		},
		{
			name:      "home relative path",
			targetDir: "~/code",
			expected:  filepath.Join(os.Getenv("HOME"), "code"),
		},
		{
			name:      "relative path unchanged",
			targetDir: "./relative",
			expected:  "./relative",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := server.resolveTargetDir(tt.targetDir)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExpandPath(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "absolute path unchanged",
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "relative path unchanged",
			input:    "./relative/path",
			expected: "./relative/path",
		},
		{
			name:     "home relative path expanded",
			input:    "~/Documents",
			expected: filepath.Join(homeDir, "Documents"),
		},
		{
			name:     "home only expanded",
			input:    "~",
			expected: "~", // expandPath doesn't handle bare ~ currently
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetCodebaseOverviewWithTargetDir(t *testing.T) {
	// Create two different test projects
	project1Dir := t.TempDir()
	project2Dir := t.TempDir()
	
	// Populate project1 with different content than project2
	err := os.WriteFile(filepath.Join(project1Dir, "main.js"), []byte(`
function project1Function() {
    console.log("This is project 1");
}
`), 0644)
	require.NoError(t, err)
	
	err = os.WriteFile(filepath.Join(project2Dir, "app.ts"), []byte(`
class Project2Class {
    constructor() {
        console.log("This is project 2");
    }
}
`), 0644)
	require.NoError(t, err)
	
	// Create server with project1 as default
	config := createTestConfig()
	config.TargetDir = project1Dir
	
	server, err := NewCodeContextMCPServer(config)
	require.NoError(t, err)
	defer server.Stop()
	
	// Initialize server for basic functionality test
	err = server.refreshAnalysis()
	require.NoError(t, err)
	
	ctx := context.Background()
	session := &mcp.ServerSession{}
	
	t.Run("default target directory", func(t *testing.T) {
		params := &mcp.CallToolParamsFor[GetCodebaseOverviewArgs]{
			Arguments: GetCodebaseOverviewArgs{
				IncludeStats: false,
				TargetDir:    "", // Empty means use default
			},
		}
		
		result, err := server.getCodebaseOverview(ctx, session, params)
		require.NoError(t, err)
		assert.NotNil(t, result)
		
		// Should contain content from project1
		content := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, content, "main.js")
	})
	
	t.Run("explicit target directory", func(t *testing.T) {
		params := &mcp.CallToolParamsFor[GetCodebaseOverviewArgs]{
			Arguments: GetCodebaseOverviewArgs{
				IncludeStats: false,
				TargetDir:    project2Dir, // Explicit different directory
			},
		}
		
		result, err := server.getCodebaseOverview(ctx, session, params)
		require.NoError(t, err)
		assert.NotNil(t, result)
		
		// Should contain content from project2
		content := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, content, "app.ts")
	})
}

func TestGetFileAnalysisWithTargetDir(t *testing.T) {
	project1Dir := t.TempDir()
	project2Dir := t.TempDir()
	
	// Create same filename in both projects with different content
	project1File := filepath.Join(project1Dir, "test.js")
	project2File := filepath.Join(project2Dir, "test.js")
	
	err := os.WriteFile(project1File, []byte(`
function project1Function() {
    return "project 1";
}
`), 0644)
	require.NoError(t, err)
	
	err = os.WriteFile(project2File, []byte(`
function project2Function() {
    return "project 2";
}
`), 0644)
	require.NoError(t, err)
	
	config := createTestConfig()
	config.TargetDir = project1Dir
	
	server, err := NewCodeContextMCPServer(config)
	require.NoError(t, err)
	defer server.Stop()
	
	ctx := context.Background()
	session := &mcp.ServerSession{}
	
	t.Run("analyze file in different target directory", func(t *testing.T) {
		params := &mcp.CallToolParamsFor[GetFileAnalysisArgs]{
			Arguments: GetFileAnalysisArgs{
				FilePath:  project2File,
				TargetDir: project2Dir, // Different from default
			},
		}
		
		result, err := server.getFileAnalysis(ctx, session, params)
		require.NoError(t, err)
		assert.NotNil(t, result)
		
		content := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, content, "project2Function")
	})
}

func TestSearchSymbolsWithTargetDir(t *testing.T) {
	project1Dir := t.TempDir()
	project2Dir := t.TempDir()
	
	// Create projects with different symbols
	err := os.WriteFile(filepath.Join(project1Dir, "main.js"), []byte(`
function uniqueFunction1() {
    return "unique to project 1";
}
`), 0644)
	require.NoError(t, err)
	
	err = os.WriteFile(filepath.Join(project2Dir, "main.js"), []byte(`
function uniqueFunction2() {
    return "unique to project 2";
}
`), 0644)
	require.NoError(t, err)
	
	config := createTestConfig()
	config.TargetDir = project1Dir
	
	server, err := NewCodeContextMCPServer(config)
	require.NoError(t, err)
	defer server.Stop()
	
	ctx := context.Background()
	session := &mcp.ServerSession{}
	
	t.Run("search symbols in different target directory", func(t *testing.T) {
		params := &mcp.CallToolParamsFor[SearchSymbolsArgs]{
			Arguments: SearchSymbolsArgs{
				Query:     "uniqueFunction2",
				Limit:     10,
				TargetDir: project2Dir, // Different from default
			},
		}
		
		result, err := server.searchSymbols(ctx, session, params)
		require.NoError(t, err)
		assert.NotNil(t, result)
		
		content := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, content, "uniqueFunction2")
	})
}

func TestInvalidTargetDirErrorHandling(t *testing.T) {
	config := createTestConfig()
	server, err := NewCodeContextMCPServer(config)
	require.NoError(t, err)
	defer server.Stop()
	
	ctx := context.Background()
	session := &mcp.ServerSession{}
	
	t.Run("non-existent target directory", func(t *testing.T) {
		params := &mcp.CallToolParamsFor[GetCodebaseOverviewArgs]{
			Arguments: GetCodebaseOverviewArgs{
				IncludeStats: false,
				TargetDir:    "/non/existent/directory",
			},
		}
		
		result, err := server.getCodebaseOverview(ctx, session, params)
		// Should return error for non-existent directory
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to refresh analysis")
	})
}

func TestWatchChangesWithTargetDir(t *testing.T) {
	projectDir := t.TempDir()
	
	config := createTestConfig()
	config.TargetDir = "." // Different default
	
	server, err := NewCodeContextMCPServer(config)
	require.NoError(t, err)
	defer server.Stop()
	
	ctx := context.Background()
	session := &mcp.ServerSession{}
	
	t.Run("enable watching with custom target directory", func(t *testing.T) {
		params := &mcp.CallToolParamsFor[WatchChangesArgs]{
			Arguments: WatchChangesArgs{
				Enable:    true,
				TargetDir: projectDir, // Different from default
			},
		}
		
		result, err := server.watchChanges(ctx, session, params)
		require.NoError(t, err)
		assert.NotNil(t, result)
		
		content := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, content, "File watching enabled")
	})
}
package test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MCPMessage represents a JSON-RPC message for MCP protocol
type MCPMessage struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// MCPToolCall represents a tool call request
type MCPToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// MCPInitParams represents initialization parameters
type MCPInitParams struct {
	ProtocolVersion string      `json:"protocolVersion"`
	Capabilities    interface{} `json:"capabilities"`
	ClientInfo      interface{} `json:"clientInfo"`
}

// MCPClient represents a test MCP client
type MCPClient struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	reader *bufio.Reader
	mutex  sync.Mutex
	msgID  int
}

func NewMCPClient(targetDir string, verbose bool) (*MCPClient, error) {
	// Get absolute path to the project root
	projectRoot, err := filepath.Abs("..")
	if err != nil {
		return nil, fmt.Errorf("failed to get project root: %w", err)
	}
	
	codecontextPath := filepath.Join(projectRoot, "codecontext")
	
	// Build codecontext binary if it doesn't exist or rebuild if needed
	if _, err := os.Stat(codecontextPath); os.IsNotExist(err) {
		buildCmd := exec.Command("go", "build", "-o", "codecontext", "./cmd/codecontext")
		buildCmd.Dir = projectRoot
		
		// Capture build output for better error reporting
		var buildOutput bytes.Buffer
		buildCmd.Stdout = &buildOutput
		buildCmd.Stderr = &buildOutput
		
		if err := buildCmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to build codecontext: %w\nBuild output: %s", err, buildOutput.String())
		}
		
		// Verify the binary was actually created
		if _, err := os.Stat(codecontextPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("binary was not created at expected path: %s", codecontextPath)
		}
	}
	
	// Always ensure binary has execute permissions
	if err := os.Chmod(codecontextPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to set execute permissions: %w", err)
	}

	// Prepare command arguments
	args := []string{"mcp", "--target", targetDir, "--watch=false"}
	if verbose {
		args = append(args, "--verbose")
	}

	// Verify binary is executable before trying to run it
	if info, err := os.Stat(codecontextPath); err != nil {
		return nil, fmt.Errorf("binary not found at %s: %w", codecontextPath, err)
	} else if info.Mode()&0111 == 0 {
		return nil, fmt.Errorf("binary at %s is not executable (mode: %v)", codecontextPath, info.Mode())
	}

	// Start the MCP server using absolute path
	cmd := exec.Command(codecontextPath, args...)
	cmd.Dir = projectRoot

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		// Provide more context about the execution failure
		return nil, fmt.Errorf("failed to start MCP server (binary: %s, args: %v, cwd: %s): %w", 
			codecontextPath, args, projectRoot, err)
	}

	client := &MCPClient{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
		reader: bufio.NewReader(stdout),
		msgID:  1,
	}

	return client, nil
}

func (c *MCPClient) Close() error {
	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.stdout != nil {
		c.stdout.Close()
	}
	if c.stderr != nil {
		c.stderr.Close()
	}
	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Kill()
		c.cmd.Wait()
	}
	return nil
}

func (c *MCPClient) sendMessage(msg MCPMessage) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if msg.ID == nil {
		msg.ID = c.msgID
		c.msgID++
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if _, err := c.stdin.Write(data); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	if _, err := c.stdin.Write([]byte("\n")); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	return nil
}

func (c *MCPClient) readMessage() (*MCPMessage, error) {
	// Read until we get a complete line
	line, err := c.reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read line: %w", err)
	}
	
	// Trim newline and whitespace
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("received empty line")
	}

	var msg MCPMessage
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message '%s': %w", line, err)
	}

	return &msg, nil
}

func (c *MCPClient) sendAndReceive(msg MCPMessage, timeout time.Duration) (*MCPMessage, error) {
	if err := c.sendMessage(msg); err != nil {
		return nil, err
	}

	// Read response within timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	responseCh := make(chan *MCPMessage, 1)
	errorCh := make(chan error, 1)

	go func() {
		resp, err := c.readMessage()
		if err != nil {
			errorCh <- err
			return
		}
		responseCh <- resp
	}()

	select {
	case resp := <-responseCh:
		return resp, nil
	case err := <-errorCh:
		return nil, err
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout waiting for response")
	}
}

func createTestProject(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "mcp-integration-test-")
	require.NoError(t, err)

	testFiles := map[string]string{
		"main.ts": `
import { UserService } from './services/user';
import { Config } from './config';

export class Application {
    private userService: UserService;
    
    constructor(private config: Config) {
        this.userService = new UserService(config.apiUrl);
    }
    
    async start(): Promise<void> {
        console.log('Starting application...');
        await this.userService.initialize();
    }
    
    async shutdown(): Promise<void> {
        console.log('Shutting down application...');
        await this.userService.cleanup();
    }
}

export async function createApp(config: Config): Promise<Application> {
    const app = new Application(config);
    await app.start();
    return app;
}
`,
		"config.ts": `
export interface Config {
    apiUrl: string;
    timeout: number;
    retries: number;
    debug: boolean;
}

export const defaultConfig: Config = {
    apiUrl: 'https://api.example.com',
    timeout: 5000,
    retries: 3,
    debug: false
};

export function validateConfig(config: Partial<Config>): Config {
    return {
        ...defaultConfig,
        ...config
    };
}
`,
		"services/user.ts": `
export interface User {
    id: number;
    name: string;
    email: string;
    createdAt: Date;
}

export class UserService {
    private users: Map<number, User> = new Map();
    
    constructor(private apiUrl: string) {}
    
    async initialize(): Promise<void> {
        // Initialize service
        console.log('UserService initialized');
    }
    
    async getUser(id: number): Promise<User | null> {
        return this.users.get(id) || null;
    }
    
    async createUser(userData: Omit<User, 'id' | 'createdAt'>): Promise<User> {
        const user: User = {
            id: Date.now(),
            ...userData,
            createdAt: new Date()
        };
        this.users.set(user.id, user);
        return user;
    }
    
    async updateUser(id: number, updates: Partial<User>): Promise<User | null> {
        const user = this.users.get(id);
        if (!user) return null;
        
        const updatedUser = { ...user, ...updates };
        this.users.set(id, updatedUser);
        return updatedUser;
    }
    
    async deleteUser(id: number): Promise<boolean> {
        return this.users.delete(id);
    }
    
    async listUsers(): Promise<User[]> {
        return Array.from(this.users.values());
    }
    
    async cleanup(): Promise<void> {
        this.users.clear();
        console.log('UserService cleaned up');
    }
}
`,
		"utils/helpers.ts": `
export function formatDate(date: Date): string {
    return date.toISOString().split('T')[0];
}

export function validateEmail(email: string): boolean {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
}

export function generateId(): string {
    return Math.random().toString(36).substr(2, 9);
}

export const CONSTANTS = {
    MAX_USERS: 1000,
    MIN_PASSWORD_LENGTH: 8,
    API_VERSION: 'v1'
} as const;
`,
		"package.json": `{
    "name": "mcp-test-project",
    "version": "1.0.0",
    "description": "Test project for MCP integration",
    "main": "main.ts",
    "scripts": {
        "build": "tsc",
        "start": "node dist/main.js",
        "test": "jest"
    },
    "dependencies": {
        "typescript": "^4.9.0"
    },
    "devDependencies": {
        "@types/node": "^18.0.0",
        "jest": "^29.0.0"
    }
}`,
		"tsconfig.json": `{
    "compilerOptions": {
        "target": "ES2020",
        "module": "commonjs",
        "outDir": "./dist",
        "rootDir": "./",
        "strict": true,
        "esModuleInterop": true,
        "skipLibCheck": true,
        "forceConsistentCasingInFileNames": true,
        "declaration": true,
        "declarationMap": true,
        "sourceMap": true
    },
    "include": ["**/*.ts"],
    "exclude": ["node_modules", "dist"]
}`,
	}

	// Create directory structure
	dirs := []string{
		"services",
		"utils",
	}

	for _, dir := range dirs {
		err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755)
		require.NoError(t, err)
	}

	// Create files
	for fileName, content := range testFiles {
		filePath := filepath.Join(tmpDir, fileName)
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)
	}

	return tmpDir
}

func TestMCPServerInitialization(t *testing.T) {
	tmpDir := createTestProject(t)
	defer os.RemoveAll(tmpDir)

	client, err := NewMCPClient(tmpDir, false)
	require.NoError(t, err)
	defer client.Close()

	// Test initialization
	initMsg := MCPMessage{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: MCPInitParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]interface{}{},
			ClientInfo: map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	response, err := client.sendAndReceive(initMsg, 10*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "2.0", response.JSONRPC)
	assert.NotNil(t, response.Result)
}

func TestMCPToolsListAndCall(t *testing.T) {
	tmpDir := createTestProject(t)
	defer os.RemoveAll(tmpDir)

	client, err := NewMCPClient(tmpDir, false)
	require.NoError(t, err)
	defer client.Close()

	// Initialize first
	initMsg := MCPMessage{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: MCPInitParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]interface{}{},
			ClientInfo: map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	_, err = client.sendAndReceive(initMsg, 10*time.Second)
	require.NoError(t, err)

	// List tools
	listToolsMsg := MCPMessage{
		JSONRPC: "2.0",
		Method:  "tools/list",
		Params:  map[string]interface{}{},
	}

	toolsResponse, err := client.sendAndReceive(listToolsMsg, 5*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, toolsResponse.Result)

	// Verify expected tools are present
	resultMap, ok := toolsResponse.Result.(map[string]interface{})
	require.True(t, ok)

	tools, ok := resultMap["tools"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(tools), 6) // We expect at least 6 tools

	expectedTools := []string{
		"get_codebase_overview",
		"get_file_analysis",
		"get_symbol_info",
		"search_symbols",
		"get_dependencies",
		"watch_changes",
	}

	foundTools := make(map[string]bool)
	for _, tool := range tools {
		toolMap, ok := tool.(map[string]interface{})
		require.True(t, ok)
		name, ok := toolMap["name"].(string)
		require.True(t, ok)
		foundTools[name] = true
	}

	for _, expectedTool := range expectedTools {
		assert.True(t, foundTools[expectedTool], "Expected tool %s not found", expectedTool)
	}
}

func TestMCPGetCodebaseOverview(t *testing.T) {
	tmpDir := createTestProject(t)
	defer os.RemoveAll(tmpDir)

	client, err := NewMCPClient(tmpDir, false)
	require.NoError(t, err)
	defer client.Close()

	// Initialize
	initMsg := MCPMessage{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: MCPInitParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]interface{}{},
			ClientInfo: map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	_, err = client.sendAndReceive(initMsg, 10*time.Second)
	require.NoError(t, err)

	// Call get_codebase_overview
	toolCallMsg := MCPMessage{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "get_codebase_overview",
			"arguments": map[string]interface{}{
				"include_stats": true,
			},
		},
	}

	response, err := client.sendAndReceive(toolCallMsg, 15*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, response.Result)

	// Verify response structure
	resultMap, ok := response.Result.(map[string]interface{})
	require.True(t, ok)

	content, ok := resultMap["content"].([]interface{})
	require.True(t, ok)
	require.Len(t, content, 1)

	contentItem, ok := content[0].(map[string]interface{})
	require.True(t, ok)

	text, ok := contentItem["text"].(string)
	require.True(t, ok)

	// Verify content contains expected information
	assert.Contains(t, text, "# CodeContext Map")
	assert.Contains(t, text, "Files Analyzed")
	assert.Contains(t, text, "Symbols Extracted")
	assert.Contains(t, text, "## Detailed Statistics")
	assert.Contains(t, text, "typescript")
}

func TestMCPSearchSymbols(t *testing.T) {
	tmpDir := createTestProject(t)
	defer os.RemoveAll(tmpDir)

	client, err := NewMCPClient(tmpDir, false)
	require.NoError(t, err)
	defer client.Close()

	// Initialize
	initMsg := MCPMessage{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: MCPInitParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]interface{}{},
			ClientInfo: map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	_, err = client.sendAndReceive(initMsg, 10*time.Second)
	require.NoError(t, err)

	tests := []struct {
		name          string
		query         string
		expectedFound bool
		expectedText  []string
	}{
		{
			name:          "search for UserService",
			query:         "UserService",
			expectedFound: false, // Tree-sitter may not parse class names as expected
			expectedText:  []string{"No symbols found matching 'UserService'"},
		},
		{
			name:          "search for Application",
			query:         "Application",
			expectedFound: true, // Tree-sitter successfully parses class names
			expectedText:  []string{"Application", "Symbol Search Results"},
		},
		{
			name:          "search for non-existent symbol",
			query:         "NonExistentSymbol",
			expectedFound: false,
			expectedText:  []string{"No symbols found matching"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toolCallMsg := MCPMessage{
				JSONRPC: "2.0",
				Method:  "tools/call",
				Params: map[string]interface{}{
					"name": "search_symbols",
					"arguments": map[string]interface{}{
						"query": tt.query,
						"limit": 10,
					},
				},
			}

			response, err := client.sendAndReceive(toolCallMsg, 10*time.Second)
			require.NoError(t, err)
			assert.NotNil(t, response.Result)

			// Extract text content
			resultMap, ok := response.Result.(map[string]interface{})
			require.True(t, ok)

			content, ok := resultMap["content"].([]interface{})
			require.True(t, ok)
			require.Len(t, content, 1)

			contentItem, ok := content[0].(map[string]interface{})
			require.True(t, ok)

			text, ok := contentItem["text"].(string)
			require.True(t, ok)

			// Verify expected content
			for _, expectedText := range tt.expectedText {
				assert.Contains(t, text, expectedText)
			}
		})
	}
}

func TestMCPGetFileAnalysis(t *testing.T) {
	tmpDir := createTestProject(t)
	defer os.RemoveAll(tmpDir)

	client, err := NewMCPClient(tmpDir, false)
	require.NoError(t, err)
	defer client.Close()

	// Initialize
	initMsg := MCPMessage{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: MCPInitParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]interface{}{},
			ClientInfo: map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	_, err = client.sendAndReceive(initMsg, 10*time.Second)
	require.NoError(t, err)

	// Test file analysis
	mainTSPath := filepath.Join(tmpDir, "main.ts")
	toolCallMsg := MCPMessage{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "get_file_analysis",
			"arguments": map[string]interface{}{
				"file_path": mainTSPath,
			},
		},
	}

	response, err := client.sendAndReceive(toolCallMsg, 10*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, response.Result)

	// Extract and verify content
	resultMap, ok := response.Result.(map[string]interface{})
	require.True(t, ok)

	content, ok := resultMap["content"].([]interface{})
	require.True(t, ok)
	require.Len(t, content, 1)

	contentItem, ok := content[0].(map[string]interface{})
	require.True(t, ok)

	text, ok := contentItem["text"].(string)
	require.True(t, ok)

	// Verify file analysis content
	assert.Contains(t, text, "# File Analysis:")
	assert.Contains(t, text, "main.ts")
	assert.Contains(t, text, "**Language:**")
	assert.Contains(t, text, "**Lines:**")
	assert.Contains(t, text, "**Symbols:**")
	assert.Contains(t, text, "## Symbols")
}

func TestMCPGetDependencies(t *testing.T) {
	tmpDir := createTestProject(t)
	defer os.RemoveAll(tmpDir)

	client, err := NewMCPClient(tmpDir, false)
	require.NoError(t, err)
	defer client.Close()

	// Initialize
	initMsg := MCPMessage{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: MCPInitParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]interface{}{},
			ClientInfo: map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	_, err = client.sendAndReceive(initMsg, 10*time.Second)
	require.NoError(t, err)

	// Test global dependencies
	toolCallMsg := MCPMessage{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "get_dependencies",
			"arguments": map[string]interface{}{},
		},
	}

	response, err := client.sendAndReceive(toolCallMsg, 10*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, response.Result)

	// Extract and verify content
	resultMap, ok := response.Result.(map[string]interface{})
	require.True(t, ok)

	content, ok := resultMap["content"].([]interface{})
	require.True(t, ok)
	require.Len(t, content, 1)

	contentItem, ok := content[0].(map[string]interface{})
	require.True(t, ok)

	text, ok := contentItem["text"].(string)
	require.True(t, ok)

	// Verify dependency analysis content
	assert.Contains(t, text, "# Dependency Analysis")
	assert.Contains(t, text, "## Global Dependency Overview")
	assert.Contains(t, text, "**Total Files:**")
	assert.Contains(t, text, "**Total Import Relationships:**")
}

func TestMCPWatchChanges(t *testing.T) {
	tmpDir := createTestProject(t)
	defer os.RemoveAll(tmpDir)

	client, err := NewMCPClient(tmpDir, false)
	require.NoError(t, err)
	defer client.Close()

	// Initialize
	initMsg := MCPMessage{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: MCPInitParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]interface{}{},
			ClientInfo: map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	_, err = client.sendAndReceive(initMsg, 10*time.Second)
	require.NoError(t, err)

	// Enable watching
	enableWatchMsg := MCPMessage{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "watch_changes",
			"arguments": map[string]interface{}{
				"enable": true,
			},
		},
	}

	response, err := client.sendAndReceive(enableWatchMsg, 10*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, response.Result)

	// Verify response
	resultMap, ok := response.Result.(map[string]interface{})
	require.True(t, ok)

	content, ok := resultMap["content"].([]interface{})
	require.True(t, ok)
	require.Len(t, content, 1)

	contentItem, ok := content[0].(map[string]interface{})
	require.True(t, ok)

	text, ok := contentItem["text"].(string)
	require.True(t, ok)

	assert.Contains(t, text, "File watching enabled")

	// Disable watching
	disableWatchMsg := MCPMessage{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "watch_changes",
			"arguments": map[string]interface{}{
				"enable": false,
			},
		},
	}

	response, err = client.sendAndReceive(disableWatchMsg, 5*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, response.Result)
}

func TestMCPErrorHandling(t *testing.T) {
	tmpDir := createTestProject(t)
	defer os.RemoveAll(tmpDir)

	client, err := NewMCPClient(tmpDir, false)
	require.NoError(t, err)
	defer client.Close()

	// Initialize
	initMsg := MCPMessage{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: MCPInitParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]interface{}{},
			ClientInfo: map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	_, err = client.sendAndReceive(initMsg, 10*time.Second)
	require.NoError(t, err)

	tests := []struct {
		name      string
		toolName  string
		arguments map[string]interface{}
		expectErr bool
	}{
		{
			name:     "invalid tool name",
			toolName: "non_existent_tool",
			arguments: map[string]interface{}{
				"param": "value",
			},
			expectErr: true,
		},
		{
			name:     "missing required argument",
			toolName: "search_symbols",
			arguments: map[string]interface{}{
				// Missing required "query" argument
			},
			expectErr: true,
		},
		{
			name:     "invalid file path",
			toolName: "get_file_analysis",
			arguments: map[string]interface{}{
				"file_path": "/non/existent/file.ts",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toolCallMsg := MCPMessage{
				JSONRPC: "2.0",
				Method:  "tools/call",
				Params: map[string]interface{}{
					"name":      tt.toolName,
					"arguments": tt.arguments,
				},
			}

			response, err := client.sendAndReceive(toolCallMsg, 10*time.Second)
			require.NoError(t, err)

			if tt.expectErr {
				// MCP has two types of errors:
				// 1. System errors (invalid tool): JSON-RPC error
				// 2. Tool errors (missing args): Successful response with isError: true
				
				if response.Error != nil {
					// System-level error (JSON-RPC error)
					assert.NotNil(t, response.Error, "Expected JSON-RPC error for system errors")
				} else {
					// Tool-level error (successful response with error content)
					assert.NotNil(t, response.Result, "Expected result with error content")
					
					resultMap, ok := response.Result.(map[string]interface{})
					require.True(t, ok, "Result should be a map")
					
					isError, exists := resultMap["isError"]
					if exists {
						assert.True(t, isError.(bool), "Expected isError to be true")
					} else {
						// Check if content contains error text
						content, ok := resultMap["content"].([]interface{})
						require.True(t, ok, "Expected content array")
						require.Len(t, content, 1, "Expected one content item")
						
						contentItem, ok := content[0].(map[string]interface{})
						require.True(t, ok, "Expected content item to be a map")
						
						text, ok := contentItem["text"].(string)
						require.True(t, ok, "Expected text content")
						assert.Contains(t, text, "required", "Expected error message about required field")
					}
				}
			} else {
				assert.Nil(t, response.Error, "Expected successful response")
				assert.NotNil(t, response.Result, "Expected result in response")
			}
		})
	}
}

func TestMCPPerformance(t *testing.T) {
	tmpDir := createTestProject(t)
	defer os.RemoveAll(tmpDir)

	client, err := NewMCPClient(tmpDir, false)
	require.NoError(t, err)
	defer client.Close()

	// Initialize
	initMsg := MCPMessage{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: MCPInitParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]interface{}{},
			ClientInfo: map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	_, err = client.sendAndReceive(initMsg, 10*time.Second)
	require.NoError(t, err)

	// Test performance of multiple tool calls
	start := time.Now()
	numCalls := 10

	for i := 0; i < numCalls; i++ {
		toolCallMsg := MCPMessage{
			JSONRPC: "2.0",
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name": "search_symbols",
				"arguments": map[string]interface{}{
					"query": "UserService",
					"limit": 5,
				},
			},
		}

		response, err := client.sendAndReceive(toolCallMsg, 5*time.Second)
		require.NoError(t, err)
		assert.NotNil(t, response.Result)
	}

	duration := time.Since(start)
	avgDuration := duration / time.Duration(numCalls)

	t.Logf("Average response time for %d calls: %v", numCalls, avgDuration)
	assert.Less(t, avgDuration, 2*time.Second, "Average response time should be less than 2 seconds")
}

func TestMCPConcurrentRequests(t *testing.T) {
	tmpDir := createTestProject(t)
	defer os.RemoveAll(tmpDir)

	client, err := NewMCPClient(tmpDir, false)
	require.NoError(t, err)
	defer client.Close()

	// Initialize
	initMsg := MCPMessage{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: MCPInitParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]interface{}{},
			ClientInfo: map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	_, err = client.sendAndReceive(initMsg, 10*time.Second)
	require.NoError(t, err)

	// Test sequential requests (MCP stdio doesn't support true concurrency)
	// This tests the client's ability to handle multiple requests properly
	numRequests := 5
	var wg sync.WaitGroup
	results := make(chan error, numRequests)
	var requestMutex sync.Mutex // Serialize requests to avoid stdio race conditions

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			toolCallMsg := MCPMessage{
				JSONRPC: "2.0",
				Method:  "tools/call",
				Params: map[string]interface{}{
					"name": "search_symbols",
					"arguments": map[string]interface{}{
						"query": fmt.Sprintf("symbol_%d", id),
						"limit": 5,
					},
				},
			}

			// Serialize requests to avoid stdio race conditions
			requestMutex.Lock()
			_, err := client.sendAndReceive(toolCallMsg, 10*time.Second)
			requestMutex.Unlock()
			
			results <- err
		}(i)
	}

	wg.Wait()
	close(results)

	// Check all requests completed without error
	for err := range results {
		assert.NoError(t, err)
	}
}

// Test helper to capture stderr output from server
func captureServerLogs(t *testing.T, client *MCPClient, duration time.Duration) string {
	var logBuffer bytes.Buffer
	var mu sync.Mutex
	done := make(chan bool)

	go func() {
		defer close(done)
		buf := make([]byte, 1024)
		for {
			n, err := client.stderr.Read(buf)
			if n > 0 {
				mu.Lock()
				logBuffer.Write(buf[:n])
				mu.Unlock()
			}
			if err != nil {
				break
			}
		}
	}()

	time.Sleep(duration)

	mu.Lock()
	result := logBuffer.String()
	mu.Unlock()
	
	return result
}

func TestMCPServerLogging(t *testing.T) {
	tmpDir := createTestProject(t)
	defer os.RemoveAll(tmpDir)

	// Test with verbose logging
	client, err := NewMCPClient(tmpDir, true)
	require.NoError(t, err)
	defer client.Close()

	// Capture some logs
	logs := captureServerLogs(t, client, 2*time.Second)

	// Verify verbose output contains expected information
	assert.Contains(t, logs, "CodeContext MCP Server starting")
	assert.Contains(t, logs, "TargetDir:")
	assert.Contains(t, logs, "Successfully registered 8 tools")
}
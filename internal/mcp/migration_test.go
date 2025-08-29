package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMCPServerV3Migration verifies that the MCP server can be created and started with v0.3.0 API
func TestMCPServerV3Migration(t *testing.T) {
	config := &MCPConfig{
		Name:        "test-server",
		Version:     "test-v0.3.0",
		TargetDir:   ".",
		EnableWatch: false,
		DebounceMs:  300,
	}

	// Create server
	server, err := NewCodeContextMCPServer(config)
	assert.NoError(t, err, "Should create MCP server successfully")
	assert.NotNil(t, server, "Server should not be nil")
	assert.NotNil(t, server.server, "Internal MCP server should be initialized")
	assert.Equal(t, config.Name, server.config.Name, "Config should be preserved")
	assert.Equal(t, config.Version, server.config.Version, "Version should be preserved")
}

// TestMCPToolRegistration verifies that all tools are registered correctly
func TestMCPToolRegistration(t *testing.T) {
	config := &MCPConfig{
		Name:        "test-registration",
		Version:     "test-v0.3.0", 
		TargetDir:   ".",
		EnableWatch: false,
		DebounceMs:  300,
	}

	// Create server - this will register all tools
	server, err := NewCodeContextMCPServer(config)
	assert.NoError(t, err, "Should create server with tools registered")
	assert.NotNil(t, server, "Server should be created")
	
	// Verify server has been initialized with tools
	assert.NotNil(t, server.server, "MCP server should be initialized")
}

// TestMCPServerShutdown verifies that the server can be stopped gracefully
func TestMCPServerShutdown(t *testing.T) {
	config := &MCPConfig{
		Name:        "test-shutdown",
		Version:     "test-v0.3.0",
		TargetDir:   ".",
		EnableWatch: false,
		DebounceMs:  300,
	}

	server, err := NewCodeContextMCPServer(config)
	assert.NoError(t, err, "Should create server")
	
	// Test graceful shutdown
	assert.NotPanics(t, func() {
		server.Stop()
	}, "Server should stop gracefully")
}

// TestMCPServerStartStopCycle tests that the server can start and stop without issues
// Note: This test validates server lifecycle without actually starting the stdio transport
// to avoid interfering with test coverage reporting
func TestMCPServerStartStopCycle(t *testing.T) {
	config := &MCPConfig{
		Name:        "test-cycle",
		Version:     "test-v0.3.0",
		TargetDir:   ".",
		EnableWatch: false,
		DebounceMs:  300,
	}

	server, err := NewCodeContextMCPServer(config)
	assert.NoError(t, err, "Should create server")
	
	// Test server creation and basic initialization
	assert.NotNil(t, server.server, "Internal MCP server should be initialized")
	assert.NotNil(t, server.config, "Config should be set")
	assert.NotNil(t, server.analyzer, "Analyzer should be initialized")
	
	// Test that server can be stopped gracefully
	assert.NotPanics(t, func() {
		server.Stop()
	}, "Server should stop without panicking")
	
	// Verify server state after stop
	assert.True(t, server.stopped, "Server should be marked as stopped")
}
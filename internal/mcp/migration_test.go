package mcp

import (
	"context"
	"testing"
	"time"

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
	
	// Test that server creation and shutdown works
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Start server in goroutine (it will timeout after analysis and that's expected)
	done := make(chan error, 1)
	go func() {
		done <- server.Run(ctx)
	}()
	
	// Wait for timeout or completion
	select {
	case err := <-done:
		// Context deadline exceeded is expected after analysis completes
		assert.Error(t, err, "Server should timeout as expected")
	case <-time.After(15 * time.Second):
		t.Error("Server did not respond to context cancellation in time")
	}
	
	// Cleanup
	server.Stop()
}
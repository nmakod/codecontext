# Model Context Protocol (MCP) Server

**Version:** 2.1 - Official SDK Implementation  
**Status:** ✅ PRODUCTION READY with Official MCP Go SDK  
**Updated:** July 2025

CodeContext includes a production-ready Model Context Protocol (MCP) server built with the official MCP Go SDK that provides real-time codebase context to AI assistants and development tools.

## Overview

The MCP server exposes CodeContext's powerful code analysis capabilities through a standardized protocol, enabling AI assistants to understand your codebase structure, search symbols, analyze dependencies, and monitor changes in real-time.

## Quick Start

### Starting the MCP Server

```bash
# Start MCP server for current directory
codecontext mcp

# With custom settings
codecontext mcp --target ./src --watch --debounce 300 --verbose
```

### Available Tools

The MCP server provides eight powerful tools with **dynamic project targeting**:

1. **`get_codebase_overview`** - Complete repository analysis
2. **`get_file_analysis`** - Detailed file breakdown with symbols  
3. **`get_symbol_info`** - Symbol definitions and usage
4. **`search_symbols`** - Search symbols across codebase
5. **`get_dependencies`** - Import/dependency analysis
6. **`watch_changes`** - Real-time change notifications
7. **`get_semantic_neighborhoods`** - Git-pattern based file relationships
8. **`get_framework_analysis`** - Framework-specific analysis

### 🚀 **Multi-Project Support**

All tools support an optional `target_dir` parameter, allowing you to analyze different projects without changing configurations:

```bash
# Simple configuration - works with any project
{
  "codecontext": {
    "command": "codecontext", 
    "args": ["mcp"]
  }
}
```

**Usage Examples:**
- "Analyze ~/code/my-react-app"
- "Compare ~/code/backend with ~/code/frontend" 
- "Check dependencies in /Users/dev/projects/api"

## Configuration

### Command-Line Options

```bash
codecontext mcp [flags]

Flags:
  -t, --target string     target directory to analyze (default ".")
  -w, --watch            enable real-time file watching (default true)
  -d, --debounce int     debounce interval for file changes (ms) (default 500)
  -n, --name string      MCP server name (default "codecontext")
  -v, --verbose          verbose output

Global Flags:
      --config string   config file (default is .codecontext/config.yaml)
  -o, --output string   output file (default "CLAUDE.md")
```

### Configuration File

Add MCP settings to `.codecontext/config.yaml`:

```yaml
mcp:
  name: "my-codebase-server"
  target: "./src"
  watch: true
  debounce: 500
  extensions:
    - ".ts"
    - ".tsx" 
    - ".js"
    - ".jsx"
    - ".go"
    - ".py"
```

## Tool Usage Examples

### 1. Get Codebase Overview

```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "get_codebase_overview",
    "arguments": {
      "include_stats": true
    }
  },
  "id": 1
}
```

**Response includes:**
- File count and language breakdown
- Symbol extraction summary
- Import relationship analysis
- Detailed statistics (if requested)

### 2. Search Symbols

```json
{
  "jsonrpc": "2.0",
  "method": "tools/call", 
  "params": {
    "name": "search_symbols",
    "arguments": {
      "query": "UserService",
      "file_type": "typescript",
      "limit": 10
    }
  },
  "id": 2
}
```

**Response includes:**
- Matching symbol names
- File locations with line numbers
- Symbol types (class, function, etc.)

### 3. Analyze File Dependencies

```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "get_dependencies", 
    "arguments": {
      "file_path": "/path/to/file.ts",
      "direction": "imports"
    }
  },
  "id": 3
}
```

**Response includes:**
- Import relationships
- Dependent files
- Dependency graph analysis

### 4. Enable Real-time Watching

```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "watch_changes",
    "arguments": {
      "enable": true
    }
  },
  "id": 4
}
```

**Features:**
- Debounced file change detection
- Automatic analysis refresh
- Memory-efficient incremental updates

## AI Assistant Integration

### Claude Desktop

Add to your Claude Desktop MCP configuration:

```json
{
  "mcpServers": {
    "codecontext": {
      "command": "codecontext",
      "args": ["mcp", "--target", "/path/to/your/project"]
    }
  }
}
```

### VSCode Extensions

Use MCP client extensions that support the standard protocol:

```json
{
  "mcp.servers": [
    {
      "name": "codecontext",
      "command": "codecontext mcp --target ${workspaceFolder}"
    }
  ]
}
```

### Custom Applications

Create MCP clients using standard JSON-RPC 2.0 over stdio:

```typescript
import { spawn } from 'child_process';

const mcpServer = spawn('codecontext', ['mcp', '--target', './src']);

// Send initialization
mcpServer.stdin.write(JSON.stringify({
  jsonrpc: "2.0",
  method: "initialize",
  params: {
    protocolVersion: "2024-11-05",
    clientInfo: { name: "my-client", version: "1.0.0" }
  },
  id: 1
}) + '\n');
```

## Real-time Features

### File Watching

When enabled (`--watch`), the MCP server:
- Monitors file system changes in real-time
- Debounces rapid changes (configurable interval)
- Automatically refreshes code analysis
- Provides immediate context updates

### Performance Optimizations

- **Incremental Analysis**: Only re-analyzes changed files
- **Memory Management**: Built-in garbage collection monitoring
- **Efficient Parsing**: Tree-sitter AST parsing with caching
- **Concurrent Processing**: Parallel file processing support

## Protocol Details

### Official MCP SDK Implementation

**SDK Version:** github.com/modelcontextprotocol/go-sdk v0.3.0  
**Protocol Version:** 2024-11-05  
**Compatibility:** Full Claude Desktop integration

### Transport

- **Primary**: Standard I/O (stdin/stdout) via official `mcp.NewStdioTransport()`
- **Format**: JSON-RPC 2.0 messages
- **Encoding**: UTF-8 with proper stderr/stdout separation
- **Logging**: All debug output redirected to stderr for clean protocol communication

### Enhanced Capabilities

The server advertises enhanced capabilities using official SDK patterns:

```json
{
  "capabilities": {
    "tools": {
      "listChanged": true
    },
    "logging": {},
    "experimental": {}
  },
  "serverInfo": {
    "name": "codecontext",
    "version": "2.1.0"
  }
}
```

### SDK Integration Benefits

- **Protocol Compliance**: 100% compatibility with Claude Desktop
- **Type Safety**: Generic type-safe tool parameter handling
- **Error Handling**: Structured error responses with proper codes
- **Performance**: Optimized transport layer with official implementation

### Error Handling

Standard JSON-RPC error responses:

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": "file_path is required"
  },
  "id": 1
}
```

## Advanced Usage

### Custom File Extensions

Configure which files to analyze:

```bash
codecontext mcp --target . --config custom-config.yaml
```

```yaml
# custom-config.yaml
mcp:
  extensions:
    - ".rs"    # Rust
    - ".cpp"   # C++
    - ".java"  # Java
    - ".py"    # Python
```

### Integration with CI/CD

Use in automated workflows:

```bash
# Generate analysis and exit
codecontext mcp --target ./src --watch=false < /dev/null
```

### Development Debugging

Enable verbose logging:

```bash
codecontext mcp --verbose --target ./src 2> mcp-server.log
```

## Security Considerations

### Access Control

- Server only analyzes files in specified target directory
- No file modification capabilities
- Read-only access to codebase
- No network connections

### Sandboxing

- Runs in user context only
- No elevated privileges required
- Standard file system permissions apply
- Memory usage monitoring and limits

## Performance Benchmarks

Typical performance characteristics:

| Operation | Small Project (<100 files) | Large Project (1000+ files) |
|-----------|---------------------------|------------------------------|
| Initial Analysis | < 1 second | < 10 seconds |
| Symbol Search | < 50ms | < 200ms |
| File Analysis | < 100ms | < 500ms |
| Dependency Analysis | < 200ms | < 1 second |

### Optimization Tips

1. **Target Specific Directories**: Use `--target ./src` instead of entire repo
2. **Exclude Build Artifacts**: Configure `.gitignore` patterns
3. **Adjust Debounce**: Increase `--debounce` for busy file systems
4. **Disable Watch**: Use `--watch=false` for one-time analysis

## Troubleshooting

### Common Issues

1. **Server not responding**
   ```bash
   # Check if process is running
   ps aux | grep codecontext
   
   # Verify target directory exists
   ls -la /path/to/target
   ```

2. **High memory usage**
   ```bash
   # Monitor memory
   codecontext mcp --verbose --target ./src
   
   # Reduce scope
   codecontext mcp --target ./src/specific-module
   ```

3. **Tree-sitter errors**
   ```bash
   # Verify installation
   codecontext generate --target ./test-files
   ```

### Debug Mode

Enable detailed logging:

```bash
export CODECONTEXT_DEBUG=1
codecontext mcp --verbose --target ./src
```

### Log Analysis

Server logs include:
- Initialization status
- Tool registration
- Analysis performance metrics
- Error details and stack traces

## API Reference

### Tool Schemas

Each tool has a JSON schema defining its parameters:

#### get_codebase_overview
```json
{
  "type": "object",
  "properties": {
    "include_stats": {
      "type": "boolean",
      "description": "Include detailed statistics"
    }
  }
}
```

#### search_symbols
```json
{
  "type": "object",
  "properties": {
    "query": {
      "type": "string", 
      "description": "Search query for symbols",
      "required": true
    },
    "file_type": {
      "type": "string",
      "description": "Filter by file type"
    },
    "limit": {
      "type": "integer",
      "description": "Maximum results (default: 20)"
    }
  }
}
```

### Response Formats

All tools return structured content:

```json
{
  "jsonrpc": "2.0",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "# Analysis Results\n\n..."
      }
    ]
  },
  "id": 1
}
```

## Contributing

### Adding New Tools

1. Define tool function in `internal/mcp/server.go`
2. Register in `registerTools()` method
3. Add comprehensive tests
4. Update documentation

### Extending Language Support

1. Add Tree-sitter grammar dependency
2. Extend parser in `internal/parser/`
3. Add language-specific test cases
4. Update configuration options

For more details, see the [development documentation](../CONTRIBUTING.md).
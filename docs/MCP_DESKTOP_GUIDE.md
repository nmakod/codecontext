# CodeContext MCP Desktop Integration Guide

## Quick Start for Claude Desktop Users

### 1. Installation (30 seconds)

#### macOS
```bash
# Using Homebrew (recommended)
brew tap nmakod/tap
brew install codecontext

# Or download directly
curl -L https://github.com/nmakod/codecontext/releases/latest/download/codecontext-darwin-$(uname -m) -o codecontext
chmod +x codecontext
sudo mv codecontext /usr/local/bin/
```

#### Windows
```powershell
# Download the latest release
Invoke-WebRequest -Uri "https://github.com/nmakod/codecontext/releases/latest/download/codecontext.exe" -OutFile "codecontext.exe"

# Add to PATH or move to a directory in PATH
```

#### Linux
```bash
# Download for your architecture
curl -L https://github.com/nmakod/codecontext/releases/latest/download/codecontext-linux-amd64 -o codecontext
chmod +x codecontext
sudo mv codecontext /usr/local/bin/
```

### 2. Start MCP Server (10 seconds)

```bash
# Navigate to your project
cd /path/to/your/project

# Start the MCP server
codecontext mcp

# The server is now ready for Claude Desktop!
```

### 3. Connect with Claude Desktop

1. Open Claude Desktop settings
2. Add new MCP server
3. Enter the connection details provided by codecontext
4. Start chatting with full codebase context!

## Powerful Features for Desktop Users

### 🚀 Instant Codebase Overview
Ask Claude: "Give me an overview of this codebase"
- File statistics
- Language breakdown  
- Complexity metrics
- Architecture insights

### 🔍 Intelligent Symbol Search
Ask Claude: "Find all authentication-related functions"
- Semantic search beyond simple text matching
- Understands code relationships
- Finds related symbols across files

### 🧩 Semantic Neighborhoods
Ask Claude: "What files are related to user management?"
- Uses Git history to find truly related files
- Identifies hidden dependencies
- Perfect for understanding unfamiliar codebases

### 📊 Dependency Analysis
Ask Claude: "Show me the dependency graph for the API module"
- Visualizes import relationships
- Identifies circular dependencies
- Suggests refactoring opportunities

### 👀 Real-Time Monitoring
Ask Claude: "Watch for changes and update the context"
- Monitors file system changes
- Updates context automatically
- Perfect for active development

### 🏗️ Framework Intelligence
Ask Claude: "Analyze the React components structure"
- Understands framework patterns
- Provides framework-specific insights
- Helps maintain consistency

## Desktop-Optimized Performance

### Resource Usage
- **Memory**: ~20MB (less than a browser tab)
- **CPU**: Minimal when idle, efficient during analysis
- **Disk**: No cache bloat, minimal footprint

### Speed Benchmarks
| Operation | CodeContext | Typical Node.js MCP |
|-----------|-------------|---------------------|
| Startup | <100ms | 2-3 seconds |
| 10k files analysis | 1 second | 30+ seconds |
| Memory usage | 20MB | 150-300MB |
| Binary size | 15MB | 500MB+ with dependencies |

## Common Use Cases

### 1. New Project Onboarding
```
You: "I just cloned this repo. Help me understand the architecture"
Claude: [Uses get_codebase_overview and get_framework_analysis to provide comprehensive introduction]
```

### 2. Bug Investigation
```
You: "There's a bug in user authentication. What files should I check?"
Claude: [Uses get_semantic_neighborhoods to find all auth-related files based on Git patterns]
```

### 3. Code Generation
```
You: "Create a new API endpoint following this project's patterns"
Claude: [Uses get_framework_analysis and search_symbols to understand patterns and generate consistent code]
```

### 4. Refactoring
```
You: "I want to refactor the payment module. What's connected to it?"
Claude: [Uses get_dependencies to map all connections and suggest safe refactoring approach]
```

## 🎯 Multi-Project Workflows

### Zero-Configuration Multi-Project Support

CodeContext now supports **dynamic project targeting** - analyze any project without editing configurations!

#### Simple Setup (One-Time)
```json
{
  "mcpServers": {
    "codecontext": {
      "command": "codecontext",
      "args": ["mcp"]
    }
  }
}
```

#### Multi-Project Usage Examples

**1. Project Comparison**
```
You: "Compare the authentication systems in ~/code/web-app and ~/code/mobile-app"

Claude: 
1. Analyzing ~/code/web-app authentication...
2. Analyzing ~/code/mobile-app authentication...
3. Here's the comparison...
```

**2. Cross-Project Learning**
```  
You: "How does ~/code/senior-dev-project handle error boundaries? Apply similar patterns to ~/code/my-project"

Claude: [Analyzes both projects and suggests improvements]
```

**3. Multi-Repo Development**
```
You: "Check if the API changes in ~/code/backend break anything in ~/code/frontend"

Claude: [Analyzes dependencies and interfaces across both projects]
```

**4. Architecture Analysis**
```
You: "What patterns can I learn from ~/code/open-source-project for my ~/code/startup-project?"

Claude: [Compares architectures and suggests best practices]
```

#### Supported Path Formats

- **Absolute paths**: `/Users/john/code/my-project`
- **Home-relative**: `~/code/my-project` 
- **Environment variables**: `$HOME/code/my-project`
- **Default fallback**: Uses configured project when no path specified

#### Benefits

✅ **No JSON Editing** - Switch projects in conversation  
✅ **Compare Projects** - Analyze multiple codebases simultaneously  
✅ **Learn Patterns** - Study different architectural approaches  
✅ **Context Switching** - Move between projects seamlessly  
✅ **Backward Compatible** - Existing configurations work unchanged  

## Advanced Configuration

### Custom Settings
```bash
# Start with custom settings
codecontext mcp \
  --watch \                    # Enable file watching
  --debounce 200 \            # Debounce file changes (ms)
  --name "MyProject" \        # Custom server name
  --verbose                   # Detailed logging
```

### Performance Tuning
```bash
# For large codebases
codecontext mcp \
  --cache-size 1000 \         # Increase cache size
  --workers 8 \               # Parallel processing
  --exclude "node_modules,dist,build"  # Skip unnecessary directories
```

## Troubleshooting

### Server Won't Start
```bash
# Check if port is in use
lsof -i :50051

# Try a different port
codecontext mcp --port 50052
```

### Claude Can't Connect
```bash
# Verify server is running
codecontext mcp --verbose

# Check firewall settings
# Ensure localhost connections are allowed
```

### Performance Issues
```bash
# Exclude large directories
codecontext mcp --exclude "node_modules,vendor,dist"

# Limit file size processing
codecontext mcp --max-file-size 1MB
```

## Why CodeContext for Desktop?

### 🎯 **Built for Desktop First**
- Single binary - no installation complexity
- Minimal resource usage
- Native performance
- Works offline

### 🛡️ **Privacy & Security**
- Runs entirely locally
- No data leaves your machine
- Open source and auditable
- No telemetry

### ⚡ **Unmatched Performance**
- Analyzes thousands of files per second
- Instant startup
- Real-time updates
- Efficient memory usage

### 🔧 **Zero Maintenance**
- No npm updates
- No dependency conflicts
- No breaking changes
- Just works

## Getting Help

- **Documentation**: https://github.com/nmakod/codecontext/tree/main/docs
- **Issues**: https://github.com/nmakod/codecontext/issues
- **Discussions**: https://github.com/nmakod/codecontext/discussions

---

*CodeContext - Empowering Claude Desktop with deep codebase understanding*
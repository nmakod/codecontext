# 🤖 CodeContext - AI-Powered Development Context Maps

**Intelligent context maps for seamless AI development workflows with Claude**

[![Release](https://img.shields.io/github/v/release/nmakod/codecontext)](https://github.com/nmakod/codecontext/releases)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.19+-blue.svg)](https://golang.org)

## 🎯 What is CodeContext?

CodeContext automatically generates **intelligent, token-optimized context maps** of your codebase specifically designed for AI development workflows. Instead of manually copying files or explaining your project structure to Claude, CodeContext creates comprehensive context that enables AI to understand your entire codebase instantly.

### ⚡ Quick Example

```bash
# Generate context for your project
codecontext generate

# Copy the generated CLAUDE.md and paste into Claude
# Claude now understands your entire codebase structure!
```

**Result**: Claude can now help with architecture decisions, debug complex issues, suggest refactoring, and implement features with full understanding of your project.

## 🚀 Key Features

### 🔍 **Real Tree-sitter Analysis**
- **JavaScript/TypeScript**: Full AST parsing with symbol extraction
- **Go Language**: Complete language support  
- **Multi-language**: JSON, YAML, and extensible architecture
- **Symbol Recognition**: Functions, classes, interfaces, imports, variables

### 🧠 **AI-Optimized Context**
- **Token Efficient**: Optimized output format for AI consumption
- **Relationship Mapping**: File dependencies and import relationships
- **Smart Filtering**: Focus on relevant code, exclude noise
- **Incremental Updates**: Only regenerate what's changed

### ⚡ **Enhanced Diff Algorithms (v2.0)**
- **Semantic vs Structural Diffs**: Understand code changes beyond text
- **Symbol Rename Detection**: 6 similarity algorithms + 5 heuristic patterns
- **Import Dependency Tracking**: Comprehensive change impact analysis
- **Confidence Scoring**: Evidence-based change classification

### 🛠️ **Developer Experience**
- **Watch Mode**: Real-time context updates during development
- **Compaction**: Reduce context size for large projects
- **CLI Tools**: Professional command-line interface
- **Cross-Platform**: macOS, Linux, Windows support

## 📦 Installation

### Homebrew (macOS - Recommended)
```bash
brew install --HEAD --build-from-source https://raw.githubusercontent.com/nmakod/codecontext/main/Formula/codecontext.rb
```

### Download Binary
```bash
# macOS (Apple Silicon)
curl -L https://github.com/nmakod/codecontext/releases/download/v2.0.0/codecontext-2.0.0-darwin-arm64.tar.gz | tar xz
sudo mv codecontext /usr/local/bin/

# Other platforms available at: https://github.com/nmakod/codecontext/releases
```

### Build from Source
```bash
git clone https://github.com/nmakod/codecontext.git
cd codecontext
make build
sudo make install
```

## 🚀 Quick Start with Claude

### 1. Initialize Your Project
```bash
cd your-project
codecontext init
```

### 2. Generate Context Map
```bash
codecontext generate
```

### 3. Use with Claude
Copy the generated `CLAUDE.md` content and start your Claude conversation:

```
I'm working on a [project description]. Here's my codebase context:

[Paste CLAUDE.md content]

I need help with [specific task].
```

### 4. Iterative Development
```bash
# Make changes to your code
# Update context
codecontext update

# Share updated context with Claude for continued assistance
```

## 📊 Example Output

```markdown
# CodeContext Map

**Generated:** 2025-07-12T17:45:38+05:30  
**Analysis Time:** 35ms  
**Status:** Real Tree-sitter Analysis

## 📊 Overview
- **Files Analyzed**: 15 files
- **Symbols Extracted**: 142 symbols  
- **Languages Detected**: 3 (TypeScript, JavaScript, JSON)
- **Import Relationships**: 28 dependencies

## 📁 File Analysis
| File | Language | Symbols | Type |
|------|----------|---------|------|
| `src/components/UserCard.tsx` | typescript | 8 | component |
| `src/services/userService.ts` | typescript | 12 | service |
| `src/utils/validation.ts` | typescript | 6 | utility |

## 🔍 Symbol Analysis  
| Symbol | Type | File | Line | Signature |
|--------|------|------|------|----------|
| `UserCard` | class | `src/components/UserCard.tsx` | 12 | `class UserCard` |
| `validateEmail` | function | `src/utils/validation.ts` | 25 | `validateEmail(email: string)` |

## 🔗 Import Relationships
- `src/components/UserCard.tsx` → [`services/userService`, `utils/validation`]
- `src/services/userService.ts` → [`utils/api`, `types/user`]
```

## 🛠️ Advanced Usage

### Watch Mode for Active Development
```bash
# Auto-update context as you code
codecontext watch

# Claude conversations stay in sync with your changes!
```

### Compaction for Large Projects
```bash
# Reduce context size while preserving key information
codecontext compact --level balanced

# Perfect for large codebases that exceed token limits
```

### Focused Analysis
```bash
# Generate context for specific directories
codecontext generate src/components/ src/services/

# Include/exclude patterns in config
codecontext generate --exclude "**/*.test.*"
```

### Configuration
```yaml
# .codecontext/config.yaml
project:
  name: "my-awesome-app"
  
analysis:
  include_patterns:
    - "src/**"
    - "components/**"
  exclude_patterns:
    - "**/*.test.*"
    - "node_modules/**"
    - "dist/**"

output:
  format: "markdown"
  include_stats: true
  max_file_size: 1048576  # 1MB
```

## 🎯 Use Cases with Claude

### 🏗️ **Architecture Planning**
```
Based on this codebase structure: [context]
What's the best way to implement user authentication?
```

### 🐛 **Debugging Complex Issues**
```
I'm getting this error: [error details]
Here's my codebase context: [context]
Can you help identify the root cause?
```

### 🔄 **Refactoring Guidance**
```
I want to refactor the UserService class: [context]
How can I improve this while maintaining compatibility?
```

### ✨ **Feature Implementation**
```
I need to add real-time notifications: [context]
What's the best approach given my current architecture?
```

### 📋 **Code Reviews**
```
Here's my updated codebase after implementing the new feature: [context]
Can you review for best practices and potential issues?
```

## 📚 Documentation

- **[🤖 Complete Claude Integration Guide](docs/CLAUDE_INTEGRATION.md)** - Comprehensive workflow guide
- **[🚀 Real-World Example](examples/CLAUDE_WORKFLOW.md)** - Step-by-step authentication system example  
- **[⚡ Quick Reference](CLAUDE_QUICKSTART.md)** - Essential commands and templates
- **[🏗️ Architecture](docs/ARCHITECTURE.md)** - Technical implementation details

## 🎯 Roadmap

### ✅ **Phase 1: Foundation (Completed)**
- CLI framework and configuration
- Basic file analysis and output generation
- Tree-sitter integration

### ✅ **Phase 2.1: Enhanced Diff Algorithms (v2.0.0)**
- Semantic vs structural diff analysis
- Symbol rename detection with confidence scoring
- Import dependency change tracking
- Language-specific AST diffing

### 🔄 **Phase 2.2: Multi-Level Caching (Coming Soon)**
- LRU cache for parsed ASTs
- Diff result caching with TTL
- Persistent cache across CLI invocations
- Cache invalidation strategies

### 🔄 **Phase 2.3: Watch Mode Optimization (Coming Soon)**
- Debounced file changes (300ms default)
- Batch processing of multiple changes
- Priority queuing for critical files
- Resource throttling for large repositories

### 🔮 **Phase 3: Advanced Features (Future)**
- IDE integrations (VS Code, IntelliJ)
- Git integration for change tracking
- Team collaboration features
- Custom output formats

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup
```bash
git clone https://github.com/nmakod/codecontext.git
cd codecontext
go mod download
make build
```

### Running Tests
```bash
make test
```

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🌟 Why CodeContext?

### **Before CodeContext**
```
You: "I have a React app with TypeScript, Express backend, and I'm trying to implement user authentication..."
Claude: "I'd be happy to help! Can you show me your current file structure and the relevant code?"
You: [Copies multiple files manually, explains project structure]
Claude: [Provides help based on limited context]
```

### **With CodeContext**
```bash
codecontext generate
```
```
You: "I need to implement user authentication. Here's my project context: [paste CLAUDE.md]"
Claude: "I can see your full architecture! Based on your current structure with UserService in src/services/ and your existing TypeScript types, here's the best approach..."
```

**Result**: Faster development, better code quality, more accurate suggestions, and seamless AI collaboration.

## 📈 Performance

- **Analysis Speed**: 35ms for 15 files, 142 symbols
- **Memory Efficient**: <50MB for large projects  
- **Token Optimized**: Compressed context maintains quality while reducing size
- **Incremental Updates**: Only regenerate changed files

## 🔧 Technical Details

### Supported Languages
- **TypeScript/JavaScript**: Full Tree-sitter AST parsing
- **Go**: Complete language support with Tree-sitter
- **JSON/YAML**: Basic parsing and structure analysis
- **Extensible**: Plugin architecture for additional languages

### Architecture
- **Virtual Graph Engine**: Incremental analysis with shadow/actual graph pattern
- **Multi-threaded**: Parallel file processing for performance
- **Caching Layer**: Smart caching for faster subsequent runs
- **Cross-platform**: Go-based with CGO for Tree-sitter integration

---

**Start building better software with AI assistance today! 🚀**

**[Download CodeContext v2.0.0](https://github.com/nmakod/codecontext/releases/tag/v2.0.0)**
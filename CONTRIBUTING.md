# Contributing to CodeContext

Thank you for your interest in contributing to CodeContext! This document provides guidelines and information for contributors.

## 🚀 Getting Started

### Prerequisites

- Go 1.19 or higher
- Git
- Basic understanding of Tree-sitter parsers (helpful but not required)

### Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/nmakod/codecontext.git
   cd codecontext
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Build the project**
   ```bash
   go build ./cmd/codecontext
   ```

4. **Run tests**
   ```bash
   go test ./...
   ```

## 📋 How to Contribute

### Reporting Issues

- Use the GitHub issue tracker
- Search existing issues before creating new ones
- Provide clear steps to reproduce bugs
- Include system information (OS, Go version, etc.)

### Suggesting Features

- Open an issue with the "enhancement" label
- Describe the problem you're trying to solve
- Explain why this feature would be valuable
- Consider implementation complexity

### Code Contributions

1. **Fork the repository**
2. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make your changes**
   - Write clear, readable code
   - Follow existing code style
   - Add tests for new functionality
   - Update documentation if needed

4. **Test your changes**
   ```bash
   go test ./...
   go build ./cmd/codecontext
   ./codecontext --help  # Basic smoke test
   ```

5. **Commit your changes**
   ```bash
   git commit -m "Add feature: your feature description"
   ```

6. **Push and create a pull request**
   ```bash
   git push origin feature/your-feature-name
   ```

## 🏗️ Project Structure

```
codecontext/
├── cmd/codecontext/         # Main application entry point
├── internal/
│   ├── analyzer/           # Code analysis and graph building
│   ├── parser/             # Tree-sitter language parsers
│   ├── git/                # Git integration and pattern detection
│   ├── mcp/                # Model Context Protocol server
│   ├── cli/                # Command-line interface
│   └── ...
├── pkg/types/              # Public types and interfaces
└── test/                   # Integration tests
```

## 🎯 Code Guidelines

### Go Style

- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and reasonably sized

### Testing

- Write unit tests for new functionality
- Aim for good test coverage
- Use table-driven tests where appropriate
- Include integration tests for major features

### Git Commits

- Use clear, descriptive commit messages
- Start with a verb in present tense ("Add", "Fix", "Update")
- Keep the first line under 50 characters
- Add detailed description if needed

Example:
```
Add support for Rust language parsing

- Integrate Tree-sitter Rust grammar
- Add Rust-specific symbol extraction
- Update file type detection
- Add comprehensive test coverage
```

## 🔧 Development Tips

### Adding Language Support

1. Add Tree-sitter grammar to `internal/parser/manager.go`
2. Update file extension detection in `internal/analyzer/graph.go`
3. Add language-specific symbol extraction logic
4. Write comprehensive tests
5. Update documentation

### Testing Changes

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Test specific package
go test ./internal/parser/

# Run integration tests
go test ./test/
```

### Debugging

- Use `debug.log` for development debugging
- The `-v` flag enables verbose output
- Test with various codebases to ensure robustness

## 📚 Resources

- [Tree-sitter Documentation](https://tree-sitter.github.io/tree-sitter/)
- [Model Context Protocol](https://modelcontextprotocol.io/)
- [Go Documentation](https://golang.org/doc/)

## 🤝 Community

- Be respectful and inclusive
- Help others learn and grow
- Share knowledge and best practices
- Provide constructive feedback

## 📄 License

By contributing to CodeContext, you agree that your contributions will be licensed under the MIT License.

## ❓ Questions?

Feel free to open an issue for questions about contributing, or reach out to the maintainers.

Thank you for contributing to CodeContext! 🎉
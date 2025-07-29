# CodeContext Assistant Guide

This file helps AI assistants understand the CodeContext project structure and conventions.

## Project Overview

CodeContext is a code analysis tool that generates comprehensive context maps for AI assistants. It uses Tree-sitter for parsing and provides MCP (Model Context Protocol) integration.

## Key Technologies

- **Language**: Go 1.24.5
- **Parser**: Tree-sitter with CGO bindings
- **Architecture**: Modular with clean interfaces
- **CI/CD**: GitHub Actions with automated releases

## Project Structure

```
.
├── cmd/codecontext/      # Main application entry point
├── internal/             # Internal packages
│   ├── analyzer/         # Code analysis and graph building
│   ├── mcp/             # MCP server implementation
│   ├── parser/          # Tree-sitter parser management
│   └── watcher/         # File system monitoring
├── pkg/types/           # Public type definitions
├── scripts/             # Utility scripts
└── test/                # Integration tests
```

## Development Conventions

See [.claude/conventions.md](.claude/conventions.md) for:
- Commit message format (Conventional Commits)
- Code style guidelines
- Testing requirements
- Release process

## Common Tasks

### Running Tests
```bash
go test ./...
```

### Building
```bash
go build -o codecontext ./cmd/codecontext
```

### Creating a Release
```bash
# Update CHANGELOG.md first
./scripts/tag-release.sh
git push origin v2.4.1
```

## Important Context

1. **MCP Integration**: The MCP server communicates via stdin/stdout with JSON-RPC. Never output non-JSON to stdout in MCP mode.

2. **Tree-sitter**: Requires CGO, which means we need platform-specific builds. Cannot cross-compile easily.

3. **Version Management**: Version is stored in the `VERSION` file and read by Makefile.

4. **Security**: Always run `govulncheck` before releases. We use Nancy, govulncheck, and Gosec in CI.

## Recent Changes

- Fixed MCP stdout corruption by redirecting logs to stderr
- Simplified CI/CD workflows to single build.yml
- Added Release Please for automated changelog management
- Security fix: Updated mapstructure to v2.3.0

## AI Assistant Tips

1. Always check existing patterns before implementing new features
2. Run tests to verify changes
3. Use conventional commits for version management
4. Keep security in mind - run security checks
5. Maintain backward compatibility
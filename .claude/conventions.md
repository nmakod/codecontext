# CodeContext Development Conventions

## Commit Message Format

We use [Conventional Commits](https://www.conventionalcommits.org/) for clear and automated versioning.

### Quick Reference

```
<type>(<scope>): <subject>
```

### Common Types

| Type | Description | Version Bump | Example |
|------|-------------|--------------|---------|
| `feat` | New feature | MINOR | `feat: add MCP tool for analysis` |
| `fix` | Bug fix | PATCH | `fix: resolve memory leak` |
| `docs` | Documentation | None | `docs: update API guide` |
| `style` | Code style | None | `style: format with gofmt` |
| `refactor` | Code restructure | None | `refactor: simplify workflow` |
| `test` | Tests | None | `test: add integration tests` |
| `chore` | Maintenance | None | `chore: update dependencies` |
| `perf` | Performance | PATCH | `perf: optimize parser` |
| `ci` | CI/CD changes | None | `ci: update build workflow` |

### Breaking Changes

Add `!` after type for breaking changes:
```bash
feat!: redesign API structure
```

### Examples from This Project

```bash
# Features
feat: add automated versioned releases
feat(mcp): implement file watcher tool

# Fixes
fix: redirect logs to stderr for MCP
fix(security): update vulnerable dependency

# Others
docs: update CHANGELOG for v2.4.1
chore: prepare release v2.4.1
test: fix flaky TestMCPWatchChanges
```

## Code Style

- Use `gofmt` for Go code formatting
- Follow Go idioms and best practices
- Keep functions focused and testable
- Document exported functions

## Testing

- Run tests before committing: `go test ./...`
- Maintain test coverage above 80%
- Fix flaky tests immediately

## Release Process

1. Update `CHANGELOG.md` with changes
2. Commit with conventional message
3. Create tag: `./scripts/tag-release.sh`
4. Push tag: `git push origin v2.4.1`

## Security

- Run security checks: `make security`
- Update dependencies regularly
- Never commit secrets or API keys
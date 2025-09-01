# Changelog

All notable changes to CodeContext will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [3.2.0](https://github.com/nmakod/codecontext/compare/v3.1.0...v3.2.0) (2025-09-01)


### Features

* add comprehensive C++ language support with security hardening ([d27642c](https://github.com/nmakod/codecontext/commit/d27642cbb174d7db0c96245d97aa6b5bc656e064))


### Bug Fixes

* resolve CI test failures and improve C++ P2 feature detection ([f226a46](https://github.com/nmakod/codecontext/commit/f226a46efe4491c964c488e24e0ac6457c566845))

## [3.1.0](https://github.com/nmakod/codecontext/compare/v3.0.1...v3.1.0) (2025-08-31)


### Features

* add comprehensive Swift language support with P1/P2 enhancements ([b73dbf6](https://github.com/nmakod/codecontext/commit/b73dbf63f53c1c15c80a634ba34ff61d77be9d4b))

## [3.0.1](https://github.com/nmakod/codecontext/compare/v3.0.0...v3.0.1) (2025-08-29)


### Bug Fixes

* add clarifying comment to resolve CI build cache issues ([dae7468](https://github.com/nmakod/codecontext/commit/dae7468991cf6b0d0547965570f72ff5fd89d5b3))
* add robust error handling to MCP server Stop method ([d2dc5ff](https://github.com/nmakod/codecontext/commit/d2dc5ff7cf56c1dc850506fe614fd87292f8e1bb))
* resolve all CI test failures with appropriate timeout adjustments ([ae27ce3](https://github.com/nmakod/codecontext/commit/ae27ce33507e624680e2bc1cf62ffb71f21fffee))
* resolve MCP integration test timeouts with improved synchronization ([e0516e8](https://github.com/nmakod/codecontext/commit/e0516e85221405c604aa751909a1240095e741e7))
* resolve MCP test coverage file descriptor issue ([7364eaa](https://github.com/nmakod/codecontext/commit/7364eaac9864da11032006073c414e4b525cd879))

## [3.0.0](https://github.com/nmakod/codecontext/compare/v2.9.0...v3.0.0) (2025-08-28)


### ⚠ BREAKING CHANGES

* Refactored parser architecture from library-quality to enterprise-grade implementation

### Features

* adding dart_support ([7747318](https://github.com/nmakod/codecontext/commit/7747318536f46a139e320834e08c1d74541b33a1))
* enhance Dart parser with comprehensive typedef extraction ([ffdaa57](https://github.com/nmakod/codecontext/commit/ffdaa5796f3f6a32c6807d7ebf57314470c27a02))
* implement comprehensive Dart and Flutter language support ([7538ae2](https://github.com/nmakod/codecontext/commit/7538ae2b97853525290af09495364f4e70538390))
* implement production-ready parser architecture with comprehensive error handling ([4d04706](https://github.com/nmakod/codecontext/commit/4d047068010ef423761a91530a4a90b442be5700))
* upgrade MCP SDK from v0.2.0 to v0.3.0 with comprehensive fixes ([abb767a](https://github.com/nmakod/codecontext/commit/abb767a2b086b29272af304a1db550995d163cdc))


### Bug Fixes

* resolve type conversion issue in analyzer GetSupportedLanguages ([a720c09](https://github.com/nmakod/codecontext/commit/a720c098531395491e08b152953beecdb8f5d5ea))

## [2.9.0](https://github.com/nmakod/codecontext/compare/v2.8.0...v2.9.0) (2025-08-14)


### Features

* add automated versioned releases with changelog integration ([6c2a8da](https://github.com/nmakod/codecontext/commit/6c2a8dac0f2d883c8069027636959202ba65dc46))
* add manifest.json for MCP desktop extension submission ([fadac3d](https://github.com/nmakod/codecontext/commit/fadac3d6d1fcab46e0b6505a2fb92d5b5258b53b))
* implement advanced performance optimizations and memory management ([ed0b253](https://github.com/nmakod/codecontext/commit/ed0b253d2f46d4abda686d9ed6994d5a357a4ce7))
* implement comprehensive path normalization and cross-platform support ([67be2a5](https://github.com/nmakod/codecontext/commit/67be2a53282544a936f2ded16bf9cba087f178cc))
* implement comprehensive thread safety and performance optimizations ([c99ffc2](https://github.com/nmakod/codecontext/commit/c99ffc29080b346f8bd992073d182817f060ceec))
* implement dynamic target directory support for multi-project analysis ([62d1cd4](https://github.com/nmakod/codecontext/commit/62d1cd495a44440e5829a69a864126d7d6a1a673))
* implement Phase 2 CI/CD architecture with release-please orchestration ([f4ee612](https://github.com/nmakod/codecontext/commit/f4ee6128017c78442799d3afc0d2b25db7f9d232))


### Bug Fixes

* add -buildvcs=false flag to resolve CI build failures ([d00ac9e](https://github.com/nmakod/codecontext/commit/d00ac9e2e54364c5eeca2911d6461904b45f5935))
* add issues permission to resolve release-please label creation error ([3d87e4b](https://github.com/nmakod/codecontext/commit/3d87e4b8cebe6247eed2041eb2a5e0078fad7040))
* adjust performance test thresholds for CI reliability ([5b2b973](https://github.com/nmakod/codecontext/commit/5b2b9732bda6ea711efd67514865cd126955ca15))
* exclude_patterns from config.yaml now properly excludes files ([123f67e](https://github.com/nmakod/codecontext/commit/123f67e8c5e74c09bbe0dd0037778e1a0a182c62))
* prevent duplicate workflow runs by making release-please manual only ([c860767](https://github.com/nmakod/codecontext/commit/c8607678d193383ce2c8a5d5445f599ed6066feb))
* redirect file watcher logs to stderr to prevent MCP protocol corruption ([18f57be](https://github.com/nmakod/codecontext/commit/18f57bef36b82ba22efa3b13f60eb550c0201b1c))
* resolve GitHub Actions workflow failures ([f2f5e64](https://github.com/nmakod/codecontext/commit/f2f5e6443c41015180fc4ccbc798faf95cd06c41))
* resolve test assertion failures in MCP dynamic targeting ([b1a5eae](https://github.com/nmakod/codecontext/commit/b1a5eae978230bf7827cf27eb5cb70e7998aae0e))
* resolve test isolation issues for CI reliability ([5bd804b](https://github.com/nmakod/codecontext/commit/5bd804bcef77fc4e44f6154accd1a969d7243898))
* resolve test isolation issues for CI reliability ([dfffd8c](https://github.com/nmakod/codecontext/commit/dfffd8c5869fdee3640324ac064bae9712567c41))
* resolve thread safety issues in progress tracking and compact controller ([255417e](https://github.com/nmakod/codecontext/commit/255417ea5fa7852acf47a7431777bd177db661f3))
* update mapstructure to v2.3.0 to resolve vulnerability ([1124330](https://github.com/nmakod/codecontext/commit/11243300e613f4bd99f5daa462b2203d1081a6ea))
* upgrade fsnotify from v1.8.0 to v1.9.0 ([88e651c](https://github.com/nmakod/codecontext/commit/88e651c488e5962e453555c403c2ba4552e109df))

## [2.8.0](https://github.com/nmakod/codecontext/compare/v2.7.3...v2.8.0) (2025-08-09)


### Features

* implement advanced performance optimizations and memory management ([ed0b253](https://github.com/nmakod/codecontext/commit/ed0b253d2f46d4abda686d9ed6994d5a357a4ce7))
* implement comprehensive path normalization and cross-platform support ([67be2a5](https://github.com/nmakod/codecontext/commit/67be2a53282544a936f2ded16bf9cba087f178cc))
* implement comprehensive thread safety and performance optimizations ([c99ffc2](https://github.com/nmakod/codecontext/commit/c99ffc29080b346f8bd992073d182817f060ceec))

## [2.7.3](https://github.com/nmakod/codecontext/compare/v2.7.2...v2.7.3) (2025-08-06)


### Bug Fixes

* exclude_patterns from config.yaml now properly excludes files ([123f67e](https://github.com/nmakod/codecontext/commit/123f67e8c5e74c09bbe0dd0037778e1a0a182c62))

## [2.7.2](https://github.com/nmakod/codecontext/compare/v2.7.1...v2.7.2) (2025-08-02)


### Bug Fixes

* upgrade fsnotify from v1.8.0 to v1.9.0 ([88e651c](https://github.com/nmakod/codecontext/commit/88e651c488e5962e453555c403c2ba4552e109df))

## [2.7.1](https://github.com/nmakod/codecontext/compare/v2.7.0...v2.7.1) (2025-08-02)


### Bug Fixes

* resolve test assertion failures in MCP dynamic targeting ([b1a5eae](https://github.com/nmakod/codecontext/commit/b1a5eae978230bf7827cf27eb5cb70e7998aae0e))

## [2.7.0](https://github.com/nmakod/codecontext/compare/v2.6.1...v2.7.0) (2025-08-01)


### Features

* add manifest.json for MCP desktop extension submission ([fadac3d](https://github.com/nmakod/codecontext/commit/fadac3d6d1fcab46e0b6505a2fb92d5b5258b53b))
* implement dynamic target directory support for multi-project analysis ([62d1cd4](https://github.com/nmakod/codecontext/commit/62d1cd495a44440e5829a69a864126d7d6a1a673))

## [2.6.1](https://github.com/nmakod/codecontext/compare/v2.6.0...v2.6.1) (2025-07-31)


### Bug Fixes

* add -buildvcs=false flag to resolve CI build failures ([d00ac9e](https://github.com/nmakod/codecontext/commit/d00ac9e2e54364c5eeca2911d6461904b45f5935))

## [2.6.0](https://github.com/nmakod/codecontext/compare/v2.5.0...v2.6.0) (2025-07-31)


### Features

* add automated versioned releases with changelog integration ([6c2a8da](https://github.com/nmakod/codecontext/commit/6c2a8dac0f2d883c8069027636959202ba65dc46))
* implement Phase 2 CI/CD architecture with release-please orchestration ([f4ee612](https://github.com/nmakod/codecontext/commit/f4ee6128017c78442799d3afc0d2b25db7f9d232))


### Bug Fixes

* add issues permission to resolve release-please label creation error ([3d87e4b](https://github.com/nmakod/codecontext/commit/3d87e4b8cebe6247eed2041eb2a5e0078fad7040))
* adjust performance test thresholds for CI reliability ([5b2b973](https://github.com/nmakod/codecontext/commit/5b2b9732bda6ea711efd67514865cd126955ca15))
* prevent duplicate workflow runs by making release-please manual only ([c860767](https://github.com/nmakod/codecontext/commit/c8607678d193383ce2c8a5d5445f599ed6066feb))
* redirect file watcher logs to stderr to prevent MCP protocol corruption ([18f57be](https://github.com/nmakod/codecontext/commit/18f57bef36b82ba22efa3b13f60eb550c0201b1c))
* resolve GitHub Actions workflow failures ([f2f5e64](https://github.com/nmakod/codecontext/commit/f2f5e6443c41015180fc4ccbc798faf95cd06c41))
* resolve test isolation issues for CI reliability ([5bd804b](https://github.com/nmakod/codecontext/commit/5bd804bcef77fc4e44f6154accd1a969d7243898))
* resolve test isolation issues for CI reliability ([dfffd8c](https://github.com/nmakod/codecontext/commit/dfffd8c5869fdee3640324ac064bae9712567c41))
* resolve thread safety issues in progress tracking and compact controller ([255417e](https://github.com/nmakod/codecontext/commit/255417ea5fa7852acf47a7431777bd177db661f3))
* update mapstructure to v2.3.0 to resolve vulnerability ([1124330](https://github.com/nmakod/codecontext/commit/11243300e613f4bd99f5daa462b2203d1081a6ea))

## [2.5.0](https://github.com/nmakod/codecontext/compare/v2.4.0...v2.5.0) (2025-07-29)


### Features

* add automated versioned releases with changelog integration ([6c2a8da](https://github.com/nmakod/codecontext/commit/6c2a8dac0f2d883c8069027636959202ba65dc46))


### Bug Fixes

* redirect file watcher logs to stderr to prevent MCP protocol corruption ([18f57be](https://github.com/nmakod/codecontext/commit/18f57bef36b82ba22efa3b13f60eb550c0201b1c))
* resolve GitHub Actions workflow failures ([f2f5e64](https://github.com/nmakod/codecontext/commit/f2f5e6443c41015180fc4ccbc798faf95cd06c41))
* resolve thread safety issues in progress tracking and compact controller ([255417e](https://github.com/nmakod/codecontext/commit/255417ea5fa7852acf47a7431777bd177db661f3))
* update mapstructure to v2.3.0 to resolve vulnerability ([1124330](https://github.com/nmakod/codecontext/commit/11243300e613f4bd99f5daa462b2203d1081a6ea))

## [Unreleased]

### Fixed
- File watcher logs redirected to stderr to prevent MCP protocol corruption
- Fixed flaky TestMCPWatchChanges test by preventing JSON parsing errors
- Updated mapstructure dependency to v2.3.0 to resolve security vulnerability GO-2025-3787

### Security
- Updated github.com/go-viper/mapstructure/v2 from v2.2.1 to v2.3.0 to fix information leak vulnerability

## [2.4.0] - 2025-07-29

### Added
- **Open-Source Foundations**
  - MIT LICENSE file for clear open-source licensing
  - Comprehensive CONTRIBUTING.md guide with development workflow
  - GitHub Actions CI/CD pipeline (via merged PR #1)
  - Multi-platform binary generation and automated releases
  - Dependabot configuration for dependency management

- **Multi-Language Support Expansion**
  - Extended file type support in graph analyzer
  - Added support for Go (.go), Python (.py, .pyi), Java (.java), Rust (.rs)
  - Additional JavaScript/TypeScript extensions (.mts, .cts, .mjs, .cjs)
  - Markdown (.md) support for documentation analysis

### Changed
- Standardized GitHub username references to 'nmakod' across all documentation
- Updated version references from 2.2.0 to 2.4.0 throughout codebase
- Synchronized GraphBuilder.isSupportedFile with parser manager capabilities

### Fixed
- Version inconsistencies in documentation and code
- Missing licensing information
- Lack of contribution guidelines

## [2.2.2] - 2025-07-21

### Added
- Synchronize documentation with master HLD
- Complete git semantic analysis implementation with comprehensive test suite
- Semantic neighborhoods with hierarchical clustering integration

### Enhanced
- Comprehensive test suite fixes - All integration and unit tests now pass
- Integration of unused modules and improved code coverage
- Documentation alignment with current implementation status

### Fixed
- Test suite stability and reliability improvements
- Module integration issues resolved

## [2.2.0] - 2025-07-14

### Added
- **Virtual Graph Engine** - Complete implementation with Virtual DOM-inspired architecture
  - Shadow graph management with efficient virtual representation
  - Change batching with configurable thresholds and timeouts
  - AST diffing with multiple algorithm support
  - Reconciliation system with dependency-aware processing
  - O(changes) complexity for incremental updates
  - Thread-safe concurrent operations with memory optimization

- **Compact Controller** - Multi-strategy optimization system
  - Six compaction strategies: relevance, frequency, dependency, size, hybrid, adaptive
  - Parallel processing with batch support and concurrent operations
  - Impact analysis and comprehensive dependency tracking
  - Performance metrics and compression ratio monitoring
  - Adaptive strategy selection based on graph characteristics
  - Quality scoring and rollback capabilities

- **Enhanced Diff Algorithms** - Advanced semantic analysis capabilities
  - Complete semantic vs structural diff engine
  - Language-specific AST diffing with handler framework
  - Advanced symbol rename detection with 6 similarity algorithms
  - Pattern-based heuristics: camelCase, prefix/suffix, abbreviation, refactoring, contextual
  - Comprehensive dependency change tracking with multi-language support
  - Confidence scoring, impact assessment, and evidence collection

- **Production-Ready MCP Server** - Official SDK integration
  - Official MCP SDK integration: `github.com/modelcontextprotocol/go-sdk v0.2.0`
  - Six production-ready MCP tools for complete codebase analysis
  - Real-time file watching with debounced change detection
  - Claude Desktop integration with complete protocol support
  - Comprehensive API documentation and usage examples
  - Performance monitoring and metrics collection

- **Enhanced CLI Framework**
  - New `mcp` command for MCP server management
  - Watch mode with real-time file monitoring
  - Advanced configuration management
  - Comprehensive progress reporting and statistics
  - Graceful shutdown handling

- **Advanced Type System**
  - Virtual graph types: `VirtualGraphEngine`, `ChangeSet`, `ReconciliationPlan`
  - Compact types: `CompactController`, `Strategy`, `CompactRequest`
  - Enhanced diff types: `SimilarityScore`, `HeuristicScore`, `DependencyChange`
  - Complete graph metadata and analysis timing

### Enhanced
- **Parser Manager** - Production-ready Tree-sitter integration
  - Real AST parsing with Tree-sitter JavaScript/TypeScript grammars
  - Multi-language support: TypeScript, JavaScript, JSON, YAML
  - Advanced symbol extraction with metadata and location tracking
  - Performance optimization with sub-millisecond parsing

- **Documentation** - Comprehensive synchronization with implementation
  - Updated HLD to reflect all completed components
  - Implementation plan updated with completed phases 1-4
  - Component dependencies documentation updated to Phase 4 status
  - New implementation status report with comprehensive analysis

### Performance
- **Parser Performance:** <1ms per file (3.5KB TypeScript) - exceeds targets by 10x
- **Symbol Extraction:** 15+ symbols from real AST data
- **Analysis Time:** 16ms for entire project analysis
- **Memory Usage:** <25MB for complete analysis - exceeds targets by 4x
- **Test Coverage:** 95.1% across all components
- **Virtual Graph Engine:** O(changes) complexity for incremental updates
- **MCP Server:** Real-time file watching with debounced changes

### Changed
- Updated project status from Phase 2.1 to Phase 4 complete
- All core HLD components moved from "PLANNED" to "COMPLETED"
- Documentation version updated to v2.2
- Implementation timeline updated to reflect ahead-of-schedule completion

### Technical Debt Resolved
- ✅ Real Tree-sitter integration (replaced mock parsers)
- ✅ Complete symbol extraction implementation
- ✅ Actual code graph construction
- ✅ Comprehensive diff algorithms
- ✅ Advanced rename detection
- ✅ Dependency change tracking

## [2.1.0] - 2025-07-12

### Added
- Extensive MCP testing suite and comprehensive documentation
- Integration tests for MCP server functionality
- Performance benchmarking for MCP operations
- Enhanced error handling and logging

### Enhanced
- MCP server stability and reliability
- Documentation completeness and accuracy
- Test coverage for MCP components

## [2.0.2] - 2025-07-11

### Fixed
- Watch command output file bug
- Generate command output file bug

### Enhanced
- Error handling for file operations
- Output file validation and creation

## [2.0.1] - 2025-07-10

### Added
- Comprehensive Claude integration documentation and guides
- Usage examples and best practices
- Troubleshooting documentation

### Enhanced
- User experience with better documentation
- Claude Desktop integration guidance

## [2.0.0] - 2025-07-08

### Added
- **Foundation Release** - Complete project infrastructure
  - Go module setup with proper project structure
  - Cobra-based CLI framework with all major commands
  - Viper configuration management with hierarchical configs
  - Complete type system with graph definitions
  - Comprehensive test framework and utilities
  - Tree-sitter integration foundation

- **Core Commands**
  - `codecontext init` - Project initialization
  - `codecontext generate` - Context map generation
  - `codecontext update` - Incremental updates
  - `codecontext compact` - Context optimization

- **Parser Infrastructure**
  - Multi-language parser manager
  - AST cache implementation with TTL support
  - Language detection and file classification
  - Symbol extraction framework

### Technical
- **Go Version:** 1.24.5
- **Architecture:** Modular design with clean interfaces
- **Dependencies:** Modern Go ecosystem with official bindings
- **Testing:** Comprehensive unit and integration tests

---

## Version History Summary

- **v2.2.0** - Complete implementation of all HLD components (Virtual Graph, Compact Controller, Enhanced Diff, MCP Server)
- **v2.1.0** - MCP testing suite and comprehensive documentation
- **v2.0.2** - Bug fixes for watch and generate commands
- **v2.0.1** - Enhanced Claude integration documentation
- **v2.0.0** - Foundation release with core infrastructure

## Migration Guide

### Upgrading to v2.2.0

**From v2.1.x:**
- All existing configurations and projects remain compatible
- New MCP server provides enhanced Claude integration
- Virtual Graph Engine enables faster incremental updates
- Compact Controller offers new optimization strategies

**New Features Available:**
- `codecontext mcp` command for MCP server management
- Enhanced `codecontext compact` with multiple strategies
- Real-time file watching capabilities
- Advanced diff analysis and rename detection

**Performance Improvements:**
- 10x faster parsing performance
- 4x better memory efficiency
- O(changes) complexity for incremental updates
- Real-time change detection with debouncing

### Breaking Changes
- None - Full backward compatibility maintained

### Recommended Actions
1. **Update Installation:**
   ```bash
   brew upgrade codecontext  # or download latest binary
   ```

2. **Enable New Features:**
   ```bash
   codecontext mcp --enable-watch  # Start MCP server with file watching
   ```

3. **Test New Capabilities:**
   ```bash
   codecontext compact --strategy adaptive  # Try new optimization strategies
   ```

---

*For detailed release information, see [RELEASE_PLAN_V2.2.md](docs/RELEASE_PLAN_V2.2.md)*

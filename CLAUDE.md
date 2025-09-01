# CodeContext Map

**Generated:** 2025-08-31T18:16:48+05:30  
**Version:** 2.0.0  
**Analysis Time:** 10.366986959s  
**Status:** Real Tree-sitter Analysis

## 📊 Overview

This context map was generated using **real Tree-sitter parsing** and provides comprehensive analysis of your codebase:

- **Files Analyzed**: 120 files
- **Symbols Extracted**: 2265 symbols  
- **Languages Detected**: 3 languages
- **Import Relationships**: 110 file dependencies

### 🎯 Analysis Capabilities
- ✅ **Real AST Parsing** - Tree-sitter JavaScript/TypeScript grammars
- ✅ **Symbol Extraction** - Functions, classes, methods, variables, imports
- ✅ **Dependency Analysis** - File-to-file relationship mapping
- ✅ **Multi-language Support** - TypeScript, JavaScript, JSON, YAML

## 📁 File Analysis

| File | Language | Lines | Symbols | Imports | Type |
|------|----------|-------|---------|---------|------|
| `.claude/settings.local.json` | json | 27 | 0 | 0 | source |
| `.codecontext/config.yaml` | yaml | 80 | 0 | 0 | source |
| `.github/dependabot.yml` | yaml | 23 | 0 | 0 | source |
| `.github/release.yml` | yaml | 36 | 0 | 0 | source |
| `.release-please-manifest.json` | json | 3 | 0 | 0 | source |
| `cmd/codecontext/main.go` | go | 24 | 3 | 1 | source |
| `internal/analyzer/graph.go` | go | 1261 | 45 | 1 | source |
| `internal/analyzer/graph_test.go` | go | 1532 | 48 | 1 | test |
| `internal/analyzer/incremental.go` | go | 589 | 24 | 1 | source |
| `internal/analyzer/incremental_test.go` | go | 680 | 15 | 1 | test |
| `internal/analyzer/markdown.go` | go | 748 | 34 | 1 | source |
| `internal/analyzer/relationships.go` | go | 486 | 22 | 1 | source |
| `internal/analyzer/relationships_test.go` | go | 473 | 12 | 1 | test |
| `internal/cache/persistent.go` | go | 597 | 37 | 1 | source |
| `internal/cache/persistent_test.go` | go | 526 | 14 | 1 | test |
| `internal/cli/compact.go` | go | 198 | 9 | 1 | source |
| `internal/cli/compact_test.go` | go | 129 | 4 | 1 | test |
| `internal/cli/generate.go` | go | 175 | 5 | 1 | source |
| `internal/cli/init.go` | go | 203 | 4 | 1 | source |
| `internal/cli/init_test.go` | go | 115 | 3 | 1 | test |
| `internal/cli/integration_test.go` | go | 799 | 31 | 1 | test |
| `internal/cli/mcp.go` | go | 119 | 4 | 1 | source |
| `internal/cli/mcp_test.go` | go | 656 | 11 | 1 | test |
| `internal/cli/progress.go` | go | 611 | 44 | 1 | source |
| `internal/cli/progress_test.go` | go | 500 | 30 | 1 | test |
| `internal/cli/root.go` | go | 86 | 6 | 1 | source |
| `internal/cli/shutdown.go` | go | 408 | 31 | 1 | source |
| `internal/cli/update.go` | go | 116 | 6 | 1 | source |
| `internal/cli/watch.go` | go | 584 | 27 | 1 | source |
| `internal/compact/controller.go` | go | 466 | 28 | 1 | source |
| `internal/compact/controller_test.go` | go | 715 | 20 | 1 | test |
| `internal/compact/strategy.go` | go | 859 | 49 | 1 | source |
| `internal/compact/strategy_test.go` | go | 698 | 23 | 1 | test |
| `internal/config/config.go` | go | 34 | 2 | 0 | source |
| `internal/diff/ast.go` | go | 641 | 59 | 1 | source |
| `internal/diff/dependency.go` | go | 781 | 88 | 1 | source |
| `internal/diff/engine.go` | go | 672 | 40 | 1 | source |
| `internal/diff/heuristics.go` | go | 641 | 51 | 1 | source |
| `internal/diff/rename.go` | go | 513 | 37 | 1 | source |
| `internal/diff/semantic.go` | go | 526 | 38 | 1 | source |
| `internal/diff/similarity.go` | go | 760 | 64 | 1 | source |
| `internal/diff/utils.go` | go | 39 | 4 | 0 | source |
| `internal/generator/markdown.go` | go | 69 | 5 | 1 | source |
| `internal/git/analyzer.go` | go | 354 | 25 | 1 | source |
| `internal/git/analyzer_test.go` | go | 425 | 14 | 1 | test |
| `internal/git/error_handling_test.go` | go | 119 | 15 | 1 | test |
| `internal/git/integration.go` | go | 1279 | 71 | 1 | source |
| `internal/git/integration_flow_test.go` | go | 402 | 6 | 1 | test |
| `internal/git/integration_test.go` | go | 779 | 27 | 1 | test |
| `internal/git/interfaces.go` | go | 23 | 3 | 1 | source |
| `internal/git/pattern_detection_integration_test.go` | go | 424 | 9 | 1 | test |
| `internal/git/patterns.go` | go | 667 | 34 | 1 | source |
| `internal/git/patterns_ignore_test.go` | go | 188 | 5 | 1 | test |
| `internal/git/patterns_test.go` | go | 341 | 13 | 1 | test |
| `internal/git/performance_benchmark_test.go` | go | 231 | 17 | 1 | test |
| `internal/git/semantic.go` | go | 495 | 31 | 1 | source |
| `internal/git/semantic_analysis_e2e_test.go` | go | 265 | 17 | 1 | test |
| `internal/git/semantic_test.go` | go | 477 | 14 | 1 | test |
| `internal/git/simple_patterns.go` | go | 219 | 14 | 1 | source |
| `internal/git/simple_patterns_test.go` | go | 457 | 11 | 1 | test |
| `internal/mcp/migration_test.go` | go | 93 | 5 | 1 | test |
| `internal/mcp/server.go` | go | 1343 | 50 | 1 | source |
| `internal/mcp/server_test.go` | go | 1040 | 22 | 1 | test |
| `internal/parser/builder.go` | go | 266 | 14 | 1 | source |
| `internal/parser/cache.go` | go | 194 | 17 | 1 | source |
| `internal/parser/cache_test.go` | go | 300 | 10 | 1 | test |
| `internal/parser/config.go` | go | 111 | 4 | 1 | source |
| `internal/parser/cpp_framework_test.go` | go | 324 | 3 | 1 | test |
| `internal/parser/cpp_integration_test.go` | go | 534 | 5 | 1 | test |
| `internal/parser/cpp_simple_test.go` | go | 222 | 10 | 1 | test |
| `internal/parser/cpp_templates_test.go` | go | 208 | 3 | 1 | test |
| `internal/parser/dart.go` | go | 2092 | 85 | 1 | source |
| `internal/parser/dart_debug_test.go` | go | 71 | 4 | 1 | test |
| `internal/parser/dart_enums_types_test.go` | go | 580 | 10 | 1 | test |
| `internal/parser/dart_mixins_extensions_test.go` | go | 652 | 16 | 1 | test |
| `internal/parser/dart_part_files_test.go` | go | 333 | 12 | 1 | test |
| `internal/parser/dart_performance_test.go` | go | 242 | 6 | 1 | test |
| `internal/parser/dart_simple_test.go` | go | 84 | 4 | 1 | test |
| `internal/parser/dart_test.go` | go | 401 | 10 | 1 | test |
| `internal/parser/errors.go` | go | 100 | 15 | 1 | source |
| `internal/parser/flutter.go` | go | 439 | 15 | 1 | source |
| `internal/parser/flutter_accuracy_test.go` | go | 452 | 3 | 1 | test |
| `internal/parser/flutter_advanced_build_test.go` | go | 267 | 2 | 1 | test |
| `internal/parser/flutter_advanced_test.go` | go | 518 | 6 | 1 | test |
| `internal/parser/flutter_integration_validation_test.go` | go | 782 | 8 | 1 | test |
| `internal/parser/flutter_symbol_classification_test.go` | go | 430 | 7 | 1 | test |
| `internal/parser/flutter_test.go` | go | 294 | 5 | 1 | test |
| `internal/parser/framework.go` | go | 398 | 15 | 1 | source |
| `internal/parser/integration_dart_test.go` | go | 130 | 7 | 1 | test |
| `internal/parser/integration_test.go` | go | 316 | 5 | 1 | test |
| `internal/parser/interfaces.go` | go | 154 | 11 | 1 | source |
| `internal/parser/logger.go` | go | 226 | 32 | 1 | source |
| `internal/parser/manager.go` | go | 2319 | 85 | 1 | source |
| `internal/parser/manager_test.go` | go | 693 | 12 | 1 | test |
| `internal/parser/panic_handler.go` | go | 153 | 17 | 1 | source |
| `internal/parser/swift.go` | go | 1020 | 24 | 1 | source |
| `internal/parser/swift_advanced_test.go` | go | 390 | 9 | 1 | test |
| `internal/parser/swift_framework_test.go` | go | 134 | 3 | 1 | test |
| `internal/parser/swift_integration_test.go` | go | 661 | 4 | 1 | test |
| `internal/parser/swift_p1_p2_test.go` | go | 706 | 5 | 1 | test |
| `internal/parser/swift_simple_test.go` | go | 192 | 8 | 1 | test |
| `internal/performance/monitor.go` | go | 457 | 36 | 1 | source |
| `internal/performance/monitor_test.go` | go | 382 | 22 | 1 | test |
| `internal/testutils/filesystem.go` | go | 1088 | 45 | 1 | source |
| `internal/testutils/filesystem_test.go` | go | 667 | 24 | 1 | test |
| `internal/vgraph/batcher.go` | go | 217 | 15 | 1 | source |
| `internal/vgraph/differ.go` | go | 563 | 38 | 1 | source |
| `internal/vgraph/engine.go` | go | 469 | 33 | 1 | source |
| `internal/vgraph/engine_test.go` | go | 518 | 13 | 1 | test |
| `internal/vgraph/reconciler.go` | go | 622 | 35 | 1 | source |
| `internal/watcher/watcher.go` | go | 322 | 16 | 1 | source |
| `internal/watcher/watcher_test.go` | go | 270 | 6 | 1 | test |
| `manifest.json` | json | 51 | 0 | 0 | source |
| `pkg/types/compact.go` | go | 221 | 14 | 1 | source |
| `pkg/types/dart_types.go` | go | 44 | 2 | 0 | source |
| `pkg/types/graph.go` | go | 231 | 21 | 1 | source |
| `pkg/types/graph_test.go` | go | 161 | 5 | 1 | test |
| `pkg/types/vgraph.go` | go | 172 | 17 | 1 | source |
| `release-please-config.json` | json | 28 | 0 | 0 | source |
| `test/mcp_integration_test.go` | go | 1553 | 32 | 1 | test |


## 🔍 Symbol Analysis

### Symbol Types

- 🏷️ **type**: 321
- 📦 **variable**: 259
- 📥 **import**: 110
- 🔧 **function**: 604
- ⚙️ **method**: 971


## 📈 Language Statistics

| Language | Files | Percentage |
|----------|-------|------------|
| go | 113 | 94.2% |
| json | 4 | 3.3% |
| yaml | 3 | 2.5% |


## 🔗 Import Analysis

- **Total Import Statements**: 110
- **Internal Imports**: 0 (relative paths)
- **External Imports**: 110 (packages/modules)
- **Unique Modules**: 1

### Most Imported Modules

| Module | Import Count |
|--------|-------------|
| `` | 110 |


## 🔗 Relationship Analysis

### 📊 Relationship Summary

- **Total Relationships**: 110
- **File-to-File**: 110
- **Symbol-to-Symbol**: 0
- **Cross-File References**: 0

### 🔍 Relationship Types

| Type | Count | Description |
|------|-------|-------------|
| imports | 110 | File imports another file |
| references | 0 | Symbol references another symbol |
| uses | 0 | Symbol uses another symbol |
| calls | 0 | Function/method calls another function/method |

### ✅ No Circular Dependencies

No circular dependencies detected in the codebase.

### 🏝️ Isolated Files

Files with no import/export relationships:

- `dart_types.go`
- `.release-please-manifest.json`
- `settings.local.json`
- `release.yml`
- `utils.go`
- `config.yaml`
- `release-please-config.json`
- `config.go`
- `dependabot.yml`
- `manifest.json`



## 🏘️ Semantic Code Neighborhoods

### 📊 Analysis Overview

This analysis uses **git history patterns** and **hierarchical clustering** to identify semantic code neighborhoods:

- **Analysis Period**: 30 days
- **Files with Patterns**: 21 files
- **Basic Neighborhoods**: 15 groups
- **Clustered Groups**: 3 clusters
- **Average Cluster Size**: 5.0 files
- **Analysis Time**: 9.439292833s
- **Clustering Quality**: Excellent


### 🔍 Semantic Neighborhoods

Files grouped by git change patterns and correlation:

#### swift + swift_integration_test

- **Correlation Strength**: 1.00
- **Change Frequency**: 1 changes
- **Last Changed**: 2025-08-31
- **Files**: 2 files

**Files in this neighborhood:**
- `swift.go`
- `swift_integration_test.go`

**Common Operations:**
- coordinated_changes

#### swift_p1_p2_test + swift_simple_test

- **Correlation Strength**: 1.00
- **Change Frequency**: 1 changes
- **Last Changed**: 2025-08-31
- **Files**: 2 files

**Files in this neighborhood:**
- `swift_p1_p2_test.go`
- `swift_simple_test.go`

**Common Operations:**
- coordinated_changes

#### generate + integration_test

- **Correlation Strength**: 1.00
- **Change Frequency**: 2 changes
- **Last Changed**: 2025-08-09
- **Files**: 2 files

**Files in this neighborhood:**
- `generate.go`
- `integration_test.go`

**Common Operations:**
- coordinated_changes

#### logger + manager_test

- **Correlation Strength**: 1.00
- **Change Frequency**: 1 changes
- **Last Changed**: 2025-08-25
- **Files**: 2 files

**Files in this neighborhood:**
- `logger.go`
- `manager_test.go`

**Common Operations:**
- coordinated_changes

#### flutter + integration_dart_test

- **Correlation Strength**: 1.00
- **Change Frequency**: 2 changes
- **Last Changed**: 2025-08-25
- **Files**: 2 files

**Files in this neighborhood:**
- `flutter.go`
- `integration_dart_test.go`

**Common Operations:**
- coordinated_changes

#### config + interfaces

- **Correlation Strength**: 1.00
- **Change Frequency**: 1 changes
- **Last Changed**: 2025-08-25
- **Files**: 2 files

**Files in this neighborhood:**
- `config.go`
- `interfaces.go`

**Common Operations:**
- coordinated_changes

#### dart + integration_dart_test

- **Correlation Strength**: 1.00
- **Change Frequency**: 2 changes
- **Last Changed**: 2025-08-25
- **Files**: 2 files

**Files in this neighborhood:**
- `dart.go`
- `integration_dart_test.go`

**Common Operations:**
- coordinated_changes

#### dart_enums_types_test + dart_mixins_extensions_test

- **Correlation Strength**: 1.00
- **Change Frequency**: 1 changes
- **Last Changed**: 2025-08-24
- **Files**: 2 files

**Files in this neighborhood:**
- `dart_enums_types_test.go`
- `dart_mixins_extensions_test.go`

**Common Operations:**
- coordinated_changes

#### config + errors

- **Correlation Strength**: 1.00
- **Change Frequency**: 1 changes
- **Last Changed**: 2025-08-25
- **Files**: 2 files

**Files in this neighborhood:**
- `config.go`
- `errors.go`

**Common Operations:**
- coordinated_changes

#### dart_mixins_extensions_test + flutter_test

- **Correlation Strength**: 1.00
- **Change Frequency**: 1 changes
- **Last Changed**: 2025-08-24
- **Files**: 2 files

**Files in this neighborhood:**
- `dart_mixins_extensions_test.go`
- `flutter_test.go`

**Common Operations:**
- coordinated_changes


### 🎯 Advanced Clustering Analysis

Neighborhoods grouped using **hierarchical clustering with Ward linkage**:

#### Cluster 1: + Group

- **Description**: Cluster of 11 neighborhoods containing 22 files with 0.60 average combined score
- **Size**: 11 files
- **Strength**: 0.773
- **Silhouette Score**: 0.945
- **Davies-Bouldin Index**: 0.927
- **Cohesion**: 0.945
- **Density**: 0.073

**Recommended Tasks:**
- testing
- debugging
- quality_assurance
- configuration
- setup
- deployment

**Why**: Strong cluster with good cohesion and regular co-changes

**Files in this cluster:**
- `dart_enums_types_test.go`
- `flutter_advanced_build_test.go`
- `dart_mixins_extensions_test.go`
- `flutter_symbol_classification_test.go`
- `swift.go`
- `swift_p1_p2_test.go`
- `dart_debug_test.go`
- `dart_test.go`
- `dart_performance_test.go`
- `flutter_test.go`
- `generate.go`
- `integration_test.go`
- `dart.go`
- `integration_dart_test.go`
- `builder.go`
- `manager_test.go`
- `flutter.go`
- `dart_part_files_test.go`

#### Cluster 2: + Group

- **Description**: Cluster of 2 neighborhoods containing 4 files with 0.61 average combined score
- **Size**: 2 files
- **Strength**: 0.806
- **Silhouette Score**: 1.000
- **Davies-Bouldin Index**: 1.000
- **Cohesion**: 1.000
- **Density**: 0.000

**Recommended Tasks:**
- testing
- debugging
- quality_assurance
- configuration
- setup
- deployment

**Why**: Very strong cluster with high cohesion and frequent co-changes

**Files in this cluster:**
- `flutter_accuracy_test.go`
- `flutter_symbol_classification_test.go`
- `dart.go`
- `flutter.go`

#### Cluster 3: + Group

- **Description**: Cluster of 2 neighborhoods containing 4 files with 0.62 average combined score
- **Size**: 2 files
- **Strength**: 0.812
- **Silhouette Score**: 0.999
- **Davies-Bouldin Index**: 1.000
- **Cohesion**: 0.999
- **Density**: 0.000

**Recommended Tasks:**
- testing
- debugging
- quality_assurance
- configuration
- setup
- deployment

**Why**: Very strong cluster with high cohesion and frequent co-changes

**Files in this cluster:**
- `config.go`
- `panic_handler.go`
- `flutter_accuracy_test.go`
- `flutter_advanced_build_test.go`


### 📈 Clustering Quality Assessment

**Overall Clustering Performance:**

- **Average Silhouette Score**: 0.981
- **Average Davies-Bouldin Index**: 0.976
- **Overall Quality Rating**: Excellent

**Quality Metrics Interpretation:**

- **Silhouette Score**: Measures how similar files are to their own cluster vs. other clusters
  - Range: -1 to 1 (higher is better)
  - >0.7: Excellent clustering, >0.5: Good, >0.25: Fair, <0.25: Poor
- **Davies-Bouldin Index**: Measures cluster separation and compactness
  - Range: 0+ (lower is better)
  - Values closer to 0 indicate better clustering

**Clustering Algorithm:**

- **Method**: Hierarchical Clustering with Ward Linkage
- **Features**: Git patterns + dependency analysis + structural similarity
- **Optimization**: Elbow method for optimal cluster count
- **Quality**: Real-time silhouette and Davies-Bouldin scoring



## 📁 Project Structure

```
.release-please-manifest.json
manifest.json
release-please-config.json
.claude/
├── settings.local.json
.codecontext/
├── config.yaml
.github/
├── dependabot.yml
├── release.yml
cmd/codecontext/
├── main.go
internal/analyzer/
├── graph.go
├── graph_test.go
├── incremental.go
├── incremental_test.go
├── markdown.go
├── relationships.go
├── relationships_test.go
internal/cache/
├── persistent.go
├── persistent_test.go
internal/cli/
├── compact.go
├── compact_test.go
├── generate.go
├── init.go
├── init_test.go
├── integration_test.go
├── mcp.go
├── mcp_test.go
├── progress.go
├── progress_test.go
├── root.go
├── shutdown.go
├── update.go
├── watch.go
internal/compact/
├── controller.go
├── controller_test.go
├── strategy.go
├── strategy_test.go
internal/config/
├── config.go
internal/diff/
├── ast.go
├── dependency.go
├── engine.go
├── heuristics.go
├── rename.go
├── semantic.go
├── similarity.go
├── utils.go
internal/generator/
├── markdown.go
internal/git/
├── analyzer.go
├── analyzer_test.go
├── error_handling_test.go
├── integration.go
├── integration_flow_test.go
├── integration_test.go
├── interfaces.go
├── pattern_detection_integration_test.go
├── patterns.go
├── patterns_ignore_test.go
├── patterns_test.go
├── performance_benchmark_test.go
├── semantic.go
├── semantic_analysis_e2e_test.go
├── semantic_test.go
├── simple_patterns.go
├── simple_patterns_test.go
internal/mcp/
├── migration_test.go
├── server.go
├── server_test.go
internal/parser/
├── builder.go
├── cache.go
├── cache_test.go
├── config.go
├── cpp_framework_test.go
├── cpp_integration_test.go
├── cpp_simple_test.go
├── cpp_templates_test.go
├── dart.go
├── dart_debug_test.go
├── dart_enums_types_test.go
├── dart_mixins_extensions_test.go
├── dart_part_files_test.go
├── dart_performance_test.go
├── dart_simple_test.go
├── dart_test.go
├── errors.go
├── flutter.go
├── flutter_accuracy_test.go
├── flutter_advanced_build_test.go
├── flutter_advanced_test.go
├── flutter_integration_validation_test.go
├── flutter_symbol_classification_test.go
├── flutter_test.go
├── framework.go
├── integration_dart_test.go
├── integration_test.go
├── interfaces.go
├── logger.go
├── manager.go
├── manager_test.go
├── panic_handler.go
├── swift.go
├── swift_advanced_test.go
├── swift_framework_test.go
├── swift_integration_test.go
├── swift_p1_p2_test.go
├── swift_simple_test.go
internal/performance/
├── monitor.go
├── monitor_test.go
internal/testutils/
├── filesystem.go
├── filesystem_test.go
internal/vgraph/
├── batcher.go
├── differ.go
├── engine.go
├── engine_test.go
├── reconciler.go
internal/watcher/
├── watcher.go
├── watcher_test.go
pkg/types/
├── compact.go
├── dart_types.go
├── graph.go
├── graph_test.go
├── vgraph.go
test/
├── mcp_integration_test.go
```


---

*Generated by CodeContext v2.0.0 with real Tree-sitter parsing*  
*Analysis completed in 10.366986959s*
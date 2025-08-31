# API Interfaces Specification

**Version:** 3.1.1  
**Status:** Production Ready - Enhanced with C++ Language Support and Security Hardening  
**Last Updated:** August 2025

## Overview

This document defines all core interfaces and APIs implemented in CodeContext v3.1.1. These interfaces reflect the production-ready implementation that exceeds the original HLD scope, including new components like Git Integration, MCP Server, Framework-Specific Analysis capabilities, comprehensive Swift language support, and security-hardened C++ language support.

## Table of Contents

1. [Core Interfaces](#core-interfaces)
2. [Framework Detection Interfaces](#framework-detection-interfaces) **Enhanced v3.0.1**
3. [Git Integration Interfaces](#git-integration-interfaces) **NEW**
4. [Enhanced Diff Engine Interfaces](#enhanced-diff-engine-interfaces) **NEW**
5. [MCP Server Interfaces](#mcp-server-interfaces) **Enhanced v3.0.1**
6. [Virtual Graph Interfaces](#virtual-graph-interfaces)
7. [Parser Interfaces](#parser-interfaces) **Enhanced v3.1.1 with C++ and Swift Support**
8. [Compact Controller Interfaces](#compact-controller-interfaces)
9. [Storage Interfaces](#storage-interfaces)
10. [REST API Specifications](#rest-api-specifications)
11. [GraphQL Schema](#graphql-schema)
12. [CLI Interface](#cli-interface)

## Core Interfaces

### Orchestrator Interface
```go
type Orchestrator interface {
    // Main operations
    GenerateMap(ctx context.Context, config ProjectConfig) (*MapResult, error)
    UpdateMap(ctx context.Context, changes FileChanges) (*MapResult, error)
    UpdateMapIncremental(ctx context.Context, changes FileChanges) (*PatchResult, error)
    
    // Configuration
    ValidateConfig(config ProjectConfig) ValidationResult
    GetSupportedLanguages() []Language
    
    // Operation control
    CancelOperation(operationId string) error
    ResumeFromCheckpoint(checkpointId string) (*MapResult, error)
    
    // Compact operations
    ExecuteCompactCommand(ctx context.Context, command CompactCommand) (*CompactResult, error)
}

type OrchestrationContext struct {
    Context        context.Context
    Cancel         context.CancelFunc
    Checkpoint     *Checkpoint
    ProgressChan   chan Progress
    VirtualGraph   VirtualGraphEngine
    Metrics        *MetricsCollector
}

type MapResult struct {
    Graph          *CodeGraph
    ProcessingTime time.Duration
    TokenCount     int
    FilesProcessed int
    Errors         []ProcessingError
    Checkpoints    []Checkpoint
    Metadata       map[string]interface{}
}

type PatchResult struct {
    Patches        []GraphPatch
    AffectedNodes  []NodeId
    TokenDelta     int
    ProcessingTime time.Duration
    Applied        bool
    Conflicts      []Conflict
}
```

### Progress Tracking
```go
type Progress struct {
    OperationId    string
    Stage          string
    Completed      int
    Total          int
    Percentage     float64
    CurrentFile    string
    EstimatedTime  time.Duration
    Message        string
    Timestamp      time.Time
}

type ProgressReporter interface {
    ReportProgress(progress Progress)
    SetTotal(total int)
    Increment(delta int)
    UpdateMessage(message string)
    Finish()
}
```

## Framework Detection Interfaces

### Framework Detector Interface
```go
type FrameworkDetector interface {
    // Main detection methods
    DetectFramework(filePath, language, content string) string
    
    // Strategy-specific detection
    DetectByFileExtension(filePath string) string
    DetectByImports(content string) string
    DetectByPackageJson(filePath string) string
    DetectPythonFramework(content string) string
    DetectJavaFramework(content string) string
    DetectSwiftFramework(content string) string
    
    // Caching and performance
    ClearCache() error
    GetCacheStats() CacheStats
}

type PackageInfo struct {
    Dependencies    map[string]string `json:"dependencies"`
    DevDependencies map[string]string `json:"devDependencies"`
    Scripts         map[string]string `json:"scripts"`
}

type CacheStats struct {
    PackageCacheSize   int
    FrameworkCacheSize int
    HitRate           float64
    MissRate          float64
}
```

### Framework-Specific Symbol Types
```go
// Framework-specific symbol types (added to SymbolType enum)
const (
    SymbolTypeComponent    SymbolType = "component"    // React, Vue, Angular, Svelte components
    SymbolTypeHook         SymbolType = "hook"         // React hooks
    SymbolTypeDirective    SymbolType = "directive"    // Angular directives
    SymbolTypeService      SymbolType = "service"      // Angular services
    SymbolTypeStore        SymbolType = "store"        // Svelte stores, Vue stores
    SymbolTypeComputed     SymbolType = "computed"     // Vue computed properties
    SymbolTypeWatcher      SymbolType = "watcher"      // Vue watchers
    SymbolTypeLifecycle    SymbolType = "lifecycle"    // Lifecycle methods/hooks
    SymbolTypeRoute        SymbolType = "route"        // Next.js pages, API routes
    SymbolTypeMiddleware   SymbolType = "middleware"   // Next.js middleware
    SymbolTypeAction       SymbolType = "action"       // Svelte actions, Vue actions
    // Swift-specific symbol types (NEW v3.0.1)
    SymbolTypeActor        SymbolType = "actor"        // Swift actors for concurrency
    SymbolTypePropertyWrapper SymbolType = "property_wrapper" // Swift property wrappers
    SymbolTypeResultBuilder   SymbolType = "result_builder"   // Swift result builders
    SymbolTypeMacro           SymbolType = "macro"            // Swift 5.9+ macros
)
```

### Framework Classification
```go
type FileClassification struct {
    Language    Language `json:"language"`
    FileType    string   `json:"file_type"` // "source", "test", "config", "documentation"
    IsGenerated bool     `json:"is_generated"`
    IsTest      bool     `json:"is_test"`
    Framework   string   `json:"framework,omitempty"` // NEW: Framework detection result
    Confidence  float64  `json:"confidence"`
}
```

### Framework Analysis Results
```go
type FrameworkAnalysis struct {
    Framework          string                    `json:"framework"`
    SymbolDistribution map[string]int           `json:"symbol_distribution"`
    TotalSymbols       int                      `json:"total_symbols"`
    Insights           []FrameworkInsight       `json:"insights"`
    BestPractices      []BestPractice          `json:"best_practices"`
    Recommendations    []Recommendation        `json:"recommendations"`
}

type FrameworkInsight struct {
    Type        string  `json:"type"`        // "pattern", "anti-pattern", "optimization"
    Message     string  `json:"message"`
    Confidence  float64 `json:"confidence"`
    Impact      string  `json:"impact"`      // "high", "medium", "low"
    FileCount   int     `json:"file_count"`
}

type BestPractice struct {
    Category    string   `json:"category"`    // "structure", "performance", "maintainability"
    Description string   `json:"description"`
    Examples    []string `json:"examples"`
    References  []string `json:"references"`
}

type Recommendation struct {
    Title       string   `json:"title"`
    Description string   `json:"description"`
    Priority    string   `json:"priority"`    // "high", "medium", "low"
    ActionItems []string `json:"action_items"`
    Effort      string   `json:"effort"`      // "low", "medium", "high"
}
```

## Git Integration Interfaces

### Git Analyzer Interface
```go
type GitAnalyzer interface {
    // Repository information
    IsGitRepository() bool
    GetBranchInfo() (string, error)
    GetRemoteInfo() (string, error)
    
    // Change analysis
    GetFileChangeHistory(days int) ([]FileChange, error)
    GetCommitHistory(days int) ([]CommitInfo, error)
    GetFileCoOccurrences(days int) (map[string][]string, error)
    GetChangeFrequency(days int) (map[string]int, error)
    GetLastModified() (map[string]time.Time, error)
    
    // Command execution
    ExecuteGitCommand(ctx context.Context, args ...string) ([]byte, error)
}

type PatternDetector interface {
    // Pattern detection
    DetectChangePatterns(days int) ([]ChangePattern, error)
    DetectFileRelationships(days int) ([]FileRelationship, error)
    DetectModuleGroups(days int) ([]ModuleGroup, error)
    
    // Configuration
    SetThresholds(minSupport, minConfidence float64)
}

type SemanticAnalyzer interface {
    // Main analysis
    AnalyzeRepository() (*SemanticAnalysisResult, error)
    GetContextRecommendationsForFile(filePath string) ([]ContextRecommendation, error)
}

type GraphIntegration interface {
    // Enhanced neighborhood building
    BuildEnhancedNeighborhoods() ([]EnhancedNeighborhood, error)
    BuildClusteredNeighborhoods() ([]ClusteredNeighborhood, error)
    
    // Clustering operations
    PerformHierarchicalClustering(neighborhoods []EnhancedNeighborhood) ([]Cluster, error)
    CalculateClusterQuality(clusters []Cluster) (*ClusterQuality, error)
    DetermineOptimalClusters(clusters []Cluster) (int, error)
}
```

### Git Integration Types
```go
type ChangePattern struct {
    Name           string
    Files          []string
    Frequency      int
    Confidence     float64
    LastOccurrence time.Time
    AvgInterval    time.Duration
    Metadata       map[string]string
}

type FileRelationship struct {
    File1       string
    File2       string
    Correlation float64
    Frequency   int
    Strength    string // "strong", "moderate", "weak"
}

type EnhancedNeighborhood struct {
    Name                 string
    Files                []string
    Dependencies         []string
    StructuralSimilarity float64
    GitPatternScore      float64
    CombinedScore        float64
    Metadata             map[string]interface{}
}

type ClusteredNeighborhood struct {
    ClusterID    int
    Neighborhoods []EnhancedNeighborhood
    Quality      ClusterQuality
    TaskType     string
    Description  string
}
```

## Enhanced Diff Engine Interfaces

### Core Diff Engine
```go
type DiffEngine interface {
    // File comparison
    CompareFiles(ctx context.Context, oldFile, newFile *types.FileInfo) (*DiffResult, error)
    CompareSymbols(ctx context.Context, oldSymbol, newSymbol *types.Symbol) (*DiffResult, error)
    
    // Configuration
    SetConfig(config *DiffConfig) error
    GetConfig() *DiffConfig
}

type DiffConfig struct {
    EnableSemanticDiff    bool
    EnableStructuralDiff  bool
    EnableRenameDetection bool
    EnableDepTracking     bool
    SimilarityThreshold   float64
    RenameThreshold       float64
    MaxDiffDepth          int
    Timeout               time.Duration
    EnableCaching         bool
    CacheTTL              time.Duration
}
```

### Similarity Algorithms
```go
type SimilarityAlgorithm interface {
    CalculateSimilarity(old, new *types.Symbol) SimilarityScore
    GetWeight() float64
    GetName() string
}

type SimilarityScore struct {
    Score      float64 `json:"score"`      // 0.0 to 1.0
    Confidence float64 `json:"confidence"` // 0.0 to 1.0
    Evidence   string  `json:"evidence"`   // Description of evidence
    Algorithm  string  `json:"algorithm"`  // Algorithm that produced this score
}

// Implemented algorithms
type NameSimilarity struct{}
type SignatureSimilarity struct{}
type StructuralSimilarity struct{}
type LocationSimilarity struct{}
type DocumentationSimilarity struct{}
type SemanticSimilarity struct{}
```

### Rename Detection
```go
type RenameDetector interface {
    DetectRenames(ctx context.Context, oldSymbols, newSymbols []*Symbol) ([]Rename, error)
    CalculateSimilarity(old, new *Symbol) float64
    ApplyHeuristics(old, new *Symbol, context *RenameContext) float64
}

type RenameHeuristic interface {
    EvaluateRename(old, new *Symbol, context *RenameContext) HeuristicScore
    GetWeight() float64
    GetName() string
}

// Implemented heuristics
type CamelCaseHeuristic struct{}
type PrefixSuffixHeuristic struct{}
type AbbreviationHeuristic struct{}
type RefactoringPatternHeuristic struct{}
type ContextualHeuristic struct{}
```

### Dependency Tracking
```go
type DependencyTracker interface {
    TrackDependencyChanges(ctx context.Context, oldFile, newFile *FileInfo) ([]Change, error)
}

type DependencyDetector interface {
    ExtractDependencies(file *FileInfo) ([]Dependency, error)
    ParseImportStatement(line string) (*Import, error)
    GetImportKeywords() []string
    IsRelativeImport(importPath string) bool
    NormalizeImportPath(importPath string) string
}

// Implemented detectors
type JavaScriptDetector struct{}
type TypeScriptDetector struct{}
type GoDetector struct{}
type PythonDetector struct{}
type JavaDetector struct{}
type CSharpDetector struct{}
type CppDetector struct{}
type CppParser struct{}
```

## MCP Server Interfaces

### MCP Server Core
```go
type MCPServer interface {
    // Server lifecycle
    Start(ctx context.Context) error
    Stop() error
    GetStatus() ServerStatus
    
    // Tool registration
    RegisterTool(tool MCPTool) error
    ListTools() []ToolInfo
    
    // Request handling
    HandleRequest(request MCPRequest) (MCPResponse, error)
}

type MCPTool interface {
    GetName() string
    GetDescription() string
    GetInputSchema() map[string]interface{}
    Execute(params map[string]interface{}) (interface{}, error)
}
```

### Implemented MCP Tools (v3.1.1 - 8 Tools Enhanced with C++ and Swift Support)
```go
// 1. Get codebase overview
type GetCodebaseOverviewTool struct {
    analyzer *analyzer.GraphBuilder
}

// 2. Get file analysis  
type GetFileAnalysisTool struct {
    analyzer *analyzer.GraphBuilder
}

// 3. Get symbol information (Enhanced with framework-specific insights + C++ and Swift support)
type GetSymbolInfoTool struct {
    analyzer *analyzer.GraphBuilder
    frameworkDetector *parser.FrameworkDetector
    swiftParser *parser.SwiftParser
    cppParser *parser.CppParser
}

// 4. Search symbols (Enhanced with framework-aware filtering + C++ and Swift support)
type SearchSymbolsTool struct {
    analyzer *analyzer.GraphBuilder
    frameworkDetector *parser.FrameworkDetector
    swiftParser *parser.SwiftParser
    cppParser *parser.CppParser
}

// 5. Get dependencies
type GetDependenciesTool struct {
    analyzer *analyzer.GraphBuilder
}

// 6. Watch changes
type WatchChangesTool struct {
    watcher *watcher.FileWatcher
}

// 7. Get semantic neighborhoods
type GetSemanticNeighborhoodsTool struct {
    gitIntegration *git.GraphIntegration
}

// 8. Get framework analysis (NEW in v2.4)
type GetFrameworkAnalysisTool struct {
    analyzer *analyzer.GraphBuilder
    frameworkDetector *parser.FrameworkDetector
}
```

### Enhanced MCP Tool Arguments (v2.4)
```go
// Enhanced get_symbol_info arguments with framework support
type GetSymbolInfoArgs struct {
    SymbolName    string `json:"symbol_name"`
    FilePath      string `json:"file_path,omitempty"`
    FrameworkType string `json:"framework_type,omitempty"` // NEW: Framework filtering
}

// Enhanced search_symbols arguments with framework filtering
type SearchSymbolsArgs struct {
    Query         string `json:"query"`
    Limit         int    `json:"limit,omitempty"`
    FileType      string `json:"file_type,omitempty"`
    SymbolType    string `json:"symbol_type,omitempty"`
    FrameworkType string `json:"framework_type,omitempty"` // NEW: Framework filtering
}

// NEW: Framework analysis arguments
type GetFrameworkAnalysisArgs struct {
    Framework    string `json:"framework,omitempty"`    // Filter by specific framework
    IncludeStats bool   `json:"include_stats,omitempty"` // Include detailed statistics
}
```

### MCP Protocol Types
```go
type MCPRequest struct {
    Method string                 `json:"method"`
    Params map[string]interface{} `json:"params"`
    ID     string                 `json:"id"`
}

type MCPResponse struct {
    Result interface{} `json:"result,omitempty"`
    Error  *MCPError   `json:"error,omitempty"`
    ID     string      `json:"id"`
}

type ToolInfo struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"inputSchema"`
}
```

## Virtual Graph Interfaces

### Core Virtual Graph Engine
```go
type VirtualGraphEngine interface {
    // State management
    GetShadowGraph() *CodeGraph
    GetActualGraph() *CodeGraph
    GetPendingChanges() []ChangeSet
    
    // Core operations
    Diff(oldAST AST, newAST AST) (*ASTDiff, error)
    BatchChange(change Change) error
    Reconcile() (*ReconciliationPlan, error)
    Commit(plan *ReconciliationPlan) (*CodeGraph, error)
    Rollback(checkpoint *GraphCheckpoint) error
    
    // Optimization
    ShouldBatch(change Change) bool
    OptimizePlan(plan *ReconciliationPlan) (*OptimizedPlan, error)
    
    // Metrics and monitoring
    GetChangeMetrics() *ChangeMetrics
    ValidateIntegrity() []ValidationError
    GetMemoryUsage() MemoryStats
}

type ASTDiffer interface {
    // Core diff operations
    StructuralDiff(oldAST AST, newAST AST) (*StructuralDiff, error)
    SemanticDiff(oldAST AST, newAST AST) (*SemanticDiff, error)
    
    // Symbol tracking
    TrackSymbolChanges(diff *StructuralDiff) (*SymbolChangeSet, error)
    ComputeImpact(changes *SymbolChangeSet) (*ImpactGraph, error)
    
    // Configuration
    SetDiffAlgorithm(algorithm string) error
    SetMaxDepth(depth int)
    EnableMemoization(enabled bool)
}

type Reconciler interface {
    // Plan generation
    GeneratePlan(changes []ChangeSet) (*ReconciliationPlan, error)
    OptimizePlan(plan *ReconciliationPlan) (*OptimizedPlan, error)
    ValidatePlan(plan *ReconciliationPlan) []ValidationError
    
    // Plan execution
    ExecutePlan(plan *ReconciliationPlan) (*ExecutionResult, error)
    PreviewPlan(plan *ReconciliationPlan) (*PlanPreview, error)
    
    // Rollback support
    CreateCheckpoint() (*GraphCheckpoint, error)
    RestoreCheckpoint(checkpoint *GraphCheckpoint) error
}
```

### Change Management
```go
type Change struct {
    Id        string
    Type      ChangeType    // Add, Remove, Modify, Move
    Target    interface{}   // NodeId, FileLocation, etc.
    Data      interface{}   // Change-specific data
    Metadata  map[string]interface{}
    Timestamp time.Time
    Source    ChangeSource  // FileSystem, UserEdit, Refactor
}

type ChangeSet struct {
    Id        string
    Changes   []Change
    Timestamp time.Time
    Source    string
    BatchSize int
    Metadata  map[string]interface{}
}

type ChangeSource string
const (
    ChangeSourceFileSystem ChangeSource = "filesystem"
    ChangeSourceUserEdit   ChangeSource = "user_edit"
    ChangeSourceRefactor   ChangeSource = "refactor"
    ChangeSourceImport     ChangeSource = "import"
)
```

## Parser Interfaces

### Parser Manager
```go
type ParserManager interface {
    // Core parsing
    ParseFile(filePath string, language Language) (*AST, error)
    ParseFileVersioned(filePath, content, version string) (*VersionedAST, error)
    ParseContent(content string, language Language) (*AST, error)
    
    // Symbol extraction
    ExtractSymbols(ast *AST) ([]*Symbol, error)
    ExtractImports(ast *AST) ([]*Import, error)
    ResolveImportAlias(importPath, fromFile string) (string, error)
    
    // Language support
    GetSupportedLanguages() []Language
    ClassifyFile(filePath string) (*FileClassification, error)
    DetectLanguage(filePath string) (*Language, error)
    
    // Cache management
    GetASTCache() ASTCache
    InvalidateCache(fileId string) error
    ClearCache() error
}

type LanguageParser interface {
    // Basic parsing
    Parse(content string) (*AST, error)
    ParseWithOptions(content string, options ParseOptions) (*AST, error)
    
    // Language-specific features
    ExtractSymbols(ast *AST) ([]*Symbol, error)
    ExtractComments(ast *AST) ([]*Comment, error)
    ExtractImports(ast *AST) ([]*Import, error)
    
    // Validation
    Validate(ast *AST) []ValidationError
    GetParseErrors() []ParseError
    
    // Configuration
    SetQueryPath(path string) error
    LoadQueries() error
    GetLanguageInfo() LanguageInfo
}

type ASTCache interface {
    // Basic cache operations
    Get(fileId string, version ...string) (*VersionedAST, error)
    Set(fileId string, ast *VersionedAST) error
    Invalidate(fileId string) error
    Clear() error
    
    // Diff cache
    GetDiffCache(fileId string) ([]*ASTDiff, error)
    SetDiffCache(fileId string, diffs []*ASTDiff) error
    
    // Cache management
    Size() int
    Stats() CacheStats
    SetMaxSize(size int)
    SetTTL(ttl time.Duration)
}
```

### Language Support
```go
type Language struct {
    Name         string
    Extensions   []string
    Parser       string
    TreeSitter   string
    QueryPath    string
    Enabled      bool
    Features     LanguageFeatures
    Configuration map[string]interface{}
}

type LanguageFeatures struct {
    SupportsClasses     bool
    SupportsInterfaces  bool
    SupportsGenerics    bool
    SupportsModules     bool
    SupportsDecorators  bool
    SupportsAsync       bool
    TypeSystem          TypeSystemLevel
}

type TypeSystemLevel string
const (
    TypeSystemNone     TypeSystemLevel = "none"
    TypeSystemDynamic  TypeSystemLevel = "dynamic"
    TypeSystemStatic   TypeSystemLevel = "static"
    TypeSystemGradual  TypeSystemLevel = "gradual"
)
```

## Compact Controller Interfaces

### Core Compaction
```go
type CompactController interface {
    // Command execution
    ExecuteCommand(command string, context *CompactContext) (*CompactResult, error)
    ParseCommand(command string) (*CompactCommand, error)
    
    // Predefined strategies
    CompactMinimal() (*CompactResult, error)
    CompactBalanced() (*CompactResult, error)
    CompactAggressive() (*CompactResult, error)
    
    // Task-specific compaction
    CompactForTask(task TaskType) (*CompactResult, error)
    CompactToTokenLimit(maxTokens int) (*CompactResult, error)
    CompactWithFocus(focusFiles []string) (*CompactResult, error)
    
    // Interactive features
    PreviewCompaction(strategy CompactStrategy) (*CompactPreview, error)
    UndoCompaction() error
    GetCompactionHistory() ([]*CompactHistory, error)
    
    // Strategy management
    SaveStrategy(name string, strategy CompactStrategy) error
    LoadStrategy(name string) (*CompactStrategy, error)
    ListStrategies() ([]string, error)
}

type CompactStrategy interface {
    // Core methods
    Apply(graph *CodeGraph) (*CodeGraph, error)
    Preview(graph *CodeGraph) (*CompactPreview, error)
    CalculateQuality(original, compacted *CodeGraph) (*QualityScore, error)
    
    // Configuration
    GetName() string
    GetDescription() string
    GetTokenTarget() float64
    GetRules() []CompactRule
    
    // Validation
    Validate() []ValidationError
    IsReversible() bool
}

type QualityAnalyzer interface {
    // Quality assessment
    CalculateQuality(original, compacted *CodeGraph) (*QualityScore, error)
    AnalyzeSymbolCoverage(original, compacted *CodeGraph) float64
    AnalyzeRelationshipPreservation(original, compacted *CodeGraph) float64
    AnalyzeContextCoherence(original, compacted *CodeGraph) float64
    
    // Validation
    ValidateQuality(score *QualityScore, threshold float64) []QualityWarning
    SuggestImprovements(score *QualityScore) []QualityImprovement
}
```

### Task Management
```go
type TaskType interface {
    GetName() string
    GetDescription() string
    GetPreservePatterns() []SymbolPattern
    GetRemovePatterns() []SymbolPattern
    GetPriorityWeights() map[string]float64
    
    ShouldPreserve(symbol *Symbol) bool
    ShouldRemove(symbol *Symbol) bool
    CalculatePriority(symbol *Symbol) float64
}

type SymbolPattern interface {
    Match(symbol *Symbol) bool
    GetPattern() string
    GetType() PatternType
    GetPriority() int
}

type PatternType string
const (
    PatternTypeExact  PatternType = "exact"
    PatternTypeRegex  PatternType = "regex"
    PatternTypeGlob   PatternType = "glob"
    PatternTypeCustom PatternType = "custom"
)
```

## Storage Interfaces

### Graph Storage
```go
type GraphStorage interface {
    // Graph operations
    SaveGraph(graph *CodeGraph) error
    LoadGraph(id string) (*CodeGraph, error)
    DeleteGraph(id string) error
    ListGraphs() ([]GraphMetadata, error)
    
    // Version management
    SaveVersion(graphId string, version *GraphVersion) error
    LoadVersion(graphId, versionId string) (*CodeGraph, error)
    ListVersions(graphId string) ([]GraphVersion, error)
    
    // Patch management
    SavePatch(patch *GraphPatch) error
    LoadPatches(graphId string, since time.Time) ([]GraphPatch, error)
    ApplyPatches(graphId string, patches []GraphPatch) error
}

type CacheStorage interface {
    // Cache operations
    Get(key string) ([]byte, error)
    Set(key string, value []byte, ttl time.Duration) error
    Delete(key string) error
    Clear() error
    
    // Batch operations
    GetMulti(keys []string) (map[string][]byte, error)
    SetMulti(items map[string][]byte, ttl time.Duration) error
    DeleteMulti(keys []string) error
    
    // Cache management
    Size() int
    Stats() CacheStats
    Cleanup() error
}

type ConfigStorage interface {
    // Configuration management
    LoadConfig(path string) (*Config, error)
    SaveConfig(path string, config *Config) error
    ValidateConfig(config *Config) []ValidationError
    
    // Defaults and templates
    GetDefaultConfig() *Config
    GetConfigTemplate(template string) (*Config, error)
    ListConfigTemplates() ([]string, error)
}
```

### Metrics and Monitoring
```go
type MetricsCollector interface {
    // Counter metrics
    IncrementCounter(name string, tags map[string]string)
    AddCounter(name string, value float64, tags map[string]string)
    
    // Gauge metrics
    SetGauge(name string, value float64, tags map[string]string)
    
    // Histogram metrics
    RecordHistogram(name string, value float64, tags map[string]string)
    RecordTiming(name string, duration time.Duration, tags map[string]string)
    
    // Custom metrics
    RecordCustom(metric CustomMetric)
    
    // Export
    Export() ([]byte, error)
    ExportPrometheus() ([]byte, error)
    ExportJSON() ([]byte, error)
}

type HealthChecker interface {
    // Health checks
    CheckHealth() (*HealthStatus, error)
    CheckComponent(component string) (*ComponentHealth, error)
    RegisterHealthCheck(name string, check HealthCheckFunc)
    
    // Monitoring
    GetHealthHistory() ([]HealthSnapshot, error)
    GetAlerts() ([]HealthAlert, error)
    SetHealthThreshold(component string, threshold HealthThreshold)
}
```

## REST API Specifications

### Core Endpoints
```yaml
openapi: 3.0.3
info:
  title: CodeContext API
  version: 2.0.0

paths:
  /api/v1/projects:
    post:
      summary: Create a new project
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ProjectConfig'
      responses:
        '201':
          description: Project created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Project'

  /api/v1/projects/{projectId}/generate:
    post:
      summary: Generate context map
      parameters:
        - name: projectId
          in: path
          required: true
          schema:
            type: string
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/GenerateRequest'
      responses:
        '200':
          description: Map generated successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/MapResult'

  /api/v1/projects/{projectId}/update:
    post:
      summary: Update context map incrementally
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateRequest'
      responses:
        '200':
          description: Map updated successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PatchResult'

  /api/v1/projects/{projectId}/compact:
    post:
      summary: Execute compact command
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CompactRequest'
      responses:
        '200':
          description: Compaction executed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CompactResult'

components:
  schemas:
    ProjectConfig:
      type: object
      properties:
        name:
          type: string
        path:
          type: string
        languages:
          type: array
          items:
            $ref: '#/components/schemas/Language'
        excludePatterns:
          type: array
          items:
            type: string

    CompactRequest:
      type: object
      properties:
        level:
          type: string
          enum: [minimal, balanced, aggressive]
        task:
          type: string
          enum: [debugging, refactoring, documentation, review, testing]
        maxTokens:
          type: integer
        preview:
          type: boolean
        focusFiles:
          type: array
          items:
            type: string
```

## GraphQL Schema

### Core Schema
```graphql
type Query {
  # Project queries
  project(id: ID!): Project
  projects(filter: ProjectFilter): [Project!]!
  
  # Graph queries
  graph(projectId: ID!): CodeGraph
  symbols(projectId: ID!, filter: SymbolFilter): [Symbol!]!
  
  # Search
  searchSymbols(query: String!, limit: Int): [Symbol!]!
  searchFiles(query: String!, limit: Int): [File!]!
  
  # Analytics
  getMetrics(projectId: ID!, timeRange: TimeRange): Metrics
  getHealth(): HealthStatus
}

type Mutation {
  # Project operations
  createProject(config: ProjectConfigInput!): Project!
  updateProject(id: ID!, config: ProjectConfigInput!): Project!
  deleteProject(id: ID!): Boolean!
  
  # Map operations
  generateMap(projectId: ID!, config: GenerateConfigInput): MapResult!
  updateMap(projectId: ID!, changes: [FileChangeInput!]!): PatchResult!
  
  # Compact operations
  compactMap(projectId: ID!, command: CompactCommandInput!): CompactResult!
  undoCompaction(projectId: ID!): Boolean!
  
  # Configuration
  saveStrategy(projectId: ID!, strategy: CompactStrategyInput!): CompactStrategy!
  deleteStrategy(projectId: ID!, name: String!): Boolean!
}

type Subscription {
  # Real-time updates
  mapUpdates(projectId: ID!): MapUpdate!
  processingProgress(operationId: ID!): Progress!
  healthChanges: HealthStatus!
}

# Core types
type Project {
  id: ID!
  name: String!
  path: String!
  config: ProjectConfig!
  createdAt: DateTime!
  updatedAt: DateTime!
  status: ProjectStatus!
}

type CodeGraph {
  nodes: [GraphNode!]!
  edges: [GraphEdge!]!
  metadata: GraphMetadata!
  version: GraphVersion!
}

type Symbol {
  id: ID!
  name: String!
  type: SymbolType!
  location: FileLocation!
  signature: String
  documentation: String
  language: String!
  importance: Float!
}
```

## CLI Interface

### Command Structure
```bash
codecontext [global-flags] <command> [command-flags] [arguments]

Global Flags:
  --config, -c     Config file path
  --verbose, -v    Verbose output
  --output, -o     Output file
  --help, -h       Help
  --version        Version

Commands:
  init             Initialize project
  generate         Generate context map
  update           Update incrementally
  compact          Optimize context
  config           Manage configuration
  cache            Cache operations
  graph            Graph operations
  metrics          Show metrics
  health           Health check
```

### Command Interfaces
```go
type CLICommand interface {
    // Command metadata
    Name() string
    Description() string
    Usage() string
    Examples() []string
    
    // Flag management
    DefineFlags() []Flag
    ValidateFlags(flags map[string]interface{}) error
    
    // Execution
    Execute(ctx *CLIContext) error
    
    // Help and completion
    GetHelp() string
    GetCompletion(partial string) []string
}

type CLIContext struct {
    Command     string
    Flags       map[string]interface{}
    Arguments   []string
    Config      *Config
    Output      io.Writer
    Input       io.Reader
    Logger      Logger
    Metrics     MetricsCollector
}

type Flag struct {
    Name        string
    Short       string
    Type        FlagType
    Required    bool
    Default     interface{}
    Description string
    Validation  func(interface{}) error
}
```

---

*This API specification serves as the definitive reference for all interfaces in the CodeContext system and should be kept synchronized with the implementation.*
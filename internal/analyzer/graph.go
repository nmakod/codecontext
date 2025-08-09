package analyzer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/nuthan-ms/codecontext/internal/cache"
	"github.com/nuthan-ms/codecontext/internal/git"
	"github.com/nuthan-ms/codecontext/internal/parser"
	"github.com/nuthan-ms/codecontext/pkg/types"
)

// Constants for configuration
const (
	DefaultProgressInterval = 10
	MinProgressInterval     = 1
	MaxCachedPatterns       = 1000 // Prevent memory leaks from excessive caching
	MaxNormalizationCache   = 1000 // Maximum entries in normalization caches
)

// Memory pools for hot path allocations to reduce GC pressure
var (
	// Pool for string slices used in path component splitting
	stringSlicePool = sync.Pool{
		New: func() interface{} {
			// Pre-allocate with capacity for typical path depth (8 components)
			// This covers most real-world paths efficiently
			return make([]string, 0, 8)
		},
	}
)

// getStringSlice retrieves a string slice from the pool for hot path allocations
func getStringSlice() []string {
	slice := stringSlicePool.Get().([]string)
	return slice[:0] // Reset length but keep capacity
}

// putStringSlice returns a string slice to the pool for reuse
func putStringSlice(slice []string) {
	if cap(slice) <= 32 { // Only pool reasonably sized slices to prevent memory bloat
		stringSlicePool.Put(slice)
	}
}

// deduplicate removes duplicate patterns from a slice
// Empty patterns are automatically filtered out
func deduplicate(patterns []string) []string {
	if len(patterns) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(patterns)) // struct{} uses 0 bytes per entry
	result := make([]string, 0, len(patterns))

	for _, pattern := range patterns {
		if pattern == "" { // Skip empty patterns early
			continue
		}
		if _, exists := seen[pattern]; !exists {
			seen[pattern] = struct{}{}
			result = append(result, pattern)
		}
	}

	return result
}

// Default exclude patterns with lazy initialization for thread safety
var (
	defaultPatterns     []string
	defaultPatternsOnce sync.Once
)

// getDefaultExcludePatterns returns the default exclude patterns with lazy initialization
// This ensures thread-safe access and avoids package initialization order issues
func getDefaultExcludePatterns() []string {
	defaultPatternsOnce.Do(func() {
		defaultPatterns = deduplicate([]string{
			// JavaScript/TypeScript
			"node_modules/**",
			".next/**",
			".nuxt/**",
			".cache/**",
			".parcel-cache/**",
			"bower_components/**",

			// Python
			"__pycache__/**",
			"*.py[cod]",
			"*$py.class",
			".venv/**",
			"venv/**",
			"env/**",
			".Python",
			".pytest_cache/**",
			".mypy_cache/**",
			"*.egg-info/**",
			".tox/**",
			".eggs/**",
			"htmlcov/**",
			".hypothesis/**",
			".coverage",
			"*.cover",
			".coverage.*",

			// Java/Kotlin/Scala
			"target/**",
			".gradle/**",
			"*.class",
			"*.jar",

			// Go
			"vendor/**",

			// Rust
			"target/**",

			// Ruby
			".bundle/**",
			"vendor/bundle/**",

			// PHP
			"vendor/**",
			".phpunit.cache/**",

			// .NET
			"bin/**",
			"obj/**",
			"packages/**",
			".vs/**",
			"*.dll",
			"*.exe",

			// Build outputs
			"dist/**",
			"build/**",
			"out/**",
			"_build/**",

			// Testing
			"coverage/**",
			".nyc_output/**",
			"test-results/**",
			"jest-cache/**",

			// Version control
			".git/**",
			".svn/**",
			".hg/**",
			".bzr/**",

			// IDE
			".idea/**",
			".vscode/**",
			"*.swp",
			"*.swo",
			"*~",
			".project",
			".classpath",
			".settings/**",

			// OS
			".DS_Store",
			"Thumbs.db",
			"desktop.ini",

			// Logs and temp
			"*.log",
			"logs/**",
			"tmp/**",
			"temp/**",
			"*.tmp",
			"*.temp",
			"*.bak",
			"*.backup",
			"*.old",

			// Package managers
			"npm-debug.log*",
			"yarn-debug.log*",
			"yarn-error.log*",
			"pnpm-debug.log*",
			".pnpm-store/**",

			// CI/CD
			".terraform/**",
			".serverless/**",
			".github/workflows/**",
			".gitlab/**",

			// Documentation builds
			"_site/**",
			".docusaurus/**",
			"site/**",

			// Mobile
			".expo/**",
			".expo-shared/**",

			// Certificates and secrets (safety)
			"*.pem",
			"*.key",
			"*.cert",
			"*.crt",
			".env.local",
			".env.*.local",
		})
	})
	return defaultPatterns
}

// GraphBuilder builds code graphs from parsed files
// ProgressConfig configures progress reporting behavior
type ProgressConfig struct {
	Interval       int  // Update progress every N files (default: 10)
	ShowPercentage bool // Show percentage progress if total count is known
}

type GraphBuilder struct {
	parser             *parser.Manager
	graph              *types.CodeGraph
	cache              *cache.PersistentCache
	progressCallback   func(string)
	progressConfig     ProgressConfig
	excludePatterns    []string
	includePatterns    []string // Negation patterns (starting with !)
	useDefaultExcludes bool

	// Thread-safe pattern caching
	patternMu      sync.RWMutex
	cachedPatterns []string // Cached merged patterns to avoid repeated allocation
	patternsDirty  bool     // Whether cached patterns need to be regenerated

	// Path normalization cache to avoid redundant operations
	normCacheMu     sync.RWMutex
	normalizeCache  map[string]string // Cache for normalizePath results
	patternCache    map[string]string // Cache for normalizeForPattern results

	// Error handling
	logger *log.Logger // Optional logger for pattern errors
}

// NewGraphBuilder creates a new graph builder
func NewGraphBuilder() *GraphBuilder {
	return &GraphBuilder{
		parser: parser.NewManager(),
		graph: &types.CodeGraph{
			Nodes:    make(map[types.NodeId]*types.GraphNode),
			Edges:    make(map[types.EdgeId]*types.GraphEdge),
			Files:    make(map[string]*types.FileNode),
			Symbols:  make(map[types.SymbolId]*types.Symbol),
			Metadata: &types.GraphMetadata{},
		},
		progressConfig: ProgressConfig{
			Interval:       DefaultProgressInterval,
			ShowPercentage: false, // Default: don't show percentage (requires pre-counting)
		},
		useDefaultExcludes: true, // Use default exclude patterns by default
		excludePatterns:    []string{},
		includePatterns:    []string{},
		patternsDirty:      true, // Force initial cache build
		
		// Initialize normalization caches with reasonable initial capacity
		normalizeCache:  make(map[string]string, 256),
		patternCache:    make(map[string]string, 256),
	}
}

// SetLogger sets a logger for pattern error reporting
func (gb *GraphBuilder) SetLogger(logger *log.Logger) {
	gb.logger = logger
}

// SetCache sets the persistent cache for the graph builder
func (gb *GraphBuilder) SetCache(c *cache.PersistentCache) {
	gb.cache = c
}

// Path normalization helpers for cross-platform compatibility and security

// normalizePath ensures consistent path format across platforms
func (gb *GraphBuilder) normalizePath(path string) string {
	// Check cache first (fast path with read lock)
	gb.normCacheMu.RLock()
	if cached, exists := gb.normalizeCache[path]; exists {
		gb.normCacheMu.RUnlock()
		return cached
	}
	gb.normCacheMu.RUnlock()
	
	// Clean the path to remove redundant elements like "." and ".."
	normalized := filepath.Clean(path)
	
	// Cache the result (write lock)
	gb.normCacheMu.Lock()
	// Check cache size to prevent memory leaks
	if len(gb.normalizeCache) < MaxNormalizationCache {
		gb.normalizeCache[path] = normalized
	}
	gb.normCacheMu.Unlock()
	
	return normalized
}

// normalizeForPattern converts path to forward slashes for consistent pattern matching
// This is essential for cross-platform glob pattern matching
func (gb *GraphBuilder) normalizeForPattern(path string) string {
	// Check cache first (fast path with read lock)
	gb.normCacheMu.RLock()
	if cached, exists := gb.patternCache[path]; exists {
		gb.normCacheMu.RUnlock()
		return cached
	}
	gb.normCacheMu.RUnlock()
	
	var normalized string
	
	// Handle UNC paths specially to preserve the double slash prefix
	if strings.HasPrefix(path, "\\\\") {
		// UNC path: \\server\share -> //server/share
		// Use strings.Builder for efficient string construction
		var builder strings.Builder
		builder.WriteString("//")
		builder.WriteString(strings.TrimPrefix(path, "\\\\"))
		unc := builder.String()
		
		// Replace remaining backslashes with forward slashes
		unc = strings.ReplaceAll(unc, "\\", "/")
		// Clean the path but preserve the UNC prefix
		cleaned := filepath.Clean(unc)
		
		// filepath.Clean might convert // to /, so restore it
		if !strings.HasPrefix(cleaned, "//") {
			builder.Reset()
			builder.WriteString("//")
			builder.WriteString(strings.TrimPrefix(cleaned, "/"))
			normalized = builder.String()
		} else {
			normalized = cleaned
		}
	} else {
		// First normalize backslashes to forward slashes for cross-platform consistency
		temp := strings.ReplaceAll(path, "\\", "/")
		// Then clean the path and convert to forward slashes
		normalized = filepath.ToSlash(filepath.Clean(temp))
	}
	
	// Cache the result (write lock)
	gb.normCacheMu.Lock()
	// Check cache size to prevent memory leaks
	if len(gb.patternCache) < MaxNormalizationCache {
		gb.patternCache[path] = normalized
	}
	gb.normCacheMu.Unlock()
	
	return normalized
}

// validateImportPath ensures import paths don't escape the project directory
// This prevents directory traversal attacks when resolving relative imports
func (gb *GraphBuilder) validateImportPath(importPath, baseDir string) error {
	cleaned := filepath.Clean(importPath)
	
	// Check for actual directory traversal attempts (not just files with dots)
	// We need to look for "../" patterns or standalone ".." components
	hasTraversal := false
	
	// Split path into components to check for actual ".." directory references
	// Use pooled slice to reduce allocations in hot path
	components := getStringSlice()
	components = append(components, strings.Split(strings.ReplaceAll(cleaned, "\\", "/"), "/")...)
	
	for _, component := range components {
		if component == ".." {
			hasTraversal = true
			break
		}
	}
	
	putStringSlice(components)
	
	if hasTraversal {
		// Resolve to absolute path and verify it's within project boundaries
		// We need to find the project root, not just the current file's directory
		abs := filepath.Join(baseDir, cleaned)
		abs = filepath.Clean(abs)
		
		// Get absolute base directory
		baseDirAbs, err := filepath.Abs(baseDir)
		if err != nil {
			baseDirAbs = filepath.Clean(baseDir)
		}
		
		// Get absolute resolved path
		resolvedAbs, err := filepath.Abs(abs)
		if err != nil {
			resolvedAbs = abs
		}
		
		// For import paths, we should allow going up to sibling directories
		// but not beyond reasonable project boundaries
		_, err = filepath.Rel(baseDirAbs, resolvedAbs)
		if err != nil {
			return fmt.Errorf("cannot determine relative path for import: %s", importPath)
		}
		
		// Count upward levels in the original import path
		// Handle both forward and back slashes
		normalizedPath := strings.ReplaceAll(cleaned, "\\", "/")
		upwardLevels := strings.Count(normalizedPath, "../")
		
		// Also count standalone ".." at the end
		if strings.HasSuffix(normalizedPath, "/..") || normalizedPath == ".." {
			upwardLevels++
		}
		
		// Allow reasonable traversal (max 2 levels up) but block obvious attacks
		if upwardLevels > 2 {
			return fmt.Errorf("import path escapes project directory: %s", importPath)
		}
		
		// Additional check: if resolved path contains suspicious system paths, block it
		if strings.Contains(resolvedAbs, "/etc/") || 
		   strings.Contains(resolvedAbs, "/bin/") ||
		   strings.Contains(resolvedAbs, "/sbin/") ||
		   strings.HasSuffix(resolvedAbs, "/etc/passwd") ||
		   strings.HasSuffix(resolvedAbs, "/bin/sh") {
			return fmt.Errorf("import path escapes project directory: %s", importPath)
		}
	}
	
	return nil
}

// SetExcludePatterns sets the exclude patterns for the graph builder
// Patterns starting with ! are treated as include patterns (negations)
func (gb *GraphBuilder) SetExcludePatterns(patterns []string) {
	gb.patternMu.Lock()
	defer gb.patternMu.Unlock()

	gb.excludePatterns = []string{}
	gb.includePatterns = []string{}

	for _, pattern := range patterns {
		if strings.HasPrefix(pattern, "!") {
			// Remove the ! prefix and add to include patterns
			trimmed, _ := strings.CutPrefix(pattern, "!")
			gb.includePatterns = append(gb.includePatterns, trimmed)
		} else {
			gb.excludePatterns = append(gb.excludePatterns, pattern)
		}
	}

	gb.patternsDirty = true // Mark patterns as dirty to force cache rebuild
	
	// Clear normalization caches since patterns have changed
	gb.clearNormalizationCaches()
}

// clearNormalizationCaches clears the path normalization caches
func (gb *GraphBuilder) clearNormalizationCaches() {
	gb.normCacheMu.Lock()
	defer gb.normCacheMu.Unlock()
	
	// Clear both caches to ensure fresh normalization
	gb.normalizeCache = make(map[string]string, 256)
	gb.patternCache = make(map[string]string, 256)
}

// SetUseDefaultExcludes sets whether to use default exclude patterns
func (gb *GraphBuilder) SetUseDefaultExcludes(use bool) {
	gb.patternMu.Lock()
	defer gb.patternMu.Unlock()

	if gb.useDefaultExcludes != use {
		gb.useDefaultExcludes = use
		gb.patternsDirty = true // Mark patterns as dirty since defaults changed
		gb.clearNormalizationCaches() // Clear caches when default patterns change
	}
}

// SetProgressCallback sets a callback function for progress updates
func (gb *GraphBuilder) SetProgressCallback(callback func(string)) {
	gb.progressCallback = callback
}

// SetProgressInterval sets how often progress updates are sent (every N files)
func (gb *GraphBuilder) SetProgressInterval(interval int) {
	if interval >= MinProgressInterval {
		gb.progressConfig.Interval = interval
	}
}

// SetProgressConfig sets the complete progress configuration
func (gb *GraphBuilder) SetProgressConfig(config ProgressConfig) {
	if config.Interval >= MinProgressInterval {
		gb.progressConfig = config
	}
}

// AnalyzeDirectory analyzes a directory and builds a complete code graph
func (gb *GraphBuilder) AnalyzeDirectory(targetDir string) (*types.CodeGraph, error) {
	start := time.Now()

	// Initialize graph metadata
	gb.graph.Metadata = &types.GraphMetadata{
		Generated:    time.Now(),
		Version:      "2.0.0",
		TotalFiles:   0,
		TotalSymbols: 0,
		Languages:    make(map[string]int),
	}

	// Walk directory and process files
	fileCount := 0
	err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Normalize path immediately for consistent handling
		path = gb.normalizePath(path)

		// Skip directories and unsupported files
		if info.IsDir() || !gb.isSupportedFile(path) {
			return nil
		}

		// Skip certain directories
		// Convert to relative path for better pattern matching
		relPath, err := filepath.Rel(targetDir, path)
		if err != nil {
			relPath = path // fallback to absolute path
		}
		relPath = gb.normalizePath(relPath)
		
		if gb.shouldSkipPath(relPath) || gb.shouldSkipPath(path) {
			return nil
		}

		fileCount++

		// Update progress at configured intervals for staged display
		if gb.progressCallback != nil && fileCount%gb.progressConfig.Interval == 0 {
			gb.progressCallback(fmt.Sprintf("📄 Parsing files... (%d files)", fileCount))
		}

		return gb.processFile(path)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to analyze directory: %w", err)
	}

	// Show completion of parsing stage
	if gb.progressCallback != nil {
		gb.progressCallback(fmt.Sprintf("✅ Parsing complete (%d files)", fileCount))
	}

	// Build relationships between files
	if gb.progressCallback != nil {
		gb.progressCallback("🔗 Building relationships...")
	}
	gb.buildFileRelationships()

	if gb.progressCallback != nil {
		gb.progressCallback("✅ Relationships built")
	}

	// Build semantic neighborhoods if git repository
	if gb.progressCallback != nil {
		gb.progressCallback("📊 Analyzing git history...")
	}
	semanticResult, err := gb.buildSemanticNeighborhoods(targetDir)
	if err == nil && semanticResult != nil {
		// Add semantic analysis results to metadata
		if gb.graph.Metadata.Configuration == nil {
			gb.graph.Metadata.Configuration = make(map[string]interface{})
		}
		gb.graph.Metadata.Configuration["semantic_neighborhoods"] = semanticResult

		if gb.progressCallback != nil {
			gb.progressCallback("✅ Git analysis complete")
		}
	} else if gb.progressCallback != nil {
		gb.progressCallback("⚠️ Git analysis skipped")
	}

	// Update metadata
	gb.graph.Metadata.TotalFiles = len(gb.graph.Files)
	gb.graph.Metadata.TotalSymbols = len(gb.graph.Symbols)
	gb.graph.Metadata.AnalysisTime = time.Since(start)

	return gb.graph, nil
}

// processFile processes a single file and extracts symbols
func (gb *GraphBuilder) processFile(filePath string) error {
	// Normalize path before any processing to ensure consistency
	filePath = gb.normalizePath(filePath)
	
	// Detect language
	classification, err := gb.parser.ClassifyFile(filePath)
	if err != nil {
		// Skip files we can't classify
		return nil
	}

	// Parse the file
	ast, err := gb.parser.ParseFile(filePath, classification.Language)
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	// Extract symbols
	symbols, err := gb.parser.ExtractSymbols(ast)
	if err != nil {
		return fmt.Errorf("failed to extract symbols from %s: %w", filePath, err)
	}

	// Extract imports
	imports, err := gb.parser.ExtractImports(ast)
	if err != nil {
		return fmt.Errorf("failed to extract imports from %s: %w", filePath, err)
	}

	// Create file node
	fileNode := &types.FileNode{
		Path:         filePath,
		Language:     classification.Language.Name,
		Size:         len(ast.Content),
		Lines:        strings.Count(ast.Content, "\n") + 1,
		SymbolCount:  len(symbols),
		ImportCount:  len(imports),
		IsTest:       classification.IsTest,
		IsGenerated:  classification.IsGenerated,
		LastModified: time.Now(),
		Symbols:      make([]types.SymbolId, 0, len(symbols)),
		Imports:      imports,
	}

	// Add symbols to graph and file
	for _, symbol := range symbols {
		gb.graph.Symbols[symbol.Id] = symbol
		fileNode.Symbols = append(fileNode.Symbols, symbol.Id)

		// Create symbol node
		symbolNode := &types.GraphNode{
			Id:       types.NodeId(fmt.Sprintf("symbol-%s", symbol.Id)),
			Type:     "symbol",
			Label:    symbol.Name,
			FilePath: filePath,
			Metadata: map[string]interface{}{
				"symbolType": symbol.Type,
				"language":   symbol.Language,
				"signature":  symbol.Signature,
				"line":       symbol.Location.StartLine,
			},
		}
		gb.graph.Nodes[symbolNode.Id] = symbolNode
	}

	// Add file to graph
	gb.graph.Files[filePath] = fileNode

	// Update language statistics
	if gb.graph.Metadata.Languages == nil {
		gb.graph.Metadata.Languages = make(map[string]int)
	}
	gb.graph.Metadata.Languages[classification.Language.Name]++

	return nil
}

// buildFileRelationships analyzes imports to build file-to-file relationships
func (gb *GraphBuilder) buildFileRelationships() {
	// Use the enhanced relationship analyzer
	analyzer := NewRelationshipAnalyzer(gb.graph)

	// Perform comprehensive relationship analysis
	metrics, err := analyzer.AnalyzeAllRelationships()
	if err != nil {
		// Fall back to basic relationship building if analysis fails
		gb.buildBasicFileRelationships()
		return
	}

	// Store relationship metrics in graph metadata
	if gb.graph.Metadata.Configuration == nil {
		gb.graph.Metadata.Configuration = make(map[string]interface{})
	}
	gb.graph.Metadata.Configuration["relationship_metrics"] = metrics
}

// buildBasicFileRelationships provides fallback basic relationship building
func (gb *GraphBuilder) buildBasicFileRelationships() {
	for filePath, fileNode := range gb.graph.Files {
		for _, imp := range fileNode.Imports {
			targetFile := gb.resolveImportPath(imp.Path, filePath)
			if targetFile != "" && gb.graph.Files[targetFile] != nil {
				// Create edge for file dependency
				edgeId := types.EdgeId(fmt.Sprintf("import-%s-%s", filePath, targetFile))
				edge := &types.GraphEdge{
					Id:     edgeId,
					From:   types.NodeId(fmt.Sprintf("file-%s", filePath)),
					To:     types.NodeId(fmt.Sprintf("file-%s", targetFile)),
					Type:   "imports",
					Weight: 1.0,
					Metadata: map[string]interface{}{
						"importPath": imp.Path,
						"specifiers": imp.Specifiers,
						"isDefault":  imp.IsDefault,
					},
				}
				gb.graph.Edges[edgeId] = edge
			}
		}
	}
}

// resolveImportPath attempts to resolve an import path to an actual file
func (gb *GraphBuilder) resolveImportPath(importPath, fromFile string) string {
	// Normalize the fromFile path
	fromFile = gb.normalizePath(fromFile)
	
	// Handle relative imports
	if strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") {
		dir := filepath.Dir(fromFile)
		
		// Validate import path for security - prevent directory traversal
		if err := gb.validateImportPath(importPath, dir); err != nil {
			if gb.logger != nil {
				gb.logger.Printf("Invalid import path: %v", err)
			}
			return ""
		}
		
		resolved := gb.normalizePath(filepath.Join(dir, importPath))

		// Try common extensions
		extensions := []string{".ts", ".tsx", ".js", ".jsx"}
		for _, ext := range extensions {
			candidate := gb.normalizePath(resolved + ext)
			if _, exists := gb.graph.Files[candidate]; exists {
				return candidate
			}
		}

		// Try with index files
		for _, ext := range extensions {
			candidate := gb.normalizePath(filepath.Join(resolved, "index"+ext))
			if _, exists := gb.graph.Files[candidate]; exists {
				return candidate
			}
		}
	}

	// For now, we don't resolve node_modules or absolute imports
	// This could be enhanced later
	return ""
}

// isSupportedFile checks if a file is supported for parsing
func (gb *GraphBuilder) isSupportedFile(path string) bool {
	ext := filepath.Ext(path)
	// Include all languages supported by the parser manager
	supportedExtensions := []string{
		// JavaScript/TypeScript
		".ts", ".tsx", ".js", ".jsx", ".mts", ".cts", ".mjs", ".cjs",
		// Go
		".go",
		// Python
		".py", ".pyi",
		// Java
		".java",
		// Rust
		".rs",
		// Config files
		".json", ".yaml", ".yml",
		// Markdown (for documentation)
		".md",
	}

	return slices.Contains(supportedExtensions, ext)
}

// getMergedPatterns returns the combined exclude patterns (defaults + user patterns)
// with thread-safe caching and memory leak protection
func (gb *GraphBuilder) getMergedPatterns() []string {
	// Fast path: try read lock first
	gb.patternMu.RLock()
	if !gb.patternsDirty && gb.cachedPatterns != nil {
		defer gb.patternMu.RUnlock()
		return gb.cachedPatterns
	}
	gb.patternMu.RUnlock()

	// Slow path: need to rebuild cache
	gb.patternMu.Lock()
	defer gb.patternMu.Unlock()

	// Double-check pattern (classic concurrent programming)
	if !gb.patternsDirty && gb.cachedPatterns != nil {
		return gb.cachedPatterns
	}

	// Check for memory leak prevention
	defaultPatterns := getDefaultExcludePatterns()
	totalPatterns := len(gb.excludePatterns)
	if gb.useDefaultExcludes {
		totalPatterns += len(defaultPatterns)
	}

	if totalPatterns > MaxCachedPatterns {
		// Don't cache very large pattern sets to prevent memory leaks
		return gb.buildPatternsUncached(defaultPatterns)
	}

	// Rebuild cache
	if gb.useDefaultExcludes {
		// Merge default and user patterns
		gb.cachedPatterns = make([]string, 0, totalPatterns)
		gb.cachedPatterns = append(gb.cachedPatterns, defaultPatterns...)
		gb.cachedPatterns = append(gb.cachedPatterns, gb.excludePatterns...)
	} else {
		// Use only user patterns
		gb.cachedPatterns = make([]string, len(gb.excludePatterns))
		copy(gb.cachedPatterns, gb.excludePatterns)
	}

	gb.patternsDirty = false
	return gb.cachedPatterns
}

// buildPatternsUncached builds patterns without caching for large pattern sets
func (gb *GraphBuilder) buildPatternsUncached(defaultPatterns []string) []string {
	if gb.useDefaultExcludes {
		result := make([]string, 0, len(defaultPatterns)+len(gb.excludePatterns))
		result = append(result, defaultPatterns...)
		result = append(result, gb.excludePatterns...)
		return result
	}

	// Return copy to avoid external modification
	result := make([]string, len(gb.excludePatterns))
	copy(result, gb.excludePatterns)
	return result
}

// shouldSkipPath checks if a path should be skipped during analysis
func (gb *GraphBuilder) shouldSkipPath(path string) bool {
	// Normalize path for consistent comparison across platforms
	path = gb.normalizePath(path)
	
	// First check if path is explicitly included (negation patterns)
	if gb.matchesPattern(path, gb.includePatterns) {
		return false // Explicitly included, don't skip
	}

	// Check against all exclude patterns (cached to avoid repeated allocations)
	return gb.matchesPattern(path, gb.getMergedPatterns())
}

// matchesPattern checks if a path matches any of the given patterns
// Returns true if any pattern matches, false otherwise
func (gb *GraphBuilder) matchesPattern(path string, patterns []string) bool {
	// Normalize path for consistent cross-platform matching
	path = gb.normalizePath(path)
	
	// Use forward slashes for pattern matching (cross-platform consistency)
	patternPath := gb.normalizeForPattern(path)
	
	for _, pattern := range patterns {
		// Skip empty patterns (these are filtered during deduplication)
		if pattern == "" {
			continue
		}

		// Normalize pattern for cross-platform compatibility
		normalizedPattern := filepath.ToSlash(pattern)

		// Check full path match using normalized paths
		if matched, err := gb.checkPatternMatch(normalizedPattern, patternPath); err != nil {
			gb.logPatternError(pattern, err)
			continue
		} else if matched {
			return true
		}

		// Also check against just the filename for patterns like *.test.*
		if gb.hasDirectorySeparator(path) {
			baseName := filepath.Base(patternPath)
			if matched, err := gb.checkPatternMatch(normalizedPattern, baseName); err != nil {
				gb.logPatternError(pattern, err)
				continue
			} else if matched {
				return true
			}
		}

		// Check individual path components for patterns like *temp*
		// This allows matching directory names within paths
		pathComponents := getStringSlice()
		pathComponents = append(pathComponents, strings.Split(patternPath, "/")...)
		
		matched := false
		for _, component := range pathComponents {
			if component != "" { // Skip empty components
				if m, err := gb.checkPatternMatch(normalizedPattern, component); err != nil {
					gb.logPatternError(pattern, err)
					continue
				} else if m {
					matched = true
					break
				}
			}
		}
		
		putStringSlice(pathComponents)
		if matched {
			return true
		}

		// Handle ** patterns which filepath.Match doesn't support natively
		if strings.Contains(normalizedPattern, "**") {
			if gb.matchesDoubleStarPattern(patternPath, normalizedPattern) {
				return true
			}
		}
	}

	return false
}

// checkPatternMatch performs a single pattern match with error handling
func (gb *GraphBuilder) checkPatternMatch(pattern, path string) (bool, error) {
	matched, err := filepath.Match(pattern, path)
	return matched, err
}

// logPatternError logs pattern errors using the configured logger
func (gb *GraphBuilder) logPatternError(pattern string, err error) {
	if gb.logger != nil {
		gb.logger.Printf("Invalid glob pattern %q: %v", pattern, err)
	}
	// Still send to progress callback for backward compatibility
	if gb.progressCallback != nil {
		gb.progressCallback(fmt.Sprintf("⚠️  Invalid pattern %q: %v", pattern, err))
	}
}

// hasDirectorySeparator checks if path contains directory separators
// Works with both forward slashes and backslashes for cross-platform compatibility
func (gb *GraphBuilder) hasDirectorySeparator(path string) bool {
	// Normalize path first to handle both / and \ separators
	normalized := gb.normalizeForPattern(path)
	return strings.Contains(normalized, "/")
}

// matchesDoubleStarPattern handles patterns with ** wildcards
func (gb *GraphBuilder) matchesDoubleStarPattern(path, pattern string) bool {
	// Handle different ** patterns:
	// 1. **/filename.ext - match filename at any depth
	// 2. prefix/**/suffix - match prefix and suffix with any levels between
	// 3. **/*.ext - match any file with extension at any depth
	
	// Use pooled slices to reduce allocations in recursive pattern matching
	pathParts := getStringSlice()
	patternParts := getStringSlice()
	
	pathParts = append(pathParts, strings.Split(path, "/")...)
	patternParts = append(patternParts, strings.Split(pattern, "/")...)
	
	result := gb.matchDoubleStarRecursive(pathParts, patternParts, 0, 0)
	
	putStringSlice(pathParts)
	putStringSlice(patternParts)
	
	return result
}

// matchDoubleStarRecursive recursively matches path and pattern parts handling **
func (gb *GraphBuilder) matchDoubleStarRecursive(pathParts, patternParts []string, pathIdx, patternIdx int) bool {
	// If we've matched all pattern parts, check if we've consumed all path parts
	if patternIdx >= len(patternParts) {
		return pathIdx >= len(pathParts)
	}
	
	// If we've consumed all path parts but have more pattern parts
	if pathIdx >= len(pathParts) {
		// Only OK if remaining patterns are all ** (which can match zero directories)
		for i := patternIdx; i < len(patternParts); i++ {
			if patternParts[i] != "**" {
				return false
			}
		}
		return true
	}
	
	currentPattern := patternParts[patternIdx]
	
	// Handle ** - it can match zero or more directory levels
	if currentPattern == "**" {
		// Try matching ** with zero directories (skip it)
		if gb.matchDoubleStarRecursive(pathParts, patternParts, pathIdx, patternIdx+1) {
			return true
		}
		
		// Try matching ** with one or more directories
		for i := pathIdx + 1; i <= len(pathParts); i++ {
			if gb.matchDoubleStarRecursive(pathParts, patternParts, i, patternIdx+1) {
				return true
			}
		}
		return false
	}
	
	// Handle regular pattern matching for current part
	currentPath := pathParts[pathIdx]
	matched, err := gb.checkPatternMatch(currentPattern, currentPath)
	if err != nil {
		gb.logPatternError(currentPattern, err)
		return false
	}
	
	if !matched {
		return false
	}
	
	// Continue with next parts
	return gb.matchDoubleStarRecursive(pathParts, patternParts, pathIdx+1, patternIdx+1)
}

// GetSupportedLanguages returns the list of supported languages
func (gb *GraphBuilder) GetSupportedLanguages() []types.Language {
	return gb.parser.GetSupportedLanguages()
}

// GetFileStats returns statistics about the analyzed files
func (gb *GraphBuilder) GetFileStats() map[string]interface{} {
	if gb.graph.Metadata == nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"totalFiles":   gb.graph.Metadata.TotalFiles,
		"totalSymbols": gb.graph.Metadata.TotalSymbols,
		"languages":    gb.graph.Metadata.Languages,
		"analysisTime": gb.graph.Metadata.AnalysisTime,
	}
}

// SemanticAnalysisResult contains the results of semantic neighborhood analysis
type SemanticAnalysisResult struct {
	SemanticNeighborhoods  []git.SemanticNeighborhood  `json:"semantic_neighborhoods"`
	EnhancedNeighborhoods  []git.EnhancedNeighborhood  `json:"enhanced_neighborhoods"`
	ClusteredNeighborhoods []git.ClusteredNeighborhood `json:"clustered_neighborhoods"`
	AnalysisMetadata       SemanticAnalysisMetadata    `json:"analysis_metadata"`
	Error                  string                      `json:"error,omitempty"`
}

// SemanticAnalysisMetadata contains metadata about the semantic analysis
type SemanticAnalysisMetadata struct {
	IsGitRepository    bool          `json:"is_git_repository"`
	AnalysisPeriodDays int           `json:"analysis_period_days"`
	TotalNeighborhoods int           `json:"total_neighborhoods"`
	TotalClusters      int           `json:"total_clusters"`
	FilesWithPatterns  int           `json:"files_with_patterns"`
	AverageClusterSize float64       `json:"average_cluster_size"`
	AnalysisTime       time.Duration `json:"analysis_time"`
	QualityScores      QualityScores `json:"quality_scores"`
}

// QualityScores contains overall quality metrics for the clustering
type QualityScores struct {
	AverageSilhouetteScore    float64 `json:"average_silhouette_score"`
	AverageDaviesBouldinIndex float64 `json:"average_davies_bouldin_index"`
	OverallQualityRating      string  `json:"overall_quality_rating"`
}

// buildSemanticNeighborhoods analyzes git patterns and builds semantic neighborhoods
func (gb *GraphBuilder) buildSemanticNeighborhoods(targetDir string) (*SemanticAnalysisResult, error) {
	start := time.Now()

	// Initialize git analyzer
	gitAnalyzer, err := git.NewGitAnalyzer(targetDir)
	if err != nil {
		return &SemanticAnalysisResult{
			Error: fmt.Sprintf("Failed to create git analyzer: %v", err),
			AnalysisMetadata: SemanticAnalysisMetadata{
				IsGitRepository: false,
				AnalysisTime:    time.Since(start),
			},
		}, nil
	}

	// Check if this is a git repository
	if !gitAnalyzer.IsGitRepository() {
		return &SemanticAnalysisResult{
			AnalysisMetadata: SemanticAnalysisMetadata{
				IsGitRepository: false,
				AnalysisTime:    time.Since(start),
			},
		}, nil
	}

	// Create semantic analyzer with default config
	semanticConfig := git.DefaultSemanticConfig()
	semanticAnalyzer, err := git.NewSemanticAnalyzer(targetDir, semanticConfig)
	if err != nil {
		return &SemanticAnalysisResult{
			Error: fmt.Sprintf("Failed to create semantic analyzer: %v", err),
			AnalysisMetadata: SemanticAnalysisMetadata{
				IsGitRepository: true,
				AnalysisTime:    time.Since(start),
			},
		}, nil
	}

	// Perform semantic analysis
	analysisResult, err := semanticAnalyzer.AnalyzeRepository()
	if err != nil {
		return &SemanticAnalysisResult{
			Error: fmt.Sprintf("Failed to analyze repository: %v", err),
			AnalysisMetadata: SemanticAnalysisMetadata{
				IsGitRepository: true,
				AnalysisTime:    time.Since(start),
			},
		}, nil
	}

	// Build enhanced neighborhoods using graph integration
	integrationConfig := git.DefaultIntegrationConfig()
	graphIntegration := git.NewGraphIntegration(semanticAnalyzer, gb.graph, integrationConfig)

	enhancedNeighborhoods, err := graphIntegration.BuildEnhancedNeighborhoods()
	if err != nil {
		return &SemanticAnalysisResult{
			SemanticNeighborhoods: analysisResult.Neighborhoods,
			Error:                 fmt.Sprintf("Failed to build enhanced neighborhoods: %v", err),
			AnalysisMetadata: SemanticAnalysisMetadata{
				IsGitRepository:    true,
				AnalysisPeriodDays: semanticConfig.AnalysisPeriodDays,
				TotalNeighborhoods: len(analysisResult.Neighborhoods),
				FilesWithPatterns:  analysisResult.AnalysisSummary.ActiveFiles,
				AnalysisTime:       time.Since(start),
			},
		}, nil
	}

	// Build clustered neighborhoods
	clusteredNeighborhoods, err := graphIntegration.BuildClusteredNeighborhoods()
	if err != nil {
		return &SemanticAnalysisResult{
			SemanticNeighborhoods: analysisResult.Neighborhoods,
			EnhancedNeighborhoods: enhancedNeighborhoods,
			Error:                 fmt.Sprintf("Failed to build clustered neighborhoods: %v", err),
			AnalysisMetadata: SemanticAnalysisMetadata{
				IsGitRepository:    true,
				AnalysisPeriodDays: semanticConfig.AnalysisPeriodDays,
				TotalNeighborhoods: len(analysisResult.Neighborhoods),
				FilesWithPatterns:  analysisResult.AnalysisSummary.ActiveFiles,
				AnalysisTime:       time.Since(start),
			},
		}, nil
	}

	// Calculate quality scores
	qualityScores := gb.calculateQualityScores(clusteredNeighborhoods)

	// Calculate average cluster size
	avgClusterSize := 0.0
	if len(clusteredNeighborhoods) > 0 {
		totalSize := 0
		for _, cluster := range clusteredNeighborhoods {
			totalSize += cluster.Cluster.Size
		}
		avgClusterSize = float64(totalSize) / float64(len(clusteredNeighborhoods))
	}

	return &SemanticAnalysisResult{
		SemanticNeighborhoods:  analysisResult.Neighborhoods,
		EnhancedNeighborhoods:  enhancedNeighborhoods,
		ClusteredNeighborhoods: clusteredNeighborhoods,
		AnalysisMetadata: SemanticAnalysisMetadata{
			IsGitRepository:    true,
			AnalysisPeriodDays: semanticConfig.AnalysisPeriodDays,
			TotalNeighborhoods: len(analysisResult.Neighborhoods),
			TotalClusters:      len(clusteredNeighborhoods),
			FilesWithPatterns:  analysisResult.AnalysisSummary.ActiveFiles,
			AverageClusterSize: avgClusterSize,
			AnalysisTime:       time.Since(start),
			QualityScores:      qualityScores,
		},
	}, nil
}

// calculateQualityScores calculates overall quality metrics from clustered neighborhoods
func (gb *GraphBuilder) calculateQualityScores(clusteredNeighborhoods []git.ClusteredNeighborhood) QualityScores {
	if len(clusteredNeighborhoods) == 0 {
		return QualityScores{
			OverallQualityRating: "No clusters",
		}
	}

	totalSilhouette := 0.0
	totalDaviesBouldin := 0.0
	validClusters := 0

	for _, cluster := range clusteredNeighborhoods {
		if cluster.QualityMetrics.SilhouetteScore > 0 {
			totalSilhouette += cluster.QualityMetrics.SilhouetteScore
			totalDaviesBouldin += cluster.QualityMetrics.DaviesBouldinIndex
			validClusters++
		}
	}

	if validClusters == 0 {
		return QualityScores{
			OverallQualityRating: "Insufficient data",
		}
	}

	avgSilhouette := totalSilhouette / float64(validClusters)
	avgDaviesBouldin := totalDaviesBouldin / float64(validClusters)

	// Determine overall quality rating
	qualityRating := "Poor"
	if avgSilhouette > 0.7 {
		qualityRating = "Excellent"
	} else if avgSilhouette > 0.5 {
		qualityRating = "Good"
	} else if avgSilhouette > 0.25 {
		qualityRating = "Fair"
	}

	return QualityScores{
		AverageSilhouetteScore:    avgSilhouette,
		AverageDaviesBouldinIndex: avgDaviesBouldin,
		OverallQualityRating:      qualityRating,
	}
}

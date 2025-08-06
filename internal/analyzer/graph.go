package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nuthan-ms/codecontext/internal/cache"
	"github.com/nuthan-ms/codecontext/internal/git"
	"github.com/nuthan-ms/codecontext/internal/parser"
	"github.com/nuthan-ms/codecontext/pkg/types"
)

// GraphBuilder builds code graphs from parsed files
// ProgressConfig configures progress reporting behavior
type ProgressConfig struct {
	Interval       int  // Update progress every N files (default: 10)
	ShowPercentage bool // Show percentage progress if total count is known
}

type GraphBuilder struct {
	parser           *parser.Manager
	graph            *types.CodeGraph
	cache            *cache.PersistentCache
	progressCallback func(string)
	progressConfig   ProgressConfig
	excludePatterns  []string
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
			Interval:       10,    // Default: update every 10 files
			ShowPercentage: false, // Default: don't show percentage (requires pre-counting)
		},
	}
}

// SetCache sets the persistent cache for the graph builder
func (gb *GraphBuilder) SetCache(c *cache.PersistentCache) {
	gb.cache = c
}

// SetExcludePatterns sets the exclude patterns for the graph builder
func (gb *GraphBuilder) SetExcludePatterns(patterns []string) {
	gb.excludePatterns = patterns
}

// SetProgressCallback sets a callback function for progress updates
func (gb *GraphBuilder) SetProgressCallback(callback func(string)) {
	gb.progressCallback = callback
}

// SetProgressInterval sets how often progress updates are sent (every N files)
func (gb *GraphBuilder) SetProgressInterval(interval int) {
	if interval > 0 {
		gb.progressConfig.Interval = interval
	}
}

// SetProgressConfig sets the complete progress configuration
func (gb *GraphBuilder) SetProgressConfig(config ProgressConfig) {
	if config.Interval > 0 {
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
	// Handle relative imports
	if strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") {
		dir := filepath.Dir(fromFile)
		resolved := filepath.Join(dir, importPath)

		// Try common extensions
		extensions := []string{".ts", ".tsx", ".js", ".jsx"}
		for _, ext := range extensions {
			candidate := resolved + ext
			if _, exists := gb.graph.Files[candidate]; exists {
				return candidate
			}
		}

		// Try with index files
		for _, ext := range extensions {
			candidate := filepath.Join(resolved, "index"+ext)
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

	for _, supported := range supportedExtensions {
		if ext == supported {
			return true
		}
	}
	return false
}

// shouldSkipPath checks if a path should be skipped during analysis
func (gb *GraphBuilder) shouldSkipPath(path string) bool {
	// Default skip directories
	skipDirs := []string{
		"node_modules", ".git", ".codecontext", "dist", "build",
		"coverage", ".nyc_output",
	}

	// Check default skip directories
	for _, skipDir := range skipDirs {
		// Check if the path contains the directory as a segment or if it's the exact directory name
		if strings.Contains(path, "/"+skipDir+"/") || strings.HasSuffix(path, "/"+skipDir) || path == skipDir {
			return true
		}
	}

	// Check exclude patterns from config
	for _, pattern := range gb.excludePatterns {
		// Use filepath.Match for glob pattern matching
		matched, err := filepath.Match(pattern, path)
		if err == nil && matched {
			return true
		}

		// Also check against just the filename for patterns like *.test.*
		baseName := filepath.Base(path)
		if baseName != path { // Only if path contains directories
			matched, err = filepath.Match(pattern, baseName)
			if err == nil && matched {
				return true
			}
		}

		// Also check if the pattern matches any part of the path (for patterns like "node_modules/**")
		// This handles patterns with ** which filepath.Match doesn't support natively
		if strings.Contains(pattern, "**") {
			// Convert ** to a simpler pattern for basic matching
			simplePattern := strings.ReplaceAll(pattern, "/**", "")
			simplePattern = strings.ReplaceAll(simplePattern, "**", "*")

			// Check if any segment of the path matches
			parts := strings.Split(path, string(filepath.Separator))
			for i := range parts {
				subPath := filepath.Join(parts[:i+1]...)
				if matched, err := filepath.Match(simplePattern, subPath); err == nil && matched {
					return true
				}
			}
		}
	}

	return false
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

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nuthan-ms/codecontext/internal/analyzer"
	"github.com/nuthan-ms/codecontext/internal/git"
	"github.com/nuthan-ms/codecontext/internal/watcher"
	"github.com/nuthan-ms/codecontext/pkg/types"
)

// MCPConfig holds configuration for the MCP server
type MCPConfig struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	TargetDir   string `json:"target_dir"`
	EnableWatch bool   `json:"enable_watch"`
	DebounceMs  int    `json:"debounce_ms"`
}

// CodeContextMCPServer provides codecontext functionality via MCP
type CodeContextMCPServer struct {
	server   *mcp.Server
	config   *MCPConfig
	watcher  *watcher.FileWatcher
	graph    *types.CodeGraph
	analyzer *analyzer.GraphBuilder
	stopMutex sync.RWMutex // Protect against concurrent stop operations
	stopped   bool         // Track server state
}

// Tool argument structs
type GetCodebaseOverviewArgs struct {
	IncludeStats bool   `json:"include_stats"`
	TargetDir    string `json:"target_dir,omitempty"` // Optional: directory to analyze
}

type GetFileAnalysisArgs struct {
	FilePath  string `json:"file_path"`
	TargetDir string `json:"target_dir,omitempty"` // Optional: directory to analyze
}

type GetSymbolInfoArgs struct {
	SymbolName    string `json:"symbol_name"`
	FilePath      string `json:"file_path,omitempty"`
	FrameworkType string `json:"framework_type,omitempty"`
	TargetDir     string `json:"target_dir,omitempty"` // Optional: directory to analyze
}

type SearchSymbolsArgs struct {
	Query         string `json:"query"`
	FileType      string `json:"file_type,omitempty"`
	SymbolType    string `json:"symbol_type,omitempty"`
	FrameworkType string `json:"framework_type,omitempty"`
	Limit         int    `json:"limit,omitempty"`
	TargetDir     string `json:"target_dir,omitempty"` // Optional: directory to analyze
}

type GetDependenciesArgs struct {
	FilePath  string `json:"file_path,omitempty"`
	Direction string `json:"direction,omitempty"`
	TargetDir string `json:"target_dir,omitempty"` // Optional: directory to analyze
}

type WatchChangesArgs struct {
	Enable    bool   `json:"enable"`
	TargetDir string `json:"target_dir,omitempty"` // Optional: directory to watch
}

type GetSemanticNeighborhoodsArgs struct {
	FilePath     string `json:"file_path,omitempty"`
	IncludeBasic bool   `json:"include_basic,omitempty"`
	IncludeQuality bool `json:"include_quality,omitempty"`
	MaxResults   int    `json:"max_results,omitempty"`
	TargetDir    string `json:"target_dir,omitempty"` // Optional: directory to analyze
}

type GetFrameworkAnalysisArgs struct {
	Framework    string `json:"framework,omitempty"`
	IncludeStats bool   `json:"include_stats,omitempty"`
	TargetDir    string `json:"target_dir,omitempty"` // Optional: directory to analyze
}

// NewCodeContextMCPServer creates a new MCP server instance
func NewCodeContextMCPServer(config *MCPConfig) (*CodeContextMCPServer, error) {
	// Redirect all logging to stderr for MCP compatibility
	log.SetOutput(os.Stderr)
	log.Printf("[MCP] Creating new CodeContext MCP server with config: %+v", config)
	
	// Create server with official SDK pattern
	server := mcp.NewServer(&mcp.Implementation{
		Name:    config.Name,
		Version: config.Version,
	}, nil)
	log.Printf("[MCP] Created MCP server with name=%s, version=%s", config.Name, config.Version)
	
	s := &CodeContextMCPServer{
		server:   server,
		config:   config,
		analyzer: analyzer.NewGraphBuilder(),
	}
	log.Printf("[MCP] Created CodeContextMCPServer instance")

	// Register tools
	log.Printf("[MCP] Registering tools...")
	s.registerTools()
	log.Printf("[MCP] All tools registered successfully")
	
	return s, nil
}

// registerTools registers all MCP tools
func (s *CodeContextMCPServer) registerTools() {
	// Tool 1: Get codebase overview
	log.Printf("[MCP] Registering tool: get_codebase_overview")
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_codebase_overview",
		Description: "Get comprehensive overview of a codebase. Optional target_dir parameter allows analyzing different projects (supports ~/path and absolute paths).",
	}, s.getCodebaseOverview)

	// Tool 2: Get file analysis
	log.Printf("[MCP] Registering tool: get_file_analysis")
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_file_analysis",
		Description: "Get detailed analysis of a specific file. Optional target_dir parameter allows analyzing files in different projects.",
	}, s.getFileAnalysis)

	// Tool 3: Get symbol information
	log.Printf("[MCP] Registering tool: get_symbol_info")
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_symbol_info",
		Description: "Get detailed information about a specific symbol, including framework-specific details (React components, Vue stores, Angular services, etc.). Optional target_dir parameter allows searching symbols in different projects.",
	}, s.getSymbolInfo)

	// Tool 4: Search symbols
	log.Printf("[MCP] Registering tool: search_symbols")
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "search_symbols",
		Description: "Search for symbols across a codebase with framework-aware filtering (components, hooks, services, stores, etc.). Optional target_dir parameter allows searching in different projects.",
	}, s.searchSymbols)

	// Tool 5: Get dependencies
	log.Printf("[MCP] Registering tool: get_dependencies")
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_dependencies",
		Description: "Analyze import dependencies and relationships. Optional target_dir parameter allows analyzing dependencies in different projects.",
	}, s.getDependencies)

	// Tool 6: Watch changes (real-time)
	log.Printf("[MCP] Registering tool: watch_changes")
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "watch_changes",
		Description: "Enable/disable real-time change notifications. Optional target_dir parameter allows watching different project directories.",
	}, s.watchChanges)

	// Tool 7: Get semantic neighborhoods
	log.Printf("[MCP] Registering tool: get_semantic_neighborhoods")
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_semantic_neighborhoods",
		Description: "Get semantic code neighborhoods using git patterns and hierarchical clustering. Optional target_dir parameter allows analyzing neighborhoods in different projects.",
	}, s.getSemanticNeighborhoods)

	// Tool 8: Get framework analysis
	log.Printf("[MCP] Registering tool: get_framework_analysis")
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_framework_analysis",
		Description: "Get comprehensive framework-specific analysis including component relationships, hook usage patterns, and framework-specific metrics. Optional target_dir parameter allows analyzing different projects.",
	}, s.getFrameworkAnalysis)
	
	log.Printf("[MCP] Successfully registered 8 tools")
}

// Tool implementations

func (s *CodeContextMCPServer) getCodebaseOverview(ctx context.Context, req *mcp.CallToolRequest, args GetCodebaseOverviewArgs) (*mcp.CallToolResult, any, error) {
	log.Printf("[MCP] Tool called: get_codebase_overview with args: %+v", args)
	start := time.Now()
	
	// Resolve target directory
	targetDir := s.resolveTargetDir(args.TargetDir)
	
	// Ensure we have fresh analysis
	log.Printf("[MCP] Refreshing analysis for codebase overview...")
	if err := s.refreshAnalysisWithTargetDir(targetDir); err != nil {
		log.Printf("[MCP] ERROR: Failed to refresh analysis: %v", err)
		return nil, nil, fmt.Errorf("failed to refresh analysis: %w", err)
	}

	log.Printf("[MCP] Generating markdown content...")
	generator := analyzer.NewMarkdownGenerator(s.graph)
	content := generator.GenerateContextMap()
	log.Printf("[MCP] Generated markdown content (%d chars)", len(content))

	if args.IncludeStats {
		log.Printf("[MCP] Including detailed statistics...")
		stats := s.analyzer.GetFileStats()
		statsJson, _ := json.MarshalIndent(stats, "", "  ")
		content += "\n\n## Detailed Statistics\n```json\n" + string(statsJson) + "\n```"
		log.Printf("[MCP] Added statistics to content")
	}

	elapsed := time.Since(start)
	log.Printf("[MCP] Tool completed: get_codebase_overview (took %v)", elapsed)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: content}},
	}, nil, nil
}

func (s *CodeContextMCPServer) getFileAnalysis(ctx context.Context, req *mcp.CallToolRequest, args GetFileAnalysisArgs) (*mcp.CallToolResult, any, error) {
	log.Printf("[MCP] Tool called: get_file_analysis with args: %+v", args)
	start := time.Now()
	
	if args.FilePath == "" {
		log.Printf("[MCP] ERROR: file_path is required")
		return nil, nil, fmt.Errorf("file_path is required")
	}

	// Resolve target directory
	targetDir := s.resolveTargetDir(args.TargetDir)

	// Ensure we have fresh analysis
	log.Printf("[MCP] Refreshing analysis for file: %s", args.FilePath)
	if err := s.refreshAnalysisWithTargetDir(targetDir); err != nil {
		log.Printf("[MCP] ERROR: Failed to refresh analysis: %v", err)
		return nil, nil, fmt.Errorf("failed to refresh analysis: %w", err)
	}

	// Find the file in our graph
	log.Printf("[MCP] Looking up file in graph: %s", args.FilePath)
	fileNode, exists := s.graph.Files[args.FilePath]
	if !exists {
		log.Printf("[MCP] ERROR: File not found in graph: %s (available files: %d)", args.FilePath, len(s.graph.Files))
		return nil, nil, fmt.Errorf("file not found: %s", args.FilePath)
	}
	log.Printf("[MCP] Found file in graph: %s (language: %s, lines: %d, symbols: %d)", args.FilePath, fileNode.Language, fileNode.Lines, len(fileNode.Symbols))

	// Build detailed file analysis
	analysis := fmt.Sprintf("# File Analysis: %s\n\n", args.FilePath)
	analysis += fmt.Sprintf("**Language:** %s\n", fileNode.Language)
	analysis += fmt.Sprintf("**Lines:** %d\n", fileNode.Lines)
	analysis += fmt.Sprintf("**Symbols:** %d\n\n", len(fileNode.Symbols))

	// List symbols in this file
	if len(fileNode.Symbols) > 0 {
		analysis += "## Symbols\n\n"
		for _, symbolId := range fileNode.Symbols {
			if symbol, exists := s.graph.Symbols[symbolId]; exists {
				analysis += fmt.Sprintf("- **%s** (%s) - Line %d\n", 
					symbol.Name, symbol.Kind, symbol.Location.StartLine)
			}
		}
	}

	// List imports for this file
	log.Printf("[MCP] Analyzing dependencies for file: %s", args.FilePath)
	analysis += "\n## Dependencies\n\n"
	importCount := 0
	for _, edge := range s.graph.Edges {
		if edge.Type == "imports" && edge.From == types.NodeId(args.FilePath) {
			if importCount == 0 {
				analysis += "### Imports:\n"
			}
			analysis += fmt.Sprintf("- %s\n", edge.To)
			importCount++
		}
	}
	if importCount == 0 {
		analysis += "No imports found.\n"
	}
	log.Printf("[MCP] Found %d imports for file: %s", importCount, args.FilePath)

	elapsed := time.Since(start)
	log.Printf("[MCP] Tool completed: get_file_analysis (took %v)", elapsed)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: analysis}},
	}, nil, nil
}

func (s *CodeContextMCPServer) getSymbolInfo(ctx context.Context, req *mcp.CallToolRequest, args GetSymbolInfoArgs) (*mcp.CallToolResult, any, error) {
	log.Printf("[MCP] Tool called: get_symbol_info with args: %+v", args)
	start := time.Now()
	
	if args.SymbolName == "" {
		log.Printf("[MCP] ERROR: symbol_name is required")
		return nil, nil, fmt.Errorf("symbol_name is required")
	}

	// Resolve target directory
	targetDir := s.resolveTargetDir(args.TargetDir)

	// Ensure we have fresh analysis
	log.Printf("[MCP] Refreshing analysis for symbol lookup: %s", args.SymbolName)
	if err := s.refreshAnalysisWithTargetDir(targetDir); err != nil {
		log.Printf("[MCP] ERROR: Failed to refresh analysis: %v", err)
		return nil, nil, fmt.Errorf("failed to refresh analysis: %w", err)
	}

	log.Printf("[MCP] Searching for symbol: %s in %d symbols", args.SymbolName, len(s.graph.Symbols))
	var foundSymbols []*types.Symbol
	for _, symbol := range s.graph.Symbols {
		if symbol.Name == args.SymbolName {
			foundSymbols = append(foundSymbols, symbol)
		}
	}

	log.Printf("[MCP] Found %d symbols matching '%s'", len(foundSymbols), args.SymbolName)
	if len(foundSymbols) == 0 {
		log.Printf("[MCP] ERROR: Symbol not found: %s", args.SymbolName)
		return nil, nil, fmt.Errorf("symbol '%s' not found", args.SymbolName)
	}

	result := fmt.Sprintf("# Symbol Information: %s\n\n", args.SymbolName)
	
	for i, symbol := range foundSymbols {
		if i > 0 {
			result += "\n---\n\n"
		}
		result += fmt.Sprintf("**Line:** %d\n", symbol.Location.StartLine)
		result += fmt.Sprintf("**Type:** %s\n", symbol.Kind)
		
		// Add framework-specific information
		if symbol.Type != "" && string(symbol.Type) != symbol.Kind {
			result += fmt.Sprintf("**Framework Type:** %s\n", symbol.Type)
			result += s.getFrameworkSpecificDescription(string(symbol.Type))
		}
		
		if symbol.Signature != "" {
			result += fmt.Sprintf("**Signature:** `%s`\n", symbol.Signature)
		}
		if symbol.Documentation != "" {
			result += fmt.Sprintf("**Documentation:** %s\n", symbol.Documentation)
		}
		
		// Add framework-specific insights
		if frameworkInsights := s.getFrameworkInsights(symbol); frameworkInsights != "" {
			result += fmt.Sprintf("**Framework Insights:** %s\n", frameworkInsights)
		}
	}

	elapsed := time.Since(start)
	log.Printf("[MCP] Tool completed: get_symbol_info (took %v)", elapsed)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: result}},
	}, nil, nil
}

func (s *CodeContextMCPServer) searchSymbols(ctx context.Context, req *mcp.CallToolRequest, args SearchSymbolsArgs) (*mcp.CallToolResult, any, error) {
	log.Printf("[MCP] Tool called: search_symbols with args: %+v", args)
	start := time.Now()
	
	if args.Query == "" {
		log.Printf("[MCP] ERROR: query is required")
		return nil, nil, fmt.Errorf("query is required")
	}

	// Set default limit
	if args.Limit <= 0 {
		args.Limit = 20
	}
	log.Printf("[MCP] Searching symbols with query='%s', limit=%d", args.Query, args.Limit)

	// Resolve target directory
	targetDir := s.resolveTargetDir(args.TargetDir)

	// Ensure we have fresh analysis
	log.Printf("[MCP] Refreshing analysis for symbol search...")
	if err := s.refreshAnalysisWithTargetDir(targetDir); err != nil {
		log.Printf("[MCP] ERROR: Failed to refresh analysis: %v", err)
		return nil, nil, fmt.Errorf("failed to refresh analysis: %w", err)
	}

	var matches []*types.Symbol
	query := strings.ToLower(args.Query)
	log.Printf("[MCP] Searching through %d symbols for query: %s", len(s.graph.Symbols), query)

	for _, symbol := range s.graph.Symbols {
		// Check name match
		nameMatch := strings.Contains(strings.ToLower(symbol.Name), query)
		
		// Check framework type filter
		frameworkMatch := true
		if args.FrameworkType != "" {
			frameworkMatch = s.matchesFramework(symbol, args.FrameworkType)
		}
		
		// Check symbol type filter
		symbolTypeMatch := true
		if args.SymbolType != "" {
			symbolTypeMatch = strings.EqualFold(string(symbol.Type), args.SymbolType)
		}
		
		if nameMatch && frameworkMatch && symbolTypeMatch {
			matches = append(matches, symbol)
			if len(matches) >= args.Limit {
				log.Printf("[MCP] Reached limit of %d matches", args.Limit)
				break
			}
		}
	}

	if len(matches) == 0 {
		result := fmt.Sprintf("No symbols found matching '%s'", args.Query)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: result}},
		}, nil, nil
	}

	result := fmt.Sprintf("# Symbol Search Results: '%s'\n\n", args.Query)
	if args.SymbolType != "" || args.FrameworkType != "" {
		result += fmt.Sprintf("**Filters Applied:** ")
		if args.SymbolType != "" {
			result += fmt.Sprintf("Symbol Type: %s ", args.SymbolType)
		}
		if args.FrameworkType != "" {
			result += fmt.Sprintf("Framework: %s ", args.FrameworkType)
		}
		result += "\n\n"
	}
	result += fmt.Sprintf("Found %d matches:\n\n", len(matches))

	for _, symbol := range matches {
		frameworkInfo := ""
		if symbol.Type != "" && string(symbol.Type) != symbol.Kind {
			frameworkInfo = fmt.Sprintf(" [%s]", symbol.Type)
		}
		result += fmt.Sprintf("- **%s**%s (%s) - Line %d\n", 
			symbol.Name, frameworkInfo, symbol.Kind, symbol.Location.StartLine)
		
		// Add framework-specific details
		if insight := s.getFrameworkInsights(symbol); insight != "" {
			result += fmt.Sprintf("  *%s*\n", insight)
		}
	}

	elapsed := time.Since(start)
	log.Printf("[MCP] Tool completed: search_symbols (took %v, found %d matches)", elapsed, len(matches))
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: result}},
	}, nil, nil
}

func (s *CodeContextMCPServer) getDependencies(ctx context.Context, req *mcp.CallToolRequest, args GetDependenciesArgs) (*mcp.CallToolResult, any, error) {
	log.Printf("[MCP] Tool called: get_dependencies with args: %+v", args)
	start := time.Now()
	
	// Resolve target directory
	targetDir := s.resolveTargetDir(args.TargetDir)
	
	// Ensure we have fresh analysis
	log.Printf("[MCP] Refreshing analysis for dependency analysis...")
	if err := s.refreshAnalysisWithTargetDir(targetDir); err != nil {
		log.Printf("[MCP] ERROR: Failed to refresh analysis: %v", err)
		return nil, nil, fmt.Errorf("failed to refresh analysis: %w", err)
	}

	result := "# Dependency Analysis\n\n"
	log.Printf("[MCP] Analyzing %d edges for dependencies", len(s.graph.Edges))

	if args.FilePath != "" {
		// File-specific dependencies
		result += fmt.Sprintf("## Dependencies for: %s\n\n", args.FilePath)
		
		if args.Direction == "" || args.Direction == "imports" {
			result += "### Imports:\n"
			found := false
			for _, edge := range s.graph.Edges {
				if edge.Type == "imports" && edge.From == types.NodeId(args.FilePath) {
					result += fmt.Sprintf("- %s\n", edge.To)
					found = true
				}
			}
			if !found {
				result += "No imports found.\n"
			}
		}

		if args.Direction == "" || args.Direction == "dependents" {
			result += "\n### Dependents (files that import this):\n"
			found := false
			for _, edge := range s.graph.Edges {
				if edge.Type == "imports" && edge.To == types.NodeId(args.FilePath) {
					result += fmt.Sprintf("- %s\n", edge.From)
					found = true
				}
			}
			if !found {
				result += "No dependents found.\n"
			}
		}
	} else {
		// Global dependency overview
		result += "## Global Dependency Overview\n\n"
		
		fileCount := len(s.graph.Files)
		importCount := 0
		for _, edge := range s.graph.Edges {
			if edge.Type == "imports" {
				importCount++
			}
		}
		
		result += fmt.Sprintf("- **Total Files:** %d\n", fileCount)
		result += fmt.Sprintf("- **Total Import Relationships:** %d\n", importCount)
		
		// Most imported files
		dependentCounts := make(map[string]int)
		for _, edge := range s.graph.Edges {
			if edge.Type == "imports" {
				dependentCounts[string(edge.To)]++
			}
		}
		
		if len(dependentCounts) > 0 {
			result += "\n### Most Imported Files:\n"
			// Simple top 5 most imported
			count := 0
			for file, deps := range dependentCounts {
				if count >= 5 {
					break
				}
				result += fmt.Sprintf("- %s (%d imports)\n", file, deps)
				count++
			}
		}
	}

	elapsed := time.Since(start)
	log.Printf("[MCP] Tool completed: get_dependencies (took %v)", elapsed)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: result}},
	}, nil, nil
}

func (s *CodeContextMCPServer) watchChanges(ctx context.Context, req *mcp.CallToolRequest, args WatchChangesArgs) (*mcp.CallToolResult, any, error) {
	log.Printf("[MCP] Tool called: watch_changes with args: %+v", args)
	start := time.Now()
	
	// Check if server is being stopped
	s.stopMutex.RLock()
	if s.stopped {
		s.stopMutex.RUnlock()
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Server is shutting down, cannot process watch changes"}},
		}, nil, nil
	}
	s.stopMutex.RUnlock()
	
	if args.Enable {
		log.Printf("[MCP] Enabling file watching...")
		if s.watcher != nil {
			log.Printf("[MCP] File watching is already enabled")
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "File watching is already enabled"}},
			}, nil, nil
		}
		
		// Resolve target directory
		targetDir := s.resolveTargetDir(args.TargetDir)
		
		// Create watcher config
		config := watcher.Config{
			TargetDir:    targetDir,
			OutputFile:   "CLAUDE.md", // Not used in MCP mode
			DebounceTime: time.Duration(s.config.DebounceMs) * time.Millisecond,
			IncludeExts:  []string{".ts", ".tsx", ".js", ".jsx", ".go", ".py", ".java", ".cpp", ".c", ".rs"},
		}
		
		// Start file watcher
		log.Printf("[MCP] Creating file watcher with config: %+v", config)
		fileWatcher, err := watcher.NewFileWatcher(config)
		if err != nil {
			log.Printf("[MCP] ERROR: Failed to create file watcher: %v", err)
			return nil, nil, fmt.Errorf("failed to start file watcher: %w", err)
		}
		
		s.watcher = fileWatcher
		log.Printf("[MCP] File watcher created successfully")
		
		// Start watching in a goroutine
		watchCtx := context.Background()
		log.Printf("[MCP] Starting file watcher goroutine...")
		go func() {
			if err := fileWatcher.Start(watchCtx); err != nil {
				log.Printf("[MCP] ERROR: File watcher error: %v", err)
			}
		}()
		
		elapsed := time.Since(start)
		log.Printf("[MCP] Tool completed: watch_changes (enable) (took %v)", elapsed)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "File watching enabled. Real-time change notifications are now active."}},
		}, nil, nil
	} else {
		log.Printf("[MCP] Disabling file watching...")
		if s.watcher == nil {
			log.Printf("[MCP] File watching is not currently enabled")
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "File watching is not currently enabled"}},
			}, nil, nil
		}
		
		log.Printf("[MCP] Stopping file watcher...")
		s.watcher.Stop()
		s.watcher = nil
		log.Printf("[MCP] File watcher stopped")
		
		elapsed := time.Since(start)
		log.Printf("[MCP] Tool completed: watch_changes (disable) (took %v)", elapsed)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "File watching disabled"}},
		}, nil, nil
	}
}

func (s *CodeContextMCPServer) getSemanticNeighborhoods(ctx context.Context, req *mcp.CallToolRequest, args GetSemanticNeighborhoodsArgs) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	log.Printf("[MCP] Tool called: get_semantic_neighborhoods with args: %+v", args)

	// Resolve target directory
	targetDir := s.resolveTargetDir(args.TargetDir)

	// Ensure we have fresh analysis
	if s.graph == nil {
		if err := s.refreshAnalysisWithTargetDir(targetDir); err != nil {
			log.Printf("[MCP] Failed to refresh analysis: %v", err)
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "Failed to analyze codebase: " + err.Error()}},
			}, nil, nil
		}
	}

	// Get semantic neighborhoods from metadata
	semanticData, err := s.getSemanticNeighborhoodsData()
	if err != nil {
		log.Printf("[MCP] Failed to get semantic neighborhoods: %v", err)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Failed to get semantic neighborhoods: " + err.Error()}},
		}, nil, nil
	}

	// Build response based on arguments
	response := s.buildSemanticNeighborhoodsResponse(semanticData, args)

	elapsed := time.Since(start)
	log.Printf("[MCP] Tool completed: get_semantic_neighborhoods (took %v)", elapsed)
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: response}},
	}, nil, nil
}

// Helper methods

func (s *CodeContextMCPServer) refreshAnalysis() error {
	return s.refreshAnalysisWithTargetDir(s.config.TargetDir)
}

func (s *CodeContextMCPServer) refreshAnalysisWithTargetDir(targetDir string) error {
	log.Printf("[MCP] Starting analysis of directory: %s", targetDir)
	graph, err := s.analyzer.AnalyzeDirectory(targetDir)
	if err != nil {
		log.Printf("[MCP] Analysis failed: %v", err)
		return err
	}
	log.Printf("[MCP] Analysis completed successfully - %d files, %d symbols", len(graph.Files), len(graph.Symbols))
	s.graph = graph
	return nil
}

func (s *CodeContextMCPServer) resolveTargetDir(targetDir string) string {
	if targetDir != "" {
		return expandPath(targetDir)
	}
	return s.config.TargetDir
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

// getSemanticNeighborhoodsData extracts semantic neighborhoods from the graph metadata
func (s *CodeContextMCPServer) getSemanticNeighborhoodsData() (*analyzer.SemanticAnalysisResult, error) {
	if s.graph == nil || s.graph.Metadata == nil || s.graph.Metadata.Configuration == nil {
		return nil, fmt.Errorf("no graph metadata available")
	}

	semanticInterface, exists := s.graph.Metadata.Configuration["semantic_neighborhoods"]
	if !exists {
		return nil, fmt.Errorf("no semantic neighborhoods data found - ensure this is a git repository")
	}

	semanticResult, ok := semanticInterface.(*analyzer.SemanticAnalysisResult)
	if !ok {
		return nil, fmt.Errorf("invalid semantic neighborhoods data format")
	}

	return semanticResult, nil
}

// buildSemanticNeighborhoodsResponse builds the response string for semantic neighborhoods
func (s *CodeContextMCPServer) buildSemanticNeighborhoodsResponse(data *analyzer.SemanticAnalysisResult, args GetSemanticNeighborhoodsArgs) string {
	var response strings.Builder
	
	response.WriteString("# Semantic Code Neighborhoods Analysis\n\n")
	
	// Check if git repository
	if !data.AnalysisMetadata.IsGitRepository {
		response.WriteString("❌ **Not a Git Repository**: This directory is not a git repository. Semantic neighborhoods require git history for pattern analysis.\n")
		return response.String()
	}
	
	// Handle errors
	if data.Error != "" {
		response.WriteString(fmt.Sprintf("⚠️ **Analysis Error**: %s\n\n", data.Error))
	}
	
	// Analysis overview
	metadata := data.AnalysisMetadata
	response.WriteString("## 📊 Analysis Overview\n\n")
	response.WriteString("**Git-based pattern analysis with hierarchical clustering:**\n\n")
	response.WriteString(fmt.Sprintf("- **Analysis Period**: %d days\n", metadata.AnalysisPeriodDays))
	response.WriteString(fmt.Sprintf("- **Files with Patterns**: %d files\n", metadata.FilesWithPatterns))
	response.WriteString(fmt.Sprintf("- **Basic Neighborhoods**: %d groups\n", metadata.TotalNeighborhoods))
	response.WriteString(fmt.Sprintf("- **Clustered Groups**: %d clusters\n", metadata.TotalClusters))
	response.WriteString(fmt.Sprintf("- **Average Cluster Size**: %.1f files\n", metadata.AverageClusterSize))
	response.WriteString(fmt.Sprintf("- **Analysis Time**: %v\n", metadata.AnalysisTime))
	
	if metadata.QualityScores.OverallQualityRating != "" {
		response.WriteString(fmt.Sprintf("- **Clustering Quality**: %s\n", metadata.QualityScores.OverallQualityRating))
	}
	response.WriteString("\n")
	
	// Context recommendations based on file path
	if args.FilePath != "" {
		response.WriteString(s.buildFileContextRecommendations(data, args.FilePath))
	}
	
	// Basic neighborhoods (if requested)
	if args.IncludeBasic && len(data.SemanticNeighborhoods) > 0 {
		response.WriteString("## 🔍 Basic Semantic Neighborhoods\n\n")
		response.WriteString(s.buildBasicNeighborhoodsResponse(data.SemanticNeighborhoods, args.MaxResults))
	}
	
	// Clustered neighborhoods (always include if available)
	if len(data.ClusteredNeighborhoods) > 0 {
		response.WriteString("## 🎯 Clustered Neighborhoods\n\n")
		response.WriteString(s.buildClusteredNeighborhoodsResponse(data.ClusteredNeighborhoods, args.MaxResults))
	}
	
	// Quality metrics (if requested)
	if args.IncludeQuality && len(data.ClusteredNeighborhoods) > 0 {
		response.WriteString("## 📈 Quality Metrics\n\n")
		response.WriteString(s.buildQualityMetricsResponse(data))
	}
	
	// No neighborhoods found
	if len(data.SemanticNeighborhoods) == 0 && len(data.ClusteredNeighborhoods) == 0 {
		response.WriteString("## 🏷️ No Neighborhoods Found\n\n")
		response.WriteString("No semantic neighborhoods were detected. This could mean:\n")
		response.WriteString("- Files don't frequently change together\n")
		response.WriteString("- Insufficient git history (need at least a few commits)\n")
		response.WriteString("- Repository primarily contains single-purpose files\n")
		response.WriteString("- Analysis period too short (default: 30 days)\n")
	}
	
	return response.String()
}

// buildFileContextRecommendations builds context recommendations for a specific file
func (s *CodeContextMCPServer) buildFileContextRecommendations(data *analyzer.SemanticAnalysisResult, filePath string) string {
	var response strings.Builder
	
	response.WriteString(fmt.Sprintf("## 🎯 Context Recommendations for `%s`\n\n", filePath))
	
	// Find neighborhoods containing this file
	relatedNeighborhoods := []string{}
	relatedClusters := []string{}
	
	// Check basic neighborhoods
	for _, neighborhood := range data.SemanticNeighborhoods {
		for _, file := range neighborhood.Files {
			if strings.Contains(file, filePath) || strings.Contains(filePath, file) {
				relatedNeighborhoods = append(relatedNeighborhoods, neighborhood.Name)
				break
			}
		}
	}
	
	// Check clustered neighborhoods
	for i, clustered := range data.ClusteredNeighborhoods {
		for _, neighborhood := range clustered.Neighborhoods {
			for _, file := range neighborhood.Files {
				if strings.Contains(file, filePath) || strings.Contains(filePath, file) {
					relatedClusters = append(relatedClusters, fmt.Sprintf("Cluster %d: %s", i+1, clustered.Cluster.Name))
					break
				}
			}
		}
	}
	
	if len(relatedNeighborhoods) > 0 {
		response.WriteString("**Related Neighborhoods:**\n")
		for _, neighborhood := range relatedNeighborhoods {
			response.WriteString(fmt.Sprintf("- %s\n", neighborhood))
		}
		response.WriteString("\n")
	}
	
	if len(relatedClusters) > 0 {
		response.WriteString("**Related Clusters:**\n")
		for _, cluster := range relatedClusters {
			response.WriteString(fmt.Sprintf("- %s\n", cluster))
		}
		response.WriteString("\n")
	}
	
	if len(relatedNeighborhoods) == 0 && len(relatedClusters) == 0 {
		response.WriteString("**No direct relationships found.** This file may be independent or have weak patterns with other files.\n\n")
	}
	
	return response.String()
}

// buildBasicNeighborhoodsResponse builds the basic neighborhoods response
func (s *CodeContextMCPServer) buildBasicNeighborhoodsResponse(neighborhoods []git.SemanticNeighborhood, maxResults int) string {
	var response strings.Builder
	
	// Sort by correlation strength
	sortedNeighborhoods := make([]git.SemanticNeighborhood, len(neighborhoods))
	copy(sortedNeighborhoods, neighborhoods)
	
	limit := len(sortedNeighborhoods)
	if maxResults > 0 && maxResults < limit {
		limit = maxResults
	}
	
	for i := 0; i < limit; i++ {
		neighborhood := sortedNeighborhoods[i]
		response.WriteString(fmt.Sprintf("### %s\n\n", neighborhood.Name))
		response.WriteString(fmt.Sprintf("- **Correlation**: %.2f\n", neighborhood.CorrelationStrength))
		response.WriteString(fmt.Sprintf("- **Changes**: %d\n", neighborhood.ChangeFrequency))
		response.WriteString(fmt.Sprintf("- **Files**: %d\n", len(neighborhood.Files)))
		response.WriteString(fmt.Sprintf("- **Last Changed**: %s\n", neighborhood.LastChanged.Format("2006-01-02")))
		
		if len(neighborhood.Files) > 0 {
			response.WriteString("\n**Files:**\n")
			for _, file := range neighborhood.Files {
				response.WriteString(fmt.Sprintf("- `%s`\n", file))
			}
		}
		response.WriteString("\n")
	}
	
	return response.String()
}

// buildClusteredNeighborhoodsResponse builds the clustered neighborhoods response
func (s *CodeContextMCPServer) buildClusteredNeighborhoodsResponse(clusteredNeighborhoods []git.ClusteredNeighborhood, maxResults int) string {
	var response strings.Builder
	
	limit := len(clusteredNeighborhoods)
	if maxResults > 0 && maxResults < limit {
		limit = maxResults
	}
	
	for i := 0; i < limit; i++ {
		clustered := clusteredNeighborhoods[i]
		cluster := clustered.Cluster
		
		response.WriteString(fmt.Sprintf("### Cluster %d: %s\n\n", i+1, cluster.Name))
		response.WriteString(fmt.Sprintf("- **Description**: %s\n", cluster.Description))
		response.WriteString(fmt.Sprintf("- **Size**: %d files\n", cluster.Size))
		response.WriteString(fmt.Sprintf("- **Strength**: %.3f\n", cluster.Strength))
		response.WriteString(fmt.Sprintf("- **Silhouette Score**: %.3f\n", clustered.QualityMetrics.SilhouetteScore))
		response.WriteString(fmt.Sprintf("- **Cohesion**: %.3f\n", cluster.IntraMetrics.Cohesion))
		
		if len(cluster.OptimalTasks) > 0 {
			response.WriteString("\n**Recommended Tasks:**\n")
			for _, task := range cluster.OptimalTasks {
				response.WriteString(fmt.Sprintf("- %s\n", task))
			}
		}
		
		if cluster.RecommendationReason != "" {
			response.WriteString(fmt.Sprintf("\n**Why**: %s\n", cluster.RecommendationReason))
		}
		
		response.WriteString("\n")
	}
	
	return response.String()
}

// buildQualityMetricsResponse builds the quality metrics response
func (s *CodeContextMCPServer) buildQualityMetricsResponse(data *analyzer.SemanticAnalysisResult) string {
	var response strings.Builder
	
	scores := data.AnalysisMetadata.QualityScores
	
	response.WriteString("**Overall Clustering Performance:**\n\n")
	response.WriteString(fmt.Sprintf("- **Average Silhouette Score**: %.3f\n", scores.AverageSilhouetteScore))
	response.WriteString(fmt.Sprintf("- **Average Davies-Bouldin Index**: %.3f\n", scores.AverageDaviesBouldinIndex))
	response.WriteString(fmt.Sprintf("- **Quality Rating**: %s\n\n", scores.OverallQualityRating))
	
	response.WriteString("**Interpretation:**\n")
	response.WriteString("- **Silhouette Score**: 0.7+ Excellent, 0.5+ Good, 0.25+ Fair, <0.25 Poor\n")
	response.WriteString("- **Davies-Bouldin**: Lower values indicate better clustering\n")
	response.WriteString("- **Algorithm**: Hierarchical clustering with Ward linkage\n")
	
	return response.String()
}

// Run starts the MCP server
func (s *CodeContextMCPServer) Run(ctx context.Context) error {
	log.Printf("[MCP] CodeContext MCP Server starting - will analyze %s", s.config.TargetDir)
	
	// Initial analysis
	if err := s.refreshAnalysis(); err != nil {
		log.Printf("[MCP] Initial analysis failed, server will not start: %v", err)
		return fmt.Errorf("failed to perform initial analysis: %w", err)
	}
	
	log.Printf("[MCP] CodeContext MCP Server ready - analysis complete")
	
	// Run the MCP server with stdio transport
	return s.server.Run(ctx, mcp.NewStdioTransport())
}

// Stop gracefully stops the MCP server
func (s *CodeContextMCPServer) Stop() {
	log.Printf("[MCP] Stopping MCP server...")
	
	// Set stopped flag to prevent new operations
	s.stopMutex.Lock()
	s.stopped = true
	s.stopMutex.Unlock()
	
	if s.watcher != nil {
		log.Printf("[MCP] Stopping file watcher...")
		s.watcher.Stop()
		s.watcher = nil
		log.Printf("[MCP] File watcher stopped")
	}
	log.Printf("[MCP] MCP server stopped successfully")
}

// Framework-specific helper functions

// getFrameworkSpecificDescription returns a description for framework-specific symbol types
func (s *CodeContextMCPServer) getFrameworkSpecificDescription(symbolType string) string {
	switch symbolType {
	case "component":
		return "**Description:** A reusable UI component that encapsulates functionality and presentation.\n"
	case "hook":
		return "**Description:** A React hook that provides stateful logic and side effects.\n"
	case "service":
		return "**Description:** An Angular service that provides shared functionality and data.\n"
	case "directive":
		return "**Description:** An Angular directive that extends HTML with custom behavior.\n"
	case "store":
		return "**Description:** A state management store for centralized application state.\n"
	case "computed":
		return "**Description:** A Vue computed property that derives data reactively.\n"
	case "watcher":
		return "**Description:** A Vue watcher that observes data changes and reacts accordingly.\n"
	case "route":
		return "**Description:** A Next.js route handler for page or API endpoint.\n"
	case "middleware":
		return "**Description:** Next.js middleware that runs before request completion.\n"
	case "action":
		return "**Description:** A Svelte action that adds behavior to DOM elements.\n"
	case "lifecycle":
		return "**Description:** A framework lifecycle method that handles component state changes.\n"
	default:
		return ""
	}
}

// getFrameworkInsights provides framework-specific insights for symbols
func (s *CodeContextMCPServer) getFrameworkInsights(symbol *types.Symbol) string {
	switch string(symbol.Type) {
	case "component":
		return "Consider: Props interface, state management, performance optimization"
	case "hook":
		return "Consider: Dependencies array, cleanup functions, memoization"
	case "service":
		return "Consider: Dependency injection, singleton pattern, testing"
	case "store":
		return "Consider: State mutations, subscriptions, persistence"
	case "route":
		filePath := s.getFilePathForSymbol(symbol)
		if strings.Contains(filePath, "/api/") {
			return "API Route: Consider request validation, error handling, response types"
		}
		return "Page Route: Consider SEO, data fetching, loading states"
	default:
		return ""
	}
}

// matchesFramework checks if a symbol matches a specific framework
func (s *CodeContextMCPServer) matchesFramework(symbol *types.Symbol, framework string) bool {
	// Get file classification to determine framework
	if s.graph != nil && s.graph.Files != nil {
		filePath := s.getFilePathForSymbol(symbol)
		if _, exists := s.graph.Files[filePath]; exists {
			// Check if file has framework metadata
			// For now, do a simple string match on framework types
			symbolType := string(symbol.Type)
			switch strings.ToLower(framework) {
			case "react":
				return symbolType == "component" || symbolType == "hook" || 
					   strings.Contains(filePath, ".jsx") || 
					   strings.Contains(filePath, ".tsx")
			case "vue":
				return symbolType == "component" || symbolType == "computed" || 
					   symbolType == "watcher" || strings.Contains(filePath, ".vue")
			case "angular":
				return symbolType == "component" || symbolType == "service" || 
					   symbolType == "directive" || strings.Contains(filePath, ".component.")
			case "svelte":
				return symbolType == "component" || symbolType == "store" || 
					   symbolType == "action" || strings.Contains(filePath, ".svelte")
			case "nextjs", "next.js":
				return symbolType == "route" || symbolType == "middleware" ||
					   strings.Contains(filePath, "/pages/") ||
					   strings.Contains(filePath, "/app/")
			}
		}
	}
	return false
}

// getFrameworkAnalysis provides comprehensive framework-specific analysis
func (s *CodeContextMCPServer) getFrameworkAnalysis(ctx context.Context, req *mcp.CallToolRequest, args GetFrameworkAnalysisArgs) (*mcp.CallToolResult, any, error) {

	// Resolve target directory
	targetDir := s.resolveTargetDir(args.TargetDir)

	// Ensure we have fresh analysis
	if err := s.refreshAnalysisWithTargetDir(targetDir); err != nil {
		return nil, nil, fmt.Errorf("failed to refresh analysis: %w", err)
	}

	if s.graph == nil {
		return nil, nil, fmt.Errorf("no graph available - ensure analysis has been performed")
	}

	// Get all framework-specific symbols
	frameworkSymbols := make(map[string][]*types.Symbol)
	frameworkCounts := make(map[string]map[string]int)
	
	for _, symbol := range s.graph.Symbols {
		if symbol.Type == types.SymbolTypeComponent || 
		   symbol.Type == types.SymbolTypeHook || 
		   symbol.Type == types.SymbolTypeDirective || 
		   symbol.Type == types.SymbolTypeService || 
		   symbol.Type == types.SymbolTypeStore || 
		   symbol.Type == types.SymbolTypeComputed || 
		   symbol.Type == types.SymbolTypeWatcher || 
		   symbol.Type == types.SymbolTypeLifecycle || 
		   symbol.Type == types.SymbolTypeRoute || 
		   symbol.Type == types.SymbolTypeMiddleware || 
		   symbol.Type == types.SymbolTypeAction {
			
			// Determine framework from file classification
			filePath := s.getFilePathForSymbol(symbol)
			framework := s.getFrameworkForFile(filePath)
			if framework == "" {
				framework = "Unknown"
			}
			
			// Filter by requested framework if specified
			if args.Framework != "" && !strings.EqualFold(framework, args.Framework) {
				continue
			}
			
			frameworkSymbols[framework] = append(frameworkSymbols[framework], symbol)
			
			if frameworkCounts[framework] == nil {
				frameworkCounts[framework] = make(map[string]int)
			}
			frameworkCounts[framework][string(symbol.Type)]++
		}
	}

	response := s.buildFrameworkAnalysisResponse(frameworkSymbols, frameworkCounts, args)

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: response}},
	}, nil, nil
}

// getFrameworkForFile determines the framework for a given file path
func (s *CodeContextMCPServer) getFrameworkForFile(filePath string) string {
	// Check if we have file classification data
	for _, file := range s.graph.Files {
		if file.Path == filePath {
			// Try to get framework from metadata or file patterns
			if strings.Contains(filePath, ".vue") {
				return "Vue"
			} else if strings.Contains(filePath, ".svelte") {
				return "Svelte"
			} else if strings.Contains(filePath, ".astro") {
				return "Astro"
			} else if strings.Contains(filePath, ".component.") {
				return "Angular"
			} else if strings.Contains(filePath, ".jsx") || strings.Contains(filePath, ".tsx") {
				return "React"
			} else if strings.Contains(filePath, "/pages/") || strings.Contains(filePath, "/app/") {
				return "Next.js"
			}
		}
	}
	
	// Fallback to basic pattern matching
	if strings.Contains(filePath, ".vue") {
		return "Vue"
	} else if strings.Contains(filePath, ".svelte") {
		return "Svelte"
	} else if strings.Contains(filePath, ".astro") {
		return "Astro"
	} else if strings.Contains(filePath, ".component.") {
		return "Angular"
	} else if strings.Contains(filePath, ".jsx") || strings.Contains(filePath, ".tsx") {
		return "React"
	} else if strings.Contains(filePath, "/pages/") || strings.Contains(filePath, "/app/") {
		return "Next.js"
	}
	
	return ""
}

// buildFrameworkAnalysisResponse builds the comprehensive framework analysis response
func (s *CodeContextMCPServer) buildFrameworkAnalysisResponse(frameworkSymbols map[string][]*types.Symbol, frameworkCounts map[string]map[string]int, args GetFrameworkAnalysisArgs) string {
	var response strings.Builder
	
	response.WriteString("# 🚀 Framework Analysis Report\n\n")
	
	if args.Framework != "" {
		response.WriteString(fmt.Sprintf("**Focused Analysis for: %s**\n\n", args.Framework))
	} else {
		response.WriteString("**Comprehensive Multi-Framework Analysis**\n\n")
	}
	
	if len(frameworkSymbols) == 0 {
		response.WriteString("❌ **No framework-specific symbols found**\n")
		response.WriteString("This codebase doesn't appear to use any detected frameworks, or symbols haven't been properly extracted.\n")
		return response.String()
	}
	
	// Overview statistics
	if args.IncludeStats {
		response.WriteString("## 📊 Framework Overview\n\n")
		totalSymbols := 0
		for framework, symbols := range frameworkSymbols {
			count := len(symbols)
			totalSymbols += count
			response.WriteString(fmt.Sprintf("- **%s**: %d symbols\n", framework, count))
		}
		response.WriteString(fmt.Sprintf("\n**Total Framework Symbols**: %d\n\n", totalSymbols))
	}
	
	// Detailed framework analysis
	for framework, symbols := range frameworkSymbols {
		response.WriteString(fmt.Sprintf("## 🎯 %s Framework Analysis\n\n", framework))
		
		// Symbol type breakdown
		counts := frameworkCounts[framework]
		response.WriteString("### Symbol Distribution\n\n")
		for symbolType, count := range counts {
			emoji := s.getSymbolTypeEmoji(symbolType)
			response.WriteString(fmt.Sprintf("- %s **%s**: %d\n", emoji, symbolType, count))
		}
		response.WriteString("\n")
		
		// Framework-specific insights
		insights := s.getFrameworkAnalysisInsights(framework, symbols, counts)
		if insights != "" {
			response.WriteString("### 💡 Framework Insights\n\n")
			response.WriteString(insights)
			response.WriteString("\n")
		}
		
		// Key symbols (top 5 by name)
		response.WriteString("### 🔑 Key Symbols\n\n")
		for i, symbol := range symbols {
			if i >= 5 { // Limit to top 5
				break
			}
			emoji := s.getSymbolTypeEmoji(string(symbol.Type))
			filePath := s.getFilePathForSymbol(symbol)
			location := fmt.Sprintf("%s:%d", filePath, symbol.Location.StartLine)
			response.WriteString(fmt.Sprintf("- %s **%s** (`%s`) - %s\n", emoji, symbol.Name, symbol.Type, location))
		}
		response.WriteString("\n")
	}
	
	// Cross-framework recommendations
	if len(frameworkSymbols) > 1 {
		response.WriteString("## 🔄 Multi-Framework Observations\n\n")
		response.WriteString("This codebase uses multiple frameworks. Consider:\n")
		response.WriteString("- **Consistency**: Ensure similar patterns across frameworks\n")
		response.WriteString("- **Separation**: Keep framework-specific code in separate modules\n")
		response.WriteString("- **Shared utilities**: Extract common logic to framework-agnostic utilities\n\n")
	}
	
	return response.String()
}

// getFrameworkAnalysisInsights provides framework-specific insights based on symbol analysis
func (s *CodeContextMCPServer) getFrameworkAnalysisInsights(framework string, symbols []*types.Symbol, counts map[string]int) string {
	var insights strings.Builder
	
	switch strings.ToLower(framework) {
	case "react":
		componentCount := counts["component"]
		hookCount := counts["hook"]
		if componentCount > 0 && hookCount > 0 {
			ratio := float64(hookCount) / float64(componentCount)
			if ratio > 0.5 {
				insights.WriteString("✅ **Good hook usage**: High hook-to-component ratio suggests good state logic separation\n")
			} else {
				insights.WriteString("💡 **Consider more hooks**: Low hook-to-component ratio - consider extracting stateful logic\n")
			}
		}
		if componentCount > 10 {
			insights.WriteString("📦 **Large codebase**: Consider component composition and code splitting\n")
		}
		
	case "vue":
		componentCount := counts["component"]
		computedCount := counts["computed"]
		if computedCount > 0 {
			insights.WriteString("✅ **Good reactive patterns**: Using computed properties for derived state\n")
		}
		if componentCount > computedCount*2 {
			insights.WriteString("💡 **Consider computed properties**: Many components without computed properties\n")
		}
		
	case "angular":
		componentCount := counts["component"]
		serviceCount := counts["service"]
		if serviceCount > 0 {
			ratio := float64(serviceCount) / float64(componentCount)
			if ratio > 0.3 {
				insights.WriteString("✅ **Good service usage**: Good separation of concerns with services\n")
			} else {
				insights.WriteString("💡 **Consider more services**: Extract business logic into services\n")
			}
		}
		
	case "svelte":
		componentCount := counts["component"]
		storeCount := counts["store"]
		if storeCount > 0 {
			insights.WriteString("✅ **Using stores**: Good global state management with Svelte stores\n")
		}
		if componentCount > 5 && storeCount == 0 {
			insights.WriteString("💡 **Consider stores**: Large component count without stores - consider global state management\n")
		}
		
	case "next.js":
		routeCount := counts["route"]
		middlewareCount := counts["middleware"]
		if middlewareCount > 0 {
			insights.WriteString("✅ **Using middleware**: Good request processing patterns\n")
		}
		if routeCount > 20 {
			insights.WriteString("📊 **Large application**: Consider route organization and lazy loading\n")
		}
	}
	
	return insights.String()
}

// getFilePathForSymbol finds the file path for a given symbol
func (s *CodeContextMCPServer) getFilePathForSymbol(symbol *types.Symbol) string {
	// Look through all files to find which one contains this symbol
	for filePath, fileNode := range s.graph.Files {
		for _, symbolId := range fileNode.Symbols {
			if symbolId == symbol.Id {
				return filePath
			}
		}
	}
	return ""
}

// getSymbolTypeEmoji returns an emoji for each symbol type
func (s *CodeContextMCPServer) getSymbolTypeEmoji(symbolType string) string {
	switch symbolType {
	case "component":
		return "🧩"
	case "hook":
		return "🪝"
	case "directive":
		return "📋"
	case "service":
		return "⚙️"
	case "store":
		return "🗄️"
	case "computed":
		return "🧮"
	case "watcher":
		return "👁️"
	case "lifecycle":
		return "🔄"
	case "route":
		return "🛣️"
	case "middleware":
		return "🔀"
	case "action":
		return "⚡"
	default:
		return "📦"
	}
}
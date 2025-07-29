package compact

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/nuthan-ms/codecontext/pkg/types"
)

// CompactController manages context optimization using various strategies
type CompactController struct {
	strategies map[string]Strategy
	config     *CompactConfig
	metrics    *CompactMetrics
	metricsMux sync.RWMutex
}

// CompactConfig holds configuration for the compact controller
type CompactConfig struct {
	EnableCompaction   bool                   `json:"enable_compaction"`
	DefaultStrategy    string                 `json:"default_strategy"`
	MaxContextSize     int                    `json:"max_context_size"`
	CompressionRatio   float64                `json:"compression_ratio"`
	PriorityThreshold  float64                `json:"priority_threshold"`
	CacheEnabled       bool                   `json:"cache_enabled"`
	CacheSize          int                    `json:"cache_size"`
	MetricsEnabled     bool                   `json:"metrics_enabled"`
	StrategyConfig     map[string]interface{} `json:"strategy_config"`
	AdaptiveEnabled    bool                   `json:"adaptive_enabled"`
	AdaptiveThreshold  float64                `json:"adaptive_threshold"`
	BatchSize          int                    `json:"batch_size"`
	ParallelProcessing bool                   `json:"parallel_processing"`
}

// CompactMetrics tracks compaction performance
type CompactMetrics struct {
	TotalCompactions   int64            `json:"total_compactions"`
	CompressionRatio   float64          `json:"compression_ratio"`
	AverageTime        time.Duration    `json:"average_time"`
	LastCompaction     time.Time        `json:"last_compaction"`
	StrategiesUsed     map[string]int64 `json:"strategies_used"`
	CacheHitRate       float64          `json:"cache_hit_rate"`
	MemorySaved        int64            `json:"memory_saved"`
	ContextSizeReduced int64            `json:"context_size_reduced"`
	AdaptiveTriggers   int64            `json:"adaptive_triggers"`
}

// CompactRequest represents a request for context optimization
type CompactRequest struct {
	Graph        *types.CodeGraph       `json:"graph"`
	Strategy     string                 `json:"strategy"`
	MaxSize      int                    `json:"max_size"`
	Priorities   map[string]float64     `json:"priorities"`
	Requirements *CompactRequirements   `json:"requirements"`
	Context      map[string]interface{} `json:"context"`
}

// CompactRequirements specifies what must be preserved during compaction
type CompactRequirements struct {
	PreserveFiles   []string           `json:"preserve_files"`
	PreserveSymbols []types.SymbolId   `json:"preserve_symbols"`
	MinDepth        int                `json:"min_depth"`
	RequiredTypes   []types.SymbolType `json:"required_types"`
	LanguageFilter  []string           `json:"language_filter"`
	PreservePaths   []string           `json:"preserve_paths"`
}

// CompactResult contains the result of context optimization
type CompactResult struct {
	CompactedGraph   *types.CodeGraph       `json:"compacted_graph"`
	OriginalSize     int                    `json:"original_size"`
	CompactedSize    int                    `json:"compacted_size"`
	CompressionRatio float64                `json:"compression_ratio"`
	Strategy         string                 `json:"strategy"`
	ExecutionTime    time.Duration          `json:"execution_time"`
	RemovedItems     *RemovedItems          `json:"removed_items"`
	Metadata         map[string]interface{} `json:"metadata"`
	Warnings         []string               `json:"warnings"`
}

// RemovedItems tracks what was removed during compaction
type RemovedItems struct {
	Files   []string         `json:"files"`
	Symbols []types.SymbolId `json:"symbols"`
	Edges   []types.EdgeId   `json:"edges"`
	Nodes   []types.NodeId   `json:"nodes"`
	Reason  string           `json:"reason"`
	Impact  *ImpactAnalysis  `json:"impact"`
}

// ImpactAnalysis analyzes the impact of removing items
type ImpactAnalysis struct {
	DependentFiles   []string `json:"dependent_files"`
	BrokenReferences int      `json:"broken_references"`
	IsolatedSymbols  int      `json:"isolated_symbols"`
	RiskLevel        string   `json:"risk_level"`
	Recommendations  []string `json:"recommendations"`
}

// NewCompactController creates a new compact controller
func NewCompactController(config *CompactConfig) *CompactController {
	if config == nil {
		config = DefaultCompactConfig()
	}

	controller := &CompactController{
		strategies: make(map[string]Strategy),
		config:     config,
		metrics: &CompactMetrics{
			StrategiesUsed: make(map[string]int64),
		},
	}

	// Register default strategies
	controller.RegisterStrategy("relevance", NewRelevanceStrategy())
	controller.RegisterStrategy("frequency", NewFrequencyStrategy())
	controller.RegisterStrategy("dependency", NewDependencyStrategy())
	controller.RegisterStrategy("size", NewSizeStrategy())
	controller.RegisterStrategy("hybrid", NewHybridStrategy())
	controller.RegisterStrategy("adaptive", NewAdaptiveStrategy())

	return controller
}

// DefaultCompactConfig returns default configuration
func DefaultCompactConfig() *CompactConfig {
	return &CompactConfig{
		EnableCompaction:   true,
		DefaultStrategy:    "hybrid",
		MaxContextSize:     10000,
		CompressionRatio:   0.7,
		PriorityThreshold:  0.5,
		CacheEnabled:       true,
		CacheSize:          100,
		MetricsEnabled:     true,
		StrategyConfig:     make(map[string]interface{}),
		AdaptiveEnabled:    true,
		AdaptiveThreshold:  0.8,
		BatchSize:          50,
		ParallelProcessing: true,
	}
}

// RegisterStrategy registers a new compaction strategy
func (cc *CompactController) RegisterStrategy(name string, strategy Strategy) {
	cc.strategies[name] = strategy
}

// GetStrategy retrieves a strategy by name
func (cc *CompactController) GetStrategy(name string) (Strategy, bool) {
	strategy, exists := cc.strategies[name]
	return strategy, exists
}

// ListStrategies returns all registered strategy names
func (cc *CompactController) ListStrategies() []string {
	names := make([]string, 0, len(cc.strategies))
	for name := range cc.strategies {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Compact performs context optimization using the specified strategy
func (cc *CompactController) Compact(ctx context.Context, request *CompactRequest) (*CompactResult, error) {
	if !cc.config.EnableCompaction {
		return cc.noCompactionResult(request.Graph), nil
	}

	start := time.Now()

	// Validate request
	if err := cc.validateRequest(request); err != nil {
		return nil, fmt.Errorf("invalid compact request: %w", err)
	}

	// Select strategy
	strategyName := request.Strategy
	if strategyName == "" {
		strategyName = cc.config.DefaultStrategy
	}

	// Check if strategy exists first
	strategy, exists := cc.strategies[strategyName]
	if !exists {
		return nil, fmt.Errorf("unknown strategy: %s", strategyName)
	}

	// Apply adaptive strategy selection if enabled and strategy is empty or adaptive
	if cc.config.AdaptiveEnabled && (request.Strategy == "" || request.Strategy == "adaptive") {
		adaptiveStrategy := cc.selectAdaptiveStrategy(request)
		if adaptiveStrategy != "" {
			strategyName = adaptiveStrategy
			strategy = cc.strategies[strategyName]
			cc.metricsMux.Lock()
			cc.metrics.AdaptiveTriggers++
			cc.metricsMux.Unlock()
		}
	}

	// Track strategy usage
	cc.metricsMux.Lock()
	cc.metrics.StrategiesUsed[strategyName]++
	cc.metricsMux.Unlock()

	// Perform compaction
	result, err := strategy.Compact(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("compaction failed with strategy %s: %w", strategyName, err)
	}

	// Update result metadata
	result.Strategy = strategyName
	result.ExecutionTime = time.Since(start)

	// Add adaptive metadata if adaptive selection was used
	if cc.config.AdaptiveEnabled && (request.Strategy == "" || request.Strategy == "adaptive") && strategyName != request.Strategy {
		if result.Metadata == nil {
			result.Metadata = make(map[string]interface{})
		}
		result.Metadata["adaptive_choice"] = strategyName
	}

	// Calculate compression ratio
	originalSize := cc.calculateGraphSize(request.Graph)
	compactedSize := cc.calculateGraphSize(result.CompactedGraph)
	result.OriginalSize = originalSize
	result.CompactedSize = compactedSize
	if originalSize > 0 {
		result.CompressionRatio = float64(compactedSize) / float64(originalSize)
	}

	// Update metrics
	cc.updateMetrics(result)

	return result, nil
}

// CompactMultiple performs batch compaction using different strategies
func (cc *CompactController) CompactMultiple(ctx context.Context, requests []*CompactRequest) ([]*CompactResult, error) {
	results := make([]*CompactResult, len(requests))

	if cc.config.ParallelProcessing && len(requests) > 1 {
		// Process in parallel
		type resultWithIndex struct {
			result *CompactResult
			err    error
			index  int
		}

		resultChan := make(chan resultWithIndex, len(requests))

		for i, request := range requests {
			go func(idx int, req *CompactRequest) {
				result, err := cc.Compact(ctx, req)
				resultChan <- resultWithIndex{result, err, idx}
			}(i, request)
		}

		// Collect results
		for i := 0; i < len(requests); i++ {
			res := <-resultChan
			if res.err != nil {
				return nil, fmt.Errorf("batch compaction failed at index %d: %w", res.index, res.err)
			}
			results[res.index] = res.result
		}
	} else {
		// Process sequentially
		for i, request := range requests {
			result, err := cc.Compact(ctx, request)
			if err != nil {
				return nil, fmt.Errorf("batch compaction failed at index %d: %w", i, err)
			}
			results[i] = result
		}
	}

	return results, nil
}

// AnalyzeCompactionPotential analyzes how much a graph could be compacted
func (cc *CompactController) AnalyzeCompactionPotential(graph *types.CodeGraph) *CompactionAnalysis {
	analysis := &CompactionAnalysis{
		TotalFiles:   len(graph.Files),
		TotalSymbols: len(graph.Symbols),
		TotalNodes:   len(graph.Nodes),
		TotalEdges:   len(graph.Edges),
		Strategies:   make(map[string]*StrategyAnalysis),
	}

	// Analyze each strategy
	for name, strategy := range cc.strategies {
		if analyzer, ok := strategy.(StrategyAnalyzer); ok {
			strategyAnalysis := analyzer.AnalyzePotential(graph)
			analysis.Strategies[name] = strategyAnalysis
		}
	}

	// Calculate overall potential
	analysis.calculateOverallPotential()

	return analysis
}

// GetMetrics returns current compaction metrics
func (cc *CompactController) GetMetrics() *CompactMetrics {
	cc.metricsMux.RLock()
	defer cc.metricsMux.RUnlock()
	
	// Return a copy to avoid race conditions
	metricsCopy := *cc.metrics
	metricsCopy.StrategiesUsed = make(map[string]int64)
	for k, v := range cc.metrics.StrategiesUsed {
		metricsCopy.StrategiesUsed[k] = v
	}
	return &metricsCopy
}

// ResetMetrics resets all metrics
func (cc *CompactController) ResetMetrics() {
	cc.metricsMux.Lock()
	defer cc.metricsMux.Unlock()
	
	cc.metrics = &CompactMetrics{
		StrategiesUsed: make(map[string]int64),
	}
}

// Helper methods

func (cc *CompactController) validateRequest(request *CompactRequest) error {
	if request.Graph == nil {
		return fmt.Errorf("graph cannot be nil")
	}

	if request.MaxSize <= 0 {
		request.MaxSize = cc.config.MaxContextSize
	}

	if request.Requirements == nil {
		request.Requirements = &CompactRequirements{}
	}

	return nil
}

func (cc *CompactController) selectAdaptiveStrategy(request *CompactRequest) string {
	graphSize := cc.calculateGraphSize(request.Graph)

	// Use different strategies based on graph characteristics
	if graphSize > 50000 {
		return "size" // For very large graphs, prioritize size reduction
	}

	if len(request.Graph.Files) > 1000 {
		return "dependency" // For many files, focus on dependencies
	}

	if request.Requirements != nil && len(request.Requirements.PreserveFiles) > 0 {
		return "relevance" // When specific files need preservation
	}

	// Default to hybrid for balanced approach
	return "hybrid"
}

func (cc *CompactController) calculateGraphSize(graph *types.CodeGraph) int {
	return len(graph.Files) + len(graph.Symbols) + len(graph.Nodes) + len(graph.Edges)
}

func (cc *CompactController) noCompactionResult(graph *types.CodeGraph) *CompactResult {
	size := cc.calculateGraphSize(graph)
	return &CompactResult{
		CompactedGraph:   graph,
		OriginalSize:     size,
		CompactedSize:    size,
		CompressionRatio: 1.0,
		Strategy:         "none",
		ExecutionTime:    0,
		RemovedItems:     &RemovedItems{},
		Metadata:         map[string]interface{}{"compaction": "disabled"},
	}
}

func (cc *CompactController) updateMetrics(result *CompactResult) {
	cc.metricsMux.Lock()
	defer cc.metricsMux.Unlock()
	
	cc.metrics.TotalCompactions++
	cc.metrics.LastCompaction = time.Now()

	// Update average time
	if cc.metrics.TotalCompactions == 1 {
		cc.metrics.AverageTime = result.ExecutionTime
	} else {
		// Exponential moving average
		alpha := 0.1
		cc.metrics.AverageTime = time.Duration(float64(cc.metrics.AverageTime)*(1-alpha) + float64(result.ExecutionTime)*alpha)
	}

	// Update compression ratio
	if cc.metrics.TotalCompactions == 1 {
		cc.metrics.CompressionRatio = result.CompressionRatio
	} else {
		alpha := 0.1
		cc.metrics.CompressionRatio = cc.metrics.CompressionRatio*(1-alpha) + result.CompressionRatio*alpha
	}

	// Update memory saved
	memorySaved := int64(result.OriginalSize - result.CompactedSize)
	cc.metrics.MemorySaved += memorySaved
	cc.metrics.ContextSizeReduced += memorySaved
}

// CompactionAnalysis represents the analysis of compaction potential
type CompactionAnalysis struct {
	TotalFiles          int                          `json:"total_files"`
	TotalSymbols        int                          `json:"total_symbols"`
	TotalNodes          int                          `json:"total_nodes"`
	TotalEdges          int                          `json:"total_edges"`
	Strategies          map[string]*StrategyAnalysis `json:"strategies"`
	RecommendedStrategy string                       `json:"recommended_strategy"`
	MaxCompressionRatio float64                      `json:"max_compression_ratio"`
	EstimatedSavings    int                          `json:"estimated_savings"`
}

// StrategyAnalysis represents analysis for a specific strategy
type StrategyAnalysis struct {
	EstimatedCompression float64  `json:"estimated_compression"`
	RemovableFiles       int      `json:"removable_files"`
	RemovableSymbols     int      `json:"removable_symbols"`
	Confidence           float64  `json:"confidence"`
	Warnings             []string `json:"warnings"`
}

func (ca *CompactionAnalysis) calculateOverallPotential() {
	bestRatio := 1.0
	bestStrategy := "hybrid" // Default strategy

	for strategy, analysis := range ca.Strategies {
		if analysis.EstimatedCompression < bestRatio {
			bestRatio = analysis.EstimatedCompression
			bestStrategy = strategy
		}
	}

	// Ensure we always have a recommended strategy
	if bestStrategy == "" && len(ca.Strategies) > 0 {
		// If no strategy found, pick the first one
		for strategy := range ca.Strategies {
			bestStrategy = strategy
			break
		}
	}

	ca.RecommendedStrategy = bestStrategy
	ca.MaxCompressionRatio = bestRatio
	totalSize := ca.TotalFiles + ca.TotalSymbols + ca.TotalNodes + ca.TotalEdges
	ca.EstimatedSavings = int(float64(totalSize) * (1.0 - bestRatio))
}

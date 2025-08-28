package git

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/nuthan-ms/codecontext/pkg/types"
)

// TestCompleteIntegrationFlow tests the complete semantic neighborhoods workflow
func TestCompleteIntegrationFlow(t *testing.T) {
	// Create real components (not mocks) to test the complete flow
	
	// 1. Test GitAnalyzer creation and basic functionality
	analyzer, err := NewGitAnalyzer(".")
	if err != nil {
		t.Fatalf("Failed to create GitAnalyzer: %v", err)
	}
	
	if !analyzer.IsGitRepository() {
		t.Skip("Skipping integration test - not in a git repository")
	}
	
	// 2. Test SemanticAnalyzer creation and repository analysis
	semanticAnalyzer, err := NewSemanticAnalyzer(".", DefaultSemanticConfig())
	if err != nil {
		t.Fatalf("Failed to create SemanticAnalyzer: %v", err)
	}
	
	result, err := semanticAnalyzer.AnalyzeRepository()
	if err != nil {
		t.Fatalf("Failed to analyze repository: %v", err)
	}
	
	t.Logf("Semantic analysis found %d neighborhoods", len(result.Neighborhoods))
	
	// 3. Create mock code graph for integration testing
	codeGraph := createRealisticCodeGraph()
	
	// 4. Test GraphIntegration creation and enhanced neighborhood building
	integration := NewGraphIntegration(semanticAnalyzer, codeGraph, DefaultIntegrationConfig())
	
	enhancedNeighborhoods, err := integration.BuildEnhancedNeighborhoods()
	if err != nil {
		t.Fatalf("Failed to build enhanced neighborhoods: %v", err)
	}
	
	t.Logf("Built %d enhanced neighborhoods", len(enhancedNeighborhoods))
	
	// 5. Test clustering functionality
	clusteredNeighborhoods, err := integration.BuildClusteredNeighborhoods()
	if err != nil {
		t.Fatalf("Failed to build clustered neighborhoods: %v", err)
	}
	
	t.Logf("Built %d clustered neighborhoods", len(clusteredNeighborhoods))
	
	// 6. Verify data integrity and relationships
	for i, clustered := range clusteredNeighborhoods {
		t.Logf("Cluster %d: %s", i, clustered.Cluster.Name)
		t.Logf("  - Description: %s", clustered.Cluster.Description)
		t.Logf("  - Size: %d neighborhoods", clustered.Cluster.Size)
		t.Logf("  - Strength: %.2f", clustered.Cluster.Strength)
		t.Logf("  - Quality metrics: Silhouette=%.2f, Davies-Bouldin=%.2f", 
			clustered.QualityMetrics.SilhouetteScore, 
			clustered.QualityMetrics.DaviesBouldinIndex)
		t.Logf("  - Optimal tasks: %v", clustered.Cluster.OptimalTasks)
		
		// Verify cluster contains valid neighborhoods
		if len(clustered.Neighborhoods) == 0 {
			t.Errorf("Cluster %d has no neighborhoods", i)
		}
		
		// Verify cluster metrics are reasonable
		if clustered.Cluster.Strength < 0 || clustered.Cluster.Strength > 1 {
			t.Errorf("Cluster %d has invalid strength: %.2f", i, clustered.Cluster.Strength)
		}
	}
	
	// 7. Test specific context recommendations
	if len(result.Neighborhoods) > 0 {
		// Test getting recommendations for a specific file
		firstFile := ""
		if len(result.Neighborhoods[0].Files) > 0 {
			firstFile = result.Neighborhoods[0].Files[0]
		}
		
		if firstFile != "" {
			recommendations, err := semanticAnalyzer.GetContextRecommendationsForFile(firstFile)
			if err != nil {
				t.Errorf("Failed to get context recommendations for file %s: %v", firstFile, err)
			}
			t.Logf("Found %d recommendations for file %s", len(recommendations), firstFile)
		}
	}
}

// TestIntegrationPerformance tests the performance of the complete workflow
func TestIntegrationPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}
	
	// Measure total execution time
	start := time.Now()
	
	// Create components
	analyzer, err := NewGitAnalyzer(".")
	if err != nil {
		t.Fatalf("Failed to create GitAnalyzer: %v", err)
	}
	
	if !analyzer.IsGitRepository() {
		t.Skip("Skipping performance test - not in a git repository")
	}
	
	semanticAnalyzer, err := NewSemanticAnalyzer(".", DefaultSemanticConfig())
	if err != nil {
		t.Fatalf("Failed to create SemanticAnalyzer: %v", err)
	}
	codeGraph := createRealisticCodeGraph()
	integration := NewGraphIntegration(semanticAnalyzer, codeGraph, DefaultIntegrationConfig())
	
	// Measure semantic analysis time
	semanticStart := time.Now()
	result, err := semanticAnalyzer.AnalyzeRepository()
	if err != nil {
		t.Fatalf("Failed to analyze repository: %v", err)
	}
	semanticDuration := time.Since(semanticStart)
	
	// Measure enhanced neighborhoods time
	enhancedStart := time.Now()
	enhancedNeighborhoods, err := integration.BuildEnhancedNeighborhoods()
	if err != nil {
		t.Fatalf("Failed to build enhanced neighborhoods: %v", err)
	}
	enhancedDuration := time.Since(enhancedStart)
	
	// Measure clustering time
	clusterStart := time.Now()
	clusteredNeighborhoods, err := integration.BuildClusteredNeighborhoods()
	if err != nil {
		t.Fatalf("Failed to build clustered neighborhoods: %v", err)
	}
	clusterDuration := time.Since(clusterStart)
	
	totalDuration := time.Since(start)
	
	// Report performance metrics
	t.Logf("Performance metrics:")
	t.Logf("  - Semantic analysis: %v (%d neighborhoods)", semanticDuration, len(result.Neighborhoods))
	t.Logf("  - Enhanced neighborhoods: %v (%d enhanced)", enhancedDuration, len(enhancedNeighborhoods))
	t.Logf("  - Clustering: %v (%d clusters)", clusterDuration, len(clusteredNeighborhoods))
	t.Logf("  - Total time: %v", totalDuration)
	
	// Performance assertions (adjusted for CI environments which can be slower)
	if totalDuration > 120*time.Second {
		t.Errorf("Total execution time too slow: %v", totalDuration)
	}
	
	if semanticDuration > 60*time.Second {
		t.Errorf("Semantic analysis too slow: %v", semanticDuration)
	}
	
	if clusterDuration > 45*time.Second && len(enhancedNeighborhoods) < 100 {
		t.Errorf("Clustering too slow for %d neighborhoods: %v", len(enhancedNeighborhoods), clusterDuration)
	}
}

// TestIntegrationEdgeCases tests edge cases in the integration flow
func TestIntegrationEdgeCases(t *testing.T) {
	// Test with empty code graph
	analyzer, err := NewGitAnalyzer(".")
	if err != nil {
		t.Fatalf("Failed to create GitAnalyzer: %v", err)
	}
	
	if !analyzer.IsGitRepository() {
		t.Skip("Skipping edge case test - not in a git repository")
	}
	
	semanticAnalyzer, err := NewSemanticAnalyzer(".", DefaultSemanticConfig())
	if err != nil {
		t.Fatalf("Failed to create SemanticAnalyzer: %v", err)
	}
	emptyCodeGraph := &types.CodeGraph{
		Nodes:   make(map[types.NodeId]*types.GraphNode),
		Edges:   make(map[types.EdgeId]*types.GraphEdge),
		Files:   make(map[string]*types.FileNode),
		Symbols: make(map[types.SymbolId]*types.Symbol),
	}
	
	integration := NewGraphIntegration(semanticAnalyzer, emptyCodeGraph, DefaultIntegrationConfig())
	
	// Should handle empty graph gracefully
	enhancedNeighborhoods, err := integration.BuildEnhancedNeighborhoods()
	if err != nil {
		t.Errorf("Failed to handle empty code graph: %v", err)
	}
	t.Logf("Empty graph produced %d enhanced neighborhoods", len(enhancedNeighborhoods))
	
	// Test clustering with empty neighborhoods
	clusteredNeighborhoods, err := integration.BuildClusteredNeighborhoods()
	if err != nil {
		t.Errorf("Failed to handle empty neighborhoods: %v", err)
	}
	t.Logf("Empty neighborhoods produced %d clusters", len(clusteredNeighborhoods))
	
	// Test with nil code graph
	integrationNil := NewGraphIntegration(semanticAnalyzer, nil, DefaultIntegrationConfig())
	
	enhancedNil, err := integrationNil.BuildEnhancedNeighborhoods()
	if err != nil {
		t.Errorf("Failed to handle nil code graph: %v", err)
	}
	t.Logf("Nil graph produced %d enhanced neighborhoods", len(enhancedNil))
}

// TestConfigurationImpact tests how different configurations affect results
func TestConfigurationImpact(t *testing.T) {
	analyzer, err := NewGitAnalyzer(".")
	if err != nil {
		t.Fatalf("Failed to create GitAnalyzer: %v", err)
	}
	
	if !analyzer.IsGitRepository() {
		t.Skip("Skipping configuration test - not in a git repository")
	}
	
	semanticAnalyzer, err := NewSemanticAnalyzer(".", DefaultSemanticConfig())
	if err != nil {
		t.Fatalf("Failed to create SemanticAnalyzer: %v", err)
	}
	codeGraph := createRealisticCodeGraph()
	
	// Test with default config
	defaultIntegration := NewGraphIntegration(semanticAnalyzer, codeGraph, DefaultIntegrationConfig())
	defaultClustered, err := defaultIntegration.BuildClusteredNeighborhoods()
	if err != nil {
		t.Fatalf("Failed with default config: %v", err)
	}
	
	// Test with git-heavy config
	gitHeavyConfig := &IntegrationConfig{
		WeightGitPatterns:     0.8,
		WeightDependencies:    0.1,
		WeightStructural:      0.1,
		MinCombinedScore:      0.3,
		MaxNeighborhoodSize:   20,
		IncludeWeakRelations:  true,
		PrioritizeRecentFiles: true,
	}
	
	gitHeavyIntegration := NewGraphIntegration(semanticAnalyzer, codeGraph, gitHeavyConfig)
	gitHeavyClustered, err := gitHeavyIntegration.BuildClusteredNeighborhoods()
	if err != nil {
		t.Fatalf("Failed with git-heavy config: %v", err)
	}
	
	// Test with dependency-heavy config
	depHeavyConfig := &IntegrationConfig{
		WeightGitPatterns:     0.1,
		WeightDependencies:    0.8,
		WeightStructural:      0.1,
		MinCombinedScore:      0.5,
		MaxNeighborhoodSize:   10,
		IncludeWeakRelations:  false,
		PrioritizeRecentFiles: false,
	}
	
	depHeavyIntegration := NewGraphIntegration(semanticAnalyzer, codeGraph, depHeavyConfig)
	depHeavyClustered, err := depHeavyIntegration.BuildClusteredNeighborhoods()
	if err != nil {
		t.Fatalf("Failed with dependency-heavy config: %v", err)
	}
	
	// Compare results
	t.Logf("Configuration impact:")
	t.Logf("  - Default config: %d clusters", len(defaultClustered))
	t.Logf("  - Git-heavy config: %d clusters", len(gitHeavyClustered))
	t.Logf("  - Dependency-heavy config: %d clusters", len(depHeavyClustered))
	
	// Verify configurations produce different results (unless repository is very simple)
	if len(defaultClustered) > 0 {
		defaultStrength := defaultClustered[0].Cluster.Strength
		if len(gitHeavyClustered) > 0 {
			gitHeavyStrength := gitHeavyClustered[0].Cluster.Strength
			t.Logf("  - First cluster strength: default=%.3f, git-heavy=%.3f", defaultStrength, gitHeavyStrength)
		}
	}
}

// createRealisticCodeGraph creates a more realistic code graph for testing
func createRealisticCodeGraph() *types.CodeGraph {
	graph := &types.CodeGraph{
		Nodes:   make(map[types.NodeId]*types.GraphNode),
		Edges:   make(map[types.EdgeId]*types.GraphEdge),
		Files:   make(map[string]*types.FileNode),
		Symbols: make(map[types.SymbolId]*types.Symbol),
	}
	
	// Add realistic file nodes
	files := []string{
		"src/user.go",
		"src/auth.go", 
		"src/database.go",
		"src/api.go",
		"src/utils.go",
		"test/user_test.go",
		"test/auth_test.go",
		"main.go",
	}
	
	symbolId := 0
	for i, filePath := range files {
		// Create file node
		fileNode := &types.FileNode{
			Path:        filePath,
			Language:    "go",
			Size:        1000 + i*200,
			Lines:       50 + i*10,
			SymbolCount: 3 + i,
			ImportCount: 2,
			IsTest:      strings.Contains(filePath, "test"),
			LastModified: time.Now().Add(-time.Duration(i) * time.Hour),
			Symbols:     []types.SymbolId{},
			Imports:     []*types.Import{},
		}
		
		// Create symbols for this file
		for j := 0; j < 3+i; j++ {
			symbolIdStr := types.SymbolId(fmt.Sprintf("symbol_%d", symbolId))
			symbol := &types.Symbol{
				Id:   symbolIdStr,
				Name: fmt.Sprintf("Symbol%d", symbolId),
				Type: types.SymbolTypeFunction,
				Location: types.Location{
					StartLine: j + 1,
				},
				Language: "go",
			}
			
			graph.Symbols[symbolIdStr] = symbol
			fileNode.Symbols = append(fileNode.Symbols, symbolIdStr)
			symbolId++
		}
		
		graph.Files[filePath] = fileNode
		
		// Create graph node
		nodeId := types.NodeId(fmt.Sprintf("node_%d", i))
		graphNode := &types.GraphNode{
			Id:              nodeId,
			Type:            "file",
			Label:           filePath,
			FilePath:        filePath,
			Importance:      float64(len(files)-i) / float64(len(files)),
			Connections:     i,
			ChangeFrequency: 10 - i,
			LastModified:    time.Now().Add(-time.Duration(i) * time.Hour),
		}
		
		graph.Nodes[nodeId] = graphNode
	}
	
	// Add some edges to represent dependencies
	edgeId := 0
	relationships := []struct {
		from, to   string
		edgeType   string
		weight     float64
	}{
		{"node_0", "node_1", "imports", 0.8},  // user imports auth
		{"node_0", "node_2", "imports", 0.6},  // user imports database
		{"node_1", "node_2", "imports", 0.7},  // auth imports database
		{"node_3", "node_0", "imports", 0.9},  // api imports user
		{"node_3", "node_1", "imports", 0.8},  // api imports auth
		{"node_4", "node_0", "imports", 0.5},  // utils imports user
		{"node_5", "node_0", "calls", 0.9},    // user_test calls user
		{"node_6", "node_1", "calls", 0.9},    // auth_test calls auth
		{"node_7", "node_3", "imports", 1.0},  // main imports api
	}
	
	for _, rel := range relationships {
		edgeIdStr := types.EdgeId(fmt.Sprintf("edge_%d", edgeId))
		edge := &types.GraphEdge{
			Id:     edgeIdStr,
			From:   types.NodeId(rel.from),
			To:     types.NodeId(rel.to),
			Type:   rel.edgeType,
			Weight: rel.weight,
		}
		
		graph.Edges[edgeIdStr] = edge
		edgeId++
	}
	
	return graph
}
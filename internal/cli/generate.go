package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nuthan-ms/codecontext/internal/analyzer"
	"github.com/nuthan-ms/codecontext/internal/cache"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate initial context map",
	Long: `Generate a comprehensive context map of the codebase.
This command analyzes the entire repository and creates an intelligent
context map optimized for AI-powered development tools.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return generateContextMap(cmd)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringP("target", "t", ".", "target directory to analyze")
	generateCmd.Flags().BoolP("watch", "w", false, "enable watch mode for continuous updates")
	generateCmd.Flags().StringP("format", "f", "markdown", "output format (markdown, json, yaml)")

	// Bind flags to viper with error handling
	if err := viper.BindPFlag("target", generateCmd.Flags().Lookup("target")); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bind target flag: %v\n", err)
	}
	if err := viper.BindPFlag("watch", generateCmd.Flags().Lookup("watch")); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bind watch flag: %v\n", err)
	}
	if err := viper.BindPFlag("format", generateCmd.Flags().Lookup("format")); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bind format flag: %v\n", err)
	}
}

func generateContextMap(cmd *cobra.Command) error {
	start := time.Now()

	// Initialize progress manager
	progressManager := NewProgressManager()
	defer progressManager.Stop()

	if viper.GetBool("verbose") {
		fmt.Println("🔍 Starting context map generation...")
	}

	// Get target directory from flags - try direct flag first, then viper fallback
	targetDir, err := cmd.Flags().GetString("target")
	if err != nil || targetDir == "" {
		targetDir = viper.GetString("target")
		if targetDir == "" {
			targetDir = "."
		}
	}

	outputFile := viper.GetString("output")
	if outputFile == "" {
		outputFile = "CLAUDE.md"
	}

	if viper.GetBool("verbose") {
		fmt.Printf("📁 Analyzing directory: %s\n", targetDir)
		fmt.Printf("📄 Output file: %s\n", outputFile)
	}

	// Initialize cache for better performance
	cacheDir := filepath.Join(os.TempDir(), "codecontext", "cache")
	cacheConfig := &cache.Config{
		Directory:     cacheDir,
		MaxSize:       1000,
		TTL:           24 * time.Hour,
		EnableLRU:     true,
		EnableMetrics: true,
	}

	persistentCache, err := cache.NewPersistentCache(cacheConfig)
	if err != nil {
		// Log warning but don't fail - cache is optional
		if viper.GetBool("verbose") {
			fmt.Printf("⚠️  Cache initialization failed: %v\n", err)
		}
	}

	// Start analysis with progress tracking
	progressManager.StartIndeterminate("🔍 Initializing analysis...")

	// Create graph builder and analyze directory
	builder := analyzer.NewGraphBuilder()

	// Set cache if available
	if persistentCache != nil {
		builder.SetCache(persistentCache)
	}

	// Set use_default_excludes from config (default true)
	useDefaultExcludes := true
	if viper.IsSet("use_default_excludes") {
		useDefaultExcludes = viper.GetBool("use_default_excludes")
	}
	builder.SetUseDefaultExcludes(useDefaultExcludes)
	
	// Set exclude patterns from config
	excludePatterns := viper.GetStringSlice("exclude_patterns")
	if len(excludePatterns) > 0 {
		builder.SetExcludePatterns(excludePatterns)
		if viper.GetBool("verbose") {
			// Count include patterns (starting with !)
			includeCount := 0
			for _, p := range excludePatterns {
				if strings.HasPrefix(p, "!") {
					includeCount++
				}
			}
			excludeCount := len(excludePatterns) - includeCount
			
			fmt.Printf("🚫 Exclude patterns: %d, Include overrides: %d\n", excludeCount, includeCount)
			if !useDefaultExcludes {
				fmt.Println("   ⚠️  Default excludes disabled")
			}
		}
	}

	// Set up progress callback for real-time updates
	builder.SetProgressCallback(func(message string) {
		progressManager.UpdateIndeterminate(message)
	})

	graph, err := builder.AnalyzeDirectory(targetDir)
	if err != nil {
		return fmt.Errorf("failed to analyze directory: %w", err)
	}

	progressManager.UpdateIndeterminate("📝 Generating context map...")

	if viper.GetBool("verbose") {
		stats := builder.GetFileStats()
		fmt.Printf("📊 Analysis complete: %d files, %d symbols\n",
			stats["totalFiles"], stats["totalSymbols"])
	}

	// Generate markdown content from real data
	generator := analyzer.NewMarkdownGenerator(graph)
	content := generator.GenerateContextMap()

	progressManager.UpdateIndeterminate("💾 Writing output file...")

	// Write real content
	if err := writeOutputFile(outputFile, content); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	progressManager.UpdateIndeterminate("✅ Complete")

	progressManager.Stop()

	duration := time.Since(start)
	fmt.Printf("✅ Context map generated successfully in %v\n", duration)
	fmt.Printf("   Output file: %s\n", outputFile)

	return nil
}

func writeOutputFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

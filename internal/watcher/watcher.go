package watcher

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/nuthan-ms/codecontext/internal/analyzer"
)

// FileWatcher monitors filesystem changes and triggers incremental updates
type FileWatcher struct {
	watcher    *fsnotify.Watcher
	analyzer   *analyzer.GraphBuilder
	targetDir  string
	outputFile string
	debounce   time.Duration
	changes    chan FileChange
	done       chan struct{}

	// Configuration
	excludePatterns []string
	includeExts     []string
}

// FileChange represents a file system change event
type FileChange struct {
	Path      string
	Operation string
	Timestamp time.Time
}

// Config holds configuration for the file watcher
type Config struct {
	TargetDir       string
	OutputFile      string
	DebounceTime    time.Duration
	ExcludePatterns []string
	IncludeExts     []string
}

// NewFileWatcher creates a new file watcher instance
func NewFileWatcher(config Config) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	if config.DebounceTime == 0 {
		config.DebounceTime = 500 * time.Millisecond
	}

	if len(config.IncludeExts) == 0 {
		config.IncludeExts = []string{".ts", ".tsx", ".js", ".jsx", ".json", ".yaml", ".yml"}
	}

	if len(config.ExcludePatterns) == 0 {
		config.ExcludePatterns = []string{
			"node_modules",
			".git",
			".codecontext",
			"dist",
			"build",
			"coverage",
			"*.log",
			"*.tmp",
		}
	}

	return &FileWatcher{
		watcher:         watcher,
		analyzer:        analyzer.NewGraphBuilder(),
		targetDir:       config.TargetDir,
		outputFile:      config.OutputFile,
		debounce:        config.DebounceTime,
		changes:         make(chan FileChange, 100),
		done:            make(chan struct{}),
		excludePatterns: config.ExcludePatterns,
		includeExts:     config.IncludeExts,
	}, nil
}

// Start begins watching for file changes
func (fw *FileWatcher) Start(ctx context.Context) error {
	// Add target directory to watcher
	err := fw.addDirectory(fw.targetDir)
	if err != nil {
		return fmt.Errorf("failed to add directory to watcher: %w", err)
	}

	// Start change processor
	go fw.processChanges(ctx)

	// Start file system event handler
	go fw.handleEvents(ctx)

	log.Printf("🔍 File watcher started for: %s", fw.targetDir)
	log.Printf("   Debounce time: %v", fw.debounce)
	log.Printf("   Watching extensions: %v", fw.includeExts)

	return nil
}

// Stop stops the file watcher
func (fw *FileWatcher) Stop() error {
	close(fw.done)
	return fw.watcher.Close()
}

// StopWatching stops the file watcher (alias for Stop)
func (fw *FileWatcher) StopWatching() error {
	return fw.Stop()
}

// addDirectory recursively adds a directory and its subdirectories to the watcher
func (fw *FileWatcher) addDirectory(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		// Skip excluded directories
		if fw.shouldExclude(path) {
			return filepath.SkipDir
		}

		return fw.watcher.Add(path)
	})
}

// shouldExclude checks if a path should be excluded from watching
func (fw *FileWatcher) shouldExclude(path string) bool {
	for _, pattern := range fw.excludePatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

// shouldInclude checks if a file should be included based on extension
func (fw *FileWatcher) shouldInclude(path string) bool {
	ext := filepath.Ext(path)
	for _, includeExt := range fw.includeExts {
		if ext == includeExt {
			return true
		}
	}
	return false
}

// handleEvents processes filesystem events
func (fw *FileWatcher) handleEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-fw.done:
			return
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			// Skip if file should be excluded
			if fw.shouldExclude(event.Name) {
				continue
			}

			// Skip if file extension not supported
			if !fw.shouldInclude(event.Name) {
				continue
			}

			// Create change event
			change := FileChange{
				Path:      event.Name,
				Operation: event.Op.String(),
				Timestamp: time.Now(),
			}

			// Send to change processor
			select {
			case fw.changes <- change:
			default:
				// Channel full, skip this event
			}

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("❌ File watcher error: %v", err)
		}
	}
}

// processChanges handles debounced file changes
func (fw *FileWatcher) processChanges(ctx context.Context) {
	var pendingChanges []FileChange
	timer := time.NewTimer(fw.debounce)
	timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-fw.done:
			return
		case change := <-fw.changes:
			pendingChanges = append(pendingChanges, change)

			// Reset debounce timer
			timer.Reset(fw.debounce)

		case <-timer.C:
			if len(pendingChanges) > 0 {
				err := fw.processFileChanges(pendingChanges)
				if err != nil {
					log.Printf("❌ Error processing file changes: %v", err)
				}
				pendingChanges = nil
			}
		}
	}
}

// processFileChanges performs incremental analysis on changed files
func (fw *FileWatcher) processFileChanges(changes []FileChange) error {
	start := time.Now()

	log.Printf("🔄 Processing %d file changes...", len(changes))

	// Group changes by type
	changedFiles := make(map[string]string)
	for _, change := range changes {
		changedFiles[change.Path] = change.Operation
	}

	// Perform incremental analysis
	graph, err := fw.analyzer.AnalyzeDirectory(fw.targetDir)
	if err != nil {
		return fmt.Errorf("failed to analyze directory: %w", err)
	}

	// Generate updated context map
	generator := analyzer.NewMarkdownGenerator(graph)
	content := generator.GenerateContextMap()

	// Write to output file
	err = fw.writeOutput(content)
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	duration := time.Since(start)
	log.Printf("✅ Context map updated in %v", duration)
	log.Printf("   Files processed: %d", len(changedFiles))

	return nil
}

// writeOutput writes the generated content to the output file
func (fw *FileWatcher) writeOutput(content string) error {
	return os.WriteFile(fw.outputFile, []byte(content), 0644)
}

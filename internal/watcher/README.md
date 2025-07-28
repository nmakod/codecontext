# File Watcher Package

This package provides real-time file system monitoring capabilities for CodeContext, enabling automatic context map updates when source files change.

## Features

- **Real-time monitoring**: Uses `fsnotify` for efficient file system watching
- **Debounced updates**: Batches file changes to avoid excessive rebuilds
- **Configurable filtering**: Include/exclude patterns for files and directories
- **Incremental analysis**: Leverages the analyzer package for code graph updates
- **Graceful shutdown**: Supports context-based cancellation

## Usage

### Basic Usage

```go
import "github.com/nmakod/codecontext/internal/watcher"

config := watcher.Config{
    TargetDir:    "/path/to/project",
    OutputFile:   "/path/to/output.md",
    DebounceTime: 500 * time.Millisecond,
}

fileWatcher, err := watcher.NewFileWatcher(config)
if err != nil {
    log.Fatal(err)
}
defer fileWatcher.Stop()

ctx := context.Background()
err = fileWatcher.Start(ctx)
if err != nil {
    log.Fatal(err)
}
```

### CLI Usage

```bash
# Watch for changes and update automatically
codecontext update --watch

# Custom debounce time
codecontext update --watch --debounce 1s

# Verbose output
codecontext update --watch --verbose
```

## Configuration

### Config Structure

```go
type Config struct {
    TargetDir       string        // Directory to watch
    OutputFile      string        // Output file path
    DebounceTime    time.Duration // Debounce time for batching changes
    ExcludePatterns []string      // Patterns to exclude from watching
    IncludeExts     []string      // File extensions to include
}
```

### Default Values

- **DebounceTime**: 500ms
- **ExcludePatterns**: `["node_modules", ".git", ".codecontext", "dist", "build", "coverage", "*.log", "*.tmp"]`
- **IncludeExts**: `[".ts", ".tsx", ".js", ".jsx", ".json", ".yaml", ".yml"]`

## Architecture

### Components

1. **FileWatcher**: Main component that orchestrates file monitoring
2. **Event Handler**: Processes file system events from `fsnotify`
3. **Change Processor**: Debounces and batches file changes
4. **Analyzer Integration**: Triggers incremental analysis via the analyzer package

### Data Flow

```
File System → fsnotify → Event Handler → Change Processor → Analyzer → Output
```

### Event Processing

1. **File Change Detection**: `fsnotify` detects file system events
2. **Filtering**: Events are filtered based on include/exclude patterns
3. **Debouncing**: Changes are batched using a configurable debounce timer
4. **Analysis**: Batched changes trigger incremental analysis
5. **Output Generation**: Updated context map is written to output file

## Performance Considerations

### Memory Usage

- **Event Buffer**: 100 events maximum in the change channel
- **Debounce Timer**: Single timer per watcher instance
- **Analyzer**: Reuses existing analyzer instances for efficiency

### File System Monitoring

- **Recursive Watching**: Automatically watches subdirectories
- **Smart Filtering**: Excludes common build/cache directories
- **Efficient Updates**: Only processes supported file types

### Debouncing Strategy

The watcher uses a debounce strategy to batch file changes:

1. **Timer Reset**: Each new change resets the debounce timer
2. **Batch Processing**: When the timer expires, all pending changes are processed together
3. **Configurable Delay**: Default 500ms, adjustable via configuration

## Error Handling

### Graceful Degradation

- **Watcher Errors**: Logged but don't stop the process
- **Analysis Errors**: Reported but don't crash the watcher
- **File System Errors**: Handled with appropriate error messages

### Recovery Mechanisms

- **Context Cancellation**: Supports graceful shutdown
- **Resource Cleanup**: Properly closes file system watchers
- **Error Reporting**: Comprehensive error messages for debugging

## Testing

### Unit Tests

```bash
go test ./internal/watcher/...
```

### Integration Tests

The package includes integration tests that:

1. Create temporary test directories
2. Set up file watchers
3. Modify files and verify updates
4. Test debouncing behavior
5. Validate output generation

### Test Coverage

- **Unit Tests**: Core functionality and edge cases
- **Integration Tests**: End-to-end file watching workflow
- **Performance Tests**: Memory and CPU usage validation

## Examples

### Basic File Watching

```go
config := watcher.Config{
    TargetDir:    "./src",
    OutputFile:   "./context.md",
    DebounceTime: 300 * time.Millisecond,
}

watcher, err := watcher.NewFileWatcher(config)
// Handle error and start watcher
```

### Custom Filtering

```go
config := watcher.Config{
    TargetDir:       "./src",
    OutputFile:      "./context.md",
    ExcludePatterns: []string{"node_modules", "*.test.js"},
    IncludeExts:     []string{".ts", ".js"},
}

watcher, err := watcher.NewFileWatcher(config)
// Handle error and start watcher
```

### Context-Based Shutdown

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

err := watcher.Start(ctx)
// Watcher will stop when context is cancelled
```

## Future Enhancements

### Planned Features

1. **Virtual Graph Integration**: Connect with Virtual Graph Engine for more efficient updates
2. **Change Batching**: Intelligent batching based on file relationships
3. **Performance Metrics**: Detailed monitoring and performance reporting
4. **Custom Filters**: User-defined filtering rules
5. **Multiple Output Formats**: Support for various output formats

### Performance Optimizations

1. **Selective Updates**: Only update affected parts of the context map
2. **Parallel Processing**: Process independent file changes in parallel
3. **Cache Integration**: Leverage caching for faster updates
4. **Memory Optimization**: Reduce memory footprint for large projects

## Dependencies

- **fsnotify**: File system event monitoring
- **analyzer**: Code analysis and graph building
- **context**: Graceful shutdown and cancellation

## License

This package is part of the CodeContext project and follows the same licensing terms.
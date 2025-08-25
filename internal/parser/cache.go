package parser

import (
	"fmt"
	"sync"
	"time"

	"github.com/nuthan-ms/codecontext/pkg/types"
)

// ASTCache implements the AST cache interface
type ASTCache struct {
	astCache   map[string]*types.VersionedAST
	diffCache  map[string][]*types.ASTDiff
	mu         sync.RWMutex
	maxSize    int
	ttl        time.Duration
	timestamps map[string]time.Time
}

// NewASTCache creates a new AST cache
func NewASTCache() *ASTCache {
	return &ASTCache{
		astCache:   make(map[string]*types.VersionedAST),
		diffCache:  make(map[string][]*types.ASTDiff),
		maxSize:    1000,
		ttl:        time.Hour,
		timestamps: make(map[string]time.Time),
	}
}

// Get retrieves an AST from the cache
func (c *ASTCache) Get(fileId string, version ...string) (*types.VersionedAST, error) {
	var key string
	if len(version) > 0 {
		key = fmt.Sprintf("%s:%s", fileId, version[0])
	} else {
		key = fileId
	}

	// First, check under read lock
	c.mu.RLock()
	ast, exists := c.astCache[key]
	timestamp, timestampExists := c.timestamps[key]
	c.mu.RUnlock()

	// If entry exists and is valid, return it
	if exists && timestampExists && time.Since(timestamp) < c.ttl {
		return ast, nil
	}

	// If entry exists but is expired, clean it up under write lock
	if exists && timestampExists {
		c.mu.Lock()
		// Double-check after acquiring write lock (entry might have been cleaned by another goroutine)
		if t, ok := c.timestamps[key]; ok && time.Since(t) >= c.ttl {
			delete(c.astCache, key)
			delete(c.timestamps, key)
		}
		c.mu.Unlock()
	}

	return nil, fmt.Errorf("AST not found in cache: %s", key)
}

// Set stores an AST in the cache
func (c *ASTCache) Set(fileId string, ast *types.VersionedAST) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if cache is full
	if len(c.astCache) >= c.maxSize {
		c.evictOldest()
	}

	key := fmt.Sprintf("%s:%s", fileId, ast.Version)
	c.astCache[key] = ast
	c.timestamps[key] = time.Now()

	return nil
}

// GetDiffCache retrieves diffs from the cache
func (c *ASTCache) GetDiffCache(fileId string) ([]*types.ASTDiff, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if diffs, exists := c.diffCache[fileId]; exists {
		return diffs, nil
	}

	return nil, fmt.Errorf("diff cache not found for file: %s", fileId)
}

// SetDiffCache stores diffs in the cache
func (c *ASTCache) SetDiffCache(fileId string, diffs []*types.ASTDiff) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.diffCache[fileId] = diffs
	return nil
}

// Invalidate removes an entry from the cache
func (c *ASTCache) Invalidate(fileId string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove all versions of this file
	for key := range c.astCache {
		if key == fileId || (len(key) > len(fileId) && key[:len(fileId)] == fileId && key[len(fileId)] == ':') {
			delete(c.astCache, key)
			delete(c.timestamps, key)
		}
	}

	// Remove diff cache
	delete(c.diffCache, fileId)

	return nil
}

// Clear removes all entries from the cache
func (c *ASTCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.astCache = make(map[string]*types.VersionedAST)
	c.diffCache = make(map[string][]*types.ASTDiff)
	c.timestamps = make(map[string]time.Time)

	return nil
}

// Size returns the current size of the cache
func (c *ASTCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.astCache)
}

// Stats returns cache statistics
func (c *ASTCache) Stats() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]any{
		"ast_entries":  len(c.astCache),
		"diff_entries": len(c.diffCache),
		"max_size":     c.maxSize,
		"ttl_seconds":  c.ttl.Seconds(),
	}
}

// evictOldest removes the oldest entry from the cache
func (c *ASTCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, timestamp := range c.timestamps {
		if oldestKey == "" || timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = timestamp
		}
	}

	if oldestKey != "" {
		delete(c.astCache, oldestKey)
		delete(c.timestamps, oldestKey)
	}
}

// SetMaxSize sets the maximum cache size
func (c *ASTCache) SetMaxSize(size int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.maxSize = size

	// Evict entries if current size exceeds new max size
	for len(c.astCache) > c.maxSize {
		c.evictOldest()
	}
}

// SetTTL sets the time-to-live for cache entries
func (c *ASTCache) SetTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ttl = ttl
}

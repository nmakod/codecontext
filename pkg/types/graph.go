package types

import (
	"time"
)

// SymbolId represents a unique identifier for a symbol
type SymbolId string

// NodeId represents a unique identifier for a graph node
type NodeId string

// SymbolType represents the type of a symbol
type SymbolType string

const (
	SymbolTypeFunction  SymbolType = "function"
	SymbolTypeClass     SymbolType = "class"
	SymbolTypeInterface SymbolType = "interface"
	SymbolTypeType      SymbolType = "type"
	SymbolTypeVariable  SymbolType = "variable"
	SymbolTypeConstant  SymbolType = "constant"
	SymbolTypeImport    SymbolType = "import"
	SymbolTypeNamespace SymbolType = "namespace"
	SymbolTypeMethod    SymbolType = "method"
	SymbolTypeProperty  SymbolType = "property"
	
	// Framework-specific symbol types
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
	
	// C++ specific symbol types
	SymbolTypeConstructor  SymbolType = "constructor"  // C++ constructors
	SymbolTypeDestructor   SymbolType = "destructor"   // C++ destructors
	SymbolTypeOperator     SymbolType = "operator"     // C++ operator overloads
	SymbolTypeTemplate     SymbolType = "template"     // C++ templates
	SymbolTypeCppTypedef   SymbolType = "cpp_typedef"  // C++ typedefs
	SymbolTypeCppUsing     SymbolType = "cpp_using"    // C++ using declarations
)

// FileLocation represents a location in a file
type FileLocation struct {
	FilePath  string `json:"file_path"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	EndLine   int    `json:"end_line"`
	EndColumn int    `json:"end_column"`
}

// Symbol represents a code symbol
type Symbol struct {
	Id                 SymbolId   `json:"id"`
	Name               string     `json:"name"`
	Type               SymbolType `json:"type"`
	Kind               string     `json:"kind"` // For diff compatibility
	FullyQualifiedName string     `json:"fully_qualified_name"`
	Location           Location   `json:"location"`
	Signature          string     `json:"signature,omitempty"`
	Documentation      string     `json:"documentation,omitempty"`
	Visibility         string     `json:"visibility,omitempty"`
	Language           string     `json:"language"`
	Hash               string     `json:"hash"`
	LastModified       time.Time  `json:"last_modified"`
}

// GraphNode represents a node in the code graph
type GraphNode struct {
	Id              NodeId                 `json:"id"`
	Type            string                 `json:"type"`
	Label           string                 `json:"label"`
	FilePath        string                 `json:"file_path,omitempty"`
	Symbol          *Symbol                `json:"symbol,omitempty"`
	Importance      float64                `json:"importance"`
	Connections     int                    `json:"connections"`
	ChangeFrequency int                    `json:"change_frequency"`
	LastModified    time.Time              `json:"last_modified"`
	Tags            []string               `json:"tags,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// Edge represents a connection between graph nodes
type Edge struct {
	From     NodeId                 `json:"from"`
	To       NodeId                 `json:"to"`
	Type     string                 `json:"type"`
	Weight   float64                `json:"weight"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// GraphVersion represents a version of the graph
type GraphVersion struct {
	Major       int       `json:"major"`
	Minor       int       `json:"minor"`
	Patch       int       `json:"patch"`
	Timestamp   time.Time `json:"timestamp"`
	ChangeCount int       `json:"change_count"`
	Hash        string    `json:"hash"`
}

// GraphMetadata contains metadata about the graph
type GraphMetadata struct {
	ProjectName    string                 `json:"project_name"`
	ProjectPath    string                 `json:"project_path"`
	TotalFiles     int                    `json:"total_files"`
	TotalSymbols   int                    `json:"total_symbols"`
	Languages      map[string]int         `json:"languages"`
	Generated      time.Time              `json:"generated"`
	GeneratedAt    time.Time              `json:"generated_at"`
	ProcessingTime time.Duration          `json:"processing_time"`
	AnalysisTime   time.Duration          `json:"analysis_time"`
	Version        string                 `json:"version"`
	TokenCount     int                    `json:"token_count"`
	Configuration  map[string]interface{} `json:"configuration,omitempty"`
}

// CodeGraph represents the complete code graph
type CodeGraph struct {
	Nodes        map[NodeId]*GraphNode `json:"nodes"`
	Edges        map[EdgeId]*GraphEdge `json:"edges"`
	Files        map[string]*FileNode  `json:"files"`
	Symbols      map[SymbolId]*Symbol  `json:"symbols"`
	Metadata     *GraphMetadata        `json:"metadata"`
	Version      GraphVersion          `json:"version"`
	PatchHistory []GraphPatch          `json:"patch_history,omitempty"`
}

// GraphPatch represents a change to the graph
type GraphPatch struct {
	Id           string                 `json:"id"`
	Type         string                 `json:"type"` // "add", "remove", "modify", "reorder"
	TargetNode   NodeId                 `json:"target_node"`
	Changes      []PropertyChange       `json:"changes"`
	Dependencies []NodeId               `json:"dependencies"`
	Timestamp    time.Time              `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// PropertyChange represents a change to a property
type PropertyChange struct {
	Property string      `json:"property"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
}

// Import represents an import statement
type Import struct {
	Path       string       `json:"path"`
	Alias      string       `json:"alias,omitempty"`
	Specifiers []string     `json:"specifiers,omitempty"`
	IsDefault  bool         `json:"is_default"`
	Location   FileLocation `json:"location"`
}

// Language represents a programming language configuration
type Language struct {
	Name       string   `json:"name"`
	Extensions []string `json:"extensions"`
	Parser     string   `json:"parser"`
	TreeSitter string   `json:"tree_sitter"`
	QueryPath  string   `json:"query_path,omitempty"`
	Enabled    bool     `json:"enabled"`
}

// FileClassification represents the classification of a file
type FileClassification struct {
	Language    Language `json:"language"`
	FileType    string   `json:"file_type"` // "source", "test", "config", "documentation"
	IsGenerated bool     `json:"is_generated"`
	IsTest      bool     `json:"is_test"`
	Framework   string   `json:"framework,omitempty"`
	Confidence  float64  `json:"confidence"`
}

// EdgeId represents a unique identifier for an edge
type EdgeId string

// GraphEdge represents an edge in the code graph with metadata
type GraphEdge struct {
	Id       EdgeId                 `json:"id"`
	From     NodeId                 `json:"from"`
	To       NodeId                 `json:"to"`
	Type     string                 `json:"type"`
	Weight   float64                `json:"weight"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// FileNode represents a file in the codebase
type FileNode struct {
	Path         string     `json:"path"`
	Language     string     `json:"language"`
	Size         int        `json:"size"`
	Lines        int        `json:"lines"`
	SymbolCount  int        `json:"symbol_count"`
	ImportCount  int        `json:"import_count"`
	IsTest       bool       `json:"is_test"`
	IsGenerated  bool       `json:"is_generated"`
	LastModified time.Time  `json:"last_modified"`
	Symbols      []SymbolId `json:"symbols"`
	Imports      []*Import  `json:"imports"`
}

// FileInfo represents file information for diff operations
type FileInfo struct {
	Path     string      `json:"path"`
	Language string      `json:"language"`
	Content  string      `json:"content"`
	ModTime  time.Time   `json:"mod_time"`
	Symbols  []*Symbol   `json:"symbols"`
	AST      interface{} `json:"ast,omitempty"`
	Hash     string      `json:"hash,omitempty"`
	Size     int         `json:"size"`
	Lines    int         `json:"lines"`
}

// Location represents a location in source code (for diff compatibility)
type Location struct {
	StartLine   int `json:"start_line"`
	StartColumn int `json:"start_column"`
	EndLine     int `json:"end_line"`
	EndColumn   int `json:"end_column"`
}

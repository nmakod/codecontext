package parser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/nuthan-ms/codecontext/pkg/types"
	sitter "github.com/tree-sitter/go-tree-sitter"
	javascript "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	python "github.com/tree-sitter/tree-sitter-python/bindings/go"
	java "github.com/tree-sitter/tree-sitter-java/bindings/go"
	golang "github.com/tree-sitter/tree-sitter-go/bindings/go"
	rust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
	// csharp "github.com/zzctmac/go-tree-sitter/csharp" // TODO: Fix type compatibility
)

// Manager implements the parser manager interface
type Manager struct {
	parsers           map[string]*sitter.Parser
	languages         map[string]*sitter.Language
	cache             Cache
	frameworkDetector *FrameworkDetector
	mu                sync.RWMutex
	
	// Injected dependencies
	logger       Logger
	panicHandler *PanicHandler
	config       *ParserConfig
}

// NewManager creates a new parser manager using default configuration
// Deprecated: Use NewManagerBuilder() for better control over dependencies
func NewManager() *Manager {
	manager, err := NewManagerBuilder().Build()
	if err != nil {
		// This should not happen with default configuration
		panic("failed to create manager with default configuration: " + err.Error())
	}
	return manager
}

// NewManagerWithRoot creates a new parser manager with a specified project root
// Deprecated: Use NewManagerBuilder().WithProjectRoot(root).Build() for better control
func NewManagerWithRoot(projectRoot string) *Manager {
	manager, err := NewManagerBuilder().WithProjectRoot(projectRoot).Build()
	if err != nil {
		// This should not happen with default configuration
		panic("failed to create manager with default configuration: " + err.Error())
	}
	return manager
}

// initLanguages initializes the supported languages with real Tree-sitter grammars
func (m *Manager) initLanguages() {
	// JavaScript grammar using official bindings
	jsLang := sitter.NewLanguage(javascript.Language())
	m.languages["javascript"] = jsLang

	jsParser := sitter.NewParser()
	jsParser.SetLanguage(jsLang)
	m.parsers["javascript"] = jsParser

	// TypeScript - use JavaScript grammar as fallback until TypeScript bindings are fixed
	// Both JS and TS have similar syntax and this provides basic parsing capability
	tsLang := sitter.NewLanguage(javascript.Language())
	m.languages["typescript"] = tsLang

	tsParser := sitter.NewParser()
	tsParser.SetLanguage(tsLang)
	m.parsers["typescript"] = tsParser

	// Python grammar using official bindings
	pythonLang := sitter.NewLanguage(python.Language())
	m.languages["python"] = pythonLang

	pythonParser := sitter.NewParser()
	pythonParser.SetLanguage(pythonLang)
	m.parsers["python"] = pythonParser

	// Java grammar using official bindings
	javaLang := sitter.NewLanguage(java.Language())
	m.languages["java"] = javaLang

	javaParser := sitter.NewParser()
	javaParser.SetLanguage(javaLang)
	m.parsers["java"] = javaParser

	// Go grammar using official bindings
	goLang := sitter.NewLanguage(golang.Language())
	m.languages["go"] = goLang

	goParser := sitter.NewParser()
	goParser.SetLanguage(goLang)
	m.parsers["go"] = goParser

	// Rust grammar using official bindings
	rustLang := sitter.NewLanguage(rust.Language())
	m.languages["rust"] = rustLang

	rustParser := sitter.NewParser()
	rustParser.SetLanguage(rustLang)
	m.parsers["rust"] = rustParser

	// Swift grammar - using basic parsing approach for now
	// Will be replaced with tree-sitter bindings when official Go bindings are available
	m.languages["swift"] = nil // No tree-sitter language for now
	basicSwiftParser := sitter.NewParser()
	m.parsers["swift"] = basicSwiftParser

	// Dart grammar - using regex-based approach for now
	// Will be replaced with tree-sitter bindings when available
	m.languages["dart"] = nil // No tree-sitter language for now
	basicDartParser := sitter.NewParser()
	m.parsers["dart"] = basicDartParser

	// C# grammar - temporarily disabled due to type compatibility issues
	// TODO: Fix type compatibility between official and community bindings
	// csharpLang := csharp.GetLanguage()
	// m.languages["csharp"] = csharpLang
	// csharpParser := sitter.NewParser()
	// csharpParser.SetLanguage(csharpLang)
	// m.parsers["csharp"] = csharpParser

	// For JSON and YAML, we'll use basic parsers for now
	// These can be extended with proper grammars later
	basicParser := sitter.NewParser()
	m.languages["json"] = nil
	m.languages["yaml"] = nil
	m.parsers["json"] = basicParser
	m.parsers["yaml"] = basicParser

	// Framework-specific file types use basic parsing for now
	// Framework detection is handled separately by FrameworkDetector
	frameworkParser1 := sitter.NewParser()
	frameworkParser2 := sitter.NewParser()
	frameworkParser3 := sitter.NewParser()
	m.languages["vue"] = nil
	m.languages["svelte"] = nil
	m.languages["astro"] = nil
	m.parsers["vue"] = frameworkParser1
	m.parsers["svelte"] = frameworkParser2
	m.parsers["astro"] = frameworkParser3
}

// ParseFile parses a file and returns an AST
func (m *Manager) ParseFile(filePath string, language types.Language) (*types.AST, error) {
	// Read file content from disk
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return m.parseContent(string(content), language, filePath)
}

// ParseFileVersioned parses a file with version information
func (m *Manager) ParseFileVersioned(filePath, content, version string) (*types.VersionedAST, error) {
	// Detect language
	lang := m.detectLanguage(filePath)
	if lang == nil {
		return nil, fmt.Errorf("unsupported file type: %s", filePath)
	}

	// Parse the content
	ast, err := m.parseContent(content, *lang, filePath)
	if err != nil {
		return nil, err
	}

	ast.Version = version

	versionedAST := &types.VersionedAST{
		AST:       ast,
		Version:   version,
		Hash:      calculateHash(content),
		Timestamp: time.Now(),
	}

	return versionedAST, nil
}

// ExtractSymbols extracts symbols from an AST
func (m *Manager) ExtractSymbols(ast *types.AST) ([]*types.Symbol, error) {
	if ast.Root == nil {
		return nil, fmt.Errorf("AST root is nil")
	}

	var symbols []*types.Symbol
	m.extractSymbolsRecursiveWithContent(ast.Root, ast.FilePath, ast.Language, ast.Content, &symbols)

	return symbols, nil
}

// ExtractImports extracts imports from an AST
func (m *Manager) ExtractImports(ast *types.AST) ([]*types.Import, error) {
	if ast.Root == nil {
		return nil, fmt.Errorf("AST root is nil")
	}

	var imports []*types.Import
	m.extractImportsRecursive(ast.Root, &imports)

	return imports, nil
}

// Parse parses source code content and returns an AST
func (m *Manager) Parse(content, filePath string) (*types.AST, error) {
	return m.ParseWithContext(context.Background(), content, filePath)
}

// ParseWithContext parses source code content with context for better error reporting
func (m *Manager) ParseWithContext(ctx context.Context, content, filePath string) (*types.AST, error) {
	// Add context information for better error reporting
	ctx = WithFilePath(ctx, filePath)
	
	result, err := m.panicHandler.WithOperationReturn(ctx, "parse_content", func() (any, error) {
		// Detect language from file path
		lang := m.detectLanguage(filePath)
		if lang == nil {
			return nil, NewParseError("detect_language", filePath, "", ErrUnsupportedLanguage)
		}
		
		// Add language to context
		ctx = WithLanguage(ctx, lang.Name)
		
		return m.parseContentWithContext(ctx, content, *lang, filePath)
	})
	
	if err != nil {
		return nil, err
	}
	
	if ast, ok := result.(*types.AST); ok {
		return ast, nil
	}
	
	return nil, NewParseError("parse_content", filePath, "", fmt.Errorf("internal error: unexpected return type"))
}

// GetSupportedLanguages returns the list of supported languages
func (m *Manager) GetSupportedLanguages() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var languages []string
	for name, _ := range m.languages {
		languages = append(languages, name)
	}
	
	// Add Dart support
	languages = append(languages, "dart")
	
	// Add Swift support
	languages = append(languages, "swift")

	return languages
}

// ClassifyFile classifies a file based on its path and content
func (m *Manager) ClassifyFile(filePath string) (*types.FileClassification, error) {
	ext := filepath.Ext(filePath)
	baseName := filepath.Base(filePath)

	// Detect language
	lang := m.detectLanguage(filePath)
	if lang == nil {
		return nil, fmt.Errorf("unsupported file type: %s", filePath)
	}

	// Determine file type
	fileType := "source"
	isTest := false

	if strings.Contains(strings.ToLower(baseName), "test") || strings.Contains(strings.ToLower(baseName), "spec") {
		fileType = "test"
		isTest = true
	} else if strings.Contains(baseName, "config") || ext == ".json" || ext == ".yaml" || ext == ".yml" {
		fileType = "config"
	}

	// Check if generated
	isGenerated := strings.Contains(baseName, "generated") ||
		strings.Contains(baseName, "auto") ||
		strings.HasSuffix(baseName, ".gen.ts") ||
		strings.HasSuffix(baseName, ".generated.ts")

	// Detect framework - we need file content for better detection
	var framework string
	if content, err := os.ReadFile(filePath); err == nil {
		framework = m.frameworkDetector.DetectFramework(filePath, lang.Name, string(content))
	} else {
		// Fallback to filename-based detection only
		framework = m.frameworkDetector.DetectFramework(filePath, lang.Name, "")
	}

	return &types.FileClassification{
		Language:    *lang,
		FileType:    fileType,
		IsGenerated: isGenerated,
		IsTest:      isTest,
		Framework:   framework,
		Confidence:  0.95,
	}, nil
}

// GetASTCache returns the AST cache
func (m *Manager) GetASTCache() Cache {
	return m.cache
}

// Helper methods

func (m *Manager) detectLanguage(filePath string) *types.Language {
	ext := filepath.Ext(filePath)

	switch ext {
	case ".ts", ".tsx":
		return &types.Language{
			Name:       "typescript",
			Extensions: []string{".ts", ".tsx"},
			Parser:     "tree-sitter-typescript",
			Enabled:    true,
		}
	case ".js", ".jsx":
		return &types.Language{
			Name:       "javascript",
			Extensions: []string{".js", ".jsx"},
			Parser:     "tree-sitter-javascript",
			Enabled:    true,
		}
	case ".json":
		return &types.Language{
			Name:       "json",
			Extensions: []string{".json"},
			Parser:     "tree-sitter-json",
			Enabled:    true,
		}
	case ".yaml", ".yml":
		return &types.Language{
			Name:       "yaml",
			Extensions: []string{".yaml", ".yml"},
			Parser:     "tree-sitter-yaml",
			Enabled:    true,
		}
	case ".py":
		return &types.Language{
			Name:       "python",
			Extensions: []string{".py"},
			Parser:     "tree-sitter-python",
			Enabled:    true,
		}
	case ".java":
		return &types.Language{
			Name:       "java",
			Extensions: []string{".java"},
			Parser:     "tree-sitter-java",
			Enabled:    true,
		}
	case ".go":
		return &types.Language{
			Name:       "go",
			Extensions: []string{".go"},
			Parser:     "tree-sitter-go",
			Enabled:    true,
		}
	case ".rs":
		return &types.Language{
			Name:       "rust",
			Extensions: []string{".rs"},
			Parser:     "tree-sitter-rust",
			Enabled:    true,
		}
	case ".swift":
		return &types.Language{
			Name:       "swift",
			Extensions: []string{".swift"},
			Parser:     "tree-sitter-swift",
			Enabled:    true,
		}
	case ".dart":
		return &types.Language{
			Name:       "dart",
			Extensions: []string{".dart"},
			Parser:     "tree-sitter-dart",
			Enabled:    true,
		}
	case ".vue":
		return &types.Language{
			Name:       "vue",
			Extensions: []string{".vue"},
			Parser:     "vue-template", // Framework-specific handling
			Enabled:    true,
		}
	case ".svelte":
		return &types.Language{
			Name:       "svelte",
			Extensions: []string{".svelte"},
			Parser:     "svelte-template", // Framework-specific handling
			Enabled:    true,
		}
	case ".astro":
		return &types.Language{
			Name:       "astro",
			Extensions: []string{".astro"},
			Parser:     "astro-template", // Framework-specific handling
			Enabled:    true,
		}
	// case ".cs":
	//	return &types.Language{
	//		Name:       "csharp",
	//		Extensions: []string{".cs"},
	//		Parser:     "tree-sitter-csharp",
	//		Enabled:    true,
	//	}
	default:
		return nil
	}
}

func (m *Manager) parseContent(content string, language types.Language, filePath ...string) (*types.AST, error) {
	return m.parseContentWithContext(context.Background(), content, language, filePath...)
}

func (m *Manager) parseContentWithContext(ctx context.Context, content string, language types.Language, filePath ...string) (*types.AST, error) {
	// Handle Dart specially with our custom parser
	if language.Name == "dart" {
		filePathStr := ""
		if len(filePath) > 0 {
			filePathStr = filePath[0]
		}
		return m.parseDartContentWithContext(ctx, content, filePathStr)
	}

	// Handle Swift specially with our custom parser
	if language.Name == "swift" {
		filePathStr := ""
		if len(filePath) > 0 {
			filePathStr = filePath[0]
		}
		return m.parseSwiftContentWithContext(ctx, content, filePathStr)
	}

	m.mu.RLock()
	parser, exists := m.parsers[language.Name]
	treeSitterLang := m.languages[language.Name]
	m.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("unsupported language: %s", language.Name)
	}

	// For languages without real grammars (JSON, YAML), create mock AST
	if treeSitterLang == nil {
		ast := &types.AST{
			Language:       language.Name,
			Content:        content,
			Hash:           calculateHash(content),
			Version:        "1.0",
			ParsedAt:       time.Now(),
			TreeSitterTree: nil,
		}

		if len(filePath) > 0 {
			ast.FilePath = filePath[0]
		}

		// Create a basic root node for unsupported languages
		ast.Root = &types.ASTNode{
			Id:   "root",
			Type: "document",
			Location: types.FileLocation{
				FilePath: ast.FilePath,
				Line:     1,
				Column:   1,
			},
			Value: content,
		}

		return ast, nil
	}

	// Parse using real Tree-sitter grammar
	tree := parser.Parse([]byte(content), nil)
	defer tree.Close()

	// Create AST with real Tree-sitter data
	ast := &types.AST{
		Language:       language.Name,
		Content:        content,
		Hash:           calculateHash(content),
		Version:        "1.0",
		ParsedAt:       time.Now(),
		TreeSitterTree: tree,
	}

	if len(filePath) > 0 {
		ast.FilePath = filePath[0]
	}

	// Convert Tree-sitter root node to our AST format
	if tree.RootNode() != nil {
		ast.Root = m.convertTreeSitterNode(tree.RootNode(), content)
		if ast.Root != nil {
			ast.Root.Location.FilePath = ast.FilePath
		}
	}

	return ast, nil
}

// convertTreeSitterNode converts a tree-sitter node to our AST node format
func (m *Manager) convertTreeSitterNode(node *sitter.Node, content string) *types.ASTNode {
	if node == nil {
		return nil
	}

	startPos := node.StartPosition()
	endPos := node.EndPosition()

	astNode := &types.ASTNode{
		Id:   fmt.Sprintf("node-%d-%d", node.StartByte(), node.EndByte()),
		Type: node.Kind(),
		Location: types.FileLocation{
			Line:      int(startPos.Row) + 1,
			Column:    int(startPos.Column) + 1,
			EndLine:   int(endPos.Row) + 1,
			EndColumn: int(endPos.Column) + 1,
		},
	}

	// Extract text content for the node
	if int(node.StartByte()) < len(content) && int(node.EndByte()) <= len(content) {
		astNode.Value = content[node.StartByte():node.EndByte()]
	}

	// Convert children (limit depth to prevent excessive memory usage)
	childCount := int(node.ChildCount())
	if childCount > 0 && childCount < 1000 { // Reasonable limit
		for i := 0; i < childCount; i++ {
			child := node.Child(uint(i))
			if child != nil {
				if childAST := m.convertTreeSitterNode(child, content); childAST != nil {
					astNode.Children = append(astNode.Children, childAST)
				}
			}
		}
	}

	return astNode
}

func (m *Manager) extractSymbolsRecursive(node *types.ASTNode, filePath, language string, symbols *[]*types.Symbol) {
	if node == nil {
		return
	}

	// Check if this node represents a symbol
	if symbol := m.nodeToSymbol(node, filePath, language); symbol != nil {
		*symbols = append(*symbols, symbol)
	}

	// Recursively extract from children
	for _, child := range node.Children {
		m.extractSymbolsRecursive(child, filePath, language, symbols)
	}
}

func (m *Manager) extractSymbolsRecursiveWithContent(node *types.ASTNode, filePath, language, content string, symbols *[]*types.Symbol) {
	if node == nil {
		return
	}

	// Check if this node represents a symbol
	if symbol := m.nodeToSymbolWithContent(node, filePath, language, content); symbol != nil {
		*symbols = append(*symbols, symbol)
	}

	// Recursively extract from children
	for _, child := range node.Children {
		m.extractSymbolsRecursiveWithContent(child, filePath, language, content, symbols)
	}
}

func (m *Manager) nodeToSymbol(node *types.ASTNode, filePath, language string) *types.Symbol {
	return m.nodeToSymbolWithContent(node, filePath, language, "")
}

func (m *Manager) nodeToSymbolWithContent(node *types.ASTNode, filePath, language, content string) *types.Symbol {
	// First check for framework-specific symbols  
	if frameworkSymbol := m.extractFrameworkSymbolWithContent(node, filePath, language, content); frameworkSymbol != nil {
		return frameworkSymbol
	}

	// Language-specific symbol extraction using real Tree-sitter node types
	switch language {
	case "dart":
		return m.nodeToSymbolDart(node, filePath, language)
	case "python":
		return m.nodeToSymbolPython(node, filePath, language)
	case "java":
		return m.nodeToSymbolJava(node, filePath, language)
	case "go":
		return m.nodeToSymbolGo(node, filePath, language)
	case "rust":
		return m.nodeToSymbolRust(node, filePath, language)
	case "swift":
		return m.nodeToSymbolSwift(node, filePath, language)
	case "vue", "svelte", "astro":
		// Framework-specific files are treated as JavaScript/TypeScript for parsing
		return m.nodeToSymbolJS(node, filePath, language)
	// case "csharp":
	//	return m.nodeToSymbolCSharp(node, filePath, language)
	default:
		// Default JavaScript/TypeScript handling
		return m.nodeToSymbolJS(node, filePath, language)
	}
}

func (m *Manager) nodeToSymbolJS(node *types.ASTNode, filePath, language string) *types.Symbol {
	// Enhanced symbol extraction for JavaScript/TypeScript using real Tree-sitter node types
	switch node.Type {
	case "function_declaration", "function", "function_expression", "arrow_function":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("func-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeFunction,
			Location:     convertLocation(node.Location),
			Signature:    m.extractFunctionSignature(node),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "class_declaration", "class", "class_expression":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("class-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeClass,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "interface_declaration", "interface":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("interface-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeInterface,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "type_alias_declaration", "type_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("type-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeType,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "variable_declaration", "lexical_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("var-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeVariable,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "method_definition", "method_signature":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("method-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeMethod,
			Location:     convertLocation(node.Location),
			Signature:    m.extractFunctionSignature(node),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "import_statement", "import_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("import-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractImportName(node),
			Type:         types.SymbolTypeImport,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "export_statement", "export_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("export-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeNamespace, // Using namespace for exports
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	default:
		return nil
	}
}

// nodeToSymbolPython extracts symbols for Python language
func (m *Manager) nodeToSymbolPython(node *types.ASTNode, filePath, language string) *types.Symbol {
	switch node.Type {
	case "function_definition":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("func-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeFunction,
			Location:     convertLocation(node.Location),
			Signature:    m.extractFunctionSignature(node),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "class_definition":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("class-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeClass,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "import_statement", "import_from_statement":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("import-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractImportName(node),
			Type:         types.SymbolTypeImport,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "assignment":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("var-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeVariable,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	default:
		return nil
	}
}

// nodeToSymbolJava extracts symbols for Java language
func (m *Manager) nodeToSymbolJava(node *types.ASTNode, filePath, language string) *types.Symbol {
	switch node.Type {
	case "method_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("method-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeMethod,
			Location:     convertLocation(node.Location),
			Signature:    m.extractFunctionSignature(node),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "class_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("class-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeClass,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "interface_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("interface-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeInterface,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "field_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("field-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeVariable,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "import_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("import-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractImportName(node),
			Type:         types.SymbolTypeImport,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	default:
		return nil
	}
}

// nodeToSymbolGo extracts symbols for Go language
func (m *Manager) nodeToSymbolGo(node *types.ASTNode, filePath, language string) *types.Symbol {
	switch node.Type {
	case "function_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("func-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeFunction,
			Location:     convertLocation(node.Location),
			Signature:    m.extractFunctionSignature(node),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "method_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("method-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeMethod,
			Location:     convertLocation(node.Location),
			Signature:    m.extractFunctionSignature(node),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "type_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("type-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeType,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "var_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("var-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeVariable,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "import_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("import-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractImportName(node),
			Type:         types.SymbolTypeImport,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	default:
		return nil
	}
}

// nodeToSymbolRust extracts symbols for Rust language
func (m *Manager) nodeToSymbolRust(node *types.ASTNode, filePath, language string) *types.Symbol {
	switch node.Type {
	case "function_item":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("func-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeFunction,
			Location:     convertLocation(node.Location),
			Signature:    m.extractFunctionSignature(node),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "impl_item":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("impl-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeClass,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "struct_item":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("struct-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeClass,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "enum_item":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("enum-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeType,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "trait_item":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("trait-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeInterface,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "use_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("import-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractImportName(node),
			Type:         types.SymbolTypeImport,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	default:
		return nil
	}
}

// nodeToSymbolSwift extracts symbols for Swift language
func (m *Manager) nodeToSymbolSwift(node *types.ASTNode, filePath, language string) *types.Symbol {
	switch node.Type {
	case "class_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("class-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeClass,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "struct_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("struct-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeClass,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "protocol_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("protocol-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeInterface,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "actor_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("actor-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeClass,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "typealias_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("typealias-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeType,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "function_declaration":
		// Determine if it's a method (inside a class/struct) or top-level function
		symbolType := types.SymbolTypeFunction
		if strings.Contains(node.Value, "    func") || strings.Contains(node.Value, "\tfunc") {
			// Indented function = method
			symbolType = types.SymbolTypeMethod
		}
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("func-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         symbolType,
			Location:     convertLocation(node.Location),
			Signature:    m.extractFunctionSignature(node),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "init_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("init-%s-%d", filePath, node.Location.Line)),
			Name:         "init",
			Type:         types.SymbolTypeMethod,
			Location:     convertLocation(node.Location),
			Signature:    m.extractFunctionSignature(node),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "deinit_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("deinit-%s-%d", filePath, node.Location.Line)),
			Name:         "deinit",
			Type:         types.SymbolTypeMethod,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "import_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("import-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractImportName(node),
			Type:         types.SymbolTypeImport,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "extension_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("extension-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeNamespace,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "property_declaration":
		// Determine property type based on metadata
		symbolType := types.SymbolTypeProperty
		if node.Metadata != nil {
			if _, isWrapped := node.Metadata["is_wrapped"]; isWrapped {
				symbolType = types.SymbolTypeProperty
			} else if _, isComputed := node.Metadata["is_computed"]; isComputed {
				symbolType = types.SymbolTypeProperty
			}
		}
		
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("property-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         symbolType,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "associatedtype_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("associatedtype-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeType,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "async_function_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("async-func-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeFunction,
			Location:     convertLocation(node.Location),
			Signature:    m.extractFunctionSignature(node),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "subscript_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("subscript-%s-%d", filePath, node.Location.Line)),
			Name:         "subscript",
			Type:         types.SymbolTypeMethod,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "operator_function_declaration":
		operatorSymbol := "operator"
		if node.Metadata != nil {
			if op, exists := node.Metadata["operator_symbol"]; exists {
				operatorSymbol = op.(string)
			}
		}
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("operator-%s-%d", filePath, node.Location.Line)),
			Name:         operatorSymbol,
			Type:         types.SymbolTypeFunction,
			Location:     convertLocation(node.Location),
			Signature:    m.extractFunctionSignature(node),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "result_builder_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("result-builder-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeClass,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "macro_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("macro-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeFunction,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	default:
		return nil
	}
}

// nodeToSymbolCSharp extracts symbols for C# language
func (m *Manager) nodeToSymbolCSharp(node *types.ASTNode, filePath, language string) *types.Symbol {
	switch node.Type {
	case "method_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("method-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeMethod,
			Location:     convertLocation(node.Location),
			Signature:    m.extractFunctionSignature(node),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "class_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("class-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeClass,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "interface_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("interface-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeInterface,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "struct_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("struct-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeClass,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "enum_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("enum-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeType,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "property_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("property-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeVariable,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "using_directive":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("import-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractImportName(node),
			Type:         types.SymbolTypeImport,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	default:
		return nil
	}
}

func (m *Manager) extractImportsRecursive(node *types.ASTNode, imports *[]*types.Import) {
	if node == nil {
		return
	}

	// Check if this node represents an import
	if imp := m.nodeToImport(node); imp != nil {
		*imports = append(*imports, imp)
	}

	// Recursively extract from children
	for _, child := range node.Children {
		m.extractImportsRecursive(child, imports)
	}
}

func (m *Manager) nodeToImport(node *types.ASTNode) *types.Import {
	if node.Type != "import_statement" && node.Type != "import_declaration" {
		return nil
	}

	imp := &types.Import{
		Location: node.Location,
	}

	// Extract import path and specifiers from children
	for _, child := range node.Children {
		switch child.Type {
		case "string", "string_literal":
			imp.Path = strings.Trim(child.Value, `"'`)
		case "import_specifier":
			if name := m.extractSymbolName(child); name != "unknown" {
				imp.Specifiers = append(imp.Specifiers, name)
			}
		case "namespace_import":
			if name := m.extractSymbolName(child); name != "unknown" {
				imp.Alias = name
			}
		case "identifier":
			// Default import
			imp.IsDefault = true
			if name := strings.TrimSpace(child.Value); name != "" {
				imp.Specifiers = append(imp.Specifiers, name)
			}
		}
	}

	return imp
}

func (m *Manager) getExtensionsForLanguage(name string) []string {
	switch name {
	case "typescript":
		return []string{".ts", ".tsx"}
	case "javascript":
		return []string{".js", ".jsx"}
	case "dart":
		return []string{".dart"}
	case "swift":
		return []string{".swift"}
	default:
		return []string{}
	}
}

// Helper functions

func calculateHash(content string) string {
	// Simple hash implementation - in production, use crypto/sha256
	return fmt.Sprintf("hash-%d", len(content))
}

// convertLocation converts FileLocation to new Location type for diff compatibility
func convertLocation(loc types.FileLocation) types.Location {
	return types.Location{
		StartLine:   loc.Line,
		StartColumn: loc.Column,
		EndLine:     loc.Line,
		EndColumn:   loc.Column + 10, // Approximate end column
	}
}

// extractSymbolName extracts the name from a symbol node
func (m *Manager) extractSymbolName(node *types.ASTNode) string {
	if node == nil {
		return "unknown"
	}

	// Look for identifier children that represent the symbol name
	for _, child := range node.Children {
		if child.Type == "identifier" || child.Type == "type_identifier" {
			return strings.TrimSpace(child.Value)
		}

		// For some nodes, the name might be nested deeper
		if child.Type == "property_identifier" || child.Type == "name" {
			return strings.TrimSpace(child.Value)
		}
		
		// For part directives, the filename is stored in string_literal children
		if child.Type == "string_literal" {
			return strings.TrimSpace(child.Value)
		}
	}

	// Fallback: extract from node value using heuristics
	value := strings.TrimSpace(node.Value)
	if value != "" {
		// Try to extract first word after keywords
		lines := strings.Split(value, "\n")
		if len(lines) > 0 {
			firstLine := strings.TrimSpace(lines[0])
			words := strings.Fields(firstLine)

			// Look for name after common keywords
			for i, word := range words {
				if word == "function" || word == "class" || word == "interface" ||
					word == "type" || word == "const" || word == "let" || word == "var" {
					if i+1 < len(words) {
						name := strings.TrimSuffix(words[i+1], "(")
						name = strings.TrimSuffix(name, "{")
						return name
					}
				}
			}

			// If no keyword found, return first identifier-like word
			for _, word := range words {
				if isValidIdentifier(word) {
					return word
				}
			}
		}
	}

	return "unknown"
}

// extractFunctionSignature extracts function signature information
func (m *Manager) extractFunctionSignature(node *types.ASTNode) string {
	if node == nil {
		return ""
	}

	// Look for parameter list and return type
	for _, child := range node.Children {
		if child.Type == "formal_parameters" || child.Type == "parameters" {
			return strings.TrimSpace(child.Value)
		}
	}

	// Fallback: extract first line of the node
	value := strings.TrimSpace(node.Value)
	lines := strings.Split(value, "\n")
	if len(lines) > 0 {
		signature := strings.TrimSpace(lines[0])
		// Remove opening brace if present
		if idx := strings.Index(signature, "{"); idx != -1 {
			signature = strings.TrimSpace(signature[:idx])
		}
		return signature
	}

	return ""
}

// extractImportName extracts name from import nodes
func (m *Manager) extractImportName(node *types.ASTNode) string {
	if node == nil {
		return "unknown"
	}

	// Look for import specifiers
	for _, child := range node.Children {
		if child.Type == "import_specifier" || child.Type == "namespace_import" {
			if name := m.extractSymbolName(child); name != "unknown" {
				return name
			}
		}
		if child.Type == "identifier" {
			// Direct identifier (Swift imports)
			return strings.TrimSpace(child.Value)
		}
		if child.Type == "string" || child.Type == "string_literal" {
			// This is likely the import path
			path := strings.Trim(child.Value, `"'`)
			// Extract module name from path
			if lastSlash := strings.LastIndex(path, "/"); lastSlash != -1 {
				return path[lastSlash+1:]
			}
			return path
		}
	}

	return "unknown"
}

// isValidIdentifier checks if a string looks like a valid identifier
func isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}

	// Remove common symbols
	s = strings.TrimSuffix(s, "(")
	s = strings.TrimSuffix(s, "{")
	s = strings.TrimSuffix(s, ":")
	s = strings.TrimSuffix(s, ";")

	// Check if it looks like an identifier
	if len(s) == 0 {
		return false
	}

	// Simple heuristic: starts with letter or underscore, contains only alphanumeric and underscore
	first := s[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_' || first == '$') {
		return false
	}

	for _, r := range s[1:] {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '$') {
			return false
		}
	}

	return true
}

// extractFrameworkSymbol detects and extracts framework-specific symbols
func (m *Manager) extractFrameworkSymbol(node *types.ASTNode, filePath, language string) *types.Symbol {
	return m.extractFrameworkSymbolWithContent(node, filePath, language, "")
}

// extractFrameworkSymbolWithContent detects and extracts framework-specific symbols with provided content
func (m *Manager) extractFrameworkSymbolWithContent(node *types.ASTNode, filePath, language, providedContent string) *types.Symbol {
	// Get framework from content analysis
	var content string
	if providedContent != "" {
		content = providedContent
	} else if fileContent, err := os.ReadFile(filePath); err == nil {
		content = string(fileContent)
	} else {
		// For test scenarios where the file doesn't exist, try to get content from the AST
		content = node.Value
		if content == "" {
			// If we can't get content, we can't detect framework-specific symbols
			return nil
		}
	}
	framework := m.frameworkDetector.DetectFramework(filePath, language, content)

	switch framework {
	case "React":
		return m.extractReactSymbol(node, filePath, language, content)
	case "Vue":
		return m.extractVueSymbol(node, filePath, language, content)
	case "Angular":
		return m.extractAngularSymbol(node, filePath, language, content)
	case "Svelte":
		return m.extractSvelteSymbol(node, filePath, language, content)
	case "Next.js":
		return m.extractNextJSSymbol(node, filePath, language, content)
	case "Astro":
		return m.extractAstroSymbol(node, filePath, language, content)
	default:
		return nil
	}
}

// extractReactSymbol extracts React-specific symbols
func (m *Manager) extractReactSymbol(node *types.ASTNode, filePath, language, content string) *types.Symbol {
	// React Component Detection
	if m.isReactComponent(node, content) {
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("react-component-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeComponent,
			Location:     convertLocation(node.Location),
			Signature:    m.extractComponentProps(node),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	}

	// React Hook Detection
	if m.isReactHook(node, content) {
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("react-hook-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeHook,
			Location:     convertLocation(node.Location),
			Signature:    m.extractFunctionSignature(node),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	}

	return nil
}

// extractVueSymbol extracts Vue-specific symbols
func (m *Manager) extractVueSymbol(node *types.ASTNode, filePath, language, content string) *types.Symbol {
	// Vue Component Detection
	if m.isVueComponent(node, content) {
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("vue-component-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeComponent,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	}

	// Vue Computed Property Detection
	if m.isVueComputed(node, content) {
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("vue-computed-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeComputed,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	}

	// Vue Watcher Detection
	if m.isVueWatcher(node, content) {
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("vue-watcher-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeWatcher,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	}

	return nil
}

// extractAngularSymbol extracts Angular-specific symbols
func (m *Manager) extractAngularSymbol(node *types.ASTNode, filePath, language, content string) *types.Symbol {
	// Angular Component Detection
	if m.isAngularComponent(node, content) {
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("angular-component-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeComponent,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	}

	// Angular Service Detection
	if m.isAngularService(node, content) {
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("angular-service-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeService,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	}

	// Angular Directive Detection
	if m.isAngularDirective(node, content) {
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("angular-directive-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeDirective,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	}

	return nil
}

// extractSvelteSymbol extracts Svelte-specific symbols
func (m *Manager) extractSvelteSymbol(node *types.ASTNode, filePath, language, content string) *types.Symbol {
	// Svelte Component Detection
	if m.isSvelteComponent(node, content) {
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("svelte-component-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeComponent,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	}

	// Svelte Store Detection
	if m.isSvelteStore(node, content) {
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("svelte-store-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeStore,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	}

	// Svelte Action Detection
	if m.isSvelteAction(node, content) {
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("svelte-action-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeAction,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	}

	return nil
}

// extractNextJSSymbol extracts Next.js-specific symbols
func (m *Manager) extractNextJSSymbol(node *types.ASTNode, filePath, language, content string) *types.Symbol {
	// Next.js Page Detection
	if m.isNextJSPage(node, filePath, content) {
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("nextjs-page-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeRoute,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	}

	// Next.js API Route Detection
	if m.isNextJSAPIRoute(node, filePath, content) {
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("nextjs-api-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeRoute,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	}

	// Next.js Middleware Detection
	if m.isNextJSMiddleware(node, filePath, content) {
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("nextjs-middleware-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeMiddleware,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	}

	return nil
}

// extractAstroSymbol extracts Astro-specific symbols
func (m *Manager) extractAstroSymbol(node *types.ASTNode, filePath, language, content string) *types.Symbol {
	// Astro Component Detection
	if m.isAstroComponent(node, content) {
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("astro-component-%s-%d", filePath, node.Location.Line)),
			Name:         m.extractSymbolName(node),
			Type:         types.SymbolTypeComponent,
			Location:     convertLocation(node.Location),
			Language:     language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	}

	return nil
}

// React Helper Functions

// isReactComponent checks if a node represents a React component
func (m *Manager) isReactComponent(node *types.ASTNode, content string) bool {
	if node.Type == "function_declaration" || node.Type == "function_expression" || node.Type == "arrow_function" {
		// Check if function returns JSX
		return strings.Contains(node.Value, "return") && (strings.Contains(node.Value, "<") && strings.Contains(node.Value, "/>")) ||
			(strings.Contains(node.Value, "<") && strings.Contains(node.Value, "</"))
	}
	if node.Type == "class_declaration" && strings.Contains(content, "React.Component") {
		return true
	}
	return false
}

// isReactHook checks if a node represents a React hook
func (m *Manager) isReactHook(node *types.ASTNode, content string) bool {
	if node.Type == "function_declaration" || node.Type == "function_expression" || node.Type == "arrow_function" {
		name := m.extractSymbolName(node)
		return strings.HasPrefix(name, "use") && len(name) > 3 && 
			(name[3] >= 'A' && name[3] <= 'Z') // Starts with "use" followed by capital letter
	}
	return false
}

// extractComponentProps extracts component props signature
func (m *Manager) extractComponentProps(node *types.ASTNode) string {
	for _, child := range node.Children {
		if child.Type == "formal_parameters" || child.Type == "parameters" {
			return strings.TrimSpace(child.Value)
		}
	}
	return ""
}

// Vue Helper Functions

// isVueComponent checks if a node represents a Vue component
func (m *Manager) isVueComponent(node *types.ASTNode, content string) bool {
	// Check for Vue 3 Composition API
	if node.Type == "function_declaration" && strings.Contains(content, "defineComponent") {
		return true
	}
	// Check for Vue SFC script setup
	if strings.Contains(content, "<script setup>") {
		return true
	}
	// Check for export default with Vue options
	if node.Type == "export_statement" && strings.Contains(node.Value, "export default") {
		return strings.Contains(node.Value, "template:") || strings.Contains(node.Value, "render:") ||
			strings.Contains(node.Value, "data()") || strings.Contains(node.Value, "computed:")
	}
	return false
}

// isVueComputed checks if a node represents a Vue computed property
func (m *Manager) isVueComputed(node *types.ASTNode, content string) bool {
	// Check within computed object
	if node.Type == "property" || node.Type == "method_definition" {
		parent := m.findParentWithType(node, "object")
		if parent != nil && strings.Contains(parent.Value, "computed:") {
			return true
		}
	}
	// Check for Composition API computed
	if (node.Type == "variable_declaration" || node.Type == "lexical_declaration") && 
		strings.Contains(node.Value, "computed(") {
		return true
	}
	return false
}

// isVueWatcher checks if a node represents a Vue watcher
func (m *Manager) isVueWatcher(node *types.ASTNode, content string) bool {
	// Check within watch object
	if node.Type == "property" || node.Type == "method_definition" {
		parent := m.findParentWithType(node, "object")
		if parent != nil && strings.Contains(parent.Value, "watch:") {
			return true
		}
	}
	// Check for Composition API watch
	if (node.Type == "variable_declaration" || node.Type == "lexical_declaration") && 
		(strings.Contains(node.Value, "watch(") || strings.Contains(node.Value, "watchEffect(")) {
		return true
	}
	return false
}

// Angular Helper Functions

// isAngularComponent checks if a node represents an Angular component
func (m *Manager) isAngularComponent(node *types.ASTNode, content string) bool {
	if node.Type == "class_declaration" {
		// Check for @Component decorator
		return strings.Contains(content, "@Component") && 
			strings.Contains(node.Value, "class") &&
			(strings.Contains(content, "templateUrl:") || strings.Contains(content, "template:"))
	}
	return false
}

// isAngularService checks if a node represents an Angular service
func (m *Manager) isAngularService(node *types.ASTNode, content string) bool {
	if node.Type == "class_declaration" {
		// Check for @Injectable decorator
		return strings.Contains(content, "@Injectable") && strings.Contains(node.Value, "class")
	}
	return false
}

// isAngularDirective checks if a node represents an Angular directive
func (m *Manager) isAngularDirective(node *types.ASTNode, content string) bool {
	if node.Type == "class_declaration" {
		// Check for @Directive decorator
		return strings.Contains(content, "@Directive") && strings.Contains(node.Value, "class")
	}
	return false
}

// Svelte Helper Functions

// isSvelteComponent checks if a node represents a Svelte component
func (m *Manager) isSvelteComponent(node *types.ASTNode, content string) bool {
	// Svelte components are typically the entire .svelte file
	return strings.HasSuffix(node.Location.FilePath, ".svelte") && node.Type == "document"
}

// isSvelteStore checks if a node represents a Svelte store
func (m *Manager) isSvelteStore(node *types.ASTNode, content string) bool {
	if node.Type == "variable_declaration" || node.Type == "lexical_declaration" {
		return strings.Contains(node.Value, "writable(") || 
			strings.Contains(node.Value, "readable(") ||
			strings.Contains(node.Value, "derived(")
	}
	return false
}

// isSvelteAction checks if a node represents a Svelte action
func (m *Manager) isSvelteAction(node *types.ASTNode, content string) bool {
	if node.Type == "function_declaration" || node.Type == "function_expression" {
		// Svelte actions typically take a node parameter and return an object with destroy method
		return strings.Contains(node.Value, "destroy") && 
			(strings.Contains(node.Value, "node") || strings.Contains(node.Value, "element"))
	}
	return false
}

// Next.js Helper Functions

// isNextJSPage checks if a node represents a Next.js page
func (m *Manager) isNextJSPage(node *types.ASTNode, filePath, content string) bool {
	// Check if file is in pages directory or app directory
	if strings.Contains(filePath, "/pages/") || strings.Contains(filePath, "/app/") {
		if node.Type == "export_statement" && strings.Contains(node.Value, "export default") {
			return true
		}
		if node.Type == "function_declaration" && strings.Contains(content, "export default") {
			return true
		}
	}
	return false
}

// isNextJSAPIRoute checks if a node represents a Next.js API route
func (m *Manager) isNextJSAPIRoute(node *types.ASTNode, filePath, content string) bool {
	// Check if file is in pages/api or app/api directory
	if strings.Contains(filePath, "/pages/api/") || strings.Contains(filePath, "/app/api/") {
		if node.Type == "export_statement" && 
			(strings.Contains(node.Value, "GET") || strings.Contains(node.Value, "POST") ||
				strings.Contains(node.Value, "PUT") || strings.Contains(node.Value, "DELETE")) {
			return true
		}
		if node.Type == "function_declaration" && 
			(strings.Contains(content, "req") && strings.Contains(content, "res")) {
			return true
		}
	}
	return false
}

// isNextJSMiddleware checks if a node represents Next.js middleware
func (m *Manager) isNextJSMiddleware(node *types.ASTNode, filePath, content string) bool {
	// Check if file is middleware.ts/js
	if strings.Contains(filePath, "middleware.") {
		if node.Type == "export_statement" && strings.Contains(node.Value, "middleware") {
			return true
		}
		if node.Type == "function_declaration" && strings.Contains(content, "NextRequest") {
			return true
		}
	}
	return false
}

// Astro Helper Functions

// isAstroComponent checks if a node represents an Astro component
func (m *Manager) isAstroComponent(node *types.ASTNode, content string) bool {
	// Astro components are typically the entire .astro file
	return strings.HasSuffix(node.Location.FilePath, ".astro") && node.Type == "document"
}

// Helper function to find parent node with specific type
func (m *Manager) findParentWithType(node *types.ASTNode, targetType string) *types.ASTNode {
	// This is a simplified implementation - in a real scenario, you'd need to maintain parent references
	// For now, we'll return nil as this would require AST traversal changes
	return nil
}

// Close performs cleanup when shutting down the parser manager
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Clear cache
	if m.cache != nil {
		if err := m.cache.Clear(); err != nil {
			return fmt.Errorf("failed to clear cache: %w", err)
		}
	}
	
	// Close all parsers
	for lang, parser := range m.parsers {
		if parser != nil {
			// Tree-sitter parsers don't have a Close method, but we can clean up references
			delete(m.parsers, lang)
		}
	}
	
	// Clear languages map
	for lang := range m.languages {
		delete(m.languages, lang)
	}
	
	return nil
}

// GetParser returns a parser for the specified language
func (m *Manager) GetParser(language string) (Parser, error) {
	// For now, return self as all parsing goes through Manager
	// In a more sophisticated implementation, we might return language-specific parsers
	if m.detectLanguage(fmt.Sprintf("test.%s", getExtensionForLanguage(language))) != nil {
		return m, nil
	}
	return nil, fmt.Errorf("unsupported language: %s", language)
}

// RegisterParser registers a new parser for a language
func (m *Manager) RegisterParser(language string, parser Parser) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// For now, we don't support external parser registration as we use tree-sitter directly
	return fmt.Errorf("external parser registration not supported for language: %s", language)
}

// SetCache configures the cache implementation
func (m *Manager) SetCache(cache Cache) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if astCache, ok := cache.(*ASTCache); ok {
		m.cache = astCache
	}
}

// SetLogger configures the logger implementation  
func (m *Manager) SetLogger(logger Logger) {
	// For now, we don't have structured logging implemented
	// This would be implemented when we add proper logging
}

// SetMetrics configures the metrics implementation
func (m *Manager) SetMetrics(metrics Metrics) {
	// For now, we don't have metrics implemented
	// This would be implemented when we add proper metrics collection
}

// SetConfig updates the parser configuration
func (m *Manager) SetConfig(config *ParserConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Apply cache configuration
	if m.cache != nil {
		m.cache.SetMaxSize(config.Cache.MaxSize)
		m.cache.SetTTL(config.Cache.TTL)
	}
	
	return nil
}



// getExtensionForLanguage returns the primary file extension for a language
func getExtensionForLanguage(language string) string {
	switch language {
	case "typescript":
		return "ts"
	case "javascript":
		return "js"
	case "python":
		return "py"
	case "java":
		return "java"
	case "go":
		return "go"
	case "rust":
		return "rs"
	case "dart":
		return "dart"
	default:
		return language
	}
}

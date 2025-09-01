package parser

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nuthan-ms/codecontext/pkg/types"
	sitter "github.com/tree-sitter/go-tree-sitter"
	cpp "github.com/tree-sitter/tree-sitter-cpp/bindings/go"
)

// Custom error types for better error categorization
type CppParserError struct {
	Type    string
	Message string
	Cause   error
}

func (e *CppParserError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("C++ parser %s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("C++ parser %s: %s", e.Type, e.Message)
}

func (e *CppParserError) Unwrap() error {
	return e.Cause
}

// Error type constructors
func NewInitializationError(message string, cause error) error {
	return &CppParserError{Type: "initialization", Message: message, Cause: cause}
}

func NewParsingError(message string, cause error) error {
	return &CppParserError{Type: "parsing", Message: message, Cause: cause}
}

func NewValidationError(message string) error {
	return &CppParserError{Type: "validation", Message: message}
}

func NewASTError(message string, cause error) error {
	return &CppParserError{Type: "ast", Message: message, Cause: cause}
}

// CppParser handles C++ specific parsing logic
type CppParser struct {
	parser   *sitter.Parser
	language *sitter.Language
	logger   Logger
	config   *ParserConfig
}

// NewCppParser creates a new C++ parser with error handling
func NewCppParser(logger Logger) (*CppParser, error) {
	return NewCppParserWithConfig(logger, DefaultConfig())
}

// NewCppParserWithConfig creates a new C++ parser with custom configuration
func NewCppParserWithConfig(logger Logger, config *ParserConfig) (*CppParser, error) {
	if logger == nil {
		logger = NopLogger{} // Safe default
	}
	if config == nil {
		config = DefaultConfig()
	}
	
	logger.Debug("initializing C++ parser", 
		LogField{Key: "component", Value: "cpp_parser"},
		LogField{Key: "max_nesting_depth", Value: config.Cpp.MaxNestingDepth},
		LogField{Key: "parse_timeout", Value: config.Cpp.ParseTimeout})
	
	cppLang := sitter.NewLanguage(cpp.Language())
	if cppLang == nil {
		err := NewInitializationError("failed to initialize C++ tree-sitter language", nil)
		logger.Error("failed to initialize C++ language", err)
		return nil, err
	}
	
	cppParser := sitter.NewParser()
	if cppParser == nil {
		err := NewInitializationError("failed to create tree-sitter parser", nil)
		logger.Error("failed to create parser", err)
		return nil, err
	}
	
	cppParser.SetLanguage(cppLang)
	
	logger.Info("C++ parser initialized successfully", 
		LogField{Key: "config_validation", Value: "passed"})

	return &CppParser{
		parser:   cppParser,
		language: cppLang,
		logger:   logger,
		config:   config,
	}, nil
}

// ParseContent parses C++ content and returns an AST with enhanced feature detection
func (cp *CppParser) ParseContent(ctx context.Context, content, filePath string) (*types.AST, error) {
	if cp == nil {
		return nil, NewValidationError("CppParser is nil")
	}
	
	start := time.Now()
	cp.logger.Debug("starting C++ content parsing", 
		LogField{Key: "file", Value: filePath},
		LogField{Key: "content_size", Value: len(content)})
	
	// Validate inputs and apply limits
	if err := cp.validateInputs(content, filePath); err != nil {
		return nil, err
	}
	
	// Parse content with tree-sitter
	tree, err := cp.parseWithTreeSitter(ctx, content, filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if tree != nil {
			tree.Close()
			cp.logger.Debug("tree-sitter resources cleaned up", LogField{Key: "file", Value: filePath})
		}
	}()
	
	// Build and return AST
	ast := cp.buildAST(tree, content, filePath, start)
	
	parseTime := time.Since(start)
	cp.logger.Info("C++ parsing completed", 
		LogField{Key: "file", Value: filePath},
		LogField{Key: "parse_time", Value: parseTime},
		LogField{Key: "content_size", Value: len(content)})

	return ast, nil
}

// validateInputs validates parser inputs and applies configuration limits
func (cp *CppParser) validateInputs(content, filePath string) error {
	if cp.parser == nil {
		err := NewValidationError("tree-sitter parser is nil")
		cp.logger.Error("parser validation failed", err)
		return err
	}
	if content == "" {
		err := NewValidationError("content is empty")
		cp.logger.Error("empty content provided", err, LogField{Key: "file", Value: filePath})
		return err
	}
	
	// Apply configuration limits
	if len(content) > cp.config.Cpp.MaxFileSize {
		err := NewValidationError(fmt.Sprintf("file too large: %d > %d bytes", len(content), cp.config.Cpp.MaxFileSize))
		cp.logger.Error("file size limit exceeded", err, 
			LogField{Key: "file", Value: filePath},
			LogField{Key: "size", Value: len(content)},
			LogField{Key: "limit", Value: cp.config.Cpp.MaxFileSize})
		return err
	}
	
	return nil
}

// parseWithTreeSitter performs the actual tree-sitter parsing with timeout monitoring
func (cp *CppParser) parseWithTreeSitter(ctx context.Context, content, filePath string) (*sitter.Tree, error) {
	// Create parsing context with timeout
	parseCtx, cancel := context.WithTimeout(ctx, cp.config.Cpp.ParseTimeout)
	defer cancel()
	
	cp.logger.Debug("parsing with tree-sitter", 
		LogField{Key: "file", Value: filePath},
		LogField{Key: "timeout", Value: cp.config.Cpp.ParseTimeout})
	
	// Check if context is already cancelled
	select {
	case <-parseCtx.Done():
		err := NewParsingError("parsing cancelled before start", parseCtx.Err())
		cp.logger.Error("parsing context cancelled", err, LogField{Key: "file", Value: filePath})
		return nil, err
	default:
	}
	
	parseStart := time.Now()
	tree := cp.parser.Parse([]byte(content), nil)
	parseTime := time.Since(parseStart)
	
	// Check if parsing took too long
	if parseTime > cp.config.Cpp.ParseTimeout {
		timeoutErr := NewParsingError("parsing exceeded timeout", nil)
		cp.logger.Error("parsing exceeded timeout", timeoutErr,
			LogField{Key: "file", Value: filePath},
			LogField{Key: "parse_time", Value: parseTime},
			LogField{Key: "timeout", Value: cp.config.Cpp.ParseTimeout})
		
		// Strict timeout enforcement option
		if cp.config.Cpp.StrictTimeoutEnforcement {
			return nil, timeoutErr
		}
	}
	
	if tree == nil {
		err := NewParsingError("failed to parse content with tree-sitter", nil)
		cp.logger.Error("tree-sitter parsing failed", err, LogField{Key: "file", Value: filePath})
		return nil, err
	}
	
	return tree, nil
}

// buildAST creates the AST structure from the parsed tree
func (cp *CppParser) buildAST(tree *sitter.Tree, content, filePath string, startTime time.Time) *types.AST {
	// Create AST with real Tree-sitter data
	ast := &types.AST{
		Language:       "cpp",
		Content:        content,
		Hash:           calculateHash(content),
		Version:        "1.0",
		ParsedAt:       startTime,
		TreeSitterTree: tree,
		FilePath:       filePath,
	}

	// Convert Tree-sitter root node to our AST format
	if tree.RootNode() != nil {
		ast.Root = cp.convertTreeSitterNode(tree.RootNode(), content)
		if ast.Root != nil {
			ast.Root.Location.FilePath = ast.FilePath

			// Add comprehensive C++ feature detection metadata
			ast.Root.Metadata = cp.detectCppFeatures(tree.RootNode(), content)
		}
	}

	return ast
}

// NodeToSymbol extracts C++ symbols from AST nodes with enhanced classification
func (cp *CppParser) NodeToSymbol(node *types.ASTNode, filePath, language, content string, parentContext *CppParentContext) *types.Symbol {
	// Legacy method - delegates to new context-based method
	ctx := &SymbolExtractionContext{
		FilePath:  filePath,
		Language:  language,
		Content:   content,
		ParentCtx: parentContext,
	}
	return cp.NodeToSymbolWithContext(node, ctx)
}

// NodeToSymbolWithContext extracts C++ symbols using context struct for better maintainability
func (cp *CppParser) NodeToSymbolWithContext(node *types.ASTNode, ctx *SymbolExtractionContext) *types.Symbol {
	if node == nil || ctx == nil {
		return nil
	}
	switch node.Type {
	case "class_specifier":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("class-%s-%d", ctx.FilePath, node.Location.Line)),
			Name:         cp.extractCppClassName(node),
			Type:         types.SymbolTypeClass,
			Location:     convertLocation(node.Location),
			Language:     ctx.Language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
			Visibility:   cp.extractVisibility(node, ctx.ParentCtx),
		}
	case "struct_specifier":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("struct-%s-%d", ctx.FilePath, node.Location.Line)),
			Name:         cp.extractCppClassName(node),
			Type:         types.SymbolTypeClass,
			Location:     convertLocation(node.Location),
			Language:     ctx.Language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
			Visibility:   "public", // structs default to public
		}
	case "function_definition", "function_declaration", "declaration":
		// Check if this declaration is a function/constructor/destructor
		if node.Type == "declaration" && !cp.isFunctionDeclaration(node) {
			return nil // Not a function declaration
		}
		
		symbolType, visibility := cp.classifyFunction(node, ctx.ParentCtx)
		signature := cp.extractEnhancedFunctionSignature(node)
		
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("func-%s-%d", ctx.FilePath, node.Location.Line)),
			Name:         cp.extractCppFunctionName(node),
			Type:         symbolType,
			Location:     convertLocation(node.Location),
			Signature:    signature,
			Language:     ctx.Language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
			Visibility:   visibility,
		}
	case "namespace_definition":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("namespace-%s-%d", ctx.FilePath, node.Location.Line)),
			Name:         cp.extractCppNamespaceName(node),
			Type:         types.SymbolTypeNamespace,
			Location:     convertLocation(node.Location),
			Language:     ctx.Language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	case "field_declaration":
		// Check if this field_declaration is actually a function declaration (like operator overloads)
		if cp.isFunctionDeclaration(node) {
			symbolType, visibility := cp.classifyFunction(node, ctx.ParentCtx)
			signature := cp.extractEnhancedFunctionSignature(node)
			
			return &types.Symbol{
				Id:           types.SymbolId(fmt.Sprintf("func-%s-%d", ctx.FilePath, node.Location.Line)),
				Name:         cp.extractCppFunctionName(node),
				Type:         symbolType,
				Location:     convertLocation(node.Location),
				Signature:    signature,
				Language:     ctx.Language,
				Hash:         calculateHash(node.Value),
				LastModified: time.Now(),
				Visibility:   visibility,
			}
		} else {
			// Regular field declaration
			fieldName := cp.extractCppFieldName(node)
			if fieldName == "" {
				return nil // Don't create symbols for fields without proper names
			}
			return &types.Symbol{
				Id:           types.SymbolId(fmt.Sprintf("field-%s-%d", ctx.FilePath, node.Location.Line)),
				Name:         fieldName,
				Type:         types.SymbolTypeVariable,
				Location:     convertLocation(node.Location),
				Language:     ctx.Language,
				Hash:         calculateHash(node.Value),
				LastModified: time.Now(),
				Visibility:   cp.extractVisibility(node, ctx.ParentCtx),
			}
		}
	case "template_declaration":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("template-%s-%d", ctx.FilePath, node.Location.Line)),
			Name:         cp.extractTemplateName(node),
			Type:         types.SymbolTypeTemplate,
			Location:     convertLocation(node.Location),
			Language:     ctx.Language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
			Signature:    cp.extractTemplateSignature(node),
		}
	case "preproc_include":
		return &types.Symbol{
			Id:           types.SymbolId(fmt.Sprintf("include-%s-%d", ctx.FilePath, node.Location.Line)),
			Name:         cp.extractIncludeName(node),
			Type:         types.SymbolTypeImport,
			Location:     convertLocation(node.Location),
			Language:     ctx.Language,
			Hash:         calculateHash(node.Value),
			LastModified: time.Now(),
		}
	default:
		return nil
	}
}

// CppParentContext tracks the current parsing context for better symbol classification
type CppParentContext struct {
	InClass       bool
	ClassName     string
	CurrentAccess string // "private", "public", "protected"
	InNamespace   bool
	NamespaceName string
	TemplateDepth int
}

// SymbolExtractionContext groups related parameters for symbol extraction
type SymbolExtractionContext struct {
	FilePath     string
	Language     string
	Content      string
	ParentCtx    *CppParentContext
}

// classifyFunction determines if a function is a method, constructor, destructor, or regular function
func (cp *CppParser) classifyFunction(node *types.ASTNode, parentContext *CppParentContext) (types.SymbolType, string) {
	functionName := cp.extractCppFunctionName(node)
	
	if parentContext == nil {
		return types.SymbolTypeFunction, "public"
	}

	// If we're in a class context
	if parentContext.InClass {
		// Check for constructor (function name matches class name)
		if functionName == parentContext.ClassName {
			return types.SymbolTypeConstructor, parentContext.CurrentAccess
		}
		
		// Check for destructor (starts with ~)
		if strings.HasPrefix(functionName, "~") {
			return types.SymbolTypeDestructor, parentContext.CurrentAccess
		}
		
		// Check for operator overload
		if strings.Contains(functionName, "operator") {
			return types.SymbolTypeOperator, parentContext.CurrentAccess
		}
		
		// Regular method
		return types.SymbolTypeMethod, parentContext.CurrentAccess
	}

	// Top-level function
	return types.SymbolTypeFunction, "public"
}

// extractVisibility determines the visibility of a symbol based on context
func (cp *CppParser) extractVisibility(node *types.ASTNode, parentContext *CppParentContext) string {
	if parentContext == nil {
		return "public" // default for top-level symbols
	}
	
	if parentContext.InClass {
		return parentContext.CurrentAccess
	}
	
	return "public"
}

// extractEnhancedFunctionSignature extracts function signature with virtual/override/final info
func (cp *CppParser) extractEnhancedFunctionSignature(node *types.ASTNode) string {
	signature := cp.extractBasicSignature(node)
	
	// Check for virtual/override/final qualifiers
	qualifiers := []string{}
	if cp.isVirtualFunction(node) {
		qualifiers = append(qualifiers, "virtual")
	}
	if cp.isOverrideFunction(node) {
		qualifiers = append(qualifiers, "override")
	}
	if cp.isFinalFunction(node) {
		qualifiers = append(qualifiers, "final")
	}
	if cp.isPureVirtualFunction(node) {
		qualifiers = append(qualifiers, "pure virtual")
	}
	
	if len(qualifiers) > 0 {
		signature += " [" + strings.Join(qualifiers, ", ") + "]"
	}
	
	return signature
}

// AST-based Virtual/Override/Final detection methods
func (cp *CppParser) isVirtualFunction(node *types.ASTNode) bool {
	// Check for direct virtual keyword as sibling
	for _, child := range node.Children {
		if child.Type == "virtual" {
			return true
		}
	}
	return false
}

func (cp *CppParser) isOverrideFunction(node *types.ASTNode) bool {
	// Check for virtual_specifier containing override
	return cp.hasVirtualSpecifier(node, "override")
}

func (cp *CppParser) isFinalFunction(node *types.ASTNode) bool {
	// Check for virtual_specifier containing final
	return cp.hasVirtualSpecifier(node, "final")
}

func (cp *CppParser) isPureVirtualFunction(node *types.ASTNode) bool {
	// Check for = 0 pattern by looking for = followed by number_literal "0"
	foundEquals := false
	for _, child := range node.Children {
		if child.Type == "=" {
			foundEquals = true
		} else if foundEquals && child.Type == "number_literal" && strings.TrimSpace(child.Value) == "0" {
			return true
		}
	}
	return false
}

// hasVirtualSpecifier checks for virtual_specifier nodes containing the given specifier
func (cp *CppParser) hasVirtualSpecifier(node *types.ASTNode, specifier string) bool {
	for _, child := range node.Children {
		if child.Type == "function_declarator" {
			for _, grandchild := range child.Children {
				if grandchild.Type == "virtual_specifier" {
					for _, ggchild := range grandchild.Children {
						if ggchild.Type == specifier {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

// extractBasicSignature extracts the basic function signature
func (cp *CppParser) extractBasicSignature(node *types.ASTNode) string {
	// Look for function_declarator pattern
	for _, child := range node.Children {
		if child.Type == "function_declarator" {
			return strings.TrimSpace(child.Value)
		}
	}
	
	// Fallback to extracting from full node value
	lines := strings.Split(node.Value, "\n")
	if len(lines) > 0 {
		// Return first line which usually contains the signature
		return strings.TrimSpace(lines[0])
	}
	
	return ""
}

// C++ specific symbol name extraction helpers
func (cp *CppParser) extractCppClassName(node *types.ASTNode) string {
	// Look for type_identifier child in class_specifier
	for _, child := range node.Children {
		if child.Type == "type_identifier" {
			return strings.TrimSpace(child.Value)
		}
	}
	return cp.extractGenericSymbolName(node)
}

func (cp *CppParser) extractCppFunctionName(node *types.ASTNode) string {
	// Look for function_declarator -> field_identifier, identifier, destructor_name, template_function, or operator_name
	for _, child := range node.Children {
		if child.Type == "function_declarator" {
			for _, grandchild := range child.Children {
				if grandchild.Type == "field_identifier" || grandchild.Type == "identifier" {
					return strings.TrimSpace(grandchild.Value)
				}
				if grandchild.Type == "operator_name" {
					return strings.TrimSpace(grandchild.Value)
				}
				if grandchild.Type == "destructor_name" {
					return strings.TrimSpace(grandchild.Value)
				}
				if grandchild.Type == "template_function" {
					// Extract function name from template_function (e.g., processValue<std::string>)
					for _, ggchild := range grandchild.Children {
						if ggchild.Type == "identifier" {
							return strings.TrimSpace(ggchild.Value)
						}
					}
					// Fallback: extract base name from full template_function value
					templateFunc := strings.TrimSpace(grandchild.Value)
					if idx := strings.Index(templateFunc, "<"); idx > 0 {
						return templateFunc[:idx]
					}
					return templateFunc
				}
			}
		}
	}
	return cp.extractGenericSymbolName(node)
}

func (cp *CppParser) extractCppFieldName(node *types.ASTNode) string {
	// Look for field_identifier in field_declaration
	for _, child := range node.Children {
		if child.Type == "field_identifier" {
			return strings.TrimSpace(child.Value)
		}
	}
	
	// If we can't find a field_identifier, this might not be a valid field
	// Don't extract it as a symbol
	return ""
}

func (cp *CppParser) extractCppNamespaceName(node *types.ASTNode) string {
	// Look for namespace_identifier in namespace_definition
	for _, child := range node.Children {
		if child.Type == "namespace_identifier" {
			return strings.TrimSpace(child.Value)
		}
		if child.Type == "identifier" {
			return strings.TrimSpace(child.Value)
		}
	}
	return cp.extractGenericSymbolName(node)
}

func (cp *CppParser) extractTemplateName(node *types.ASTNode) string {
	// Extract template name from template_declaration
	for _, child := range node.Children {
		if child.Type == "class_specifier" {
			// Look for type_identifier (primary template) or template_type (specialization)
			for _, grandchild := range child.Children {
				if grandchild.Type == "type_identifier" {
					return strings.TrimSpace(grandchild.Value)
				}
				if grandchild.Type == "template_type" {
					// For specializations like MyTemplate<int>, extract the base name
					for _, ggchild := range grandchild.Children {
						if ggchild.Type == "type_identifier" {
							return strings.TrimSpace(ggchild.Value)
						}
					}
					// Fallback: extract from the full template_type value
					templateValue := strings.TrimSpace(grandchild.Value)
					if idx := strings.Index(templateValue, "<"); idx > 0 {
						return templateValue[:idx]
					}
					return templateValue
				}
			}
			return cp.extractGenericSymbolName(child)
		}
		if child.Type == "function_definition" || child.Type == "function_declaration" {
			return cp.extractCppFunctionName(child)
		}
	}
	return "template"
}

func (cp *CppParser) extractTemplateSignature(node *types.ASTNode) string {
	// Look for template_parameter_list
	signature := ""
	specializationInfo := ""
	
	for _, child := range node.Children {
		if child.Type == "template_parameter_list" {
			signature = strings.TrimSpace(child.Value)
		}
		
		// Check for class template specialization
		if child.Type == "class_specifier" {
			for _, grandchild := range child.Children {
				if grandchild.Type == "template_type" {
					// This is a class template specialization
					specializationInfo = " (specialization: " + strings.TrimSpace(grandchild.Value) + ")"
				}
			}
		}
		
		// Check for function template specialization
		if child.Type == "function_definition" || child.Type == "function_declaration" {
			for _, grandchild := range child.Children {
				if grandchild.Type == "function_declarator" {
					for _, ggchild := range grandchild.Children {
						if ggchild.Type == "template_function" {
							// This is a function template specialization
							specializationInfo = " (specialization: " + strings.TrimSpace(ggchild.Value) + ")"
						}
					}
				}
			}
		}
	}
	
	return signature + specializationInfo
}

func (cp *CppParser) extractIncludeName(node *types.ASTNode) string {
	// Extract include path from preproc_include
	for _, child := range node.Children {
		if child.Type == "string_literal" || child.Type == "system_lib_string" {
			return strings.TrimSpace(child.Value)
		}
	}
	return strings.TrimSpace(node.Value)
}

func (cp *CppParser) extractGenericSymbolName(node *types.ASTNode) string {
	// Generic symbol name extraction fallback
	for _, child := range node.Children {
		if child.Type == "identifier" || child.Type == "field_identifier" || child.Type == "type_identifier" {
			return strings.TrimSpace(child.Value)
		}
	}
	
	// Fallback to first identifier-like word in the value
	words := strings.Fields(node.Value)
	for _, word := range words {
		if len(word) > 0 && (word[0] >= 'A' && word[0] <= 'Z' || word[0] >= 'a' && word[0] <= 'z' || word[0] == '_') {
			// Remove punctuation
			word = strings.Trim(word, "(){}[];,<>")
			if len(word) > 0 {
				return word
			}
		}
	}
	
	return "unknown"
}

// ExtractSymbolsWithContext extracts symbols from AST with proper parent context tracking
func (cp *CppParser) ExtractSymbolsWithContext(root *types.ASTNode, filePath, content string) ([]*types.Symbol, error) {
	if cp == nil {
		return nil, fmt.Errorf("CppParser is nil")
	}
	if root == nil {
		return nil, fmt.Errorf("AST root is nil")
	}
	if filePath == "" {
		return nil, fmt.Errorf("filePath is empty")
	}
	
	var symbols []*types.Symbol
	
	// Start with empty context
	context := &CppParentContext{
		CurrentAccess: "private", // C++ class default is private
	}
	
	if err := cp.extractSymbolsRecursive(root, filePath, content, context, &symbols); err != nil {
		return nil, fmt.Errorf("failed to extract symbols: %w", err)
	}
	
	return symbols, nil
}

// extractSymbolsRecursive recursively extracts symbols while tracking parent context
func (cp *CppParser) extractSymbolsRecursive(node *types.ASTNode, filePath, content string, context *CppParentContext, symbols *[]*types.Symbol) error {
	if node == nil || symbols == nil {
		return NewValidationError("invalid parameters: node or symbols is nil")
	}
	if cp == nil {
		return NewValidationError("CppParser is nil")
	}
	if context == nil {
		// Create default context if none provided
		context = &CppParentContext{
			CurrentAccess: "private", // C++ class default
		}
	}

	// Create new context for this node level using efficient copying
	newContext := cp.copyContext(context)

	// Update context based on current node
	cp.updateContext(node, newContext)

	// For class field_declaration_list, handle access specifiers specially
	if node.Type == "field_declaration_list" {
		if err := cp.processClassBody(node, filePath, content, newContext, symbols); err != nil {
			return NewASTError("failed to process class body", err)
		}
		return nil
	}

	// Extract symbol if this node represents one (but not for access specifiers themselves)
	if node.Type != "access_specifier" {
		if symbol := cp.NodeToSymbol(node, filePath, "cpp", content, newContext); symbol != nil {
			*symbols = append(*symbols, symbol)
		}
	}

	// Recursively process children with updated context
	for _, child := range node.Children {
		if err := cp.extractSymbolsRecursive(child, filePath, content, newContext, symbols); err != nil {
			return NewASTError(fmt.Sprintf("failed to process child node %s", child.Type), err)
		}
	}
	
	return nil
}

// processClassBody handles the special case of class body with access specifiers
func (cp *CppParser) processClassBody(bodyNode *types.ASTNode, filePath, content string, context *CppParentContext, symbols *[]*types.Symbol) error {
	if bodyNode == nil || symbols == nil {
		return fmt.Errorf("invalid parameters: bodyNode or symbols is nil")
	}
	if context == nil {
		return fmt.Errorf("context is nil")
	}
	
	currentAccess := context.CurrentAccess // Start with the class default
	
	for _, child := range bodyNode.Children {
		if child.Type == "access_specifier" {
			// Update access level for subsequent declarations
			currentAccess = strings.TrimSpace(child.Value)
		} else if child.Type == "field_declaration" || child.Type == "function_definition" || child.Type == "function_declaration" || child.Type == "template_declaration" {
			// Create context with current access level using efficient copying
			childContext := cp.copyContext(context)
			childContext.CurrentAccess = currentAccess
			
			// Extract symbol with proper access level
			childCtx := &SymbolExtractionContext{
				FilePath:  filePath,
				Language:  "cpp",
				Content:   content,
				ParentCtx: childContext,
			}
			if symbol := cp.NodeToSymbolWithContext(child, childCtx); symbol != nil {
				*symbols = append(*symbols, symbol)
			}
			
			// Process any nested content (but skip the direct symbol extraction since we already did it)
			for _, grandchild := range child.Children {
				if err := cp.extractSymbolsRecursive(grandchild, filePath, content, childContext, symbols); err != nil {
					return fmt.Errorf("failed to process grandchild node %s: %w", grandchild.Type, err)
				}
			}
		} else {
			// For other nodes, just process recursively
			childContext := cp.copyContext(context)
			childContext.CurrentAccess = currentAccess
			if err := cp.extractSymbolsRecursive(child, filePath, content, childContext, symbols); err != nil {
				return fmt.Errorf("failed to process child node %s: %w", child.Type, err)
			}
		}
	}
	
	return nil
}

// updateContext updates the parent context based on the current node
func (cp *CppParser) updateContext(node *types.ASTNode, context *CppParentContext) {
	switch node.Type {
	case "class_specifier", "struct_specifier":
		context.InClass = true
		context.ClassName = cp.extractCppClassName(node)
		if node.Type == "struct_specifier" {
			context.CurrentAccess = "public" // structs default to public
		} else {
			context.CurrentAccess = "private" // classes default to private
		}
		
	case "namespace_definition":
		context.InNamespace = true
		context.NamespaceName = cp.extractCppNamespaceName(node)
		
	case "template_declaration":
		context.TemplateDepth++
		
	case "access_specifier":
		// Update current access level - node.Value should be just "private", "public", or "protected"
		accessValue := strings.TrimSpace(node.Value)
		context.CurrentAccess = accessValue
		
	// Also check for access labels that might have different node types
	default:
		// Check if this node contains access specifier labels
		nodeValue := strings.TrimSpace(node.Value)
		if strings.HasSuffix(nodeValue, ":") && context.InClass {
			switch nodeValue {
			case "private:", "public:", "protected:":
				access := strings.TrimSuffix(nodeValue, ":")
				context.CurrentAccess = access
			}
		}
	}
}

// detectCppFeatures performs comprehensive C++ feature detection using AST traversal
func (cp *CppParser) detectCppFeatures(rootNode *sitter.Node, content string) map[string]interface{} {
	features := make(map[string]interface{})
	
	// Initialize all features to false
	cp.initializeFeatureFlags(features)
	
	// Perform AST-based detection
	cp.detectFeaturesFromAST(rootNode, features, content)
	
	// Supplement with pattern-based detection for complex features
	cp.detectFeaturesFromPatterns(content, features)
	
	return features
}

// initializeFeatureFlags initializes all feature flags to false
func (cp *CppParser) initializeFeatureFlags(features map[string]interface{}) {
	// Core features (Phase 1)
	coreFeatures := []string{
		"has_classes", "has_structs", "has_functions", "has_namespaces",
		"has_constructors", "has_destructors", "has_inheritance", "has_includes",
	}
	
	// P1 features (Phase 2)
	p1Features := []string{
		"has_templates", "has_auto_keyword", "has_lambdas", "has_range_for",
		"has_smart_pointers", "has_constexpr", "has_operator_overload",
	}
	
	// P2 features (Phase 3)
	p2Features := []string{
		"has_concepts", "has_structured_binding", "has_if_constexpr",
		"has_coroutines", "has_modules",
	}
	
	// Framework features
	frameworkFeatures := []string{
		"has_qt", "has_boost", "has_opencv", "has_unreal", "has_stl",
	}
	
	// Initialize all to false
	allFeatures := append(coreFeatures, p1Features...)
	allFeatures = append(allFeatures, p2Features...)
	allFeatures = append(allFeatures, frameworkFeatures...)
	
	for _, feature := range allFeatures {
		features[feature] = false
	}
}

// detectFeaturesFromAST performs AST-based feature detection
func (cp *CppParser) detectFeaturesFromAST(node *sitter.Node, features map[string]interface{}, content string) {
	if node == nil {
		return
	}
	
	nodeType := node.Kind()
	
	// Core feature detection
	switch nodeType {
	case "class_specifier":
		features["has_classes"] = true
	case "struct_specifier":
		features["has_structs"] = true
	case "function_definition":
		features["has_functions"] = true
	case "namespace_definition":
		features["has_namespaces"] = true
	case "preproc_include":
		features["has_includes"] = true
	case "template_declaration":
		features["has_templates"] = true
	}
	
	// Check node content for specific patterns using safe extraction
	nodeContent := cp.safeExtractNodeContent(node, content)
	if nodeContent != "" {
		
		// P1 feature detection
		if strings.Contains(nodeContent, "auto ") {
			features["has_auto_keyword"] = true
		}
		if strings.Contains(nodeContent, "constexpr") {
			features["has_constexpr"] = true
		}
		if strings.Contains(nodeContent, "operator") {
			features["has_operator_overload"] = true
		}
	}
	
	// Recursively check children
	for i := 0; i < int(node.ChildCount()); i++ {
		cp.detectFeaturesFromAST(node.Child(uint(i)), features, content)
	}
}

// detectFeaturesFromPatterns supplements AST detection with pattern matching
func (cp *CppParser) detectFeaturesFromPatterns(content string, features map[string]interface{}) {
	// Constructor detection
	if cp.detectConstructors(content) {
		features["has_constructors"] = true
	}
	
	// Enhanced destructor detection
	if cp.detectDestructors(content) {
		features["has_destructors"] = true
	}
	
	// Detect special member functions
	specialFunctions := cp.detectSpecialMemberFunctions(content)
	for key, value := range specialFunctions {
		features[key] = value
	}
	
	// Inheritance detection
	if cp.detectInheritance(content) {
		features["has_inheritance"] = true
	}
	
	// Lambda detection
	if cp.detectLambdas(content) {
		features["has_lambdas"] = true
	}
	
	// Range-based for detection
	if cp.detectRangeBasedFor(content) {
		features["has_range_for"] = true
	}
	
	// Smart pointer detection
	if cp.detectSmartPointers(content) {
		features["has_smart_pointers"] = true
	}
	
	// P2 features
	if cp.detectConcepts(content) {
		features["has_concepts"] = true
	}
	
	if cp.detectStructuredBinding(content) {
		features["has_structured_binding"] = true
	}
	
	if cp.detectIfConstexpr(content) {
		features["has_if_constexpr"] = true
	}
	
	if cp.detectCoroutines(content) {
		features["has_coroutines"] = true
	}
	
	if cp.detectModules(content) {
		features["has_modules"] = true
	}
	
	// Framework detection
	cp.detectFrameworks(content, features)
}

// Enhanced pattern detection methods
func (cp *CppParser) detectConstructors(content string) bool {
	// Create validator for constructor patterns
	excludePatterns := []string{
		"return ", "if (", "while (", "for (", "switch (",
		"sizeof(", "typeof(", "decltype(", "#define", "#include",
	}
	
	includePatterns := []string{
		" : ",        // member initializer list
		"{}",         // brace initialization  
		"= default",  // defaulted constructor
		"= delete",   // deleted constructor
		"explicit ",  // explicit constructor
		"constexpr ", // constexpr constructor
		"noexcept",   // noexcept constructor
		"[[",         // attribute specifiers (C++11+)
	}
	
	validator := NewPatternValidator(excludePatterns, includePatterns)
	return validator.validateDeclarationLines(content)
}


// detectDestructors enhances destructor detection beyond simple tilde matching
func (cp *CppParser) detectDestructors(content string) bool {
	// Must contain tilde for destructors
	if !strings.Contains(content, "~") {
		return false
	}
	
	// Create validator for destructor patterns
	excludePatterns := []string{
		"return ", "if (", "while (", "for (", "switch (",
		"& ~", "| ~", "^ ~", "= ~", "( ~", // bitwise operations
	}
	
	includePatterns := []string{
		"virtual ~",   // virtual destructor
		"~",          // basic destructor (will be filtered by exclude patterns)
		"= default",  // defaulted destructor
		"= delete",   // deleted destructor  
		"noexcept",   // noexcept destructor
	}
	
	validator := NewPatternValidator(excludePatterns, includePatterns)
	return validator.validateDeclarationLines(content)
}


// detectSpecialMemberFunctions detects copy/move constructors and assignment operators
func (cp *CppParser) detectSpecialMemberFunctions(content string) map[string]bool {
	features := map[string]bool{
		"has_copy_constructor":      false,
		"has_move_constructor":      false,
		"has_copy_assignment":       false,
		"has_move_assignment":       false,
		"has_default_constructor":   false,
		"has_explicit_constructor":  false,
		"has_constexpr_constructor": false,
	}
	
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip comments
		if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Copy constructor patterns
		if (strings.Contains(line, "(const ") && strings.Contains(line, "&")) ||
		   strings.Contains(line, "= default") {
			features["has_copy_constructor"] = true
		}
		
		// Move constructor patterns  
		if strings.Contains(line, "&&") && strings.Contains(line, "(") {
			features["has_move_constructor"] = true
		}
		
		// Assignment operator patterns
		if strings.Contains(line, "operator=") {
			if strings.Contains(line, "&&") {
				features["has_move_assignment"] = true
			} else if strings.Contains(line, "&") {
				features["has_copy_assignment"] = true
			}
		}
		
		// Special constructor types
		if strings.Contains(line, "= default") && strings.Contains(line, "(") {
			features["has_default_constructor"] = true
		}
		
		if strings.Contains(line, "explicit ") {
			features["has_explicit_constructor"] = true
		}
		
		if strings.Contains(line, "constexpr ") && strings.Contains(line, "(") {
			features["has_constexpr_constructor"] = true
		}
	}
	
	return features
}

func (cp *CppParser) detectInheritance(content string) bool {
	// Look for inheritance patterns: class Derived : [access] Base
	return strings.Contains(content, " : ") && 
		   (strings.Contains(content, "class ") || strings.Contains(content, "struct "))
}

func (cp *CppParser) detectLambdas(content string) bool {
	// Enhanced lambda detection
	return strings.Contains(content, "[") && strings.Contains(content, "](") &&
		   (strings.Contains(content, "{") || strings.Contains(content, "->"))
}

func (cp *CppParser) detectRangeBasedFor(content string) bool {
	// Range-based for loop: for (type var : container)
	return strings.Contains(content, "for (") && strings.Contains(content, " : ") &&
		   !strings.Contains(content, "for (;;") // not C-style for loop
}

func (cp *CppParser) detectSmartPointers(content string) bool {
	smartPointers := []string{"unique_ptr", "shared_ptr", "weak_ptr", "make_unique", "make_shared"}
	for _, ptr := range smartPointers {
		if strings.Contains(content, ptr) {
			return true
		}
	}
	return false
}

func (cp *CppParser) detectConcepts(content string) bool {
	// C++20 concepts can be defined with or without explicit requires clause
	return strings.Contains(content, "concept ") && 
		   (strings.Contains(content, "requires") || 
		    strings.Contains(content, "std::integral") || 
		    strings.Contains(content, "std::floating_point") ||
		    strings.Contains(content, "= "))
}

func (cp *CppParser) detectStructuredBinding(content string) bool {
	// auto [a, b, c] = expression
	return strings.Contains(content, "auto [") && strings.Contains(content, "] =")
}

func (cp *CppParser) detectIfConstexpr(content string) bool {
	return strings.Contains(content, "if constexpr")
}

func (cp *CppParser) detectCoroutines(content string) bool {
	coroutineKeywords := []string{"co_await", "co_return", "co_yield"}
	for _, keyword := range coroutineKeywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	return false
}

func (cp *CppParser) detectModules(content string) bool {
	// C++20 modules: import std.core; or module mymodule;
	return (strings.Contains(content, "import ") && !strings.Contains(content, "#include")) ||
		   strings.Contains(content, "module ")
}

func (cp *CppParser) detectFrameworks(content string, features map[string]interface{}) {
	// Qt framework
	qtPatterns := []string{"#include <Q", "QObject", "Q_OBJECT", "SIGNAL", "SLOT"}
	for _, pattern := range qtPatterns {
		if strings.Contains(content, pattern) {
			features["has_qt"] = true
			break
		}
	}
	
	// Boost framework
	boostPatterns := []string{"#include <boost/", "boost::", "BOOST_"}
	for _, pattern := range boostPatterns {
		if strings.Contains(content, pattern) {
			features["has_boost"] = true
			break
		}
	}
	
	// OpenCV framework
	opencvPatterns := []string{"#include <opencv2/", "cv::", "cv::Mat"}
	for _, pattern := range opencvPatterns {
		if strings.Contains(content, pattern) {
			features["has_opencv"] = true
			break
		}
	}
	
	// Unreal Engine framework
	unrealPatterns := []string{"UCLASS", "UFUNCTION", "UPROPERTY", "#include \"CoreMinimal.h\""}
	for _, pattern := range unrealPatterns {
		if strings.Contains(content, pattern) {
			features["has_unreal"] = true
			break
		}
	}
	
	// STL framework
	stlPatterns := []string{"std::", "#include <vector>", "#include <string>", "#include <memory>"}
	for _, pattern := range stlPatterns {
		if strings.Contains(content, pattern) {
			features["has_stl"] = true
			break
		}
	}
}

// Maximum depth for AST conversion to prevent stack overflow
const MaxASTConversionDepth = 1000

// convertTreeSitterNode converts a Tree-sitter node to our AST node format
func (cp *CppParser) convertTreeSitterNode(tsNode *sitter.Node, content string) *types.ASTNode {
	return cp.convertTreeSitterNodeWithDepth(tsNode, content, 0)
}

// convertTreeSitterNodeWithDepth converts a Tree-sitter node with depth limiting
func (cp *CppParser) convertTreeSitterNodeWithDepth(tsNode *sitter.Node, content string, depth int) *types.ASTNode {
	if tsNode == nil {
		return nil
	}
	
	// Prevent stack overflow from deeply nested structures
	if depth > MaxASTConversionDepth {
		return &types.ASTNode{
			Id:   fmt.Sprintf("truncated-node-%d-%d", tsNode.StartByte(), tsNode.EndByte()),
			Type: tsNode.Kind() + "_truncated",
			Location: types.FileLocation{
				Line:      1,
				Column:    1,
				EndLine:   1,
				EndColumn: 1,
			},
			Value:    fmt.Sprintf("// Truncated at depth %d", depth),
			Children: make([]*types.ASTNode, 0),
		}
	}

	startPos := tsNode.StartPosition()
	endPos := tsNode.EndPosition()
	
	astNode := &types.ASTNode{
		Id:   fmt.Sprintf("node-%d-%d", tsNode.StartByte(), tsNode.EndByte()),
		Type: tsNode.Kind(),
		Location: types.FileLocation{
			Line:      int(startPos.Row) + 1,
			Column:    int(startPos.Column) + 1,
			EndLine:   int(endPos.Row) + 1,
			EndColumn: int(endPos.Column) + 1,
		},
		Children: make([]*types.ASTNode, 0),
	}

	// Extract text content for the node using safe bounds checking
	astNode.Value = cp.safeExtractNodeContent(tsNode, content)

	// Convert children with incremented depth
	for i := 0; i < int(tsNode.ChildCount()); i++ {
		child := cp.convertTreeSitterNodeWithDepth(tsNode.Child(uint(i)), content, depth+1)
		if child != nil {
			astNode.Children = append(astNode.Children, child)
		}
	}

	return astNode
}

// Helper functions for AST conversion and hash calculation
func (cp *CppParser) getFullContent() []byte {
	// This is a placeholder - in real implementation, 
	// the content would be passed through the parsing context
	return []byte{}
}

// isFunctionDeclaration checks if a field_declaration is actually a function declaration
func (cp *CppParser) isFunctionDeclaration(node *types.ASTNode) bool {
	// Look for function_declarator child, which indicates this is a function
	for _, child := range node.Children {
		if child.Type == "function_declarator" {
			return true
		}
	}
	return false
}

// copyContext creates an efficient copy of the parent context
func (cp *CppParser) copyContext(src *CppParentContext) *CppParentContext {
	if src == nil {
		return &CppParentContext{
			CurrentAccess: "private", // C++ class default
		}
	}
	
	// Explicit field copying for better performance and clarity
	return &CppParentContext{
		InClass:       src.InClass,
		ClassName:     src.ClassName,
		CurrentAccess: src.CurrentAccess,
		InNamespace:   src.InNamespace,
		NamespaceName: src.NamespaceName,
		TemplateDepth: src.TemplateDepth,
	}
}

// safeExtractNodeContent safely extracts content from Tree-sitter nodes with bounds checking
func (cp *CppParser) safeExtractNodeContent(node *sitter.Node, content string) string {
	if node == nil {
		return ""
	}
	
	start, end := int(node.StartByte()), int(node.EndByte())
	if start < 0 || end < 0 || start >= len(content) || end > len(content) || start > end {
		return ""  // Invalid bounds
	}
	return content[start:end]
}

// safeExtractASTNodeContent safely extracts content from AST nodes with bounds checking
func (cp *CppParser) safeExtractASTNodeContent(node *types.ASTNode, content string) string {
	if node == nil {
		return ""
	}
	
	// Convert location to byte positions (approximate)
	lines := strings.Split(content, "\n")
	if node.Location.Line <= 0 || node.Location.Line > len(lines) {
		return ""
	}
	
	// For safety, return the line content rather than attempting byte-level extraction
	return strings.TrimSpace(lines[node.Location.Line-1])
}

// PatternValidator provides common validation logic for C++ code patterns
type PatternValidator struct {
	excludePatterns []string
	includePatterns []string
}

// NewPatternValidator creates a validator with exclude and include patterns
func NewPatternValidator(excludePatterns, includePatterns []string) *PatternValidator {
	return &PatternValidator{
		excludePatterns: excludePatterns,
		includePatterns: includePatterns,
	}
}

// isValidDeclaration checks if a line represents a valid declaration based on patterns
func (pv *PatternValidator) isValidDeclaration(line string) bool {
	line = strings.TrimSpace(line)
	
	// Skip comments and preprocessor directives
	if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
		return false
	}
	
	// Check exclude patterns first
	for _, pattern := range pv.excludePatterns {
		if strings.Contains(line, pattern) {
			return false
		}
	}
	
	// Check include patterns
	for _, pattern := range pv.includePatterns {
		if strings.Contains(line, pattern) {
			return true
		}
	}
	
	return false
}

// validateDeclarationLines processes multiple lines with the validator
func (pv *PatternValidator) validateDeclarationLines(content string) bool {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if pv.isValidDeclaration(line) {
			return true
		}
	}
	return false
}

// Note: calculateHash and convertLocation are defined in manager.go
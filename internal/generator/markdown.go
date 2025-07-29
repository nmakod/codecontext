package generator

import (
	"fmt"
	"strings"
	"time"

	"github.com/nuthan-ms/codecontext/internal/config"
	"github.com/nuthan-ms/codecontext/pkg/types"
)

// MarkdownGenerator generates markdown output
type MarkdownGenerator struct {
	config *config.Config
}

// NewMarkdownGenerator creates a new markdown generator
func NewMarkdownGenerator(cfg *config.Config) *MarkdownGenerator {
	return &MarkdownGenerator{
		config: cfg,
	}
}

// Generate generates markdown output from the code graph
func (mg *MarkdownGenerator) Generate(graph *types.CodeGraph) (string, error) {
	var sb strings.Builder

	// Header
	sb.WriteString("# CodeContext Map\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n", time.Now().Format(time.RFC3339)))
	sb.WriteString("**Version:** 2.4.0\n")
	sb.WriteString("**Status:** Generated\n\n")

	// Files section
	if len(graph.Files) > 0 {
		sb.WriteString("## Files\n\n")
		for path, file := range graph.Files {
			sb.WriteString(fmt.Sprintf("### %s\n\n", path))
			sb.WriteString(fmt.Sprintf("- **Size:** %d bytes\n", file.Size))
			sb.WriteString(fmt.Sprintf("- **Lines:** %d\n", file.Lines))
			sb.WriteString(fmt.Sprintf("- **Language:** %s\n\n", file.Language))
		}
	}

	// Symbols section
	if len(graph.Symbols) > 0 {
		sb.WriteString("## Symbols\n\n")
		for _, symbol := range graph.Symbols {
			sb.WriteString(fmt.Sprintf("### %s\n\n", symbol.Name))
			sb.WriteString(fmt.Sprintf("- **Type:** %s\n", symbol.Kind))
			sb.WriteString(fmt.Sprintf("- **File:** %s\n", symbol.FullyQualifiedName))
			if symbol.Documentation != "" {
				sb.WriteString(fmt.Sprintf("- **Documentation:** %s\n", symbol.Documentation))
			}
			sb.WriteString("\n")
		}
	}

	// Dependencies section - derived from graph edges
	if len(graph.Edges) > 0 {
		sb.WriteString("## Dependencies\n\n")
		sb.WriteString("- Graph contains relationship data\n")
		sb.WriteString(fmt.Sprintf("- Total edges: %d\n", len(graph.Edges)))
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

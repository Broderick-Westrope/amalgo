package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/Broderick-Westrope/amalgo/internal/parser"
	"github.com/Broderick-Westrope/amalgo/internal/traverse"
	"github.com/Broderick-Westrope/amalgo/internal/utils"
)

// Generator handles the creation of the consolidated output file
type Generator struct {
	registry *parser.Registry
	output   string
}

// NewGenerator creates a new output generator
func NewGenerator(output string, registry *parser.Registry) *Generator {
	return &Generator{
		registry: registry,
		output:   output,
	}
}

// Generate creates the complete output file
func (g *Generator) Generate(paths []traverse.PathInfo, opts Options) error {
	// Write header with timestamp
	header := fmt.Sprintf("## Generated: %s\n\n", utils.FormatTimestamp())
	if err := utils.WriteOutput(g.output, header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Generate and write directory tree if requested
	if !opts.NoTree {
		tree := utils.GenerateTree(paths)
		if err := utils.AppendOutput(g.output, tree+"\n"); err != nil {
			return fmt.Errorf("failed to write directory tree: %w", err)
		}
	}

	// Generate language-specific outlines if requested
	if opts.Outline {
		if err := g.generateOutlines(paths); err != nil {
			return fmt.Errorf("failed to generate outlines: %w", err)
		}
	}

	// Dump file contents if requested
	if !opts.NoDump {
		if err := g.dumpFiles(paths); err != nil {
			return fmt.Errorf("failed to dump files: %w", err)
		}
	}

	return nil
}

// Options configures the output generation
type Options struct {
	NoTree  bool
	NoDump  bool
	Outline bool
}

func (g *Generator) generateOutlines(paths []traverse.PathInfo) error {
	content := "\n## Language Outlines\n\n"
	if err := utils.AppendOutput(g.output, content); err != nil {
		return err
	}

	for _, path := range paths {
		if path.IsDir {
			continue
		}

		// Skip if no parser available for this file type
		if !g.registry.IsSupported(path.Path) {
			continue
		}

		content := fmt.Sprintf("\n### File: %s\n", path.RelativePath)
		if err := utils.AppendOutput(g.output, content); err != nil {
			return err
		}

		if err := g.processFileOutline(path.Path); err != nil {
			return fmt.Errorf("failed to process outline for %s: %w", path.Path, err)
		}
	}

	return nil
}

func (g *Generator) processFileOutline(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	parser := g.registry.GetParser(filePath)
	if parser == nil {
		return fmt.Errorf("no parser found for %s", filePath)
	}

	outline, err := parser.Parse(content, filePath)
	if err != nil {
		return err
	}

	if len(outline.Errors) > 0 {
		var errMsgs []string
		for _, err := range outline.Errors {
			errMsgs = append(errMsgs, err.Error())
		}
		content := fmt.Sprintf("Parsing errors:\n%s\n", strings.Join(errMsgs, "\n"))
		return utils.AppendOutput(g.output, content)
	}

	return g.writeSymbols(outline.Symbols, 0)
}

func (g *Generator) writeSymbols(symbols []*parser.Symbol, depth int) error {
	indent := strings.Repeat("  ", depth)

	for _, symbol := range symbols {
		// Write symbol header
		content := fmt.Sprintf("%s%s: %s", indent, strings.ToUpper(symbol.Type), symbol.Name)
		if symbol.Signature != "" {
			content += fmt.Sprintf(" (%s)", symbol.Signature)
		}
		content += "\n"

		// Write decorators if present
		if len(symbol.Decorators) > 0 {
			content += fmt.Sprintf("%s  Decorators: %s\n", indent, strings.Join(symbol.Decorators, ", "))
		}

		// Write docstring if present
		if symbol.Docstring != "" {
			docLines := strings.Split(strings.TrimSpace(symbol.Docstring), "\n")
			content += fmt.Sprintf("%s  Documentation:\n", indent)
			for _, line := range docLines {
				content += fmt.Sprintf("%s    %s\n", indent, line)
			}
		}

		if err := utils.AppendOutput(g.output, content); err != nil {
			return err
		}

		// Recursively write children
		if len(symbol.Children) > 0 {
			if err := g.writeSymbols(symbol.Children, depth+1); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *Generator) dumpFiles(paths []traverse.PathInfo) error {
	content := "\n## File Contents\n\n"
	if err := utils.AppendOutput(g.output, content); err != nil {
		return err
	}

	for _, path := range paths {
		if path.IsDir {
			continue
		}

		// Check if file is binary
		isBinary, err := utils.IsBinaryFile(path.Path)
		if err != nil {
			return fmt.Errorf("failed to check if file is binary: %w", err)
		}

		if isBinary {
			content := fmt.Sprintf("--- File: %s\n<binary file>\n\n", path.RelativePath)
			if err := utils.AppendOutput(g.output, content); err != nil {
				return err
			}
			continue
		}

		// Read and write file content
		fileContent, err := os.ReadFile(path.Path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path.Path, err)
		}

		content := fmt.Sprintf("--- File: %s\n%s\n\n", path.RelativePath, string(fileContent))
		if err := utils.AppendOutput(g.output, content); err != nil {
			return err
		}
	}

	return nil
}

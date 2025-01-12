package internal

import (
	"fmt"
	"os"
	"strings"

	"github.com/Broderick-Westrope/amalgo/internal/parser"
)

type OutputFormat string

const (
	OutputFormatDefault = "default"
	OutputFormatJSON    = "json"
)

// Options configures the output generation
type OutputOptions struct {
	NoTree     bool
	NoDump     bool
	Outline    bool
	SkipBinary bool
	Format     OutputFormat
}

// GenerateOutput creates the complete output string
func GenerateOutput(paths []PathInfo, registry *parser.Registry, opts OutputOptions) (string, error) {
	if opts.Format == OutputFormatJSON {
		return generateOutputJSON(paths, registry, opts)
	}

	output := fmt.Sprintf("## Generated with Amalgo at: %s\n\n", FormatTimestamp())

	if !opts.NoTree {
		output += generateTree(paths)
	}

	if opts.Outline {
		outlines, err := generateOutlines(paths, registry)
		if err != nil {
			return "", fmt.Errorf("generating outlines: %w", err)
		}
		output += outlines
	}

	if !opts.NoDump {
		filesDump, err := dumpFiles(paths, opts.SkipBinary)
		if err != nil {
			return "", fmt.Errorf("dumping files: %w", err)
		}
		output += filesDump
	}
	return output, nil
}

func generateOutlines(paths []PathInfo, registry *parser.Registry) (string, error) {
	output := "## Language-Specific Outlines\n\n"

	var temp string
	var err error
	for _, path := range paths {
		if path.IsDir {
			continue
		}

		// Skip if no parser available for this file type
		if !registry.IsSupported(path.Path) {
			continue
		}

		temp, err = processFileOutline(path.Path, registry)
		if err != nil {
			return "", fmt.Errorf("processing outline for %q: %w", path.Path, err)
		}
		output += fmt.Sprintf("\n### File: %s\n%s", path.RelativePath, temp)
	}
	return output, nil
}

func processFileOutline(filePath string, registry *parser.Registry) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	parser := registry.GetParser(filePath)
	if parser == nil {
		return "", fmt.Errorf("no parser found for %q", filePath)
	}

	outline, err := parser.Parse(content, filePath)
	if err != nil {
		return "", fmt.Errorf("parsing file %q: %w", filePath, err)
	}

	if len(outline.Errors) > 0 {
		var errMsgs []string
		for _, err := range outline.Errors {
			errMsgs = append(errMsgs, err.Error())
		}
		return fmt.Sprintf("Parsing errors:\n%s\n", strings.Join(errMsgs, "\n")), nil
	}

	return writeSymbols(outline.Symbols, 0)
}

func writeSymbols(symbols []*parser.Symbol, depth int) (string, error) {
	indent := strings.Repeat("  ", depth)
	var output string
	for _, symbol := range symbols {
		// Write symbol header
		output += fmt.Sprintf("%s%s: %s", indent, strings.ToUpper(symbol.Type), symbol.Name)
		if symbol.Signature != "" {
			output += fmt.Sprintf(" (%s)", symbol.Signature)
		}
		output += "\n"

		// Write decorators if present
		if len(symbol.Decorators) > 0 {
			output += fmt.Sprintf("%s  Decorators: %s\n", indent, strings.Join(symbol.Decorators, ", "))
		}

		// Write docstring if present
		if symbol.Docstring != "" {
			docLines := strings.Split(strings.TrimSpace(symbol.Docstring), "\n")
			output += fmt.Sprintf("%s  Documentation:\n", indent)
			for _, line := range docLines {
				output += fmt.Sprintf("%s    %s\n", indent, line)
			}
		}

		// Recursively write children
		if len(symbol.Children) > 0 {
			temp, err := writeSymbols(symbol.Children, depth+1)
			if err != nil {
				return "", err
			}
			output += temp
		}
	}
	return output, nil
}

func dumpFiles(paths []PathInfo, skipBinary bool) (string, error) {
	var sb strings.Builder
	sb.WriteString("## File Contents\n\n")

	for _, path := range paths {
		if path.IsDir {
			continue
		}

		if skipBinary {
			// Check if file is binary
			isBinary, err := IsBinaryFile(path.Path)
			if err != nil {
				return "", fmt.Errorf("failed to check if file is binary: %w", err)
			}

			if isBinary {
				sb.WriteString(
					fmt.Sprintf("--- File: %s\n<binary file>\n\n", path.RelativePath),
				)
				continue
			}
		}

		// Read and write file content
		fileContent, err := os.ReadFile(path.Path)
		if err != nil {
			return "", fmt.Errorf("failed to read file %s: %w", path.Path, err)
		}

		sb.WriteString(
			fmt.Sprintf("--- File: %s\n%s\n\n", path.RelativePath, string(fileContent)),
		)
	}
	return sb.String(), nil
}

func generateTree(paths []PathInfo) string {
	tree := GenerateTree(paths)
	return fmt.Sprintf("## File Tree\n\n%s\n", tree)
}

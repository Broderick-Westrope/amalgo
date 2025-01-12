package internal

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Broderick-Westrope/amalgo/internal/parser"
)

type JSONDocument struct {
	Timestamp string            `json:"timestamp"`
	Tree      string            `json:"tree,omitempty"`
	Files     []JSONFile        `json:"files,omitempty"`
	Outlines  []JSONFileOutline `json:"outlines,omitempty"`
}

type JSONFile struct {
	Path    string `json:"path"`
	Content string `json:"content,omitempty"`
	Binary  bool   `json:"binary,omitempty"`
}

// JSONFileOutline represents the parsed structure of a source file
type JSONFileOutline struct {
	Path    string       `json:"path"`
	Symbols []JSONSymbol `json:"symbols,omitempty"`
	Errors  []string     `json:"errors,omitempty"`
}

// JSONSymbol represents a parsed symbol (function, type, class, etc.)
type JSONSymbol struct {
	Type          string       `json:"type"`                    // e.g., "function", "class", "interface"
	Name          string       `json:"name"`                    // Name of the symbol
	Signature     string       `json:"signature,omitempty"`     // Full signature for functions/methods
	Documentation string       `json:"documentation,omitempty"` // Associated documentation
	Decorators    []string     `json:"decorators,omitempty"`    // Any decorators/annotations
	Children      []JSONSymbol `json:"children,omitempty"`      // Nested symbols (e.g., methods in a class)
	Metadata      any          `json:"metadata,omitempty"`      // Additional language-specific metadata
}

func generateOutputJSON(paths []PathInfo, registry *parser.Registry, opts OutputOptions) (string, error) {
	doc := JSONDocument{
		Timestamp: FormatTimestamp(),
	}

	if !opts.NoTree {
		doc.Tree = GenerateTree(paths)
	}

	if !opts.NoDump {
		files, err := generateFilesJSON(paths, opts.SkipBinary)
		if err != nil {
			return "", fmt.Errorf("dumping files: %w", err)
		}
		doc.Files = files
	}

	if opts.Outline {
		outlines, err := generateOutlinesJSON(paths, registry)
		if err != nil {
			return "", fmt.Errorf("generating outlines: %w", err)
		}
		doc.Outlines = outlines
	}

	output, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling JSON: %w", err)
	}

	return string(output), nil
}

func generateFilesJSON(paths []PathInfo, skipBinary bool) ([]JSONFile, error) {
	files := make([]JSONFile, 0, len(paths))

	for _, path := range paths {
		if path.IsDir {
			continue
		}

		if skipBinary {
			isBinary, err := IsBinaryFile(path.Path)
			if err != nil {
				return nil, fmt.Errorf("checking if binary: %w", err)
			}

			if isBinary {
				files = append(files, JSONFile{
					Path:   path.RelativePath,
					Binary: true,
				})
				continue
			}
		}

		content, err := os.ReadFile(path.Path)
		if err != nil {
			return nil, fmt.Errorf("reading file %s: %w", path.Path, err)
		}

		files = append(files, JSONFile{
			Path:    path.RelativePath,
			Content: string(content),
		})
	}

	return files, nil
}

func generateOutlinesJSON(paths []PathInfo, registry *parser.Registry) ([]JSONFileOutline, error) {
	outlines := make([]JSONFileOutline, 0)

	for _, path := range paths {
		if path.IsDir || !registry.IsSupported(path.Path) {
			continue
		}

		content, err := os.ReadFile(path.Path)
		if err != nil {
			return nil, fmt.Errorf("reading file %s: %w", path.Path, err)
		}

		parser := registry.GetParser(path.Path)
		parsedOutline, err := parser.Parse(content, path.Path)
		if err != nil {
			return nil, fmt.Errorf("parsing file %s: %w", path.Path, err)
		}

		outline := JSONFileOutline{
			Path:    path.RelativePath,
			Symbols: make([]JSONSymbol, 0, len(parsedOutline.Symbols)),
		}

		// Convert parser.Symbols to our JSON Symbol type
		for _, sym := range parsedOutline.Symbols {
			symbol, err := convertSymbol(sym)
			if err != nil {
				return nil, fmt.Errorf("converting symbol in %s: %w", path.Path, err)
			}
			outline.Symbols = append(outline.Symbols, symbol)
		}

		// Add any parsing errors
		if len(parsedOutline.Errors) > 0 {
			outline.Errors = make([]string, len(parsedOutline.Errors))
			for i, err := range parsedOutline.Errors {
				outline.Errors[i] = err.Error()
			}
		}

		outlines = append(outlines, outline)
	}

	return outlines, nil
}

func convertSymbol(ps *parser.Symbol) (JSONSymbol, error) {
	children := make([]JSONSymbol, 0, len(ps.Children))
	for _, child := range ps.Children {
		converted, err := convertSymbol(child)
		if err != nil {
			return JSONSymbol{}, fmt.Errorf("converting child symbol: %w", err)
		}
		children = append(children, converted)
	}

	return JSONSymbol{
		Type:          ps.Type,
		Name:          ps.Name,
		Signature:     ps.Signature,
		Documentation: ps.Docstring,
		Decorators:    ps.Decorators,
		Children:      children,
		Metadata:      ps.Metadata,
	}, nil
}

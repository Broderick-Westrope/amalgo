// Package parser provides language-specific parsing capabilities
package parser

import "path/filepath"

// Parser defines the interface for language-specific parsers
type Parser interface {
	// Parse analyzes the content of a file and returns a structured outline
	Parse(content []byte, filename string) (*FileOutline, error)

	// Extensions returns the file extensions this parser handles
	Extensions() []string
}

// Symbol represents a parsed symbol (function, type, class, etc.)
type Symbol struct {
	Type       string         // e.g., "function", "class", "interface", etc.
	Name       string         // Name of the symbol
	Signature  string         // Full signature for functions/methods
	Docstring  string         // Associated documentation
	Decorators []string       // Any decorators/annotations
	Children   []*Symbol      // Nested symbols (e.g., methods in a class)
	Metadata   map[string]any // Additional language-specific metadata
}

// FileOutline represents the parsed structure of a source file
type FileOutline struct {
	Filename string    // Name of the parsed file
	Symbols  []*Symbol // Top-level symbols in the file
	Errors   []error   // Any errors encountered during parsing
}

// Registry manages the available parsers
type Registry struct {
	parsers map[string]Parser
}

// NewRegistry creates a new parser registry
func NewRegistry() *Registry {
	return &Registry{
		parsers: make(map[string]Parser),
	}
}

// Register adds a parser to the registry
func (r *Registry) Register(parser Parser) {
	for _, ext := range parser.Extensions() {
		r.parsers[ext] = parser
	}
}

// GetParser returns the appropriate parser for a file extension
func (r *Registry) GetParser(filename string) Parser {
	ext := filepath.Ext(filename)
	return r.parsers[ext]
}

// IsSupported checks if there's a parser available for the given file extension
func (r *Registry) IsSupported(filename string) bool {
	ext := filepath.Ext(filename)
	_, ok := r.parsers[ext]
	return ok
}

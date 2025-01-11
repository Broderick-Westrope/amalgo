package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// GoParser implements Parser for Go source files
type GoParser struct{}

// NewGoParser creates a new Go parser
func NewGoParser() *GoParser {
	return &GoParser{}
}

func (p *GoParser) Extensions() []string {
	return []string{".go"}
}

func (p *GoParser) Parse(content []byte, filename string) (*FileOutline, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
	if err != nil {
		return &FileOutline{
			Filename: filename,
			Errors:   []error{err},
		}, nil
	}

	outline := &FileOutline{
		Filename: filename,
		Symbols:  make([]*Symbol, 0),
	}

	// Process package-level declarations
	for _, decl := range file.Decls {
		symbols := p.processDecl(decl, file)
		outline.Symbols = append(outline.Symbols, symbols...)
	}

	return outline, nil
}

func (p *GoParser) processDecl(decl ast.Decl, file *ast.File) []*Symbol {
	var symbols []*Symbol

	switch d := decl.(type) {
	case *ast.FuncDecl:
		symbol := &Symbol{
			Type:      "function",
			Name:      d.Name.Name,
			Signature: p.getFunctionSignature(d),
			Docstring: p.getDocstring(d.Doc),
		}

		// Handle methods
		if d.Recv != nil {
			symbol.Type = "method"
			if len(d.Recv.List) > 0 {
				recvType := p.typeToString(d.Recv.List[0].Type)
				symbol.Name = recvType + "." + d.Name.Name
			}
		}

		symbols = append(symbols, symbol)

	case *ast.GenDecl:
		switch d.Tok {
		case token.TYPE:
			for _, spec := range d.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					symbol := &Symbol{
						Type:      p.getTypeSymbolType(typeSpec),
						Name:      typeSpec.Name.Name,
						Docstring: p.getDocstring(d.Doc),
					}

					// Handle interface methods and struct fields
					if symbol.Type == "interface" {
						if iface, ok := typeSpec.Type.(*ast.InterfaceType); ok {
							symbol.Children = p.processInterface(iface)
						}
					} else if symbol.Type == "struct" {
						if structType, ok := typeSpec.Type.(*ast.StructType); ok {
							symbol.Children = p.processStruct(structType)
						}
					}

					symbols = append(symbols, symbol)
				}
			}

		case token.CONST, token.VAR:
			for _, spec := range d.Specs {
				if valSpec, ok := spec.(*ast.ValueSpec); ok {
					for _, name := range valSpec.Names {
						symbol := &Symbol{
							Type:      strings.ToLower(d.Tok.String()),
							Name:      name.Name,
							Docstring: p.getDocstring(d.Doc),
						}
						if valSpec.Type != nil {
							symbol.Signature = p.typeToString(valSpec.Type)
						}
						symbols = append(symbols, symbol)
					}
				}
			}
		}
	}

	return symbols
}

func (p *GoParser) processInterface(iface *ast.InterfaceType) []*Symbol {
	var methods []*Symbol
	if iface.Methods == nil {
		return methods
	}

	for _, method := range iface.Methods.List {
		if len(method.Names) == 0 {
			continue // Skip embedded interfaces
		}

		methodType, ok := method.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		for _, name := range method.Names {
			symbol := &Symbol{
				Type:      "method",
				Name:      name.Name,
				Signature: p.getFuncTypeSignature(methodType),
				Docstring: p.getDocstring(method.Doc),
			}
			methods = append(methods, symbol)
		}
	}

	return methods
}

func (p *GoParser) processStruct(structType *ast.StructType) []*Symbol {
	var fields []*Symbol
	if structType.Fields == nil {
		return fields
	}

	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			// Anonymous/embedded field
			symbol := &Symbol{
				Type:      "field",
				Name:      p.typeToString(field.Type),
				Signature: p.typeToString(field.Type),
				Docstring: p.getDocstring(field.Doc),
			}
			fields = append(fields, symbol)
			continue
		}

		for _, name := range field.Names {
			symbol := &Symbol{
				Type:      "field",
				Name:      name.Name,
				Signature: p.typeToString(field.Type),
				Docstring: p.getDocstring(field.Doc),
			}
			fields = append(fields, symbol)
		}
	}

	return fields
}

func (p *GoParser) getFunctionSignature(fn *ast.FuncDecl) string {
	var builder strings.Builder
	builder.WriteString("func ")

	// Add receiver if it's a method
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		builder.WriteString("(")
		if len(fn.Recv.List[0].Names) > 0 {
			builder.WriteString(fn.Recv.List[0].Names[0].Name)
			builder.WriteString(" ")
		}
		builder.WriteString(p.typeToString(fn.Recv.List[0].Type))
		builder.WriteString(") ")
	}

	builder.WriteString(fn.Name.Name)
	builder.WriteString(p.getFuncTypeSignature(fn.Type))
	return builder.String()
}

func (p *GoParser) getFuncTypeSignature(ft *ast.FuncType) string {
	var builder strings.Builder
	builder.WriteString("(")

	if ft.Params != nil {
		for i, param := range ft.Params.List {
			if i > 0 {
				builder.WriteString(", ")
			}
			for j, name := range param.Names {
				if j > 0 {
					builder.WriteString(", ")
				}
				builder.WriteString(name.Name)
			}
			if len(param.Names) > 0 {
				builder.WriteString(" ")
			}
			builder.WriteString(p.typeToString(param.Type))
		}
	}

	builder.WriteString(")")

	if ft.Results != nil {
		if ft.Results.NumFields() == 1 && len(ft.Results.List[0].Names) == 0 {
			builder.WriteString(" ")
			builder.WriteString(p.typeToString(ft.Results.List[0].Type))
		} else {
			builder.WriteString(" (")
			for i, result := range ft.Results.List {
				if i > 0 {
					builder.WriteString(", ")
				}
				for j, name := range result.Names {
					if j > 0 {
						builder.WriteString(", ")
					}
					builder.WriteString(name.Name)
				}
				if len(result.Names) > 0 {
					builder.WriteString(" ")
				}
				builder.WriteString(p.typeToString(result.Type))
			}
			builder.WriteString(")")
		}
	}

	return builder.String()
}

func (p *GoParser) typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return p.typeToString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + p.typeToString(t.X)
	case *ast.ArrayType:
		return "[]" + p.typeToString(t.Elt)
	case *ast.MapType:
		return "map[" + p.typeToString(t.Key) + "]" + p.typeToString(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.ChanType:
		switch t.Dir {
		case ast.SEND:
			return "chan<- " + p.typeToString(t.Value)
		case ast.RECV:
			return "<-chan " + p.typeToString(t.Value)
		default:
			return "chan " + p.typeToString(t.Value)
		}
	case *ast.FuncType:
		return "func" + p.getFuncTypeSignature(t)
	case *ast.StructType:
		return "struct{...}"
	case *ast.Ellipsis:
		return "..." + p.typeToString(t.Elt)
	default:
		return "<unknown>"
	}
}

func (p *GoParser) getTypeSymbolType(typeSpec *ast.TypeSpec) string {
	switch typeSpec.Type.(type) {
	case *ast.InterfaceType:
		return "interface"
	case *ast.StructType:
		return "struct"
	default:
		return "type"
	}
}

func (p *GoParser) getDocstring(doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}
	return doc.Text()
}

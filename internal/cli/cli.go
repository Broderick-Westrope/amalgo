package cli

import (
	"fmt"

	"github.com/Broderick-Westrope/amalgo/internal/output"
	"github.com/Broderick-Westrope/amalgo/internal/parser"
	"github.com/Broderick-Westrope/amalgo/internal/traverse"
	"github.com/alecthomas/kong"
	"github.com/fatih/color"
)

func New(version string) *RootCmd {
	return &RootCmd{
		Version: versionFlag(version),
	}
}

type RootCmd struct {
	// Default command args and flags
	Dirs          []string `arg:"" optional:"" help:"Directories to analyze. Defaults to current directory." type:"path"`
	Output        string   `help:"Output file path. Defaults to './amalgo.txt'." short:"o" type:"path"`
	IgnoreDirs    []string `help:"Directories to ignore." short:"i" placeholder:"DIR"`
	Extensions    []string `help:"File extensions to include." short:"e" placeholder:"EXT"`
	IncludeAll    bool     `help:"Include all files regardless of extension." default:"false"`
	IncludeHidden bool     `help:"Include hidden files and directories." default:"false"`
	NoTree        bool     `help:"Skip directory tree generation." default:"false"`
	NoDump        bool     `help:"Skip file content dumping." default:"false"`
	Outline       bool     `help:"Generate language-specific outlines." default:"false"`
	NoColor       bool     `help:"Don't use color in the terminal output." default:"false"`

	// Subcommands
	Version versionFlag `help:"Print version information and quit" short:"v" name:"version"`
}

func (c *RootCmd) Run() error {
	if len(c.Dirs) == 0 {
		c.Dirs = []string{"."}
	}

	if c.Output == "" {
		c.Output = "amalgo.txt"
	}

	registry := parser.NewRegistry()
	registry.Register(parser.NewGoParser())
	// Add more parsers here as they're implemented
	// registry.Register(parser.NewPythonParser())
	// registry.Register(parser.NewJavaScriptParser())

	traverseOpts := traverse.Options{
		Extensions:    c.Extensions,
		IncludeAll:    c.IncludeAll,
		IncludeHidden: c.IncludeHidden,
		IgnoreDirs:    c.IgnoreDirs,
	}

	// Get all matching paths
	paths, err := traverse.GetPaths(c.Dirs, traverseOpts)
	if err != nil {
		return fmt.Errorf("traversing directories: %w", err)
	}

	outputOpts := output.Options{
		NoTree:  c.NoTree,
		NoDump:  c.NoDump,
		Outline: c.Outline,
	}

	generator := output.NewGenerator(c.Output, registry)
	if err := generator.Generate(paths, outputOpts); err != nil {
		return fmt.Errorf("generating output: %w", err)
	}

	// Print success message unless output is stdout
	if c.Output != "stdout" && c.Output != "-" {
		msg := fmt.Sprintf("Successfully generated output to: %s\n", c.Output)
		if !c.NoColor {
			msg = color.GreenString(msg)
		}
		fmt.Print(msg)
	}

	return nil
}

type versionFlag string

func (v versionFlag) Decode(_ *kong.DecodeContext) error { return nil }
func (v versionFlag) IsBool() bool                       { return true }
func (v versionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Println(vars["version"])
	app.Exit(0)
	return nil
}

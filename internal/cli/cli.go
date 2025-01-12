package cli

import (
	"fmt"
	"strings"

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
	Dirs          []string `arg:"" optional:"" help:"Directories to analyze." type:"path" default:"."`
	Output        string   `help:"Output file path." short:"o" type:"path" default:"amalgo.txt"`
	Stdout        bool     `help:"Write output to stdout instead of file."`
	IgnoreDirs    []string `help:"Directories to ignore." short:"i" placeholder:"DIR"`
	Filter        []string `help:"File patterns to include and exclude (e.g. '*.go,*.{js,ts}', or '*,!.md')." short:"f" placeholder:"PATTERN" default:"*"`
	IncludeHidden bool     `help:"Include hidden files and directories." default:"false"`
	NoTree        bool     `help:"Skip directory tree generation." default:"false"`
	NoDump        bool     `help:"Skip file content dumping." default:"false"`
	Outline       bool     `help:"Generate language-specific outlines." default:"false"`
	NoColor       bool     `help:"Don't use color in the terminal output." default:"false"`
	IncludeBinary bool     `help:"Include binary files." default:"false"`

	// Subcommands
	Version versionFlag `help:"Print version information and quit" short:"v" name:"version"`
}

func (c *RootCmd) validate() bool {
	issues := make([]string, 0)
	if len(c.Dirs) == 0 {
		issues = append(issues, "At least one input directory is required.")
	}
	if c.Output == "" {
		issues = append(issues, "Output cannot be empty.")
	}
	if c.NoDump && c.NoTree && !c.Outline {
		issues = append(issues, "An empty output is not allowed (no dump, no tree, and no outline).")
	}

	if len(issues) == 0 {
		return true
	}
	out := strings.Join(issues, "\n")
	if !c.NoColor {
		out = color.RedString(out)
	}
	fmt.Println(out)
	return false
}

func (c *RootCmd) Run() error {
	if !c.validate() {
		return nil
	}

	outputDest := c.Output
	if c.Stdout {
		outputDest = "stdout"
	}

	registry := parser.NewRegistry()
	registry.Register(parser.NewGoParser())

	includePatterns := make([]string, 0)
	excludePatterns := make([]string, 0)
	for _, original := range c.Filter {
		new, found := strings.CutPrefix(original, "!")
		if found {
			excludePatterns = append(excludePatterns, new)
		} else {
			includePatterns = append(includePatterns, original)
		}
	}

	traverseOpts := traverse.Options{
		IncludePatterns: includePatterns,
		ExcludePatterns: excludePatterns,
		IncludeHidden:   c.IncludeHidden,
		IgnoreDirs:      c.IgnoreDirs,
	}

	paths, err := traverse.GetPaths(c.Dirs, traverseOpts)
	if err != nil {
		return fmt.Errorf("traversing directories: %w", err)
	}

	outputOpts := output.Options{
		NoTree:     c.NoTree,
		NoDump:     c.NoDump,
		Outline:    c.Outline,
		SkipBinary: !c.IncludeBinary,
	}

	generator := output.NewGenerator(outputDest, registry)
	if err := generator.Generate(paths, outputOpts); err != nil {
		return fmt.Errorf("generating output: %w", err)
	}

	// Print success message unless output is stdout.
	if outputDest != "stdout" {
		msg := fmt.Sprintf("Successfully generated output to: %s\n", outputDest)
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

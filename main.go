package main

import (
	"fmt"
	"strings"

	"github.com/Broderick-Westrope/amalgo/internal"
	"github.com/Broderick-Westrope/amalgo/internal/parser"
	"github.com/alecthomas/kong"
	"github.com/fatih/color"
)

const (
	appName = "amalgo"
	version = "0.2.0"
)

func main() {
	cli := RootCmd{
		Version: versionFlag(version),
	}

	ctx := kong.Parse(&cli,
		kong.Name(appName),
		kong.Description("Create consolidated snapshots of source code for analysis, documentation, and sharing with LLMs."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.DefaultEnvars(appName),
		kong.Vars{
			"version": string(cli.Version),
		})
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}

type RootCmd struct {
	// Default command args and flags
	Dirs          []string `arg:"" optional:"" help:"Directories to analyze. If a file is provided the parent directory will be used." type:"path" default:"."`
	Output        string   `help:"Output file path." short:"o" type:"path" default:"amalgo.txt"`
	Stdout        bool     `help:"Write output to stdout instead of file."`
	Filter        []string `help:"Glob patterns to filter by. Prefixing a pattern with '!' makes it an exclude pattern. The default patterns include everything except hidden files and folders. (e.g. '*.go,*.{js,ts}' OR '!.md')" short:"f" default:"*,!.*"`
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
	if len(includePatterns) == 0 {
		includePatterns = []string{"*"}
	}

	paths, err := internal.TraverseDirectories(c.Dirs, includePatterns, excludePatterns)
	if err != nil {
		return fmt.Errorf("traversing directories: %w", err)
	}

	outputOpts := internal.OutputOptions{
		NoTree:     c.NoTree,
		NoDump:     c.NoDump,
		Outline:    c.Outline,
		SkipBinary: !c.IncludeBinary,
	}

	output, err := internal.GenerateOutput(paths, registry, outputOpts)
	if err != nil {
		return fmt.Errorf("generating output: %w", err)
	}

	err = internal.WriteOutput(outputDest, output)
	if err != nil {
		return fmt.Errorf("writing output: %w", err)
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

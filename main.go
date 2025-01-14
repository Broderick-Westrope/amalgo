package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Broderick-Westrope/amalgo/internal"
	"github.com/Broderick-Westrope/amalgo/internal/parser"
	"github.com/alecthomas/kong"
	"github.com/fatih/color"
)

const (
	appName = "amalgo"
	version = "0.4.1"
)

func main() {
	os.Exit(run())
}

func run() int {
	cli := RootCmd{
		Version: versionFlag(version),
	}

	exitHandler := &exitWriter{code: -1}
	ctx := kong.Parse(&cli,
		kong.Name(appName),
		kong.Description("Create consolidated snapshots of source code for analysis, documentation, and sharing with LLMs."),
		kong.UsageOnError(),
		kong.Writers(os.Stdout, exitHandler),
		kong.Exit(exitHandler.Exit),
		kong.DefaultEnvars(appName),
		kong.Vars{"version": string(cli.Version)},
	)

	switch {
	case exitHandler.code == 0:
		return exitHandler.code
	case exitHandler.code != -1:
		fmt.Fprintf(os.Stderr, "%s", exitHandler.message)
		return exitHandler.code
	}

	err := ctx.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	return 0
}

type RootCmd struct {
	// Default command args and flags
	Dir           string                `arg:"" optional:"" help:"Directory to analyze. If a file is provided it's parent directory will be used." type:"path" default:"."`
	Output        string                `help:"Specifies the destination path for the output file. The file extension will automatically adjust based on the selected format (see '--format')." short:"o" type:"path" placeholder:"amalgo.txt"`
	Stdout        bool                  `help:"Redirects all output to standard output (terminal) instead of writing to a file. Useful for piping output to other commands."`
	Filter        []string              `help:"Controls which files are processed using glob patterns. Include patterns are processed first, then exclude patterns (prefixed with '!'). Hidden files and directories are excluded by default." short:"f" default:"*,!.*"`
	NoTree        bool                  `help:"Skips the inclusion of the file tree in the output." default:"false"`
	NoDump        bool                  `help:"Skips the inclusion of file contents in the output." default:"false"`
	Outline       bool                  `help:"Includes in the output a language-aware outline of code files, showing functions, classes, and other significant elements. Only available for specific file extensions: '.go'." default:"false"`
	NoColor       bool                  `help:"Disables ANSI color codes in the output." default:"false"`
	IncludeBinary bool                  `help:"Processes binary files instead of skipping them. Use with caution as this may produce large or unreadable output." default:"false"`
	Format        internal.OutputFormat `help:"Selects an alternative output format. This affects both the structure and the file extension of the output. Options: 'default', 'json'." enum:"default,json" default:"default"`

	// Subcommands
	Version versionFlag `help:"Displays the current version of the tool and exits immediately." short:"v" name:"version"`
}

func (c *RootCmd) validate() bool {
	if c.Output == "" {
		if c.Format == internal.OutputFormatJSON {
			c.Output += "amalgo.json"
		} else {
			c.Output = "amalgo.txt"
		}
	}

	issues := make([]string, 0)
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

	paths, err := internal.TraverseDirectory(c.Dir, c.Filter)
	if err != nil {
		return fmt.Errorf("traversing directories: %w", err)
	}

	outputOpts := internal.OutputOptions{
		NoTree:     c.NoTree,
		NoDump:     c.NoDump,
		Outline:    c.Outline,
		SkipBinary: !c.IncludeBinary,
		Format:     c.Format,
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

// Custom writer that can capture output for testing
type exitWriter struct {
	code    int
	message string
}

func (w *exitWriter) Write(p []byte) (n int, err error) {
	w.message += string(p)
	return len(p), nil
}

func (w *exitWriter) Exit(code int) {
	w.code = code
}

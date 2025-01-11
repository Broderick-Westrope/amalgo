package main

import (
	"github.com/Broderick-Westrope/amalgo/internal/cli"
	"github.com/alecthomas/kong"
)

const (
	appName = "amalgo"
	version = "0.1.0"
)

func main() {
	cli := cli.New(version)
	ctx := kong.Parse(cli,
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

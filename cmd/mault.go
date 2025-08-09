package cmd

import "github.com/urfave/cli/v2"

// The main command
var Mault *cli.App = &cli.App{
	Name:                 "mault",
	Usage:                "A CLI tool to help you store your secrets",
	Args:                 false,
	EnableBashCompletion: true,
	Commands: []*cli.Command{
		initC,
		createC,
		generateC,
		listC,
		getC,
		deleteC,
	},
}

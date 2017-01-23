package main

import (
	"os"
	"path/filepath"

	"gopkg.in/urfave/cli.v2"
)

var (
	version  = "0.0.1"
	commands = []*cli.Command{}

	usage = `Provides a cli tool which executes specific orders of commands.
`
)

func main() {
	app := &cli.App{}
	app.Name = "Taskr"
	app.Version = version
	app.Commands = commands
	app.Usage = usage
	app.Action = taskRunner
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "in",
			Aliases: []string{"input"},
			Usage:   "in=task_file",
		},
	}

	app.Run(os.Args)
}

func taskRunner(ctx *cli.Context) error {
	cdir, err := os.Getwd()
	if err != nil {
		return err
	}

	infile := ctx.String("input")

	if infile == "" {
		infile = filepath.Join(cdir, "tasks.json")
	}

	_ = infile
	return nil
}

package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/influx6/clis/taskr/tasks"

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

	data, err := ioutil.ReadFile(infile)
	if err != nil {
		return err
	}

	var taskCol []*tasks.Tson

	if err := json.Unmarshal(data, &taskCol); err != nil {
		return err
	}

	tseries := tasks.New(taskCol...)

	if err := tseries.Start(); err != nil {
		return err
	}

	tseries.Wait()

	return nil
}

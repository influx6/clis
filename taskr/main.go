package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/influx6/clis/taskr/tasks"

	"gopkg.in/urfave/cli.v2"
)

var (
	version  = "0.0.1"
	commands = []*cli.Command{}

	usage = `Provides a cli tool which executes specific orders of commands.

	- Create a 'tasks.json' file easily

		> taskr init

	- Run all defined tasks

		> taskr run


	- Run tasks in a specificed task file

		> taskr run --in ./bonds/task.json
`

	template = `
[{
  "desc": "Example description",
  "write_delay": "20ms",
  "tasks": [{
    "max_runtime": "1m",
    "max_checktime": "500ms",
    "main": {
      "name": "Sample",
      "command":"echo",
      "params": ["Sample"],
      "desc": "Sample main task"
    },
    "after":[],
    "before":[]
  }]
}]
`
)

func main() {
	app := &cli.App{}
	app.Name = "Taskr"
	app.Version = version
	app.Commands = commands
	app.Usage = usage
	app.Commands = []*cli.Command{
		{
			Name:        "init",
			Usage:       "taskr init",
			Description: "Generates a initial tasks.json file for customizer",
			Action:      initJSON,
		},
		{
			Name:        "run",
			Usage:       "taskr run",
			Description: "Attempts to Load a tasks.json from the current path or the provided path to execute tasks defined in it",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "in",
					Aliases:     []string{"input"},
					Usage:       "in=tasks.json",
					DefaultText: "tasks.json",
				},
			},
			Action: taskRunner,
		},
	}

	app.Run(os.Args)
}

func initJSON(ctx *cli.Context) error {
	cdir, err := os.Getwd()
	if err != nil {
		return err
	}

	newFile := filepath.Join(cdir, "tasks.json")

	if err := ioutil.WriteFile(newFile, []byte(template), 0777); err != nil {
		return err
	}

	return nil
}

func taskRunner(ctx *cli.Context) error {
	cdir, err := os.Getwd()
	if err != nil {
		return err
	}

	userFile := ctx.String("input")

	if userFile != "" {
		if strings.HasPrefix(userFile, ".") || !strings.HasPrefix(userFile, "/") {
			userFile = filepath.Join(cdir, userFile)
		}
	} else {
		userFile = filepath.Join(cdir, "tasks.json")
	}

	data, err := ioutil.ReadFile(userFile)
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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		<-sigChan
		tseries.Stop()
	}()

	tseries.Wait()

	return nil
}

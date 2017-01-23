package tasks_test

import (
	"os"
	"testing"

	"github.com/influx6/clis/taskr/tasks"
)

func TestMasterTask(t *testing.T) {

	mtask := tasks.MasterTask{
		LockIO:    true,
		RunTimePT: 120,
		Main: &tasks.Task{
			Name:        "TodaysFileReport",
			Description: "Finds all modified files today",
			Command:     "find",
			Parameters:  []string{"~ -type f -mtime 0"},
		},
		Before: []tasks.Task{
			{
				Name:        "ListDir",
				Description: "List Current Dir",
				Command:     "ls",
			},
			{
				Name:        "EchoName",
				Description: "Echo Starting",
				Command:     "echo",
				Parameters:  []string{"'Starting Todays report'"},
			},
		},
		After: []tasks.Task{
			{
				Name:        "EchoName",
				Description: "Echo Ending",
				Command:     "echo",
				Parameters:  []string{"'Ending Todays report'"},
			},
		},
	}

	mtask.Run(os.Stdout, os.Stderr)

}

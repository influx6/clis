package tasks_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/influx6/clis/taskr/tasks"
)

func TestTSON(t *testing.T) {

	// var buf bytes.Buffer
	var tson tasks.Tson

	// tson.Sink = &buf
	tson.DirsGlob = "./*"
	tson.Description = "Manages running of all master task"
	tson.WriteDelay = "100ms"
	tson.Tasks = []tasks.MasterTask{
		{
			MaxRunTime: "20s",
			Main: &tasks.Task{
				Name:        "TodaysFileReport",
				Description: "Finds all modified files today",
				Command:     "find",
				Parameters:  []string{"~ -type f -mtime 0"},
			},
			Before: []*tasks.Task{
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
			After: []*tasks.Task{
				{
					Name:        "EchoName",
					Description: "Echo Ending",
					Command:     "echo",
					Parameters:  []string{"'Ending Todays report'"},
				},
			},
		},
	}

	if err := tson.Start(); err != nil {
		t.Fatalf("\tFailed: \t Error occurred in start tson: %q", err.Error())
	}

	go func() {
		<-time.After(5 * time.Second)
		tson.Restart()
	}()

	go func() {
		<-time.After(10 * time.Second)
		tson.Stop()
	}()

	tson.Wait()

	// if buf.Len() == 0 {
	// 	t.Fatal("Should  have contain data after tasks execution.")
	// }
}

func TestMasterTask(t *testing.T) {

	mtask := tasks.MasterTask{
		MaxRunTime: "20s",
		Main: &tasks.Task{
			Name:        "TodaysFileReport",
			Description: "Finds all modified files today",
			Command:     "find",
			Parameters:  []string{"~ -type f -mtime 0"},
		},
		Before: []*tasks.Task{
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
		After: []*tasks.Task{
			{
				Name:        "EchoName",
				Description: "Echo Ending",
				Command:     "echo",
				Parameters:  []string{"'Ending Todays report'"},
			},
		},
	}

	var buf bytes.Buffer
	mtask.Run(&buf, &buf)

	if buf.Len() == 0 {
		t.Fatal("Should  have contain data after tasks execution.")
	}

}

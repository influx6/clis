package tasks

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// Task defines a struct which holds commands which must be executed when runned.
type Task struct {
	Name        string   `json:"name"`
	Command     string   `json:"command"`
	Parameters  []string `json:"params"`
	Description string   `json:"description"`
	commando    *exec.Cmd
	stdOut      bytes.Buffer
	stdErr      bytes.Buffer
	signal      chan struct{}
}

// Stopped returns true/false if the given task has been stopped or not started.
func (t *Task) Stopped() bool {
	return t.commando == nil
}

// Run initializes the task to be invoked.
func (t *Task) Run(m io.Writer) {
	t.signal = make(chan struct{})

	t.commando = exec.Command(t.Command, t.Parameters...)
	t.commando.Stdout = &t.stdOut
	t.commando.Stderr = &t.stdErr

	if err := t.commando.Start(); err != nil {
		fmt.Fprintf(m, taskError, t.Name, t.Description, t.Command, t.Parameters, err.Error())
		return
	}

	if t.commando.ProcessState.Exited() {
		var status string

		if t.commando.ProcessState.Success() {
			status = "Passed!"
		} else {
			status = "Warning: Possible Error!"
		}

		fmt.Fprintf(m, task, t.Name, t.Description, t.Command, t.Parameters, status)

		if t.commando.ProcessState.Success() {
			fmt.Fprintf(m, taskOutput, t.stdOut.String())
		} else {
			fmt.Fprintf(m, taskErrOutput, t.stdOut.String())
		}
	}
}

// InputLoop creates loops to read out and error details to be printed into
// the writers for the task.
func (t *Task) InputLoop(outM, errM io.Writer) {
	go func() {
	inLoop:
		for {
			select {
			case <-t.signal:
				break inLoop
			case <-time.After(10 * time.Millisecond):
				fmt.Fprintf(errM, task, t.Name, t.Description, t.Command, t.Parameters, "Running!")
				fmt.Fprintf(errM, taskErrOutput, t.stdErr.Bytes())
				t.stdErr.Reset()
			}
		}
	}()

	go func() {
	inLoop:
		for {
			select {
			case <-t.signal:
				break inLoop
			case <-time.After(10 * time.Millisecond):
				fmt.Fprintf(outM, task, t.Name, t.Description, t.Command, t.Parameters, "Running!")
				fmt.Fprintf(outM, taskOutput, t.stdOut.Bytes())
				t.stdOut.Reset()
			}
		}
	}()
}

// Stop ends the task which when initialized.
func (t *Task) Stop(m io.Writer) {
	if t.commando.ProcessState.Exited() {
		return
	}

	var msg string
	var err error

	if runtime.GOOS == "windows" {
		err = t.commando.Process.Kill()
	} else {
		err = t.commando.Process.Signal(os.Interrupt)
	}

	t.commando.Wait()

	close(t.signal)

	if err != nil {
		msg = err.Error()
	} else {
		msg = "Killed Successfully!"
	}

	fmt.Fprintf(m, taskKill, t.Name, t.Description, t.Command, t.Parameters, msg)

	t.commando = nil
}

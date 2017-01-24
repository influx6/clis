package tasks

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
)

// Task defines a struct which holds commands which must be executed when runned.
type Task struct {
	Name        string   `json:"name"`
	Command     string   `json:"command"`
	Parameters  []string `json:"params"`
	Description string   `json:"description"`
	commando    *exec.Cmd
	signal      chan struct{}
}

// Wait blocks until the tasks completes or it gets stopped.
func (t *Task) Wait() {
	if t.commando == nil {
		return
	}

	t.commando.Wait()
}

// Stopped returns true/false if the given task has been stopped or not started.
func (t *Task) Stopped() bool {
	return t.commando == nil
}

// Run initializes the task to be invoked.
func (t *Task) Run(outw io.Writer, errw io.Writer) {
	t.signal = make(chan struct{})
	close(t.signal)

	t.commando = exec.Command(t.Command, t.Parameters...)

	fmt.Fprintf(outw, task, t.Name, t.Description, t.Command, t.Parameters, "Executing")
	t.inputLoop(outw, errw)

	if err := t.commando.Start(); err != nil {
		fmt.Fprintf(outw, taskError, t.Name, t.Description, t.Command, t.Parameters, err.Error())
		return
	}

	t.commando.Wait()

	if t.commando.ProcessState != nil {
		fmt.Fprintf(outw, taskLogs, t.commando.ProcessState.String())

		if t.commando.ProcessState.Exited() {
			if t.commando.ProcessState.Success() {
				fmt.Fprintf(outw, taskOutput, "Done!")
			} else {
				fmt.Fprintf(outw, taskErrOutput, "Failed!")
			}
		}
	}
}

// inputLoop creates loops to read out and error details to be printed into
// the writers for the task.
func (t *Task) inputLoop(outM, errM io.Writer) {
	fmt.Fprintf(outM, taskBegin, t.Name, t.Description)

	outReader, err := t.commando.StdoutPipe()
	if err != nil {
		fmt.Fprintf(outM, taskError, t.Name, t.Description, t.Command, t.Parameters, err.Error())
	} else {
		go t.readInput(outReader, outM)
	}

	errReader, err := t.commando.StderrPipe()
	if err != nil {
		fmt.Fprintf(errM, taskError, t.Name, t.Description, t.Command, t.Parameters, err.Error())
	} else {
		go t.readInput(errReader, errM)
	}

}

func (t *Task) readInput(reader io.ReadCloser, out io.Writer) {
	scanner := bufio.NewScanner(reader)

	for {
		select {
		case <-t.signal:
			if scanner.Scan() {
				fmt.Fprintf(out, taskLogs, scanner.Text())
			}
		default:
			return
		}
	}
}

// Stop ends the task which when initialized.
func (t *Task) Stop(m io.Writer) {
	if t.signal == nil {
		return
	}

	var msg string
	var err error

	if runtime.GOOS == "windows" {
		err = t.commando.Process.Kill()
	} else {
		err = t.commando.Process.Signal(os.Interrupt)
	}

	if errz := t.commando.Wait(); errz != nil {
		fmt.Fprintf(m, taskError, t.Name, t.Description, t.Command, t.Parameters, errz.Error())
	}

	if err != nil {
		msg = err.Error()
	} else {
		msg = "Killed Successfully!"
	}

	fmt.Fprintf(m, taskKill, t.Name, t.Description, t.Command, t.Parameters, msg)

	t.commando = nil
}

package tasks

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"syscall"
	"time"
)

// Task defines a struct which holds commands which must be executed when runned.
type Task struct {
	Name        string   `json:"name"`
	Command     string   `json:"command"`
	Parameters  []string `json:"params"`
	Description string   `json:"desc"`
	EndCheck    time.Duration
	commando    *exec.Cmd
	running     bool
	rl          sync.Mutex
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
	if t.commando != nil && t.commando.ProcessState != nil {
		return t.commando.ProcessState.Exited()
	}

	return t.commando == nil
}

// Run initializes the task to be invoked.
func (t *Task) Run(outw io.Writer, errw io.Writer) {
	t.rl.Lock()
	t.running = true
	t.rl.Unlock()

	t.commando = exec.Command(t.Command, t.Parameters...)

	fmt.Fprintf(outw, task, t.Name, t.Description, t.Command, t.Parameters, "Executing")
	t.inputLoop(outw, errw)

	if err := t.commando.Start(); err != nil {
		fmt.Fprintf(outw, taskError, t.Name, t.Description, t.Command, t.Parameters, err.Error())
		return
	}

	// Lunch a timer to check if the main had finished executing.
	go func() {
		for {
			<-time.After(t.EndCheck)

			if t.Stopped() {
				t.Stop(outw)
				return
			}
		}
	}()

	t.commando.Wait()

	if t.commando.ProcessState != nil {
		fmt.Fprintf(outw, taskLogs, t.commando.ProcessState.String())
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

	for scanner.Scan() {
		t.rl.Lock()
		if !t.running {
			t.rl.Unlock()
			break
		}
		t.rl.Unlock()

		fmt.Fprintf(out, taskLogs, scanner.Text())
	}
}

// Stop ends the task which when initialized.
func (t *Task) Stop(m io.Writer) {
	if t.commando == nil {
		return
	}

	t.rl.Lock()
	if !t.running {
		t.rl.Unlock()
		return
	}
	t.rl.Unlock()

	t.running = false

	var err error

	if t.commando != nil {
		if t.commando.Process != nil {

			if runtime.GOOS == "windows" {
				err = t.commando.Process.Kill()
			} else {
				err = t.commando.Process.Signal(os.Interrupt)
			}

		}

		if t.commando.ProcessState != nil {
			if ws, ok := t.commando.ProcessState.Sys().(syscall.WaitStatus); ok {
				if ws.ExitStatus() != 0 && err != nil {
					fmt.Fprintf(m, taskKill, t.Name, t.Description, t.Command, t.Parameters, err.Error())
				}
			}
		}
	}

	t.commando = nil
}

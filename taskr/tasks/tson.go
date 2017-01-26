package tasks

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/influx6/faux/utils"
)

// TsonSeries defines a higher level Tson manager which handles the management
// of a series of independent tasks providers.
type TsonSeries struct {
	Tasks []*Tson
	wg    sync.WaitGroup
}

// New returns a new instance of a TsonSeries.
func New(tsons ...*Tson) *TsonSeries {
	var series TsonSeries
	series.Tasks = append(series.Tasks, tsons...)

	return &series
}

// Start launches the series of internal Tson tasks managers, returning an error
// if any fails to start.
func (ts *TsonSeries) Start() error {
	for _, tson := range ts.Tasks {
		if err := tson.Start(); err != nil {
			return err
		}

		ts.wg.Add(1)
	}

	return nil
}

// Stop stops the series of internal Tson tasks managers, returning an error
// if any fails to start.
func (ts *TsonSeries) Stop() error {
	for _, tson := range ts.Tasks {
		tson.Stop()
		ts.wg.Done()
	}

	return nil
}

// Wait calls the tson task runner to await all end calls for all tasks shutting
// down the file watchers as well.
func (ts *TsonSeries) Wait() {
	ts.wg.Wait()
}

//==============================================================================

// Tson defines a struct which initializes and sets up a collection of tasks
// which will be printed in accordance with the state of all tasks.
type Tson struct {
	Description string        `json:"desc"`
	Tasks       []*MasterTask `json:"tasks"`
	DirsGlob    string        `json:"dirs_glob,omitempty"`
	FilesGlob   string        `json:"files_glob,omitempty"`
	Files       []string      `json:"files,omitempty"`
	WriteDelay  string        `json:"write_delay"` // in seconds
	Events      string        `json:"events"`
	writedelay  time.Duration
	Sink        io.Writer
	killer      chan struct{}
	restarter   chan struct{}
	starter     chan struct{}
	watcher     *FileSystemWatch
	twriters    *TsonWriter
	wg          sync.WaitGroup
}

// Wait calls the tson task runner to await all end calls for all tasks shutting
// down the file watchers as well.
func (t *Tson) Wait() {
	t.wg.Wait()
}

// Restart restarts the tson task runner.
func (t *Tson) Restart() {
	t.restarter <- struct{}{}
}

// Stop ends the tson task runner.
func (t *Tson) Stop() {
	t.killer <- struct{}{}
}

// Start intializes all internal structure for the runner and initializes each
// individual task runner.
func (t *Tson) Start() error {
	delay, err := utils.GetDuration(t.WriteDelay)
	if err != nil {
		return err
	}

	t.writedelay = delay

	if t.DirsGlob != "" || t.FilesGlob != "" || t.Files != nil {
		watcher, err := FileSystemWatchFromGlob(t.FilesGlob, t.DirsGlob, func(ev fsnotify.Event) {
			if t.Events == "" {
				t.restarter <- struct{}{}
				return
			}

			if t.Events == ev.Op.String() {
				t.restarter <- struct{}{}
				return
			}
		}, nil)

		if err != nil {
			return err
		}

		t.watcher = watcher
	}

	t.wg.Add(1)

	if t.Sink == nil {
		t.Sink = os.Stdout
	}

	t.writeLog(bytes.NewBufferString(fmt.Sprintf("TSON TaskManager: %q\n", t.Description)))
	t.writeLog(bytes.NewBufferString(fmt.Sprintf("TSON Watchers For Event: %q\n", t.Events)))
	t.writeLog(bytes.NewBufferString(fmt.Sprintf("TSON Watchers Files: %+q\n", t.Files)))
	t.writeLog(bytes.NewBufferString(fmt.Sprintf("TSON Watchers DirGlob: %q\n", t.DirsGlob)))
	t.writeLog(bytes.NewBufferString(fmt.Sprintf("TSON Watchers FilesGlob: %q\n", t.FilesGlob)))

	t.killer = make(chan struct{})
	t.starter = make(chan struct{})
	t.restarter = make(chan struct{})
	t.twriters = NewTsonWriter(len(t.Tasks), t.writedelay, t.writeLog)

	if t.watcher != nil {
		if err := t.watcher.Begin(); err != nil {
			return err
		}
	}

	go t.manage()

	t.starter <- struct{}{}

	return nil
}

// writeLog wrties the task output logs.
func (t *Tson) writeLog(bu *bytes.Buffer) {
	fmt.Fprint(t.Sink, bu.String())
}

// restartTasks restarts all tasks in the log.
func (t *Tson) restartTasks() {
	for index, task := range t.Tasks {
		task.Stop(t.twriters.Writer(index))
	}

	for index, task := range t.Tasks {
		go func(ind int, ts *MasterTask) {
			wm := t.twriters.Writer(ind)
			ts.Run(wm, wm)
		}(index, task)
	}
}

// manage handles the managed of the operations of the tson task runner.
func (t *Tson) manage() {
	{
		defer t.wg.Done()

		for {
			select {
			case <-t.starter:
				for index, task := range t.Tasks {
					go func(ind int, ts *MasterTask) {
						wm := t.twriters.Writer(ind)
						ts.Run(wm, wm)
					}(index, task)
				}

			case <-t.restarter:
				t.restartTasks()

			case <-t.killer:
				for index, task := range t.Tasks {
					task.Stop(t.twriters.Writer(index))
				}

				return
			case <-time.After(30 * time.Second):
				continue
			}
		}
	}
}

//==============================================================================

// WriteBlock defines an inteface which exposes certain methods for a block writer.
type WriteBlock interface {
	Reset()
	Bytes() []byte
	Write([]byte) (int, error)
}

// TsonWriter defines a custom writer for the all tasks.
type TsonWriter struct {
	maxWriters int
	wait       time.Duration
	ticker     *time.Timer
	writers    []WriteBlock
	handler    func(*bytes.Buffer)
}

// NewTsonWriter returns a new instance of a TsonWriter.
func NewTsonWriter(maxWriters int, wait time.Duration, handle func(*bytes.Buffer)) *TsonWriter {
	tson := &TsonWriter{
		handler:    handle,
		maxWriters: maxWriters,
		wait:       wait,
	}

	for index := 0; index < maxWriters; index++ {
		tson.writers = append(tson.writers, NewTickWriter(index, tson.tick))
	}

	return tson
}

// Writer calls the giving index with the provided byte.
func (ts *TsonWriter) Writer(index int) io.Writer {
	return ts.writers[index]
}

// Reset resets the writers for all blocks. Basically empties them all out.
func (ts *TsonWriter) Reset() {
	for _, bx := range ts.writers {
		bx.Reset()
	}
}

// tick is called for all internal tson writers that have updates.
func (ts *TsonWriter) tick(index int) {
	if ts.ticker == nil {
		ts.ticker = time.NewTimer(ts.wait)

		go func() {
			<-ts.ticker.C

			var bu bytes.Buffer

			for _, bx := range ts.writers {
				bu.Write(bx.Bytes())
				bx.Reset()
			}

			ts.handler(&bu)

			ts.ticker = nil
		}()

		return
	}

	ts.ticker.Reset(ts.wait)
}

//==============================================================================

// TickWriter defines a writer which calls a function for all writes.
type TickWriter struct {
	*bytes.Buffer
	index  int
	ticker func(int)
}

// NewTickWriter returns a ne instance of a TickWriter.
func NewTickWriter(index int, ticker func(int)) *TickWriter {
	return &TickWriter{
		index:  index,
		ticker: ticker,
		Buffer: bytes.NewBuffer(nil),
	}
}

// Write calls the tickWriter ticker function after writing to update the
// handler of a write.
func (t *TickWriter) Write(bu []byte) (int, error) {
	n, err := t.Buffer.Write(bu)

	if t.ticker != nil {
		t.ticker(t.index)
	}

	return n, err
}

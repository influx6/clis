package tasks

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
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

		go func(tsn *Tson) {
			tson.Wait()
			ts.wg.Done()
		}(tson)

		ts.wg.Add(1)
	}

	return nil
}

// Stop stops the series of internal Tson tasks managers, returning an error
// if any fails to start.
func (ts *TsonSeries) Stop() error {
	for _, tson := range ts.Tasks {
		tson.Stop()
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
	Description   string        `json:"desc"`
	Tasks         []*MasterTask `json:"tasks"`
	DirsGlob      string        `json:"dirs_glob,omitempty"`
	FilesGlob     string        `json:"files_glob,omitempty"`
	Files         []string      `json:"files,omitempty"`
	WriteDelay    string        `json:"write_delay"`
	DebounceDelay string        `json:"debounce_delay"`
	Events        string        `json:"events"`
	writedelay    time.Duration
	Sink          io.Writer
	singleRun     chan struct{}
	killer        chan struct{}
	restarter     chan struct{}
	starter       chan struct{}
	rebooting     int64
	watcher       *FileSystemWatch
	twriters      *TsonWriter
	wg            sync.WaitGroup
	debounce      int64
	ticker        *time.Ticker
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
		debounce, err := utils.GetDuration(t.DebounceDelay)
		if err != nil {
			t.ticker = time.NewTicker(10 * time.Second)
		} else {
			t.ticker = time.NewTicker(debounce)
		}

		watcher, err := FileSystemWatchFromGlob(t.FilesGlob, t.DirsGlob, func(ev fsnotify.Event) {
			fmt.Printf("DB: %q\n", ev.String())
			if atomic.LoadInt64(&t.debounce) > 0 {
				atomic.StoreInt64(&t.debounce, 1)
				return
			}

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
	} else {
		t.ticker = time.NewTicker(30 * time.Second)
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
	t.singleRun = make(chan struct{})
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

// startTasks restarts all tasks in the log.
func (t *Tson) startTasks() {
	atomic.StoreInt64(&t.rebooting, 1)

	for index, task := range t.Tasks {
		go func(ind int, ts *MasterTask) {
			wm := t.twriters.Writer(ind)
			ts.Run(wm, wm)
			t.singleRun <- struct{}{}
		}(index, task)
	}

	atomic.StoreInt64(&t.rebooting, 0)
}

// restartTasks restarts all tasks in the log.
func (t *Tson) restartTasks() {
	atomic.StoreInt64(&t.rebooting, 1)
	for index, task := range t.Tasks {
		task.Stop(t.twriters.Writer(index))
	}

	for index, task := range t.Tasks {
		go func(ind int, ts *MasterTask) {
			wm := t.twriters.Writer(ind)
			ts.Run(wm, wm)
			t.singleRun <- struct{}{}
		}(index, task)
	}

	atomic.StoreInt64(&t.rebooting, 0)
}

// isBooting returns true/false if the task is rebooting.
func (t *Tson) isBooting() bool {
	return atomic.LoadInt64(&t.rebooting) == 1
}

// manage handles the managed of the operations of the tson task runner.
func (t *Tson) manage() {
	var totalDone int64
	totalTask := len(t.Tasks)

	{
		defer t.wg.Done()

		for {
			select {
			case <-t.ticker.C:
				if atomic.LoadInt64(&t.debounce) > 0 {
					atomic.StoreInt64(&t.debounce, 0)
					continue
				}

				atomic.StoreInt64(&t.debounce, 1)
				continue

			case <-t.starter:
				t.startTasks()

			case <-t.singleRun:
				atomic.AddInt64(&totalDone, 1)

				done := int(atomic.LoadInt64(&totalDone))

				if done == totalTask && t.watcher == nil {
					atomic.StoreInt64(&totalDone, 0)

					// Create goroutine to wait until write ends and then kill.
					go func() {
						t.twriters.Wait()
						t.killer <- struct{}{}
					}()
				}

			case <-t.restarter:
				t.restartTasks()

			case <-t.killer:
				for index, task := range t.Tasks {
					task.Stop(t.twriters.Writer(index))
				}

				if t.ticker != nil {
					t.ticker.Stop()
				}

				return
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
	wg         sync.WaitGroup
}

// NewTsonWriter returns a new instance of a TsonWriter.
func NewTsonWriter(maxWriters int, wait time.Duration, handle func(*bytes.Buffer)) *TsonWriter {
	var tson TsonWriter
	tson.handler = handle
	tson.maxWriters = maxWriters
	tson.wait = wait

	for index := 0; index < maxWriters; index++ {
		tson.writers = append(tson.writers, NewTickWriter(index, tson.tick))
	}

	return &tson
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

// Wait checks if the timer has finished else waits for it.
func (ts *TsonWriter) Wait() {
	ts.wg.Wait()
}

// tick is called for all internal tson writers that have updates.
func (ts *TsonWriter) tick(index int) {
	if ts.ticker == nil {
		ts.wg.Add(1)
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
			ts.wg.Done()
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

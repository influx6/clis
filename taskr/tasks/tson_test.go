package tasks_test

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/influx6/clis/taskr/tasks"
)

func TestTson(t *testing.T) {
	var ws sync.WaitGroup
	ws.Add(1)

	tsm := tasks.NewTsonWriter(2, 16*time.Millisecond, func(dl bytes.Buffer) {
		ws.Done()
	})

	tsm.Writer(0).Write([]byte("bottoms down\n"))

	<-time.After(10 * time.Millisecond)

	tsm.Writer(1).Write([]byte("tops down\n"))

	ws.Wait()
}

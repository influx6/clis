package tasks

import (
	"io"
	"time"
)

// MasterTask provides higher level structure which provides a series of tasks
// which would be run in order where the main task is allowed a consistent hold on
// the input and output writers.
// Before and After tasks cant not down the calls, they are given a maximum of
// 5min and then killed.
type MasterTask struct {
	Main      *Task `json:"main"`
	RunTimePT int64 `json:"run_time_per_task"` // values held in seconds.

	Before []*Task `json:"before"`
	After  []*Task `json:"after"`
}

// Stop ends all it's internal tasks.
func (mt *MasterTask) Stop(m io.Writer) {

	// Stop the before tasks.
	for _, tk := range mt.Before {
		if tk.Stopped() {
			continue
		}

		tk.Stop(m)
	}

	mt.Stop(m)

	// Stop the after tasks.
	for _, tk := range mt.After {
		if tk.Stopped() {
			continue
		}

		tk.Stop(m)
	}
}

// Run executes the givin master tasks in the other expected, passing the
// provided writer to collect all responses.
func (mt *MasterTask) Run(mout, merr io.Writer) {

	// Execute the before tasks.
	for _, tk := range mt.Before {
		go func() {
			<-time.After(time.Duration(mt.RunTimePT) * time.Second)
			tk.Stop(mout)
		}()

		tk.Run(mout, merr)
		// tk.Wait()
	}

	// Execute the main tasks and allow it hold io.
	// go func() {
	mt.Main.Run(mout, merr)
	// }()

	// Execute the after tasks.
	for _, tk := range mt.After {
		go func() {
			<-time.After(time.Duration(mt.RunTimePT) * time.Second)
			tk.Stop(mout)
		}()

		tk.Run(mout, merr)
		// tk.Wait()
	}

}

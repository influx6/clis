package tasks

import (
	"github.com/fsnotify/fsnotify"
)

// FileSystemWatch provides a structure which watches a directory and a series of
// provided files for different changes which will be notified by the handle
// passed in.
type FileSystemWatch struct {
	notifier *fsnotify.Watcher
	dirs     []string
	files    []string
	done     chan struct{}
	errors   func(error)
	events   func(fsnotify.Event)
}

// Stop ends the watcher, returning an error if the watcher fails to end appropriately.
func (fs *FileSystemWatch) Stop() error {
	return fs.notifier.Close()
}

// Begin starts the file watchers, adds the directories and file paths which
// should be watched for.
func (fs *FileSystemWatch) Begin() error {
	wc, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	fs.done = make(chan struct{})
	fs.notifier = wc

	go func() {
		for {
			select {
			case <-fs.done:
				return
			case event := <-wc.Events:
				if fs.events != nil {
					fs.events(event)
				}
			case err := <-wc.Errors:
				if fs.errors != nil {
					fs.errors(err)
				}
			}
		}
	}()

	for _, file := range fs.files {
		if err := fs.notifier.Add(file); err != nil {
			return err
		}
	}

	for _, dir := range fs.dirs {
		if err := fs.notifier.Add(dir); err != nil {
			return err
		}
	}

	return nil
}

package tasks

import (
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// FileSystemWatch provides a structure which watches a directory and a series of
// provided files for different changes which will be notified by the handle
// passed in.
type FileSystemWatch struct {
	files    []string
	done     chan struct{}
	errors   func(error)
	events   func(fsnotify.Event)
	notifier *fsnotify.Watcher
}

// FileSystemWatchFromGlob returns a new instance of a FileSystemWatch using the glob
// file and dirs path.
func FileSystemWatchFromGlob(filesGlob []string, ev func(fsnotify.Event), errs func(error)) (*FileSystemWatch, error) {
	var watches []string

	for _, file := range filesGlob {
		files, err := filepath.Glob(file)
		if err != nil {
			return nil, err
		}

		watches = append(watches, files...)
	}

	return NewFileSystemWatch(watches, ev, errs), nil
}

// NewFileSystemWatch returns a new instance of a FileSystemWatch.
func NewFileSystemWatch(files []string, ev func(fsnotify.Event), errs func(error)) *FileSystemWatch {
	return &FileSystemWatch{
		files:  files,
		events: ev,
		errors: errs,
	}
}

// Add add the giving sets of path into the watchers file lists ensuring they
// are added for reuse when restarting and are added into the watcher if started.
func (fs *FileSystemWatch) Add(ms ...string) error {
	if fs.notifier == nil {
		fs.files = append(fs.files, ms...)
		return nil
	}

	fs.files = append(fs.files, ms...)

	for _, file := range ms {
		if err := fs.notifier.Add(file); err != nil {
			return err
		}
	}

	return nil
}

// Stop ends the watcher, returning an error if the watcher fails to end appropriately.
func (fs *FileSystemWatch) Stop() error {
	if err := fs.notifier.Close(); err != nil {
		return err
	}

	fs.notifier = nil
	return nil
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

	return nil
}

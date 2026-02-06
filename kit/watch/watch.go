package watch

import (
	"io/fs"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// Monitors file system paths for changes.
//
// Listens for filesystem events on specified paths and delivers them to a
// callback function. Supports adding and removing paths dynamically, as well
// as recursive watching of directories. Listens for create, write, remove,
// and rename events, delivering events to a given callback. The watcher runs
// until Close is called or the callback returns an error. It may also stop due
// to internal errors.
type Watcher struct {
	watcher *fsnotify.Watcher // Underlying fsnotify watcher
	stop    chan struct{}     // Channel to signal stopping
	stopped chan struct{}     // Channel closed when stopped
	err     error             // Error that caused stop, if any
	errMu   sync.Mutex        // Mutex to protect err
}

// Stops the watcher and releases resources.
//
// After calling Close, the watcher will stop and no further events will be
// delivered. Any blocked calls to Wait will return once the watcher has stopped.
func (w *Watcher) Close() error {
	close(w.stop)
	return w.watcher.Close()
}

// Waits for the watcher to stop and returns any error that caused it to stop.
//
// This blocks until the watcher has fully stopped, either due to Close being
// called, a callback error, or an internal error.
func (w *Watcher) Wait() error {
	<-w.Done()
	return w.err
}

// Returns a channel that is closed when the watcher stops.
//
// This can be used to detect when watching has ended, whether due to Close
// being called, a callback error, or an internal error.
func (w *Watcher) Done() <-chan struct{} {
	return w.stopped
}

// Returns the error that caused the watcher to stop, if any.
//
// Returns nil if the watcher was closed normally via Close or if the callback
// returned an error. Only returns an error for internal failures.
func (w *Watcher) Err() error {
	w.errMu.Lock()
	defer w.errMu.Unlock()
	return w.err
}

func (w *Watcher) setErr(err error) {
	w.errMu.Lock()
	defer w.errMu.Unlock()
	w.err = err
}

// Monitors a single path for file system events.
func Watch(path string, callback EventCallback) (*Watcher, error) {
	return WatchAll([]string{path}, callback)
}

// Monitors a path and all subdirectories for file system events.
//
// Subdirectories are collected at startup. Directories created after the watch
// starts are not automatically watched; use Add or AddRecursive in the callback
// to watch them.
func WatchRecursive(path string, callback EventCallback) (*Watcher, error) {

	dirs, err := collectDirs(path)
	if err != nil {
		return nil, err
	}

	return WatchAll(dirs, callback)
}

// Monitors multiple paths for file system events.
//
// Starts to watch all specified paths and deliver events to the callback. An
// event loop is started in a separate goroutine, which runs until Close is
// called or an error occurs. Returns the Watcher instance or an error if the
// watcher could not be created.
func WatchAll(paths []string, callback EventCallback) (*Watcher, error) {

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Add all paths to watcher
	for _, path := range paths {
		if err := fsw.Add(path); err != nil {
			fsw.Close()
			return nil, err
		}
	}

	w := &Watcher{
		watcher: fsw,
		stop:    make(chan struct{}),
		stopped: make(chan struct{}),
	}

	// Start event loop
	go w.run(callback)

	return w, nil
}

// Adds a path to the watcher.
//
// Events for the path will be delivered to the callback. Returns an error if
// the path cannot be watched.
func (w *Watcher) Add(path string) error {
	return w.watcher.Add(path)
}

// Adds a path and all its subdirectories to the watcher.
//
// Subdirectories are collected at the time of the call. Returns an error if
// any path cannot be watched.
func (w *Watcher) AddRecursive(path string) error {

	// Collect all directories in path
	dirs, err := collectDirs(path)
	if err != nil {
		return err
	}

	return w.AddAll(dirs)
}

// Adds multiple paths to the watcher.
func (w *Watcher) AddAll(paths []string) error {

	for _, path := range paths {
		if err := w.watcher.Add(path); err != nil {
			return err
		}
	}

	return nil
}

// Stops watching a path.
func (w *Watcher) Remove(path string) error {
	return w.watcher.Remove(path)
}

// Stops watching a path and all its subdirectories.
func (w *Watcher) RemoveRecursive(path string) error {
	dirs, err := collectDirs(path)
	if err != nil {
		// Path may already be deleted; try removing just the root
		return w.watcher.Remove(path)
	}

	return w.RemoveAll(dirs)
}

// Stops watching multiple paths.
func (w *Watcher) RemoveAll(paths []string) error {

	for _, path := range paths {
		if err := w.watcher.Remove(path); err != nil {
			return err
		}
	}

	return nil
}

// Implements the main event loop for the watcher.
//
// It listens for events and errors, invoking the callback for each event. The
// loop exits when the stop channel is closed or an error occurs.
func (w *Watcher) run(callback EventCallback) {
	defer close(w.stopped)

	for {
		select {
		case <-w.stop:
			return

		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if err := callback(&WatchEvent{event: event}); err != nil {
				w.setErr(err)
				return
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			w.setErr(err)
			return
		}
	}
}

// Returns a list of all directories under root, including root itself.
func collectDirs(root string) ([]string, error) {
	var dirs []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			dirs = append(dirs, path)
		}

		return nil
	})

	return dirs, err
}

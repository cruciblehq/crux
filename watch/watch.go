package watch

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// Monitors file system paths for changes.
//
// Delivers create, write, remove, and rename events to a callback. Supports
// adding and removing paths dynamically, including recursive directory
// watching. The watcher runs until Close is called, the callback returns an
// error, or the underlying watcher fails.
type Watcher struct {
	watcher   *fsnotify.Watcher // Underlying fsnotify watcher
	stop      chan struct{}     // Channel to signal stopping
	stopped   chan struct{}     // Channel closed when stopped
	err       error             // Error that caused stop, if any
	errMu     sync.Mutex        // Mutex to protect err
	closeOnce sync.Once         // Ensures Close logic runs only once
}

// Stops the watcher and releases resources.
//
// Signals the event loop to stop and blocks until the watcher has fully shut
// down, including closing the underlying file system monitor. After returning,
// no further events will be delivered. Returns any error that caused the
// watcher to stop. If the watcher was closed normally, returns any error from
// releasing the underlying resources. Safe to call multiple times; subsequent
// calls return immediately.
func (w *Watcher) Close() error {
	w.closeOnce.Do(func() {
		close(w.stop)
	})
	<-w.stopped
	return w.Err()
}

// Waits for the watcher to stop and returns any error that caused it to stop.
//
// This blocks until the watcher has fully stopped, either due to Close being
// called, a callback error, or an internal error.
func (w *Watcher) Wait() error {
	<-w.Done()
	return w.Err()
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
// Returns nil if the watcher stopped without error. Returns a callback error
// if the callback returned one, an internal error if the underlying watcher
// failed, or a close error if releasing resources failed.
func (w *Watcher) Err() error {
	w.errMu.Lock()
	defer w.errMu.Unlock()
	return w.err
}

// Sets the error that caused the watcher to stop.
func (w *Watcher) setErr(err error) {
	w.errMu.Lock()
	defer w.errMu.Unlock()
	w.err = err
}

// Monitors a single path for file system events.
func Watch(path string, callback Callback) (*Watcher, error) {
	return WatchAll([]string{path}, callback)
}

// Monitors a path and all subdirectories for file system events.
//
// Subdirectories are collected at startup. Directories created after the watch
// starts are not automatically watched; use Add or AddRecursive in the callback
// to watch them.
func WatchRecursive(path string, callback Callback) (*Watcher, error) {

	dirs, err := collectDirs(path)
	if err != nil {
		return nil, err
	}

	return WatchAll(dirs, callback)
}

// Monitors multiple paths for file system events.
//
// Starts watching all specified paths and delivers events to the callback. An
// event loop is started in a separate goroutine, which runs until Close is
// called or an error occurs. If any path cannot be watched, no paths are
// watched and the error is returned.
func WatchAll(paths []string, callback Callback) (*Watcher, error) {

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		watcher: fsw,
		stop:    make(chan struct{}),
		stopped: make(chan struct{}),
	}

	if err := w.AddAll(paths); err != nil {
		fsw.Close()
		return nil, err
	}

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
//
// If any path fails to watch, all previously added paths in this call are
// rolled back and the error is returned. Either all paths are watched or none.
func (w *Watcher) AddAll(paths []string) error {

	for i, path := range paths {
		if err := w.watcher.Add(path); err != nil {
			w.RemoveAll(paths[:i])
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
//
// If the path has already been deleted and its subdirectories cannot be
// collected, only the root path is removed.
func (w *Watcher) RemoveRecursive(path string) error {
	dirs, err := collectDirs(path)
	if err != nil {
		if os.IsNotExist(err) {
			return w.watcher.Remove(path)
		}
		return err
	}

	return w.RemoveAll(dirs)
}

// Stops watching multiple paths.
//
// All paths are removed on a best-effort basis. If any removals fail, the
// errors are joined and returned. Paths that succeed are unwatched regardless
// of failures in other paths.
func (w *Watcher) RemoveAll(paths []string) error {
	var errs []error

	for _, path := range paths {
		if err := w.watcher.Remove(path); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// Implements the main event loop for the watcher.
//
// It listens for events and errors, invoking the callback for each event. The
// loop exits when the stop channel is closed or an error occurs. The underlying
// watcher is closed when the loop exits; any close error is stored only if no
// prior error exists.
func (w *Watcher) run(callback Callback) {
	defer close(w.stopped)
	defer func() {
		if err := w.watcher.Close(); err != nil && w.Err() == nil {
			w.setErr(err)
		}
	}()

	for {
		select {
		case <-w.stop:
			return

		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if err := callback(&Event{event: event}); err != nil {
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

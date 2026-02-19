package watch

import "github.com/fsnotify/fsnotify"

// Invoked for each file system event.
//
// Return an error to stop the watcher. The error is stored and can be
// retrieved with [Watcher.Err] or [Watcher.Wait].
type Callback func(*Event) error

// Represents a file system event.
type Event struct {
	event fsnotify.Event
}

// Returns the path of the file or directory that triggered the event.
func (e *Event) Path() string {
	return e.event.Name
}

// Returns true if a file or directory was created.
func (e *Event) IsCreate() bool {
	return e.event.Op&fsnotify.Create != 0
}

// Returns true if a file was written to.
func (e *Event) IsWrite() bool {
	return e.event.Op&fsnotify.Write != 0
}

// Returns true if a file or directory was removed.
func (e *Event) IsRemove() bool {
	return e.event.Op&fsnotify.Remove != 0
}

// Returns true if a file or directory was renamed.
func (e *Event) IsRename() bool {
	return e.event.Op&fsnotify.Rename != 0
}

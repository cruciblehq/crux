package watch

import "github.com/fsnotify/fsnotify"

// EventCallback is invoked for each file system event.
//
// Return an error to stop the watcher. The error is not stored; use it for
// control flow only.
type EventCallback func(*WatchEvent) error

// Represents a file system event.
type WatchEvent struct {
	event fsnotify.Event
}

// Returns the path of the file or directory that triggered the event.
func (e *WatchEvent) Path() string {
	return e.event.Name
}

// Returns true if a file or directory was created.
func (e *WatchEvent) IsCreate() bool {
	return e.event.Op&fsnotify.Create != 0
}

// Returns true if a file was written to.
func (e *WatchEvent) IsWrite() bool {
	return e.event.Op&fsnotify.Write != 0
}

// Returns true if a file or directory was removed.
func (e *WatchEvent) IsRemove() bool {
	return e.event.Op&fsnotify.Remove != 0
}

// Returns true if a file or directory was renamed.
func (e *WatchEvent) IsRename() bool {
	return e.event.Op&fsnotify.Rename != 0
}

// Package watch provides file system monitoring.
//
// A [Watcher] observes one or more paths for changes and invokes a callback on
// every event. The callback receives an [Event] that wraps the underlying
// fsnotify event with convenience methods for inspecting the operation kind.
// If the callback returns an error, the watcher shuts down and the error is
// surfaced through [Watcher.Wait] and [Watcher.Err].
//
// [Watch] observes a single path. [WatchRecursive] adds a directory and all of
// its existing subdirectories; newly created subdirectories are not
// automatically watched and must be added from within the callback using
// [Watcher.Add] or [Watcher.AddRecursive]. Paths can also be removed at
// runtime with [Watcher.Remove]. [Watcher.Close] stops the watcher and blocks
// until the event loop has fully shut down. [Watcher.Wait] also blocks until
// the event loop exits and returns any error.
//
//	w, err := watch.Watch("/path/to/file", func(e *watch.Event) error {
//	    if e.IsWrite() {
//	        fmt.Println("modified:", e.Path())
//	    }
//	    return nil
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	c := make(chan os.Signal, 1)
//	signal.Notify(c, os.Interrupt)
//	<-c
//
//	w.Close()
package watch

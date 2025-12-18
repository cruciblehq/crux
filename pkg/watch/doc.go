// Package watch provides file system monitoring.
//
// Basic usage:
//
//	w, err := watch.Watch("/path/to/file", func(e *watch.WatchEvent) error {
//	    fmt.Printf("%s: %s\n", e.Op(), e.Path())
//	    return nil
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Block until interrupt
//	c := make(chan os.Signal, 1)
//	signal.Notify(c, os.Interrupt)
//	<-c
//
//	w.Close()
//
// Recursive watching:
//
//	w, err := watch.WatchRecursive("/path/to/dir", callback)
//
// Dynamic path management:
//
//	w.Add("/another/path")
//	w.AddRecursive("/another/dir")
//	w.Remove("/path/to/file")
//
// Waiting for completion:
//
//	if err := w.Wait(); err != nil {
//	    log.Printf("watcher error: %v", err)
//	}
package watch

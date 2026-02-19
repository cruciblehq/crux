package watch

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatch_Create(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "testwatch.txt")

	// Buffer to avoid blocking the watcher goroutine
	got := make(chan *Event, 10)

	w, err := Watch(dir, func(e *Event) error {
		got <- e
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	// Create file
	if err := os.WriteFile(file, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	// Wait for event
	select {
	case e := <-got:
		if !e.IsCreate() {
			t.Error("expected create event")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestWatch_Write(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "testwatchwrite.txt")

	// Create file first
	if err := os.WriteFile(file, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	// Buffer to avoid blocking the watcher goroutine
	got := make(chan *Event, 10)

	w, err := Watch(dir, func(e *Event) error {
		got <- e
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	// Modify file
	if err := os.WriteFile(file, []byte("world"), 0644); err != nil {
		t.Fatal(err)
	}

	// Wait for write event
	timeout := time.After(time.Second)
	for {
		select {
		case e := <-got:
			if e.IsWrite() {
				return // Success
			}
		case <-timeout:
			t.Fatal("timeout waiting for write event")
		}
	}
}

func TestWatch_Remove(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "testwatchremove.txt")

	// Create file first
	if err := os.WriteFile(file, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	// Buffer to avoid blocking the watcher goroutine
	got := make(chan *Event, 10)

	w, err := Watch(dir, func(e *Event) error {
		got <- e
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	// Remove file
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}

	timeout := time.After(time.Second)
	for {
		select {
		case e := <-got:
			if e.IsRemove() {
				return
			}
		case <-timeout:
			t.Fatal("timeout waiting for remove event")
		}
	}
}

func TestWatch_Rename(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "testwatchrename.txt")
	renamed := filepath.Join(dir, "renamed.txt")

	// Create file first
	if err := os.WriteFile(file, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	// Buffer to avoid blocking the watcher goroutine
	got := make(chan *Event, 10)

	w, err := Watch(dir, func(e *Event) error {
		got <- e
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	// Rename file
	if err := os.Rename(file, renamed); err != nil {
		t.Fatal(err)
	}

	timeout := time.After(time.Second)
	for {
		select {
		case e := <-got:
			if e.IsRename() {
				return
			}
		case <-timeout:
			t.Fatal("timeout waiting for rename event")
		}
	}
}

func TestWatch_NonexistentPath(t *testing.T) {
	_, err := Watch("/doesnotexist/path", func(e *Event) error {
		return nil
	})
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestWatch_CallbackError(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "testwatchcallbackerror.txt")

	callbackErr := errors.New("callback error")

	w, err := Watch(dir, func(e *Event) error {
		return callbackErr
	})
	if err != nil {
		t.Fatal(err)
	}

	// Trigger event
	if err := os.WriteFile(file, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	// Wait for watcher to stop
	select {
	case <-w.Done():
	case <-time.After(time.Second):
		t.Fatal("watcher did not stop after callback error")
	}

	// Callback errors should be stored
	if w.Err() == nil {
		t.Errorf("expected non-nil error, got %v", w.Err())
	}
}

func TestWatch_Close(t *testing.T) {
	dir := t.TempDir()

	w, err := Watch(dir, func(e *Event) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := w.Close(); err != nil {
		t.Errorf("unexpected error from Close: %v", err)
	}

	select {
	case <-w.Done():
	case <-time.After(time.Second):
		t.Fatal("Done not closed after Close")
	}
}

func TestWatch_Wait(t *testing.T) {
	dir := t.TempDir()

	w, err := Watch(dir, func(e *Event) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error)
	go func() {
		done <- w.Wait()
	}()

	w.Close()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("unexpected error from Wait: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Wait did not return after Close")
	}
}

func TestWatchRecursive_Success(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "subdir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	// Buffer to avoid blocking the watcher goroutine
	got := make(chan *Event, 10)

	w, err := WatchRecursive(dir, func(e *Event) error {
		got <- e
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	// Create file in subdir
	file := filepath.Join(subdir, "testwatchrecursive.txt")
	if err := os.WriteFile(file, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case e := <-got:
		if e.Path() != file {
			t.Errorf("expected %q, got %q", file, e.Path())
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event from subdirectory")
	}
}

func TestWatchAll_Success(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	// Buffer to avoid blocking the watcher goroutine
	got := make(chan *Event, 10)

	w, err := WatchAll([]string{dir1, dir2}, func(e *Event) error {
		got <- e
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	// Create file in first dir
	file1 := filepath.Join(dir1, "test1.txt")
	if err := os.WriteFile(file1, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-got:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event from dir1")
	}

	// Create file in second dir
	file2 := filepath.Join(dir2, "test2.txt")
	if err := os.WriteFile(file2, []byte("world"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-got:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event from dir2")
	}
}

func TestWatchAll_PartialFailure(t *testing.T) {
	dir := t.TempDir()

	_, err := WatchAll([]string{dir, "/nonexistent/path"}, func(e *Event) error {
		return nil
	})
	if err == nil {
		t.Error("expected error for partial failure")
	}
}

func TestWatcher_Add(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	// Buffer to avoid blocking the watcher goroutine
	got := make(chan *Event, 10)

	w, err := Watch(dir1, func(e *Event) error {
		got <- e
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	if err := w.Add(dir2); err != nil {
		t.Fatal(err)
	}

	// Create file in second dir
	file := filepath.Join(dir2, "testwatchadd.txt")
	if err := os.WriteFile(file, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case e := <-got:
		if e.Path() != file {
			t.Errorf("expected %q, got %q", file, e.Path())
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event from added directory")
	}
}

func TestWatcher_AddRecursive(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "subdir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	watchDir := t.TempDir()

	// Buffer to avoid blocking the watcher goroutine
	got := make(chan *Event, 10)

	w, err := Watch(watchDir, func(e *Event) error {
		got <- e
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	if err := w.AddRecursive(dir); err != nil {
		t.Fatal(err)
	}

	// Create file in subdir
	file := filepath.Join(subdir, "test.txt")
	if err := os.WriteFile(file, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case e := <-got:
		if e.Path() != file {
			t.Errorf("expected path %q, got %q", file, e.Path())
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event from recursively added directory")
	}
}

func TestWatcher_Remove(t *testing.T) {
	dir := t.TempDir()

	// Buffer to avoid blocking the watcher goroutine
	got := make(chan *Event, 10)

	w, err := Watch(dir, func(e *Event) error {
		got <- e
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	if err := w.Remove(dir); err != nil {
		t.Fatal(err)
	}

	// Create file - should not trigger event
	file := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(file, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-got:
		t.Error("unexpected event after Remove")
	case <-time.After(100 * time.Millisecond):
		// Expected - no event
	}
}

func TestWatcher_RemoveRecursive(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "subdir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	// Buffer to avoid blocking the watcher goroutine
	got := make(chan *Event, 10)

	w, err := WatchRecursive(dir, func(e *Event) error {
		got <- e
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	if err := w.RemoveRecursive(dir); err != nil {
		t.Fatal(err)
	}

	// Create file in subdir - should not trigger event
	file := filepath.Join(subdir, "test.txt")
	if err := os.WriteFile(file, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-got:
		t.Error("unexpected event after RemoveRecursive")
	case <-time.After(100 * time.Millisecond):
		// Expected - no event
	}
}

func TestWatcher_RemoveAll(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	// Buffer to avoid blocking the watcher goroutine
	got := make(chan *Event, 10)

	w, err := WatchAll([]string{dir1, dir2}, func(e *Event) error {
		got <- e
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	if err := w.RemoveAll([]string{dir1, dir2}); err != nil {
		t.Fatal(err)
	}

	// Create files - should not trigger events
	if err := os.WriteFile(filepath.Join(dir1, "test.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir2, "test.txt"), []byte("world"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-got:
		t.Error("unexpected event after RemoveAll")
	case <-time.After(100 * time.Millisecond):
		// Expected - no event
	}
}

func TestCollectDirs_Success(t *testing.T) {
	dir := t.TempDir()
	sub1 := filepath.Join(dir, "sub1")
	sub2 := filepath.Join(dir, "sub1", "sub2")

	if err := os.MkdirAll(sub2, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a file (should not be collected)
	file := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(file, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	dirs, err := collectDirs(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(dirs) != 3 {
		t.Errorf("expected 3 directories, got %d: %v", len(dirs), dirs)
	}

	expected := map[string]bool{dir: false, sub1: false, sub2: false}
	for _, d := range dirs {
		if _, ok := expected[d]; ok {
			expected[d] = true
		}
	}
	for d, found := range expected {
		if !found {
			t.Errorf("missing directory: %s", d)
		}
	}
}

func TestCollectDirs_NonexistentPath(t *testing.T) {
	_, err := collectDirs("/nonexistent/path")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestWatcher_AddAll_Rollback(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	got := make(chan *Event, 10)

	w, err := Watch(t.TempDir(), func(e *Event) error {
		got <- e
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	// Add two valid dirs and one invalid; all should be rolled back
	err = w.AddAll([]string{dir1, dir2, "/nonexistent/path"})
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}

	// dir1 and dir2 should not be watched after rollback
	if err := os.WriteFile(filepath.Join(dir1, "test.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir2, "test.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case e := <-got:
		t.Errorf("unexpected event after rollback: %s", e.Path())
	case <-time.After(100 * time.Millisecond):
		// Expected - no events
	}
}

func TestWatcher_RemoveAll_PartialFailure(t *testing.T) {
	dir := t.TempDir()

	w, err := Watch(dir, func(e *Event) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	// Remove one valid path and one that was never watched
	err = w.RemoveAll([]string{dir, "/never/watched"})
	if err == nil {
		t.Fatal("expected error for unwatched path")
	}

	// The valid path should still have been removed despite the error.
	// Confirm the error was returned.
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestWatcher_Close_Idempotent(t *testing.T) {
	dir := t.TempDir()

	w, err := Watch(dir, func(e *Event) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := w.Close(); err != nil {
		t.Errorf("first Close returned error: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Errorf("second Close returned error: %v", err)
	}

	select {
	case <-w.Done():
	case <-time.After(time.Second):
		t.Fatal("Done not closed after double Close")
	}
}

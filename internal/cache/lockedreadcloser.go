package cache

import "os"

// Wraps an [os.File] with an unlock function that is called on [Close].
type lockedReadCloser struct {
	file   *os.File
	unlock func()
}

// Delegates to the underlying file's [Read] method.
func (r *lockedReadCloser) Read(p []byte) (int, error) {
	return r.file.Read(p)
}

// Closes the underlying file and releases the lock.
func (r *lockedReadCloser) Close() error {
	defer r.unlock()
	return r.file.Close()
}

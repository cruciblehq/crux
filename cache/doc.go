// Package cache provides a local cache for downloaded resources, guarded by a
// file lock.
//
// The cache stores downloaded resource archives locally, avoiding redundant
// downloads from remote registries. All operations are protected by a file
// lock across processes and a mutex within the process.
//
// The cache root (<cache-root> below) is the registry directory under the crux
// XDG cache directory (e.g. ~/Library/Caches/crux/registry on macOS,
// ~/.cache/crux/registry on Linux).
//
//	<cache-root>/
//	  cache.lock                                  File lock
//	  archives/                                   Downloaded archives
//	    <namespace>/<resource>/<version>/
//	      meta.json                               Version metadata (JSON)
//	      archive.tar.zst                         Compressed archive
//	  extracted/                                  Extracted contents (on demand)
//	    <namespace>/<resource>/<version>/
//	      ...                                     Archive contents
//
// Archives and extracted contents are stored in parallel trees. Removing an
// entry via [Cache.Delete] or [Cache.Clear] removes both the archive and any
// extracted contents.
//
// Opening the cache and storing an archive:
//
//	c, err := cache.Open()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer c.Close()
//
//	ver, err := c.Put("my-namespace", "my-resource", "1.0.0", archiveReader)
//
// Extracting cached contents:
//
//	dir, err := c.Extract("my-namespace", "my-resource", "1.0.0")
package cache

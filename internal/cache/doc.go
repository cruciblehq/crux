// Package cache provides a local file-locked cache for downloaded resources.
//
// The cache stores downloaded resource archives locally, avoiding redundant
// downloads from remote registries. Artifacts are organized into
// <namespace>/<resource>/<version> directories with metadata stored as JSON
// alongside the archive files. All operations are protected by file locks to
// allow safe concurrent access from multiple processes.
//
// Opening the cache and storing an archive:
//
//	c, err := cache.Open(ctx, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer c.Close()
//
//	ver, err := c.Put(ctx, "my-namespace", "my-resource", "1.0.0", archiveReader)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Retrieving a cached version:
//
//	ver, err := c.Get(ctx, "my-namespace", "my-resource", "1.0.0")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(ver.String, *ver.Digest)
package cache

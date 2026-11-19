package lima

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cruciblehq/utils-go/file"
)

// Removes the staged image file id if it lies inside tempDir.
//
// Images outside tempDir belong to the caller and are left in place.
func RemoveImageIfStaged(ctx context.Context, id, tempDir string) error {
	abs, err := filepath.Abs(id)
	if err != nil {
		return err
	}
	prefix, err := filepath.Abs(tempDir)
	if err != nil {
		return err
	}
	if !file.IsInside(abs, prefix) {
		return nil
	}
	return os.Remove(abs)
}

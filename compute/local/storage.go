package local

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cruciblehq/utils-go/crex"
)

// Creates a volume directory under root with the given permissions.
//
// Returns the absolute path to the created directory.
func CreateVolume(root, id string, mode os.FileMode) (string, error) {
	if mode == 0 {
		mode = DefaultVolumeMode
	}
	dir := filepath.Join(root, VolumesDir, id)
	if err := os.MkdirAll(dir, mode); err != nil {
		return "", crex.Wrap(ErrStorage, err)
	}
	return dir, nil
}

// Removes a volume directory and its contents.
func RemoveVolume(root, id string) error {
	dir := filepath.Join(root, VolumesDir, id)
	if err := os.RemoveAll(dir); err != nil {
		return crex.Wrap(ErrStorage, err)
	}
	return nil
}

// Lists volume directories present under root.
//
// Each subdirectory of <root>/volumes is reported as a volume ID.
func ListVolumes(root string) ([]string, error) {
	dir := filepath.Join(root, VolumesDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, crex.Wrap(ErrStorage, err)
	}
	var ids []string
	for _, e := range entries {
		if e.IsDir() {
			ids = append(ids, e.Name())
		}
	}
	return ids, nil
}

// Returns the host-local path to a volume directory.
func VolumePath(root, id string) string {
	return filepath.Join(root, VolumesDir, id)
}

// Resolves a machine image path, removing the staged copy if it resides in
// the given temp directory.
func RemoveImageIfStaged(ctx context.Context, id, tempDir string) error {
	abs, err := filepath.Abs(id)
	if err != nil {
		return err
	}
	prefix, err := filepath.Abs(tempDir)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(prefix, abs)
	if err != nil || len(rel) > 0 && rel[0] == '.' {
		return nil
	}
	return os.Remove(abs)
}

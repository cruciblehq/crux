package machines

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cruciblehq/spec/trust"
	"github.com/cruciblehq/utils-go/codec"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/file"
)

// What crux knows about one machine.
type Record struct {
	ID    trust.MachineID `codec:"id"`              // Identifier crux assigned.
	Token []byte          `codec:"token,omitempty"` // Enrollment token, empty once enrolled.
}

// Whether the machine has completed enrollment.
func (r Record) Enrolled() bool {
	return len(r.Token) == 0
}

// Returns a record for a new machine with a fresh identifier and token.
func New() Record {
	return Record{
		ID:    trust.NewMachineID(),
		Token: trust.NewToken(),
	}
}

// Reads the record for the machine name from dir.
//
// Returns [ErrNotFound] when there is no record.
func Load(dir, name string) (Record, error) {
	data, err := os.ReadFile(path(dir, name))
	if errors.Is(err, fs.ErrNotExist) {
		return Record{}, crex.Newf(ErrNotFound, "no record for machine %q", name)
	}
	if err != nil {
		return Record{}, err
	}
	var r Record
	if err := codec.Decode(data, &r, codec.JSON); err != nil {
		return Record{}, crex.Wrap(ErrInvalidRecord, err)
	}
	if err := r.ID.Validate(); err != nil {
		return Record{}, crex.Wrap(ErrInvalidRecord, err)
	}
	return r, nil
}

// Writes the record for the machine name to dir.
func Save(dir, name string, r Record) error {
	if err := os.MkdirAll(dir, file.OwnerDirMode); err != nil {
		return err
	}
	data, err := codec.Encode(r, codec.JSON)
	if err != nil {
		return err
	}
	return file.WriteFile(path(dir, name), data, file.OwnerFileMode)
}

// Deletes the record for the machine name from dir, if there is one.
func Remove(dir, name string) error {
	if err := os.Remove(path(dir, name)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

// Returns the file of the record for the machine name.
func path(dir, name string) string {
	return filepath.Join(dir, fmt.Sprintf("%s.json", name))
}

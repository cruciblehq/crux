package fcap

// Selects how file capabilities are granted on a binary.
//
// Effective mode also sets the file effective bit, which causes the granted
// capabilities to be active immediately after execve without the new process
// having to raise them itself. Inheritable mode requires the calling process
// to already hold the capabilities in its inheritable set.
type Mode string

const (

	// File-permitted plus effective bit.
	//
	// Capabilities are immediately active after execve without the process
	// needing to raise them.
	ModeEffective Mode = "effective"

	// File-inheritable.
	//
	// Capabilities take effect only if the caller already holds them in its
	// inheritable set.
	ModeInheritable Mode = "inheritable"
)

// Parses s as a Mode and returns the validated value.
//
// Returns ErrUnknownFcapMode if s does not match a known mode.
func ParseMode(s string) (Mode, error) {
	m := Mode(s)
	switch m {
	case ModeEffective, ModeInheritable:
		// Valid modes.
	default:
		return "", ErrUnknownFcapMode
	}
	return m, nil
}

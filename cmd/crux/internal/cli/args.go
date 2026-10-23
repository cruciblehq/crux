package cli

// Argument separator used by kong to distinguish between command flags and
// positional arguments.
const argSeparator = "--"

// Drops a single leading "--" from args (kong's argument separator).
//
// Used by passthrough commands so the user can write `crux ... -- cmd a b`
// without the separator reaching the executed command.
func stripArgSeparator(args []string) []string {
	if len(args) > 0 && args[0] == argSeparator {
		return args[1:]
	}
	return args
}

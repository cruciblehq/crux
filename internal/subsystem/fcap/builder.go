package fcap

// Compiles file capability grants from rule strings.
//
// Owns the accumulated fcap state and deduplicates grants across multiple
// Build calls within a single compilation session.
type Builder struct {
	state *State
}

// Returns a new Builder with no accumulated state.
func NewBuilder() *Builder {
	return &Builder{}
}

// Returns a new Builder initialized with existing state.
func NewBuilderWithState(state *State) *Builder {
	return &Builder{state: cloneState(state)}
}

// Parses an fcap rule and merges it into the accumulated state.
//
// Rules have the form "path mode cap" where path is an absolute filesystem
// path, mode is "effective" or "inheritable", and cap is a capability name
// without the CAP_ prefix. The mode determines whether capabilities become
// file-permitted (effective on exec) or file-inheritable (only effective if
// the caller also holds them in its inheritable set).
func (b *Builder) Build(rule string) error {
	g, err := Parse(rule)
	if err != nil {
		return err
	}
	if b.state == nil {
		b.state = NewState()
	}
	_, err = b.state.Apply(g)
	return err
}

// Returns a snapshot of the accumulated fcap state.
//
// The returned state is a deep copy of the accumulated state. If the state is
// nil or empty, returns nil. Otherwise, the returned state is non-nil with a
// non-empty Entries map.
func (b *Builder) State() *State {
	return cloneState(b.state)
}

// Merges another fcap state into the accumulated state.
//
// If the input state is nil or empty, does nothing. Otherwise, merges all
// entries from the input state into the accumulated state, creating entries
// as needed.
func (b *Builder) Merge(other *State) error {
	if other == nil {
		return nil
	}
	if b.state == nil {
		b.state = cloneState(other)
		return nil
	}
	b.state.Merge(other)
	return nil
}

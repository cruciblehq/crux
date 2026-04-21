package cap

// Compiles capability grants from rule strings.
//
// Owns the accumulated capability state and deduplicates grants across
// multiple calls within a single compilation session.
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

// Parses a capability rule and merges it into the accumulated state.
//
// Rules begin with an optional mode keyword followed by a capability name
// (e.g. "effective net_admin"). The mode selects which of the five kernel
// capability sets to populate. The mode defaults to ModeFull.
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

// Returns a copy of the accumulated capability state.
func (b *Builder) State() *State {
	return cloneState(b.state)
}

// Merges another capability state into the accumulated state.
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

package cap

// Compiles capability grants from rule strings.
//
// Owns the accumulated capability model and deduplicates grants across
// multiple calls within a compilation session.
type Builder struct {
	model *Model // Accumulated capability model.
}

// Returns a new Builder with no accumulated model.
//
// The returned Builder owns its own accumulated model. The Builder may be
// reused across multiple Build calls within a single compilation session,
// accumulating grants into the model across calls.
func NewBuilder() *Builder {
	return &Builder{}
}

// Returns a new Builder initialized with existing model.
//
// The input model is deep-copied to ensure the Builder owns its own copy of
// the accumulated model. The input model is not modified by the Builder and
// may be safely reused across multiple Builder instances.
func NewBuilderWithModel(model *Model) *Builder {
	return &Builder{model: cloneModel(model)}
}

// Parses a capability rule and merges it into the accumulated model.
//
// Rules begin with an optional mode keyword followed by a capability name
// (e.g. "effective net_admin"). The mode selects which of the five kernel
// capability sets to populate. The mode defaults to ModeFull.
func (b *Builder) Build(rule string) error {
	delta, err := Parse(rule)
	if err != nil {
		return err
	}
	return b.Merge(delta)
}

// Returns a copy of the accumulated capability model.
//
// The returned model is a deep copy of the accumulated model. If the model
// is nil or empty, returns nil. Otherwise, the returned model is non-nil
// with at least one non-empty capability set.
func (b *Builder) Model() *Model {
	return cloneModel(b.model)
}

// Merges another capability model into the accumulated model.
//
// If the input model is nil or empty, does nothing. Otherwise, incorporates
// all capability names from each set.
func (b *Builder) Merge(other *Model) error {
	if other == nil {
		return nil
	}
	if b.model == nil {
		b.model = cloneModel(other)
		return nil
	}
	b.model.Merge(other)
	return nil
}

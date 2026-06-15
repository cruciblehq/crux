package cgroup

// Constraint for slice elements that participate in identity-keyed merges.
//
// The three methods define the entry's split between identity (the fields
// whose equality makes two entries "the same entry") and payload (everything
// else, which is checked for conflicts and merged).
type mergeable[T any] interface {
	*T

	// Whether the receiver and other share the same identity keys.
	//
	// Each type defines its own identity keys. For example, for the I/O cost
	// model, the identity keys are the device major and minor numbers, while
	// the payload is the cost parameters.
	equal(T) bool

	// Whether the receiver and other have conflicting payload values.
	//
	// Returns nil if they are not the same entry or if they are identical, and
	// ErrConflict if they have the same identity but disagree on their payload.
	check(T) error

	// Folds other into the receiver in place.
	//
	// Returns whether the receiver was changed. The merge logic only calls
	// merge if the entries have the same identity and non-conflicting payloads,
	// so implementations can assume that other is safe to merge.
	merge(T) bool
}

// Folds src into slice by element identity, appending or merging.
//
// Walks slice once. On the first identity match, runs check; on conflict
// the wrapped ErrConflict is returned unchanged. Otherwise merges in place
// and returns whether the existing entry actually changed. With no match,
// appends src and reports change=true. PT is the pointer-to-element type
// so that merge can mutate the slice slot through *T.
func merge[T any, PT mergeable[T]](slice *[]T, src T) (bool, error) {
	for i := range *slice {
		e := PT(&(*slice)[i])
		if !e.equal(src) {
			continue
		}
		if err := e.check(src); err != nil {
			return false, err
		}
		return e.merge(src), nil
	}
	*slice = append(*slice, src)
	return true, nil
}

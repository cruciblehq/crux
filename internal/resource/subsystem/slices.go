package subsystem

// Appends s to the slice at dst if s is not already present.
//
// Returns true if s was added.
func appendUnique(dst *[]string, s string) bool {
	for _, existing := range *dst {
		if existing == s {
			return false
		}
	}
	*dst = append(*dst, s)
	return true
}

// Merges src into dst via [appendUnique].
//
// Returns true if any element was added.
func mergeSlice(dst *[]string, src []string) bool {
	changed := false
	for _, s := range src {
		if appendUnique(dst, s) {
			changed = true
		}
	}
	return changed
}

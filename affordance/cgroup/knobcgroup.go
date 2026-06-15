package cgroup

import (
	"slices"
	"strings"

	"github.com/cruciblehq/crux/crex"
)

// Core cgroup hierarchy controls.
//
// These controls apply to the cgroup hierarchy as a whole rather than to any
// particular resource controller, covering freezing, node type, descendant
// and depth caps, pressure reporting, and the subtree control delegation
// list.
type cgroup struct {
	Freeze         bool         `knob:"freeze" default:"true" json:"freeze,omitempty"`   // Whether the cgroup is frozen (processes within it are paused).
	Type           nodeType     `knob:"type" default:"domain" json:"type,omitempty"`     // Cgroup node type (domain or threaded).
	MaxDescendants uint32       `knob:"max.descendants" json:"maxDescendants,omitempty"` // Maximum number of descendant cgroups.
	MaxDepth       uint32       `knob:"max.depth" json:"maxDepth,omitempty"`             // Maximum depth of the cgroup hierarchy.
	Pressure       bool         `knob:"pressure" json:"pressure,omitempty"`              // Whether pressure stall information is enabled.
	SubtreeControl []controller `knob:"subtree_control" json:"subtreeControl,omitempty"` // List of controllers delegated to child cgroups.
}

// Node type within the cgroup hierarchy.
//
// cgroup v2 supports both process-granularity and thread-granularity cgroups.
// The node type determines which granularity to use. The domain type creates
// process-granularity cgroups where all threads of a process must be in the
// same cgroup. The threaded type creates thread-granularity cgroups where
// threads of the same process can be in different cgroups.
type nodeType string

const (
	nodeTypeDomain   nodeType = "domain"   // Process-granularity cgroup (default).
	nodeTypeThreaded nodeType = "threaded" // Thread-granularity cgroup.
)

// Parses a cgroup node type name.
func parseNodeType(value string) (nodeType, error) {
	s := strings.TrimSpace(value)
	switch nodeType(s) {
	case nodeTypeDomain, nodeTypeThreaded:
		return nodeType(s), nil
	default:
		return "", crex.Newf(ErrInvalidGrant, "invalid cgroup node type %q", value)
	}
}

// Merges an incoming subtree_control list into s.
//
// Returns added=true when the list was written; returns an error when the
// existing list is non-empty and differs from incoming.
func (s *spec) mergeSubtreeControl(incoming []controller) (bool, error) {
	if len(incoming) == 0 {
		return false, nil
	}
	if len(s.Cgroup.SubtreeControl) > 0 {
		if slices.Equal(s.Cgroup.SubtreeControl, incoming) {
			return false, nil
		}
		return false, crex.Newf(ErrConflict, "%s already set", subtreeControlKnob)
	}
	s.Cgroup.SubtreeControl = slices.Clone(incoming)
	return true, nil
}

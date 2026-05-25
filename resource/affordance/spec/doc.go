// Package spec provides the unified runtime spec accumulated by the affordance
// builder.
//
// NewSpec returns a Spec initialised to the strictest possible state; every
// subsystem can only loosen it. ToSpec converts the accumulated Spec to the
// serialisable manifest form at the end of a build session.
package spec

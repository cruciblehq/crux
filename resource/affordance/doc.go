// Package affordance implements the grant scope resolution pipeline for
// Crucible affordances.
//
// [Builder.Resolve] expands affordance references and dispatches domain grants
// to the appropriate runtime subsystem. [PullAffordance] fetches an affordance
// resource from the registry.
//
//	b := affordance.NewBuilder()
//	resolved, err := b.Resolve(ctx, scopes, source)
//	spec := b.Spec()
package affordance

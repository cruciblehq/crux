package compute

import (
	"context"

	speccompute "github.com/cruciblehq/spec/compute"
)

// Make sure that the direct containerd implementation satisfies the spec.
//
// [Container] satisfies [speccompute.Container] directly. [Client] satisfies
// [speccompute.Runtime] except that its Load returns the concrete [Container];
// runtimeAdapter bridges that difference by returning the interface type. The
// crux client replaces this adapter with the HTTP-backed implementation in a
// later phase; here it proves the direct Client implements every runtime method
// with the contract's signatures.
var (
	_ speccompute.Container = (*Container)(nil)
	_ speccompute.Runtime   = runtimeAdapter{}
)

// Adapts [Client] to [speccompute.Runtime].
//
// Embeds the client to promote its image and lifecycle methods and overrides
// only Load so that it returns the [speccompute.Container] interface rather than
// the concrete [Container].
type runtimeAdapter struct {
	*Client
}

// Creates a container and returns it as the contract interface type.
func (a runtimeAdapter) Load(ctx context.Context, ref speccompute.ImageRef, opts speccompute.Options) (speccompute.Container, error) {
	return a.Client.Load(ctx, ref, opts)
}

package provision

// Resource requirements for one service, declared by .provision grants.
//
// The blueprint builder reads this after processing a service's affordances.
// It uses the values to bin-pack services onto compute units and select a
// concrete instance type that satisfies the totals.
type Spec struct {

	// CPU allocation.
	//
	// Measured in millicores (1 vCPU = 1000m). The blueprint builder rounds up
	// to the nearest vCPU when selecting instance types, so fractional millicore
	// values can express requirements that are less than 1 vCPU.
	CPU uint64

	// Memory allocation.
	//
	// Measured in bytes. The blueprint builder rounds up to the nearest MiB
	// when selecting instance types.
	Memory uint64

	// Root disk allocation.
	//
	// Measured in bytes. The blueprint builder rounds up to the nearest MiB
	// when selecting instance types.
	Disk uint64
}

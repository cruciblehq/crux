package cgroup

// Memory limits and reclaim protection for the cgroup.
//
// Groups the cgroup v2 memory controller knobs that bound RAM, swap, and
// zswap usage and that influence the kernel's reclaim and OOM behavior.
type memory struct {
	Max            uint64 `knob:"max" json:"max,omitempty"`                           // Maximum memory usage in bytes.
	High           uint64 `knob:"high" json:"high,omitempty"`                         // Soft memory limit in bytes.
	Min            uint64 `knob:"min" json:"min,omitempty"`                           // Minimum memory limit in bytes.
	Low            uint64 `knob:"low" json:"low,omitempty"`                           // Low memory limit in bytes.
	SwapMax        uint64 `knob:"swap.max" json:"swapMax,omitempty"`                  // Maximum swap usage in bytes.
	SwapHigh       uint64 `knob:"swap.high" json:"swapHigh,omitempty"`                // Soft swap limit in bytes.
	OOMGroup       bool   `knob:"oom.group" default:"true" json:"oomGroup,omitempty"` // Whether to kill the entire cgroup atomically on OOM (true contains blast radius).
	ZswapMax       uint64 `knob:"zswap.max" json:"zswapMax,omitempty"`                // Maximum zswap usage in bytes.
	ZswapWriteback bool   `knob:"zswap.writeback" json:"zswapWriteback,omitempty"`    // Whether to write back zswap pages to the backing device when memory is under pressure.
}

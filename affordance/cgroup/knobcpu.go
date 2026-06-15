package cgroup

// CPU bandwidth and scheduling controls for the cgroup.
//
// Groups the cgroup v2 cpu controller knobs that bound runtime per
// scheduling period, set proportional weights, and request utilization
// clamps from the scheduler.
type cpu struct {
	Max        uint64  `knob:"max" json:"max,omitempty"`                             // Maximum CPU time in microseconds.
	Period     uint64  `knob:"period" default:"100000" json:"period,omitempty"`      // CPU time period in microseconds.
	Burst      uint64  `knob:"burst" json:"burst,omitempty"`                         // Maximum CPU burst time in microseconds.
	WeightNice int16   `knob:"weight.nice" default:"19" json:"weightNice,omitempty"` // Nice value for CPU weight adjustment (19 = lowest priority).
	Weight     uint16  `knob:"weight" default:"1" json:"weight,omitempty"`           // CPU scheduling weight.
	Idle       bool    `knob:"idle" default:"true" json:"idle,omitempty"`            // Whether the CPU is idle.
	UclampMin  float64 `knob:"uclamp.min" json:"uclampMin,omitempty"`                // Minimum CPU utilization clamp.
	UclampMax  float64 `knob:"uclamp.max" json:"uclampMax,omitempty"`                // Maximum CPU utilization clamp.
}

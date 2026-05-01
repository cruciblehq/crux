package cgroup

import (
	"strings"

	"github.com/cruciblehq/crux/internal/crex"
)

// RDMA resource limit for one HCA device.
//
// The rdma controller bounds the number of HCA handles and protected
// objects a cgroup may pin on a Host Channel Adapter, preventing one cgroup
// from exhausting per-device verb resources shared with the host.
type rdma struct {
	Device    string `knob:"device" json:"device,omitempty"`       // Device name.
	HcaHandle uint32 `knob:"hcaHandle" json:"hcaHandle,omitempty"` // HCA handle.
	HcaObject uint32 `knob:"hcaObject" json:"hcaObject,omitempty"` // HCA object.
}

// Parses an rdma entry.
//
// Expects an HCA device name followed by zero or more space-separated
// key=value pairs drawn from hca_handle and hca_object. The device name is
// required and identifies the entry; omitted limits remain at zero and are
// treated as "not set" when reconciling against earlier overrides for the
// same device.
func parseRDMA(value string) (rdma, error) {
	device, rest, _ := strings.Cut(strings.TrimSpace(value), " ")
	if device == "" {
		return rdma{}, crex.Wrapf(ErrInvalidGrant, "device name required")
	}
	r := rdma{Device: device}
	if rest != "" {
		if err := parseArgs(rest, map[string]func(string) error{
			"hca_handle": func(v string) error { return parseUint32(&r.HcaHandle, v) },
			"hca_object": func(v string) error { return parseUint32(&r.HcaObject, v) },
		}); err != nil {
			return rdma{}, err
		}
	}
	return r, nil
}

// Whether e and other address the same RDMA device.
func (e rdma) equal(other rdma) bool {
	return e.Device == other.Device
}

// Returns an error if e and other share identity with conflicting values.
func (e rdma) check(other rdma) error {
	if !e.equal(other) || e == other {
		return nil
	}
	return crex.Wrapf(ErrConflict, "%s %s already set", "rdma", other.Device)
}

// Leaves e unchanged and always reports no change.
//
// Same-identity entries with conflicting values are rejected upstream by
// check, and identical entries need no merge, so there is nothing to do.
func (e *rdma) merge(other rdma) bool {
	return false
}

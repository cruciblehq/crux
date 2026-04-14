package subsystem

import (
	"context"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/manifest"
)

// Implements [Subsystem] for [DomainDevice].
//
// Accumulates device provisioning grants keyed by device name. Identity fields
// (path, type, major, minor, uid, gid) and special mode bits (setuid, setgid,
// and sticky) must match across grants for the same name. The lower permission
// bits are relaxed via bitwise OR.
type DeviceSubsystem struct {
	devices map[string]*device
}

// Feeds a device grant into the subsystem.
//
// Validates the device name and properties from the expression and args.
// Returns an error wrapping [ErrGrantConflict] if identity fields or special
// mode bits differ for the same device name.
func (s *DeviceSubsystem) Build(_ context.Context, domain Domain, input manifest.Grant) ([]manifest.Grant, error) {
	crex.Assertf(domain == DomainDevice, "unexpected device domain %q", domain)

	d, err := parseDeviceExpr(input.Expr, input.Args)
	if err != nil {
		return nil, err
	}

	applied, err := s.merge(&d)
	if err != nil {
		return nil, err
	}
	if !applied {
		return nil, nil
	}

	return []manifest.Grant{{Subsystem: string(domain), Expr: input.Expr, Args: input.Args}}, nil
}

// Merges a parsed [device] into the accumulated model.
//
// If the name is new, the entry is stored directly. If the name already
// exists, identity fields and special mode bits must match; the lower 9
// permission bits are relaxed via OR. Returns true if the effective model
// changed.
func (s *DeviceSubsystem) merge(d *device) (bool, error) {
	if s.devices == nil {
		s.devices = make(map[string]*device)
	}

	existing, ok := s.devices[d.Name]
	if !ok {
		s.devices[d.Name] = d
		return true, nil
	}

	return existing.merge(d)
}

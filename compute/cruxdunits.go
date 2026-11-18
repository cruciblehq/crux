package compute

import (
	"context"
	"io"
	"net/http"

	speccompute "github.com/cruciblehq/spec/compute"
	"github.com/cruciblehq/utils-go/crex"
)

// Make sure the HTTP client satisfies the spec.
var _ speccompute.Units = cruxdUnits{}

// Units on a machine, reached through cruxd over HTTPS.
type cruxdUnits struct {
	runtime *cruxdRuntime // Connection to cruxd.
}

// Returns the units kept running on the machine.
func (r *cruxdRuntime) Units() speccompute.Units {
	return cruxdUnits{runtime: r}
}

// Makes the unit run as defined and returns its status.
func (u cruxdUnits) Apply(ctx context.Context, unit speccompute.Unit) (speccompute.UnitStatus, error) {
	var st speccompute.UnitStatus
	if err := u.runtime.call(ctx, speccompute.EndpointApplyUnit, speccompute.MediaTypeUnit, unit, speccompute.MediaTypeUnitStatus, &st); err != nil {
		return speccompute.UnitStatus{}, err
	}
	return st, nil
}

// Stops the unit and removes it with its filesystem and logs.
func (u cruxdUnits) Remove(ctx context.Context, id speccompute.UnitID) error {
	return u.runtime.call(ctx, speccompute.EndpointRemoveUnit, speccompute.MediaTypeRemoveUnitRequest, speccompute.RemoveUnitRequest{ID: id}, "", nil)
}

// Returns the status of every unit, ordered by identifier.
func (u cruxdUnits) List(ctx context.Context) ([]speccompute.UnitStatus, error) {
	resp, err := u.runtime.request(ctx, http.MethodGet, speccompute.EndpointListUnits, nil, "", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out speccompute.ListUnitsResponse
	if err := decodeResponse(resp, speccompute.MediaTypeListUnitsResponse, &out); err != nil {
		return nil, err
	}
	return out.Units, nil
}

// Returns the status of the unit.
func (u cruxdUnits) Inspect(ctx context.Context, id speccompute.UnitID) (speccompute.UnitStatus, error) {
	var st speccompute.UnitStatus
	if err := u.runtime.call(ctx, speccompute.EndpointInspectUnit, speccompute.MediaTypeInspectUnitRequest, speccompute.InspectUnitRequest{ID: id}, speccompute.MediaTypeUnitStatus, &st); err != nil {
		return speccompute.UnitStatus{}, err
	}
	return st, nil
}

// Writes the unit's output to w, and keeps writing new output when follow is set.
//
// Following ends without an error when ctx ends.
func (u cruxdUnits) Logs(ctx context.Context, id speccompute.UnitID, follow bool, w io.Writer) error {
	resp, err := u.runtime.send(ctx, speccompute.EndpointUnitLogs, speccompute.MediaTypeUnitLogsRequest, speccompute.UnitLogsRequest{ID: id, Follow: follow})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := checkResponse(resp, speccompute.MediaTypeUnitLogs.String()); err != nil {
		return err
	}
	if _, err := io.Copy(w, resp.Body); err != nil && ctx.Err() == nil {
		return crex.Wrap(ErrCruxd, err)
	}
	return nil
}

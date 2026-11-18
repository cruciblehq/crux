package compute

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	speccompute "github.com/cruciblehq/spec/compute"
)

func TestUnitsApply(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointApplyUnit, func(w http.ResponseWriter, r *http.Request) {
		var u speccompute.Unit
		readJSON(t, r, &u)
		if u.ID != "web" || u.Image != "img:1" {
			t.Errorf("unit = %+v, want web from img:1", u)
		}
		writeJSON(t, w, speccompute.MediaTypeUnitStatus, speccompute.UnitStatus{Unit: u, State: speccompute.UnitStarting})
	})
	units := serveCruxd(t, mux).Units()

	st, err := units.Apply(context.Background(), speccompute.Unit{ID: "web", Image: "img:1", Restart: speccompute.RestartAlways})
	if err != nil {
		t.Fatalf("Apply() = %v, want nil", err)
	}
	if st.State != speccompute.UnitStarting {
		t.Errorf("state = %q, want %q", st.State, speccompute.UnitStarting)
	}
}

func TestUnitsRemoveNotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointRemoveUnit, func(w http.ResponseWriter, r *http.Request) {
		writeError(t, w, http.StatusNotFound, speccompute.ErrorCodeNotFound, "unit not found")
	})
	units := serveCruxd(t, mux).Units()

	err := units.Remove(context.Background(), "web")
	var opErr *speccompute.Error
	if !errors.As(err, &opErr) || opErr.Code != speccompute.ErrorCodeNotFound {
		t.Fatalf("Remove() = %v, want a not_found spec error", err)
	}
}

func TestUnitsList(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointListUnits, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		writeJSON(t, w, speccompute.MediaTypeListUnitsResponse, speccompute.ListUnitsResponse{Units: []speccompute.UnitStatus{
			{Unit: speccompute.Unit{ID: "api"}, State: speccompute.UnitRunning},
			{Unit: speccompute.Unit{ID: "web"}, State: speccompute.UnitExited},
		}})
	})
	units := serveCruxd(t, mux).Units()

	statuses, err := units.List(context.Background())
	if err != nil {
		t.Fatalf("List() = %v, want nil", err)
	}
	if len(statuses) != 2 || statuses[1].State != speccompute.UnitExited {
		t.Errorf("statuses = %+v, want api running and web exited", statuses)
	}
}

func TestUnitsInspect(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointInspectUnit, func(w http.ResponseWriter, r *http.Request) {
		var req speccompute.InspectUnitRequest
		readJSON(t, r, &req)
		writeJSON(t, w, speccompute.MediaTypeUnitStatus, speccompute.UnitStatus{Unit: speccompute.Unit{ID: req.ID}, State: speccompute.UnitRunning, Restarts: 2})
	})
	units := serveCruxd(t, mux).Units()

	st, err := units.Inspect(context.Background(), "web")
	if err != nil {
		t.Fatalf("Inspect() = %v, want nil", err)
	}
	if st.Unit.ID != "web" || st.Restarts != 2 {
		t.Errorf("status = %+v, want web with 2 restarts", st)
	}
}

func TestUnitsLogs(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointUnitLogs, func(w http.ResponseWriter, r *http.Request) {
		var req speccompute.UnitLogsRequest
		readJSON(t, r, &req)
		if !req.Follow {
			t.Error("follow = false, want true")
		}
		w.Header().Set("Content-Type", speccompute.MediaTypeUnitLogs.String())
		_, _ = io.WriteString(w, "line one\nline two\n")
	})
	units := serveCruxd(t, mux).Units()

	var out bytes.Buffer
	if err := units.Logs(context.Background(), "web", true, &out); err != nil {
		t.Fatalf("Logs() = %v, want nil", err)
	}
	if out.String() != "line one\nline two\n" {
		t.Errorf("logs = %q, want both lines", out.String())
	}
}

func TestUnitsLogsNotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointUnitLogs, func(w http.ResponseWriter, r *http.Request) {
		writeError(t, w, http.StatusNotFound, speccompute.ErrorCodeNotFound, "unit not found")
	})
	units := serveCruxd(t, mux).Units()

	err := units.Logs(context.Background(), "web", false, &bytes.Buffer{})
	var opErr *speccompute.Error
	if !errors.As(err, &opErr) || opErr.Code != speccompute.ErrorCodeNotFound {
		t.Fatalf("Logs() = %v, want a not_found spec error", err)
	}
}

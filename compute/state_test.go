package compute

import "testing"

func TestStateString(t *testing.T) {
	cases := map[State]string{
		StateNotProvisioned: "not provisioned",
		StateRunning:        "running",
		StateStopped:        "stopped",
		State(99):           "unknown",
	}
	for state, want := range cases {
		if got := state.String(); got != want {
			t.Fatalf("State(%d).String() = %q, want %q", state, got, want)
		}
	}
}

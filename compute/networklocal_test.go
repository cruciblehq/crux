package compute

import (
	"net"
	"strings"
	"testing"

	"github.com/cruciblehq/crux/compute/local"
)

func TestBridgeName(t *testing.T) {
	if got, want := bridgeName("net-1234567890"), "br-1234567890"; got != want {
		t.Fatalf("bridgeName() = %q, want %q", got, want)
	}

	longID := strings.Repeat("x", 100)
	if got, want := bridgeName("net-"+longID), "br-"+strings.Repeat("x", local.InterfaceNameMax-3); got != want {
		t.Fatalf("bridgeName() = %q, want %q", got, want)
	}
}

func TestNextHost(t *testing.T) {
	ip := net.ParseIP("10.0.0.0")
	got := nextHost(ip)
	want := net.ParseIP("10.0.0.1")
	if !got.Equal(want) {
		t.Fatalf("nextHost() = %v, want %v", got, want)
	}
}

func TestTruncate(t *testing.T) {
	if got, want := truncate("abcdef", 4), "abcd"; got != want {
		t.Fatalf("truncate() = %q, want %q", got, want)
	}
	if got, want := truncate("short", 32), "short"; got != want {
		t.Fatalf("truncate() = %q, want %q", got, want)
	}
}

func TestBuildNftRule(t *testing.T) {
	rule := Rule{
		Direction: Ingress,
		Protocol:  "tcp",
		Port:      8080,
		Peer:      "10.0.0.2",
	}
	got := buildNftRule(rule)
	want := []string{"nft", "add", "rule", "inet", "filter", "input", "meta", "l4proto", "tcp", "th", "dport", "8080", "ip", "saddr", "10.0.0.2", "accept"}
	if len(got) != len(want) {
		t.Fatalf("buildNftRule() len = %d, want %d; got %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("buildNftRule()[%d] = %q, want %q; full = %v", i, got[i], want[i], got)
		}
	}

	if args := buildNftRule(Rule{Direction: Direction(99)}); args != nil {
		t.Fatalf("buildNftRule() for unknown direction = %v, want nil", args)
	}
}

func TestLocalStateToState(t *testing.T) {
	if got, want := localStateToState(local.StateRunning), StateRunning; got != want {
		t.Fatalf("localStateToState() = %v, want %v", got, want)
	}
	if got, want := localStateToState(local.StateStopped), StateStopped; got != want {
		t.Fatalf("localStateToState() = %v, want %v", got, want)
	}
	if got, want := localStateToState(local.StateNotProvisioned), StateNotProvisioned; got != want {
		t.Fatalf("localStateToState() = %v, want %v", got, want)
	}
}

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

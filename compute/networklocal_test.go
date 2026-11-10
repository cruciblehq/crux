package compute

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"slices"
	"strings"
	"testing"

	"github.com/cruciblehq/crux/compute/local"
)

type fakeExec struct {
	calls  []string
	output map[string]string
	fail   string
}

func (f *fakeExec) run(_ context.Context, stdout, _ io.Writer, command string, args ...string) (int, error) {
	line := strings.Join(append([]string{command}, args...), " ")
	f.calls = append(f.calls, line)
	if f.fail != "" && strings.HasPrefix(line, f.fail) {
		return 1, nil
	}
	if out, ok := f.output[line]; ok {
		_, _ = io.WriteString(stdout, out)
	}
	return 0, nil
}

func (f *fakeExec) ran(prefix string) bool {
	return slices.ContainsFunc(f.calls, func(line string) bool {
		return strings.HasPrefix(line, prefix)
	})
}

func TestBridgeName(t *testing.T) {
	if got, want := bridgeName("net-1234567890"), "br-1234567890"; got != want {
		t.Fatalf("bridgeName() = %q, want %q", got, want)
	}

	longID := strings.Repeat("x", 100)
	if got, want := bridgeName(fmt.Sprintf("net-%s", longID)), fmt.Sprintf("br-%s", strings.Repeat("x", local.InterfaceNameMax-3)); got != want {
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

func TestNetworkCreate(t *testing.T) {
	fx := &fakeExec{}
	id, err := newNetworkLocal(fx.run).Create(context.Background(), NetworkOptions{Subnet: "10.1.0.0/24"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !strings.HasPrefix(id, local.NetworkIDPrefix) {
		t.Fatalf("Create() id = %q, want prefix %q", id, local.NetworkIDPrefix)
	}
	bridge := bridgeName(id)
	if !fx.ran(fmt.Sprintf("ip link add %s type bridge", bridge)) {
		t.Errorf("bridge %q was not created; calls = %v", bridge, fx.calls)
	}
	if !fx.ran(fmt.Sprintf("ip addr add 10.1.0.1/24 dev %s", bridge)) {
		t.Errorf("gateway was not assigned; calls = %v", fx.calls)
	}
}

func TestNetworkCreateMissingSubnet(t *testing.T) {
	fx := &fakeExec{}
	_, err := newNetworkLocal(fx.run).Create(context.Background(), NetworkOptions{})
	if !errors.Is(err, ErrNetwork) {
		t.Fatalf("Create() error = %v, want ErrNetwork", err)
	}
	if len(fx.calls) != 0 {
		t.Errorf("calls = %v, want none", fx.calls)
	}
}

func TestNetworkCreateInvalidSubnet(t *testing.T) {
	_, err := newNetworkLocal((&fakeExec{}).run).Create(context.Background(), NetworkOptions{Subnet: "not-a-cidr"})
	if !errors.Is(err, ErrNetwork) {
		t.Fatalf("Create() error = %v, want ErrNetwork", err)
	}
}

func TestNetworkCreateCommandFails(t *testing.T) {
	fx := &fakeExec{fail: "ip link add"}
	_, err := newNetworkLocal(fx.run).Create(context.Background(), NetworkOptions{Subnet: "10.1.0.0/24"})
	if !errors.Is(err, ErrNetwork) {
		t.Fatalf("Create() error = %v, want ErrNetwork", err)
	}
}

func TestNetworkRemove(t *testing.T) {
	fx := &fakeExec{}
	if err := newNetworkLocal(fx.run).Remove(context.Background(), "net-abc"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if !fx.ran("ip link delete br-abc") {
		t.Errorf("bridge was not deleted; calls = %v", fx.calls)
	}
}

func TestNetworkRemoveFails(t *testing.T) {
	fx := &fakeExec{fail: "ip link delete"}
	if err := newNetworkLocal(fx.run).Remove(context.Background(), "net-abc"); !errors.Is(err, ErrNetwork) {
		t.Fatalf("Remove() error = %v, want ErrNetwork", err)
	}
}

func TestNetworkAttach(t *testing.T) {
	fx := &fakeExec{}
	if err := newNetworkLocal(fx.run).Attach(context.Background(), "unit1", "net-abc"); err != nil {
		t.Fatalf("Attach() error = %v", err)
	}
	if !fx.ran("ip netns add unit1") {
		t.Errorf("namespace was not created; calls = %v", fx.calls)
	}
	if !fx.ran(fmt.Sprintf("ip link set %s", local.VethPrefix)) {
		t.Errorf("veth was not configured; calls = %v", fx.calls)
	}
}

func TestNetworkAttachFails(t *testing.T) {
	fx := &fakeExec{fail: "ip link add"}
	if err := newNetworkLocal(fx.run).Attach(context.Background(), "unit1", "net-abc"); !errors.Is(err, ErrNetwork) {
		t.Fatalf("Attach() error = %v, want ErrNetwork", err)
	}
}

func TestNetworkDetach(t *testing.T) {
	fx := &fakeExec{output: map[string]string{
		"ip -o link show master br-abc": "5: ve-1234@if4: <UP>\n6: other0: <UP>\n",
	}}
	if err := newNetworkLocal(fx.run).Detach(context.Background(), "unit1", "net-abc"); err != nil {
		t.Fatalf("Detach() error = %v", err)
	}
	if !fx.ran("ip link delete ve-1234@if4") {
		t.Errorf("veth was not deleted; calls = %v", fx.calls)
	}
	if fx.ran("ip link delete other0") {
		t.Errorf("interface without the veth prefix was deleted; calls = %v", fx.calls)
	}
}

func TestNetworkDetachFails(t *testing.T) {
	fx := &fakeExec{fail: "ip -o link show"}
	if err := newNetworkLocal(fx.run).Detach(context.Background(), "unit1", "net-abc"); !errors.Is(err, ErrNetwork) {
		t.Fatalf("Detach() error = %v, want ErrNetwork", err)
	}
}

func TestNetworkApply(t *testing.T) {
	fx := &fakeExec{}
	rules := []Rule{
		{Direction: Ingress, Protocol: "tcp", Port: 8080, Peer: "10.0.0.2"},
		{Direction: Direction(99)},
	}
	if err := newNetworkLocal(fx.run).Apply(context.Background(), "unit1", rules); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if !fx.ran("ip netns exec unit1 nft flush chain inet filter input") {
		t.Errorf("input chain was not flushed; calls = %v", fx.calls)
	}
	if !fx.ran("ip netns exec unit1 nft add rule inet filter input meta l4proto tcp th dport 8080") {
		t.Errorf("rule was not applied; calls = %v", fx.calls)
	}
}

func TestNetworkList(t *testing.T) {
	fx := &fakeExec{output: map[string]string{
		"ip -o link show type bridge":   "3: br-abc: <UP>\n4: docker0: <UP>\n",
		"ip -o -4 addr show dev br-abc": "3: br-abc    inet 10.1.0.1/24 scope global br-abc\n",
	}}
	got, err := newNetworkLocal(fx.run).List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	want := []NetworkInfo{{ID: "net-abc", Subnet: "10.1.0.1/24"}}
	if !slices.Equal(got, want) {
		t.Fatalf("List() = %v, want %v", got, want)
	}
}

func TestNetworkListFails(t *testing.T) {
	fx := &fakeExec{fail: "ip -o link show"}
	if _, err := newNetworkLocal(fx.run).List(context.Background()); !errors.Is(err, ErrNetwork) {
		t.Fatalf("List() error = %v, want ErrNetwork", err)
	}
}

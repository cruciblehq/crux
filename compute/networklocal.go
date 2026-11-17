package compute

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"strings"

	"github.com/cruciblehq/crux/compute/local"
	"github.com/cruciblehq/utils-go/crex"
)

// Shim that adapts the local network implementation to the [Network] interface.
//
// Translates Network method calls into the local package's bridge, namespace,
// and nftables functions. The shim is stateless; the executor records all
// metadata in the deployment state file.
type networkLocalShim struct {
	exec local.ExecFunc // Function to execute commands on the local host.
}

// Returns a [Network] backed by the local provider.
func newNetworkLocal(exec local.ExecFunc) Network {
	return &networkLocalShim{exec: exec}
}

// Creates a new Linux bridge on the VM.
func (s *networkLocalShim) Create(ctx context.Context, opts NetworkOptions) (string, error) {
	id := fmt.Sprintf("%s%s", local.NetworkIDPrefix, randHex(idLen))
	bridge := bridgeName(id)

	cidr := opts.Subnet
	if cidr == "" {
		return "", crex.Newf(ErrNetwork, "subnet is required")
	}

	_, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", crex.Newf(ErrNetwork, "invalid subnet %q", cidr)
	}

	gateway := nextHost(subnet.IP)
	bits, _ := subnet.Mask.Size()

	if err := local.CreateBridge(ctx, s.exec, bridge, gateway, bits); err != nil {
		return "", crex.Wrap(ErrNetwork, err)
	}

	return id, nil
}

// Removes a Linux bridge from the VM.
func (s *networkLocalShim) Remove(ctx context.Context, id string) error {
	if err := local.RemoveBridge(ctx, s.exec, bridgeName(id)); err != nil {
		return crex.Wrap(ErrNetwork, err)
	}
	return nil
}

// Attaches a unit to a network.
func (s *networkLocalShim) Attach(ctx context.Context, unitID, networkID string) error {
	bridge := bridgeName(networkID)
	veth := truncate(fmt.Sprintf("%s%s", local.VethPrefix, randHex(local.VethRandLen)), local.InterfaceNameMax)
	peer := truncate(fmt.Sprintf("%s%s", local.PeerPrefix, randHex(local.VethRandLen)), local.InterfaceNameMax)

	if err := local.AttachUnit(ctx, s.exec, unitID, bridge, veth, peer); err != nil {
		return crex.Wrap(ErrNetwork, err)
	}
	local.InstallBaseline(ctx, s.exec, unitID)
	return nil
}

// Detaches a unit from a network.
func (s *networkLocalShim) Detach(ctx context.Context, unitID, networkID string) error {
	bridge := bridgeName(networkID)
	ifaces, err := local.ListBridgeInterfaces(ctx, s.exec, bridge)
	if err != nil {
		return crex.Wrap(ErrNetwork, err)
	}
	for _, ifname := range ifaces {
		if strings.HasPrefix(ifname, local.VethPrefix) {
			local.DetachUnit(ctx, s.exec, ifname)
		}
	}
	return nil
}

// Replaces the firewall rules for a unit.
func (s *networkLocalShim) Apply(ctx context.Context, unitID string, rules []Rule) error {
	var nftRules [][]string
	for _, r := range rules {
		switch r.Direction {
		case Ingress:
			nftRules = append(nftRules, local.InputRule(r.Protocol, r.Port, r.Peer))
		case Egress:
			nftRules = append(nftRules, local.OutputRule(r.Protocol, r.Port, r.Peer))
		}
	}
	local.ApplyRules(ctx, s.exec, unitID, nftRules)
	return nil
}

// Lists bridges on the VM matching the Crucible naming convention.
func (s *networkLocalShim) List(ctx context.Context) ([]NetworkInfo, error) {
	bridges, err := local.ListBridges(ctx, s.exec, local.BridgePrefix)
	if err != nil {
		return nil, crex.Wrap(ErrNetwork, err)
	}

	var result []NetworkInfo
	for _, name := range bridges {
		id := fmt.Sprintf("%s%s", local.NetworkIDPrefix, strings.TrimPrefix(name, local.BridgePrefix))
		info := NetworkInfo{ID: id}
		if addr, err := local.InterfaceAddress(ctx, s.exec, name); err == nil {
			info.Subnet = addr
		}
		result = append(result, info)
	}
	return result, nil
}

// Derives the bridge interface name from a network ID.
func bridgeName(networkID string) string {
	return truncate(fmt.Sprintf("%s%s", local.BridgePrefix, strings.TrimPrefix(networkID, local.NetworkIDPrefix)), local.InterfaceNameMax)
}

// Returns the first host address in a subnet.
func nextHost(ip net.IP) net.IP {
	host := make(net.IP, len(ip))
	copy(host, ip)
	host[len(host)-1]++
	return host
}

// Truncates s to at most max bytes.
func truncate(s string, maxCh int) string {
	if len(s) > maxCh {
		return s[:maxCh]
	}
	return s
}

// Returns n random bytes as a lowercase hex string.
//
// crypto/rand.Read never fails, so there is no error to return.
func randHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

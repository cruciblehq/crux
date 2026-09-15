package compute

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/cruciblehq/crux/compute/local"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/crypto"
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
	id := fmt.Sprintf("%s%s", local.NetworkIDPrefix, crypto.RandHex(idLen))
	bridge := bridgeName(id)

	cidr := opts.Subnet
	if cidr == "" {
		return "", crex.New("subnet is required")
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
	veth := truncate(fmt.Sprintf("%s%s", local.VethPrefix, crypto.RandHex(local.VethRandLen)), local.InterfaceNameMax)
	peer := truncate(fmt.Sprintf("%s%s", local.PeerPrefix, crypto.RandHex(local.VethRandLen)), local.InterfaceNameMax)

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
		if args := buildNftRule(r); args != nil {
			nftRules = append(nftRules, args)
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
		id := local.NetworkIDPrefix + strings.TrimPrefix(name, local.BridgePrefix)
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
	return truncate(local.BridgePrefix+strings.TrimPrefix(networkID, local.NetworkIDPrefix), local.InterfaceNameMax)
}

// Translates a [Rule] into an nft command argument list.
func buildNftRule(r Rule) []string {
	var chain string
	switch r.Direction {
	case Ingress:
		chain = "input"
	case Egress:
		chain = "output"
	default:
		return nil
	}

	args := []string{"nft", "add", "rule", "inet", "filter", chain}

	proto := strings.ToLower(r.Protocol)
	switch proto {
	case "tcp", "udp", "sctp", "dccp":
		args = append(args, "meta", "l4proto", proto)
		if r.Port > 0 {
			args = append(args, "th", "dport", fmt.Sprintf("%d", r.Port))
		}
	case "icmp", "icmpv6":
		args = append(args, "meta", "l4proto", proto)
	default:
		args = append(args, "meta", "l4proto", proto)
	}

	if r.Peer != "" {
		switch r.Direction {
		case Ingress:
			args = append(args, "ip", "saddr", r.Peer)
		case Egress:
			args = append(args, "ip", "daddr", r.Peer)
		}
	}

	args = append(args, "accept")
	return args
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

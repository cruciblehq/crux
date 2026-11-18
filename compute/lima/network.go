package lima

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/cruciblehq/spec/manifest"
)

// Creates a Linux bridge on the host with a deny-all forward chain.
//
// bridge is the interface name (must be at most [InterfaceNameMax] characters).
// gateway is the first host address in the subnet, assigned to the bridge
// interface. bits is the subnet mask length.
func CreateBridge(ctx context.Context, exec ExecFunc, bridge string, gateway net.IP, bits int) error {
	if err := run(ctx, exec, ipCommand, linkSubcommand, "add", bridge, "type", nftFamilyBridge); err != nil {
		return err
	}
	if err := run(ctx, exec, ipCommand, addrSubcommand, "add", fmt.Sprintf("%s/%d", gateway, bits), "dev", bridge); err != nil {
		run(ctx, exec, ipCommand, linkSubcommand, "delete", bridge)
		return err
	}
	if err := run(ctx, exec, ipCommand, linkSubcommand, "set", bridge, "up"); err != nil {
		run(ctx, exec, ipCommand, linkSubcommand, "delete", bridge)
		return err
	}
	if err := run(ctx, exec, nftCommand, "add", "table", nftFamilyBridge, bridge); err != nil {
		run(ctx, exec, ipCommand, linkSubcommand, "delete", bridge)
		return err
	}
	if err := run(ctx, exec, nftCommand, "add", "chain", nftFamilyBridge, bridge, nftChainForward, "{ type filter hook forward priority 0; policy drop; }"); err != nil {
		run(ctx, exec, nftCommand, "delete", "table", nftFamilyBridge, bridge)
		run(ctx, exec, ipCommand, linkSubcommand, "delete", bridge)
		return err
	}
	return nil
}

// Removes a Linux bridge and its nftables table from the host.
func RemoveBridge(ctx context.Context, exec ExecFunc, bridge string) error {
	run(ctx, exec, nftCommand, "delete", "table", nftFamilyBridge, bridge)
	return run(ctx, exec, ipCommand, linkSubcommand, "delete", bridge)
}

// Attaches a unit to a bridge by creating a veth pair and network namespace.
//
// Creates a named network namespace for the unit (if one does not already
// exist), creates a veth pair connecting the namespace to the bridge, and
// brings up all interfaces.
func AttachUnit(ctx context.Context, exec ExecFunc, unitID, bridge, veth, peer string) error {
	run(ctx, exec, ipCommand, netnsSubcommand, "add", unitID)

	if err := run(ctx, exec, ipCommand, linkSubcommand, "add", veth, "type", "veth", "peer", "name", peer); err != nil {
		return err
	}
	if err := run(ctx, exec, ipCommand, linkSubcommand, "set", peer, netnsSubcommand, unitID); err != nil {
		run(ctx, exec, ipCommand, linkSubcommand, "delete", veth)
		return err
	}
	if err := run(ctx, exec, ipCommand, linkSubcommand, "set", veth, "master", bridge); err != nil {
		nsRun(ctx, exec, unitID, ipCommand, linkSubcommand, "delete", peer)
		return err
	}
	if err := run(ctx, exec, ipCommand, linkSubcommand, "set", veth, "up"); err != nil {
		run(ctx, exec, ipCommand, linkSubcommand, "delete", veth)
		return err
	}

	nsRun(ctx, exec, unitID, ipCommand, linkSubcommand, "set", peer, "up")
	nsRun(ctx, exec, unitID, ipCommand, linkSubcommand, "set", loopback, "up")

	return nil
}

// Installs a deny-all nftables baseline inside a unit's network namespace.
//
// Creates an inet filter table with input and output chains set to policy drop.
// Loopback traffic is allowed.
func InstallBaseline(ctx context.Context, exec ExecFunc, unitID string) {
	nsRun(ctx, exec, unitID, nftCommand, "add", "table", nftFamilyInet, nftTableFilter)
	nsRun(ctx, exec, unitID, nftCommand, "add", "chain", nftFamilyInet, nftTableFilter, nftChainInput, "{ type filter hook input priority 0; policy drop; }")
	nsRun(ctx, exec, unitID, nftCommand, "add", "chain", nftFamilyInet, nftTableFilter, nftChainOutput, "{ type filter hook output priority 0; policy drop; }")
	nsRun(ctx, exec, unitID, nftCommand, "add", "rule", nftFamilyInet, nftTableFilter, nftChainInput, "iif", loopback, "accept")
	nsRun(ctx, exec, unitID, nftCommand, "add", "rule", nftFamilyInet, nftTableFilter, nftChainOutput, "oif", loopback, "accept")
}

// Detaches a unit from a bridge by removing the host-side veth.
//
// Removing the host-side veth also removes its peer inside the namespace.
func DetachUnit(ctx context.Context, exec ExecFunc, veth string) error {
	return run(ctx, exec, ipCommand, linkSubcommand, "delete", veth)
}

// Replaces all nftables rules inside a unit's network namespace.
//
// Flushes the input and output chains, re-adds loopback rules, then applies
// the provided nft argument lists. Each entry in rules is a complete nft
// command argument list (e.g. ["nft", "add", "rule", "inet", "filter", ...]).
func ApplyRules(ctx context.Context, exec ExecFunc, unitID string, rules [][]string) {
	nsRun(ctx, exec, unitID, nftCommand, "flush", "chain", nftFamilyInet, nftTableFilter, nftChainInput)
	nsRun(ctx, exec, unitID, nftCommand, "flush", "chain", nftFamilyInet, nftTableFilter, nftChainOutput)

	nsRun(ctx, exec, unitID, nftCommand, "add", "rule", nftFamilyInet, nftTableFilter, nftChainInput, "iif", loopback, "accept")
	nsRun(ctx, exec, unitID, nftCommand, "add", "rule", nftFamilyInet, nftTableFilter, nftChainOutput, "oif", loopback, "accept")

	for _, args := range rules {
		cmd := append([]string{ipCommand, netnsSubcommand, execSubcommand, unitID}, args...)
		run(ctx, exec, cmd[0], cmd[1:]...)
	}
}

// Returns the nft arguments of a rule that accepts traffic entering a unit.
//
// proto is a protocol keyword or number from the manifest. The "ip" wildcard
// or an empty value matches any protocol. port is matched only when it is not
// zero and proto has ports. peer is the source address, matched only when it
// is not empty.
func InputRule(proto string, port int, peer string) []string {
	return nftAcceptRule(nftChainInput, "saddr", proto, port, peer)
}

// Returns the nft arguments of a rule that accepts traffic leaving a unit.
//
// Takes the same arguments as [InputRule], except that peer is the
// destination address.
func OutputRule(proto string, port int, peer string) []string {
	return nftAcceptRule(nftChainOutput, "daddr", proto, port, peer)
}

// Returns the nft arguments of an accept rule in chain.
//
// peerField is the address field that holds peer.
func nftAcceptRule(chain, peerField, proto string, port int, peer string) []string {
	args := []string{nftCommand, "add", "rule", nftFamilyInet, nftTableFilter, chain}
	proto = strings.ToLower(proto)
	if proto != "" && proto != manifest.ProtocolAny {
		args = append(args, "meta", "l4proto", proto)
		if port > 0 && hasPorts(proto) {
			args = append(args, "th", "dport", strconv.Itoa(port))
		}
	}
	if peer != "" {
		args = append(args, "ip", peerField, peer)
	}
	return append(args, "accept")
}

// Whether proto carries port numbers.
func hasPorts(proto string) bool {
	switch proto {
	case manifest.ProtocolTCP, manifest.ProtocolUDP, manifest.ProtocolSCTP:
		return true
	}
	return false
}

// Lists bridge interface names on the host that match the given prefix.
func ListBridges(ctx context.Context, exec ExecFunc, prefix string) ([]string, error) {
	out, err := execOutput(ctx, exec, ipCommand, "-o", linkSubcommand, "show", "type", nftFamilyBridge)
	if err != nil {
		return nil, err
	}
	var bridges []string
	for _, name := range parseLinkNames(out) {
		if strings.HasPrefix(name, prefix) {
			bridges = append(bridges, name)
		}
	}
	return bridges, nil
}

// Queries the IPv4 address assigned to an interface on the host.
func InterfaceAddress(ctx context.Context, exec ExecFunc, iface string) (string, error) {
	out, err := execOutput(ctx, exec, ipCommand, "-o", "-4", addrSubcommand, "show", "dev", iface)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		for i, f := range fields {
			if f == "inet" && i+1 < len(fields) {
				return fields[i+1], nil
			}
		}
	}
	return "", nil
}

// Lists host-side veth interfaces attached to a bridge.
func ListBridgeInterfaces(ctx context.Context, exec ExecFunc, bridge string) ([]string, error) {
	out, err := execOutput(ctx, exec, ipCommand, "-o", linkSubcommand, "show", "master", bridge)
	if err != nil {
		return nil, err
	}
	return parseLinkNames(out), nil
}

// Returns the interface names in the output of "ip -o link show".
//
// Each line looks like "3: name: <FLAGS> ...", so the name is the second field
// without its trailing colon.
func parseLinkNames(out string) []string {
	const nameField = 1
	var names []string
	for line := range strings.SplitSeq(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) <= nameField {
			continue
		}
		names = append(names, strings.TrimSuffix(fields[nameField], ":"))
	}
	return names
}

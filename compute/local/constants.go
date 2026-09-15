package local

import "os"

// Linux interface name constraints.
const (
	InterfaceNameMax = 15 // Maximum interface name length on Linux.
)

// Naming prefixes for provider-assigned identifiers.
const (
	NetworkIDPrefix = "net-" // Prefix for network identifiers.
	BridgePrefix    = "br-"  // Prefix for bridge interface names.
	VethPrefix      = "ve-"  // Prefix for host-side veth interface names.
	PeerPrefix      = "vp-"  // Prefix for namespace-side veth interface names.
	VolumeIDPrefix  = "vol-" // Prefix for volume identifiers.
)

// Random identifier lengths in bytes.
const (
	VethRandLen = 4 // Random bytes in veth interface names.
)

// Filesystem defaults.
const (
	VolumesDir                    = "volumes" // Subdirectory name for volume storage.
	DefaultVolumeMode os.FileMode = 0o700     // Default permission mode for volume directories.
)

// Command names and subcommands used by the network implementation.
const (
	ipCommand  = "ip"  // iproute2 command.
	nftCommand = "nft" // nftables command.
	loopback   = "lo"  // Loopback interface name.

	// ip subcommands and keywords.
	netnsSubcommand = "netns"
	execSubcommand  = "exec"
	linkSubcommand  = "link"
	addrSubcommand  = "addr"
	routeSubcommand = "route"
)

// nftables table, chain, and family names.
const (
	nftFamilyBridge = "bridge"  // nftables bridge family for forward chains.
	nftFamilyInet   = "inet"    // nftables inet family for input/output chains.
	nftTableFilter  = "filter"  // nftables filter table name.
	nftChainForward = "forward" // nftables forward chain name.
	nftChainInput   = "input"   // nftables input chain name.
	nftChainOutput  = "output"  // nftables output chain name.
)

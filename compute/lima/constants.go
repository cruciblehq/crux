package lima

// Linux interface name constraints.
const (
	InterfaceNameMax = 15 // Maximum interface name length on Linux.
)

// Prefixes of the identifiers the provider assigns.
const (
	NetworkIDPrefix = "net-" // Prefix for network identifiers.
	BridgePrefix    = "br-"  // Prefix for bridge interface names.
	VethPrefix      = "ve-"  // Prefix for veth interface names in the VM's root namespace.
	PeerPrefix      = "vp-"  // Prefix for veth interface names inside a container's namespace.
	VolumeIDPrefix  = "vol-" // Prefix for volume identifiers.
)

// Random identifier lengths in bytes.
const (
	VethRandLen = 4 // Random bytes in veth interface names.
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

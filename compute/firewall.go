package compute

// Traffic direction for firewall rules.
type Direction int

const (
	Ingress Direction = iota // Inbound traffic to the instance.
	Egress                   // Outbound traffic from the instance.
)

// A single firewall allow rule.
//
// Rules are derived from the affordance model's .net grants. All rules are
// permits against a deny-all baseline; there is no action field. The executor
// resolves symbolic service names from the blueprint's connectivity graph into
// concrete addresses before passing rules to the provider.
type Rule struct {
	Direction Direction `codec:"direction"`      // Whether the rule applies to inbound or outbound traffic.
	Protocol  string    `codec:"protocol"`       // Transport protocol: tcp, udp, icmp, etc.
	Port      int       `codec:"port,omitempty"` // Port number; zero means protocol-level (no port).
	Peer      string    `codec:"peer,omitempty"` // Resolved address of the remote end.
}

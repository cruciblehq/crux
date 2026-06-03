package manifest

import (
	"errors"
	"testing"

	affcap "github.com/cruciblehq/crux/security/fcap"
	affmac "github.com/cruciblehq/crux/security/mac"
	afnet "github.com/cruciblehq/crux/security/net"
)

func TestContainerValidateEmpty(t *testing.T) {
	if err := (&Container{}).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestContainerValidateOK(t *testing.T) {
	c := &Container{
		Fcap: affcap.Fcap{Entries: map[string]*affcap.FcapCapabilities{
			"/bin/foo": {Permitted: []string{"cap_net_admin"}},
		}},
		MAC: affmac.MAC{Rules: []*affmac.MACAllow{{Hook: "file_open"}}},
		Network: afnet.NetworkPolicy{
			Ingress: []afnet.NetworkIngressRule{{Protocol: "tcp", Port: 8080}},
			Egress:  []afnet.NetworkEgressRule{{Protocol: "tcp", Destination: "api.example.com"}},
		},
	}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestContainerValidatePropagatesFcapError(t *testing.T) {
	c := &Container{
		Fcap: affcap.Fcap{Entries: map[string]*affcap.FcapCapabilities{
			"": {Permitted: []string{"cap_net_admin"}}, // empty path key
		}},
	}
	err := c.Validate()
	if !errors.Is(err, ErrInvalidContainer) {
		t.Fatalf("err = %v, want ErrInvalidContainer", err)
	}
}

func TestContainerValidatePropagatesMACError(t *testing.T) {
	c := &Container{
		MAC: affmac.MAC{Rules: []*affmac.MACAllow{{Hook: ""}}}, // empty hook
	}
	err := c.Validate()
	if !errors.Is(err, ErrInvalidContainer) {
		t.Fatalf("err = %v, want ErrInvalidContainer", err)
	}
}

func TestContainerValidatePropagatesNetworkError(t *testing.T) {
	c := &Container{
		Network: afnet.NetworkPolicy{
			Ingress: []afnet.NetworkIngressRule{{Protocol: "bogus", Port: 80}},
		},
	}
	err := c.Validate()
	if !errors.Is(err, ErrInvalidContainer) {
		t.Fatalf("err = %v, want ErrInvalidContainer", err)
	}
}

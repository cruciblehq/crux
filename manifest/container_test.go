package manifest

import (
	"errors"
	"testing"

	affcap "github.com/cruciblehq/crux/affordance/fcap"
	affmac "github.com/cruciblehq/crux/affordance/mac"
	afnet "github.com/cruciblehq/crux/affordance/net"
)

func TestContainerValidateEmpty(t *testing.T) {
	if err := (&Container{}).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestContainerValidateOK(t *testing.T) {
	c := &Container{
		Fcap: affcap.Spec{Entries: map[string]*affcap.Capabilities{
			"/bin/foo": {Permitted: []string{"cap_net_admin"}},
		}},
		MAC: affmac.Spec{Rules: []*affmac.MACAllow{{Hook: "file_open"}}},
		Network: afnet.Spec{
			Ingress: []afnet.IngressRule{{Protocol: "tcp", Port: 8080}},
			Egress:  []afnet.EgressRule{{Protocol: "tcp", Destination: "api.example.com"}},
		},
	}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestContainerValidatePropagatesFcapError(t *testing.T) {
	c := &Container{
		Fcap: affcap.Spec{Entries: map[string]*affcap.Capabilities{
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
		MAC: affmac.Spec{Rules: []*affmac.MACAllow{{Hook: ""}}}, // empty hook
	}
	err := c.Validate()
	if !errors.Is(err, ErrInvalidContainer) {
		t.Fatalf("err = %v, want ErrInvalidContainer", err)
	}
}

func TestContainerValidatePropagatesNetworkError(t *testing.T) {
	c := &Container{
		Network: afnet.Spec{
			Ingress: []afnet.IngressRule{{Protocol: "bogus", Port: 80}},
		},
	}
	err := c.Validate()
	if !errors.Is(err, ErrInvalidContainer) {
		t.Fatalf("err = %v, want ErrInvalidContainer", err)
	}
}

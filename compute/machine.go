package compute

import (
	"context"
	"crypto/x509"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/cruciblehq/crux/authority"
	"github.com/cruciblehq/crux/dirs"
	"github.com/cruciblehq/crux/machines"
	"github.com/cruciblehq/spec/agent"
	"github.com/cruciblehq/spec/trust"
	"github.com/cruciblehq/utils-go/codec"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/retry"
)

// Machine readiness timing.
const (
	machineReadyTimeout = 15 * time.Minute // Longest wait for a machine to enroll and report healthy.
	machinePollInterval = 2 * time.Second  // Wait between readiness attempts.
	machineTryTimeout   = 5 * time.Second  // Longest wait for one readiness attempt.
)

// Opens crux's certificate authority, with its private keys in the secrets
// directory.
func openAuthority() (*authority.Authority, error) {
	return authority.Open(dirs.Trust(), dirs.Secrets())
}

// Creates the record for a new machine and returns its enrollment data.
//
// The record keeps the machine identifier and token so crux can enroll the
// machine once it boots. The provider places the returned JSON where the
// agent on the machine reads it.
func newMachine(name string) ([]byte, error) {
	auth, err := openAuthority()
	if err != nil {
		return nil, err
	}
	rec := machines.New()
	if err := machines.Save(dirs.Machines(), name, rec); err != nil {
		return nil, err
	}
	return codec.Encode(trust.Enrollment{
		Machine:     rec.ID,
		Root:        auth.Root().Raw,
		Revocations: auth.Revocations().Raw,
		Token:       rec.Token,
	}, codec.JSON)
}

// Connects to the agent on the enrolled machine name at addr over mutual TLS.
//
// crux accepts only the agent certificate the root signed for this machine,
// and sends its revocation list in the handshake.
func connectMachine(ctx context.Context, name, addr string) (agent.Agent, error) {
	auth, err := openAuthority()
	if err != nil {
		return nil, err
	}
	rec, err := machines.Load(dirs.Machines(), name)
	if err != nil {
		return nil, err
	}
	if !rec.Enrolled() {
		return nil, crex.Newf(ErrNotEnrolled, "machine %q has not enrolled", name)
	}
	cert, err := auth.Client()
	if err != nil {
		return nil, err
	}
	cfg := trust.ClientConfig(cert, auth.Root(), auth.Revocations(), rec.ID)
	return Dial(ctx, addr, cfg, auth.Revocations().Raw)
}

// Enrolls the machine name at addr.
//
// The agent proves it holds the token from the machine's record, crux signs
// the agent's certificate and installs it, and the token is then dropped from
// the record.
func enrollMachine(ctx context.Context, name, addr string) error {
	auth, err := openAuthority()
	if err != nil {
		return err
	}
	rec, err := machines.Load(dirs.Machines(), name)
	if err != nil {
		return err
	}
	r, err := dialEnrollment(auth, addr)
	if err != nil {
		return err
	}
	defer r.Close()

	csr, err := requestEnrollment(ctx, r, name, rec.Token)
	if err != nil {
		return err
	}
	if err := certifyMachine(ctx, r, auth, csr, rec.ID); err != nil {
		return err
	}

	rec.Token = nil
	return machines.Save(dirs.Machines(), name, rec)
}

// Returns a client for enrolling the agent on the machine at addr.
//
// TLS cannot verify an agent that has not enrolled, so crux trusts it through
// the token proof instead (see [trust.EnrollConfig]).
func dialEnrollment(auth *authority.Authority, addr string) (*client, error) {
	cert, err := auth.Client()
	if err != nil {
		return nil, err
	}
	return &client{
		base: (&url.URL{Scheme: "https", Host: addr}).String(),
		client: &http.Client{Transport: &http.Transport{
			TLSClientConfig:   trust.EnrollConfig(cert),
			ForceAttemptHTTP2: true,
		}},
	}, nil
}

// Asks the agent on the machine name for its certificate request and checks
// it.
//
// The agent must prove it holds token on this TLS session, and the request
// must carry the key the agent used in TLS.
func requestEnrollment(ctx context.Context, r *client, name string, token []byte) (*x509.CertificateRequest, error) {
	resp, err := r.send(ctx, agent.EndpointEnroll, agent.MediaTypeEnrollRequest, agent.EnrollRequest{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var er agent.EnrollResponse
	if err := decodeResponse(resp, agent.MediaTypeEnrollResponse, &er); err != nil {
		return nil, err
	}
	if resp.TLS == nil {
		return nil, crex.Newf(ErrNotEnrolled, "enrollment response for machine %q did not arrive over TLS", name)
	}
	return trust.VerifyEnrollment(*resp.TLS, er.Request, er.Proof, token)
}

// Signs the agent certificate for csr with the root and sends it to the agent
// to install.
func certifyMachine(ctx context.Context, r *client, auth *authority.Authority, csr *x509.CertificateRequest, id trust.MachineID) error {
	der, err := auth.SignAgent(csr, id)
	if err != nil {
		return err
	}
	req := agent.CertifyRequest{Certificate: der}
	return r.call(ctx, agent.EndpointCertify, agent.MediaTypeCertifyRequest, req, agent.MediaTypeCertifyResponse, &agent.CertifyResponse{})
}

// Blocks until the machine name at addr is enrolled and healthy.
//
// Retries every [machinePollInterval], because the machine may still be
// booting. Returns an error when ctx ends or [machineReadyTimeout] passes.
func waitForMachine(ctx context.Context, name, addr string) error {
	err := retry.Poll(ctx, machineReadyTimeout, machinePollInterval, func(ctx context.Context) error {
		return machineReady(ctx, name, addr)
	})
	if errors.Is(err, retry.ErrTimeout) {
		return crex.SystemErrorf("machine not ready", "cruxd did not become healthy within %s", machineReadyTimeout).
			Recovery("Recreate the machine and retry.").
			Cause(err).
			Err()
	}
	return err
}

// Enrolls the machine if it has not enrolled yet, then checks its health.
func machineReady(ctx context.Context, name, addr string) error {
	ctx, cancel := context.WithTimeout(ctx, machineTryTimeout)
	defer cancel()

	rec, err := machines.Load(dirs.Machines(), name)
	if err != nil {
		return err
	}
	if !rec.Enrolled() {
		if err := enrollMachine(ctx, name, addr); err != nil {
			return err
		}
	}
	return machineHealthy(ctx, name, addr)
}

// Connects to the agent on the enrolled machine and checks its health.
func machineHealthy(ctx context.Context, name, addr string) error {
	ctx, cancel := context.WithTimeout(ctx, machineTryTimeout)
	defer cancel()

	a, err := connectMachine(ctx, name, addr)
	if err != nil {
		return err
	}
	defer a.Close()
	return a.Health(ctx)
}

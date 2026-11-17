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
	speccompute "github.com/cruciblehq/spec/compute"
	"github.com/cruciblehq/spec/trust"
	"github.com/cruciblehq/utils-go/codec"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/pki/keychain"
	"github.com/cruciblehq/utils-go/retry"
)

// Keychain service under which crux keeps its private keys.
const keychainService = "crux"

// Machine readiness timing.
const (
	machineReadyTimeout = 15 * time.Minute // Longest wait for a machine to enroll and report healthy.
	machinePollInterval = 2 * time.Second  // Wait between readiness attempts.
	machineTryTimeout   = 5 * time.Second  // Longest wait for one readiness attempt.
)

// Opens crux's certificate authority, with its private keys in the OS keychain.
func openAuthority() (*authority.Authority, error) {
	return authority.Open(dirs.Trust(), keychain.New(keychainService))
}

// Creates the record for a new machine and returns its enrollment data.
//
// The record keeps the machine identifier and token so crux can enroll the
// machine once it boots. The returned JSON goes into the data the machine
// boots from, at [trust.EnrollmentPath].
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

// Connects to the enrolled machine name at addr over mutual TLS.
//
// crux accepts only the certificate the root signed for this machine, and
// sends its revocation list in the handshake.
func connectMachine(ctx context.Context, name, addr string) (speccompute.Runtime, error) {
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
// The machine proves it holds the token from its record, crux signs its
// certificate and installs it, and the token is then dropped from the record.
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

// Returns a runtime for enrolling the machine at addr.
//
// TLS cannot verify a machine that has not enrolled, so crux trusts it through
// the token proof instead (see [trust.EnrollConfig]).
func dialEnrollment(auth *authority.Authority, addr string) (*cruxdRuntime, error) {
	cert, err := auth.Client()
	if err != nil {
		return nil, err
	}
	return &cruxdRuntime{
		base: (&url.URL{Scheme: "https", Host: addr}).String(),
		client: &http.Client{Transport: &http.Transport{
			TLSClientConfig:   trust.EnrollConfig(cert),
			ForceAttemptHTTP2: true,
		}},
	}, nil
}

// Asks the machine name for its certificate request and checks it.
//
// The machine must prove it holds token on this TLS session, and the request
// must carry the key the machine used in TLS.
func requestEnrollment(ctx context.Context, r *cruxdRuntime, name string, token []byte) (*x509.CertificateRequest, error) {
	resp, err := r.post(ctx, speccompute.EndpointEnroll, nil, "", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var er speccompute.EnrollResponse
	if err := decodeResponse(resp, speccompute.MediaTypeEnrollResponse, &er); err != nil {
		return nil, err
	}
	if resp.TLS == nil {
		return nil, crex.Newf(ErrNotEnrolled, "enrollment response for machine %q did not arrive over TLS", name)
	}
	return trust.VerifyEnrollment(*resp.TLS, er.Request, er.Proof, token)
}

// Signs the machine certificate for csr with the root and installs it.
func certifyMachine(ctx context.Context, r *cruxdRuntime, auth *authority.Authority, csr *x509.CertificateRequest, id trust.MachineID) error {
	der, err := auth.SignMachine(csr, id)
	if err != nil {
		return err
	}
	req := speccompute.CertifyRequest{Certificate: der}
	return r.call(ctx, speccompute.EndpointCertify, speccompute.MediaTypeCertifyRequest, req, "", nil)
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

// Connects to the enrolled machine and checks its health.
func machineHealthy(ctx context.Context, name, addr string) error {
	ctx, cancel := context.WithTimeout(ctx, machineTryTimeout)
	defer cancel()

	rt, err := connectMachine(ctx, name, addr)
	if err != nil {
		return err
	}
	defer rt.Close()
	return rt.Health(ctx)
}

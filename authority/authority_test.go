package authority

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/cruciblehq/spec/trust"
	"github.com/cruciblehq/utils-go/file"
	"github.com/cruciblehq/utils-go/pki"
)

func openTest(t *testing.T) (*Authority, string, pki.KeyStore) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "trust")
	keys := pki.NewDirKeyStore(filepath.Join(t.TempDir(), "keys"))
	a, err := Open(dir, keys)
	if err != nil {
		t.Fatalf("Open() = %v, want nil", err)
	}
	return a, dir, keys
}

func TestOpenCreates(t *testing.T) {
	a, dir, keys := openTest(t)

	if !a.Root().IsCA {
		t.Error("Root().IsCA = false, want true")
	}
	if a.Revocations().Number.Int64() != 1 {
		t.Errorf("Revocations().Number = %s, want 1", a.Revocations().Number)
	}
	if err := a.issuer.CheckSignatureFrom(a.Root()); err != nil {
		t.Errorf("issuer is not signed by the root: %v", err)
	}
	for _, name := range []string{rootFile, issuerFile, revocationsFile} {
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("stat %s: %v", name, err)
		}
		if info.Mode().Perm() != file.OwnerFileMode {
			t.Errorf("%s mode = %v, want %v", name, info.Mode().Perm(), file.OwnerFileMode)
		}
	}
	if _, err := keys.Load(rootKeyName); err != nil {
		t.Errorf("root key not in key store: %v", err)
	}
}

func TestOpenReloads(t *testing.T) {
	a, dir, keys := openTest(t)

	b, err := Open(dir, keys)
	if err != nil {
		t.Fatalf("Open() = %v, want nil", err)
	}
	if !b.Root().Equal(a.Root()) {
		t.Error("reopened root differs, want the same root")
	}
	if !b.issuer.Equal(a.issuer) {
		t.Error("reopened issuer differs, want the same issuer")
	}
}

func TestOpenRenewsMissingIssuer(t *testing.T) {
	a, dir, keys := openTest(t)
	if err := os.Remove(filepath.Join(dir, issuerFile)); err != nil {
		t.Fatal(err)
	}

	b, err := Open(dir, keys)
	if err != nil {
		t.Fatalf("Open() = %v, want nil", err)
	}
	if b.issuer.Equal(a.issuer) {
		t.Error("issuer was not renewed")
	}
	if err := b.issuer.CheckSignatureFrom(b.Root()); err != nil {
		t.Errorf("renewed issuer is not signed by the root: %v", err)
	}
}

func TestOpenMissingRootKey(t *testing.T) {
	_, dir, _ := openTest(t)
	if err := os.Remove(filepath.Join(dir, issuerFile)); err != nil {
		t.Fatal(err)
	}

	if _, err := Open(dir, pki.NewDirKeyStore(t.TempDir())); !errors.Is(err, pki.ErrKeyNotFound) {
		t.Errorf("Open() = %v, want %v", err, pki.ErrKeyNotFound)
	}
}

func TestOpenCorruptRevocations(t *testing.T) {
	_, dir, keys := openTest(t)
	if err := os.WriteFile(filepath.Join(dir, revocationsFile), []byte("garbage"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := Open(dir, keys); !errors.Is(err, trust.ErrInvalidRevocations) {
		t.Errorf("Open() = %v, want %v", err, trust.ErrInvalidRevocations)
	}
}

func TestClient(t *testing.T) {
	a, _, _ := openTest(t)

	cert, err := a.Client()
	if err != nil {
		t.Fatalf("Client() = %v, want nil", err)
	}
	if len(cert.Certificate) != 2 {
		t.Fatalf("len(Certificate) = %d, want 2", len(cert.Certificate))
	}
	roots := x509.NewCertPool()
	roots.AddCert(a.Root())
	intermediates := x509.NewCertPool()
	intermediates.AddCert(a.issuer)
	chains, err := cert.Leaf.Verify(x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	})
	if err != nil {
		t.Fatalf("client certificate does not verify: %v", err)
	}
	if err := trust.VerifyClient(chains, a.Revocations()); err != nil {
		t.Errorf("VerifyClient() = %v, want nil", err)
	}
}

func TestSignAgent(t *testing.T) {
	a, _, _ := openTest(t)
	_, key, _ := ed25519.GenerateKey(rand.Reader)
	der, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{}, key)
	if err != nil {
		t.Fatal(err)
	}
	csr, err := x509.ParseCertificateRequest(der)
	if err != nil {
		t.Fatal(err)
	}
	id := trust.NewMachineID()

	signed, err := a.SignAgent(csr, id)
	if err != nil {
		t.Fatalf("SignAgent() = %v, want nil", err)
	}
	cert, err := x509.ParseCertificate(signed)
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(a.Root())
	chains, err := cert.Verify(x509.VerifyOptions{
		Roots:     roots,
		DNSName:   id.Name(),
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	})
	if err != nil {
		t.Fatalf("agent certificate does not verify: %v", err)
	}
	if err := trust.VerifyServer(chains, a.Revocations()); err != nil {
		t.Errorf("VerifyServer() = %v, want nil", err)
	}
	if !cert.NotAfter.Equal(agentNotAfter) {
		t.Errorf("NotAfter = %v, want %v", cert.NotAfter, agentNotAfter)
	}
}

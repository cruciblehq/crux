package authority

import (
	"crypto/ed25519"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/cruciblehq/spec/trust"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/file"
)

// Files in the key directory.
const (
	rootKeyName   = "ca-root.key"   // Root private key.
	issuerKeyName = "ca-issuer.key" // Issuer private key.
)

// Common names of the certificates crux issues.
const (
	rootName   = "crux root"   // Root certificate.
	issuerName = "crux issuer" // Issuer certificates.
	clientName = "crux"        // Client certificates.
)

// Files in the authority directory.
const (
	rootFile        = "root.crt"        // DER root certificate.
	issuerFile      = "issuer.crt"      // DER issuer certificate.
	revocationsFile = "revocations.crl" // DER revocation list.
)

// Expiry of agent certificates.
//
// RFC 5280 reserves this date for certificates with no well defined
// expiration. Agent certificates last for the life of their machine
// and are withdrawn only through the revocation list.
var agentNotAfter = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)

// Certificate authority backed by a directory of public files and a directory
// of private keys.
type Authority struct {
	dir       string               // Directory of the public files.
	keyDir    string               // Directory holding the root and issuer keys.
	root      *x509.Certificate    // Root certificate.
	issuer    *x509.Certificate    // Current issuer certificate.
	issuerKey ed25519.PrivateKey   // Current issuer key.
	crl       *x509.RevocationList // Current revocation list.
}

// Opens the authority in dir with its private keys in keyDir, creating it on
// first use.
//
// The first call creates the root, an empty revocation list, and an issuer.
// Later calls renew the issuer when it is missing, when its key cannot be
// loaded, or when a client certificate signed now could outlive it.
func Open(dir, keyDir string) (*Authority, error) {
	if err := os.MkdirAll(dir, file.OwnerDirMode); err != nil {
		return nil, err
	}
	a := &Authority{dir: dir, keyDir: keyDir}

	root, err := readCertificate(filepath.Join(dir, rootFile))
	if errors.Is(err, fs.ErrNotExist) {
		if err := a.create(); err != nil {
			return nil, err
		}
		return a, nil
	}
	if err != nil {
		return nil, err
	}
	a.root = root

	if a.crl, err = a.loadRevocations(); err != nil {
		return nil, err
	}
	if err := a.loadIssuer(); err != nil {
		return nil, err
	}
	return a, nil
}

// Reads and parses the DER certificate at path.
//
// A missing file returns an error that wraps [fs.ErrNotExist].
func readCertificate(path string) (*x509.Certificate, error) {
	der, err := os.ReadFile(path)
	if err != nil {
		return nil, crex.Wrap(ErrInvalidState, err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, crex.Wrap(ErrInvalidState, err)
	}
	return cert, nil
}

// Reads the saved revocation list and checks that the root signed it.
func (a *Authority) loadRevocations() (*x509.RevocationList, error) {
	der, err := os.ReadFile(filepath.Join(a.dir, revocationsFile))
	if err != nil {
		return nil, crex.Wrap(ErrInvalidState, err)
	}
	return trust.ParseRevocations(der, a.root)
}

// Reads the private key named name from the key directory.
//
// A missing file returns an error that wraps [fs.ErrNotExist].
func (a *Authority) loadKey(name string) (ed25519.PrivateKey, error) {
	der, err := os.ReadFile(filepath.Join(a.keyDir, name))
	if err != nil {
		return nil, err
	}
	return trust.DecodeKey(der)
}

// Writes key to the key directory under name, replacing any previous key.
func (a *Authority) storeKey(name string, key ed25519.PrivateKey) error {
	der, err := trust.EncodeKey(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(a.keyDir, file.OwnerDirMode); err != nil {
		return err
	}
	return file.WriteFile(filepath.Join(a.keyDir, name), der, file.OwnerFileMode)
}

// Loads the saved issuer, or renews it.
//
// The issuer is renewed when it is missing, when its key cannot be loaded, or
// when a client certificate signed now could outlive it.
func (a *Authority) loadIssuer() error {
	issuer, err := readCertificate(filepath.Join(a.dir, issuerFile))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if issuer != nil && time.Now().Add(trust.ClientLifetime).Before(issuer.NotAfter) {
		if key, err := a.loadKey(issuerKeyName); err == nil {
			a.issuer, a.issuerKey = issuer, key
			return nil
		}
	}
	rootKey, err := a.loadKey(rootKeyName)
	if err != nil {
		return err
	}
	return a.renewIssuer(rootKey)
}

// Creates the root, an empty revocation list, and the first issuer.
func (a *Authority) create() error {
	pub, key, err := ed25519.GenerateKey(nil)
	if err != nil {
		return err
	}
	root, err := trust.Sign(trust.AuthorityTemplate(rootName, trust.ValidFor(trust.RootLifetime), false), nil, pub, key)
	if err != nil {
		return err
	}
	if err := a.storeKey(rootKeyName, key); err != nil {
		return err
	}
	if err := file.WriteFile(filepath.Join(a.dir, rootFile), root.Raw, file.OwnerFileMode); err != nil {
		return err
	}
	a.root = root

	crl, err := trust.SignRevocationList(1, trust.ValidFor(trust.IssuerLifetime), nil, root, key)
	if err != nil {
		return err
	}
	if err := file.WriteFile(filepath.Join(a.dir, revocationsFile), crl.Raw, file.OwnerFileMode); err != nil {
		return err
	}
	a.crl = crl

	return a.renewIssuer(key)
}

// Creates a new issuer signed by the root.
func (a *Authority) renewIssuer(rootKey ed25519.PrivateKey) error {
	pub, key, err := ed25519.GenerateKey(nil)
	if err != nil {
		return err
	}
	issuer, err := trust.Sign(trust.AuthorityTemplate(issuerName, trust.ValidFor(trust.IssuerLifetime), true), a.root, pub, rootKey)
	if err != nil {
		return err
	}
	if err := a.storeKey(issuerKeyName, key); err != nil {
		return err
	}
	if err := file.WriteFile(filepath.Join(a.dir, issuerFile), issuer.Raw, file.OwnerFileMode); err != nil {
		return err
	}
	a.issuer, a.issuerKey = issuer, key
	return nil
}

// Returns the root certificate.
func (a *Authority) Root() *x509.Certificate {
	return a.root
}

// Returns the current revocation list.
func (a *Authority) Revocations() *x509.RevocationList {
	return a.crl
}

// Returns a new client certificate followed by its issuer.
func (a *Authority) Client() (tls.Certificate, error) {
	pub, key, err := ed25519.GenerateKey(nil)
	if err != nil {
		return tls.Certificate{}, err
	}
	leaf, err := trust.Sign(trust.LeafTemplate(clientName, trust.ValidFor(trust.ClientLifetime), x509.ExtKeyUsageClientAuth), a.issuer, pub, a.issuerKey)
	if err != nil {
		return tls.Certificate{}, err
	}
	return tls.Certificate{
		Certificate: [][]byte{leaf.Raw, a.issuer.Raw},
		PrivateKey:  key,
		Leaf:        leaf,
	}, nil
}

// Signs the agent certificate for csr with the root.
//
// The certificate names machine id, is for server authentication, and lasts
// for the life of the machine. Returns the DER certificate.
func (a *Authority) SignAgent(csr *x509.CertificateRequest, id trust.MachineID) ([]byte, error) {
	rootKey, err := a.loadKey(rootKeyName)
	if err != nil {
		return nil, err
	}
	pub, ok := csr.PublicKey.(ed25519.PublicKey)
	if !ok {
		return nil, crex.Newf(trust.ErrInvalidEnrollment, "unexpected request key type %T", csr.PublicKey)
	}
	tmpl := trust.LeafTemplate(id.Name(), trust.ValidUntil(agentNotAfter), x509.ExtKeyUsageServerAuth, id.Name())
	cert, err := trust.Sign(tmpl, a.root, pub, rootKey)
	if err != nil {
		return nil, err
	}
	return cert.Raw, nil
}

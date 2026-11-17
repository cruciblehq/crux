package authority

import (
	"crypto"
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
	"github.com/cruciblehq/utils-go/pki"
)

// Names of the private keys in the key store.
const (
	rootKeyName   = "root"   // Root private key.
	issuerKeyName = "issuer" // Issuer private key.
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

// Expiry of machine certificates.
//
// RFC 5280 reserves this date for certificates with no well defined
// expiration. Machine certificates last for the life of the machine
// and are withdrawn only through the revocation list.
var machineNotAfter = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)

// Certificate authority backed by a directory and a key store.
type Authority struct {
	dir       string               // Directory of the public files.
	keys      pki.KeyStore         // Store holding the root and issuer keys.
	root      *x509.Certificate    // Root certificate.
	issuer    *x509.Certificate    // Current issuer certificate.
	issuerKey crypto.Signer        // Current issuer key.
	crl       *x509.RevocationList // Current revocation list.
}

// Opens the authority in dir with its private keys in keys, creating it on
// first use.
//
// The first call creates the root, an empty revocation list, and an issuer.
// Later calls renew the issuer when a client certificate minted now could
// outlive it.
func Open(dir string, keys pki.KeyStore) (*Authority, error) {
	if err := os.MkdirAll(dir, file.OwnerDirMode); err != nil {
		return nil, err
	}
	a := &Authority{dir: dir, keys: keys}

	root, err := pki.ReadCertificate(filepath.Join(dir, rootFile))
	if errors.Is(err, fs.ErrNotExist) {
		if err := a.create(); err != nil {
			return nil, err
		}
		return a, nil
	}
	if err != nil {
		return nil, crex.Wrap(ErrInvalidState, err)
	}
	a.root = root

	crlDER, err := os.ReadFile(filepath.Join(dir, revocationsFile))
	if err != nil {
		return nil, crex.Wrap(ErrInvalidState, err)
	}
	if a.crl, err = trust.ParseRevocations(crlDER, root); err != nil {
		return nil, err
	}

	issuer, err := pki.ReadCertificate(filepath.Join(dir, issuerFile))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, crex.Wrap(ErrInvalidState, err)
	}
	if issuer != nil && time.Now().Add(trust.ClientLifetime).Before(issuer.NotAfter) {
		if key, err := keys.Load(issuerKeyName); err == nil {
			a.issuer, a.issuerKey = issuer, key
			return a, nil
		}
	}
	rootKey, err := keys.Load(rootKeyName)
	if err != nil {
		return nil, err
	}
	if err := a.renewIssuer(rootKey); err != nil {
		return nil, err
	}
	return a, nil
}

// Creates the root, an empty revocation list, and the first issuer.
func (a *Authority) create() error {
	pub, key, err := ed25519.GenerateKey(nil)
	if err != nil {
		return err
	}
	root, err := pki.Sign(pki.AuthorityTemplate(rootName, pki.ValidFor(trust.RootLifetime), false), nil, pub, key)
	if err != nil {
		return err
	}
	if err := a.keys.Store(rootKeyName, key); err != nil {
		return err
	}
	if err := file.WriteFile(filepath.Join(a.dir, rootFile), root.Raw, file.OwnerFileMode); err != nil {
		return err
	}
	a.root = root

	crl, err := pki.SignRevocationList(1, pki.ValidFor(trust.IssuerLifetime), nil, root, key)
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
func (a *Authority) renewIssuer(rootKey crypto.Signer) error {
	pub, key, err := ed25519.GenerateKey(nil)
	if err != nil {
		return err
	}
	issuer, err := pki.Sign(pki.AuthorityTemplate(issuerName, pki.ValidFor(trust.IssuerLifetime), true), a.root, pub, rootKey)
	if err != nil {
		return err
	}
	if err := a.keys.Store(issuerKeyName, key); err != nil {
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
	leaf, err := pki.Sign(pki.LeafTemplate(clientName, pki.ValidFor(trust.ClientLifetime), x509.ExtKeyUsageClientAuth), a.issuer, pub, a.issuerKey)
	if err != nil {
		return tls.Certificate{}, err
	}
	return tls.Certificate{
		Certificate: [][]byte{leaf.Raw, a.issuer.Raw},
		PrivateKey:  key,
		Leaf:        leaf,
	}, nil
}

// Signs the machine certificate for csr with the root.
//
// The certificate names machine id, is for server authentication, and lasts
// for the life of the machine. Returns the DER certificate.
func (a *Authority) SignMachine(csr *x509.CertificateRequest, id trust.MachineID) ([]byte, error) {
	rootKey, err := a.keys.Load(rootKeyName)
	if err != nil {
		return nil, err
	}
	pub, ok := csr.PublicKey.(ed25519.PublicKey)
	if !ok {
		return nil, crex.Newf(trust.ErrInvalidEnrollment, "unexpected request key type %T", csr.PublicKey)
	}
	tmpl := pki.LeafTemplate(id.Name(), pki.ValidUntil(machineNotAfter), x509.ExtKeyUsageServerAuth, id.Name())
	cert, err := pki.Sign(tmpl, a.root, pub, rootKey)
	if err != nil {
		return nil, err
	}
	return cert.Raw, nil
}

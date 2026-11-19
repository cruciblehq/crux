// Package authority is crux's certificate authority.
//
// The root and issuer private keys are kept in a [pki.KeyStore], which crux
// backs with the OS keychain, so they are never written to disk and are
// unlocked by the OS login. Their certificates and the revocation list are
// public and live in a directory.
//
// The root signs the issuer and every agent certificate. The issuer signs a
// new client certificate each time crux runs, and is renewed from the root
// before a client certificate could outlive it.
package authority

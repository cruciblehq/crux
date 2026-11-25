// Package authority is crux's certificate authority.
//
// The root and issuer private keys are files readable only by the user, kept in
// their own directory. Their certificates and the revocation list are public
// and live in another directory.
//
// The root signs the issuer and every agent certificate. The issuer signs a
// new client certificate each time crux runs, and is renewed from the root
// before a client certificate could outlive it.
package authority

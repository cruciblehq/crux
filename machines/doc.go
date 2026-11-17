// Package machines keeps what crux knows about each machine it created.
//
// A record holds the machine identifier crux checks on every connection and,
// until enrollment completes, the one-time enrollment token. Records are
// readable only by the user, because the token lets a machine enroll.
package machines

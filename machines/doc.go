// Package machines keeps what crux knows about each machine it created.
//
// A record holds the machine identifier crux checks on every connection and,
// until enrollment completes, the enrollment token, which is used once.
// Records are readable only by the user, because the token lets an agent
// enroll its machine.
package machines

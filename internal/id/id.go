// Package id generates short, sortable, prefixed identifiers.
package id

import (
	"crypto/rand"
	"encoding/hex"
)

// New returns prefix + "_" + 16 random hex characters, e.g. "evt_9f3a...".
func New(prefix string) string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("id: reading random bytes: " + err.Error())
	}
	return prefix + "_" + hex.EncodeToString(b[:])
}

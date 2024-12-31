package tpkg

import (
	axongo "github.com/axonfibre/axon.go/v4"
)

// Must panics if the given error is not nil.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// ReferenceUnlock returns a reference unlock with the given index.
func ReferenceUnlock(index uint16) *axongo.ReferenceUnlock {
	return &axongo.ReferenceUnlock{Reference: index}
}

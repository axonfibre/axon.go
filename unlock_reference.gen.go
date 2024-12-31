package axongo

import (
	"github.com/axonfibre/fibre.go/serializer/v2"
)

// ReferenceUnlock is an Unlock which references a previous input/unlock.
type ReferenceUnlock struct {
	// The other input/unlock this ReferenceUnlock references to.
	Reference uint16 `serix:""`
}

func (r *ReferenceUnlock) Clone() Unlock {
	return &ReferenceUnlock{
		Reference: r.Reference,
	}
}

func (r *ReferenceUnlock) SourceAllowed(address Address) bool {
	_, ok := address.(ChainAddress)

	return !ok
}

func (r *ReferenceUnlock) Chainable() bool {
	return false
}

func (r *ReferenceUnlock) ReferencedInputIndex() uint16 {
	return r.Reference
}

func (r *ReferenceUnlock) Type() UnlockType {
	return UnlockReference
}

func (r *ReferenceUnlock) Size() int {
	// UnlockType + Reference
	return serializer.SmallTypeDenotationByteSize + serializer.UInt16ByteSize
}

func (r *ReferenceUnlock) WorkScore(_ *WorkScoreParameters) (WorkScore, error) {
	return 0, nil
}

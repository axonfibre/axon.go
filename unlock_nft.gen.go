package axongo

import (
	"github.com/axonfibre/fibre.go/serializer/v2"
)

// NFTUnlock is an Unlock which references a previous input/unlock.
type NFTUnlock struct {
	// The other input/unlock this NFTUnlock references to.
	Reference uint16 `serix:""`
}

func (r *NFTUnlock) Clone() Unlock {
	return &NFTUnlock{
		Reference: r.Reference,
	}
}

func (r *NFTUnlock) SourceAllowed(address Address) bool {
	_, ok := address.(*NFTAddress)

	return ok
}

func (r *NFTUnlock) Chainable() bool {
	return true
}

func (r *NFTUnlock) ReferencedInputIndex() uint16 {
	return r.Reference
}

func (r *NFTUnlock) Type() UnlockType {
	return UnlockNFT
}

func (r *NFTUnlock) Size() int {
	// UnlockType + Reference
	return serializer.SmallTypeDenotationByteSize + serializer.UInt16ByteSize
}

func (r *NFTUnlock) WorkScore(_ *WorkScoreParameters) (WorkScore, error) {
	return 0, nil
}

package axongo

import (
	"github.com/axonfibre/fibre.go/serializer/v2"
)

// AccountUnlock is an Unlock which references a previous input/unlock.
type AccountUnlock struct {
	// The other input/unlock this AccountUnlock references to.
	Reference uint16 `serix:""`
}

func (r *AccountUnlock) Clone() Unlock {
	return &AccountUnlock{
		Reference: r.Reference,
	}
}

func (r *AccountUnlock) SourceAllowed(address Address) bool {
	_, ok := address.(*AccountAddress)

	return ok
}

func (r *AccountUnlock) Chainable() bool {
	return true
}

func (r *AccountUnlock) ReferencedInputIndex() uint16 {
	return r.Reference
}

func (r *AccountUnlock) Type() UnlockType {
	return UnlockAccount
}

func (r *AccountUnlock) Size() int {
	// UnlockType + Reference
	return serializer.SmallTypeDenotationByteSize + serializer.UInt16ByteSize
}

func (r *AccountUnlock) WorkScore(_ *WorkScoreParameters) (WorkScore, error) {
	return 0, nil
}

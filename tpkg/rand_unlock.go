package tpkg

import (
	axongo "github.com/axonfibre/axon.go/v4"
)

// RandUnlock returns a random unlock (except Signature, Reference, Account, Anchor, NFT).
func RandUnlock(allowEmptyUnlock bool) axongo.Unlock {
	unlockTypes := []axongo.UnlockType{axongo.UnlockSignature, axongo.UnlockReference, axongo.UnlockAccount, axongo.UnlockAnchor, axongo.UnlockNFT}

	if allowEmptyUnlock {
		unlockTypes = append(unlockTypes, axongo.UnlockEmpty)
	}

	unlockType := unlockTypes[RandInt(len(unlockTypes))]

	//nolint:exhaustive
	switch unlockType {
	case axongo.UnlockSignature:
		return RandEd25519SignatureUnlock()
	case axongo.UnlockReference:
		return RandReferenceUnlock()
	case axongo.UnlockAccount:
		return RandAccountUnlock()
	case axongo.UnlockAnchor:
		return RandAnchorUnlock()
	case axongo.UnlockNFT:
		return RandNFTUnlock()
	case axongo.UnlockEmpty:
		return &axongo.EmptyUnlock{}
	default:
		panic("all supported unlock types should be handled above")
	}
}

// RandEd25519SignatureUnlock returns a random Ed25519 signature unlock.
func RandEd25519SignatureUnlock() *axongo.SignatureUnlock {
	return &axongo.SignatureUnlock{Signature: RandEd25519Signature()}
}

// RandReferenceUnlock returns a random reference unlock.
func RandReferenceUnlock() *axongo.ReferenceUnlock {
	return ReferenceUnlock(uint16(RandInt(1000)))
}

// RandAccountUnlock returns a random account unlock.
func RandAccountUnlock() *axongo.AccountUnlock {
	return &axongo.AccountUnlock{Reference: uint16(RandInt(1000))}
}

// RandAnchorUnlock returns a random anchor unlock.
func RandAnchorUnlock() *axongo.AnchorUnlock {
	return &axongo.AnchorUnlock{Reference: uint16(RandInt(1000))}
}

// RandNFTUnlock returns a random account unlock.
func RandNFTUnlock() *axongo.NFTUnlock {
	return &axongo.NFTUnlock{Reference: uint16(RandInt(1000))}
}

// RandMultiUnlock returns a random multi unlock.
func RandMultiUnlock() *axongo.MultiUnlock {
	// at least 2 unlocks but max 10 unlocks
	unlockCnt := RandInt(9) + 2
	unlocks := make([]axongo.Unlock, 0, unlockCnt)

	for range unlockCnt {
		unlocks = append(unlocks, RandUnlock(true))
	}

	return &axongo.MultiUnlock{
		Unlocks: unlocks,
	}
}

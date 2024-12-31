package tpkg

import (
	axongo "github.com/axonfibre/axon.go/v4"
)

func RandIdentifier() axongo.Identifier {
	return Rand32ByteArray()
}

// RandBlockID produces a random block ID.
func RandBlockID() axongo.BlockID {
	return Rand36ByteArray()
}

// SortedRandBlockIDs returned random block IDs.
func SortedRandBlockIDs(count int) axongo.BlockIDs {
	slice := make([]axongo.BlockID, count)
	for i, ele := range SortedRand36ByteArray(count) {
		slice[i] = ele
	}

	return slice
}

func RandAccountID() axongo.AccountID {
	alias := axongo.AccountID{}
	copy(alias[:], RandBytes(axongo.AccountIDLength))

	return alias
}

func RandAnchorID() axongo.AnchorID {
	anchorID := axongo.AnchorID{}
	copy(anchorID[:], RandBytes(axongo.AnchorIDLength))

	return anchorID
}

func RandNFTID() axongo.NFTID {
	nft := axongo.NFTID{}
	copy(nft[:], RandBytes(axongo.NFTIDLength))

	return nft
}

func RandDelegationID() axongo.DelegationID {
	delegation := axongo.DelegationID{}
	copy(delegation[:], RandBytes(axongo.DelegationIDLength))

	return delegation
}

func RandNativeTokenID() axongo.NativeTokenID {
	var nativeTokenID axongo.NativeTokenID
	copy(nativeTokenID[:], RandBytes(axongo.NativeTokenIDLength))

	// the underlying address needs to be an account address
	nativeTokenID[0] = byte(axongo.AddressAccount)

	// set the simple token scheme type
	nativeTokenID[axongo.FoundryIDLength-axongo.FoundryTokenSchemeLength] = byte(axongo.TokenSchemeSimple)

	return nativeTokenID
}

package tpkg

import (
	"bytes"
	"fmt"
	"slices"

	axongo "github.com/axonfibre/axon.go/v4"
)

// RandEd25519Address returns a random Ed25519 address.
func RandEd25519Address() *axongo.Ed25519Address {
	edAddr := &axongo.Ed25519Address{}
	addr := RandBytes(axongo.Ed25519AddressBytesLength)
	copy(edAddr[:], addr)

	return edAddr
}

// RandAccountAddress returns a random AccountAddress.
func RandAccountAddress() *axongo.AccountAddress {
	addr := &axongo.AccountAddress{}
	accountID := RandBytes(axongo.AccountAddressBytesLength)
	copy(addr[:], accountID)

	return addr
}

// RandNFTAddress returns a random NFTAddress.
func RandNFTAddress() *axongo.NFTAddress {
	addr := &axongo.NFTAddress{}
	nftID := RandBytes(axongo.NFTAddressBytesLength)
	copy(addr[:], nftID)

	return addr
}

// RandAnchorAddress returns a random AnchorAddress.
func RandAnchorAddress() *axongo.AnchorAddress {
	addr := &axongo.AnchorAddress{}
	anchorID := RandBytes(axongo.AnchorAddressBytesLength)
	copy(addr[:], anchorID)

	return addr
}

// RandImplicitAccountCreationAddress returns a random ImplicitAccountCreationAddress.
func RandImplicitAccountCreationAddress() *axongo.ImplicitAccountCreationAddress {
	iacAddr := &axongo.ImplicitAccountCreationAddress{}
	addr := RandBytes(axongo.Ed25519AddressBytesLength)
	copy(iacAddr[:], addr)

	return iacAddr
}

// RandMultiAddress returns a random MultiAddress.
func RandMultiAddress() *axongo.MultiAddress {
	addrCnt := RandInt(9) + 2 // at least 2 addresses but max 10 addresses

	cumulativeWeight := 0
	addresses := make([]*axongo.AddressWithWeight, 0, addrCnt)
	for range addrCnt {
		weight := RandInt(8) + 1
		cumulativeWeight += weight
		addresses = append(addresses, &axongo.AddressWithWeight{
			Address: RandAddress(),
			Weight:  byte(weight),
		})
	}

	slices.SortFunc(addresses, func(a *axongo.AddressWithWeight, b *axongo.AddressWithWeight) int {
		return bytes.Compare(a.Address.ID(), b.Address.ID())
	})

	threshold := RandInt(cumulativeWeight) + 1

	return &axongo.MultiAddress{
		Addresses: addresses,
		Threshold: uint16(threshold),
	}
}

// RandRestrictedEd25519Address returns a random restricted Ed25519 address.
func RandRestrictedEd25519Address(capabilities axongo.AddressCapabilitiesBitMask) *axongo.RestrictedAddress {
	return &axongo.RestrictedAddress{
		Address:             RandEd25519Address(),
		AllowedCapabilities: capabilities,
	}
}

// RandRestrictedAccountAddress returns a random restricted account address.
func RandRestrictedAccountAddress(capabilities axongo.AddressCapabilitiesBitMask) *axongo.RestrictedAddress {
	return &axongo.RestrictedAddress{
		Address:             RandAccountAddress(),
		AllowedCapabilities: capabilities,
	}
}

// RandRestrictedNFTAddress returns a random restricted NFT address.
func RandRestrictedNFTAddress(capabilities axongo.AddressCapabilitiesBitMask) *axongo.RestrictedAddress {
	return &axongo.RestrictedAddress{
		Address:             RandNFTAddress(),
		AllowedCapabilities: capabilities,
	}
}

// RandRestrictedAnchorAddress returns a random restricted anchor address.
func RandRestrictedAnchorAddress(capabilities axongo.AddressCapabilitiesBitMask) *axongo.RestrictedAddress {
	return &axongo.RestrictedAddress{
		Address:             RandAnchorAddress(),
		AllowedCapabilities: capabilities,
	}
}

// RandRestrictedMultiAddress returns a random restricted multi address.
func RandRestrictedMultiAddress(capabilities axongo.AddressCapabilitiesBitMask) *axongo.RestrictedAddress {
	return &axongo.RestrictedAddress{
		Address:             RandMultiAddress(),
		AllowedCapabilities: capabilities,
	}
}

// RandAddress returns a random address (Ed25519, Account, NFT, Anchor).
func RandAddress(addressType ...axongo.AddressType) axongo.Address {
	var addrType axongo.AddressType
	if len(addressType) > 0 {
		addrType = addressType[0]
	} else {
		addressTypes := []axongo.AddressType{axongo.AddressEd25519, axongo.AddressAccount, axongo.AddressNFT, axongo.AddressAnchor}
		addrType = addressTypes[RandInt(len(addressTypes))]
	}

	//nolint:exhaustive
	switch addrType {
	case axongo.AddressEd25519:
		return RandEd25519Address()
	case axongo.AddressAccount:
		return RandAccountAddress()
	case axongo.AddressNFT:
		return RandNFTAddress()
	case axongo.AddressAnchor:
		return RandAnchorAddress()
	default:
		panic(fmt.Sprintf("unknown address type %d", addrType))
	}
}

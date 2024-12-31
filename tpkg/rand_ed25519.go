package tpkg

import (
	"bytes"
	"crypto/ed25519"
	cryptorand "crypto/rand"
	"fmt"
	"slices"

	hiveEd25519 "github.com/axonfibre/fibre.go/crypto/ed25519"
	axongo "github.com/axonfibre/axon.go/v4"
)

// RandEd25519Signature returns a random Ed25519 signature.
func RandEd25519Signature() *axongo.Ed25519Signature {
	edSig := &axongo.Ed25519Signature{}
	pub := RandBytes(ed25519.PublicKeySize)
	sig := RandBytes(ed25519.SignatureSize)
	copy(edSig.PublicKey[:], pub)
	copy(edSig.Signature[:], sig)

	return edSig
}

// RandEd25519PrivateKey returns a random Ed25519 private key.
func RandEd25519PrivateKey() ed25519.PrivateKey {
	seed := RandEd25519Seed()

	return ed25519.NewKeyFromSeed(seed[:])
}

func RandEd25519PublicKey() hiveEd25519.PublicKey {
	//nolint:forcetypeassert // we can safely assume that this is an ed25519.PublicKey
	return hiveEd25519.PublicKey(RandEd25519PrivateKey().Public().(ed25519.PublicKey))
}

// RandEd25519Seed returns a random Ed25519 seed.
func RandEd25519Seed() [ed25519.SeedSize]byte {
	var b [ed25519.SeedSize]byte
	read, err := cryptorand.Read(b[:])
	if read != ed25519.SeedSize {
		panic(fmt.Sprintf("could not read %d required bytes from secure RNG", ed25519.SeedSize))
	}
	if err != nil {
		panic(err)
	}

	return b
}

// RandEd25519Identity produces a random Ed25519 identity.
func RandEd25519Identity() (ed25519.PrivateKey, *axongo.Ed25519Address, axongo.AddressKeys) {
	edSk := RandEd25519PrivateKey()
	//nolint:forcetypeassert // we can safely assume that this is an ed25519.PublicKey
	edAddr := axongo.Ed25519AddressFromPubKey(edSk.Public().(ed25519.PublicKey))
	addrKeys := axongo.NewAddressKeysForEd25519Address(edAddr, edSk)

	return edSk, edAddr, addrKeys
}

// RandEd25519IdentitiesSortedByAddress returns random Ed25519 addresses and keys lexically sorted by the address.
func RandEd25519IdentitiesSortedByAddress(count int) ([]axongo.Address, []axongo.AddressKeys) {
	addresses := make([]axongo.Address, count)
	addressKeys := make([]axongo.AddressKeys, count)
	for i := range count {
		_, addresses[i], addressKeys[i] = RandEd25519Identity()
	}

	// addressses need to be lexically ordered in the MultiAddress
	slices.SortFunc(addresses, func(a axongo.Address, b axongo.Address) int {
		return bytes.Compare(a.ID(), b.ID())
	})

	// addressses need to be lexically ordered in the MultiAddress
	slices.SortFunc(addressKeys, func(a axongo.AddressKeys, b axongo.AddressKeys) int {
		return bytes.Compare(a.Address.ID(), b.Address.ID())
	})

	return addresses, addressKeys
}

// RandImplicitAccountIdentity produces a random Implicit Account identity.
func RandImplicitAccountIdentity() (ed25519.PrivateKey, *axongo.ImplicitAccountCreationAddress, axongo.AddressKeys) {
	edSk := RandEd25519PrivateKey()
	//nolint:forcetypeassert // we can safely assume that this is an ed25519.PublicKey
	implicitAccAddr := axongo.ImplicitAccountCreationAddressFromPubKey(edSk.Public().(ed25519.PublicKey))
	addrKeys := axongo.NewAddressKeysForImplicitAccountCreationAddress(implicitAccAddr, edSk)

	return edSk, implicitAccAddr, addrKeys
}

func RandBlockIssuerKey() axongo.BlockIssuerKey {
	return axongo.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(RandEd25519PublicKey())
}

func RandBlockIssuerKeys(count ...int) axongo.BlockIssuerKeys {
	// We always generate at least one key.
	length := RandInt(10) + 1

	if len(count) > 0 {
		length = count[0]
	}

	blockIssuerKeys := axongo.NewBlockIssuerKeys()
	for range length {
		blockIssuerKeys.Add(RandBlockIssuerKey())
	}
	blockIssuerKeys.Sort()

	return blockIssuerKeys
}

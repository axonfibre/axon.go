//nolint:dupl,revive
package axongo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hiveEd25519 "github.com/axonfibre/fibre.go/crypto/ed25519"
	"github.com/axonfibre/fibre.go/serializer/v2"
	"github.com/axonfibre/fibre.go/serializer/v2/serix"
	"github.com/axonfibre/fibre.go/serializer/v2/stream"
	axongo "github.com/axonfibre/axon.go/v4"
	"github.com/axonfibre/axon.go/v4/tpkg"
	"github.com/axonfibre/axon.go/v4/tpkg/frameworks"
)

func TestAddressDeSerialize(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name:   "ok - Ed25519Address",
			Source: tpkg.RandEd25519Address(),
			Target: &axongo.Ed25519Address{},
		},
		{
			Name:   "ok - AccountAddress",
			Source: tpkg.RandAccountAddress(),
			Target: &axongo.AccountAddress{},
		},
		{
			Name:   "ok - NFTAddress",
			Source: tpkg.RandNFTAddress(),
			Target: &axongo.NFTAddress{},
		},
		{
			Name:   "ok - AnchorAddress",
			Source: tpkg.RandAnchorAddress(),
			Target: &axongo.AnchorAddress{},
		},
		{
			Name:   "ok - ImplicitAccountCreationAddress",
			Source: tpkg.RandImplicitAccountCreationAddress(),
			Target: &axongo.ImplicitAccountCreationAddress{},
		},
		{
			Name:   "ok - MultiAddress",
			Source: tpkg.RandMultiAddress(),
			Target: &axongo.MultiAddress{},
		},
		{
			Name:   "ok - RestrictedEd25519Address without capabilities",
			Source: tpkg.RandRestrictedEd25519Address(axongo.AddressCapabilitiesBitMask{}),
			Target: &axongo.RestrictedAddress{},
		},
		{
			Name:   "ok - RestrictedEd25519Address with capabilities",
			Source: tpkg.RandRestrictedEd25519Address(axongo.AddressCapabilitiesBitMask{0xff}),
			Target: &axongo.RestrictedAddress{},
		},
		{
			Name:   "ok - RestrictedAccountAddress with capabilities",
			Source: tpkg.RandRestrictedAccountAddress(axongo.AddressCapabilitiesBitMask{0xff}),
			Target: &axongo.RestrictedAddress{},
		},
		{
			Name:   "ok - RestrictedNFTAddress with capabilities",
			Source: tpkg.RandRestrictedNFTAddress(axongo.AddressCapabilitiesBitMask{0xff}),
			Target: &axongo.RestrictedAddress{},
		},
		{
			Name:   "ok - RestrictedAnchorAddress with capabilities",
			Source: tpkg.RandRestrictedAnchorAddress(axongo.AddressCapabilitiesBitMask{0xff}),
			Target: &axongo.RestrictedAddress{},
		},
		{
			Name:   "ok - RestrictedMultiAddress with capabilities",
			Source: tpkg.RandRestrictedMultiAddress(axongo.AddressCapabilitiesBitMask{0xff}),
			Target: &axongo.RestrictedAddress{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.Run(t)

			// test the AddressFromReader func
			//nolint:forcetypeassert
			address := tt.Source.(axongo.Address)
			addressBytes, err := tpkg.ZeroCostTestAPI.Encode(address, serix.WithValidation())
			require.NoError(t, err)

			reader := stream.NewByteReader(addressBytes)

			addr, err := axongo.AddressFromReader(reader)
			require.NoError(t, err)

			assert.Equal(t, address, addr)
		})
	}
}

type bech32Test struct {
	name       string
	network    axongo.NetworkPrefix
	addr       axongo.Address
	targetAddr axongo.Address
	bech32     string
}

var bech32Tests = []*bech32Test{
	func() *bech32Test {
		addr := &axongo.Ed25519Address{0x52, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49}
		return &bech32Test{
			name:       "RFC example: Ed25519 mainnet",
			network:    axongo.PrefixMainnet,
			addr:       addr,
			targetAddr: addr,
			bech32:     "iota1qpf0mlq8yxpx2nck8a0slxnzr4ef2ek8f5gqxlzd0wasgp73utryj430ldu",
		}
	}(),
	func() *bech32Test {
		addr := &axongo.Ed25519Address{0x52, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49}
		return &bech32Test{
			name:       "RFC example: Ed25519 testnet",
			network:    axongo.PrefixTestnet,
			addr:       addr,
			targetAddr: addr,
			bech32:     "rms1qpf0mlq8yxpx2nck8a0slxnzr4ef2ek8f5gqxlzd0wasgp73utryjkxa9q5",
		}
	}(),
	func() *bech32Test {
		addr := &axongo.MultiAddress{
			Addresses: []*axongo.AddressWithWeight{
				{
					Address: &axongo.Ed25519Address{0x52, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
					Weight:  1,
				},
				{
					Address: &axongo.Ed25519Address{0x53, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
					Weight:  1,
				},
				{
					Address: &axongo.Ed25519Address{0x54, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
					Weight:  1,
				},
				{
					Address: &axongo.AccountAddress{0x55, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
					Weight:  2,
				},
				{
					Address: &axongo.NFTAddress{0x56, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
					Weight:  3,
				},
				{
					Address: &axongo.AnchorAddress{0x57, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
					Weight:  4,
				},
			},
			Threshold: 2,
		}

		return &bech32Test{
			name:       "Multi Address",
			network:    axongo.PrefixTestnet,
			addr:       addr,
			targetAddr: axongo.NewMultiAddressReferenceFromMultiAddress(addr),
			bech32:     "rms19zt4pdt7fl3lqqgnxduzdyzx45c2pc95jq7xccfqzuncep4zjtxmj4skzzd",
		}
	}(),
	func() *bech32Test {
		addr := &axongo.RestrictedAddress{
			Address:             &axongo.Ed25519Address{0x52, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
			AllowedCapabilities: axongo.AddressCapabilitiesBitMask{0x55},
		}

		return &bech32Test{
			name:       "Restricted Ed25519 Address",
			network:    axongo.PrefixTestnet,
			addr:       addr,
			targetAddr: addr,
			bech32:     "rms1xqq99l0uquscye20zcl47ru6vgwh99txcax3qqmuf4amkpq8683vvjgp25npuutf",
		}
	}(),
	func() *bech32Test {
		addr := &axongo.RestrictedAddress{
			Address:             &axongo.AccountAddress{0x52, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
			AllowedCapabilities: axongo.AddressCapabilitiesBitMask{0x55},
		}

		return &bech32Test{
			name:       "Restricted Account Address",
			network:    axongo.PrefixTestnet,
			addr:       addr,
			targetAddr: addr,
			bech32:     "rms1xqy99l0uquscye20zcl47ru6vgwh99txcax3qqmuf4amkpq8683vvjgp254k6s7n",
		}
	}(),
	func() *bech32Test {
		addr := &axongo.RestrictedAddress{
			Address:             &axongo.NFTAddress{0x52, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
			AllowedCapabilities: axongo.AddressCapabilitiesBitMask{0x55},
		}

		return &bech32Test{
			name:       "Restricted NFT Address",
			network:    axongo.PrefixTestnet,
			addr:       addr,
			targetAddr: addr,
			bech32:     "rms1xqg99l0uquscye20zcl47ru6vgwh99txcax3qqmuf4amkpq8683vvjgp25lxsyg5",
		}
	}(),
	func() *bech32Test {
		addr := &axongo.RestrictedAddress{
			Address:             &axongo.AnchorAddress{0x52, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
			AllowedCapabilities: axongo.AddressCapabilitiesBitMask{0x55},
		}

		return &bech32Test{
			name:       "Restricted Anchor Address",
			network:    axongo.PrefixTestnet,
			addr:       addr,
			targetAddr: addr,
			bech32:     "rms1xqv99l0uquscye20zcl47ru6vgwh99txcax3qqmuf4amkpq8683vvjgp25e3kgaw",
		}
	}(),
	func() *bech32Test {
		multiAddr := &axongo.MultiAddress{
			Addresses: []*axongo.AddressWithWeight{
				{
					Address: &axongo.Ed25519Address{0x52, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
					Weight:  1,
				},
				{
					Address: &axongo.Ed25519Address{0x53, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
					Weight:  1,
				},
				{
					Address: &axongo.Ed25519Address{0x54, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
					Weight:  1,
				},
				{
					Address: &axongo.AccountAddress{0x55, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
					Weight:  2,
				},
				{
					Address: &axongo.NFTAddress{0x56, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
					Weight:  3,
				},
				{
					Address: &axongo.AnchorAddress{0x57, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
					Weight:  3,
				},
			},
			Threshold: 2,
		}

		addr := &axongo.RestrictedAddress{
			Address:             multiAddr,
			AllowedCapabilities: axongo.AddressCapabilitiesBitMask{0x55},
		}

		return &bech32Test{
			name:    "Restricted Multi Address",
			network: axongo.PrefixTestnet,
			addr:    addr,
			targetAddr: &axongo.RestrictedAddress{
				Address:             axongo.NewMultiAddressReferenceFromMultiAddress(multiAddr),
				AllowedCapabilities: addr.AllowedCapabilities,
			},
			bech32: "rms1xq5vf6xg2ysksnazu9zfuaq93erznxpqjsus79yhkaz9xarpskm2ckqp25kthwqv",
		}
	}(),
}

func TestBech32(t *testing.T) {
	for _, tt := range bech32Tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.bech32, tt.addr.Bech32(tt.network))
		})
	}
}

func TestParseBech32(t *testing.T) {
	for _, tt := range bech32Tests {
		t.Run(tt.name, func(t *testing.T) {
			network, addr, err := axongo.ParseBech32(tt.bech32)
			assert.NoError(t, err)
			assert.Equal(t, tt.network, network)
			assert.Equal(t, tt.targetAddr.ID(), addr.ID(), "parsed bech32 address does not match the given target address: %s != %s", tt.targetAddr.Bech32(tt.network), addr.Bech32(tt.network))
		})
	}
}

func TestImplicitAccountCreationAddressCapabilities(t *testing.T) {
	address := axongo.ImplicitAccountCreationAddressFromPubKey(hiveEd25519.PublicKey(tpkg.Rand32ByteArray()).ToEd25519())
	require.False(t, address.CannotReceiveNativeTokens())
	require.False(t, address.CannotReceiveMana())
	require.True(t, address.CannotReceiveOutputsWithTimelockUnlockCondition())
	require.True(t, address.CannotReceiveOutputsWithExpirationUnlockCondition())
	require.True(t, address.CannotReceiveOutputsWithStorageDepositReturnUnlockCondition())
	require.True(t, address.CannotReceiveAccountOutputs())
	require.True(t, address.CannotReceiveAnchorOutputs())
	require.True(t, address.CannotReceiveNFTOutputs())
	require.True(t, address.CannotReceiveDelegationOutputs())
}

func assertRestrictedAddresses(t *testing.T, addresses []*axongo.RestrictedAddress) {
	t.Helper()

	for i, addr := range addresses {
		// fmt.Println(addr.Bech32(axongo.PrefixMainnet))

		j, err := tpkg.ZeroCostTestAPI.JSONEncode(addr)
		_ = j
		require.NoError(t, err)
		// fmt.Println(string(j))

		b, err := tpkg.ZeroCostTestAPI.Encode(addr)
		require.NoError(t, err)
		// fmt.Println(hexutil.Encode(b))

		addrChecks := []func() bool{
			addr.CannotReceiveNativeTokens,
			addr.CannotReceiveMana,
			addr.CannotReceiveOutputsWithTimelockUnlockCondition,
			addr.CannotReceiveOutputsWithExpirationUnlockCondition,
			addr.CannotReceiveOutputsWithStorageDepositReturnUnlockCondition,
			addr.CannotReceiveAccountOutputs,
			addr.CannotReceiveAnchorOutputs,
			addr.CannotReceiveNFTOutputs,
			addr.CannotReceiveDelegationOutputs,
		}

		require.Equal(t, addr.Size(), len(b))

		setBit := func(bit int) axongo.AddressCapabilitiesBitMask {
			return axongo.AddressCapabilitiesBitMask(axongo.BitMaskSetBit([]byte{}, uint(bit)))
		}

		setAllBits := func(bit int) axongo.AddressCapabilitiesBitMask {
			var bitMask []byte
			for i := 0; i < bit; i++ {
				bitMask = axongo.BitMaskSetBit(bitMask, uint(i))
			}

			return axongo.AddressCapabilitiesBitMask(bitMask)
		}

		capabilitiesCount := 9
		indexModuloTestAmount := (i % (capabilitiesCount + 2)) // + 2 because we also test the "all enabled" and "all disabled" capabilities bit mask
		addressCapabilitiesBytesSize := 2 + (indexModuloTestAmount / 8)

		switch indexModuloTestAmount {
		default:
			for checkIndex, check := range addrChecks {
				require.Equalf(t, indexModuloTestAmount != checkIndex, check(), "index: %d", i)
			}
			require.Equalf(t, setBit(indexModuloTestAmount), addr.AllowedCapabilitiesBitMask(), "index: %d", i)

			require.Equalf(t, addressCapabilitiesBytesSize, addr.AllowedCapabilitiesBitMask().Size(), "index: %d", i)

		case capabilitiesCount: // all capabilities enabled
			for _, check := range addrChecks {
				require.Falsef(t, check(), "index: %d", i)
			}
			require.Equalf(t, setAllBits(indexModuloTestAmount), addr.AllowedCapabilitiesBitMask(), "index: %d", i)
			require.Equalf(t, addressCapabilitiesBytesSize, addr.AllowedCapabilitiesBitMask().Size(), "index: %d", i)

		case capabilitiesCount + 1: // all capabilities disabled
			for _, check := range addrChecks {
				require.Truef(t, check(), "index: %d", i)
			}
			require.Equalf(t, axongo.AddressCapabilitiesBitMask{}, addr.AllowedCapabilitiesBitMask(), "index: %d", i)
			require.Equalf(t, 1, addr.AllowedCapabilitiesBitMask().Size(), "index: %d", i)
		}
	}
}

//nolint:dupl // we have a lot of similar tests
func TestRestrictedAddressCapabilities(t *testing.T) {
	edAddr := tpkg.RandEd25519Address()
	accountAddr := tpkg.RandAccountAddress()
	nftAddr := tpkg.RandNFTAddress()
	anchorAddr := tpkg.RandAnchorAddress()
	multiAddress := tpkg.RandMultiAddress()

	addresses := []*axongo.RestrictedAddress{
		axongo.RestrictedAddressWithCapabilities(edAddr, axongo.WithAddressCanReceiveNativeTokens(true)),
		axongo.RestrictedAddressWithCapabilities(edAddr, axongo.WithAddressCanReceiveMana(true)),
		axongo.RestrictedAddressWithCapabilities(edAddr, axongo.WithAddressCanReceiveOutputsWithTimelockUnlockCondition(true)),
		axongo.RestrictedAddressWithCapabilities(edAddr, axongo.WithAddressCanReceiveOutputsWithExpirationUnlockCondition(true)),
		axongo.RestrictedAddressWithCapabilities(edAddr, axongo.WithAddressCanReceiveOutputsWithStorageDepositReturnUnlockCondition(true)),
		axongo.RestrictedAddressWithCapabilities(edAddr, axongo.WithAddressCanReceiveAccountOutputs(true)),
		axongo.RestrictedAddressWithCapabilities(edAddr, axongo.WithAddressCanReceiveAnchorOutputs(true)),
		axongo.RestrictedAddressWithCapabilities(edAddr, axongo.WithAddressCanReceiveNFTOutputs(true)),
		axongo.RestrictedAddressWithCapabilities(edAddr, axongo.WithAddressCanReceiveDelegationOutputs(true)),
		axongo.RestrictedAddressWithCapabilities(edAddr, axongo.WithAddressCanReceiveAnything()),
		axongo.RestrictedAddressWithCapabilities(edAddr),

		axongo.RestrictedAddressWithCapabilities(accountAddr, axongo.WithAddressCanReceiveNativeTokens(true)),
		axongo.RestrictedAddressWithCapabilities(accountAddr, axongo.WithAddressCanReceiveMana(true)),
		axongo.RestrictedAddressWithCapabilities(accountAddr, axongo.WithAddressCanReceiveOutputsWithTimelockUnlockCondition(true)),
		axongo.RestrictedAddressWithCapabilities(accountAddr, axongo.WithAddressCanReceiveOutputsWithExpirationUnlockCondition(true)),
		axongo.RestrictedAddressWithCapabilities(accountAddr, axongo.WithAddressCanReceiveOutputsWithStorageDepositReturnUnlockCondition(true)),
		axongo.RestrictedAddressWithCapabilities(accountAddr, axongo.WithAddressCanReceiveAccountOutputs(true)),
		axongo.RestrictedAddressWithCapabilities(accountAddr, axongo.WithAddressCanReceiveAnchorOutputs(true)),
		axongo.RestrictedAddressWithCapabilities(accountAddr, axongo.WithAddressCanReceiveNFTOutputs(true)),
		axongo.RestrictedAddressWithCapabilities(accountAddr, axongo.WithAddressCanReceiveDelegationOutputs(true)),
		axongo.RestrictedAddressWithCapabilities(accountAddr, axongo.WithAddressCanReceiveAnything()),
		axongo.RestrictedAddressWithCapabilities(accountAddr),

		axongo.RestrictedAddressWithCapabilities(nftAddr, axongo.WithAddressCanReceiveNativeTokens(true)),
		axongo.RestrictedAddressWithCapabilities(nftAddr, axongo.WithAddressCanReceiveMana(true)),
		axongo.RestrictedAddressWithCapabilities(nftAddr, axongo.WithAddressCanReceiveOutputsWithTimelockUnlockCondition(true)),
		axongo.RestrictedAddressWithCapabilities(nftAddr, axongo.WithAddressCanReceiveOutputsWithExpirationUnlockCondition(true)),
		axongo.RestrictedAddressWithCapabilities(nftAddr, axongo.WithAddressCanReceiveOutputsWithStorageDepositReturnUnlockCondition(true)),
		axongo.RestrictedAddressWithCapabilities(nftAddr, axongo.WithAddressCanReceiveAccountOutputs(true)),
		axongo.RestrictedAddressWithCapabilities(nftAddr, axongo.WithAddressCanReceiveAnchorOutputs(true)),
		axongo.RestrictedAddressWithCapabilities(nftAddr, axongo.WithAddressCanReceiveNFTOutputs(true)),
		axongo.RestrictedAddressWithCapabilities(nftAddr, axongo.WithAddressCanReceiveDelegationOutputs(true)),
		axongo.RestrictedAddressWithCapabilities(nftAddr, axongo.WithAddressCanReceiveAnything()),
		axongo.RestrictedAddressWithCapabilities(nftAddr),

		axongo.RestrictedAddressWithCapabilities(anchorAddr, axongo.WithAddressCanReceiveNativeTokens(true)),
		axongo.RestrictedAddressWithCapabilities(anchorAddr, axongo.WithAddressCanReceiveMana(true)),
		axongo.RestrictedAddressWithCapabilities(anchorAddr, axongo.WithAddressCanReceiveOutputsWithTimelockUnlockCondition(true)),
		axongo.RestrictedAddressWithCapabilities(anchorAddr, axongo.WithAddressCanReceiveOutputsWithExpirationUnlockCondition(true)),
		axongo.RestrictedAddressWithCapabilities(anchorAddr, axongo.WithAddressCanReceiveOutputsWithStorageDepositReturnUnlockCondition(true)),
		axongo.RestrictedAddressWithCapabilities(anchorAddr, axongo.WithAddressCanReceiveAccountOutputs(true)),
		axongo.RestrictedAddressWithCapabilities(anchorAddr, axongo.WithAddressCanReceiveAnchorOutputs(true)),
		axongo.RestrictedAddressWithCapabilities(anchorAddr, axongo.WithAddressCanReceiveNFTOutputs(true)),
		axongo.RestrictedAddressWithCapabilities(anchorAddr, axongo.WithAddressCanReceiveDelegationOutputs(true)),
		axongo.RestrictedAddressWithCapabilities(anchorAddr, axongo.WithAddressCanReceiveAnything()),
		axongo.RestrictedAddressWithCapabilities(anchorAddr),

		axongo.RestrictedAddressWithCapabilities(multiAddress, axongo.WithAddressCanReceiveNativeTokens(true)),
		axongo.RestrictedAddressWithCapabilities(multiAddress, axongo.WithAddressCanReceiveMana(true)),
		axongo.RestrictedAddressWithCapabilities(multiAddress, axongo.WithAddressCanReceiveOutputsWithTimelockUnlockCondition(true)),
		axongo.RestrictedAddressWithCapabilities(multiAddress, axongo.WithAddressCanReceiveOutputsWithExpirationUnlockCondition(true)),
		axongo.RestrictedAddressWithCapabilities(multiAddress, axongo.WithAddressCanReceiveOutputsWithStorageDepositReturnUnlockCondition(true)),
		axongo.RestrictedAddressWithCapabilities(multiAddress, axongo.WithAddressCanReceiveAccountOutputs(true)),
		axongo.RestrictedAddressWithCapabilities(multiAddress, axongo.WithAddressCanReceiveAnchorOutputs(true)),
		axongo.RestrictedAddressWithCapabilities(multiAddress, axongo.WithAddressCanReceiveNFTOutputs(true)),
		axongo.RestrictedAddressWithCapabilities(multiAddress, axongo.WithAddressCanReceiveDelegationOutputs(true)),
		axongo.RestrictedAddressWithCapabilities(multiAddress, axongo.WithAddressCanReceiveAnything()),
		axongo.RestrictedAddressWithCapabilities(multiAddress),
	}

	assertRestrictedAddresses(t, addresses)
}

//nolint:dupl // we have a lot of similar tests
func TestRestrictedAddressCapabilitiesBitMask(t *testing.T) {

	type test struct {
		name    string
		addr    *axongo.RestrictedAddress
		wantErr error
	}

	tests := []*test{
		{
			name: "ok - no trailing zero bytes",
			addr: &axongo.RestrictedAddress{
				Address:             tpkg.RandEd25519Address(),
				AllowedCapabilities: axongo.AddressCapabilitiesBitMask{0x01, 0x02},
			},
			wantErr: nil,
		},
		{
			name: "ok - empty capabilities",
			addr: &axongo.RestrictedAddress{
				Address:             tpkg.RandEd25519Address(),
				AllowedCapabilities: axongo.AddressCapabilitiesBitMask{},
			},
			wantErr: nil,
		},
		{
			name: "fail - trailing zero bytes",
			addr: &axongo.RestrictedAddress{
				Address:             tpkg.RandEd25519Address(),
				AllowedCapabilities: axongo.AddressCapabilitiesBitMask{0x01, 0x00},
			},
			wantErr: axongo.ErrBitmaskTrailingZeroBytes,
		},
		{
			name: "fail - single zero bytes",
			addr: &axongo.RestrictedAddress{
				Address:             tpkg.RandEd25519Address(),
				AllowedCapabilities: axongo.AddressCapabilitiesBitMask{0x00},
			},
			wantErr: axongo.ErrBitmaskTrailingZeroBytes,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := tpkg.ZeroCostTestAPI.Encode(test.addr, serix.WithValidation())
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)

				return
			}
			require.NoError(t, err)
		})
	}
}

type outputsSyntacticalValidationTest struct {
	// the name of the testcase
	name string
	// the amount of randomly created ed25519 addresses with private keys
	ed25519AddrCnt int
	// used to create outputs for the test
	outputsFunc func(ed25519Addresses []axongo.Address) axongo.TxEssenceOutputs
	// expected error during serialization of the transaction
	wantErr error
}

func runOutputsSyntacticalValidationTest(t *testing.T, testAPI axongo.API, test *outputsSyntacticalValidationTest) {
	t.Helper()

	t.Run(test.name, func(t *testing.T) {
		// generate random ed25519 addresses
		ed25519Addresses, _ := tpkg.RandEd25519IdentitiesSortedByAddress(test.ed25519AddrCnt)

		_, err := testAPI.Encode(test.outputsFunc(ed25519Addresses), serix.WithValidation())
		if test.wantErr != nil {
			require.ErrorIs(t, err, test.wantErr)
			return
		}
		require.NoError(t, err)
	})
}

func TestRestrictedAddressSyntacticalValidation(t *testing.T) {

	defaultAmount := OneIOTA

	tests := []*outputsSyntacticalValidationTest{
		// ok - Valid address types nested inside of a RestrictedAddress
		func() *outputsSyntacticalValidationTest {
			return &outputsSyntacticalValidationTest{
				name:           "ok - Valid address types nested inside of a RestrictedAddress",
				ed25519AddrCnt: 2,
				outputsFunc: func(ed25519Addresses []axongo.Address) axongo.TxEssenceOutputs {
					return axongo.TxEssenceOutputs{
						&axongo.BasicOutput{
							Amount: defaultAmount,
							UnlockConditions: axongo.BasicOutputUnlockConditions{
								&axongo.AddressUnlockCondition{Address: &axongo.RestrictedAddress{
									Address:             ed25519Addresses[0],
									AllowedCapabilities: axongo.AddressCapabilitiesBitMask{},
								}},
							},
						},
						&axongo.BasicOutput{
							Amount: defaultAmount,
							UnlockConditions: axongo.BasicOutputUnlockConditions{
								&axongo.AddressUnlockCondition{Address: &axongo.RestrictedAddress{
									Address:             &axongo.AccountAddress{},
									AllowedCapabilities: axongo.AddressCapabilitiesBitMask{},
								}},
							},
						},
						&axongo.BasicOutput{
							Amount: defaultAmount,
							UnlockConditions: axongo.BasicOutputUnlockConditions{
								&axongo.AddressUnlockCondition{Address: &axongo.RestrictedAddress{
									Address:             &axongo.NFTAddress{},
									AllowedCapabilities: axongo.AddressCapabilitiesBitMask{},
								}},
							},
						},
						&axongo.BasicOutput{
							Amount: defaultAmount,
							UnlockConditions: axongo.BasicOutputUnlockConditions{
								&axongo.AddressUnlockCondition{Address: &axongo.RestrictedAddress{
									Address:             &axongo.AnchorAddress{},
									AllowedCapabilities: axongo.AddressCapabilitiesBitMask{},
								}},
							},
						},
						&axongo.BasicOutput{
							Amount: defaultAmount,
							UnlockConditions: axongo.BasicOutputUnlockConditions{
								&axongo.AddressUnlockCondition{Address: &axongo.RestrictedAddress{
									Address: &axongo.MultiAddress{
										Addresses: []*axongo.AddressWithWeight{
											{
												Address: ed25519Addresses[0],
												Weight:  1,
											},
											{
												Address: ed25519Addresses[1],
												Weight:  1,
											},
										},
										Threshold: 2,
									},
									AllowedCapabilities: axongo.AddressCapabilitiesBitMask{},
								}},
							},
						},
					}
				},
				wantErr: nil,
			}
		}(),

		// fail - ImplicitAccountCreationAddress nested inside of a RestrictedAddress
		func() *outputsSyntacticalValidationTest {
			return &outputsSyntacticalValidationTest{
				name:           "fail - ImplicitAccountCreationAddress nested inside of a RestrictedAddress",
				ed25519AddrCnt: 0,
				outputsFunc: func(ed25519Addresses []axongo.Address) axongo.TxEssenceOutputs {
					return axongo.TxEssenceOutputs{
						&axongo.BasicOutput{
							Amount: defaultAmount,
							UnlockConditions: axongo.BasicOutputUnlockConditions{
								&axongo.AddressUnlockCondition{Address: &axongo.RestrictedAddress{
									Address: &axongo.ImplicitAccountCreationAddress{},
								}},
							},
						},
					}
				},
				wantErr: axongo.ErrInvalidNestedAddressType,
			}
		}(),

		// fail - RestrictedAddress nested inside of a RestrictedAddress
		func() *outputsSyntacticalValidationTest {
			return &outputsSyntacticalValidationTest{
				name:           "fail - RestrictedAddress nested inside of a RestrictedAddress",
				ed25519AddrCnt: 1,
				outputsFunc: func(ed25519Addresses []axongo.Address) axongo.TxEssenceOutputs {
					return axongo.TxEssenceOutputs{
						&axongo.BasicOutput{
							Amount: defaultAmount,
							UnlockConditions: axongo.BasicOutputUnlockConditions{
								&axongo.AddressUnlockCondition{Address: &axongo.RestrictedAddress{
									Address: &axongo.RestrictedAddress{
										Address:             ed25519Addresses[0],
										AllowedCapabilities: axongo.AddressCapabilitiesBitMask{},
									},
								}},
							},
						},
					}
				},
				wantErr: axongo.ErrInvalidNestedAddressType,
			}
		}(),
	}

	testAPI := tpkg.ZeroCostTestAPI

	for _, tt := range tests {
		runOutputsSyntacticalValidationTest(t, testAPI, tt)
	}
}

func TestMultiAddressSyntacticalValidation(t *testing.T) {

	defaultAmount := OneIOTA

	tests := []*outputsSyntacticalValidationTest{
		// fail - threshold > cumulativeWeight
		func() *outputsSyntacticalValidationTest {
			return &outputsSyntacticalValidationTest{
				name:           "fail - threshold > cumulativeWeight",
				ed25519AddrCnt: 2,
				outputsFunc: func(ed25519Addresses []axongo.Address) axongo.TxEssenceOutputs {
					return axongo.TxEssenceOutputs{
						&axongo.BasicOutput{
							Amount: defaultAmount,
							UnlockConditions: axongo.BasicOutputUnlockConditions{
								&axongo.AddressUnlockCondition{Address: &axongo.MultiAddress{
									Addresses: []*axongo.AddressWithWeight{
										{
											Address: ed25519Addresses[0],
											Weight:  1,
										},
										{
											Address: ed25519Addresses[1],
											Weight:  1,
										},
									},
									Threshold: 3,
								}},
							},
						},
					}
				},
				wantErr: axongo.ErrMultiAddressThresholdInvalid,
			}
		}(),

		// fail - threshold < 1
		func() *outputsSyntacticalValidationTest {
			return &outputsSyntacticalValidationTest{
				name:           "fail - threshold < 1",
				ed25519AddrCnt: 1,
				outputsFunc: func(ed25519Addresses []axongo.Address) axongo.TxEssenceOutputs {
					return axongo.TxEssenceOutputs{
						&axongo.BasicOutput{
							Amount: defaultAmount,
							UnlockConditions: axongo.BasicOutputUnlockConditions{
								&axongo.AddressUnlockCondition{Address: &axongo.MultiAddress{
									Addresses: []*axongo.AddressWithWeight{
										{
											Address: ed25519Addresses[0],
											Weight:  1,
										},
									},
									Threshold: 0,
								}},
							},
						},
					}
				},
				wantErr: axongo.ErrMultiAddressThresholdInvalid,
			}
		}(),

		// fail - address weight == 0
		func() *outputsSyntacticalValidationTest {
			return &outputsSyntacticalValidationTest{
				name:           "fail - address weight == 0",
				ed25519AddrCnt: 2,
				outputsFunc: func(ed25519Addresses []axongo.Address) axongo.TxEssenceOutputs {
					return axongo.TxEssenceOutputs{
						&axongo.BasicOutput{
							Amount: defaultAmount,
							UnlockConditions: axongo.BasicOutputUnlockConditions{
								&axongo.AddressUnlockCondition{Address: &axongo.MultiAddress{
									Addresses: []*axongo.AddressWithWeight{
										{
											Address: ed25519Addresses[0],
											Weight:  0,
										},
										{
											Address: ed25519Addresses[1],
											Weight:  1,
										},
									},
									Threshold: 1,
								}},
							},
						},
					}
				},
				wantErr: axongo.ErrMultiAddressWeightInvalid,
			}
		}(),

		// fail - empty MultiAddress
		func() *outputsSyntacticalValidationTest {
			return &outputsSyntacticalValidationTest{
				name:           "fail - empty MultiAddress",
				ed25519AddrCnt: 2,
				outputsFunc: func(ed25519Addresses []axongo.Address) axongo.TxEssenceOutputs {
					return axongo.TxEssenceOutputs{
						&axongo.BasicOutput{
							Amount: defaultAmount,
							UnlockConditions: axongo.BasicOutputUnlockConditions{
								&axongo.AddressUnlockCondition{Address: &axongo.MultiAddress{
									Addresses: []*axongo.AddressWithWeight{},
									Threshold: 1,
								}},
							},
						},
					}
				},
				wantErr: axongo.ErrMultiAddressThresholdInvalid,
			}
		}(),

		// fail - MultiAddress limit exceeded
		func() *outputsSyntacticalValidationTest {
			return &outputsSyntacticalValidationTest{
				name:           "fail - MultiAddress limit exceeded",
				ed25519AddrCnt: 13,
				outputsFunc: func(ed25519Addresses []axongo.Address) axongo.TxEssenceOutputs {
					return axongo.TxEssenceOutputs{
						&axongo.BasicOutput{
							Amount: defaultAmount,
							UnlockConditions: axongo.BasicOutputUnlockConditions{
								&axongo.AddressUnlockCondition{Address: &axongo.MultiAddress{
									Addresses: []*axongo.AddressWithWeight{
										{Address: ed25519Addresses[2], Weight: 1},
										{Address: ed25519Addresses[3], Weight: 1},
										{Address: ed25519Addresses[4], Weight: 1},
										{Address: ed25519Addresses[5], Weight: 1},
										{Address: ed25519Addresses[6], Weight: 1},
										{Address: ed25519Addresses[7], Weight: 1},
										{Address: ed25519Addresses[8], Weight: 1},
										{Address: ed25519Addresses[9], Weight: 1},
										{Address: ed25519Addresses[10], Weight: 1},
										{Address: ed25519Addresses[11], Weight: 1},
										{Address: ed25519Addresses[12], Weight: 1},
									},
									Threshold: 11,
								}},
							},
						},
					}
				},
				wantErr: serializer.ErrArrayValidationMaxElementsExceeded,
			}
		}(),

		// fail - the binary encoding of all addresses inside a MultiAddress need to be unique
		func() *outputsSyntacticalValidationTest {
			return &outputsSyntacticalValidationTest{
				name:           "fail - the binary encoding of all addresses inside a MultiAddress need to be unique",
				ed25519AddrCnt: 1,
				outputsFunc: func(ed25519Addresses []axongo.Address) axongo.TxEssenceOutputs {
					return axongo.TxEssenceOutputs{
						&axongo.BasicOutput{
							Amount: defaultAmount,
							UnlockConditions: axongo.BasicOutputUnlockConditions{
								&axongo.AddressUnlockCondition{Address: &axongo.MultiAddress{
									Addresses: []*axongo.AddressWithWeight{
										// both have the same pubKeyHash
										{
											Address: &axongo.Ed25519Address{},
											Weight:  1,
										},
										{
											Address: &axongo.Ed25519Address{},
											Weight:  1,
										},
									},
									Threshold: 1,
								}},
							},
						},
					}
				},
				wantErr: axongo.ErrArrayValidationViolatesUniqueness,
			}
		}(),

		// fail - ImplicitAccountCreationAddress nested inside of a MultiAddress
		func() *outputsSyntacticalValidationTest {
			return &outputsSyntacticalValidationTest{
				name:           "fail - ImplicitAccountCreationAddress nested inside of a MultiAddress",
				ed25519AddrCnt: 1,
				outputsFunc: func(ed25519Addresses []axongo.Address) axongo.TxEssenceOutputs {
					return axongo.TxEssenceOutputs{
						&axongo.BasicOutput{
							Amount: defaultAmount,
							UnlockConditions: axongo.BasicOutputUnlockConditions{
								&axongo.AddressUnlockCondition{Address: &axongo.RestrictedAddress{
									Address: &axongo.MultiAddress{
										Addresses: []*axongo.AddressWithWeight{
											{
												Address: &axongo.ImplicitAccountCreationAddress{},
												Weight:  1,
											},
										},
										Threshold: 1,
									},
								}},
							},
						},
					}
				},
				wantErr: axongo.ErrInvalidNestedAddressType,
			}
		}(),

		// fail - MultiAddress nested inside of a MultiAddress
		func() *outputsSyntacticalValidationTest {
			return &outputsSyntacticalValidationTest{
				name:           "fail - MultiAddress nested inside of a MultiAddress",
				ed25519AddrCnt: 2,
				outputsFunc: func(ed25519Addresses []axongo.Address) axongo.TxEssenceOutputs {
					return axongo.TxEssenceOutputs{
						&axongo.BasicOutput{
							Amount: defaultAmount,
							UnlockConditions: axongo.BasicOutputUnlockConditions{
								&axongo.AddressUnlockCondition{Address: &axongo.MultiAddress{
									Addresses: []*axongo.AddressWithWeight{
										{
											Address: &axongo.MultiAddress{
												Addresses: axongo.AddressesWithWeight{
													{
														Address: ed25519Addresses[1],
														Weight:  1,
													},
												},
												Threshold: 1,
											},
											Weight: 1,
										},
									},
									Threshold: 1,
								}},
							},
						},
					}
				},
				wantErr: axongo.ErrInvalidNestedAddressType,
			}
		}(),

		// fail - RestrictedAddress nested inside of a MultiAddress
		func() *outputsSyntacticalValidationTest {
			return &outputsSyntacticalValidationTest{
				name:           "fail - RestrictedAddress nested inside of a MultiAddress",
				ed25519AddrCnt: 1,
				outputsFunc: func(ed25519Addresses []axongo.Address) axongo.TxEssenceOutputs {
					return axongo.TxEssenceOutputs{
						&axongo.BasicOutput{
							Amount: defaultAmount,
							UnlockConditions: axongo.BasicOutputUnlockConditions{
								&axongo.AddressUnlockCondition{Address: &axongo.MultiAddress{
									Addresses: []*axongo.AddressWithWeight{
										{
											Address: &axongo.RestrictedAddress{
												Address:             ed25519Addresses[0],
												AllowedCapabilities: axongo.AddressCapabilitiesBitMask{},
											},
											Weight: 1,
										},
									},
									Threshold: 1,
								}},
							},
						},
					}
				},
				wantErr: axongo.ErrInvalidNestedAddressType,
			}
		}(),
	}

	testAPI := tpkg.ZeroCostTestAPI

	for _, tt := range tests {
		runOutputsSyntacticalValidationTest(t, testAPI, tt)
	}
}

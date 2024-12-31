//nolint:dupl
package axongo_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/axonfibre/fibre.go/serializer/v2/serix"
	axongo "github.com/axonfibre/axon.go/v4"
	"github.com/axonfibre/axon.go/v4/tpkg"
)

func TestNativeTokenDeSerialization(t *testing.T) {
	ntIn := &axongo.NativeTokenFeature{
		ID:     tpkg.Rand38ByteArray(),
		Amount: new(big.Int).SetUint64(1000),
	}

	ntBytes, err := tpkg.ZeroCostTestAPI.Encode(ntIn, serix.WithValidation())
	require.NoError(t, err)

	ntOut := &axongo.NativeTokenFeature{}
	_, err = tpkg.ZeroCostTestAPI.Decode(ntBytes, ntOut, serix.WithValidation())
	require.NoError(t, err)

	require.EqualValues(t, ntIn, ntOut)
}

func TestNativeToken_SyntacticalValidation(t *testing.T) {
	nativeTokenFeature := tpkg.RandNativeTokenFeature()
	accountAddress, err := nativeTokenFeature.ID.AccountAddress()
	require.NoError(t, err)

	type test struct {
		name               string
		nativeTokenFeature *axongo.NativeTokenFeature
		wantErr            error
	}

	tests := []*test{
		{
			name:               "ok - NativeTokenFeature token ID == FoundryID",
			nativeTokenFeature: nativeTokenFeature,
			wantErr:            nil,
		},
		{
			name:               "fail - NativeTokenFeature token ID != FoundryID",
			nativeTokenFeature: tpkg.RandNativeTokenFeature(),
			wantErr:            axongo.ErrFoundryIDNativeTokenIDMismatch,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			foundryIn := &axongo.FoundryOutput{
				Amount:       1000,
				SerialNumber: nativeTokenFeature.ID.FoundrySerialNumber(),
				TokenScheme: &axongo.SimpleTokenScheme{
					MintedTokens:  big.NewInt(100),
					MeltedTokens:  big.NewInt(0),
					MaximumSupply: big.NewInt(100),
				},
				UnlockConditions: axongo.FoundryOutputUnlockConditions{
					&axongo.ImmutableAccountUnlockCondition{
						Address: accountAddress,
					},
				},
				Features: axongo.FoundryOutputFeatures{
					test.nativeTokenFeature,
				},
				ImmutableFeatures: axongo.FoundryOutputImmFeatures{},
			}

			foundryBytes, err := tpkg.ZeroCostTestAPI.Encode(foundryIn, serix.WithValidation())
			if err == nil {
				err = axongo.OutputsSyntacticalFoundry()(0, foundryIn)
			}
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)

			foundryOut := &axongo.FoundryOutput{}
			_, err = tpkg.ZeroCostTestAPI.Decode(foundryBytes, foundryOut, serix.WithValidation())
			require.NoError(t, err)

			require.True(t, foundryIn.Equal(foundryOut))
		})
	}
}

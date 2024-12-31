package builder_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	iotago "github.com/axonfibre/axon.go/v4"
	"github.com/axonfibre/axon.go/v4/builder"
	"github.com/axonfibre/axon.go/v4/tpkg"
)

func TestBasicOutputBuilder(t *testing.T) {
	var (
		targetAddr                          = tpkg.RandEd25519Address()
		amount             iotago.BaseToken = 1337
		nativeTokenFeature                  = tpkg.RandNativeTokenFeature()
		expirationTarget                    = tpkg.RandEd25519Address()
		metadataEntries                     = iotago.MetadataFeatureEntries{"data": []byte("123456")}
		slotTimeProvider                    = iotago.NewTimeProvider(0, time.Now().Unix(), 10, 10)
	)
	timelock := slotTimeProvider.SlotFromTime(time.Now().Add(5 * time.Minute))
	expiration := slotTimeProvider.SlotFromTime(time.Now().Add(10 * time.Minute))

	basicOutput, err := builder.NewBasicOutputBuilder(targetAddr, amount).
		NativeToken(nativeTokenFeature).
		Timelock(timelock).
		Expiration(expirationTarget, expiration).
		Metadata(metadataEntries).
		Build()
	require.NoError(t, err)

	require.Equal(t, &iotago.BasicOutput{
		Amount: 1337,
		UnlockConditions: iotago.BasicOutputUnlockConditions{
			&iotago.AddressUnlockCondition{Address: targetAddr},
			&iotago.TimelockUnlockCondition{Slot: timelock},
			&iotago.ExpirationUnlockCondition{ReturnAddress: expirationTarget, Slot: expiration},
		},
		Features: iotago.BasicOutputFeatures{
			&iotago.MetadataFeature{Entries: metadataEntries},
			nativeTokenFeature,
		},
	}, basicOutput)
}

func TestAccountOutputBuilder(t *testing.T) {
	var (
		addr                                = tpkg.RandEd25519Address()
		amount             iotago.BaseToken = 1337
		metadataEntries                     = iotago.MetadataFeatureEntries{"data": []byte("123456")}
		immMetadataEntries                  = iotago.MetadataFeatureEntries{"data": []byte("654321")}
		immIssuer                           = tpkg.RandEd25519Address()

		blockIssuerKey1    = iotago.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray())
		blockIssuerKey2    = iotago.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray())
		blockIssuerKey3    = iotago.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray())
		newBlockIssuerKey1 = iotago.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray())
		newBlockIssuerKey2 = iotago.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray())
	)

	accountOutput, err := builder.NewAccountOutputBuilder(addr, amount).
		Metadata(metadataEntries).
		Staking(amount, 1, 1000).
		BlockIssuer(iotago.NewBlockIssuerKeys(blockIssuerKey1, blockIssuerKey2, blockIssuerKey3), 100000).
		ImmutableMetadata(immMetadataEntries).
		ImmutableIssuer(immIssuer).
		FoundriesToGenerate(5).
		Build()
	require.NoError(t, err)

	expectedBlockIssuerKeys := iotago.NewBlockIssuerKeys(blockIssuerKey1, blockIssuerKey2, blockIssuerKey3)

	expected := &iotago.AccountOutput{
		Amount:         1337,
		FoundryCounter: 5,
		UnlockConditions: iotago.AccountOutputUnlockConditions{
			&iotago.AddressUnlockCondition{Address: addr},
		},
		Features: iotago.AccountOutputFeatures{
			&iotago.MetadataFeature{Entries: metadataEntries},
			&iotago.BlockIssuerFeature{
				BlockIssuerKeys: expectedBlockIssuerKeys,
				ExpirySlot:      100000,
			},
			&iotago.StakingFeature{
				StakedAmount: amount,
				FixedCost:    1,
				StartEpoch:   1000,
				EndEpoch:     iotago.MaxEpochIndex,
			},
		},
		ImmutableFeatures: iotago.AccountOutputImmFeatures{
			&iotago.IssuerFeature{Address: immIssuer},
			&iotago.MetadataFeature{Entries: immMetadataEntries},
		},
	}
	require.True(t, expected.Equal(accountOutput), "account output should be equal")

	const newAmount iotago.BaseToken = 7331
	//nolint:forcetypeassert // we can safely assume that this is an AccountOutput
	expectedCpy := expected.Clone().(*iotago.AccountOutput)
	expectedCpy.Amount = newAmount

	updatedOutput, err := builder.NewAccountOutputBuilderFromPrevious(accountOutput).
		Amount(newAmount).Build()
	require.NoError(t, err)
	require.Equal(t, expectedCpy, updatedOutput)

	updatedFeatures, err := builder.NewAccountOutputBuilderFromPrevious(accountOutput).
		BlockIssuerTransition().
		AddKeys(newBlockIssuerKey2, newBlockIssuerKey1).
		RemoveKey(blockIssuerKey3).
		RemoveKey(blockIssuerKey1).
		ExpirySlot(1500).
		Builder().
		StakingTransition().
		EndEpoch(2000).
		Builder().Build()
	require.NoError(t, err)

	expectedUpdatedBlockIssuerKeys := iotago.NewBlockIssuerKeys(blockIssuerKey2, newBlockIssuerKey1, newBlockIssuerKey2)

	expectedFeatures := &iotago.AccountOutput{
		Amount:         1337,
		FoundryCounter: 5,
		UnlockConditions: iotago.AccountOutputUnlockConditions{
			&iotago.AddressUnlockCondition{Address: addr},
		},
		Features: iotago.AccountOutputFeatures{
			&iotago.MetadataFeature{Entries: metadataEntries},
			&iotago.BlockIssuerFeature{
				BlockIssuerKeys: expectedUpdatedBlockIssuerKeys,
				ExpirySlot:      1500,
			},
			&iotago.StakingFeature{
				StakedAmount: amount,
				FixedCost:    1,
				StartEpoch:   1000,
				EndEpoch:     2000,
			},
		},
		ImmutableFeatures: iotago.AccountOutputImmFeatures{
			&iotago.IssuerFeature{Address: immIssuer},
			&iotago.MetadataFeature{Entries: immMetadataEntries},
		},
	}
	require.True(t, expectedFeatures.Equal(updatedFeatures), "features should be equal")
}

func TestAnchorOutputBuilder(t *testing.T) {
	var (
		stateCtrl                             = tpkg.RandEd25519Address()
		stateCtrlNew                          = tpkg.RandEd25519Address()
		gov                                   = tpkg.RandEd25519Address()
		amount               iotago.BaseToken = 1337
		stateMetadataEntries                  = iotago.StateMetadataFeatureEntries{"data": []byte("123456")}
		immMetadataEntries                    = iotago.MetadataFeatureEntries{"data": []byte("654321")}
		immIssuer                             = tpkg.RandEd25519Address()
	)

	anchorOutput, err := builder.NewAnchorOutputBuilder(stateCtrl, gov, amount).
		StateMetadata(stateMetadataEntries).
		ImmutableMetadata(immMetadataEntries).
		ImmutableIssuer(immIssuer).
		Build()
	require.NoError(t, err)

	expected := &iotago.AnchorOutput{
		Amount:     amount,
		StateIndex: 0,
		UnlockConditions: iotago.AnchorOutputUnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl},
			&iotago.GovernorAddressUnlockCondition{Address: gov},
		},
		Features: iotago.AnchorOutputFeatures{
			&iotago.StateMetadataFeature{Entries: stateMetadataEntries},
		},
		ImmutableFeatures: iotago.AnchorOutputImmFeatures{
			&iotago.IssuerFeature{Address: immIssuer},
			&iotago.MetadataFeature{Entries: immMetadataEntries},
		},
	}
	require.True(t, expected.Equal(anchorOutput), "anchor output should be equal")

	const newAmount iotago.BaseToken = 7331
	newStateMetadataEntries := iotago.StateMetadataFeatureEntries{"newData": []byte("newState")}

	//nolint:forcetypeassert // we can safely assume that this is an AnchorOutput
	expectedCpy := expected.Clone().(*iotago.AnchorOutput)
	expectedCpy.Amount = newAmount
	expectedCpy.StateIndex++
	expectedCpy.Features.Upsert(&iotago.StateMetadataFeature{Entries: newStateMetadataEntries})

	updatedOutput, err := builder.NewAnchorOutputBuilderFromPrevious(anchorOutput).StateTransition().
		Amount(newAmount).
		StateMetadata(newStateMetadataEntries).
		Builder().Build()
	require.NoError(t, err)
	require.Equal(t, expectedCpy, updatedOutput)
	require.True(t, expectedCpy.Equal(updatedOutput), "outputs should be equal")

	updatedOutput2, err := builder.NewAnchorOutputBuilderFromPrevious(anchorOutput).GovernanceTransition().
		StateController(stateCtrlNew).Builder().Build()
	require.NoError(t, err)

	expectedOutput2 := &iotago.AnchorOutput{
		Amount:     amount,
		StateIndex: 0,
		UnlockConditions: iotago.AnchorOutputUnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: stateCtrlNew},
			&iotago.GovernorAddressUnlockCondition{Address: gov},
		},
		Features: iotago.AnchorOutputFeatures{
			&iotago.StateMetadataFeature{Entries: stateMetadataEntries},
		},
		ImmutableFeatures: iotago.AnchorOutputImmFeatures{
			&iotago.IssuerFeature{Address: immIssuer},
			&iotago.MetadataFeature{Entries: immMetadataEntries},
		},
	}
	require.True(t, expectedOutput2.Equal(updatedOutput2), "outputs should be equal")
}

func TestDelegationOutputBuilder(t *testing.T) {
	var (
		address                           = tpkg.RandEd25519Address()
		updatedAddress                    = tpkg.RandEd25519Address()
		amount           iotago.BaseToken = 1337
		updatedAmount    iotago.BaseToken = 127
		validatorAddress                  = tpkg.RandAccountAddress()
		delegationID                      = tpkg.RandDelegationID()
	)

	delegationOutput, err := builder.NewDelegationOutputBuilder(validatorAddress, address, amount).
		DelegatedAmount(amount).
		StartEpoch(1000).
		Build()
	require.NoError(t, err)

	expected := &iotago.DelegationOutput{
		Amount:           1337,
		DelegatedAmount:  1337,
		DelegationID:     iotago.EmptyDelegationID(),
		ValidatorAddress: validatorAddress,
		StartEpoch:       1000,
		EndEpoch:         0,
		UnlockConditions: iotago.DelegationOutputUnlockConditions{
			&iotago.AddressUnlockCondition{Address: address},
		},
	}
	require.Equal(t, expected, delegationOutput)

	updatedOutput, err := builder.NewDelegationOutputBuilderFromPrevious(delegationOutput).
		DelegationID(delegationID).
		DelegatedAmount(updatedAmount).
		Amount(updatedAmount).
		EndEpoch(1500).
		Address(updatedAddress).
		Build()
	require.NoError(t, err)

	expectedOutput := &iotago.DelegationOutput{
		Amount:           127,
		DelegatedAmount:  127,
		ValidatorAddress: validatorAddress,
		DelegationID:     delegationID,
		StartEpoch:       1000,
		EndEpoch:         1500,
		UnlockConditions: iotago.DelegationOutputUnlockConditions{
			&iotago.AddressUnlockCondition{Address: updatedAddress},
		},
	}
	require.Equal(t, expectedOutput, updatedOutput)
}

func TestFoundryOutputBuilder(t *testing.T) {
	var (
		accountAddr                  = tpkg.RandAccountAddress()
		amount      iotago.BaseToken = 1337
		tokenScheme                  = &iotago.SimpleTokenScheme{
			MintedTokens:  big.NewInt(0),
			MeltedTokens:  big.NewInt(0),
			MaximumSupply: big.NewInt(1000),
		}
		nativeTokenFeature = tpkg.RandNativeTokenFeature()
		metadataEntries    = iotago.MetadataFeatureEntries{"data": []byte("123456")}
		immMetadataEntries = iotago.MetadataFeatureEntries{"data": []byte("654321")}
	)

	foundryOutput, err := builder.NewFoundryOutputBuilder(accountAddr, amount, 12345, tokenScheme).
		NativeToken(nativeTokenFeature).
		Metadata(metadataEntries).
		ImmutableMetadata(immMetadataEntries).
		Build()
	require.NoError(t, err)

	require.Equal(t, &iotago.FoundryOutput{
		Amount:       1337,
		SerialNumber: 12345,
		TokenScheme:  tokenScheme,
		UnlockConditions: iotago.FoundryOutputUnlockConditions{
			&iotago.ImmutableAccountUnlockCondition{Address: accountAddr},
		},
		Features: iotago.FoundryOutputFeatures{
			&iotago.MetadataFeature{Entries: metadataEntries},
			nativeTokenFeature,
		},
		ImmutableFeatures: iotago.FoundryOutputImmFeatures{
			&iotago.MetadataFeature{Entries: immMetadataEntries},
		},
	}, foundryOutput)
}

func TestNFTOutputBuilder(t *testing.T) {
	var (
		targetAddr                          = tpkg.RandAccountAddress()
		amount             iotago.BaseToken = 1337
		metadataEntries                     = iotago.MetadataFeatureEntries{"data": []byte("123456")}
		immMetadataEntries                  = iotago.MetadataFeatureEntries{"data": []byte("654321")}
	)

	nftOutput, err := builder.NewNFTOutputBuilder(targetAddr, amount).
		Metadata(metadataEntries).
		ImmutableMetadata(immMetadataEntries).
		Build()
	require.NoError(t, err)

	require.Equal(t, &iotago.NFTOutput{
		Amount: 1337,
		UnlockConditions: iotago.NFTOutputUnlockConditions{
			&iotago.AddressUnlockCondition{Address: targetAddr},
		},
		Features: iotago.NFTOutputFeatures{
			&iotago.MetadataFeature{Entries: metadataEntries},
		},
		ImmutableFeatures: iotago.NFTOutputImmFeatures{
			&iotago.MetadataFeature{Entries: immMetadataEntries},
		},
	}, nftOutput)
}

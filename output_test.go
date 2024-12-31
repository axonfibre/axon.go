package axongo_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	axongo "github.com/axonfibre/axon.go/v4"
	"github.com/axonfibre/axon.go/v4/tpkg"
	"github.com/axonfibre/axon.go/v4/tpkg/frameworks"
)

func TestOutputTypeString(t *testing.T) {
	tests := []struct {
		outputType       axongo.OutputType
		outputTypeString string
	}{
		{axongo.OutputBasic, "BasicOutput"},
		{axongo.OutputAccount, "AccountOutput"},
		{axongo.OutputAnchor, "AnchorOutput"},
		{axongo.OutputFoundry, "FoundryOutput"},
		{axongo.OutputNFT, "NFTOutput"},
		{axongo.OutputDelegation, "DelegationOutput"},
	}
	for _, tt := range tests {
		require.Equal(t, tt.outputType.String(), tt.outputTypeString)
	}
}

func TestOutputsDeSerialize(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name: "ok - BasicOutput",
			Source: &axongo.BasicOutput{
				Amount: 1337,
				Mana:   500,
				UnlockConditions: axongo.BasicOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&axongo.StorageDepositReturnUnlockCondition{
						ReturnAddress: tpkg.RandEd25519Address(),
						Amount:        1000,
					},
					&axongo.TimelockUnlockCondition{Slot: 1337},
					&axongo.ExpirationUnlockCondition{
						ReturnAddress: tpkg.RandEd25519Address(),
						Slot:          4000,
					},
				},
				Features: axongo.BasicOutputFeatures{
					&axongo.SenderFeature{Address: tpkg.RandEd25519Address()},
					&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": tpkg.RandBytes(100)}},
					&axongo.TagFeature{Tag: tpkg.RandBytes(32)},
					tpkg.RandNativeTokenFeature(),
				},
			},
			Target: &axongo.BasicOutput{},
		},
		{
			Name: "ok - AccountOutput",
			Source: &axongo.AccountOutput{
				Amount:         1337,
				Mana:           500,
				AccountID:      tpkg.RandAccountAddress().AccountID(),
				FoundryCounter: 1337,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: axongo.AccountOutputFeatures{
					&axongo.SenderFeature{Address: tpkg.RandEd25519Address()},
					&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": tpkg.RandBytes(100)}},
					&axongo.BlockIssuerFeature{
						ExpirySlot:      1337,
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(),
					},
					&axongo.StakingFeature{
						StakedAmount: 1337,
						FixedCost:    10,
						StartEpoch:   1,
						EndEpoch:     2,
					},
				},
				ImmutableFeatures: axongo.AccountOutputImmFeatures{
					&axongo.IssuerFeature{Address: tpkg.RandEd25519Address()},
					&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"immutable": tpkg.RandBytes(100)}},
				},
			},
			Target: &axongo.AccountOutput{},
		},
		{
			Name: "ok - AnchorOutput",
			Source: &axongo.AnchorOutput{
				Amount:     1337,
				Mana:       500,
				AnchorID:   tpkg.RandAnchorAddress().AnchorID(),
				StateIndex: 10,
				UnlockConditions: axongo.AnchorOutputUnlockConditions{
					&axongo.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&axongo.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: axongo.AnchorOutputFeatures{
					&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": tpkg.RandBytes(100)}},
					&axongo.StateMetadataFeature{Entries: axongo.StateMetadataFeatureEntries{"data": tpkg.RandBytes(100)}},
				},
				ImmutableFeatures: axongo.AnchorOutputImmFeatures{
					&axongo.IssuerFeature{Address: tpkg.RandEd25519Address()},
					&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": tpkg.RandBytes(100)}},
				},
			},
			Target: &axongo.AnchorOutput{},
		},
		{
			Name: "ok - FoundryOutput",
			Source: &axongo.FoundryOutput{
				Amount:       1337,
				SerialNumber: 0,
				TokenScheme: &axongo.SimpleTokenScheme{
					MintedTokens:  new(big.Int).SetUint64(100),
					MeltedTokens:  big.NewInt(50),
					MaximumSupply: new(big.Int).SetUint64(1000),
				},
				UnlockConditions: axongo.FoundryOutputUnlockConditions{
					&axongo.ImmutableAccountUnlockCondition{Address: tpkg.RandAccountAddress()},
				},
				Features: axongo.FoundryOutputFeatures{
					&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": tpkg.RandBytes(100)}},
					tpkg.RandNativeTokenFeature(),
				},
				ImmutableFeatures: axongo.FoundryOutputImmFeatures{
					&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"immutable": tpkg.RandBytes(100)}},
				},
			},
			Target: &axongo.FoundryOutput{},
		},
		{
			Name: "ok - NFTOutput",
			Source: &axongo.NFTOutput{
				Amount: 1337,
				Mana:   500,
				NFTID:  tpkg.Rand32ByteArray(),
				UnlockConditions: axongo.NFTOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&axongo.StorageDepositReturnUnlockCondition{
						ReturnAddress: tpkg.RandEd25519Address(),
						Amount:        1000,
					},
					&axongo.TimelockUnlockCondition{Slot: 1337},
					&axongo.ExpirationUnlockCondition{
						ReturnAddress: tpkg.RandEd25519Address(),
						Slot:          4000,
					},
				},
				Features: axongo.NFTOutputFeatures{
					&axongo.SenderFeature{Address: tpkg.RandEd25519Address()},
					&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": tpkg.RandBytes(100)}},
					&axongo.TagFeature{Tag: tpkg.RandBytes(32)},
				},
				ImmutableFeatures: axongo.NFTOutputImmFeatures{
					&axongo.IssuerFeature{Address: tpkg.RandEd25519Address()},
					&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"immutable": tpkg.RandBytes(10)}},
				},
			},
			Target: &axongo.NFTOutput{},
		},
		{
			Name: "ok - DelegationOutput",
			Source: &axongo.DelegationOutput{
				Amount:           1337,
				DelegatedAmount:  1337,
				DelegationID:     tpkg.Rand32ByteArray(),
				ValidatorAddress: tpkg.RandAccountAddress(),
				StartEpoch:       axongo.EpochIndex(32),
				EndEpoch:         axongo.EpochIndex(37),
				UnlockConditions: axongo.DelegationOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			Target: &axongo.DelegationOutput{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

func TestOutputsSyntacticalDepositAmount(t *testing.T) {
	protoParams := tpkg.IOTAMainnetV3TestProtocolParameters

	var minAmount axongo.BaseToken = 14100

	tests := []struct {
		name        string
		protoParams axongo.ProtocolParameters
		outputs     axongo.Outputs[axongo.Output]
		wantErr     error
	}{
		{
			name:        "ok",
			protoParams: tpkg.ZeroCostTestAPI.ProtocolParameters(),
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.BasicOutput{
					Amount:           protoParams.TokenSupply(),
					UnlockConditions: axongo.BasicOutputUnlockConditions{&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()}},
					Mana:             500,
				},
			},
			wantErr: nil,
		},
		{
			name:        "ok - storage deposit covered",
			protoParams: protoParams,
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.BasicOutput{
					Amount:           minAmount, // min amount
					UnlockConditions: axongo.BasicOutputUnlockConditions{&axongo.AddressUnlockCondition{Address: tpkg.RandAccountAddress()}},
				},
			},
			wantErr: nil,
		},
		{
			name:        "ok - storage deposit return",
			protoParams: protoParams,
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.BasicOutput{
					Amount: 100000,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
						&axongo.StorageDepositReturnUnlockCondition{
							ReturnAddress: tpkg.RandAccountAddress(),
							Amount:        minAmount, // min amount
						},
					},
					Mana: 500,
				},
			},
			wantErr: nil,
		},
		{
			name:        "fail - storage deposit return less than min storage deposit",
			protoParams: protoParams,
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.BasicOutput{
					Amount: 100000,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
						&axongo.StorageDepositReturnUnlockCondition{
							ReturnAddress: tpkg.RandAccountAddress(),
							Amount:        minAmount - 1, // off by 1
						},
					},
				},
			},
			wantErr: axongo.ErrStorageDepositLessThanMinReturnOutputStorageDeposit,
		},
		{
			name:        "fail - storage deposit more than target output deposit",
			protoParams: protoParams,
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.BasicOutput{
					Amount: OneIOTA,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
						&axongo.StorageDepositReturnUnlockCondition{
							ReturnAddress: tpkg.RandAccountAddress(),
							// off by one from the deposit
							Amount: OneIOTA + 1,
						},
					},
					Mana: 500,
				},
			},
			wantErr: axongo.ErrStorageDepositExceedsTargetOutputAmount,
		},
		{
			name:        "fail - storage deposit not covered",
			protoParams: protoParams,
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.BasicOutput{
					Amount: minAmount - 1,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
				},
			},
			wantErr: axongo.ErrStorageDepositNotCovered,
		},
		{
			name:        "fail - zero deposit",
			protoParams: tpkg.ZeroCostTestAPI.ProtocolParameters(),
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.BasicOutput{
					Amount: 0,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			wantErr: axongo.ErrAmountMustBeGreaterThanZero,
		},
		{
			name:        "fail - more than total supply on single output",
			protoParams: tpkg.ZeroCostTestAPI.ProtocolParameters(),
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.BasicOutput{
					Amount: protoParams.TokenSupply() + 1,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			wantErr: axongo.ErrOutputsSumExceedsTotalSupply,
		},
		{
			name:        "fail - sum more than total supply over multiple outputs",
			protoParams: tpkg.ZeroCostTestAPI.ProtocolParameters(),
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.BasicOutput{
					Amount: protoParams.TokenSupply() - 1,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				&axongo.BasicOutput{
					Amount: protoParams.TokenSupply() - 1,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			wantErr: axongo.ErrOutputsSumExceedsTotalSupply,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valFunc := axongo.OutputsSyntacticalDepositAmount(tt.protoParams, axongo.NewStorageScoreStructure(tt.protoParams.StorageScoreParameters()))
			var runErr error
			for index, output := range tt.outputs {
				if err := valFunc(index, output); err != nil {
					runErr = err
				}
			}
			fmt.Println(tt.name)
			require.ErrorIs(t, runErr, tt.wantErr, tt.name)
		})
	}
}

func TestOutputsSyntacticalExpirationAndTimelock(t *testing.T) {
	tests := []struct {
		name    string
		outputs axongo.TxEssenceOutputs
		wantErr error
	}{
		{
			name: "ok",
			outputs: axongo.TxEssenceOutputs{
				&axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						&axongo.ExpirationUnlockCondition{
							ReturnAddress: tpkg.RandEd25519Address(),
							Slot:          1337,
						},
					},
				},
				&axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						&axongo.TimelockUnlockCondition{
							Slot: 1337,
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - zero expiration time",
			outputs: axongo.TxEssenceOutputs{
				&axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						&axongo.ExpirationUnlockCondition{
							ReturnAddress: tpkg.RandEd25519Address(),
							Slot:          0,
						},
					},
				},
			},
			wantErr: axongo.ErrExpirationConditionZero,
		},
		{
			name: "fail - zero timelock time",
			outputs: axongo.TxEssenceOutputs{
				&axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						&axongo.TimelockUnlockCondition{
							Slot: 0,
						},
					},
				},
			},
			wantErr: axongo.ErrTimelockConditionZero,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valFunc := axongo.OutputsSyntacticalExpirationAndTimelock()
			var runErr error
			for index, output := range tt.outputs {
				if err := valFunc(index, output); err != nil {
					runErr = err
				}
			}
			require.ErrorIs(t, runErr, tt.wantErr)
		})
	}
}

func TestOutputsSyntacticalNativeTokensCount(t *testing.T) {
	tests := []struct {
		name    string
		outputs axongo.Outputs[axongo.Output]
		wantErr error
	}{
		{
			name: "ok",
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.BasicOutput{
					Amount: 1,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
					Features: axongo.BasicOutputFeatures{
						tpkg.RandNativeTokenFeature(),
					},
				},
				&axongo.BasicOutput{
					Amount: 1,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
					Features: axongo.BasicOutputFeatures{
						tpkg.RandNativeTokenFeature(),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - native token with zero amount",
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.BasicOutput{
					Amount: 1,
					Features: axongo.BasicOutputFeatures{
						&axongo.NativeTokenFeature{
							ID:     axongo.NativeTokenID{},
							Amount: big.NewInt(0),
						},
					},
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			wantErr: axongo.ErrNativeTokenAmountLessThanEqualZero,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valFunc := axongo.OutputsSyntacticalNativeTokens()
			var runErr error
			for index, output := range tt.outputs {
				if err := valFunc(index, output); err != nil {
					runErr = err
				}
			}
			require.ErrorIs(t, runErr, tt.wantErr)
		})
	}
}

func TestOutputsSyntacticalAccount(t *testing.T) {
	exampleBlockIssuerFeature := &axongo.BlockIssuerFeature{
		ExpirySlot:      3,
		BlockIssuerKeys: tpkg.RandBlockIssuerKeys(2),
	}

	tests := []struct {
		name    string
		outputs axongo.Outputs[axongo.Output]
		wantErr error
	}{
		{
			name: "ok - empty state",
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.AccountOutput{
					Amount:         OneIOTA,
					AccountID:      axongo.AccountID{},
					FoundryCounter: 0,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - non empty state",
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.AccountOutput{
					Amount:         OneIOTA,
					AccountID:      tpkg.Rand32ByteArray(),
					FoundryCounter: 1337,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - foundry counter non zero on empty account ID",
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.AccountOutput{
					Amount:         OneIOTA,
					AccountID:      axongo.AccountID{},
					FoundryCounter: 1,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
				},
			},
			wantErr: axongo.ErrAccountOutputNonEmptyState,
		},
		{
			name: "fail - account's unlock condition contains its own account address",
			outputs: axongo.Outputs[axongo.Output]{
				func() *axongo.AccountOutput {
					accountID := axongo.AccountID(tpkg.Rand32ByteArray())

					return &axongo.AccountOutput{
						Amount:         OneIOTA,
						AccountID:      accountID,
						FoundryCounter: 1337,
						UnlockConditions: axongo.AccountOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: accountID.ToAddress()},
						},
					}
				}(),
			},
			wantErr: axongo.ErrAccountOutputCyclicAddress,
		},
		{
			name: "ok - staked amount equal to amount",
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.AccountOutput{
					Amount:         OneIOTA,
					AccountID:      tpkg.Rand32ByteArray(),
					FoundryCounter: 1337,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
					Features: axongo.AccountOutputFeatures{
						exampleBlockIssuerFeature,
						&axongo.StakingFeature{StakedAmount: OneIOTA},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - staked amount less than amount",
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.AccountOutput{
					Amount:         OneIOTA + 1,
					AccountID:      tpkg.Rand32ByteArray(),
					FoundryCounter: 1337,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
					Features: axongo.AccountOutputFeatures{
						exampleBlockIssuerFeature,
						&axongo.StakingFeature{StakedAmount: OneIOTA},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - staked amount greater than amount",
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.AccountOutput{
					Amount:         OneIOTA,
					AccountID:      tpkg.Rand32ByteArray(),
					FoundryCounter: 1337,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
					Features: axongo.AccountOutputFeatures{
						exampleBlockIssuerFeature,
						&axongo.StakingFeature{StakedAmount: OneIOTA + 1},
					},
				},
			},
			wantErr: axongo.ErrAccountOutputAmountLessThanStakedAmount,
		},
		{
			name: "ok - staking feature present with block issuer feature",
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.AccountOutput{
					Amount:         OneIOTA,
					AccountID:      tpkg.Rand32ByteArray(),
					FoundryCounter: 1337,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
					Features: axongo.AccountOutputFeatures{
						exampleBlockIssuerFeature,
						&axongo.StakingFeature{StakedAmount: OneIOTA},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - staking feature present without block issuer feature",
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.AccountOutput{
					Amount:         OneIOTA,
					AccountID:      tpkg.Rand32ByteArray(),
					FoundryCounter: 1337,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.StakingFeature{StakedAmount: OneIOTA},
					},
				},
			},
			wantErr: axongo.ErrStakingBlockIssuerFeatureMissing,
		},
	}
	valFunc := axongo.OutputsSyntacticalAccount()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var runErr error
			for index, output := range tt.outputs {
				if err := valFunc(index, output); err != nil {
					runErr = err
				}
			}
			require.ErrorIs(t, runErr, tt.wantErr)
		})
	}
}

func TestOutputsSyntacticalAnchor(t *testing.T) {
	tests := []struct {
		name    string
		outputs axongo.Outputs[axongo.Output]
		wantErr error
	}{
		{
			name: "ok - empty state",
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.AnchorOutput{
					Amount:     OneIOTA,
					AnchorID:   axongo.AnchorID{},
					StateIndex: 0,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: tpkg.RandAnchorAddress()},
						&axongo.GovernorAddressUnlockCondition{Address: tpkg.RandAnchorAddress()},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - non empty state",
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.AnchorOutput{
					Amount:     OneIOTA,
					AnchorID:   tpkg.Rand32ByteArray(),
					StateIndex: 10,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: tpkg.RandAnchorAddress()},
						&axongo.GovernorAddressUnlockCondition{Address: tpkg.RandAnchorAddress()},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - state index non zero on empty anchor ID",
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.AnchorOutput{
					Amount:     OneIOTA,
					AnchorID:   axongo.AnchorID{},
					StateIndex: 1,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: tpkg.RandAnchorAddress()},
						&axongo.GovernorAddressUnlockCondition{Address: tpkg.RandAnchorAddress()},
					},
				},
			},
			wantErr: axongo.ErrAnchorOutputNonEmptyState,
		},
		{
			name: "fail - anchors's state controller address unlock condition contains its own anchor address",
			outputs: axongo.Outputs[axongo.Output]{
				func() *axongo.AnchorOutput {
					anchorID := axongo.AnchorID(tpkg.Rand32ByteArray())

					return &axongo.AnchorOutput{
						Amount:     OneIOTA,
						AnchorID:   anchorID,
						StateIndex: 10,
						UnlockConditions: axongo.AnchorOutputUnlockConditions{
							&axongo.StateControllerAddressUnlockCondition{Address: anchorID.ToAddress()},
							&axongo.GovernorAddressUnlockCondition{Address: tpkg.RandAnchorAddress()},
						},
					}
				}(),
			},
			wantErr: axongo.ErrAnchorOutputCyclicAddress,
		},
		{
			name: "fail - anchors's governor address unlock condition contains its own anchor address",
			outputs: axongo.Outputs[axongo.Output]{
				func() *axongo.AnchorOutput {
					anchorID := axongo.AnchorID(tpkg.Rand32ByteArray())

					return &axongo.AnchorOutput{
						Amount:     OneIOTA,
						AnchorID:   anchorID,
						StateIndex: 10,
						UnlockConditions: axongo.AnchorOutputUnlockConditions{
							&axongo.StateControllerAddressUnlockCondition{Address: tpkg.RandAnchorAddress()},
							&axongo.GovernorAddressUnlockCondition{Address: anchorID.ToAddress()},
						},
					}
				}(),
			},
			wantErr: axongo.ErrAnchorOutputCyclicAddress,
		},
	}
	valFunc := axongo.OutputsSyntacticalAnchor()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var runErr error
			for index, output := range tt.outputs {
				if err := valFunc(index, output); err != nil {
					runErr = err
				}
			}
			require.ErrorIs(t, runErr, tt.wantErr)
		})
	}
}

func TestOutputsSyntacticalFoundry(t *testing.T) {
	tests := []struct {
		name    string
		outputs axongo.Outputs[axongo.Output]
		wantErr error
	}{
		{
			name: "ok",
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.FoundryOutput{
					Amount:       1337,
					SerialNumber: 5,
					TokenScheme: &axongo.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(5),
						MeltedTokens:  big.NewInt(2),
						MaximumSupply: new(big.Int).SetUint64(10),
					},
					UnlockConditions: axongo.FoundryOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
					Features: nil,
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - minted and max supply same",
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.FoundryOutput{
					Amount:       1337,
					SerialNumber: 5,
					TokenScheme: &axongo.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(10),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(10),
					},
					UnlockConditions: axongo.FoundryOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
					Features: nil,
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - invalid maximum supply",
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.FoundryOutput{
					Amount:       1337,
					SerialNumber: 5,
					TokenScheme: &axongo.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(5),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(0),
					},
					UnlockConditions: axongo.FoundryOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
					Features: nil,
				},
			},
			wantErr: axongo.ErrSimpleTokenSchemeInvalidMaximumSupply,
		},
		{
			name: "fail - minted less than melted",
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.FoundryOutput{
					Amount:       1337,
					SerialNumber: 5,
					TokenScheme: &axongo.SimpleTokenScheme{
						MintedTokens:  big.NewInt(5),
						MeltedTokens:  big.NewInt(10),
						MaximumSupply: new(big.Int).SetUint64(100),
					},
					UnlockConditions: axongo.FoundryOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
					Features: nil,
				},
			},
			wantErr: axongo.ErrSimpleTokenSchemeInvalidMintedMeltedTokens,
		},
		{
			name: "fail - minted melted delta is bigger than maximum supply",
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.FoundryOutput{
					Amount:       1337,
					SerialNumber: 5,
					TokenScheme: &axongo.SimpleTokenScheme{
						MintedTokens:  big.NewInt(50),
						MeltedTokens:  big.NewInt(20),
						MaximumSupply: new(big.Int).SetUint64(10),
					},
					UnlockConditions: axongo.FoundryOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandAccountAddress()},
					},
					Features: nil,
				},
			},
			wantErr: axongo.ErrSimpleTokenSchemeInvalidMintedMeltedTokens,
		},
	}
	valFunc := axongo.OutputsSyntacticalFoundry()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var runErr error
			for index, output := range tt.outputs {
				if err := valFunc(index, output); err != nil {
					runErr = err
				}
			}
			require.ErrorIs(t, runErr, tt.wantErr)
		})
	}
}

func TestOutputsSyntacticalNFT(t *testing.T) {
	tests := []struct {
		name    string
		outputs axongo.Outputs[axongo.Output]
		wantErr error
	}{
		{
			name: "ok",
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.NFTOutput{
					Amount: OneIOTA,
					NFTID:  axongo.NFTID{},
					UnlockConditions: axongo.NFTOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
		},
		{
			name: "fail - NFT's address unlock condition contains its own NFT address",
			outputs: axongo.Outputs[axongo.Output]{
				func() *axongo.NFTOutput {
					nftID := axongo.NFTID(tpkg.Rand32ByteArray())

					return &axongo.NFTOutput{
						Amount: OneIOTA,
						NFTID:  nftID,
						UnlockConditions: axongo.NFTOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: nftID.ToAddress()},
						},
					}
				}(),
			},
			wantErr: axongo.ErrNFTOutputCyclicAddress,
		},
	}
	valFunc := axongo.OutputsSyntacticalNFT()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var runErr error
			for index, output := range tt.outputs {
				if err := valFunc(index, output); err != nil {
					runErr = err
				}
			}
			require.ErrorIs(t, runErr, tt.wantErr)
		})
	}
}

func TestOutputsSyntacticaDelegation(t *testing.T) {
	emptyAccountAddress := axongo.AccountAddress{}

	tests := []struct {
		name    string
		outputs axongo.Outputs[axongo.Output]
		wantErr error
	}{
		{
			name: "ok",
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.DelegationOutput{
					Amount:           OneIOTA,
					DelegatedAmount:  OneIOTA,
					DelegationID:     axongo.EmptyDelegationID(),
					ValidatorAddress: tpkg.RandAccountAddress(),
					UnlockConditions: axongo.DelegationOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
		},
		{
			name: "fail - Delegation Output contains empty validator address",
			outputs: axongo.Outputs[axongo.Output]{
				&axongo.DelegationOutput{
					Amount:           OneIOTA,
					DelegationID:     axongo.EmptyDelegationID(),
					ValidatorAddress: &emptyAccountAddress,
					UnlockConditions: axongo.DelegationOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			wantErr: axongo.ErrDelegationValidatorAddressEmpty,
		},
	}
	valFunc := axongo.OutputsSyntacticalDelegation()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var runErr error
			for index, output := range tt.outputs {
				if err := valFunc(index, output); err != nil {
					runErr = err
				}
			}
			require.ErrorIs(t, runErr, tt.wantErr)
		})
	}
}

func TestOwnerTransitionIndependentOutput_UnlockableBy(t *testing.T) {
	type test struct {
		name                string
		output              axongo.OwnerTransitionIndependentOutput
		targetAddr          axongo.Address
		commitmentInputTime axongo.SlotIndex
		minCommittableAge   axongo.SlotIndex
		maxCommittableAge   axongo.SlotIndex
		canUnlock           bool
	}
	tests := []*test{
		func() *test {
			receiverAddr := tpkg.RandEd25519Address()
			return &test{
				name: "can unlock - target is source (no unlock conditions)",
				output: &axongo.BasicOutput{
					Amount: OneIOTA,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: receiverAddr},
					},
				},
				targetAddr:          receiverAddr,
				commitmentInputTime: axongo.SlotIndex(0),
				minCommittableAge:   axongo.SlotIndex(0),
				maxCommittableAge:   axongo.SlotIndex(0),
				canUnlock:           true,
			}
		}(),
		func() *test {
			return &test{
				name: "can not unlock - target is not source (no timelocks or expiration unlock conditions)",
				output: &axongo.BasicOutput{
					Amount: OneIOTA,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				targetAddr:          tpkg.RandEd25519Address(),
				commitmentInputTime: axongo.SlotIndex(0),
				minCommittableAge:   axongo.SlotIndex(0),
				maxCommittableAge:   axongo.SlotIndex(0),
				canUnlock:           false,
			}
		}(),
		func() *test {
			receiverAddr := tpkg.RandEd25519Address()
			return &test{
				name: "expiration - receiver addr can unlock",
				output: &axongo.BasicOutput{
					Amount: OneIOTA,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: receiverAddr},
						&axongo.ExpirationUnlockCondition{
							ReturnAddress: tpkg.RandEd25519Address(),
							Slot:          26,
						},
					},
				},
				targetAddr:          receiverAddr,
				commitmentInputTime: axongo.SlotIndex(5),
				minCommittableAge:   axongo.SlotIndex(10),
				maxCommittableAge:   axongo.SlotIndex(20),
				canUnlock:           true,
			}
		}(),
		func() *test {
			receiverAddr := tpkg.RandEd25519Address()
			returnAddr := tpkg.RandEd25519Address()
			return &test{
				name: "expiration - receiver addr can not unlock",
				output: &axongo.BasicOutput{
					Amount: OneIOTA,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: receiverAddr},
						&axongo.ExpirationUnlockCondition{
							ReturnAddress: returnAddr,
							Slot:          25,
						},
					},
				},
				targetAddr:          receiverAddr,
				commitmentInputTime: axongo.SlotIndex(5),
				minCommittableAge:   axongo.SlotIndex(10),
				maxCommittableAge:   axongo.SlotIndex(20),
				canUnlock:           false,
			}
		}(),
		func() *test {
			receiverAddr := tpkg.RandEd25519Address()
			returnAddr := tpkg.RandEd25519Address()
			return &test{
				name: "expiration - return addr can unlock",
				output: &axongo.BasicOutput{
					Amount: OneIOTA,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: receiverAddr},
						&axongo.ExpirationUnlockCondition{
							ReturnAddress: returnAddr,
							Slot:          15,
						},
					},
				},
				targetAddr:          returnAddr,
				commitmentInputTime: axongo.SlotIndex(5),
				minCommittableAge:   axongo.SlotIndex(10),
				maxCommittableAge:   axongo.SlotIndex(20),
				canUnlock:           true,
			}
		}(),
		func() *test {
			receiverAddr := tpkg.RandEd25519Address()
			returnAddr := tpkg.RandEd25519Address()
			return &test{
				name: "expiration - return addr can not unlock",
				output: &axongo.BasicOutput{
					Amount: OneIOTA,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: receiverAddr},
						&axongo.ExpirationUnlockCondition{
							ReturnAddress: returnAddr,
							Slot:          16,
						},
					},
				},
				targetAddr:          returnAddr,
				commitmentInputTime: axongo.SlotIndex(5),
				minCommittableAge:   axongo.SlotIndex(10),
				maxCommittableAge:   axongo.SlotIndex(20),
				canUnlock:           false,
			}
		}(),
		func() *test {
			receiverAddr := tpkg.RandEd25519Address()
			return &test{
				name: "timelock - expired timelock is unlockable",
				output: &axongo.BasicOutput{
					Amount: OneIOTA,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: receiverAddr},
						&axongo.TimelockUnlockCondition{Slot: 15},
					},
				},
				targetAddr:          receiverAddr,
				commitmentInputTime: axongo.SlotIndex(5),
				minCommittableAge:   axongo.SlotIndex(10),
				maxCommittableAge:   axongo.SlotIndex(20),
				canUnlock:           true,
			}
		}(),
		func() *test {
			receiverAddr := tpkg.RandEd25519Address()
			return &test{
				name: "timelock - non-expired timelock is not unlockable",
				output: &axongo.BasicOutput{
					Amount: OneIOTA,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: receiverAddr},
						&axongo.TimelockUnlockCondition{Slot: 16},
					},
				},
				targetAddr:          receiverAddr,
				commitmentInputTime: axongo.SlotIndex(5),
				minCommittableAge:   axongo.SlotIndex(10),
				maxCommittableAge:   axongo.SlotIndex(20),
				canUnlock:           false,
			}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.canUnlock, tt.output.UnlockableBy(tt.targetAddr, tt.commitmentInputTime+tt.maxCommittableAge, tt.commitmentInputTime+tt.minCommittableAge))
		})
	}
}

func TestAnchorOutput_UnlockableBy(t *testing.T) {
	type test struct {
		name                 string
		current              axongo.OwnerTransitionDependentOutput
		next                 axongo.OwnerTransitionDependentOutput
		targetAddr           axongo.Address
		addrCanUnlockInstead axongo.Address
		commitmentInputTime  axongo.SlotIndex
		minCommittableAge    axongo.SlotIndex
		maxCommittableAge    axongo.SlotIndex
		wantErr              error
		canUnlock            bool
	}
	tests := []*test{
		func() *test {
			stateCtrl := tpkg.RandEd25519Address()
			govCtrl := tpkg.RandEd25519Address()

			return &test{
				name: "state ctrl can unlock - state index increase",
				current: &axongo.AnchorOutput{
					Amount:     OneIOTA,
					StateIndex: 0,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: stateCtrl},
						&axongo.GovernorAddressUnlockCondition{Address: govCtrl},
					},
				},
				next: &axongo.AnchorOutput{
					Amount:     OneIOTA,
					StateIndex: 1,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: stateCtrl},
						&axongo.GovernorAddressUnlockCondition{Address: govCtrl},
					},
				},
				targetAddr:          stateCtrl,
				commitmentInputTime: axongo.SlotIndex(0),
				minCommittableAge:   axongo.SlotIndex(0),
				maxCommittableAge:   axongo.SlotIndex(0),
				canUnlock:           true,
			}
		}(),
		func() *test {
			stateCtrl := tpkg.RandEd25519Address()
			govCtrl := tpkg.RandEd25519Address()

			return &test{
				name: "state ctrl can not unlock - state index same",
				current: &axongo.AnchorOutput{
					Amount:     OneIOTA,
					StateIndex: 0,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: stateCtrl},
						&axongo.GovernorAddressUnlockCondition{Address: govCtrl},
					},
				},
				next: &axongo.AnchorOutput{
					Amount:     OneIOTA,
					StateIndex: 0,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: stateCtrl},
						&axongo.GovernorAddressUnlockCondition{Address: govCtrl},
					},
				},
				targetAddr:           stateCtrl,
				addrCanUnlockInstead: govCtrl,
				commitmentInputTime:  axongo.SlotIndex(0),
				minCommittableAge:    axongo.SlotIndex(0),
				maxCommittableAge:    axongo.SlotIndex(0),
				canUnlock:            false,
			}
		}(),
		func() *test {
			stateCtrl := tpkg.RandEd25519Address()
			govCtrl := tpkg.RandEd25519Address()

			return &test{
				name: "state ctrl can not unlock - transition destroy",
				current: &axongo.AnchorOutput{
					Amount:     OneIOTA,
					StateIndex: 0,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: stateCtrl},
						&axongo.GovernorAddressUnlockCondition{Address: govCtrl},
					},
				},
				next:                 nil,
				targetAddr:           stateCtrl,
				addrCanUnlockInstead: govCtrl,
				commitmentInputTime:  axongo.SlotIndex(0),
				minCommittableAge:    axongo.SlotIndex(0),
				maxCommittableAge:    axongo.SlotIndex(0),
				canUnlock:            false,
			}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canUnlock, err := tt.current.UnlockableBy(tt.targetAddr, tt.next, tt.commitmentInputTime+tt.maxCommittableAge, tt.commitmentInputTime+tt.minCommittableAge)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)

				return
			}
			require.Equal(t, tt.canUnlock, canUnlock)
			if tt.addrCanUnlockInstead == nil {
				return
			}
			canUnlockInstead, err := tt.current.UnlockableBy(tt.addrCanUnlockInstead, tt.next, tt.commitmentInputTime+tt.maxCommittableAge, tt.commitmentInputTime+tt.minCommittableAge)
			require.NoError(t, err)
			require.True(t, canUnlockInstead)
		})
	}
}

func TestOutputsSyntacticDisallowedImplicitAccountCreationAddress(t *testing.T) {
	type test struct {
		name    string
		output  axongo.Output
		wantErr error
	}

	tests := []test{
		{
			name: "fail - Account Output contains Implicit Account Creation Address",
			output: &axongo.AccountOutput{
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandImplicitAccountCreationAddress()},
				},
			},
			wantErr: axongo.ErrImplicitAccountCreationAddressInInvalidOutput,
		},
		{
			name: "fail - Anchor Output contains Implicit Account Creation Address as State Controller",
			output: &axongo.AnchorOutput{
				UnlockConditions: axongo.AnchorOutputUnlockConditions{
					&axongo.StateControllerAddressUnlockCondition{Address: tpkg.RandImplicitAccountCreationAddress()},
					&axongo.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			wantErr: axongo.ErrImplicitAccountCreationAddressInInvalidOutput,
		},
		{
			name: "fail - Anchor Output contains Implicit Account Creation Address as Governor",
			output: &axongo.AnchorOutput{
				UnlockConditions: axongo.AnchorOutputUnlockConditions{
					&axongo.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&axongo.GovernorAddressUnlockCondition{Address: tpkg.RandImplicitAccountCreationAddress()},
				},
			},
			wantErr: axongo.ErrImplicitAccountCreationAddressInInvalidOutput,
		},
		{
			name: "fail - NFT Output contains Implicit Account Creation Address",
			output: &axongo.NFTOutput{
				UnlockConditions: axongo.NFTOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandImplicitAccountCreationAddress()},
				},
			},
			wantErr: axongo.ErrImplicitAccountCreationAddressInInvalidOutput,
		},
		{
			name: "fail - Delegation Output contains Implicit Account Creation Address",
			output: &axongo.DelegationOutput{
				Amount:           1337,
				DelegatedAmount:  1337,
				DelegationID:     tpkg.Rand32ByteArray(),
				ValidatorAddress: tpkg.RandAccountAddress(),
				StartEpoch:       axongo.EpochIndex(32),
				EndEpoch:         axongo.EpochIndex(37),
				UnlockConditions: axongo.DelegationOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandImplicitAccountCreationAddress()},
				},
			},
			wantErr: axongo.ErrImplicitAccountCreationAddressInInvalidOutput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			implicitAccountCreationAddressValidatorFunc := axongo.OutputsSyntacticalImplicitAccountCreationAddress()

			err := implicitAccountCreationAddressValidatorFunc(0, tt.output)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			}
		})
	}

}

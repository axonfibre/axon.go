//nolint:forcetypeassert,dupl,nlreturn
package nova_test

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	axongo "github.com/axonfibre/axon.go/v4"
	"github.com/axonfibre/axon.go/v4/tpkg"
	"github.com/axonfibre/axon.go/v4/vm"
)

type fieldMutations map[string]interface{}

//nolint:thelper
func copyObjectAndMutate(t *testing.T, source any, mutations fieldMutations) any {
	srcBytes, err := tpkg.ZeroCostTestAPI.Encode(source)
	require.NoError(t, err)

	ptrToCpyOfSrc := reflect.New(reflect.ValueOf(source).Elem().Type())

	cpySeri := ptrToCpyOfSrc.Interface()
	_, err = tpkg.ZeroCostTestAPI.Decode(srcBytes, cpySeri)
	require.NoError(t, err)

	for fieldName, newVal := range mutations {
		ptrToCpyOfSrc.Elem().FieldByName(fieldName).Set(reflect.ValueOf(newVal))
	}

	return cpySeri
}

func TestAccountOutput_ValidateStateTransition(t *testing.T) {
	exampleIssuer := tpkg.RandEd25519Address()
	exampleAccountID := tpkg.RandAccountAddress().AccountID()

	exampleAddress := tpkg.RandEd25519Address()

	exampleExistingFoundryOutput := &axongo.FoundryOutput{
		Amount:       100,
		SerialNumber: 5,
		TokenScheme: &axongo.SimpleTokenScheme{
			MintedTokens:  new(big.Int).SetInt64(1000),
			MeltedTokens:  big.NewInt(0),
			MaximumSupply: new(big.Int).SetInt64(10000),
		},
		UnlockConditions: axongo.FoundryOutputUnlockConditions{
			&axongo.ImmutableAccountUnlockCondition{Address: exampleAccountID.ToAddress().(*axongo.AccountAddress)},
		},
	}
	exampleExistingFoundryOutputFoundryID := exampleExistingFoundryOutput.MustFoundryID()

	currentEpoch := axongo.EpochIndex(20)
	currentSlot := tpkg.ZeroCostTestAPI.TimeProvider().EpochStart(currentEpoch)

	blockIssuerPubKey := axongo.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray())
	exampleBlockIssuerFeature := &axongo.BlockIssuerFeature{
		BlockIssuerKeys: axongo.NewBlockIssuerKeys(blockIssuerPubKey),
		ExpirySlot:      currentSlot + tpkg.ZeroCostTestAPI.ProtocolParameters().MaxCommittableAge(),
	}

	exampleBIC := map[axongo.AccountID]axongo.BlockIssuanceCredits{
		exampleAccountID: 100,
	}

	type test struct {
		name      string
		input     *vm.ChainOutputWithIDs
		next      *axongo.AccountOutput
		nextMut   map[string]fieldMutations
		transType axongo.ChainTransitionType
		svCtx     *vm.Params
		wantErr   error
	}

	tests := []*test{
		{
			name: "ok - genesis transition",
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: axongo.AccountID{},
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				ImmutableFeatures: axongo.AccountOutputImmFeatures{
					&axongo.IssuerFeature{Address: exampleIssuer},
				},
			},
			input:     nil,
			transType: axongo.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleIssuer.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - block issuer genesis transition",
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: axongo.AccountID{},
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				ImmutableFeatures: axongo.AccountOutputImmFeatures{
					&axongo.IssuerFeature{Address: exampleIssuer},
				},
				Features: axongo.AccountOutputFeatures{
					&axongo.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
						ExpirySlot:      1000,
					},
				},
			},
			input:     nil,
			transType: axongo.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: 900,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleIssuer.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: 900,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - block issuer genesis expiry too early",
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: axongo.AccountID{},
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				ImmutableFeatures: axongo.AccountOutputImmFeatures{
					&axongo.IssuerFeature{Address: exampleIssuer},
				},
				Features: axongo.AccountOutputFeatures{
					&axongo.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
						ExpirySlot:      1000,
					},
				},
			},
			input:     nil,
			transType: axongo.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: 10001,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleIssuer.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: 10001,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrBlockIssuerExpiryTooEarly,
		},
		{
			name: "fail - block issuer genesis expired but within MCA",
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: axongo.AccountID{},
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				ImmutableFeatures: axongo.AccountOutputImmFeatures{
					&axongo.IssuerFeature{Address: exampleIssuer},
				},
				Features: axongo.AccountOutputFeatures{
					&axongo.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
						ExpirySlot:      1000,
					},
				},
			},
			input:     nil,
			transType: axongo.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: 991,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleIssuer.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: 991,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrBlockIssuerExpiryTooEarly,
		},
		{
			name: "ok - staking genesis transition",
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: axongo.AccountID{},
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: axongo.AccountOutputFeatures{
					&axongo.StakingFeature{
						StakedAmount: 50,
						FixedCost:    5,
						StartEpoch:   currentEpoch,
						EndEpoch:     axongo.MaxEpochIndex,
					},
					exampleBlockIssuerFeature,
				},
			},
			input:     nil,
			transType: axongo.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: currentSlot,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleIssuer.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - staking genesis start epoch invalid",
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: axongo.AccountID{},
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: axongo.AccountOutputFeatures{
					&axongo.StakingFeature{
						StakedAmount: 50,
						FixedCost:    5,
						StartEpoch:   currentEpoch - 2,
						EndEpoch:     axongo.MaxEpochIndex,
					},
					exampleBlockIssuerFeature,
				},
			},
			input:     nil,
			transType: axongo.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: currentSlot,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleIssuer.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: axongo.ErrStakingStartEpochInvalid,
		},
		{
			name: "fail - staking genesis end epoch too early",
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: axongo.AccountID{},
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: axongo.AccountOutputFeatures{
					&axongo.StakingFeature{
						StakedAmount: 50,
						FixedCost:    5,
						StartEpoch:   currentEpoch,
						EndEpoch:     currentEpoch + tpkg.ZeroCostTestAPI.ProtocolParameters().StakingUnbondingPeriod() - 1,
					},
					exampleBlockIssuerFeature,
				},
			},
			input:     nil,
			transType: axongo.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: currentSlot,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleIssuer.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: axongo.ErrStakingEndEpochTooEarly,
		},
		{
			name: "ok - valid staking transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.StakingFeature{
							StakedAmount: 100,
							FixedCost:    50,
							StartEpoch:   currentEpoch,
							EndEpoch:     currentEpoch + 10000,
						},
						exampleBlockIssuerFeature,
					},
				},
			},
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: axongo.AccountOutputFeatures{
					&axongo.StakingFeature{
						StakedAmount: 100,
						FixedCost:    50,
						StartEpoch:   currentEpoch,
						EndEpoch:     currentEpoch + tpkg.ZeroCostTestAPI.ProtocolParameters().StakingUnbondingPeriod(),
					},
					exampleBlockIssuerFeature,
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: currentSlot,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleIssuer.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
				},
			},
		},
		{
			name: "ok - adding staking feature in account state transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						exampleBlockIssuerFeature,
					},
				},
			},
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: axongo.AccountOutputFeatures{
					&axongo.StakingFeature{
						StakedAmount: 100,
						FixedCost:    50,
						StartEpoch:   currentEpoch,
						EndEpoch:     currentEpoch + tpkg.ZeroCostTestAPI.ProtocolParameters().StakingUnbondingPeriod(),
					},
					exampleBlockIssuerFeature,
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: currentSlot,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleIssuer.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
				},
			},
		},
		{
			name: "fail - adding staking feature in account state transition with start epoch set too early",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						exampleBlockIssuerFeature,
					},
				},
			},
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: axongo.AccountOutputFeatures{
					&axongo.StakingFeature{
						StakedAmount: 100,
						FixedCost:    50,
						StartEpoch:   currentEpoch - 5,
						EndEpoch:     currentEpoch + tpkg.ZeroCostTestAPI.ProtocolParameters().StakingUnbondingPeriod(),
					},
					exampleBlockIssuerFeature,
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: currentSlot,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleIssuer.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: axongo.ErrStakingStartEpochInvalid,
		},
		{
			name: "fail - negative BIC during account state transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						exampleBlockIssuerFeature,
					},
				},
			},
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: axongo.AccountOutputFeatures{
					exampleBlockIssuerFeature,
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: currentSlot,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleIssuer.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
					BIC: map[axongo.AccountID]axongo.BlockIssuanceCredits{
						exampleAccountID: -100,
					},
				},
			},
			wantErr: axongo.ErrAccountLocked,
		},
		{
			name: "fail - removing staking feature before end epoch",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.StakingFeature{
							StakedAmount: 100,
							FixedCost:    50,
							StartEpoch:   currentEpoch,
							EndEpoch:     currentEpoch + 10000,
						},
						exampleBlockIssuerFeature,
					},
				},
			},
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: axongo.AccountOutputFeatures{
					exampleBlockIssuerFeature,
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: currentSlot,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleIssuer.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: axongo.ErrStakingFeatureRemovedBeforeUnbonding,
		},
		{
			name: "fail - changing staking feature's staked amount",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.StakingFeature{
							StakedAmount: 100,
							FixedCost:    50,
							StartEpoch:   currentEpoch,
							EndEpoch:     currentEpoch + 10000,
						},
						exampleBlockIssuerFeature,
					},
				},
			},
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: axongo.AccountOutputFeatures{
					&axongo.StakingFeature{
						StakedAmount: 90,
						FixedCost:    50,
						StartEpoch:   currentEpoch,
						EndEpoch:     currentEpoch + 10000,
					},
					exampleBlockIssuerFeature,
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: currentSlot,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleIssuer.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: axongo.ErrStakingFeatureModifiedBeforeUnbonding,
		},
		{
			name: "fail - reducing staking feature's end epoch by more than the unbonding period",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.StakingFeature{
							StakedAmount: 100,
							FixedCost:    50,
							StartEpoch:   currentEpoch,
							EndEpoch:     currentEpoch + 10000,
						},
						exampleBlockIssuerFeature,
					},
				},
			},
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: axongo.AccountOutputFeatures{
					&axongo.StakingFeature{
						StakedAmount: 100,
						FixedCost:    50,
						StartEpoch:   currentEpoch,
						EndEpoch:     currentEpoch + tpkg.ZeroCostTestAPI.ProtocolParameters().StakingUnbondingPeriod() - 5,
					},
					exampleBlockIssuerFeature,
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: currentSlot,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleIssuer.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: axongo.ErrStakingEndEpochTooEarly,
		},
		{
			name: "fail - expired staking feature removed without specifying reward input",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(1000, 0),
				ChainID:  exampleAccountID,
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.StakingFeature{
							StakedAmount: 50,
							FixedCost:    5,
							StartEpoch:   currentEpoch - 10,
							EndEpoch:     currentEpoch - 5,
						},
						exampleBlockIssuerFeature,
					},
				},
			},
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: axongo.AccountOutputFeatures{
					exampleBlockIssuerFeature,
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: currentSlot,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleIssuer.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: axongo.ErrStakingRewardInputMissing,
		},
		{
			name: "fail - changing an expired staking feature without claiming",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(1000, 0),
				ChainID:  exampleAccountID,
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.StakingFeature{
							StakedAmount: 50,
							FixedCost:    5,
							StartEpoch:   currentEpoch - 10,
							EndEpoch:     currentEpoch - 5,
						},
						exampleBlockIssuerFeature,
					},
				},
			},
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: axongo.AccountOutputFeatures{
					&axongo.StakingFeature{
						StakedAmount: 80,
						FixedCost:    5,
						StartEpoch:   currentEpoch,
						EndEpoch:     currentEpoch + testProtoParams.StakingUnbondingPeriod(),
					},
					exampleBlockIssuerFeature,
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: currentSlot,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleIssuer.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: axongo.ErrStakingRewardInputMissing,
		},
		{
			name: "fail - claiming rewards of an expired staking feature without resetting start epoch",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(1000, 0),
				ChainID:  exampleAccountID,
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.StakingFeature{
							StakedAmount: 50,
							FixedCost:    5,
							StartEpoch:   currentEpoch - 10,
							EndEpoch:     currentEpoch - 5,
						},
						exampleBlockIssuerFeature,
					},
				},
			},
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: axongo.AccountOutputFeatures{
					&axongo.StakingFeature{
						StakedAmount: 50,
						FixedCost:    5,
						StartEpoch:   currentEpoch - 10,
						EndEpoch:     currentEpoch + 10,
					},
					exampleBlockIssuerFeature,
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: currentSlot,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleIssuer.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
					Rewards: map[axongo.ChainID]axongo.Mana{
						exampleAccountID: 200,
					},
				},
			},
			wantErr: axongo.ErrStakingStartEpochInvalid,
		},
		{
			name: "fail - claiming rewards without removing staking feature",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(1000, 0),
				ChainID:  exampleAccountID,
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.StakingFeature{
							StakedAmount: 50,
							FixedCost:    5,
							StartEpoch:   currentEpoch - 10,
							EndEpoch:     currentEpoch - 5,
						},
						exampleBlockIssuerFeature,
					},
				},
			},
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: axongo.AccountOutputFeatures{
					&axongo.StakingFeature{
						StakedAmount: 50,
						FixedCost:    5,
						StartEpoch:   currentEpoch - 10,
						EndEpoch:     currentEpoch - 5,
					},
					exampleBlockIssuerFeature,
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: currentSlot,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleIssuer.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
					Rewards: map[axongo.ChainID]axongo.Mana{
						exampleAccountID: 200,
					},
				},
			},
			wantErr: axongo.ErrStakingRewardClaimingInvalid,
		},
		{
			name: "fail - destroy account with expired staking feature but without claiming rewards",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(1000, 0),
				ChainID:  exampleAccountID,
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.StakingFeature{
							StakedAmount: 50,
							FixedCost:    5,
							StartEpoch:   currentEpoch - 10,
							EndEpoch:     currentEpoch - 5,
						},
						&axongo.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      currentSlot - 50,
						},
					},
				},
			},
			next:      nil,
			transType: axongo.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: currentSlot,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleIssuer.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
					BIC:     exampleBIC,
					Rewards: map[axongo.ChainID]axongo.Mana{},
				},
			},
			wantErr: axongo.ErrStakingRewardInputMissing,
		},
		{
			name: "ok - destroy transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: tpkg.RandAccountAddress().AccountID(),
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
				},
			},
			next:      nil,
			transType: axongo.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - destroy block issuer account with negative BIC",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      1000,
						},
					},
				},
			},
			next:      nil,
			transType: axongo.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: 1001,
					},
					UnlockedAddrs: vm.UnlockedAddresses{},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: 1001,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
					BIC: map[axongo.AccountID]axongo.BlockIssuanceCredits{
						exampleAccountID: -1,
					},
				},
			},
			wantErr: axongo.ErrAccountLocked,
		},
		{
			name: "fail - destroy block issuer account no BIC provided",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      1000,
						},
					},
				},
			},
			next:      nil,
			transType: axongo.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: 1001,
					},
					UnlockedAddrs: vm.UnlockedAddresses{},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: 1001,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrBlockIssuanceCreditInputMissing,
		},
		{
			name: "fail - non-expired block issuer destroy transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: tpkg.RandAccountAddress().AccountID(),
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      1000,
						},
					},
				},
			},
			next:      nil,
			transType: axongo.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: 1000,
					},
					UnlockedAddrs: vm.UnlockedAddresses{},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: 1000,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrBlockIssuerNotExpired,
		},
		{
			name: "ok - expired block issuer destroy transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      1000,
						},
					},
				},
			},
			next:      nil,
			transType: axongo.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: 1001,
					},
					UnlockedAddrs: vm.UnlockedAddresses{},
					BIC: map[axongo.AccountID]axongo.BlockIssuanceCredits{
						exampleAccountID: 0,
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							Inputs:       nil,
							CreationSlot: 1001,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "failed - remove non-expired block issuer feature transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      1000,
						},
					},
				},
			},
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: axongo.AccountOutputFeatures{},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: 999,
					},
					UnlockedAddrs: vm.UnlockedAddresses{},
					BIC: map[axongo.AccountID]axongo.BlockIssuanceCredits{
						exampleAccountID: 0,
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							Inputs:       nil,
							CreationSlot: 999,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrBlockIssuerNotExpired,
		},
		{
			name: "ok - remove expired block issuer feature transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      1000,
						},
					},
				},
			},
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: axongo.AccountOutputFeatures{},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: 1001,
					},
					UnlockedAddrs: vm.UnlockedAddresses{},
					BIC: map[axongo.AccountID]axongo.BlockIssuanceCredits{
						exampleAccountID: 0,
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							Inputs:       nil,
							CreationSlot: 1001,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - foundry counter increased by number of new foundries",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					FoundryCounter: 5,
				},
			},
			next: &axongo.AccountOutput{
				Amount:    200,
				AccountID: exampleAccountID,
				// mutating owner
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				FoundryCounter: 7,
				Features: axongo.AccountOutputFeatures{
					&axongo.SenderFeature{Address: exampleAddress},
					&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("1337")}},
					&axongo.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
						ExpirySlot:      1015,
					},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleAddress.Key(): {UnlockedAtInputIndex: 0},
					},
					Commitment: &axongo.Commitment{
						Slot: 990,
					},
					BIC: exampleBIC,
					InChains: map[axongo.ChainID]*vm.ChainOutputWithIDs{
						// serial number 5
						exampleExistingFoundryOutputFoundryID: {
							ChainID:  exampleExistingFoundryOutputFoundryID,
							OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
							Output:   exampleExistingFoundryOutput,
						},
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: 900,
							Inputs:       nil,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything())},
						Outputs: axongo.TxEssenceOutputs{
							&axongo.FoundryOutput{
								Amount:       100,
								SerialNumber: 6,
								TokenScheme:  &axongo.SimpleTokenScheme{},
								UnlockConditions: axongo.FoundryOutputUnlockConditions{
									&axongo.ImmutableAccountUnlockCondition{Address: exampleAccountID.ToAddress().(*axongo.AccountAddress)},
								},
							},
							&axongo.FoundryOutput{
								Amount:       100,
								SerialNumber: 7,
								TokenScheme:  &axongo.SimpleTokenScheme{},
								UnlockConditions: axongo.FoundryOutputUnlockConditions{
									&axongo.ImmutableAccountUnlockCondition{Address: exampleAccountID.ToAddress().(*axongo.AccountAddress)},
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - update expired block issuer feature without extending expiration after MCA",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      1000,
						},
					},
				},
			},
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: axongo.AccountOutputFeatures{
					&axongo.SenderFeature{Address: exampleAddress},
					&axongo.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
						ExpirySlot:      1000,
					},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: 990,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleAddress.Key(): {UnlockedAtInputIndex: 0},
					},
					BIC: map[axongo.AccountID]axongo.BlockIssuanceCredits{
						exampleAccountID: 10,
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							Inputs:       nil,
							CreationSlot: 990,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - update account immutable features",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					ImmutableFeatures: axongo.AccountOutputImmFeatures{
						&axongo.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      900,
						},
					},
				},
			},
			next: &axongo.AccountOutput{
				Amount:    200,
				AccountID: exampleAccountID,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: exampleAddress},
				},
				ImmutableFeatures: axongo.AccountOutputImmFeatures{
					&axongo.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
						ExpirySlot:      999,
					},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleAddress.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrChainOutputImmutableFeaturesChanged,
		},
		{
			name: "fail - update expired block issuer feature with extending expiration before MCA",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      900,
						},
					},
				},
			},
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: axongo.AccountOutputFeatures{
					&axongo.SenderFeature{Address: exampleAddress},
					&axongo.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
						ExpirySlot:      999,
					},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: 990,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleAddress.Key(): {UnlockedAtInputIndex: 0},
					},
					BIC: map[axongo.AccountID]axongo.BlockIssuanceCredits{
						exampleAccountID: 10,
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							Inputs:       nil,
							CreationSlot: 990,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrBlockIssuerExpiryTooEarly,
		},
		{
			name: "fail - update expired block issuer feature with extending expiration to the past before MCA",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      1100,
						},
					},
				},
			},
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: axongo.AccountOutputFeatures{
					&axongo.SenderFeature{Address: exampleAddress},
					&axongo.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
						ExpirySlot:      999,
					},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: 990,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleAddress.Key(): {UnlockedAtInputIndex: 0},
					},
					BIC: map[axongo.AccountID]axongo.BlockIssuanceCredits{
						exampleAccountID: 10,
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							Inputs:       nil,
							CreationSlot: 990,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrBlockIssuerExpiryTooEarly,
		},
		{
			name: "fail - update block issuer account with negative BIC",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      1000,
						},
					},
				},
			},
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: axongo.AccountOutputFeatures{
					&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("1337")}},
					&axongo.SenderFeature{Address: exampleAddress},
					&axongo.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
						ExpirySlot:      1000,
					},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: 900,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleAddress.Key(): {UnlockedAtInputIndex: 0},
					},
					BIC: map[axongo.AccountID]axongo.BlockIssuanceCredits{
						exampleAccountID: -1,
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							Inputs:       nil,
							CreationSlot: 900,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrAccountLocked,
		},
		{
			name: "fail - update block issuer account without BIC provided",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      1000,
						},
					},
				},
			},
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: axongo.AccountOutputFeatures{
					&axongo.SenderFeature{Address: exampleAddress},
					&axongo.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
						ExpirySlot:      1000,
					},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: 900,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleAddress.Key(): {UnlockedAtInputIndex: 0},
					},

					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: 900,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrBlockIssuanceCreditInputMissing,
		},
		{
			name: "ok - update block issuer feature expiration to earlier slot",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      1000,
						},
					},
				},
			},
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: axongo.AccountOutputFeatures{
					&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("1337")}},
					&axongo.SenderFeature{Address: exampleAddress},
					&axongo.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
						ExpirySlot:      999,
					},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: 900,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleAddress.Key(): {UnlockedAtInputIndex: 0},
					},
					BIC: map[axongo.AccountID]axongo.BlockIssuanceCredits{
						exampleAccountID: 10,
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: 900,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - non-expired block issuer replace key",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &axongo.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      1000,
						},
					},
					FoundryCounter: 5,
				},
			},
			next: &axongo.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: exampleAddress},
				},
				FoundryCounter: 5,
				Features: axongo.AccountOutputFeatures{
					&axongo.SenderFeature{Address: exampleAddress},
					&axongo.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
						ExpirySlot:      1000,
					},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &axongo.Commitment{
						Slot: 0,
					},
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleAddress.Key(): {UnlockedAtInputIndex: 0},
					},
					InChains: map[axongo.ChainID]*vm.ChainOutputWithIDs{
						// serial number 5
						exampleExistingFoundryOutputFoundryID: {
							ChainID:  exampleExistingFoundryOutputFoundryID,
							OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
							Output:   exampleExistingFoundryOutput,
						},
					},
					BIC: map[axongo.AccountID]axongo.BlockIssuanceCredits{
						exampleAccountID: 10,
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							Inputs:       nil,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything())},
						Outputs: axongo.TxEssenceOutputs{
							&axongo.FoundryOutput{
								Amount:       100,
								SerialNumber: 6,
								TokenScheme:  &axongo.SimpleTokenScheme{},
								UnlockConditions: axongo.FoundryOutputUnlockConditions{
									&axongo.ImmutableAccountUnlockCondition{Address: exampleAccountID.ToAddress().(*axongo.AccountAddress)},
								},
							},
							&axongo.FoundryOutput{
								Amount:       100,
								SerialNumber: 7,
								TokenScheme:  &axongo.SimpleTokenScheme{},
								UnlockConditions: axongo.FoundryOutputUnlockConditions{
									&axongo.ImmutableAccountUnlockCondition{Address: exampleAccountID.ToAddress().(*axongo.AccountAddress)},
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - invalid foundry counters",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &axongo.AccountOutput{
					Amount:         100,
					AccountID:      exampleAccountID,
					FoundryCounter: 5,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: exampleAddress},
					},
					ImmutableFeatures: axongo.AccountOutputImmFeatures{
						&axongo.IssuerFeature{Address: exampleIssuer},
					},
				},
			},
			nextMut: map[string]fieldMutations{
				"foundry_counter_lower_than_current": {
					"FoundryCounter": uint32(4),
				},
				"foundries_not_created": {
					"FoundryCounter": uint32(7),
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					InChains:      vm.ChainInputSet{},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrAccountInvalidFoundryCounter,
		},
	}

	for _, tt2 := range tests {
		tt := tt2

		if tt.nextMut != nil {
			for mutName, muts := range tt.nextMut {
				t.Run(fmt.Sprintf("%s_%s", tt.name, mutName), func(t *testing.T) {
					cpy := copyObjectAndMutate(t, tt.input.Output, muts).(*axongo.AccountOutput)

					createWorkingSet(t, tt.input, tt.svCtx.WorkingSet)

					err := novaVM.ChainSTVF(tt.svCtx, tt.transType, tt.input, cpy)
					if tt.wantErr != nil {
						require.ErrorIs(t, err, tt.wantErr)
						return
					}
					require.NoError(t, err)
				})
			}
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			createWorkingSet(t, tt.input, tt.svCtx.WorkingSet)

			err := novaVM.ChainSTVF(tt.svCtx, tt.transType, tt.input, tt.next)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestAnchorOutput_ValidateStateTransition(t *testing.T) {
	exampleIssuer := tpkg.RandEd25519Address()
	exampleAnchorID := tpkg.RandAnchorAddress().AnchorID()

	exampleStateCtrl := tpkg.RandEd25519Address()
	exampleGovCtrl := tpkg.RandEd25519Address()

	type test struct {
		name      string
		input     *vm.ChainOutputWithIDs
		next      *axongo.AnchorOutput
		nextMut   map[string]fieldMutations
		transType axongo.ChainTransitionType
		svCtx     *vm.Params
		wantErr   error
	}

	tests := []*test{
		{
			name: "ok - genesis transition",
			next: &axongo.AnchorOutput{
				Amount:   100,
				AnchorID: axongo.AnchorID{},
				UnlockConditions: axongo.AnchorOutputUnlockConditions{
					&axongo.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&axongo.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				ImmutableFeatures: axongo.AnchorOutputImmFeatures{
					&axongo.IssuerFeature{Address: exampleIssuer},
				},
			},
			input:     nil,
			transType: axongo.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleIssuer.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - destroy transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &axongo.AnchorOutput{
					Amount:   100,
					AnchorID: tpkg.RandAnchorAddress().AnchorID(),
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&axongo.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
				},
			},
			next:      nil,
			transType: axongo.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - gov transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &axongo.AnchorOutput{
					Amount:   100,
					AnchorID: exampleAnchorID,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&axongo.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					StateIndex: 10,
					Features: axongo.AnchorOutputFeatures{
						&axongo.StateMetadataFeature{Entries: axongo.StateMetadataFeatureEntries{"data": []byte("1337")}},
					},
				},
			},
			next: &axongo.AnchorOutput{
				Amount:     100,
				AnchorID:   exampleAnchorID,
				StateIndex: 10,
				// mutating controllers
				UnlockConditions: axongo.AnchorOutputUnlockConditions{
					&axongo.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&axongo.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: axongo.AnchorOutputFeatures{
					&axongo.SenderFeature{Address: exampleGovCtrl},
					&axongo.StateMetadataFeature{Entries: axongo.StateMetadataFeatureEntries{"data": []byte("1337")}},
					// adding metadata feature
					&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("1338")}},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleGovCtrl.Key(): {UnlockedAtInputIndex: 0},
					},
					Commitment: &axongo.Commitment{
						Slot: 990,
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: 900,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - state transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &axongo.AnchorOutput{
					Amount:   100,
					AnchorID: exampleAnchorID,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&axongo.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					StateIndex: 10,
					Features: axongo.AnchorOutputFeatures{
						&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("1338")}},
					},
				},
			},
			next: &axongo.AnchorOutput{
				Amount:   200,
				AnchorID: exampleAnchorID,
				UnlockConditions: axongo.AnchorOutputUnlockConditions{
					&axongo.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&axongo.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				StateIndex: 11,
				Features: axongo.AnchorOutputFeatures{
					&axongo.SenderFeature{Address: exampleStateCtrl},
					&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("1338")}},
					// adding state metadata feature
					&axongo.StateMetadataFeature{Entries: axongo.StateMetadataFeatureEntries{"data": []byte("1337")}},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleStateCtrl.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							Inputs:       nil,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - update anchor immutable features in gov transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &axongo.AnchorOutput{
					Amount:   100,
					AnchorID: exampleAnchorID,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&axongo.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					ImmutableFeatures: axongo.AnchorOutputImmFeatures{
						&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("1337")}},
					},
					StateIndex: 10,
				},
			},
			next: &axongo.AnchorOutput{
				Amount:   100,
				AnchorID: exampleAnchorID,
				UnlockConditions: axongo.AnchorOutputUnlockConditions{
					&axongo.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&axongo.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				StateIndex: 10,
				ImmutableFeatures: axongo.AnchorOutputImmFeatures{
					&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("1338")}},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleStateCtrl.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrAnchorInvalidGovernanceTransition,
		},
		{
			name: "fail - update anchor immutable features in state transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &axongo.AnchorOutput{
					Amount:   100,
					AnchorID: exampleAnchorID,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&axongo.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					ImmutableFeatures: axongo.AnchorOutputImmFeatures{
						&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("1337")}},
					},
					StateIndex: 10,
				},
			},
			next: &axongo.AnchorOutput{
				Amount:   200,
				AnchorID: exampleAnchorID,
				UnlockConditions: axongo.AnchorOutputUnlockConditions{
					&axongo.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&axongo.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				StateIndex: 11,
				ImmutableFeatures: axongo.AnchorOutputImmFeatures{
					&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("1338")}},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleStateCtrl.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrAnchorInvalidStateTransition,
		},
		{
			name: "fail - gov transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &axongo.AnchorOutput{
					Amount:     100,
					AnchorID:   exampleAnchorID,
					StateIndex: 10,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&axongo.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: axongo.AnchorOutputFeatures{
						&axongo.StateMetadataFeature{
							Entries: axongo.StateMetadataFeatureEntries{
								"data": []byte("foo"),
							},
						},
					},
				},
			},
			nextMut: map[string]fieldMutations{
				"amount": {
					"Amount": axongo.BaseToken(1337),
				},
				"state_metadata_feature_changed": {
					"Features": axongo.AnchorOutputFeatures{
						&axongo.StateMetadataFeature{
							Entries: axongo.StateMetadataFeatureEntries{
								"data": []byte("bar"),
							},
						},
					},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrAnchorInvalidGovernanceTransition,
		},
		{
			name: "fail - state transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &axongo.AnchorOutput{
					Amount:     100,
					AnchorID:   exampleAnchorID,
					StateIndex: 10,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&axongo.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: axongo.AnchorOutputFeatures{
						&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("foo")}},
					},
					ImmutableFeatures: axongo.AnchorOutputImmFeatures{
						&axongo.IssuerFeature{Address: exampleIssuer},
					},
				},
			},
			nextMut: map[string]fieldMutations{
				"state_controller": {
					"StateIndex": uint32(11),
					"UnlockConditions": axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						&axongo.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
				},
				"governance_controller": {
					"StateIndex": uint32(11),
					"UnlockConditions": axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&axongo.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				"state_index_lower": {
					"StateIndex": uint32(4),
				},
				"state_index_bigger_more_than_1": {
					"StateIndex": uint32(7),
				},
				"metadata_feature_changed": {
					"StateIndex": uint32(11),
					"Features": axongo.AnchorOutputFeatures{
						&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("bar")}},
					},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					InChains:      vm.ChainInputSet{},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrAnchorInvalidStateTransition,
		},
	}

	for _, tt2 := range tests {
		tt := tt2

		if tt.nextMut != nil {
			for mutName, muts := range tt.nextMut {
				t.Run(fmt.Sprintf("%s_%s", tt.name, mutName), func(t *testing.T) {
					cpy := copyObjectAndMutate(t, tt.input.Output, muts).(*axongo.AnchorOutput)

					createWorkingSet(t, tt.input, tt.svCtx.WorkingSet)

					err := novaVM.ChainSTVF(tt.svCtx, tt.transType, tt.input, cpy)
					if tt.wantErr != nil {
						require.ErrorIs(t, err, tt.wantErr)
						return
					}
					require.NoError(t, err)
				})
			}
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			createWorkingSet(t, tt.input, tt.svCtx.WorkingSet)

			err := novaVM.ChainSTVF(tt.svCtx, tt.transType, tt.input, tt.next)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestFoundryOutput_ValidateStateTransition(t *testing.T) {
	exampleAccountAddr := tpkg.RandAccountAddress()

	startingSupply := new(big.Int).SetUint64(100)
	exampleFoundry := &axongo.FoundryOutput{
		Amount:       100,
		SerialNumber: 6,
		TokenScheme: &axongo.SimpleTokenScheme{
			MintedTokens:  startingSupply,
			MeltedTokens:  big.NewInt(0),
			MaximumSupply: new(big.Int).SetUint64(1000),
		},
		UnlockConditions: axongo.FoundryOutputUnlockConditions{
			&axongo.ImmutableAccountUnlockCondition{Address: exampleAccountAddr},
		},
	}

	toBeDestoyedFoundry := &axongo.FoundryOutput{
		Amount:       100,
		SerialNumber: 6,
		TokenScheme: &axongo.SimpleTokenScheme{
			MintedTokens:  startingSupply,
			MeltedTokens:  startingSupply,
			MaximumSupply: new(big.Int).SetUint64(1000),
		},
		UnlockConditions: axongo.FoundryOutputUnlockConditions{
			&axongo.ImmutableAccountUnlockCondition{Address: exampleAccountAddr},
		},
	}

	type test struct {
		name      string
		input     *vm.ChainOutputWithIDs
		next      *axongo.FoundryOutput
		nextMut   map[string]fieldMutations
		transType axongo.ChainTransitionType
		svCtx     *vm.Params
		wantErr   error
	}

	tests := []*test{
		{
			name:      "ok - genesis transition",
			next:      exampleFoundry,
			input:     nil,
			transType: axongo.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
						Outputs: axongo.TxEssenceOutputs{exampleFoundry},
					},
					InChains: vm.ChainInputSet{
						exampleAccountAddr.AccountID(): &vm.ChainOutputWithIDs{
							ChainID:  exampleAccountAddr.AccountID(),
							OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
							Output:   &axongo.AccountOutput{FoundryCounter: 5},
						},
					},
					OutChains: map[axongo.ChainID]axongo.ChainOutput{
						exampleAccountAddr.AccountID(): &axongo.AccountOutput{FoundryCounter: 6},
					},
					OutNativeTokens: map[axongo.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): startingSupply,
					},
				},
			},
			wantErr: nil,
		},
		{
			name:      "fail - genesis transition - mint supply not equal to out",
			next:      exampleFoundry,
			input:     nil,
			transType: axongo.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
						Outputs: axongo.TxEssenceOutputs{exampleFoundry},
					},
					InChains: vm.ChainInputSet{
						exampleAccountAddr.AccountID(): &vm.ChainOutputWithIDs{
							ChainID:  exampleAccountAddr.AccountID(),
							OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
							Output:   &axongo.AccountOutput{FoundryCounter: 5},
						},
					},
					OutChains: map[axongo.ChainID]axongo.ChainOutput{
						exampleAccountAddr.AccountID(): &axongo.AccountOutput{FoundryCounter: 6},
					},
					OutNativeTokens: map[axongo.NativeTokenID]*big.Int{
						// absent but should be there
					},
				},
			},
			wantErr: axongo.ErrSimpleTokenSchemeGenesisInvalid,
		},
		{
			name:      "fail - genesis transition - serial number not in interval",
			next:      exampleFoundry,
			input:     nil,
			transType: axongo.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
						Outputs: axongo.TxEssenceOutputs{exampleFoundry},
					},
					InChains: vm.ChainInputSet{
						exampleAccountAddr.AccountID(): &vm.ChainOutputWithIDs{
							ChainID:  exampleAccountAddr.AccountID(),
							OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
							Output:   &axongo.AccountOutput{FoundryCounter: 6},
						},
					},
					OutChains: map[axongo.ChainID]axongo.ChainOutput{
						exampleAccountAddr.AccountID(): &axongo.AccountOutput{FoundryCounter: 7},
					},
					OutNativeTokens: map[axongo.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): startingSupply,
					},
				},
			},
			wantErr: axongo.ErrFoundrySerialInvalid,
		},
		{
			name:      "fail - genesis transition - foundries unsorted",
			next:      exampleFoundry,
			input:     nil,
			transType: axongo.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
						Outputs: axongo.TxEssenceOutputs{
							&axongo.FoundryOutput{
								Amount: 100,
								// exampleFoundry has serial number 6
								SerialNumber: 7,
								TokenScheme: &axongo.SimpleTokenScheme{
									MintedTokens:  startingSupply,
									MeltedTokens:  big.NewInt(0),
									MaximumSupply: new(big.Int).SetUint64(1000),
								},
								UnlockConditions: axongo.FoundryOutputUnlockConditions{
									&axongo.ImmutableAccountUnlockCondition{Address: exampleAccountAddr},
								},
							},
							exampleFoundry,
						},
					},
					InChains: vm.ChainInputSet{
						exampleAccountAddr.AccountID(): &vm.ChainOutputWithIDs{
							ChainID:  exampleAccountAddr.AccountID(),
							OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
							Output:   &axongo.AccountOutput{FoundryCounter: 5},
						},
					},
					OutChains: map[axongo.ChainID]axongo.ChainOutput{
						exampleAccountAddr.AccountID(): &axongo.AccountOutput{FoundryCounter: 7},
					},
					OutNativeTokens: map[axongo.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): startingSupply,
					},
				},
			},
			wantErr: axongo.ErrFoundrySerialInvalid,
		},
		{
			name: "ok - state transition - metadata feature",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleFoundry,
			},
			nextMut: map[string]fieldMutations{
				"change_metadata": {
					"Features": axongo.FoundryOutputFeatures{
						&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": tpkg.RandBytes(20)}},
					},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					OutNativeTokens: map[axongo.NativeTokenID]*big.Int{},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - state transition - mint",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleFoundry,
			},
			nextMut: map[string]fieldMutations{
				"+300": {
					"TokenScheme": &axongo.SimpleTokenScheme{
						MintedTokens:  big.NewInt(400),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					OutNativeTokens: map[axongo.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): new(big.Int).SetUint64(300),
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - state transition - melt",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleFoundry,
			},
			nextMut: map[string]fieldMutations{
				"-50": {
					"TokenScheme": &axongo.SimpleTokenScheme{
						MintedTokens:  big.NewInt(100),
						MeltedTokens:  big.NewInt(50),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					InNativeTokens: map[axongo.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): startingSupply,
					},
					OutNativeTokens: map[axongo.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): new(big.Int).SetUint64(50),
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - state transition - burn",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleFoundry,
			},
			nextMut:   map[string]fieldMutations{},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					InNativeTokens: map[axongo.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): startingSupply,
					},
					OutNativeTokens: map[axongo.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): new(big.Int).SetUint64(50),
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - state transition - melt complete supply",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleFoundry,
			},
			nextMut: map[string]fieldMutations{
				"-100": {
					"TokenScheme": &axongo.SimpleTokenScheme{
						MintedTokens:  big.NewInt(100),
						MeltedTokens:  big.NewInt(100),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					InNativeTokens: map[axongo.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): startingSupply,
					},
					OutNativeTokens: map[axongo.NativeTokenID]*big.Int{},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - state transition - mint (out: excess)",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleFoundry,
			},
			nextMut: map[string]fieldMutations{
				"+100": {
					"TokenScheme": &axongo.SimpleTokenScheme{
						MintedTokens:  big.NewInt(200),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					OutNativeTokens: map[axongo.NativeTokenID]*big.Int{
						// 100 excess
						exampleFoundry.MustNativeTokenID(): new(big.Int).SetUint64(200),
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrSimpleTokenSchemeMintingInvalid,
		},
		{
			name: "fail - state transition - mint (out: deficit)",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleFoundry,
			},
			nextMut: map[string]fieldMutations{
				"+100": {
					"TokenScheme": &axongo.SimpleTokenScheme{
						MintedTokens:  big.NewInt(200),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					OutNativeTokens: map[axongo.NativeTokenID]*big.Int{
						// 50 deficit
						exampleFoundry.MustNativeTokenID(): new(big.Int).SetUint64(50),
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrSimpleTokenSchemeMintingInvalid,
		},
		{
			name: "fail - state transition - melt (out: excess)",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleFoundry,
			},
			nextMut: map[string]fieldMutations{
				"-50": {
					"TokenScheme": &axongo.SimpleTokenScheme{
						MintedTokens:  big.NewInt(100),
						MeltedTokens:  big.NewInt(50),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					InNativeTokens: map[axongo.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): startingSupply,
					},
					OutNativeTokens: map[axongo.NativeTokenID]*big.Int{
						// 25 excess
						exampleFoundry.MustNativeTokenID(): new(big.Int).SetUint64(75),
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrSimpleTokenSchemeMeltingInvalid,
		},
		{
			name: "fail - state transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleFoundry,
			},
			nextMut: map[string]fieldMutations{
				"maximum_supply": {
					"TokenScheme": &axongo.SimpleTokenScheme{
						MintedTokens:  startingSupply,
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: big.NewInt(1337),
					},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrSimpleTokenSchemeMaximumSupplyChanged,
		},
		{
			name: "ok - destroy transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   toBeDestoyedFoundry,
			},
			transType: axongo.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					InNativeTokens:  map[axongo.NativeTokenID]*big.Int{},
					OutNativeTokens: map[axongo.NativeTokenID]*big.Int{},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - destroy transition - foundry token unbalanced",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleFoundry,
			},
			transType: axongo.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					InNativeTokens: map[axongo.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): startingSupply,
					},
					OutNativeTokens: map[axongo.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): new(big.Int).Mul(startingSupply, new(big.Int).SetUint64(2)),
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrNativeTokenSumUnbalanced,
		},
	}

	for _, tt := range tests {

		if tt.nextMut != nil {
			for mutName, muts := range tt.nextMut {
				t.Run(fmt.Sprintf("%s_%s", tt.name, mutName), func(t *testing.T) {
					cpy := copyObjectAndMutate(t, tt.input.Output, muts).(*axongo.FoundryOutput)

					createWorkingSet(t, tt.input, tt.svCtx.WorkingSet)

					err := novaVM.ChainSTVF(tt.svCtx, tt.transType, tt.input, cpy)
					if tt.wantErr != nil {
						require.ErrorIs(t, err, tt.wantErr)
						return
					}
					require.NoError(t, err)
				})
			}
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			createWorkingSet(t, tt.input, tt.svCtx.WorkingSet)

			err := novaVM.ChainSTVF(tt.svCtx, tt.transType, tt.input, tt.next)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestNFTOutput_ValidateStateTransition(t *testing.T) {
	exampleIssuer := tpkg.RandEd25519Address()

	exampleCurrentNFTOutput := &axongo.NFTOutput{
		Amount: 100,
		NFTID:  axongo.NFTID{},
		UnlockConditions: axongo.NFTOutputUnlockConditions{
			&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
		},
		ImmutableFeatures: axongo.NFTOutputImmFeatures{
			&axongo.IssuerFeature{Address: exampleIssuer},
			&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("some-ipfs-link")}},
		},
	}

	type test struct {
		name      string
		input     *vm.ChainOutputWithIDs
		next      *axongo.NFTOutput
		nextMut   map[string]fieldMutations
		transType axongo.ChainTransitionType
		svCtx     *vm.Params
		wantErr   error
	}

	tests := []*test{
		{
			name:      "ok - genesis transition",
			next:      exampleCurrentNFTOutput,
			input:     nil,
			transType: axongo.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleIssuer.Key(): {UnlockedAtInputIndex: 0},
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - destroy transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleCurrentNFTOutput,
			},
			next:      nil,
			transType: axongo.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - state transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleCurrentNFTOutput,
			},
			nextMut: map[string]fieldMutations{
				"amount": {
					"Amount": axongo.BaseToken(1337),
				},
				"address": {
					"UnlockConditions": axongo.NFTOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - state transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleCurrentNFTOutput,
			},
			nextMut: map[string]fieldMutations{
				"immutable_metadata": {
					"ImmutableFeatures": axongo.NFTOutputImmFeatures{
						&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("link-to-cat.gif")}},
					},
				},
				"issuer": {
					"ImmutableFeatures": axongo.NFTOutputImmFeatures{
						&axongo.IssuerFeature{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrChainOutputImmutableFeaturesChanged,
		},
	}

	for _, tt := range tests {

		if tt.nextMut != nil {
			for mutName, muts := range tt.nextMut {
				t.Run(fmt.Sprintf("%s_%s", tt.name, mutName), func(t *testing.T) {
					cpy := copyObjectAndMutate(t, tt.input.Output, muts).(*axongo.NFTOutput)

					createWorkingSet(t, tt.input, tt.svCtx.WorkingSet)

					err := novaVM.ChainSTVF(tt.svCtx, tt.transType, tt.input, cpy)
					if tt.wantErr != nil {
						require.ErrorIs(t, err, tt.wantErr)
						return
					}
					require.NoError(t, err)
				})
			}
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			createWorkingSet(t, tt.input, tt.svCtx.WorkingSet)

			err := novaVM.ChainSTVF(tt.svCtx, tt.transType, tt.input, tt.next)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestDelegationOutput_ValidateStateTransition(t *testing.T) {
	currentEpoch := axongo.EpochIndex(20)
	epochStartSlot := tpkg.ZeroCostTestAPI.TimeProvider().EpochStart(currentEpoch)
	epochEndSlot := tpkg.ZeroCostTestAPI.TimeProvider().EpochEnd(currentEpoch)
	minCommittableAge := tpkg.ZeroCostTestAPI.ProtocolParameters().MinCommittableAge()
	maxCommittableAge := tpkg.ZeroCostTestAPI.ProtocolParameters().MaxCommittableAge()

	// Commitment indices that will always end up being in the current epoch no matter if
	// future or past bounded.
	epochStartCommitmentIndex := epochStartSlot - minCommittableAge
	epochEndCommitmentIndex := epochEndSlot - maxCommittableAge

	exampleDelegationID := axongo.DelegationIDFromOutputID(tpkg.RandOutputID(0))

	type test struct {
		name      string
		input     *vm.ChainOutputWithIDs
		next      *axongo.DelegationOutput
		nextMut   map[string]fieldMutations
		transType axongo.ChainTransitionType
		svCtx     *vm.Params
		wantErr   error
	}

	tests := []*test{
		{
			name: "ok - valid genesis",
			next: &axongo.DelegationOutput{
				Amount:           100,
				DelegatedAmount:  100,
				DelegationID:     axongo.EmptyDelegationID(),
				ValidatorAddress: tpkg.RandAccountAddress(),
				StartEpoch:       currentEpoch + 1,
				EndEpoch:         0,
				UnlockConditions: axongo.DelegationOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			input:     nil,
			transType: axongo.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					Commitment: &axongo.Commitment{
						Slot: epochStartCommitmentIndex,
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - invalid genesis - non-zero delegation ID",
			next: &axongo.DelegationOutput{
				Amount:           100,
				DelegatedAmount:  100,
				DelegationID:     exampleDelegationID,
				ValidatorAddress: tpkg.RandAccountAddress(),
				StartEpoch:       currentEpoch + 1,
				EndEpoch:         0,
				UnlockConditions: axongo.DelegationOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			input:     nil,
			transType: axongo.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					Commitment: &axongo.Commitment{
						Slot: epochStartCommitmentIndex,
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrNewChainOutputHasNonZeroedID,
		},
		{
			name: "fail - invalid genesis - delegated amount does not match amount",
			next: &axongo.DelegationOutput{
				Amount:           100,
				DelegatedAmount:  120,
				DelegationID:     axongo.EmptyDelegationID(),
				ValidatorAddress: tpkg.RandAccountAddress(),
				StartEpoch:       currentEpoch + 1,
				EndEpoch:         0,
				UnlockConditions: axongo.DelegationOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			input:     nil,
			transType: axongo.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					Commitment: &axongo.Commitment{
						Slot: epochStartCommitmentIndex,
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrDelegationAmountMismatch,
		},
		{
			name: "fail - invalid genesis - non-zero end epoch",
			next: &axongo.DelegationOutput{
				Amount:           100,
				DelegatedAmount:  100,
				DelegationID:     axongo.EmptyDelegationID(),
				ValidatorAddress: tpkg.RandAccountAddress(),
				StartEpoch:       currentEpoch + 1,
				EndEpoch:         currentEpoch + 5,
				UnlockConditions: axongo.DelegationOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			input:     nil,
			transType: axongo.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					Commitment: &axongo.Commitment{
						Slot: epochStartCommitmentIndex,
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrDelegationEndEpochNotZero,
		},
		{
			name: "fail - invalid transition - start epoch not set to expected epoch",
			next: &axongo.DelegationOutput{
				Amount:           100,
				DelegatedAmount:  100,
				DelegationID:     axongo.EmptyDelegationID(),
				ValidatorAddress: tpkg.RandAccountAddress(),
				StartEpoch:       currentEpoch - 3,
				EndEpoch:         0,
				UnlockConditions: axongo.DelegationOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			transType: axongo.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					Commitment: &axongo.Commitment{
						Slot: epochStartCommitmentIndex,
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrDelegationStartEpochInvalid,
		},
		{
			name: "fail - invalid transition - non-zero delegation id on input",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &axongo.DelegationOutput{
					Amount:           100,
					DelegatedAmount:  100,
					DelegationID:     tpkg.RandDelegationID(),
					ValidatorAddress: tpkg.RandAccountAddress(),
					StartEpoch:       currentEpoch + 1,
					EndEpoch:         0,
					UnlockConditions: axongo.DelegationOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			next:      &axongo.DelegationOutput{},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					Commitment: &axongo.Commitment{
						Slot: epochStartCommitmentIndex,
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrDelegationOutputTransitionedTwice,
		},
		{
			name: "fail - invalid transition - modified delegated amount, start epoch and validator id",
			input: &vm.ChainOutputWithIDs{
				ChainID:  exampleDelegationID,
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &axongo.DelegationOutput{
					Amount:           100,
					DelegatedAmount:  100,
					DelegationID:     axongo.EmptyDelegationID(),
					ValidatorAddress: tpkg.RandAccountAddress(),
					StartEpoch:       currentEpoch + 1,
					EndEpoch:         0,
					UnlockConditions: axongo.DelegationOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			nextMut: map[string]fieldMutations{
				"delegated_amount_modified": {
					"DelegatedAmount": axongo.BaseToken(1337),
					"Amount":          axongo.BaseToken(5),
					"DelegationID":    exampleDelegationID,
					"EndEpoch":        currentEpoch,
				},
				"start_epoch_modified": {
					"StartEpoch":   axongo.EpochIndex(3),
					"DelegationID": exampleDelegationID,
					"EndEpoch":     currentEpoch,
				},
				"validator_address_modified": {
					"ValidatorAddress": tpkg.RandAccountAddress(),
					"DelegationID":     exampleDelegationID,
					"EndEpoch":         currentEpoch,
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					Commitment: &axongo.Commitment{
						Slot: epochStartCommitmentIndex,
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrDelegationModified,
		},
		{
			name: "fail - invalid pre-registration slot transition - end epoch not set to expected epoch",
			input: &vm.ChainOutputWithIDs{
				ChainID:  exampleDelegationID,
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &axongo.DelegationOutput{
					Amount:           100,
					DelegatedAmount:  100,
					DelegationID:     axongo.EmptyDelegationID(),
					ValidatorAddress: tpkg.RandAccountAddress(),
					StartEpoch:       currentEpoch + 1,
					EndEpoch:         0,
					UnlockConditions: axongo.DelegationOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			nextMut: map[string]fieldMutations{
				"end_epoch_-1": {
					"DelegationID": exampleDelegationID,
					"EndEpoch":     currentEpoch - 1,
				},
				"end_epoch_+1": {
					"DelegationID": exampleDelegationID,
					"EndEpoch":     currentEpoch + 1,
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					Commitment: &axongo.Commitment{
						Slot: epochStartCommitmentIndex,
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrDelegationEndEpochInvalid,
		},
		{
			name: "fail - invalid post-registration slot transition - end epoch not set to expected epoch",
			input: &vm.ChainOutputWithIDs{
				ChainID:  exampleDelegationID,
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &axongo.DelegationOutput{
					Amount:           100,
					DelegatedAmount:  100,
					DelegationID:     axongo.EmptyDelegationID(),
					ValidatorAddress: tpkg.RandAccountAddress(),
					StartEpoch:       currentEpoch + 1,
					EndEpoch:         0,
					UnlockConditions: axongo.DelegationOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			nextMut: map[string]fieldMutations{
				"end_epoch_current": {
					"DelegationID": exampleDelegationID,
					"EndEpoch":     currentEpoch,
				},
				"end_epoch_+2": {
					"DelegationID": exampleDelegationID,
					"EndEpoch":     currentEpoch + 2,
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					Commitment: &axongo.Commitment{
						Slot: epochEndCommitmentIndex,
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrDelegationEndEpochInvalid,
		},
		{
			name: "fail - invalid transition - cannot claim rewards during transition",
			input: &vm.ChainOutputWithIDs{
				ChainID:  exampleDelegationID,
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &axongo.DelegationOutput{
					Amount:           100,
					DelegatedAmount:  100,
					DelegationID:     axongo.EmptyDelegationID(),
					ValidatorAddress: tpkg.RandAccountAddress(),
					StartEpoch:       currentEpoch + 1,
					EndEpoch:         0,
					UnlockConditions: axongo.DelegationOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			next:      &axongo.DelegationOutput{},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					Commitment: &axongo.Commitment{
						Slot: epochStartCommitmentIndex,
					},
					Rewards: map[axongo.ChainID]axongo.Mana{
						exampleDelegationID: 1,
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrDelegationRewardsClaimingInvalid,
		},
		{
			name: "ok - valid destruction",
			input: &vm.ChainOutputWithIDs{
				ChainID:  exampleDelegationID,
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &axongo.DelegationOutput{
					Amount:           100,
					DelegatedAmount:  100,
					DelegationID:     axongo.EmptyDelegationID(),
					ValidatorAddress: tpkg.RandAccountAddress(),
					StartEpoch:       currentEpoch + 1,
					EndEpoch:         0,
					UnlockConditions: axongo.DelegationOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			nextMut:   nil,
			transType: axongo.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					Commitment: &axongo.Commitment{
						Slot: epochStartCommitmentIndex,
					},
					Rewards: map[axongo.ChainID]axongo.Mana{
						exampleDelegationID: 0,
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - invalid destruction - missing reward input",
			input: &vm.ChainOutputWithIDs{
				ChainID:  exampleDelegationID,
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &axongo.DelegationOutput{
					Amount:           100,
					DelegatedAmount:  100,
					DelegationID:     axongo.EmptyDelegationID(),
					ValidatorAddress: tpkg.RandAccountAddress(),
					StartEpoch:       currentEpoch + 1,
					EndEpoch:         0,
					UnlockConditions: axongo.DelegationOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			nextMut:   nil,
			transType: axongo.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					Commitment: &axongo.Commitment{
						Slot: epochStartCommitmentIndex,
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrDelegationRewardInputMissing,
		},
		{
			name: "fail - invalid genesis - missing commitment input",
			next: &axongo.DelegationOutput{
				Amount:           100,
				DelegatedAmount:  100,
				DelegationID:     axongo.EmptyDelegationID(),
				ValidatorAddress: tpkg.RandAccountAddress(),
				StartEpoch:       currentEpoch + 1,
				EndEpoch:         0,
				UnlockConditions: axongo.DelegationOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			transType: axongo.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{},
					Rewards: map[axongo.ChainID]axongo.Mana{
						exampleDelegationID: 0,
					},
					Tx: &axongo.Transaction{
						TransactionEssence: &axongo.TransactionEssence{
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrDelegationCommitmentInputMissing,
		},
	}

	for _, tt := range tests {
		if tt.nextMut != nil {
			for mutName, muts := range tt.nextMut {
				t.Run(fmt.Sprintf("%s_%s", tt.name, mutName), func(t *testing.T) {
					cpy := copyObjectAndMutate(t, tt.input.Output, muts).(*axongo.DelegationOutput)

					createWorkingSet(t, tt.input, tt.svCtx.WorkingSet)

					err := novaVM.ChainSTVF(tt.svCtx, tt.transType, tt.input, cpy)
					if tt.wantErr != nil {
						require.ErrorIs(t, err, tt.wantErr)
						return
					}
					require.NoError(t, err)
				})
			}
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			createWorkingSet(t, tt.input, tt.svCtx.WorkingSet)

			err := novaVM.ChainSTVF(tt.svCtx, tt.transType, tt.input, tt.next)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestImplicitAccountOutput_ValidateStateTransition(t *testing.T) {
	exampleIssuer := tpkg.RandEd25519Address()
	exampleAccountID := tpkg.RandAccountAddress().AccountID()

	currentEpoch := axongo.EpochIndex(20)
	currentSlot := tpkg.ZeroCostTestAPI.TimeProvider().EpochStart(currentEpoch)
	blockIssuerPubKey := axongo.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray())
	exampleBlockIssuerFeature := &axongo.BlockIssuerFeature{
		BlockIssuerKeys: axongo.NewBlockIssuerKeys(blockIssuerPubKey),
		ExpirySlot:      currentSlot + tpkg.ZeroCostTestAPI.ProtocolParameters().MaxCommittableAge(),
	}

	exampleBIC := map[axongo.AccountID]axongo.BlockIssuanceCredits{
		exampleAccountID: 100,
	}

	type test struct {
		name      string
		input     *vm.ChainOutputWithIDs
		next      *axongo.AccountOutput
		transType axongo.ChainTransitionType
		svCtx     *vm.Params
		wantErr   error
	}

	implicitAccountCreationAddr := axongo.ImplicitAccountCreationAddressFromPubKey(tpkg.RandEd25519Signature().PublicKey[:])
	exampleAmount := axongo.BaseToken(100_000)

	tests := []*test{
		{
			name: "ok - implicit account conversion transition",
			next: &axongo.AccountOutput{
				Amount:    exampleAmount,
				AccountID: exampleAccountID,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: axongo.AccountOutputFeatures{
					exampleBlockIssuerFeature,
				},
			},
			input: &vm.ChainOutputWithIDs{
				ChainID: exampleAccountID,
				Output: &vm.ImplicitAccountOutput{
					BasicOutput: &axongo.BasicOutput{
						Amount: exampleAmount,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{
								Address: implicitAccountCreationAddr,
							},
						},
					},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{

				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleIssuer.Key(): {UnlockedAtInputIndex: 0},
					},
					BIC: exampleBIC,
					Commitment: &axongo.Commitment{
						Slot: currentSlot,
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - explicit account lacks block issuer feature",
			next: &axongo.AccountOutput{
				Amount:    exampleAmount,
				AccountID: exampleAccountID,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: axongo.AccountOutputFeatures{},
			},
			input: &vm.ChainOutputWithIDs{
				ChainID: exampleAccountID,
				Output: &vm.ImplicitAccountOutput{
					BasicOutput: &axongo.BasicOutput{
						Amount: exampleAmount,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{
								Address: implicitAccountCreationAddr,
							},
						},
					},
				},
			},
			transType: axongo.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedAddrs: vm.UnlockedAddresses{
						exampleIssuer.Key(): {UnlockedAtInputIndex: 0},
					},
					BIC: exampleBIC,
					Commitment: &axongo.Commitment{
						Slot: currentSlot,
					},
					Tx: &axongo.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &axongo.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: axongo.ErrBlockIssuerNotExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createWorkingSet(t, tt.input, tt.svCtx.WorkingSet)

			err := novaVM.ChainSTVF(tt.svCtx, tt.transType, tt.input, tt.next)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func createWorkingSet(t *testing.T, input *vm.ChainOutputWithIDs, workingSet *vm.WorkingSet) {
	t.Helper()

	if input != nil {
		// create the working set for the test
		if workingSet.UTXOInputsSet == nil {
			workingSet.UTXOInputsSet = make(vm.InputSet)
		}
		workingSet.UTXOInputsSet[input.OutputID] = input.Output

		totalManaIn, err := vm.TotalManaIn(
			tpkg.ZeroCostTestAPI.ManaDecayProvider(),
			tpkg.ZeroCostTestAPI.StorageScoreStructure(),
			workingSet.Tx.CreationSlot,
			workingSet.UTXOInputsSet,
			workingSet.Rewards,
		)
		require.NoError(t, err)
		workingSet.TotalManaIn = totalManaIn

		totalManaOut, err := vm.TotalManaOut(
			workingSet.Tx.Outputs,
			workingSet.Tx.Allotments,
		)
		require.NoError(t, err)
		workingSet.TotalManaOut = totalManaOut
	}
}

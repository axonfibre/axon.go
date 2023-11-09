//nolint:scopelint
package iotago_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestTransactionEssence_DeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok",
			source: tpkg.RandTransaction(tpkg.TestAPI),
			target: &iotago.Transaction{API: tpkg.TestAPI},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestChainConstrainedOutputUniqueness(t *testing.T) {
	ident1 := tpkg.RandEd25519Address()

	inputIDs := tpkg.RandOutputIDs(1)

	accountAddress := iotago.AccountAddressFromOutputID(inputIDs[0])
	accountID := accountAddress.AccountID()

	anchorAddress := iotago.AnchorAddressFromOutputID(inputIDs[0])
	anchorID := anchorAddress.AnchorID()

	nftAddress := iotago.NFTAddressFromOutputID(inputIDs[0])
	nftID := nftAddress.NFTID()

	tests := []deSerializeTest{
		{
			// we transition the same Account twice
			name: "transition the same Account twice",
			source: tpkg.RandSignedTransactionWithTransaction(tpkg.TestAPI,
				&iotago.Transaction{
					API: tpkg.TestAPI,
					TransactionEssence: &iotago.TransactionEssence{
						NetworkID:     tpkg.TestNetworkID,
						ContextInputs: iotago.TxEssenceContextInputs{},
						Inputs:        inputIDs.UTXOInputs(),
						Allotments:    iotago.Allotments{},
						Capabilities:  iotago.TransactionCapabilitiesBitMask{},
					},
					Outputs: iotago.TxEssenceOutputs{
						&iotago.AccountOutput{
							Amount:    OneIOTA,
							AccountID: accountID,
							UnlockConditions: iotago.AccountOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: ident1},
							},
							Features: nil,
						},
						&iotago.AccountOutput{
							Amount:    OneIOTA,
							AccountID: accountID,
							UnlockConditions: iotago.AccountOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: ident1},
							},
							Features: nil,
						},
					},
				}),
			target:    &iotago.SignedTransaction{},
			seriErr:   iotago.ErrNonUniqueChainOutputs,
			deSeriErr: nil,
		},
		{
			// we transition the same Anchor twice
			name: "transition the same Anchor twice",
			source: tpkg.RandSignedTransactionWithTransaction(tpkg.TestAPI,
				&iotago.Transaction{
					API: tpkg.TestAPI,
					TransactionEssence: &iotago.TransactionEssence{
						NetworkID:     tpkg.TestNetworkID,
						ContextInputs: iotago.TxEssenceContextInputs{},
						Inputs:        inputIDs.UTXOInputs(),
						Allotments:    iotago.Allotments{},
						Capabilities:  iotago.TransactionCapabilitiesBitMask{},
					},
					Outputs: iotago.TxEssenceOutputs{
						&iotago.AnchorOutput{
							Amount:   OneIOTA,
							AnchorID: anchorID,
							UnlockConditions: iotago.AnchorOutputUnlockConditions{
								&iotago.StateControllerAddressUnlockCondition{Address: ident1},
								&iotago.GovernorAddressUnlockCondition{Address: ident1},
							},
							Features: nil,
						},
						&iotago.AnchorOutput{
							Amount:   OneIOTA,
							AnchorID: anchorID,
							UnlockConditions: iotago.AnchorOutputUnlockConditions{
								&iotago.StateControllerAddressUnlockCondition{Address: ident1},
								&iotago.GovernorAddressUnlockCondition{Address: ident1},
							},
							Features: nil,
						},
					},
				}),
			target:    &iotago.SignedTransaction{},
			seriErr:   iotago.ErrNonUniqueChainOutputs,
			deSeriErr: nil,
		},
		{
			// we transition the same NFT twice
			name: "transition the same NFT twice",
			source: tpkg.RandSignedTransactionWithTransaction(tpkg.TestAPI,
				&iotago.Transaction{
					API: tpkg.TestAPI,
					TransactionEssence: &iotago.TransactionEssence{
						NetworkID:    tpkg.TestNetworkID,
						Inputs:       inputIDs.UTXOInputs(),
						Capabilities: iotago.TransactionCapabilitiesBitMask{},
					},
					Outputs: iotago.TxEssenceOutputs{
						&iotago.NFTOutput{
							Amount: OneIOTA,
							NFTID:  nftID,
							UnlockConditions: iotago.NFTOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: ident1},
							},
							Features: nil,
						},
						&iotago.NFTOutput{
							Amount: OneIOTA,
							NFTID:  nftID,
							UnlockConditions: iotago.NFTOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: ident1},
							},
							Features: nil,
						},
					},
				}),
			target:    &iotago.SignedTransaction{},
			seriErr:   iotago.ErrNonUniqueChainOutputs,
			deSeriErr: nil,
		},
		{
			// we transition the same Foundry twice
			name: "transition the same Foundry twice",
			source: tpkg.RandSignedTransactionWithTransaction(tpkg.TestAPI,
				&iotago.Transaction{
					API: tpkg.TestAPI,
					TransactionEssence: &iotago.TransactionEssence{
						NetworkID:    tpkg.TestNetworkID,
						Inputs:       inputIDs.UTXOInputs(),
						Capabilities: iotago.TransactionCapabilitiesBitMask{},
					},
					Outputs: iotago.TxEssenceOutputs{
						&iotago.AccountOutput{
							Amount:    OneIOTA,
							AccountID: accountID,
							UnlockConditions: iotago.AccountOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: ident1},
							},
							Features: nil,
						},
						&iotago.FoundryOutput{
							Amount:       OneIOTA,
							SerialNumber: 1,
							TokenScheme: &iotago.SimpleTokenScheme{
								MintedTokens:  big.NewInt(50),
								MeltedTokens:  big.NewInt(0),
								MaximumSupply: big.NewInt(50),
							},
							UnlockConditions: iotago.FoundryOutputUnlockConditions{
								&iotago.ImmutableAccountUnlockCondition{Address: accountAddress},
							},
							Features: nil,
						},
						&iotago.FoundryOutput{
							Amount:       OneIOTA,
							SerialNumber: 1,
							TokenScheme: &iotago.SimpleTokenScheme{
								MintedTokens:  big.NewInt(50),
								MeltedTokens:  big.NewInt(0),
								MaximumSupply: big.NewInt(50),
							},
							UnlockConditions: iotago.FoundryOutputUnlockConditions{
								&iotago.ImmutableAccountUnlockCondition{Address: accountAddress},
							},
							Features: nil,
						},
					},
				}),
			target:    &iotago.SignedTransaction{},
			seriErr:   iotago.ErrNonUniqueChainOutputs,
			deSeriErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestAllotmentUniqueness(t *testing.T) {
	inputIDs := tpkg.RandOutputIDs(1)

	accountAddress := iotago.AccountAddressFromOutputID(inputIDs[0])
	accountID := accountAddress.AccountID()

	tests := []deSerializeTest{
		{
			name: "allot to the same account twice",
			source: tpkg.RandSignedTransactionWithTransaction(tpkg.TestAPI,
				&iotago.Transaction{
					API: tpkg.TestAPI,
					TransactionEssence: &iotago.TransactionEssence{
						NetworkID:     tpkg.TestNetworkID,
						ContextInputs: iotago.TxEssenceContextInputs{},
						Inputs:        inputIDs.UTXOInputs(),
						Allotments: iotago.Allotments{
							&iotago.Allotment{
								AccountID: accountID,
								Mana:      0,
							},
							&iotago.Allotment{
								AccountID: tpkg.RandAccountID(),
								Mana:      12,
							},
							&iotago.Allotment{
								AccountID: accountID,
								Mana:      12,
							},
						},
						Capabilities: iotago.TransactionCapabilitiesBitMask{},
					},
					Outputs: iotago.TxEssenceOutputs{
						tpkg.RandBasicOutput(iotago.AddressEd25519),
					},
				}),
			target:    &iotago.SignedTransaction{},
			seriErr:   serializer.ErrArrayValidationOrderViolatesLexicalOrder,
			deSeriErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestTransactionEssenceCapabilitiesBitMask(t *testing.T) {

	type test struct {
		name    string
		tx      *iotago.Transaction
		wantErr error
	}

	randTransactionWithCapabilities := func(capabilities iotago.TransactionCapabilitiesBitMask) *iotago.Transaction {
		tx := tpkg.RandTransaction(tpkg.TestAPI)
		tx.Capabilities = capabilities
		return tx
	}

	tests := []*test{
		{
			name:    "ok - no trailing zero bytes",
			tx:      randTransactionWithCapabilities(iotago.TransactionCapabilitiesBitMask{0x01}),
			wantErr: nil,
		},
		{
			name:    "ok - empty capabilities",
			tx:      randTransactionWithCapabilities(iotago.TransactionCapabilitiesBitMask{}),
			wantErr: nil,
		},
		{
			name:    "fail - single zero byte",
			tx:      randTransactionWithCapabilities(iotago.TransactionCapabilitiesBitMask{0x00}),
			wantErr: iotago.ErrBitmaskTrailingZeroBytes,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.tx.SyntacticallyValidate(tpkg.TestAPI)
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)

				return
			}
			require.NoError(t, err)
		})
	}
}

func TestTransactionSyntacticMaxMana(t *testing.T) {
	type test struct {
		name    string
		tx      *iotago.Transaction
		wantErr error
	}

	basicOutputWithMana := func(mana iotago.Mana) *iotago.BasicOutput {
		return &iotago.BasicOutput{
			Amount: OneIOTA,
			Mana:   mana,
			Conditions: iotago.BasicOutputUnlockConditions{
				&iotago.AddressUnlockCondition{
					Address: tpkg.RandEd25519Address(),
				},
			},
		}
	}

	allotmentWithMana := func(mana iotago.Mana) *iotago.Allotment {
		return &iotago.Allotment{
			Mana:      mana,
			AccountID: tpkg.RandAccountID(),
		}
	}

	var maxManaValue iotago.Mana = (1 << tpkg.TestAPI.ProtocolParameters().ManaParameters().BitsCount) - 1
	tests := []*test{
		{
			name: "ok - stored mana sum below max value",
			tx: tpkg.RandTransactionWithOptions(tpkg.TestAPI,
				func(tx *iotago.Transaction) {
					tx.Outputs = iotago.TxEssenceOutputs{basicOutputWithMana(1), basicOutputWithMana(maxManaValue - 1)}
				},
			),
			wantErr: nil,
		},
		{
			name: "fail - one output's stored mana exceeds max mana value",
			tx: tpkg.RandTransactionWithOptions(tpkg.TestAPI,
				func(tx *iotago.Transaction) {
					tx.Outputs = iotago.TxEssenceOutputs{basicOutputWithMana(maxManaValue + 1)}
				},
			),
			wantErr: iotago.ErrMaxManaExceeded,
		},
		{
			name: "fail - sum of stored mana exceeds max mana value",
			tx: tpkg.RandTransactionWithOptions(tpkg.TestAPI,
				func(tx *iotago.Transaction) {
					tx.Outputs = iotago.TxEssenceOutputs{basicOutputWithMana(maxManaValue - 1), basicOutputWithMana(maxManaValue - 1)}
				},
			),
			wantErr: iotago.ErrMaxManaExceeded,
		},
		{
			name: "ok - allotted mana sum below max value",
			tx: tpkg.RandTransactionWithOptions(tpkg.TestAPI,
				func(tx *iotago.Transaction) {
					tx.Allotments = iotago.Allotments{allotmentWithMana(1), allotmentWithMana(maxManaValue - 1)}
				},
			),
			wantErr: nil,
		},
		{
			name: "fail - one allotment's mana exceeds max value",
			tx: tpkg.RandTransactionWithOptions(tpkg.TestAPI,
				func(tx *iotago.Transaction) {
					tx.Allotments = iotago.Allotments{allotmentWithMana(maxManaValue + 1)}
				},
			),
			wantErr: iotago.ErrMaxManaExceeded,
		},
		{
			name: "fail - sum of allotted mana exceeds max value",
			tx: tpkg.RandTransactionWithOptions(tpkg.TestAPI,
				func(tx *iotago.Transaction) {
					tx.Allotments = iotago.Allotments{allotmentWithMana(maxManaValue - 1), allotmentWithMana(maxManaValue - 1)}
				},
			),
			wantErr: iotago.ErrMaxManaExceeded,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.tx.SyntacticallyValidate(tpkg.TestAPI)
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)

				return
			}
			require.NoError(t, err)
		})
	}
}

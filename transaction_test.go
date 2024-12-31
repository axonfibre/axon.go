package axongo_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/axonfibre/fibre.go/lo"
	"github.com/axonfibre/fibre.go/serializer/v2"
	axongo "github.com/axonfibre/axon.go/v4"
	"github.com/axonfibre/axon.go/v4/builder"
	"github.com/axonfibre/axon.go/v4/tpkg"
	"github.com/axonfibre/axon.go/v4/tpkg/frameworks"
)

func TestTransactionEssence_DeSerialize(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name:   "ok",
			Source: tpkg.RandTransaction(tpkg.ZeroCostTestAPI),
			Target: &axongo.Transaction{API: tpkg.ZeroCostTestAPI},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

func TestChainConstrainedOutputUniqueness(t *testing.T) {
	addr1 := tpkg.RandEd25519Address()

	inputIDs := tpkg.RandOutputIDs(1)

	accountAddress := axongo.AccountAddressFromOutputID(inputIDs[0])
	accountID := accountAddress.AccountID()

	anchorAddress := axongo.AnchorAddressFromOutputID(inputIDs[0])
	anchorID := anchorAddress.AnchorID()

	nftAddress := axongo.NFTAddressFromOutputID(inputIDs[0])
	nftID := nftAddress.NFTID()

	tests := []*frameworks.DeSerializeTest{
		{
			// we transition the same Account twice
			Name: "transition the same Account twice",
			Source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				&axongo.Transaction{
					API: tpkg.ZeroCostTestAPI,
					TransactionEssence: &axongo.TransactionEssence{
						NetworkID:     tpkg.TestNetworkID,
						ContextInputs: axongo.TxEssenceContextInputs{},
						Inputs:        inputIDs.UTXOInputs(),
						Allotments:    axongo.Allotments{},
						Capabilities:  axongo.TransactionCapabilitiesBitMask{},
					},
					Outputs: axongo.TxEssenceOutputs{
						&axongo.AccountOutput{
							Amount:    OneIOTA,
							AccountID: accountID,
							UnlockConditions: axongo.AccountOutputUnlockConditions{
								&axongo.AddressUnlockCondition{Address: addr1},
							},
							Features: nil,
						},
						&axongo.AccountOutput{
							Amount:    OneIOTA,
							AccountID: accountID,
							UnlockConditions: axongo.AccountOutputUnlockConditions{
								&axongo.AddressUnlockCondition{Address: addr1},
							},
							Features: nil,
						},
					},
				}),
			Target:    &axongo.SignedTransaction{},
			SeriErr:   axongo.ErrNonUniqueChainOutputs,
			DeSeriErr: axongo.ErrNonUniqueChainOutputs,
		},
		{
			// we transition the same Anchor twice
			Name: "transition the same Anchor twice",
			Source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				&axongo.Transaction{
					API: tpkg.ZeroCostTestAPI,
					TransactionEssence: &axongo.TransactionEssence{
						NetworkID:     tpkg.TestNetworkID,
						ContextInputs: axongo.TxEssenceContextInputs{},
						Inputs:        inputIDs.UTXOInputs(),
						Allotments:    axongo.Allotments{},
						Capabilities:  axongo.TransactionCapabilitiesBitMask{},
					},
					Outputs: axongo.TxEssenceOutputs{
						&axongo.AnchorOutput{
							Amount:   OneIOTA,
							AnchorID: anchorID,
							UnlockConditions: axongo.AnchorOutputUnlockConditions{
								&axongo.StateControllerAddressUnlockCondition{Address: addr1},
								&axongo.GovernorAddressUnlockCondition{Address: addr1},
							},
							Features: nil,
						},
						&axongo.AnchorOutput{
							Amount:   OneIOTA,
							AnchorID: anchorID,
							UnlockConditions: axongo.AnchorOutputUnlockConditions{
								&axongo.StateControllerAddressUnlockCondition{Address: addr1},
								&axongo.GovernorAddressUnlockCondition{Address: addr1},
							},
							Features: nil,
						},
					},
				}),
			Target:    &axongo.SignedTransaction{},
			SeriErr:   axongo.ErrNonUniqueChainOutputs,
			DeSeriErr: axongo.ErrNonUniqueChainOutputs,
		},
		{
			// we transition the same NFT twice
			Name: "transition the same NFT twice",
			Source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				&axongo.Transaction{
					API: tpkg.ZeroCostTestAPI,
					TransactionEssence: &axongo.TransactionEssence{
						NetworkID:    tpkg.TestNetworkID,
						Inputs:       inputIDs.UTXOInputs(),
						Capabilities: axongo.TransactionCapabilitiesBitMask{},
					},
					Outputs: axongo.TxEssenceOutputs{
						&axongo.NFTOutput{
							Amount: OneIOTA,
							NFTID:  nftID,
							UnlockConditions: axongo.NFTOutputUnlockConditions{
								&axongo.AddressUnlockCondition{Address: addr1},
							},
							Features: nil,
						},
						&axongo.NFTOutput{
							Amount: OneIOTA,
							NFTID:  nftID,
							UnlockConditions: axongo.NFTOutputUnlockConditions{
								&axongo.AddressUnlockCondition{Address: addr1},
							},
							Features: nil,
						},
					},
				}),
			Target:    &axongo.SignedTransaction{},
			SeriErr:   axongo.ErrNonUniqueChainOutputs,
			DeSeriErr: axongo.ErrNonUniqueChainOutputs,
		},
		{
			// we transition the same Foundry twice
			Name: "transition the same Foundry twice",
			Source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				&axongo.Transaction{
					API: tpkg.ZeroCostTestAPI,
					TransactionEssence: &axongo.TransactionEssence{
						NetworkID:    tpkg.TestNetworkID,
						Inputs:       inputIDs.UTXOInputs(),
						Capabilities: axongo.TransactionCapabilitiesBitMask{},
					},
					Outputs: axongo.TxEssenceOutputs{
						&axongo.AccountOutput{
							Amount:    OneIOTA,
							AccountID: accountID,
							UnlockConditions: axongo.AccountOutputUnlockConditions{
								&axongo.AddressUnlockCondition{Address: addr1},
							},
							Features: nil,
						},
						&axongo.FoundryOutput{
							Amount:       OneIOTA,
							SerialNumber: 1,
							TokenScheme: &axongo.SimpleTokenScheme{
								MintedTokens:  big.NewInt(50),
								MeltedTokens:  big.NewInt(0),
								MaximumSupply: big.NewInt(50),
							},
							UnlockConditions: axongo.FoundryOutputUnlockConditions{
								&axongo.ImmutableAccountUnlockCondition{Address: accountAddress},
							},
							Features: nil,
						},
						&axongo.FoundryOutput{
							Amount:       OneIOTA,
							SerialNumber: 1,
							TokenScheme: &axongo.SimpleTokenScheme{
								MintedTokens:  big.NewInt(50),
								MeltedTokens:  big.NewInt(0),
								MaximumSupply: big.NewInt(50),
							},
							UnlockConditions: axongo.FoundryOutputUnlockConditions{
								&axongo.ImmutableAccountUnlockCondition{Address: accountAddress},
							},
							Features: nil,
						},
					},
				}),
			Target:    &axongo.SignedTransaction{},
			SeriErr:   axongo.ErrNonUniqueChainOutputs,
			DeSeriErr: axongo.ErrNonUniqueChainOutputs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

func TestAllotmentUniqueness(t *testing.T) {
	inputIDs := tpkg.RandOutputIDs(1)

	accountAddress := axongo.AccountAddressFromOutputID(inputIDs[0])
	accountID := accountAddress.AccountID()

	tests := []*frameworks.DeSerializeTest{
		{
			Name: "allot to the same account twice",
			Source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				&axongo.Transaction{
					API: tpkg.ZeroCostTestAPI,
					TransactionEssence: &axongo.TransactionEssence{
						NetworkID:     tpkg.TestNetworkID,
						ContextInputs: axongo.TxEssenceContextInputs{},
						Inputs:        inputIDs.UTXOInputs(),
						Allotments: axongo.Allotments{
							&axongo.Allotment{
								AccountID: accountID,
								Mana:      0,
							},
							&axongo.Allotment{
								AccountID: accountID,
								Mana:      12,
							},
							&axongo.Allotment{
								AccountID: tpkg.RandAccountID(),
								Mana:      12,
							},
						},
						Capabilities: axongo.TransactionCapabilitiesBitMask{},
					},
					Outputs: axongo.TxEssenceOutputs{
						tpkg.RandBasicOutput(axongo.AddressEd25519),
					},
				}),
			Target:    &axongo.SignedTransaction{},
			SeriErr:   axongo.ErrArrayValidationViolatesUniqueness,
			DeSeriErr: axongo.ErrArrayValidationViolatesUniqueness,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

func TestTransactionEssenceCapabilitiesBitMask(t *testing.T) {

	type test struct {
		name    string
		tx      *axongo.Transaction
		wantErr error
	}

	randTransactionWithCapabilities := func(capabilities axongo.TransactionCapabilitiesBitMask) *axongo.Transaction {
		tx := tpkg.RandTransaction(tpkg.ZeroCostTestAPI)
		tx.Capabilities = capabilities
		return tx
	}

	tests := []*test{
		{
			name:    "ok - no trailing zero bytes",
			tx:      randTransactionWithCapabilities(axongo.TransactionCapabilitiesBitMask{0x01}),
			wantErr: nil,
		},
		{
			name:    "ok - empty capabilities",
			tx:      randTransactionWithCapabilities(axongo.TransactionCapabilitiesBitMask{}),
			wantErr: nil,
		},
		{
			name:    "fail - single zero byte",
			tx:      randTransactionWithCapabilities(axongo.TransactionCapabilitiesBitMask{0x00}),
			wantErr: axongo.ErrBitmaskTrailingZeroBytes,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.tx.SyntacticallyValidate(tpkg.ZeroCostTestAPI)
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
		tx      *axongo.Transaction
		wantErr error
	}

	basicOutputWithMana := func(mana axongo.Mana) *axongo.BasicOutput {
		return &axongo.BasicOutput{
			Amount: OneIOTA,
			Mana:   mana,
			UnlockConditions: axongo.BasicOutputUnlockConditions{
				&axongo.AddressUnlockCondition{
					Address: tpkg.RandEd25519Address(),
				},
			},
		}
	}

	allotmentWithMana := func(mana axongo.Mana) *axongo.Allotment {
		return &axongo.Allotment{
			Mana:      mana,
			AccountID: tpkg.RandAccountID(),
		}
	}

	var maxManaValue axongo.Mana = (1 << tpkg.ZeroCostTestAPI.ProtocolParameters().ManaParameters().BitsCount) - 1
	tests := []*test{
		{
			name: "ok - stored mana sum below max value",
			tx: tpkg.RandTransactionWithOptions(tpkg.ZeroCostTestAPI,
				func(tx *axongo.Transaction) {
					tx.Outputs = axongo.TxEssenceOutputs{basicOutputWithMana(1), basicOutputWithMana(maxManaValue - 1)}
				},
			),
			wantErr: nil,
		},
		{
			name: "fail - one output's stored mana exceeds max mana value",
			tx: tpkg.RandTransactionWithOptions(tpkg.ZeroCostTestAPI,
				func(tx *axongo.Transaction) {
					tx.Outputs = axongo.TxEssenceOutputs{basicOutputWithMana(maxManaValue + 1)}
				},
			),
			wantErr: axongo.ErrMaxManaExceeded,
		},
		{
			name: "fail - sum of stored mana exceeds max mana value",
			tx: tpkg.RandTransactionWithOptions(tpkg.ZeroCostTestAPI,
				func(tx *axongo.Transaction) {
					tx.Outputs = axongo.TxEssenceOutputs{basicOutputWithMana(maxManaValue - 1), basicOutputWithMana(maxManaValue - 1)}
				},
			),
			wantErr: axongo.ErrMaxManaExceeded,
		},
		{
			name: "ok - allotted mana sum below max value",
			tx: tpkg.RandTransactionWithOptions(tpkg.ZeroCostTestAPI,
				func(tx *axongo.Transaction) {
					tx.Allotments = axongo.Allotments{allotmentWithMana(1), allotmentWithMana(maxManaValue - 1)}
					tx.Allotments.Sort()
				},
			),
			wantErr: nil,
		},
		{
			name: "fail - one allotment's mana exceeds max value",
			tx: tpkg.RandTransactionWithOptions(tpkg.ZeroCostTestAPI,
				func(tx *axongo.Transaction) {
					tx.Allotments = axongo.Allotments{allotmentWithMana(maxManaValue + 1)}
					tx.Allotments.Sort()
				},
			),
			wantErr: axongo.ErrMaxManaExceeded,
		},
		{
			name: "fail - sum of allotted mana exceeds max value",
			tx: tpkg.RandTransactionWithOptions(tpkg.ZeroCostTestAPI,
				func(tx *axongo.Transaction) {
					tx.Allotments = axongo.Allotments{allotmentWithMana(maxManaValue - 1), allotmentWithMana(maxManaValue - 1)}
					tx.Allotments.Sort()
				},
			),
			wantErr: axongo.ErrMaxManaExceeded,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.tx.SyntacticallyValidate(tpkg.ZeroCostTestAPI)
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)

				return
			}
			require.NoError(t, err)
		})
	}
}

func TestTransactionInputUniqueness(t *testing.T) {
	type test struct {
		name      string
		inputs    axongo.TxEssenceInputs
		seriErr   error
		deseriErr error
	}

	input1 := axongo.MustOutputIDFromHexString("0x2668778ef0362d601c36ea05c742185daa1740dfcdaee0acfde6a9806a1f2ed20d8566fd0800")
	input2 := axongo.MustOutputIDFromHexString("0x3f34a869f47f8454e7cb233943cd31a0e3bd8b9551b1390039ec582b0a196856eff185120400")
	input3 := axongo.MustOutputIDFromHexString("0xfdad2fee88cc4f1020848dce710124ac9060cdbee840a72b750c1f6901502576422f83b50500")
	// Differs from input3 only in the output index.
	input4 := axongo.MustOutputIDFromHexString("0xfdad2fee88cc4f1020848dce710124ac9060cdbee840a72b750c1f6901502576422f83b50600")

	tests := []test{
		{
			name: "ok - inputs unique",
			inputs: axongo.TxEssenceInputs{
				input3.UTXOInput(),
				input1.UTXOInput(),
				input4.UTXOInput(),
				input2.UTXOInput(),
			},
			seriErr: nil,
		},
		{
			name: "fail - duplicate inputs",
			inputs: axongo.TxEssenceInputs{
				input1.UTXOInput(),
				input2.UTXOInput(),
				input2.UTXOInput(),
			},
			seriErr:   axongo.ErrInputUTXORefsNotUnique,
			deseriErr: axongo.ErrInputUTXORefsNotUnique,
		},
	}

	for _, test := range tests {
		stx := tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI, &axongo.Transaction{
			API: tpkg.ZeroCostTestAPI,
			TransactionEssence: &axongo.TransactionEssence{
				Allotments:    axongo.Allotments{},
				ContextInputs: axongo.TxEssenceContextInputs{},
				Capabilities:  axongo.TransactionCapabilitiesBitMaskWithCapabilities(),
				NetworkID:     tpkg.ZeroCostTestAPI.ProtocolParameters().NetworkID(),
				Inputs:        test.inputs,
			},
			Outputs: axongo.TxEssenceOutputs{
				tpkg.RandBasicOutput(),
			},
		})

		tst := &frameworks.DeSerializeTest{
			Name:      test.name,
			Source:    stx,
			Target:    &axongo.SignedTransaction{},
			SeriErr:   test.seriErr,
			DeSeriErr: test.deseriErr,
		}

		t.Run(tst.Name, tst.Run)
	}
}

func TestTransactionContextInputLexicalOrderAndUniqueness(t *testing.T) {
	type test struct {
		name          string
		contextInputs axongo.TxEssenceContextInputs
		wantErr       error
	}

	accountID1 := axongo.MustAccountIDFromHexString("0x2668778ef0362d601c36ea05c742185daa1740dfcdaee0acfde6a9806a1f2ed2")
	accountID2 := axongo.MustAccountIDFromHexString("0x4e7cb233943cd31a0e3bd8b92668778ef0362d601c36ea05c742039ec582b0af")
	commitmentID1 := axongo.MustCommitmentIDFromHexString("0x3f34a869f47f8454e7cb233943cd31a0e3bd8b9551b1390039ec582b0a196856e50500fd")
	commitmentID2 := axongo.MustCommitmentIDFromHexString("0x90039ec582b0a196856e50500fd3f34a869f47f8454e7cb233943cd31a0e3bd8b9551ac4")

	tests := []test{
		{
			name: "ok - context inputs lexically ordered and unique",
			contextInputs: axongo.TxEssenceContextInputs{
				&axongo.CommitmentInput{
					CommitmentID: commitmentID1,
				}, // type 0
				&axongo.BlockIssuanceCreditInput{
					AccountID: accountID1,
				}, // type 1
				&axongo.BlockIssuanceCreditInput{
					AccountID: accountID2,
				}, // type 1
				&axongo.RewardInput{
					Index: 0,
				}, // type 2
				&axongo.RewardInput{
					Index: 1,
				}, // type 2
			},
			wantErr: nil,
		},
		{
			name: "fail - context inputs lexically unordered",
			contextInputs: axongo.TxEssenceContextInputs{
				&axongo.CommitmentInput{
					CommitmentID: commitmentID1,
				},
				&axongo.BlockIssuanceCreditInput{
					AccountID: accountID1,
				}, // type 1
				&axongo.CommitmentInput{
					CommitmentID: commitmentID1,
				}, // type 0
				&axongo.RewardInput{
					Index: 0,
				}, // type 2
			},
			wantErr: axongo.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - block issuance credits inputs lexically unordered",
			contextInputs: axongo.TxEssenceContextInputs{
				&axongo.CommitmentInput{
					CommitmentID: commitmentID1,
				},
				&axongo.BlockIssuanceCreditInput{
					AccountID: accountID2,
				}, // type 1
				&axongo.BlockIssuanceCreditInput{
					AccountID: accountID1,
				}, // type 1
			},
			wantErr: axongo.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - reward inputs lexically unordered",
			contextInputs: axongo.TxEssenceContextInputs{
				&axongo.CommitmentInput{
					CommitmentID: commitmentID1,
				},
				&axongo.RewardInput{
					Index: 1,
				}, // type 2
				&axongo.RewardInput{
					Index: 0,
				}, // type 2
			},
			wantErr: axongo.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - duplicate block issuance credit inputs",
			contextInputs: axongo.TxEssenceContextInputs{
				&axongo.CommitmentInput{
					CommitmentID: commitmentID1,
				},
				&axongo.BlockIssuanceCreditInput{
					AccountID: accountID1,
				}, // type 1
				&axongo.BlockIssuanceCreditInput{
					AccountID: accountID1,
				}, // type 1
			},
			wantErr: axongo.ErrArrayValidationViolatesUniqueness,
		},
		{
			name: "fail - duplicate reward inputs",
			contextInputs: axongo.TxEssenceContextInputs{
				&axongo.CommitmentInput{
					CommitmentID: commitmentID1,
				},
				&axongo.RewardInput{
					Index: 0,
				}, // type 2
				&axongo.RewardInput{
					Index: 0,
				}, // type 2
			},
			wantErr: axongo.ErrArrayValidationViolatesUniqueness,
		},
		{
			// At most one commitment input is allowed.
			name: "fail - duplicate commitment inputs",
			contextInputs: axongo.TxEssenceContextInputs{
				&axongo.CommitmentInput{
					CommitmentID: commitmentID2,
				}, // type 0
				&axongo.CommitmentInput{
					CommitmentID: commitmentID1,
				}, // type 0
				&axongo.RewardInput{
					Index: 1,
				}, // type 2
			},
			wantErr: axongo.ErrArrayValidationViolatesUniqueness,
		},
		{
			name: "fail - block issuance credit input without commitment input",
			contextInputs: axongo.TxEssenceContextInputs{
				&axongo.BlockIssuanceCreditInput{
					AccountID: accountID1,
				},
			},
			wantErr: axongo.ErrCommitmentInputMissing,
		},
		{
			name: "fail - reward input without commitment input",
			contextInputs: axongo.TxEssenceContextInputs{
				&axongo.RewardInput{
					Index: 0,
				},
			},
			wantErr: axongo.ErrCommitmentInputMissing,
		},
	}

	for _, test := range tests {
		// We need to build the transaction manually, since the builder and rand funcs would sort the context inputs.
		tx := &axongo.Transaction{
			API: tpkg.ZeroCostTestAPI,
			TransactionEssence: &axongo.TransactionEssence{
				Allotments:   axongo.Allotments{},
				Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(),
				NetworkID:    tpkg.ZeroCostTestAPI.ProtocolParameters().NetworkID(),
				CreationSlot: 5,
				Inputs: axongo.TxEssenceInputs{
					tpkg.RandUTXOInput(),
					tpkg.RandUTXOInput(),
					tpkg.RandUTXOInput(),
				},
				ContextInputs: test.contextInputs,
			},
			Outputs: axongo.TxEssenceOutputs{
				tpkg.RandBasicOutput(),
			},
		}

		stx := tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI, tx)

		tst := &frameworks.DeSerializeTest{
			Name:      test.name,
			Source:    stx,
			Target:    &axongo.SignedTransaction{},
			SeriErr:   test.wantErr,
			DeSeriErr: test.wantErr,
		}

		t.Run(test.name, tst.Run)
	}
}

type transactionSerializeTest struct {
	name      string
	output    axongo.Output
	seriErr   error
	deseriErr error
}

func (test *transactionSerializeTest) ToDeserializeTest() *frameworks.DeSerializeTest {
	_, addr, addrKeys := tpkg.RandEd25519Identity()

	txBuilder := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, axongo.NewInMemoryAddressSigner(addrKeys))
	txBuilder.WithTransactionCapabilities(
		axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanBurnNativeTokens(true)),
	)

	txBuilder.AddInput(&builder.TxInput{
		UnlockTarget: addr,
		InputID:      tpkg.RandUTXOInput().OutputID(),
		Input:        tpkg.RandBasicOutput(),
	})

	txBuilder.AddOutput(test.output)

	tx := lo.PanicOnErr(txBuilder.Build())

	return &frameworks.DeSerializeTest{
		Name:      test.name,
		Source:    tx,
		Target:    &axongo.SignedTransaction{},
		SeriErr:   test.seriErr,
		DeSeriErr: test.deseriErr,
	}
}

// Tests that lexical order & uniqueness are checked for unlock conditions across all relevant outputs.
func TestTransactionOutputUnlockConditionsLexicalOrderAndUniqueness(t *testing.T) {
	addressUnlockCond := &axongo.AddressUnlockCondition{
		Address: tpkg.RandEd25519Address(),
	}
	addressUnlockCond2 := &axongo.AddressUnlockCondition{
		Address: tpkg.RandEd25519Address(),
	}
	stateCtrlUnlockCond := &axongo.StateControllerAddressUnlockCondition{
		Address: tpkg.RandEd25519Address(),
	}
	govUnlockCond := &axongo.GovernorAddressUnlockCondition{
		Address: tpkg.RandEd25519Address(),
	}
	immAccUnlockCond := &axongo.ImmutableAccountUnlockCondition{
		Address: tpkg.RandAccountAddress(),
	}
	immAccUnlockCond2 := &axongo.ImmutableAccountUnlockCondition{
		Address: tpkg.RandAccountAddress(),
	}

	timelockUnlockCond := &axongo.TimelockUnlockCondition{Slot: 1337}
	timelockUnlockCond2 := &axongo.TimelockUnlockCondition{Slot: 1000}

	expirationUnlockCond := &axongo.ExpirationUnlockCondition{
		ReturnAddress: tpkg.RandEd25519Address(),
		Slot:          1000,
	}

	tests := []transactionSerializeTest{
		{
			name: "fail - BasicOutput contains lexically unordered unlock conditions",
			output: &axongo.BasicOutput{
				Amount: 10_000_000,
				UnlockConditions: axongo.BasicOutputUnlockConditions{
					addressUnlockCond,    // type 0
					expirationUnlockCond, // type 3
					timelockUnlockCond,   // type 2
				},
				Features: axongo.BasicOutputFeatures{},
			},
			seriErr:   axongo.ErrArrayValidationOrderViolatesLexicalOrder,
			deseriErr: axongo.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - AnchorOutput contains lexically unordered unlock conditions",
			output: &axongo.AnchorOutput{
				Amount: 10_000_000,
				UnlockConditions: axongo.AnchorOutputUnlockConditions{
					govUnlockCond,       // type 5
					stateCtrlUnlockCond, // type 4
				},
				Features: axongo.AnchorOutputFeatures{},
			},
			seriErr:   axongo.ErrArrayValidationOrderViolatesLexicalOrder,
			deseriErr: axongo.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - NFTOutput contains lexically unordered unlock conditions",
			output: &axongo.NFTOutput{
				Amount: 10_000_000,
				UnlockConditions: axongo.NFTOutputUnlockConditions{
					addressUnlockCond,    // type 0
					expirationUnlockCond, // type 3
					timelockUnlockCond,   // type 2
				},
				Features: axongo.NFTOutputFeatures{},
			},
			seriErr:   axongo.ErrArrayValidationOrderViolatesLexicalOrder,
			deseriErr: axongo.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - BasicOutput contains duplicate unlock conditions",
			output: &axongo.BasicOutput{
				Amount: 10_000_000,
				UnlockConditions: axongo.BasicOutputUnlockConditions{
					addressUnlockCond,   // type 0
					timelockUnlockCond,  // type 2
					timelockUnlockCond2, // type 2
				},
			},
			seriErr:   axongo.ErrArrayValidationViolatesUniqueness,
			deseriErr: axongo.ErrArrayValidationViolatesUniqueness,
		},
		{
			name: "fail - AccountOutput contains duplicate unlock conditions",
			output: &axongo.AccountOutput{
				Amount: 10_000_000,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					addressUnlockCond,  // type 0
					addressUnlockCond2, // type 0
				},
			},
			seriErr: axongo.ErrArrayValidationViolatesUniqueness,
			// During decoding, we encounter the max size error before the custom validator runs.
			deseriErr: serializer.ErrArrayValidationMaxElementsExceeded,
		},
		{
			name: "fail - AnchorOutput contains duplicate unlock conditions",
			output: &axongo.AnchorOutput{
				Amount: 10_000_000,
				UnlockConditions: axongo.AnchorOutputUnlockConditions{
					stateCtrlUnlockCond, // type 4
					stateCtrlUnlockCond, // type 4
					govUnlockCond,       // type 5
				},
				Features: axongo.AnchorOutputFeatures{},
			},
			seriErr: axongo.ErrArrayValidationViolatesUniqueness,
			// During decoding, we encounter the max size error before the custom validator runs.
			deseriErr: serializer.ErrArrayValidationMaxElementsExceeded,
		},
		{
			name: "fail - FoundryOutput contains duplicate unlock conditions",
			output: &axongo.FoundryOutput{
				Amount:      10_000_000,
				TokenScheme: tpkg.RandTokenScheme(),
				UnlockConditions: axongo.FoundryOutputUnlockConditions{
					immAccUnlockCond,  // type 6
					immAccUnlockCond2, // type 6
				},
				Features: axongo.FoundryOutputFeatures{},
			},
			seriErr:   axongo.ErrArrayValidationViolatesUniqueness,
			deseriErr: serializer.ErrArrayValidationMaxElementsExceeded,
		},
		{
			name: "fail - NFTOutput contains duplicate unlock conditions",
			output: &axongo.NFTOutput{
				Amount: 10_000_000,
				UnlockConditions: axongo.NFTOutputUnlockConditions{
					addressUnlockCond,   // type 0
					timelockUnlockCond,  // type 2
					timelockUnlockCond2, // type 2
				},
			},
			seriErr:   axongo.ErrArrayValidationViolatesUniqueness,
			deseriErr: axongo.ErrArrayValidationViolatesUniqueness,
		},
		{
			name: "fail - DelegationOutput contains duplicate unlock conditions",
			output: &axongo.DelegationOutput{
				Amount:           10_000_000,
				ValidatorAddress: tpkg.RandAccountAddress(),
				UnlockConditions: axongo.DelegationOutputUnlockConditions{
					addressUnlockCond,  // type 0
					addressUnlockCond2, // type 0
				},
			},
			seriErr: axongo.ErrArrayValidationViolatesUniqueness,
			// During decoding, we encounter the max size error before the custom validator runs.
			deseriErr: serializer.ErrArrayValidationMaxElementsExceeded,
		},
	}

	for _, test := range tests {
		t.Run(test.name, test.ToDeserializeTest().Run)
	}
}

// Tests that lexical order & uniqueness are checked for features across all relevant outputs.
func TestTransactionOutputFeatureLexicalOrderAndUniqueness(t *testing.T) {
	addressUnlockCond := &axongo.AddressUnlockCondition{
		Address: tpkg.RandEd25519Address(),
	}
	immutableAccountAddressUnlockCond := &axongo.ImmutableAccountUnlockCondition{
		Address: tpkg.RandAccountAddress(),
	}
	stateCtrlUnlockCond := &axongo.StateControllerAddressUnlockCondition{
		Address: tpkg.RandEd25519Address(),
	}
	govUnlockCond := &axongo.GovernorAddressUnlockCondition{
		Address: tpkg.RandEd25519Address(),
	}

	senderFeat := &axongo.SenderFeature{
		Address: tpkg.RandEd25519Address(),
	}
	senderFeat2 := &axongo.SenderFeature{
		Address: tpkg.RandEd25519Address(),
	}

	metadataFeat := &axongo.MetadataFeature{
		Entries: axongo.MetadataFeatureEntries{
			"key": []byte("val"),
		},
	}
	metadataFeat2 := &axongo.MetadataFeature{
		Entries: axongo.MetadataFeatureEntries{
			"entry": []byte("theval"),
		},
	}

	stateMetadataFeat := &axongo.StateMetadataFeature{
		Entries: axongo.StateMetadataFeatureEntries{
			"key": []byte("value"),
		},
	}

	tagFeat := &axongo.TagFeature{
		Tag: tpkg.RandBytes(3),
	}
	tagFeat2 := &axongo.TagFeature{
		Tag: tpkg.RandBytes(6),
	}

	nativeTokenFeat := tpkg.RandNativeTokenFeature()

	tests := []transactionSerializeTest{
		{
			name: "fail - BasicOutput contains lexically unordered features",
			output: &axongo.BasicOutput{
				Amount: 1337,
				Mana:   500,
				UnlockConditions: axongo.BasicOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: axongo.BasicOutputFeatures{
					tagFeat,    // type 4
					senderFeat, // type 0
				},
			},
			seriErr:   axongo.ErrArrayValidationOrderViolatesLexicalOrder,
			deseriErr: axongo.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - AccountOutput contains lexically unordered features",
			output: &axongo.AccountOutput{
				Amount: 1_000_000,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					addressUnlockCond,
				},
				Features: axongo.AccountOutputFeatures{
					metadataFeat, // type 2
					senderFeat,   // type 0
				},
			},
			seriErr:   axongo.ErrArrayValidationOrderViolatesLexicalOrder,
			deseriErr: axongo.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - AnchorOutput contains lexically unordered features",
			output: &axongo.AnchorOutput{
				Amount: 1_000_000,
				UnlockConditions: axongo.AnchorOutputUnlockConditions{
					stateCtrlUnlockCond,
					govUnlockCond,
				},
				Features: axongo.AnchorOutputFeatures{
					stateMetadataFeat, // type 3
					metadataFeat,      // type 2
				},
			},
			seriErr:   axongo.ErrArrayValidationOrderViolatesLexicalOrder,
			deseriErr: axongo.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - FoundryOutput contains lexically unordered features",
			output: &axongo.FoundryOutput{
				Amount:      1_000_000,
				TokenScheme: tpkg.RandTokenScheme(),
				UnlockConditions: axongo.FoundryOutputUnlockConditions{
					immutableAccountAddressUnlockCond,
				},
				Features: axongo.FoundryOutputFeatures{
					nativeTokenFeat, // type 5
					metadataFeat,    // type 2
				},
			},
			seriErr:   axongo.ErrArrayValidationOrderViolatesLexicalOrder,
			deseriErr: axongo.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - NFTOutput contains lexically unordered features",
			output: &axongo.NFTOutput{
				Amount: 1337,
				UnlockConditions: axongo.NFTOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: axongo.NFTOutputFeatures{
					tagFeat,    // type 4
					senderFeat, // type 0
				},
			},
			seriErr:   axongo.ErrArrayValidationOrderViolatesLexicalOrder,
			deseriErr: axongo.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - BasicOutput contains duplicate features",
			output: &axongo.BasicOutput{
				Amount: 1337,
				UnlockConditions: axongo.BasicOutputUnlockConditions{
					addressUnlockCond,
				},
				Features: axongo.BasicOutputFeatures{
					tagFeat,  // type 4
					tagFeat2, // type 4
				},
			},
			seriErr:   axongo.ErrArrayValidationViolatesUniqueness,
			deseriErr: axongo.ErrArrayValidationViolatesUniqueness,
		},
		{
			name: "fail - AccountOutput contains duplicate features",
			output: &axongo.AccountOutput{
				Amount: 1337,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					addressUnlockCond,
				},
				Features: axongo.AccountOutputFeatures{
					senderFeat,  // type 0
					senderFeat2, // type 0
				},
			},
			seriErr:   axongo.ErrArrayValidationViolatesUniqueness,
			deseriErr: axongo.ErrArrayValidationViolatesUniqueness,
		},
		{
			name: "fail - AnchorOutput contains duplicate features",
			output: &axongo.AnchorOutput{
				Amount: 1337,
				UnlockConditions: axongo.AnchorOutputUnlockConditions{
					stateCtrlUnlockCond,
					govUnlockCond,
				},
				Features: axongo.AnchorOutputFeatures{
					metadataFeat,  // type 2
					metadataFeat2, // type 2
				},
			},
			seriErr:   axongo.ErrArrayValidationViolatesUniqueness,
			deseriErr: axongo.ErrArrayValidationViolatesUniqueness,
		},
		{
			name: "fail - FoundryOutput contains duplicate features",
			output: &axongo.FoundryOutput{
				Amount:      1_000_000,
				TokenScheme: tpkg.RandTokenScheme(),
				UnlockConditions: axongo.FoundryOutputUnlockConditions{
					immutableAccountAddressUnlockCond,
				},
				Features: axongo.FoundryOutputFeatures{
					metadataFeat,  // type 2
					metadataFeat2, // type 2
				},
			},
			seriErr:   axongo.ErrArrayValidationViolatesUniqueness,
			deseriErr: axongo.ErrArrayValidationViolatesUniqueness,
		},
		{
			name: "fail - NFTOutput contains duplicate features",
			output: &axongo.NFTOutput{
				Amount: 1337,
				UnlockConditions: axongo.NFTOutputUnlockConditions{
					addressUnlockCond,
				},
				Features: axongo.NFTOutputFeatures{
					tagFeat,  // type 4
					tagFeat2, // type 4
				},
			},
			seriErr:   axongo.ErrArrayValidationViolatesUniqueness,
			deseriErr: axongo.ErrArrayValidationViolatesUniqueness,
		},
	}

	for _, test := range tests {
		t.Run(test.name, test.ToDeserializeTest().Run)
	}
}

// Tests that lexical order & uniqueness are checked for immutable features across all relevant outputs.
func TestTransactionOutputImmutableFeatureLexicalOrderAndUniqueness(t *testing.T) {
	addressUnlockCond := &axongo.AddressUnlockCondition{
		Address: tpkg.RandEd25519Address(),
	}
	stateCtrlUnlockCond := &axongo.StateControllerAddressUnlockCondition{
		Address: tpkg.RandEd25519Address(),
	}
	govUnlockCond := &axongo.GovernorAddressUnlockCondition{
		Address: tpkg.RandEd25519Address(),
	}

	issuerFeat := &axongo.IssuerFeature{
		Address: tpkg.RandEd25519Address(),
	}
	// Create a second issuer feature to ensure uniqueness is checked based on the type of the feature.
	issuerFeat2 := &axongo.IssuerFeature{
		Address: tpkg.RandEd25519Address(),
	}

	metadataFeat := &axongo.MetadataFeature{
		Entries: axongo.MetadataFeatureEntries{
			"key": []byte("val"),
		},
	}
	metadataFeat2 := &axongo.MetadataFeature{
		Entries: axongo.MetadataFeatureEntries{
			"key": []byte("value"),
		},
	}

	tests := []transactionSerializeTest{
		{
			name: "fail - AccountOutput contains lexically unordered immutable features",
			output: &axongo.AccountOutput{
				Amount: 1_000_000,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					addressUnlockCond,
				},
				ImmutableFeatures: axongo.AccountOutputImmFeatures{
					metadataFeat, // type 2
					issuerFeat,   // type 1
				},
			},
			seriErr:   axongo.ErrArrayValidationOrderViolatesLexicalOrder,
			deseriErr: axongo.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - AnchorOutput contains lexically unordered immutable features",
			output: &axongo.AnchorOutput{
				Amount: 1_000_000,
				UnlockConditions: axongo.AnchorOutputUnlockConditions{
					stateCtrlUnlockCond,
					govUnlockCond,
				},
				ImmutableFeatures: axongo.AnchorOutputImmFeatures{
					metadataFeat, // type 2
					issuerFeat,   // type 1
				},
			},
			seriErr:   axongo.ErrArrayValidationOrderViolatesLexicalOrder,
			deseriErr: axongo.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - NFTOutput contains lexically unordered immutable features",
			output: &axongo.NFTOutput{
				Amount: 1_000_000,
				UnlockConditions: axongo.NFTOutputUnlockConditions{
					addressUnlockCond,
				},
				ImmutableFeatures: axongo.NFTOutputImmFeatures{
					metadataFeat, // type 2
					issuerFeat,   // type 1
				},
			},
			seriErr:   axongo.ErrArrayValidationOrderViolatesLexicalOrder,
			deseriErr: axongo.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - AccountOutput contains duplicate immutable features",
			output: &axongo.AccountOutput{
				Amount: 1_000_000,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					addressUnlockCond,
				},
				ImmutableFeatures: axongo.AccountOutputImmFeatures{
					issuerFeat,  // type 1
					issuerFeat2, // type 1
				},
			},
			seriErr:   axongo.ErrArrayValidationViolatesUniqueness,
			deseriErr: axongo.ErrArrayValidationViolatesUniqueness,
		},
		{
			name: "fail - AnchorOutput contains duplicate immutable features",
			output: &axongo.AnchorOutput{
				Amount: 1_000_000,
				UnlockConditions: axongo.AnchorOutputUnlockConditions{
					stateCtrlUnlockCond,
					govUnlockCond,
				},
				ImmutableFeatures: axongo.AnchorOutputImmFeatures{
					issuerFeat,  // type 1
					issuerFeat2, // type 1
				},
			},
			seriErr:   axongo.ErrArrayValidationViolatesUniqueness,
			deseriErr: axongo.ErrArrayValidationViolatesUniqueness,
		},
		{
			name: "fail - FoundryOutput contains duplicate immutable features",
			output: &axongo.FoundryOutput{
				Amount:      1_000_000,
				TokenScheme: tpkg.RandTokenScheme(),
				UnlockConditions: axongo.FoundryOutputUnlockConditions{
					&axongo.ImmutableAccountUnlockCondition{
						Address: tpkg.RandAccountAddress(),
					},
				},
				ImmutableFeatures: axongo.FoundryOutputImmFeatures{
					metadataFeat,  // type 2
					metadataFeat2, // type 2
				},
			},
			seriErr: axongo.ErrArrayValidationViolatesUniqueness,
			// During decoding, we encounter the max size error before the custom validator runs.
			deseriErr: serializer.ErrArrayValidationMaxElementsExceeded,
		},
		{
			name: "fail - NFTOutput contains duplicate immutable features",
			output: &axongo.NFTOutput{
				Amount: 1_000_000,
				UnlockConditions: axongo.NFTOutputUnlockConditions{
					addressUnlockCond,
				},
				ImmutableFeatures: axongo.NFTOutputImmFeatures{
					issuerFeat,  // type 1
					issuerFeat2, // type 1
				},
			},
			seriErr:   axongo.ErrArrayValidationViolatesUniqueness,
			deseriErr: axongo.ErrArrayValidationViolatesUniqueness,
		},
	}

	for _, test := range tests {
		t.Run(test.name, test.ToDeserializeTest().Run)
	}
}

// Helper struct for testing JSON encoding, since slices cannot be serialized directly.
type transactionIDTestHelper struct {
	IDs axongo.TransactionIDs `serix:""`
}

// Tests that lexical order & uniqueness are checked for TransactionIDs.
func TestTransactionIDsLexicalOrderAndUniqueness(t *testing.T) {
	txID1 := axongo.MustTransactionIDFromHexString("0x8f63d1473a0417e89d01c5174ac5802402f2a49159cad1de811786367da7db3d0a0d3d78")
	txID2 := axongo.MustTransactionIDFromHexString("0xc988b403f48b71adbd0a0dba3b2c0665283f8c3290028e220eab35d1c86c60f747eb2624")
	txID3 := axongo.MustTransactionIDFromHexString("0xfe25a362ae9483819ec35387a47476408e7a65d868651832d7714935fd5ca7596aa8828b")

	tests := []*frameworks.DeSerializeTest{
		{
			Name: "ok - transaction ids lexically ordered and unique",
			Source: &transactionIDTestHelper{
				IDs: axongo.TransactionIDs{
					txID1,
					txID2,
					txID3,
				},
			},
			Target: &transactionIDTestHelper{},
		},
		{
			Name: "fail - transaction ids lexically unordered",
			Source: &transactionIDTestHelper{
				IDs: axongo.TransactionIDs{
					txID1,
					txID3,
					txID2,
				},
			},
			Target:    &transactionIDTestHelper{},
			SeriErr:   axongo.ErrArrayValidationOrderViolatesLexicalOrder,
			DeSeriErr: axongo.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			Name: "fail - transaction ids contains duplicates",
			Source: &transactionIDTestHelper{
				IDs: axongo.TransactionIDs{
					txID1,
					txID2,
					txID2,
				},
			},
			Target:    &transactionIDTestHelper{},
			SeriErr:   axongo.ErrArrayValidationViolatesUniqueness,
			DeSeriErr: axongo.ErrArrayValidationViolatesUniqueness,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

func TestCommitmentInputSyntacticalValidation(t *testing.T) {
	accountWithFeatures := func(feats axongo.AccountOutputFeatures) *axongo.AccountOutput {
		return &axongo.AccountOutput{
			Amount: 100_000_000,
			UnlockConditions: axongo.AccountOutputUnlockConditions{
				&axongo.AddressUnlockCondition{
					Address: tpkg.RandAccountAddress(),
				},
			},
			ImmutableFeatures: axongo.AccountOutputImmFeatures{},
			Features:          feats,
		}
	}

	tests := []*frameworks.DeSerializeTest{
		// fail - BlockIssuerFeature on output side without Commitment Input
		{
			Name: "fail - BlockIssuerFeature on output side without Commitment Input",
			Source: tpkg.RandSignedTransaction(tpkg.ZeroCostTestAPI, func(t *axongo.Transaction) {
				t.Outputs = axongo.TxEssenceOutputs{
					accountWithFeatures(
						axongo.AccountOutputFeatures{
							&axongo.BlockIssuerFeature{
								ExpirySlot:      100,
								BlockIssuerKeys: tpkg.RandBlockIssuerKeys(3),
							},
						},
					),
				}
				// Make sure there are no Context Inputs added by the rand function for this test.
				t.TransactionEssence.ContextInputs = nil
			}),
			Target:    &axongo.SignedTransaction{},
			SeriErr:   axongo.ErrBlockIssuerCommitmentInputMissing,
			DeSeriErr: axongo.ErrBlockIssuerCommitmentInputMissing,
		},
		// fail - StakingFeature on output side without Commitment Input
		{
			Name: "fail - StakingFeature on output side without Commitment Input",
			Source: tpkg.RandSignedTransaction(tpkg.ZeroCostTestAPI, func(t *axongo.Transaction) {
				t.Outputs = axongo.TxEssenceOutputs{
					accountWithFeatures(
						axongo.AccountOutputFeatures{
							&axongo.BlockIssuerFeature{
								ExpirySlot:      100,
								BlockIssuerKeys: tpkg.RandBlockIssuerKeys(3),
							},
							&axongo.StakingFeature{
								StakedAmount: 1,
								FixedCost:    1,
								StartEpoch:   10,
								EndEpoch:     12,
							},
						},
					),
				}
				// Make sure there are no Context Inputs added by the rand function for this test.
				t.TransactionEssence.ContextInputs = nil
			}),
			Target:    &axongo.SignedTransaction{},
			SeriErr:   axongo.ErrStakingCommitmentInputMissing,
			DeSeriErr: axongo.ErrStakingCommitmentInputMissing,
		},
		// fail - Delegation Output on output side without Commitment Input
		{
			Name: "fail - Delegation Output on output side without Commitment Input",
			Source: tpkg.RandSignedTransaction(tpkg.ZeroCostTestAPI, func(t *axongo.Transaction) {
				t.Outputs = axongo.TxEssenceOutputs{
					&axongo.DelegationOutput{
						Amount:           10,
						DelegatedAmount:  10,
						DelegationID:     tpkg.RandDelegationID(),
						ValidatorAddress: tpkg.RandAccountAddress(),
						StartEpoch:       10,
						EndEpoch:         12,
						UnlockConditions: axongo.DelegationOutputUnlockConditions{
							&axongo.AddressUnlockCondition{
								Address: tpkg.RandEd25519Address(),
							},
						},
					},
				}
				// Make sure there are no Context Inputs added by the rand function for this test.
				t.TransactionEssence.ContextInputs = nil
			}),
			Target:    &axongo.SignedTransaction{},
			SeriErr:   axongo.ErrDelegationCommitmentInputMissing,
			DeSeriErr: axongo.ErrDelegationCommitmentInputMissing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

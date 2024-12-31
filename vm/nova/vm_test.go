//nolint:forcetypeassert,dupl,nlreturn,revive
package nova_test

import (
	"bytes"
	"crypto/ed25519"
	"math/big"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/axonfibre/fibre.go/lo"
	"github.com/axonfibre/fibre.go/serializer/v2/serix"
	axongo "github.com/axonfibre/axon.go/v4"
	"github.com/axonfibre/axon.go/v4/builder"
	"github.com/axonfibre/axon.go/v4/tpkg"
	"github.com/axonfibre/axon.go/v4/vm"
	"github.com/axonfibre/axon.go/v4/vm/nova"
)

const (
	OneIOTA axongo.BaseToken = 1_000_000
)

var (
	novaVM = nova.NewVirtualMachine()

	testProtoParams = tpkg.IOTAMainnetV3TestProtocolParameters

	testAPI = axongo.V3API(testProtoParams)
)

func TestNFTTransition(t *testing.T) {
	_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()

	inputIDs := tpkg.RandOutputIDs(1)
	inputs := vm.InputSet{
		inputIDs[0]: &axongo.NFTOutput{
			Amount: OneIOTA,
			NFTID:  axongo.NFTID{},
			UnlockConditions: axongo.NFTOutputUnlockConditions{
				&axongo.AddressUnlockCondition{Address: addr1},
			},
			Features: nil,
		},
	}

	nftAddr := axongo.NFTAddressFromOutputID(inputIDs[0])
	nftID := nftAddr.NFTID()

	transaction := &axongo.Transaction{
		API: testAPI,
		TransactionEssence: &axongo.TransactionEssence{
			Inputs:       inputIDs.UTXOInputs(),
			Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
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
		},
	}

	sigs, err := transaction.Sign(addr1AddrKeys)
	require.NoError(t, err)

	tx := &axongo.SignedTransaction{
		API:         testAPI,
		Transaction: transaction,
		Unlocks: axongo.Unlocks{
			&axongo.SignatureUnlock{Signature: sigs[0]},
		},
	}

	require.NoError(t, validateAndExecuteSignedTransaction(tx, vm.ResolvedInputs{InputSet: inputs}))
}

func TestCirculatingSupplyMelting(t *testing.T) {
	_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
	accountaddr1 := tpkg.RandAccountAddress()

	inputIDs := tpkg.RandOutputIDs(3)
	inputs := vm.InputSet{
		inputIDs[0]: &axongo.BasicOutput{
			Amount: OneIOTA,
			UnlockConditions: axongo.BasicOutputUnlockConditions{
				&axongo.AddressUnlockCondition{Address: addr1},
			},
		},
		inputIDs[1]: &axongo.AccountOutput{
			Amount:         OneIOTA,
			AccountID:      accountaddr1.AccountID(),
			FoundryCounter: 1,
			UnlockConditions: axongo.AccountOutputUnlockConditions{
				&axongo.AddressUnlockCondition{Address: addr1},
			},
			Features: nil,
		},
		inputIDs[2]: &axongo.FoundryOutput{
			Amount:       OneIOTA,
			SerialNumber: 1,
			TokenScheme: &axongo.SimpleTokenScheme{
				MintedTokens:  big.NewInt(50),
				MeltedTokens:  big.NewInt(0),
				MaximumSupply: big.NewInt(50),
			},
			UnlockConditions: axongo.FoundryOutputUnlockConditions{
				&axongo.ImmutableAccountUnlockCondition{Address: accountaddr1},
			},
			Features: nil,
		},
	}

	// set input BasicOutput NativeToken to 50 which get melted
	foundryNativeTokenID := inputs[inputIDs[2]].(*axongo.FoundryOutput).MustNativeTokenID()
	inputs[inputIDs[0]].(*axongo.BasicOutput).Features.Upsert(&axongo.NativeTokenFeature{
		ID:     foundryNativeTokenID,
		Amount: new(big.Int).SetInt64(50),
	})

	transaction := &axongo.Transaction{
		API: testAPI,
		TransactionEssence: &axongo.TransactionEssence{
			Inputs:       inputIDs.UTXOInputs(),
			Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
		},
		Outputs: axongo.TxEssenceOutputs{
			&axongo.AccountOutput{
				Amount:         OneIOTA,
				AccountID:      accountaddr1.AccountID(),
				FoundryCounter: 1,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: addr1},
				},
				Features: nil,
			},
			&axongo.FoundryOutput{
				Amount:       2 * OneIOTA,
				SerialNumber: 1,
				TokenScheme: &axongo.SimpleTokenScheme{
					MintedTokens:  big.NewInt(50),
					MeltedTokens:  big.NewInt(50),
					MaximumSupply: big.NewInt(50),
				},
				UnlockConditions: axongo.FoundryOutputUnlockConditions{
					&axongo.ImmutableAccountUnlockCondition{Address: accountaddr1},
				},
				Features: nil,
			},
		},
	}

	sigs, err := transaction.Sign(addr1AddrKeys)
	require.NoError(t, err)

	tx := &axongo.SignedTransaction{
		API:         testAPI,
		Transaction: transaction,
		Unlocks: axongo.Unlocks{
			&axongo.SignatureUnlock{Signature: sigs[0]},
			&axongo.ReferenceUnlock{Reference: 0},
			&axongo.AccountUnlock{Reference: 1},
		},
	}

	require.NoError(t, validateAndExecuteSignedTransaction(tx, vm.ResolvedInputs{InputSet: inputs}))
}

func TestNovaTransactionExecution(t *testing.T) {
	type test struct {
		name           string
		vmParams       *vm.Params
		resolvedInputs vm.ResolvedInputs
		tx             *axongo.SignedTransaction
		wantErr        error
	}
	tests := []*test{
		// ok
		func() *test {
			var (
				_, addr1, addr1AddrKeys = tpkg.RandEd25519Identity()
				_, addr2, addr2AddrKeys = tpkg.RandEd25519Identity()
				_, addr3, addr3AddrKeys = tpkg.RandEd25519Identity()
				_, addr4, addr4AddrKeys = tpkg.RandEd25519Identity()
				_, addr5, _             = tpkg.RandEd25519Identity()
			)

			var (
				defaultAmount        = OneIOTA
				storageDepositReturn = OneIOTA / 2
				nativeTokenTransfer1 = tpkg.RandNativeTokenFeature()
				nativeTokenTransfer2 = tpkg.RandNativeTokenFeature()
			)

			var (
				nft1ID = tpkg.Rand32ByteArray()
				nft2ID = tpkg.Rand32ByteArray()
			)

			inputIDs := tpkg.RandOutputIDs(18)

			account1InputID := inputIDs[6]

			account1AccountID := axongo.AccountIDFromOutputID(account1InputID)
			account1AccountAddress := account1AccountID.ToAddress().(*axongo.AccountAddress)

			anchor1InputID := inputIDs[8]
			anchor2InputID := inputIDs[9]

			anchor1AnchorID := axongo.AnchorIDFromOutputID(anchor1InputID)
			anchor2AnchorID := axongo.AnchorIDFromOutputID(anchor2InputID)

			foundry1InputID := inputIDs[11]
			foundry2InputID := inputIDs[12]
			foundry3InputID := inputIDs[13]
			foundry4InputID := inputIDs[14]

			nft1InputID := inputIDs[15]

			inputs := vm.InputSet{
				// basic output with no features [defaultAmount] (owned by addr1)
				// => output 0: change ownership to addr5
				inputIDs[0]: &axongo.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},

				// basic output with native token feature - nativeTokenTransfer1 [defaultAmount] (owned by addr2)
				// => output 1: change ownership to addr3
				inputIDs[1]: &axongo.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr2},
					},
					Features: axongo.BasicOutputFeatures{
						nativeTokenTransfer1,
					},
				},

				// basic output with native token feature - nativeTokenTransfer2 [defaultAmount] (owned by addr2)
				// => output 2: change ownership to addr4
				inputIDs[2]: &axongo.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr2},
					},
					Features: axongo.BasicOutputFeatures{
						nativeTokenTransfer2,
					},
				},

				// basic output with expiration unlock condition - slot: 500, return: addr1 [defaultAmount] (originally owned by addr2 => creation slot 750 => owned by addr1)
				// => output 3: remove expiration unlock condition
				inputIDs[3]: &axongo.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr2},
						&axongo.ExpirationUnlockCondition{
							ReturnAddress: addr1,
							Slot:          500,
						},
					},
				},

				// basic output with timelock unlock condition - slot: 500 [defaultAmount] (owned by addr2 => creation slot 750 => can be unlocked)
				// => output 4: remove timelock unlock condition
				inputIDs[4]: &axongo.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr2},
						&axongo.TimelockUnlockCondition{
							Slot: 500,
						},
					},
				},

				// basic output [defaultAmount + storageDepositReturn] (owned by addr2 => creation slot 750 => can be unlocked, owned by addr2)
				//					 storage deposit return unlock condition - return: addr1
				// 			       	 timelock unlock condition 				 - slot 500
				// 			       	 expiration unlock condition 			 - slot: 900, return: addr1
				// => output 5: storageDepositReturn to addr1
				// => output 14: defaultAmount
				inputIDs[5]: &axongo.BasicOutput{
					Amount: defaultAmount + storageDepositReturn,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr2},
						&axongo.StorageDepositReturnUnlockCondition{
							ReturnAddress: addr1,
							Amount:        storageDepositReturn,
						},
						&axongo.TimelockUnlockCondition{
							Slot: 500,
						},
						&axongo.ExpirationUnlockCondition{
							ReturnAddress: addr1,
							Slot:          900,
						},
					},
				},

				// account output with no features - foundry counter 5 [defaultAmount] (owned by addr3) => going to be transitioned
				// => output 6: output transition (foundry counter 5 => 6, added metadata)
				account1InputID: &axongo.AccountOutput{
					Amount:         defaultAmount,
					AccountID:      account1AccountID,
					FoundryCounter: 5,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr3},
					},
					Features: nil,
				},

				// account output with no features [defaultAmount] (owned by addr3) => going to be destroyed
				// => output 7: destroyed and new account output created
				inputIDs[7]: &axongo.AccountOutput{
					Amount:         defaultAmount,
					AccountID:      axongo.AccountID{},
					FoundryCounter: 0,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr3},
					},
					Features: nil,
				},

				// anchor output with no features - state index 0 [defaultAmount] (owned by - state: addr3, gov: addr4) => going to be governance transitioned
				// => output 8: governance transition (added metadata)
				anchor1InputID: &axongo.AnchorOutput{
					Amount:     defaultAmount,
					AnchorID:   anchor1AnchorID,
					StateIndex: 0,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: addr3},
						&axongo.GovernorAddressUnlockCondition{Address: addr4},
					},
					Features: axongo.AnchorOutputFeatures{
						&axongo.StateMetadataFeature{Entries: axongo.StateMetadataFeatureEntries{"data": []byte("gov transitioning")}},
					},
				},

				// anchor output with no features - state index 5 [defaultAmount] (owned by - state: addr3, gov: addr4) => going to be state transitioned
				// => output 9: state transition (state index 5 => 6, changed state metadata)
				anchor2InputID: &axongo.AnchorOutput{
					Amount:     defaultAmount,
					AnchorID:   anchor2AnchorID,
					StateIndex: 5,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: addr3},
						&axongo.GovernorAddressUnlockCondition{Address: addr4},
					},
					Features: axongo.AnchorOutputFeatures{
						&axongo.StateMetadataFeature{Entries: axongo.StateMetadataFeatureEntries{"data": []byte("state transitioning")}},
					},
				},

				// anchor output with no features - state index 0 [defaultAmount] (owned by - state: addr3, gov: addr3) => going to be destroyed
				// => output 10: destroyed and new anchor output created
				inputIDs[10]: &axongo.AnchorOutput{
					Amount:     defaultAmount,
					AnchorID:   axongo.AnchorID{},
					StateIndex: 0,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: addr3},
						&axongo.GovernorAddressUnlockCondition{Address: addr3},
					},
					Features: axongo.AnchorOutputFeatures{
						&axongo.StateMetadataFeature{Entries: axongo.StateMetadataFeatureEntries{"data": []byte("going to be destroyed")}},
					},
				},

				// foundry output - serialNumber: 1, minted: 100, melted: 0, max: 1000 [defaultAmount] (owned by account1AccountAddress)
				// => output 11: mint 100 new tokens
				foundry1InputID: &axongo.FoundryOutput{
					Amount:       defaultAmount,
					SerialNumber: 1,
					TokenScheme: &axongo.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(100),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
					UnlockConditions: axongo.FoundryOutputUnlockConditions{
						&axongo.ImmutableAccountUnlockCondition{Address: account1AccountAddress},
					},
					Features: nil,
				},

				// foundry output - serialNumber: 2, minted: 100, melted: 0, max: 1000 [defaultAmount] (owned by account1AccountAddress)
				//				  - native token balance later updated to 100 (still on input side)
				// => output 12: melt 50 tokens
				foundry2InputID: &axongo.FoundryOutput{
					Amount:       defaultAmount,
					SerialNumber: 2,
					TokenScheme: &axongo.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(100),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
					UnlockConditions: axongo.FoundryOutputUnlockConditions{
						&axongo.ImmutableAccountUnlockCondition{Address: account1AccountAddress},
					},
					Features: axongo.FoundryOutputFeatures{
						// native token feature added later
					},
				},

				// foundry output - serialNumber: 3, minted: 100, melted: 0, max: 1000 [defaultAmount] (owned by account1AccountAddress)
				// => output 13: add metadata
				foundry3InputID: &axongo.FoundryOutput{
					Amount:       defaultAmount,
					SerialNumber: 3,
					TokenScheme: &axongo.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(100),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
					UnlockConditions: axongo.FoundryOutputUnlockConditions{
						&axongo.ImmutableAccountUnlockCondition{Address: account1AccountAddress},
					},
					Features: nil,
				},

				// foundry output - serialNumber: 4, minted: 100, melted: 0, max: 1000 [defaultAmount] (owned by account1AccountAddress)
				//				  - native token balance later updated to 50 (still on input side)
				// => output 15: foundry destroyed
				foundry4InputID: &axongo.FoundryOutput{
					Amount:       defaultAmount,
					SerialNumber: 4,
					TokenScheme: &axongo.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(100),
						MeltedTokens:  big.NewInt(50),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
					UnlockConditions: axongo.FoundryOutputUnlockConditions{
						&axongo.ImmutableAccountUnlockCondition{Address: account1AccountAddress},
					},
					Features: nil,
				},

				// NFT output with issuer (addr3) and immutable metadata feature [defaultAmount] (owned by addr3) => going to be transferred to addr4
				// => output 16: transfer to addr4
				nft1InputID: &axongo.NFTOutput{
					Amount: defaultAmount,
					NFTID:  nft1ID,
					UnlockConditions: axongo.NFTOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr3},
					},
					Features: axongo.NFTOutputFeatures{},
					ImmutableFeatures: axongo.NFTOutputImmFeatures{
						&axongo.IssuerFeature{Address: addr3},
						&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("transfer to 4")}},
					},
				},

				// NFT output with immutable features [defaultAmount] (owned by addr4) => going to be destroyed
				// => output 17: destroyed and new NFT output created
				inputIDs[16]: &axongo.NFTOutput{
					Amount: defaultAmount,
					NFTID:  nft2ID,
					UnlockConditions: axongo.NFTOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr4},
					},
					Features: axongo.NFTOutputFeatures{},
					ImmutableFeatures: axongo.NFTOutputImmFeatures{
						&axongo.IssuerFeature{Address: addr3},
						&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("going to be destroyed")}},
					},
				},

				// basic output with no features [defaultAmount] (owned by nft1ID)
				// => output 18: change ownership to addr5
				inputIDs[17]: &axongo.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: axongo.NFTID(nft1ID).ToAddress()},
					},
				},
			}

			foundry1addr3NativeTokenID := inputs[foundry1InputID].(*axongo.FoundryOutput).MustNativeTokenID()
			foundry2addr3NativeTokenID := inputs[foundry2InputID].(*axongo.FoundryOutput).MustNativeTokenID()
			foundry4addr3NativeTokenID := inputs[foundry4InputID].(*axongo.FoundryOutput).MustNativeTokenID()

			inputs[foundry2InputID].(*axongo.FoundryOutput).Features.Upsert(&axongo.NativeTokenFeature{
				ID:     foundry2addr3NativeTokenID,
				Amount: big.NewInt(100),
			})

			inputs[foundry4InputID].(*axongo.FoundryOutput).Features.Upsert(&axongo.NativeTokenFeature{
				ID:     foundry4addr3NativeTokenID,
				Amount: big.NewInt(50),
			})

			// new foundry output - serialNumber: 6, minted: 100, melted: 0, max: 1000 (owned by account1AccountAddress)
			//					  - native token balance 100
			newFoundryWithInitialSupply := &axongo.FoundryOutput{
				Amount:       defaultAmount,
				SerialNumber: 6,
				TokenScheme: &axongo.SimpleTokenScheme{
					MintedTokens:  big.NewInt(100),
					MeltedTokens:  big.NewInt(0),
					MaximumSupply: new(big.Int).SetInt64(1000),
				},
				UnlockConditions: axongo.FoundryOutputUnlockConditions{
					&axongo.ImmutableAccountUnlockCondition{Address: account1AccountAddress},
				},
				Features: nil,
			}
			newFoundryNativeTokenID := newFoundryWithInitialSupply.MustNativeTokenID()
			newFoundryWithInitialSupply.Features.Upsert(&axongo.NativeTokenFeature{
				ID:     newFoundryNativeTokenID,
				Amount: big.NewInt(100),
			})

			creationSlot := axongo.SlotIndex(750)
			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs:       inputIDs.UTXOInputs(),
					CreationSlot: creationSlot,
					Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
				},
				Outputs: axongo.TxEssenceOutputs{
					// basic output [defaultAmount] (owned by addr5)
					// => input 0
					&axongo.BasicOutput{
						Amount: defaultAmount,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr5},
						},
					},

					// basic output with native token feature - nativeTokenTransfer1 [defaultAmount] (owned by addr3)
					// => input 1
					&axongo.BasicOutput{
						Amount: defaultAmount,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr3},
						},
						Features: axongo.BasicOutputFeatures{
							nativeTokenTransfer1,
						},
					},

					// basic output with native token feature - nativeTokenTransfer2 [defaultAmount] (owned by addr4)
					// => input 2
					&axongo.BasicOutput{
						Amount: defaultAmount,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr4},
						},
						Features: axongo.BasicOutputFeatures{
							nativeTokenTransfer2,
						},
					},

					// basic output [defaultAmount] (owned by addr2)
					// => input 3
					&axongo.BasicOutput{
						Amount: defaultAmount,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr2},
						},
					},

					// basic output [defaultAmount] (owned by addr2)
					// => input 4
					&axongo.BasicOutput{
						Amount: defaultAmount,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr2},
						},
					},

					// basic output [storageDepositReturn] (owned by addr1)
					// => input 5
					&axongo.BasicOutput{
						Amount: storageDepositReturn,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr1},
						},
					},

					// transitioned account output [defaultAmount] (owned by addr3)
					// => input 6
					&axongo.AccountOutput{
						Amount:         defaultAmount,
						AccountID:      account1AccountID,
						FoundryCounter: 6,
						UnlockConditions: axongo.AccountOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr3},
						},
						Features: axongo.AccountOutputFeatures{
							&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("transitioned")}},
						},
					},

					// new account output [defaultAmount] (owned by addr3)
					// => input 7
					&axongo.AccountOutput{
						Amount:         defaultAmount,
						AccountID:      axongo.AccountID{},
						FoundryCounter: 0,
						UnlockConditions: axongo.AccountOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr3},
						},
						Features: axongo.AccountOutputFeatures{
							&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("new")}},
						},
					},

					// governance transitioned anchor output [defaultAmount] (owned by - state: addr3, gov: addr4)
					// => input 8
					&axongo.AnchorOutput{
						Amount:     defaultAmount,
						AnchorID:   anchor1AnchorID,
						StateIndex: 0,
						UnlockConditions: axongo.AnchorOutputUnlockConditions{
							&axongo.StateControllerAddressUnlockCondition{Address: addr3},
							&axongo.GovernorAddressUnlockCondition{Address: addr4},
						},
						Features: axongo.AnchorOutputFeatures{
							&axongo.StateMetadataFeature{Entries: axongo.StateMetadataFeatureEntries{"data": []byte("gov transitioning")}},
							&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("the gov mutation on this output")}},
						},
					},

					// state transitioned anchor output [defaultAmount] (owned by - state: addr3, gov: addr4)
					// => input 9
					&axongo.AnchorOutput{
						Amount:     defaultAmount,
						AnchorID:   anchor2AnchorID,
						StateIndex: 6,
						UnlockConditions: axongo.AnchorOutputUnlockConditions{
							&axongo.StateControllerAddressUnlockCondition{Address: addr3},
							&axongo.GovernorAddressUnlockCondition{Address: addr4},
						},
						Features: axongo.AnchorOutputFeatures{
							&axongo.StateMetadataFeature{Entries: axongo.StateMetadataFeatureEntries{
								"data":  []byte("state transitioning"),
								"added": []byte("next state"),
							}},
						},
					},

					// new anchor output [defaultAmount] (owned by - state: addr3, gov: addr4)
					// => input 10
					&axongo.AnchorOutput{
						Amount:     defaultAmount,
						AnchorID:   axongo.AnchorID{},
						StateIndex: 0,
						UnlockConditions: axongo.AnchorOutputUnlockConditions{
							&axongo.StateControllerAddressUnlockCondition{Address: addr3},
							&axongo.GovernorAddressUnlockCondition{Address: addr4},
						},
						Features: axongo.AnchorOutputFeatures{
							&axongo.StateMetadataFeature{Entries: axongo.StateMetadataFeatureEntries{"data": []byte("a new anchor output")}},
						},
					},

					// foundry output - serialNumber: 1, minted: 200, melted: 0, max: 1000 [defaultAmount] (owned by account1AccountAddress)
					//				  - native token balance 100 (freshly minted)
					// => input 11
					&axongo.FoundryOutput{
						Amount:       defaultAmount,
						SerialNumber: 1,
						TokenScheme: &axongo.SimpleTokenScheme{
							MintedTokens:  new(big.Int).SetInt64(200),
							MeltedTokens:  big.NewInt(0),
							MaximumSupply: new(big.Int).SetInt64(1000),
						},
						UnlockConditions: axongo.FoundryOutputUnlockConditions{
							&axongo.ImmutableAccountUnlockCondition{Address: account1AccountAddress},
						},
						Features: axongo.FoundryOutputFeatures{
							&axongo.NativeTokenFeature{
								ID:     foundry1addr3NativeTokenID,
								Amount: new(big.Int).SetUint64(100), // freshly minted
							},
						},
					},

					// foundry output - serialNumber: 2, minted: 100, melted: 50, max: 1000 [defaultAmount] (owned by account1AccountAddress)
					//				  - native token balance 50 (melted 50)
					// => input 12
					&axongo.FoundryOutput{
						Amount:       defaultAmount,
						SerialNumber: 2,
						TokenScheme: &axongo.SimpleTokenScheme{
							MintedTokens:  new(big.Int).SetInt64(100),
							MeltedTokens:  big.NewInt(50),
							MaximumSupply: new(big.Int).SetInt64(1000),
						},
						UnlockConditions: axongo.FoundryOutputUnlockConditions{
							&axongo.ImmutableAccountUnlockCondition{Address: account1AccountAddress},
						},
						Features: axongo.FoundryOutputFeatures{
							&axongo.NativeTokenFeature{
								ID:     foundry2addr3NativeTokenID,
								Amount: new(big.Int).SetUint64(50), // melted to 50
							},
						},
					},

					// foundry output - serialNumber: 3, minted: 100, melted: 0, max: 1000 [defaultAmount] (owned by account1AccountAddress)
					// => input 13
					&axongo.FoundryOutput{
						Amount:       defaultAmount,
						SerialNumber: 3,
						TokenScheme: &axongo.SimpleTokenScheme{
							MintedTokens:  new(big.Int).SetInt64(100),
							MeltedTokens:  big.NewInt(0),
							MaximumSupply: new(big.Int).SetInt64(1000),
						},
						UnlockConditions: axongo.FoundryOutputUnlockConditions{
							&axongo.ImmutableAccountUnlockCondition{Address: account1AccountAddress},
						},
						Features: axongo.FoundryOutputFeatures{
							&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("interesting metadata")}},
						},
					},

					// foundry output - serialNumber: 6, minted: 100, melted: 0, max: 1000 [defaultAmount] (owned by account1AccountAddress)
					//				  - native token balance 100
					// => input 5
					newFoundryWithInitialSupply,

					// basic output [defaultAmount] (owned by addr3)
					// => input 14 (foundry 4 destruction remainder)
					&axongo.BasicOutput{
						Amount: defaultAmount,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr3},
						},
					},

					// NFT output transitioned and changed ownership [defaultAmount] (owned by addr4)
					// => input 15
					&axongo.NFTOutput{
						Amount: defaultAmount,
						NFTID:  nft1ID,
						UnlockConditions: axongo.NFTOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr4},
						},
						Features: axongo.NFTOutputFeatures{},
						ImmutableFeatures: axongo.NFTOutputImmFeatures{
							&axongo.IssuerFeature{Address: addr3},
							&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("transfer to 4")}},
						},
					},

					// new NFT output [defaultAmount] (owned by addr4)
					// => input 16
					&axongo.NFTOutput{
						Amount: defaultAmount,
						NFTID:  axongo.NFTID{},
						UnlockConditions: axongo.NFTOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr4},
						},
						Features: nil,
						ImmutableFeatures: axongo.NFTOutputImmFeatures{
							&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("immutable metadata")}},
						},
					},

					// basic output [defaultAmount] (owned by addr5)
					// => input 17
					&axongo.BasicOutput{
						Amount: defaultAmount,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr5},
						},
					},
				},
			}

			sigs, err := transaction.Sign(addr1AddrKeys, addr2AddrKeys, addr3AddrKeys, addr4AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet:        inputs,
					CommitmentInput: &axongo.Commitment{Slot: creationSlot},
				},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						// basic
						&axongo.SignatureUnlock{Signature: sigs[0]}, // basic output (owned by addr1) => (addr1 == Reference 0)
						&axongo.SignatureUnlock{Signature: sigs[1]}, // basic output (owned by addr2) => (addr2 == Reference 1)
						&axongo.ReferenceUnlock{Reference: 1},       // basic output (owned by addr2)
						&axongo.ReferenceUnlock{Reference: 0},       // basic output (owned by addr1)
						&axongo.ReferenceUnlock{Reference: 1},       // basic output (owned by addr2)
						&axongo.ReferenceUnlock{Reference: 1},       // basic output (owned by addr2)
						// account
						&axongo.SignatureUnlock{Signature: sigs[2]}, // account output (owned by addr3) => (addr3 == Reference 6)
						&axongo.ReferenceUnlock{Reference: 6},       // account output (owned by addr3)
						// anchor
						&axongo.SignatureUnlock{Signature: sigs[3]}, // anchor output (owned by state: addr3, gov: addr4) => governance transitioned => (addr4 == Reference 8)
						&axongo.ReferenceUnlock{Reference: 6},       // anchor output (owned by state: addr3, gov: addr4) => state transitioned
						&axongo.ReferenceUnlock{Reference: 6},       // anchor output (owned by state: addr3, gov: addr3) => governance transitioned
						// foundries
						&axongo.AccountUnlock{Reference: 6}, // foundry output (owned by account1AccountAddress)
						&axongo.AccountUnlock{Reference: 6}, // foundry output (owned by account1AccountAddress)
						&axongo.AccountUnlock{Reference: 6}, // foundry output (owned by account1AccountAddress)
						&axongo.AccountUnlock{Reference: 6}, // foundry output (owned by account1AccountAddress)
						// nfts
						&axongo.ReferenceUnlock{Reference: 6}, // NFT output (owned by addr3)
						&axongo.ReferenceUnlock{Reference: 8}, // NFT output (owned by addr4)
						&axongo.NFTUnlock{Reference: 15},      // basic output (owned by nft1ID)
					},
				},
				wantErr: nil,
			}
		}(),

		// fail - changed immutable account address unlock
		func() *test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, addr1, addr1AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(2)
			inFoundry := &axongo.FoundryOutput{
				Amount:       100,
				SerialNumber: 5,
				TokenScheme: &axongo.SimpleTokenScheme{
					MintedTokens:  new(big.Int).SetInt64(1000),
					MeltedTokens:  big.NewInt(0),
					MaximumSupply: new(big.Int).SetInt64(10000),
				},
				UnlockConditions: axongo.FoundryOutputUnlockConditions{
					&axongo.ImmutableAccountUnlockCondition{Address: accountAddr1},
				},
			}
			outFoundry := inFoundry.Clone().(*axongo.FoundryOutput)
			// change the immutable account address unlock
			outFoundry.UnlockConditions = axongo.FoundryOutputUnlockConditions{
				&axongo.ImmutableAccountUnlockCondition{Address: tpkg.RandAccountAddress()},
			}

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.AccountOutput{
					Amount:    100,
					AccountID: accountAddr1.AccountID(),
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
				inputIDs[1]: inFoundry,
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs:       inputIDs.UTXOInputs(),
					Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.AccountOutput{
						Amount:    100,
						AccountID: accountAddr1.AccountID(),
						UnlockConditions: axongo.AccountOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr1},
						},
					},
					outFoundry,
				},
			}

			sigs, err := transaction.Sign(addr1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - changed immutable account address unlock",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
						// should be an AccountUnlock
						&axongo.AccountUnlock{Reference: 0},
					},
				},
				// Changing the immutable account address unlock changes foundryID, therefore the chain is broken.
				// Next state of the foundry is empty, meaning it is interpreted as a destroy operation, and native tokens
				// are not balanced.
				wantErr: axongo.ErrNativeTokenSumUnbalanced,
			}
		}(),

		// ok - modify block issuer account
		func() *test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, addr1, addr1AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.AccountOutput{
					Amount:    100,
					AccountID: accountAddr1.AccountID(),
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							BlockIssuerKeys: axongo.NewBlockIssuerKeys(),
							ExpirySlot:      100,
						},
					},
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					CreationSlot: 110,
					ContextInputs: axongo.TxEssenceContextInputs{
						&axongo.BlockIssuanceCreditInput{
							AccountID: accountAddr1.AccountID(),
						},
					},
					Inputs:       inputIDs.UTXOInputs(),
					Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.AccountOutput{
						Amount:    100,
						AccountID: accountAddr1.AccountID(),
						Features: axongo.AccountOutputFeatures{
							&axongo.BlockIssuerFeature{
								BlockIssuerKeys: axongo.NewBlockIssuerKeys(),
								ExpirySlot:      1000,
							},
						},
						UnlockConditions: axongo.AccountOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr1},
						},
					},
				},
			}

			bicInputs := vm.BlockIssuanceCreditInputSet{
				accountAddr1.AccountID(): 0,
			}

			commitmentInput := &axongo.Commitment{
				Slot: 110,
			}

			sigs, err := transaction.Sign(addr1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - modify block issuer account",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs, BlockIssuanceCreditInputSet: bicInputs, CommitmentInput: commitmentInput},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - set block issuer expiry to max value
		func() *test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, addr1, addr1AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.AccountOutput{
					Amount:    100,
					AccountID: accountAddr1.AccountID(),
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							BlockIssuerKeys: axongo.NewBlockIssuerKeys(),
							ExpirySlot:      100,
						},
					},
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					CreationSlot: 110,
					ContextInputs: axongo.TxEssenceContextInputs{
						&axongo.BlockIssuanceCreditInput{
							AccountID: accountAddr1.AccountID(),
						},
					},
					Inputs:       inputIDs.UTXOInputs(),
					Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.AccountOutput{
						Amount:    100,
						AccountID: accountAddr1.AccountID(),
						Features: axongo.AccountOutputFeatures{
							&axongo.BlockIssuerFeature{
								BlockIssuerKeys: axongo.NewBlockIssuerKeys(),
								ExpirySlot:      axongo.MaxSlotIndex,
							},
						},
						UnlockConditions: axongo.AccountOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr1},
						},
					},
				},
			}

			bicInputs := vm.BlockIssuanceCreditInputSet{
				accountAddr1.AccountID(): 0,
			}

			commitmentInput := &axongo.Commitment{
				Slot: 110,
			}

			sigs, err := transaction.Sign(addr1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - set block issuer expiry to max value",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs, BlockIssuanceCreditInputSet: bicInputs, CommitmentInput: commitmentInput},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - remove expired block issuer feature from new account
		func() *test {
			_, addr1, addr1AddressKeys := tpkg.RandEd25519Identity()

			creationSlot := axongo.SlotIndex(110)
			inputIDs := tpkg.RandOutputIDs(1)
			accountID := axongo.AccountIDFromOutputID(inputIDs[0])

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.AccountOutput{
					Amount:    100,
					AccountID: axongo.EmptyAccountID,
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							BlockIssuerKeys: axongo.NewBlockIssuerKeys(),
							ExpirySlot:      100,
						},
					},
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					CreationSlot: creationSlot,
					ContextInputs: axongo.TxEssenceContextInputs{
						&axongo.BlockIssuanceCreditInput{
							AccountID: accountID,
						},
					},
					Inputs:       inputIDs.UTXOInputs(),
					Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.AccountOutput{
						Amount:    100,
						AccountID: accountID,
						Features:  axongo.AccountOutputFeatures{},
						UnlockConditions: axongo.AccountOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr1},
						},
					},
				},
			}

			bicInputs := vm.BlockIssuanceCreditInputSet{
				accountID: 0,
			}

			commitmentInput := &axongo.Commitment{
				Slot: creationSlot,
			}

			sigs, err := transaction.Sign(addr1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - remove expired block issuer feature from new account",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet:                    inputs,
					BlockIssuanceCreditInputSet: bicInputs,
					CommitmentInput:             commitmentInput,
				},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		// fail - destroy block issuer account with expiry at slot with max value
		func() *test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, addr1, addr1AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.AccountOutput{
					Amount:    100,
					AccountID: accountAddr1.AccountID(),
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							BlockIssuerKeys: axongo.NewBlockIssuerKeys(),
							ExpirySlot:      axongo.MaxSlotIndex,
						},
					},
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					CreationSlot: 110,
					ContextInputs: axongo.TxEssenceContextInputs{
						&axongo.BlockIssuanceCreditInput{
							AccountID: accountAddr1.AccountID(),
						},
					},
					Inputs:       inputIDs.UTXOInputs(),
					Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.BasicOutput{
						Amount: 100,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr1},
						},
					},
				},
			}

			bicInputs := vm.BlockIssuanceCreditInputSet{
				accountAddr1.AccountID(): 0,
			}

			commitmentInput := &axongo.Commitment{
				Slot: 110,
			}

			sigs, err := transaction.Sign(addr1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - destroy block issuer account with expiry at slot with max value",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs, BlockIssuanceCreditInputSet: bicInputs, CommitmentInput: commitmentInput},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},

				wantErr: axongo.ErrBlockIssuerNotExpired,
			}
		}(),

		// ok - destroy block issuer account
		func() *test {
			_, addr1, addr1AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(1)
			// Simulate the scenario where the input account's ID is unset.
			accountID := axongo.AccountIDFromOutputID(inputIDs[0])

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.AccountOutput{
					Amount:    100,
					AccountID: axongo.EmptyAccountID,
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							BlockIssuerKeys: axongo.NewBlockIssuerKeys(),
							ExpirySlot:      100,
						},
					},
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					CreationSlot: 110,
					ContextInputs: axongo.TxEssenceContextInputs{
						&axongo.BlockIssuanceCreditInput{
							AccountID: accountID,
						},
					},
					Inputs:       inputIDs.UTXOInputs(),
					Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.BasicOutput{
						Amount: 100,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr1},
						},
					},
				},
			}

			bicInputs := vm.BlockIssuanceCreditInputSet{
				accountID: 0,
			}

			commitment := &axongo.Commitment{
				Slot: 110,
			}

			sigs, err := transaction.Sign(addr1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - destroy block issuer account",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs, BlockIssuanceCreditInputSet: bicInputs, CommitmentInput: commitment},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		// fail - destroy block issuer account without supplying BIC
		func() *test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, addr1, addr1AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.AccountOutput{
					Amount:    100,
					AccountID: accountAddr1.AccountID(),
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							BlockIssuerKeys: axongo.NewBlockIssuerKeys(),
							ExpirySlot:      100,
						},
					},
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					CreationSlot: 110,
					Inputs:       inputIDs.UTXOInputs(),
					Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.BasicOutput{
						Amount: 100,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr1},
						},
					},
				},
			}

			commitment := &axongo.Commitment{
				Slot: 110,
			}

			sigs, err := transaction.Sign(addr1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - destroy block issuer account without supplying BIC",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs, CommitmentInput: commitment},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: axongo.ErrBlockIssuanceCreditInputMissing,
			}
		}(),

		// fail - modify block issuer without supplying BIC
		func() *test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, addr1, addr1AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.AccountOutput{
					Amount:    100,
					AccountID: accountAddr1.AccountID(),
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							BlockIssuerKeys: axongo.NewBlockIssuerKeys(),
							ExpirySlot:      100,
						},
					},
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					CreationSlot: 110,
					Inputs:       inputIDs.UTXOInputs(),
					Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.AccountOutput{
						Amount:    100,
						AccountID: accountAddr1.AccountID(),
						Features: axongo.AccountOutputFeatures{
							&axongo.BlockIssuerFeature{
								BlockIssuerKeys: axongo.NewBlockIssuerKeys(),
								ExpirySlot:      1000,
							},
						},
						UnlockConditions: axongo.AccountOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr1},
						},
					},
				},
			}

			sigs, err := transaction.Sign(addr1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - modify block issuer without supplying BIC",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: axongo.ErrBlockIssuanceCreditInputMissing,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAndExecuteSignedTransaction(tt.tx, tt.resolvedInputs)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

type txBuilder struct {
	// the amount of randomly created ed25519 addresses with private keys
	ed25519AddrCnt int
	// used to created own addresses for the test
	addressesFunc func(ed25519Addresses []axongo.Address) []axongo.Address
	// used to create inputs for the test
	inputsFunc func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output
	// used to create outputs for the test (optional)
	outputsFunc func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address, totalInputAmount axongo.BaseToken, totalInputMana axongo.Mana) axongo.TxEssenceOutputs
	// used to create unlocks for the test
	unlocksFunc func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks
}

type txExecTest struct {
	// the name of the testcase
	name string
	// the txBuilder that builds the transaction for the testcase
	txBuilder *txBuilder
	// hook that gets executed before the transaction is signed (optional)
	txPreSignHook func(t *axongo.Transaction)
	// expected error during execution of the transaction
	wantErr error
}

func runNovaTransactionExecutionTest(t *testing.T, test *txExecTest) {
	t.Helper()

	t.Run(test.name, func(t *testing.T) {
		// generate random ed25519 addresses
		ed25519Addresses, ed25519AddressesWithKeys := tpkg.RandEd25519IdentitiesSortedByAddress(test.txBuilder.ed25519AddrCnt)

		// pass the ed25519 testAddresses and get the complete list of testAddresses
		testAddresses := make([]axongo.Address, 0)
		if test.txBuilder.addressesFunc != nil {
			testAddresses = test.txBuilder.addressesFunc(ed25519Addresses)
		}

		inputs := test.txBuilder.inputsFunc(ed25519Addresses, testAddresses)
		if len(inputs) == 0 {
			require.FailNow(t, "no outputs given")
		}

		// create the input set
		inputIDs := tpkg.RandOutputIDsWithCreationSlot(0, uint16(len(inputs)))
		inputSet := vm.InputSet{}
		var totalInputAmount axongo.BaseToken
		for idx, output := range inputs {
			inputSet[inputIDs[idx]] = output
			totalInputAmount += output.BaseTokenAmount()
		}

		// calculate the mana on input side
		// HINT: all outputs are created at slot 0 and the transaction is executed at slot 10000
		var txCreationSlot axongo.SlotIndex = 10000

		totalInputMana, err := vm.TotalManaIn(testAPI.ManaDecayProvider(), testAPI.StorageScoreStructure(), txCreationSlot, inputSet, vm.RewardsInputSet{})
		require.NoError(t, err)

		outputs := axongo.TxEssenceOutputs{
			// collect everything on a basic output with a random ed25519 address
			&axongo.BasicOutput{
				Amount: totalInputAmount,
				UnlockConditions: axongo.BasicOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
		}
		if test.txBuilder.outputsFunc != nil {
			outputs = test.txBuilder.outputsFunc(ed25519Addresses, testAddresses, totalInputAmount, totalInputMana)
		}

		// create the transaction
		tx := &axongo.Transaction{
			API: testAPI,
			TransactionEssence: &axongo.TransactionEssence{
				NetworkID:     testProtoParams.NetworkID(),
				CreationSlot:  txCreationSlot,
				ContextInputs: axongo.TxEssenceContextInputs{},
				Inputs:        inputIDs.UTXOInputs(),
				Allotments:    axongo.Allotments{},
				Capabilities:  axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
			},
			Outputs: outputs,
		}

		// execute the pre sign hook
		if test.txPreSignHook != nil {
			test.txPreSignHook(tx)
		}

		// sign the transaction essence
		sigs, err := tx.Sign(ed25519AddressesWithKeys...)
		require.NoError(t, err)

		// pass the signatures and get the unlock conditions
		unlocks := test.txBuilder.unlocksFunc(sigs, testAddresses)

		signedTx := &axongo.SignedTransaction{
			API:         testAPI,
			Transaction: tx,
			Unlocks:     unlocks,
		}

		txBytes, err := testAPI.Encode(signedTx, serix.WithValidation())
		require.NoError(t, err)

		// we deserialize to be sure that all serix rules are applied (like lexically ordering or multi addresses)
		signedTx = &axongo.SignedTransaction{}
		_, err = testAPI.Decode(txBytes, signedTx, serix.WithValidation())
		require.NoError(t, err)

		// execute the transaction
		err = validateAndExecuteSignedTransaction(signedTx, vm.ResolvedInputs{InputSet: inputSet})
		if test.wantErr != nil {
			require.ErrorIs(t, err, test.wantErr)
			return
		}
		require.NoError(t, err)
	})
}

func TestNovaTransactionExecution_RestrictedAddress(t *testing.T) {

	defaultAmount := OneIOTA

	tests := []*txExecTest{
		// ok - restricted ed25519 address unlock
		func() *txExecTest {
			return &txExecTest{
				name: "ok - restricted ed25519 address unlock",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 1,
					addressesFunc: func(ed25519Addresses []axongo.Address) []axongo.Address {
						return []axongo.Address{
							&axongo.RestrictedAddress{
								Address:             ed25519Addresses[0],
								AllowedCapabilities: axongo.AddressCapabilitiesBitMask{},
							},
						}
					},
					inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
						return []axongo.Output{
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
						}
					},
					outputsFunc: nil,
					unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
						return axongo.Unlocks{
							&axongo.SignatureUnlock{Signature: sigs[0]},
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - restricted account address unlock
		func() *txExecTest {
			return &txExecTest{
				name: "ok - restricted account address unlock",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 1,
					addressesFunc: func(ed25519Addresses []axongo.Address) []axongo.Address {
						accountAddress := tpkg.RandAccountAddress()
						return []axongo.Address{
							accountAddress,
							&axongo.RestrictedAddress{
								Address:             accountAddress,
								AllowedCapabilities: axongo.AddressCapabilitiesBitMask{},
							},
						}
					},
					inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
						return []axongo.Output{
							// we add an output with a Ed25519 address to be able to check the AccountUnlock in the RestrictedAddress
							&axongo.AccountOutput{
								Amount:         defaultAmount,
								AccountID:      testAddresses[0].(*axongo.AccountAddress).AccountID(),
								FoundryCounter: 0,
								UnlockConditions: axongo.AccountOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
								Features: nil,
							},
							// owned by restricted account address
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: testAddresses[1]},
								},
							},
						}
					},
					outputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address, totalInputAmount axongo.BaseToken, totalInputMana axongo.Mana) axongo.TxEssenceOutputs {
						return axongo.TxEssenceOutputs{
							&axongo.AccountOutput{
								Amount:         defaultAmount,
								AccountID:      testAddresses[0].(*axongo.AccountAddress).AccountID(),
								FoundryCounter: 0,
								UnlockConditions: axongo.AccountOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
								Features: nil,
							},
							&axongo.BasicOutput{
								Amount: totalInputAmount - defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
							},
						}
					},
					unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
						return axongo.Unlocks{
							&axongo.SignatureUnlock{Signature: sigs[0]}, // account unlock
							&axongo.AccountUnlock{Reference: 0},
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - restricted anchor address unlock
		func() *txExecTest {
			return &txExecTest{
				name: "ok - restricted anchor address unlock",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 2,
					addressesFunc: func(ed25519Addresses []axongo.Address) []axongo.Address {
						anchorAddress := tpkg.RandAnchorAddress()
						return []axongo.Address{
							anchorAddress,
							&axongo.RestrictedAddress{
								Address:             anchorAddress,
								AllowedCapabilities: axongo.AddressCapabilitiesBitMask{},
							},
						}
					},
					inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
						return []axongo.Output{
							// we add an output with a Ed25519 address to be able to check the AnchorUnlock in the RestrictedAddress
							&axongo.AnchorOutput{
								Amount:     defaultAmount,
								AnchorID:   testAddresses[0].(*axongo.AnchorAddress).AnchorID(),
								StateIndex: 1,
								UnlockConditions: axongo.AnchorOutputUnlockConditions{
									&axongo.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
									&axongo.GovernorAddressUnlockCondition{Address: ed25519Addresses[1]},
								},
								Features: axongo.AnchorOutputFeatures{
									&axongo.StateMetadataFeature{Entries: axongo.StateMetadataFeatureEntries{"data": []byte("current state")}},
								},
							},
							// owned by restricted anchor address
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: testAddresses[1]},
								},
							},
						}
					},
					outputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address, totalInputAmount axongo.BaseToken, totalInputMana axongo.Mana) axongo.TxEssenceOutputs {
						return axongo.TxEssenceOutputs{
							// the anchor unlock needs to be a state transition (governor doesn't work for anchor reference unlocks)
							&axongo.AnchorOutput{
								Amount:     defaultAmount,
								AnchorID:   testAddresses[0].(*axongo.AnchorAddress).AnchorID(),
								StateIndex: 2,
								UnlockConditions: axongo.AnchorOutputUnlockConditions{
									&axongo.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
									&axongo.GovernorAddressUnlockCondition{Address: ed25519Addresses[1]},
								},
								Features: axongo.AnchorOutputFeatures{
									&axongo.StateMetadataFeature{Entries: axongo.StateMetadataFeatureEntries{"data": []byte("next state")}},
								},
							},
							&axongo.BasicOutput{
								Amount: totalInputAmount - defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
							},
						}
					},
					unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
						return axongo.Unlocks{
							&axongo.SignatureUnlock{Signature: sigs[0]}, // anchor state controller unlock
							&axongo.AnchorUnlock{Reference: 0},
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - restricted NFT unlock
		func() *txExecTest {
			return &txExecTest{
				name: "ok - restricted NFT unlock",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 2,
					addressesFunc: func(ed25519Addresses []axongo.Address) []axongo.Address {
						nftAddress := tpkg.RandNFTAddress()
						return []axongo.Address{
							nftAddress,
							&axongo.RestrictedAddress{
								Address:             nftAddress,
								AllowedCapabilities: axongo.AddressCapabilitiesBitMask{},
							},
						}
					},
					inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
						return []axongo.Output{
							// we add an output with a Ed25519 address to be able to check the NFT Unlock in the RestrictedAddress
							&axongo.NFTOutput{
								Amount: defaultAmount,
								NFTID:  testAddresses[0].(*axongo.NFTAddress).NFTID(),
								UnlockConditions: axongo.NFTOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
								Features: axongo.NFTOutputFeatures{},
								ImmutableFeatures: axongo.NFTOutputImmFeatures{
									&axongo.IssuerFeature{Address: ed25519Addresses[1]},
									&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("immutable")}},
								},
							},
							// owned by restricted NFT address
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: testAddresses[1]},
								},
							},
						}
					},
					outputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address, totalInputAmount axongo.BaseToken, totalInputMana axongo.Mana) axongo.TxEssenceOutputs {
						return axongo.TxEssenceOutputs{
							&axongo.NFTOutput{
								Amount: defaultAmount,
								NFTID:  testAddresses[0].(*axongo.NFTAddress).NFTID(),
								UnlockConditions: axongo.NFTOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
								Features: axongo.NFTOutputFeatures{
									&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("some new metadata")}},
								},
								ImmutableFeatures: axongo.NFTOutputImmFeatures{
									&axongo.IssuerFeature{Address: ed25519Addresses[1]},
									&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("immutable")}},
								},
							},
							&axongo.BasicOutput{
								Amount: totalInputAmount - defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
							},
						}
					},
					unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
						return axongo.Unlocks{
							&axongo.SignatureUnlock{Signature: sigs[0]}, // NFT unlock
							&axongo.NFTUnlock{Reference: 0},
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - restricted multi address unlock
		func() *txExecTest {
			return &txExecTest{
				name: "ok - restricted multi address unlock",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 1,
					addressesFunc: func(ed25519Addresses []axongo.Address) []axongo.Address {
						nftAddress := tpkg.RandNFTAddress()
						return []axongo.Address{
							nftAddress,
							&axongo.RestrictedAddress{
								Address: &axongo.MultiAddress{
									Addresses: []*axongo.AddressWithWeight{
										{
											Address: ed25519Addresses[0],
											Weight:  1,
										},
										{
											Address: nftAddress,
											Weight:  1,
										},
									},
									Threshold: 2,
								},
								AllowedCapabilities: axongo.AddressCapabilitiesBitMask{},
							},
						}
					},
					inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
						return []axongo.Output{
							// we add an output with a Ed25519 address to be able to check the NFT Unlock in the RestrictedAddress multi address
							&axongo.NFTOutput{
								Amount: defaultAmount,
								NFTID:  testAddresses[0].(*axongo.NFTAddress).NFTID(),
								UnlockConditions: axongo.NFTOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
								Features: axongo.NFTOutputFeatures{},
								ImmutableFeatures: axongo.NFTOutputImmFeatures{
									&axongo.IssuerFeature{Address: ed25519Addresses[0]},
									&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("immutable")}},
								},
							},
							// owned by restricted multi address
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: testAddresses[1]},
								},
							},
						}
					},
					outputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address, totalInputAmount axongo.BaseToken, totalInputMana axongo.Mana) axongo.TxEssenceOutputs {
						return axongo.TxEssenceOutputs{
							&axongo.NFTOutput{
								Amount: defaultAmount,
								NFTID:  testAddresses[0].(*axongo.NFTAddress).NFTID(),
								UnlockConditions: axongo.NFTOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
								Features: axongo.NFTOutputFeatures{
									&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("some new metadata")}},
								},
								ImmutableFeatures: axongo.NFTOutputImmFeatures{
									&axongo.IssuerFeature{Address: ed25519Addresses[0]},
									&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("immutable")}},
								},
							},
							&axongo.BasicOutput{
								Amount: totalInputAmount - defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
							},
						}
					},
					unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
						return axongo.Unlocks{
							&axongo.SignatureUnlock{Signature: sigs[0]}, // NFT unlock
							&axongo.MultiUnlock{
								Unlocks: []axongo.Unlock{
									&axongo.ReferenceUnlock{Reference: 0},
									&axongo.NFTUnlock{Reference: 0},
								},
							},
						}
					},
				},
				wantErr: nil,
			}
		}(),
	}
	for _, tt := range tests {
		runNovaTransactionExecutionTest(t, tt)
	}
}

func TestNovaTransactionExecution_MultiAddress(t *testing.T) {

	defaultAmount := OneIOTA

	tests := []*txExecTest{
		// ok - threshold == cumulativeWeight (threshold reached)
		func() *txExecTest {
			return &txExecTest{
				name: "ok - threshold == cumulativeWeight (threshold reached)",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 2,
					addressesFunc: func(ed25519Addresses []axongo.Address) []axongo.Address {
						return []axongo.Address{
							// only 2 mandatory addresses
							&axongo.MultiAddress{
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
						}
					},
					inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
						return []axongo.Output{
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
						}
					},
					outputsFunc: nil,
					unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
						return axongo.Unlocks{
							&axongo.MultiUnlock{
								Unlocks: []axongo.Unlock{
									&axongo.SignatureUnlock{Signature: sigs[0]},
									&axongo.SignatureUnlock{Signature: sigs[1]},
								},
							},
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - threshold < cumulativeWeight (threshold reached)
		func() *txExecTest {
			return &txExecTest{
				name: "ok - threshold < cumulativeWeight (threshold reached)",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 2,
					addressesFunc: func(ed25519Addresses []axongo.Address) []axongo.Address {
						return []axongo.Address{
							// only 2 mandatory addresses
							&axongo.MultiAddress{
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
								Threshold: 1,
							},
						}
					},
					inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
						return []axongo.Output{
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
						}
					},
					outputsFunc: nil,
					unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
						return axongo.Unlocks{
							&axongo.MultiUnlock{
								Unlocks: []axongo.Unlock{
									&axongo.SignatureUnlock{Signature: sigs[0]},
									&axongo.SignatureUnlock{Signature: sigs[1]},
								},
							},
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// fail - threshold == cumulativeWeight (threshold not reached)
		func() *txExecTest {
			return &txExecTest{
				name: "fail - threshold == cumulativeWeight (threshold not reached)",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 2,
					addressesFunc: func(ed25519Addresses []axongo.Address) []axongo.Address {
						return []axongo.Address{
							// only 2 mandatory addresses
							&axongo.MultiAddress{
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
						}
					},
					inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
						return []axongo.Output{
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
						}
					},
					outputsFunc: nil,
					unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
						return axongo.Unlocks{
							&axongo.MultiUnlock{
								Unlocks: []axongo.Unlock{
									// we only unlock one of the addresses
									&axongo.SignatureUnlock{Signature: sigs[0]},
									&axongo.EmptyUnlock{},
								},
							},
						}
					},
				},
				wantErr: axongo.ErrMultiAddressUnlockThresholdNotReached,
			}
		}(),

		// fail - threshold < cumulativeWeight (threshold not reached)
		func() *txExecTest {
			return &txExecTest{
				name: "fail - threshold < cumulativeWeight (threshold not reached)",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 2,
					addressesFunc: func(ed25519Addresses []axongo.Address) []axongo.Address {
						return []axongo.Address{
							// only 2 mandatory addresses
							&axongo.MultiAddress{
								Addresses: []*axongo.AddressWithWeight{
									{
										Address: ed25519Addresses[0],
										Weight:  2,
									},
									{
										Address: ed25519Addresses[1],
										Weight:  2,
									},
								},
								Threshold: 3,
							},
						}
					},
					inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
						return []axongo.Output{
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
						}
					},
					outputsFunc: nil,
					unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
						return axongo.Unlocks{
							&axongo.MultiUnlock{
								Unlocks: []axongo.Unlock{
									// we only unlock one of the addresses
									&axongo.SignatureUnlock{Signature: sigs[0]},
									&axongo.EmptyUnlock{},
								},
							},
						}
					},
				},
				wantErr: axongo.ErrMultiAddressUnlockThresholdNotReached,
			}
		}(),

		// fail - len(multiAddr) != len(multiUnlock)
		func() *txExecTest {
			return &txExecTest{
				name: "fail - len(multiAddr) != len(multiUnlock)",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 3,
					addressesFunc: func(ed25519Addresses []axongo.Address) []axongo.Address {
						return []axongo.Address{
							&axongo.MultiAddress{
								Addresses: []*axongo.AddressWithWeight{
									{
										Address: ed25519Addresses[0],
										Weight:  1,
									},
									{
										Address: ed25519Addresses[1],
										Weight:  1,
									},
									{
										Address: ed25519Addresses[2],
										Weight:  1,
									},
								},
								Threshold: 2,
							},
						}
					},
					inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
						return []axongo.Output{
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
						}
					},
					outputsFunc: nil,
					unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
						return axongo.Unlocks{
							&axongo.MultiUnlock{
								Unlocks: []axongo.Unlock{
									&axongo.SignatureUnlock{Signature: sigs[0]},
									&axongo.SignatureUnlock{Signature: sigs[1]},
									// Empty unlock missing here
								},
							},
						}
					},
				},
				wantErr: axongo.ErrMultiAddressLengthUnlockLengthMismatch,
			}
		}(),

		// ok - Reference unlock
		func() *txExecTest {
			return &txExecTest{
				name: "ok - Reference unlock",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 2,
					addressesFunc: func(ed25519Addresses []axongo.Address) []axongo.Address {
						return []axongo.Address{
							// only 2 mandatory addresses
							&axongo.MultiAddress{
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
						}
					},
					inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
						return []axongo.Output{
							// we add a basic output with a Ed25519 address to be able to check the RefUnlock in the MultiAddress
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: ed25519Addresses[1]},
								},
							},
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
						}
					},
					outputsFunc: nil,
					unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
						return axongo.Unlocks{
							&axongo.SignatureUnlock{Signature: sigs[1]},
							&axongo.MultiUnlock{
								Unlocks: []axongo.Unlock{
									&axongo.SignatureUnlock{Signature: sigs[0]},
									&axongo.ReferenceUnlock{Reference: 0},
								},
							},
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - MultiAddress Reference unlock
		func() *txExecTest {
			return &txExecTest{
				name: "ok - MultiAddress Reference unlock",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 2,
					addressesFunc: func(ed25519Addresses []axongo.Address) []axongo.Address {
						return []axongo.Address{
							&axongo.MultiAddress{
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
						}
					},
					inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
						return []axongo.Output{
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
						}
					},
					outputsFunc: nil,
					unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
						return axongo.Unlocks{
							&axongo.MultiUnlock{
								Unlocks: []axongo.Unlock{
									&axongo.SignatureUnlock{Signature: sigs[0]},
									&axongo.SignatureUnlock{Signature: sigs[1]},
								},
							},
							&axongo.ReferenceUnlock{Reference: 0},
							&axongo.ReferenceUnlock{Reference: 0},
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - Account unlock
		func() *txExecTest {
			return &txExecTest{
				name: "ok - Account unlock",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 2,
					addressesFunc: func(ed25519Addresses []axongo.Address) []axongo.Address {
						accountAddress := tpkg.RandAccountAddress()
						return []axongo.Address{
							accountAddress,
							// ed25519 address + account address
							&axongo.MultiAddress{
								Addresses: []*axongo.AddressWithWeight{
									{
										Address: ed25519Addresses[1],
										Weight:  1,
									},
									{
										Address: accountAddress,
										Weight:  1,
									},
								},
								Threshold: 2,
							},
						}
					},
					inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
						return []axongo.Output{
							// we add an output with a Ed25519 address to be able to check the AccountUnlock in the MultiAddress
							&axongo.AccountOutput{
								Amount:         defaultAmount,
								AccountID:      testAddresses[0].(*axongo.AccountAddress).AccountID(),
								FoundryCounter: 0,
								UnlockConditions: axongo.AccountOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
								Features: nil,
							},
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: ed25519Addresses[1]},
								},
							},
							// owned by ed25519 address + account address
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: testAddresses[1]},
								},
							},
						}
					},
					outputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address, totalInputAmount axongo.BaseToken, totalInputMana axongo.Mana) axongo.TxEssenceOutputs {
						return axongo.TxEssenceOutputs{
							&axongo.AccountOutput{
								Amount:         defaultAmount,
								AccountID:      testAddresses[0].(*axongo.AccountAddress).AccountID(),
								FoundryCounter: 0,
								UnlockConditions: axongo.AccountOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
								Features: nil,
							},
							&axongo.BasicOutput{
								Amount: totalInputAmount - defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
							},
						}
					},
					unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
						// this is a bit complicated in the test, because the addresses are generated randomly,
						// but the MultiAddresses get sorted lexically, so we have to find out the correct order in the MultiUnlock.

						accountAddress := testAddresses[0]
						multiAddress := testAddresses[1].(*axongo.MultiAddress)

						// sort the addresses in the multi like the serializer will do
						slices.SortFunc(multiAddress.Addresses, func(a *axongo.AddressWithWeight, b *axongo.AddressWithWeight) int {
							return bytes.Compare(a.Address.ID(), b.Address.ID())
						})

						// search the index of the account address in the multi address
						foundAccountAddressIndex := -1
						for idx, address := range multiAddress.Addresses {
							if address.Address.Equal(accountAddress) {
								foundAccountAddressIndex = idx
								break
							}
						}

						var multiUnlock *axongo.MultiUnlock

						switch foundAccountAddressIndex {
						case -1:
							require.FailNow(t, "account address not found in multi address")

						case 0:
							multiUnlock = &axongo.MultiUnlock{
								Unlocks: []axongo.Unlock{
									&axongo.AccountUnlock{Reference: 0},
									&axongo.ReferenceUnlock{Reference: 1},
								},
							}

						case 1:
							multiUnlock = &axongo.MultiUnlock{
								Unlocks: []axongo.Unlock{
									&axongo.ReferenceUnlock{Reference: 1},
									&axongo.AccountUnlock{Reference: 0},
								},
							}

						default:
							require.FailNow(t, "unknown account address index found in multi address")
						}

						return axongo.Unlocks{
							&axongo.SignatureUnlock{Signature: sigs[0]}, // account unlock
							&axongo.SignatureUnlock{Signature: sigs[1]}, // basic output unlock
							multiUnlock,
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - Anchor unlock (state transition)
		func() *txExecTest {
			return &txExecTest{
				name: "ok - Anchor unlock (state transition)",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 2,
					addressesFunc: func(ed25519Addresses []axongo.Address) []axongo.Address {
						anchorAddress := tpkg.RandAnchorAddress()
						return []axongo.Address{
							anchorAddress,
							// ed25519 address + anchor address
							&axongo.MultiAddress{
								Addresses: []*axongo.AddressWithWeight{
									{
										Address: ed25519Addresses[1],
										Weight:  1,
									},
									{
										Address: anchorAddress,
										Weight:  1,
									},
								},
								Threshold: 2,
							},
						}
					},
					inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
						return []axongo.Output{
							// we add an output with a Ed25519 address to be able to check the AnchorUnlock in the MultiAddress
							&axongo.AnchorOutput{
								Amount:     defaultAmount,
								AnchorID:   testAddresses[0].(*axongo.AnchorAddress).AnchorID(),
								StateIndex: 1,
								UnlockConditions: axongo.AnchorOutputUnlockConditions{
									&axongo.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
									&axongo.GovernorAddressUnlockCondition{Address: ed25519Addresses[1]},
								},
								Features: axongo.AnchorOutputFeatures{
									&axongo.StateMetadataFeature{Entries: axongo.StateMetadataFeatureEntries{"data": []byte("current state")}},
								},
							},
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: ed25519Addresses[1]},
								},
							},
							// owned by ed25519 address + anchor address
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: testAddresses[1]},
								},
							},
						}
					},
					outputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address, totalInputAmount axongo.BaseToken, totalInputMana axongo.Mana) axongo.TxEssenceOutputs {
						return axongo.TxEssenceOutputs{
							// the anchor unlock needs to be a state transition (governor doesn't work for anchor reference unlocks)
							&axongo.AnchorOutput{
								Amount:     defaultAmount,
								AnchorID:   testAddresses[0].(*axongo.AnchorAddress).AnchorID(),
								StateIndex: 2,
								UnlockConditions: axongo.AnchorOutputUnlockConditions{
									&axongo.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
									&axongo.GovernorAddressUnlockCondition{Address: ed25519Addresses[1]},
								},
								Features: axongo.AnchorOutputFeatures{
									&axongo.StateMetadataFeature{Entries: axongo.StateMetadataFeatureEntries{"data": []byte("next state")}},
								},
							},
							&axongo.BasicOutput{
								Amount: totalInputAmount - defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
							},
						}
					},
					unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
						// this is a bit complicated in the test, because the addresses are generated randomly,
						// but the MultiAddresses get sorted lexically, so we have to find out the correct order in the MultiUnlock.

						anchorAddress := testAddresses[0]
						multiAddress := testAddresses[1].(*axongo.MultiAddress)

						// sort the addresses in the multi like the serializer will do
						slices.SortFunc(multiAddress.Addresses, func(a *axongo.AddressWithWeight, b *axongo.AddressWithWeight) int {
							return bytes.Compare(a.Address.ID(), b.Address.ID())
						})

						// search the index of the anchor address in the multi address
						foundAnchorAddressIndex := -1
						for idx, address := range multiAddress.Addresses {
							if address.Address.Equal(anchorAddress) {
								foundAnchorAddressIndex = idx
								break
							}
						}

						var multiUnlock *axongo.MultiUnlock

						switch foundAnchorAddressIndex {
						case -1:
							require.FailNow(t, "anchor address not found in multi address")

						case 0:
							multiUnlock = &axongo.MultiUnlock{
								Unlocks: []axongo.Unlock{
									&axongo.AnchorUnlock{Reference: 0},
									&axongo.ReferenceUnlock{Reference: 1},
								},
							}

						case 1:
							multiUnlock = &axongo.MultiUnlock{
								Unlocks: []axongo.Unlock{
									&axongo.ReferenceUnlock{Reference: 1},
									&axongo.AnchorUnlock{Reference: 0},
								},
							}

						default:
							require.FailNow(t, "unknown anchor address index found in multi address")
						}

						return axongo.Unlocks{
							&axongo.SignatureUnlock{Signature: sigs[0]}, // anchor state controller unlock
							&axongo.SignatureUnlock{Signature: sigs[1]}, // basic output unlock
							multiUnlock,
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// fail - Anchor unlock (governance transition)
		func() *txExecTest {
			return &txExecTest{
				name: "fail - Anchor unlock (governance transition)",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 2,
					addressesFunc: func(ed25519Addresses []axongo.Address) []axongo.Address {
						anchorAddress := tpkg.RandAnchorAddress()
						return []axongo.Address{
							anchorAddress,
							// ed25519 address + anchor address
							&axongo.MultiAddress{
								Addresses: []*axongo.AddressWithWeight{
									{
										Address: ed25519Addresses[0],
										Weight:  1,
									},
									{
										Address: anchorAddress,
										Weight:  1,
									},
								},
								Threshold: 2,
							},
						}
					},
					inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
						return []axongo.Output{
							// we add an output with a Ed25519 address to be able to check the AnchorUnlock in the MultiAddress
							&axongo.AnchorOutput{
								Amount:     defaultAmount,
								AnchorID:   testAddresses[0].(*axongo.AnchorAddress).AnchorID(),
								StateIndex: 1,
								UnlockConditions: axongo.AnchorOutputUnlockConditions{
									&axongo.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
									&axongo.GovernorAddressUnlockCondition{Address: ed25519Addresses[1]},
								},
								Features: axongo.AnchorOutputFeatures{
									&axongo.StateMetadataFeature{Entries: axongo.StateMetadataFeatureEntries{"data": []byte("governance transition")}},
								},
							},
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
							},
							// owned by ed25519 address + anchor address
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: testAddresses[1]},
								},
							},
						}
					},
					outputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address, totalInputAmount axongo.BaseToken, totalInputMana axongo.Mana) axongo.TxEssenceOutputs {
						return axongo.TxEssenceOutputs{
							// the anchor unlock needs to be a state transition (governor doesn't work for anchor reference unlocks)
							&axongo.AnchorOutput{
								Amount:     defaultAmount,
								AnchorID:   testAddresses[0].(*axongo.AnchorAddress).AnchorID(),
								StateIndex: 1,
								UnlockConditions: axongo.AnchorOutputUnlockConditions{
									&axongo.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
									&axongo.GovernorAddressUnlockCondition{Address: ed25519Addresses[1]},
								},
								Features: axongo.AnchorOutputFeatures{
									&axongo.StateMetadataFeature{Entries: axongo.StateMetadataFeatureEntries{"data": []byte("governance transition")}},
								},
							},
							&axongo.BasicOutput{
								Amount: totalInputAmount - defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
							},
						}
					},
					unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
						// this is a bit complicated in the test, because the addresses are generated randomly,
						// but the MultiAddresses get sorted lexically, so we have to find out the correct order in the MultiUnlock.

						anchorAddress := testAddresses[0]
						multiAddress := testAddresses[1].(*axongo.MultiAddress)

						// sort the addresses in the multi like the serializer will do
						slices.SortFunc(multiAddress.Addresses, func(a *axongo.AddressWithWeight, b *axongo.AddressWithWeight) int {
							return bytes.Compare(a.Address.ID(), b.Address.ID())
						})

						// search the index of the anchor address in the multi address
						foundAnchorAddressIndex := -1
						for idx, address := range multiAddress.Addresses {
							if address.Address.Equal(anchorAddress) {
								foundAnchorAddressIndex = idx
								break
							}
						}

						var multiUnlock *axongo.MultiUnlock

						switch foundAnchorAddressIndex {
						case -1:
							require.FailNow(t, "anchor address not found in multi address")

						case 0:
							multiUnlock = &axongo.MultiUnlock{
								Unlocks: []axongo.Unlock{
									&axongo.AnchorUnlock{Reference: 0},
									&axongo.ReferenceUnlock{Reference: 1},
								},
							}

						case 1:
							multiUnlock = &axongo.MultiUnlock{
								Unlocks: []axongo.Unlock{
									&axongo.ReferenceUnlock{Reference: 1},
									&axongo.AnchorUnlock{Reference: 0},
								},
							}

						default:
							require.FailNow(t, "unknown anchor address index found in multi address")
						}

						return axongo.Unlocks{
							&axongo.SignatureUnlock{Signature: sigs[1]}, // anchor governor unlock
							&axongo.SignatureUnlock{Signature: sigs[0]}, // basic output unlock
							multiUnlock,
						}
					},
				},
				wantErr: axongo.ErrMultiAddressUnlockInvalid,
			}
		}(),

		// ok - NFT unlock
		func() *txExecTest {
			return &txExecTest{
				name: "ok - NFT unlock",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 2,
					addressesFunc: func(ed25519Addresses []axongo.Address) []axongo.Address {
						nftAddress := tpkg.RandNFTAddress()
						return []axongo.Address{
							nftAddress,
							// ed25519 address + NFT address
							&axongo.MultiAddress{
								Addresses: []*axongo.AddressWithWeight{
									{
										Address: ed25519Addresses[1],
										Weight:  1,
									},
									{
										Address: nftAddress,
										Weight:  1,
									},
								},
								Threshold: 2,
							},
						}
					},
					inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
						return []axongo.Output{
							// we add an output with a Ed25519 address to be able to check the NFT Unlock in the MultiAddress
							&axongo.NFTOutput{
								Amount: defaultAmount,
								NFTID:  testAddresses[0].(*axongo.NFTAddress).NFTID(),
								UnlockConditions: axongo.NFTOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
								Features: axongo.NFTOutputFeatures{},
								ImmutableFeatures: axongo.NFTOutputImmFeatures{
									&axongo.IssuerFeature{Address: ed25519Addresses[1]},
									&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("immutable")}},
								},
							},
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: ed25519Addresses[1]},
								},
							},
							// owned by ed25519 address + NFT address
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: testAddresses[1]},
								},
							},
						}
					},
					outputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address, totalInputAmount axongo.BaseToken, totalInputMana axongo.Mana) axongo.TxEssenceOutputs {
						return axongo.TxEssenceOutputs{
							&axongo.NFTOutput{
								Amount: defaultAmount,
								NFTID:  testAddresses[0].(*axongo.NFTAddress).NFTID(),
								UnlockConditions: axongo.NFTOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
								Features: axongo.NFTOutputFeatures{
									&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("some new metadata")}},
								},
								ImmutableFeatures: axongo.NFTOutputImmFeatures{
									&axongo.IssuerFeature{Address: ed25519Addresses[1]},
									&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("immutable")}},
								},
							},
							&axongo.BasicOutput{
								Amount: totalInputAmount - defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
							},
						}
					},
					unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
						// this is a bit complicated in the test, because the addresses are generated randomly,
						// but the MultiAddresses get sorted lexically, so we have to find out the correct order in the MultiUnlock.

						nftAddress := testAddresses[0]
						multiAddress := testAddresses[1].(*axongo.MultiAddress)

						// sort the addresses in the multi like the serializer will do
						slices.SortFunc(multiAddress.Addresses, func(a *axongo.AddressWithWeight, b *axongo.AddressWithWeight) int {
							return bytes.Compare(a.Address.ID(), b.Address.ID())
						})

						// search the index of the NFT address in the multi address
						foundNFTAddressIndex := -1
						for idx, address := range multiAddress.Addresses {
							if address.Address.Equal(nftAddress) {
								foundNFTAddressIndex = idx
								break
							}
						}

						var multiUnlock *axongo.MultiUnlock

						switch foundNFTAddressIndex {
						case -1:
							require.FailNow(t, "NFT address not found in multi address")

						case 0:
							multiUnlock = &axongo.MultiUnlock{
								Unlocks: []axongo.Unlock{
									&axongo.NFTUnlock{Reference: 0},
									&axongo.ReferenceUnlock{Reference: 1},
								},
							}

						case 1:
							multiUnlock = &axongo.MultiUnlock{
								Unlocks: []axongo.Unlock{
									&axongo.ReferenceUnlock{Reference: 1},
									&axongo.NFTUnlock{Reference: 0},
								},
							}

						default:
							require.FailNow(t, "unknown NFT address index found in multi address")
						}

						return axongo.Unlocks{
							&axongo.SignatureUnlock{Signature: sigs[0]}, // NFT unlock
							&axongo.SignatureUnlock{Signature: sigs[1]}, // basic output unlock
							multiUnlock,
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - multiple MultiAddresses in one TX - no signature reuse
		func() *txExecTest {
			return &txExecTest{
				name: "ok - multiple MultiAddresses in one TX - no signature reuse",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 4,
					addressesFunc: func(ed25519Addresses []axongo.Address) []axongo.Address {
						return []axongo.Address{
							&axongo.MultiAddress{
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
							&axongo.MultiAddress{
								Addresses: []*axongo.AddressWithWeight{
									{
										// optional
										Address: ed25519Addresses[0],
										Weight:  1,
									},
									{
										// optional
										Address: ed25519Addresses[2],
										Weight:  1,
									},
									{
										// mandatory
										Address: ed25519Addresses[3],
										Weight:  2,
									},
								},
								Threshold: 3,
							},
						}
					},
					inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
						return []axongo.Output{
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: testAddresses[1]},
								},
							},
						}
					},
					outputsFunc: nil,
					unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
						return axongo.Unlocks{
							&axongo.MultiUnlock{
								Unlocks: []axongo.Unlock{
									&axongo.SignatureUnlock{Signature: sigs[0]},
									&axongo.SignatureUnlock{Signature: sigs[1]},
								},
							},
							&axongo.MultiUnlock{
								Unlocks: []axongo.Unlock{
									&axongo.EmptyUnlock{},
									&axongo.SignatureUnlock{Signature: sigs[2]},
									&axongo.SignatureUnlock{Signature: sigs[3]},
								},
							},
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - multiple MultiAddresses in one TX - signature reuse in different multi unlocks
		func() *txExecTest {
			return &txExecTest{
				name: "ok - multiple MultiAddresses in one TX - signature reuse in different multi unlocks",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 4,
					addressesFunc: func(ed25519Addresses []axongo.Address) []axongo.Address {
						return []axongo.Address{
							&axongo.MultiAddress{
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
							&axongo.MultiAddress{
								Addresses: []*axongo.AddressWithWeight{
									{
										// optional
										Address: ed25519Addresses[0],
										Weight:  1,
									},
									{
										// optional
										Address: ed25519Addresses[2],
										Weight:  1,
									},
									{
										// mandatory
										Address: ed25519Addresses[3],
										Weight:  2,
									},
								},
								Threshold: 3,
							},
						}
					},
					inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
						return []axongo.Output{
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
							&axongo.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: axongo.BasicOutputUnlockConditions{
									&axongo.AddressUnlockCondition{Address: testAddresses[1]},
								},
							},
						}
					},
					outputsFunc: nil,
					unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
						return axongo.Unlocks{
							&axongo.MultiUnlock{
								Unlocks: []axongo.Unlock{
									&axongo.SignatureUnlock{Signature: sigs[0]},
									&axongo.SignatureUnlock{Signature: sigs[1]},
								},
							},
							&axongo.MultiUnlock{
								Unlocks: []axongo.Unlock{
									&axongo.SignatureUnlock{Signature: sigs[0]},
									&axongo.EmptyUnlock{},
									&axongo.SignatureUnlock{Signature: sigs[3]},
								},
							},
						}
					},
				},
				wantErr: nil,
			}
		}(),
	}
	for _, tt := range tests {
		runNovaTransactionExecutionTest(t, tt)
	}
}

func TestNovaTransactionExecution_TxCapabilities(t *testing.T) {

	defaultAmount := OneIOTA

	// builds a transaction that burns native tokens
	burnNativeTokenTxBuilder := &txBuilder{
		ed25519AddrCnt: 1,
		inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
			return []axongo.Output{
				&axongo.BasicOutput{
					Amount: defaultAmount,
					// add native tokens
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
					Features: axongo.BasicOutputFeatures{
						tpkg.RandNativeTokenFeature(),
					},
				},
			}
		},
		outputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address, totalInputAmount axongo.BaseToken, totalInputMana axongo.Mana) axongo.TxEssenceOutputs {
			return axongo.TxEssenceOutputs{
				&axongo.BasicOutput{
					Amount: totalInputAmount,
					Mana:   totalInputMana,
					// burn the native tokens
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
				},
			}
		},
		unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
			return axongo.Unlocks{
				&axongo.SignatureUnlock{Signature: sigs[0]},
			}
		},
	}

	// builds a transaction that melts native tokens
	meltNativeTokenTxBuilder := &txBuilder{
		ed25519AddrCnt: 1,
		addressesFunc: func(ed25519Addresses []axongo.Address) []axongo.Address {
			accountAddress := tpkg.RandAccountAddress()
			return []axongo.Address{
				accountAddress,
			}
		},
		inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
			foundryID, err := axongo.FoundryIDFromAddressAndSerialNumberAndTokenScheme(testAddresses[0], 1, axongo.TokenSchemeSimple)
			require.NoError(t, err)

			return []axongo.Output{
				&axongo.AccountOutput{
					Amount:         defaultAmount,
					Mana:           0,
					AccountID:      testAddresses[0].(*axongo.AccountAddress).AccountID(),
					FoundryCounter: 1,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
					Features:          axongo.AccountOutputFeatures{},
					ImmutableFeatures: axongo.AccountOutputImmFeatures{},
				},
				&axongo.FoundryOutput{
					Amount:       defaultAmount,
					SerialNumber: 1,
					TokenScheme: &axongo.SimpleTokenScheme{
						MintedTokens:  big.NewInt(100),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: big.NewInt(100),
					},
					UnlockConditions: axongo.FoundryOutputUnlockConditions{
						&axongo.ImmutableAccountUnlockCondition{
							Address: testAddresses[0].(*axongo.AccountAddress),
						},
					},
					Features:          axongo.FoundryOutputFeatures{},
					ImmutableFeatures: axongo.FoundryOutputImmFeatures{},
				},
				&axongo.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
					Features: axongo.BasicOutputFeatures{
						&axongo.NativeTokenFeature{
							ID:     foundryID,
							Amount: big.NewInt(100),
						},
					},
				},
			}
		},
		outputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address, totalInputAmount axongo.BaseToken, totalInputMana axongo.Mana) axongo.TxEssenceOutputs {
			foundryID, err := axongo.FoundryIDFromAddressAndSerialNumberAndTokenScheme(testAddresses[0], 1, axongo.TokenSchemeSimple)
			require.NoError(t, err)

			return axongo.TxEssenceOutputs{
				&axongo.AccountOutput{
					Amount:         totalInputAmount - defaultAmount,
					Mana:           totalInputMana,
					AccountID:      testAddresses[0].(*axongo.AccountAddress).AccountID(),
					FoundryCounter: 1,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
					Features:          axongo.AccountOutputFeatures{},
					ImmutableFeatures: axongo.AccountOutputImmFeatures{},
				},
				&axongo.FoundryOutput{
					Amount:       defaultAmount,
					SerialNumber: 1,
					TokenScheme: &axongo.SimpleTokenScheme{
						// melt the native tokens
						MintedTokens:  big.NewInt(100),
						MeltedTokens:  big.NewInt(50),
						MaximumSupply: big.NewInt(100),
					},
					UnlockConditions: axongo.FoundryOutputUnlockConditions{
						&axongo.ImmutableAccountUnlockCondition{
							Address: testAddresses[0].(*axongo.AccountAddress),
						},
					},
					Features: axongo.FoundryOutputFeatures{
						&axongo.NativeTokenFeature{
							ID:     foundryID,
							Amount: big.NewInt(50),
						},
					},
					ImmutableFeatures: axongo.FoundryOutputImmFeatures{},
				},
			}
		},
		unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
			return axongo.Unlocks{
				&axongo.SignatureUnlock{Signature: sigs[0]},
				&axongo.AccountUnlock{Reference: 0},
				&axongo.ReferenceUnlock{Reference: 0},
			}
		},
	}

	// builds a transaction that burns and melts native tokens
	burnAndMeltNativeTokenTxBuilder := &txBuilder{
		ed25519AddrCnt: 1,
		addressesFunc: func(ed25519Addresses []axongo.Address) []axongo.Address {
			accountAddress := tpkg.RandAccountAddress()
			return []axongo.Address{
				accountAddress,
			}
		},
		inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
			foundryID, err := axongo.FoundryIDFromAddressAndSerialNumberAndTokenScheme(testAddresses[0], 1, axongo.TokenSchemeSimple)
			require.NoError(t, err)

			return []axongo.Output{
				&axongo.AccountOutput{
					Amount:         defaultAmount,
					Mana:           0,
					AccountID:      testAddresses[0].(*axongo.AccountAddress).AccountID(),
					FoundryCounter: 1,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
					Features:          axongo.AccountOutputFeatures{},
					ImmutableFeatures: axongo.AccountOutputImmFeatures{},
				},
				&axongo.FoundryOutput{
					Amount:       defaultAmount,
					SerialNumber: 1,
					TokenScheme: &axongo.SimpleTokenScheme{
						MintedTokens:  big.NewInt(100),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: big.NewInt(100),
					},
					UnlockConditions: axongo.FoundryOutputUnlockConditions{
						&axongo.ImmutableAccountUnlockCondition{
							Address: testAddresses[0].(*axongo.AccountAddress),
						},
					},
					Features:          axongo.FoundryOutputFeatures{},
					ImmutableFeatures: axongo.FoundryOutputImmFeatures{},
				},
				&axongo.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
					Features: axongo.BasicOutputFeatures{
						&axongo.NativeTokenFeature{
							ID:     foundryID,
							Amount: big.NewInt(100),
						},
					},
				},
			}
		},
		outputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address, totalInputAmount axongo.BaseToken, totalInputMana axongo.Mana) axongo.TxEssenceOutputs {
			return axongo.TxEssenceOutputs{
				&axongo.AccountOutput{
					Amount:         totalInputAmount - defaultAmount,
					Mana:           totalInputMana,
					AccountID:      testAddresses[0].(*axongo.AccountAddress).AccountID(),
					FoundryCounter: 1,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
					Features:          axongo.AccountOutputFeatures{},
					ImmutableFeatures: axongo.AccountOutputImmFeatures{},
				},
				&axongo.FoundryOutput{
					Amount:       defaultAmount,
					SerialNumber: 1,
					TokenScheme: &axongo.SimpleTokenScheme{
						// melt the native tokens
						MintedTokens:  big.NewInt(100),
						MeltedTokens:  big.NewInt(50),
						MaximumSupply: big.NewInt(100),
					},
					UnlockConditions: axongo.FoundryOutputUnlockConditions{
						&axongo.ImmutableAccountUnlockCondition{
							Address: testAddresses[0].(*axongo.AccountAddress),
						},
					},
					Features:          axongo.FoundryOutputFeatures{},
					ImmutableFeatures: axongo.FoundryOutputImmFeatures{},
				},
			}
		},
		unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
			return axongo.Unlocks{
				&axongo.SignatureUnlock{Signature: sigs[0]},
				&axongo.AccountUnlock{Reference: 0},
				&axongo.ReferenceUnlock{Reference: 0},
			}
		},
	}

	// builds a transaction that burns mana
	burnManaTxBuilder := &txBuilder{
		ed25519AddrCnt: 1,
		inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
			return []axongo.Output{
				&axongo.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
				},
			}
		},
		outputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address, totalInputAmount axongo.BaseToken, totalInputMana axongo.Mana) axongo.TxEssenceOutputs {
			return axongo.TxEssenceOutputs{
				&axongo.BasicOutput{
					Amount: totalInputAmount,
					// burn mana
					Mana: totalInputMana - 10,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
				},
			}
		},
		unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
			return axongo.Unlocks{
				&axongo.SignatureUnlock{Signature: sigs[0]},
			}
		},
	}

	// builds a transaction that destroys an account
	destroyAccountTxBuilder := &txBuilder{
		ed25519AddrCnt: 1,
		inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
			return []axongo.Output{
				&axongo.AccountOutput{
					Amount: defaultAmount,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
				},
			}
		},
		outputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address, totalInputAmount axongo.BaseToken, totalInputMana axongo.Mana) axongo.TxEssenceOutputs {
			return axongo.TxEssenceOutputs{
				// destroy the account output
				&axongo.BasicOutput{
					Amount: totalInputAmount,
					Mana:   totalInputMana,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
				},
			}
		},
		unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
			return axongo.Unlocks{
				&axongo.SignatureUnlock{Signature: sigs[0]},
			}
		},
	}

	// builds a transaction that destroys an anchor
	destroyAnchorTxBuilder := &txBuilder{
		ed25519AddrCnt: 1,
		inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
			return []axongo.Output{
				&axongo.AnchorOutput{
					Amount: defaultAmount,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.GovernorAddressUnlockCondition{Address: ed25519Addresses[0]},
						&axongo.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
					},
				},
			}
		},
		outputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address, totalInputAmount axongo.BaseToken, totalInputMana axongo.Mana) axongo.TxEssenceOutputs {
			return axongo.TxEssenceOutputs{
				// destroy the anchor output
				&axongo.BasicOutput{
					Amount: totalInputAmount,
					Mana:   totalInputMana,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
				},
			}
		},
		unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
			return axongo.Unlocks{
				&axongo.SignatureUnlock{Signature: sigs[0]},
			}
		},
	}

	// builds a transaction that destroys a foundry
	destroyFoundryTxBuilder := &txBuilder{
		ed25519AddrCnt: 1,
		addressesFunc: func(ed25519Addresses []axongo.Address) []axongo.Address {
			accountAddress := tpkg.RandAccountAddress()
			return []axongo.Address{
				accountAddress,
			}
		},
		inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
			return []axongo.Output{
				&axongo.AccountOutput{
					Amount:         defaultAmount,
					Mana:           0,
					AccountID:      testAddresses[0].(*axongo.AccountAddress).AccountID(),
					FoundryCounter: 1,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
					Features:          axongo.AccountOutputFeatures{},
					ImmutableFeatures: axongo.AccountOutputImmFeatures{},
				},
				&axongo.FoundryOutput{
					Amount:       defaultAmount,
					SerialNumber: 1,
					TokenScheme: &axongo.SimpleTokenScheme{
						MintedTokens:  big.NewInt(100),
						MeltedTokens:  big.NewInt(100),
						MaximumSupply: big.NewInt(100),
					},
					UnlockConditions: axongo.FoundryOutputUnlockConditions{
						&axongo.ImmutableAccountUnlockCondition{
							Address: testAddresses[0].(*axongo.AccountAddress),
						},
					},
					Features:          axongo.FoundryOutputFeatures{},
					ImmutableFeatures: axongo.FoundryOutputImmFeatures{},
				},
			}
		},
		outputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address, totalInputAmount axongo.BaseToken, totalInputMana axongo.Mana) axongo.TxEssenceOutputs {
			return axongo.TxEssenceOutputs{
				// destroy the foundry output
				&axongo.AccountOutput{
					Amount:         totalInputAmount,
					Mana:           totalInputMana,
					AccountID:      testAddresses[0].(*axongo.AccountAddress).AccountID(),
					FoundryCounter: 1,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
					Features:          axongo.AccountOutputFeatures{},
					ImmutableFeatures: axongo.AccountOutputImmFeatures{},
				},
			}
		},
		unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
			return axongo.Unlocks{
				&axongo.SignatureUnlock{Signature: sigs[0]},
				&axongo.AccountUnlock{Reference: 0},
			}
		},
	}

	// builds a transaction that destroys a NFT
	destroyNFTTxBuilder := &txBuilder{
		ed25519AddrCnt: 1,
		inputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address) []axongo.Output {
			return []axongo.Output{
				&axongo.NFTOutput{
					Amount: defaultAmount,
					UnlockConditions: axongo.NFTOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
				},
			}
		},
		outputsFunc: func(ed25519Addresses []axongo.Address, testAddresses []axongo.Address, totalInputAmount axongo.BaseToken, totalInputMana axongo.Mana) axongo.TxEssenceOutputs {
			return axongo.TxEssenceOutputs{
				// destroy the NFT output
				&axongo.BasicOutput{
					Amount: totalInputAmount,
					Mana:   totalInputMana,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
				},
			}
		},
		unlocksFunc: func(sigs []axongo.Signature, testAddresses []axongo.Address) axongo.Unlocks {
			return axongo.Unlocks{
				&axongo.SignatureUnlock{Signature: sigs[0]},
			}
		},
	}

	tests := []*txExecTest{
		// ok - burn native tokens (burning enabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "ok - burn native tokens (burning enabled)",
				txBuilder: burnNativeTokenTxBuilder,
				txPreSignHook: func(t *axongo.Transaction) {
					t.Capabilities = axongo.TransactionCapabilitiesBitMaskWithCapabilities(
						axongo.WithTransactionCanBurnNativeTokens(true),
					)
				},
				wantErr: nil,
			}
		}(),

		// fail - burn native tokens (burning disabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "fail - burn native tokens (burning disabled)",
				txBuilder: burnNativeTokenTxBuilder,
				txPreSignHook: func(t *axongo.Transaction) {
					t.Capabilities = axongo.TransactionCapabilitiesBitMaskWithCapabilities(
						axongo.WithTransactionCanDoAnything(),
						axongo.WithTransactionCanBurnNativeTokens(false),
					)
				},
				wantErr: axongo.ErrTxCapabilitiesNativeTokenBurningNotAllowed,
			}
		}(),

		// ok - melt native tokens (burning enabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "ok - melt native tokens (burning enabled)",
				txBuilder: meltNativeTokenTxBuilder,
				txPreSignHook: func(t *axongo.Transaction) {
					t.Capabilities = axongo.TransactionCapabilitiesBitMaskWithCapabilities(
						axongo.WithTransactionCanBurnNativeTokens(true),
					)
				},
				wantErr: nil,
			}
		}(),

		// ok - melt native tokens (burning disabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "ok - melt native tokens (burning disabled)",
				txBuilder: meltNativeTokenTxBuilder,
				txPreSignHook: func(t *axongo.Transaction) {
					t.Capabilities = axongo.TransactionCapabilitiesBitMaskWithCapabilities(
						axongo.WithTransactionCanDoAnything(),
						axongo.WithTransactionCanBurnNativeTokens(false),
					)
				},
				wantErr: nil,
			}
		}(),

		// fail - burn and melt native tokens (burning enabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "fail - burn and melt native tokens (burning enabled)",
				txBuilder: burnAndMeltNativeTokenTxBuilder,
				txPreSignHook: func(t *axongo.Transaction) {
					t.Capabilities = axongo.TransactionCapabilitiesBitMaskWithCapabilities(
						axongo.WithTransactionCanBurnNativeTokens(true),
					)
				},
				wantErr: axongo.ErrSimpleTokenSchemeMeltingInvalid,
			}
		}(),

		// fail - burn and melt native tokens (burning disabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "fail - burn and melt native tokens (burning disabled)",
				txBuilder: burnAndMeltNativeTokenTxBuilder,
				txPreSignHook: func(t *axongo.Transaction) {
					t.Capabilities = axongo.TransactionCapabilitiesBitMaskWithCapabilities(
						axongo.WithTransactionCanDoAnything(),
						axongo.WithTransactionCanBurnNativeTokens(false),
					)
				},
				wantErr: axongo.ErrSimpleTokenSchemeMeltingInvalid,
			}
		}(),

		// ok - burn mana (burning enabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "ok - burn mana (burning enabled)",
				txBuilder: burnManaTxBuilder,
				txPreSignHook: func(t *axongo.Transaction) {
					t.Capabilities = axongo.TransactionCapabilitiesBitMaskWithCapabilities(
						axongo.WithTransactionCanBurnMana(true),
					)
				},
				wantErr: nil,
			}
		}(),

		// fail - burn mana (burning disabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "fail - burn mana (burning disabled)",
				txBuilder: burnManaTxBuilder,
				txPreSignHook: func(t *axongo.Transaction) {
					t.Capabilities = axongo.TransactionCapabilitiesBitMaskWithCapabilities(
						axongo.WithTransactionCanDoAnything(),
						axongo.WithTransactionCanBurnMana(false),
					)
				},
				wantErr: axongo.ErrTxCapabilitiesManaBurningNotAllowed,
			}
		}(),

		// ok - destroy account (destruction enabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "ok - destroy account (destruction enabled)",
				txBuilder: destroyAccountTxBuilder,
				txPreSignHook: func(t *axongo.Transaction) {
					t.Capabilities = axongo.TransactionCapabilitiesBitMaskWithCapabilities(
						axongo.WithTransactionCanDestroyAccountOutputs(true),
					)
				},
				wantErr: nil,
			}
		}(),

		// fail - destroy account (destruction disabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "fail - destroy account (destruction disabled)",
				txBuilder: destroyAccountTxBuilder,
				txPreSignHook: func(t *axongo.Transaction) {
					t.Capabilities = axongo.TransactionCapabilitiesBitMaskWithCapabilities(
						axongo.WithTransactionCanDoAnything(),
						axongo.WithTransactionCanDestroyAccountOutputs(false),
					)
				},
				wantErr: axongo.ErrTxCapabilitiesAccountDestructionNotAllowed,
			}
		}(),

		// ok - destroy anchor (destruction enabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "ok - destroy anchor (destruction enabled)",
				txBuilder: destroyAnchorTxBuilder,
				txPreSignHook: func(t *axongo.Transaction) {
					t.Capabilities = axongo.TransactionCapabilitiesBitMaskWithCapabilities(
						axongo.WithTransactionCanDestroyAnchorOutputs(true),
					)
				},
				wantErr: nil,
			}
		}(),

		// fail - destroy anchor (destruction disabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "fail - destroy anchor (destruction disabled)",
				txBuilder: destroyAnchorTxBuilder,
				txPreSignHook: func(t *axongo.Transaction) {
					t.Capabilities = axongo.TransactionCapabilitiesBitMaskWithCapabilities(
						axongo.WithTransactionCanDoAnything(),
						axongo.WithTransactionCanDestroyAnchorOutputs(false),
					)
				},
				wantErr: axongo.ErrTxCapabilitiesAnchorDestructionNotAllowed,
			}
		}(),

		// ok - destroy foundry (destruction enabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "ok - destroy foundry (destruction enabled)",
				txBuilder: destroyFoundryTxBuilder,
				txPreSignHook: func(t *axongo.Transaction) {
					t.Capabilities = axongo.TransactionCapabilitiesBitMaskWithCapabilities(
						axongo.WithTransactionCanDestroyFoundryOutputs(true),
					)
				},
				wantErr: nil,
			}
		}(),

		// fail - destroy foundry (destruction disabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "fail - destroy foundry (destruction disabled)",
				txBuilder: destroyFoundryTxBuilder,
				txPreSignHook: func(t *axongo.Transaction) {
					t.Capabilities = axongo.TransactionCapabilitiesBitMaskWithCapabilities(
						axongo.WithTransactionCanDoAnything(),
						axongo.WithTransactionCanDestroyFoundryOutputs(false),
					)
				},
				wantErr: axongo.ErrTxCapabilitiesFoundryDestructionNotAllowed,
			}
		}(),

		// ok - destroy NFT (destruction enabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "ok - destroy NFT (destruction enabled)",
				txBuilder: destroyNFTTxBuilder,
				txPreSignHook: func(t *axongo.Transaction) {
					t.Capabilities = axongo.TransactionCapabilitiesBitMaskWithCapabilities(
						axongo.WithTransactionCanDestroyNFTOutputs(true),
					)
				},
				wantErr: nil,
			}
		}(),

		// fail - destroy NFT (destruction disabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "fail - destroy NFT (destruction disabled)",
				txBuilder: destroyNFTTxBuilder,
				txPreSignHook: func(t *axongo.Transaction) {
					t.Capabilities = axongo.TransactionCapabilitiesBitMaskWithCapabilities(
						axongo.WithTransactionCanDoAnything(),
						axongo.WithTransactionCanDestroyNFTOutputs(false),
					)
				},
				wantErr: axongo.ErrTxCapabilitiesNFTDestructionNotAllowed,
			}
		}(),
	}

	for _, tt := range tests {
		runNovaTransactionExecutionTest(t, tt)
	}
}

func TestTxSemanticInputUnlocks(t *testing.T) {
	type test struct {
		name           string
		vmParams       *vm.Params
		resolvedInputs vm.ResolvedInputs
		tx             *axongo.SignedTransaction
		wantErr        error
	}
	tests := []*test{
		// ok
		func() *test {
			_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
			_, addr2, addr2AddrKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(12)

			accountInputID := inputIDs[3]
			accountaddr1 := axongo.AccountAddressFromOutputID(accountInputID)

			anchorInputID := inputIDs[8]
			anchoraddr1 := axongo.AnchorAddressFromOutputID(anchorInputID)

			nftaddr1 := tpkg.RandNFTAddress()
			nftaddr2 := tpkg.RandNFTAddress()

			defaultAmount := OneIOTA

			inputs := vm.InputSet{
				// basic output to create a signature unlock (owned by addr1)
				inputIDs[0]: &axongo.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
				// basic output unlockable by sender as expired (owned by addr2)
				inputIDs[1]: &axongo.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
						&axongo.ExpirationUnlockCondition{
							ReturnAddress: addr2,
							Slot:          5,
						},
					},
				},
				// basic output not unlockable by sender as not expired (owned by addr1)
				inputIDs[2]: &axongo.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
						&axongo.ExpirationUnlockCondition{
							ReturnAddress: addr2,
							Slot:          30,
						},
					},
				},

				// account output that ownes the following outputs (owned by addr1)
				accountInputID: &axongo.AccountOutput{
					Amount:    defaultAmount,
					AccountID: axongo.AccountID{}, // empty on purpose as validation should resolve
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
				// basic output (owned by accountaddr1)
				inputIDs[4]: &axongo.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: accountaddr1},
					},
				},
				// NFT output (owned by accountaddr1)
				inputIDs[5]: &axongo.NFTOutput{
					Amount: defaultAmount,
					NFTID:  nftaddr1.NFTID(),
					UnlockConditions: axongo.NFTOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: accountaddr1},
					},
				},
				// basic output (owned by nftaddr1)
				inputIDs[6]: &axongo.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: nftaddr1},
					},
				},
				// foundry output (owned by accountaddr1)
				inputIDs[7]: &axongo.FoundryOutput{
					Amount:       defaultAmount,
					SerialNumber: 0,
					TokenScheme: &axongo.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetInt64(100),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetInt64(1000),
					},
					UnlockConditions: axongo.FoundryOutputUnlockConditions{
						&axongo.ImmutableAccountUnlockCondition{Address: accountaddr1},
					},
				},

				// anchor output that ownes the following outputs (owned by addr1)
				inputIDs[8]: &axongo.AnchorOutput{
					Amount:   defaultAmount,
					AnchorID: axongo.AnchorID{}, // empty on purpose as validation should resolve
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: addr1},
						&axongo.GovernorAddressUnlockCondition{Address: addr1},
					},
				},
				// basic output (owned by anchoraddr1)
				inputIDs[9]: &axongo.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: anchoraddr1},
					},
				},
				// NFT output (owned by anchoraddr1)
				inputIDs[10]: &axongo.NFTOutput{
					Amount: defaultAmount,
					NFTID:  nftaddr2.NFTID(),
					UnlockConditions: axongo.NFTOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: anchoraddr1},
					},
				},
				// basic output (owned by nftaddr2)
				inputIDs[11]: &axongo.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: nftaddr2},
					},
				},
			}

			creationSlot := axongo.SlotIndex(10)
			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs:       inputIDs.UTXOInputs(),
					CreationSlot: creationSlot,
					Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.AccountOutput{
						Amount:    defaultAmount / 2,
						AccountID: accountaddr1.AccountID(),
						UnlockConditions: axongo.AccountOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr1},
						},
					},
					&axongo.AnchorOutput{
						Amount:     defaultAmount / 2,
						AnchorID:   anchoraddr1.AnchorID(),
						StateIndex: 1,
						UnlockConditions: axongo.AnchorOutputUnlockConditions{
							&axongo.StateControllerAddressUnlockCondition{Address: addr1},
							&axongo.GovernorAddressUnlockCondition{Address: addr1},
						},
					},
				},
			}

			sigs, err := transaction.Sign(addr1AddrKeys, addr2AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
					CommitmentInput: &axongo.Commitment{
						Slot: axongo.SlotIndex(0),
					},
				},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]}, // basic output (owned by addr1)
						&axongo.SignatureUnlock{Signature: sigs[1]}, // basic output (owned by addr2)
						&axongo.ReferenceUnlock{Reference: 0},       // basic output (owned by addr1)
						&axongo.ReferenceUnlock{Reference: 0},       // account output (owned by addr1)
						&axongo.AccountUnlock{Reference: 3},         // basic output (owned by accountaddr1)
						&axongo.AccountUnlock{Reference: 3},         // NFT output (owned by accountaddr1)
						&axongo.NFTUnlock{Reference: 5},             // basic output (owned by nftaddr1)
						&axongo.AccountUnlock{Reference: 3},         // foundry output (owned by accountaddr1)
						&axongo.ReferenceUnlock{Reference: 0},       // anchor output (owned by addr1)
						&axongo.AnchorUnlock{Reference: 8},          // basic output (owned by anchoraddr1)
						&axongo.AnchorUnlock{Reference: 8},          // NFT output (owned by anchoraddr1)
						&axongo.NFTUnlock{Reference: 10},            // basic output (owned by nftaddr2
					},
				},
				wantErr: nil,
			}
		}(),

		// fail - invalid signature
		func() *test {
			addr1Sk, addr1, _ := tpkg.RandEd25519Identity()
			_, _, addr2AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
			}

			transaction := &axongo.Transaction{
				API:                testAPI,
				TransactionEssence: &axongo.TransactionEssence{Inputs: inputIDs.UTXOInputs()},
			}

			sigs, err := transaction.Sign(addr2AddrKeys)
			require.NoError(t, err)

			copy(sigs[0].(*axongo.Ed25519Signature).PublicKey[:], addr1Sk.Public().(ed25519.PublicKey))

			return &test{
				name: "fail - invalid signature",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: axongo.ErrEd25519SignatureInvalid,
			}
		}(),

		// fail - should contain reference unlock
		func() *test {
			_, addr1, addr1AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
				inputIDs[1]: &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
			}

			transaction := &axongo.Transaction{API: testAPI, TransactionEssence: &axongo.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
			}}

			sigs, err := transaction.Sign(addr1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - should contain reference unlock",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: axongo.ErrDirectUnlockableAddressUnlockInvalid,
			}
		}(),

		// fail - should contain account unlock
		func() *test {
			_, addr1, addr1AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)

			accountaddr1 := axongo.AccountAddressFromOutputID(inputIDs[0])
			inputs := vm.InputSet{
				inputIDs[0]: &axongo.AccountOutput{
					Amount:    100,
					AccountID: axongo.AccountID{},
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
				inputIDs[1]: &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: accountaddr1},
					},
				},
			}

			transaction := &axongo.Transaction{API: testAPI, TransactionEssence: &axongo.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
			}}

			sigs, err := transaction.Sign(addr1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - should contain account unlock",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
						&axongo.ReferenceUnlock{Reference: 0},
					},
				},
				wantErr: axongo.ErrChainAddressUnlockInvalid,
			}
		}(),

		// fail - should contain anchor unlock
		func() *test {
			_, addr1, addr1AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)

			anchoraddr1 := axongo.AnchorAddressFromOutputID(inputIDs[0])
			inputs := vm.InputSet{
				inputIDs[0]: &axongo.AnchorOutput{
					Amount:   100,
					AnchorID: axongo.AnchorID{},
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: addr1},
						&axongo.GovernorAddressUnlockCondition{Address: addr1},
					},
				},
				inputIDs[1]: &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: anchoraddr1},
					},
				},
			}

			transaction := &axongo.Transaction{API: testAPI, TransactionEssence: &axongo.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
			}}

			sigs, err := transaction.Sign(addr1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - should contain anchor unlock",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
						&axongo.ReferenceUnlock{Reference: 0},
					},
				},
				wantErr: axongo.ErrChainAddressUnlockInvalid,
			}
		}(),

		// fail - should contain NFT unlock
		func() *test {
			_, addr1, addr1AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)

			nftaddr1 := axongo.NFTAddressFromOutputID(inputIDs[0])
			inputs := vm.InputSet{
				inputIDs[0]: &axongo.NFTOutput{
					Amount: 100,
					NFTID:  axongo.NFTID{},
					UnlockConditions: axongo.NFTOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
				inputIDs[1]: &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: nftaddr1},
					},
				},
			}

			transaction := &axongo.Transaction{API: testAPI, TransactionEssence: &axongo.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
			}}

			sigs, err := transaction.Sign(addr1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - should contain NFT unlock",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
						&axongo.ReferenceUnlock{Reference: 0},
					},
				},
				wantErr: axongo.ErrChainAddressUnlockInvalid,
			}
		}(),

		// fail - circular NFT unlock
		func() *test {
			inputIDs := tpkg.RandOutputIDs(2)

			nftaddr1 := axongo.NFTAddressFromOutputID(inputIDs[0])
			nftaddr2 := axongo.NFTAddressFromOutputID(inputIDs[1])

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.NFTOutput{
					Amount: 100,
					NFTID:  nftaddr1.NFTID(),
					UnlockConditions: axongo.NFTOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: nftaddr2},
					},
				},
				inputIDs[1]: &axongo.NFTOutput{
					Amount: 100,
					NFTID:  nftaddr2.NFTID(),
					UnlockConditions: axongo.NFTOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: nftaddr2},
					},
				},
			}

			transaction := &axongo.Transaction{API: testAPI, TransactionEssence: &axongo.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
			}}
			_, err := transaction.Sign()
			require.NoError(t, err)
			return &test{
				name: "fail - circular NFT unlock",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.NFTUnlock{Reference: 1},
						&axongo.NFTUnlock{Reference: 0},
					},
				},
				wantErr: axongo.ErrChainAddressUnlockInvalid,
			}
		}(),

		// fail - should contain sig unlock
		func() *test {
			_, addr1, addr1AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.AccountOutput{
					Amount:    100,
					AccountID: axongo.AccountID{},
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
			}

			transaction := &axongo.Transaction{API: testAPI, TransactionEssence: &axongo.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
			}}

			_, err := transaction.Sign(addr1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - should contain sig unlock",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.AccountUnlock{Reference: 0},
					},
				},
				wantErr: axongo.ErrDirectUnlockableAddressUnlockInvalid,
			}
		}(),

		// fail - reference unlock pointee invalid
		func() *test {
			_, addr1, addr1AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.AccountOutput{
					Amount:    100,
					AccountID: axongo.AccountID{},
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
			}

			transaction := &axongo.Transaction{API: testAPI, TransactionEssence: &axongo.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
			}}

			_, err := transaction.Sign(addr1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - reference unlock pointee invalid",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.ReferenceUnlock{Reference: 0},
					},
				},
				wantErr: axongo.ErrDirectUnlockableAddressUnlockInvalid,
			}
		}(),

		// fail - sender can not unlock yet
		func() *test {
			_, addr1, _ := tpkg.RandEd25519Identity()
			_, addr2, addr2AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
						&axongo.ExpirationUnlockCondition{
							ReturnAddress: addr2,
							Slot:          20,
						},
					},
				},
			}

			creationSlot := axongo.SlotIndex(5)
			transaction := &axongo.Transaction{API: testAPI, TransactionEssence: &axongo.TransactionEssence{Inputs: inputIDs.UTXOInputs(), CreationSlot: creationSlot}}

			sigs, err := transaction.Sign(addr2AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - sender can not unlock yet",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
					CommitmentInput: &axongo.Commitment{
						Slot: axongo.SlotIndex(0),
					},
				},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: axongo.ErrExpirationNotUnlockable,
			}
		}(),

		// fail - receiver can not unlock anymore
		func() *test {
			_, addr1, addr1AddressKeys := tpkg.RandEd25519Identity()
			_, addr2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
						&axongo.ExpirationUnlockCondition{
							ReturnAddress: addr2,
							Slot:          10,
						},
					},
				},
			}

			creationSlot := axongo.SlotIndex(10)
			transaction := &axongo.Transaction{API: testAPI, TransactionEssence: &axongo.TransactionEssence{Inputs: inputIDs.UTXOInputs(), CreationSlot: creationSlot}}

			sigs, err := transaction.Sign(addr1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - receiver can not unlock anymore",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
					CommitmentInput: &axongo.Commitment{
						Slot: creationSlot,
					},
				},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: axongo.ErrEd25519PubKeyAndAddrMismatch,
			}
		}(),

		// fail - referencing other account unlocked by source account
		func() *test {
			_, addr1, addr1AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(3)

			var (
				accountAddr1 = tpkg.RandAccountAddress()
				accountAddr2 = tpkg.RandAccountAddress()
				accountAddr3 = tpkg.RandAccountAddress()
			)

			inputs := vm.InputSet{
				// owned by addr1
				inputIDs[0]: &axongo.AccountOutput{
					Amount:    100,
					AccountID: accountAddr1.AccountID(),
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
				// owned by account1
				inputIDs[1]: &axongo.AccountOutput{
					Amount:    100,
					AccountID: accountAddr2.AccountID(),
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: accountAddr1},
					},
				},
				// owned by account1
				inputIDs[2]: &axongo.AccountOutput{
					Amount:    100,
					AccountID: accountAddr3.AccountID(),
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: accountAddr1},
					},
				},
			}

			transaction := &axongo.Transaction{API: testAPI, TransactionEssence: &axongo.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
			}}

			sigs, err := transaction.Sign(addr1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - referencing other account unlocked by source account",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
						&axongo.AccountUnlock{Reference: 0},
						// error, should be 0, because account3 is unlocked by account1, not account2
						&axongo.AccountUnlock{Reference: 1},
					},
				},
				wantErr: axongo.ErrChainAddressUnlockInvalid,
			}
		}(),

		// fail - referencing other anchor unlocked by source anchor
		func() *test {
			_, addr1, addr1AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(3)

			var (
				anchorAddr1 = tpkg.RandAnchorAddress()
				anchorAddr2 = tpkg.RandAnchorAddress()
				anchorAddr3 = tpkg.RandAnchorAddress()
			)

			inputs := vm.InputSet{
				// owned by addr1
				inputIDs[0]: &axongo.AnchorOutput{
					Amount:   100,
					AnchorID: anchorAddr1.AnchorID(),
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: addr1},
						&axongo.GovernorAddressUnlockCondition{Address: addr1},
					},
				},
				// owned by anchor1
				inputIDs[1]: &axongo.AnchorOutput{
					Amount:   100,
					AnchorID: anchorAddr2.AnchorID(),
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: anchorAddr1},
						&axongo.GovernorAddressUnlockCondition{Address: anchorAddr1},
					},
				},
				// owned by anchor1
				inputIDs[2]: &axongo.AnchorOutput{
					Amount:   100,
					AnchorID: anchorAddr3.AnchorID(),
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: anchorAddr1},
						&axongo.GovernorAddressUnlockCondition{Address: anchorAddr1},
					},
				},
			}

			transaction := &axongo.Transaction{API: testAPI, TransactionEssence: &axongo.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
			}}

			sigs, err := transaction.Sign(addr1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - referencing other anchor unlocked by source anchor",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
						&axongo.AnchorUnlock{Reference: 0},
						// error, should be 0, because anchor3 is unlocked by anchor1, not anchor2
						&axongo.AnchorUnlock{Reference: 1},
					},
				},
				wantErr: axongo.ErrChainAddressUnlockInvalid,
			}
		}(),

		// fail - anchor output not state transitioning
		func() *test {
			_, addr1, _ := tpkg.RandEd25519Identity()
			_, addr2, addr2AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)

			anchorAddr1 := tpkg.RandAnchorAddress()

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.AnchorOutput{
					Amount:   100,
					AnchorID: anchorAddr1.AnchorID(),
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: addr1},
						&axongo.GovernorAddressUnlockCondition{Address: addr2},
					},
				},
				inputIDs[1]: &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: anchorAddr1},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.AnchorOutput{
						Amount:   100,
						AnchorID: anchorAddr1.AnchorID(),
						UnlockConditions: axongo.AnchorOutputUnlockConditions{
							&axongo.StateControllerAddressUnlockCondition{Address: addr1},
							&axongo.GovernorAddressUnlockCondition{Address: addr2},
						},
					},
				},
			}

			sigs, err := transaction.Sign(addr2AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - anchor output not state transitioning",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
						&axongo.AnchorUnlock{Reference: 0},
					},
				},
				wantErr: axongo.ErrChainAddressUnlockInvalid,
			}
		}(),

		// fail - wrong unlock for foundry
		func() *test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, addr1, addr1AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(2)
			foundryOutput := &axongo.FoundryOutput{
				Amount:       100,
				SerialNumber: 5,
				TokenScheme: &axongo.SimpleTokenScheme{
					MintedTokens:  new(big.Int).SetInt64(1000),
					MeltedTokens:  big.NewInt(0),
					MaximumSupply: new(big.Int).SetInt64(10000),
				},
				UnlockConditions: axongo.FoundryOutputUnlockConditions{
					&axongo.ImmutableAccountUnlockCondition{Address: accountAddr1},
				},
			}

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.AccountOutput{
					Amount:    100,
					AccountID: accountAddr1.AccountID(),
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
				inputIDs[1]: foundryOutput,
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.AccountOutput{
						Amount:    100,
						AccountID: accountAddr1.AccountID(),
						UnlockConditions: axongo.AccountOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr1},
						},
					},
					foundryOutput,
				},
			}

			sigs, err := transaction.Sign(addr1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - wrong unlock for foundry",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
						// should be an AccountUnlock
						&axongo.ReferenceUnlock{Reference: 0},
					},
				},
				wantErr: axongo.ErrChainAddressUnlockInvalid,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := novaVM.ValidateUnlocks(tt.tx, tt.resolvedInputs)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestTxSemanticDeposit(t *testing.T) {
	type test struct {
		name           string
		vmParams       *vm.Params
		resolvedInputs vm.ResolvedInputs
		tx             *axongo.SignedTransaction
		wantErr        error
	}
	tests := []*test{
		// ok
		func() *test {
			_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
			_, addr2, addr2AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(3)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
				// unlocked by addr1 as it is not expired
				inputIDs[1]: &axongo.BasicOutput{
					Amount: 500,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
						&axongo.StorageDepositReturnUnlockCondition{
							ReturnAddress: addr2,
							Amount:        420,
						},
						&axongo.ExpirationUnlockCondition{
							ReturnAddress: addr2,
							Slot:          30,
						},
					},
				},
				// unlocked by addr2 as it is expired
				inputIDs[2]: &axongo.BasicOutput{
					Amount: 500,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
						&axongo.StorageDepositReturnUnlockCondition{
							ReturnAddress: addr2,
							Amount:        420,
						},
						&axongo.ExpirationUnlockCondition{
							ReturnAddress: addr2,
							Slot:          2,
						},
					},
				},
			}

			creationSlot := axongo.SlotIndex(5)
			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs:       inputIDs.UTXOInputs(),
					CreationSlot: creationSlot,
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.BasicOutput{
						Amount: 180,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
					&axongo.BasicOutput{
						// return via addr1 + reclaim
						Amount: 420 + 500,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr2},
						},
					},
				},
			}
			sigs, err := transaction.Sign(addr1AddrKeys, addr2AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
					CommitmentInput: &axongo.Commitment{
						Slot: creationSlot,
					},
				},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
						&axongo.ReferenceUnlock{Reference: 0},
						&axongo.SignatureUnlock{Signature: sigs[1]},
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - more storage deposit returned via more outputs
		func() *test {
			_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
			_, addr2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 1000,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
						&axongo.StorageDepositReturnUnlockCondition{
							ReturnAddress: addr2,
							Amount:        420,
						},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.BasicOutput{
						// returns 200 to addr2
						Amount: 200,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr2},
						},
					},
					&axongo.BasicOutput{
						// returns 221 to addr2
						Amount: 221,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr2},
						},
					},
					&axongo.BasicOutput{
						// remainder to random address
						Amount: 579,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := transaction.Sign(addr1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - more storage deposit returned via more outputs",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		// fail - unbalanced, more on output than input
		func() *test {
			_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 50,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs:       inputIDs.UTXOInputs(),
					CreationSlot: 5,
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.BasicOutput{
						Amount: 100,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := transaction.Sign(addr1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - unbalanced, more on output than input",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: axongo.ErrInputOutputBaseTokenMismatch,
			}
		}(),

		// fail - unbalanced, more on input than output
		func() *test {
			_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs:       inputIDs.UTXOInputs(),
					CreationSlot: 5,
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.BasicOutput{
						Amount: 50,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := transaction.Sign(addr1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - unbalanced, more on input than output",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: axongo.ErrInputOutputBaseTokenMismatch,
			}
		}(),

		// fail - return not fulfilled
		func() *test {
			_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
			_, addr2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 500,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
						&axongo.StorageDepositReturnUnlockCondition{
							ReturnAddress: addr2,
							Amount:        420,
						},
						// not yet expired, so addr1 needs to unlock
						&axongo.ExpirationUnlockCondition{
							ReturnAddress: addr2,
							Slot:          30,
						},
					},
				},
			}

			creationSlot := axongo.SlotIndex(5)
			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs:       inputIDs.UTXOInputs(),
					CreationSlot: creationSlot,
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.BasicOutput{
						Amount: 500,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr1},
						},
					},
				},
			}
			sigs, err := transaction.Sign(addr1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - return not fulfilled",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
					CommitmentInput: &axongo.Commitment{
						Slot: creationSlot,
					},
				},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: axongo.ErrReturnAmountNotFulFilled,
			}
		}(),

		// fail - storage deposit return not basic output
		func() *test {
			_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
			_, addr2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 500,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
						&axongo.StorageDepositReturnUnlockCondition{
							ReturnAddress: addr2,
							Amount:        420,
						},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.BasicOutput{
						Amount: 80,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr1},
						},
					},
					&axongo.NFTOutput{
						Amount: 420,
						UnlockConditions: axongo.NFTOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr2},
						},
					},
				},
			}
			sigs, err := transaction.Sign(addr1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - storage deposit return not basic output",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: axongo.ErrReturnAmountNotFulFilled,
			}
		}(),

		// fail - storage deposit return has additional unlocks
		func() *test {
			_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
			_, addr2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 500,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
						&axongo.StorageDepositReturnUnlockCondition{
							ReturnAddress: addr2,
							Amount:        420,
						},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.BasicOutput{
						Amount: 80,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr1},
						},
					},
					&axongo.BasicOutput{
						Amount: 420,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr2},
							&axongo.ExpirationUnlockCondition{
								ReturnAddress: addr1,
								Slot:          10,
							},
						},
					},
				},
			}
			sigs, err := transaction.Sign(addr1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - storage deposit return has additional unlocks",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: axongo.ErrReturnAmountNotFulFilled,
			}
		}(),

		// fail - storage deposit return has feature
		func() *test {
			_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
			_, addr2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 500,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
						&axongo.StorageDepositReturnUnlockCondition{
							ReturnAddress: addr2,
							Amount:        420,
						},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.BasicOutput{
						Amount: 80,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr1},
						},
					},
					&axongo.BasicOutput{
						Amount: 420,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr2},
						},
						Features: axongo.BasicOutputFeatures{
							&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": []byte("foo")}},
						},
					},
				},
			}
			sigs, err := transaction.Sign(addr1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - storage deposit return has feature",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: axongo.ErrReturnAmountNotFulFilled,
			}
		}(),

		// fail - storage deposit return has native tokens
		func() *test {
			_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
			_, addr2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)
			ntID := tpkg.Rand38ByteArray()

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 500,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
						&axongo.StorageDepositReturnUnlockCondition{
							ReturnAddress: addr2,
							Amount:        420,
						},
					},
					Features: axongo.BasicOutputFeatures{
						&axongo.NativeTokenFeature{
							ID:     ntID,
							Amount: new(big.Int).SetUint64(1000),
						},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.BasicOutput{
						Amount: 80,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr1},
						},
					},
					&axongo.BasicOutput{
						Amount: 420,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr2},
						},
						Features: axongo.BasicOutputFeatures{
							&axongo.NativeTokenFeature{
								ID:     ntID,
								Amount: new(big.Int).SetUint64(1000),
							},
						},
					},
				},
			}
			sigs, err := transaction.Sign(addr1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - storage deposit return has native tokens",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: axongo.ErrReturnAmountNotFulFilled,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAndExecuteSignedTransaction(tt.tx, tt.resolvedInputs, vm.ExecFuncBalancedBaseTokens())
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestTxSemanticNativeTokens(t *testing.T) {
	foundryAccountAddr := tpkg.RandAccountAddress()
	foundryMaxSupply := new(big.Int).SetInt64(1000)
	foundryMintedSupply := new(big.Int).SetInt64(500)

	inUnrelatedFoundryOutput := &axongo.FoundryOutput{
		Amount:       100,
		SerialNumber: 0,
		TokenScheme: &axongo.SimpleTokenScheme{
			MintedTokens:  foundryMintedSupply,
			MeltedTokens:  big.NewInt(0),
			MaximumSupply: foundryMaxSupply,
		},
		UnlockConditions: axongo.FoundryOutputUnlockConditions{
			&axongo.ImmutableAccountUnlockCondition{Address: foundryAccountAddr},
		},
	}

	outUnrelatedFoundryOutput := &axongo.FoundryOutput{
		Amount:       100,
		SerialNumber: 0,
		TokenScheme: &axongo.SimpleTokenScheme{
			MintedTokens:  foundryMintedSupply,
			MeltedTokens:  big.NewInt(0),
			MaximumSupply: foundryMaxSupply,
		},
		UnlockConditions: axongo.FoundryOutputUnlockConditions{
			&axongo.ImmutableAccountUnlockCondition{Address: foundryAccountAddr},
		},
	}

	type test struct {
		name           string
		vmParams       *vm.Params
		resolvedInputs vm.ResolvedInputs
		tx             *axongo.SignedTransaction
		wantErr        error
	}
	tests := []*test{
		// ok
		func() *test {
			inputIDs := tpkg.RandOutputIDs(2)

			nativeTokenFeature1 := tpkg.RandNativeTokenFeature()
			nativeTokenFeature2 := tpkg.RandNativeTokenFeature()

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
					Features: axongo.BasicOutputFeatures{
						nativeTokenFeature1,
					},
				},
				inputIDs[1]: &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
					Features: axongo.BasicOutputFeatures{
						nativeTokenFeature2,
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.BasicOutput{
						Amount: 100,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: axongo.BasicOutputFeatures{
							nativeTokenFeature1,
						},
					},
					&axongo.BasicOutput{
						Amount: 100,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: axongo.BasicOutputFeatures{
							nativeTokenFeature2,
						},
					},
				},
			}

			return &test{
				name: "ok",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks:     axongo.Unlocks{},
				},
				wantErr: nil,
			}
		}(),

		// ok - consolidate native token (same type)
		func() *test {
			inputIDs := tpkg.RandOutputIDs(axongo.MaxInputsCount)
			nativeToken := tpkg.RandNativeTokenFeature()

			inputs := vm.InputSet{}
			for i := 0; i < axongo.MaxInputsCount; i++ {
				inputs[inputIDs[i]] = &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
					Features: axongo.BasicOutputFeatures{
						&axongo.NativeTokenFeature{
							ID:     nativeToken.ID,
							Amount: big.NewInt(1),
						},
					},
				}
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.BasicOutput{
						Amount: 100 * axongo.MaxInputsCount,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: axongo.BasicOutputFeatures{
							&axongo.NativeTokenFeature{
								ID:     nativeToken.ID,
								Amount: big.NewInt(axongo.MaxInputsCount),
							},
						},
					},
				},
			}

			return &test{
				name: "ok - consolidate native token (same type)",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks:     axongo.Unlocks{},
				},
				wantErr: nil,
			}
		}(),

		// ok - most possible tokens in a tx
		func() *test {
			inputIDs := tpkg.RandOutputIDs(axongo.MaxInputsCount)

			nativeTokenFeatures := make([]*axongo.NativeTokenFeature, axongo.MaxInputsCount)
			for i := 0; i < axongo.MaxInputsCount; i++ {
				nativeTokenFeatures[i] = tpkg.RandNativeTokenFeature()
			}

			inputs := vm.InputSet{}
			for i := 0; i < axongo.MaxInputsCount; i++ {
				inputs[inputIDs[i]] = &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
					Features: axongo.BasicOutputFeatures{
						nativeTokenFeatures[i],
					},
				}
			}

			outputs := make(axongo.TxEssenceOutputs, axongo.MaxOutputsCount)
			for i := range outputs {
				outputs[i] = &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
					Features: axongo.BasicOutputFeatures{
						nativeTokenFeatures[i],
					},
				}
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: outputs,
			}

			return &test{
				name: "ok - most possible tokens in a tx",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks:     axongo.Unlocks{},
				},
				wantErr: nil,
			}
		}(),

		// fail - unbalanced on output
		func() *test {
			inputIDs := tpkg.RandOutputIDs(1)

			nativeTokenFeature := tpkg.RandNativeTokenFeature()

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
					Features: axongo.BasicOutputFeatures{
						nativeTokenFeature,
					},
				},
			}

			// unbalance by making one token be excess on the output side
			cpyNativeTokenFeature := nativeTokenFeature.Clone()
			cpyNativeTokenFeature.(*axongo.NativeTokenFeature).Amount = big.NewInt(0).Add(nativeTokenFeature.Amount, big.NewInt(1))

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.BasicOutput{
						Amount: 100,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: axongo.BasicOutputFeatures{
							cpyNativeTokenFeature,
						},
					},
				},
			}

			return &test{
				name: "fail - unbalanced on output",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks:     axongo.Unlocks{},
				},
				wantErr: axongo.ErrNativeTokenSumUnbalanced,
			}
		}(),

		// fail - unbalanced with unrelated foundry in term of new output tokens
		func() *test {
			inputIDs := tpkg.RandOutputIDs(3)

			nativeTokenFeature1 := tpkg.RandNativeTokenFeature()
			nativeTokenFeature2 := nativeTokenFeature1.Clone()

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
					Features: axongo.BasicOutputFeatures{
						nativeTokenFeature1,
					},
				},
				inputIDs[1]: &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
					Features: axongo.BasicOutputFeatures{
						nativeTokenFeature2,
					},
				},
				inputIDs[2]: inUnrelatedFoundryOutput,
			}

			// unbalance by making one token be excess on the output side
			cpyNativeTokenFeature := nativeTokenFeature1.Clone()
			cpyNativeTokenFeature.(*axongo.NativeTokenFeature).Amount = big.NewInt(0).Add(nativeTokenFeature1.Amount, big.NewInt(1))

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.BasicOutput{
						Amount: 100,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: axongo.BasicOutputFeatures{
							cpyNativeTokenFeature,
						},
					},
					&axongo.BasicOutput{
						Amount: 100,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: axongo.BasicOutputFeatures{
							nativeTokenFeature2,
						},
					},
					outUnrelatedFoundryOutput,
				},
			}

			return &test{
				name: "fail - unbalanced with unrelated foundry in term of new output tokens",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks:     axongo.Unlocks{},
				},
				wantErr: axongo.ErrNativeTokenSumUnbalanced,
			}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := novaVM.Execute(tt.tx.Transaction, tt.resolvedInputs, make(vm.UnlockedAddresses), vm.ExecFuncBalancedNativeTokens())
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestTxSemanticOutputsSender(t *testing.T) {
	type test struct {
		name           string
		vmParams       *vm.Params
		resolvedInputs vm.ResolvedInputs
		tx             *axongo.SignedTransaction
		wantErr        error
	}
	tests := []*test{
		// ok
		func() *test {
			_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(4)
			accountAddr := tpkg.RandAccountAddress()
			anchorAddr := tpkg.RandAnchorAddress()
			nftAddr := tpkg.RandNFTAddress()

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
				inputIDs[1]: &axongo.AccountOutput{
					Amount:    100,
					AccountID: accountAddr.AccountID(),
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
				inputIDs[2]: &axongo.AnchorOutput{
					Amount:     100,
					AnchorID:   anchorAddr.AnchorID(),
					StateIndex: 1,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: addr1},
						&axongo.GovernorAddressUnlockCondition{Address: addr1},
					},
				},
				inputIDs[3]: &axongo.NFTOutput{
					Amount: 100,
					NFTID:  nftAddr.NFTID(),
					UnlockConditions: axongo.NFTOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: func() axongo.TxEssenceOutputs {
					outputs := make(axongo.TxEssenceOutputs, 0)

					// we need to do a state transition to unlock the sender feature for the anchor output
					outputs = append(outputs, &axongo.AnchorOutput{
						Amount:     100,
						AnchorID:   anchorAddr.AnchorID(),
						StateIndex: 2,
						UnlockConditions: axongo.AnchorOutputUnlockConditions{
							&axongo.StateControllerAddressUnlockCondition{Address: addr1},
							&axongo.GovernorAddressUnlockCondition{Address: addr1},
						},
					})

					for _, sender := range []axongo.Address{addr1, accountAddr, anchorAddr, nftAddr} {
						outputs = append(outputs, &axongo.BasicOutput{
							Amount: 1337,
							UnlockConditions: axongo.BasicOutputUnlockConditions{
								&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
							},
							Features: axongo.BasicOutputFeatures{
								&axongo.SenderFeature{Address: sender},
							},
						})

						outputs = append(outputs, &axongo.AccountOutput{
							Amount: 1337,
							UnlockConditions: axongo.AccountOutputUnlockConditions{
								&axongo.AddressUnlockCondition{Address: addr1},
							},
							Features: axongo.AccountOutputFeatures{
								&axongo.SenderFeature{Address: sender},
							},
						})

						outputs = append(outputs, &axongo.NFTOutput{
							Amount: 1337,
							UnlockConditions: axongo.NFTOutputUnlockConditions{
								&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
							},
							Features: axongo.NFTOutputFeatures{
								&axongo.SenderFeature{Address: sender},
							},
						})
					}
					return outputs
				}(),
			}
			sigs, err := transaction.Sign(addr1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
						&axongo.ReferenceUnlock{Reference: 0},
						&axongo.ReferenceUnlock{Reference: 0},
						&axongo.ReferenceUnlock{Reference: 0},
					},
				},
				wantErr: nil,
			}
		}(),

		// fail - sender not unlocked
		func() *test {
			_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.BasicOutput{
						Amount: 1337,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: axongo.BasicOutputFeatures{
							&axongo.SenderFeature{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := transaction.Sign(addr1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - sender not unlocked",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: axongo.ErrSenderFeatureNotUnlocked,
			}
		}(),

		// fail - sender not unlocked due to governance transition
		func() *test {
			_, stateController, _ := tpkg.RandEd25519Identity()
			_, governor, governorAddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)
			anchorAddr := tpkg.RandAnchorAddress()
			anchorID := anchorAddr.AnchorID()

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.AnchorOutput{
					Amount:   100,
					AnchorID: anchorID,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: stateController},
						&axongo.GovernorAddressUnlockCondition{Address: governor},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.AnchorOutput{
						Amount:   50,
						AnchorID: anchorID,
						UnlockConditions: axongo.AnchorOutputUnlockConditions{
							&axongo.StateControllerAddressUnlockCondition{Address: stateController},
							&axongo.GovernorAddressUnlockCondition{Address: governor},
						},
						Features: axongo.AnchorOutputFeatures{},
					},
					&axongo.BasicOutput{
						Amount: 50,
						Mana:   0,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: anchorAddr},
						},
						Features: axongo.BasicOutputFeatures{
							&axongo.SenderFeature{Address: anchorAddr},
						},
					},
				},
			}
			sigs, err := transaction.Sign(governorAddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - sender not unlocked due to governance transition",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: axongo.ErrSenderFeatureNotUnlocked,
			}
		}(),

		// ok - anchor addr unlocked with state transition
		func() *test {
			_, stateController, stateControllerAddrKeys := tpkg.RandEd25519Identity()
			_, governor, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)
			anchorAddr := tpkg.RandAnchorAddress()
			anchorID := anchorAddr.AnchorID()
			currentStateIndex := uint32(1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.AnchorOutput{
					Amount:     100,
					AnchorID:   anchorID,
					StateIndex: currentStateIndex,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: stateController},
						&axongo.GovernorAddressUnlockCondition{Address: governor},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.AnchorOutput{
						Amount:     50,
						AnchorID:   anchorID,
						StateIndex: currentStateIndex + 1,
						UnlockConditions: axongo.AnchorOutputUnlockConditions{
							&axongo.StateControllerAddressUnlockCondition{Address: stateController},
							&axongo.GovernorAddressUnlockCondition{Address: governor},
						},
						Features: axongo.AnchorOutputFeatures{},
					},
					&axongo.BasicOutput{
						Amount: 50,
						Mana:   0,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: anchorAddr},
						},
						Features: axongo.BasicOutputFeatures{
							&axongo.SenderFeature{Address: anchorAddr},
						},
					},
				},
			}
			sigs, err := transaction.Sign(stateControllerAddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - anchor addr unlocked with state transition",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - sender is governor address
		func() *test {
			_, stateController, _ := tpkg.RandEd25519Identity()
			_, governor, governorAddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)
			anchorAddr := tpkg.RandAnchorAddress()
			anchorID := anchorAddr.AnchorID()

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.AnchorOutput{
					Amount:   100,
					AnchorID: anchorID,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: stateController},
						&axongo.GovernorAddressUnlockCondition{Address: governor},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.AnchorOutput{
						Amount:   50,
						AnchorID: anchorID,
						UnlockConditions: axongo.AnchorOutputUnlockConditions{
							&axongo.StateControllerAddressUnlockCondition{Address: stateController},
							&axongo.GovernorAddressUnlockCondition{Address: governor},
						},
						Features: axongo.AnchorOutputFeatures{},
					},
					&axongo.BasicOutput{
						Amount: 50,
						Mana:   0,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: anchorAddr},
						},
						Features: axongo.BasicOutputFeatures{
							&axongo.SenderFeature{Address: governor},
						},
					},
				},
			}
			sigs, err := transaction.Sign(governorAddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - sender is governor address",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - multi address in sender feature
		func() *test {
			_, addr1, addr1Keys := tpkg.RandEd25519Identity()
			_, addr2, addr2Keys := tpkg.RandEd25519Identity()

			multiAddr := axongo.MultiAddress{
				Addresses: axongo.AddressesWithWeight{
					{
						Address: addr1,
						Weight:  5,
					},
					{
						Address: addr2,
						Weight:  10,
					},
					{
						Address: tpkg.RandNFTAddress(),
						Weight:  1,
					},
				},
				Threshold: 12,
			}

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: &multiAddr},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.BasicOutput{
						Amount: 100,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: axongo.BasicOutputFeatures{
							&axongo.SenderFeature{Address: &multiAddr},
						},
					},
				},
			}

			sigs, err := transaction.Sign(addr1Keys, addr2Keys)
			require.NoError(t, err)

			return &test{
				name: "ok - multi address in sender feature",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.MultiUnlock{
							Unlocks: axongo.Unlocks{
								&axongo.SignatureUnlock{Signature: sigs[0]},
								&axongo.SignatureUnlock{Signature: sigs[1]},
								&axongo.EmptyUnlock{},
							},
						},
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - restricted multi address in sender and issuer feature
		func() *test {
			_, addr1, addr1Keys := tpkg.RandEd25519Identity()
			_, addr2, addr2Keys := tpkg.RandEd25519Identity()

			multiAddr := axongo.MultiAddress{
				Addresses: axongo.AddressesWithWeight{
					{
						Address: addr1,
						Weight:  5,
					},
					{
						Address: addr2,
						Weight:  10,
					},
					{
						Address: tpkg.RandAccountAddress(),
						Weight:  1,
					},
				},
				Threshold: 12,
			}

			restrictedAddr := axongo.RestrictedAddress{
				Address:             &multiAddr,
				AllowedCapabilities: axongo.AddressCapabilitiesBitMaskWithCapabilities(axongo.WithAddressCanReceiveMana(true)),
			}

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: &restrictedAddr},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.NFTOutput{
						Amount: 100,
						UnlockConditions: axongo.NFTOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: axongo.NFTOutputFeatures{
							// We can use the restricted address...
							&axongo.SenderFeature{Address: &restrictedAddr},
						},
						ImmutableFeatures: axongo.NFTOutputImmFeatures{
							// ...or the underlying address.
							&axongo.IssuerFeature{Address: &multiAddr},
						},
					},
				},
			}

			sigs, err := transaction.Sign(addr1Keys, addr2Keys)
			require.NoError(t, err)

			return &test{
				name: "ok - restricted multi address in sender and issuer feature",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.MultiUnlock{
							Unlocks: axongo.Unlocks{
								&axongo.SignatureUnlock{Signature: sigs[0]},
								&axongo.SignatureUnlock{Signature: sigs[1]},
								&axongo.EmptyUnlock{},
							},
						},
					},
				},
				wantErr: nil,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAndExecuteSignedTransaction(tt.tx, tt.resolvedInputs, vm.ExecFuncSenderUnlocked())
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestTxSemanticOutputsIssuer(t *testing.T) {
	type test struct {
		name           string
		vmParams       *vm.Params
		resolvedInputs vm.ResolvedInputs
		tx             *axongo.SignedTransaction
		wantErr        error
	}
	tests := []*test{
		// fail - issuer not unlocked due to governance transition
		func() *test {
			_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
			_, stateController, _ := tpkg.RandEd25519Identity()
			_, governor, governorAddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)
			anchorAddr := tpkg.RandAnchorAddress()
			anchorID := anchorAddr.AnchorID()

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.AnchorOutput{
					Amount:   100,
					AnchorID: anchorID,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: stateController},
						&axongo.GovernorAddressUnlockCondition{Address: governor},
					},
				},
				inputIDs[1]: &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.AnchorOutput{
						Amount:   100,
						AnchorID: anchorID,
						UnlockConditions: axongo.AnchorOutputUnlockConditions{
							&axongo.StateControllerAddressUnlockCondition{Address: stateController},
							&axongo.GovernorAddressUnlockCondition{Address: governor},
						},
					},
					&axongo.NFTOutput{
						Amount: 100,
						UnlockConditions: axongo.NFTOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr1},
						},
						ImmutableFeatures: axongo.NFTOutputImmFeatures{
							&axongo.IssuerFeature{Address: anchorAddr},
						},
					},
				},
			}
			sigs, err := transaction.Sign(governorAddrKeys, addr1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - issuer not unlocked due to governance transition",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
						&axongo.SignatureUnlock{Signature: sigs[1]},
					},
				},
				wantErr: axongo.ErrIssuerFeatureNotUnlocked,
			}
		}(),

		// ok - issuer unlocked with state transition
		func() *test {
			_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
			_, stateController, stateControllerAddrKeys := tpkg.RandEd25519Identity()
			_, governor, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)
			anchorAddr := tpkg.RandAnchorAddress()
			anchorID := anchorAddr.AnchorID()
			currentStateIndex := uint32(1)

			nftAddr := tpkg.RandNFTAddress()

			inputs := vm.InputSet{
				// possible issuers: anchorAddr, stateController, nftAddr, addr1
				inputIDs[0]: &axongo.AnchorOutput{
					Amount:     100,
					AnchorID:   anchorID,
					StateIndex: currentStateIndex,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: stateController},
						&axongo.GovernorAddressUnlockCondition{Address: governor},
					},
				},
				inputIDs[1]: &axongo.NFTOutput{
					Amount: 900,
					NFTID:  nftAddr.NFTID(),
					UnlockConditions: axongo.NFTOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: axongo.TxEssenceOutputs{
					// transitioned anchor + nft
					&axongo.AnchorOutput{
						Amount:     100,
						AnchorID:   anchorID,
						StateIndex: currentStateIndex + 1,
						UnlockConditions: axongo.AnchorOutputUnlockConditions{
							&axongo.StateControllerAddressUnlockCondition{Address: stateController},
							&axongo.GovernorAddressUnlockCondition{Address: governor},
						},
					},
					&axongo.NFTOutput{
						Amount: 100,
						NFTID:  nftAddr.NFTID(),
						UnlockConditions: axongo.NFTOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr1},
						},
					},
					// issuer is anchorAddr
					&axongo.NFTOutput{
						Amount: 100,
						UnlockConditions: axongo.NFTOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr1},
						},
						ImmutableFeatures: axongo.NFTOutputImmFeatures{
							&axongo.IssuerFeature{Address: anchorAddr},
						},
					},
					&axongo.AnchorOutput{
						Amount: 100,
						UnlockConditions: axongo.AnchorOutputUnlockConditions{
							&axongo.StateControllerAddressUnlockCondition{Address: stateController},
							&axongo.GovernorAddressUnlockCondition{Address: governor},
						},
						ImmutableFeatures: axongo.AnchorOutputImmFeatures{
							&axongo.IssuerFeature{Address: anchorAddr},
						},
					},
					// issuer is stateController
					&axongo.NFTOutput{
						Amount: 100,
						UnlockConditions: axongo.NFTOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr1},
						},
						ImmutableFeatures: axongo.NFTOutputImmFeatures{
							&axongo.IssuerFeature{Address: stateController},
						},
					},
					&axongo.AnchorOutput{
						Amount: 100,
						UnlockConditions: axongo.AnchorOutputUnlockConditions{
							&axongo.StateControllerAddressUnlockCondition{Address: stateController},
							&axongo.GovernorAddressUnlockCondition{Address: governor},
						},
						ImmutableFeatures: axongo.AnchorOutputImmFeatures{
							&axongo.IssuerFeature{Address: stateController},
						},
					},
					// issuer is nftAddr
					&axongo.NFTOutput{
						Amount: 100,
						UnlockConditions: axongo.NFTOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr1},
						},
						ImmutableFeatures: axongo.NFTOutputImmFeatures{
							&axongo.IssuerFeature{Address: nftAddr},
						},
					},
					&axongo.AnchorOutput{
						Amount: 100,
						UnlockConditions: axongo.AnchorOutputUnlockConditions{
							&axongo.StateControllerAddressUnlockCondition{Address: stateController},
							&axongo.GovernorAddressUnlockCondition{Address: governor},
						},
						ImmutableFeatures: axongo.AnchorOutputImmFeatures{
							&axongo.IssuerFeature{Address: nftAddr},
						},
					},
					// issuer is addr1
					&axongo.NFTOutput{
						Amount: 100,
						UnlockConditions: axongo.NFTOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr1},
						},
						ImmutableFeatures: axongo.NFTOutputImmFeatures{
							&axongo.IssuerFeature{Address: addr1},
						},
					},
					&axongo.AnchorOutput{
						Amount: 100,
						UnlockConditions: axongo.AnchorOutputUnlockConditions{
							&axongo.StateControllerAddressUnlockCondition{Address: stateController},
							&axongo.GovernorAddressUnlockCondition{Address: governor},
						},
						ImmutableFeatures: axongo.AnchorOutputImmFeatures{
							&axongo.IssuerFeature{Address: addr1},
						},
					},
				},
			}
			sigs, err := transaction.Sign(stateControllerAddrKeys, addr1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - issuer unlocked with state transition",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
						&axongo.SignatureUnlock{Signature: sigs[1]},
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - issuer is the governor
		func() *test {
			_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
			_, stateController, _ := tpkg.RandEd25519Identity()
			_, governor, governorAddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)
			anchorAddr := tpkg.RandAnchorAddress()
			anchorID := anchorAddr.AnchorID()

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.AnchorOutput{
					Amount:   100,
					AnchorID: anchorID,
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: stateController},
						&axongo.GovernorAddressUnlockCondition{Address: governor},
					},
				},
				inputIDs[1]: &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.AnchorOutput{
						Amount:   100,
						AnchorID: anchorID,
						UnlockConditions: axongo.AnchorOutputUnlockConditions{
							&axongo.StateControllerAddressUnlockCondition{Address: stateController},
							&axongo.GovernorAddressUnlockCondition{Address: governor},
						},
					},
					&axongo.NFTOutput{
						Amount: 100,
						UnlockConditions: axongo.NFTOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: addr1},
						},
						ImmutableFeatures: axongo.NFTOutputImmFeatures{
							&axongo.IssuerFeature{Address: governor},
						},
					},
				},
			}
			sigs, err := transaction.Sign(governorAddrKeys, addr1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - issuer is the governor",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
						&axongo.SignatureUnlock{Signature: sigs[1]},
					},
				},
				wantErr: nil,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAndExecuteSignedTransaction(tt.tx, tt.resolvedInputs)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestTxSemanticTimelocks(t *testing.T) {
	type test struct {
		name           string
		vmParams       *vm.Params
		resolvedInputs vm.ResolvedInputs
		tx             *axongo.SignedTransaction
		wantErr        error
	}
	tests := []*test{
		// ok
		func() *test {
			_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
						&axongo.TimelockUnlockCondition{
							Slot: 5,
						},
					},
				},
			}

			creationSlot := axongo.SlotIndex(10)
			transaction := &axongo.Transaction{API: testAPI, TransactionEssence: &axongo.TransactionEssence{Inputs: inputIDs.UTXOInputs(), CreationSlot: creationSlot}}
			sigs, err := transaction.Sign(addr1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
					CommitmentInput: &axongo.Commitment{
						Slot: creationSlot,
					},
				},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		// fail - timelock not expired
		func() *test {
			_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
						&axongo.TimelockUnlockCondition{
							Slot: 25,
						},
					},
				},
			}

			creationSlot := axongo.SlotIndex(10)
			transaction := &axongo.Transaction{API: testAPI, TransactionEssence: &axongo.TransactionEssence{Inputs: inputIDs.UTXOInputs(), CreationSlot: creationSlot}}
			sigs, err := transaction.Sign(addr1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - timelock not expired",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
					CommitmentInput: &axongo.Commitment{
						Slot: creationSlot,
					},
				},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: axongo.ErrTimelockNotExpired,
			}
		}(),

		// fail - no commitment input for timelock
		func() *test {
			_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 100,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
						&axongo.TimelockUnlockCondition{
							Slot: 1000,
						},
					},
				},
			}

			creationSlot := axongo.SlotIndex(1005)
			transaction := &axongo.Transaction{API: testAPI, TransactionEssence: &axongo.TransactionEssence{Inputs: inputIDs.UTXOInputs(), CreationSlot: creationSlot}}
			sigs, err := transaction.Sign(addr1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - no commitment input for timelock",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
				},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: axongo.ErrTimelockCommitmentInputMissing,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := novaVM.Execute(tt.tx.Transaction, tt.resolvedInputs, make(vm.UnlockedAddresses), vm.ExecFuncTimelocks())
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

// TODO: add some more failing test cases.
func TestTxSemanticMana(t *testing.T) {
	type test struct {
		name           string
		vmParams       *vm.Params
		resolvedInputs vm.ResolvedInputs
		tx             *axongo.SignedTransaction
		wantErr        error
	}
	tests := []*test{
		// ok - stored Mana only without allotment"
		func() *test {
			_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDsWithCreationSlot(10, 1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: OneIOTA,
					Mana:   axongo.MaxMana,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs:       inputIDs.UTXOInputs(),
					CreationSlot: 10 + 100*testProtoParams.ParamEpochDurationInSlots(),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.BasicOutput{
						Amount: OneIOTA,
						Mana: func() axongo.Mana {
							var creationSlot axongo.SlotIndex = 10
							targetSlot := 10 + 100*testProtoParams.ParamEpochDurationInSlots()

							input := inputs[inputIDs[0]]
							storageScoreStructure := axongo.NewStorageScoreStructure(testProtoParams.StorageScoreParameters())
							potentialMana, err := axongo.PotentialMana(testAPI.ManaDecayProvider(), storageScoreStructure, input, creationSlot, targetSlot)
							require.NoError(t, err)

							storedMana, err := testAPI.ManaDecayProvider().DecayManaBySlots(axongo.MaxMana, creationSlot, targetSlot)
							require.NoError(t, err)

							return potentialMana + storedMana
						}(),
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := transaction.Sign(addr1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - stored Mana only without allotment",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - stored and allotted
		func() *test {
			_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDsWithCreationSlot(10, 1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: OneIOTA,
					Mana:   axongo.MaxMana,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
					Allotments: axongo.Allotments{
						&axongo.Allotment{Mana: 50},
					},
					CreationSlot: 10 + 100*testProtoParams.ParamEpochDurationInSlots(),
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.BasicOutput{
						Amount: OneIOTA,
						Mana: func() axongo.Mana {
							var creationSlot axongo.SlotIndex = 10
							targetSlot := 10 + 100*testProtoParams.ParamEpochDurationInSlots()

							input := inputs[inputIDs[0]]
							storageScoreStructure := axongo.NewStorageScoreStructure(testProtoParams.StorageScoreParameters())
							potentialMana, err := axongo.PotentialMana(testAPI.ManaDecayProvider(), storageScoreStructure, input, creationSlot, targetSlot)
							require.NoError(t, err)

							storedMana, err := testAPI.ManaDecayProvider().DecayManaBySlots(axongo.MaxMana, creationSlot, targetSlot)
							require.NoError(t, err)

							// generated mana + decay - allotment
							return potentialMana + storedMana - 50
						}(),
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := transaction.Sign(addr1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - stored and allotted",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		// fail - input created after tx
		func() *test {
			_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDsWithCreationSlot(20, 1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 5,
					Mana:   10,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs:       inputIDs.UTXOInputs(),
					CreationSlot: 15,
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.BasicOutput{
						Amount: 5,
						Mana:   35,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := transaction.Sign(addr1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - input created after tx",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: axongo.ErrInputCreationAfterTxCreation,
			}
		}(),

		// ok - input created in same slot as tx
		func() *test {
			_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDsWithCreationSlot(15, 1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 5,
					Mana:   10,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs:       inputIDs.UTXOInputs(),
					CreationSlot: 15,
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.BasicOutput{
						Amount: 5,
						Mana:   10,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := transaction.Sign(addr1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - input created in same slot as tx",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		// fail - mana overflow on the input side sum
		func() *test {
			_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDsWithCreationSlot(15, 2)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 5,
					Mana:   axongo.MaxMana,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
				inputIDs[1]: &axongo.BasicOutput{
					Amount: 5,
					Mana:   10,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs:       inputIDs.UTXOInputs(),
					CreationSlot: 15,
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.BasicOutput{
						Amount: 5,
						Mana:   9,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := transaction.Sign(addr1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - mana overflow on the input side sum",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
						&axongo.ReferenceUnlock{Reference: 0},
					},
				},
				wantErr: axongo.ErrManaOverflow,
			}
		}(),

		// fail - mana overflow on the output side sum
		func() *test {
			_, addr1, addr1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDsWithCreationSlot(15, 1)

			inputs := vm.InputSet{
				inputIDs[0]: &axongo.BasicOutput{
					Amount: 5,
					Mana:   10,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: addr1},
					},
				},
			}

			transaction := &axongo.Transaction{
				API: testAPI,
				TransactionEssence: &axongo.TransactionEssence{
					Inputs:       inputIDs.UTXOInputs(),
					CreationSlot: 15,
				},
				Outputs: axongo.TxEssenceOutputs{
					&axongo.BasicOutput{
						Amount: 5,
						Mana:   1,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
					&axongo.BasicOutput{
						Amount: 5,
						Mana:   axongo.MaxMana,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := transaction.Sign(addr1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - mana overflow on the output side sum",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: axongo.ErrManaOverflow,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAndExecuteSignedTransaction(tt.tx, tt.resolvedInputs, vm.ExecFuncBalancedMana())
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestManaRewardsClaimingStaking(t *testing.T) {
	_, addr, addrAddrKeys := tpkg.RandEd25519Identity()
	accountAddr := tpkg.RandAccountAddress()
	accountID := accountAddr.AccountID()

	var manaRewardAmount axongo.Mana = 200
	currentEpoch := axongo.EpochIndex(20)
	currentSlot := testAPI.TimeProvider().EpochStart(currentEpoch)

	blockIssuerFeature := &axongo.BlockIssuerFeature{
		BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
		ExpirySlot:      currentSlot + 500,
	}

	var creationSlot axongo.SlotIndex = 100
	balance := OneIOTA * 10

	inputIDs := tpkg.RandOutputIDsWithCreationSlot(creationSlot, 1)
	inputs := vm.InputSet{
		inputIDs[0]: &axongo.AccountOutput{
			Amount:         balance,
			AccountID:      accountID,
			Mana:           0,
			FoundryCounter: 0,
			UnlockConditions: axongo.AccountOutputUnlockConditions{
				&axongo.AddressUnlockCondition{Address: addr},
			},
			Features: axongo.AccountOutputFeatures{
				blockIssuerFeature,
				&axongo.StakingFeature{
					StakedAmount: 100,
					FixedCost:    50,
					StartEpoch:   currentEpoch - 10,
					EndEpoch:     currentEpoch - 1,
				},
			},
		},
	}

	inputMinDeposit := lo.PanicOnErr(testAPI.StorageScoreStructure().MinDeposit(inputs[inputIDs[0]]))

	transaction := &axongo.Transaction{
		API: testAPI,
		TransactionEssence: &axongo.TransactionEssence{
			Inputs:       inputIDs.UTXOInputs(),
			CreationSlot: currentSlot,
		},
		Outputs: axongo.TxEssenceOutputs{
			&axongo.AccountOutput{
				Amount:         OneIOTA * 5,
				Mana:           lo.PanicOnErr(testAPI.ManaDecayProvider().GenerateManaAndDecayBySlots(balance-inputMinDeposit, creationSlot, currentSlot)),
				AccountID:      accountID,
				FoundryCounter: 0,
				UnlockConditions: axongo.AccountOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: addr},
				},
				Features: axongo.AccountOutputFeatures{
					blockIssuerFeature,
				},
			},
			&axongo.BasicOutput{
				Amount: OneIOTA * 5,
				Mana:   manaRewardAmount,
				UnlockConditions: axongo.BasicOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: accountAddr},
				},
				Features: nil,
			},
		},
	}

	sigs, err := transaction.Sign(addrAddrKeys)
	require.NoError(t, err)

	tx := &axongo.SignedTransaction{
		API:         testAPI,
		Transaction: transaction,
		Unlocks: axongo.Unlocks{
			&axongo.SignatureUnlock{Signature: sigs[0]},
		},
	}

	resolvedInputs := vm.ResolvedInputs{
		InputSet: inputs,
		RewardsInputSet: map[axongo.ChainID]axongo.Mana{
			accountID: manaRewardAmount,
		},
		CommitmentInput: &axongo.Commitment{
			Slot: currentSlot,
		},
		BlockIssuanceCreditInputSet: vm.BlockIssuanceCreditInputSet{
			accountID: 1000,
		},
	}
	require.NoError(t, validateAndExecuteSignedTransaction(tx, resolvedInputs))
}

func TestManaRewardsClaimingDelegation(t *testing.T) {
	_, addr, addrAddrKeys := tpkg.RandEd25519Identity()

	const manaRewardAmount axongo.Mana = 200
	currentSlot := 20 * testProtoParams.ParamEpochDurationInSlots()
	currentEpoch := testAPI.TimeProvider().EpochFromSlot(currentSlot)

	inputIDs := tpkg.RandOutputIDs(1)
	inputs := vm.InputSet{
		inputIDs[0]: &axongo.DelegationOutput{
			Amount:           OneIOTA * 10,
			DelegatedAmount:  OneIOTA * 10,
			DelegationID:     axongo.EmptyDelegationID(),
			ValidatorAddress: &axongo.AccountAddress{},
			StartEpoch:       currentEpoch,
			EndEpoch:         currentEpoch + 5,
			UnlockConditions: axongo.DelegationOutputUnlockConditions{
				&axongo.AddressUnlockCondition{Address: addr},
			},
		},
	}
	delegationID := axongo.DelegationIDFromOutputID(inputIDs[0])

	transaction := &axongo.Transaction{
		API: testAPI,
		TransactionEssence: &axongo.TransactionEssence{
			Inputs:       inputIDs.UTXOInputs(),
			CreationSlot: currentSlot,
			Capabilities: axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanDoAnything()),
		},
		Outputs: axongo.TxEssenceOutputs{
			&axongo.BasicOutput{
				Amount: OneIOTA * 10,
				Mana:   manaRewardAmount,
				UnlockConditions: axongo.BasicOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: addr},
				},
				Features: nil,
			},
		},
	}

	sigs, err := transaction.Sign(addrAddrKeys)
	require.NoError(t, err)

	tx := &axongo.SignedTransaction{
		API:         testAPI,
		Transaction: transaction,
		Unlocks: axongo.Unlocks{
			&axongo.SignatureUnlock{Signature: sigs[0]},
		},
	}

	resolvedInputs := vm.ResolvedInputs{
		InputSet: inputs,
		RewardsInputSet: map[axongo.ChainID]axongo.Mana{
			delegationID: manaRewardAmount,
		},
	}
	require.NoError(t, validateAndExecuteSignedTransaction(tx, resolvedInputs))
}

func TestTxSyntacticAddressRestrictions(t *testing.T) {
	type testParameters struct {
		name    string
		address axongo.Address
		wantErr error
	}
	type test struct {
		createTestOutput     func(address axongo.Address) axongo.Output
		createTestParameters []func() testParameters
	}

	_, addr, addrAddrKeys := tpkg.RandEd25519Identity()
	randAddr := tpkg.RandEd25519Address()

	tests := []*test{
		{
			createTestOutput: func(address axongo.Address) axongo.Output {
				return &axongo.BasicOutput{
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: address},
					},
					Features: axongo.BasicOutputFeatures{
						tpkg.RandNativeTokenFeature(),
					},
				}
			},
			createTestParameters: []func() testParameters{
				func() testParameters {
					return testParameters{
						name:    "ok - Native Token Address in Output with Native Tokens",
						address: axongo.RestrictedAddressWithCapabilities(randAddr, axongo.WithAddressCanReceiveNativeTokens(true)),
						wantErr: nil,
					}
				},
				func() testParameters {
					return testParameters{
						name:    "fail - Non Native Token Address in Output with Native Tokens",
						address: axongo.RestrictedAddressWithCapabilities(randAddr),
						wantErr: axongo.ErrAddressCannotReceiveNativeTokens,
					}
				},
			},
		},
		{
			createTestOutput: func(address axongo.Address) axongo.Output {
				return &axongo.BasicOutput{
					Mana: axongo.Mana(4),
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: address},
					},
				}
			},
			createTestParameters: []func() testParameters{
				func() testParameters {
					return testParameters{
						name:    "ok - Mana Address in Output with Mana",
						address: axongo.RestrictedAddressWithCapabilities(randAddr, axongo.WithAddressCanReceiveMana(true)),
						wantErr: nil,
					}
				},
				func() testParameters {
					return testParameters{
						name:    "fail - Non Mana Address in Output with Mana",
						address: axongo.RestrictedAddressWithCapabilities(randAddr),
						wantErr: axongo.ErrAddressCannotReceiveMana,
					}
				},
			},
		},
		{
			createTestOutput: func(address axongo.Address) axongo.Output {
				return &axongo.BasicOutput{
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: address},
						&axongo.TimelockUnlockCondition{
							Slot: 500,
						},
					},
				}
			},
			createTestParameters: []func() testParameters{
				func() testParameters {
					return testParameters{
						name:    "ok - Timelock Unlock Condition Address in Output with Timelock Unlock Condition",
						address: axongo.RestrictedAddressWithCapabilities(randAddr, axongo.WithAddressCanReceiveOutputsWithTimelockUnlockCondition(true)),
						wantErr: nil,
					}
				},
				func() testParameters {
					return testParameters{
						name:    "fail - Non Timelock Unlock Condition Address in Output with Timelock Unlock Condition",
						address: axongo.RestrictedAddressWithCapabilities(randAddr),
						wantErr: axongo.ErrAddressCannotReceiveTimelockUnlockCondition,
					}
				},
			},
		},
		{
			createTestOutput: func(address axongo.Address) axongo.Output {
				return &axongo.BasicOutput{
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: address},
						&axongo.ExpirationUnlockCondition{
							Slot:          500,
							ReturnAddress: addr,
						},
					},
				}
			},
			createTestParameters: []func() testParameters{
				func() testParameters {
					return testParameters{
						name:    "ok - Expiration Unlock Condition Address in Output with Expiration Unlock Condition",
						address: axongo.RestrictedAddressWithCapabilities(randAddr, axongo.WithAddressCanReceiveOutputsWithExpirationUnlockCondition(true)),
						wantErr: nil,
					}
				},
				func() testParameters {
					return testParameters{
						name:    "fail - Non Expiration Unlock Condition Address in Output with Expiration Unlock Condition",
						address: axongo.RestrictedAddressWithCapabilities(randAddr),
						wantErr: axongo.ErrAddressCannotReceiveExpirationUnlockCondition,
					}
				},
			},
		},
		{
			createTestOutput: func(address axongo.Address) axongo.Output {
				return &axongo.BasicOutput{
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: address},
						&axongo.StorageDepositReturnUnlockCondition{
							ReturnAddress: addr,
						},
					},
				}
			},
			createTestParameters: []func() testParameters{
				func() testParameters {
					return testParameters{
						name:    "ok - Storage Deposit Return Unlock Condition Address in Output with Storage Deposit Return Unlock Condition",
						address: axongo.RestrictedAddressWithCapabilities(randAddr, axongo.WithAddressCanReceiveOutputsWithStorageDepositReturnUnlockCondition(true)),
						wantErr: nil,
					}
				},
				func() testParameters {
					return testParameters{
						name:    "fail - Non Storage Deposit Return Unlock Condition Address in Output with Storage Deposit Return Unlock Condition",
						address: axongo.RestrictedAddressWithCapabilities(randAddr),
						wantErr: axongo.ErrAddressCannotReceiveStorageDepositReturnUnlockCondition,
					}
				},
			},
		},
		{
			createTestOutput: func(address axongo.Address) axongo.Output {
				return &axongo.AccountOutput{
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: address},
					},
				}
			},
			createTestParameters: []func() testParameters{
				func() testParameters {
					return testParameters{
						name:    "ok - Account Output Address in Account Output",
						address: axongo.RestrictedAddressWithCapabilities(randAddr, axongo.WithAddressCanReceiveAccountOutputs(true)),
						wantErr: nil,
					}
				},
				func() testParameters {
					return testParameters{
						name:    "fail - Non Account Output Address in Account Output",
						address: axongo.RestrictedAddressWithCapabilities(randAddr),
						wantErr: axongo.ErrAddressCannotReceiveAccountOutput,
					}
				},
			},
		},
		{
			createTestOutput: func(address axongo.Address) axongo.Output {
				return &axongo.AnchorOutput{
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.StateControllerAddressUnlockCondition{Address: address},
					},
				}
			},
			createTestParameters: []func() testParameters{
				func() testParameters {
					return testParameters{
						name:    "ok - Anchor Output Address in State Controller UC in Anchor Output",
						address: axongo.RestrictedAddressWithCapabilities(randAddr, axongo.WithAddressCanReceiveAnchorOutputs(true)),
						wantErr: nil,
					}
				},
				func() testParameters {
					return testParameters{
						name:    "fail - Non Anchor Output Address in State Controller UC in Anchor Output",
						address: axongo.RestrictedAddressWithCapabilities(randAddr),
						wantErr: axongo.ErrAddressCannotReceiveAnchorOutput,
					}
				},
			},
		},
		{
			createTestOutput: func(address axongo.Address) axongo.Output {
				return &axongo.AnchorOutput{
					UnlockConditions: axongo.AnchorOutputUnlockConditions{
						&axongo.GovernorAddressUnlockCondition{Address: address},
					},
				}
			},
			createTestParameters: []func() testParameters{
				func() testParameters {
					return testParameters{
						name:    "ok - Anchor Output Address in Governor UC in Anchor Output",
						address: axongo.RestrictedAddressWithCapabilities(randAddr, axongo.WithAddressCanReceiveAnchorOutputs(true)),
						wantErr: nil,
					}
				},
				func() testParameters {
					return testParameters{
						name:    "fail - Non Anchor Output Address in Governor UC in Anchor Output",
						address: axongo.RestrictedAddressWithCapabilities(randAddr),
						wantErr: axongo.ErrAddressCannotReceiveAnchorOutput,
					}
				},
			},
		},
		{
			createTestOutput: func(address axongo.Address) axongo.Output {
				return &axongo.NFTOutput{
					UnlockConditions: axongo.NFTOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: address},
					},
				}
			},
			createTestParameters: []func() testParameters{
				func() testParameters {
					return testParameters{
						name:    "ok - NFT Output Address in NFT Output",
						address: axongo.RestrictedAddressWithCapabilities(randAddr, axongo.WithAddressCanReceiveNFTOutputs(true)),
						wantErr: nil,
					}
				},
				func() testParameters {
					return testParameters{
						name:    "fail - Non NFT Output Address in NFT Output",
						address: axongo.RestrictedAddressWithCapabilities(randAddr),
						wantErr: axongo.ErrAddressCannotReceiveNFTOutput,
					}
				},
			},
		},
		{
			createTestOutput: func(address axongo.Address) axongo.Output {
				return &axongo.DelegationOutput{
					UnlockConditions: axongo.DelegationOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: address},
					},
					ValidatorAddress: &axongo.AccountAddress{},
				}
			},
			createTestParameters: []func() testParameters{
				func() testParameters {
					return testParameters{
						name:    "ok - Delegation Output Address in Delegation Output",
						address: axongo.RestrictedAddressWithCapabilities(randAddr, axongo.WithAddressCanReceiveDelegationOutputs(true)),
						wantErr: nil,
					}
				},
				func() testParameters {
					return testParameters{
						name:    "fail - Non Delegation Output Address in Delegation Output",
						address: axongo.RestrictedAddressWithCapabilities(randAddr),
						wantErr: axongo.ErrAddressCannotReceiveDelegationOutput,
					}
				},
			},
		},
		{
			createTestOutput: func(address axongo.Address) axongo.Output {
				return &axongo.BasicOutput{
					Mana: 42,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						&axongo.ExpirationUnlockCondition{
							// only the return address is restricted here
							ReturnAddress: address,
						},
					},
				}
			},
			createTestParameters: []func() testParameters{
				func() testParameters {
					return testParameters{
						name:    "ok - Mana Return Address in Output with Mana",
						address: axongo.RestrictedAddressWithCapabilities(randAddr, axongo.WithAddressCanReceiveMana(true), axongo.WithAddressCanReceiveOutputsWithExpirationUnlockCondition(true)),
						wantErr: nil,
					}
				},
				func() testParameters {
					return testParameters{
						name:    "fail - Non Mana Return Address in Output with Mana",
						address: axongo.RestrictedAddressWithCapabilities(randAddr),
						wantErr: axongo.ErrAddressCannotReceiveMana,
					}
				},
			},
		},
	}

	makeTransaction := func(output axongo.Output) (vm.InputSet, axongo.Signature, *axongo.Transaction) {
		inputIDs := tpkg.RandOutputIDsWithCreationSlot(10, 1)

		inputs := vm.InputSet{
			inputIDs[0]: &axongo.BasicOutput{
				Amount: axongo.BaseToken(1_000_000),
				UnlockConditions: axongo.BasicOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: addr},
				},
			},
		}

		transaction := &axongo.Transaction{
			API: testAPI,
			TransactionEssence: &axongo.TransactionEssence{
				NetworkID:    testAPI.ProtocolParameters().NetworkID(),
				Inputs:       inputIDs.UTXOInputs(),
				CreationSlot: 10,
			},
			Outputs: axongo.TxEssenceOutputs{
				output,
			},
		}
		sigs, err := transaction.Sign(addrAddrKeys)
		require.NoError(t, err)

		return inputs, sigs[0], transaction
	}

	for _, tt := range tests {
		for _, makeTestInput := range tt.createTestParameters {
			testInput := makeTestInput()
			t.Run(testInput.name, func(t *testing.T) {
				testOutput := tt.createTestOutput(testInput.address)

				_, sig, transaction := makeTransaction(testOutput)

				tx := &axongo.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: axongo.Unlocks{
						&axongo.SignatureUnlock{Signature: sig},
					},
				}

				addressRestrictionFunc := axongo.OutputsSyntacticalAddressRestrictions()

				for index, output := range tx.Transaction.Outputs {
					err := addressRestrictionFunc(index, output)

					if testInput.wantErr != nil {
						require.ErrorIs(t, err, testInput.wantErr)
						return
					}

					require.NoError(t, err)
				}
			})
		}
	}
}

func TestTxSemanticImplicitAccountCreationAndTransition(t *testing.T) {
	type TestInput struct {
		inputID      axongo.OutputID
		input        axongo.Output
		unlockTarget axongo.Address
	}

	type test struct {
		name                    string
		inputs                  []TestInput
		keys                    []axongo.AddressKeys
		resolvedCommitmentInput *axongo.Commitment
		resolvedBICInputSet     vm.BlockIssuanceCreditInputSet
		outputs                 []axongo.Output
		wantErr                 error
	}

	_, edAddr, edAddrAddrKeys := tpkg.RandEd25519Identity()
	_, implicitAccountAddr, implicitAccountAddrAddrKeys := tpkg.RandImplicitAccountIdentity()
	exampleAmount := axongo.BaseToken(1_000_000)
	exampleMana := axongo.Mana(10_000_000)
	exampleNativeTokenFeature := tpkg.RandNativeTokenFeature()
	outputID1 := tpkg.RandOutputID(0)
	outputID2 := tpkg.RandOutputID(1)
	accountID1 := axongo.AccountIDFromOutputID(outputID1)
	accountID2 := axongo.AccountIDFromOutputID(outputID2)
	currentSlot := axongo.SlotIndex(10)
	commitmentSlot := currentSlot - testAPI.ProtocolParameters().MaxCommittableAge()

	dummyImplicitAccount := &axongo.BasicOutput{
		Amount: 0,
		UnlockConditions: axongo.BasicOutputUnlockConditions{
			&axongo.AddressUnlockCondition{Address: implicitAccountAddr},
		},
	}
	exampleMetadataFeature := &axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": tpkg.RandBytes(40)}}
	exampleMetadataFeatureStorageDeposit := axongo.BaseToken(exampleMetadataFeature.Size()*int(testAPI.StorageScoreStructure().FactorData())) * testAPI.StorageScoreStructure().StorageCost()

	storageScore := dummyImplicitAccount.StorageScore(testAPI.StorageScoreStructure(), nil)
	minAmountImplicitAccount := testAPI.StorageScoreStructure().StorageCost() * axongo.BaseToken(storageScore)

	exampleInputs := []TestInput{
		{
			inputID: outputID1,
			input: &axongo.BasicOutput{
				Amount: exampleAmount,
				UnlockConditions: axongo.BasicOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: edAddr},
				},
				Features: axongo.BasicOutputFeatures{
					exampleNativeTokenFeature,
				},
			},
			unlockTarget: edAddr,
		},
	}

	tests := []*test{
		{
			name:   "ok - implicit account creation",
			inputs: exampleInputs,
			outputs: []axongo.Output{
				&axongo.BasicOutput{
					Amount: exampleAmount,
					Mana:   0,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: implicitAccountAddr},
					},
				},
			},
			keys:    []axongo.AddressKeys{edAddrAddrKeys},
			wantErr: nil,
		},
		{
			name:   "fail - implicit account contains timelock unlock conditions",
			inputs: exampleInputs,
			outputs: []axongo.Output{
				&axongo.BasicOutput{
					Amount: exampleAmount,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: implicitAccountAddr},
						&axongo.TimelockUnlockCondition{Slot: 500},
					},
				},
			},
			keys:    []axongo.AddressKeys{edAddrAddrKeys},
			wantErr: axongo.ErrAddressCannotReceiveTimelockUnlockCondition,
		},
		{
			name:   "fail - implicit account contains expiration unlock conditions",
			inputs: exampleInputs,
			outputs: []axongo.Output{
				&axongo.BasicOutput{
					Amount: exampleAmount,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: implicitAccountAddr},
						&axongo.ExpirationUnlockCondition{
							// The implicit account creation address should disallow this expiration UC.
							ReturnAddress: tpkg.RandEd25519Address(),
							Slot:          500,
						},
					},
				},
			},
			keys:    []axongo.AddressKeys{edAddrAddrKeys},
			wantErr: axongo.ErrAddressCannotReceiveExpirationUnlockCondition,
		},
		{
			name:   "fail - implicit account contains storage deposit return unlock conditions",
			inputs: exampleInputs,
			outputs: []axongo.Output{
				&axongo.BasicOutput{
					Amount: exampleAmount,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: implicitAccountAddr},
						&axongo.StorageDepositReturnUnlockCondition{
							// The implicit account creation address should disallow this SDRUC.
							ReturnAddress: tpkg.RandEd25519Address(),
							Amount:        20_000,
						},
					},
				},
			},
			keys:    []axongo.AddressKeys{edAddrAddrKeys},
			wantErr: axongo.ErrAddressCannotReceiveStorageDepositReturnUnlockCondition,
		},
		{
			name:   "ok - implicit account contains features",
			inputs: exampleInputs,
			outputs: []axongo.Output{
				&axongo.BasicOutput{
					Amount: exampleAmount,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: implicitAccountAddr},
					},
					Features: axongo.BasicOutputFeatures{
						&axongo.SenderFeature{
							Address: edAddrAddrKeys.Address,
						},
						&axongo.MetadataFeature{Entries: axongo.MetadataFeatureEntries{"data": tpkg.RandBytes(40)}},
						&axongo.TagFeature{
							Tag: tpkg.RandBytes(12),
						},
						&axongo.NativeTokenFeature{
							ID:     exampleNativeTokenFeature.ID,
							Amount: exampleNativeTokenFeature.Amount,
						},
					},
				},
			},
			keys:    []axongo.AddressKeys{edAddrAddrKeys},
			wantErr: nil,
		},
		{
			name: "ok - implicit account transitioned to account with block issuer feature",
			inputs: []TestInput{
				{
					inputID: outputID1,
					input: &axongo.BasicOutput{
						Amount: exampleAmount,
						Mana:   exampleMana,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: implicitAccountAddr},
						},
					},
					unlockTarget: implicitAccountAddr,
				},
			},
			resolvedBICInputSet: vm.BlockIssuanceCreditInputSet{
				accountID1: axongo.BlockIssuanceCredits(0),
			},
			resolvedCommitmentInput: &axongo.Commitment{
				Slot: commitmentSlot,
			},
			outputs: []axongo.Output{
				&axongo.AccountOutput{
					Amount:    exampleAmount,
					Mana:      exampleMana,
					AccountID: accountID1,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{
							Address: edAddr,
						},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							ExpirySlot: axongo.MaxSlotIndex,
							BlockIssuerKeys: axongo.NewBlockIssuerKeys(
								axongo.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
								axongo.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
								axongo.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
							),
						},
					},
				},
			},
			keys:    []axongo.AddressKeys{implicitAccountAddrAddrKeys},
			wantErr: nil,
		},
		{
			name: "ok - implicit account with native tokens transitioned to account with block issuer feature",
			inputs: []TestInput{
				{
					inputID: outputID1,
					input: &axongo.BasicOutput{
						Amount: exampleAmount,
						Mana:   exampleMana,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: implicitAccountAddr},
						},
						Features: axongo.BasicOutputFeatures{
							exampleNativeTokenFeature,
						},
					},
					unlockTarget: implicitAccountAddr,
				},
			},
			resolvedBICInputSet: vm.BlockIssuanceCreditInputSet{
				accountID1: axongo.BlockIssuanceCredits(0),
			},
			resolvedCommitmentInput: &axongo.Commitment{
				Slot: commitmentSlot,
			},
			outputs: []axongo.Output{
				// a basic output will hold the native tokens
				&axongo.BasicOutput{
					Amount: 21200,
					Mana:   0,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: edAddr},
					},
					Features: axongo.BasicOutputFeatures{
						exampleNativeTokenFeature,
					},
				},
				&axongo.AccountOutput{
					Amount:    exampleAmount - 21200,
					Mana:      exampleMana,
					AccountID: accountID1,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{
							Address: edAddr,
						},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							ExpirySlot: axongo.MaxSlotIndex,
							BlockIssuerKeys: axongo.NewBlockIssuerKeys(
								axongo.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
								axongo.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
								axongo.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
							),
						},
					},
				},
			},
			keys:    []axongo.AddressKeys{implicitAccountAddrAddrKeys},
			wantErr: nil,
		},
		{
			name: "fail - implicit account transitioned to account without block issuer feature",
			inputs: []TestInput{
				{
					inputID: outputID1,
					input: &axongo.BasicOutput{
						Amount: exampleAmount,
						Mana:   exampleMana,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: implicitAccountAddr},
						},
					},
					unlockTarget: implicitAccountAddr,
				},
			},
			resolvedBICInputSet: vm.BlockIssuanceCreditInputSet{
				accountID1: axongo.BlockIssuanceCredits(0),
			},
			resolvedCommitmentInput: &axongo.Commitment{
				Slot: commitmentSlot,
			},
			outputs: []axongo.Output{
				&axongo.AccountOutput{
					Amount:    exampleAmount,
					Mana:      exampleMana,
					AccountID: accountID1,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{
							Address: edAddr,
						},
					},
					Features: axongo.AccountOutputFeatures{},
				},
			},
			keys:    []axongo.AddressKeys{implicitAccountAddrAddrKeys},
			wantErr: axongo.ErrBlockIssuerNotExpired,
		},
		{
			name: "fail - attempt to destroy implicit account",
			inputs: []TestInput{
				{
					inputID: outputID1,
					input: &axongo.BasicOutput{
						Amount: exampleAmount,
						Mana:   exampleMana,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: implicitAccountAddr},
						},
					},
					unlockTarget: implicitAccountAddr,
				},
			},
			resolvedBICInputSet: vm.BlockIssuanceCreditInputSet{
				accountID1: axongo.BlockIssuanceCredits(0),
			},
			resolvedCommitmentInput: &axongo.Commitment{
				Slot: commitmentSlot,
			},
			outputs: []axongo.Output{
				&axongo.BasicOutput{
					Amount: exampleAmount,
					Mana:   exampleMana,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: edAddr},
					},
				},
			},
			keys:    []axongo.AddressKeys{implicitAccountAddrAddrKeys},
			wantErr: axongo.ErrImplicitAccountDestructionDisallowed,
		},
		{
			name: "ok - implicit account with OffsetImplicitAccountCreationAddress can be transitioned",
			inputs: []TestInput{
				{
					inputID: outputID1,
					input: &axongo.BasicOutput{
						Amount: minAmountImplicitAccount,
						Mana:   0,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: implicitAccountAddr},
						},
					},
					unlockTarget: implicitAccountAddr,
				},
			},
			resolvedBICInputSet: vm.BlockIssuanceCreditInputSet{
				accountID1: axongo.BlockIssuanceCredits(0),
			},
			resolvedCommitmentInput: &axongo.Commitment{
				Slot: commitmentSlot,
			},
			outputs: []axongo.Output{
				&axongo.AccountOutput{
					Amount:    minAmountImplicitAccount,
					Mana:      0,
					AccountID: accountID1,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{
							Address: edAddr,
						},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							ExpirySlot: axongo.MaxSlotIndex,
							BlockIssuerKeys: axongo.NewBlockIssuerKeys(
								axongo.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
							),
						},
					},
				},
			},
			keys:    []axongo.AddressKeys{implicitAccountAddrAddrKeys},
			wantErr: nil,
		},
		{
			name: "ok - implicit account with minimal amount and metadata feat can be transitioned",
			inputs: []TestInput{
				{
					inputID: outputID1,
					input: &axongo.BasicOutput{
						Amount: minAmountImplicitAccount + exampleMetadataFeatureStorageDeposit,
						Mana:   0,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: implicitAccountAddr},
						},
						Features: axongo.BasicOutputFeatures{
							exampleMetadataFeature,
						},
					},
					unlockTarget: implicitAccountAddr,
				},
			},
			resolvedBICInputSet: vm.BlockIssuanceCreditInputSet{
				accountID1: axongo.BlockIssuanceCredits(0),
			},
			resolvedCommitmentInput: &axongo.Commitment{
				Slot: commitmentSlot,
			},
			outputs: []axongo.Output{
				&axongo.AccountOutput{
					Amount:    minAmountImplicitAccount + exampleMetadataFeatureStorageDeposit,
					Mana:      0,
					AccountID: accountID1,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{
							Address: edAddr,
						},
					},
					Features: axongo.AccountOutputFeatures{
						exampleMetadataFeature,
						&axongo.BlockIssuerFeature{
							ExpirySlot: axongo.MaxSlotIndex,
							BlockIssuerKeys: axongo.NewBlockIssuerKeys(
								axongo.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
							),
						},
					},
				},
			},
			keys:    []axongo.AddressKeys{implicitAccountAddrAddrKeys},
			wantErr: nil,
		},
		{
			name: "ok - implicit account conversion transaction can contain other non-implicit-account outputs",
			inputs: []TestInput{
				{
					inputID: outputID1,
					input: &axongo.BasicOutput{
						Amount: minAmountImplicitAccount,
						Mana:   0,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: implicitAccountAddr},
						},
					},
					unlockTarget: implicitAccountAddr,
				},
				{
					inputID: tpkg.RandOutputID(1),
					input: &axongo.BasicOutput{
						Amount: exampleAmount,
						Mana:   0,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: edAddr},
						},
					},
					unlockTarget: edAddr,
				},
			},
			resolvedBICInputSet: vm.BlockIssuanceCreditInputSet{
				accountID1: axongo.BlockIssuanceCredits(0),
			},
			resolvedCommitmentInput: &axongo.Commitment{
				Slot: commitmentSlot,
			},
			outputs: []axongo.Output{
				&axongo.AccountOutput{
					// Fund new account with additional base tokens from another output.
					Amount:    minAmountImplicitAccount + exampleAmount,
					Mana:      0,
					AccountID: accountID1,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{
							Address: edAddr,
						},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							ExpirySlot: axongo.MaxSlotIndex,
							BlockIssuerKeys: axongo.NewBlockIssuerKeys(
								axongo.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
							),
						},
					},
				},
			},
			keys:    []axongo.AddressKeys{implicitAccountAddrAddrKeys, edAddrAddrKeys},
			wantErr: nil,
		},
		{
			name: "fail - transaction contains more than one implicit account on the input side",
			inputs: []TestInput{
				{
					inputID: outputID1,
					input: &axongo.BasicOutput{
						Amount: minAmountImplicitAccount,
						Mana:   0,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: implicitAccountAddr},
						},
					},
					unlockTarget: implicitAccountAddr,
				},
				{
					inputID: outputID2,
					input: &axongo.BasicOutput{
						Amount: exampleAmount,
						Mana:   0,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: implicitAccountAddr},
						},
					},
					unlockTarget: implicitAccountAddr,
				},
			},
			resolvedBICInputSet: vm.BlockIssuanceCreditInputSet{
				accountID1: axongo.BlockIssuanceCredits(0),
				accountID2: axongo.BlockIssuanceCredits(0),
			},
			resolvedCommitmentInput: &axongo.Commitment{
				Slot: commitmentSlot,
			},
			outputs: []axongo.Output{
				&axongo.AccountOutput{
					Amount:    minAmountImplicitAccount,
					Mana:      0,
					AccountID: accountID1,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{
							Address: edAddr,
						},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							ExpirySlot: axongo.MaxSlotIndex,
							BlockIssuerKeys: axongo.NewBlockIssuerKeys(
								axongo.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
							),
						},
					},
				},
				&axongo.AccountOutput{
					Amount:    exampleAmount,
					Mana:      0,
					AccountID: accountID2,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{
							Address: edAddr,
						},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							ExpirySlot: axongo.MaxSlotIndex,
							BlockIssuerKeys: axongo.NewBlockIssuerKeys(
								axongo.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
							),
						},
					},
				},
			},
			keys:    []axongo.AddressKeys{implicitAccountAddrAddrKeys, edAddrAddrKeys},
			wantErr: axongo.ErrMultipleImplicitAccountCreationAddresses,
		},
		{
			name: "fail - transaction moves mana off an implicit account",
			inputs: []TestInput{
				{
					inputID: outputID1,
					input: &axongo.BasicOutput{
						Amount: exampleAmount,
						Mana:   exampleMana,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: implicitAccountAddr},
						},
					},
					unlockTarget: implicitAccountAddr,
				},
			},
			resolvedBICInputSet: vm.BlockIssuanceCreditInputSet{
				accountID1: axongo.BlockIssuanceCredits(0),
			},
			resolvedCommitmentInput: &axongo.Commitment{
				Slot: commitmentSlot,
			},
			outputs: []axongo.Output{
				&axongo.AccountOutput{
					Amount:    minAmountImplicitAccount,
					Mana:      exampleMana / 2,
					AccountID: accountID1,
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{
							Address: edAddr,
						},
					},
					Features: axongo.AccountOutputFeatures{
						&axongo.BlockIssuerFeature{
							ExpirySlot: axongo.MaxSlotIndex,
							BlockIssuerKeys: axongo.NewBlockIssuerKeys(
								axongo.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
							),
						},
					},
				},
				&axongo.BasicOutput{
					Amount: exampleAmount - minAmountImplicitAccount,
					Mana:   exampleMana / 2,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{
							Address: edAddr,
						},
					},
				},
			},
			keys:    []axongo.AddressKeys{implicitAccountAddrAddrKeys},
			wantErr: axongo.ErrManaMovedOffBlockIssuerAccount,
		},
	}

	for idx, tt := range tests {
		resolvedInputs := vm.ResolvedInputs{
			InputSet: vm.InputSet{},
		}

		txBuilder := builder.NewTransactionBuilder(testAPI, axongo.NewInMemoryAddressSigner(tt.keys...))
		txBuilder.WithTransactionCapabilities(
			axongo.TransactionCapabilitiesBitMaskWithCapabilities(axongo.WithTransactionCanBurnNativeTokens(true)),
		)

		// Add the BIC and Commitment Inputs to the TX builder since they are required syntactically.
		// Note that this has no effect on the actual test.
		for accountID := range tests[idx].resolvedBICInputSet {
			txBuilder.AddBlockIssuanceCreditInput(&axongo.BlockIssuanceCreditInput{
				AccountID: accountID,
			})
		}
		if tests[idx].resolvedCommitmentInput != nil {
			txBuilder.AddCommitmentInput(&axongo.CommitmentInput{
				CommitmentID: tests[idx].resolvedCommitmentInput.MustID(),
			})
		}

		for _, input := range tests[idx].inputs {
			txBuilder.AddInput(&builder.TxInput{
				UnlockTarget: input.unlockTarget,
				InputID:      input.inputID,
				Input:        input.input,
			},
			)

			resolvedInputs.InputSet[input.inputID] = input.input
		}

		for _, output := range tests[idx].outputs {
			txBuilder.AddOutput(output)
		}
		tx := lo.PanicOnErr(txBuilder.Build())

		resolvedInputs.BlockIssuanceCreditInputSet = tests[idx].resolvedBICInputSet
		resolvedInputs.CommitmentInput = tests[idx].resolvedCommitmentInput

		t.Run(tt.name, func(t *testing.T) {
			var err error
			// Some constraints are implicitly tested as part of the address restrictions, which are syntactic checks.
			err = tx.Transaction.SyntacticallyValidate(tx.API)
			if err == nil {
				err = validateAndExecuteSignedTransaction(tx, resolvedInputs)
			}

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

// Ensure that the storage score offset for implicit accounts is the
// minimum required for a full block issuer account.
func TestTxSyntacticImplicitAccountMinDeposit(t *testing.T) {
	_, implicitAccountAddr, _ := tpkg.RandImplicitAccountIdentity()

	implicitAccount := &axongo.BasicOutput{
		Amount: 0,
		UnlockConditions: axongo.BasicOutputUnlockConditions{
			&axongo.AddressUnlockCondition{Address: implicitAccountAddr},
		},
	}
	storageScore := implicitAccount.StorageScore(testAPI.StorageScoreStructure(), nil)
	minAmount := testAPI.StorageScoreStructure().StorageCost() * axongo.BaseToken(storageScore)
	implicitAccount.Amount = minAmount
	depositValidationFunc := axongo.OutputsSyntacticalDepositAmount(testAPI.ProtocolParameters(), testAPI.StorageScoreStructure())
	require.NoError(t, depositValidationFunc(0, implicitAccount))

	convertedAccount := &axongo.AccountOutput{
		Amount: implicitAccount.Amount,
		UnlockConditions: axongo.AccountOutputUnlockConditions{
			&axongo.AddressUnlockCondition{
				Address: &axongo.Ed25519Address{},
			},
		},
		Features: axongo.AccountOutputFeatures{
			&axongo.BlockIssuerFeature{
				BlockIssuerKeys: axongo.BlockIssuerKeys{
					&axongo.Ed25519PublicKeyHashBlockIssuerKey{},
				},
			},
		},
	}

	require.NoError(t, depositValidationFunc(0, convertedAccount))
}

func validateAndExecuteSignedTransaction(tx *axongo.SignedTransaction, resolvedInputs vm.ResolvedInputs, execFunctions ...vm.ExecFunc) (err error) {
	unlockedAddrs, err := novaVM.ValidateUnlocks(tx, resolvedInputs)
	if err != nil {
		return err
	}

	return lo.Return2(novaVM.Execute(tx.Transaction, resolvedInputs, unlockedAddrs, execFunctions...))
}

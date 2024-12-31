//nolint:forcetypeassert
package builder_test

import (
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axonfibre/fibre.go/ierrors"
	axongo "github.com/axonfibre/axon.go/v4"
	"github.com/axonfibre/axon.go/v4/builder"
	"github.com/axonfibre/axon.go/v4/tpkg"
)

func TestTransactionBuilder(t *testing.T) {
	prvKey := tpkg.RandEd25519PrivateKey()
	pubKey := prvKey.Public().(ed25519.PublicKey)
	inputAddrEd25519 := axongo.Ed25519AddressFromPubKey(pubKey)
	inputAddrRestricted := axongo.RestrictedAddressWithCapabilities(inputAddrEd25519, axongo.WithAddressCanReceiveAnything())
	inputAddrImplicitAccountCreation := axongo.ImplicitAccountCreationAddressFromPubKey(pubKey)
	signer := axongo.NewInMemoryAddressSignerFromEd25519PrivateKeys(prvKey)

	output := &axongo.BasicOutput{
		Amount: 50,
		UnlockConditions: axongo.BasicOutputUnlockConditions{
			&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
		},
	}

	type test struct {
		name     string
		builder  *builder.TransactionBuilder
		buildErr error
	}

	tests := []*test{
		// ok - 1 input/output - Ed25519 address
		func() *test {
			inputUTXO1 := &axongo.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 0}
			input := tpkg.RandOutputOnAddress(axongo.OutputBasic, inputAddrEd25519)
			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, signer).
				AddInput(&builder.TxInput{UnlockTarget: inputAddrEd25519, InputID: inputUTXO1.OutputID(), Input: input}).
				AddOutput(output)

			return &test{
				name:    "ok - 1 input/output - Ed25519 address",
				builder: bdl,
			}
		}(),

		// ok - 1 input/output - Restricted address with underlying Ed25519 address
		func() *test {
			inputUTXO1 := &axongo.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 0}
			input := tpkg.RandOutputOnAddress(axongo.OutputBasic, inputAddrRestricted)
			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, signer).
				AddInput(&builder.TxInput{UnlockTarget: inputAddrRestricted, InputID: inputUTXO1.OutputID(), Input: input}).
				AddOutput(output)

			return &test{
				name:    "ok - 1 input/output - Restricted address with underlying Ed25519 address",
				builder: bdl,
			}
		}(),

		// ok - 1 input/output - Implicit account creation address
		func() *test {
			inputUTXO1 := &axongo.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 0}
			input := tpkg.RandOutputOnAddress(axongo.OutputBasic, inputAddrImplicitAccountCreation)
			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, signer).
				AddInput(&builder.TxInput{UnlockTarget: inputAddrImplicitAccountCreation, InputID: inputUTXO1.OutputID(), Input: input}).
				AddOutput(output)

			return &test{
				name:    "ok - 1 input/output - Implicit account creation address",
				builder: bdl,
			}
		}(),

		// ok - Implicit account creation address with basic input
		func() *test {
			inputUTXO1 := &axongo.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 0}
			input1 := tpkg.RandOutputOnAddress(axongo.OutputBasic, inputAddrImplicitAccountCreation)

			inputUTXO2 := &axongo.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 1}
			input2 := &axongo.BasicOutput{
				Amount:           1000,
				UnlockConditions: axongo.BasicOutputUnlockConditions{&axongo.AddressUnlockCondition{Address: inputAddrEd25519}},
			}

			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, signer).
				AddInput(&builder.TxInput{UnlockTarget: inputAddrImplicitAccountCreation, InputID: inputUTXO1.OutputID(), Input: input1}).
				AddInput(&builder.TxInput{UnlockTarget: inputAddrEd25519, InputID: inputUTXO2.OutputID(), Input: input2}).
				AddOutput(output)

			return &test{
				name:    "ok - Implicit account creation address with basic input",
				builder: bdl,
			}
		}(),

		// ok - mix basic+chain outputs
		func() *test {
			var (
				inputID1 = &axongo.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 0}
				inputID2 = &axongo.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 1}
				inputID3 = &axongo.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 4}
				inputID4 = &axongo.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 8}
			)

			var (
				basicOutput = &axongo.BasicOutput{
					Amount:           1000,
					UnlockConditions: axongo.BasicOutputUnlockConditions{&axongo.AddressUnlockCondition{Address: inputAddrEd25519}},
				}

				nftOutput = &axongo.NFTOutput{
					Amount:            1000,
					NFTID:             tpkg.Rand32ByteArray(),
					UnlockConditions:  axongo.NFTOutputUnlockConditions{&axongo.AddressUnlockCondition{Address: inputAddrEd25519}},
					Features:          nil,
					ImmutableFeatures: nil,
				}

				accountOwnedByNFT = &axongo.AccountOutput{
					Amount:    1000,
					AccountID: tpkg.Rand32ByteArray(),
					UnlockConditions: axongo.AccountOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: nftOutput.ChainID().ToAddress()},
					},
				}

				basicOwnedByAccount = &axongo.BasicOutput{
					Amount:           1000,
					UnlockConditions: axongo.BasicOutputUnlockConditions{&axongo.AddressUnlockCondition{Address: accountOwnedByNFT.ChainID().ToAddress()}},
				}
			)

			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, signer).
				AddInput(&builder.TxInput{UnlockTarget: inputAddrEd25519, InputID: inputID1.OutputID(), Input: basicOutput}).
				AddInput(&builder.TxInput{UnlockTarget: inputAddrEd25519, InputID: inputID2.OutputID(), Input: nftOutput}).
				AddInput(&builder.TxInput{UnlockTarget: nftOutput.ChainID().ToAddress(), InputID: inputID3.OutputID(), Input: accountOwnedByNFT}).
				AddInput(&builder.TxInput{UnlockTarget: accountOwnedByNFT.ChainID().ToAddress(), InputID: inputID4.OutputID(), Input: basicOwnedByAccount}).
				AddOutput(output)

			return &test{
				name:    "ok - mix basic+chain outputs",
				builder: bdl,
			}
		}(),

		// ok - with tagged data payload
		func() *test {
			inputUTXO1 := &axongo.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 0}

			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, signer).
				AddInput(&builder.TxInput{UnlockTarget: inputAddrEd25519, InputID: inputUTXO1.OutputID(), Input: tpkg.RandOutputOnAddress(axongo.OutputBasic, inputAddrEd25519)}).
				AddOutput(output).
				AddTaggedDataPayload(&axongo.TaggedData{Tag: []byte("index"), Data: nil})

			return &test{
				name:    "ok - with tagged data payload",
				builder: bdl,
			}
		}(),

		// ok - with context inputs
		func() *test {
			inputUTXO1 := &axongo.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 0}

			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, signer).
				AddInput(&builder.TxInput{UnlockTarget: inputAddrEd25519, InputID: inputUTXO1.OutputID(), Input: tpkg.RandOutputOnAddress(axongo.OutputBasic, inputAddrEd25519)}).
				AddOutput(output).
				AddCommitmentInput(&axongo.CommitmentInput{CommitmentID: tpkg.Rand36ByteArray()}).
				AddBlockIssuanceCreditInput(&axongo.BlockIssuanceCreditInput{AccountID: tpkg.RandAccountID()}).
				AddRewardInput(&axongo.RewardInput{Index: 0}, 100)

			return &test{
				name:    "ok - with context inputs",
				builder: bdl,
			}
		}(),

		// ok - allot all mana
		func() *test {
			inputUTXO1 := &axongo.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 0}

			basicOutput := &axongo.BasicOutput{
				Amount:           1000_000_000,
				UnlockConditions: axongo.BasicOutputUnlockConditions{&axongo.AddressUnlockCondition{Address: inputAddrEd25519}},
			}

			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, signer).
				AddInput(&builder.TxInput{UnlockTarget: inputAddrEd25519, InputID: inputUTXO1.OutputID(), Input: basicOutput}).
				AddOutput(output).
				AllotAllMana(inputUTXO1.CreationSlot()+6, tpkg.RandAccountID(), 20)

			return &test{
				name:    "ok - allot all mana",
				builder: bdl,
			}
		}(),

		// ok - with mana lock condition
		func() *test {
			inputUTXO1 := &axongo.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 0}

			accountAddr := axongo.AccountAddressFromOutputID(inputUTXO1.OutputID())
			basicOutput := &axongo.BasicOutput{
				Amount: 1000,
				UnlockConditions: axongo.BasicOutputUnlockConditions{
					&axongo.AddressUnlockCondition{Address: accountAddr},
					&axongo.TimelockUnlockCondition{Slot: inputUTXO1.CreationSlot()},
				}}

			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, signer).
				AddInput(&builder.TxInput{
					UnlockTarget: inputAddrImplicitAccountCreation,
					InputID:      inputUTXO1.OutputID(),
					Input:        tpkg.RandOutputOnAddress(axongo.OutputBasic, inputAddrImplicitAccountCreation),
				}).
				SetCreationSlot(10).
				AddOutput(basicOutput).
				StoreRemainingManaInOutputAndAllotRemainingAccountBoundMana(inputUTXO1.CreationSlot(), 0)

			return &test{
				name:    "ok - with mana lock condition",
				builder: bdl,
			}
		}(),

		// err - missing address keys (wrong address)
		func() *test {
			inputUTXO1 := &axongo.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 0}

			// wrong address/keys
			wrongAddress := tpkg.RandEd25519PrivateKey()
			//nolint:forcetypeassert // we can safely assume that this is a ed25519.PublicKey
			wrongAddr := axongo.Ed25519AddressFromPubKey(wrongAddress.Public().(ed25519.PublicKey))
			wrongAddrKeys := axongo.AddressKeys{Address: wrongAddr, Keys: wrongAddress}

			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, axongo.NewInMemoryAddressSigner(wrongAddrKeys)).
				AddInput(&builder.TxInput{UnlockTarget: inputAddrEd25519, InputID: inputUTXO1.OutputID(), Input: tpkg.RandOutputOnAddress(axongo.OutputBasic, inputAddrEd25519)}).
				AddOutput(output)

			return &test{
				name:     "err - missing address keys (wrong address)",
				builder:  bdl,
				buildErr: axongo.ErrAddressKeysNotMapped,
			}
		}(),

		// err - missing address keys (no keys given at all)
		func() *test {
			inputUTXO1 := &axongo.UTXOInput{TransactionID: tpkg.Rand36ByteArray(), TransactionOutputIndex: 0}

			bdl := builder.NewTransactionBuilder(tpkg.ZeroCostTestAPI, axongo.NewInMemoryAddressSigner()).
				AddInput(&builder.TxInput{UnlockTarget: inputAddrEd25519, InputID: inputUTXO1.OutputID(), Input: tpkg.RandOutputOnAddress(axongo.OutputBasic, inputAddrEd25519)}).
				AddOutput(output)

			return &test{
				name:     "err - missing address keys (no keys given at all)",
				builder:  bdl,
				buildErr: axongo.ErrAddressKeysNotMapped,
			}
		}(),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.builder.Build()
			if test.buildErr != nil {
				assert.True(t, ierrors.Is(err, test.buildErr), "wrong error : %s != %s", err, test.buildErr)

				return
			}
			assert.NoError(t, err)
		})
	}
}

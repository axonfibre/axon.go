//nolint:forcetypeassert
package axongo_test

import (
	"crypto/ed25519"
	"testing"

	hiveEd25519 "github.com/axonfibre/fibre.go/crypto/ed25519"
	"github.com/axonfibre/fibre.go/serializer/v2/serix"
	axongo "github.com/axonfibre/axon.go/v4"
	"github.com/axonfibre/axon.go/v4/tpkg"
)

var (
	benchLargeTx = &axongo.SignedTransaction{
		API: tpkg.ZeroCostTestAPI,
		Transaction: &axongo.Transaction{
			API: tpkg.ZeroCostTestAPI,
			TransactionEssence: &axongo.TransactionEssence{
				NetworkID:     tpkg.TestNetworkID,
				ContextInputs: axongo.TxEssenceContextInputs{},
				Inputs: func() axongo.TxEssenceInputs {
					var inputs axongo.TxEssenceInputs
					for i := 0; i < axongo.MaxInputsCount; i++ {
						inputs = append(inputs, &axongo.UTXOInput{
							TransactionID:          tpkg.Rand36ByteArray(),
							TransactionOutputIndex: 0,
						})
					}

					return inputs
				}(),
				Allotments:   axongo.Allotments{},
				Capabilities: axongo.TransactionCapabilitiesBitMask{},
				Payload:      nil,
			},
			Outputs: func() axongo.TxEssenceOutputs {
				var outputs axongo.TxEssenceOutputs
				for i := 0; i < axongo.MaxOutputsCount; i++ {
					outputs = append(outputs, &axongo.BasicOutput{
						Amount: 100,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					})
				}

				return outputs
			}(),
		},
		Unlocks: func() axongo.Unlocks {
			var unlocks axongo.Unlocks
			for i := 0; i < axongo.MaxInputsCount; i++ {
				unlocks = append(unlocks, &axongo.SignatureUnlock{
					Signature: tpkg.RandEd25519Signature(),
				})
			}

			return unlocks
		}(),
	}
)

func BenchmarkDeserializationLargeTxPayload(b *testing.B) {
	data, err := tpkg.ZeroCostTestAPI.Encode(benchLargeTx, serix.WithValidation())
	if err != nil {
		b.Fatal(err)
	}

	b.Run("reflection with validation", func(b *testing.B) {
		target := &axongo.SignedTransaction{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tpkg.ZeroCostTestAPI.Decode(data, target, serix.WithValidation())
		}
	})

	b.Run("reflection without validation", func(b *testing.B) {
		target := &axongo.SignedTransaction{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tpkg.ZeroCostTestAPI.Decode(data, target)
		}
	})
}

func BenchmarkDeserializationOneIOTxPayload(b *testing.B) {
	data, err := tpkg.ZeroCostTestAPI.Encode(tpkg.OneInputOutputTransaction(), serix.WithValidation())
	if err != nil {
		b.Fatal(err)
	}

	b.Run("reflection with validation", func(b *testing.B) {
		target := &axongo.SignedTransaction{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tpkg.ZeroCostTestAPI.Decode(data, target, serix.WithValidation())
		}
	})

	b.Run("reflection without validation", func(b *testing.B) {
		target := &axongo.SignedTransaction{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tpkg.ZeroCostTestAPI.Decode(data, target)
		}
	})
}

func BenchmarkSerializationOneIOTxPayload(b *testing.B) {

	b.Run("reflection with validation", func(b *testing.B) {
		txPayload := tpkg.OneInputOutputTransaction()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tpkg.ZeroCostTestAPI.Encode(txPayload, serix.WithValidation())
		}
	})

	b.Run("reflection without validation", func(b *testing.B) {
		txPayload := tpkg.OneInputOutputTransaction()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tpkg.ZeroCostTestAPI.Encode(txPayload)
		}
	})
}

func BenchmarkSignEd25519OneIOTxEssence(b *testing.B) {
	txPayload := tpkg.OneInputOutputTransaction()
	b.ResetTimer()

	txEssenceData, err := txPayload.Transaction.SigningMessage()
	tpkg.Must(err)

	seed := tpkg.RandEd25519Seed()
	prvKey := ed25519.NewKeyFromSeed(seed[:])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ed25519.Sign(prvKey, txEssenceData)
	}
}

func BenchmarkVerifyEd25519OneIOTxEssence(b *testing.B) {
	txPayload := tpkg.OneInputOutputTransaction()
	b.ResetTimer()

	txEssenceData, err := txPayload.Transaction.SigningMessage()
	tpkg.Must(err)

	seed := tpkg.RandEd25519Seed()
	prvKey := ed25519.NewKeyFromSeed(seed[:])

	sig := ed25519.Sign(prvKey, txEssenceData)

	pubKey := prvKey.Public().(ed25519.PublicKey)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hiveEd25519.Verify(pubKey, txEssenceData, sig)
	}
}

func BenchmarkSerializeAndHashBlockWithTransactionPayload(b *testing.B) {
	txPayload := tpkg.OneInputOutputTransaction()

	m := &axongo.Block{
		API: tpkg.ZeroCostTestAPI,
		Header: axongo.BlockHeader{
			ProtocolVersion: tpkg.ZeroCostTestAPI.Version(),
		},
		Body: &axongo.BasicBlockBody{
			API:                tpkg.ZeroCostTestAPI,
			StrongParents:      tpkg.SortedRandBlockIDs(2),
			WeakParents:        axongo.BlockIDs{},
			ShallowLikeParents: axongo.BlockIDs{},
			Payload:            txPayload,
		},
	}
	for i := 0; i < b.N; i++ {
		_, _ = m.ID()
	}
}

package iotago_test

import (
	"testing"

	"github.com/iotaledger/iota.go/v2"
	"github.com/iotaledger/iota.go/v2/ed25519"
)

func BenchmarkDeserializeWithValidationOneIOTxPayload(b *testing.B) {
	data, err := oneInputOutputTransaction().Serialize(iotago.DeSeriModeNoValidation)
	if err != nil {
		b.Fatal(err)
	}

	target := &iotago.Transaction{}
	_, err = target.Deserialize(data, iotago.DeSeriModeNoValidation)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		target.Deserialize(data, iotago.DeSeriModePerformValidation)
	}
}

func BenchmarkDeserializeWithoutValidationOneIOTxPayload(b *testing.B) {
	data, err := oneInputOutputTransaction().Serialize(iotago.DeSeriModeNoValidation)
	if err != nil {
		b.Fatal(err)
	}

	target := &iotago.Transaction{}
	_, err = target.Deserialize(data, iotago.DeSeriModeNoValidation)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		target.Deserialize(data, iotago.DeSeriModeNoValidation)
	}
}

func BenchmarkSerializeWithValidationOneIOTxPayload(b *testing.B) {
	txPayload := oneInputOutputTransaction()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		txPayload.Serialize(iotago.DeSeriModePerformValidation)
	}
}

func BenchmarkSerializeWithoutValidationOneIOTxPayload(b *testing.B) {
	sigTxPayload := oneInputOutputTransaction()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sigTxPayload.Serialize(iotago.DeSeriModeNoValidation)
	}
}

func BenchmarkSignEd25519OneIOTxEssence(b *testing.B) {
	txPayload := oneInputOutputTransaction()
	b.ResetTimer()

	txEssenceData, err := txPayload.Essence.(*iotago.TransactionEssence).SigningMessage()
	must(err)

	seed := randEd25519Seed()
	prvKey := ed25519.NewKeyFromSeed(seed[:])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ed25519.Sign(prvKey, txEssenceData)
	}
}

func BenchmarkVerifyEd25519OneIOTxEssence(b *testing.B) {
	txPayload := oneInputOutputTransaction()
	b.ResetTimer()

	txEssenceData, err := txPayload.Essence.(*iotago.TransactionEssence).SigningMessage()
	must(err)

	seed := randEd25519Seed()
	prvKey := ed25519.NewKeyFromSeed(seed[:])

	sig := ed25519.Sign(prvKey, txEssenceData)

	pubKey := prvKey.Public().(ed25519.PublicKey)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ed25519.Verify(pubKey, txEssenceData, sig)
	}
}

func BenchmarkSerializeAndHashMessageWithTransactionPayload(b *testing.B) {
	txPayload := oneInputOutputTransaction()

	m := &iotago.Message{
		Parents: sortedRand32ByteHashes(2),
		Payload: txPayload,
		Nonce:   0,
	}
	for i := 0; i < b.N; i++ {
		_, _ = m.ID()
	}
}

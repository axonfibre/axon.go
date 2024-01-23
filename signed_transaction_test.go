package iotago_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestTransactionDeSerialize(t *testing.T) {
	tests := []*deSerializeTest{
		{
			name:   "ok - UTXO",
			source: tpkg.RandSignedTransaction(tpkg.ZeroCostTestAPI),
			target: &iotago.SignedTransaction{},
		},
		{
			name: "ok -  Commitment",
			source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				tpkg.RandTransactionWithOptions(
					tpkg.ZeroCostTestAPI,
					tpkg.WithContextInputs(iotago.TxEssenceContextInputs{
						&iotago.CommitmentInput{
							CommitmentID: iotago.CommitmentID{},
						},
					}),
				)),
			target:    &iotago.SignedTransaction{},
			seriErr:   nil,
			deSeriErr: nil,
		},
		{
			name: "ok - BIC",
			source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				tpkg.RandTransactionWithOptions(
					tpkg.ZeroCostTestAPI,
					tpkg.WithContextInputs(iotago.TxEssenceContextInputs{
						&iotago.BlockIssuanceCreditInput{
							AccountID: tpkg.RandAccountID(),
						},
					}),
				)),
			target:    &iotago.SignedTransaction{},
			seriErr:   nil,
			deSeriErr: nil,
		},
		{
			name: "ok - Commitment + BIC",
			source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				tpkg.RandTransactionWithOptions(
					tpkg.ZeroCostTestAPI,
					tpkg.WithContextInputs(iotago.TxEssenceContextInputs{
						&iotago.CommitmentInput{
							CommitmentID: iotago.CommitmentID{},
						},
						&iotago.BlockIssuanceCreditInput{
							AccountID: tpkg.RandAccountID(),
						},
					}),
				)),
			target:    &iotago.SignedTransaction{},
			seriErr:   nil,
			deSeriErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestTransactionDeSerialize_MaxInputsCount(t *testing.T) {
	tests := []*deSerializeTest{
		{
			name: "ok",
			source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				tpkg.RandTransactionWithOptions(
					tpkg.ZeroCostTestAPI,
					tpkg.WithUTXOInputCount(iotago.MaxInputsCount),
					tpkg.WithBlockIssuanceCreditInputCount(iotago.MaxContextInputsCount/2),
					tpkg.WithRewardInputCount(iotago.MaxContextInputsCount/2-1),
					tpkg.WithCommitmentInput(),
				)),
			target:    &iotago.SignedTransaction{},
			seriErr:   nil,
			deSeriErr: nil,
		},
		{
			name: "too many inputs",
			source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				tpkg.RandTransactionWithOptions(
					tpkg.ZeroCostTestAPI,
					tpkg.WithUTXOInputCount(iotago.MaxInputsCount+1),
				)),
			target:    &iotago.SignedTransaction{},
			seriErr:   serializer.ErrArrayValidationMaxElementsExceeded,
			deSeriErr: serializer.ErrArrayValidationMaxElementsExceeded,
		},
		{
			name: "too many context inputs",
			source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				tpkg.RandTransactionWithOptions(
					tpkg.ZeroCostTestAPI,
					tpkg.WithBlockIssuanceCreditInputCount(iotago.MaxContextInputsCount-1),
					tpkg.WithCommitmentInput(),
					func(tx *iotago.Transaction) {
						tx.TransactionEssence.ContextInputs = append(tx.TransactionEssence.ContextInputs, &iotago.RewardInput{
							Index: 0,
						})
					},
				)),
			target:    &iotago.SignedTransaction{},
			seriErr:   serializer.ErrArrayValidationMaxElementsExceeded,
			deSeriErr: serializer.ErrArrayValidationMaxElementsExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestTransactionDeSerialize_MaxOutputsCount(t *testing.T) {
	tests := []*deSerializeTest{
		{
			name:      "ok",
			source:    tpkg.RandSignedTransactionWithOutputCount(tpkg.ZeroCostTestAPI, iotago.MaxOutputsCount),
			target:    &iotago.SignedTransaction{},
			seriErr:   nil,
			deSeriErr: nil,
		},
		{
			name:      "too many outputs",
			source:    tpkg.RandSignedTransactionWithOutputCount(tpkg.ZeroCostTestAPI, iotago.MaxOutputsCount+1),
			target:    &iotago.SignedTransaction{},
			seriErr:   serializer.ErrArrayValidationMaxElementsExceeded,
			deSeriErr: serializer.ErrArrayValidationMaxElementsExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestTransactionDeSerialize_MaxAllotmentsCount(t *testing.T) {
	tests := []*deSerializeTest{
		{
			name:      "ok",
			source:    tpkg.RandSignedTransactionWithAllotmentCount(tpkg.ZeroCostTestAPI, iotago.MaxAllotmentCount),
			target:    &iotago.SignedTransaction{},
			seriErr:   nil,
			deSeriErr: nil,
		},
		{
			name:      "too many allotments",
			source:    tpkg.RandSignedTransactionWithAllotmentCount(tpkg.ZeroCostTestAPI, iotago.MaxAllotmentCount+1),
			target:    &iotago.SignedTransaction{},
			seriErr:   serializer.ErrArrayValidationMaxElementsExceeded,
			deSeriErr: serializer.ErrArrayValidationMaxElementsExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestTransactionDeSerialize_RefUTXOIndexMax(t *testing.T) {
	tests := []*deSerializeTest{
		{
			name: "ok",
			source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				tpkg.RandTransactionWithOptions(tpkg.ZeroCostTestAPI, tpkg.WithInputs(iotago.TxEssenceInputs{
					&iotago.UTXOInput{
						TransactionID:          tpkg.RandTransactionID(),
						TransactionOutputIndex: iotago.RefUTXOIndexMax,
					},
				}))),
			target:    &iotago.SignedTransaction{},
			seriErr:   nil,
			deSeriErr: nil,
		},
		{
			name: "wrong ref index",
			source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				tpkg.RandTransactionWithOptions(tpkg.ZeroCostTestAPI, tpkg.WithInputs(iotago.TxEssenceInputs{
					&iotago.UTXOInput{
						TransactionID:          tpkg.RandTransactionID(),
						TransactionOutputIndex: iotago.RefUTXOIndexMax + 1,
					},
				}))),
			target:    &iotago.SignedTransaction{},
			seriErr:   iotago.ErrRefUTXOIndexInvalid,
			deSeriErr: iotago.ErrRefUTXOIndexInvalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestTransaction_InputTypes(t *testing.T) {
	utxoInput1 := &iotago.UTXOInput{
		TransactionID:          tpkg.RandTransactionID(),
		TransactionOutputIndex: 13,
	}

	utxoInput2 := &iotago.UTXOInput{
		TransactionID:          tpkg.RandTransactionID(),
		TransactionOutputIndex: 11,
	}

	commitmentInput1 := &iotago.CommitmentInput{
		CommitmentID: iotago.CommitmentIDRepresentingData(10, tpkg.RandBytes(32)),
	}

	bicInput1 := &iotago.BlockIssuanceCreditInput{
		AccountID: tpkg.RandAccountID(),
	}
	bicInput2 := &iotago.BlockIssuanceCreditInput{
		AccountID: tpkg.RandAccountID(),
	}

	rewardInput1 := &iotago.RewardInput{
		Index: 3,
	}
	rewardInput2 := &iotago.RewardInput{
		Index: 2,
	}

	signedTransaction := tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
		tpkg.RandTransactionWithOptions(
			tpkg.ZeroCostTestAPI,
			tpkg.WithInputs(iotago.TxEssenceInputs{
				utxoInput1,
				utxoInput2,
			}),
			tpkg.WithContextInputs(iotago.TxEssenceContextInputs{
				commitmentInput1,
				bicInput1,
				bicInput2,
				rewardInput1,
				rewardInput2,
			}),
		))

	utxoInputs, err := signedTransaction.Transaction.Inputs()
	require.NoError(t, err)

	commitmentInput := signedTransaction.Transaction.CommitmentInput()
	require.NotNil(t, commitmentInput)

	bicInputs, err := signedTransaction.Transaction.BICInputs()
	require.NoError(t, err)

	rewardInputs, err := signedTransaction.Transaction.RewardInputs()
	require.NoError(t, err)

	require.Equal(t, 2, len(utxoInputs))
	require.Equal(t, 2, len(bicInputs))
	require.Equal(t, 2, len(rewardInputs))

	require.Contains(t, utxoInputs, utxoInput1)
	require.Contains(t, utxoInputs, utxoInput2)

	require.Equal(t, commitmentInput, commitmentInput1)

	require.Contains(t, bicInputs, bicInput1)
	require.Contains(t, bicInputs, bicInput2)

	require.Contains(t, rewardInputs, rewardInput1)
	require.Contains(t, rewardInputs, rewardInput2)
}

func TestTransaction_Clone(t *testing.T) {
	transaction := tpkg.RandSignedTransaction(tpkg.ZeroCostTestAPI)
	txID, err := transaction.ID()
	require.NoError(t, err)

	//nolint:forcetypeassert
	cpy := transaction.Clone().(*iotago.SignedTransaction)

	cpyTxID, err := cpy.ID()
	require.NoError(t, err)

	require.EqualValues(t, txID, cpyTxID)
}

package axongo_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/axonfibre/fibre.go/serializer/v2"
	axongo "github.com/axonfibre/axon.go/v4"
	"github.com/axonfibre/axon.go/v4/tpkg"
	"github.com/axonfibre/axon.go/v4/tpkg/frameworks"
)

func TestTransactionDeSerialize(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name:   "ok - UTXO",
			Source: tpkg.RandSignedTransaction(tpkg.ZeroCostTestAPI),
			Target: &axongo.SignedTransaction{},
		},
		{
			Name: "ok - Commitment",
			Source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				tpkg.RandTransactionWithOptions(
					tpkg.ZeroCostTestAPI,
					tpkg.WithContextInputs(axongo.TxEssenceContextInputs{
						&axongo.CommitmentInput{
							CommitmentID: axongo.CommitmentID{},
						},
					}),
				)),
			Target:    &axongo.SignedTransaction{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name: "ok - BIC",
			Source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				tpkg.RandTransactionWithOptions(
					tpkg.ZeroCostTestAPI,
					tpkg.WithContextInputs(axongo.TxEssenceContextInputs{
						&axongo.CommitmentInput{},
						&axongo.BlockIssuanceCreditInput{
							AccountID: tpkg.RandAccountID(),
						},
					}),
				)),
			Target:    &axongo.SignedTransaction{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name: "ok - Commitment + BIC",
			Source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				tpkg.RandTransactionWithOptions(
					tpkg.ZeroCostTestAPI,
					tpkg.WithContextInputs(axongo.TxEssenceContextInputs{
						&axongo.CommitmentInput{
							CommitmentID: axongo.CommitmentID{},
						},
						&axongo.BlockIssuanceCreditInput{
							AccountID: tpkg.RandAccountID(),
						},
					}),
				)),
			Target:    &axongo.SignedTransaction{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

func TestTransactionDeSerialize_MaxInputsCount(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name: "ok",
			Source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				tpkg.RandTransactionWithOptions(
					tpkg.ZeroCostTestAPI,
					tpkg.WithUTXOInputCount(axongo.MaxInputsCount),
					tpkg.WithBlockIssuanceCreditInputCount(axongo.MaxContextInputsCount/2),
					tpkg.WithRewardInputCount(axongo.MaxContextInputsCount/2-1),
					tpkg.WithCommitmentInput(),
				)),
			Target:    &axongo.SignedTransaction{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name: "too many inputs",
			Source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				tpkg.RandTransactionWithOptions(
					tpkg.ZeroCostTestAPI,
					tpkg.WithUTXOInputCount(axongo.MaxInputsCount+1),
				)),
			Target:    &axongo.SignedTransaction{},
			SeriErr:   serializer.ErrArrayValidationMaxElementsExceeded,
			DeSeriErr: serializer.ErrArrayValidationMaxElementsExceeded,
		},
		{
			Name: "too many context inputs",
			Source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				tpkg.RandTransactionWithOptions(
					tpkg.ZeroCostTestAPI,
					tpkg.WithBlockIssuanceCreditInputCount(axongo.MaxContextInputsCount-1),
					tpkg.WithCommitmentInput(),
					func(tx *axongo.Transaction) {
						tx.TransactionEssence.ContextInputs = append(tx.TransactionEssence.ContextInputs, &axongo.RewardInput{
							Index: 0,
						})
					},
				)),
			Target:    &axongo.SignedTransaction{},
			SeriErr:   serializer.ErrArrayValidationMaxElementsExceeded,
			DeSeriErr: serializer.ErrArrayValidationMaxElementsExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

func TestTransactionDeSerialize_MaxOutputsCount(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name:      "ok",
			Source:    tpkg.RandSignedTransactionWithOutputCount(tpkg.ZeroCostTestAPI, axongo.MaxOutputsCount),
			Target:    &axongo.SignedTransaction{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name:      "too many outputs",
			Source:    tpkg.RandSignedTransactionWithOutputCount(tpkg.ZeroCostTestAPI, axongo.MaxOutputsCount+1),
			Target:    &axongo.SignedTransaction{},
			SeriErr:   serializer.ErrArrayValidationMaxElementsExceeded,
			DeSeriErr: serializer.ErrArrayValidationMaxElementsExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

func TestTransactionDeSerialize_MaxAllotmentsCount(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name:      "ok",
			Source:    tpkg.RandSignedTransactionWithAllotmentCount(tpkg.ZeroCostTestAPI, axongo.MaxAllotmentCount),
			Target:    &axongo.SignedTransaction{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name:      "too many allotments",
			Source:    tpkg.RandSignedTransactionWithAllotmentCount(tpkg.ZeroCostTestAPI, axongo.MaxAllotmentCount+1),
			Target:    &axongo.SignedTransaction{},
			SeriErr:   serializer.ErrArrayValidationMaxElementsExceeded,
			DeSeriErr: serializer.ErrArrayValidationMaxElementsExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

func TestTransactionDeSerialize_RefUTXOIndexMax(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name: "ok",
			Source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				tpkg.RandTransactionWithOptions(tpkg.ZeroCostTestAPI, tpkg.WithInputs(axongo.TxEssenceInputs{
					&axongo.UTXOInput{
						TransactionID:          tpkg.RandTransactionID(),
						TransactionOutputIndex: axongo.RefUTXOIndexMax,
					},
				}))),
			Target:    &axongo.SignedTransaction{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name: "wrong ref index",
			Source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				tpkg.RandTransactionWithOptions(tpkg.ZeroCostTestAPI, tpkg.WithInputs(axongo.TxEssenceInputs{
					&axongo.UTXOInput{
						TransactionID:          tpkg.RandTransactionID(),
						TransactionOutputIndex: axongo.RefUTXOIndexMax + 1,
					},
				}))),
			Target:    &axongo.SignedTransaction{},
			SeriErr:   axongo.ErrRefUTXOIndexInvalid,
			DeSeriErr: axongo.ErrRefUTXOIndexInvalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

func TestTransaction_InputTypes(t *testing.T) {
	utxoInput1 := &axongo.UTXOInput{
		TransactionID:          tpkg.RandTransactionID(),
		TransactionOutputIndex: 13,
	}

	utxoInput2 := &axongo.UTXOInput{
		TransactionID:          tpkg.RandTransactionID(),
		TransactionOutputIndex: 11,
	}

	commitmentInput1 := &axongo.CommitmentInput{
		CommitmentID: axongo.CommitmentIDRepresentingData(10, tpkg.RandBytes(32)),
	}

	bicInput1 := &axongo.BlockIssuanceCreditInput{
		AccountID: tpkg.RandAccountID(),
	}
	bicInput2 := &axongo.BlockIssuanceCreditInput{
		AccountID: tpkg.RandAccountID(),
	}

	rewardInput1 := &axongo.RewardInput{
		Index: 3,
	}
	rewardInput2 := &axongo.RewardInput{
		Index: 2,
	}

	signedTransaction := tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
		tpkg.RandTransactionWithOptions(
			tpkg.ZeroCostTestAPI,
			tpkg.WithInputs(axongo.TxEssenceInputs{
				utxoInput1,
				utxoInput2,
			}),
			tpkg.WithContextInputs(axongo.TxEssenceContextInputs{
				commitmentInput1,
				bicInput1,
				bicInput2,
				rewardInput1,
				rewardInput2,
			}),
		))

	utxoInputs := signedTransaction.Transaction.Inputs()

	commitmentInput := signedTransaction.Transaction.CommitmentInput()
	require.NotNil(t, commitmentInput)

	bicInputs := signedTransaction.Transaction.BICInputs()

	rewardInputs := signedTransaction.Transaction.RewardInputs()

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
	cpy := transaction.Clone().(*axongo.SignedTransaction)

	cpyTxID, err := cpy.ID()
	require.NoError(t, err)

	require.EqualValues(t, txID, cpyTxID)
}

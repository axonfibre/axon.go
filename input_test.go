package axongo_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	axongo "github.com/axonfibre/axon.go/v4"
	"github.com/axonfibre/axon.go/v4/tpkg"
	"github.com/axonfibre/axon.go/v4/tpkg/frameworks"
)

func TestInputsSyntacticalUnique(t *testing.T) {
	tests := []struct {
		name    string
		inputs  axongo.Inputs[axongo.Input]
		wantErr error
	}{
		{
			name: "ok",
			inputs: axongo.Inputs[axongo.Input]{
				&axongo.UTXOInput{
					TransactionID:          [36]byte{},
					TransactionOutputIndex: 0,
				},
				&axongo.UTXOInput{
					TransactionID:          [36]byte{},
					TransactionOutputIndex: 1,
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - addr not unique",
			inputs: axongo.Inputs[axongo.Input]{
				&axongo.UTXOInput{
					TransactionID:          [36]byte{},
					TransactionOutputIndex: 0,
				},
				&axongo.UTXOInput{
					TransactionID:          [36]byte{},
					TransactionOutputIndex: 0,
				},
			},
			wantErr: axongo.ErrInputUTXORefsNotUnique,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valFunc := axongo.InputsSyntacticalUnique()
			var runErr error
			for index, input := range tt.inputs {
				if err := valFunc(index, input); err != nil {
					runErr = err
				}
			}
			require.ErrorIs(t, runErr, tt.wantErr)
		})
	}
}

func TestContextInputsRewardInputMaxIndex(t *testing.T) {
	tests := []struct {
		name    string
		inputs  axongo.ContextInputs[axongo.ContextInput]
		wantErr error
	}{
		{
			name: "ok",
			inputs: axongo.ContextInputs[axongo.ContextInput]{
				&axongo.CommitmentInput{
					CommitmentID: tpkg.Rand36ByteArray(),
				},
				&axongo.BlockIssuanceCreditInput{
					AccountID: tpkg.RandAccountID(),
				},
				&axongo.RewardInput{
					Index: 2,
				},
				&axongo.BlockIssuanceCreditInput{
					AccountID: tpkg.RandAccountID(),
				},
				&axongo.RewardInput{
					Index: 4,
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - reward input references index equal to inputs count",
			inputs: axongo.ContextInputs[axongo.ContextInput]{
				&axongo.RewardInput{
					Index: 1,
				},
				&axongo.RewardInput{
					Index: axongo.MaxInputsCount / 2,
				},
			},
			wantErr: axongo.ErrInputRewardIndexExceedsMaxInputsCount,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valFunc := axongo.ContextInputsRewardInputMaxIndex(axongo.MaxInputsCount / 2)
			var runErr error
			for index, input := range tt.inputs {
				if err := valFunc(index, input); err != nil {
					runErr = err
				}
			}
			require.ErrorIs(t, runErr, tt.wantErr)
		})
	}
}

func TestInputsSyntacticalIndicesWithinBounds(t *testing.T) {
	tests := []struct {
		name    string
		inputs  axongo.Inputs[axongo.Input]
		wantErr error
	}{
		{
			name: "ok",
			inputs: axongo.Inputs[axongo.Input]{
				&axongo.UTXOInput{
					TransactionID:          [36]byte{},
					TransactionOutputIndex: 0,
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - invalid reference UTXO index",
			inputs: axongo.Inputs[axongo.Input]{
				&axongo.UTXOInput{
					TransactionID:          [36]byte{},
					TransactionOutputIndex: 250,
				},
			},
			wantErr: axongo.ErrRefUTXOIndexInvalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valFunc := axongo.InputsSyntacticalIndicesWithinBounds()
			var runErr error
			for index, input := range tt.inputs {
				if err := valFunc(index, input); err != nil {
					runErr = err
				}
			}
			require.ErrorIs(t, runErr, tt.wantErr)
		})
	}
}

func TestInputDeSerialize(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name: "ok - UTXO",
			Source: &axongo.UTXOInput{
				TransactionID:          [36]byte{},
				TransactionOutputIndex: 0,
			},
			Target:    &axongo.UTXOInput{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name: "ok - Commitment",
			Source: &axongo.CommitmentInput{
				CommitmentID: axongo.CommitmentID{},
			},
			Target:    &axongo.CommitmentInput{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name: "ok - BIC",
			Source: &axongo.BlockIssuanceCreditInput{
				AccountID: tpkg.RandAccountID(),
			},
			Target:    &axongo.BlockIssuanceCreditInput{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
		{
			Name: "ok - Reward",
			Source: &axongo.RewardInput{
				Index: 6,
			},
			Target:    &axongo.RewardInput{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

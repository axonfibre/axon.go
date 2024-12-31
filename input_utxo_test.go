package axongo_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	axongo "github.com/axonfibre/axon.go/v4"
	"github.com/axonfibre/axon.go/v4/tpkg"
	"github.com/axonfibre/axon.go/v4/tpkg/frameworks"
)

func TestUTXOInput_DeSerialize(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name:   "",
			Source: tpkg.RandUTXOInput(),
			Target: &axongo.UTXOInput{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

func TestUTXOInput_Equals(t *testing.T) {
	input1 := &axongo.UTXOInput{axongo.TransactionID{1, 2, 3, 4, 5, 6, 7}, 10}
	input2 := &axongo.UTXOInput{axongo.TransactionID{1, 2, 3, 4, 5, 6, 7}, 10}
	input3 := &axongo.UTXOInput{axongo.TransactionID{1, 2, 3, 4, 5, 6, 8}, 10}
	input4 := &axongo.UTXOInput{axongo.TransactionID{1, 2, 3, 4, 5, 6, 7}, 12}
	//nolint:gocritic // false positive
	require.True(t, input1.Equals(input1))
	require.True(t, input1.Equals(input2))
	require.False(t, input1.Equals(input3))
	require.False(t, input1.Equals(input4))
	require.False(t, input3.Equals(input4))
}

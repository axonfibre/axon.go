//nolint:forcetypeassert
package builder_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	axongo "github.com/axonfibre/axon.go/v4"
	"github.com/axonfibre/axon.go/v4/builder"
	"github.com/axonfibre/axon.go/v4/tpkg"
)

func TestBasicBlockBuilder(t *testing.T) {
	parents := tpkg.SortedRandBlockIDs(4)

	taggedDataPayload := &axongo.TaggedData{
		Tag:  []byte("hello world"),
		Data: []byte{1, 2, 3, 4},
	}
	block, err := builder.NewBasicBlockBuilder(axongo.V3API(tpkg.IOTAMainnetV3TestProtocolParameters)).
		Payload(taggedDataPayload).
		StrongParents(parents).
		CalculateAndSetMaxBurnedMana(100).
		Build()
	require.NoError(t, err)

	require.Equal(t, axongo.BlockBodyTypeBasic, block.Body.Type())

	basicBlock := block.Body.(*axongo.BasicBlockBody)
	expectedBurnedMana, err := block.ManaCost(100)
	require.NoError(t, err)
	require.EqualValues(t, expectedBurnedMana, basicBlock.MaxBurnedMana)
}

func TestValidationBlockBuilder(t *testing.T) {
	parents := tpkg.SortedRandBlockIDs(4)

	block, err := builder.NewValidationBlockBuilder(tpkg.ZeroCostTestAPI).
		StrongParents(parents).
		HighestSupportedVersion(100).
		Build()
	require.NoError(t, err)

	require.Equal(t, axongo.BlockBodyTypeValidation, block.Body.Type())

	basicBlock := block.Body.(*axongo.ValidationBlockBody)
	require.EqualValues(t, 100, basicBlock.HighestSupportedVersion)
}

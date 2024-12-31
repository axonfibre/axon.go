package axongo_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/axonfibre/fibre.go/lo"
	axongo "github.com/axonfibre/axon.go/v4"
	"github.com/axonfibre/axon.go/v4/builder"
	"github.com/axonfibre/axon.go/v4/tpkg"
)

func TestAttestation(t *testing.T) {
	block, err := builder.NewValidationBlockBuilder(tpkg.ZeroCostTestAPI).
		StrongParents(tpkg.SortedRandBlockIDs(2)).
		Sign(tpkg.RandAccountID(), tpkg.RandEd25519PrivateKey()).
		Build()

	require.NoError(t, err)
	require.Equal(t, axongo.BlockBodyTypeValidation, block.Body.Type())

	attestation := axongo.NewAttestation(tpkg.ZeroCostTestAPI, block)

	// Compare fields of block and attestation.
	{
		require.Equal(t, block.Header, attestation.Header)
		require.Equal(t, lo.PanicOnErr(block.Body.Hash()), attestation.BodyHash)
		require.Equal(t, block.Signature, attestation.Signature)
	}

	// Compare block ID and attestation block ID.
	{
		blockID, err := block.ID()
		require.NoError(t, err)

		blockIDFromAttestation, err := attestation.BlockID()
		require.NoError(t, err)

		require.Equal(t, blockID, blockIDFromAttestation)
	}

	// Check validity of signature.
	{
		valid, err := attestation.VerifySignature()
		require.NoError(t, err)
		require.True(t, valid)
	}
}

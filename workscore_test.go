package axongo_test

import (
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/require"

	hiveEd25519 "github.com/axonfibre/fibre.go/crypto/ed25519"
	axongo "github.com/axonfibre/axon.go/v4"
	"github.com/axonfibre/axon.go/v4/builder"
	"github.com/axonfibre/axon.go/v4/tpkg"
)

func TestTransactionEssenceWorkScore(t *testing.T) {
	keyPair := hiveEd25519.GenerateKeyPair()
	keyPair2 := hiveEd25519.GenerateKeyPair()
	// Derive a dummy account from addr.
	addr := axongo.Ed25519AddressFromPubKey(keyPair.PublicKey[:])

	output1 := &axongo.BasicOutput{
		Amount: 100000,
		UnlockConditions: axongo.BasicOutputUnlockConditions{
			&axongo.AddressUnlockCondition{
				Address: addr,
			},
		},
		Features: axongo.BasicOutputFeatures{
			tpkg.RandNativeTokenFeature(),
		},
	}
	output2 := &axongo.AccountOutput{
		Amount: 1_000_000,
		UnlockConditions: axongo.AccountOutputUnlockConditions{
			&axongo.AddressUnlockCondition{addr},
		},
		Features: axongo.AccountOutputFeatures{
			&axongo.BlockIssuerFeature{
				ExpirySlot: 300,
				BlockIssuerKeys: axongo.BlockIssuerKeys{
					axongo.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(keyPair.PublicKey),
					axongo.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(keyPair2.PublicKey),
				},
			},
			&axongo.StakingFeature{
				StakedAmount: 500_00,
				FixedCost:    500,
			},
		},
	}

	api := axongo.V3API(
		axongo.NewV3SnapshotProtocolParameters(
			axongo.WithStorageOptions(0, 0, 0, 0, 0, 0),
			axongo.WithWorkScoreOptions(1, 2, 3, 4, 5, 6, 7, 8, 9, 10),
		),
	)

	tx, err := builder.NewTransactionBuilder(api, axongo.NewInMemoryAddressSigner(axongo.AddressKeys{Address: addr, Keys: ed25519.PrivateKey(keyPair.PrivateKey[:])})).
		AddInput(&builder.TxInput{
			UnlockTarget: addr,
			InputID:      tpkg.RandOutputID(0),
			Input:        output1,
		}).
		AddInput(&builder.TxInput{
			UnlockTarget: addr,
			InputID:      tpkg.RandOutputID(0),
			Input:        output1,
		}).
		AddOutput(output1).
		AddOutput(output2).
		AddCommitmentInput(&axongo.CommitmentInput{CommitmentID: axongo.NewCommitmentID(85, tpkg.Rand32ByteArray())}).
		AddBlockIssuanceCreditInput(&axongo.BlockIssuanceCreditInput{AccountID: tpkg.RandAccountID()}).
		AddRewardInput(&axongo.RewardInput{Index: 0}, 0).
		IncreaseAllotment(tpkg.RandAccountID(), tpkg.RandMana(10000)+1).
		IncreaseAllotment(tpkg.RandAccountID(), tpkg.RandMana(10000)+1).
		Build()
	require.NoError(t, err)

	block, err := builder.NewBasicBlockBuilder(api).Payload(tx).Build()
	require.NoError(t, err)

	workScore, err := block.WorkScore()
	require.NoError(t, err)

	workScoreParameters := api.ProtocolParameters().WorkScoreParameters()

	// Calculate work score as defined in TIP-45 for verification.
	expectedWorkScore := workScoreParameters.DataByte*axongo.WorkScore(tx.Size()) +
		workScoreParameters.Block*1 +
		workScoreParameters.Input*2 +
		workScoreParameters.ContextInput*3 +
		workScoreParameters.Output*2 +
		workScoreParameters.NativeToken*1 +
		workScoreParameters.Staking +
		workScoreParameters.BlockIssuer +
		workScoreParameters.Allotment*2 +
		// Accounts for one Signature unlock.
		workScoreParameters.SignatureEd25519

	require.Equal(t, expectedWorkScore, workScore, "work score expected: %d, actual: %d", expectedWorkScore, workScore)
}

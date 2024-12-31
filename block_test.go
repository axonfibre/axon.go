package axongo_test

import (
	"crypto/ed25519"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	hiveEd25519 "github.com/axonfibre/fibre.go/crypto/ed25519"
	"github.com/axonfibre/fibre.go/lo"
	"github.com/axonfibre/fibre.go/serializer/v2"
	"github.com/axonfibre/fibre.go/serializer/v2/serix"
	axongo "github.com/axonfibre/axon.go/v4"
	"github.com/axonfibre/axon.go/v4/builder"
	"github.com/axonfibre/axon.go/v4/hexutil"
	"github.com/axonfibre/axon.go/v4/tpkg"
	"github.com/axonfibre/axon.go/v4/tpkg/frameworks"
)

func TestBlock_DeSerialize(t *testing.T) {
	blockID1 := axongo.MustBlockIDFromHexString("0x960192696d2c99fe338a212f223f96e72c11147ca23490806c1bb18e4d76995ccbfb91ae")
	blockID2 := axongo.MustBlockIDFromHexString("0xc9e20c8bf3b1655b6fc385aebde8e25a668bd4109f5c698eb1b30b31fbbcfb5e6b9dd933")
	blockID3 := axongo.MustBlockIDFromHexString("0xf2520bde652b46d7119a6d2a3b83947ce2d8a79867d37262e91f129215e5098f3f011d8e")

	tests := []*frameworks.DeSerializeTest{
		{
			Name:   "ok - no payload",
			Source: tpkg.RandBlock(tpkg.RandBasicBlockBody(tpkg.ZeroCostTestAPI, 255), tpkg.ZeroCostTestAPI, 0),
			Target: &axongo.Block{},
		},
		{
			Name:   "ok - transaction",
			Source: tpkg.RandBlock(tpkg.RandBasicBlockBody(tpkg.ZeroCostTestAPI, axongo.PayloadSignedTransaction), tpkg.ZeroCostTestAPI, 0),
			Target: &axongo.Block{},
		},
		{
			Name:   "ok - tagged data",
			Source: tpkg.RandBlock(tpkg.RandBasicBlockBody(tpkg.ZeroCostTestAPI, axongo.PayloadTaggedData), tpkg.ZeroCostTestAPI, 0),
			Target: &axongo.Block{},
		},
		{
			Name:   "ok - validation block",
			Source: tpkg.RandBlock(tpkg.RandValidationBlockBody(tpkg.ZeroCostTestAPI), tpkg.ZeroCostTestAPI, 0),
			Target: &axongo.Block{},
		},
		{
			Name: "ok - basic block parent ids sorted",
			Source: func() *axongo.Block {
				block := tpkg.RandBlock(tpkg.RandBasicBlockBody(tpkg.ZeroCostTestAPI, axongo.PayloadTaggedData), tpkg.ZeroCostTestAPI, 1)
				//nolint:forcetypeassert
				basicBlockBody := block.Body.(*axongo.BasicBlockBody)
				basicBlockBody.ShallowLikeParents = axongo.BlockIDs{}
				basicBlockBody.StrongParents = axongo.BlockIDs{
					blockID1,
					blockID2,
					blockID3,
				}
				basicBlockBody.WeakParents = axongo.BlockIDs{}

				return block
			}(),
			Target: &axongo.Block{},
		},
		{
			Name: "fail - basic block strong parent ids unsorted",
			Source: func() *axongo.Block {
				block := tpkg.RandBlock(tpkg.RandBasicBlockBody(tpkg.ZeroCostTestAPI, axongo.PayloadTaggedData), tpkg.ZeroCostTestAPI, 1)
				//nolint:forcetypeassert
				basicBlockBody := block.Body.(*axongo.BasicBlockBody)
				basicBlockBody.ShallowLikeParents = axongo.BlockIDs{}
				basicBlockBody.StrongParents = axongo.BlockIDs{
					blockID1,
					blockID3,
					blockID2,
				}
				basicBlockBody.WeakParents = axongo.BlockIDs{}

				return block
			}(),
			Target:    &axongo.Block{},
			SeriErr:   axongo.ErrArrayValidationOrderViolatesLexicalOrder,
			DeSeriErr: axongo.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			Name: "fail - validation block weak parent ids unsorted",
			Source: func() *axongo.Block {
				block := tpkg.RandBlock(tpkg.RandBasicBlockBody(tpkg.ZeroCostTestAPI, axongo.PayloadTaggedData), tpkg.ZeroCostTestAPI, 1)
				//nolint:forcetypeassert
				basicBlockBody := block.Body.(*axongo.BasicBlockBody)
				basicBlockBody.ShallowLikeParents = axongo.BlockIDs{}
				basicBlockBody.StrongParents = axongo.BlockIDs{
					tpkg.RandBlockID(),
				}
				basicBlockBody.WeakParents = axongo.BlockIDs{
					blockID1,
					blockID3,
					blockID2,
				}

				return block
			}(),
			Target:    &axongo.Block{},
			SeriErr:   axongo.ErrArrayValidationOrderViolatesLexicalOrder,
			DeSeriErr: axongo.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			Name: "fail - max block size exceeded",
			Source: func() *axongo.Block {
				bigBasicOutput := func() *axongo.BasicOutput {
					return &axongo.BasicOutput{
						Amount: 10_000_000,
						UnlockConditions: axongo.BasicOutputUnlockConditions{
							&axongo.AddressUnlockCondition{
								Address: tpkg.RandEd25519Address(),
							},
						},
						Features: axongo.BasicOutputFeatures{
							&axongo.MetadataFeature{
								Entries: axongo.MetadataFeatureEntries{
									"x": tpkg.RandBytes(8150),
								},
							},
						},
					}
				}

				tx := tpkg.RandSignedTransaction(tpkg.ZeroCostTestAPI, func(t *axongo.Transaction) {
					t.Outputs = axongo.TxEssenceOutputs{
						bigBasicOutput(),
						bigBasicOutput(),
						bigBasicOutput(),
						bigBasicOutput(),
					}
				})
				block := tpkg.RandBlock(tpkg.RandBasicBlockBodyWithPayload(tpkg.ZeroCostTestAPI, tx), tpkg.ZeroCostTestAPI, 1)

				return block
			}(),
			Target:    &axongo.Block{},
			SeriErr:   axongo.ErrBlockMaxSizeExceeded,
			DeSeriErr: axongo.ErrBlockMaxSizeExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

func createBlockWithParents(t *testing.T, strongParents, weakParents, shallowLikeParent axongo.BlockIDs, apiProvider *axongo.EpochBasedProvider) error {
	t.Helper()

	apiForSlot := apiProvider.LatestAPI()

	block, err := builder.NewBasicBlockBuilder(apiForSlot).
		StrongParents(strongParents).
		WeakParents(weakParents).
		ShallowLikeParents(shallowLikeParent).
		IssuingTime(time.Now()).
		SlotCommitmentID(axongo.NewCommitment(apiForSlot.Version(), apiForSlot.TimeProvider().CurrentSlot()-apiForSlot.ProtocolParameters().MinCommittableAge(), axongo.CommitmentID{}, axongo.Identifier{}, 0, 0).MustID()).
		Build()
	require.NoError(t, err)

	return lo.Return2(apiForSlot.Encode(block, serix.WithValidation()))
}

func createBlockAtSlot(t *testing.T, blockIndex, commitmentIndex axongo.SlotIndex, apiProvider *axongo.EpochBasedProvider) error {
	t.Helper()

	apiForSlot := apiProvider.APIForSlot(blockIndex)

	block, err := builder.NewBasicBlockBuilder(apiForSlot).
		StrongParents(axongo.BlockIDs{tpkg.RandBlockID()}).
		IssuingTime(apiForSlot.TimeProvider().SlotStartTime(blockIndex)).
		SlotCommitmentID(axongo.NewCommitment(apiForSlot.Version(), commitmentIndex, axongo.CommitmentID{}, axongo.Identifier{}, 0, 0).MustID()).
		Build()
	require.NoError(t, err)

	return lo.Return2(apiForSlot.Encode(block, serix.WithValidation()))
}

func createBlockAtSlotWithVersion(t *testing.T, blockIndex axongo.SlotIndex, version axongo.Version, apiProvider *axongo.EpochBasedProvider) error {
	t.Helper()

	apiForSlot := apiProvider.APIForSlot(blockIndex)
	block, err := builder.NewBasicBlockBuilder(apiForSlot).
		ProtocolVersion(version).
		StrongParents(axongo.BlockIDs{axongo.BlockID{}}).
		IssuingTime(apiForSlot.TimeProvider().SlotStartTime(blockIndex)).
		SlotCommitmentID(axongo.NewCommitment(apiForSlot.Version(), blockIndex-apiForSlot.ProtocolParameters().MinCommittableAge(), axongo.CommitmentID{}, axongo.Identifier{}, 0, 0).MustID()).
		Build()
	require.NoError(t, err)

	return lo.Return2(apiForSlot.Encode(block, serix.WithValidation()))
}

//nolint:unparam // in the test we always issue at blockIndex=100, but let's keep this flexibility.
func createBlockAtSlotWithPayload(t *testing.T, blockIndex, commitmentIndex axongo.SlotIndex, payload axongo.ApplicationPayload, apiProvider *axongo.EpochBasedProvider) error {
	t.Helper()

	apiForSlot := apiProvider.APIForSlot(blockIndex)

	block, err := builder.NewBasicBlockBuilder(apiForSlot).
		StrongParents(axongo.BlockIDs{tpkg.RandBlockID()}).
		IssuingTime(apiForSlot.TimeProvider().SlotStartTime(blockIndex)).
		SlotCommitmentID(axongo.NewCommitment(apiForSlot.Version(), commitmentIndex, axongo.CommitmentID{}, axongo.Identifier{}, 0, 0).MustID()).
		Payload(payload).
		Build()
	require.NoError(t, err)

	return lo.Return2(apiForSlot.Encode(block, serix.WithValidation()))
}

func TestBlock_ProtocolVersionSyntactical(t *testing.T) {
	apiProvider := axongo.NewEpochBasedProvider(
		axongo.WithAPIForMissingVersionCallback(
			func(parameters axongo.ProtocolParameters) (axongo.API, error) {
				return axongo.V3API(axongo.NewV3SnapshotProtocolParameters(axongo.WithVersion(parameters.Version()))), nil
			},
		),
	)
	apiProvider.AddProtocolParametersAtEpoch(axongo.NewV3SnapshotProtocolParameters(), 0)
	apiProvider.AddProtocolParametersAtEpoch(axongo.NewV3SnapshotProtocolParameters(axongo.WithVersion(4)), 3)

	timeProvider := apiProvider.CommittedAPI().TimeProvider()

	require.ErrorIs(t, createBlockAtSlotWithVersion(t, timeProvider.EpochStart(1), 2, apiProvider), axongo.ErrInvalidBlockVersion)

	require.NoError(t, createBlockAtSlotWithVersion(t, timeProvider.EpochEnd(1), 3, apiProvider))

	require.NoError(t, createBlockAtSlotWithVersion(t, timeProvider.EpochEnd(2), 3, apiProvider))

	require.ErrorIs(t, createBlockAtSlotWithVersion(t, timeProvider.EpochStart(3), 3, apiProvider), axongo.ErrInvalidBlockVersion)

	require.NoError(t, createBlockAtSlotWithVersion(t, timeProvider.EpochStart(3), 4, apiProvider))

	require.NoError(t, createBlockAtSlotWithVersion(t, timeProvider.EpochEnd(3), 4, apiProvider))

	require.NoError(t, createBlockAtSlotWithVersion(t, timeProvider.EpochStart(5), 4, apiProvider))

	apiProvider.AddProtocolParametersAtEpoch(axongo.NewV3SnapshotProtocolParameters(axongo.WithVersion(5)), 10)

	require.NoError(t, createBlockAtSlotWithVersion(t, timeProvider.EpochEnd(9), 4, apiProvider))

	require.ErrorIs(t, createBlockAtSlotWithVersion(t, timeProvider.EpochStart(10), 4, apiProvider), axongo.ErrInvalidBlockVersion)

	require.NoError(t, createBlockAtSlotWithVersion(t, timeProvider.EpochStart(10), 5, apiProvider))
}

func TestBlock_Commitments(t *testing.T) {
	// with the following parameters, a block issued in slot 100 can commit between slot 80 and 90
	apiProvider := axongo.NewEpochBasedProvider()
	apiProvider.AddProtocolParametersAtEpoch(
		axongo.NewV3SnapshotProtocolParameters(
			axongo.WithTimeProviderOptions(0, time.Now().Add(-20*time.Minute).Unix(), 10, 13),
			axongo.WithLivenessOptions(15, 30, 11, 21, 60),
		), 0)

	require.ErrorIs(t, createBlockAtSlot(t, 100, 78, apiProvider), axongo.ErrCommitmentTooOld)

	require.ErrorIs(t, createBlockAtSlot(t, 100, 90, apiProvider), axongo.ErrCommitmentTooRecent)

	require.NoError(t, createBlockAtSlot(t, 100, 89, apiProvider))

	require.NoError(t, createBlockAtSlot(t, 100, 80, apiProvider))

	require.NoError(t, createBlockAtSlot(t, 100, 85, apiProvider))
}

func TestBlock_Commitments1(t *testing.T) {
	// with the following parameters, a block issued in slot 100 can commit between slot 80 and 90
	apiProvider := axongo.NewEpochBasedProvider()
	apiProvider.AddProtocolParametersAtEpoch(
		axongo.NewV3SnapshotProtocolParameters(
			axongo.WithTimeProviderOptions(0, time.Now().Add(-20*time.Minute).Unix(), 10, 13),
			axongo.WithLivenessOptions(15, 30, 7, 21, 60),
		), 0)

	require.ErrorIs(t, createBlockAtSlot(t, 10, 4, apiProvider), axongo.ErrCommitmentTooRecent)

}

func TestBlock_TransactionCreationTime(t *testing.T) {
	keyPair := hiveEd25519.GenerateKeyPair()
	// We derive a dummy account from addr.
	addr := axongo.Ed25519AddressFromPubKey(keyPair.PublicKey[:])
	output := &axongo.BasicOutput{
		Amount: 100000,
		UnlockConditions: axongo.BasicOutputUnlockConditions{
			&axongo.AddressUnlockCondition{
				Address: addr,
			},
		},
	}
	// with the following parameters, block issued in slot 110 can contain a transaction with commitment input referencing
	// commitments between 90 and slot that the block commits to (100 at most)
	apiProvider := axongo.NewEpochBasedProvider()
	apiProvider.AddProtocolParametersAtEpoch(
		axongo.NewV3SnapshotProtocolParameters(
			axongo.WithTimeProviderOptions(0, time.Now().Add(-20*time.Minute).Unix(), 10, 13),
			axongo.WithLivenessOptions(15, 30, 7, 21, 60),
		), 0)

	creationSlotTooRecent, err := builder.NewTransactionBuilder(apiProvider.LatestAPI(), axongo.NewInMemoryAddressSigner(axongo.AddressKeys{Address: addr, Keys: ed25519.PrivateKey(keyPair.PrivateKey[:])})).
		AddInput(&builder.TxInput{
			UnlockTarget: addr,
			InputID:      tpkg.RandOutputID(0),
			Input:        output,
		}).
		AddOutput(output).
		SetCreationSlot(101).
		AddCommitmentInput(&axongo.CommitmentInput{CommitmentID: axongo.NewCommitmentID(78, tpkg.Rand32ByteArray())}).
		Build()

	require.NoError(t, err)

	require.ErrorIs(t, createBlockAtSlotWithPayload(t, 100, 79, creationSlotTooRecent, apiProvider), axongo.ErrTransactionCreationSlotTooRecent)

	creationSlotCorrectEqual, err := builder.NewTransactionBuilder(apiProvider.LatestAPI(), axongo.NewInMemoryAddressSigner(axongo.AddressKeys{Address: addr, Keys: ed25519.PrivateKey(keyPair.PrivateKey[:])})).
		AddInput(&builder.TxInput{
			UnlockTarget: addr,
			InputID:      tpkg.RandOutputID(0),
			Input:        output,
		}).
		AddOutput(output).
		SetCreationSlot(100).
		Build()

	require.NoError(t, err)

	require.NoError(t, createBlockAtSlotWithPayload(t, 100, 89, creationSlotCorrectEqual, apiProvider))

	creationSlotCorrectSmallerThanCommitment, err := builder.NewTransactionBuilder(apiProvider.LatestAPI(), axongo.NewInMemoryAddressSigner(axongo.AddressKeys{Address: addr, Keys: ed25519.PrivateKey(keyPair.PrivateKey[:])})).
		AddInput(&builder.TxInput{
			UnlockTarget: addr,
			InputID:      tpkg.RandOutputID(0),
			Input:        output,
		}).
		AddOutput(output).
		SetCreationSlot(1).
		Build()

	require.NoError(t, err)

	require.NoError(t, createBlockAtSlotWithPayload(t, 100, 89, creationSlotCorrectSmallerThanCommitment, apiProvider))

	creationSlotCorrectLargerThanCommitment, err := builder.NewTransactionBuilder(apiProvider.LatestAPI(), axongo.NewInMemoryAddressSigner(axongo.AddressKeys{Address: addr, Keys: ed25519.PrivateKey(keyPair.PrivateKey[:])})).
		AddInput(&builder.TxInput{
			UnlockTarget: addr,
			InputID:      tpkg.RandOutputID(0),
			Input:        output,
		}).
		AddOutput(output).
		SetCreationSlot(99).
		Build()

	require.NoError(t, err)

	require.NoError(t, createBlockAtSlotWithPayload(t, 100, 89, creationSlotCorrectLargerThanCommitment, apiProvider))
}

func TestBlock_WeakParents(t *testing.T) {
	// with the following parameters, a block issued in slot 100 can commit between slot 80 and 90
	apiProvider := axongo.NewEpochBasedProvider()
	apiProvider.AddProtocolParametersAtEpoch(
		axongo.NewV3SnapshotProtocolParameters(
			axongo.WithTimeProviderOptions(0, time.Now().Add(-20*time.Minute).Unix(), 10, 13),
			axongo.WithLivenessOptions(15, 30, 10, 20, 60),
		), 0)
	strongParent1 := tpkg.RandBlockID()
	strongParent2 := tpkg.RandBlockID()
	weakParent1 := tpkg.RandBlockID()
	weakParent2 := tpkg.RandBlockID()
	shallowLikeParent1 := tpkg.RandBlockID()
	shallowLikeParent2 := tpkg.RandBlockID()
	require.ErrorIs(t, createBlockWithParents(
		t,
		axongo.BlockIDs{strongParent1, strongParent2},
		axongo.BlockIDs{weakParent1, weakParent2, shallowLikeParent2},
		axongo.BlockIDs{shallowLikeParent1, shallowLikeParent2},
		apiProvider,
	), axongo.ErrWeakParentsInvalid)

	require.ErrorIs(t, createBlockWithParents(
		t,
		axongo.BlockIDs{strongParent1, strongParent2},
		axongo.BlockIDs{weakParent1, weakParent2, strongParent2},
		axongo.BlockIDs{shallowLikeParent1, shallowLikeParent2},
		apiProvider,
	), axongo.ErrWeakParentsInvalid)

	require.NoError(t, createBlockWithParents(
		t,
		axongo.BlockIDs{strongParent1, strongParent2},
		axongo.BlockIDs{weakParent1, weakParent2},
		axongo.BlockIDs{shallowLikeParent1, shallowLikeParent2},
		apiProvider,
	))

	require.NoError(t, createBlockWithParents(
		t,
		axongo.BlockIDs{strongParent1, strongParent2},
		axongo.BlockIDs{weakParent1, weakParent2},
		axongo.BlockIDs{shallowLikeParent1, shallowLikeParent2, strongParent2},
		apiProvider,
	))
}

func TestBlock_TransactionCommitmentInput(t *testing.T) {
	keyPair := hiveEd25519.GenerateKeyPair()
	// We derive a dummy account from addr.
	addr := axongo.Ed25519AddressFromPubKey(keyPair.PublicKey[:])
	output := &axongo.BasicOutput{
		Amount: 100000,
		UnlockConditions: axongo.BasicOutputUnlockConditions{
			&axongo.AddressUnlockCondition{
				Address: addr,
			},
		},
	}
	// with the following parameters, block issued in slot 110 can contain a transaction with commitment input referencing
	// commitments between 90 and slot that the block commits to (100 at most)
	apiProvider := axongo.NewEpochBasedProvider()
	apiProvider.AddProtocolParametersAtEpoch(
		axongo.NewV3SnapshotProtocolParameters(
			axongo.WithTimeProviderOptions(0, time.Now().Add(-20*time.Minute).Unix(), 10, 13),
			axongo.WithLivenessOptions(15, 30, 11, 21, 60),
		), 0)

	commitmentInputTooOld, err := builder.NewTransactionBuilder(apiProvider.LatestAPI(), axongo.NewInMemoryAddressSigner(axongo.AddressKeys{Address: addr, Keys: ed25519.PrivateKey(keyPair.PrivateKey[:])})).
		AddInput(&builder.TxInput{
			UnlockTarget: addr,
			InputID:      tpkg.RandOutputID(0),
			Input:        output,
		}).
		AddOutput(output).
		AddCommitmentInput(&axongo.CommitmentInput{CommitmentID: axongo.NewCommitmentID(78, tpkg.Rand32ByteArray())}).
		Build()

	require.NoError(t, err)

	require.ErrorIs(t, createBlockAtSlotWithPayload(t, 100, 79, commitmentInputTooOld, apiProvider), axongo.ErrCommitmentInputTooOld)

	commitmentInputTooRecent, err := builder.NewTransactionBuilder(apiProvider.LatestAPI(), axongo.NewInMemoryAddressSigner(axongo.AddressKeys{Address: addr, Keys: ed25519.PrivateKey(keyPair.PrivateKey[:])})).
		AddInput(&builder.TxInput{
			UnlockTarget: addr,
			InputID:      tpkg.RandOutputID(0),
			Input:        output,
		}).
		AddOutput(output).
		AddCommitmentInput(&axongo.CommitmentInput{CommitmentID: axongo.NewCommitmentID(90, tpkg.Rand32ByteArray())}).
		Build()

	require.NoError(t, err)

	require.ErrorIs(t, createBlockAtSlotWithPayload(t, 100, 89, commitmentInputTooRecent, apiProvider), axongo.ErrCommitmentInputTooRecent)

	commitmentInputNewerThanBlockCommitment, err := builder.NewTransactionBuilder(apiProvider.LatestAPI(), axongo.NewInMemoryAddressSigner(axongo.AddressKeys{Address: addr, Keys: ed25519.PrivateKey(keyPair.PrivateKey[:])})).
		AddInput(&builder.TxInput{
			UnlockTarget: addr,
			InputID:      tpkg.RandOutputID(0),
			Input:        output,
		}).
		AddOutput(output).
		AddCommitmentInput(&axongo.CommitmentInput{CommitmentID: axongo.NewCommitmentID(85, tpkg.Rand32ByteArray())}).
		Build()

	require.NoError(t, err)

	require.ErrorIs(t, createBlockAtSlotWithPayload(t, 100, 79, commitmentInputNewerThanBlockCommitment, apiProvider), axongo.ErrCommitmentInputNewerThanCommitment)

	commitmentCorrect, err := builder.NewTransactionBuilder(apiProvider.LatestAPI(), axongo.NewInMemoryAddressSigner(axongo.AddressKeys{Address: addr, Keys: ed25519.PrivateKey(keyPair.PrivateKey[:])})).
		AddInput(&builder.TxInput{
			UnlockTarget: addr,
			InputID:      tpkg.RandOutputID(0),
			Input:        output,
		}).
		AddOutput(output).
		AddCommitmentInput(&axongo.CommitmentInput{CommitmentID: axongo.NewCommitmentID(79, tpkg.Rand32ByteArray())}).
		Build()

	require.NoError(t, err)

	require.NoError(t, createBlockAtSlotWithPayload(t, 100, 89, commitmentCorrect, apiProvider))

	commitmentCorrectOldest, err := builder.NewTransactionBuilder(apiProvider.LatestAPI(), axongo.NewInMemoryAddressSigner(axongo.AddressKeys{Address: addr, Keys: ed25519.PrivateKey(keyPair.PrivateKey[:])})).
		AddInput(&builder.TxInput{
			UnlockTarget: addr,
			InputID:      tpkg.RandOutputID(0),
			Input:        output,
		}).
		AddOutput(output).
		AddCommitmentInput(&axongo.CommitmentInput{CommitmentID: axongo.NewCommitmentID(79, tpkg.Rand32ByteArray())}).
		Build()

	require.NoError(t, err)

	require.NoError(t, createBlockAtSlotWithPayload(t, 100, 79, commitmentCorrectOldest, apiProvider))

	commitmentCorrectNewest, err := builder.NewTransactionBuilder(apiProvider.LatestAPI(), axongo.NewInMemoryAddressSigner(axongo.AddressKeys{Address: addr, Keys: ed25519.PrivateKey(keyPair.PrivateKey[:])})).
		AddInput(&builder.TxInput{
			UnlockTarget: addr,
			InputID:      tpkg.RandOutputID(0),
			Input:        output,
		}).
		AddOutput(output).
		AddCommitmentInput(&axongo.CommitmentInput{CommitmentID: axongo.NewCommitmentID(89, tpkg.Rand32ByteArray())}).
		Build()

	require.NoError(t, err)

	require.NoError(t, createBlockAtSlotWithPayload(t, 100, 89, commitmentCorrectNewest, apiProvider))

	commitmentCorrectMiddle, err := builder.NewTransactionBuilder(apiProvider.LatestAPI(), axongo.NewInMemoryAddressSigner(axongo.AddressKeys{Address: addr, Keys: ed25519.PrivateKey(keyPair.PrivateKey[:])})).
		AddInput(&builder.TxInput{
			UnlockTarget: addr,
			InputID:      tpkg.RandOutputID(0),
			Input:        output,
		}).
		AddOutput(output).
		AddCommitmentInput(&axongo.CommitmentInput{CommitmentID: axongo.NewCommitmentID(85, tpkg.Rand32ByteArray())}).
		Build()

	require.NoError(t, err)

	require.NoError(t, createBlockAtSlotWithPayload(t, 100, 89, commitmentCorrectMiddle, apiProvider))
}

func TestBlock_DeserializationNotEnoughData(t *testing.T) {
	blockBytes := []byte{byte(tpkg.ZeroCostTestAPI.Version()), 1}

	block := &axongo.Block{}
	_, err := tpkg.ZeroCostTestAPI.Decode(blockBytes, block)
	require.ErrorIs(t, err, serializer.ErrDeserializationNotEnoughData)
}

func TestBasicBlock_MinSize(t *testing.T) {
	minBlock := &axongo.Block{
		API: tpkg.ZeroCostTestAPI,
		Header: axongo.BlockHeader{
			ProtocolVersion:  tpkg.ZeroCostTestAPI.Version(),
			NetworkID:        tpkg.ZeroCostTestAPI.ProtocolParameters().NetworkID(),
			IssuingTime:      tpkg.RandUTCTime(),
			SlotCommitmentID: axongo.NewEmptyCommitment(tpkg.ZeroCostTestAPI).MustID(),
		},
		Signature: tpkg.RandEd25519Signature(),
		Body: &axongo.BasicBlockBody{
			API:                tpkg.ZeroCostTestAPI,
			StrongParents:      tpkg.SortedRandBlockIDs(1),
			WeakParents:        axongo.BlockIDs{},
			ShallowLikeParents: axongo.BlockIDs{},
			Payload:            nil,
		},
	}

	blockBytes, err := tpkg.ZeroCostTestAPI.Encode(minBlock)
	require.NoError(t, err)

	block2 := &axongo.Block{}
	consumedBytes, err := tpkg.ZeroCostTestAPI.Decode(blockBytes, block2, serix.WithValidation())
	require.NoError(t, err)
	require.Equal(t, minBlock, block2)
	require.Equal(t, len(blockBytes), consumedBytes)
}

func TestValidationBlock_MinSize(t *testing.T) {
	minBlock := &axongo.Block{
		API: tpkg.ZeroCostTestAPI,
		Header: axongo.BlockHeader{
			ProtocolVersion:  tpkg.ZeroCostTestAPI.Version(),
			NetworkID:        tpkg.ZeroCostTestAPI.ProtocolParameters().NetworkID(),
			IssuingTime:      tpkg.RandUTCTime(),
			SlotCommitmentID: axongo.NewEmptyCommitment(tpkg.ZeroCostTestAPI).MustID(),
		},
		Signature: tpkg.RandEd25519Signature(),
		Body: &axongo.ValidationBlockBody{
			API:                     tpkg.ZeroCostTestAPI,
			StrongParents:           tpkg.SortedRandBlockIDs(1),
			WeakParents:             axongo.BlockIDs{},
			ShallowLikeParents:      axongo.BlockIDs{},
			HighestSupportedVersion: tpkg.ZeroCostTestAPI.Version(),
		},
	}

	blockBytes, err := tpkg.ZeroCostTestAPI.Encode(minBlock)
	require.NoError(t, err)

	block2 := &axongo.Block{}
	consumedBytes, err := tpkg.ZeroCostTestAPI.Decode(blockBytes, block2, serix.WithValidation())
	require.NoError(t, err)
	require.Equal(t, minBlock, block2)
	require.Equal(t, len(blockBytes), consumedBytes)
}

func TestValidationBlock_HighestSupportedVersion(t *testing.T) {
	block := &axongo.Block{
		API: tpkg.ZeroCostTestAPI,
		Header: axongo.BlockHeader{
			ProtocolVersion:  tpkg.ZeroCostTestAPI.Version(),
			NetworkID:        tpkg.ZeroCostTestAPI.ProtocolParameters().NetworkID(),
			IssuingTime:      tpkg.RandUTCTime(),
			SlotCommitmentID: axongo.NewEmptyCommitment(tpkg.ZeroCostTestAPI).MustID(),
		},
		Signature: tpkg.RandEd25519Signature(),
	}

	// Invalid HighestSupportedVersion.
	{
		block.Body = &axongo.ValidationBlockBody{
			API:                     tpkg.ZeroCostTestAPI,
			StrongParents:           tpkg.SortedRandBlockIDs(1),
			WeakParents:             axongo.BlockIDs{},
			ShallowLikeParents:      axongo.BlockIDs{},
			HighestSupportedVersion: tpkg.ZeroCostTestAPI.Version() - 1,
		}
		blockBytes, err := tpkg.ZeroCostTestAPI.Encode(block)
		require.NoError(t, err)

		block2 := &axongo.Block{}
		_, err = tpkg.ZeroCostTestAPI.Decode(blockBytes, block2, serix.WithValidation())
		require.ErrorIs(t, err, axongo.ErrHighestSupportedVersionTooSmall)
	}

	// Valid HighestSupportedVersion.
	{
		block.Body = &axongo.ValidationBlockBody{
			API:                     tpkg.ZeroCostTestAPI,
			StrongParents:           tpkg.SortedRandBlockIDs(1),
			WeakParents:             axongo.BlockIDs{},
			ShallowLikeParents:      axongo.BlockIDs{},
			HighestSupportedVersion: tpkg.ZeroCostTestAPI.Version(),
		}
		blockBytes, err := tpkg.ZeroCostTestAPI.Encode(block)
		require.NoError(t, err)

		block2 := &axongo.Block{}
		consumedBytes, err := tpkg.ZeroCostTestAPI.Decode(blockBytes, block2, serix.WithValidation())
		require.NoError(t, err)
		require.Equal(t, block, block2)
		require.Equal(t, len(blockBytes), consumedBytes)
	}
}

func TestBlockJSONMarshalling(t *testing.T) {
	networkID := axongo.NetworkIDFromString("xxxNetwork")
	issuingTime := tpkg.RandUTCTime()
	commitmentID := axongo.NewEmptyCommitment(tpkg.ZeroCostTestAPI).MustID()
	issuerID := tpkg.RandAccountID()
	signature := tpkg.RandEd25519Signature()
	strongParents := tpkg.SortedRandBlockIDs(1)
	validationBlock := &axongo.Block{
		API: tpkg.ZeroCostTestAPI,
		Header: axongo.BlockHeader{
			ProtocolVersion:  tpkg.ZeroCostTestAPI.Version(),
			IssuingTime:      issuingTime,
			IssuerID:         issuerID,
			NetworkID:        networkID,
			SlotCommitmentID: commitmentID,
		},
		Body: &axongo.ValidationBlockBody{
			API:                     tpkg.ZeroCostTestAPI,
			StrongParents:           strongParents,
			HighestSupportedVersion: tpkg.ZeroCostTestAPI.Version(),
		},
		Signature: signature,
	}

	blockJSON := fmt.Sprintf(`{"header":{"protocolVersion":%d,"networkId":"%d","issuingTime":"%s","slotCommitmentId":"%s","latestFinalizedSlot":0,"issuerId":"%s"},"body":{"type":%d,"strongParents":["%s"],"highestSupportedVersion":%d,"protocolParametersHash":"0x0000000000000000000000000000000000000000000000000000000000000000"},"signature":{"type":%d,"publicKey":"%s","signature":"%s"}}`,
		tpkg.ZeroCostTestAPI.Version(),
		networkID,
		strconv.FormatUint(serializer.TimeToUint64(issuingTime), 10),
		commitmentID.ToHex(),
		issuerID.ToHex(),
		axongo.BlockBodyTypeValidation,
		strongParents[0].ToHex(),
		tpkg.ZeroCostTestAPI.Version(),
		axongo.SignatureEd25519,
		hexutil.EncodeHex(signature.PublicKey[:]),
		hexutil.EncodeHex(signature.Signature[:]),
	)

	jsonEncode, err := tpkg.ZeroCostTestAPI.JSONEncode(validationBlock)

	fmt.Println(string(jsonEncode))

	require.NoError(t, err)
	require.Equal(t, blockJSON, string(jsonEncode))
}

// TODO: add tests
//  - parents parameters basic block
//  - parents parameters validator block

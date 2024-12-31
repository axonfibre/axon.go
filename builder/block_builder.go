package builder

import (
	"crypto/ed25519"
	"time"

	"github.com/axonfibre/fibre.go/ierrors"
	axongo "github.com/axonfibre/axon.go/v4"
)

// NewBasicBlockBuilder creates a new BasicBlockBuilder.
func NewBasicBlockBuilder(api axongo.API) *BasicBlockBuilder {
	// TODO: burn the correct amount of Mana in all cases according to block work and RMC with issue #285
	basicBlock := &axongo.BasicBlockBody{
		API:                api,
		StrongParents:      axongo.BlockIDs{},
		WeakParents:        axongo.BlockIDs{},
		ShallowLikeParents: axongo.BlockIDs{},
	}

	protocolBlock := &axongo.Block{
		API: api,
		Header: axongo.BlockHeader{
			ProtocolVersion:  api.ProtocolParameters().Version(),
			SlotCommitmentID: axongo.EmptyCommitmentID,
			NetworkID:        api.ProtocolParameters().NetworkID(),
			IssuingTime:      time.Now().UTC(),
		},
		Signature: &axongo.Ed25519Signature{},
		Body:      basicBlock,
	}

	return &BasicBlockBuilder{
		protocolBlock: protocolBlock,
		basicBlock:    basicBlock,
	}
}

// BasicBlockBuilder is used to easily build up a Basic Block.
type BasicBlockBuilder struct {
	basicBlock *axongo.BasicBlockBody

	protocolBlock *axongo.Block
	err           error
}

// Build builds the Block or returns any error which occurred during the build steps.
func (b *BasicBlockBuilder) Build() (*axongo.Block, error) {
	b.basicBlock.ShallowLikeParents.Sort()
	b.basicBlock.WeakParents.Sort()
	b.basicBlock.StrongParents.Sort()

	if b.err != nil {
		return nil, b.err
	}

	return b.protocolBlock, nil
}

// ProtocolVersion sets the protocol version.
func (b *BasicBlockBuilder) ProtocolVersion(version axongo.Version) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.protocolBlock.Header.ProtocolVersion = version

	return b
}

func (b *BasicBlockBuilder) IssuingTime(time time.Time) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.protocolBlock.Header.IssuingTime = time.UTC()

	return b
}

// SlotCommitmentID sets the slot commitment.
func (b *BasicBlockBuilder) SlotCommitmentID(commitment axongo.CommitmentID) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.protocolBlock.Header.SlotCommitmentID = commitment

	return b
}

// LatestFinalizedSlot sets the latest finalized slot.
func (b *BasicBlockBuilder) LatestFinalizedSlot(slot axongo.SlotIndex) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.protocolBlock.Header.LatestFinalizedSlot = slot

	return b
}

func (b *BasicBlockBuilder) Sign(accountID axongo.AccountID, privKey ed25519.PrivateKey) *BasicBlockBuilder {
	//nolint:forcetypeassert // we can safely assume that this is an ed25519.PublicKey
	pubKey := privKey.Public().(ed25519.PublicKey)
	ed25519Address := axongo.Ed25519AddressFromPubKey(pubKey)

	signer := axongo.NewInMemoryAddressSigner(
		axongo.NewAddressKeysForEd25519Address(ed25519Address, privKey),
	)

	return b.SignWithSigner(accountID, signer, ed25519Address)
}

func (b *BasicBlockBuilder) SignWithSigner(accountID axongo.AccountID, signer axongo.AddressSigner, addr axongo.Address) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.protocolBlock.Header.IssuerID = accountID

	signature, err := b.protocolBlock.Sign(signer, addr)
	if err != nil {
		b.err = ierrors.Wrap(err, "failed to sign basic block")

		return b
	}

	edSig, isEdSig := signature.(*axongo.Ed25519Signature)
	if !isEdSig {
		panic("unsupported signature type")
	}

	b.protocolBlock.Signature = edSig

	return b
}

// StrongParents sets the strong parents.
func (b *BasicBlockBuilder) StrongParents(parents axongo.BlockIDs) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.basicBlock.StrongParents = parents.RemoveDupsAndSort()

	return b
}

// WeakParents sets the weak parents.
func (b *BasicBlockBuilder) WeakParents(parents axongo.BlockIDs) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.basicBlock.WeakParents = parents.RemoveDupsAndSort()

	return b
}

// ShallowLikeParents sets the shallow like parents.
func (b *BasicBlockBuilder) ShallowLikeParents(parents axongo.BlockIDs) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.basicBlock.ShallowLikeParents = parents.RemoveDupsAndSort()

	return b
}

// Payload sets the payload.
func (b *BasicBlockBuilder) Payload(payload axongo.ApplicationPayload) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.basicBlock.Payload = payload

	return b
}

// MaxBurnedMana sets the maximum amount of mana allowed to be burned by the block.
func (b *BasicBlockBuilder) MaxBurnedMana(maxBurnedMana axongo.Mana) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.basicBlock.MaxBurnedMana = maxBurnedMana

	return b
}

// CalculateAndSetMaxBurnedMana sets the maximum amount of mana allowed to be burned by the block based on the provided reference mana cost.
func (b *BasicBlockBuilder) CalculateAndSetMaxBurnedMana(rmc axongo.Mana) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	burnedMana, err := b.protocolBlock.ManaCost(rmc)
	if err != nil {
		b.err = ierrors.Wrap(err, "failed to calculate mana cost")
		return b
	}

	b.basicBlock.MaxBurnedMana = burnedMana

	return b
}

// NewValidationBlockBuilder creates a new ValidationBlockBuilder.
func NewValidationBlockBuilder(api axongo.API) *ValidationBlockBuilder {
	validationBlock := &axongo.ValidationBlockBody{
		API:                api,
		StrongParents:      axongo.BlockIDs{},
		WeakParents:        axongo.BlockIDs{},
		ShallowLikeParents: axongo.BlockIDs{},
	}

	protocolBlock := &axongo.Block{
		API: api,
		Header: axongo.BlockHeader{
			ProtocolVersion:  api.ProtocolParameters().Version(),
			SlotCommitmentID: axongo.NewEmptyCommitment(api).MustID(),
			NetworkID:        api.ProtocolParameters().NetworkID(),
			IssuingTime:      time.Now().UTC(),
		},
		Signature: &axongo.Ed25519Signature{},
		Body:      validationBlock,
	}

	return &ValidationBlockBuilder{
		protocolBlock:   protocolBlock,
		validationBlock: validationBlock,
	}
}

// ValidationBlockBuilder is used to easily build up a Validation Block.
type ValidationBlockBuilder struct {
	validationBlock *axongo.ValidationBlockBody

	protocolBlock *axongo.Block
	err           error
}

// Build builds the Block or returns any error which occurred during the build steps.
func (v *ValidationBlockBuilder) Build() (*axongo.Block, error) {
	v.validationBlock.ShallowLikeParents.Sort()
	v.validationBlock.WeakParents.Sort()
	v.validationBlock.StrongParents.Sort()

	if v.err != nil {
		return nil, v.err
	}

	return v.protocolBlock, nil
}

// ProtocolVersion sets the protocol version.
func (v *ValidationBlockBuilder) ProtocolVersion(version axongo.Version) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.protocolBlock.Header.ProtocolVersion = version

	return v
}

func (v *ValidationBlockBuilder) IssuingTime(time time.Time) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.protocolBlock.Header.IssuingTime = time.UTC()

	return v
}

// SlotCommitmentID sets the slot commitment.
func (v *ValidationBlockBuilder) SlotCommitmentID(commitmentID axongo.CommitmentID) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.protocolBlock.Header.SlotCommitmentID = commitmentID

	return v
}

// LatestFinalizedSlot sets the latest finalized slot.
func (v *ValidationBlockBuilder) LatestFinalizedSlot(slot axongo.SlotIndex) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.protocolBlock.Header.LatestFinalizedSlot = slot

	return v
}

func (v *ValidationBlockBuilder) Sign(accountID axongo.AccountID, privKey ed25519.PrivateKey) *ValidationBlockBuilder {
	//nolint:forcetypeassert // we can safely assume that this is an ed25519.PublicKey
	pubKey := privKey.Public().(ed25519.PublicKey)
	ed25519Address := axongo.Ed25519AddressFromPubKey(pubKey)

	signer := axongo.NewInMemoryAddressSigner(
		axongo.NewAddressKeysForEd25519Address(ed25519Address, privKey),
	)

	return v.SignWithSigner(accountID, signer, ed25519Address)
}

func (v *ValidationBlockBuilder) SignWithSigner(accountID axongo.AccountID, signer axongo.AddressSigner, addr axongo.Address) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.protocolBlock.Header.IssuerID = accountID

	signature, err := v.protocolBlock.Sign(signer, addr)
	if err != nil {
		v.err = ierrors.Wrap(err, "failed to sign validation block")

		return v
	}

	edSig, isEdSig := signature.(*axongo.Ed25519Signature)
	if !isEdSig {
		panic("unsupported signature type")
	}

	v.protocolBlock.Signature = edSig

	return v
}

// StrongParents sets the strong parents.
func (v *ValidationBlockBuilder) StrongParents(parents axongo.BlockIDs) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.validationBlock.StrongParents = parents.RemoveDupsAndSort()

	return v
}

// WeakParents sets the weak parents.
func (v *ValidationBlockBuilder) WeakParents(parents axongo.BlockIDs) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.validationBlock.WeakParents = parents.RemoveDupsAndSort()

	return v
}

// ShallowLikeParents sets the shallow like parents.
func (v *ValidationBlockBuilder) ShallowLikeParents(parents axongo.BlockIDs) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.validationBlock.ShallowLikeParents = parents.RemoveDupsAndSort()

	return v
}

// HighestSupportedVersion sets the highest supported version.
func (v *ValidationBlockBuilder) HighestSupportedVersion(highestSupportedVersion axongo.Version) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.validationBlock.HighestSupportedVersion = highestSupportedVersion

	return v
}

// ProtocolParametersHash sets the ProtocolParametersHash of the highest supported version.
func (v *ValidationBlockBuilder) ProtocolParametersHash(hash axongo.Identifier) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.validationBlock.ProtocolParametersHash = hash

	return v
}

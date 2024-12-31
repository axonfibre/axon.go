package builder

import (
	"github.com/axonfibre/fibre.go/ierrors"
	axongo "github.com/axonfibre/axon.go/v4"
)

// NewNFTOutputBuilder creates a new NFTOutputBuilder with the address and base token amount.
func NewNFTOutputBuilder(targetAddr axongo.Address, amount axongo.BaseToken) *NFTOutputBuilder {
	return &NFTOutputBuilder{output: &axongo.NFTOutput{
		Amount: amount,
		Mana:   0,
		NFTID:  axongo.EmptyNFTID(),
		UnlockConditions: axongo.NFTOutputUnlockConditions{
			&axongo.AddressUnlockCondition{Address: targetAddr},
		},
		Features:          axongo.NFTOutputFeatures{},
		ImmutableFeatures: axongo.NFTOutputImmFeatures{},
	}}
}

// NewNFTOutputBuilderFromPrevious creates a new NFTOutputBuilder starting from a copy of the previous axongo.NFTOutput.
func NewNFTOutputBuilderFromPrevious(previous *axongo.NFTOutput) *NFTOutputBuilder {
	return &NFTOutputBuilder{
		prev: previous,
		//nolint:forcetypeassert // we can safely assume that this is a NFTOutput
		output: previous.Clone().(*axongo.NFTOutput),
	}
}

// NFTOutputBuilder builds an axongo.NFTOutput.
type NFTOutputBuilder struct {
	prev   *axongo.NFTOutput
	output *axongo.NFTOutput
}

// Amount sets the base token amount of the output.
func (builder *NFTOutputBuilder) Amount(amount axongo.BaseToken) *NFTOutputBuilder {
	builder.output.Amount = amount

	return builder
}

// Amount sets the mana of the output.
func (builder *NFTOutputBuilder) Mana(mana axongo.Mana) *NFTOutputBuilder {
	builder.output.Mana = mana

	return builder
}

// NFTID sets the axongo.NFTID of this output.
// Do not call this function if the underlying axongo.NFTID is not new.
func (builder *NFTOutputBuilder) NFTID(nftID axongo.NFTID) *NFTOutputBuilder {
	builder.output.NFTID = nftID

	return builder
}

// Address sets/modifies an axongo.AddressUnlockCondition on the output.
func (builder *NFTOutputBuilder) Address(addr axongo.Address) *NFTOutputBuilder {
	builder.output.UnlockConditions.Upsert(&axongo.AddressUnlockCondition{Address: addr})

	return builder
}

// StorageDepositReturn sets/modifies an axongo.StorageDepositReturnUnlockCondition on the output.
func (builder *NFTOutputBuilder) StorageDepositReturn(returnAddr axongo.Address, amount axongo.BaseToken) *NFTOutputBuilder {
	builder.output.UnlockConditions.Upsert(&axongo.StorageDepositReturnUnlockCondition{ReturnAddress: returnAddr, Amount: amount})

	return builder
}

// Timelock sets/modifies an axongo.TimelockUnlockCondition on the output.
func (builder *NFTOutputBuilder) Timelock(untilSlot axongo.SlotIndex) *NFTOutputBuilder {
	builder.output.UnlockConditions.Upsert(&axongo.TimelockUnlockCondition{Slot: untilSlot})

	return builder
}

// Expiration sets/modifies an axongo.ExpirationUnlockCondition on the output.
func (builder *NFTOutputBuilder) Expiration(returnAddr axongo.Address, expiredAfterSlot axongo.SlotIndex) *NFTOutputBuilder {
	builder.output.UnlockConditions.Upsert(&axongo.ExpirationUnlockCondition{ReturnAddress: returnAddr, Slot: expiredAfterSlot})

	return builder
}

// Sender sets/modifies an axongo.SenderFeature on the output.
func (builder *NFTOutputBuilder) Sender(senderAddr axongo.Address) *NFTOutputBuilder {
	builder.output.Features.Upsert(&axongo.SenderFeature{Address: senderAddr})

	return builder
}

// Metadata sets/modifies an axongo.MetadataFeature on the output.
func (builder *NFTOutputBuilder) Metadata(entries axongo.MetadataFeatureEntries) *NFTOutputBuilder {
	builder.output.Features.Upsert(&axongo.MetadataFeature{Entries: entries})

	return builder
}

// Tag sets/modifies an axongo.TagFeature on the output.
func (builder *NFTOutputBuilder) Tag(tag []byte) *NFTOutputBuilder {
	builder.output.Features.Upsert(&axongo.TagFeature{Tag: tag})

	return builder
}

// ImmutableIssuer sets/modifies an axongo.IssuerFeature as an immutable feature on the output.
// Only call this function on a new axongo.NFTOutput.
func (builder *NFTOutputBuilder) ImmutableIssuer(issuer axongo.Address) *NFTOutputBuilder {
	builder.output.ImmutableFeatures.Upsert(&axongo.IssuerFeature{Address: issuer})

	return builder
}

// ImmutableMetadata sets/modifies an axongo.MetadataFeature as an immutable feature on the output.
// Only call this function on a new axongo.NFTOutput.
func (builder *NFTOutputBuilder) ImmutableMetadata(entries axongo.MetadataFeatureEntries) *NFTOutputBuilder {
	builder.output.ImmutableFeatures.Upsert(&axongo.MetadataFeature{Entries: entries})

	return builder
}

// Build builds the axongo.FoundryOutput.
func (builder *NFTOutputBuilder) Build() (*axongo.NFTOutput, error) {
	if builder.prev != nil {
		if !builder.prev.ImmutableFeatures.Equal(builder.output.ImmutableFeatures) {
			return nil, ierrors.New("immutable features are not allowed to be changed")
		}
	}

	builder.output.UnlockConditions.Sort()
	builder.output.Features.Sort()
	builder.output.ImmutableFeatures.Sort()

	return builder.output, nil
}

// MustBuild works like Build() but panics if an error is encountered.
func (builder *NFTOutputBuilder) MustBuild() *axongo.NFTOutput {
	output, err := builder.Build()
	if err != nil {
		panic(err)
	}

	return output
}

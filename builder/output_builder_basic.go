package builder

import (
	axongo "github.com/axonfibre/axon.go/v4"
)

// NewBasicOutputBuilder creates a new BasicOutputBuilder with the required target address and base token amount.
func NewBasicOutputBuilder(targetAddr axongo.Address, amount axongo.BaseToken) *BasicOutputBuilder {
	return &BasicOutputBuilder{output: &axongo.BasicOutput{
		Amount: amount,
		Mana:   0,
		UnlockConditions: axongo.BasicOutputUnlockConditions{
			&axongo.AddressUnlockCondition{Address: targetAddr},
		},
		Features: axongo.BasicOutputFeatures{},
	}}
}

// NewBasicOutputBuilderFromPrevious creates a new BasicOutputBuilder starting from a copy of the previous axongo.BasicOutput.
func NewBasicOutputBuilderFromPrevious(previous *axongo.BasicOutput) *BasicOutputBuilder {
	//nolint:forcetypeassert // we can safely assume that this is a BasicOutput
	return &BasicOutputBuilder{output: previous.Clone().(*axongo.BasicOutput)}
}

// BasicOutputBuilder builds an axongo.BasicOutput.
type BasicOutputBuilder struct {
	output *axongo.BasicOutput
}

// Amount sets the base token amount of the output.
func (builder *BasicOutputBuilder) Amount(amount axongo.BaseToken) *BasicOutputBuilder {
	builder.output.Amount = amount

	return builder
}

// Mana sets the mana of the output.
func (builder *BasicOutputBuilder) Mana(mana axongo.Mana) *BasicOutputBuilder {
	builder.output.Mana = mana

	return builder
}

// Address sets/modifies an axongo.AddressUnlockCondition on the output.
func (builder *BasicOutputBuilder) Address(addr axongo.Address) *BasicOutputBuilder {
	builder.output.UnlockConditions.Upsert(&axongo.AddressUnlockCondition{Address: addr})

	return builder
}

// StorageDepositReturn sets/modifies an axongo.StorageDepositReturnUnlockCondition on the output.
func (builder *BasicOutputBuilder) StorageDepositReturn(returnAddr axongo.Address, amount axongo.BaseToken) *BasicOutputBuilder {
	builder.output.UnlockConditions.Upsert(&axongo.StorageDepositReturnUnlockCondition{ReturnAddress: returnAddr, Amount: amount})

	return builder
}

// Timelock sets/modifies an axongo.TimelockUnlockCondition on the output.
func (builder *BasicOutputBuilder) Timelock(untilSlot axongo.SlotIndex) *BasicOutputBuilder {
	builder.output.UnlockConditions.Upsert(&axongo.TimelockUnlockCondition{Slot: untilSlot})

	return builder
}

// Expiration sets/modifies an axongo.ExpirationUnlockCondition on the output.
func (builder *BasicOutputBuilder) Expiration(returnAddr axongo.Address, expiredAfterSlot axongo.SlotIndex) *BasicOutputBuilder {
	builder.output.UnlockConditions.Upsert(&axongo.ExpirationUnlockCondition{ReturnAddress: returnAddr, Slot: expiredAfterSlot})

	return builder
}

// Sender sets/modifies an axongo.SenderFeature on the output.
func (builder *BasicOutputBuilder) Sender(senderAddr axongo.Address) *BasicOutputBuilder {
	builder.output.Features.Upsert(&axongo.SenderFeature{Address: senderAddr})

	return builder
}

// Metadata sets/modifies an axongo.MetadataFeature on the output.
func (builder *BasicOutputBuilder) Metadata(entries axongo.MetadataFeatureEntries) *BasicOutputBuilder {
	builder.output.Features.Upsert(&axongo.MetadataFeature{Entries: entries})

	return builder
}

// Tag sets/modifies an axongo.TagFeature on the output.
func (builder *BasicOutputBuilder) Tag(tag []byte) *BasicOutputBuilder {
	builder.output.Features.Upsert(&axongo.TagFeature{Tag: tag})

	return builder
}

// NativeToken adds/modifies a native token to/on the output.
func (builder *BasicOutputBuilder) NativeToken(nt *axongo.NativeTokenFeature) *BasicOutputBuilder {
	builder.output.Features.Upsert(nt)

	return builder
}

// Build builds the axongo.BasicOutput.
func (builder *BasicOutputBuilder) Build() (*axongo.BasicOutput, error) {
	builder.output.UnlockConditions.Sort()
	builder.output.Features.Sort()

	return builder.output, nil
}

// MustBuild works like Build() but panics if an error is encountered.
func (builder *BasicOutputBuilder) MustBuild() *axongo.BasicOutput {
	output, err := builder.Build()
	if err != nil {
		panic(err)
	}

	return output
}

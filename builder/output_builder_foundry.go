package builder

import (
	"github.com/axonfibre/fibre.go/ierrors"
	axongo "github.com/axonfibre/axon.go/v4"
)

// NewFoundryOutputBuilder creates a new FoundryOutputBuilder with the account address, serial number, token scheme and base token amount.
func NewFoundryOutputBuilder(accountAddr *axongo.AccountAddress, amount axongo.BaseToken, serialNumber uint32, tokenScheme axongo.TokenScheme) *FoundryOutputBuilder {
	return &FoundryOutputBuilder{output: &axongo.FoundryOutput{
		Amount:       amount,
		SerialNumber: serialNumber,
		TokenScheme:  tokenScheme,
		UnlockConditions: axongo.FoundryOutputUnlockConditions{
			&axongo.ImmutableAccountUnlockCondition{Address: accountAddr},
		},
		Features:          axongo.FoundryOutputFeatures{},
		ImmutableFeatures: axongo.FoundryOutputImmFeatures{},
	}}
}

// NewFoundryOutputBuilderFromPrevious creates a new FoundryOutputBuilder starting from a copy of the previous axongo.FoundryOutput.
func NewFoundryOutputBuilderFromPrevious(previous *axongo.FoundryOutput) *FoundryOutputBuilder {
	return &FoundryOutputBuilder{
		prev: previous,
		//nolint:forcetypeassert // we can safely assume that this is a FoundryOutput
		output: previous.Clone().(*axongo.FoundryOutput),
	}
}

// FoundryOutputBuilder builds an axongo.FoundryOutput.
type FoundryOutputBuilder struct {
	prev   *axongo.FoundryOutput
	output *axongo.FoundryOutput
}

// Amount sets the base token amount of the output.
func (builder *FoundryOutputBuilder) Amount(amount axongo.BaseToken) *FoundryOutputBuilder {
	builder.output.Amount = amount

	return builder
}

// Metadata sets/modifies an axongo.MetadataFeature on the output.
func (builder *FoundryOutputBuilder) Metadata(entries axongo.MetadataFeatureEntries) *FoundryOutputBuilder {
	builder.output.Features.Upsert(&axongo.MetadataFeature{Entries: entries})

	return builder
}

// NativeToken adds/modifies a native token to/on the output.
func (builder *FoundryOutputBuilder) NativeToken(nt *axongo.NativeTokenFeature) *FoundryOutputBuilder {
	builder.output.Features.Upsert(nt)

	return builder
}

// ImmutableMetadata sets/modifies an axongo.MetadataFeature as an immutable feature on the output.
// Only call this function on a new axongo.FoundryOutput.
func (builder *FoundryOutputBuilder) ImmutableMetadata(entries axongo.MetadataFeatureEntries) *FoundryOutputBuilder {
	builder.output.ImmutableFeatures.Upsert(&axongo.MetadataFeature{Entries: entries})

	return builder
}

func (builder *FoundryOutputBuilder) TokenScheme(tokenScheme axongo.TokenScheme) *FoundryOutputBuilder {
	builder.output.TokenScheme = tokenScheme

	return builder
}

// Build builds the axongo.FoundryOutput.
func (builder *FoundryOutputBuilder) Build() (*axongo.FoundryOutput, error) {
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
func (builder *FoundryOutputBuilder) MustBuild() *axongo.FoundryOutput {
	output, err := builder.Build()

	if err != nil {
		panic(err)
	}

	return output
}

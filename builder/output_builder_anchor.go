package builder

import (
	"github.com/axonfibre/fibre.go/ierrors"
	axongo "github.com/axonfibre/axon.go/v4"
)

// NewAnchorOutputBuilder creates a new AnchorOutputBuilder with the required state controller/governor addresses and base token amount.
func NewAnchorOutputBuilder(stateCtrl axongo.Address, govAddr axongo.Address, amount axongo.BaseToken) *AnchorOutputBuilder {
	return &AnchorOutputBuilder{output: &axongo.AnchorOutput{
		Amount:     amount,
		Mana:       0,
		AnchorID:   axongo.EmptyAnchorID,
		StateIndex: 0,
		UnlockConditions: axongo.AnchorOutputUnlockConditions{
			&axongo.StateControllerAddressUnlockCondition{Address: stateCtrl},
			&axongo.GovernorAddressUnlockCondition{Address: govAddr},
		},
		Features:          axongo.AnchorOutputFeatures{},
		ImmutableFeatures: axongo.AnchorOutputImmFeatures{},
	}}
}

// NewAnchorOutputBuilderFromPrevious creates a new AnchorOutputBuilder starting from a copy of the previous axongo.AnchorOutput.
func NewAnchorOutputBuilderFromPrevious(previous *axongo.AnchorOutput) *AnchorOutputBuilder {
	return &AnchorOutputBuilder{
		prev: previous,
		//nolint:forcetypeassert // we can safely assume that this is an AnchorOutput
		output: previous.Clone().(*axongo.AnchorOutput),
	}
}

// AnchorOutputBuilder builds an axongo.AnchorOutput.
type AnchorOutputBuilder struct {
	prev         *axongo.AnchorOutput
	output       *axongo.AnchorOutput
	stateCtrlReq bool
	govCtrlReq   bool
}

// Amount sets the base token amount of the output.
func (builder *AnchorOutputBuilder) Amount(amount axongo.BaseToken) *AnchorOutputBuilder {
	builder.output.Amount = amount
	builder.stateCtrlReq = true

	return builder
}

// Mana sets the mana of the output.
func (builder *AnchorOutputBuilder) Mana(mana axongo.Mana) *AnchorOutputBuilder {
	builder.output.Mana = mana

	return builder
}

// AnchorID sets the axongo.AnchorID of this output.
// Do not call this function if the underlying axongo.AnchorOutput is not new.
func (builder *AnchorOutputBuilder) AnchorID(anchorID axongo.AnchorID) *AnchorOutputBuilder {
	builder.output.AnchorID = anchorID

	return builder
}

// StateController sets the axongo.StateControllerAddressUnlockCondition of the output.
func (builder *AnchorOutputBuilder) StateController(stateCtrl axongo.Address) *AnchorOutputBuilder {
	builder.output.UnlockConditions.Upsert(&axongo.StateControllerAddressUnlockCondition{Address: stateCtrl})
	builder.govCtrlReq = true

	return builder
}

// Governor sets the axongo.GovernorAddressUnlockCondition of the output.
func (builder *AnchorOutputBuilder) Governor(governor axongo.Address) *AnchorOutputBuilder {
	builder.output.UnlockConditions.Upsert(&axongo.GovernorAddressUnlockCondition{Address: governor})
	builder.govCtrlReq = true

	return builder
}

// Metadata sets/modifies an axongo.MetadataFeature on the output.
func (builder *AnchorOutputBuilder) Metadata(entries axongo.MetadataFeatureEntries) *AnchorOutputBuilder {
	builder.output.Features.Upsert(&axongo.MetadataFeature{Entries: entries})
	builder.govCtrlReq = true

	return builder
}

// StateMetadata sets/modifies an axongo.StateMetadataFeature on the output.
func (builder *AnchorOutputBuilder) StateMetadata(entries axongo.StateMetadataFeatureEntries) *AnchorOutputBuilder {
	builder.output.Features.Upsert(&axongo.StateMetadataFeature{Entries: entries})
	builder.stateCtrlReq = true

	return builder
}

// ImmutableIssuer sets/modifies an axongo.IssuerFeature as an immutable feature on the output.
// Only call this function on a new axongo.AnchorOutput.
func (builder *AnchorOutputBuilder) ImmutableIssuer(issuer axongo.Address) *AnchorOutputBuilder {
	builder.output.ImmutableFeatures.Upsert(&axongo.IssuerFeature{Address: issuer})

	return builder
}

// ImmutableMetadata sets/modifies an axongo.MetadataFeature as an immutable feature on the output.
// Only call this function on a new axongo.AnchorOutput.
func (builder *AnchorOutputBuilder) ImmutableMetadata(entries axongo.MetadataFeatureEntries) *AnchorOutputBuilder {
	builder.output.ImmutableFeatures.Upsert(&axongo.MetadataFeature{Entries: entries})

	return builder
}

// Build builds the axongo.AnchorOutput.
func (builder *AnchorOutputBuilder) Build() (*axongo.AnchorOutput, error) {
	if builder.prev != nil && builder.govCtrlReq && builder.stateCtrlReq {
		return nil, ierrors.New("builder calls require both state and governor transitions which is not possible")
	}

	if builder.prev != nil {
		if builder.stateCtrlReq {
			builder.output.StateIndex++
		}
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
func (builder *AnchorOutputBuilder) MustBuild() *axongo.AnchorOutput {
	output, err := builder.Build()
	if err != nil {
		panic(err)
	}

	return output
}

type AnchorStateTransition struct {
	builder *AnchorOutputBuilder
}

// StateTransition narrows the builder functions to the ones available for an anchor state transition.
func (builder *AnchorOutputBuilder) StateTransition() *AnchorStateTransition {
	return &AnchorStateTransition{builder: builder}
}

// Amount sets the base token amount of the output.
func (trans *AnchorStateTransition) Amount(amount axongo.BaseToken) *AnchorStateTransition {
	return trans.builder.Amount(amount).StateTransition()
}

// Mana sets the mana of the output.
func (trans *AnchorStateTransition) Mana(mana axongo.Mana) *AnchorStateTransition {
	return trans.builder.Mana(mana).StateTransition()
}

// StateMetadata sets/modifies an axongo.StateMetadataFeature on the output.
func (trans *AnchorStateTransition) StateMetadata(entries axongo.StateMetadataFeatureEntries) *AnchorStateTransition {
	return trans.builder.StateMetadata(entries).StateTransition()
}

// Builder returns the AnchorOutputBuilder.
func (trans *AnchorStateTransition) Builder() *AnchorOutputBuilder {
	return trans.builder
}

type AnchorGovernanceTransition struct {
	builder *AnchorOutputBuilder
}

// GovernanceTransition narrows the builder functions to the ones available for an anchor governance transition.
func (builder *AnchorOutputBuilder) GovernanceTransition() *AnchorGovernanceTransition {
	return &AnchorGovernanceTransition{builder: builder}
}

// StateController sets the axongo.StateControllerAddressUnlockCondition of the output.
func (trans *AnchorGovernanceTransition) StateController(stateCtrl axongo.Address) *AnchorGovernanceTransition {
	return trans.builder.StateController(stateCtrl).GovernanceTransition()
}

// Governor sets the axongo.GovernorAddressUnlockCondition of the output.
func (trans *AnchorGovernanceTransition) Governor(governor axongo.Address) *AnchorGovernanceTransition {
	return trans.builder.Governor(governor).GovernanceTransition()
}

// Metadata sets/modifies an axongo.MetadataFeature as a mutable feature on the output.
func (trans *AnchorGovernanceTransition) Metadata(entries axongo.MetadataFeatureEntries) *AnchorGovernanceTransition {
	return trans.builder.Metadata(entries).GovernanceTransition()
}

// Builder returns the AnchorOutputBuilder.
func (trans *AnchorGovernanceTransition) Builder() *AnchorOutputBuilder {
	return trans.builder
}

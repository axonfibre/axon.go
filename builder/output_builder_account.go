package builder

import (
	"github.com/axonfibre/fibre.go/ierrors"
	axongo "github.com/axonfibre/axon.go/v4"
)

// NewAccountOutputBuilder creates a new AccountOutputBuilder with the address and base token amount.
func NewAccountOutputBuilder(targetAddr axongo.Address, amount axongo.BaseToken) *AccountOutputBuilder {
	return &AccountOutputBuilder{output: &axongo.AccountOutput{
		Amount:         amount,
		Mana:           0,
		AccountID:      axongo.EmptyAccountID,
		FoundryCounter: 0,
		UnlockConditions: axongo.AccountOutputUnlockConditions{
			&axongo.AddressUnlockCondition{Address: targetAddr},
		},
		Features:          axongo.AccountOutputFeatures{},
		ImmutableFeatures: axongo.AccountOutputImmFeatures{},
	}}
}

// NewAccountOutputBuilderFromPrevious creates a new AccountOutputBuilder starting from a copy of the previous axongo.AccountOutput.
func NewAccountOutputBuilderFromPrevious(previous *axongo.AccountOutput) *AccountOutputBuilder {
	return &AccountOutputBuilder{
		prev: previous,
		//nolint:forcetypeassert // we can safely assume that this is an AccountOutput
		output: previous.Clone().(*axongo.AccountOutput),
	}
}

// AccountOutputBuilder builds an axongo.AccountOutput.
type AccountOutputBuilder struct {
	prev   *axongo.AccountOutput
	output *axongo.AccountOutput
}

// Amount sets the base token amount of the output.
func (builder *AccountOutputBuilder) Amount(amount axongo.BaseToken) *AccountOutputBuilder {
	builder.output.Amount = amount

	return builder
}

// Mana sets the mana of the output.
func (builder *AccountOutputBuilder) Mana(mana axongo.Mana) *AccountOutputBuilder {
	builder.output.Mana = mana

	return builder
}

// AccountID sets the axongo.AccountID of this output.
// Do not call this function if the underlying axongo.AccountOutput is not new.
func (builder *AccountOutputBuilder) AccountID(accountID axongo.AccountID) *AccountOutputBuilder {
	builder.output.AccountID = accountID

	return builder
}

// FoundriesToGenerate bumps the output's foundry counter by the amount of foundries to generate.
func (builder *AccountOutputBuilder) FoundriesToGenerate(count uint32) *AccountOutputBuilder {
	builder.output.FoundryCounter += count

	return builder
}

// Address sets/modifies an axongo.AddressUnlockCondition on the output.
func (builder *AccountOutputBuilder) Address(addr axongo.Address) *AccountOutputBuilder {
	builder.output.UnlockConditions.Upsert(&axongo.AddressUnlockCondition{Address: addr})

	return builder
}

// Sender sets/modifies an axongo.SenderFeature as a mutable feature on the output.
func (builder *AccountOutputBuilder) Sender(senderAddr axongo.Address) *AccountOutputBuilder {
	builder.output.Features.Upsert(&axongo.SenderFeature{Address: senderAddr})

	return builder
}

// Metadata sets/modifies an axongo.MetadataFeature on the output.
func (builder *AccountOutputBuilder) Metadata(entries axongo.MetadataFeatureEntries) *AccountOutputBuilder {
	builder.output.Features.Upsert(&axongo.MetadataFeature{Entries: entries})

	return builder
}

// BlockIssuer sets/modifies an axongo.BlockIssuerFeature as a mutable feature on the output.
func (builder *AccountOutputBuilder) BlockIssuer(keys axongo.BlockIssuerKeys, expirySlot axongo.SlotIndex) *AccountOutputBuilder {
	builder.output.Features.Upsert(&axongo.BlockIssuerFeature{
		BlockIssuerKeys: keys,
		ExpirySlot:      expirySlot,
	})

	return builder
}

// Staking sets/modifies an axongo.StakingFeature as a mutable feature on the output.
func (builder *AccountOutputBuilder) Staking(amount axongo.BaseToken, fixedCost axongo.Mana, startEpoch axongo.EpochIndex, optEndEpoch ...axongo.EpochIndex) *AccountOutputBuilder {
	endEpoch := axongo.MaxEpochIndex
	if len(optEndEpoch) > 0 {
		endEpoch = optEndEpoch[0]
	}

	builder.output.Features.Upsert(&axongo.StakingFeature{
		StakedAmount: amount,
		FixedCost:    fixedCost,
		StartEpoch:   startEpoch,
		EndEpoch:     endEpoch,
	})

	return builder
}

// ImmutableIssuer sets/modifies an axongo.IssuerFeature as an immutable feature on the output.
// Only call this function on a new axongo.AccountOutput.
func (builder *AccountOutputBuilder) ImmutableIssuer(issuer axongo.Address) *AccountOutputBuilder {
	builder.output.ImmutableFeatures.Upsert(&axongo.IssuerFeature{Address: issuer})

	return builder
}

// ImmutableMetadata sets/modifies an axongo.MetadataFeature as an immutable feature on the output.
// Only call this function on a new axongo.AccountOutput.
func (builder *AccountOutputBuilder) ImmutableMetadata(entries axongo.MetadataFeatureEntries) *AccountOutputBuilder {
	builder.output.ImmutableFeatures.Upsert(&axongo.MetadataFeature{Entries: entries})

	return builder
}

// Build builds the axongo.AccountOutput.
func (builder *AccountOutputBuilder) Build() (*axongo.AccountOutput, error) {
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
func (builder *AccountOutputBuilder) MustBuild() *axongo.AccountOutput {
	output, err := builder.Build()
	if err != nil {
		panic(err)
	}

	return output
}

// RemoveFeature removes a feature from the output.
func (builder *AccountOutputBuilder) RemoveFeature(featureType axongo.FeatureType) *AccountOutputBuilder {
	builder.output.Features.Remove(featureType)

	return builder
}

// BlockIssuerTransition narrows the builder functions to the ones available for an axongo.BlockIssuerFeature transition.
// If BlockIssuerFeature does not exist, it creates and sets an empty feature.
func (builder *AccountOutputBuilder) BlockIssuerTransition() *BlockIssuerTransition {
	blockIssuerFeature := builder.output.FeatureSet().BlockIssuer()
	if blockIssuerFeature == nil {
		blockIssuerFeature = &axongo.BlockIssuerFeature{
			BlockIssuerKeys: axongo.NewBlockIssuerKeys(),
			ExpirySlot:      0,
		}
	}

	return &BlockIssuerTransition{
		feature: blockIssuerFeature,
		builder: builder,
	}
}

// StakingTransition narrows the builder functions to the ones available for an axongo.StakingFeature transition.
// If StakingFeature does not exist, it creates and sets an empty feature.
func (builder *AccountOutputBuilder) StakingTransition() *StakingTransition {
	stakingFeature := builder.output.FeatureSet().Staking()
	if stakingFeature == nil {
		stakingFeature = &axongo.StakingFeature{
			StakedAmount: 0,
			FixedCost:    0,
			StartEpoch:   0,
			EndEpoch:     0,
		}
	}

	return &StakingTransition{
		feature: stakingFeature,
		builder: builder,
	}
}

type BlockIssuerTransition struct {
	feature *axongo.BlockIssuerFeature
	builder *AccountOutputBuilder
}

// AddKeys adds the keys of the BlockIssuerFeature.
func (trans *BlockIssuerTransition) AddKeys(keys ...axongo.BlockIssuerKey) *BlockIssuerTransition {
	for _, blockIssuerKey := range keys {
		trans.feature.BlockIssuerKeys.Add(blockIssuerKey)
	}

	return trans
}

// RemoveKey deletes the key of the axongo.BlockIssuerFeature.
func (trans *BlockIssuerTransition) RemoveKey(keyToDelete axongo.BlockIssuerKey) *BlockIssuerTransition {
	trans.feature.BlockIssuerKeys.Remove(keyToDelete)

	return trans
}

// Keys sets the keys of the axongo.BlockIssuerFeature.
func (trans *BlockIssuerTransition) Keys(keys axongo.BlockIssuerKeys) *BlockIssuerTransition {
	trans.feature.BlockIssuerKeys = keys

	return trans
}

// ExpirySlot sets the ExpirySlot of axongo.BlockIssuerFeature.
func (trans *BlockIssuerTransition) ExpirySlot(slot axongo.SlotIndex) *BlockIssuerTransition {
	trans.feature.ExpirySlot = slot

	return trans
}

// Builder returns the AccountOutputBuilder.
func (trans *BlockIssuerTransition) Builder() *AccountOutputBuilder {
	return trans.builder
}

type StakingTransition struct {
	feature *axongo.StakingFeature
	builder *AccountOutputBuilder
}

// StakedAmount sets the StakedAmount of axongo.StakingFeature.
func (trans *StakingTransition) StakedAmount(amount axongo.BaseToken) *StakingTransition {
	trans.feature.StakedAmount = amount

	return trans
}

// FixedCost sets the FixedCost of axongo.StakingFeature.
func (trans *StakingTransition) FixedCost(fixedCost axongo.Mana) *StakingTransition {
	trans.feature.FixedCost = fixedCost

	return trans
}

// StartEpoch sets the StartEpoch of axongo.StakingFeature.
func (trans *StakingTransition) StartEpoch(epoch axongo.EpochIndex) *StakingTransition {
	trans.feature.StartEpoch = epoch

	return trans
}

// EndEpoch sets the EndEpoch of axongo.StakingFeature.
func (trans *StakingTransition) EndEpoch(epoch axongo.EpochIndex) *StakingTransition {
	trans.feature.EndEpoch = epoch

	return trans
}

// Builder returns the AccountOutputBuilder.
func (trans *StakingTransition) Builder() *AccountOutputBuilder {
	return trans.builder
}

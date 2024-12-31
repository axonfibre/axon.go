package builder

import (
	axongo "github.com/axonfibre/axon.go/v4"
)

// NewDelegationOutputBuilder creates a new DelegationOutputBuilder with the account address, serial number, token scheme and base token amount.
func NewDelegationOutputBuilder(validatorAddress *axongo.AccountAddress, addr axongo.Address, amount axongo.BaseToken) *DelegationOutputBuilder {
	return &DelegationOutputBuilder{output: &axongo.DelegationOutput{
		Amount:           amount,
		DelegatedAmount:  0,
		DelegationID:     axongo.DelegationID{},
		ValidatorAddress: validatorAddress,
		StartEpoch:       0,
		EndEpoch:         0,
		UnlockConditions: axongo.DelegationOutputUnlockConditions{
			&axongo.AddressUnlockCondition{Address: addr},
		},
	}}
}

// NewDelegationOutputBuilderFromPrevious creates a new DelegationOutputBuilder starting from a copy of the previous axongo.DelegationOutput.
func NewDelegationOutputBuilderFromPrevious(previous *axongo.DelegationOutput) *DelegationOutputBuilder {
	return &DelegationOutputBuilder{
		//nolint:forcetypeassert // we can safely assume that this is a DelegationOutput
		output: previous.Clone().(*axongo.DelegationOutput),
	}
}

// DelegationOutputBuilder builds an axongo.DelegationOutput.
type DelegationOutputBuilder struct {
	output *axongo.DelegationOutput
}

// Amount sets the base token amount of the output.
func (builder *DelegationOutputBuilder) Amount(amount axongo.BaseToken) *DelegationOutputBuilder {
	builder.output.Amount = amount

	return builder
}

// DelegatedAmount sets the delegated amount of the output.
func (builder *DelegationOutputBuilder) DelegatedAmount(delegatedAmount axongo.BaseToken) *DelegationOutputBuilder {
	builder.output.DelegatedAmount = delegatedAmount

	return builder
}

// ValidatorAddress sets the validator address of the output.
func (builder *DelegationOutputBuilder) ValidatorAddress(validatorAddress *axongo.AccountAddress) *DelegationOutputBuilder {
	builder.output.ValidatorAddress = validatorAddress

	return builder
}

// DelegationID sets the delegation ID of the output.
func (builder *DelegationOutputBuilder) DelegationID(delegationID axongo.DelegationID) *DelegationOutputBuilder {
	builder.output.DelegationID = delegationID

	return builder
}

// StartEpoch sets the delegation start epoch.
func (builder *DelegationOutputBuilder) StartEpoch(startEpoch axongo.EpochIndex) *DelegationOutputBuilder {
	builder.output.StartEpoch = startEpoch

	return builder
}

// EndEpoch sets the delegation end epoch.
func (builder *DelegationOutputBuilder) EndEpoch(endEpoch axongo.EpochIndex) *DelegationOutputBuilder {
	builder.output.EndEpoch = endEpoch

	return builder
}

// Address sets/modifies an axongo.AddressUnlockCondition on the output.
func (builder *DelegationOutputBuilder) Address(addr axongo.Address) *DelegationOutputBuilder {
	builder.output.UnlockConditions.Upsert(&axongo.AddressUnlockCondition{Address: addr})

	return builder
}

// Build builds the axongo.DelegationOutput.
func (builder *DelegationOutputBuilder) Build() (*axongo.DelegationOutput, error) {
	builder.output.UnlockConditions.Sort()

	return builder.output, nil
}

// MustBuild works like Build() but panics if an error is encountered.
func (builder *DelegationOutputBuilder) MustBuild() *axongo.DelegationOutput {
	output, err := builder.Build()
	if err != nil {
		panic(err)
	}

	return output
}

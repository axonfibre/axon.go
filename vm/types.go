package vm

import (
	"fmt"

	axongo "github.com/axonfibre/axon.go/v4"
)

// InputSet is a map of OutputID to Output.
type InputSet map[axongo.OutputID]axongo.Output

func (inputSet InputSet) OutputSet() axongo.OutputSet {
	outputs := make(axongo.OutputSet, len(inputSet))
	for outputID := range inputSet {
		outputs[outputID] = inputSet[outputID]
	}

	return outputs
}

type ChainOutputWithIDs struct {
	ChainID  axongo.ChainID
	OutputID axongo.OutputID
	Output   axongo.ChainOutput
}

// ChainInputSet returns a ChainInputSet for all ChainOutputs in the InputSet.
func (inputSet InputSet) ChainInputSet() ChainInputSet {
	set := make(ChainInputSet)
	for utxoInputID, input := range inputSet {
		chainOutput, is := input.(axongo.ChainOutput)
		if !is {
			continue
		}

		chainID := chainOutput.ChainID()
		if chainID.Empty() {
			if utxoIDChainID, is := chainID.(axongo.UTXOIDChainID); is {
				chainID = utxoIDChainID.FromOutputID(utxoInputID)
			}
		}

		if chainID.Empty() {
			panic(fmt.Sprintf("output of type %s has empty chain ID but is not utxo dependable", chainOutput.Type()))
		}

		set[chainID] = &ChainOutputWithIDs{
			ChainID:  chainID,
			OutputID: utxoInputID,
			Output:   chainOutput,
		}
	}

	return set
}

// ChainInputSet is a map of ChainID to ChainOutput.
type ChainInputSet map[axongo.ChainID]*ChainOutputWithIDs

type BlockIssuanceCreditInputSet map[axongo.AccountID]axongo.BlockIssuanceCredits

// A map of either DelegationID or AccountID to their mana reward amount.
type RewardsInputSet map[axongo.ChainID]axongo.Mana

//nolint:revive // the VM at the beginning makes it more clear
type VMCommitmentInput *axongo.Commitment

type ResolvedInputs struct {
	InputSet
	BlockIssuanceCreditInputSet
	CommitmentInput VMCommitmentInput
	RewardsInputSet
}

type ImplicitAccountOutput struct {
	*axongo.BasicOutput
}

func (o *ImplicitAccountOutput) ChainID() axongo.ChainID {
	return axongo.EmptyAccountID
}

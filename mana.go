package iotago

import (
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
)

// Mana Structure defines the parameters used in mana calculations.
type ManaStructure struct {
	// ManaBitsCount is the number of bits used to represent Mana.
	ManaBitsCount uint8 `serix:"0,mapKey=manaBitsCount"`
	// ManaGenerationRate is the amount of potential Mana generated by 1 IOTA in 1 slot.
	ManaGenerationRate uint8 `serix:"1,mapKey=manaGenerationRate"`
	// ManaGenerationRateExponent is the scaling of ManaGenerationRate expressed as an exponent of 2.
	ManaGenerationRateExponent uint8 `serix:"2,mapKey=manaGenerationRateExponent"`
	// ManaDecayFactors is a lookup table of epoch index diff to mana decay factor (slice index 0 = 1 epoch).
	ManaDecayFactors []uint32 `serix:"3,lengthPrefixType=uint16,mapKey=manaDecayFactors"`
	// ManaDecayFactorsExponent is the scaling of ManaDecayFactors expressed as an exponent of 2.
	ManaDecayFactorsExponent uint8 `serix:"4,mapKey=manaDecayFactorsExponent"`
	// ManaDecayFactorEpochsSum is an integer approximation of the sum of decay over epochs.
	ManaDecayFactorEpochsSum uint32 `serix:"5,mapKey=manaDecayFactorEpochsSum"`
	// ManaDecayFactorEpochsSumExponent is the scaling of ManaDecayFactorEpochsSum expressed as an exponent of 2.
	ManaDecayFactorEpochsSumExponent uint8 `serix:"6,mapKey=manaDecayFactorEpochsSumExponent"`
}

func (m ManaStructure) Equals(other ManaStructure) bool {
	return m.ManaBitsCount == other.ManaBitsCount &&
		m.ManaGenerationRate == other.ManaGenerationRate &&
		m.ManaGenerationRateExponent == other.ManaGenerationRateExponent &&
		lo.Equal(m.ManaDecayFactors, other.ManaDecayFactors) &&
		m.ManaDecayFactorsExponent == other.ManaDecayFactorsExponent &&
		m.ManaDecayFactorEpochsSum == other.ManaDecayFactorEpochsSum &&
		m.ManaDecayFactorEpochsSumExponent == other.ManaDecayFactorEpochsSumExponent
}

type RewardsParameters struct {
	// ValidatorBlocksPerSlot is the number of validation blocks that should be issued by a selected validator per slot during its epoch duties.
	ValidatorBlocksPerSlot uint8 `serix:"0,mapKey=validatorBlocksPerSlot"`
	// ProfitMarginExponent is used for shift operation for calculation of profit margin.
	ProfitMarginExponent uint8 `serix:"1,mapKey=profitMarginExponent"`
	// BootstrappinDuration is the length in epochs of the bootstrapping phase, (approx 3 years).
	BootstrappingDuration EpochIndex `serix:"2,mapKey=bootstrappinDuration"`
	// RewardsManaShareCoefficient is the coefficient used for calculation of initial rewards, relative to the term theta/(1-theta) from the Whitepaper, with theta = 2/3.
	RewardsManaShareCoefficient uint64 `serix:"3,mapKey=rewardsManaShareCoefficient"`
	// DecayBalancingConstantExponent is the exponent used for calculation of the initial reward.
	DecayBalancingConstantExponent uint8 `serix:"4,mapKey=decayBalancingConstantExponent"`
	// DecayBalancingConstant needs to be an integer approc  calculated based on chosen DecayBalancingConstantExponent.
	DecayBalancingConstant uint64 `serix:"5,mapKey=decayBalancingConstant"`
	// PoolCoefficientExponent is the exponent used for shifting operation in the pool rewards calculations.
	PoolCoefficientExponent uint8 `serix:"6,mapKey=poolCoefficientExponent"`
}

func (r RewardsParameters) Equals(other RewardsParameters) bool {
	return r.ValidatorBlocksPerSlot == other.ValidatorBlocksPerSlot &&
		r.ProfitMarginExponent == other.ProfitMarginExponent && r.BootstrappingDuration == other.BootstrappingDuration &&
		r.RewardsManaShareCoefficient == other.RewardsManaShareCoefficient &&
		r.DecayBalancingConstantExponent == other.DecayBalancingConstantExponent &&
		r.DecayBalancingConstant == other.DecayBalancingConstant &&
		r.PoolCoefficientExponent == other.PoolCoefficientExponent
}

func (r RewardsParameters) TargetReward(index EpochIndex, api API) (Mana, error) {
	if index > r.BootstrappingDuration {
		return Mana(api.ComputedFinalReward()), nil
	}

	decayedInitialReward, err := api.ManaDecayProvider().RewardsWithDecay(Mana(api.ComputedInitialReward()), index, index)
	if err != nil {
		return 0, ierrors.Errorf("failed to calculate decayed initial reward: %w", err)
	}

	return decayedInitialReward, nil
}

package iotago

import (
	"github.com/iotaledger/hive.go/core/safemath"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
)

// Mana Structure defines the parameters used in mana calculations.
type ManaParameters struct {
	// BitsCount is the number of bits used to represent Mana.
	BitsCount uint8 `serix:""`
	// GenerationRate is the amount of potential Mana generated by 1 microIOTA in 1 slot multiplied by 2^GenerationRateExponent.
	GenerationRate uint8 `serix:""`
	// GenerationRateExponent is the scaling of GenerationRate expressed as an exponent of 2.
	// The actual generation rate of Mana is given by GenerationRate * 2^(-GenerationRateExponent).
	GenerationRateExponent uint8 `serix:""`
	// DecayFactors is a lookup table of epoch diff to mana decay factor (slice index 0 = 1 epoch).
	// The actual decay factor is given by DecayFactors[epochDiff] * 2^(-DecayFactorsExponent).
	DecayFactors []uint32 `serix:",lenPrefix=uint16"`
	// DecayFactorsExponent is the scaling of DecayFactors expressed as an exponent of 2.
	DecayFactorsExponent uint8 `serix:""`
	// DecayFactorEpochsSum is an integer approximation of the sum of decay over epochs.
	DecayFactorEpochsSum uint32 `serix:""`
	// DecayFactorEpochsSumExponent is the scaling of DecayFactorEpochsSum expressed as an exponent of 2.
	DecayFactorEpochsSumExponent uint8 `serix:""`
	// AnnualDecayFactorPercentage is the decay factor for 1 year.
	AnnualDecayFactorPercentage uint8 `serix:""`
}

func (m ManaParameters) Equals(other ManaParameters) bool {
	return m.BitsCount == other.BitsCount &&
		m.GenerationRate == other.GenerationRate &&
		m.GenerationRateExponent == other.GenerationRateExponent &&
		lo.Equal(m.DecayFactors, other.DecayFactors) &&
		m.DecayFactorsExponent == other.DecayFactorsExponent &&
		m.DecayFactorEpochsSum == other.DecayFactorEpochsSum &&
		m.DecayFactorEpochsSumExponent == other.DecayFactorEpochsSumExponent &&
		m.AnnualDecayFactorPercentage == other.AnnualDecayFactorPercentage
}

type RewardsParameters struct {
	// ProfitMarginExponent is used for shift operation for calculation of profit margin.
	ProfitMarginExponent uint8 `serix:""`
	// BootstrappingDuration is the length in epochs of the bootstrapping phase, (approx 3 years).
	BootstrappingDuration EpochIndex `serix:""`
	// RewardToGenerationRatio is the ratio of the final rewards rate to the generation rate of Mana.
	RewardToGenerationRatio uint8 `serix:""`
	// InitialRewardsRate is the rate of Mana rewards at the start of the bootstrapping phase.
	InitialRewardsRate Mana `serix:""`
	// FinalRewardsRate is the rate of Mana rewards after the bootstrapping phase.
	FinalRewardsRate Mana `serix:""`
	// PoolCoefficientExponent is the exponent used for shifting operation in the pool rewards calculations.
	PoolCoefficientExponent uint8 `serix:""`
	// The number of epochs for which rewards are retained.
	RetentionPeriod uint16 `serix:""`
}

func (r RewardsParameters) Equals(other RewardsParameters) bool {
	return r.ProfitMarginExponent == other.ProfitMarginExponent &&
		r.BootstrappingDuration == other.BootstrappingDuration &&
		r.RewardToGenerationRatio == other.RewardToGenerationRatio &&
		r.InitialRewardsRate == other.InitialRewardsRate &&
		r.FinalRewardsRate == other.FinalRewardsRate &&
		r.PoolCoefficientExponent == other.PoolCoefficientExponent &&
		r.RetentionPeriod == other.RetentionPeriod
}

func (r RewardsParameters) TargetReward(epoch EpochIndex, api API) (Mana, error) {
	if epoch > r.BootstrappingDuration {
		return api.ProtocolParameters().RewardsParameters().FinalRewardsRate, nil
	}

	// Rewards start at epoch 0.
	decayedInitialReward, err := api.ManaDecayProvider().DecayManaByEpochs(api.ProtocolParameters().RewardsParameters().InitialRewardsRate, 0, epoch)
	if err != nil {
		return 0, ierrors.Errorf("failed to calculate decayed initial reward: %w", err)
	}

	return decayedInitialReward, nil
}

func ManaCost(rmc Mana, workScore WorkScore) (Mana, error) {
	return safemath.SafeMul(rmc, Mana(workScore))
}

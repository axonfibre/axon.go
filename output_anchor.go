package iotago

import (
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	// ErrInvalidAnchorStateTransition gets returned when an anchor is doing an invalid state transition.
	ErrInvalidAnchorStateTransition = ierrors.New("invalid anchor state transition")
	// ErrInvalidAnchorGovernanceTransition gets returned when an anchor is doing an invalid governance transition.
	ErrInvalidAnchorGovernanceTransition = ierrors.New("invalid anchor governance transition")
	// ErrAnchorMissing gets returned when an anchor is missing.
	ErrAnchorMissing = ierrors.New("anchor is missing")
)

type (
	AnchorOutputUnlockCondition  interface{ UnlockCondition }
	AnchorOutputFeature          interface{ Feature }
	AnchorOutputImmFeature       interface{ Feature }
	AnchorOutputUnlockConditions = UnlockConditions[AnchorOutputUnlockCondition]
	AnchorOutputFeatures         = Features[AnchorOutputFeature]
	AnchorOutputImmFeatures      = Features[AnchorOutputImmFeature]
)

// AnchorOutput is an output type which represents an anchor.
type AnchorOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount BaseToken `serix:""`
	// The stored mana held by the output.
	Mana Mana `serix:""`
	// The identifier for this anchor.
	AnchorID AnchorID `serix:""`
	// The index of the state.
	StateIndex uint32 `serix:""`
	// The unlock conditions on this output.
	UnlockConditions AnchorOutputUnlockConditions `serix:",omitempty"`
	// The features on the output.
	Features AnchorOutputFeatures `serix:",omitempty"`
	// The immutable feature on the output.
	ImmutableFeatures AnchorOutputImmFeatures `serix:",omitempty"`
}

func (a *AnchorOutput) GovernorAddress() Address {
	return a.UnlockConditions.MustSet().GovernorAddress().Address
}

func (a *AnchorOutput) StateController() Address {
	return a.UnlockConditions.MustSet().StateControllerAddress().Address
}

func (a *AnchorOutput) Clone() Output {
	return &AnchorOutput{
		Amount:            a.Amount,
		Mana:              a.Mana,
		AnchorID:          a.AnchorID,
		StateIndex:        a.StateIndex,
		UnlockConditions:  a.UnlockConditions.Clone(),
		Features:          a.Features.Clone(),
		ImmutableFeatures: a.ImmutableFeatures.Clone(),
	}
}

func (a *AnchorOutput) Equal(other Output) bool {
	otherOutput, isSameType := other.(*AnchorOutput)
	if !isSameType {
		return false
	}

	if a.Amount != otherOutput.Amount {
		return false
	}

	if a.Mana != otherOutput.Mana {
		return false
	}

	if a.AnchorID != otherOutput.AnchorID {
		return false
	}

	if a.StateIndex != otherOutput.StateIndex {
		return false
	}

	if !a.UnlockConditions.Equal(otherOutput.UnlockConditions) {
		return false
	}

	if !a.Features.Equal(otherOutput.Features) {
		return false
	}

	if !a.ImmutableFeatures.Equal(otherOutput.ImmutableFeatures) {
		return false
	}

	return true
}

func (a *AnchorOutput) UnlockableBy(ident Address, next TransDepIdentOutput, pastBoundedSlotIndex SlotIndex, futureBoundedSlotIndex SlotIndex) (bool, error) {
	return outputUnlockableBy(a, next, ident, pastBoundedSlotIndex, futureBoundedSlotIndex)
}

func (a *AnchorOutput) StorageScore(storageScoreStruct *StorageScoreStructure, _ StorageScoreFunc) StorageScore {
	return storageScoreStruct.OffsetOutput +
		storageScoreStruct.FactorData().Multiply(StorageScore(a.Size())) +
		a.UnlockConditions.StorageScore(storageScoreStruct, nil) +
		a.Features.StorageScore(storageScoreStruct, nil) +
		a.ImmutableFeatures.StorageScore(storageScoreStruct, nil)
}

func (a *AnchorOutput) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	workScoreConditions, err := a.UnlockConditions.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	workScoreFeatures, err := a.Features.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	workScoreImmutableFeatures, err := a.ImmutableFeatures.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	return workScoreParameters.Output.Add(workScoreConditions, workScoreFeatures, workScoreImmutableFeatures)
}

func (a *AnchorOutput) Ident(nextState TransDepIdentOutput) (Address, error) {
	// if there isn't a next state, then only the governance address can destroy the anchor
	if nextState == nil {
		return a.GovernorAddress(), nil
	}
	otherAnchorOutput, isAnchorOutput := nextState.(*AnchorOutput)
	if !isAnchorOutput {
		return nil, ierrors.Wrapf(ErrTransDepIdentOutputNextInvalid, "expected AnchorOutput but got %s for ident computation", nextState.Type())
	}
	switch {
	case a.StateIndex == otherAnchorOutput.StateIndex:
		return a.GovernorAddress(), nil
	case a.StateIndex+1 == otherAnchorOutput.StateIndex:
		return a.StateController(), nil
	default:
		return nil, ierrors.Wrap(ErrTransDepIdentOutputNextInvalid, "can not compute right ident for anchor output as state index delta is invalid")
	}
}

func (a *AnchorOutput) ChainID() ChainID {
	return a.AnchorID
}

func (a *AnchorOutput) FeatureSet() FeatureSet {
	return a.Features.MustSet()
}

func (a *AnchorOutput) UnlockConditionSet() UnlockConditionSet {
	return a.UnlockConditions.MustSet()
}

func (a *AnchorOutput) ImmutableFeatureSet() FeatureSet {
	return a.ImmutableFeatures.MustSet()
}

func (a *AnchorOutput) BaseTokenAmount() BaseToken {
	return a.Amount
}

func (a *AnchorOutput) StoredMana() Mana {
	return a.Mana
}

func (a *AnchorOutput) Target() (Address, error) {
	addr := new(AnchorAddress)
	copy(addr[:], a.AnchorID[:])

	return addr, nil
}

func (a *AnchorOutput) Type() OutputType {
	return OutputAnchor
}

func (a *AnchorOutput) Size() int {
	// OutputType
	return serializer.OneByte +
		BaseTokenSize +
		ManaSize +
		AnchorIDLength +
		// StateIndex
		serializer.UInt32ByteSize +
		a.UnlockConditions.Size() +
		a.Features.Size() +
		a.ImmutableFeatures.Size()
}

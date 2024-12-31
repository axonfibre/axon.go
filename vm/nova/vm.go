package nova

import (
	"fmt"

	"github.com/axonfibre/fibre.go/core/safemath"
	"github.com/axonfibre/fibre.go/ierrors"
	axongo "github.com/axonfibre/axon.go/v4"
	"github.com/axonfibre/axon.go/v4/vm"
)

// NewVirtualMachine returns an VirtualMachine adhering to the Nova protocol.
func NewVirtualMachine() vm.VirtualMachine {
	return &virtualMachine{
		execList: []vm.ExecFunc{
			vm.ExecFuncTimelocks(),
			vm.ExecFuncSenderUnlocked(),
			vm.ExecFuncBalancedBaseTokens(),
			vm.ExecFuncBalancedNativeTokens(),
			vm.ExecFuncChainTransitions(),
			vm.ExecFuncBalancedMana(),
			vm.ExecFuncAtMostOneImplicitAccountCreationAddress(),
		},
	}
}

type virtualMachine struct {
	execList []vm.ExecFunc
}

func NewVMParamsWorkingSet(api axongo.API, t *axongo.Transaction, resolvedInputs vm.ResolvedInputs) (*vm.WorkingSet, error) {
	var err error
	utxoInputsSet := constructInputSet(resolvedInputs.InputSet)
	workingSet := &vm.WorkingSet{}
	workingSet.Tx = t
	workingSet.UnlockedAddrs = make(vm.UnlockedAddresses)
	workingSet.UTXOInputsSet = utxoInputsSet
	workingSet.InputIDToInputIndex = make(map[axongo.OutputID]uint16)
	for inputIndex, txInput := range workingSet.Tx.TransactionEssence.Inputs {
		//nolint:forcetypeassert // we can safely assume that this is an UTXOInput
		txInputID := txInput.(*axongo.UTXOInput).OutputID()
		workingSet.InputIDToInputIndex[txInputID] = uint16(inputIndex)
		input, ok := workingSet.UTXOInputsSet[txInputID]
		if !ok {
			panic(fmt.Sprintf("UTXO for input %d should be supplied %s", inputIndex, txInputID))
		}
		workingSet.UTXOInputs = append(workingSet.UTXOInputs, input)
	}

	txID, err := workingSet.Tx.ID()
	if err != nil {
		panic(fmt.Sprintf("transaction ID computation should have succeeded: %s", err.Error()))
	}

	workingSet.InChains = utxoInputsSet.ChainInputSet()
	workingSet.OutChains = workingSet.Tx.Outputs.ChainOutputSet(txID)

	workingSet.BIC = resolvedInputs.BlockIssuanceCreditInputSet
	workingSet.Commitment = resolvedInputs.CommitmentInput
	workingSet.Rewards = resolvedInputs.RewardsInputSet

	workingSet.TotalManaIn, err = vm.TotalManaIn(
		api.ManaDecayProvider(),
		api.StorageScoreStructure(),
		workingSet.Tx.CreationSlot,
		workingSet.UTXOInputsSet,
		workingSet.Rewards,
	)
	if err != nil {
		return nil, err
	}

	workingSet.TotalManaOut, err = vm.TotalManaOut(
		workingSet.Tx.Outputs,
		workingSet.Tx.Allotments,
	)
	if err != nil {
		return nil, err
	}

	return workingSet, nil
}

func constructInputSet(inputSet vm.InputSet) vm.InputSet {
	utxoInputsSet := vm.InputSet{}
	for outputID, outputWithCreationSlot := range inputSet {
		if basicOutput, isBasic := outputWithCreationSlot.(*axongo.BasicOutput); isBasic {
			if addressUnlock := basicOutput.UnlockConditionSet().Address(); addressUnlock != nil {
				if addressUnlock.Address.Type() == axongo.AddressImplicitAccountCreation {
					utxoInputsSet[outputID] = &vm.ImplicitAccountOutput{BasicOutput: basicOutput}

					continue
				}
			}
		}
		utxoInputsSet[outputID] = outputWithCreationSlot
	}

	return utxoInputsSet
}

func (novaVM *virtualMachine) ValidateUnlocks(signedTransaction *axongo.SignedTransaction, inputs vm.ResolvedInputs) (unlockedAddrs vm.UnlockedAddresses, err error) {
	return vm.ValidateUnlocks(signedTransaction, inputs)
}

func (novaVM *virtualMachine) Execute(transaction *axongo.Transaction, resolvedInputs vm.ResolvedInputs, unlockedAddrs vm.UnlockedAddresses, execFunctions ...vm.ExecFunc) (outputs []axongo.Output, err error) {
	vmParams := &vm.Params{
		API: transaction.API,
	}

	if vmParams.WorkingSet, err = NewVMParamsWorkingSet(vmParams.API, transaction, resolvedInputs); err != nil {
		return nil, ierrors.Wrap(err, "failed to create working set")
	}
	vmParams.WorkingSet.UnlockedAddrs = unlockedAddrs

	if len(execFunctions) == 0 {
		execFunctions = novaVM.execList
	}

	err = vm.RunVMFuncs(novaVM, vmParams, execFunctions...)
	if err != nil {
		return nil, ierrors.Wrap(err, "failed to execute transaction")
	}

	outputs = make([]axongo.Output, len(transaction.Outputs))
	for i, output := range transaction.Outputs {
		outputs[i] = output
	}

	return outputs, nil
}

func (novaVM *virtualMachine) ChainSTVF(vmParams *vm.Params, transType axongo.ChainTransitionType, input *vm.ChainOutputWithIDs, next axongo.ChainOutput) error {
	transitionState := next
	if transType != axongo.ChainTransitionTypeGenesis {
		transitionState = input.Output
	}

	var ok bool
	switch castedInput := transitionState.(type) {
	case *axongo.AccountOutput:
		var nextAccount *axongo.AccountOutput
		if next != nil {
			if nextAccount, ok = next.(*axongo.AccountOutput); !ok {
				return ierrors.New("can only state transition to another account output")
			}
		}

		return accountSTVF(vmParams, input, transType, nextAccount)

	case *vm.ImplicitAccountOutput:
		var nextAccount *axongo.AccountOutput
		if next != nil {
			if nextAccount, ok = next.(*axongo.AccountOutput); !ok {
				return ierrors.New("can only state transition implicit account to an account output")
			}
		}

		err := implicitAccountSTVF(vmParams, castedInput, input.OutputID, nextAccount, transType)
		if err != nil {
			return ierrors.Wrapf(err, "transition failed for implicit account with output ID %s", input.OutputID.ToHex())
		}

		return nil

	case *axongo.AnchorOutput:
		var nextAnchor *axongo.AnchorOutput
		if next != nil {
			if nextAnchor, ok = next.(*axongo.AnchorOutput); !ok {
				return ierrors.New("can only state transition to another anchor output")
			}
		}

		return anchorSTVF(vmParams, input, transType, nextAnchor)

	case *axongo.FoundryOutput:
		var nextFoundry *axongo.FoundryOutput
		if next != nil {
			if nextFoundry, ok = next.(*axongo.FoundryOutput); !ok {
				return ierrors.New("can only state transition to another foundry output")
			}
		}

		return foundrySTVF(vmParams, input, transType, nextFoundry)

	case *axongo.NFTOutput:
		var nextNFT *axongo.NFTOutput
		if next != nil {
			if nextNFT, ok = next.(*axongo.NFTOutput); !ok {
				return ierrors.New("can only state transition to another NFT output")
			}
		}

		return nftSTVF(vmParams, input, transType, nextNFT)

	case *axongo.DelegationOutput:
		var nextDelegationOutput *axongo.DelegationOutput
		if next != nil {
			if nextDelegationOutput, ok = next.(*axongo.DelegationOutput); !ok {
				return ierrors.New("can only state transition to another Delegation output")
			}
		}

		return delegationSTVF(vmParams, input, transType, nextDelegationOutput)

	default:
		panic(fmt.Sprintf("invalid output type %v passed to Nova virtual machine", input.Output))
	}
}

// For implicit account conversion, there must be a basic output as input, and an account output as output with an AccountID matching the input.
func implicitAccountSTVF(vmParams *vm.Params, implicitAccount *vm.ImplicitAccountOutput, outputID axongo.OutputID, next *axongo.AccountOutput, transType axongo.ChainTransitionType) error {
	if transType == axongo.ChainTransitionTypeDestroy {
		return axongo.ErrImplicitAccountDestructionDisallowed
	}

	// Create a wrapper around the implicit account.
	implicitAccountChainOutput := &vm.ChainOutputWithIDs{
		ChainID:  next.AccountID,
		OutputID: outputID,
		Output:   implicitAccount,
	}

	if err := accountBlockIssuanceCreditLocked(implicitAccountChainOutput, vmParams.WorkingSet.BIC); err != nil {
		return err
	}

	implicitAccountBlockIssuerFeature := &axongo.BlockIssuerFeature{
		BlockIssuerKeys: axongo.NewBlockIssuerKeys(),
		// Setting MaxSlotIndex means one cannot remove the block issuer feature in the transition, but it does allow for setting
		// the expiry slot to a lower value, which is the behavior we want.
		ExpirySlot: axongo.MaxSlotIndex,
	}

	if err := accountBlockIssuerSTVF(vmParams, implicitAccountChainOutput, implicitAccountBlockIssuerFeature, next); err != nil {
		return err
	}

	return accountGenesisValid(vmParams, next, false)
}

// For output AccountOutput(s) with non-zeroed AccountID, there must be a corresponding input AccountOutput where either its
// AccountID is zeroed and FoundryCounter is zero or an input AccountOutput with the same AccountID.
func accountSTVF(vmParams *vm.Params, input *vm.ChainOutputWithIDs, transType axongo.ChainTransitionType, next *axongo.AccountOutput) error {
	// Whether the transaction is claiming Mana rewards for this account.
	isClaimingRewards := false
	if vmParams.WorkingSet.Rewards != nil && input != nil {
		_, isClaimingRewards = vmParams.WorkingSet.Rewards[input.ChainID]
	}

	// Whether the account is removing the staking feature.
	isRemovingStakingFeatureValue := false
	isRemovingStakingFeature := &isRemovingStakingFeatureValue

	switch transType {
	case axongo.ChainTransitionTypeGenesis:
		if err := accountGenesisValid(vmParams, next, true); err != nil {
			return err
		}
	case axongo.ChainTransitionTypeStateChange:
		if err := accountStateChangeValid(vmParams, input, next, isRemovingStakingFeature); err != nil {
			return err
		}
	case axongo.ChainTransitionTypeDestroy:
		if err := accountDestructionValid(vmParams, input, isRemovingStakingFeature); err != nil {
			return err
		}
	default:
		panic("unknown chain transition type in AccountOutput")
	}

	if isClaimingRewards && !*isRemovingStakingFeature {
		return axongo.ErrStakingRewardClaimingInvalid
	}

	if !isClaimingRewards && *isRemovingStakingFeature {
		return axongo.ErrStakingRewardInputMissing
	}

	return nil
}

func accountGenesisValid(vmParams *vm.Params, next *axongo.AccountOutput, accountIDMustBeZeroed bool) error {
	if accountIDMustBeZeroed && !next.AccountID.Empty() {
		return axongo.ErrNewChainOutputHasNonZeroedID
	}

	if nextBlockIssuerFeat := next.FeatureSet().BlockIssuer(); nextBlockIssuerFeat != nil {
		if vmParams.WorkingSet.Commitment == nil {
			panic("commitment input should be present for block issuer features on the output side which should be validated syntactically")
		}

		pastBoundedSlot := vmParams.PastBoundedSlotIndex(vmParams.WorkingSet.Commitment.Slot)
		if nextBlockIssuerFeat.ExpirySlot < pastBoundedSlot {
			return ierrors.WithMessagef(axongo.ErrBlockIssuerExpiryTooEarly, "is %d, must be >= %d", nextBlockIssuerFeat.ExpirySlot, pastBoundedSlot)
		}
	}

	if stakingFeat := next.FeatureSet().Staking(); stakingFeat != nil {
		if err := accountStakingGenesisValidation(vmParams, stakingFeat); err != nil {
			return err
		}
	}

	return vm.IsIssuerOnOutputUnlocked(next, vmParams.WorkingSet.UnlockedAddrs)
}

func accountStateChangeValid(vmParams *vm.Params, input *vm.ChainOutputWithIDs, next *axongo.AccountOutput, isRemovingStakingFeature *bool) error {
	//nolint:forcetypeassert // we can safely assume that this is an AccountOutput
	current := input.Output.(*axongo.AccountOutput)
	if !current.ImmutableFeatures.Equal(next.ImmutableFeatures) {
		return ierrors.WithMessagef(
			axongo.ErrChainOutputImmutableFeaturesChanged, "old state %s, next state %s", current.ImmutableFeatures, next.ImmutableFeatures,
		)
	}

	// If a Block Issuer Feature is present on the input side of the transaction,
	// and the BIC is negative, the account is locked.
	if current.FeatureSet().BlockIssuer() != nil {
		if err := accountBlockIssuanceCreditLocked(input, vmParams.WorkingSet.BIC); err != nil {
			return err
		}
	}

	if err := accountStakingSTVF(vmParams, current, next, isRemovingStakingFeature); err != nil {
		return err
	}

	if err := accountFoundryCounterSTVF(vmParams, current, next); err != nil {
		return err
	}

	return accountBlockIssuerSTVF(vmParams, input, input.Output.FeatureSet().BlockIssuer(), next)
}

// If an account output has a block issuer feature, the following conditions for its transition must be checked.
// The block issuer credit must be non-negative.
// The expiry time of the block issuer feature, if creating new account or expired already, must be set at least MaxCommittableAge greater than the Commitment Input.
// Check that at least one Block Issuer Key is present.
func accountBlockIssuerSTVF(vmParams *vm.Params, input *vm.ChainOutputWithIDs, currentBlockIssuerFeat *axongo.BlockIssuerFeature, next *axongo.AccountOutput) error {
	current := input.Output
	nextBlockIssuerFeat := next.FeatureSet().BlockIssuer()

	// if the account has no block issuer feature.
	if currentBlockIssuerFeat == nil && nextBlockIssuerFeat == nil {
		return nil
	}

	if vmParams.WorkingSet.Commitment == nil {
		return axongo.ErrBlockIssuerCommitmentInputMissing
	}

	commitmentInputSlot := vmParams.WorkingSet.Commitment.Slot
	pastBoundedSlot := vmParams.PastBoundedSlotIndex(commitmentInputSlot)

	if currentBlockIssuerFeat != nil && currentBlockIssuerFeat.ExpirySlot >= commitmentInputSlot {
		// if the block issuer feature has not expired, it can not be removed.
		if nextBlockIssuerFeat == nil {
			return ierrors.WithMessagef(axongo.ErrBlockIssuerNotExpired, "commitment slot: %d, expiry slot: %d", commitmentInputSlot, currentBlockIssuerFeat.ExpirySlot)
		}
		if nextBlockIssuerFeat.ExpirySlot != currentBlockIssuerFeat.ExpirySlot && nextBlockIssuerFeat.ExpirySlot < pastBoundedSlot {
			return ierrors.WithMessagef(axongo.ErrBlockIssuerExpiryTooEarly, "is %d, must be >= %d", nextBlockIssuerFeat.ExpirySlot, pastBoundedSlot)
		}
	} else if nextBlockIssuerFeat != nil {
		// The block issuer feature was newly added,
		// or the current feature has expired but it was not removed.
		// In both cases the expiry slot must be set sufficiently far in the future.
		if nextBlockIssuerFeat.ExpirySlot < pastBoundedSlot {
			return ierrors.WithMessagef(axongo.ErrBlockIssuerExpiryTooEarly, "is %d, must be >= %d", nextBlockIssuerFeat.ExpirySlot, pastBoundedSlot)
		}
	}

	// the Mana on the account on the input side must not be moved to any other outputs or accounts.
	manaDecayProvider := vmParams.API.ManaDecayProvider()
	storageScoreStructure := vmParams.API.StorageScoreStructure()

	manaIn := vmParams.WorkingSet.TotalManaIn
	manaOut := vmParams.WorkingSet.TotalManaOut

	// AccountInStored
	manaStoredAccount, err := manaDecayProvider.DecayManaBySlots(current.StoredMana(), input.OutputID.CreationSlot(), vmParams.WorkingSet.Tx.CreationSlot)
	if err != nil {
		return ierrors.Wrapf(err, "account %s stored mana calculation failed", next.AccountID)
	}
	manaIn, err = safemath.SafeSub(manaIn, manaStoredAccount)
	if err != nil {
		return ierrors.Wrapf(err, "account %s stored mana in exceeds total remaining mana in, manaStoredAccountIn: %d, manaIn: %d", next.AccountID, manaStoredAccount, manaIn)
	}

	// AccountInPotential - the potential mana from the input side of the account in question
	manaPotentialAccount, err := axongo.PotentialMana(manaDecayProvider, storageScoreStructure, input.Output, input.OutputID.CreationSlot(), vmParams.WorkingSet.Tx.CreationSlot)

	if err != nil {
		return ierrors.Wrapf(err, "account %s potential mana calculation failed", next.AccountID)
	}

	manaIn, err = safemath.SafeSub(manaIn, manaPotentialAccount)
	if err != nil {
		return ierrors.Wrapf(err, "account %s potential mana in exceeds total remaining mana in, manaPotentialAccountIn: %d, manaIn: %d", next.AccountID, manaPotentialAccount, manaIn)
	}

	// AccountOutStored - stored Mana on the output side of the account in question
	manaOut, err = safemath.SafeSub(manaOut, next.Mana)
	if err != nil {
		return ierrors.Wrapf(err, "account %s stored mana out exceeds total remaining mana out, storedManaOut: %d, manaOut: %d", next.AccountID, next.Mana, manaOut)
	}

	// AccountOutAllotted - allotments to the account in question
	accountOutAllotted := vmParams.WorkingSet.Tx.Allotments.Get(next.AccountID)
	manaOut, err = safemath.SafeSub(manaOut, accountOutAllotted)
	if err != nil {
		return ierrors.Wrapf(err, "account %s allotment exceeds total remaining mana out, accountAllotted: %d, manaOut: %d", next.AccountID, accountOutAllotted, manaOut)
	}

	// AccountOutLocked - outputs with manalock conditions
	minManalockedSlot := pastBoundedSlot + vmParams.API.ProtocolParameters().MaxCommittableAge()
	for outputIndex, output := range vmParams.WorkingSet.Tx.Outputs {
		if output.UnlockConditionSet().HasManalockCondition(next.AccountID, minManalockedSlot) {
			manaOut, err = safemath.SafeSub(manaOut, output.StoredMana())
			if err != nil {
				return ierrors.Wrapf(err, "account %s manalocked output mana exceeds total remaining mana out, outputIndex: %d, outputStoredMana: %d, manaOut: %d", next.AccountID, outputIndex, output.StoredMana(), manaOut)
			}
		}
	}

	if manaIn < manaOut {
		return ierrors.WithMessagef(axongo.ErrManaMovedOffBlockIssuerAccount, "mana in %d, mana out %d", manaIn, manaOut)
	}

	return nil
}

func accountStakingSTVF(vmParams *vm.Params, current *axongo.AccountOutput, next *axongo.AccountOutput, isRemovingStakingFeature *bool) error {
	currentStakingFeat := current.FeatureSet().Staking()
	nextStakingFeat := next.FeatureSet().Staking()

	if currentStakingFeat != nil {
		commitment := vmParams.WorkingSet.Commitment
		if commitment == nil {
			return axongo.ErrStakingCommitmentInputMissing
		}

		timeProvider := vmParams.API.TimeProvider()
		pastBoundedSlot := vmParams.PastBoundedSlotIndex(commitment.Slot)
		pastBoundedEpoch := timeProvider.EpochFromSlot(pastBoundedSlot)
		futureBoundedSlot := vmParams.FutureBoundedSlotIndex(commitment.Slot)
		futureBoundedEpoch := timeProvider.EpochFromSlot(futureBoundedSlot)

		if futureBoundedEpoch <= currentStakingFeat.EndEpoch {
			earliestUnbondingEpoch := pastBoundedEpoch + vmParams.API.ProtocolParameters().StakingUnbondingPeriod()

			return accountStakingNonExpiredValidation(
				currentStakingFeat, nextStakingFeat, earliestUnbondingEpoch,
			)
		}

		return accountStakingExpiredValidation(vmParams, currentStakingFeat, nextStakingFeat, isRemovingStakingFeature)
	} else if nextStakingFeat != nil {
		return accountStakingGenesisValidation(vmParams, nextStakingFeat)
	}

	return nil
}

// Validates the rules for a newly added Staking Feature in an account,
// or one which was effectively removed and added within the same transaction.
// This is allowed as long as the epoch range of the old and new feature are disjoint.
func accountStakingGenesisValidation(vmParams *vm.Params, stakingFeat *axongo.StakingFeature) error {
	commitment := vmParams.WorkingSet.Commitment
	if commitment == nil {
		panic("commitment input should be present for staking features on the output side which should be validated syntactically")
	}

	pastBoundedSlot := vmParams.PastBoundedSlotIndex(commitment.Slot)
	timeProvider := vmParams.API.TimeProvider()
	pastBoundedEpoch := timeProvider.EpochFromSlot(pastBoundedSlot)

	if stakingFeat.StartEpoch != pastBoundedEpoch {
		return ierrors.WithMessagef(axongo.ErrStakingStartEpochInvalid, "is %d, expected %d", stakingFeat.StartEpoch, pastBoundedEpoch)
	}

	unbondingEpoch := pastBoundedEpoch + vmParams.API.ProtocolParameters().StakingUnbondingPeriod()
	if stakingFeat.EndEpoch < unbondingEpoch {
		return ierrors.WithMessagef(axongo.ErrStakingEndEpochTooEarly, "end epoch %d should be >= %d", stakingFeat.EndEpoch, unbondingEpoch)
	}

	return nil
}

// Validates a staking feature's transition if the feature is not expired,
// i.e. the current epoch is before the end epoch.
func accountStakingNonExpiredValidation(
	currentStakingFeat *axongo.StakingFeature,
	nextStakingFeat *axongo.StakingFeature,
	earliestUnbondingEpoch axongo.EpochIndex,
) error {
	if nextStakingFeat == nil {
		return axongo.ErrStakingFeatureRemovedBeforeUnbonding
	}

	if currentStakingFeat.StakedAmount != nextStakingFeat.StakedAmount ||
		currentStakingFeat.FixedCost != nextStakingFeat.FixedCost ||
		currentStakingFeat.StartEpoch != nextStakingFeat.StartEpoch {
		return axongo.ErrStakingFeatureModifiedBeforeUnbonding
	}

	if currentStakingFeat.EndEpoch != nextStakingFeat.EndEpoch &&
		nextStakingFeat.EndEpoch < earliestUnbondingEpoch {
		return ierrors.WithMessagef(axongo.ErrStakingEndEpochTooEarly, "end epoch %d should be >= %d or the end epoch must match on input and output side", nextStakingFeat.EndEpoch, earliestUnbondingEpoch)
	}

	return nil
}

// Validates a staking feature's transition if the feature is expired,
// i.e. the current epoch is equal or after the end epoch.
func accountStakingExpiredValidation(
	vmParams *vm.Params,
	currentStakingFeat *axongo.StakingFeature,
	nextStakingFeat *axongo.StakingFeature,
	isRemovingStakingFeature *bool,
) error {
	if nextStakingFeat == nil {
		*isRemovingStakingFeature = true
	} else if !currentStakingFeat.Equal(nextStakingFeat) {
		// If an expired feature is changed it must be transitioned as if newly added.
		if err := accountStakingGenesisValidation(vmParams, nextStakingFeat); err != nil {
			return ierrors.Wrap(err, "rewards claiming without removing the feature requires updating the feature")
		}
		// If staking feature genesis validation succeeds, the start epoch has been reset which means the new epoch range
		// is disjoint from the current staking feature's, which can therefore be considered as removing and re-adding
		// the feature.
		*isRemovingStakingFeature = true
	}

	return nil
}

func accountFoundryCounterSTVF(vmParams *vm.Params, current *axongo.AccountOutput, next *axongo.AccountOutput) error {
	if current.FoundryCounter > next.FoundryCounter {
		return ierrors.WithMessagef(axongo.ErrAccountInvalidFoundryCounter,
			"foundry counter of next state is less than previous, in %d / out %d", current.FoundryCounter, next.FoundryCounter,
		)
	}

	// check that for a foundry counter change, X amount of foundries were actually created
	if current.FoundryCounter == next.FoundryCounter {
		return nil
	}

	var seenNewFoundriesOfAccount uint32
	for _, output := range vmParams.WorkingSet.Tx.Outputs {
		foundryOutput, is := output.(*axongo.FoundryOutput)
		if !is {
			continue
		}

		if _, notNew := vmParams.WorkingSet.InChains[foundryOutput.MustFoundryID()]; notNew {
			continue
		}

		//nolint:forcetypeassert // we can safely assume that this is an AccountAddress
		foundryAccountID := foundryOutput.Owner().(*axongo.AccountAddress).ChainID()
		if !foundryAccountID.Matches(next.AccountID) {
			continue
		}
		seenNewFoundriesOfAccount++
	}

	expectedNewFoundriesCount := next.FoundryCounter - current.FoundryCounter
	if expectedNewFoundriesCount != seenNewFoundriesOfAccount {
		return ierrors.WithMessagef(axongo.ErrAccountInvalidFoundryCounter,
			"%d new foundries were created but the account output's foundry counter changed by %d",
			seenNewFoundriesOfAccount,
			expectedNewFoundriesCount,
		)
	}

	return nil
}

func accountDestructionValid(vmParams *vm.Params, input *vm.ChainOutputWithIDs, isRemovingStakingFeature *bool) error {
	if vmParams.WorkingSet.Tx.Capabilities.CannotDestroyAccountOutputs() {
		return axongo.ErrTxCapabilitiesAccountDestructionNotAllowed
	}

	//nolint:forcetypeassert // we can safely assume that this is an AccountOutput
	outputToDestroy := input.Output.(*axongo.AccountOutput)

	blockIssuerFeat := outputToDestroy.FeatureSet().BlockIssuer()
	if blockIssuerFeat != nil {
		if vmParams.WorkingSet.Commitment == nil {
			return axongo.ErrBlockIssuerCommitmentInputMissing
		}

		if blockIssuerFeat.ExpirySlot >= vmParams.WorkingSet.Commitment.Slot {
			return ierrors.WithMessagef(axongo.ErrBlockIssuerNotExpired, "commitment slot: %d, expiry slot: %d",
				vmParams.WorkingSet.Commitment.Slot, blockIssuerFeat.ExpirySlot)
		}

		if err := accountBlockIssuanceCreditLocked(input, vmParams.WorkingSet.BIC); err != nil {
			return err
		}
	}

	stakingFeat := outputToDestroy.FeatureSet().Staking()
	if stakingFeat != nil {
		// This case should never occur as the staking feature requires the presence of a block issuer feature,
		// which also requires a commitment input.
		commitment := vmParams.WorkingSet.Commitment
		if commitment == nil {
			return axongo.ErrStakingCommitmentInputMissing
		}

		timeProvider := vmParams.API.TimeProvider()
		futureBoundedSlot := vmParams.FutureBoundedSlotIndex(commitment.Slot)
		futureBoundedEpoch := timeProvider.EpochFromSlot(futureBoundedSlot)

		if futureBoundedEpoch <= stakingFeat.EndEpoch {
			return ierrors.WithMessagef(
				axongo.ErrStakingFeatureRemovedBeforeUnbonding, "future bounded epoch is %d, must be > %d", futureBoundedEpoch, stakingFeat.EndEpoch,
			)
		}

		*isRemovingStakingFeature = true
	}

	return nil
}

func accountBlockIssuanceCreditLocked(input *vm.ChainOutputWithIDs, bicSet vm.BlockIssuanceCreditInputSet) error {
	accountID, is := input.ChainID.(axongo.AccountID)
	if !is {
		return ierrors.WithMessagef(axongo.ErrBlockIssuanceCreditInputMissing, "cannot convert chain ID %s to account ID",
			input.ChainID.ToHex())
	}

	if bic, exists := bicSet[accountID]; !exists {
		return axongo.ErrBlockIssuanceCreditInputMissing
	} else if bic < 0 {
		return axongo.ErrAccountLocked
	}

	return nil
}

// For output AnchorOutput(s) with non-zeroed AnchorID, there must be a corresponding input AnchorOutput where either its
// AnchorID is zeroed and StateIndex is zero or an input AnchorOutput with the same AnchorID.
//
// On anchor state transitions: The StateIndex must be incremented by 1 and Only Amount, StateIndex and StateMetadata can be mutated.
//
// On anchor governance transition: Only StateController, GovernanceController and the MetadataBlock can be mutated.
func anchorSTVF(vmParams *vm.Params, input *vm.ChainOutputWithIDs, transType axongo.ChainTransitionType, next *axongo.AnchorOutput) error {
	switch transType {
	case axongo.ChainTransitionTypeGenesis:
		if err := anchorGenesisValid(vmParams, next, true); err != nil {
			return err
		}
	case axongo.ChainTransitionTypeStateChange:
		if err := anchorStateChangeValid(input, next, vmParams); err != nil {
			return err
		}
	case axongo.ChainTransitionTypeDestroy:
		if err := anchorDestructionValid(vmParams); err != nil {
			return err
		}
	default:
		panic("unknown chain transition type in AnchorOutput")
	}

	return nil
}

func anchorGenesisValid(vmParams *vm.Params, current *axongo.AnchorOutput, anchorIDMustBeZeroed bool) error {
	if anchorIDMustBeZeroed && !current.AnchorID.Empty() {
		return ierrors.Join(axongo.ErrAnchorInvalidStateTransition, axongo.ErrNewChainOutputHasNonZeroedID)
	}

	return vm.IsIssuerOnOutputUnlocked(current, vmParams.WorkingSet.UnlockedAddrs)
}

func anchorStateChangeValid(input *vm.ChainOutputWithIDs, next *axongo.AnchorOutput, vmParams *vm.Params) error {
	//nolint:forcetypeassert // we can safely assume that this is an AnchorOutput
	current := input.Output.(*axongo.AnchorOutput)

	isGovTransition := current.StateIndex == next.StateIndex
	if !current.ImmutableFeatures.Equal(next.ImmutableFeatures) {
		err := axongo.ErrAnchorInvalidStateTransition
		if isGovTransition {
			err = axongo.ErrAnchorInvalidGovernanceTransition
		}

		return ierrors.Join(err,
			ierrors.WithMessagef(
				axongo.ErrChainOutputImmutableFeaturesChanged,
				"old state %s, next state %s", current.ImmutableFeatures, next.ImmutableFeatures,
			))
	}

	if isGovTransition {
		return anchorGovernanceSTVF(input, next)
	}

	return anchorStateSTVF(input, next, vmParams)
}

func anchorGovernanceSTVF(input *vm.ChainOutputWithIDs, next *axongo.AnchorOutput) error {
	//nolint:forcetypeassert // we can safely assume that this is an AnchorOutput
	current := input.Output.(*axongo.AnchorOutput)

	switch {
	case current.Amount != next.Amount:
		return ierrors.WithMessagef(axongo.ErrAnchorInvalidGovernanceTransition, "amount changed, in %d / out %d ", current.Amount, next.Amount)
	case current.StateIndex != next.StateIndex:
		return ierrors.WithMessagef(axongo.ErrAnchorInvalidGovernanceTransition, "state index changed, in %d / out %d", current.StateIndex, next.StateIndex)
	}

	if err := axongo.FeatureUnchanged(axongo.FeatureStateMetadata, current.Features.MustSet(), next.Features.MustSet()); err != nil {
		return ierrors.Join(axongo.ErrAnchorInvalidGovernanceTransition, err)
	}

	return nil
}

func anchorStateSTVF(input *vm.ChainOutputWithIDs, next *axongo.AnchorOutput, vmParams *vm.Params) error {
	//nolint:forcetypeassert // we can safely assume that this is an AnchorOutput
	current := input.Output.(*axongo.AnchorOutput)
	switch {
	case !current.StateController().Equal(next.StateController()):
		return ierrors.WithMessagef(axongo.ErrAnchorInvalidStateTransition,
			"state controller changed, in %s / out %s",
			current.StateController().Bech32(vmParams.API.ProtocolParameters().Bech32HRP()),
			next.StateController().Bech32(vmParams.API.ProtocolParameters().Bech32HRP()),
		)
	case !current.GovernorAddress().Equal(next.GovernorAddress()):
		return ierrors.WithMessagef(axongo.ErrAnchorInvalidStateTransition,
			"governance controller changed, in %s / out %s",
			current.GovernorAddress().Bech32(vmParams.API.ProtocolParameters().Bech32HRP()),
			next.GovernorAddress().Bech32(vmParams.API.ProtocolParameters().Bech32HRP()),
		)
	case current.StateIndex+1 != next.StateIndex:
		return ierrors.WithMessagef(axongo.ErrAnchorInvalidStateTransition, "state index %d on the input side but %d on the output side", current.StateIndex, next.StateIndex)
	}

	if err := axongo.FeatureUnchanged(axongo.FeatureMetadata, current.Features.MustSet(), next.Features.MustSet()); err != nil {
		return ierrors.Join(axongo.ErrAnchorInvalidStateTransition, err)
	}

	return nil
}

func anchorDestructionValid(vmParams *vm.Params) error {
	if vmParams.WorkingSet.Tx.Capabilities.CannotDestroyAnchorOutputs() {
		return ierrors.Join(axongo.ErrAnchorInvalidStateTransition, axongo.ErrTxCapabilitiesAnchorDestructionNotAllowed)
	}

	return nil
}

func foundrySTVF(vmParams *vm.Params, input *vm.ChainOutputWithIDs, transType axongo.ChainTransitionType, next *axongo.FoundryOutput) error {
	inSums := vmParams.WorkingSet.InNativeTokens
	outSums := vmParams.WorkingSet.OutNativeTokens

	switch transType {
	case axongo.ChainTransitionTypeGenesis:
		if err := foundryGenesisValid(vmParams, next, next.MustFoundryID(), outSums); err != nil {
			return err
		}
	case axongo.ChainTransitionTypeStateChange:
		//nolint:forcetypeassert // we can safely assume that this is a FoundryOutput
		current := input.Output.(*axongo.FoundryOutput)
		if err := foundryStateChangeValid(current, next, inSums, outSums); err != nil {
			return err
		}
	case axongo.ChainTransitionTypeDestroy:
		//nolint:forcetypeassert // we can safely assume that this is a FoundryOutput
		current := input.Output.(*axongo.FoundryOutput)
		if err := foundryDestructionValid(vmParams, current, inSums, outSums); err != nil {
			return err
		}
	default:
		panic("unknown chain transition type in FoundryOutput")
	}

	return nil
}

func foundryGenesisValid(vmParams *vm.Params, current *axongo.FoundryOutput, thisFoundryID axongo.FoundryID, outSums axongo.NativeTokenSum) error {
	nativeTokenID := current.MustNativeTokenID()
	if err := current.TokenScheme.StateTransition(axongo.ChainTransitionTypeGenesis, nil, nil, outSums.ValueOrBigInt0(nativeTokenID)); err != nil {
		return err
	}

	// grab foundry counter from transitioning AccountOutput
	//nolint:forcetypeassert // we can safely assume that this is an AccountAddress
	accountID := current.Owner().(*axongo.AccountAddress).AccountID()
	inAccount, ok := vmParams.WorkingSet.InChains[accountID]
	if !ok {
		return ierrors.WithMessagef(axongo.ErrFoundryTransitionWithoutAccount,
			"missing input transitioning account output %s for new foundry output %s", accountID, thisFoundryID.ToHex(),
		)
	}

	outAccount, ok := vmParams.WorkingSet.OutChains[accountID]
	if !ok {
		return ierrors.WithMessagef(axongo.ErrFoundryTransitionWithoutAccount,
			"missing output transitioning account output %s for new foundry output %s", accountID, thisFoundryID.ToHex(),
		)
	}

	//nolint:forcetypeassert // we can safely assume that this is an AccountOutput
	return foundrySerialNumberValid(vmParams, current, inAccount.Output.(*axongo.AccountOutput), outAccount.(*axongo.AccountOutput), thisFoundryID)
}

func foundrySerialNumberValid(vmParams *vm.Params, current *axongo.FoundryOutput, inAccount *axongo.AccountOutput, outAccount *axongo.AccountOutput, thisFoundryID axongo.FoundryID) error {
	// this new foundry's serial number must be between the given foundry counter interval
	startSerial := inAccount.FoundryCounter
	endIncSerial := outAccount.FoundryCounter
	if startSerial >= current.SerialNumber || current.SerialNumber > endIncSerial {
		return ierrors.WithMessagef(
			axongo.ErrFoundrySerialInvalid,
			"new foundry output %s's serial number %d is not between the foundry counter interval of [%d,%d)", thisFoundryID.ToHex(), current.SerialNumber, startSerial, endIncSerial,
		)
	}

	// OPTIMIZE: this loop happens on every STVF of every new foundry output
	// check order of serial number
	for outputIndex, output := range vmParams.WorkingSet.Tx.Outputs {
		otherFoundryOutput, is := output.(*axongo.FoundryOutput)
		if !is {
			continue
		}

		if !otherFoundryOutput.Owner().Equal(current.Owner()) {
			continue
		}

		otherFoundryID, err := otherFoundryOutput.FoundryID()
		if err != nil {
			return err
		}

		if _, isNotNew := vmParams.WorkingSet.InChains[otherFoundryID]; isNotNew {
			continue
		}

		// only check up to own foundry whether it is ordered
		if otherFoundryID == thisFoundryID {
			break
		}

		if otherFoundryOutput.SerialNumber >= current.SerialNumber {
			return ierrors.WithMessagef(
				axongo.ErrFoundrySerialInvalid,
				"new foundry output %s at index %d has bigger equal serial number than this foundry %s", otherFoundryID.ToHex(), outputIndex, thisFoundryID.ToHex(),
			)
		}
	}

	return nil
}

func foundryStateChangeValid(current *axongo.FoundryOutput, next *axongo.FoundryOutput, inSums axongo.NativeTokenSum, outSums axongo.NativeTokenSum) error {
	if !current.ImmutableFeatures.Equal(next.ImmutableFeatures) {
		return ierrors.WithMessagef(
			axongo.ErrChainOutputImmutableFeaturesChanged,
			"old state %s, next state %s", current.ImmutableFeatures, next.ImmutableFeatures,
		)
	}

	// the check for the serial number and token scheme not being mutated is implicit
	// as a change would cause the foundry ID to be different, which would result in
	// no matching foundry to be found to validate the state transition against
	if current.MustFoundryID() != next.MustFoundryID() {
		// impossible invariant as the STVF should be called via the matching next foundry output
		panic(fmt.Sprintf("foundry IDs mismatch in state transition validation function: have %s got %s", current.MustFoundryID().ToHex(), next.MustFoundryID().ToHex()))
	}

	nativeTokenID := current.MustNativeTokenID()

	return current.TokenScheme.StateTransition(axongo.ChainTransitionTypeStateChange, next.TokenScheme, inSums.ValueOrBigInt0(nativeTokenID), outSums.ValueOrBigInt0(nativeTokenID))
}

func foundryDestructionValid(vmParams *vm.Params, current *axongo.FoundryOutput, inSums axongo.NativeTokenSum, outSums axongo.NativeTokenSum) error {
	if vmParams.WorkingSet.Tx.Capabilities.CannotDestroyFoundryOutputs() {
		return axongo.ErrTxCapabilitiesFoundryDestructionNotAllowed
	}

	nativeTokenID := current.MustNativeTokenID()

	return current.TokenScheme.StateTransition(axongo.ChainTransitionTypeDestroy, nil, inSums.ValueOrBigInt0(nativeTokenID), outSums.ValueOrBigInt0(nativeTokenID))
}

func nftSTVF(vmParams *vm.Params, input *vm.ChainOutputWithIDs, transType axongo.ChainTransitionType, next *axongo.NFTOutput) error {
	switch transType {
	case axongo.ChainTransitionTypeGenesis:
		if err := nftGenesisValid(vmParams, next); err != nil {
			return err
		}
	case axongo.ChainTransitionTypeStateChange:
		//nolint:forcetypeassert // we can safely assume that this is an NFTOutput
		current := input.Output.(*axongo.NFTOutput)
		if err := nftStateChangeValid(current, next); err != nil {
			return err
		}
	case axongo.ChainTransitionTypeDestroy:
		if err := nftDestructionValid(vmParams); err != nil {
			return err
		}
	default:
		panic("unknown chain transition type in NFTOutput")
	}

	return nil
}

func nftGenesisValid(vmParams *vm.Params, current *axongo.NFTOutput) error {
	if !current.NFTID.Empty() {
		return axongo.ErrNewChainOutputHasNonZeroedID
	}

	return vm.IsIssuerOnOutputUnlocked(current, vmParams.WorkingSet.UnlockedAddrs)
}

func nftStateChangeValid(current *axongo.NFTOutput, next *axongo.NFTOutput) error {
	if !current.ImmutableFeatures.Equal(next.ImmutableFeatures) {
		return ierrors.WithMessagef(axongo.ErrChainOutputImmutableFeaturesChanged,
			"old state %s, next state %s", current.ImmutableFeatures, next.ImmutableFeatures,
		)
	}

	return nil
}

func nftDestructionValid(vmParams *vm.Params) error {
	if vmParams.WorkingSet.Tx.Capabilities.CannotDestroyNFTOutputs() {
		return axongo.ErrTxCapabilitiesNFTDestructionNotAllowed
	}

	return nil
}

func delegationSTVF(vmParams *vm.Params, input *vm.ChainOutputWithIDs, transType axongo.ChainTransitionType, next *axongo.DelegationOutput) error {
	switch transType {
	case axongo.ChainTransitionTypeGenesis:
		if err := delegationGenesisValid(vmParams, next); err != nil {
			return err
		}
	case axongo.ChainTransitionTypeStateChange:
		_, isClaiming := vmParams.WorkingSet.Rewards[input.ChainID]
		if isClaiming {
			return ierrors.WithMessage(axongo.ErrDelegationRewardsClaimingInvalid, "cannot claim rewards during delegation output transition")
		}
		//nolint:forcetypeassert // we can safely assume that this is an DelegationOutput
		current := input.Output.(*axongo.DelegationOutput)
		if err := delegationStateChangeValid(vmParams, current, next); err != nil {
			return err
		}
	case axongo.ChainTransitionTypeDestroy:
		_, isClaiming := vmParams.WorkingSet.Rewards[input.ChainID]
		if !isClaiming {
			return ierrors.WithMessage(axongo.ErrDelegationRewardInputMissing, "cannot destroy delegation output without a rewards input")
		}

		return nil
	default:
		panic("unknown chain transition type in DelegationOutput")
	}

	return nil
}

func delegationGenesisValid(vmParams *vm.Params, current *axongo.DelegationOutput) error {
	if !current.DelegationID.Empty() {
		return axongo.ErrNewChainOutputHasNonZeroedID
	}

	timeProvider := vmParams.API.TimeProvider()
	commitment := vmParams.WorkingSet.Commitment
	if commitment == nil {
		return axongo.ErrDelegationCommitmentInputMissing
	}
	pastBoundedSlot := vmParams.PastBoundedSlotIndex(commitment.Slot)
	pastBoundedEpoch := timeProvider.EpochFromSlot(pastBoundedSlot)
	registrationSlot := registrationSlot(vmParams, pastBoundedEpoch)

	var expectedStartEpoch axongo.EpochIndex
	if pastBoundedSlot <= registrationSlot {
		expectedStartEpoch = pastBoundedEpoch + 1
	} else {
		expectedStartEpoch = pastBoundedEpoch + 2
	}

	if current.StartEpoch != expectedStartEpoch {
		return ierrors.WithMessagef(axongo.ErrDelegationStartEpochInvalid, "is %d, expected %d", current.StartEpoch, expectedStartEpoch)
	}

	if current.DelegatedAmount != current.Amount {
		return ierrors.WithMessagef(axongo.ErrDelegationAmountMismatch, "delegated amount %d, amount %d", current.DelegatedAmount, current.Amount)
	}

	if current.EndEpoch != 0 {
		return axongo.ErrDelegationEndEpochNotZero
	}

	return nil
}

func delegationStateChangeValid(vmParams *vm.Params, current *axongo.DelegationOutput, next *axongo.DelegationOutput) error {
	// State transitioning a Delegation Output is always a transition to the delayed claiming state.
	// Since they can only be transitioned once, the input will always need to have a zeroed ID.
	if !current.DelegationID.Empty() {
		return ierrors.WithMessagef(axongo.ErrDelegationOutputTransitionedTwice,
			"delegation output can only be transitioned if it has a zeroed ID",
		)
	}

	if current.DelegatedAmount != next.DelegatedAmount ||
		!current.ValidatorAddress.Equal(next.ValidatorAddress) ||
		current.StartEpoch != next.StartEpoch {
		return axongo.ErrDelegationModified
	}

	timeProvider := vmParams.API.TimeProvider()
	commitment := vmParams.WorkingSet.Commitment
	if commitment == nil {
		return axongo.ErrDelegationCommitmentInputMissing
	}
	futureBoundedSlot := vmParams.FutureBoundedSlotIndex(commitment.Slot)
	futureBoundedEpoch := timeProvider.EpochFromSlot(futureBoundedSlot)
	registrationSlot := registrationSlot(vmParams, futureBoundedEpoch)

	var expectedEndEpoch axongo.EpochIndex
	if futureBoundedSlot <= registrationSlot {
		expectedEndEpoch = futureBoundedEpoch
	} else {
		expectedEndEpoch = futureBoundedEpoch + 1
	}

	if next.EndEpoch != expectedEndEpoch {
		return ierrors.WithMessagef(axongo.ErrDelegationEndEpochInvalid, "is %d, expected %d", next.EndEpoch, expectedEndEpoch)
	}

	return nil
}

// registrationSlot returns the slot at the end of which the validator and delegator registration ends and the voting power
// for the epoch with index epoch + 1 is calculated.
func registrationSlot(vmParams *vm.Params, epoch axongo.EpochIndex) axongo.SlotIndex {
	return vmParams.API.TimeProvider().EpochEnd(epoch) - vmParams.API.ProtocolParameters().EpochNearingThreshold()
}

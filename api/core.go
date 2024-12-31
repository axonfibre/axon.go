package api

import (
	"time"

	"github.com/axonfibre/fibre.go/ierrors"
	"github.com/axonfibre/fibre.go/serializer/v2"
	iotago "github.com/axonfibre/axon.go/v4"
)

type BlockState byte

const (
	// BlockStateUnknown indicates that the state of the block can not be determined by the node for some reason.
	BlockStateUnknown BlockState = iota
	// BlockStatePending indicates that the block has been booked by the node but not yet accepted.
	BlockStatePending
	// BlockStateAccepted indicates that the block has been referenced by the super majority of the online committee.
	BlockStateAccepted
	// BlockStateConfirmed indicates that the block has been referenced by the super majority of the total committee.
	BlockStateConfirmed
	// BlockStateFinalized indicates that the commitment containing the block has been finalized.
	// This state is computed based on the accepted/confirmed block's slot being smaller or equal than the latest finalized slot.
	BlockStateFinalized
	// BlockStateDropped indicates that the block has been dropped due to congestion control.
	BlockStateDropped
	// BlockStateOrphaned indicates that the block's slot has been committed by the node without the block being included.
	// In this case, the block will never be finalized unless there is a chain switch.
	// This state is computed based on the pending block's slot being smaller or equal than the latest committed slot.
	BlockStateOrphaned
)

func (b BlockState) String() string {
	return []string{
		"unknown",
		"pending",
		"accepted",
		"confirmed",
		"finalized",
		"dropped",
		"orphaned",
	}[b]
}

func (b BlockState) Bytes() ([]byte, error) {
	return []byte{byte(b)}, nil
}

func (b BlockState) EncodeJSON() (any, error) {
	if b > BlockStateOrphaned {
		return nil, ierrors.Errorf("invalid block state: %d", b)
	}

	return b.String(), nil
}

func (b *BlockState) DecodeJSON(state any) error {
	if state == nil {
		return ierrors.New("given block state is nil")
	}

	blockState, ok := state.(string)
	if !ok {
		return ierrors.Errorf("invalid type: %T", state)
	}

	switch blockState {
	case "unknown":
		*b = BlockStateUnknown
	case "pending":
		*b = BlockStatePending
	case "accepted":
		*b = BlockStateAccepted
	case "confirmed":
		*b = BlockStateConfirmed
	case "finalized":
		*b = BlockStateFinalized
	case "dropped":
		*b = BlockStateDropped
	case "orphaned":
		*b = BlockStateOrphaned
	default:
		return ierrors.Errorf("invalid block state: %s", blockState)
	}

	return nil
}

type TransactionState byte
type TransactionFailureReason byte

const (
	TransactionStateLength         = serializer.OneByte
	TransactionFailureReasonLength = serializer.OneByte
)

const (
	// TransactionStateUnknown indicates that the state of the transaction can not be determined by the node for some reason.
	TransactionStateUnknown TransactionState = iota
	// TransactionStatePending indicates that the transaction has been booked by the node but not yet accepted.
	TransactionStatePending
	// TransactionStateAccepted indicates that the transaction meets the following 4 conditions:
	//	- Signatures of the transaction are valid.
	//  - The transaction has been approved by the super majority of the online committee (potential conflicts are resolved by this time).
	//	- The transactions that created the inputs were accepted (monotonicity).
	//  - At least one valid attachment was accepted.
	TransactionStateAccepted
	// TransactionStateCommitted indicates that the slot of the earliest accepted attachment of the transaction was committed.
	TransactionStateCommitted
	// TransactionStateFinalized indicates that the transaction is accepted and the slot containing the transaction has been finalized by the node.
	// This state is computed based on the accepted transaction's earliest included attachment slot being smaller or equal than the latest finalized slot.
	TransactionStateFinalized
	// TransactionStateFailed indicates that the transaction has not been executed by the node due to a failure during processing.
	TransactionStateFailed
)

func (t TransactionState) String() string {
	return []string{
		"unknown",
		"pending",
		"accepted",
		"committed",
		"finalized",
		"failed",
	}[t]
}

func (t TransactionState) Bytes() ([]byte, error) {
	return []byte{byte(t)}, nil
}

func TransactionStateFromBytes(b []byte) (TransactionState, int, error) {
	if len(b) < TransactionStateLength {
		return 0, 0, ierrors.New("invalid transaction state size")
	}

	return TransactionState(b[0]), TransactionStateLength, nil
}

func (t TransactionState) EncodeJSON() (any, error) {
	if t > TransactionStateFailed {
		return nil, ierrors.Errorf("invalid transaction state: %d", t)
	}

	return t.String(), nil
}

func (t *TransactionState) DecodeJSON(state any) error {
	if state == nil {
		return ierrors.New("given transaction state is nil")
	}

	transactionState, ok := state.(string)
	if !ok {
		return ierrors.Errorf("invalid type: %T", state)
	}

	switch transactionState {
	case "unknown":
		*t = TransactionStateUnknown
	case "pending":
		*t = TransactionStatePending
	case "accepted":
		*t = TransactionStateAccepted
	case "committed":
		*t = TransactionStateCommitted
	case "finalized":
		*t = TransactionStateFinalized
	case "failed":
		*t = TransactionStateFailed
	default:
		return ierrors.Errorf("invalid transaction state: %s", transactionState)
	}

	return nil
}

const (
	TxFailureNone TransactionFailureReason = 0

	TxFailureConflictRejected TransactionFailureReason = 1
	TxFailureOrphaned         TransactionFailureReason = 2

	TxFailureInputAlreadySpent            TransactionFailureReason = 3
	TxFailureInputCreationAfterTxCreation TransactionFailureReason = 4
	TxFailureUnlockSignatureInvalid       TransactionFailureReason = 5

	TxFailureChainAddressUnlockInvalid            TransactionFailureReason = 6
	TxFailureDirectUnlockableAddressUnlockInvalid TransactionFailureReason = 7
	TxFailureMultiAddressUnlockInvalid            TransactionFailureReason = 8

	TxFailureCommitmentInputReferenceInvalid TransactionFailureReason = 9
	TxFailureBICInputReferenceInvalid        TransactionFailureReason = 10
	TxFailureRewardInputReferenceInvalid     TransactionFailureReason = 11

	TxFailureStakingRewardCalculationFailure    TransactionFailureReason = 12
	TxFailureDelegationRewardCalculationFailure TransactionFailureReason = 13

	TxFailureInputOutputBaseTokenMismatch TransactionFailureReason = 14

	TxFailureManaOverflow                             TransactionFailureReason = 15
	TxFailureInputOutputManaMismatch                  TransactionFailureReason = 16
	TxFailureManaDecayCreationIndexExceedsTargetIndex TransactionFailureReason = 17

	TxFailureNativeTokenSumUnbalanced TransactionFailureReason = 18

	TxFailureSimpleTokenSchemeMintedMeltedTokenDecrease TransactionFailureReason = 19
	TxFailureSimpleTokenSchemeMintingInvalid            TransactionFailureReason = 20
	TxFailureSimpleTokenSchemeMeltingInvalid            TransactionFailureReason = 21
	TxFailureSimpleTokenSchemeMaximumSupplyChanged      TransactionFailureReason = 22
	TxFailureSimpleTokenSchemeGenesisInvalid            TransactionFailureReason = 23

	TxFailureMultiAddressLengthUnlockLengthMismatch TransactionFailureReason = 24
	TxFailureMultiAddressUnlockThresholdNotReached  TransactionFailureReason = 25

	TxFailureSenderFeatureNotUnlocked TransactionFailureReason = 26

	TxFailureIssuerFeatureNotUnlocked TransactionFailureReason = 27

	TxFailureStakingRewardInputMissing             TransactionFailureReason = 28
	TxFailureStakingCommitmentInputMissing         TransactionFailureReason = 29
	TxFailureStakingRewardClaimingInvalid          TransactionFailureReason = 30
	TxFailureStakingFeatureRemovedBeforeUnbonding  TransactionFailureReason = 31
	TxFailureStakingFeatureModifiedBeforeUnbonding TransactionFailureReason = 32
	TxFailureStakingStartEpochInvalid              TransactionFailureReason = 33
	TxFailureStakingEndEpochTooEarly               TransactionFailureReason = 34

	TxFailureBlockIssuerCommitmentInputMissing TransactionFailureReason = 35
	TxFailureBlockIssuanceCreditInputMissing   TransactionFailureReason = 36
	TxFailureBlockIssuerNotExpired             TransactionFailureReason = 37
	TxFailureBlockIssuerExpiryTooEarly         TransactionFailureReason = 38
	TxFailureManaMovedOffBlockIssuerAccount    TransactionFailureReason = 39
	TxFailureAccountLocked                     TransactionFailureReason = 40

	TxFailureTimelockCommitmentInputMissing TransactionFailureReason = 41
	TxFailureTimelockNotExpired             TransactionFailureReason = 42

	TxFailureExpirationCommitmentInputMissing TransactionFailureReason = 43
	TxFailureExpirationNotUnlockable          TransactionFailureReason = 44

	TxFailureReturnAmountNotFulFilled TransactionFailureReason = 45

	TxFailureNewChainOutputHasNonZeroedID        TransactionFailureReason = 46
	TxFailureChainOutputImmutableFeaturesChanged TransactionFailureReason = 47

	TxFailureImplicitAccountDestructionDisallowed     TransactionFailureReason = 48
	TxFailureMultipleImplicitAccountCreationAddresses TransactionFailureReason = 49

	TxFailureAccountInvalidFoundryCounter TransactionFailureReason = 50

	TxFailureAnchorInvalidStateTransition      TransactionFailureReason = 51
	TxFailureAnchorInvalidGovernanceTransition TransactionFailureReason = 52

	TxFailureFoundryTransitionWithoutAccount TransactionFailureReason = 53
	TxFailureFoundrySerialInvalid            TransactionFailureReason = 54

	TxFailureDelegationCommitmentInputMissing  TransactionFailureReason = 55
	TxFailureDelegationRewardInputMissing      TransactionFailureReason = 56
	TxFailureDelegationRewardsClaimingInvalid  TransactionFailureReason = 57
	TxFailureDelegationOutputTransitionedTwice TransactionFailureReason = 58
	TxFailureDelegationModified                TransactionFailureReason = 59
	TxFailureDelegationStartEpochInvalid       TransactionFailureReason = 60
	TxFailureDelegationAmountMismatch          TransactionFailureReason = 61
	TxFailureDelegationEndEpochNotZero         TransactionFailureReason = 62
	TxFailureDelegationEndEpochInvalid         TransactionFailureReason = 63

	TxFailureCapabilitiesNativeTokenBurningNotAllowed TransactionFailureReason = 64
	TxFailureCapabilitiesManaBurningNotAllowed        TransactionFailureReason = 65
	TxFailureCapabilitiesAccountDestructionNotAllowed TransactionFailureReason = 66
	TxFailureCapabilitiesAnchorDestructionNotAllowed  TransactionFailureReason = 67
	TxFailureCapabilitiesFoundryDestructionNotAllowed TransactionFailureReason = 68
	TxFailureCapabilitiesNFTDestructionNotAllowed     TransactionFailureReason = 69

	TxFailureSemanticValidationFailed TransactionFailureReason = 255
)

var txErrorsFailureReasonMap = map[error]TransactionFailureReason{
	// ================================
	// Pre-Transaction Execution Errors
	// ================================

	// tx level errors
	iotago.ErrTxConflictRejected: TxFailureConflictRejected,
	iotago.ErrTxOrphaned:         TxFailureOrphaned,

	// input
	iotago.ErrInputAlreadySpent:            TxFailureInputAlreadySpent,
	iotago.ErrInputCreationAfterTxCreation: TxFailureInputCreationAfterTxCreation,
	iotago.ErrUnlockSignatureInvalid:       TxFailureUnlockSignatureInvalid,

	// unlock
	iotago.ErrChainAddressUnlockInvalid:            TxFailureChainAddressUnlockInvalid,
	iotago.ErrDirectUnlockableAddressUnlockInvalid: TxFailureDirectUnlockableAddressUnlockInvalid,
	iotago.ErrMultiAddressUnlockInvalid:            TxFailureMultiAddressUnlockInvalid,

	// context inputs
	iotago.ErrCommitmentInputReferenceInvalid: TxFailureCommitmentInputReferenceInvalid,
	iotago.ErrBICInputReferenceInvalid:        TxFailureBICInputReferenceInvalid,
	iotago.ErrRewardInputReferenceInvalid:     TxFailureRewardInputReferenceInvalid,

	// reward calculation
	iotago.ErrStakingRewardCalculationFailure:    TxFailureStakingRewardCalculationFailure,
	iotago.ErrDelegationRewardCalculationFailure: TxFailureDelegationRewardCalculationFailure,

	// ============================
	// Transaction Execution Errors
	// ============================

	// amount
	iotago.ErrInputOutputBaseTokenMismatch: TxFailureInputOutputBaseTokenMismatch,

	// mana
	iotago.ErrManaOverflow:                             TxFailureManaOverflow,
	iotago.ErrInputOutputManaMismatch:                  TxFailureInputOutputManaMismatch,
	iotago.ErrManaDecayCreationIndexExceedsTargetIndex: TxFailureManaDecayCreationIndexExceedsTargetIndex,

	// native token
	iotago.ErrNativeTokenSumUnbalanced: TxFailureNativeTokenSumUnbalanced,

	// simple token scheme
	iotago.ErrSimpleTokenSchemeMintedMeltedTokenDecrease: TxFailureSimpleTokenSchemeMintedMeltedTokenDecrease,
	iotago.ErrSimpleTokenSchemeMintingInvalid:            TxFailureSimpleTokenSchemeMintingInvalid,
	iotago.ErrSimpleTokenSchemeMeltingInvalid:            TxFailureSimpleTokenSchemeMeltingInvalid,
	iotago.ErrSimpleTokenSchemeMaximumSupplyChanged:      TxFailureSimpleTokenSchemeMaximumSupplyChanged,
	iotago.ErrSimpleTokenSchemeGenesisInvalid:            TxFailureSimpleTokenSchemeGenesisInvalid,

	// multi address
	iotago.ErrMultiAddressLengthUnlockLengthMismatch: TxFailureMultiAddressLengthUnlockLengthMismatch,
	iotago.ErrMultiAddressUnlockThresholdNotReached:  TxFailureMultiAddressUnlockThresholdNotReached,

	// sender feature
	iotago.ErrSenderFeatureNotUnlocked: TxFailureSenderFeatureNotUnlocked,

	// issuer feature
	iotago.ErrIssuerFeatureNotUnlocked: TxFailureIssuerFeatureNotUnlocked,

	// staking feature
	iotago.ErrStakingRewardInputMissing:             TxFailureStakingRewardInputMissing,
	iotago.ErrStakingCommitmentInputMissing:         TxFailureStakingCommitmentInputMissing,
	iotago.ErrStakingRewardClaimingInvalid:          TxFailureStakingRewardClaimingInvalid,
	iotago.ErrStakingFeatureRemovedBeforeUnbonding:  TxFailureStakingFeatureRemovedBeforeUnbonding,
	iotago.ErrStakingFeatureModifiedBeforeUnbonding: TxFailureStakingFeatureModifiedBeforeUnbonding,
	iotago.ErrStakingStartEpochInvalid:              TxFailureStakingStartEpochInvalid,
	iotago.ErrStakingEndEpochTooEarly:               TxFailureStakingEndEpochTooEarly,

	// block issuer feature
	iotago.ErrBlockIssuerCommitmentInputMissing: TxFailureBlockIssuerCommitmentInputMissing,
	iotago.ErrBlockIssuanceCreditInputMissing:   TxFailureBlockIssuanceCreditInputMissing,
	iotago.ErrBlockIssuerNotExpired:             TxFailureBlockIssuerNotExpired,
	iotago.ErrBlockIssuerExpiryTooEarly:         TxFailureBlockIssuerExpiryTooEarly,
	iotago.ErrManaMovedOffBlockIssuerAccount:    TxFailureManaMovedOffBlockIssuerAccount,
	iotago.ErrAccountLocked:                     TxFailureAccountLocked,

	// timelock unlock condition
	iotago.ErrTimelockCommitmentInputMissing: TxFailureTimelockCommitmentInputMissing,
	iotago.ErrTimelockNotExpired:             TxFailureTimelockNotExpired,

	// expiration unlock condition
	iotago.ErrExpirationCommitmentInputMissing: TxFailureExpirationCommitmentInputMissing,
	iotago.ErrExpirationNotUnlockable:          TxFailureExpirationNotUnlockable,

	// storage deposit return unlock condition
	iotago.ErrReturnAmountNotFulFilled: TxFailureReturnAmountNotFulFilled,

	// generic chain output errors
	iotago.ErrNewChainOutputHasNonZeroedID:        TxFailureNewChainOutputHasNonZeroedID,
	iotago.ErrChainOutputImmutableFeaturesChanged: TxFailureChainOutputImmutableFeaturesChanged,

	// implicit account
	iotago.ErrImplicitAccountDestructionDisallowed:     TxFailureImplicitAccountDestructionDisallowed,
	iotago.ErrMultipleImplicitAccountCreationAddresses: TxFailureMultipleImplicitAccountCreationAddresses,

	// account
	iotago.ErrAccountInvalidFoundryCounter: TxFailureAccountInvalidFoundryCounter,

	iotago.ErrAnchorInvalidStateTransition:      TxFailureAnchorInvalidStateTransition,
	iotago.ErrAnchorInvalidGovernanceTransition: TxFailureAnchorInvalidGovernanceTransition,

	// foundry
	iotago.ErrFoundryTransitionWithoutAccount: TxFailureFoundryTransitionWithoutAccount,
	iotago.ErrFoundrySerialInvalid:            TxFailureFoundrySerialInvalid,

	// delegation
	iotago.ErrDelegationCommitmentInputMissing:  TxFailureDelegationCommitmentInputMissing,
	iotago.ErrDelegationRewardInputMissing:      TxFailureDelegationRewardInputMissing,
	iotago.ErrDelegationRewardsClaimingInvalid:  TxFailureDelegationRewardsClaimingInvalid,
	iotago.ErrDelegationOutputTransitionedTwice: TxFailureDelegationOutputTransitionedTwice,
	iotago.ErrDelegationModified:                TxFailureDelegationModified,
	iotago.ErrDelegationStartEpochInvalid:       TxFailureDelegationStartEpochInvalid,
	iotago.ErrDelegationAmountMismatch:          TxFailureDelegationAmountMismatch,
	iotago.ErrDelegationEndEpochNotZero:         TxFailureDelegationEndEpochNotZero,
	iotago.ErrDelegationEndEpochInvalid:         TxFailureDelegationEndEpochInvalid,

	// tx capabilities
	iotago.ErrTxCapabilitiesNativeTokenBurningNotAllowed: TxFailureCapabilitiesNativeTokenBurningNotAllowed,
	iotago.ErrTxCapabilitiesManaBurningNotAllowed:        TxFailureCapabilitiesManaBurningNotAllowed,
	iotago.ErrTxCapabilitiesAccountDestructionNotAllowed: TxFailureCapabilitiesAccountDestructionNotAllowed,
	iotago.ErrTxCapabilitiesAnchorDestructionNotAllowed:  TxFailureCapabilitiesAnchorDestructionNotAllowed,
	iotago.ErrTxCapabilitiesFoundryDestructionNotAllowed: TxFailureCapabilitiesFoundryDestructionNotAllowed,
	iotago.ErrTxCapabilitiesNFTDestructionNotAllowed:     TxFailureCapabilitiesNFTDestructionNotAllowed,
}

func (t TransactionFailureReason) Bytes() ([]byte, error) {
	return []byte{byte(t)}, nil
}

func TransactionFailureReasonFromBytes(b []byte) (TransactionFailureReason, int, error) {
	if len(b) < TransactionFailureReasonLength {
		return 0, 0, ierrors.New("invalid transaction failure reason size")
	}

	return TransactionFailureReason(b[0]), TransactionFailureReasonLength, nil
}

func (t TransactionFailureReason) String() string {
	return []string{
		"none",

		"conflict_rejected",
		"orphaned",

		"input_already_spent",
		"input_creation_after_tx_creation",
		"unlock_signature_invalid",

		"chain_address_unlock_invalid",
		"direct_unlockable_address_unlock_invalid",
		"multi_address_unlock_invalid",

		"commitment_input_reference_invalid",
		"bic_input_reference_invalid",
		"reward_input_reference_invalid",

		"staking_reward_calculation_failure",
		"delegation_reward_calculation_failure",

		"input_output_base_token_mismatch",

		"mana_overflow",
		"input_output_mana_mismatch",
		"mana_decay_creation_index_exceeds_target_index",

		"native_token_sum_unbalanced",

		"simple_token_scheme_minted_melted_token_decrease",
		"simple_token_scheme_minting_invalid",
		"simple_token_scheme_melting_invalid",
		"simple_token_scheme_maximum_supply_changed",
		"simple_token_scheme_genesis_invalid",

		"multi_address_length_unlock_length_mismatch",
		"multi_address_unlock_threshold_not_reached",

		"sender_feature_not_unlocked",

		"issuer_feature_not_unlocked",

		"staking_reward_input_missing",
		"staking_commitment_input_missing",
		"staking_reward_claiming_invalid",
		"staking_feature_removed_before_unbonding",
		"staking_feature_modified_before_unbonding",
		"staking_start_epoch_invalid",
		"staking_end_epoch_too_early",

		"block_issuer_commitment_input_missing",
		"block_issuance_credit_input_missing",
		"block_issuer_not_expired",
		"block_issuer_expiry_too_early",
		"mana_moved_off_block_issuer_account",
		"account_locked",

		"timelock_commitment_input_missing",
		"timelock_not_expired",

		"expiration_commitment_input_missing",
		"expiration_not_unlockable",

		"return_amount_not_fulfilled",

		"new_chain_output_has_non_zeroed_id",
		"chain_output_immutable_features_changed",

		"implicit_account_destruction_disallowed",
		"multiple_implicit_account_creation_addresses",

		"account_invalid_foundry_counter",

		"anchor_invalid_state_transition",
		"anchor_invalid_governance_transition",

		"foundry_transition_without_account",
		"foundry_serial_invalid",

		"delegation_commitment_input_missing",
		"delegation_reward_input_missing",
		"delegation_rewards_claiming_invalid",
		"delegation_output_transitioned_twice",
		"delegation_modified",
		"delegation_start_epoch_invalid",
		"delegation_amount_mismatch",
		"delegation_end_epoch_not_zero",
		"delegation_end_epoch_invalid",

		"capabilities_native_token_burning_not_allowed",
		"capabilities_mana_burning_not_allowed",
		"capabilities_account_destruction_not_allowed",
		"capabilities_anchor_destruction_not_allowed",
		"capabilities_foundry_destruction_not_allowed",
		"capabilities_nft_destruction_not_allowed",

		"semantic_validation_failed",
	}[t]
}

// Unwraps the given err into the given errList by recursively unwrapping it.
//
// In case of joined errors, the right-most error is unwrapped first, which corresponds
// to a post-order depth-traversal of err's tree.
//
// This means errList will contain the most-detailed errors first (those lower in the error tree).
func unwrapErrors(err error, errList []error) []error {
	//nolint:errorlint // false positive: we're not switching on a specific error type.
	switch x := err.(type) {
	case interface{ Unwrap() []error }:
		errors := x.Unwrap()
		// Iterate the errors in reverse, so we walk the tree in post-order.
		for i := len(errors) - 1; i >= 0; i-- {
			err := errors[i]
			if err != nil {
				errList = unwrapErrors(err, errList)
				errList = append(errList, err)
			}
		}
	case interface{ Unwrap() error }:
		err = x.Unwrap()
		if err != nil {
			errList = unwrapErrors(err, errList)
			errList = append(errList, err)
		}
	}

	return errList
}

func DetermineTransactionFailureReason(err error) TransactionFailureReason {
	errorList := make([]error, 0)
	errorList = unwrapErrors(err, errorList)

	// Map the error to the transaction failure reason.
	// The strategy is to map the first failure reason that exists in order of most-detailed to least-detailed error.
	for _, err := range errorList {
		if txFailureReason, matches := txErrorsFailureReasonMap[err]; matches {
			return txFailureReason
		}
	}

	// Use most general failure reason if no other error matches.
	return TxFailureSemanticValidationFailed
}

type (
	// InfoResponse defines the response of a GET info REST API call.
	InfoResponse struct {
		// The name of the node software.
		Name string `serix:",lenPrefix=uint8"`
		// The semver version of the node software.
		Version string `serix:",lenPrefix=uint8"`
		// The current status of this node.
		Status *InfoResNodeStatus `serix:""`
		// The protocol parameters used by this node.
		ProtocolParameters []*InfoResProtocolParameters `serix:",lenPrefix=uint8"`
		// The base token of the network.
		BaseToken *InfoResBaseToken `serix:""`
	}

	// InfoResProtocolParameters defines the protocol parameters of a node in the InfoResponse.
	InfoResProtocolParameters struct {
		StartEpoch iotago.EpochIndex         `serix:""`
		Parameters iotago.ProtocolParameters `serix:""`
	}

	// InfoResNodeStatus defines the status of the node in the InfoResponse.
	InfoResNodeStatus struct {
		// Whether the node is healthy.
		IsHealthy bool `serix:""`
		// Whether the network is healthy (finalization is not delayed).
		IsNetworkHealthy bool `serix:""`
		// The accepted tangle time.
		AcceptedTangleTime time.Time `serix:""`
		// The relative accepted tangle time.
		RelativeAcceptedTangleTime time.Time `serix:""`
		// The confirmed tangle time.
		ConfirmedTangleTime time.Time `serix:""`
		// The relative confirmed tangle time.
		RelativeConfirmedTangleTime time.Time `serix:""`
		// The id of the latest known commitment.
		LatestCommitmentID iotago.CommitmentID `serix:""`
		// The latest finalized slot.
		LatestFinalizedSlot iotago.SlotIndex `serix:""`
		// The slot of the latest accepted block.
		LatestAcceptedBlockSlot iotago.SlotIndex `serix:""`
		// The slot of the latest confirmed block.
		LatestConfirmedBlockSlot iotago.SlotIndex `serix:""`
		// The epoch at which the tangle data was pruned.
		PruningEpoch iotago.EpochIndex `serix:""`
	}

	// InfoResBaseToken defines the base token of the node in the InfoResponse.
	InfoResBaseToken struct {
		// The base token name.
		Name string `serix:",lenPrefix=uint8"`
		// The base token ticker symbol.
		TickerSymbol string `serix:",lenPrefix=uint8"`
		// The base token unit.
		Unit string `serix:",lenPrefix=uint8"`
		// The base token subunit.
		Subunit string `serix:",lenPrefix=uint8,omitempty"`
		// The base token amount of decimals.
		Decimals uint32 `serix:""`
	}

	// NetworkHealthResponse defines the network health response.
	NetworkHealthResponse struct {
		// Whether the network is healthy (finalization is not delayed).
		IsNetworkHealthy bool `serix:""`
	}

	// NetworkMetricsResponse defines the network metrics response.
	NetworkMetricsResponse struct {
		// The current rate of new blocks per second, it's updated when a commitment is committed.
		BlocksPerSecond float64 `serix:""`
		// The current rate of confirmed blocks per second, it's updated when a commitment is committed.
		ConfirmedBlocksPerSecond float64 `serix:""`
		// The ratio of confirmed blocks in relation to new blocks up until the latest commitment is committed.
		ConfirmationRate float64 `serix:""`
	}

	// IssuanceBlockHeaderResponse defines the response of a GET block issuance REST API call.
	IssuanceBlockHeaderResponse struct {
		// StrongParents are the strong parents of the block.
		StrongParents iotago.BlockIDs `serix:",lenPrefix=uint8"`
		// WeakParents are the weak parents of the block.
		WeakParents iotago.BlockIDs `serix:",lenPrefix=uint8,omitempty"`
		// ShallowLikeParents are the shallow like parents of the block.
		ShallowLikeParents iotago.BlockIDs `serix:",lenPrefix=uint8,omitempty"`
		// LatestParentBlockIssuingTime is the latest issuing time of the returned parents.
		LatestParentBlockIssuingTime time.Time `serix:""`
		// LatestFinalizedSlot is the latest finalized slot.
		LatestFinalizedSlot iotago.SlotIndex `serix:""`
		// LatestCommitment is the latest commitment of the node.
		LatestCommitment *iotago.Commitment `serix:""`
	}

	// BlockCreatedResponse defines the response of a POST blocks REST API call.
	BlockCreatedResponse struct {
		// The hex encoded block ID of the block.
		BlockID iotago.BlockID `serix:""`
	}

	// BlockMetadataResponse defines the response of a GET block metadata REST API call.
	BlockMetadataResponse struct {
		// BlockID The hex encoded block ID of the block.
		BlockID iotago.BlockID `serix:""`
		// BlockState might be unknown, pending, accepted, confirmed, finalized, dropped, orphaned.
		BlockState BlockState `serix:""`
	}

	// BlockWithMetadataResponse defines the response of a GET full block REST API call.
	BlockWithMetadataResponse struct {
		Block    *iotago.Block          `serix:""`
		Metadata *BlockMetadataResponse `serix:""`
	}

	// TransactionMetadataResponse defines the response of a GET transaction metadata REST API call.
	TransactionMetadataResponse struct {
		// TransactionID is the hex encoded transaction ID of the transaction.
		TransactionID iotago.TransactionID `serix:""`
		// TransactionState might be pending, accepted, committed, finalized, failed.
		TransactionState TransactionState `serix:""`
		// EarliestAttachmentSlot is the slot of the earliest included valid block that contains an attachment of the transaction.
		EarliestAttachmentSlot iotago.SlotIndex `serix:""`
		// TransactionFailureReason if applicable indicates the error that occurred during the transaction processing.
		TransactionFailureReason TransactionFailureReason `serix:",omitempty"`
		// TransactionFailureDetails contains the detailed error message that occurred during the transaction processing
		// if the debug mode was activated in the retainer.
		TransactionFailureDetails string `serix:",omitempty,lenPrefix=uint16"`
	}

	// OutputResponse defines the response of a GET outputs REST API call.
	OutputResponse struct {
		Output        iotago.TxEssenceOutput `serix:""`
		OutputIDProof *iotago.OutputIDProof  `serix:""`
	}

	// OutputWithID returns an output with its corresponding ID.
	OutputWithID struct {
		OutputID iotago.OutputID        `serix:""`
		Output   iotago.TxEssenceOutput `serix:""`
	}

	OutputInclusionMetadata struct {
		// Slot is the slot in which the output was included.
		Slot iotago.SlotIndex `serix:""`
		// TransactionID is the transaction ID that created the output.
		TransactionID iotago.TransactionID `serix:""`
		// CommitmentID is the commitment ID that includes the creation of the output.
		CommitmentID iotago.CommitmentID `serix:",omitempty"`
	}

	OutputConsumptionMetadata struct {
		// Slot is the slot in which the output was spent.
		Slot iotago.SlotIndex `serix:""`
		// TransactionID is the transaction ID that spent the output.
		TransactionID iotago.TransactionID `serix:""`
		// CommitmentID is the commitment ID that includes the spending of the output.
		CommitmentID iotago.CommitmentID `serix:",omitempty"`
	}

	// OutputMetadata defines the response of a GET outputs metadata REST API call.
	OutputMetadata struct {
		// OutputID is the hex encoded output ID.
		OutputID iotago.OutputID `serix:""`
		// BlockID is the block ID that contains the output.
		BlockID iotago.BlockID `serix:""`
		// Included is the metadata of the output if it is included in the ledger.
		Included *OutputInclusionMetadata `serix:""`
		// Spent is the metadata of the output if it is marked as spent in the ledger.
		Spent *OutputConsumptionMetadata `serix:",optional,omitempty"`
		// LatestCommitmentID is the latest commitment ID of a node.
		LatestCommitmentID iotago.CommitmentID `serix:""`
	}

	// OutputWithMetadataResponse defines the response of a GET full outputs REST API call.
	OutputWithMetadataResponse struct {
		Output        iotago.TxEssenceOutput `serix:""`
		OutputIDProof *iotago.OutputIDProof  `serix:""`
		Metadata      *OutputMetadata        `serix:""`
	}

	// UTXOChangesResponse defines the response for the UTXO changes per slot REST API call.
	UTXOChangesResponse struct {
		// CommitmentID is the commitment ID of the requested slot that contains the changes.
		CommitmentID iotago.CommitmentID `serix:""`
		// The outputs that are created in this slot.
		CreatedOutputs iotago.OutputIDs `serix:",lenPrefix=uint32"`
		// The outputs that are consumed in this slot.
		ConsumedOutputs iotago.OutputIDs `serix:",lenPrefix=uint32"`
	}

	// UTXOChangesFullResponse defines the response for the UTXO changes per slot REST API call.
	// It returns the full information about the outputs with their corresponding ID.
	UTXOChangesFullResponse struct {
		// CommitmentID is the commitment ID of the requested slot that contains the changes.
		CommitmentID iotago.CommitmentID `serix:""`
		// The outputs that are created in this slot.
		CreatedOutputs []*OutputWithID `serix:",lenPrefix=uint32"`
		// The outputs that are consumed in this slot.
		ConsumedOutputs []*OutputWithID `serix:",lenPrefix=uint32"`
	}

	// CongestionResponse defines the response for the congestion REST API call.
	CongestionResponse struct {
		// Slot is the slot for which the estimate is provided.
		Slot iotago.SlotIndex `serix:""`
		// Ready indicates if a node is ready to schedule a block issued by the specified account, or if the issuer should wait.
		Ready bool `serix:""`
		// ReferenceManaCost (RMC) is the mana cost a user needs to burn to issue a block in slot Slot.
		ReferenceManaCost iotago.Mana `serix:""`
		// BlockIssuanceCredits (BIC) provides the BIC of the account in slot Slot.
		// This balance needs to be non-negative, otherwise account is locked
		BlockIssuanceCredits iotago.BlockIssuanceCredits `serix:""`
	}

	// ValidatorResponse defines the response used in stakers response REST API calls.
	ValidatorResponse struct {
		// AddressBech32 is the account address of the validator.
		AddressBech32 string `serix:"address,lenPrefix=uint8"`
		// StakingEndEpoch is the epoch until which the validator registered to stake.
		StakingEndEpoch iotago.EpochIndex `serix:""`
		// PoolStake is the sum of tokens delegated to the pool and the validator stake.
		PoolStake iotago.BaseToken `serix:""`
		// ValidatorStake is the stake of the validator.
		ValidatorStake iotago.BaseToken `serix:""`
		// FixedCost is the fixed cost that the validator receives from the total pool reward.
		FixedCost iotago.Mana `serix:""`
		// Active indicates whether the validator was active recently, and would be considered during committee selection.
		Active bool `serix:""`
		// LatestSupportedProtocolVersion is the latest supported protocol version of the validator.
		LatestSupportedProtocolVersion iotago.Version `serix:""`
		// LatestSupportedProtocolHash is the protocol hash of the latest supported protocol of the validator.
		LatestSupportedProtocolHash iotago.Identifier `serix:""`
	}

	// ValidatorsResponse defines the response for the staking REST API call.
	ValidatorsResponse struct {
		Validators []*ValidatorResponse `serix:",lenPrefix=uint16"`
		PageSize   uint32               `serix:""`
		Cursor     string               `serix:",lenPrefix=uint8,omitempty"`
	}

	// ManaRewardsResponse defines the response for the mana rewards REST API call.
	ManaRewardsResponse struct {
		// StartEpoch is the first epoch for which rewards can be claimed.
		// This value is useful for checking if rewards have expired (by comparing against the staking or delegation start)
		// or would expire soon (by checking its relation to the rewards retention period).
		StartEpoch iotago.EpochIndex `serix:""`
		// EndEpoch is the last epoch for which rewards can be claimed.
		EndEpoch iotago.EpochIndex `serix:""`
		// The amount of totally available decayed rewards the requested output may claim.
		Rewards iotago.Mana `serix:""`
		// The rewards of the latest committed epoch of the staking pool to which this validator or delegator belongs.
		// The ratio of this value and the maximally possible rewards for the latest committed epoch can be used to determine
		// how well the validator of this staking pool performed in that epoch.
		// Note that if the pool was not part of the committee in the latest committed epoch, this value is 0.
		LatestCommittedEpochPoolRewards iotago.Mana `serix:""`
	}

	// CommitteeMemberResponse defines the response used in committee and staking response REST API calls.
	CommitteeMemberResponse struct {
		// AddressBech32 is the account address of the validator.
		AddressBech32 string `serix:"address,lenPrefix=uint8"`
		// PoolStake is the sum of tokens delegated to the pool and the validator stake.
		PoolStake iotago.BaseToken `serix:""`
		// ValidatorStake is the stake of the validator.
		ValidatorStake iotago.BaseToken `serix:""`
		// FixedCost is the fixed cost that the validator received from the total pool reward.
		FixedCost iotago.Mana `serix:""`
	}

	// CommitteeResponse defines the response for the staking REST API call.
	CommitteeResponse struct {
		Committee           []*CommitteeMemberResponse `serix:",lenPrefix=uint8"`
		TotalStake          iotago.BaseToken           `serix:""`
		TotalValidatorStake iotago.BaseToken           `serix:""`
		Epoch               iotago.EpochIndex          `serix:""`
	}
)

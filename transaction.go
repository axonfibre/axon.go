package iotago

import (
	"context"
	"fmt"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
)

const (
	// MaxOutputsCount defines the maximum amount of outputs within a Transaction.
	MaxOutputsCount = 128
	// MinOutputsCount defines the minimum amount of inputs within a Transaction.
	MinOutputsCount = 1
)

var (
	// ErrInvalidInputsCommitment gets returned when the inputs commitment is invalid.
	ErrInvalidInputsCommitment = ierrors.New("invalid inputs commitment")
	// ErrTxEssenceNetworkIDInvalid gets returned when a network ID within a Transaction is invalid.
	ErrTxEssenceNetworkIDInvalid = ierrors.New("invalid network ID")
	// ErrInputUTXORefsNotUnique gets returned if multiple inputs reference the same UTXO.
	ErrInputUTXORefsNotUnique = ierrors.New("inputs must each reference a unique UTXO")
	// ErrInputBICNotUnique gets returned if multiple inputs reference the same BIC.
	ErrInputBICNotUnique = ierrors.New("inputs must each reference a unique BIC")
	// ErrInputRewardInvalid gets returned if multiple reward inputs reference the same input index
	// or if they reference an index greater than max inputs count.
	ErrInputRewardInvalid = ierrors.New("invalid reward input")
	// ErrMultipleInputCommitments gets returned if multiple commitment inputs are provided.
	ErrMultipleInputCommitments = ierrors.New("there are multiple commitment inputs")
	// ErrAccountOutputNonEmptyState gets returned if an AccountOutput with zeroed AccountID contains state (counters non-zero etc.).
	ErrAccountOutputNonEmptyState = ierrors.New("account output is not empty state")
	// ErrAccountOutputCyclicAddress gets returned if an AccountOutput's AccountID results into the same address as the State/Governance controller.
	ErrAccountOutputCyclicAddress = ierrors.New("account output's AccountID corresponds to state and/or governance controller")
	// ErrNFTOutputCyclicAddress gets returned if an NFTOutput's NFTID results into the same address as the address field within the output.
	ErrNFTOutputCyclicAddress = ierrors.New("NFT output's ID corresponds to address field")
	// ErrDelegationValidatorAddressZeroed gets returned if a Delegation Output's Validator address is zeroed out.
	ErrDelegationValidatorAddressZeroed = ierrors.New("delegation output's validator address is zeroed")
	// ErrOutputsSumExceedsTotalSupply gets returned if the sum of the output deposits exceeds the total supply of tokens.
	ErrOutputsSumExceedsTotalSupply = ierrors.New("accumulated output balance exceeds total supply")
	// ErrOutputAmountMoreThanTotalSupply gets returned if an output base token amount is more than the total supply.
	ErrOutputAmountMoreThanTotalSupply = ierrors.New("an output's base token amount cannot exceed the total supply")
	// ErrStorageDepositLessThanMinReturnOutputStorageDeposit gets returned when the storage deposit condition's amount is less than the min storage deposit for the return output.
	ErrStorageDepositLessThanMinReturnOutputStorageDeposit = ierrors.New("storage deposit return amount is less than the min storage deposit needed for the return output")
	// ErrStorageDepositExceedsTargetOutputAmount gets returned when the storage deposit condition's amount exceeds the target output's base token amount.
	ErrStorageDepositExceedsTargetOutputAmount = ierrors.New("storage deposit return amount exceeds target output's base token amount")
	// ErrMaxNativeTokensCountExceeded gets returned if outputs or transactions exceed the MaxNativeTokensCount.
	ErrMaxNativeTokensCountExceeded = ierrors.New("max native tokens count exceeded")
)

type (
	TxEssenceOutputs = Outputs[TxEssenceOutput]
)

// Transaction is the part of a SignedTransaction that contains inputs and outputs.
type Transaction struct {
	API                 API
	*TransactionEssence `serix:"0"`
	// The outputs of this transaction.
	Outputs TxEssenceOutputs `serix:"1,mapKey=outputs"`
}

// ID returns the TransactionID created without the signatures.
func (t *Transaction) ID() (TransactionID, error) {
	essenceBytes, err := t.API.Encode(t.TransactionEssence)
	if err != nil {
		return TransactionID{}, ierrors.Errorf("can't compute essence bytes: %w", err)
	}

	outputBytes, err := t.API.Encode(t.Outputs)
	if err != nil {
		return TransactionID{}, ierrors.Errorf("can't compute unlock bytes: %w", err)
	}

	return TransactionIDFromData(t.CreationSlot, byteutils.ConcatBytes(essenceBytes, outputBytes)), nil
}

func (t *Transaction) SetDeserializationContext(ctx context.Context) {
	t.API = APIFromContext(ctx)
}

func (t *Transaction) Clone() *Transaction {
	return &Transaction{
		API:                t.API,
		TransactionEssence: t.TransactionEssence.Clone(),
		Outputs:            t.Outputs.Clone(),
	}
}

func (t *Transaction) Inputs() ([]*UTXOInput, error) {
	references := make([]*UTXOInput, 0, len(t.TransactionEssence.Inputs))
	for _, input := range t.TransactionEssence.Inputs {
		switch castInput := input.(type) {
		case *UTXOInput:
			references = append(references, castInput)
		default:
			return nil, ErrUnknownInputType
		}
	}

	return references, nil
}

func (t *Transaction) ContextInputs() (TransactionContextInputs, error) {
	references := make(TransactionContextInputs, 0, len(t.TransactionEssence.ContextInputs))
	for _, input := range t.TransactionEssence.ContextInputs {
		switch castInput := input.(type) {
		case *CommitmentInput, *BlockIssuanceCreditInput, *RewardInput:
			references = append(references, castInput)
		default:
			return nil, ErrUnknownContextInputType
		}
	}

	return references, nil
}

func (t *Transaction) BICInputs() ([]*BlockIssuanceCreditInput, error) {
	references := make([]*BlockIssuanceCreditInput, 0, len(t.TransactionEssence.ContextInputs))
	for _, input := range t.TransactionEssence.ContextInputs {
		switch castInput := input.(type) {
		case *BlockIssuanceCreditInput:
			references = append(references, castInput)
		case *CommitmentInput, *RewardInput:
			// ignore this type
		default:
			return nil, ErrUnknownContextInputType
		}
	}

	return references, nil
}

func (t *Transaction) RewardInputs() ([]*RewardInput, error) {
	references := make([]*RewardInput, 0, len(t.TransactionEssence.ContextInputs))
	for _, input := range t.TransactionEssence.ContextInputs {
		switch castInput := input.(type) {
		case *RewardInput:
			references = append(references, castInput)
		case *CommitmentInput, *BlockIssuanceCreditInput:
			// ignore this type
		default:
			return nil, ErrUnknownContextInputType
		}
	}

	return references, nil
}

// Returns the first commitment input in the transaction if it exists or nil.
func (t *Transaction) CommitmentInput() *CommitmentInput {
	for _, input := range t.TransactionEssence.ContextInputs {
		switch castInput := input.(type) {
		case *BlockIssuanceCreditInput, *RewardInput:
			// ignore this type
		case *CommitmentInput:
			return castInput
		default:
			return nil
		}
	}

	return nil
}

// SigningMessage returns the to be signed message.
func (t *Transaction) SigningMessage(api API) ([]byte, error) {
	essenceBytes, err := api.Encode(t)
	if err != nil {
		return nil, err
	}
	essenceBytesHash := blake2b.Sum256(essenceBytes)

	return essenceBytesHash[:], nil
}

// Sign produces signatures signing the essence for every given AddressKeys.
// The produced signatures are in the same order as the AddressKeys.
func (t *Transaction) Sign(api API, inputsCommitment []byte, addrKeys ...AddressKeys) ([]Signature, error) {
	if inputsCommitment == nil || len(inputsCommitment) != InputsCommitmentLength {
		return nil, ErrInvalidInputsCommitment
	}

	copy(t.InputsCommitment[:], inputsCommitment)

	signMsg, err := t.SigningMessage(api)
	if err != nil {
		return nil, err
	}

	sigs := make([]Signature, len(addrKeys))
	signer := NewInMemoryAddressSigner(addrKeys...)
	for i, v := range addrKeys {
		sig, err := signer.Sign(v.Address, signMsg)
		if err != nil {
			return nil, err
		}
		sigs[i] = sig
	}

	return sigs, nil
}

func (t *Transaction) Size() int {
	return t.TransactionEssence.Size() + t.Outputs.Size()
}

// syntacticallyValidate checks whether the transaction essence is syntactically valid.
// The function does not syntactically validate the input or outputs themselves.
func (t *Transaction) syntacticallyValidate(api API) error {
	protoParams := api.ProtocolParameters()

	if err := t.TransactionEssence.syntacticallyValidate(api); err != nil {
		return err
	}

	return SyntacticallyValidateOutputs(t.Outputs,
		OutputsSyntacticalDepositAmount(protoParams, api.RentStructure()),
		OutputsSyntacticalExpirationAndTimelock(),
		OutputsSyntacticalNativeTokens(),
		OutputsSyntacticalChainConstrainedOutputUniqueness(),
		OutputsSyntacticalFoundry(),
		OutputsSyntacticalAccount(),
		OutputsSyntacticalNFT(),
		OutputsSyntacticalDelegation(),
	)
}

// WorkScore calculates the Work Score of the Transaction.
func (t *Transaction) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	workscoreTransactionEssence, err := t.TransactionEssence.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

	workScoreOutputs, err := t.Outputs.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

	return workscoreTransactionEssence.Add(workScoreOutputs)
}

func (t *Transaction) String() string {
	// TODO: stringify for debugging purposes
	return fmt.Sprintf("SignedTransaction[%v %v]", t.TransactionEssence, t.Outputs)
}

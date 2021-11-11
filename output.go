package iotago

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/iotaledger/hive.go/serializer"
)

const (
	// OutputIDLength defines the length of an OutputID.
	OutputIDLength = TransactionIDLength + serializer.UInt16ByteSize
)

// OutputType defines the type of outputs.
type OutputType byte

const (
	// OutputSimple denotes a type of output which is locked by a signature and deposits onto a single address.
	OutputSimple OutputType = iota
	// OutputExtended denotes a type of output which can also hold native tokens and feature blocks.
	OutputExtended
	// OutputTreasury denotes the type of the TreasuryOutput.
	OutputTreasury
	// OutputAlias denotes the type of an AliasOutput.
	OutputAlias
	// OutputFoundry denotes the type of a FoundryOutput.
	OutputFoundry
	// OutputNFT denotes the type of an NFTOutput.
	OutputNFT
)

// OutputTypeToString returns the name of an Output given the type.
func OutputTypeToString(ty OutputType) string {
	switch ty {
	case OutputSimple:
		return "SimpleOutput"
	case OutputExtended:
		return "ExtendedOutput"
	case OutputTreasury:
		return "TreasuryOutput"
	case OutputAlias:
		return "AliasOutput"
	case OutputFoundry:
		return "FoundryOutput"
	case OutputNFT:
		return "NFTOutput"
	}
	return "unknown output"
}

var (
	// ErrDepositAmountMustBeGreaterThanZero returned if the deposit amount of an output is less or equal zero.
	ErrDepositAmountMustBeGreaterThanZero = errors.New("deposit amount must be greater than zero")
	// ErrMultiIdentOutputMismatch gets returned when MultiIdentOutput(s) aren't compatible.
	ErrMultiIdentOutputMismatch = errors.New("multi ident output mismatch")
	// ErrNonUniqueMultiIdentOutputs gets returned when multiple MultiIdentOutput(s) with the same AccountID exist within an OutputsByType.
	ErrNonUniqueMultiIdentOutputs = errors.New("non unique multi ident within outputs")
)

// Outputs is a slice of Output.
type Outputs []Output

func (o Outputs) ToSerializables() serializer.Serializables {
	seris := make(serializer.Serializables, len(o))
	for i, x := range o {
		seris[i] = x.(serializer.Serializable)
	}
	return seris
}

func (o *Outputs) FromSerializables(seris serializer.Serializables) {
	*o = make(Outputs, len(seris))
	for i, seri := range seris {
		(*o)[i] = seri.(Output)
	}
}

// ToOutputsByType converts the Outputs slice to OutputsByType.
func (o Outputs) ToOutputsByType() OutputsByType {
	outputsByType := make(OutputsByType)
	for _, output := range o {
		slice, has := outputsByType[output.Type()]
		if !has {
			slice = make(Outputs, 0)
		}
		outputsByType[output.Type()] = append(slice, output)
	}
	return outputsByType
}

// OutputsFilterFunc is a predicate function operating on an Output.
type OutputsFilterFunc func(output Output) bool

// OutputsFilterByType is an OutputsFilterFunc which filters Outputs by OutputType.
func OutputsFilterByType(ty OutputType) OutputsFilterFunc {
	return func(output Output) bool { return output.Type() == ty }
}

// Filter returns Outputs (retained order) passing the given OutputsFilterFunc.
func (o Outputs) Filter(f OutputsFilterFunc) Outputs {
	filtered := make(Outputs, 0)
	for _, output := range o {
		if !f(output) {
			continue
		}
		filtered = append(filtered, output)
	}
	return filtered
}

// OutputsByType is a map of OutputType(s) to slice of Output(s).
type OutputsByType map[OutputType][]Output

// NativeTokenOutputs returns a slice of Outputs which are NativeTokenOutput.
func (outputs OutputsByType) NativeTokenOutputs() NativeTokenOutputs {
	nativeTokenOutputs := make(NativeTokenOutputs, 0)
	for _, slice := range outputs {
		for _, output := range slice {
			nativeTokenOutput, is := output.(NativeTokenOutput)
			if !is {
				continue
			}
			nativeTokenOutputs = append(nativeTokenOutputs, nativeTokenOutput)
		}
	}
	return nativeTokenOutputs
}

// MultiIdentOutputs returns a slice of Outputs which are MultiIdentOutput.
func (outputs OutputsByType) MultiIdentOutputs() MultiIdentOutputs {
	multiIdentOutputs := make(MultiIdentOutputs, 0)
	for _, slice := range outputs {
		for _, output := range slice {
			multiIdentOutput, is := output.(MultiIdentOutput)
			if !is {
				continue
			}
			multiIdentOutputs = append(multiIdentOutputs, multiIdentOutput)
		}
	}
	return multiIdentOutputs
}

// MultiIdentOutputsSet returns a map of AccountID to MultiIdentOutput.
// If multiple MultiIdentOutput(s) exist for a given AccountID, an error is returned.
func (outputs OutputsByType) MultiIdentOutputsSet() (MultiIdentOutputsSet, error) {
	multiIdentOutputsSet := make(MultiIdentOutputsSet, 0)
	for _, output := range outputs.MultiIdentOutputs() {
		if _, has := multiIdentOutputsSet[output.Account()]; has {
			return nil, ErrNonUniqueMultiIdentOutputs
		}
		multiIdentOutputsSet[output.Account()] = output
	}
	return multiIdentOutputsSet, nil
}

// FoundryOutputs returns a slice of Outputs which are FoundryOutput.
func (outputs OutputsByType) FoundryOutputs() FoundryOutputs {
	foundryOutputs := make(FoundryOutputs, 0)
	for _, output := range outputs[OutputFoundry] {
		foundryOutput, is := output.(*FoundryOutput)
		if !is {
			continue
		}
		foundryOutputs = append(foundryOutputs, foundryOutput)
	}
	return foundryOutputs
}

// FoundryOutputsSet returns a map of FoundryID to FoundryOutput.
// If multiple FoundryOutput(s) exist for a given FoundryID, an error is returned.
func (outputs OutputsByType) FoundryOutputsSet() (FoundryOutputsSet, error) {
	foundryOutputsSet := make(FoundryOutputsSet, 0)
	for _, output := range outputs[OutputFoundry] {
		foundryOutput, is := output.(*FoundryOutput)
		if !is {
			continue
		}
		foundryID, err := foundryOutput.ID()
		if err != nil {
			return nil, err
		}
		if _, has := foundryOutputsSet[foundryID]; has {
			return nil, ErrNonUniqueFoundryOutputs
		}
		foundryOutputsSet[foundryID] = foundryOutput
	}
	return foundryOutputsSet, nil
}

// AliasOutputs returns a slice of Outputs which are AliasOutput.
func (outputs OutputsByType) AliasOutputs() AliasOutputs {
	aliasOutputs := make(AliasOutputs, 0)
	for _, output := range outputs[OutputFoundry] {
		aliasOutput, is := output.(*AliasOutput)
		if !is {
			continue
		}
		aliasOutputs = append(aliasOutputs, aliasOutput)
	}
	return aliasOutputs
}

// NonNewAliasOutputsSet returns a map of AliasID to AliasOutput.
// If multiple AliasOutput(s) exist for a given AliasID, an error is returned.
// The produced set does not include AliasOutputs of which their AliasID are zeroed.
func (outputs OutputsByType) NonNewAliasOutputsSet() (AliasOutputsSet, error) {
	aliasOutputsSet := make(AliasOutputsSet, 0)
	for _, output := range outputs[OutputFoundry] {
		aliasOutput, is := output.(*AliasOutput)
		if !is || aliasOutput.AliasEmpty() {
			continue
		}
		if _, has := aliasOutputsSet[aliasOutput.AliasID]; has {
			return nil, ErrNonUniqueAliasOutputs
		}
		aliasOutputsSet[aliasOutput.AliasID] = aliasOutput
	}
	return aliasOutputsSet, nil
}

// NativeTokenOutputs is a slice of NativeTokenOutput(s).
type NativeTokenOutputs []NativeTokenOutput

// Sum sums up the different NativeTokens occurring within the given outputs.
func (ntOutputs NativeTokenOutputs) Sum() (NativeTokenSum, error) {
	sum := make(map[NativeTokenID]*big.Int)
	for _, output := range ntOutputs {
		for _, nativeToken := range output.NativeTokenSet() {
			if sign := nativeToken.Amount.Sign(); sign == -1 || sign == 0 {
				return nil, ErrNativeTokenAmountLessThanEqualZero
			}

			val := sum[nativeToken.ID]
			if val == nil {
				val = new(big.Int)
			}

			if val.Add(val, nativeToken.Amount).Cmp(abi.MaxUint256) == 1 {
				return nil, ErrNativeTokenSumExceedsUint256
			}
			sum[nativeToken.ID] = val
		}
	}
	return sum, nil
}

// InputSet maps inputs to their origin UTXOs.
type InputSet map[UTXOInputID]Output

// NewAliases returns an AliasOutputsSet for all AliasOutputs which are new.
func (inputSet InputSet) NewAliases() AliasOutputsSet {
	set := make(AliasOutputsSet)
	for utxoInputID, output := range inputSet {
		aliasOutput, is := output.(*AliasOutput)
		if !is || !aliasOutput.AliasEmpty() {
			continue
		}
		set[AliasIDFromOutputID(utxoInputID)] = aliasOutput
	}
	return set
}

// Output defines a unit of output of a transaction.
type Output interface {
	serializer.Serializable

	// Deposit returns the amount this Output deposits.
	Deposit() (uint64, error)
	// Type returns the type of the output.
	Type() OutputType
}

// SingleIdentOutput is a type of Output where without considering its FeatureBlocks,
// only one identity needs to be unlocked.
type SingleIdentOutput interface {
	Output
	// Ident returns the identity to which this output is locked to.
	Ident() (Address, error)
}

// AccountOutput is a type of Output which encapsulates the concept of an account.
type AccountOutput interface {
	Output
	Account() AccountID
}

// MultiIdentOutputsSet is a set of MultiIdentOutput(s).
type MultiIdentOutputsSet map[AccountID]MultiIdentOutput

// MultiIdentOutputs is a slice of MultiIdentOutput(s).
type MultiIdentOutputs []MultiIdentOutput

// MultiIdentOutput is a type of Output which multiple identities can control/modify.
// Unlike the SingleIdentOutput, the MultiIdentOutput's to unlock identity is dependent
// on the transition the output does between inputs and outputs.
type MultiIdentOutput interface {
	AccountOutput
	// Ident computes the identity to which this output is locked to by examining
	// the transition to the next output state.
	// Note that it is the caller's job to ensure that the given other MultiIdentOutput
	// corresponds to this MultiIdentOutput.
	// If this MultiIdentOutput is not dependent on a transition to compute the ident,
	// nil can be passed as an argument.
	Ident(nextState MultiIdentOutput) (Address, error)
}

// NativeTokenOutput is a type of Output which can hold NativeToken.
type NativeTokenOutput interface {
	Output
	// NativeTokenSet returns the NativeToken this output defines.
	NativeTokenSet() NativeTokens
}

// FeatureBlockOutput is a type of Output which can hold FeatureBlock.
type FeatureBlockOutput interface {
	// FeatureBlocks returns the feature blocks this output defines.
	FeatureBlocks() FeatureBlocks
}

// OutputSelector implements SerializableSelectorFunc for output types.
func OutputSelector(outputType uint32) (serializer.Serializable, error) {
	var seri serializer.Serializable
	switch OutputType(outputType) {
	case OutputSimple:
		seri = &SimpleOutput{}
	case OutputExtended:
		seri = &ExtendedOutput{}
	case OutputTreasury:
		seri = &TreasuryOutput{}
	case OutputAlias:
		seri = &AliasOutput{}
	case OutputFoundry:
		seri = &FoundryOutput{}
	case OutputNFT:
		seri = &NFTOutput{}
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownOutputType, outputType)
	}
	return seri, nil
}

// OutputIDHex is the hex representation of an output ID.
type OutputIDHex string

// MustSplitParts returns the transaction ID and output index parts of the hex output ID.
// It panics if the hex output ID is invalid.
func (oih OutputIDHex) MustSplitParts() (*TransactionID, uint16) {
	txID, outputIndex, err := oih.SplitParts()
	if err != nil {
		panic(err)
	}
	return txID, outputIndex
}

// SplitParts returns the transaction ID and output index parts of the hex output ID.
func (oih OutputIDHex) SplitParts() (*TransactionID, uint16, error) {
	outputIDBytes, err := hex.DecodeString(string(oih))
	if err != nil {
		return nil, 0, err
	}
	var txID TransactionID
	copy(txID[:], outputIDBytes[:TransactionIDLength])
	outputIndex := binary.LittleEndian.Uint16(outputIDBytes[TransactionIDLength : TransactionIDLength+serializer.UInt16ByteSize])
	return &txID, outputIndex, nil
}

// MustAsUTXOInput converts the hex output ID to a UTXOInput.
// It panics if the hex output ID is invalid.
func (oih OutputIDHex) MustAsUTXOInput() *UTXOInput {
	utxoInput, err := oih.AsUTXOInput()
	if err != nil {
		panic(err)
	}
	return utxoInput
}

// AsUTXOInput converts the hex output ID to a UTXOInput.
func (oih OutputIDHex) AsUTXOInput() (*UTXOInput, error) {
	var utxoInput UTXOInput
	txID, outputIndex, err := oih.SplitParts()
	if err != nil {
		return nil, err
	}
	copy(utxoInput.TransactionID[:], txID[:])
	utxoInput.TransactionOutputIndex = outputIndex
	return &utxoInput, nil
}

// OutputsPredicateFunc which given the index of an output and the output itself, runs validations and returns an error if any should fail.
type OutputsPredicateFunc func(index int, output Output) error

// OutputsPredicateAddrUnique returns an OutputsPredicateFunc which checks that all addresses are unique per OutputType.
// Deprecated: an output set no longer needs to hold unique addresses per output.
func OutputsPredicateAddrUnique() OutputsPredicateFunc {
	set := map[OutputType]map[string]int{}
	return func(index int, dep Output) error {
		var b strings.Builder

		target, err := dep.Target()
		if err != nil {
			return fmt.Errorf("unable to get target of output: %w", err)
		}

		if target == nil {
			return nil
		}

		// can't be reduced to one b.Write()
		switch addr := target.(type) {
		case *Ed25519Address:
			if _, err := b.Write(addr[:]); err != nil {
				return fmt.Errorf("%w: unable to serialize Ed25519 address in addr unique validator", err)
			}
		}

		k := b.String()

		m, ok := set[dep.Type()]
		if !ok {
			m = make(map[string]int)
			set[dep.Type()] = m
		}

		if j, has := m[k]; has {
			return fmt.Errorf("%w: output %d and %d share the same address", ErrOutputAddrNotUnique, j, index)
		}
		m[k] = index
		return nil
	}
}

// OutputsPredicateDepositAmount returns an OutputsPredicateFunc which checks that:
//	- every output deposits more than zero
//	- every output deposits less than the total supply
//	- the sum of deposits does not exceed the total supply
// If -1 is passed to the validator func, then the sum is not aggregated over multiple calls.
func OutputsPredicateDepositAmount() OutputsPredicateFunc {
	var sum uint64
	return func(index int, dep Output) error {
		deposit, err := dep.Deposit()
		if err != nil {
			return fmt.Errorf("unable to get deposit of output: %w", err)
		}
		if deposit == 0 {
			return fmt.Errorf("%w: output %d", ErrDepositAmountMustBeGreaterThanZero, index)
		}
		if deposit > TokenSupply {
			return fmt.Errorf("%w: output %d", ErrOutputDepositsMoreThanTotalSupply, index)
		}
		if sum+deposit > TokenSupply {
			return fmt.Errorf("%w: output %d", ErrOutputsSumExceedsTotalSupply, index)
		}
		if index != -1 {
			sum += deposit
		}
		return nil
	}
}

// OutputsPredicateNativeTokensCount returns an OutputsPredicateFunc which checks that:
//	- the sum of native tokens count across all outputs does not exceed MaxNativeTokensCount
func OutputsPredicateNativeTokensCount() OutputsPredicateFunc {
	var nativeTokensCount int
	return func(index int, output Output) error {
		if nativeTokenOutput, is := output.(NativeTokenOutput); is {
			nativeTokensCount += len(nativeTokenOutput.NativeTokenSet())
			if nativeTokensCount > MaxNativeTokensCount {
				return ErrOutputsExceedMaxNativeTokensCount
			}
		}
		return nil
	}
}

// OutputsPredicateSenderFeatureBlockRequirement returns an OutputsPredicateFunc which checks that:
//	- if an output contains a SenderFeatureBlock if another FeatureBlock (example ReturnFeatureBlock) requires it
func OutputsPredicateSenderFeatureBlockRequirement() OutputsPredicateFunc {
	return func(index int, output Output) error {
		featureBlockOutput, is := output.(FeatureBlockOutput)
		if !is {
			return nil
		}
		var hasReturnFeatBlock, hasExpMsFeatBlock, hasExpUnixFeatBlock, hasSenderFeatBlock bool
		for _, featureBlock := range featureBlockOutput.FeatureBlocks() {
			switch featureBlock.(type) {
			case *ReturnFeatureBlock:
				hasReturnFeatBlock = true
			case *ExpirationMilestoneIndexFeatureBlock:
				hasExpMsFeatBlock = true
			case *ExpirationUnixFeatureBlock:
				hasExpUnixFeatBlock = true
			case *SenderFeatureBlock:
				hasSenderFeatBlock = true
			}
		}
		if (hasReturnFeatBlock || hasExpMsFeatBlock || hasExpUnixFeatBlock) && !hasSenderFeatBlock {
			return fmt.Errorf("%w: output %d", ErrOutputRequiresSenderFeatureBlock, index)
		}
		return nil
	}
}

// OutputsPredicateAlias returns an OutputsPredicateFunc which checks that AliasOutput(s)':
//	- StateIndex/FoundryCounter are zero if the AliasID is zeroed
//	- StateController and GovernanceController must be different from AliasAddress derived from AliasID
func OutputsPredicateAlias(txID *TransactionID) OutputsPredicateFunc {
	return func(index int, output Output) error {
		aliasOutput, is := output.(*AliasOutput)
		if !is {
			return nil
		}

		var outputAliasAddr AliasAddress
		if aliasOutput.AliasEmpty() {
			switch {
			case aliasOutput.StateIndex != 0:
				return fmt.Errorf("%w: output %d, state index not zero", ErrAliasOutputNonEmptyState, index)
			case aliasOutput.FoundryCounter != 0:
				return fmt.Errorf("%w: output %d, foundry counter not zero", ErrAliasOutputNonEmptyState, index)
			}

			// build AliasID using the transaction ID
			outputAliasAddr = AliasAddressFromOutputID(UTXOIDFromTransactionIDAndIndex(*txID, uint16(index)))
		}

		if outputAliasAddr == emptyAliasAddress {
			copy(outputAliasAddr[:], aliasOutput.AliasID[:])
		}

		if stateCtrlAddr, ok := aliasOutput.StateController.(*AliasAddress); ok && outputAliasAddr == *stateCtrlAddr {
			return fmt.Errorf("%w: output %d, AliasID=StateController", ErrAliasOutputCyclicAddress, index)
		}
		if govCtrlAddr, ok := aliasOutput.GovernanceController.(*AliasAddress); ok && outputAliasAddr == *govCtrlAddr {
			return fmt.Errorf("%w: output %d, AliasID=GovernanceController", ErrAliasOutputCyclicAddress, index)
		}

		return nil
	}
}

// OutputsPredicateFoundry returns an OutputsPredicateFunc which checks that FoundryOutput(s)':
//	- CirculatingSupply is less equal MaximumSupply
//	- MaximumSupply is not zero
func OutputsPredicateFoundry() OutputsPredicateFunc {
	return func(index int, output Output) error {
		foundryOutput, is := output.(*FoundryOutput)
		if !is {
			return nil
		}

		if r := foundryOutput.MaximumSupply.Cmp(new(big.Int).SetInt64(0)); r == -1 || r == 0 {
			return fmt.Errorf("%w: output %d, less than equal zero", ErrFoundryOutputInvalidMaximumSupply, index)
		}

		if r := foundryOutput.CirculatingSupply.Cmp(foundryOutput.MaximumSupply); r == 1 {
			return fmt.Errorf("%w: output %d, bigger than maximum supply", ErrFoundryOutputInvalidCirculatingSupply, index)
		}

		return nil
	}
}

// OutputsPredicateNFT returns an OutputsPredicateFunc which checks that NFTOutput(s)':
//	- Address must be different from NFTAddress derived from NFTID
func OutputsPredicateNFT(txID *TransactionID) OutputsPredicateFunc {
	return func(index int, output Output) error {
		nftOutput, is := output.(*NFTOutput)
		if !is {
			return nil
		}

		var outputNFTAddr NFTAddress
		if nftOutput.NFTID == emptyNFTID {
			outputNFTAddr = NFTAddressFromOutputID(UTXOIDFromTransactionIDAndIndex(*txID, uint16(index)))
		}

		if outputNFTAddr == emptyNFTAddress {
			copy(outputNFTAddr[:], nftOutput.NFTID[:])
		}

		if addr, ok := nftOutput.Address.(*NFTAddress); ok && outputNFTAddr == *addr {
			return fmt.Errorf("%w: output %d, AliasID=StateController", ErrNFTOutputCyclicAddress, index)
		}

		return nil
	}
}

// supposed to be called with -1 as input in order to be used over multiple calls.
var outputAmountValidator = OutputsPredicateDepositAmount()

// ValidateOutputs validates the outputs by running them against the given OutputsPredicateFunc(s).
func ValidateOutputs(outputs Outputs, funcs ...OutputsPredicateFunc) error {
	for i, output := range outputs {
		for _, f := range funcs {
			if err := f(i, output); err != nil {
				return err
			}
		}
	}
	return nil
}

// jsonOutputSelector selects the json output implementation for the given type.
func jsonOutputSelector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch OutputType(ty) {
	case OutputSimple:
		obj = &jsonSimpleOutput{}
	case OutputExtended:
		obj = &jsonExtendedOutput{}
	case OutputTreasury:
		obj = &jsonTreasuryOutput{}
	case OutputAlias:
		obj = &jsonAliasOutput{}
	case OutputFoundry:
		obj = &jsonFoundryOutput{}
	case OutputNFT:
		obj = &jsonNFTOutput{}
	default:
		return nil, fmt.Errorf("unable to decode output type from JSON: %w", ErrUnknownOutputType)
	}
	return obj, nil
}

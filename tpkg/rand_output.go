package tpkg

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"

	"github.com/axonfibre/fibre.go/lo"
	axongo "github.com/axonfibre/axon.go/v4"
)

func RandOutputIDWithCreationSlot(slot axongo.SlotIndex, index ...uint16) axongo.OutputID {
	txID := RandTransactionIDWithCreationSlot(slot)

	idx := RandUint16(126)
	if len(index) > 0 {
		idx = index[0]
	}

	var outputID axongo.OutputID
	copy(outputID[:], txID[:])
	binary.LittleEndian.PutUint16(outputID[axongo.TransactionIDLength:], idx)

	return outputID
}

func RandOutputID(index ...uint16) axongo.OutputID {
	return RandOutputIDWithCreationSlot(0, index...)
}

func RandOutputIDsWithCreationSlot(slot axongo.SlotIndex, count uint16) axongo.OutputIDs {
	outputIDs := make(axongo.OutputIDs, int(count))
	for i := range int(count) {
		outputIDs[i] = RandOutputIDWithCreationSlot(slot, count)
	}

	return outputIDs
}

func RandOutputIDs(count uint16) axongo.OutputIDs {
	outputIDs := make(axongo.OutputIDs, int(count))
	for i := range count {
		outputIDs[i] = RandOutputID(count)
	}

	return outputIDs
}

func RandOutputIDProof(api axongo.API) *axongo.OutputIDProof {
	tx := RandTransaction(api, WithOutputCount(1))
	return lo.PanicOnErr(axongo.OutputIDProofFromTransaction(tx, 0))
}

// RandBasicOutput returns a random basic output (with no features).
func RandBasicOutput(addressType ...axongo.AddressType) *axongo.BasicOutput {
	dep := &axongo.BasicOutput{
		Amount:           RandBaseToken(10000) + 1,
		UnlockConditions: axongo.BasicOutputUnlockConditions{},
		Features:         axongo.BasicOutputFeatures{},
	}

	addrType := axongo.AddressEd25519
	if len(addressType) > 0 {
		addrType = addressType[0]
	}

	//nolint:exhaustive
	switch addrType {
	case axongo.AddressEd25519:
		dep.UnlockConditions = axongo.BasicOutputUnlockConditions{&axongo.AddressUnlockCondition{Address: RandEd25519Address()}}
	default:
		panic(fmt.Sprintf("invalid addr type: %d", addrType))
	}

	return dep
}

func RandOutputType() axongo.OutputType {
	outputTypes := []axongo.OutputType{axongo.OutputBasic, axongo.OutputAccount, axongo.OutputAnchor, axongo.OutputFoundry, axongo.OutputNFT, axongo.OutputDelegation}

	return outputTypes[RandInt(len(outputTypes)-1)]
}

func RandOutput(outputType axongo.OutputType) axongo.Output {
	var addr axongo.Address
	if outputType == axongo.OutputFoundry {
		addr = RandAddress(axongo.AddressAccount)
	} else {
		addr = RandAddress(axongo.AddressEd25519)
	}

	return RandOutputOnAddress(outputType, addr)
}

func RandOutputOnAddress(outputType axongo.OutputType, address axongo.Address) axongo.Output {
	return RandOutputOnAddressWithAmount(outputType, address, RandBaseToken(axongo.MaxBaseToken))
}

func RandOutputOnAddressWithAmount(outputType axongo.OutputType, address axongo.Address, amount axongo.BaseToken) axongo.Output {
	var iotaOutput axongo.Output

	switch outputType {
	case axongo.OutputBasic:
		iotaOutput = &axongo.BasicOutput{
			Amount: amount,
			UnlockConditions: axongo.BasicOutputUnlockConditions{
				&axongo.AddressUnlockCondition{
					Address: address,
				},
			},
			Features: axongo.BasicOutputFeatures{},
		}

	case axongo.OutputAccount:
		iotaOutput = &axongo.AccountOutput{
			Amount:    amount,
			AccountID: RandAccountID(),
			UnlockConditions: axongo.AccountOutputUnlockConditions{
				&axongo.AddressUnlockCondition{
					Address: address,
				},
			},
			Features:          axongo.AccountOutputFeatures{},
			ImmutableFeatures: axongo.AccountOutputImmFeatures{},
		}

	case axongo.OutputAnchor:
		iotaOutput = &axongo.AnchorOutput{
			Amount:   amount,
			AnchorID: RandAnchorID(),
			UnlockConditions: axongo.AnchorOutputUnlockConditions{
				&axongo.StateControllerAddressUnlockCondition{
					Address: address,
				},
				&axongo.GovernorAddressUnlockCondition{
					Address: address,
				},
			},
			Features:          axongo.AnchorOutputFeatures{},
			ImmutableFeatures: axongo.AnchorOutputImmFeatures{},
		}

	case axongo.OutputFoundry:
		if address.Type() != axongo.AddressAccount {
			panic("not an alias address")
		}
		supply := new(big.Int).SetUint64(RandUint64(math.MaxUint64))

		//nolint:forcetypeassert // we already checked the type
		iotaOutput = &axongo.FoundryOutput{
			Amount:       amount,
			SerialNumber: 0,
			TokenScheme: &axongo.SimpleTokenScheme{
				MintedTokens:  supply,
				MeltedTokens:  new(big.Int).SetBytes([]byte{0}),
				MaximumSupply: supply,
			},
			UnlockConditions: axongo.FoundryOutputUnlockConditions{
				&axongo.ImmutableAccountUnlockCondition{
					Address: address.(*axongo.AccountAddress),
				},
			},
			Features:          axongo.FoundryOutputFeatures{},
			ImmutableFeatures: axongo.FoundryOutputImmFeatures{},
		}

	case axongo.OutputNFT:
		iotaOutput = &axongo.NFTOutput{
			Amount: amount,
			NFTID:  RandNFTID(),
			UnlockConditions: axongo.NFTOutputUnlockConditions{
				&axongo.AddressUnlockCondition{
					Address: address,
				},
			},
			Features:          axongo.NFTOutputFeatures{},
			ImmutableFeatures: axongo.NFTOutputImmFeatures{},
		}

	case axongo.OutputDelegation:
		iotaOutput = &axongo.DelegationOutput{
			Amount:           amount,
			DelegatedAmount:  amount,
			DelegationID:     RandDelegationID(),
			ValidatorAddress: RandAccountAddress(),
			StartEpoch:       RandEpoch(),
			EndEpoch:         axongo.MaxEpochIndex,
			UnlockConditions: axongo.DelegationOutputUnlockConditions{
				&axongo.AddressUnlockCondition{
					Address: address,
				},
			},
		}

	default:
		panic("unhandled output type")
	}

	return iotaOutput
}

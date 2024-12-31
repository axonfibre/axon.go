package tpkg

import (
	cryptorand "crypto/rand"
	"encoding/binary"

	"github.com/axonfibre/fibre.go/runtime/options"
	axongo "github.com/axonfibre/axon.go/v4"
)

func RandSignedTransactionIDWithCreationSlot(slot axongo.SlotIndex) axongo.SignedTransactionID {
	var signedTransactionID axongo.SignedTransactionID
	_, err := cryptorand.Read(signedTransactionID[:axongo.IdentifierLength])
	if err != nil {
		panic(err)
	}
	binary.LittleEndian.PutUint32(signedTransactionID[axongo.IdentifierLength:axongo.TransactionIDLength], uint32(slot))

	return signedTransactionID
}

func RandTransactionIDWithCreationSlot(slot axongo.SlotIndex) axongo.TransactionID {
	var transactionID axongo.TransactionID
	_, err := cryptorand.Read(transactionID[:axongo.IdentifierLength])
	if err != nil {
		panic(err)
	}
	binary.LittleEndian.PutUint32(transactionID[axongo.IdentifierLength:axongo.TransactionIDLength], uint32(slot))

	return transactionID
}

func RandSignedTransactionID() axongo.SignedTransactionID {
	return RandSignedTransactionIDWithCreationSlot(RandSlot())
}

func RandTransactionID() axongo.TransactionID {
	return RandTransactionIDWithCreationSlot(RandSlot())
}

// RandTransaction returns a random transaction essence.
func RandTransaction(api axongo.API, opts ...options.Option[axongo.Transaction]) *axongo.Transaction {
	return RandTransactionWithOptions(
		api,
		append([]options.Option[axongo.Transaction]{
			WithUTXOInputCount(RandInt(axongo.MaxInputsCount) + 1),
			WithOutputCount(RandInt(axongo.MaxOutputsCount) + 1),
			WithAllotmentCount(RandInt(axongo.MaxAllotmentCount) + 1),
		}, opts...)...,
	)
}

// RandTransactionWithInputCount returns a random transaction essence with a specific amount of inputs..
func RandTransactionWithInputCount(api axongo.API, inputCount int) *axongo.Transaction {
	return RandTransactionWithOptions(
		api,
		WithUTXOInputCount(inputCount),
		WithOutputCount(RandInt(axongo.MaxOutputsCount)+1),
		WithAllotmentCount(RandInt(axongo.MaxAllotmentCount)+1),
	)
}

// RandTransactionWithOutputCount returns a random transaction essence with a specific amount of outputs.
func RandTransactionWithOutputCount(api axongo.API, outputCount int) *axongo.Transaction {
	return RandTransactionWithOptions(
		api,
		WithUTXOInputCount(RandInt(axongo.MaxInputsCount)+1),
		WithOutputCount(outputCount),
		WithAllotmentCount(RandInt(axongo.MaxAllotmentCount)+1),
	)
}

// RandTransactionWithAllotmentCount returns a random transaction essence with a specific amount of outputs.
func RandTransactionWithAllotmentCount(api axongo.API, allotmentCount int) *axongo.Transaction {
	return RandTransactionWithOptions(
		api,
		WithUTXOInputCount(RandInt(axongo.MaxInputsCount)+1),
		WithOutputCount(RandInt(axongo.MaxOutputsCount)+1),
		WithAllotmentCount(allotmentCount),
	)
}

// RandTransactionWithOptions returns a random transaction essence with options applied.
func RandTransactionWithOptions(api axongo.API, opts ...options.Option[axongo.Transaction]) *axongo.Transaction {
	tx := &axongo.Transaction{
		API: api,
		TransactionEssence: &axongo.TransactionEssence{
			NetworkID:     TestNetworkID,
			ContextInputs: axongo.TxEssenceContextInputs{},
			Inputs:        axongo.TxEssenceInputs{},
			Allotments:    axongo.Allotments{},
			Capabilities:  axongo.TransactionCapabilitiesBitMask{},
		},
		Outputs: axongo.TxEssenceOutputs{},
	}

	inputCount := 1
	for i := inputCount; i > 0; i-- {
		tx.TransactionEssence.Inputs = append(tx.TransactionEssence.Inputs, RandUTXOInput())
	}

	outputCount := 1
	for i := outputCount; i > 0; i-- {
		tx.Outputs = append(tx.Outputs, RandBasicOutput(axongo.AddressEd25519))
	}

	tx = options.Apply(tx, opts)
	tx.TransactionEssence.ContextInputs.Sort()

	return tx
}

func WithBlockIssuanceCreditInputCount(inputCount int) options.Option[axongo.Transaction] {
	return func(tx *axongo.Transaction) {
		for i := inputCount; i > 0; i-- {
			tx.TransactionEssence.ContextInputs = append(tx.TransactionEssence.ContextInputs, RandBlockIssuanceCreditInput())
		}
	}
}

func WithRewardInputCount(inputCount uint16) options.Option[axongo.Transaction] {
	return func(tx *axongo.Transaction) {
		for i := inputCount; i > 0; i-- {
			rewardInput := &axongo.RewardInput{
				Index: i,
			}
			tx.TransactionEssence.ContextInputs = append(tx.TransactionEssence.ContextInputs, rewardInput)
		}
	}
}

func WithCommitmentInput() options.Option[axongo.Transaction] {
	return func(tx *axongo.Transaction) {
		tx.TransactionEssence.ContextInputs = append(tx.TransactionEssence.ContextInputs, RandCommitmentInput())
	}
}

func WithUTXOInputCount(inputCount int) options.Option[axongo.Transaction] {
	return func(tx *axongo.Transaction) {
		tx.TransactionEssence.Inputs = make(axongo.TxEssenceInputs, 0, inputCount)

		for i := inputCount; i > 0; i-- {
			tx.TransactionEssence.Inputs = append(tx.TransactionEssence.Inputs, RandUTXOInput())
		}
	}
}

func WithOutputCount(outputCount int) options.Option[axongo.Transaction] {
	return func(tx *axongo.Transaction) {
		tx.Outputs = make(axongo.TxEssenceOutputs, 0, outputCount)

		for i := outputCount; i > 0; i-- {
			tx.Outputs = append(tx.Outputs, RandBasicOutput(axongo.AddressEd25519))
		}
	}
}

func WithOutputs(outputs axongo.TxEssenceOutputs) options.Option[axongo.Transaction] {
	return func(tx *axongo.Transaction) {
		tx.Outputs = outputs
	}
}

func WithAllotmentCount(allotmentCount int) options.Option[axongo.Transaction] {
	return func(tx *axongo.Transaction) {
		tx.Allotments = RandSortAllotment(allotmentCount)
	}
}

func WithInputs(inputs axongo.TxEssenceInputs) options.Option[axongo.Transaction] {
	return func(tx *axongo.Transaction) {
		tx.TransactionEssence.Inputs = inputs
	}
}

func WithContextInputs(inputs axongo.TxEssenceContextInputs) options.Option[axongo.Transaction] {
	return func(tx *axongo.Transaction) {
		tx.TransactionEssence.ContextInputs = inputs
		tx.TransactionEssence.ContextInputs.Sort()
	}
}

func WithAllotments(allotments axongo.TxEssenceAllotments) options.Option[axongo.Transaction] {
	return func(tx *axongo.Transaction) {
		tx.Allotments = allotments
	}
}

func WithTxEssencePayload(payload axongo.TxEssencePayload) options.Option[axongo.Transaction] {
	return func(tx *axongo.Transaction) {
		tx.Payload = payload
	}
}

// RandSignedTransactionWithTransaction returns a random transaction with a specific essence.
func RandSignedTransactionWithTransaction(api axongo.API, transaction *axongo.Transaction) *axongo.SignedTransaction {
	sigTxPayload := &axongo.SignedTransaction{API: api}
	sigTxPayload.Transaction = transaction

	unlocksCount := len(transaction.TransactionEssence.Inputs)
	for i := unlocksCount; i > 0; i-- {
		sigTxPayload.Unlocks = append(sigTxPayload.Unlocks, RandEd25519SignatureUnlock())
	}

	return sigTxPayload
}

// RandSignedTransaction returns a random transaction.
func RandSignedTransaction(api axongo.API, opts ...options.Option[axongo.Transaction]) *axongo.SignedTransaction {
	return RandSignedTransactionWithTransaction(api, RandTransaction(api, opts...))
}

// RandSignedTransactionWithUTXOInputCount returns a random transaction with a specific amount of inputs.
func RandSignedTransactionWithUTXOInputCount(api axongo.API, inputCount int) *axongo.SignedTransaction {
	return RandSignedTransactionWithTransaction(api, RandTransactionWithInputCount(api, inputCount))
}

// RandSignedTransactionWithOutputCount returns a random transaction with a specific amount of outputs.
func RandSignedTransactionWithOutputCount(api axongo.API, outputCount int) *axongo.SignedTransaction {
	return RandSignedTransactionWithTransaction(api, RandTransactionWithOutputCount(api, outputCount))
}

// RandSignedTransactionWithAllotmentCount returns a random transaction with a specific amount of allotments.
func RandSignedTransactionWithAllotmentCount(api axongo.API, allotmentCount int) *axongo.SignedTransaction {
	return RandSignedTransactionWithTransaction(api, RandTransactionWithAllotmentCount(api, allotmentCount))
}

// RandSignedTransactionWithInputOutputCount returns a random transaction with a specific amount of inputs and outputs.
func RandSignedTransactionWithInputOutputCount(api axongo.API, inputCount int, outputCount int) *axongo.SignedTransaction {
	return RandSignedTransactionWithTransaction(api, RandTransactionWithOptions(api, WithUTXOInputCount(inputCount), WithOutputCount(outputCount)))
}

// OneInputOutputTransaction generates a random transaction with one input and output.
func OneInputOutputTransaction() *axongo.SignedTransaction {
	return &axongo.SignedTransaction{
		API: ZeroCostTestAPI,
		Transaction: &axongo.Transaction{
			API: ZeroCostTestAPI,
			TransactionEssence: &axongo.TransactionEssence{
				NetworkID:     TestNetworkID,
				ContextInputs: axongo.TxEssenceContextInputs{},
				Inputs: axongo.TxEssenceInputs{
					&axongo.UTXOInput{
						TransactionID: func() axongo.TransactionID {
							var b axongo.TransactionID
							copy(b[:], RandBytes(axongo.TransactionIDLength))

							return b
						}(),
						TransactionOutputIndex: 0,
					},
				},
				Allotments:   axongo.Allotments{},
				Capabilities: axongo.TransactionCapabilitiesBitMask{},
				Payload:      nil,
			},
			Outputs: axongo.TxEssenceOutputs{
				&axongo.BasicOutput{
					Amount: 1337,
					UnlockConditions: axongo.BasicOutputUnlockConditions{
						&axongo.AddressUnlockCondition{Address: RandEd25519Address()},
					},
				},
			},
		},
		Unlocks: axongo.Unlocks{
			&axongo.SignatureUnlock{
				Signature: RandEd25519Signature(),
			},
		},
	}
}

package iotago

import (
	"fmt"

	"github.com/iotaledger/hive.go/constraints"
)

const (
	// MaxPayloadSize defines the maximum size of a basic block payload.
	// MaxPayloadSize = MaxBlockSize - block header - empty basic block - one strong parent - block signature.
	MaxPayloadSize = MaxBlockSize - BlockHeaderLength - BasicBlockSizeEmptyParentsAndEmptyPayload - SlotIdentifierLength - Ed25519SignatureSerializedBytesSize
)

// PayloadType denotes a type of payload.
type PayloadType uint32

const (
	// Deprecated payload types
	// PayloadTransactionTIP7 = 0
	// PayloadMilestoneTIP8 = 1
	// PayloadIndexationTIP6 = 2
	// PayloadReceiptTIP17TIP8 = 3.

	// PayloadTreasuryTransaction denotes a TreasuryTransaction.
	PayloadTreasuryTransaction PayloadType = 4
	// PayloadTaggedData denotes a TaggedData payload.
	PayloadTaggedData PayloadType = 5
	// PayloadMilestone denotes a Milestone.
	PayloadMilestone PayloadType = 7
	// PayloadSignedTransaction denotes a SignedTransaction.
	PayloadSignedTransaction PayloadType = 8
)

func (payloadType PayloadType) String() string {
	if int(payloadType) >= len(payloadNames) {
		return fmt.Sprintf("unknown payload type: %d", payloadType)
	}

	return payloadNames[payloadType]
}

var (
	payloadNames = [PayloadMilestone + 1]string{
		"Deprecated-TransactionTIP7",
		"Deprecated-MilestoneTIP8",
		"Deprecated-IndexationTIP6",
		"Deprecated-ReceiptTIP17TIP8",
		"TreasuryTransaction",
		"TaggedData",
		"SignedTransaction",
		"Milestone",
	}
)

// Payload is an object which can be embedded into other objects.
type Payload interface {
	Sizer
	ProcessableObject
	constraints.Cloneable[Payload]

	// PayloadType returns the type of the payload.
	PayloadType() PayloadType
}

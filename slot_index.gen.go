package axongo

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/axonfibre/fibre.go/ierrors"
	"github.com/axonfibre/fibre.go/lo"
	"github.com/axonfibre/fibre.go/serializer/v2"
)

const (
	SlotIndexLength = serializer.UInt32ByteSize
	MaxSlotIndex    = SlotIndex(math.MaxUint32)
)

// SlotIndex is the index of a slot.
type SlotIndex uint32

func SlotIndexFromBytes(b []byte) (SlotIndex, int, error) {
	if len(b) < SlotIndexLength {
		return 0, 0, ierrors.Errorf("invalid length for slot index, expected at least %d bytes, got %d bytes", SlotIndexLength, len(b))
	}

	return SlotIndex(binary.LittleEndian.Uint32(b)), SlotIndexLength, nil
}

func (i SlotIndex) Bytes() ([]byte, error) {
	bytes := make([]byte, SlotIndexLength)
	binary.LittleEndian.PutUint32(bytes, uint32(i))

	return bytes, nil
}

func (i SlotIndex) MustBytes() []byte {
	return lo.PanicOnErr(i.Bytes())
}

func (i SlotIndex) String() string {
	return fmt.Sprintf("SlotIndex(%d)", i)
}

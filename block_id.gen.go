package iotago

// Code generated by go generate; DO NOT EDIT. Check gen/ directory instead.

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"

	"golang.org/x/crypto/blake2b"

	"github.com/axonfibre/fibre.go/ierrors"
	"github.com/axonfibre/axon.go/v4/hexutil"
)

const (
	BlockIDLength = IdentifierLength + SlotIndexLength
)

var (
	EmptyBlockID = BlockID{}
)

// BlockID is a 32 byte hash value together with an 4 byte slot index.
type BlockID [BlockIDLength]byte

// BlockIDRepresentingData returns a new BlockID for the given data by hashing it with blake2b and associating it with the given slot index.
func BlockIDRepresentingData(slot SlotIndex, data []byte) BlockID {
	return NewBlockID(slot, blake2b.Sum256(data))
}

func NewBlockID(slot SlotIndex, idBytes Identifier) BlockID {
	b := BlockID{}
	copy(b[:], idBytes[:])
	binary.LittleEndian.PutUint32(b[IdentifierLength:], uint32(slot))

	return b
}

// BlockIDFromHexString converts the hex to a BlockID representation.
func BlockIDFromHexString(hex string) (BlockID, error) {
	b, err := hexutil.DecodeHex(hex)
	if err != nil {
		return EmptyBlockID, err
	}

	s, _, err := BlockIDFromBytes(b)

	return s, err
}

// IsValidBlockID returns an error if the passed bytes are not a valid BlockID, otherwise nil.
func IsValidBlockID(b []byte) error {
	if len(b) != BlockIDLength {
		return ierrors.Errorf("invalid blockID length: expected %d bytes, got %d bytes", BlockIDLength, len(b))
	}

	return nil
}

// BlockIDFromBytes returns a new BlockID represented by the passed bytes.
func BlockIDFromBytes(b []byte) (BlockID, int, error) {
	if len(b) < BlockIDLength {
		return EmptyBlockID, 0, ierrors.Errorf("invalid length for blockID, expected at least %d bytes, got %d bytes", BlockIDLength, len(b))
	}

	return BlockID(b), BlockIDLength, nil
}

// MustBlockIDFromHexString converts the hex to a BlockID representation.
func MustBlockIDFromHexString(hex string) BlockID {
	b, err := BlockIDFromHexString(hex)
	if err != nil {
		panic(err)
	}

	return b
}

func (b BlockID) Bytes() ([]byte, error) {
	return b[:], nil
}

func (b BlockID) MarshalText() (text []byte, err error) {
	dst := make([]byte, hex.EncodedLen(len(EmptyBlockID)))
	hex.Encode(dst, b[:])

	return dst, nil
}

func (b *BlockID) UnmarshalText(text []byte) error {
	_, err := hex.Decode(b[:], text)

	return err
}

// Empty tells whether the BlockID is empty.
func (b BlockID) Empty() bool {
	return b == EmptyBlockID
}

// ToHex converts the Identifier to its hex representation.
func (b BlockID) ToHex() string {
	return hexutil.EncodeHex(b[:])
}

func (b BlockID) String() string {
	return fmt.Sprintf("BlockID(%s:%d)", b.Alias(), b.Slot())
}

func (b BlockID) Slot() SlotIndex {
	return SlotIndex(binary.LittleEndian.Uint32(b[IdentifierLength:]))
}

// Index returns a slot index to conform with hive's IndexedID interface.
func (b BlockID) Index() SlotIndex {
	return b.Slot()
}

func (b BlockID) Identifier() Identifier {
	return Identifier(b[:IdentifierLength])
}

var (
	// BlockIDAliases contains a dictionary of identifiers associated to their human-readable alias.
	BlockIDAliases = make(map[BlockID]string)

	// blockIDAliasesMutex is the mutex that is used to synchronize access to the previous map.
	blockIDAliasesMutex = sync.RWMutex{}
)

// RegisterAlias allows to register a human-readable alias for the Identifier which will be used as a replacement for
// the String method.
func (b BlockID) RegisterAlias(alias string) {
	blockIDAliasesMutex.Lock()
	defer blockIDAliasesMutex.Unlock()

	BlockIDAliases[b] = alias
}

// Alias returns the human-readable alias of the Identifier (or the base58 encoded bytes of no alias was set).
func (b BlockID) Alias() (alias string) {
	blockIDAliasesMutex.RLock()
	defer blockIDAliasesMutex.RUnlock()

	if existingAlias, exists := BlockIDAliases[b]; exists {
		return existingAlias
	}

	return b.ToHex()
}

// UnregisterAlias allows to unregister a previously registered alias.
func (b BlockID) UnregisterAlias() {
	blockIDAliasesMutex.Lock()
	defer blockIDAliasesMutex.Unlock()

	delete(BlockIDAliases, b)
}

// Compare compares two BlockIDs.
func (b BlockID) Compare(other BlockID) int {
	return bytes.Compare(b[:], other[:])
}

type BlockIDs []BlockID

// ToHex converts the BlockIDs to their hex representation.
func (ids BlockIDs) ToHex() []string {
	hexIDs := make([]string, len(ids))
	for i, b := range ids {
		hexIDs[i] = hexutil.EncodeHex(b[:])
	}

	return hexIDs
}

// RemoveDupsAndSort removes duplicated BlockIDs and sorts the slice by the lexical ordering.
func (ids BlockIDs) RemoveDupsAndSort() BlockIDs {
	sorted := append(BlockIDs{}, ids...)
	sort.Slice(sorted, func(i, j int) bool {
		return bytes.Compare(sorted[i][:], sorted[j][:]) == -1
	})

	var result BlockIDs
	var prev BlockID
	for i, b := range sorted {
		if i == 0 || !bytes.Equal(prev[:], b[:]) {
			result = append(result, b)
		}
		prev = b
	}

	return result
}

// Sort sorts the BlockIDs lexically and in-place.
func (ids BlockIDs) Sort() {
	sort.Slice(ids, func(i, j int) bool {
		return ids[i].Compare(ids[j]) < 0
	})
}

// BlockIDsFromHexString converts the given block IDs from their hex to BlockID representation.
func BlockIDsFromHexString(BlockIDsHex []string) (BlockIDs, error) {
	result := make(BlockIDs, len(BlockIDsHex))

	for i, hexString := range BlockIDsHex {
		BlockID, err := BlockIDFromHexString(hexString)
		if err != nil {
			return nil, err
		}
		result[i] = BlockID
	}

	return result, nil
}

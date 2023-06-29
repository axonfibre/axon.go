package iotago

import (
	"bytes"
	"crypto/ed25519"
	"fmt"
	"sort"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/blake2b"

	hiveEd25519 "github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

const (
	// BlockIDLength defines the length of a block ID.
	BlockIDLength = SlotIdentifierLength
	// MaxBlockSize defines the maximum size of a block.
	MaxBlockSize = 32768
	// BlockMinStrongParents defines the minimum amount of strong parents in a block.
	BlockMinStrongParents = 1
	// BlockMinParents defines the minimum amount of non-strong parents in a block.
	BlockMinParents = 0
	// BlockMaxParents defines the maximum amount of parents in a block.
	BlockMaxParents = 8
	// BlockTypeValidatorMaxParents defines the maximum amount of parents in a ValidatorBlock. TODO: replace number with committee size.
	BlockTypeValidatorMaxParents = BlockMaxParents + 42
)

// BlockType denotes a type of Block.
type BlockType byte

const (
	BlockTypeBasic     BlockType = 1
	BlockTypeValidator BlockType = 2
)

// EmptyBlockID returns an empty BlockID.
func EmptyBlockID() BlockID {
	return emptySlotIdentifier
}

// BlockID is the ID of a Block.
type BlockID = SlotIdentifier

// BlockIDs are IDs of blocks.
type BlockIDs []BlockID

// ToHex converts the BlockIDs to their hex representation.
func (ids BlockIDs) ToHex() []string {
	hexIDs := make([]string, len(ids))
	for i, id := range ids {
		hexIDs[i] = hexutil.EncodeHex(id[:])
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
	for i, id := range sorted {
		if i == 0 || !bytes.Equal(prev[:], id[:]) {
			result = append(result, id)
		}
		prev = id
	}
	return result
}

// BlockIDsFromHexString converts the given block IDs from their hex to BlockID representation.
func BlockIDsFromHexString(blockIDsHex []string) (BlockIDs, error) {
	result := make(BlockIDs, len(blockIDsHex))

	for i, hexString := range blockIDsHex {
		blockID, err := SlotIdentifierFromHexString(hexString)
		if err != nil {
			return nil, err
		}
		result[i] = blockID
	}

	return result, nil
}

type BlockPayload interface {
	Payload
}

const BlockHeaderLength = 1 + serializer.UInt64ByteSize + serializer.UInt64ByteSize + CommitmentIDLength + serializer.UInt64ByteSize + AccountIDLength

type BlockHeader struct {
	ProtocolVersion byte      `serix:"0,mapKey=protocolVersion"`
	NetworkID       NetworkID `serix:"1,mapKey=networkId"`

	IssuingTime         time.Time    `serix:"2,mapKey=issuingTime"`
	SlotCommitmentID    CommitmentID `serix:"3,mapKey=slotCommitment"`
	LatestFinalizedSlot SlotIndex    `serix:"4,mapKey=latestFinalizedSlot"`

	IssuerID AccountID `serix:"5,mapKey=issuerID"`
}

func (b *BlockHeader) Hash(api API) (Identifier, error) {
	headerBytes, err := api.Encode(b)
	if err != nil {
		return Identifier{}, fmt.Errorf("failed to serialize block header: %w", err)
	}

	return blake2b.Sum256(headerBytes), nil
}

type ProtocolBlock struct {
	BlockHeader `serix:"0"`

	Block Block `serix:"1,mapKey=block"`

	Signature Signature `serix:"2,mapKey=signature"`
}

func BlockIdentifierFromBlockBytes(blockBytes []byte) (Identifier, error) {
	if len(blockBytes) < BlockHeaderLength+Ed25519SignatureSerializedBytesSize {
		return Identifier{}, errors.New("not enough block bytes")
	}

	length := len(blockBytes)
	// Separate into header hash, block hash and signature bytes so that we are able to recompute the BlockID from an Attestation.
	headerHash := blake2b.Sum256(blockBytes[:BlockHeaderLength])
	blockHash := blake2b.Sum256(blockBytes[BlockHeaderLength : length-Ed25519SignatureSerializedBytesSize])
	signatureBytes := [Ed25519SignatureSerializedBytesSize]byte(blockBytes[length-Ed25519SignatureSerializedBytesSize:])

	return blockIdentifier(headerHash, blockHash, signatureBytes[:]), nil
}

func blockIdentifier(headerHash Identifier, blockHash Identifier, signatureBytes []byte) Identifier {
	return IdentifierFromData(byteutils.ConcatBytes(headerHash[:], blockHash[:], signatureBytes))
}

// SigningMessage returns the to be signed message.
// The BlockHeader and Block are separately hashed and concatenated to enable the verification of the signature for
// an Attestation where only the BlockHeader and the hash of Block is known.
func (b *ProtocolBlock) SigningMessage(api API) ([]byte, error) {
	headerHash, err := b.BlockHeader.Hash(api)
	if err != nil {
		return nil, err
	}

	blockHash, err := b.Block.Hash(api)
	if err != nil {
		return nil, err
	}

	return blockSigningMessage(headerHash, blockHash), nil
}

func blockSigningMessage(headerHash Identifier, blockHash Identifier) []byte {
	return byteutils.ConcatBytes(headerHash[:], blockHash[:])
}

// Sign produces signatures signing the essence for every given AddressKeys.
// The produced signatures are in the same order as the AddressKeys.
func (b *ProtocolBlock) Sign(api API, addrKey AddressKeys) (Signature, error) {
	signMsg, err := b.SigningMessage(api)
	if err != nil {
		return nil, err
	}

	signer := NewInMemoryAddressSigner(addrKey)

	return signer.Sign(addrKey.Address, signMsg)
}

// VerifySignature verifies the Signature of the block.
func (b *ProtocolBlock) VerifySignature(api API) (valid bool, err error) {
	signingMessage, err := b.SigningMessage(api)
	if err != nil {
		return false, err
	}

	edSig, isEdSig := b.Signature.(*Ed25519Signature)
	if !isEdSig {
		return false, fmt.Errorf("only ed2519 signatures supported, got %s", b.Signature.Type())
	}

	if edSig.PublicKey == [ed25519.PublicKeySize]byte{} {
		return false, fmt.Errorf("empty publicKeys are invalid")
	}

	return hiveEd25519.Verify(edSig.PublicKey[:], signingMessage, edSig.Signature[:]), nil
}

// ID computes the ID of the Block.
func (b *ProtocolBlock) ID(api API) (BlockID, error) {
	data, err := api.Encode(b)
	if err != nil {
		return BlockID{}, fmt.Errorf("can't compute block ID: %w", err)
	}

	id, err := BlockIdentifierFromBlockBytes(data)
	if err != nil {
		return BlockID{}, err
	}

	slotIndex := api.TimeProvider().SlotFromTime(b.IssuingTime)

	return NewSlotIdentifier(slotIndex, id), nil
}

// MustID works like ID but panics if the BlockID can't be computed.
func (b *ProtocolBlock) MustID(api API) BlockID {
	blockID, err := b.ID(api)
	if err != nil {
		panic(err)
	}
	return blockID
}

type Block interface {
	Type() BlockType

	StrongParentIDs() BlockIDs
	WeakParentIDs() BlockIDs
	ShallowLikeParentIDs() BlockIDs

	Hash(api API) (Identifier, error)
}

// strongParentIDsBasicBlock is a slice of BlockIDs the BasicBlock strongly references.
type strongParentIDsBasicBlock = BlockIDs

// weakParentIDsBasicBlock is a slice of BlockIDs the block weakly references.
type weakParentIDsBasicBlock = BlockIDs

// shallowLikeParentIDsBasicBlock is a slice of BlockIDs the block shallow like references.
type shallowLikeParentIDsBasicBlock = BlockIDs

// BasicBlock represents a basic vertex in the Tangle/BlockDAG.
type BasicBlock struct {
	// The parents the block references.
	StrongParents      strongParentIDsBasicBlock      `serix:"0,lengthPrefixType=uint8,mapKey=strongParents"`
	WeakParents        weakParentIDsBasicBlock        `serix:"1,lengthPrefixType=uint8,mapKey=weakParents"`
	ShallowLikeParents shallowLikeParentIDsBasicBlock `serix:"2,lengthPrefixType=uint8,mapKey=shallowLikeParents"`

	// The inner payload of the block. Can be nil.
	Payload BlockPayload `serix:"3,optional,mapKey=payload,omitempty"`

	BurnedMana Mana `serix:"4,mapKey=burnedMana"`
}

func (b *BasicBlock) Type() BlockType {
	return BlockTypeBasic
}

func (b *BasicBlock) StrongParentIDs() BlockIDs {
	return b.StrongParents
}

func (b *BasicBlock) WeakParentIDs() BlockIDs {
	return b.WeakParents
}

func (b *BasicBlock) ShallowLikeParentIDs() BlockIDs {
	return b.ShallowLikeParents
}

func (b *BasicBlock) Hash(api API) (Identifier, error) {
	blockBytes, err := api.Encode(b)
	if err != nil {
		return Identifier{}, fmt.Errorf("failed to serialize basic block: %w", err)
	}

	return blake2b.Sum256(blockBytes), nil
}

// strongParentIDsValidatorBlock is a slice of BlockIDs the ValidatorBlock strongly references.
type strongParentIDsValidatorBlock = BlockIDs

// weakParentIDsValidatorBlock is a slice of BlockIDs the block weakly references.
type weakParentIDsValidatorBlock = BlockIDs

// shallowLikeParentIDsValidatorBlock is a slice of BlockIDs the block shallow like references.
type shallowLikeParentIDsValidatorBlock = BlockIDs

// ValidatorBlock represents a validator vertex in the Tangle/BlockDAG.
type ValidatorBlock struct {
	// The parents the block references.
	StrongParents      strongParentIDsValidatorBlock      `serix:"0,lengthPrefixType=uint8,mapKey=strongParents"`
	WeakParents        weakParentIDsValidatorBlock        `serix:"1,lengthPrefixType=uint8,mapKey=weakParents"`
	ShallowLikeParents shallowLikeParentIDsValidatorBlock `serix:"2,lengthPrefixType=uint8,mapKey=shallowLikeParents"`

	HighestSupportedVersion byte `serix:"3,mapKey=latestFinalizedSlot"`
}

func (b *ValidatorBlock) Type() BlockType {
	return BlockTypeValidator
}

func (b *ValidatorBlock) StrongParentIDs() BlockIDs {
	return b.StrongParents
}

func (b *ValidatorBlock) WeakParentIDs() BlockIDs {
	return b.WeakParents
}

func (b *ValidatorBlock) ShallowLikeParentIDs() BlockIDs {
	return b.ShallowLikeParents
}

func (b *ValidatorBlock) Hash(api API) (Identifier, error) {
	blockBytes, err := api.Encode(b)
	if err != nil {
		return Identifier{}, fmt.Errorf("failed to serialize validator block: %w", err)
	}

	return blake2b.Sum256(blockBytes), nil
}

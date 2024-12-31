package axongo

import (
	"cmp"

	"github.com/axonfibre/fibre.go/ierrors"
	"github.com/axonfibre/fibre.go/serializer/v2"
)

const (
	// MinBlockIssuerKeysCount is the minimum amount of block issuer keys allowed for a BlockIssuerFeature.
	MinBlockIssuerKeysCount = 1
	// MaxBlockIssuerKeysCount is the maximum amount of block issuer keys allowed for a BlockIssuerFeature.
	MaxBlockIssuerKeysCount = 128
)

var (
	// ErrBlockIssuerCommitmentInputMissing gets returned when no commitment input was passed in a TX containing a Block Issuer Feature.
	ErrBlockIssuerCommitmentInputMissing = ierrors.New("commitment input missing for block issuer feature")
	// ErrBlockIssuanceCreditInputMissing gets returned when a transaction containing an account with a block issuer feature
	// does not have a Block Issuance Credit Input.
	ErrBlockIssuanceCreditInputMissing = ierrors.New("block issuance credit input missing for account with block issuer feature")
	// ErrBlockIssuerNotExpired gets returned when an unexpired block issuer feature is removed.
	ErrBlockIssuerNotExpired = ierrors.New("block issuer feature has not expired")
	// ErrBlockIssuerExpiryTooEarly gets returned when the expiry slot in a block issuer feature is set too early.
	ErrBlockIssuerExpiryTooEarly = ierrors.New("block issuer feature expiry set too early")
	// ErrManaMovedOffBlockIssuerAccount gets returned when mana is moved off block issuer accounts.
	ErrManaMovedOffBlockIssuerAccount = ierrors.New("mana cannot be moved off block issuer accounts except with manalocks")
	// ErrAccountLocked gets returned when an account has negative block issuance credits.
	ErrAccountLocked = ierrors.New("account is locked due to negative block issuance credits")
)

// BlockIssuerFeature is a feature which indicates that this account can issue blocks.
// The feature includes a block issuer address as well as an expiry slot.
type BlockIssuerFeature struct {
	ExpirySlot      SlotIndex       `serix:""`
	BlockIssuerKeys BlockIssuerKeys `serix:",lenPrefix=uint8"`
}

func (s *BlockIssuerFeature) Clone() Feature {
	return &BlockIssuerFeature{
		ExpirySlot:      s.ExpirySlot,
		BlockIssuerKeys: s.BlockIssuerKeys,
	}
}

func (s *BlockIssuerFeature) StorageScore(storageScoreStruct *StorageScoreStructure, _ StorageScoreFunc) StorageScore {
	return s.BlockIssuerKeys.StorageScore(storageScoreStruct, nil)
}

func (s *BlockIssuerFeature) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	// block issuer feature requires invocation of account and mana managers, so requires extra work.
	return workScoreParameters.BlockIssuer, nil
}

func (s *BlockIssuerFeature) Compare(other Feature) int {
	return cmp.Compare(s.Type(), other.Type())
}

func (s *BlockIssuerFeature) Equal(other Feature) bool {
	otherFeat, is := other.(*BlockIssuerFeature)
	if !is {
		return false
	}

	if s.ExpirySlot != otherFeat.ExpirySlot {
		return false
	}

	if len(s.BlockIssuerKeys) != len(otherFeat.BlockIssuerKeys) {
		return false
	}
	for i := range s.BlockIssuerKeys {
		if !s.BlockIssuerKeys[i].Equal(otherFeat.BlockIssuerKeys[i]) {
			return false
		}
	}

	return true
}

func (s *BlockIssuerFeature) Type() FeatureType {
	return FeatureBlockIssuer
}

func (s *BlockIssuerFeature) Size() int {
	// FeatureType + ExpirySlot + BlockIssuerKeys
	return serializer.SmallTypeDenotationByteSize + SlotIndexLength + s.BlockIssuerKeys.Size()
}

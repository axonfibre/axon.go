package axongo

import (
	"bytes"
	"math"
	"sort"

	"github.com/axonfibre/fibre.go/core/safemath"
	"github.com/axonfibre/fibre.go/ierrors"
	"github.com/axonfibre/fibre.go/lo"
	"github.com/axonfibre/fibre.go/serializer/v2"
)

// BlockIssuanceCredits defines the type of block issuance credits.
type BlockIssuanceCredits int64

const MaxBlockIssuanceCredits = BlockIssuanceCredits(math.MaxInt64)

// Allotment is a struct that represents a list of account IDs and an allotted value.
type Allotment struct {
	AccountID AccountID `serix:""`
	Mana      Mana      `serix:""`
}

func (a *Allotment) Clone() *Allotment {
	return &Allotment{
		AccountID: a.AccountID,
		Mana:      a.Mana,
	}
}

func (a *Allotment) Compare(other *Allotment) int {
	return bytes.Compare(a.AccountID[:], other.AccountID[:])
}

// Allotments is a slice of Allotment.
type Allotments []*Allotment

func (a Allotments) Clone() Allotments {
	return lo.CloneSlice(a)
}

// Sort sorts the allotments in lexical order.
func (a Allotments) Sort() {
	sort.Slice(a, func(i, j int) bool {
		return a[i].Compare(a[j]) < 0
	})
}

func (a Allotments) Size() int {
	// LengthPrefixType
	return serializer.UInt16ByteSize + len(a)*(AccountIDLength+ManaSize)
}

func (a Allotments) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	// Allotments requires invocation of account managers, so requires extra work.
	workScoreAllotments, err := workScoreParameters.Allotment.Multiply(len(a))
	if err != nil {
		return 0, err
	}

	return workScoreAllotments, nil
}

func (a Allotments) Get(id AccountID) Mana {
	for _, allotment := range a {
		if allotment.AccountID == id {
			return allotment.Mana
		}
	}

	return 0
}

// allotmentMaxManaValidator checks that the sum of all allotted mana does not exceed 2^(Mana Bits Count) - 1.
func allotmentMaxManaValidator(maxManaValue Mana) ElementValidationFunc[*Allotment] {
	var sum Mana

	return func(index int, next *Allotment) error {
		var err error
		sum, err = safemath.SafeAdd(sum, next.Mana)
		if err != nil {
			return ierrors.Join(ErrMaxManaExceeded, ierrors.Wrapf(err, "allotment mana sum calculation failed at allotment %d", index))
		}

		if sum > maxManaValue {
			return ierrors.WithMessagef(ErrMaxManaExceeded, "sum of allotted mana exceeds max value with allotment %d", index)
		}

		return nil
	}
}

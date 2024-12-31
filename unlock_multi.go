package iotago

import (
	"github.com/axonfibre/fibre.go/lo"
	"github.com/axonfibre/fibre.go/serializer/v2"
)

// MultiUnlock is an Unlock which holds a list of unlocks for a multi address.
type MultiUnlock struct {
	// The unlocks for this MultiUnlock.
	Unlocks []Unlock `serix:",lenPrefix=uint8,minLen=2,maxLen=10"`
}

func (u *MultiUnlock) Clone() Unlock {
	return &MultiUnlock{
		Unlocks: lo.CloneSlice(u.Unlocks),
	}
}

func (u *MultiUnlock) Type() UnlockType {
	return UnlockMulti
}

func (u *MultiUnlock) Size() int {
	// UnlockType + Unlocks Length
	sum := serializer.SmallTypeDenotationByteSize + serializer.SmallTypeDenotationByteSize

	for _, unlock := range u.Unlocks {
		sum += unlock.Size()
	}

	return sum
}

func (u *MultiUnlock) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	var sum WorkScore
	for _, unlock := range u.Unlocks {
		unlockWorkScore, err := unlock.WorkScore(workScoreParameters)
		if err != nil {
			return 0, err
		}

		sum, err = sum.Add(unlockWorkScore)
		if err != nil {
			return 0, err
		}
	}

	return sum, nil
}

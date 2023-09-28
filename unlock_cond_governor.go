//nolint:dupl
package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

// GovernorAddressUnlockCondition is an UnlockCondition defining the governor identity for an AccountOutput.
type GovernorAddressUnlockCondition struct {
	Address Address `serix:"0,mapKey=address"`
}

func (s *GovernorAddressUnlockCondition) Clone() UnlockCondition {
	return &GovernorAddressUnlockCondition{Address: s.Address.Clone()}
}

func (s *GovernorAddressUnlockCondition) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return rentStruct.VBFactorData().Multiply(serializer.SmallTypeDenotationByteSize) +
		s.Address.VBytes(rentStruct, nil)
}

func (s *GovernorAddressUnlockCondition) WorkScore(_ *WorkScoreStructure) (WorkScore, error) {
	// GovernorAddressUnlockCondition does not require a signature check on creation, only consumption.
	return 0, nil
}

func (s *GovernorAddressUnlockCondition) Equal(other UnlockCondition) bool {
	otherUnlockCond, is := other.(*GovernorAddressUnlockCondition)
	if !is {
		return false
	}

	return s.Address.Equal(otherUnlockCond.Address)
}

func (s *GovernorAddressUnlockCondition) Type() UnlockConditionType {
	return UnlockConditionGovernorAddress
}

func (s *GovernorAddressUnlockCondition) Size() int {
	// UnlockType + Address
	return serializer.SmallTypeDenotationByteSize + s.Address.Size()
}

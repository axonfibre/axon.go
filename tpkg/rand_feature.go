package tpkg

import (
	iotago "github.com/axonfibre/axon.go/v4"
)

// RandNativeTokenFeature returns a random NativeToken feature.
func RandNativeTokenFeature() *iotago.NativeTokenFeature {
	return &iotago.NativeTokenFeature{
		ID:     RandNativeTokenID(),
		Amount: RandUint256(),
	}
}

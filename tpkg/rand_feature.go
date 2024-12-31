package tpkg

import (
	axongo "github.com/axonfibre/axon.go/v4"
)

// RandNativeTokenFeature returns a random NativeToken feature.
func RandNativeTokenFeature() *axongo.NativeTokenFeature {
	return &axongo.NativeTokenFeature{
		ID:     RandNativeTokenID(),
		Amount: RandUint256(),
	}
}

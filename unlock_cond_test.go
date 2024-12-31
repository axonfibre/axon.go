package axongo_test

import (
	"testing"

	axongo "github.com/axonfibre/axon.go/v4"
	"github.com/axonfibre/axon.go/v4/tpkg"
	"github.com/axonfibre/axon.go/v4/tpkg/frameworks"
)

func TestUnlockConditionsDeSerialize(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name: "ok - AddressUnlockCondition",
			Source: &axongo.AddressUnlockCondition{
				Address: tpkg.RandEd25519Address(),
			},
			Target: &axongo.AddressUnlockCondition{},
		},
		{
			Name: "ok - StorageDepositReturnUnlockCondition",
			Source: &axongo.StorageDepositReturnUnlockCondition{
				ReturnAddress: tpkg.RandEd25519Address(),
				Amount:        1337,
			},
			Target: &axongo.StorageDepositReturnUnlockCondition{},
		},
		{
			Name: "ok - TimelockUnlockCondition",
			Source: &axongo.TimelockUnlockCondition{
				Slot: 1000,
			},
			Target: &axongo.TimelockUnlockCondition{},
		},
		{
			Name: "ok - ExpirationUnlockCondition",
			Source: &axongo.ExpirationUnlockCondition{
				ReturnAddress: tpkg.RandEd25519Address(),
				Slot:          1000,
			},
			Target: &axongo.ExpirationUnlockCondition{},
		},
		{
			Name: "ok - StateControllerAddressUnlockCondition",
			Source: &axongo.StateControllerAddressUnlockCondition{
				Address: tpkg.RandEd25519Address(),
			},
			Target: &axongo.StateControllerAddressUnlockCondition{},
		},
		{
			Name: "ok - GovernorAddressUnlockCondition",
			Source: &axongo.GovernorAddressUnlockCondition{
				Address: tpkg.RandEd25519Address(),
			},
			Target: &axongo.GovernorAddressUnlockCondition{},
		},
		{
			Name: "fail - ImplicitAccountCreationAddress in GovernorAddressUnlockCondition",
			Source: &axongo.GovernorAddressUnlockCondition{
				Address: tpkg.RandImplicitAccountCreationAddress(),
			},
			Target:    &axongo.GovernorAddressUnlockCondition{},
			SeriErr:   axongo.ErrImplicitAccountCreationAddressInInvalidUnlockCondition,
			DeSeriErr: axongo.ErrImplicitAccountCreationAddressInInvalidUnlockCondition,
		},
		{
			Name: "fail - ImplicitAccountCreationAddress in StateControllerAddressUnlockCondition",
			Source: &axongo.StateControllerAddressUnlockCondition{
				Address: tpkg.RandImplicitAccountCreationAddress(),
			},
			Target:    &axongo.StateControllerAddressUnlockCondition{},
			SeriErr:   axongo.ErrImplicitAccountCreationAddressInInvalidUnlockCondition,
			DeSeriErr: axongo.ErrImplicitAccountCreationAddressInInvalidUnlockCondition,
		},
		{
			Name: "fail - ImplicitAccountCreationAddress in ExpirationUnlockCondition",
			Source: &axongo.ExpirationUnlockCondition{
				Slot:          3,
				ReturnAddress: tpkg.RandImplicitAccountCreationAddress(),
			},
			Target:    &axongo.ExpirationUnlockCondition{},
			SeriErr:   axongo.ErrImplicitAccountCreationAddressInInvalidUnlockCondition,
			DeSeriErr: axongo.ErrImplicitAccountCreationAddressInInvalidUnlockCondition,
		},
		{
			Name: "fail - ImplicitAccountCreationAddress in StorageDepositReturnUnlockCondition",
			Source: &axongo.StorageDepositReturnUnlockCondition{
				ReturnAddress: tpkg.RandImplicitAccountCreationAddress(),
			},
			Target:    &axongo.StorageDepositReturnUnlockCondition{},
			SeriErr:   axongo.ErrImplicitAccountCreationAddressInInvalidUnlockCondition,
			DeSeriErr: axongo.ErrImplicitAccountCreationAddressInInvalidUnlockCondition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

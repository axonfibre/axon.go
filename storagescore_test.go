package axongo_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	axongo "github.com/axonfibre/axon.go/v4"
	"github.com/axonfibre/axon.go/v4/tpkg"
	"github.com/axonfibre/axon.go/v4/tpkg/frameworks"
)

func TestStorageScoreParameters_DeSerialize(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name:   "ok",
			Source: tpkg.RandStorageScoreParameters(),
			Target: &axongo.StorageScoreParameters{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

func TestStorageScoreParamtersJSONMarshalling(t *testing.T) {
	storageScoreParameters := &axongo.StorageScoreParameters{
		StorageCost:                 500,
		FactorData:                  1,
		OffsetOutputOverhead:        10,
		OffsetEd25519BlockIssuerKey: 50,
		OffsetStakingFeature:        100,
		OffsetDelegation:            100,
	}
	storageScoreParametersJSON := `{"storageCost":"500","factorData":1,"offsetOutputOverhead":"10","offsetEd25519BlockIssuerKey":"50","offsetStakingFeature":"100","offsetDelegation":"100"}`

	j, err := tpkg.ZeroCostTestAPI.JSONEncode(storageScoreParameters)
	require.NoError(t, err)

	require.Equal(t, storageScoreParametersJSON, string(j))

	decodedStorageScoreStructure := &axongo.StorageScoreParameters{}
	err = tpkg.ZeroCostTestAPI.JSONDecode([]byte(storageScoreParametersJSON), decodedStorageScoreStructure)
	require.NoError(t, err)

	require.Equal(t, storageScoreParameters, decodedStorageScoreStructure)
}

package axongo_test

import (
	"testing"

	axongo "github.com/axonfibre/axon.go/v4"
	"github.com/axonfibre/axon.go/v4/tpkg"
	"github.com/axonfibre/axon.go/v4/tpkg/frameworks"
)

func TestTaggedDataDeSerialize(t *testing.T) {
	const tag = "寿司を作って"

	tests := []*frameworks.DeSerializeTest{
		{
			Name:   "ok",
			Source: tpkg.RandTaggedData([]byte(tag)),
			Target: &axongo.TaggedData{},
		},
		{
			Name:   "empty-tag",
			Source: tpkg.RandTaggedData([]byte{}),
			Target: &axongo.TaggedData{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

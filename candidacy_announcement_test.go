package axongo_test

import (
	"testing"

	axongo "github.com/axonfibre/axon.go/v4"
	"github.com/axonfibre/axon.go/v4/tpkg/frameworks"
)

func TestCandidacyAnnouncmentDeSerialize(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name:   "ok",
			Source: &axongo.CandidacyAnnouncement{},
			Target: &axongo.CandidacyAnnouncement{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

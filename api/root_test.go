package api_test

import (
	"testing"

	axongo "github.com/axonfibre/axon.go/v4"
	"github.com/axonfibre/axon.go/v4/api"
	"github.com/axonfibre/axon.go/v4/tpkg/frameworks"
)

func Test_RootAPIDeSerialize(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name: "ok - HealthResponse",
			Source: &api.HealthResponse{
				IsHealthy: true,
			},
			Target: &api.HealthResponse{},
		},
		{
			Name: "ok - RoutesResponse",
			Source: &api.RoutesResponse{
				Routes: []axongo.PrefixedStringUint8{"route1", "route2"},
			},
			Target: &api.RoutesResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

func Test_RootAPIJSONSerialization(t *testing.T) {
	tests := []*frameworks.JSONEncodeTest{
		{
			Name: "ok - HealthResponse",
			Source: &api.HealthResponse{
				IsHealthy: true,
			},
			Target: `{
	"isHealthy": true
}`,
		},
		{
			Name: "ok - RoutesResponse",
			Source: &api.RoutesResponse{
				Routes: []axongo.PrefixedStringUint8{"route1", "route2"},
			},
			Target: `{
	"routes": [
		"route1",
		"route2"
	]
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

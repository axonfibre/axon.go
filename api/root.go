package api

import (
	axongo "github.com/axonfibre/axon.go/v4"
)

type (
	// HealthResponse defines the health response.
	HealthResponse struct {
		// Whether the node is healthy.
		IsHealthy bool `serix:""`
	}

	// RoutesResponse defines the response of a GET routes REST API call.
	RoutesResponse struct {
		Routes []axongo.PrefixedStringUint8 `serix:",lenPrefix=uint8"`
	}
)

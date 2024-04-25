package nodeclient

import (
	"context"
	"net/http"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/api"
)

type (
	// ManagementClient is a client which queries the optional management functionality of a node.
	ManagementClient interface {
		// PeerByID gets a peer by its identifier.
		PeerByID(ctx context.Context, id string) (*api.PeerInfo, error)
		// RemovePeerByID removes a peer by its identifier.
		RemovePeerByID(ctx context.Context, id string) error
		// Peers returns a list of all peers.
		Peers(ctx context.Context) (*api.PeersResponse, error)
		// AddPeer adds a new peer by libp2p multi address with optional alias.
		AddPeer(ctx context.Context, multiAddress string, alias ...string) (*api.PeerInfo, error)
		// PruneDatabaseBySize prunes the database by target size.
		PruneDatabaseBySize(ctx context.Context, targetDatabaseSize string) (*api.PruneDatabaseResponse, error)
		// PruneDatabaseByEpoch prunes the database by epoch.
		PruneDatabaseByEpoch(ctx context.Context, epoch iotago.EpochIndex) (*api.PruneDatabaseResponse, error)
		// PruneDatabaseByDepth prunes the database by depth.
		PruneDatabaseByDepth(ctx context.Context, depth iotago.EpochIndex) (*api.PruneDatabaseResponse, error)
		// CreateSnapshot creates a snapshot.
		CreateSnapshot(ctx context.Context) (*api.CreateSnapshotResponse, error)
	}

	managementClient struct {
		core *Client
	}
)

// Do executes a request against the endpoint.
// This function is only meant to be used for special routes not covered through the standard API.
func (client *managementClient) Do(ctx context.Context, method string, route string, reqObj interface{}, resObj interface{}) (*http.Response, error) {
	return client.core.Do(ctx, method, route, reqObj, resObj)
}

// DoWithRequestHeaderHook executes a request against the endpoint.
// This function is only meant to be used for special routes not covered through the standard API.
func (client *managementClient) DoWithRequestHeaderHook(ctx context.Context, method string, route string, requestHeaderHook RequestHeaderHook, reqObj interface{}, resObj interface{}) (*http.Response, error) {
	return client.core.DoWithRequestHeaderHook(ctx, method, route, requestHeaderHook, reqObj, resObj)
}

// PeerByID gets a peer by its identifier.
func (client *managementClient) PeerByID(ctx context.Context, id string) (*api.PeerInfo, error) {
	query := api.EndpointWithNamedParameterValue(api.ManagementRoutePeer, api.ParameterPeerID, id)

	res := new(api.PeerInfo)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptJSON, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// RemovePeerByID removes a peer by its identifier.
func (client *managementClient) RemovePeerByID(ctx context.Context, id string) error {
	query := api.EndpointWithNamedParameterValue(api.ManagementRoutePeer, api.ParameterPeerID, id)

	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodDelete, query, nil, nil); err != nil {
		return err
	}

	return nil
}

// Peers returns a list of all peers.
func (client *managementClient) Peers(ctx context.Context) (*api.PeersResponse, error) {
	res := new(api.PeersResponse)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, api.ManagementRoutePeers, RequestHeaderHookAcceptJSON, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// AddPeer adds a new peer by libp2p multi address with optional alias.
func (client *managementClient) AddPeer(ctx context.Context, multiAddress string, alias ...string) (*api.PeerInfo, error) {
	req := &api.AddPeerRequest{
		MultiAddress: multiAddress,
	}

	if len(alias) > 0 {
		req.Alias = alias[0]
	}

	res := new(api.PeerInfo)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodPost, api.ManagementRoutePeers, RequestHeaderHookAcceptJSON, req, res); err != nil {
		return nil, err
	}

	return res, nil
}

// PruneDatabaseBySize prunes the database by target size.
func (client *managementClient) PruneDatabaseBySize(ctx context.Context, targetDatabaseSize string) (*api.PruneDatabaseResponse, error) {
	req := &api.PruneDatabaseRequest{
		TargetDatabaseSize: targetDatabaseSize,
	}

	res := new(api.PruneDatabaseResponse)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodPost, api.ManagementRouteDatabasePrune, RequestHeaderHookAcceptJSON, req, res); err != nil {
		return nil, err
	}

	return res, nil
}

// PruneDatabaseByEpoch prunes the database by epoch.
func (client *managementClient) PruneDatabaseByEpoch(ctx context.Context, epoch iotago.EpochIndex) (*api.PruneDatabaseResponse, error) {
	req := &api.PruneDatabaseRequest{
		Epoch: epoch,
	}

	res := new(api.PruneDatabaseResponse)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodPost, api.ManagementRouteDatabasePrune, RequestHeaderHookAcceptJSON, req, res); err != nil {
		return nil, err
	}

	return res, nil
}

// PruneDatabaseByDepth prunes the database by depth.
func (client *managementClient) PruneDatabaseByDepth(ctx context.Context, depth iotago.EpochIndex) (*api.PruneDatabaseResponse, error) {
	req := &api.PruneDatabaseRequest{
		Depth: depth,
	}

	res := new(api.PruneDatabaseResponse)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodPost, api.ManagementRouteDatabasePrune, RequestHeaderHookAcceptJSON, req, res); err != nil {
		return nil, err
	}

	return res, nil
}

// CreateSnapshot creates a snapshot.
func (client *managementClient) CreateSnapshot(ctx context.Context) (*api.CreateSnapshotResponse, error) {
	res := new(api.CreateSnapshotResponse)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodPost, api.ManagementRouteSnapshotsCreate, RequestHeaderHookAcceptJSON, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

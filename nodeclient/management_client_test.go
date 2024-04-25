package nodeclient_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/api"
	"github.com/iotaledger/iota.go/v4/nodeclient"
)

func TestManagementClient_Enabled(t *testing.T) {
	defer gock.Off()

	originRoutes := &api.RoutesResponse{
		Routes: []iotago.PrefixedStringUint8{api.ManagementPluginName},
	}

	mockGetJSON(api.RouteRoutes, 200, originRoutes)

	client := nodeClient(t)

	_, err := client.Management(context.TODO())
	require.NoError(t, err)
}

func TestManagementClient_Disabled(t *testing.T) {
	defer gock.Off()

	originRoutes := &api.RoutesResponse{
		Routes: []iotago.PrefixedStringUint8{"someplugin/v1"},
	}

	mockGetJSON(api.RouteRoutes, 200, originRoutes)

	client := nodeClient(t)

	_, err := client.Management(context.TODO())
	require.Error(t, err, nodeclient.ErrManagementPluginNotAvailable)
}

func TestManagementClient_PeerByID(t *testing.T) {
	defer gock.Off()

	originRes := &api.PeerInfo{
		MultiAddresses: []iotago.PrefixedStringUint8{iotago.PrefixedStringUint8(fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID))},
		ID:             peerID,
		Connected:      true,
		Relation:       "autopeered",
		GossipMetrics: &api.PeerGossipMetrics{
			PacketsReceived: 1,
			PacketsSent:     2,
		},
	}

	originRoutes := &api.RoutesResponse{
		Routes: []iotago.PrefixedStringUint8{api.ManagementPluginName},
	}

	mockGetJSON(api.RouteRoutes, 200, originRoutes)
	mockGetJSON(api.EndpointWithNamedParameterValue(api.ManagementRoutePeer, api.ParameterPeerID, peerID), 200, originRes)

	client := nodeClient(t)

	management, err := client.Management(context.TODO())
	require.NoError(t, err)

	resp, err := management.PeerByID(context.Background(), peerID)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestManagementClient_RemovePeerByID(t *testing.T) {
	defer gock.Off()

	gock.New(nodeAPIUrl).
		Delete(api.EndpointWithNamedParameterValue(api.ManagementRoutePeer, api.ParameterPeerID, peerID)).
		Reply(200).
		Status(200)

	originRoutes := &api.RoutesResponse{
		Routes: []iotago.PrefixedStringUint8{api.ManagementPluginName},
	}

	mockGetJSON(api.RouteRoutes, 200, originRoutes)

	client := nodeClient(t)

	management, err := client.Management(context.TODO())
	require.NoError(t, err)

	err = management.RemovePeerByID(context.Background(), peerID)
	require.NoError(t, err)
}

func TestManagementClient_Peers(t *testing.T) {
	defer gock.Off()

	peerID2 := "12D3KooWFJ8Nq6gHLLvigTpPdddddsadsadscpJof8Y4y8yFAB32"

	originRes := &api.PeersResponse{
		Peers: []*api.PeerInfo{
			{
				ID:             peerID,
				MultiAddresses: []iotago.PrefixedStringUint8{iotago.PrefixedStringUint8(fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID))},
				Relation:       "autopeered",
				GossipMetrics: &api.PeerGossipMetrics{
					PacketsReceived: 1,
					PacketsSent:     2,
				},
				Connected: true,
			},
			{
				ID:             peerID2,
				MultiAddresses: []iotago.PrefixedStringUint8{iotago.PrefixedStringUint8(fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID2))},
				Alias:          "Peer2",
				Relation:       "static",
				GossipMetrics: &api.PeerGossipMetrics{
					PacketsReceived: 1,
					PacketsSent:     2,
				},
				Connected: true,
			},
		},
	}

	originRoutes := &api.RoutesResponse{
		Routes: []iotago.PrefixedStringUint8{api.ManagementPluginName},
	}

	mockGetJSON(api.RouteRoutes, 200, originRoutes)
	mockGetJSON(api.ManagementRoutePeers, 200, originRes)

	client := nodeClient(t)

	management, err := client.Management(context.TODO())
	require.NoError(t, err)

	resp, err := management.Peers(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestManagementClient_AddPeer(t *testing.T) {
	defer gock.Off()

	multiAddr := fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID)

	originRes := &api.PeerInfo{
		ID:             peerID,
		MultiAddresses: []iotago.PrefixedStringUint8{iotago.PrefixedStringUint8(multiAddr)},
		Relation:       "autopeered",
		Connected:      true,
		GossipMetrics: &api.PeerGossipMetrics{
			PacketsReceived: 1,
			PacketsSent:     2,
		},
	}

	req := &api.AddPeerRequest{MultiAddress: multiAddr}

	originRoutes := &api.RoutesResponse{
		Routes: []iotago.PrefixedStringUint8{api.ManagementPluginName},
	}

	mockGetJSON(api.RouteRoutes, 200, originRoutes)
	mockPostJSON(api.ManagementRoutePeers, 200, req, originRes)

	client := nodeClient(t)

	management, err := client.Management(context.TODO())
	require.NoError(t, err)

	resp, err := management.AddPeer(context.Background(), multiAddr)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestManagementClient_PruneDatabaseBySize(t *testing.T) {
	defer gock.Off()

	targetSize := "1GB"

	originRes := &api.PruneDatabaseResponse{
		Epoch: 1,
	}

	req := &api.PruneDatabaseRequest{TargetDatabaseSize: targetSize}

	originRoutes := &api.RoutesResponse{
		Routes: []iotago.PrefixedStringUint8{api.ManagementPluginName},
	}

	mockGetJSON(api.RouteRoutes, 200, originRoutes)
	mockPostJSON(api.ManagementRouteDatabasePrune, 200, req, originRes)

	client := nodeClient(t)

	management, err := client.Management(context.TODO())
	require.NoError(t, err)

	resp, err := management.PruneDatabaseBySize(context.Background(), targetSize)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestManagementClient_PruneDatabaseByEpoch(t *testing.T) {
	defer gock.Off()

	epoch := iotago.EpochIndex(1)

	originRes := &api.PruneDatabaseResponse{
		Epoch: 1,
	}

	req := &api.PruneDatabaseRequest{Epoch: epoch}

	originRoutes := &api.RoutesResponse{
		Routes: []iotago.PrefixedStringUint8{api.ManagementPluginName},
	}

	mockGetJSON(api.RouteRoutes, 200, originRoutes)
	mockPostJSON(api.ManagementRouteDatabasePrune, 200, req, originRes)

	client := nodeClient(t)

	management, err := client.Management(context.TODO())
	require.NoError(t, err)

	resp, err := management.PruneDatabaseByEpoch(context.Background(), epoch)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestManagementClient_PruneDatabaseByDepth(t *testing.T) {
	defer gock.Off()

	depth := iotago.EpochIndex(1)

	originRes := &api.PruneDatabaseResponse{
		Epoch: 1,
	}

	req := &api.PruneDatabaseRequest{Depth: depth}

	originRoutes := &api.RoutesResponse{
		Routes: []iotago.PrefixedStringUint8{api.ManagementPluginName},
	}

	mockGetJSON(api.RouteRoutes, 200, originRoutes)
	mockPostJSON(api.ManagementRouteDatabasePrune, 200, req, originRes)

	client := nodeClient(t)

	management, err := client.Management(context.TODO())
	require.NoError(t, err)

	resp, err := management.PruneDatabaseByDepth(context.Background(), depth)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestManagementClient_CreateSnapshot(t *testing.T) {
	defer gock.Off()

	originRes := &api.CreateSnapshotResponse{
		Slot:     1,
		FilePath: "filePath",
	}

	originRoutes := &api.RoutesResponse{
		Routes: []iotago.PrefixedStringUint8{api.ManagementPluginName},
	}

	mockGetJSON(api.RouteRoutes, 200, originRoutes)
	mockPostJSON(api.ManagementRouteSnapshotsCreate, 200, nil, originRes)

	client := nodeClient(t)

	management, err := client.Management(context.TODO())
	require.NoError(t, err)

	resp, err := management.CreateSnapshot(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

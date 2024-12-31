package api

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// APIRoot is the root path for all API endpoints.
	APIRoot = "/api"

	// CorePluginName is the name for the core API plugin.
	CorePluginName = "core/v3"

	// ManagementPluginName is the name for the management plugin.
	ManagementPluginName = "management/v1"

	// IndexerPluginName is the name for the indexer plugin.
	IndexerPluginName = "indexer/v2"

	// BlockIssuerPluginName is the name for the blockissuer plugin.
	BlockIssuerPluginName = "blockissuer/v1"

	// MQTTPluginName is the name for the MQTT plugin.
	MQTTPluginName = "mqtt/v2"
)

const (
	// ParameterPageSize is used to specify the page size.
	ParameterPageSize = "pageSize"

	// ParameterCursor is used to specify the point from which the response should continue for paginated results.
	ParameterCursor = "cursor"

	// ParameterSlot is used to identify a slot.
	ParameterSlot = "slot"

	// ParameterEpoch is used to identify an epoch.
	ParameterEpoch = "epoch"

	// ParameterBlockID is used to identify a block by its ID.
	ParameterBlockID = "blockId"

	// ParameterTransactionID is used to identify a transaction by its ID.
	ParameterTransactionID = "transactionId"

	// ParameterOutputID is used to identify an output by its ID.
	ParameterOutputID = "outputId"

	// ParameterCommitmentID is used to identify a slot commitment by its ID.
	ParameterCommitmentID = "commitmentId"

	// ParameterFoundryID is used to identify a foundry by its ID.
	ParameterFoundryID = "foundryId"

	// ParameterDelegationID is used to identify a delegation by its ID.
	ParameterDelegationID = "delegationId"

	// ParameterPeerID is used to identify a peer.
	ParameterPeerID = "peerId"

	// ParameterBech32Address is used to represent bech32 address.
	ParameterBech32Address = "bech32Address"

	// ParameterAddress is used to identify an address.
	ParameterAddress = "address"

	// ParameterAccountAddress is used to identify an account address.
	ParameterAccountAddress = "accountAddress"

	// ParameterAnchorAddress is used to identify an anchor address.
	ParameterAnchorAddress = "anchorAddress"

	// ParameterNFTAddress is used to identify an NFT address.
	ParameterNFTAddress = "nftAddress"

	// ParameterTag is used to identify a tag.
	ParameterTag = "tag"

	// ParameterCondition is used to identify an unlock condition.
	ParameterCondition = "condition"

	// ParameterWorkScore is used to identify work score.
	ParameterWorkScore = "workScore"
)

const (
	MIMEApplicationJSON                   = "application/json"
	MIMEApplicationVendorIOTASerializerV2 = "application/vnd.iota.serializer-v2"
)

var (
	// RouteHealth is the route for querying a node's health status.
	RouteHealth = "/health"

	// RouteRoutes is the route for getting the routes the node supports.
	// GET returns the nodes routes.
	RouteRoutes = route("", "/routes")
)

func route(pluginName, endpoint string) string {
	if len(pluginName) > 0 {
		return fmt.Sprintf("%s/%s%s", APIRoot, pluginName, endpoint)
	}

	return fmt.Sprintf("%s%s", APIRoot, endpoint)
}

func EndpointWithNamedParameterValue(endpoint string, parameter string, value string) string {
	return strings.Replace(endpoint, "{"+parameter+"}", value, 1)
}

func EndpointWithEchoParameters(endpoint string) string {
	return regexp.MustCompile(`\{([^}]*)\}`).ReplaceAllString(endpoint, ":$1")
}

// Core API endpoints.
const (
	// CoreEndpointInfo is the endpoint for getting the node info.
	// GET returns the node info.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointInfo = "/info"

	// CoreEndpointNetworkHealth is the endpoint for getting the network health.
	// GET returns http status code 200 if the network is healthy (finalization is not delayed).
	CoreEndpointNetworkHealth = "/network/health"

	// CoreEndpointNetworkMetrics is the endpoint for getting the network metrics.
	// GET returns the network metrics.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointNetworkMetrics = "/network/metrics"

	// CoreEndpointBlocks is the endpoint for sending new blocks.
	// POST sends a single new block and returns the new block ID.
	// "Content-Type" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointBlocks = "/blocks"

	// CoreEndpointBlock is the endpoint for getting a block by its blockID.
	// GET returns the block.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointBlock = "/blocks/{blockId}"

	// CoreEndpointBlockMetadata is the endpoint for getting block metadata by its blockID.
	// GET returns block metadata.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointBlockMetadata = "/blocks/{blockId}/metadata"

	// CoreEndpointBlockWithMetadata is the endpoint for getting a block, together with its metadata by its blockID.
	// GET returns the block and metadata.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointBlockWithMetadata = "/blocks/{blockId}/full"

	// CoreEndpointBlockIssuance is the endpoint for getting all needed information for block creation.
	// GET returns the data needed to attach a block.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointBlockIssuance = "/blocks/issuance"

	// CoreEndpointOutput is the endpoint for getting an output by its outputID (transactionHash + outputIndex). This includes the proof, that the output corresponds to the requested outputID.
	// GET returns the output.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointOutput = "/outputs/{outputId}"

	// CoreEndpointOutputMetadata is the endpoint for getting output metadata by its outputID (transactionHash + outputIndex) without getting the output itself again.
	// GET returns the output metadata.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointOutputMetadata = "/outputs/{outputId}/metadata"

	// CoreEndpointOutputWithMetadata is the endpoint for getting output, together with its metadata by its outputID (transactionHash + outputIndex).
	// GET returns the output and metadata.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointOutputWithMetadata = "/outputs/{outputId}/full"

	// CoreEndpointTransaction is the endpoint for getting a transaction by its transaction ID.
	// GET returns the transaction.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointTransaction = "/transactions/{transactionId}"

	// CoreEndpointTransactionsIncludedBlock is the endpoint for getting the block that was first confirmed for a given transaction ID.
	// GET returns the block.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointTransactionsIncludedBlock = "/transactions/{transactionId}/included-block"

	// CoreEndpointTransactionsIncludedBlockMetadata is the endpoint for getting the metadata for the block that was first confirmed in the ledger for a given transaction ID.
	// GET returns block metadata.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointTransactionsIncludedBlockMetadata = "/transactions/{transactionId}/included-block/metadata"

	// CoreEndpointTransactionsMetadata is the endpoint for getting the metadata for the given transaction ID.
	// GET returns transaction metadata.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointTransactionsMetadata = "/transactions/{transactionId}/metadata"

	// CoreEndpointCommitmentByID is the endpoint for getting a slot commitment by its ID.
	// GET returns the commitment.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointCommitmentByID = "/commitments/{commitmentId}"

	// CoreEndpointCommitmentByIDUTXOChanges is the endpoint for getting all UTXO changes of a commitment by its ID.
	// GET returns the output IDs of all UTXO changes.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointCommitmentByIDUTXOChanges = "/commitments/{commitmentId}/utxo-changes"

	// CoreEndpointCommitmentByIDUTXOChangesFull is the endpoint for getting all UTXO changes of a commitment by its ID.
	// GET returns the outputs of all UTXO changes including their corresponding output IDs.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointCommitmentByIDUTXOChangesFull = "/commitments/{commitmentId}/utxo-changes/full"

	// CoreEndpointCommitmentBySlot is the endpoint for getting a commitment by its Slot.
	// GET returns the commitment.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointCommitmentBySlot = "/commitments/by-slot/{slot}"

	// CoreEndpointCommitmentBySlotUTXOChanges is the endpoint for getting all UTXO changes of a commitment by its Slot.
	// GET returns the output IDs of all UTXO changes.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointCommitmentBySlotUTXOChanges = "/commitments/by-slot/{slot}/utxo-changes"

	// CoreEndpointCommitmentBySlotUTXOChangesFull is the endpoint for getting all UTXO changes of a commitment by its Slot.
	// GET returns the outputs of all UTXO changes including their corresponding output IDs.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointCommitmentBySlotUTXOChangesFull = "/commitments/by-slot/{slot}/utxo-changes/full"

	// CoreEndpointCongestion is the endpoint for getting the current congestion state and all account related useful details as block issuance credits.
	// GET returns the congestion state related to the specified account. (optional query parameters: "QueryParameterCommitmentID" to specify the used commitment)
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointCongestion = "/accounts/{bech32Address}/congestion"

	// CoreEndpointValidators is the endpoint for getting informations about the current registered validators.
	// GET returns the paginated response with the list of validators.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointValidators = "/validators"

	// CoreEndpointValidatorsAccount is the endpoint for getting details about the validator by its bech32 account address.
	// GET returns the validator details.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointValidatorsAccount = "/validators/{bech32Address}"

	// CoreEndpointRewards is the endpoint for getting the rewards for staking or delegation based on staking account or delegation output.
	// Rewards are decayed up to returned epochEnd index.
	// GET returns the rewards.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointRewards = "/rewards/{outputId}"

	// CoreEndpointCommittee is the endpoint for getting information about the current committee.
	// GET returns the information about the current committee. (optional query parameters: "QueryParameterEpochIndex" to specify the epoch)
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointCommittee = "/committee"
)

var (
	CoreRouteInfo                              = route(CorePluginName, CoreEndpointInfo)
	CoreRouteNetworkHealth                     = route(CorePluginName, CoreEndpointNetworkHealth)
	CoreRouteNetworkMetrics                    = route(CorePluginName, CoreEndpointNetworkMetrics)
	CoreRouteBlocks                            = route(CorePluginName, CoreEndpointBlocks)
	CoreRouteBlock                             = route(CorePluginName, CoreEndpointBlock)
	CoreRouteBlockMetadata                     = route(CorePluginName, CoreEndpointBlockMetadata)
	CoreRouteBlockWithMetadata                 = route(CorePluginName, CoreEndpointBlockWithMetadata)
	CoreRouteBlockIssuance                     = route(CorePluginName, CoreEndpointBlockIssuance)
	CoreRouteOutput                            = route(CorePluginName, CoreEndpointOutput)
	CoreRouteOutputMetadata                    = route(CorePluginName, CoreEndpointOutputMetadata)
	CoreRouteOutputWithMetadata                = route(CorePluginName, CoreEndpointOutputWithMetadata)
	CoreRouteTransaction                       = route(CorePluginName, CoreEndpointTransaction)
	CoreRouteTransactionsIncludedBlock         = route(CorePluginName, CoreEndpointTransactionsIncludedBlock)
	CoreRouteTransactionsIncludedBlockMetadata = route(CorePluginName, CoreEndpointTransactionsIncludedBlockMetadata)
	CoreRouteTransactionsMetadata              = route(CorePluginName, CoreEndpointTransactionsMetadata)
	CoreRouteCommitmentByID                    = route(CorePluginName, CoreEndpointCommitmentByID)
	CoreRouteCommitmentByIDUTXOChanges         = route(CorePluginName, CoreEndpointCommitmentByIDUTXOChanges)
	CoreRouteCommitmentByIDUTXOChangesFull     = route(CorePluginName, CoreEndpointCommitmentByIDUTXOChangesFull)
	CoreRouteCommitmentBySlot                  = route(CorePluginName, CoreEndpointCommitmentBySlot)
	CoreRouteCommitmentBySlotUTXOChanges       = route(CorePluginName, CoreEndpointCommitmentBySlotUTXOChanges)
	CoreRouteCommitmentBySlotUTXOChangesFull   = route(CorePluginName, CoreEndpointCommitmentBySlotUTXOChangesFull)
	CoreRouteCongestion                        = route(CorePluginName, CoreEndpointCongestion)
	CoreRouteValidators                        = route(CorePluginName, CoreEndpointValidators)
	CoreRouteValidatorsAccount                 = route(CorePluginName, CoreEndpointValidatorsAccount)
	CoreRouteRewards                           = route(CorePluginName, CoreEndpointRewards)
	CoreRouteCommittee                         = route(CorePluginName, CoreEndpointCommittee)
)

// Management API endpoints.
const (
	// ManagementEndpointPeer is the endpoint for getting peers by their peerID.
	// GET returns the peer.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	// DELETE deletes the peer.
	ManagementEndpointPeer = "/peers/{peerId}"

	// ManagementEndpointPeers is the endpoint for getting all peers of the node.
	// GET returns a list of all peers.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	// POST adds a new peer.
	// "Content-Type" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	ManagementEndpointPeers = "/peers"

	// ManagementEndpointDatabasePrune is the endpoint to manually prune the database.
	// POST prunes the database.
	// "Content-Type" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	ManagementEndpointDatabasePrune = "/database/prune"

	// ManagementEndpointSnapshotsCreate is the endpoint to manually create a snapshot files.
	// POST creates a full snapshot.
	// "Content-Type" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	ManagementEndpointSnapshotsCreate = "/snapshots/create"
)

var (
	ManagementRoutePeer            = route(ManagementPluginName, ManagementEndpointPeer)
	ManagementRoutePeers           = route(ManagementPluginName, ManagementEndpointPeers)
	ManagementRouteDatabasePrune   = route(ManagementPluginName, ManagementEndpointDatabasePrune)
	ManagementRouteSnapshotsCreate = route(ManagementPluginName, ManagementEndpointSnapshotsCreate)
)

// Indexer API endpoints.
const (
	// IndexerEndpointOutputs is the endpoint for getting basic, account, anchor, foundry, nft and delegation outputs filtered by the given parameters.
	// GET with query parameter returns all outputIDs that fit these filter criteria.
	// "Accept" header:
	//		MIMEApplicationJSON => json.
	//		MIMEApplicationVendorIOTASerializerV2 => bytes.
	// Query parameters: "hasNativeToken", "nativeToken", "unlockableByAddress", "createdBefore", "createdAfter"
	// Returns an empty list if no results are found.
	IndexerEndpointOutputs = "/outputs"

	// IndexerEndpointOutputsBasic is the endpoint for getting basic outputs filtered by the given parameters.
	// GET with query parameter returns all outputIDs that fit these filter criteria.
	// "Accept" header:
	//		MIMEApplicationJSON => json.
	//		MIMEApplicationVendorIOTASerializerV2 => bytes.
	// Query parameters: "hasNativeToken", "nativeToken", "address", "unlockableByAddress", "hasStorageDepositReturn", "storageDepositReturnAddress",
	// 					 "hasExpiration", "expiresBefore", "expiresAfter", "expirationReturnAddress",
	//					 "hasTimelock", "timelockedBefore", "timelockedAfter", "sender", "tag",
	//					 "createdBefore", "createdAfter"
	// Returns an empty list if no results are found.
	IndexerEndpointOutputsBasic = "/outputs/basic"

	// IndexerEndpointOutputsAccounts is the endpoint for getting accounts filtered by the given parameters.
	// GET with query parameter returns all outputIDs that fit these filter criteria.
	// "Accept" header:
	//		MIMEApplicationJSON => json.
	//		MIMEApplicationVendorIOTASerializerV2 => bytes.
	// Query parameters: "address", "issuer", "sender",
	//					 "createdBefore", "createdAfter"
	// Returns an empty list if no results are found.
	IndexerEndpointOutputsAccounts = "/outputs/account"

	// IndexerEndpointOutputsAccountByAddress is the endpoint for getting accounts by their accountID.
	// GET returns the outputIDs or 404 if no record is found.
	// "Accept" header:
	//		MIMEApplicationJSON => json.
	//		MIMEApplicationVendorIOTASerializerV2 => bytes.
	IndexerEndpointOutputsAccountByAddress = "/outputs/account/{bech32Address}"

	// IndexerEndpointOutputsAnchors is the endpoint for getting anchors filtered by the given parameters.
	// GET with query parameter returns all outputIDs that fit these filter criteria.
	// "Accept" header:
	//		MIMEApplicationJSON => json.
	//		MIMEApplicationVendorIOTASerializerV2 => bytes.
	// Query parameters: "unlockableByAddress", "stateController", "governor", "issuer", "sender",
	//					 "createdBefore", "createdAfter"
	// Returns an empty list if no results are found.
	IndexerEndpointOutputsAnchors = "/outputs/anchor"

	// IndexerEndpointOutputsAnchorByAddress is the endpoint for getting anchors by their anchorID.
	// GET returns the outputIDs or 404 if no record is found.
	// "Accept" header:
	//		MIMEApplicationJSON => json.
	//		MIMEApplicationVendorIOTASerializerV2 => bytes.
	IndexerEndpointOutputsAnchorByAddress = "/outputs/anchor/{bech32Address}"

	// IndexerEndpointOutputsFoundries is the endpoint for getting foundries filtered by the given parameters.
	// GET with query parameter returns all outputIDs that fit these filter criteria.
	// "Accept" header:
	//		MIMEApplicationJSON => json.
	//		MIMEApplicationVendorIOTASerializerV2 => bytes.
	// Query parameters: "hasNativeToken", "nativeToken", "account", "createdBefore", "createdAfter"
	// Returns an empty list if no results are found.
	IndexerEndpointOutputsFoundries = "/outputs/foundry"

	// IndexerEndpointOutputsFoundryByID is the endpoint for getting foundries by their foundryID.
	// GET returns the outputIDs or 404 if no record is found.
	// "Accept" header:
	//		MIMEApplicationJSON => json.
	//		MIMEApplicationVendorIOTASerializerV2 => bytes.
	IndexerEndpointOutputsFoundryByID = "/outputs/foundry/{foundryId}"

	// IndexerEndpointOutputsNFTs is the endpoint for getting NFT filtered by the given parameters.
	// GET with query parameter returns all outputIDs that fit these filter criteria.
	// "Accept" header:
	//		MIMEApplicationJSON => json.
	//		MIMEApplicationVendorIOTASerializerV2 => bytes.
	// Query parameters: "address", "unlockableByAddress", "hasStorageDepositReturn", "storageDepositReturnAddress",
	// 					 "hasExpiration", "expiresBefore", "expiresAfter", "expirationReturnAddress",
	//					 "hasTimelock", "timelockedBefore", "timelockedAfter", "issuer", "sender", "tag",
	//					 "createdBefore", "createdAfter"
	// Returns an empty list if no results are found.
	IndexerEndpointOutputsNFTs = "/outputs/nft"

	// IndexerEndpointOutputsNFTByAddress is the endpoint for getting NFT by their nftID.
	// GET returns the outputIDs or 404 if no record is found.
	// "Accept" header:
	//		MIMEApplicationJSON => json.
	//		MIMEApplicationVendorIOTASerializerV2 => bytes.
	IndexerEndpointOutputsNFTByAddress = "/outputs/nft/{bech32Address}"

	// IndexerEndpointOutputsDelegations is the endpoint for getting delegations filtered by the given parameters.
	// GET with query parameter returns all outputIDs that fit these filter criteria.
	// "Accept" header:
	//		MIMEApplicationJSON => json.
	//		MIMEApplicationVendorIOTASerializerV2 => bytes.
	// Query parameters: "address", "validator", "createdBefore", "createdAfter"
	// Returns an empty list if no results are found.
	IndexerEndpointOutputsDelegations = "/outputs/delegation"

	// IndexerEndpointOutputsDelegationByID is the endpoint for getting delegations by their delegationID.
	// GET returns the outputIDs or 404 if no record is found.
	// "Accept" header:
	//		MIMEApplicationJSON => json.
	//		MIMEApplicationVendorIOTASerializerV2 => bytes.
	IndexerEndpointOutputsDelegationByID = "/outputs/delegation/{delegationId}"

	// IndexerEndpointMultiAddressByAddress is the endpoint for getting the multi address unlock condition
	// of an MultiAddressReference (can be contained in a RestrictedAddress).
	// GET returns a MultiAddress.
	// "Accept" header:
	//		MIMEApplicationJSON => json.
	//		MIMEApplicationVendorIOTASerializerV2 => bytes.
	IndexerEndpointMultiAddressByAddress = "/multiaddress/{bech32Address}"
)

var (
	IndexerRouteOutputs                 = route(IndexerPluginName, IndexerEndpointOutputs)
	IndexerRouteOutputsBasic            = route(IndexerPluginName, IndexerEndpointOutputsBasic)
	IndexerRouteOutputsAccounts         = route(IndexerPluginName, IndexerEndpointOutputsAccounts)
	IndexerRouteOutputsAccountByAddress = route(IndexerPluginName, IndexerEndpointOutputsAccountByAddress)
	IndexerRouteOutputsAnchors          = route(IndexerPluginName, IndexerEndpointOutputsAnchors)
	IndexerRouteOutputsAnchorByAddress  = route(IndexerPluginName, IndexerEndpointOutputsAnchorByAddress)
	IndexerRouteOutputsFoundries        = route(IndexerPluginName, IndexerEndpointOutputsFoundries)
	IndexerRouteOutputsFoundryByID      = route(IndexerPluginName, IndexerEndpointOutputsFoundryByID)
	IndexerRouteOutputsNFTs             = route(IndexerPluginName, IndexerEndpointOutputsNFTs)
	IndexerRouteOutputsNFTByAddress     = route(IndexerPluginName, IndexerEndpointOutputsNFTByAddress)
	IndexerRouteOutputsDelegations      = route(IndexerPluginName, IndexerEndpointOutputsDelegations)
	IndexerRouteOutputsDelegationByID   = route(IndexerPluginName, IndexerEndpointOutputsDelegationByID)
	IndexerRouteMultiAddressByAddress   = route(IndexerPluginName, IndexerEndpointMultiAddressByAddress)
)

const (
	HeaderBlockIssuerProofOfWorkNonce = "X-IOTA-BlockIssuer-PoW-Nonce"
	HeaderBlockIssuerCommitmentID     = "X-IOTA-BlockIssuer-Commitment-ID"
)

// BlockIssuer API endpoints.
const (
	// BlockIssuerRouteInfo is the endpoint for getting the info of the block issuer.
	// GET returns the info.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	BlockIssuerEndpointInfo = "/info"

	// BlockIssuerRouteIssuePayload is the endpoint for issuing an ApplicationPayload.
	// POST issues the ApplicationPayload.
	// "Content-Type" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	BlockIssuerEndpointIssuePayload = "/issue"
)

var (
	BlockIssuerRouteInfo         = route(BlockIssuerPluginName, BlockIssuerEndpointInfo)
	BlockIssuerRouteIssuePayload = route(BlockIssuerPluginName, BlockIssuerEndpointIssuePayload)
)

// Event API endpoints.
const (
	EventAPITopicSuffixRaw       = "/raw"
	EventAPITopicSuffixAccepted  = "accepted"
	EventAPITopicSuffixConfirmed = "confirmed"

	// HINT: all existing topics always have a "/raw" suffix for the raw payload as well.
	EventAPITopicCommitmentsLatest    = "commitments/latest"    // axongo.Commitment
	EventAPITopicCommitmentsFinalized = "commitments/finalized" // axongo.Commitment

	EventAPITopicBlocks                              = "blocks"                                     // axongo.Block (track all incoming blocks)
	EventAPITopicBlocksValidation                    = "blocks/validation"                          // axongo.Block (track all incoming validation blocks)
	EventAPITopicBlocksBasic                         = "blocks/basic"                               // axongo.Block (track all incoming basic blocks)
	EventAPITopicBlocksBasicTaggedData               = "blocks/basic/tagged-data"                   // axongo.Block (track all incoming basic blocks with tagged data payload)
	EventAPITopicBlocksBasicTaggedDataTag            = "blocks/basic/tagged-data/{tag}"             // axongo.Block (track all incoming basic blocks with specific tagged data payload)
	EventAPITopicBlocksBasicTransaction              = "blocks/basic/transaction"                   // axongo.Block (track all incoming basic blocks with transactions)
	EventAPITopicBlocksBasicTransactionTaggedData    = "blocks/basic/transaction/tagged-data"       // axongo.Block (track all incoming basic blocks with transactions and tagged data)
	EventAPITopicBlocksBasicTransactionTaggedDataTag = "blocks/basic/transaction/tagged-data/{tag}" // axongo.Block (track all incoming basic blocks with transactions and specific tagged data)

	// single block on subscribe and changes in it's metadata (accepted, confirmed).
	EventAPITopicTransactionsIncludedBlockMetadata = "transactions/{transactionId}/included-block-metadata" // api.BlockMetadataResponse (track inclusion of a single transaction)
	EventAPITopicTransactionMetadata               = "transaction-metadata/{transactionId}"                 // api.TransactionMetadataResponse (track a specific transaction)

	// single block on subscribe and changes in it's metadata (accepted, confirmed).
	EventAPITopicBlockMetadata = "block-metadata/{blockId}" // api.BlockMetadataResponse (track changes to a single block)

	// all blocks that arrive after subscribing.
	EventAPITopicBlockMetadataAccepted  = "block-metadata/" + EventAPITopicSuffixAccepted  // api.BlockMetadataResponse (track acceptance of all blocks)
	EventAPITopicBlockMetadataConfirmed = "block-metadata/" + EventAPITopicSuffixConfirmed // api.BlockMetadataResponse (track confirmation of all blocks)

	// single output on subscribe and changes in it's metadata (accepted, committed, spent).
	EventAPITopicOutputs = "outputs/{outputId}" // api.OutputWithMetadataResponse (track changes to a single output)

	// all outputs that arrive after subscribing (on transaction accepted and transaction committed).
	EventAPITopicAccountOutputs                     = "outputs/account/{accountAddress}"     // api.OutputWithMetadataResponse (all changes of the chain output)
	EventAPITopicAnchorOutputs                      = "outputs/anchor/{anchorAddress}"       // api.OutputWithMetadataResponse (all changes of the chain output)
	EventAPITopicFoundryOutputs                     = "outputs/foundry/{foundryId}"          // api.OutputWithMetadataResponse (all changes of the chain output)
	EventAPITopicNFTOutputs                         = "outputs/nft/{nftAddress}"             // api.OutputWithMetadataResponse (all changes of the chain output)
	EventAPITopicDelegationOutputs                  = "outputs/delegation/{delegationId}"    // api.OutputWithMetadataResponse (all changes of the chain output)
	EventAPITopicOutputsByUnlockConditionAndAddress = "outputs/unlock/{condition}/{address}" // api.OutputWithMetadataResponse (all changes to outputs that match the unlock condition)
)

// EventAPIUnlockCondition denotes the different unlock conditions.
type EventAPIUnlockCondition string

const (
	EventAPIUnlockConditionAny              EventAPIUnlockCondition = "+"
	EventAPIUnlockConditionAddress          EventAPIUnlockCondition = "address"
	EventAPIUnlockConditionStorageReturn    EventAPIUnlockCondition = "storage-return"
	EventAPIUnlockConditionExpiration       EventAPIUnlockCondition = "expiration"
	EventAPIUnlockConditionStateController  EventAPIUnlockCondition = "state-controller"
	EventAPIUnlockConditionGovernor         EventAPIUnlockCondition = "governor"
	EventAPIUnlockConditionImmutableAccount EventAPIUnlockCondition = "immutable-account"
)

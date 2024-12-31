package tpkg

import (
	axongo "github.com/axonfibre/axon.go/v4"
)

// RandBlock returns a random block with the given inner payload.
func RandBlock(blockBody axongo.BlockBody, api axongo.API, rmc axongo.Mana) *axongo.Block {
	block := &axongo.Block{
		API: api,
		Header: axongo.BlockHeader{
			ProtocolVersion:  ZeroCostTestAPI.Version(),
			NetworkID:        api.ProtocolParameters().NetworkID(),
			IssuingTime:      RandUTCTime(),
			SlotCommitmentID: axongo.NewEmptyCommitment(api).MustID(),
			IssuerID:         RandAccountID(),
		},
		Body:      blockBody,
		Signature: RandEd25519Signature(),
	}

	if basicBlock, isBasic := blockBody.(*axongo.BasicBlockBody); isBasic {
		burnedMana, err := block.ManaCost(rmc)
		if err != nil {
			panic(err)
		}
		basicBlock.MaxBurnedMana = burnedMana
	}

	return block
}

func RandBasicBlockWithIssuerAndRMC(api axongo.API, issuerID axongo.AccountID, rmc axongo.Mana) *axongo.Block {
	basicBlock := RandBasicBlockBody(api, axongo.PayloadSignedTransaction)

	block := RandBlock(basicBlock, ZeroCostTestAPI, rmc)
	block.Header.IssuerID = issuerID

	return block
}

func RandBasicBlockBodyWithPayload(api axongo.API, payload axongo.ApplicationPayload) *axongo.BasicBlockBody {
	return &axongo.BasicBlockBody{
		API:                api,
		StrongParents:      SortedRandBlockIDs(1 + RandInt(axongo.BasicBlockMaxParents)),
		WeakParents:        axongo.BlockIDs{},
		ShallowLikeParents: axongo.BlockIDs{},
		Payload:            payload,
		MaxBurnedMana:      RandMana(1000),
	}
}

func RandBasicBlockBody(api axongo.API, withPayloadType axongo.PayloadType) *axongo.BasicBlockBody {
	var payload axongo.ApplicationPayload

	//nolint:exhaustive
	switch withPayloadType {
	case axongo.PayloadSignedTransaction:
		payload = RandSignedTransaction(api)
	case axongo.PayloadTaggedData:
		payload = RandTaggedData([]byte("tag"))
	case axongo.PayloadCandidacyAnnouncement:
		payload = &axongo.CandidacyAnnouncement{}
	}

	return RandBasicBlockBodyWithPayload(api, payload)
}

func RandValidationBlockBody(api axongo.API) *axongo.ValidationBlockBody {
	return &axongo.ValidationBlockBody{
		API:                     api,
		StrongParents:           SortedRandBlockIDs(1 + RandInt(axongo.ValidationBlockMaxParents)),
		WeakParents:             axongo.BlockIDs{},
		ShallowLikeParents:      axongo.BlockIDs{},
		HighestSupportedVersion: ZeroCostTestAPI.Version() + 1,
	}
}

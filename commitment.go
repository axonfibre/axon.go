package iotago

import (
	"context"
	"fmt"

	"github.com/axonfibre/fibre.go/ierrors"
	"github.com/axonfibre/fibre.go/serializer/v2"
)

type Commitment struct {
	ProtocolVersion      Version      `serix:""`
	Slot                 SlotIndex    `serix:""`
	PreviousCommitmentID CommitmentID `serix:""`
	RootsID              Identifier   `serix:""`
	CumulativeWeight     uint64       `serix:""`
	ReferenceManaCost    Mana         `serix:""`
}

func NewCommitment(version Version, slot SlotIndex, prevID CommitmentID, rootsID Identifier, cumulativeWeight uint64, rmc Mana) *Commitment {
	return &Commitment{
		ProtocolVersion:      version,
		Slot:                 slot,
		PreviousCommitmentID: prevID,
		RootsID:              rootsID,
		CumulativeWeight:     cumulativeWeight,
		ReferenceManaCost:    rmc,
	}
}

func NewEmptyCommitment(api API) *Commitment {
	return &Commitment{
		ProtocolVersion:   api.ProtocolParameters().Version(),
		Slot:              api.ProtocolParameters().GenesisSlot(),
		ReferenceManaCost: api.ProtocolParameters().CongestionControlParameters().MinReferenceManaCost,
	}
}

func (c *Commitment) ID() (CommitmentID, error) {
	data, err := CommonSerixAPI().Encode(context.TODO(), c)
	if err != nil {
		return CommitmentID{}, ierrors.Wrap(err, "failed to serialize commitment")
	}

	return CommitmentIDRepresentingData(c.Slot, data), nil
}

// MustID works like ID but panics if the CommitmentID can't be computed.
func (c *Commitment) MustID() CommitmentID {
	id, err := c.ID()
	if err != nil {
		panic(err)
	}

	return id
}

func (c *Commitment) Type() ContextInputType {
	return ContextInputCommitment
}

func (c *Commitment) Equals(other *Commitment) bool {
	return c.MustID() == other.MustID() &&
		c.ProtocolVersion == other.ProtocolVersion &&
		c.Slot == other.Slot &&
		c.PreviousCommitmentID == other.PreviousCommitmentID &&
		c.RootsID == other.RootsID &&
		c.CumulativeWeight == other.CumulativeWeight &&
		c.ReferenceManaCost == other.ReferenceManaCost
}

func (c *Commitment) String() string {
	return fmt.Sprintf("Commitment{\n\tIndex: %d\n\tPrevID: %s\n\tRootsID: %s\n\tCumulativeWeight: %d\n\tRMC: %d\n}",
		c.Slot, c.PreviousCommitmentID, c.RootsID, c.CumulativeWeight, c.ReferenceManaCost)
}

func (c *Commitment) Size() int {
	return serializer.OneByte +
		SlotIndexLength +
		CommitmentIDLength +
		IdentifierLength +
		serializer.UInt64ByteSize +
		ManaSize
}

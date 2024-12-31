package axongo

import (
	"github.com/axonfibre/fibre.go/lo"
	"github.com/axonfibre/fibre.go/serializer/v2"
)

// TaggedData is a payload which holds a tag and associated data.
type TaggedData struct {
	// The tag to use to categorize the data.
	Tag []byte `serix:",omitempty,lenPrefix=uint8,maxLen=64"`
	// The data within the payload.
	Data []byte `serix:",omitempty,lenPrefix=uint32,maxLen=8192"`
}

func (u *TaggedData) Clone() Payload {
	return &TaggedData{
		Tag:  lo.CopySlice(u.Tag),
		Data: lo.CopySlice(u.Data),
	}
}

func (u *TaggedData) PayloadType() PayloadType {
	return PayloadTaggedData
}

func (u *TaggedData) Size() int {
	// PayloadType
	return serializer.SmallTypeDenotationByteSize +
		serializer.OneByte + len(u.Tag) +
		serializer.UInt32ByteSize + len(u.Data)
}

func (u *TaggedData) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	// we account for the network traffic only on "Payload" level
	workScoreData, err := workScoreParameters.DataByte.Multiply(u.Size())
	if err != nil {
		return 0, err
	}

	// we include the block offset in the payload WorkScore
	return workScoreParameters.Block.Add(workScoreData)
}

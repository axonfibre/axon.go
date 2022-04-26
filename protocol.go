package iotago

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"golang.org/x/crypto/blake2b"
)

var (
	// ErrMissingProtocolParas is returned when ProtocolParameters are missing for operations which require them.
	ErrMissingProtocolParas = errors.New("missing protocol parameters")
)

// NetworkID defines the ID of the network on which entities operate on.
type NetworkID = uint64

// NetworkIDFromString returns the network ID string's numerical representation.
func NetworkIDFromString(networkIDStr string) NetworkID {
	networkIDBlakeHash := blake2b.Sum256([]byte(networkIDStr))
	return binary.LittleEndian.Uint64(networkIDBlakeHash[:])
}

// ProtocolParameters defines the parameters of the protocol.
type ProtocolParameters struct {
	// The version of the protocol running.
	Version byte
	// The human friendly name of the network.
	NetworkName string
	// The HRP prefix used for Bech32 addresses in the network.
	Bech32HRP NetworkPrefix
	// The minimum pow score of the network.
	MinPowScore float64
	// The rent structure used by given node/network.
	RentStructure RentStructure
	// TokenSupply defines the current token supply on the network.
	TokenSupply uint64
}

func (p ProtocolParameters) NetworkID() NetworkID {
	return NetworkIDFromString(p.NetworkName)
}

type jsonProtocolParameters struct {
	Version       byte          `json:"version"`
	NetworkName   string        `json:"networkName"`
	Bech32HRP     NetworkPrefix `json:"bech32HRP"`
	MinPowScore   float64       `json:"minPowScore"`
	RentStructure RentStructure `json:"rentStructure"`
	TokenSupply   string        `json:"tokenSupply"`
}

func (p ProtocolParameters) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonProtocolParameters{
		Version:       p.Version,
		NetworkName:   p.NetworkName,
		Bech32HRP:     p.Bech32HRP,
		MinPowScore:   p.MinPowScore,
		RentStructure: p.RentStructure,
		TokenSupply:   EncodeUint64(p.TokenSupply),
	})
}

func (p *ProtocolParameters) UnmarshalJSON(data []byte) error {
	aux := &jsonProtocolParameters{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	supply, err := DecodeUint64(aux.TokenSupply)
	if err != nil {
		return err
	}

	p.Version = aux.Version
	p.NetworkName = aux.NetworkName
	p.Bech32HRP = aux.Bech32HRP
	p.MinPowScore = aux.MinPowScore
	p.RentStructure = aux.RentStructure
	p.TokenSupply = supply

	return nil
}

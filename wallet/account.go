package wallet

import (
	"crypto/ed25519"

	"github.com/axonfibre/fibre.go/crypto"
	"github.com/axonfibre/fibre.go/ierrors"
	axongo "github.com/axonfibre/axon.go/v4"
)

// Account represents an account.
type Account interface {
	// ID returns the accountID.
	ID() axongo.AccountID

	// Address returns the account address.
	Address() *axongo.AccountAddress

	// OwnerAddress returns the account owner address.
	OwnerAddress() axongo.Address

	// PrivateKey returns the account private key for signing.
	PrivateKey() ed25519.PrivateKey
}

var _ Account = &Ed25519Account{}

// Ed25519Account is an account that uses an Ed25519 key pair.
type Ed25519Account struct {
	accountID  axongo.AccountID
	privateKey ed25519.PrivateKey
}

// NewEd25519Account creates a new Ed25519Account.
func NewEd25519Account(accountID axongo.AccountID, privateKey ed25519.PrivateKey) *Ed25519Account {
	return &Ed25519Account{
		accountID:  accountID,
		privateKey: privateKey,
	}
}

// ID returns the accountID.
func (e *Ed25519Account) ID() axongo.AccountID {
	return e.accountID
}

func (e *Ed25519Account) Address() *axongo.AccountAddress {
	//nolint:forcetypeassert // we know that this is an AccountAddress
	return e.accountID.ToAddress().(*axongo.AccountAddress)
}

func (e *Ed25519Account) OwnerAddress() axongo.Address {
	ed25519PubKey, ok := e.privateKey.Public().(ed25519.PublicKey)
	if !ok {
		panic("invalid public key type")
	}

	return axongo.Ed25519AddressFromPubKey(ed25519PubKey)
}

// PrivateKey returns the account private key for signing.
func (e *Ed25519Account) PrivateKey() ed25519.PrivateKey {
	return e.privateKey
}

func AccountFromParams(accountHex string, privateKey string) (Account, error) {
	accountID, err := axongo.AccountIDFromHexString(accountHex)
	if err != nil {
		return nil, ierrors.Wrap(err, "invalid accountID hex string")
	}

	privKey, err := crypto.ParseEd25519PrivateKeyFromString(privateKey)
	if err != nil {
		return nil, ierrors.Wrap(err, "invalid ed25519 private key string")
	}

	return NewEd25519Account(accountID, privKey), nil
}

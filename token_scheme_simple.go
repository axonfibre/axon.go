package axongo

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/axonfibre/fibre.go/ierrors"
	"github.com/axonfibre/fibre.go/serializer/v2"
)

var (
	// ErrSimpleTokenSchemeTransition gets returned when a SimpleTokenScheme transition is invalid.
	ErrSimpleTokenSchemeTransition = ierrors.New("simple token scheme transition invalid")
	// ErrSimpleTokenSchemeMintedMeltedTokenDecrease gets returned when a SimpleTokenScheme's minted/melted tokens decreased.
	ErrSimpleTokenSchemeMintedMeltedTokenDecrease = ierrors.New("simple token scheme's minted or melted tokens decreased")
	// ErrSimpleTokenSchemeMintingInvalid gets returned when a SimpleTokenScheme's minted tokens did not increase by the minted amount or melted tokens changed.
	ErrSimpleTokenSchemeMintingInvalid = ierrors.New("simple token scheme's minted tokens did not increase by the minted amount or melted tokens changed")
	// ErrSimpleTokenSchemeMeltingInvalid gets returned when a SimpleTokenScheme's melted tokens did not increase by the melted amount or minted tokens changed.
	ErrSimpleTokenSchemeMeltingInvalid = ierrors.New("simple token scheme's melted tokens did not increase by the melted amount or minted tokens changed")
	// ErrSimpleTokenSchemeMaximumSupplyChanged gets returned when a SimpleTokenScheme's maximum supply changes during a transition.
	ErrSimpleTokenSchemeMaximumSupplyChanged = ierrors.New("simple token scheme's maximum supply cannot change during transition")
	// ErrSimpleTokenSchemeInvalidMaximumSupply gets returned when a SimpleTokenScheme's max supply is invalid.
	ErrSimpleTokenSchemeInvalidMaximumSupply = ierrors.New("simple token scheme's maximum supply is invalid")
	// ErrSimpleTokenSchemeInvalidMintedMeltedTokens gets returned when a SimpleTokenScheme's minted/melted supply is invalid.
	ErrSimpleTokenSchemeInvalidMintedMeltedTokens = ierrors.New("simple token scheme's minted/melted tokens counters are invalid")
	// ErrSimpleTokenSchemeGenesisInvalid gets returned when a newly created simple token scheme's melted tokens are not zero
	// or minted tokens do not equal native token amount in transaction.
	ErrSimpleTokenSchemeGenesisInvalid = ierrors.New("newly created simple token scheme's melted tokens are not zero or minted tokens do not equal native token amount in transaction")
)

// SimpleTokenScheme is a TokenScheme which works with minted/melted/maximum supply counters.
type SimpleTokenScheme struct {
	// The amount of tokens which has been minted.
	MintedTokens *big.Int `serix:""`
	// The amount of tokens which has been melted.
	MeltedTokens *big.Int `serix:""`
	// The maximum supply of tokens controlled.
	MaximumSupply *big.Int `serix:""`
}

func (s *SimpleTokenScheme) Clone() TokenScheme {
	return &SimpleTokenScheme{
		MintedTokens:  new(big.Int).Set(s.MintedTokens),
		MeltedTokens:  new(big.Int).Set(s.MeltedTokens),
		MaximumSupply: new(big.Int).Set(s.MaximumSupply),
	}
}

func (s *SimpleTokenScheme) Equal(other TokenScheme) bool {
	otherTokenScheme, isSameType := other.(*SimpleTokenScheme)
	if !isSameType {
		return false
	}

	if s.MintedTokens.Cmp(otherTokenScheme.MintedTokens) != 0 {
		return false
	}

	if s.MeltedTokens.Cmp(otherTokenScheme.MeltedTokens) != 0 {
		return false
	}

	if s.MaximumSupply.Cmp(otherTokenScheme.MaximumSupply) != 0 {
		return false
	}

	return true
}

func (s *SimpleTokenScheme) StorageScore(_ *StorageScoreStructure, _ StorageScoreFunc) StorageScore {
	return 0
}

func (s *SimpleTokenScheme) Type() TokenSchemeType {
	return TokenSchemeSimple
}

func (s *SimpleTokenScheme) SyntacticalValidation() error {
	if r := s.MaximumSupply.Cmp(common.Big0); r != 1 {
		return ierrors.WithMessage(ErrSimpleTokenSchemeInvalidMaximumSupply, "less than equal zero")
	}

	// minted - melted > 0: can never have melted more than minted
	mintedMeltedDelta := big.NewInt(0).Sub(s.MintedTokens, s.MeltedTokens)
	if r := mintedMeltedDelta.Cmp(common.Big0); r == -1 {
		return ierrors.WithMessagef(ErrSimpleTokenSchemeInvalidMintedMeltedTokens, "minted/melted delta less than zero: %s", mintedMeltedDelta)
	}

	// minted - melted <= max supply: can never have minted more than max supply
	if r := mintedMeltedDelta.Cmp(s.MaximumSupply); r == 1 {
		return ierrors.WithMessagef(ErrSimpleTokenSchemeInvalidMintedMeltedTokens, "minted/melted delta more than maximum supply: %s (delta) vs. %s (max supply)", mintedMeltedDelta, s.MaximumSupply)
	}

	return nil
}

func (s *SimpleTokenScheme) StateTransition(transType ChainTransitionType, nextState TokenScheme, in *big.Int, out *big.Int) error {
	switch transType {
	case ChainTransitionTypeGenesis:
		return s.genesisValid(out)
	case ChainTransitionTypeStateChange:
		return s.stateChangeValid(nextState, in, out)
	case ChainTransitionTypeDestroy:
		return s.destructionValid(out, in)
	default:
		panic(fmt.Sprintf("invalid transition type in SimpleTokenScheme %d", transType))
	}
}

// checks that the melted tokens are zero on genesis and that the minted token count
// equals the amount of tokens on the output side of the transaction.
func (s *SimpleTokenScheme) genesisValid(outSum *big.Int) error {
	switch {
	case s.MeltedTokens.Cmp(common.Big0) != 0:
		return ierrors.WithMessagef(ErrSimpleTokenSchemeGenesisInvalid, "melted supply must be zero at genesis")
	case outSum.Cmp(s.MintedTokens) != 0:
		return ierrors.WithMessagef(ErrSimpleTokenSchemeGenesisInvalid, "output native token amount does not equal minted count: minted %s vs. output tokens %s", s.MintedTokens, outSum)
	}

	return nil
}

// SimpleTokenScheme enforces that all tokens that have been minted are melted when the foundry gets destroyed.
func (s *SimpleTokenScheme) destructionValid(out *big.Int, in *big.Int) error {
	tokenDiff := big.NewInt(0).Sub(out, in)
	if big.NewInt(0).Add(s.MintedTokens, tokenDiff).Cmp(s.MeltedTokens) != 0 {
		return ierrors.WithMessagef(ErrNativeTokenSumUnbalanced, "all minted tokens must have been melted up on destruction: minted (%s) + token diff (%d) != melted tokens (%s)", s.MintedTokens, tokenDiff, s.MeltedTokens)
	}

	return nil
}

// checks the balance between the in/out tokens and the invariants concerning supply counter changes.
// burning of tokens is never allowed while doing this transition.
func (s *SimpleTokenScheme) stateChangeValid(nextState TokenScheme, in *big.Int, out *big.Int) error {
	next, is := nextState.(*SimpleTokenScheme)
	if !is {
		return ierrors.WithMessagef(ErrSimpleTokenSchemeTransition, "can only transition to same type but got %s instead", nextState.Type())
	}

	switch {
	case s.MaximumSupply.Cmp(next.MaximumSupply) != 0:
		return ierrors.WithMessagef(ErrSimpleTokenSchemeMaximumSupplyChanged, "maximum supply mismatch wanted %s but got %s", s.MaximumSupply, next.MaximumSupply)
	case s.MintedTokens.Cmp(next.MintedTokens) == 1:
		return ierrors.WithMessagef(ErrSimpleTokenSchemeMintedMeltedTokenDecrease, "current minted supply (%s) bigger than next minted supply (%s)", s.MintedTokens, next.MintedTokens)
	case s.MeltedTokens.Cmp(next.MeltedTokens) == 1:
		return ierrors.WithMessagef(ErrSimpleTokenSchemeMintedMeltedTokenDecrease, "current melted supply (%s) bigger than next melted supply (%s)", s.MeltedTokens, next.MeltedTokens)
	}

	var (
		tokenDiff         = big.NewInt(0).Sub(out, in)
		tokenDiffType     = tokenDiff.Cmp(common.Big0)
		mintedSupplyDelta = big.NewInt(0).Sub(next.MintedTokens, s.MintedTokens)
		meltedSupplyDelta = big.NewInt(0).Sub(next.MeltedTokens, s.MeltedTokens)
	)

	switch {
	case tokenDiffType == 1:
		// out > in
		switch {
		case mintedSupplyDelta.Cmp(tokenDiff) != 0:
			// positive token diff requires the minted supply delta to equal the token diff
			return ierrors.WithMessagef(ErrSimpleTokenSchemeMintingInvalid, "positive token diff not balanced by minted supply change: next minted supply %s - current minted supply %s = %s != token delta %s", next.MintedTokens, s.MintedTokens, mintedSupplyDelta, tokenDiff)
		case next.MeltedTokens.Cmp(s.MeltedTokens) != 0:
			// must not change melted supply while minting
			return ierrors.WithMessagef(ErrSimpleTokenSchemeMintingInvalid, "positive token diff requires equal melted supply between current/next state: current (melted=%s), next (melted=%s)", s.MeltedTokens, next.MeltedTokens)
		}

	case tokenDiffType == -1:
		// out < in
		switch {
		case meltedSupplyDelta.Cmp(big.NewInt(0).Neg(tokenDiff)) != 0:
			// negative token diff requires the melted supply delta to equal the token diff
			return ierrors.WithMessagef(ErrSimpleTokenSchemeMeltingInvalid, "negative token diff not balanced by melted supply change: next melted supply %s - current melted supply %s = %s != token delta %s", next.MeltedTokens, s.MeltedTokens, meltedSupplyDelta, tokenDiff)
		case next.MintedTokens.Cmp(s.MintedTokens) != 0:
			// must not change minting supply while melting
			return ierrors.WithMessagef(ErrSimpleTokenSchemeMeltingInvalid, "negative token diff requires equal minted supply between current/next state: current (minted=%s), next (minted=%s)", s.MintedTokens, next.MintedTokens)
		}

	case tokenDiffType == 0:
		// out == in
		if s.MintedTokens.Cmp(next.MintedTokens) != 0 || s.MeltedTokens.Cmp(next.MeltedTokens) != 0 {
			// no mutations to minted/melted fields while balance is kept
			return ierrors.WithMessagef(ErrSimpleTokenSchemeInvalidMintedMeltedTokens, "zero token diff requires equal minted/melted supply between current/next state: current (minted/melted=%s/%s), next (minted/melted=%s/%s)", s.MintedTokens, s.MeltedTokens, next.MintedTokens, next.MeltedTokens)
		}
	}

	return nil
}

func (s *SimpleTokenScheme) Size() int {
	// TokenSchemeType + MintedTokens + MeltedTokens + MaximumSupply
	return serializer.OneByte + serializer.UInt256ByteSize + serializer.UInt256ByteSize + serializer.UInt256ByteSize
}

func (s *SimpleTokenScheme) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	// we add the offset for a native token here, since the simple token scheme requires extra work for big.Int calculations
	return workScoreParameters.NativeToken, nil
}

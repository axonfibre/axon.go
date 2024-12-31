//nolint:gosec
package tpkg

import (
	cryptorand "crypto/rand"
	"math"
	"math/big"
	"math/rand"
	"sort"
	"time"

	"github.com/axonfibre/fibre.go/serializer/v2"
	axongo "github.com/axonfibre/axon.go/v4"
)

func RandomRead(p []byte) (n int, err error) {
	return cryptorand.Read(p)
}

// RandByte returns a random byte.
func RandByte() byte {
	return byte(RandInt(256))
}

// RandBytes returns length amount random bytes.
func RandBytes(length int) []byte {
	b := make([]byte, 0, length)
	for range length {
		b = append(b, RandByte())
	}

	return b
}

func RandString(length int) string {
	b := make([]byte, 0, length)
	for range length {
		// Generate random printable ASCII values between 32 and 126 (inclusive)
		b = append(b, byte(RandInt(95)+32)) // 95 printable ASCII characters (126 - 32 + 1)
	}

	return string(b)
}

// RandInt returns a random int.
func RandInt(max int) int {
	return rand.Intn(max)
}

// RandInt8 returns a random int8.
func RandInt8(max int8) int8 {
	return int8(RandInt32(uint32(max)))
}

// RandInt16 returns a random int16.
func RandInt16(max int16) int16 {
	return int16(RandInt32(uint32(max)))
}

// RandInt32 returns a random int32.
func RandInt32(max uint32) int32 {
	return rand.Int31n(int32(max))
}

// RandInt64 returns a random int64.
func RandInt64(max uint64) int64 {
	return rand.Int63n(int64(uint32(max)))
}

// RandUint returns a random uint.
func RandUint(max uint) uint {
	return uint(RandInt(int(max)))
}

// RandUint8 returns a random uint8.
func RandUint8(max uint8) uint8 {
	return uint8(RandInt32(uint32(max)))
}

// RandUint16 returns a random uint16.
func RandUint16(max uint16) uint16 {
	return uint16(RandInt32(uint32(max)))
}

// RandUint32 returns a random uint32.
func RandUint32(max uint32) uint32 {
	return uint32(RandInt64(uint64(max)))
}

// RandUint64 returns a random uint64.
func RandUint64(max uint64) uint64 {
	return uint64(RandInt64(max))
}

// RandFloat64 returns a random float64.
func RandFloat64(max float64) float64 {
	return rand.Float64() * max
}

// RandUTCTime returns a random time from current year until now in UTC.
func RandUTCTime() time.Time {
	now := time.Now()
	beginnigOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	secTillNow := now.Unix() - beginnigOfYear.Unix()

	return time.Unix(beginnigOfYear.Unix()+RandInt64(uint64(secTillNow)), RandInt64(1e9)).UTC()
}

func RandDuration() time.Duration {
	return time.Duration(RandInt64(math.MaxInt64))
}

func RandUint256() *big.Int {
	return new(big.Int).SetUint64(rand.Uint64())
}

// Rand12ByteArray returns an array with 12 random bytes.
func Rand12ByteArray() [12]byte {
	var h [12]byte
	b := RandBytes(12)
	copy(h[:], b)

	return h
}

// Rand32ByteArray returns an array with 32 random bytes.
func Rand32ByteArray() [32]byte {
	var h [32]byte
	b := RandBytes(32)
	copy(h[:], b)

	return h
}

// Rand36ByteArray returns an array with 36 random bytes.
func Rand36ByteArray() [36]byte {
	var h [36]byte
	b := RandBytes(36)
	copy(h[:], b)

	return h
}

// Rand40ByteArray returns an array with 40 random bytes.
func Rand40ByteArray() [40]byte {
	var h [40]byte
	b := RandBytes(40)
	copy(h[:], b)

	return h
}

// Rand50ByteArray returns an array with 38 random bytes.
func Rand50ByteArray() [50]byte {
	var h [50]byte
	b := RandBytes(50)
	copy(h[:], b)

	return h
}

// Rand38ByteArray returns an array with 38 random bytes.
func Rand38ByteArray() [38]byte {
	var h [38]byte
	b := RandBytes(38)
	copy(h[:], b)

	return h
}

// Rand49ByteArray returns an array with 49 random bytes.
func Rand49ByteArray() [49]byte {
	var h [49]byte
	b := RandBytes(49)
	copy(h[:], b)

	return h
}

// Rand64ByteArray returns an array with 64 random bytes.
func Rand64ByteArray() [64]byte {
	var h [64]byte
	b := RandBytes(64)
	copy(h[:], b)

	return h
}

// SortedRand32ByteArray returns a count length slice of sorted 32 byte arrays.
func SortedRand32ByteArray(count int) [][32]byte {
	hashes := make(serializer.LexicalOrdered32ByteArrays, count)
	for i := range count {
		hashes[i] = Rand32ByteArray()
	}
	sort.Sort(hashes)

	return hashes
}

// SortedRand36ByteArray returns a count length slice of sorted 36 byte arrays.
func SortedRand36ByteArray(count int) [][36]byte {
	hashes := make(serializer.LexicalOrdered36ByteArrays, count)
	for i := range count {
		hashes[i] = Rand36ByteArray()
	}
	sort.Sort(hashes)

	return hashes
}

// SortedRand40ByteArray returns a count length slice of sorted 32 byte arrays.
func SortedRand40ByteArray(count int) [][40]byte {
	hashes := make(serializer.LexicalOrdered40ByteArrays, count)
	for i := range count {
		hashes[i] = Rand40ByteArray()
	}
	sort.Sort(hashes)

	return hashes
}

func RandSlot() axongo.SlotIndex {
	return axongo.SlotIndex(RandUint32(uint32(axongo.MaxSlotIndex)))
}

func RandEpoch() axongo.EpochIndex {
	return axongo.EpochIndex(RandUint32(uint32(axongo.MaxEpochIndex)))
}

// RandBaseToken returns a random amount of base token.
func RandBaseToken(max axongo.BaseToken) axongo.BaseToken {
	return axongo.BaseToken(rand.Int63n(int64(uint32(max))))
}

// RandMana returns a random amount of mana.
func RandMana(max axongo.Mana) axongo.Mana {
	return axongo.Mana(rand.Int63n(int64(uint32(max))))
}

// RandTaggedData returns a random tagged data payload.
func RandTaggedData(tag []byte, dataLength ...int) *axongo.TaggedData {
	var data []byte
	switch {
	case len(dataLength) > 0:
		data = RandBytes(dataLength[0])
	default:
		data = RandBytes(RandInt(200) + 1)
	}

	return &axongo.TaggedData{Tag: tag, Data: data}
}

// RandUTXOInput returns a random UTXO input.
func RandUTXOInput() *axongo.UTXOInput {
	return RandUTXOInputWithIndex(uint16(RandInt(axongo.RefUTXOIndexMax)))
}

func RandCommitmentID() axongo.CommitmentID {
	return Rand36ByteArray()
}

func RandCommitment() *axongo.Commitment {
	return &axongo.Commitment{
		ProtocolVersion:      axongo.LatestProtocolVersion(),
		Slot:                 RandSlot(),
		PreviousCommitmentID: RandCommitmentID(),
		RootsID:              RandIdentifier(),
		CumulativeWeight:     RandUint64(math.MaxUint64),
		ReferenceManaCost:    RandMana(axongo.MaxMana),
	}
}

// RandCommitmentInput returns a random Commitment input.
func RandCommitmentInput() *axongo.CommitmentInput {
	return &axongo.CommitmentInput{
		CommitmentID: Rand36ByteArray(),
	}
}

// RandBlockIssuanceCreditInput returns a random BlockIssuanceCreditInput.
func RandBlockIssuanceCreditInput() *axongo.BlockIssuanceCreditInput {
	return &axongo.BlockIssuanceCreditInput{
		AccountID: RandAccountID(),
	}
}

// RandUTXOInputWithIndex returns a random UTXO input with a specific index.
func RandUTXOInputWithIndex(index uint16) *axongo.UTXOInput {
	utxoInput := &axongo.UTXOInput{}
	txID := RandBytes(axongo.TransactionIDLength)
	copy(utxoInput.TransactionID[:], txID)

	utxoInput.TransactionOutputIndex = index

	return utxoInput
}

// RandAllotment returns a random Allotment.
func RandAllotment() *axongo.Allotment {
	return &axongo.Allotment{
		AccountID: RandAccountID(),
		Mana:      RandMana(10000) + 1,
	}
}

// RandSortAllotment returns count sorted Allotments.
func RandSortAllotment(count int) axongo.Allotments {
	var allotments axongo.Allotments
	for range count {
		allotments = append(allotments, RandAllotment())
	}
	allotments.Sort()

	return allotments
}

// RandWorkScore returns a random workscore.
func RandWorkScore(max uint32) axongo.WorkScore {
	return axongo.WorkScore(RandUint32(max))
}

// RandStorageScoreParameters produces random set of  parameters.
func RandStorageScoreParameters() *axongo.StorageScoreParameters {
	return &axongo.StorageScoreParameters{
		StorageCost:                 RandBaseToken(axongo.MaxBaseToken),
		FactorData:                  axongo.StorageScoreFactor(RandUint8(math.MaxUint8)),
		OffsetOutputOverhead:        axongo.StorageScore(RandUint64(math.MaxUint64)),
		OffsetEd25519BlockIssuerKey: axongo.StorageScore(RandUint64(math.MaxUint64)),
		OffsetStakingFeature:        axongo.StorageScore(RandUint64(math.MaxUint64)),
	}
}

// RandWorkScoreParameters produces random workscore structure.
func RandWorkScoreParameters() *axongo.WorkScoreParameters {
	return &axongo.WorkScoreParameters{
		DataByte:         RandWorkScore(math.MaxUint32),
		Block:            RandWorkScore(math.MaxUint32),
		Input:            RandWorkScore(math.MaxUint32),
		ContextInput:     RandWorkScore(math.MaxUint32),
		Output:           RandWorkScore(math.MaxUint32),
		NativeToken:      RandWorkScore(math.MaxUint32),
		Staking:          RandWorkScore(math.MaxUint32),
		BlockIssuer:      RandWorkScore(math.MaxUint32),
		Allotment:        RandWorkScore(math.MaxUint32),
		SignatureEd25519: RandWorkScore(math.MaxUint32),
	}
}

// RandProtocolParameters produces random protocol parameters.
// Some protocol parameters are subject to sanity checks when the protocol parameters are created
// so we use default values here to avoid panics rather than random ones.
func RandProtocolParameters() axongo.ProtocolParameters {
	return axongo.NewV3SnapshotProtocolParameters(
		axongo.WithStorageOptions(
			RandBaseToken(axongo.MaxBaseToken),
			axongo.StorageScoreFactor(RandUint8(math.MaxUint8)),
			axongo.StorageScore(RandUint64(math.MaxUint64)),
			axongo.StorageScore(RandUint64(math.MaxUint64)),
			axongo.StorageScore(RandUint64(math.MaxUint64)),
			axongo.StorageScore(RandUint64(math.MaxUint64)),
		),
		axongo.WithWorkScoreOptions(
			RandWorkScore(math.MaxUint32),
			RandWorkScore(math.MaxUint32),
			RandWorkScore(math.MaxUint32),
			RandWorkScore(math.MaxUint32),
			RandWorkScore(math.MaxUint32),
			RandWorkScore(math.MaxUint32),
			RandWorkScore(math.MaxUint32),
			RandWorkScore(math.MaxUint32),
			RandWorkScore(math.MaxUint32),
			RandWorkScore(math.MaxUint32),
		),
	)
}

func RandTokenScheme() axongo.TokenScheme {
	maxSupply := RandInt64(1_000_000_000)
	mintedTokens := RandInt64(uint64(maxSupply))

	return &axongo.SimpleTokenScheme{
		MintedTokens:  big.NewInt(mintedTokens),
		MaximumSupply: big.NewInt(maxSupply),
		MeltedTokens:  big.NewInt(0),
	}
}

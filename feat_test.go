package iotago_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestFeaturesDeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name: "ok - StakingFeature",
			source: &iotago.StakingFeature{
				StakedAmount: 100,
				FixedCost:    12,
				StartEpoch:   100,
				EndEpoch:     1236,
			},
			target: &iotago.StakingFeature{},
		},
		{
			name: "ok - BlockIssuerFeature",
			source: &iotago.BlockIssuerFeature{
				BlockIssuerKeys: iotago.NewBlockIssuerKeys(
					iotago.Ed25519PublicKeyBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
				),
				ExpirySlot: 10,
			},
			target: &iotago.BlockIssuerFeature{},
		},
		{
			name:   "ok - SenderFeature",
			source: &iotago.SenderFeature{Address: tpkg.RandEd25519Address()},
			target: &iotago.SenderFeature{},
		},
		{
			name:   "ok - Issuer",
			source: &iotago.IssuerFeature{Address: tpkg.RandEd25519Address()},
			target: &iotago.IssuerFeature{},
		},
		{
			name: "ok - MetadataFeature",
			source: &iotago.MetadataFeature{
				Entries: iotago.MetadataFeatureEntries{
					"hello":    []byte("world"),
					"did:iota": []byte("hello digital autonomy"),
					"":         []byte(""),
				},
			},
			target: &iotago.MetadataFeature{},
		},
		{
			name: "ok - StateMetadataFeature",
			source: &iotago.StateMetadataFeature{
				Entries: iotago.StateMetadataFeatureEntries{
					"hello":    []byte("world"),
					"did:iota": []byte("hello digital autonomy"),
					"":         []byte(""),
				},
			},
			target: &iotago.StateMetadataFeature{},
		},
		{
			name: "ok - TagFeature",
			source: &iotago.TagFeature{
				Tag: []byte("hello world"),
			},
			target: &iotago.TagFeature{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestFeaturesMetadata(t *testing.T) {
	tests := []deSerializeTest{
		{
			name: "ok - MetadataFeature",
			source: &iotago.MetadataFeature{
				Entries: iotago.MetadataFeatureEntries{
					"hello":    []byte("world"),
					"did:iota": []byte("hello digital autonomy"),
					"empty":    []byte(""),
				},
			},
			target: &iotago.MetadataFeature{},
		},
		{
			name: "fail - MetadataFeature - non ASCII char in key",
			source: &iotago.MetadataFeature{
				Entries: iotago.MetadataFeatureEntries{
					"hellö": []byte("world"),
				},
			},
			seriErr:   iotago.ErrInvalidMetadataKey,
			deSeriErr: iotago.ErrInvalidMetadataKey,
			target:    &iotago.MetadataFeature{},
		},
		{
			name: "ok - StateMetadataFeature",
			source: &iotago.StateMetadataFeature{
				Entries: iotago.StateMetadataFeatureEntries{
					"hello":    []byte("world"),
					"did:iota": []byte("hello digital autonomy"),
					"empty":    []byte(""),
				},
			},
			target: &iotago.StateMetadataFeature{},
		},
		{
			name: "fail - StateMetadataFeature - non ASCII char in key",
			source: &iotago.StateMetadataFeature{
				Entries: iotago.StateMetadataFeatureEntries{
					"hellö": []byte("world"),
				},
			},
			seriErr:   iotago.ErrInvalidStateMetadataKey,
			deSeriErr: iotago.ErrInvalidStateMetadataKey,
			target:    &iotago.StateMetadataFeature{},
		},
		{
			name: "fail - StateMetadataFeature - space char in key",
			source: &iotago.StateMetadataFeature{
				Entries: iotago.StateMetadataFeatureEntries{
					"space-> ": []byte("world"),
				},
			},
			seriErr:   iotago.ErrInvalidStateMetadataKey,
			deSeriErr: iotago.ErrInvalidStateMetadataKey,
			target:    &iotago.StateMetadataFeature{},
		},
		{
			name: "fail - StateMetadataFeature - ASCII control-character in key",
			source: &iotago.StateMetadataFeature{
				Entries: iotago.StateMetadataFeatureEntries{
					"\x07": []byte("world"),
				},
			},
			seriErr:   iotago.ErrInvalidStateMetadataKey,
			deSeriErr: iotago.ErrInvalidStateMetadataKey,
			target:    &iotago.StateMetadataFeature{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

// Tests that maps are sorted when encoded to and decoded from binary to produce a deterministic result,
// but do not have to be sorted when encoded/decoded to JSON.
func TestFeaturesMetadataLexicalOrdering(t *testing.T) {
	type metadataDeserializeTest struct {
		name   string
		source iotago.Feature
		target iotago.Feature
	}

	tests := []metadataDeserializeTest{
		{
			name: "ok - MetadataFeature",
			source: &iotago.MetadataFeature{
				Entries: iotago.MetadataFeatureEntries{
					"b": []byte("y"),
					"c": []byte("z"),
					"a": []byte("x"),
				},
			},
			target: &iotago.MetadataFeature{},
		},
		{
			name: "ok - StateMetadataFeature",
			source: &iotago.StateMetadataFeature{
				Entries: iotago.StateMetadataFeatureEntries{
					"b": []byte("y"),
					"c": []byte("z"),
					"a": []byte("x"),
				},
			},
			target: &iotago.StateMetadataFeature{},
		},
	}

	for _, test := range tests {
		// Required to avoid triggering the scopelint.
		source := test.source
		target := test.target
		featType := test.source.Type()

		t.Run(test.name, func(t *testing.T) {
			{
				serixData, err := tpkg.ZeroCostTestAPI.Encode(source, serix.WithValidation())
				require.NoError(t, err)

				expected := []byte{
					// Metadata Feature Type
					byte(featType),
					// Map Length
					3,
					// Key Length
					1,
					'a',
					// Little-endian value Length
					1, 0,
					'x',
					// Key Length
					1,
					'b',
					// Little-endian value Length
					1, 0,
					'y',
					// Key Length
					1,
					'c',
					// Little-endian value Length
					1, 0,
					'z',
				}

				require.Equal(t, expected, serixData)

				// Decoding the sorted map should succeed.
				bytesRead, err := tpkg.ZeroCostTestAPI.Decode(serixData, target, serix.WithValidation())
				require.NoError(t, err)
				require.Len(t, serixData, bytesRead)
				require.EqualValues(t, source, target)

				// Swap a and b to make it unsorted.
				serixData[3], serixData[8] = serixData[8], serixData[3]
				// Swap x and y so the maps are equal key-value-wise.
				serixData[6], serixData[11] = serixData[11], serixData[6]

				// Decoding the unsorted map should fail.
				serixTarget := reflect.New(reflect.TypeOf(target).Elem()).Interface()
				_, err = tpkg.ZeroCostTestAPI.Decode(serixData, serixTarget, serix.WithValidation())
				require.ErrorIs(t, err, serializer.ErrArrayValidationOrderViolatesLexicalOrder)
			}

			{
				sourceJSON, err := tpkg.ZeroCostTestAPI.JSONEncode(source, serix.WithValidation())
				require.NoError(t, err)

				json := string(sourceJSON)
				require.Contains(t, json, fmt.Sprintf(`"type":%d`, byte(source.Type())))
				require.Contains(t, json, `"a":"0x78"`)
				require.Contains(t, json, `"b":"0x79"`)
				require.Contains(t, json, `"c":"0x7a"`)

				sortedJSON := fmt.Sprintf(`{"type":%d,"entries":{"a":"0x78","b":"0x79","c":"0x7a"}}`, byte(source.Type()))
				unsortedJSON := fmt.Sprintf(`{"type":%d,"entries":{"b":"0x79","a":"0x78","c":"0x7a"}}`, byte(source.Type()))

				// Both sorted and unsorted input is accepted.
				for _, src := range []string{sortedJSON, unsortedJSON} {
					serixTarget := reflect.New(reflect.TypeOf(target).Elem()).Interface()
					err = tpkg.ZeroCostTestAPI.JSONDecode([]byte(src), serixTarget, serix.WithValidation())
					require.NoError(t, err)
					require.Equal(t, source, serixTarget)
				}
			}
		})
	}
}

func TestMetadataMaxSize(t *testing.T) {
	myKey := "mykey"
	myKeyLen := len(myKey)
	mapLenPrefixSize := 1
	keyLenPrefixSize := 1
	valueLenPrefixSize := 2

	tests := []transactionSerializeTest{
		{
			name: "ok - MetadataFeature size matches max allowed size",
			output: func() iotago.Output {
				output := &iotago.BasicOutput{
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{
							Address: tpkg.RandEd25519Address(),
						},
					},
				}
				output.Amount = 100_000_000
				output.Features = append(output.Features, &iotago.MetadataFeature{
					Entries: iotago.MetadataFeatureEntries{
						iotago.MetadataFeatureEntriesKey(myKey): tpkg.RandBytes(
							iotago.MaxMetadataMapSize - mapLenPrefixSize - myKeyLen - keyLenPrefixSize - valueLenPrefixSize,
						),
					},
				})
				return output
			}(),
		},
		{
			name: "fail - MetadataFeature size exceeds max allowed size by one",
			output: func() iotago.Output {
				output := tpkg.RandBasicOutput()
				output.Amount = 100_000_000
				output.Features = append(output.Features, &iotago.MetadataFeature{
					Entries: iotago.MetadataFeatureEntries{
						iotago.MetadataFeatureEntriesKey(myKey): tpkg.RandBytes(
							iotago.MaxMetadataMapSize - mapLenPrefixSize - myKeyLen - keyLenPrefixSize - valueLenPrefixSize + 1,
						),
					},
				})
				return output
			}(),
			seriErr:   iotago.ErrMetadataExceedsMaxSize,
			deseriErr: iotago.ErrMetadataExceedsMaxSize,
		},
		{
			name: "ok - StateMetadataFeature size matches max allowed size",
			output: func() iotago.Output {
				return &iotago.AnchorOutput{
					Amount: 100_000_000,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{
							Address: tpkg.RandEd25519Address(),
						},
						&iotago.GovernorAddressUnlockCondition{
							Address: tpkg.RandEd25519Address(),
						},
					},
					ImmutableFeatures: iotago.AnchorOutputImmFeatures{},
					Features: iotago.AnchorOutputFeatures{
						&iotago.StateMetadataFeature{
							Entries: iotago.StateMetadataFeatureEntries{
								iotago.StateMetadataFeatureEntriesKey(myKey): tpkg.RandBytes(
									iotago.MaxMetadataMapSize - mapLenPrefixSize - myKeyLen - keyLenPrefixSize - valueLenPrefixSize,
								),
							},
						},
					},
				}
			}(),
		},
		{
			name: "fail - StateMetadataFeature size exceeds max allowed size by one",
			output: func() iotago.Output {
				return &iotago.AnchorOutput{
					Amount: 100_000_000,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{
							Address: tpkg.RandEd25519Address(),
						},
						&iotago.GovernorAddressUnlockCondition{
							Address: tpkg.RandEd25519Address(),
						},
					},
					ImmutableFeatures: iotago.AnchorOutputImmFeatures{},
					Features: iotago.AnchorOutputFeatures{
						&iotago.MetadataFeature{
							Entries: iotago.MetadataFeatureEntries{
								"test": []byte("value_unrelated_to_test"),
							},
						},
						&iotago.StateMetadataFeature{
							Entries: iotago.StateMetadataFeatureEntries{
								iotago.StateMetadataFeatureEntriesKey(myKey): tpkg.RandBytes(
									iotago.MaxMetadataMapSize - mapLenPrefixSize - myKeyLen - keyLenPrefixSize - valueLenPrefixSize + 1,
								),
							},
						},
					},
				}
			}(),
			seriErr:   iotago.ErrMetadataExceedsMaxSize,
			deseriErr: iotago.ErrMetadataExceedsMaxSize,
		},
	}

	for _, test := range tests {
		tst := test.ToDeserializeTest()
		t.Run(test.name, tst.deSerialize)
	}
}

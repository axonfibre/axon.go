package axongo_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	axongo "github.com/axonfibre/axon.go/v4"
	"github.com/axonfibre/axon.go/v4/tpkg"
	"github.com/axonfibre/axon.go/v4/tpkg/frameworks"
)

func TestUnlock_DeSerialize(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name:   "ok - signature",
			Source: tpkg.RandEd25519SignatureUnlock(),
			Target: &axongo.SignatureUnlock{},
		},
		{
			Name:   "ok - reference",
			Source: tpkg.RandReferenceUnlock(),
			Target: &axongo.ReferenceUnlock{},
		},
		{
			Name:   "ok - account",
			Source: tpkg.RandAccountUnlock(),
			Target: &axongo.AccountUnlock{},
		},
		{
			Name:   "ok - anchor",
			Source: tpkg.RandAnchorUnlock(),
			Target: &axongo.AnchorUnlock{},
		},
		{
			Name:   "ok - NFT",
			Source: tpkg.RandNFTUnlock(),
			Target: &axongo.NFTUnlock{},
		},
		{
			Name:   "ok - Multi",
			Source: tpkg.RandMultiUnlock(),
			Target: &axongo.MultiUnlock{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

func TestSignaturesUniqueAndReferenceUnlocksValidator(t *testing.T) {
	tests := []struct {
		name    string
		unlocks axongo.Unlocks
		wantErr error
	}{
		{
			name: "ok",
			unlocks: axongo.Unlocks{
				tpkg.RandEd25519SignatureUnlock(),
				tpkg.RandEd25519SignatureUnlock(),
				&axongo.ReferenceUnlock{Reference: 0},
			},
			wantErr: nil,
		},
		{
			name: "ok - chainable referential unlock",
			unlocks: axongo.Unlocks{
				tpkg.RandEd25519SignatureUnlock(),
				&axongo.AccountUnlock{Reference: 0},
				&axongo.AccountUnlock{Reference: 1},
				&axongo.AnchorUnlock{Reference: 2},
				&axongo.NFTUnlock{Reference: 3},
			},
			wantErr: nil,
		},
		{
			name: "ok - multi unlock",
			unlocks: axongo.Unlocks{
				tpkg.RandEd25519SignatureUnlock(),
				tpkg.RandEd25519SignatureUnlock(),
				&axongo.MultiUnlock{
					Unlocks: []axongo.Unlock{
						&axongo.ReferenceUnlock{Reference: 0},
						&axongo.ReferenceUnlock{Reference: 1},
						&axongo.EmptyUnlock{},
						tpkg.RandEd25519SignatureUnlock(),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - chainable referential unlock in multi unlock",
			unlocks: axongo.Unlocks{
				tpkg.RandEd25519SignatureUnlock(),
				&axongo.AccountUnlock{Reference: 0},
				&axongo.AccountUnlock{Reference: 1},
				&axongo.AnchorUnlock{Reference: 2},
				&axongo.MultiUnlock{
					Unlocks: []axongo.Unlock{
						&axongo.NFTUnlock{Reference: 3},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - duplicate ed25519 sig block",
			unlocks: axongo.Unlocks{
				&axongo.SignatureUnlock{Signature: &axongo.Ed25519Signature{
					PublicKey: [32]byte{},
					Signature: [64]byte{},
				}},
				&axongo.SignatureUnlock{Signature: &axongo.Ed25519Signature{
					PublicKey: [32]byte{},
					Signature: [64]byte{},
				}},
			},
			wantErr: axongo.ErrSignatureUnlockNotUnique,
		},
		{
			name: "fail - signature reuse outside and inside the multi unlocks - 1",
			unlocks: axongo.Unlocks{
				&axongo.SignatureUnlock{Signature: &axongo.Ed25519Signature{
					PublicKey: [32]byte{},
					Signature: [64]byte{},
				}},
				&axongo.MultiUnlock{
					Unlocks: []axongo.Unlock{
						&axongo.SignatureUnlock{Signature: &axongo.Ed25519Signature{
							PublicKey: [32]byte{},
							Signature: [64]byte{},
						}},
					},
				},
			},
			wantErr: axongo.ErrSignatureUnlockNotUnique,
		},
		{
			name: "fail - signature reuse outside and inside the multi unlocks - 2",
			unlocks: axongo.Unlocks{
				&axongo.MultiUnlock{
					Unlocks: []axongo.Unlock{
						&axongo.SignatureUnlock{Signature: &axongo.Ed25519Signature{
							PublicKey: [32]byte{},
							Signature: [64]byte{},
						}},
						&axongo.SignatureUnlock{Signature: &axongo.Ed25519Signature{
							PublicKey: [32]byte{0x01},
							Signature: [64]byte{0x01},
						}},
					},
				},
				&axongo.SignatureUnlock{Signature: &axongo.Ed25519Signature{
					PublicKey: [32]byte{},
					Signature: [64]byte{},
				}},
			},
			wantErr: axongo.ErrSignatureUnlockNotUnique,
		},
		{
			name: "ok - duplicate ed25519 sig block in different multi unlocks",
			unlocks: axongo.Unlocks{
				&axongo.MultiUnlock{
					Unlocks: []axongo.Unlock{
						&axongo.SignatureUnlock{Signature: &axongo.Ed25519Signature{
							PublicKey: [32]byte{},
							Signature: [64]byte{},
						}},
						&axongo.SignatureUnlock{Signature: &axongo.Ed25519Signature{
							PublicKey: [32]byte{0x01},
							Signature: [64]byte{0x01},
						}},
					},
				},
				&axongo.MultiUnlock{
					Unlocks: []axongo.Unlock{
						&axongo.SignatureUnlock{Signature: &axongo.Ed25519Signature{
							PublicKey: [32]byte{},
							Signature: [64]byte{},
						}},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - duplicate multi unlock",
			unlocks: axongo.Unlocks{
				&axongo.MultiUnlock{
					Unlocks: []axongo.Unlock{
						&axongo.SignatureUnlock{Signature: &axongo.Ed25519Signature{
							PublicKey: [32]byte{},
							Signature: [64]byte{},
						}},
					},
				},
				&axongo.MultiUnlock{
					Unlocks: []axongo.Unlock{
						&axongo.SignatureUnlock{Signature: &axongo.Ed25519Signature{
							PublicKey: [32]byte{},
							Signature: [64]byte{},
						}},
					},
				},
			},
			wantErr: axongo.ErrMultiUnlockNotUnique,
		},
		{
			name: "fail - reference unlock invalid reference",
			unlocks: axongo.Unlocks{
				tpkg.RandEd25519SignatureUnlock(),
				tpkg.RandEd25519SignatureUnlock(),
				&axongo.ReferenceUnlock{Reference: 1337},
			},
			wantErr: axongo.ErrReferentialUnlockInvalid,
		},
		{
			name: "fail - reference unlock invalid reference in multi unlock",
			unlocks: axongo.Unlocks{
				tpkg.RandEd25519SignatureUnlock(),
				tpkg.RandEd25519SignatureUnlock(),
				&axongo.MultiUnlock{
					Unlocks: []axongo.Unlock{
						&axongo.ReferenceUnlock{Reference: 1337},
					},
				},
			},
			wantErr: axongo.ErrReferentialUnlockInvalid,
		},
		{
			name: "fail - reference unlock references non sig unlock",
			unlocks: axongo.Unlocks{
				tpkg.RandEd25519SignatureUnlock(),
				&axongo.ReferenceUnlock{Reference: 0},
				&axongo.ReferenceUnlock{Reference: 1},
			},
			wantErr: axongo.ErrReferentialUnlockInvalid,
		},
		{
			name: "fail - reference unlock references non sig unlock in multi unlock",
			unlocks: axongo.Unlocks{
				tpkg.RandEd25519SignatureUnlock(),
				&axongo.ReferenceUnlock{Reference: 0},
				&axongo.MultiUnlock{
					Unlocks: []axongo.Unlock{
						&axongo.ReferenceUnlock{Reference: 1},
					},
				},
			},
			wantErr: axongo.ErrReferentialUnlockInvalid,
		},
		{
			name: "fail - empty unlock outside multi unlock",
			unlocks: axongo.Unlocks{
				tpkg.RandEd25519SignatureUnlock(),
				&axongo.EmptyUnlock{},
			},
			wantErr: axongo.ErrEmptyUnlockOutsideMultiUnlock,
		},
		{
			name: "fail - nested multi unlock",
			unlocks: axongo.Unlocks{
				&axongo.MultiUnlock{
					Unlocks: []axongo.Unlock{
						&axongo.MultiUnlock{
							Unlocks: []axongo.Unlock{
								tpkg.RandEd25519SignatureUnlock(),
							},
						},
					},
				},
			},
			wantErr: axongo.ErrNestedMultiUnlock,
		},
		{
			name: "ok - referenced a multi unlock",
			unlocks: axongo.Unlocks{
				&axongo.MultiUnlock{
					Unlocks: []axongo.Unlock{
						tpkg.RandEd25519SignatureUnlock(),
					},
				},
				&axongo.ReferenceUnlock{Reference: 0},
			},
			wantErr: nil,
		},
		{
			name: "fail - referenced a multi unlock in a multi unlock",
			unlocks: axongo.Unlocks{
				&axongo.MultiUnlock{
					Unlocks: []axongo.Unlock{
						tpkg.RandEd25519SignatureUnlock(),
					},
				},
				&axongo.MultiUnlock{
					Unlocks: []axongo.Unlock{
						&axongo.ReferenceUnlock{Reference: 0},
					},
				},
			},
			wantErr: axongo.ErrReferentialUnlockInvalid,
		},
		{
			name: "fail - referenced a multi unlock in in itself",
			unlocks: axongo.Unlocks{
				&axongo.MultiUnlock{
					Unlocks: []axongo.Unlock{
						&axongo.ReferenceUnlock{Reference: 0},
					},
				},
			},
			wantErr: axongo.ErrReferentialUnlockInvalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valFunc := axongo.SignaturesUniqueAndReferenceUnlocksValidator(tpkg.ZeroCostTestAPI)
			var runErr error
			for index, unlock := range tt.unlocks {
				if err := valFunc(index, unlock); err != nil {
					runErr = err
				}
			}
			require.ErrorIs(t, runErr, tt.wantErr)
		})
	}
}

package axongo_test

import (
	"testing"

	axongo "github.com/axonfibre/axon.go/v4"
	"github.com/axonfibre/axon.go/v4/tpkg"
	"github.com/axonfibre/axon.go/v4/tpkg/frameworks"
)

func TestAllotmentDeSerialize(t *testing.T) {
	type allotmentDeSerializeTest struct {
		name      string
		source    axongo.TxEssenceAllotments
		seriErr   error
		deSeriErr error
	}

	accountID1 := axongo.MustAccountIDFromHexString("0x7238fbce2f6ae391bd4eb2ce1c51085e0945943bb1bb8e9133e29672c8ef2c74")
	accountID2 := axongo.MustAccountIDFromHexString("0x98f3e0f153461a73f09b6f9eedf7acbd11447f86d6ad20817973a2e2c9240f32")
	accountID3 := axongo.MustAccountIDFromHexString("0xf23ae970dc1359ff48f4169b7cec237873992dc30d9eeb6ccacdecc7679e4f69")

	tests := []allotmentDeSerializeTest{
		{
			name: "ok - multiple unique allotments in order",
			source: axongo.TxEssenceAllotments{
				&axongo.Allotment{
					AccountID: accountID1,
					Mana:      5,
				},
				&axongo.Allotment{
					AccountID: accountID2,
					Mana:      4,
				},
				&axongo.Allotment{
					AccountID: accountID3,
					Mana:      6,
				},
			},
		},
		{
			name: "err - account id in allotments not lexically ordered",
			source: axongo.TxEssenceAllotments{
				&axongo.Allotment{
					AccountID: accountID2,
					Mana:      500,
				},
				&axongo.Allotment{
					AccountID: accountID1,
					Mana:      800,
				},
				&axongo.Allotment{
					AccountID: accountID3,
					Mana:      800,
				},
			},
			seriErr:   axongo.ErrArrayValidationOrderViolatesLexicalOrder,
			deSeriErr: axongo.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "err - account id in allotments not unique",
			source: axongo.TxEssenceAllotments{
				&axongo.Allotment{
					AccountID: accountID1,
					Mana:      500,
				},
				&axongo.Allotment{
					AccountID: accountID1,
					Mana:      800,
				},
			},
			seriErr:   axongo.ErrArrayValidationViolatesUniqueness,
			deSeriErr: axongo.ErrArrayValidationViolatesUniqueness,
		},
	}

	for _, test := range tests {
		stx := tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI, &axongo.Transaction{
			API: tpkg.ZeroCostTestAPI,
			TransactionEssence: &axongo.TransactionEssence{
				Allotments:    test.source,
				Capabilities:  axongo.TransactionCapabilitiesBitMaskWithCapabilities(),
				ContextInputs: axongo.TxEssenceContextInputs{},
				NetworkID:     tpkg.ZeroCostTestAPI.ProtocolParameters().NetworkID(),
				Inputs: axongo.TxEssenceInputs{
					tpkg.RandUTXOInput(),
				},
			},
			Outputs: axongo.TxEssenceOutputs{
				tpkg.RandBasicOutput(),
			},
		})

		tst := &frameworks.DeSerializeTest{
			Name:      test.name,
			Source:    stx,
			Target:    &axongo.SignedTransaction{},
			SeriErr:   test.seriErr,
			DeSeriErr: test.deSeriErr,
		}

		t.Run(tst.Name, tst.Run)
	}
}

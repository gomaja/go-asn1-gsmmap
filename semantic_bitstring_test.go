// semantic_bitstring_test.go
//
// BIT STRING octet counts follow ITU-T X.690 (02/2021) §8.6.2,
// checked through the public Marshal and Parse entry points.
package gsmmap

import (
	"errors"
	"testing"

	"github.com/gomaja/go-asn1/runtime/ber"
)

// ITU-T X.690 (02/2021) §8.6.2 requires octets consistent with the bit length.
func TestMarshalBitStringOctetsMatchBitLength(t *testing.T) {
	gprs := func(e *EpsInfo, c *SGSNCapability) *UpdateGprsLocation {
		return &UpdateGprsLocation{
			IMSI: "310260311111111", SgsnNumber: "31631000001", SGSNAddress: "192.168.31.1",
			EpsInfo: e, SGSNCapability: c,
		}
	}
	csg := func(b HexBytes, bits int) *AnyTimeInterrogationRes {
		return &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{
			LocationInformation: &CSLocationInformation{
				VlrNumber:          "31612345678",
				UserCSGInformation: &UserCSGInformation{CsgID: b, CsgIDBits: bits},
			},
		}}
	}
	type marshaler interface{ Marshal() ([]byte, error) }
	for _, tc := range []struct {
		name             string
		msg              marshaler
		want             error
		path, constraint string
	}{
		{"isr-Information 3 bits in 1 octet", gprs(&EpsInfo{IsrInformation: HexBytes{0xE0}, IsrInformationBits: 3}, nil), nil, "", ""},
		{"isr-Information 8 bits in 1 octet", gprs(&EpsInfo{IsrInformation: HexBytes{0xFF}, IsrInformationBits: 8}, nil), nil, "", ""},
		{"isr-Information 3 bits in 2 octets", gprs(&EpsInfo{IsrInformation: HexBytes{0xE0, 0x00}, IsrInformationBits: 3}, nil), ber.ErrInvalidValue, "", ""},
		{"isr-Information 8 bits in 2 octets", gprs(&EpsInfo{IsrInformation: HexBytes{0xFF, 0x00}, IsrInformationBits: 8}, nil), ber.ErrInvalidValue, "", ""},
		{"isr-Information octets without bits", gprs(&EpsInfo{IsrInformation: HexBytes{0xE0}}, nil), ber.ErrInvalidValue, "", ""},
		{"supportedFeatures named bit 0", gprs(nil, &SGSNCapability{SupportedFeatures: &SupportedFeatures{OdbAllApn: true}}), nil, "", ""},
		{"supportedFeatures named bit 39", gprs(nil, &SGSNCapability{SupportedFeatures: &SupportedFeatures{NrAsSecondaryRAT: true}}), nil, "", ""},
		{"ext-SupportedFeatures named bit", gprs(nil, &SGSNCapability{ExtSupportedFeatures: &ExtSupportedFeatures{UnlicensedSpectrumAsSecondaryRAT: true}}), nil, "", ""},
		{"ext-SupportedFeatures unnamed bit 39", gprs(nil, &SGSNCapability{ExtSupportedFeatures: &ExtSupportedFeatures{UnknownBits: HexBytes{0, 0, 0, 0, 1}}}), nil, "", ""},
		{"csg-Id 27 bits in 4 octets", csg(HexBytes{0x11, 0x22, 0x33, 0x40}, 27), nil, "", ""},
		{"csg-Id 27 bits in 10 octets", csg(make(HexBytes, 10), 27), ber.ErrInvalidValue, "", ""},
		{"csg-Id 27 bits in 3 octets", csg(HexBytes{0x11, 0x22, 0x33}, 27), ber.ErrInvalidValue, "", ""},
		{"csg-Id octets without bits", csg(HexBytes{0x11, 0x22, 0x33, 0x40}, 0), nil, "subscriberInfo.locationInformation.userCSGInformation.csg-Id", "SIZE (27)"},
		{"csg-Id negative bits", csg(HexBytes{0}, -1), nil, "subscriberInfo.locationInformation.userCSGInformation.csg-Id", "SIZE (27)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := tc.msg.Marshal()
			if tc.path != "" {
				wantConstraintError(t, err, tc.path, tc.constraint)
			} else if !errors.Is(err, tc.want) {
				t.Fatalf("Marshal: err = %v, want %v", err, tc.want)
			}
			if err != nil {
				return
			}
			var perr error
			switch tc.msg.(type) {
			case *UpdateGprsLocation:
				_, perr = ParseUpdateGprsLocation(data)
			case *AnyTimeInterrogationRes:
				_, perr = ParseAnyTimeInterrogationRes(data)
			}
			if perr != nil {
				t.Errorf("Parse of the marshalled value: %v", perr)
			}
		})
	}
}

// The SIZE of each BIT STRING stays the codec's check.
func TestMarshalBitStringSize(t *testing.T) {
	in := &UpdateGprsLocation{
		IMSI: "310260311111111", SgsnNumber: "31631000001", SGSNAddress: "192.168.31.1",
		EpsInfo: &EpsInfo{IsrInformation: HexBytes{0xFF, 0x80}, IsrInformationBits: 9},
	}
	_, err := in.Marshal()
	var ce *ber.ConstraintError
	if !errors.As(err, &ce) || ce.Path != "eps-info.isr-Information" || ce.Constraint != "SIZE (3..8)" {
		t.Errorf("9 bits: err = %v, want the isr-Information SIZE (3..8) violation", err)
	}
}

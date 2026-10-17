// semantic_bitstring_test.go
//
// BIT STRING octet counts follow ITU-T X.690 (02/2021) §8.6.2,
// checked through the public Marshal and Parse entry points.
package gsmmap

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/gomaja/go-asn1/runtime/ber"
)

// wantBitStringLengthError checks the codec's component path and invalid
// BIT STRING representation error, including SEQUENCE OF element indices.
func wantBitStringLengthError(t *testing.T, err error, field string, bits, octets int) {
	t.Helper()
	detail := fmt.Sprintf("BIT STRING length %d disagrees with %d content octets", bits, octets)
	if bits < 0 {
		detail = fmt.Sprintf("negative bit string length %d", bits)
	}
	want := fmt.Sprintf("encoding %s: %v: %s", field, ber.ErrInvalidValue, detail)
	var ce *ber.ConstraintError
	if !errors.Is(err, ber.ErrInvalidValue) || errors.As(err, &ce) || !strings.Contains(err.Error(), want) {
		t.Fatalf("err = %v, want ber.ErrInvalidValue as %q", err, want)
	}
}

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
	type bitLength struct {
		field        string
		bits, octets int
	}
	for _, tc := range []struct {
		name             string
		msg              marshaler
		length           *bitLength // the length error Marshal returns
		path, constraint string     // the *ber.ConstraintError Marshal returns
	}{
		{"isr-Information 3 bits in 1 octet", gprs(&EpsInfo{IsrInformation: HexBytes{0xE0}, IsrInformationBits: 3}, nil), nil, "", ""},
		{"isr-Information 8 bits in 1 octet", gprs(&EpsInfo{IsrInformation: HexBytes{0xE0}, IsrInformationBits: 8}, nil), nil, "", ""},
		{"isr-Information 3 bits in 2 octets", gprs(&EpsInfo{IsrInformation: HexBytes{0xE0, 0x00}, IsrInformationBits: 3}, nil), &bitLength{"eps-info.isr-Information", 3, 2}, "", ""},
		{"isr-Information 8 bits in 2 octets", gprs(&EpsInfo{IsrInformation: HexBytes{0xFF, 0x00}, IsrInformationBits: 8}, nil), &bitLength{"eps-info.isr-Information", 8, 2}, "", ""},
		{"isr-Information octets without bits", gprs(&EpsInfo{IsrInformation: HexBytes{0xE0}}, nil), &bitLength{"eps-info.isr-Information", 0, 1}, "", ""},
		// ITU-T X.690 (02/2021) §8.6.2.2: malformed octet counts fail
		// before SIZE, for a CHOICE alternative and a SEQUENCE component.
		{"isr-Information 9 bits in 1 octet", gprs(&EpsInfo{IsrInformation: HexBytes{0xE0}, IsrInformationBits: 9}, nil), &bitLength{"eps-info.isr-Information", 9, 1}, "", ""},
		{"supportedFeatures named bit 0", gprs(nil, &SGSNCapability{SupportedFeatures: &SupportedFeatures{OdbAllApn: true}}), nil, "", ""},
		{"supportedFeatures named bit 39", gprs(nil, &SGSNCapability{SupportedFeatures: &SupportedFeatures{NrAsSecondaryRAT: true}}), nil, "", ""},
		{"ext-SupportedFeatures named bit", gprs(nil, &SGSNCapability{ExtSupportedFeatures: &ExtSupportedFeatures{UnlicensedSpectrumAsSecondaryRAT: true}}), nil, "", ""},
		{"ext-SupportedFeatures unnamed bit 39", gprs(nil, &SGSNCapability{ExtSupportedFeatures: &ExtSupportedFeatures{UnknownBits: HexBytes{0, 0, 0, 0, 1}}}), nil, "", ""},
		{"isr-Information 9 bits in 2 octets", gprs(&EpsInfo{IsrInformation: HexBytes{0xFF, 0x80}, IsrInformationBits: 9}, nil), nil, "eps-info.isr-Information", "SIZE (3..8)"},
		{"csg-Id 32 bits in 4 octets", csg(HexBytes{0x11, 0x22, 0x33, 0x40}, 32), nil, "subscriberInfo.locationInformation.userCSGInformation.csg-Id", "SIZE (27)"},
		{"csg-Id 27 bits in 4 octets", csg(HexBytes{0x11, 0x22, 0x33, 0x40}, 27), nil, "", ""},
		{"csg-Id 27 bits in 10 octets", csg(make(HexBytes, 10), 27), &bitLength{"subscriberInfo.locationInformation.userCSGInformation.csg-Id", 27, 10}, "", ""},
		{"csg-Id 27 bits in 3 octets", csg(HexBytes{0x11, 0x22, 0x33}, 27), &bitLength{"subscriberInfo.locationInformation.userCSGInformation.csg-Id", 27, 3}, "", ""},
		{"csg-Id octets without bits", csg(HexBytes{0x11, 0x22, 0x33, 0x40}, 0), &bitLength{"subscriberInfo.locationInformation.userCSGInformation.csg-Id", 0, 4}, "", ""},
		{"csg-Id negative bits", csg(HexBytes{0}, -1), &bitLength{"subscriberInfo.locationInformation.userCSGInformation.csg-Id", -1, 1}, "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := tc.msg.Marshal()
			switch {
			case tc.path != "":
				wantConstraintError(t, err, tc.path, tc.constraint)
			case tc.length != nil:
				wantBitStringLengthError(t, err, tc.length.field, tc.length.bits, tc.length.octets)
			case err != nil:
				t.Fatalf("Marshal: %v", err)
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

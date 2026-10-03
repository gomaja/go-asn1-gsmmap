// semantic_bitstring_test.go
//
// The octet count of a BIT STRING against its bit length (go-asn1#80),
// checked through the public Marshal and Parse entry points.
package gsmmap

import (
	"errors"
	"testing"

	"github.com/gomaja/go-asn1/runtime/ber"
)

// X.690 (02/2021) §8.6.2: the octets of a BIT STRING hold its bits, the
// unused bits of the last octet counted by the initial octet, so a value of
// n bits has exactly (n+7)/8 octets. Marshal must not emit more octets than
// the bit length accounts for: go-asn1 encodes every octet given and the
// receiver then reads a longer bit length
// (https://github.com/gomaja/go-asn1/issues/80).
func TestMarshalBitStringOctetsMatchBitLength(t *testing.T) {
	gprs := func(e *EpsInfo, c *SGSNCapability) *UpdateGprsLocation {
		return &UpdateGprsLocation{
			IMSI: "310260311111111", SGSNNumber: "31631000001", SGSNAddress: "192.168.31.1",
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
		name string
		msg  marshaler
		want error // nil: Marshal and Parse succeed
	}{
		{"isr-Information 3 bits in 1 octet", gprs(&EpsInfo{IsrInformation: HexBytes{0xE0}, IsrInformationBits: 3}, nil), nil},
		{"isr-Information 8 bits in 1 octet", gprs(&EpsInfo{IsrInformation: HexBytes{0xFF}, IsrInformationBits: 8}, nil), nil},
		{"isr-Information 3 bits in 2 octets", gprs(&EpsInfo{IsrInformation: HexBytes{0xE0, 0x00}, IsrInformationBits: 3}, nil), ErrBitStringOctetsMismatch},
		{"isr-Information 8 bits in 2 octets", gprs(&EpsInfo{IsrInformation: HexBytes{0xFF, 0x00}, IsrInformationBits: 8}, nil), ErrBitStringOctetsMismatch},
		{"isr-Information octets without bits", gprs(&EpsInfo{IsrInformation: HexBytes{0xE0}}, nil), ErrBitStringOctetsMismatch},
		{"supportedFeatures 26 bits in 4 octets", gprs(nil, &SGSNCapability{SupportedFeatures: HexBytes{0xA0, 0, 0, 0}, SupportedFeaturesBits: 26}), nil},
		{"supportedFeatures 40 bits in 5 octets", gprs(nil, &SGSNCapability{SupportedFeatures: HexBytes{0xA0, 0, 0, 0, 1}, SupportedFeaturesBits: 40}), nil},
		{"supportedFeatures 26 bits in 5 octets", gprs(nil, &SGSNCapability{SupportedFeatures: HexBytes{0xA0, 0, 0, 0, 0}, SupportedFeaturesBits: 26}), ErrBitStringOctetsMismatch},
		{"supportedFeatures octets without bits", gprs(nil, &SGSNCapability{SupportedFeatures: HexBytes{0xA0, 0, 0, 0}}), ErrBitStringOctetsMismatch},
		{"ext-SupportedFeatures 2 bits in 1 octet", gprs(nil, &SGSNCapability{ExtSupportedFeatures: HexBytes{0x80}, ExtSupportedFeaturesBits: 2}), nil},
		{"ext-SupportedFeatures 2 bits in 2 octets", gprs(nil, &SGSNCapability{ExtSupportedFeatures: HexBytes{0x80, 0x00}, ExtSupportedFeaturesBits: 2}), ErrBitStringOctetsMismatch},
		{"ext-SupportedFeatures octets without bits", gprs(nil, &SGSNCapability{ExtSupportedFeatures: HexBytes{0x80}}), ErrBitStringOctetsMismatch},
		{"csg-Id 27 bits in 4 octets", csg(HexBytes{0x11, 0x22, 0x33, 0x40}, 27), nil},
		{"csg-Id 27 bits in 10 octets", csg(make(HexBytes, 10), 27), ErrBitStringOctetsMismatch},
		{"csg-Id 27 bits in 3 octets", csg(HexBytes{0x11, 0x22, 0x33}, 27), ErrBitStringOctetsMismatch},
		{"csg-Id octets without bits", csg(HexBytes{0x11, 0x22, 0x33, 0x40}, 0), ErrBitStringOctetsMismatch},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := tc.msg.Marshal()
			if !errors.Is(err, tc.want) {
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
		IMSI: "310260311111111", SGSNNumber: "31631000001", SGSNAddress: "192.168.31.1",
		EpsInfo: &EpsInfo{IsrInformation: HexBytes{0xFF, 0x80}, IsrInformationBits: 9},
	}
	_, err := in.Marshal()
	var ce *ber.ConstraintError
	if !errors.As(err, &ce) || ce.Path != "isr-Information" || ce.Constraint != "SIZE (3..8)" {
		t.Errorf("9 bits: err = %v, want the isr-Information SIZE (3..8) violation", err)
	}
}

package gsmmap

import (
	"errors"
	"testing"

	"github.com/gomaja/go-asn1/runtime/ber"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// TS 29.002 V19.1.0 §17.7.1: Ext-SupportedFeatures SIZE (1..40).
func TestMutationBitStringsEmptyNamedFeatures(t *testing.T) {
	for _, tc := range []struct {
		name     string
		features ExtSupportedFeatures
		bits     int
		bytes    HexBytes
	}{
		{"unset", ExtSupportedFeatures{}, 1, HexBytes{0}},
		{"explicit minimum", ExtSupportedFeatures{BitLength: 1}, 1, HexBytes{0}},
		{"named bit", ExtSupportedFeatures{UnlicensedSpectrumAsSecondaryRAT: true}, 1, HexBytes{0x80}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := &UpdateGprsLocation{IMSI: "001010123456789", SgsnNumber: "1234", SGSNAddress: "192.0.2.1",
				SGSNCapability: &SGSNCapability{ExtSupportedFeatures: &tc.features}}
			data, err := in.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			var wire gsm_map.UpdateGprsLocationArg
			if err := wire.UnmarshalBER(data); err != nil {
				t.Fatal(err)
			}
			if wire.SgsnCapability == nil || wire.SgsnCapability.ExtSupportedFeatures == nil {
				t.Fatal("missing ext-SupportedFeatures")
			}
			got := wire.SgsnCapability.ExtSupportedFeatures
			wantEqual(t, "bit length", tc.bits, got.BitLength)
			wantEqual(t, "octets", []byte(tc.bytes), got.Bytes)
		})
	}
}

// X.690 (02/2021) §8.6.2.2: three bits require one content octet.
func TestMutationBitStringsMissingOctets(t *testing.T) {
	in := &UpdateGprsLocation{IMSI: "001010123456789", SgsnNumber: "1234", SGSNAddress: "192.0.2.1",
		EpsInfo: &EpsInfo{IsrInformationBits: 3}}
	data, err := in.Marshal()
	if !errors.Is(err, ber.ErrInvalidValue) {
		t.Fatalf("Marshal = %x, %v; want invalid BIT STRING storage", data, err)
	}
	if data != nil {
		t.Fatalf("Marshal returned %x with an error", data)
	}
}

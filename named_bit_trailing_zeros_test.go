package gsmmap

import (
	"testing"

	"github.com/gomaja/go-asn1/runtime"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// ITU-T X.680 (02/2021) §22.7 permits removing trailing zero bits to
// satisfy the SIZE of a BIT STRING with named bits.
func TestNamedBitStringZeroPaddingBeyondSize(t *testing.T) {
	t.Run("SupportedFeatures padded to 48 bits", func(t *testing.T) {
		// odb-all-apn (bit 0) set; bits 1..47 are 0, SIZE (26..40)
		// (3GPP TS 29.002 V19.1.0 §17.7.1).
		arg := func(features runtime.BitString) *gsm_map.UpdateGprsLocationArg {
			return &gsm_map.UpdateGprsLocationArg{
				Imsi:           gsm_map.IMSI{0x00, 0x01, 0x01, 0x21, 0x43, 0x65, 0x87, 0xf9},
				SgsnNumber:     gsm_map.ISDNAddressString{0x91, 0x21, 0x43},
				SgsnAddress:    gsm_map.GSNAddress{0x04, 192, 0, 2, 1},
				SgsnCapability: &gsm_map.SGSNCapability{SupportedFeatures: &features},
			}
		}
		if _, err := ParseUpdateGprsLocation(tolerantBER(t, arg(runtime.BitString{Bytes: []byte{0x80, 0, 0, 0, 0}, BitLength: 40}))); err != nil {
			t.Fatalf("40 bits: %v", err)
		}
		data := tolerantBER(t, arg(runtime.BitString{Bytes: []byte{0x80, 0, 0, 0, 0, 0}, BitLength: 48}))
		var decoded gsm_map.UpdateGprsLocationArg
		if err := decoded.UnmarshalBER(data); err != nil {
			t.Fatal(err)
		}
		wantEqual(t, "decoded supportedFeatures", &runtime.BitString{Bytes: []byte{0x80, 0, 0, 0}, BitLength: 26}, decoded.SgsnCapability.SupportedFeatures)
		got, err := ParseUpdateGprsLocation(data)
		if err != nil {
			t.Fatal(err)
		}
		wantEqual(t, "SupportedFeatures", &SupportedFeatures{OdbAllApn: true}, got.SGSNCapability.SupportedFeatures)
	})
	t.Run("SupportedCamelPhases padded to 17 bits", func(t *testing.T) {
		// phase1 (bit 0) set; bits 1..16 are 0, SIZE (1..16).
		arg := func(phases runtime.BitString) *gsm_map.UpdateLocationArg {
			return &gsm_map.UpdateLocationArg{
				Imsi:          gsm_map.IMSI{0x00, 0x01, 0x01, 0x21, 0x43, 0x65, 0x87, 0xf9},
				MscNumber:     gsm_map.ISDNAddressString{0x91, 0x21, 0x43},
				VlrNumber:     gsm_map.ISDNAddressString{0x91, 0x21, 0x43},
				VlrCapability: &gsm_map.VLRCapability{SupportedCamelPhases: &phases},
			}
		}
		if _, err := ParseUpdateLocation(tolerantBER(t, arg(runtime.BitString{Bytes: []byte{0x80, 0}, BitLength: 16}))); err != nil {
			t.Fatalf("16 bits: %v", err)
		}
		data := tolerantBER(t, arg(runtime.BitString{Bytes: []byte{0x80, 0, 0}, BitLength: 17}))
		var decoded gsm_map.UpdateLocationArg
		if err := decoded.UnmarshalBER(data); err != nil {
			t.Fatal(err)
		}
		wantEqual(t, "decoded supportedCamelPhases", &runtime.BitString{Bytes: []byte{0x80}, BitLength: 1}, decoded.VlrCapability.SupportedCamelPhases)
		got, err := ParseUpdateLocation(data)
		if err != nil {
			t.Fatal(err)
		}
		wantEqual(t, "SupportedCamelPhases", &SupportedCamelPhases{Phase1: true}, got.VlrCapability.SupportedCamelPhases)
	})
}

// ITU-T X.680 (02/2021) §22.7 lets encoding rules add trailing 0 bits to a
// BIT STRING with named bits. go-asn1 pads a received value shorter than the
// SIZE lower bound, so Parse accepts it and Marshal sends it at that bound.
func TestNamedBitStringPaddedToSizeLowerBound(t *testing.T) {
	// odb-all-apn (bit 0) set in 18 bits; supportedFeatures has SIZE (26..40)
	// (3GPP TS 29.002 V19.1.0 §17.7.1).
	w := &gsm_map.UpdateGprsLocationArg{
		Imsi:           gsm_map.IMSI{0x00, 0x01, 0x01, 0x21, 0x43, 0x65, 0x87, 0xf9},
		SgsnNumber:     gsm_map.ISDNAddressString{0x91, 0x21, 0x43},
		SgsnAddress:    gsm_map.GSNAddress{0x04, 192, 0, 2, 1},
		SgsnCapability: &gsm_map.SGSNCapability{SupportedFeatures: &runtime.BitString{Bytes: []byte{0x80, 0, 0}, BitLength: 18}},
	}
	got, err := ParseUpdateGprsLocation(tolerantBER(t, w))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if f := got.SGSNCapability.SupportedFeatures; f == nil || !f.OdbAllApn {
		t.Fatalf("SupportedFeatures = %+v, want odb-all-apn", f)
	}
	data, err := got.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var sent gsm_map.UpdateGprsLocationArg
	if err := sent.UnmarshalBER(data); err != nil {
		t.Fatalf("decoding the marshalled value: %v", err)
	}
	if bits := sent.SgsnCapability.SupportedFeatures.BitLength; bits != 26 {
		t.Errorf("Marshal sends %d bits, want 26", bits)
	}
}

package gsmmap

import (
	"bytes"
	"errors"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime/ber"
)

// ITU-T X.680 (02/2021) §22.7 permits removing trailing zero bits to
// satisfy the SIZE of a BIT STRING with named bits.
func TestNamedBitStringZeroPaddingBeyondSize(t *testing.T) {
	t.Run("SupportedFeatures padded to 48 bits", func(t *testing.T) {
		// odb-all-apn (bit 0) set; bits 1..47 are 0, SIZE (26..40)
		// (3GPP TS 29.002 V19.1.0 §17.7.1).
		arg := func(features runtime.BitString) []byte {
			// Raw received BER keeps the supplied length; generated encoders normalize it.
			fields := mustHex(t, "040800010121436587f90403912143040504c0000201")
			content := append([]byte{byte((8 - features.BitLength%8) % 8)}, features.Bytes...)
			return ussdTLV(0x30, append(fields, ussdTLV(0xa0, ussdTLV(0x89, content))...))
		}
		if _, err := ParseUpdateGprsLocation(arg(runtime.BitString{Bytes: []byte{0x80, 0, 0, 0, 0}, BitLength: 40})); err != nil {
			t.Fatalf("40 bits: %v", err)
		}
		data := arg(runtime.BitString{Bytes: []byte{0x80, 0, 0, 0, 0, 0}, BitLength: 48})
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
		arg := func(phases runtime.BitString) []byte {
			fields := mustHex(t, "040800010121436587f981039121430403912143")
			content := append([]byte{byte((8 - phases.BitLength%8) % 8)}, phases.Bytes...)
			return ussdTLV(0x30, append(fields, ussdTLV(0xa6, ussdTLV(0x80, content))...))
		}
		if _, err := ParseUpdateLocation(arg(runtime.BitString{Bytes: []byte{0x80, 0}, BitLength: 16})); err != nil {
			t.Fatalf("16 bits: %v", err)
		}
		data := arg(runtime.BitString{Bytes: []byte{0x80, 0, 0}, BitLength: 17})
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
	got, err := ParseUpdateGprsLocation(mustHex(t, "301e040800010121436587f90403912143040504c0000201a006890406800000"))
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

// ITU-T X.680 (02/2021) §22.7: go-asn1 chooses the shortest permitted length.
func TestMarshalNamedBitStringShortestRAT(t *testing.T) {
	for _, tc := range []struct {
		name string
		rat  SupportedRATTypes
		wire string
	}{
		{"all false", SupportedRATTypes{}, "88020600"},
		{"utran", SupportedRATTypes{UTRAN: true}, "88020680"},
		{"gan", SupportedRATTypes{GAN: true}, "88020520"},
		{"nb-iot", SupportedRATTypes{NBIoT: true}, "88020204"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := senderMobilityGPRS()
			in.SGSNCapability = &SGSNCapability{SupportedRATTypesIndicator: &tc.rat}
			got, err := in.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Contains(got, mustHex(t, tc.wire)) {
				t.Fatalf("Marshal = %x, want RAT %s", got, tc.wire)
			}
		})
	}
}

// https://github.com/gomaja/go-asn1/issues/183: a named-bit SIZE overflow
// should be a *ber.ValidationError (ITU-T X.680 (02/2021) §51.5).
// When this fails, go-asn1 has fixed #183: expect *ber.ValidationError
// and update the named-bit SIZE assertions and the README.
func TestGoASN1Issue183NamedBitSizeErrorClass(t *testing.T) {
	value := gsm_map.SGSNCapability{SupportedRATTypesIndicator: &runtime.BitString{Bytes: []byte{0x00, 0x80}, BitLength: 9}}
	_, err := value.MarshalBER()
	var validation *ber.ValidationError
	if !errors.Is(err, ber.ErrInvalidValue) || errors.As(err, &validation) {
		t.Fatalf("MarshalBER: %v, want ErrInvalidValue without ValidationError (#183)", err)
	}
}

func TestMarshalNamedBitStringShortestFeaturesAndISR(t *testing.T) {
	for _, tc := range []struct {
		name     string
		features *SupportedFeatures
		isr      *IsrInformation
		field    string
	}{
		{name: "features all false", features: &SupportedFeatures{}, field: "89050600000000"},
		{name: "features bit 0", features: &SupportedFeatures{OdbAllApn: true}, field: "89050680000000"},
		{name: "features bit 39", features: &SupportedFeatures{NrAsSecondaryRAT: true}, field: "8906000000000001"},
		{name: "ISR all false", isr: &IsrInformation{}, field: "a50481020500"},
		{name: "ISR update", isr: &IsrInformation{UpdateLocation: true}, field: "a50481020580"},
		{name: "ISR cancel", isr: &IsrInformation{CancelSGSN: true}, field: "a50481020540"},
		{name: "ISR initial attach", isr: &IsrInformation{InitialAttachIndicator: true}, field: "a50481020520"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := senderMobilityGPRS()
			if tc.features != nil {
				in.SGSNCapability = &SGSNCapability{SupportedFeatures: tc.features}
			}
			if tc.isr != nil {
				in.EpsInfo = &EpsInfo{IsrInformation: tc.isr}
			}
			out, err := in.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			// SIZE minima are 26 and 3, giving 6 and 5 unused bits respectively.
			if !bytes.Contains(out, mustHex(t, tc.field)) {
				t.Fatalf("Marshal = %x, want field %s", out, tc.field)
			}
			if err := checkDefiniteLengths(out); err != nil {
				t.Fatal(err)
			}
		})
	}
}

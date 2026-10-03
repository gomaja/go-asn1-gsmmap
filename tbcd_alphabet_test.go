// tbcd_alphabet_test.go
//
// End-to-end tests of the TBCD-STRING alphabet of TS 29.002 V19.1.0 §17.7.8
// (0-9 * # a b c, 1111 filler) through the Marshal / Parse entry points.

package gsmmap

import (
	"bytes"
	"errors"
	"testing"

	"github.com/gomaja/go-asn1-gsmmap/address"
	"github.com/gomaja/go-asn1-gsmmap/tbcd"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

func TestSriSmMSISDNStarHashRoundTrip(t *testing.T) {
	for _, msisdn := range []string{"*#12", "*123#", "#31#491711", "12abc", "**21*4917#", "0123456789*#abc"} {
		in := &SriSm{
			MSISDN:               msisdn,
			MSISDNNature:         address.NatureUnknown,
			MSISDNPlan:           address.PlanISDN,
			ServiceCentreAddress: "491234",
			SCANature:            address.NatureInternational,
			SCAPlan:              address.PlanISDN,
		}
		data, err := in.Marshal()
		if err != nil {
			t.Fatalf("%q: Marshal: %v", msisdn, err)
		}
		out, err := ParseSriSm(data)
		if err != nil {
			t.Fatalf("%q: ParseSriSm: %v", msisdn, err)
		}
		if out.MSISDN != msisdn {
			t.Errorf("MSISDN round trip = %q, want %q", out.MSISDN, msisdn)
		}
		if out.ServiceCentreAddress != "491234" || out.MSISDNPlan != address.PlanISDN {
			t.Errorf("other fields changed: %+v", out)
		}
		again, err := out.Marshal()
		if err != nil || !bytes.Equal(again, data) {
			t.Errorf("re-marshal differs: %x vs %x (%v)", again, data, err)
		}
	}
}

// The wire octets, not just the round trip: "*#12" is AddressString octet
// 0x81 followed by TBCD ba 21 (low nibble first; 1010 = '*', 1011 = '#').
func TestSriSmMSISDNStarHashWireOctets(t *testing.T) {
	in := &SriSm{
		MSISDN: "*#12", MSISDNPlan: address.PlanISDN,
		ServiceCentreAddress: "1", SCAPlan: address.PlanISDN,
	}
	data, err := in.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0x80, 0x03, 0x81, 0xba, 0x21}
	if !bytes.Contains(data, want) {
		t.Errorf("wire %x does not contain msisdn octets %x", data, want)
	}
}

func TestUSSDArgMSISDNStarHashRoundTrip(t *testing.T) {
	in := &USSDArg{
		DataCodingScheme: USSDDataCodingSchemeGSM7,
		USSDString:       []byte{0x31},
		MSISDN:           "*#06#",
		MSISDNNature:     address.NatureInternational,
		MSISDNPlan:       address.PlanISDN,
	}
	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	out, err := ParseUSSDArg(data)
	if err != nil {
		t.Fatalf("ParseUSSDArg: %v", err)
	}
	if out.MSISDN != in.MSISDN || out.MSISDNNature != in.MSISDNNature || out.MSISDNPlan != in.MSISDNPlan {
		t.Errorf("round trip = %+v, want %+v", out, in)
	}
}

func TestMarshalRejectsNonTBCDCharacter(t *testing.T) {
	for _, bad := range []string{"12f4", "12A4", "12d", "12 4", "+1234"} {
		in := &SriSm{MSISDN: bad, ServiceCentreAddress: "1"}
		if _, err := in.Marshal(); !errors.Is(err, tbcd.ErrInvalidCharacter) {
			t.Errorf("MSISDN %q: err = %v, want tbcd.ErrInvalidCharacter", bad, err)
		}
	}
}

func sriSmArgWithAddresses(t *testing.T, msisdn []byte, imsi []byte) []byte {
	t.Helper()
	w := &gsm_map.RoutingInfoForSMArg{
		Msisdn:               msisdn,
		SmRPPRI:              false,
		ServiceCentreAddress: []byte{0x91, 0x21},
	}
	if imsi != nil {
		i := imsi
		w.Imsi = &i
	}
	data, err := w.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}
	return data
}

// A filler nibble followed by a digit is malformed TBCD.
func TestParseMalformedTBCDFiller(t *testing.T) {
	cases := map[string][]byte{
		"filler octet then digits":  {0x91, 0x21, 0xff, 0x43},
		"low filler, high digit":    {0x91, 0x21, 0xaf},
		"digit after filler nibble": {0x91, 0xf1, 0x21},
	}
	for name, msisdn := range cases {
		if _, err := ParseSriSm(sriSmArgWithAddresses(t, msisdn, nil)); !errors.Is(err, tbcd.ErrMisplacedFiller) {
			t.Errorf("%s: err = %v, want tbcd.ErrMisplacedFiller", name, err)
		}
	}
	// Trailing filler octets are tolerated and not part of the digits.
	got, err := ParseSriSm(sriSmArgWithAddresses(t, []byte{0x91, 0x21, 0xf3, 0xff}, nil))
	if err != nil || got == nil || got.MSISDN != "123" {
		t.Errorf("trailing filler octet: %+v, err %v", got, err)
	}
}

// IMSI, IMEI and IMEISV are digit strings (TS 23.003): the extra TBCD symbols
// are valid TBCD but not valid identities.
func TestIdentityDigitsOnly(t *testing.T) {
	in := &SriSm{MSISDN: "1234", ServiceCentreAddress: "1", IMSI: "2040*0123"}
	if _, err := in.Marshal(); !errors.Is(err, ErrIdentityNotDigits) {
		t.Errorf("Marshal IMSI with '*': err = %v, want ErrIdentityNotDigits", err)
	}
	in.IMSI = "204080012345678"
	if _, err := in.Marshal(); err != nil {
		t.Errorf("Marshal valid IMSI: %v", err)
	}
	// Wire IMSI 2 0 4 0 # 0 ...: nibble 1011 is '#'.
	wire := sriSmArgWithAddresses(t, []byte{0x91, 0x21}, []byte{0x02, 0xb0, 0x21})
	if _, err := ParseSriSm(wire); !errors.Is(err, ErrIdentityNotDigits) {
		t.Errorf("Parse IMSI with '#': err = %v, want ErrIdentityNotDigits", err)
	}
	// A wire IMSI that is not valid TBCD at all fails with the TBCD sentinel.
	wire = sriSmArgWithAddresses(t, []byte{0x91, 0x21}, []byte{0x02, 0xff, 0x21})
	if _, err := ParseSriSm(wire); !errors.Is(err, tbcd.ErrMisplacedFiller) {
		t.Errorf("Parse IMSI with mid filler: err = %v, want tbcd.ErrMisplacedFiller", err)
	}
}

func TestDecodeAddressFieldZeroOctets(t *testing.T) {
	for _, in := range [][]byte{nil, {}} {
		if _, _, _, err := decodeAddressField(in); !errors.Is(err, ErrAddressStringEmpty) {
			t.Errorf("decodeAddressField(%#v) err = %v, want ErrAddressStringEmpty", in, err)
		}
	}
}

// semantic_isd_identity_test.go
//
// InsertSubscriberDataArg.IMSI and CorrelationID HLR-Id carry digits like
// every other IMSI in the package, through the shared identity rules: an
// IMSI has 6 to 15 decimal digits (3GPP TS 23.003 V20.1.0 §2.2, §2.3), and
// "HLR-Id ::= IMSI -- leading digits of IMSI, i.e. (MCC, MNC, leading digits
// of MSIN)" (3GPP TS 29.002 V19.1.0 §17.7.8) follows the same rule on
// Marshal. Parse also preserves five-digit MCC+MNC values from peers.
package gsmmap

import (
	"errors"
	"testing"

	identity "github.com/gomaja/go-3gpp-identity"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

func TestInsertSubscriberDataIMSIDigits(t *testing.T) {
	for _, tc := range []struct {
		imsi string
		want error
	}{
		{"123456", nil},
		{"204080012345678", nil},
		{"12345", identity.ErrIMSILength},
		{"1234567890123456", identity.ErrIMSILength},
		{"20408001234567a", identity.ErrNotDigits},
		{"20408001234567*", identity.ErrNotDigits},
	} {
		t.Run(tc.imsi, func(t *testing.T) {
			in := &InsertSubscriberDataArg{IMSI: tc.imsi}
			data, err := in.Marshal()
			if !errors.Is(err, tc.want) {
				t.Fatalf("Marshal: err = %v, want %v", err, tc.want)
			}
			if err != nil {
				return
			}
			got, err := ParseInsertSubscriberData(data)
			if err != nil {
				t.Fatalf("ParseInsertSubscriberData: %v", err)
			}
			if got.IMSI != tc.imsi {
				t.Errorf("IMSI = %q, want %q", got.IMSI, tc.imsi)
			}
		})
	}
}

func TestParseInsertSubscriberDataIMSIDigits(t *testing.T) {
	for _, tc := range []struct {
		name string
		imsi gsm_map.IMSI
		want error
		digs string
	}{
		{"three digits in SIZE (3)", gsm_map.IMSI{0x00, 0xF1, 0xFF}, identity.ErrIMSILength, ""},
		{"all filler", gsm_map.IMSI{0xFF, 0xFF, 0xFF}, identity.ErrEmpty, ""},
		{"TBCD a", gsm_map.IMSI{0x21, 0x43, 0xC5}, identity.ErrNotDigits, ""},
		{"six digits", gsm_map.IMSI{0x21, 0x43, 0x65}, nil, "123456"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			imsi := tc.imsi
			data, err := (&gsm_map.InsertSubscriberDataArg{Imsi: &imsi}).MarshalBER()
			if err != nil {
				t.Fatalf("MarshalBER: %v", err)
			}
			got, err := ParseInsertSubscriberData(data)
			if !errors.Is(err, tc.want) {
				t.Fatalf("ParseInsertSubscriberData: err = %v, want %v", err, tc.want)
			}
			if err == nil && got.IMSI != tc.digs {
				t.Errorf("IMSI = %q, want %q", got.IMSI, tc.digs)
			}
		})
	}
}

func TestCorrelationIDHlrIDDigits(t *testing.T) {
	sriSm := func(hlr string) *SriSm {
		return &SriSm{
			MSISDN: "31612345678", ServiceCentreAddress: "31600000001", SmRpPri: true,
			CorrelationID: &SriSmCorrelationID{HlrID: hlr, SipUriB: HexBytes("sip:b@example.com")},
		}
	}
	for _, tc := range []struct {
		hlr  string
		want error
	}{
		{"204080", nil},
		{"204080012345678", nil},
		{"20408", identity.ErrIMSILength},
		{"2040800123456789", identity.ErrIMSILength},
		{"20408a", identity.ErrNotDigits},
	} {
		t.Run("Marshal "+tc.hlr, func(t *testing.T) {
			data, err := sriSm(tc.hlr).Marshal()
			if !errors.Is(err, tc.want) {
				t.Fatalf("Marshal: err = %v, want %v", err, tc.want)
			}
			if err != nil {
				return
			}
			got, err := ParseSriSm(data)
			if err != nil {
				t.Fatalf("ParseSriSm: %v", err)
			}
			if got.CorrelationID.HlrID != tc.hlr {
				t.Errorf("HlrID = %q, want %q", got.CorrelationID.HlrID, tc.hlr)
			}
		})
	}
	for _, tc := range []struct {
		name       string
		hlr        gsm_map.HLRId
		want       error
		digits     string
		marshalErr error
	}{
		{"five digits", gsm_map.HLRId{0x02, 0x04, 0xF8}, nil, "20408", identity.ErrIMSILength},
		{"octets AA BB", gsm_map.HLRId{0xAA, 0xBB, 0xAA}, identity.ErrNotDigits, "", nil},
		{"six digits", gsm_map.HLRId{0x02, 0x04, 0x08}, nil, "204080", nil},
	} {
		t.Run("Parse "+tc.name, func(t *testing.T) {
			w, err := convertSriSmToArg(sriSm("204080"))
			if err != nil {
				t.Fatalf("convertSriSmToArg: %v", err)
			}
			hlr := tc.hlr
			w.CorrelationID.HlrId = &hlr
			data, err := w.MarshalBER()
			if err != nil {
				t.Fatalf("MarshalBER: %v", err)
			}
			got, err := ParseSriSm(data)
			if !errors.Is(err, tc.want) {
				t.Fatalf("ParseSriSm: err = %v, want %v", err, tc.want)
			}
			if err != nil {
				return
			}
			if got.CorrelationID.HlrID != tc.digits {
				t.Errorf("HlrID = %q, want %q", got.CorrelationID.HlrID, tc.digits)
			}
			if _, err := got.Marshal(); !errors.Is(err, tc.marshalErr) {
				t.Errorf("Marshal: err = %v, want %v", err, tc.marshalErr)
			}
		})
	}
}

// convert_slr_res_test.go
//
// Tests for the top-level SubscriberLocationReportRes converter and its
// Marshal()/ParseSubscriberLocationReportRes() entry points (opCode 86).
// Round-trip (struct→wire→struct), BER round-trip (Marshal→Parse), and
// targeted negative cases.
package gsmmap

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

func TestSLRResRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   *SubscriberLocationReportRes
	}{
		{"empty (all absent)", &SubscriberLocationReportRes{}},
		{"na-ESRK only", &SubscriberLocationReportRes{
			NaESRK:       "8005551212",
			NaESRKNature: 0x10,
			NaESRKPlan:   0x01,
		}},
		{"na-ESRD only", &SubscriberLocationReportRes{
			NaESRD:       "8005556789",
			NaESRDNature: 0x10,
			NaESRDPlan:   0x01,
		}},
		{"both na-ESRK and na-ESRD (independent optionals)", &SubscriberLocationReportRes{
			NaESRK:       "8005551212",
			NaESRKNature: 0x10,
			NaESRKPlan:   0x01,
			NaESRD:       "8005556789",
			NaESRDNature: 0x10,
			NaESRDPlan:   0x01,
		}},
		{"h-gmlc + NULL flag + lcs-ref", &SubscriberLocationReportRes{
			HGmlcAddress:              "192.0.2.10",
			MoLrShortCircuitIndicator: true,
			LcsReferenceNumber:        HexBytes{0x2a},
		}},
		{"reporting PLMN list", &SubscriberLocationReportRes{
			ReportingPLMNList: &ReportingPLMNList{
				PlmnListPrioritized: true,
				PlmnList: PLMNList{
					{PlmnId: HexBytes{0x32, 0xf4, 0x10}},
					{PlmnId: HexBytes{0x62, 0xf2, 0x20}},
				},
			},
		}},
		{"fully populated", &SubscriberLocationReportRes{
			NaESRK:                    "8005551212",
			NaESRKNature:              0x10,
			NaESRKPlan:                0x01,
			NaESRD:                    "8005556789",
			NaESRDNature:              0x10,
			NaESRDPlan:                0x01,
			HGmlcAddress:              "192.0.2.10",
			MoLrShortCircuitIndicator: true,
			ReportingPLMNList: &ReportingPLMNList{
				PlmnList: PLMNList{{PlmnId: HexBytes{0x32, 0xf4, 0x10}}},
			},
			LcsReferenceNumber: HexBytes{0x2a},
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertSubscriberLocationReportResToWire(tc.in)
			if err != nil {
				t.Fatalf("toWire: %v", err)
			}
			got, err := convertWireToSubscriberLocationReportRes(wire)
			if err != nil {
				t.Fatalf("toStruct: %v", err)
			}
			if !reflect.DeepEqual(tc.in, got) {
				t.Errorf("round-trip mismatch:\n in=%+v\ngot=%+v", tc.in, got)
			}
		})
	}
}

// TestSLRResBERRoundTrip exercises the public Marshal()/Parse() path
// through real BER encoding, including the all-absent (empty) response.
func TestSLRResBERRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   *SubscriberLocationReportRes
	}{
		{"empty", &SubscriberLocationReportRes{}},
		{"populated", &SubscriberLocationReportRes{
			NaESRK:                    "8005551212",
			NaESRKNature:              0x10,
			NaESRKPlan:                0x01,
			HGmlcAddress:              "192.0.2.10",
			MoLrShortCircuitIndicator: true,
			LcsReferenceNumber:        HexBytes{0x2a},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := tc.in.Marshal()
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			got, err := ParseSubscriberLocationReportRes(data)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if !reflect.DeepEqual(tc.in, got) {
				t.Errorf("BER round-trip mismatch:\n in=%+v\ngot=%+v", tc.in, got)
			}
		})
	}
}

func TestSLRResEncodeNegative(t *testing.T) {
	cases := []struct {
		name string
		in   *SubscriberLocationReportRes
		want error
	}{
		{"nil res", nil, ErrSLRResNil},
		{"LcsReferenceNumber wrong size", &SubscriberLocationReportRes{
			LcsReferenceNumber: HexBytes{0x01, 0x02},
		}, ErrLCSReferenceNumberInvalidSize},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := convertSubscriberLocationReportResToWire(tc.in)
			if !errors.Is(err, tc.want) {
				t.Errorf("want errors.Is(_, %v), got %v", tc.want, err)
			}
		})
	}
}

// TestSLRResDecodeNegative covers decode-side validation that the
// encoder cannot reach: present-but-empty na-ESRK/na-ESRD on the wire
// must be rejected so the string-based API round-trips faithfully.
func TestSLRResDecodeNegative(t *testing.T) {
	// AddressString header (extension=1, nature=international,
	// plan=ISDN) with no TBCD digits — present but decodes to "".
	emptyAddr := gsm_map.ISDNAddressString{0x91}

	t.Run("nil wire", func(t *testing.T) {
		_, err := convertWireToSubscriberLocationReportRes(nil)
		if !errors.Is(err, ErrSLRResNil) {
			t.Errorf("want ErrSLRResNil, got %v", err)
		}
	})
	t.Run("na-ESRK present but empty", func(t *testing.T) {
		w := &gsm_map.SubscriberLocationReportRes{NaESRK: &emptyAddr}
		_, err := convertWireToSubscriberLocationReportRes(w)
		if !errors.Is(err, ErrSLRResNaESRKDecodedEmpty) {
			t.Errorf("want ErrSLRResNaESRKDecodedEmpty, got %v", err)
		}
	})
	t.Run("na-ESRD present but empty", func(t *testing.T) {
		w := &gsm_map.SubscriberLocationReportRes{NaESRD: &emptyAddr}
		_, err := convertWireToSubscriberLocationReportRes(w)
		if !errors.Is(err, ErrSLRResNaESRDDecodedEmpty) {
			t.Errorf("want ErrSLRResNaESRDDecodedEmpty, got %v", err)
		}
	})
	t.Run("LcsReferenceNumber wrong size", func(t *testing.T) {
		ref := gsm_map.LCSReferenceNumber{0x01, 0x02}
		w := &gsm_map.SubscriberLocationReportRes{LcsReferenceNumber: &ref}
		_, err := convertWireToSubscriberLocationReportRes(w)
		if !errors.Is(err, ErrLCSReferenceNumberInvalidSize) {
			t.Errorf("want ErrLCSReferenceNumberInvalidSize, got %v", err)
		}
	})
}

// convert_sri_lcs_test.go
//
// Tests for SendRoutingInfoForLCS (opCode 85): the shared
// SubscriberIdentity CHOICE converter, the SriLcs/SriLcsRes top-level
// converters, and their Marshal()/Parse() entry points. Round-trip,
// BER round-trip, and targeted negative cases.
package gsmmap

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-tbcd"

	"github.com/gomaja/go-asn1-gsmmap/gsn"
)

// =============================================================================
// SubscriberIdentity shared converter
// =============================================================================

func TestSubscriberIdentityRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   SubscriberIdentity
	}{
		{"IMSI", SubscriberIdentity{IMSI: "204080000000001"}},
		// MSISDN with explicit International/ISDN nature+plan. (Leaving
		// them 0 would encode as 0x10/0x01 and decode back non-zero, so
		// the round-trip is expressed with the encoded values.)
		{"MSISDN international", SubscriberIdentity{MSISDN: "31612345678", MSISDNNature: 0x10, MSISDNPlan: 0x01}},
		// Non-international MSISDN must preserve its Nature/Plan — the
		// reason MSISDNNature/MSISDNPlan were added to SubscriberIdentity.
		{"MSISDN national", SubscriberIdentity{MSISDN: "0612345678", MSISDNNature: 0x20, MSISDNPlan: 0x01}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertSubscriberIdentityToWire(tc.in)
			if err != nil {
				t.Fatalf("toWire: %v", err)
			}
			got, err := convertWireToSubscriberIdentity(wire)
			if err != nil {
				t.Fatalf("toStruct: %v", err)
			}
			if !reflect.DeepEqual(tc.in, got) {
				t.Errorf("round-trip mismatch: in=%+v got=%+v", tc.in, got)
			}
		})
	}
}

func TestSubscriberIdentityEncodeNegative(t *testing.T) {
	cases := []struct {
		name string
		in   SubscriberIdentity
		want error
	}{
		{"neither set", SubscriberIdentity{}, ErrSubscriberIdentityNoAlternative},
		{"both set", SubscriberIdentity{IMSI: "204080000000001", MSISDN: "31612345678"}, ErrSubscriberIdentityMultipleAlternatives},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := convertSubscriberIdentityToWire(tc.in)
			if !errors.Is(err, tc.want) {
				t.Errorf("want %v, got %v", tc.want, err)
			}
		})
	}
}

func TestSubscriberIdentityDecodeNegative(t *testing.T) {
	emptyMsisdn := gsm_map.ISDNAddressString{0x91} // header only, no digits

	t.Run("MSISDN present but empty", func(t *testing.T) {
		w := gsm_map.NewSubscriberIdentityMsisdn(emptyMsisdn)
		_, err := convertWireToSubscriberIdentity(w)
		if !errors.Is(err, ErrSubscriberIdentityMSISDNDecodedEmpty) {
			t.Errorf("want ErrSubscriberIdentityMSISDNDecodedEmpty, got %v", err)
		}
	})
	t.Run("unknown choice", func(t *testing.T) {
		_, err := convertWireToSubscriberIdentity(gsm_map.SubscriberIdentity{Choice: 99})
		if !errors.Is(err, ErrSubscriberIdentityUnknownAlternative) {
			t.Errorf("want ErrSubscriberIdentityUnknownAlternative, got %v", err)
		}
	})

}

// =============================================================================
// SriLcs (Arg)
// =============================================================================

func minimalSriLcs() *SriLcs {
	return &SriLcs{
		MlcNumber:       "31611111111",
		MlcNumberNature: 0x10,
		MlcNumberPlan:   0x01,
		TargetMS:        SubscriberIdentity{IMSI: "204080000000001"},
	}
}

func TestSriLcsRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   *SriLcs
	}{
		{"IMSI target", minimalSriLcs()},
		{"MSISDN target", &SriLcs{
			MlcNumber:       "31611111111",
			MlcNumberNature: 0x10,
			MlcNumberPlan:   0x01,
			TargetMS:        SubscriberIdentity{MSISDN: "31612345678", MSISDNNature: 0x10, MSISDNPlan: 0x01},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			arg, err := convertSriLcsToArg(tc.in)
			if err != nil {
				t.Fatalf("toArg: %v", err)
			}
			got, err := convertArgToSriLcs(arg)
			if err != nil {
				t.Fatalf("toStruct: %v", err)
			}
			if !reflect.DeepEqual(tc.in, got) {
				t.Errorf("round-trip mismatch:\n in=%+v\ngot=%+v", tc.in, got)
			}
		})
	}
}

func TestSriLcsBERRoundTrip(t *testing.T) {
	in := minimalSriLcs()
	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseSriLcs(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !reflect.DeepEqual(in, got) {
		t.Errorf("BER round-trip mismatch:\n in=%+v\ngot=%+v", in, got)
	}
}

func TestSriLcsEncodeNegative(t *testing.T) {
	cases := []struct {
		name string
		mut  func(a *SriLcs)
		want error
	}{
		{"nil arg", nil, ErrSriLcsNil},
		{"empty MlcNumber", func(a *SriLcs) { a.MlcNumber = "" }, ErrSriLcsMlcNumberEmpty},
		{"no target identity", func(a *SriLcs) { a.TargetMS = SubscriberIdentity{} }, ErrSubscriberIdentityNoAlternative},
		{"ambiguous target", func(a *SriLcs) { a.TargetMS = SubscriberIdentity{IMSI: "204080000000001", MSISDN: "31612345678"} }, ErrSubscriberIdentityMultipleAlternatives},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var in *SriLcs
			if tc.mut != nil {
				in = minimalSriLcs()
				tc.mut(in)
			}
			_, err := convertSriLcsToArg(in)
			if !errors.Is(err, tc.want) {
				t.Errorf("want %v, got %v", tc.want, err)
			}
		})
	}
}

// TS 29.002 V19.1.0 §17.7.13: mlcNumber and targetMS are mandatory;
// §17.7.8 defines their ISDN-AddressString/TBCD encodings.
func TestSriLcsParseNegative(t *testing.T) {
	for _, tc := range []struct {
		name, wire, field string
		want              error
	}{
		{"nil converter", "", "sriLcs", ErrSriLcsNil},
		// Target [1] has 10 content octets (IMSI TLV 2+8). MLC TLVs
		// are 4 and 3 octets respectively: outer lengths 16 and 15.
		{"MlcNumber decode", "30108002911fa10a800800010121436587f9", "SriLcs.MlcNumber", tbcd.ErrMisplacedFiller},
		{"MlcNumber empty", "300f800191a10a800800010121436587f9", "MlcNumber", ErrSriLcsMlcNumberEmpty},
		// MLC TLV 4 + target wrapper 2 + digitless MSISDN TLV 3 = 9.
		{"TargetMS", "300980029121a103810191", "SriLcs.TargetMS", ErrSubscriberIdentityMSISDNDecodedEmpty},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got *SriLcs
			var err error
			if tc.wire == "" {
				got, err = convertArgToSriLcs(nil)
			} else {
				got, err = ParseSriLcs(mustHex(t, tc.wire))
			}
			if got != nil || !errors.Is(err, tc.want) {
				t.Fatalf("Parse = %+v, %v; want nil, %v", got, err, tc.want)
			}
			if !strings.Contains(err.Error(), tc.field) {
				t.Errorf("error must name %s: %v", tc.field, err)
			}
		})
	}
}

// =============================================================================
// SriLcsRes (Res)
// =============================================================================

func minimalSriLcsRes() *SriLcsRes {
	return &SriLcsRes{
		TargetMS: SubscriberIdentity{IMSI: "204080000000001"},
		LcsLocationInfo: LCSLocationInfo{
			NetworkNodeNumber:       "31650000000",
			NetworkNodeNumberNature: 0x10,
			NetworkNodeNumberPlan:   0x01,
		},
	}
}

func TestSriLcsResRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   *SriLcsRes
	}{
		{"minimal", minimalSriLcsRes()},
		{"MSISDN target + all GMLC/PPR addresses", &SriLcsRes{
			TargetMS: SubscriberIdentity{MSISDN: "31612345678", MSISDNNature: 0x10, MSISDNPlan: 0x01},
			LcsLocationInfo: LCSLocationInfo{
				NetworkNodeNumber:       "31650000000",
				NetworkNodeNumberNature: 0x10,
				NetworkNodeNumberPlan:   0x01,
				LMSI:                    HexBytes{0x01, 0x02, 0x03, 0x04},
				GprsNodeIndicator:       true,
			},
			VGmlcAddress:           "192.0.2.1",
			HGmlcAddress:           "192.0.2.2",
			PprAddress:             "192.0.2.3",
			AdditionalVGmlcAddress: "2001:db8::1",
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := convertSriLcsResToRes(tc.in)
			if err != nil {
				t.Fatalf("toRes: %v", err)
			}
			got, err := convertResToSriLcsRes(res)
			if err != nil {
				t.Fatalf("toStruct: %v", err)
			}
			if !reflect.DeepEqual(tc.in, got) {
				t.Errorf("round-trip mismatch:\n in=%+v\ngot=%+v", tc.in, got)
			}
		})
	}
}

func TestSriLcsResBERRoundTrip(t *testing.T) {
	in := minimalSriLcsRes()
	in.HGmlcAddress = "192.0.2.10"
	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseSriLcsRes(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !reflect.DeepEqual(in, got) {
		t.Errorf("BER round-trip mismatch:\n in=%+v\ngot=%+v", in, got)
	}
}

func TestSriLcsResEncodeNegative(t *testing.T) {
	cases := []struct {
		name string
		mut  func(r *SriLcsRes)
		want error
	}{
		{"nil res", nil, ErrSriLcsResNil},
		{"no target identity", func(r *SriLcsRes) { r.TargetMS = SubscriberIdentity{} }, ErrSubscriberIdentityNoAlternative},
		{"empty LcsLocationInfo node", func(r *SriLcsRes) { r.LcsLocationInfo.NetworkNodeNumber = "" }, ErrLCSLocationInfoNetworkNodeNumberEmpty},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var in *SriLcsRes
			if tc.mut != nil {
				in = minimalSriLcsRes()
				tc.mut(in)
			}
			_, err := convertSriLcsResToRes(in)
			if !errors.Is(err, tc.want) {
				t.Errorf("want %v, got %v", tc.want, err)
			}
		})
	}
}

// TS 29.002 V19.1.0 §17.7.13: targetMS and lcsLocationInfo are mandatory;
// the four optional GSN addresses use §17.7.8 (TS 23.003 V20.1.0 §5.1).
func TestSriLcsResParseNegative(t *testing.T) {
	for _, tc := range []struct {
		name, wire, field string
		want              error
	}{
		{"nil converter", "", "sriLcsRes", ErrSriLcsResNil},
		// Digitless target TLV 5 + location TLV 6 = 11.
		{"TargetMS", "300ba003810191a10404029121", "SriLcsRes.TargetMS", ErrSubscriberIdentityMSISDNDecodedEmpty},
		// Valid IMSI target TLV 12 + location TLV 6 or 5 = 18 or 17.
		{"LCSLocationInfo decode", "3012a00a800800010121436587f9a1040402911f", "SriLcsRes.LcsLocationInfo", tbcd.ErrMisplacedFiller},
		{"LCSLocationInfo empty", "3011a00a800800010121436587f9a103040191", "SriLcsRes.LcsLocationInfo", ErrLCSLocationInfoNetworkNodeNumberEmpty},
		// Target TLV 12 + location TLV 6 + GSN TLV 7 = 25. Five GSN
		// octets pass SIZE (5..17), but header 03 declares IPv4 length 3.
		{"VGmlcAddress", "3019a00a800800010121436587f9a10404029121830503c0000201", "SriLcsRes.VGmlcAddress", gsn.ErrIPv4LengthInvalid},
		{"HGmlcAddress", "3019a00a800800010121436587f9a10404029121840503c0000201", "SriLcsRes.HGmlcAddress", gsn.ErrIPv4LengthInvalid},
		{"PprAddress", "3019a00a800800010121436587f9a10404029121850503c0000201", "SriLcsRes.PprAddress", gsn.ErrIPv4LengthInvalid},
		{"AdditionalVGmlcAddress", "3019a00a800800010121436587f9a10404029121860503c0000201", "SriLcsRes.AdditionalVGmlcAddress", gsn.ErrIPv4LengthInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got *SriLcsRes
			var err error
			if tc.wire == "" {
				got, err = convertResToSriLcsRes(nil)
			} else {
				got, err = ParseSriLcsRes(mustHex(t, tc.wire))
			}
			if got != nil || !errors.Is(err, tc.want) {
				t.Fatalf("Parse = %+v, %v; want nil, %v", got, err, tc.want)
			}
			if !strings.Contains(err.Error(), tc.field) {
				t.Errorf("error must name %s: %v", tc.field, err)
			}
		})
	}
}

// TestSriLcsResInvalidGSNAddress: gsn.Build returns a plain (un-wrapped)
// error for a malformed IP, so there is no sentinel to assert via
// errors.Is. Instead verify the converter attributes the failure to the
// right field via its wrapping context.
func TestSriLcsResInvalidGSNAddress(t *testing.T) {
	in := minimalSriLcsRes()
	in.HGmlcAddress = "not-an-ip"
	_, err := convertSriLcsResToRes(in)
	if err == nil {
		t.Fatal("want an error for malformed GSN address, got nil")
	}
	if !strings.Contains(err.Error(), "SriLcsRes.HGmlcAddress") {
		t.Errorf("error should name the offending field; got %v", err)
	}
}

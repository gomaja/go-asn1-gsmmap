// convert_sri_lcs_test.go
//
// Tests for SendRoutingInfoForLCS (opCode 85): the shared
// SubscriberIdentity CHOICE converter, the SriLcs/SriLcsResp top-level
// converters, and their Marshal()/Parse() entry points. Round-trip,
// BER round-trip, and targeted negative cases.
package gsmmap

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
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
		{"MSISDN", SubscriberIdentity{MSISDN: "31612345678"}},
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
		{"neither set", SubscriberIdentity{}, ErrSubscriberIdentityNoAlt},
		{"both set", SubscriberIdentity{IMSI: "204080000000001", MSISDN: "31612345678"}, ErrSubscriberIdentityMultipleAlts},
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
		if !errors.Is(err, ErrSubscriberIdentityUnknownChoice) {
			t.Errorf("want ErrSubscriberIdentityUnknownChoice, got %v", err)
		}
	})
	t.Run("imsi choice but nil payload", func(t *testing.T) {
		_, err := convertWireToSubscriberIdentity(gsm_map.SubscriberIdentity{Choice: gsm_map.SubscriberIdentityChoiceImsi})
		if !errors.Is(err, ErrSubscriberIdentityUnknownChoice) {
			t.Errorf("want ErrSubscriberIdentityUnknownChoice, got %v", err)
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
			TargetMS:        SubscriberIdentity{MSISDN: "31612345678"},
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
		{"no target identity", func(a *SriLcs) { a.TargetMS = SubscriberIdentity{} }, ErrSubscriberIdentityNoAlt},
		{"ambiguous target", func(a *SriLcs) { a.TargetMS = SubscriberIdentity{IMSI: "204080000000001", MSISDN: "31612345678"} }, ErrSubscriberIdentityMultipleAlts},
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

// =============================================================================
// SriLcsResp (Res)
// =============================================================================

func minimalSriLcsResp() *SriLcsResp {
	return &SriLcsResp{
		TargetMS: SubscriberIdentity{IMSI: "204080000000001"},
		LcsLocationInfo: LCSLocationInfo{
			NetworkNodeNumber:       "31650000000",
			NetworkNodeNumberNature: 0x10,
			NetworkNodeNumberPlan:   0x01,
		},
	}
}

func TestSriLcsRespRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   *SriLcsResp
	}{
		{"minimal", minimalSriLcsResp()},
		{"MSISDN target + all GMLC/PPR addresses", &SriLcsResp{
			TargetMS: SubscriberIdentity{MSISDN: "31612345678"},
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
			res, err := convertSriLcsRespToRes(tc.in)
			if err != nil {
				t.Fatalf("toRes: %v", err)
			}
			got, err := convertResToSriLcsResp(res)
			if err != nil {
				t.Fatalf("toStruct: %v", err)
			}
			if !reflect.DeepEqual(tc.in, got) {
				t.Errorf("round-trip mismatch:\n in=%+v\ngot=%+v", tc.in, got)
			}
		})
	}
}

func TestSriLcsRespBERRoundTrip(t *testing.T) {
	in := minimalSriLcsResp()
	in.HGmlcAddress = "192.0.2.10"
	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseSriLcsResp(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !reflect.DeepEqual(in, got) {
		t.Errorf("BER round-trip mismatch:\n in=%+v\ngot=%+v", in, got)
	}
}

func TestSriLcsRespEncodeNegative(t *testing.T) {
	cases := []struct {
		name string
		mut  func(r *SriLcsResp)
		want error
	}{
		{"nil res", nil, ErrSriLcsRespNil},
		{"no target identity", func(r *SriLcsResp) { r.TargetMS = SubscriberIdentity{} }, ErrSubscriberIdentityNoAlt},
		{"empty LcsLocationInfo node", func(r *SriLcsResp) { r.LcsLocationInfo.NetworkNodeNumber = "" }, ErrLCSLocationInfoNetworkNodeEmpty},
		{"invalid GSN address", func(r *SriLcsResp) { r.HGmlcAddress = "not-an-ip" }, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var in *SriLcsResp
			if tc.mut != nil {
				in = minimalSriLcsResp()
				tc.mut(in)
			}
			_, err := convertSriLcsRespToRes(in)
			if tc.want == nil {
				if err == nil {
					t.Errorf("want an error, got nil")
				}
				return
			}
			if !errors.Is(err, tc.want) {
				t.Errorf("want %v, got %v", tc.want, err)
			}
		})
	}
}

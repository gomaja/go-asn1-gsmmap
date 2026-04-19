// lenient_decode_test.go
//
// These tests exercise the spec-mandated lenient-decode behavior for
// extensible ENUMERATED / INTEGER fields per 3GPP TS 29.002. The wire
// inputs are constructed in-memory (bypassing the encoder, which stays
// strict) so we can verify that the decoder applies the exception
// handling clauses from the spec when a peer sends an unknown value.
package gsmmap

import (
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"

	"github.com/gomaja/go-asn1-gsmmap/tbcd"
)

// SupportedCCBSPhase: INTEGER (1..127), spec exception: values 2..127
// shall be mapped to value 1. This decoder surfaces the raw value so the
// caller can observe what the peer sent; the mapping is application
// semantics.
func TestSriDecodeSupportedCCBSPhase_AcceptsRangeUpTo127(t *testing.T) {
	cases := []struct {
		name string
		in   gsm_map.SupportedCCBSPhase
	}{
		{"defined value 1", 1},
		{"reserved value 2", 2},
		{"reserved value 50", 50},
		{"reserved value 127", 127},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			arg := newSriArg()
			v := tc.in
			arg.SupportedCCBSPhase = &v

			s, err := convertArgToSri(arg)
			if err != nil {
				t.Fatalf("convertArgToSri: unexpected error: %v", err)
			}
			if s.SupportedCCBSPhase == nil {
				t.Fatalf("SupportedCCBSPhase: got nil, want %d", tc.in)
			}
			if int64(*s.SupportedCCBSPhase) != int64(tc.in) {
				t.Errorf("SupportedCCBSPhase: got %d, want %d", *s.SupportedCCBSPhase, tc.in)
			}
		})
	}
}

func TestSriDecodeSupportedCCBSPhase_RejectsOutOfRange(t *testing.T) {
	for _, v := range []gsm_map.SupportedCCBSPhase{0, -1, 128} {
		arg := newSriArg()
		x := v
		arg.SupportedCCBSPhase = &x
		if _, err := convertArgToSri(arg); err == nil {
			t.Errorf("SupportedCCBSPhase=%d: expected error, got nil", v)
		}
	}
}

// IstSupportIndicator: ENUMERATED { 0, 1, ... }, spec exception: values
// > 1 shall be mapped to istCommandSupported(1).
func TestSriDecodeIstSupportIndicator_MapsUnknownToOne(t *testing.T) {
	cases := []struct {
		name string
		in   gsm_map.ISTSupportIndicator
		want int
	}{
		{"basicISTSupported", 0, 0},
		{"istCommandSupported", 1, 1},
		{"future value 2 → 1", 2, 1},
		{"future value 99 → 1", 99, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			arg := newSriArg()
			v := tc.in
			arg.IstSupportIndicator = &v

			s, err := convertArgToSri(arg)
			if err != nil {
				t.Fatalf("convertArgToSri: unexpected error: %v", err)
			}
			if s.IstSupportIndicator == nil {
				t.Fatalf("IstSupportIndicator: got nil, want %d", tc.want)
			}
			if *s.IstSupportIndicator != tc.want {
				t.Errorf("IstSupportIndicator: got %d, want %d", *s.IstSupportIndicator, tc.want)
			}
		})
	}
}

func TestSriDecodeIstSupportIndicator_RejectsNegative(t *testing.T) {
	arg := newSriArg()
	v := gsm_map.ISTSupportIndicator(-1)
	arg.IstSupportIndicator = &v
	if _, err := convertArgToSri(arg); err == nil {
		t.Error("IstSupportIndicator=-1: expected error, got nil")
	}
}

func TestUpdateLocationDecodeIstSupportIndicator_MapsUnknownToOne(t *testing.T) {
	cases := []struct {
		name string
		in   gsm_map.ISTSupportIndicator
		want int
	}{
		{"basicISTSupported", 0, 0},
		{"istCommandSupported", 1, 1},
		{"future value 5 → 1", 5, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			arg := newUpdateLocationArg()
			v := tc.in
			arg.VlrCapability = &gsm_map.VLRCapability{IstSupportIndicator: &v}

			u, err := convertArgToUpdateLocation(arg)
			if err != nil {
				t.Fatalf("convertArgToUpdateLocation: unexpected error: %v", err)
			}
			if u.VlrCapability == nil || u.VlrCapability.IstSupportIndicator == nil {
				t.Fatalf("VlrCapability.IstSupportIndicator: got nil, want %d", tc.want)
			}
			if *u.VlrCapability.IstSupportIndicator != tc.want {
				t.Errorf("VlrCapability.IstSupportIndicator: got %d, want %d",
					*u.VlrCapability.IstSupportIndicator, tc.want)
			}
		})
	}
}

// NumberPortabilityStatus: ENUMERATED { 0, 1, 2, 4, 5 }, spec exception:
// reception of other values shall ignore the whole field.
func TestSriRespDecodeNumberPortabilityStatus_IgnoresUnknown(t *testing.T) {
	cases := []struct {
		name    string
		in      gsm_map.NumberPortabilityStatus
		wantNil bool
		want    NumberPortabilityStatus
	}{
		{"defined 0 (mnpNotKnownToBePorted)", 0, false, MnpNotKnownToBePorted},
		{"defined 1 (mnpOwnNumberPortedOut)", 1, false, MnpOwnNumberPortedOut},
		{"defined 2 (mnpForeignNumberPortedToForeignNetwork)", 2, false, MnpForeignNumberPortedToForeignNetwork},
		{"defined 4 (mnpOwnNumberNotPortedOut)", 4, false, MnpOwnNumberNotPortedOut},
		{"defined 5 (mnpForeignNumberPortedIn)", 5, false, MnpForeignNumberPortedIn},
		{"undefined 3 → ignored", 3, true, 0},
		{"undefined 6 → ignored", 6, true, 0},
		{"undefined 99 → ignored", 99, true, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := newSriRes()
			v := tc.in
			res.NumberPortabilityStatus = &v

			out, err := convertResToSriResp(res)
			if err != nil {
				t.Fatalf("convertResToSriResp: unexpected error: %v", err)
			}
			if tc.wantNil {
				if out.NumberPortabilityStatus != nil {
					t.Errorf("NumberPortabilityStatus: got %v, want nil", *out.NumberPortabilityStatus)
				}
				return
			}
			if out.NumberPortabilityStatus == nil {
				t.Fatalf("NumberPortabilityStatus: got nil, want %d", tc.want)
			}
			if *out.NumberPortabilityStatus != tc.want {
				t.Errorf("NumberPortabilityStatus: got %d, want %d", *out.NumberPortabilityStatus, tc.want)
			}
		})
	}
}

func TestMnpInfoResDecodeNumberPortabilityStatus_IgnoresUnknown(t *testing.T) {
	cases := []struct {
		name    string
		in      gsm_map.NumberPortabilityStatus
		wantNil bool
	}{
		{"defined 5 → kept", 5, false},
		{"undefined 3 → ignored", 3, true},
		{"undefined 99 → ignored", 99, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := tc.in
			w := &gsm_map.MNPInfoRes{NumberPortabilityStatus: &v}

			out, err := convertWireToMnpInfoRes(w)
			if err != nil {
				t.Fatalf("convertWireToMnpInfoRes: unexpected error: %v", err)
			}
			if tc.wantNil {
				if out.NumberPortabilityStatus != nil {
					t.Errorf("NumberPortabilityStatus: got %v, want nil", *out.NumberPortabilityStatus)
				}
				return
			}
			if out.NumberPortabilityStatus == nil {
				t.Errorf("NumberPortabilityStatus: got nil, want value")
			}
		})
	}
}

// UnavailabilityCause: ENUMERATED 1..6 extensible, spec exception:
// reception of unknown values shall result in service-unavailable for
// that call. The protocol decoder surfaces the raw value; the
// service-unavailable behavior is application-layer.
func TestSriRespDecodeUnavailabilityCause_AcceptsUnknown(t *testing.T) {
	cases := []struct {
		name string
		in   gsm_map.UnavailabilityCause
	}{
		{"defined 1", 1},
		{"defined 6", 6},
		{"future 7", 7},
		{"future 50", 50},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := newSriRes()
			v := tc.in
			res.UnavailabilityCause = &v

			out, err := convertResToSriResp(res)
			if err != nil {
				t.Fatalf("convertResToSriResp: unexpected error: %v", err)
			}
			if out.UnavailabilityCause == nil {
				t.Fatalf("UnavailabilityCause: got nil, want %d", tc.in)
			}
			if int64(*out.UnavailabilityCause) != int64(tc.in) {
				t.Errorf("UnavailabilityCause: got %d, want %d", *out.UnavailabilityCause, tc.in)
			}
		})
	}
}

func TestSriRespDecodeUnavailabilityCause_RejectsNegative(t *testing.T) {
	res := newSriRes()
	v := gsm_map.UnavailabilityCause(-1)
	res.UnavailabilityCause = &v
	if _, err := convertResToSriResp(res); err == nil {
		t.Error("UnavailabilityCause=-1: expected error, got nil")
	}
}

// RequestingNodeType: ENUMERATED { vlr(0), sgsn(1), s-cscf(2), bsf(3),
// gan-aaa-server(4), wlan-aaa-server(5), mme(16), mme-sgsn(17) }. Spec:
//
//	received values in the range (6-15) shall be treated as 'vlr'
//	received values greater than 17 shall be treated as 'sgsn'
func TestSaiDecodeRequestingNodeType_AppliesSpecMapping(t *testing.T) {
	cases := []struct {
		name string
		in   gsm_map.RequestingNodeType
		want RequestingNodeType
	}{
		{"vlr (0)", 0, RequestingNodeVlr},
		{"sgsn (1)", 1, RequestingNodeSgsn},
		{"s-cscf (2)", 2, RequestingNodeSCscf},
		{"wlan-aaa-server (5)", 5, RequestingNodeWlanAAAServer},
		{"reserved 6 → vlr", 6, RequestingNodeVlr},
		{"reserved 10 → vlr", 10, RequestingNodeVlr},
		{"reserved 15 → vlr", 15, RequestingNodeVlr},
		{"mme (16)", 16, RequestingNodeMme},
		{"mme-sgsn (17)", 17, RequestingNodeMmeSgsn},
		{"reserved 18 → sgsn", 18, RequestingNodeSgsn},
		{"reserved 99 → sgsn", 99, RequestingNodeSgsn},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			arg := newSaiArg()
			v := tc.in
			arg.RequestingNodeType = &v

			s, err := convertArgToSendAuthenticationInfo(arg)
			if err != nil {
				t.Fatalf("convertArgToSendAuthenticationInfo: unexpected error: %v", err)
			}
			if s.RequestingNodeType == nil {
				t.Fatalf("RequestingNodeType: got nil, want %d", tc.want)
			}
			if *s.RequestingNodeType != tc.want {
				t.Errorf("RequestingNodeType: got %d, want %d", *s.RequestingNodeType, tc.want)
			}
		})
	}
}

func TestSaiDecodeRequestingNodeType_RejectsNegative(t *testing.T) {
	arg := newSaiArg()
	v := gsm_map.RequestingNodeType(-1)
	arg.RequestingNodeType = &v
	if _, err := convertArgToSendAuthenticationInfo(arg); err == nil {
		t.Error("RequestingNodeType=-1: expected error, got nil")
	}
}

// --- helpers ---

// newSriArg returns a minimally valid SendRoutingInfoArg whose mandatory
// fields are populated, so optional fields under test can be added in
// isolation.
func newSriArg() *gsm_map.SendRoutingInfoArg {
	msisdn, err := encodeAddressField("31612345678", 1, 1)
	if err != nil {
		panic(err)
	}
	gsmscf, err := encodeAddressField("31600000000", 1, 1)
	if err != nil {
		panic(err)
	}
	return &gsm_map.SendRoutingInfoArg{
		Msisdn:              gsm_map.ISDNAddressString(msisdn),
		InterrogationType:   gsm_map.InterrogationType(0),
		GmscOrGsmSCFAddress: gsm_map.ISDNAddressString(gsmscf),
	}
}

// newSriRes returns an empty SendRoutingInfoRes — every field is
// optional in this direction.
func newSriRes() *gsm_map.SendRoutingInfoRes {
	return &gsm_map.SendRoutingInfoRes{}
}

// newUpdateLocationArg returns a minimally valid UpdateLocationArg.
func newUpdateLocationArg() *gsm_map.UpdateLocationArg {
	msc, err := encodeAddressField("31600000001", 1, 1)
	if err != nil {
		panic(err)
	}
	vlr, err := encodeAddressField("31600000002", 1, 1)
	if err != nil {
		panic(err)
	}
	imsi, err := tbcd.Encode("204080012345678")
	if err != nil {
		panic(err)
	}
	return &gsm_map.UpdateLocationArg{
		Imsi:      gsm_map.IMSI(imsi),
		MscNumber: gsm_map.ISDNAddressString(msc),
		VlrNumber: gsm_map.ISDNAddressString(vlr),
	}
}

// newSaiArg returns a minimally valid SendAuthenticationInfoArg.
func newSaiArg() *gsm_map.SendAuthenticationInfoArg {
	imsi, err := tbcd.Encode("204080012345678")
	if err != nil {
		panic(err)
	}
	return &gsm_map.SendAuthenticationInfoArg{
		Imsi:                     gsm_map.IMSI(imsi),
		NumberOfRequestedVectors: gsm_map.NumberOfRequestedVectors(1),
	}
}

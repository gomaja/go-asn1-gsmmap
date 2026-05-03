// absent_subscriber_diagnostic_sm_test.go
//
// Tests for the wrapper-level AbsentSubscriberDiagnosticSM named type.
// The upstream gsm_map.AbsentSubscriberDiagnosticSM is `type … = int64`
// (alias) because the ASN.1 declaration is `INTEGER (0..255)` rather
// than `ENUMERATED`; the wrapper promotes it to a named type so the
// String() ergonomics match the other diagnostic enums in this package.
package gsmmap

import (
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// String() values follow TS 23.040 §3.3.2 wording (camelCase /
// hyphenated). Codes outside the named set return "unknown" without
// being rejected — wire values 0..255 are still permitted by the spec.
func TestAbsentSubscriberDiagnosticSMString(t *testing.T) {
	cases := []struct {
		in   AbsentSubscriberDiagnosticSM
		want string
	}{
		{AbsentSubscriberDiagnosticNoPagingResponseViaTheMSC, "noPagingResponseViaTheMSC"},
		{AbsentSubscriberDiagnosticImsiDetached, "imsiDetached"},
		{AbsentSubscriberDiagnosticRoamingRestriction, "roamingRestriction"},
		{AbsentSubscriberDiagnosticDeregisteredInTheHLRForNonGPRS, "deregisteredInTheHLR-ForNonGPRS"},
		{AbsentSubscriberDiagnosticMsPurgedForNonGPRS, "msPurged-ForNonGPRS"},
		{AbsentSubscriberDiagnosticNoPagingResponseViaTheSGSN, "noPagingResponseViaTheSGSN"},
		{AbsentSubscriberDiagnosticGPRSDetached, "gprsDetached"},
		{AbsentSubscriberDiagnosticDeregisteredInTheHLRForGPRS, "deregisteredInTheHLR-ForGPRS"},
		{AbsentSubscriberDiagnosticMsPurgedForGPRS, "msPurged-ForGPRS"},
		{AbsentSubscriberDiagnosticUnidentifiedSubscriberViaTheMSC, "unidentifiedSubscriberViaTheMSC"},
		{AbsentSubscriberDiagnosticUnidentifiedSubscriberViaTheSGSN, "unidentifiedSubscriberViaTheSGSN"},
		// Out-of-set values map to "unknown" without rejection.
		{AbsentSubscriberDiagnosticSM(99), "unknown"},
		{AbsentSubscriberDiagnosticSM(255), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.in.String(); got != tc.want {
			t.Errorf("AbsentSubscriberDiagnosticSM(%d).String(): want %q, got %q", int64(tc.in), tc.want, got)
		}
	}
}

// Constants must resolve to the spec-defined integer values per
// TS 23.040 §3.3.2.
func TestAbsentSubscriberDiagnosticSMValues(t *testing.T) {
	cases := []struct {
		name string
		got  AbsentSubscriberDiagnosticSM
		want int64
	}{
		{"NoPagingResponseViaTheMSC", AbsentSubscriberDiagnosticNoPagingResponseViaTheMSC, 0},
		{"ImsiDetached", AbsentSubscriberDiagnosticImsiDetached, 1},
		{"RoamingRestriction", AbsentSubscriberDiagnosticRoamingRestriction, 2},
		{"DeregisteredInTheHLRForNonGPRS", AbsentSubscriberDiagnosticDeregisteredInTheHLRForNonGPRS, 3},
		{"MsPurgedForNonGPRS", AbsentSubscriberDiagnosticMsPurgedForNonGPRS, 4},
		{"NoPagingResponseViaTheSGSN", AbsentSubscriberDiagnosticNoPagingResponseViaTheSGSN, 5},
		{"GPRSDetached", AbsentSubscriberDiagnosticGPRSDetached, 6},
		{"DeregisteredInTheHLRForGPRS", AbsentSubscriberDiagnosticDeregisteredInTheHLRForGPRS, 7},
		{"MsPurgedForGPRS", AbsentSubscriberDiagnosticMsPurgedForGPRS, 8},
		{"UnidentifiedSubscriberViaTheMSC", AbsentSubscriberDiagnosticUnidentifiedSubscriberViaTheMSC, 9},
		{"UnidentifiedSubscriberViaTheSGSN", AbsentSubscriberDiagnosticUnidentifiedSubscriberViaTheSGSN, 10},
	}
	for _, tc := range cases {
		if int64(tc.got) != tc.want {
			t.Errorf("%s: want %d, got %d", tc.name, tc.want, int64(tc.got))
		}
	}
}

// The wrapper type and the upstream type alias share the underlying
// int64 representation: a value cast between them at the converter
// boundary preserves the integer exactly. This test exercises the
// cast that convertWireToAbsentSubscriberSMParam relies on.
func TestAbsentSubscriberDiagnosticSMUpstreamCastRoundTrip(t *testing.T) {
	for raw := int64(0); raw <= 10; raw++ {
		upstream := gsm_map.AbsentSubscriberDiagnosticSM(raw)
		local := AbsentSubscriberDiagnosticSM(upstream)
		back := gsm_map.AbsentSubscriberDiagnosticSM(local)
		if back != upstream {
			t.Errorf("cast round-trip failed for %d: got %d", raw, back)
		}
	}
}

// AbsentSubscriberSMParam fields use the wrapper-level named type;
// String() works directly without dropping down to gsm_map. This is
// the operational ergonomics the named-type promotion buys.
func TestAbsentSubscriberSMParamFieldStringDirect(t *testing.T) {
	diag := AbsentSubscriberDiagnosticImsiDetached
	addDiag := AbsentSubscriberDiagnosticMsPurgedForNonGPRS
	p := &AbsentSubscriberSMParam{
		AbsentSubscriberDiagnosticSM:           &diag,
		AdditionalAbsentSubscriberDiagnosticSM: &addDiag,
	}
	if got := p.AbsentSubscriberDiagnosticSM.String(); got != "imsiDetached" {
		t.Errorf("AbsentSubscriberDiagnosticSM.String(): want %q, got %q", "imsiDetached", got)
	}
	if got := p.AdditionalAbsentSubscriberDiagnosticSM.String(); got != "msPurged-ForNonGPRS" {
		t.Errorf("AdditionalAbsentSubscriberDiagnosticSM.String(): want %q, got %q", "msPurged-ForNonGPRS", got)
	}
}

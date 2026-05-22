// slr_foundation_test.go
//
// Tests for SubscriberLocationReport (opCode 86) foundation types.
// PR G1 of the staged SLR implementation — converters and top-level
// Arg/Res structs land in follow-up PRs.
package gsmmap

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// Compile-smoke: every new public type must be referenceable.
func TestSLRTypesCompile(t *testing.T) {
	var _ LCSEvent
	var _ SequenceNumber
	var _ LCSLocationInfo
	var _ DeferredmtLrData

	// Constants exist (compile-smoke). Numeric equivalence to upstream
	// values is verified by TestSLREnumsAliasUpstream below.
	_ = LCSEventEmergencyCallOrigination
	_ = LCSEventEmergencyCallRelease
	_ = LCSEventMoLr
	_ = LCSEventDeferredmtLrResponse
	_ = LCSEventDeferredmoLrTTTPInitiation
	_ = LCSEventEmergencyCallHandover
}

// Aliased enums/scalars must resolve to the same numeric values as
// upstream so callers can use either local or upstream names
// interchangeably.
func TestSLREnumsAliasUpstream(t *testing.T) {
	cases := []struct {
		name  string
		local int64
		upstr int64
	}{
		{"LCSEventEmergencyCallOrigination", int64(LCSEventEmergencyCallOrigination), int64(gsm_map.LCSEventEmergencyCallOrigination)},
		{"LCSEventEmergencyCallRelease", int64(LCSEventEmergencyCallRelease), int64(gsm_map.LCSEventEmergencyCallRelease)},
		{"LCSEventMoLr", int64(LCSEventMoLr), int64(gsm_map.LCSEventMoLr)},
		{"LCSEventDeferredmtLrResponse", int64(LCSEventDeferredmtLrResponse), int64(gsm_map.LCSEventDeferredmtLrResponse)},
		{"LCSEventDeferredmoLrTTTPInitiation", int64(LCSEventDeferredmoLrTTTPInitiation), int64(gsm_map.LCSEventDeferredmoLrTTTPInitiation)},
		{"LCSEventEmergencyCallHandover", int64(LCSEventEmergencyCallHandover), int64(gsm_map.LCSEventEmergencyCallHandover)},
	}
	for _, tc := range cases {
		if tc.local != tc.upstr {
			t.Errorf("%s: local=%d upstream=%d", tc.name, tc.local, tc.upstr)
		}
	}
}

// LCSEvent is aliased to the upstream named type, so String() works
// directly (it is an ENUMERATED upstream, unlike AbsentSubscriberDiagnosticSM).
func TestSLRLCSEventString(t *testing.T) {
	cases := []struct {
		in   LCSEvent
		want string
	}{
		{LCSEventEmergencyCallOrigination, "emergencyCallOrigination"},
		{LCSEventMoLr, "mo-lr"},
		{LCSEventDeferredmtLrResponse, "deferredmt-lrResponse"},
	}
	for _, tc := range cases {
		if got := tc.in.String(); got != tc.want {
			t.Errorf("LCSEvent(%d).String(): want %q, got %q", int64(tc.in), tc.want, got)
		}
	}
}

// SequenceNumber is an int64 alias; values flow without conversion and
// the range bounds match the spec (shares maxReportingAmount).
func TestSLRSequenceNumberAlias(t *testing.T) {
	var s SequenceNumber = 100
	if int64(s) != 100 {
		t.Fatalf("SequenceNumber alias: want 100, got %d", s)
	}
	if SequenceNumberMin != 1 {
		t.Errorf("SequenceNumberMin: want 1, got %d", SequenceNumberMin)
	}
	if SequenceNumberMax != 8639999 {
		t.Errorf("SequenceNumberMax: want 8639999 (maxReportingAmount), got %d", SequenceNumberMax)
	}
	// Direct comparison without casts.
	if s < SequenceNumberMin || s > SequenceNumberMax {
		t.Error("range check: 100 should be in [Min..Max]")
	}
}

// Sentinel errors must be defined, distinct, and detectable through
// errors.Is when wrapped via %w.
func TestSLRSentinelErrors(t *testing.T) {
	sentinels := []error{
		ErrLCSEventInvalid,
		ErrSequenceNumberOutOfRange,
		ErrLCSLocationInfoNetworkNodeEmpty,
		ErrLCSLocationInfoNetworkNodeDecodedEmpty,
		ErrLCSLocationInfoLMSIInvalidSize,
		ErrLCSLocationInfoMmeNameSize,
		ErrLCSLocationInfoAaaServerNameSize,
		ErrLCSLocationInfoSgsnNameSize,
		ErrLCSLocationInfoSgsnRealmSize,
	}
	seen := make(map[error]int, len(sentinels))
	for i, s := range sentinels {
		if s == nil {
			t.Errorf("sentinel #%d is nil", i)
			continue
		}
		if j, dup := seen[s]; dup {
			t.Errorf("sentinel #%d aliases sentinel #%d (same error value)", i, j)
		}
		seen[s] = i
		wrapped := fmt.Errorf("slr wrapper: %w", s)
		if !errors.Is(wrapped, s) {
			t.Errorf("sentinel #%d not detectable through errors.Is when wrapped with %%w", i)
		}
	}
}

// Foundation struct shapes must be zero-value safe so the public API
// can be constructed incrementally before the codec lands.
func TestSLRZeroValues(t *testing.T) {
	var li LCSLocationInfo
	if li.NetworkNodeNumber != "" {
		t.Error("LCSLocationInfo zero value should have empty NetworkNodeNumber")
	}
	if li.LMSI != nil || li.MmeName != nil {
		t.Error("LCSLocationInfo zero value should have nil LMSI/MmeName")
	}
	if li.GprsNodeIndicator {
		t.Error("LCSLocationInfo zero value should have GprsNodeIndicator=false")
	}
	if li.AdditionalNumber != nil || li.SupportedLCSCapabilitySets != nil {
		t.Error("LCSLocationInfo zero value should have nil pointer sub-fields")
	}

	var d DeferredmtLrData
	if d.TerminationCause != nil || d.LcsLocationInfo != nil {
		t.Error("DeferredmtLrData zero value should have nil optional fields")
	}
	if d.DeferredLocationEventType.MsAvailable {
		t.Error("DeferredmtLrData zero value should have empty DeferredLocationEventType")
	}
}

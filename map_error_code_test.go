// map_error_code_test.go
//
// Tests for the typed MapErrorCode enum and upstream local-error-code
// constants. PR F3 of the staged ReturnError.Parameter implementation.
package gsmmap

import (
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// MapErrorCode is an alias for gsm_map.GSMMAPLocalErrorcode; values pass without
// casts and the constants line up with the upstream values.
func TestMapErrorCodeAliasUpstream(t *testing.T) {
	cases := []struct {
		name string
		got  MapErrorCode
		want int64
	}{
		{"MapErrorUnknownSubscriber", MapErrorUnknownSubscriber, 1},
		{"MapErrorAbsentSubscriberSM", MapErrorAbsentSubscriberSM, 6},
		{"MapErrorRoamingNotAllowed", MapErrorRoamingNotAllowed, 8},
		{"MapErrorTeleserviceNotProvisioned", MapErrorTeleserviceNotProvisioned, 11},
		{"MapErrorCallBarred", MapErrorCallBarred, 13},
		{"MapErrorFacilityNotSupported", MapErrorFacilityNotSupported, 21},
		{"MapErrorAbsentSubscriber", MapErrorAbsentSubscriber, 27},
		{"MapErrorSystemFailure", MapErrorSystemFailure, 34},
		{"MapErrorDataMissing", MapErrorDataMissing, 35},
		{"MapErrorUnauthorizedRequestingNetwork", MapErrorUnauthorizedRequestingNetwork, 52},
	}
	for _, tc := range cases {
		if int64(tc.got) != tc.want {
			t.Errorf("%s: want %d, got %d", tc.name, tc.want, int64(tc.got))
		}
	}
}

// String() works on the typed enum without a cast (delegated to the
// upstream gsm_map.GSMMAPLocalErrorcode method).
func TestMapErrorCodeString(t *testing.T) {
	cases := []struct {
		got  MapErrorCode
		want string
	}{
		{MapErrorUnknownSubscriber, "unknownSubscriber"},
		{MapErrorAbsentSubscriberSM, "absentSubscriberSM"},
		{MapErrorRoamingNotAllowed, "roamingNotAllowed"},
		{MapErrorCallBarred, "callBarred"},
		{MapErrorSystemFailure, "systemFailure"},
		{MapErrorDataMissing, "dataMissing"},
	}
	for _, tc := range cases {
		if got := tc.got.String(); got != tc.want {
			t.Errorf("MapErrorCode(%d).String(): want %q, got %q", int64(tc.got), tc.want, got)
		}
	}
}

// GetErrorString continues to work — it delegates to upstream
// gsm_map.GSMMAPLocalErrorcode.String(). Existing callers passing raw
// int64 must not regress.
func TestGetErrorStringRegression(t *testing.T) {
	cases := []struct {
		errCode int64
		want    string
	}{
		{1, "unknownSubscriber"},
		{6, "absentSubscriberSM"},
		{34, "systemFailure"},
		{52, "unauthorizedRequestingNetwork"},
	}
	for _, tc := range cases {
		if got := GetErrorString(tc.errCode); got != tc.want {
			t.Errorf("GetErrorString(%d): want %q, got %q", tc.errCode, tc.want, got)
		}
	}
}

// MapErrorCode is a type alias for gsm_map.GSMMAPLocalErrorcode, so callers can
// use either form interchangeably without conversions. Verified by
// passing each constant through a function whose parameter is typed
// as the other side of the alias — if the alias relationship breaks,
// these calls stop compiling, matching the TestPSLByteAliases
// pattern used elsewhere in the package.
func TestMapErrorCodeUpstreamInterchangeable(t *testing.T) {
	// Pass a local constant where an upstream type is expected.
	asUpstream := func(v gsm_map.GSMMAPLocalErrorcode) gsm_map.GSMMAPLocalErrorcode { return v }
	if got := asUpstream(MapErrorCallBarred); got != gsm_map.GSMMAPLocalErrorcodeCallBarred {
		t.Errorf("local MapErrorCallBarred → upstream gsm_map.GSMMAPLocalErrorcodeCallBarred: want %d, got %d",
			gsm_map.GSMMAPLocalErrorcodeCallBarred, got)
	}

	// Pass an upstream constant where a local type is expected.
	asLocal := func(v MapErrorCode) MapErrorCode { return v }
	if got := asLocal(gsm_map.GSMMAPLocalErrorcodeSystemFailure); got != MapErrorSystemFailure {
		t.Errorf("upstream gsm_map.GSMMAPLocalErrorcodeSystemFailure → local MapErrorSystemFailure: want %d, got %d",
			MapErrorSystemFailure, got)
	}

	// ParseReturnErrorParameter takes int64 to match TCAP's wire
	// type; callers using MapErrorCode constants pass an explicit
	// cast.
	emptySeq := []byte{0x30, 0x00}
	if _, err := ParseReturnErrorParameter(int64(MapErrorUnknownSubscriber), emptySeq); err != nil {
		t.Errorf("ParseReturnErrorParameter(int64(MapErrorUnknownSubscriber)): %v", err)
	}
	if _, err := ParseReturnErrorParameter(int64(gsm_map.GSMMAPLocalErrorcodeUnknownSubscriber), emptySeq); err != nil {
		t.Errorf("ParseReturnErrorParameter(int64(gsm_map.GSMMAPLocalErrorcodeUnknownSubscriber)): %v", err)
	}
}

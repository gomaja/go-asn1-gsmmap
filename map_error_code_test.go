// map_error_code_test.go
//
// Tests for the MapErrorCode enum against the local error codes go-asn1
// generates from the MAP-Errors module.
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

// Each local MapErrorCode constant carries the CODE local value that
// go-asn1 generated from the MAP-Errors module, and String() returns the
// generated ASN.1 name.
func TestMapErrorCodeMatchesUpstream(t *testing.T) {
	cases := []struct {
		got      MapErrorCode
		upstream int64
		name     string
	}{
		{MapErrorUnknownSubscriber, gsm_map.GSMMAPLocalErrorcodeUnknownSubscriber, "unknownSubscriber"},
		{MapErrorAbsentSubscriberSM, gsm_map.GSMMAPLocalErrorcodeAbsentSubscriberSM, "absentSubscriberSM"},
		{MapErrorRoamingNotAllowed, gsm_map.GSMMAPLocalErrorcodeRoamingNotAllowed, "roamingNotAllowed"},
		{MapErrorIllegalSubscriber, gsm_map.GSMMAPLocalErrorcodeIllegalSubscriber, "illegalSubscriber"},
		{MapErrorTeleserviceNotProvisioned, gsm_map.GSMMAPLocalErrorcodeTeleserviceNotProvisioned, "teleserviceNotProvisioned"},
		{MapErrorIllegalEquipment, gsm_map.GSMMAPLocalErrorcodeIllegalEquipment, "illegalEquipment"},
		{MapErrorCallBarred, gsm_map.GSMMAPLocalErrorcodeCallBarred, "callBarred"},
		{MapErrorFacilityNotSupported, gsm_map.GSMMAPLocalErrorcodeFacilityNotSupported, "facilityNotSupported"},
		{MapErrorAbsentSubscriber, gsm_map.GSMMAPLocalErrorcodeAbsentSubscriber, "absentSubscriber"},
		{MapErrorSystemFailure, gsm_map.GSMMAPLocalErrorcodeSystemFailure, "systemFailure"},
		{MapErrorDataMissing, gsm_map.GSMMAPLocalErrorcodeDataMissing, "dataMissing"},
		{MapErrorUnexpectedDataValue, gsm_map.GSMMAPLocalErrorcodeUnexpectedDataValue, "unexpectedDataValue"},
		{MapErrorUnauthorizedRequestingNetwork, gsm_map.GSMMAPLocalErrorcodeUnauthorizedRequestingNetwork, "unauthorizedRequestingNetwork"},
		{MapErrorUnknownAlphabet, gsm_map.GSMMAPLocalErrorcodeUnknownAlphabet, "unknownAlphabet"},
		{MapErrorUSSDBusy, gsm_map.GSMMAPLocalErrorcodeUssdBusy, "ussd-Busy"},
	}
	for _, tc := range cases {
		if int64(tc.got) != tc.upstream {
			t.Errorf("%s: local code %d, upstream code %d", tc.name, int64(tc.got), tc.upstream)
		}
		if s := tc.got.String(); s != tc.name {
			t.Errorf("MapErrorCode(%d).String() = %q, want %q", int64(tc.got), s, tc.name)
		}
	}
	if s := MapErrorCode(250).String(); s != "250" {
		t.Errorf("MapErrorCode(250).String() = %q, want \"250\"", s)
	}
}

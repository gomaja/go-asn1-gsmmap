// map_error_code_test.go
//
// Tests for the MapErrorCode constants against the local error codes of
// 3GPP TS 29.002 V19.1.0 §17.6.6 and the names go-asn1 gives them.
package gsmmap

import "testing"

// The constants carry the CODE local values of 3GPP TS 29.002 V19.1.0
// §17.6.6.
func TestMapErrorCodeSpecValues(t *testing.T) {
	cases := []struct {
		name string
		got  MapErrorCode
		want int64
	}{
		{"MapErrorUnknownSubscriber", MapErrorUnknownSubscriber, 1},
		{"MapErrorAbsentSubscriberSM", MapErrorAbsentSubscriberSM, 6},
		{"MapErrorRoamingNotAllowed", MapErrorRoamingNotAllowed, 8},
		{"MapErrorIllegalSubscriber", MapErrorIllegalSubscriber, 9},
		{"MapErrorTeleserviceNotProvisioned", MapErrorTeleserviceNotProvisioned, 11},
		{"MapErrorIllegalEquipment", MapErrorIllegalEquipment, 12},
		{"MapErrorCallBarred", MapErrorCallBarred, 13},
		{"MapErrorFacilityNotSupported", MapErrorFacilityNotSupported, 21},
		{"MapErrorAbsentSubscriber", MapErrorAbsentSubscriber, 27},
		{"MapErrorSystemFailure", MapErrorSystemFailure, 34},
		{"MapErrorDataMissing", MapErrorDataMissing, 35},
		{"MapErrorUnexpectedDataValue", MapErrorUnexpectedDataValue, 36},
		{"MapErrorUnauthorizedRequestingNetwork", MapErrorUnauthorizedRequestingNetwork, 52},
		{"MapErrorUnknownAlphabet", MapErrorUnknownAlphabet, 71},
		{"MapErrorUSSDBusy", MapErrorUSSDBusy, 72},
	}
	for _, tc := range cases {
		if int64(tc.got) != tc.want {
			t.Errorf("%s: want %d, got %d", tc.name, tc.want, int64(tc.got))
		}
	}
}

// String returns the ASN.1 name of each MapErrorCode constant, and the
// MAPErrorCode(n) for a code without a name.
func TestMapErrorCodeNames(t *testing.T) {
	cases := []struct {
		got  MapErrorCode
		name string
	}{
		{MapErrorUnknownSubscriber, "unknownSubscriber"},
		{MapErrorAbsentSubscriberSM, "absentSubscriberSM"},
		{MapErrorRoamingNotAllowed, "roamingNotAllowed"},
		{MapErrorIllegalSubscriber, "illegalSubscriber"},
		{MapErrorTeleserviceNotProvisioned, "teleserviceNotProvisioned"},
		{MapErrorIllegalEquipment, "illegalEquipment"},
		{MapErrorCallBarred, "callBarred"},
		{MapErrorFacilityNotSupported, "facilityNotSupported"},
		{MapErrorAbsentSubscriber, "absentSubscriber"},
		{MapErrorSystemFailure, "systemFailure"},
		{MapErrorDataMissing, "dataMissing"},
		{MapErrorUnexpectedDataValue, "unexpectedDataValue"},
		{MapErrorUnauthorizedRequestingNetwork, "unauthorizedRequestingNetwork"},
		{MapErrorUnknownAlphabet, "unknownAlphabet"},
		{MapErrorUSSDBusy, "ussd-Busy"},
	}
	for _, tc := range cases {
		if s := tc.got.String(); s != tc.name {
			t.Errorf("MapErrorCode(%d).String() = %q, want %q", int64(tc.got), s, tc.name)
		}
	}
	if s := MapErrorCode(250).String(); s != "MAPErrorCode(250)" {
		t.Errorf("MapErrorCode(250).String() = %q, want \"MAPErrorCode(250)\"", s)
	}
}

// 3GPP TS 29.002 V19.1.0 §17.6 defines no error with local code 4.
func TestMapErrorCode4HasNoName(t *testing.T) {
	if s := MapErrorCode(4).String(); s != "MAPErrorCode(4)" {
		t.Errorf("MapErrorCode(4).String() = %q, want %q", s, "MAPErrorCode(4)")
	}
}

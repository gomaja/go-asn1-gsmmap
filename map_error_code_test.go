// map_error_code_test.go
//
// Tests for the MapErrorCode constants against the local error codes of
// 3GPP TS 29.002 V19.1.0 §17.6.6 and go-asn1's error code list.
package gsmmap

import (
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

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

// String returns the name from go-asn1's error code list.
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

// Each MapErrorCode constant equals go-asn1's error code, and String returns
// its ASN.1 name.
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

// https://github.com/gomaja/go-asn1/issues/135: go-asn1's error code list
// names code 4 secureTransportError, which 3GPP TS 29.002 V19.1.0 §17.6.6
// does not define. This pins today's behaviour; when it fails, go-asn1 has
// dropped the name and String returns "4", the decimal code, as for any code
// without a name.
func TestGoASN1Issue135ErrorCode4(t *testing.T) {
	if s := MapErrorCode(4).String(); s != "secureTransportError" {
		t.Errorf("MapErrorCode(4).String() = %q, want \"secureTransportError\"", s)
	}
}

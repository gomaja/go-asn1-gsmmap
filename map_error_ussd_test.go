// map_error_ussd_test.go
//
// Tests for the ReturnError parameters of the errors the USSD operations
// raise (TS 29.002 V19.1.0 §17.6.4): illegalSubscriber (9), illegalEquipment
// (12), unexpectedDataValue (36), unknownAlphabet (71) and ussd-Busy (72).
package gsmmap

import (
	"encoding/hex"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Golden parameters, each decoded by Wireshark 4.6.4 as the parameter of a
// TCAP ReturnError with the given local error code.
func TestParseReturnErrorParameterUSSDErrors(t *testing.T) {
	cases := []struct {
		name string
		code MapErrorCode
		hex  string
		want any
	}{
		{
			// Wireshark: unexpectedDataValue (36), UnexpectedDataParam,
			// unexpectedSubscriber: NULL.
			name: "unexpectedDataValue with unexpectedSubscriber",
			code: MapErrorUnexpectedDataValue,
			hex:  "30028000",
			want: &UnexpectedDataParam{UnexpectedSubscriber: true},
		},
		{
			// Wireshark: unexpectedDataValue (36), empty UnexpectedDataParam.
			name: "unexpectedDataValue empty",
			code: MapErrorUnexpectedDataValue,
			hex:  "3000",
			want: &UnexpectedDataParam{},
		},
		{
			// Wireshark: unexpectedDataValue (36), extensionContainer and
			// unexpectedSubscriber: NULL.
			name: "unexpectedDataValue with extensionContainer",
			code: MapErrorUnexpectedDataValue,
			hex:  "300430008000",
			want: &UnexpectedDataParam{UnexpectedSubscriber: true},
		},
		{
			// Wireshark: illegalSubscriber (9), IllegalSubscriberParam with
			// an empty extensionContainer.
			name: "illegalSubscriber with extensionContainer",
			code: MapErrorIllegalSubscriber,
			hex:  "30023000",
			want: &IllegalSubscriberParam{},
		},
		{
			// Wireshark: illegalEquipment (12), empty IllegalEquipmentParam.
			name: "illegalEquipment empty",
			code: MapErrorIllegalEquipment,
			hex:  "3000",
			want: &IllegalEquipmentParam{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := hex.DecodeString(tc.hex)
			if err != nil {
				t.Fatal(err)
			}
			got, err := ParseReturnErrorParameter(tc.code, data)
			if err != nil {
				t.Fatalf("ParseReturnErrorParameter(%v): %v", tc.code, err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ParseReturnErrorParameter(%v) diff (-want +got):\n%s", tc.code, diff)
			}
		})
	}
}

// unknownAlphabet and ussd-Busy are defined without a PARAMETER
// (TS 29.002 V19.1.0 §17.6.6), so there is nothing to decode whether or not
// a peer sends one.
func TestParseReturnErrorParameterErrorsWithoutParameter(t *testing.T) {
	for _, code := range []MapErrorCode{MapErrorUnknownAlphabet, MapErrorUSSDBusy} {
		for _, data := range [][]byte{nil, {0x30, 0x00}} {
			got, err := ParseReturnErrorParameter(code, data)
			if err != nil || got != nil {
				t.Errorf("ParseReturnErrorParameter(%v, %x) = %v, %v; want nil, nil", code, data, got, err)
			}
		}
	}
}

func TestParseReturnErrorParameterUSSDErrorsMalformed(t *testing.T) {
	cases := []struct {
		name string
		code MapErrorCode
		data []byte
	}{
		{"unexpectedSubscriber NULL with content", MapErrorUnexpectedDataValue, []byte{0x30, 0x03, 0x80, 0x01, 0x00}},
		{"unexpectedDataValue truncated", MapErrorUnexpectedDataValue, []byte{0x30, 0x02, 0x80}},
		{"illegalSubscriber not a SEQUENCE", MapErrorIllegalSubscriber, []byte{0x04, 0x00}},
		{"illegalEquipment trailing octets", MapErrorIllegalEquipment, []byte{0x30, 0x00, 0x00}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseReturnErrorParameter(tc.code, tc.data)
			if err == nil {
				t.Errorf("ParseReturnErrorParameter(%v, %x) = %v, want an error", tc.code, tc.data, got)
			}
			if got != nil {
				t.Errorf("ParseReturnErrorParameter(%v, %x) returned %#v together with its error", tc.code, tc.data, got)
			}
		})
	}
}

// A parameter that fails to decode yields a nil any, not an any holding a
// nil pointer of the parameter type, for every dispatched error code.
func TestParseReturnErrorParameterFailureIsUntypedNil(t *testing.T) {
	for code := range dispatchedParamTypes {
		got, err := ParseReturnErrorParameter(code, []byte{0x30, 0x02, 0x80})
		if err == nil {
			t.Errorf("code %v: truncated parameter decoded without an error", code)
		}
		if got != nil {
			t.Errorf("code %v: error came with a non-nil result %#v", code, got)
		}
	}
}

// dispatchedParamTypes is the error code → parameter type table documented
// on ParseReturnErrorParameter.
var dispatchedParamTypes = map[MapErrorCode]reflect.Type{
	MapErrorUnknownSubscriber:             reflect.TypeOf(&UnknownSubscriberParam{}),
	MapErrorAbsentSubscriberSM:            reflect.TypeOf(&AbsentSubscriberSMParam{}),
	MapErrorRoamingNotAllowed:             reflect.TypeOf(&RoamingNotAllowedParam{}),
	MapErrorIllegalSubscriber:             reflect.TypeOf(&IllegalSubscriberParam{}),
	MapErrorTeleserviceNotProvisioned:     reflect.TypeOf(&TeleservNotProvParam{}),
	MapErrorIllegalEquipment:              reflect.TypeOf(&IllegalEquipmentParam{}),
	MapErrorCallBarred:                    reflect.TypeOf(&CallBarredParam{}),
	MapErrorFacilityNotSupported:          reflect.TypeOf(&FacilityNotSupParam{}),
	MapErrorAbsentSubscriber:              reflect.TypeOf(&AbsentSubscriberParam{}),
	MapErrorSystemFailure:                 reflect.TypeOf(&SystemFailureParam{}),
	MapErrorDataMissing:                   reflect.TypeOf(&DataMissingParam{}),
	MapErrorUnexpectedDataValue:           reflect.TypeOf(&UnexpectedDataParam{}),
	MapErrorUnauthorizedRequestingNetwork: reflect.TypeOf(&UnauthorizedRequestingNetworkParam{}),
}

// FuzzParseReturnErrorParameter feeds arbitrary error codes and parameters
// to the dispatcher. It must never panic, a successful result must have the
// documented type for its error code (or be nil for codes without a decoded
// parameter), and a failure must never come with a result.
func FuzzParseReturnErrorParameter(f *testing.F) {
	seeds := []struct {
		code int64
		data string
	}{
		{36, "30028000"}, {36, "3000"}, {36, "300430008000"}, {9, "30023000"}, {12, "3000"},
		{71, ""}, {72, "3000"}, {36, "30028001"}, {36, "3002"}, {34, "0a0102"}, {13, "0a0101"},
		{6, "3000"}, {8, "30030a0100"}, {1, "3080"}, {27, "3000"}, {0, "30"}, {255, "3000"},
	}
	for _, s := range seeds {
		data, err := hex.DecodeString(s.data)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(s.code, data)
	}
	f.Fuzz(func(t *testing.T, code int64, data []byte) {
		errorCode := MapErrorCode(code)
		got, err := ParseReturnErrorParameter(errorCode, data)
		if err != nil {
			if got != nil {
				t.Fatalf("error %v came with a result %T", err, got)
			}
			return
		}
		want, dispatched := dispatchedParamTypes[errorCode]
		switch {
		case got == nil:
			if dispatched && len(data) != 0 {
				t.Fatalf("code %v with %d octets decoded to nil without an error", errorCode, len(data))
			}
		case !dispatched:
			t.Fatalf("code %v is not dispatched but returned %T", errorCode, got)
		case reflect.TypeOf(got) != want:
			t.Fatalf("code %v returned %T, want %v", errorCode, got, want)
		}
	})
}

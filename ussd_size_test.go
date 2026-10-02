// ussd_size_test.go
//
// The SIZE constraints of USSD-Arg and USSD-Res (3GPP TS 29.002 V19.1.0
// §17.7.4, ISDN-AddressString in §17.7.8) are enforced by the go-asn1 BER
// codec on both Marshal and Parse; these tests pin the boundaries.

package gsmmap

import (
	"bytes"
	"errors"
	"testing"

	"github.com/gomaja/go-asn1/runtime/ber"
)

// ussdTLV encodes a BER TLV with a short or one-octet long-form length.
func ussdTLV(tag byte, v []byte) []byte {
	if len(v) < 128 {
		return append([]byte{tag, byte(len(v))}, v...)
	}
	return append([]byte{tag, 0x81, byte(len(v))}, v...)
}

// wantConstraintError checks that err wraps a go-asn1 constraint violation
// of the given field and constraint.
func wantConstraintError(t *testing.T, err error, path, constraint string) {
	t.Helper()
	var ce *ber.ConstraintError
	if !errors.As(err, &ce) {
		t.Fatalf("err = %v, want a *ber.ConstraintError", err)
	}
	if ce.Path != path || ce.Constraint != constraint {
		t.Errorf("violation %q %q, want %q %q", ce.Path, ce.Constraint, path, constraint)
	}
}

func TestUSSDStringSize(t *testing.T) {
	for _, n := range []int{1, 160} {
		str := bytes.Repeat([]byte{0x41}, n)
		data, err := (&USSDArg{DataCodingScheme: USSDDataCodingSchemeGSM7, USSDString: str}).Marshal()
		if err != nil {
			t.Fatalf("USSDArg.Marshal with %d octets: %v", n, err)
		}
		if got, err := ParseUSSDArg(data); err != nil || len(got.USSDString) != n {
			t.Errorf("ParseUSSDArg with %d octets: %v", n, err)
		}
		data, err = (&USSDRes{DataCodingScheme: USSDDataCodingSchemeGSM7, USSDString: str}).Marshal()
		if err != nil {
			t.Fatalf("USSDRes.Marshal with %d octets: %v", n, err)
		}
		if got, err := ParseUSSDRes(data); err != nil || len(got.USSDString) != n {
			t.Errorf("ParseUSSDRes with %d octets: %v", n, err)
		}
	}
	dcs := ussdTLV(0x04, []byte{0x0F})
	for _, n := range []int{0, 161} {
		str := bytes.Repeat([]byte{0x41}, n)
		_, err := (&USSDArg{DataCodingScheme: USSDDataCodingSchemeGSM7, USSDString: str}).Marshal()
		wantConstraintError(t, err, "ussd-String", "SIZE (1..160)")
		_, err = (&USSDRes{DataCodingScheme: USSDDataCodingSchemeGSM7, USSDString: str}).Marshal()
		wantConstraintError(t, err, "ussd-String", "SIZE (1..160)")

		wire := ussdTLV(0x30, append(append([]byte{}, dcs...), ussdTLV(0x04, str)...))
		_, err = ParseUSSDArg(wire)
		wantConstraintError(t, err, "ussd-String", "SIZE (1..160)")
		_, err = ParseUSSDRes(wire)
		wantConstraintError(t, err, "ussd-String", "SIZE (1..160)")
	}
}

func TestUSSDDataCodingSchemeAndAlertingPatternSize(t *testing.T) {
	str := ussdTLV(0x04, []byte{0x31})
	for _, n := range []int{0, 2} {
		wire := ussdTLV(0x30, append(ussdTLV(0x04, bytes.Repeat([]byte{0x0F}, n)), str...))
		_, err := ParseUSSDArg(wire)
		wantConstraintError(t, err, "ussd-DataCodingScheme", "SIZE (1)")
		_, err = ParseUSSDRes(wire)
		wantConstraintError(t, err, "ussd-DataCodingScheme", "SIZE (1)")

		args := append(append(ussdTLV(0x04, []byte{0x0F}), str...), ussdTLV(0x04, bytes.Repeat([]byte{0x01}, n))...)
		_, err = ParseUSSDArg(ussdTLV(0x30, args))
		wantConstraintError(t, err, "alertingPattern", "SIZE (1)")
	}
}

func TestUSSDMSISDNSize(t *testing.T) {
	// 16 digits plus the nature/plan octet: the 9-octet maximum.
	arg := &USSDArg{DataCodingScheme: USSDDataCodingSchemeGSM7, USSDString: HexBytes{0x31},
		MSISDN: "1234567890123456", MSISDNNature: 0x10, MSISDNPlan: 0x01}
	data, err := arg.Marshal()
	if err != nil {
		t.Fatalf("Marshal with a 9-octet MSISDN: %v", err)
	}
	if _, err := ParseUSSDArg(data); err != nil {
		t.Fatalf("ParseUSSDArg with a 9-octet MSISDN: %v", err)
	}

	// 17 digits: 10 octets.
	arg.MSISDN = "12345678901234567"
	_, err = arg.Marshal()
	wantConstraintError(t, err, "msisdn", "SIZE (1..9)")

	head := append(ussdTLV(0x04, []byte{0x0F}), ussdTLV(0x04, []byte{0x31})...)
	wire := ussdTLV(0x30, append(head, ussdTLV(0x80, append([]byte{0x91}, bytes.Repeat([]byte{0x21}, 9)...))...))
	_, err = ParseUSSDArg(wire)
	wantConstraintError(t, err, "msisdn", "SIZE (1..9)")
}

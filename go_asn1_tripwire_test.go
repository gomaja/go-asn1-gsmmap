package gsmmap

import (
	"errors"
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime/ber"
)

// This test pins the go-asn1 behaviour that the README's Encoding bullet
// describes and upstream issue #271 tracks. It fails once go-asn1 changes it:
// then move its wire to TestParseRestrictedOpenValueSentinel and update the
// README. The wires are UpdateLocationRes messages whose private extType
// holds the value; for a payload TLV of P octets the response content is 19+P.

// https://github.com/gomaja/go-asn1/issues/271: go-asn1 refuses a
// GeneralizedTime whose seconds are 60, a leap second, although X.680
// (02/2021) §46.3 takes its time of day from ISO 8601.
func TestGoASN1Issue271LeapSecondRefused(t *testing.T) {
	// 20161231235959Z, accepted; payload 17=2+15, content 36.
	const control = "302404069113265476f8301aa018301606032a0304180f32303136313233313233353935395a"
	// 20161231235960Z; payload 17=2+15, content 36.
	const leap = "302404069113265476f8301aa018301606032a0304180f32303136313233313233353936305a"
	if _, err := ParseUpdateLocationRes(mustHex(t, control)); err != nil {
		t.Fatalf("control Parse = %v, want accepted", err)
	}
	if _, err := ParseUpdateLocationRes(mustHex(t, leap)); !errors.Is(err, ber.ErrInvalidValue) {
		t.Errorf("Parse = %v; go-asn1#271 looks resolved: update the README and TestParseRestrictedOpenValueSentinel", err)
	}
}

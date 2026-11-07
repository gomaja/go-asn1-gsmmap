package gsmmap

import (
	"errors"
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime/ber"
)

// These tests pin go-asn1 behaviour that the README's Encoding bullet
// describes and that upstream issues track. Each fails once go-asn1 changes
// it: then move its wires to TestParseRestrictedOpenValueSentinel and update
// the README. The wires are UpdateLocationRes messages whose private extType
// holds the value; for a payload TLV of P octets the response content is 19+P.

// https://github.com/gomaja/go-asn1/issues/272: X.680 (02/2021) §41.15 limits
// BMPString to the first 64K-2 cells less the surrogates, so U+FFFE and U+FFFF
// are not BMPString characters, yet go-asn1 accepts them in an open value.
func TestGoASN1Issue272BMPNoncharactersAccepted(t *testing.T) {
	for _, wire := range []string{
		"301704069113265476f8300da00b300906032a03041e02fffe", // U+FFFE; payload 4=2+2, content 23
		"301704069113265476f8300da00b300906032a03041e02ffff", // U+FFFF; payload 4=2+2, content 23
	} {
		if _, err := ParseUpdateLocationRes(mustHex(t, wire)); err != nil {
			t.Errorf("Parse(%s) = %v; go-asn1#272 looks fixed: expect ber.ErrInvalidValue and update the README", wire, err)
		}
	}
}

// https://github.com/gomaja/go-asn1/issues/274: X.690 (02/2021) §8.23.9 c)
// forbids SHIFT OUT in a BMPString, and §8.23.5.3 with Table 3 forbids an
// explicit escape sequence in an IA5String, yet go-asn1 accepts both in an
// open value.
func TestGoASN1Issue274EscapeAndShiftAccepted(t *testing.T) {
	for _, wire := range []string{
		"301704069113265476f8300da00b300906032a03041e02000e", // BMPString SHIFT OUT; payload 4=2+2, content 23
		"301704069113265476f8300da00b300906032a030416021b4e", // IA5String ESC N; payload 4=2+2, content 23
	} {
		if _, err := ParseUpdateLocationRes(mustHex(t, wire)); err != nil {
			t.Errorf("Parse(%s) = %v; go-asn1#274 looks fixed: expect ber.ErrInvalidValue and update the README", wire, err)
		}
	}
}

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

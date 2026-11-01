package gsmmap

import (
	"encoding/hex"
	"errors"
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime/ber"
)

// A failed ASN.1 decode in Parse is wrapped in *ber.ElementError and matches
// exactly one ber sentinel: ErrUnexpectedElement for a well-formed element
// the MAP ASN.1 does not admit at its position, and for an EXTERNAL,
// EMBEDDED PDV or CHARACTER STRING in an open value whose explicit [0]
// (single-ASN1-type or identification) holds no value or more than one
// (ITU-T X.690 (02/2021) §8.14.3), otherwise the sentinel of the X.690
// violation. Every input changes one thing in the AnyTimeInterrogationArg of
// TestAnyTimeInterrogationDecodeProfile (3GPP TS 29.002 V19.1.0 §17.7.1):
// [0] subscriberIdentity holding imsi, [1] requestedInfo and
// [3] gsmSCF-Address. The extType inputs add an extensionContainer [2] with
// one private extension, extId 1.2.3.4.
func TestParseDecodeErrorSentinel(t *testing.T) {
	sentinels := []error{
		ber.ErrUnexpectedElement, ber.ErrTruncated, ber.ErrInvalidLength,
		ber.ErrInvalidTag, ber.ErrExtraData, ber.ErrInvalidValue,
	}
	for _, tc := range []struct {
		name, wire string
		want       error
	}{
		{"SET for the SEQUENCE", "3114a00880060001012143f5a1028000830491132654", ber.ErrUnexpectedElement},
		{"missing gsmSCF-Address", "300ea00880060001012143f5a1028000", ber.ErrUnexpectedElement},
		{"gsmSCF-Address before requestedInfo", "3014a00880060001012143f5830491132654a1028000", ber.ErrUnexpectedElement},
		{"empty subscriberIdentity wrapper", "300ca000a1028000830491132654", ber.ErrUnexpectedElement},
		{"two choices in the subscriberIdentity wrapper", "301ca01080060001012143f580060001012143f5a1028000830491132654", ber.ErrUnexpectedElement},
		{"EXTERNAL extType with an empty [0]", "3028a00880060001012143f5a1028000830491132654a212a010300e06032a0304280706032a0304a000", ber.ErrUnexpectedElement},
		{"EXTERNAL extType with two values in [0]", "302ca00880060001012143f5a1028000830491132654a216a014301206032a0304280b06032a0304a00405000500", ber.ErrUnexpectedElement},
		{"EMBEDDED PDV extType with an empty [0]", "3025a00880060001012143f5a1028000830491132654a20fa00d300b06032a03042b04a0008200", ber.ErrUnexpectedElement},
		{"CHARACTER STRING extType with two values in [0]", "302fa00880060001012143f5a1028000830491132654a219a017301506032a03043d0ea00a81032a030481032a03048200", ber.ErrUnexpectedElement},
		{"truncated gsmSCF-Address", "3014a00880060001012143f5a10280008304911326", ber.ErrTruncated},
		{"reserved length octet", "30ffa00880060001012143f5a1028000830491132654", ber.ErrInvalidLength},
		{"primitive subscriberIdentity wrapper", "3014800880060001012143f5a1028000830491132654", ber.ErrInvalidTag},
		{"octets after the argument", "3014a00880060001012143f5a10280008304911326540000", ber.ErrExtraData},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := hex.DecodeString(tc.wire)
			if err != nil {
				t.Fatal(err)
			}
			_, err = ParseAnyTimeInterrogation(data)
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
			for _, s := range sentinels {
				if s != tc.want && errors.Is(err, s) {
					t.Errorf("err = %v also matches %v", err, s)
				}
			}
			var ee *ber.ElementError
			if !errors.As(err, &ee) {
				t.Errorf("err = %v, want a *ber.ElementError", err)
			}
		})
	}
	// With one value in each [0], the same extTypes parse.
	for _, wire := range []string{
		"302aa00880060001012143f5a1028000830491132654a214a012301006032a0304280906032a0304a0020500",
		"302aa00880060001012143f5a1028000830491132654a214a012301006032a03042b09a00581032a03048200",
		"302aa00880060001012143f5a1028000830491132654a214a012301006032a03043d09a00581032a03048200",
	} {
		data, err := hex.DecodeString(wire)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ParseAnyTimeInterrogation(data); err != nil {
			t.Errorf("ParseAnyTimeInterrogation(%s) = %v", wire, err)
		}
	}
}

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
// (ITU-T X.690 (02/2021) §8.14.3) or whose components are not those of its
// associated sequence type, once each and in order (X.680 (02/2021) §§36.5,
// 37.5, 44.5; X.690 §§8.17.1, 8.18.1, 8.24.1); ErrInvalidValue for an
// EXTERNAL without a direct or indirect reference (X.690 §8.18.4) and for
// identification contents X.690 forbids; otherwise the sentinel of the X.690
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
		{"EXTERNAL extType with only a direct reference", "3024a00880060001012143f5a1028000830491132654a20ea00c300a06032a0304280306012a", ber.ErrUnexpectedElement},
		{"EXTERNAL extType with a [0] after its encoding", "302aa00880060001012143f5a1028000830491132654a214a012301006032a0304280906012aa0020500a000", ber.ErrUnexpectedElement},
		{"EMBEDDED PDV extType without a data-value", "3025a00880060001012143f5a1028000830491132654a20fa00d300b06032a03042b04a0028500", ber.ErrUnexpectedElement},
		{"EMBEDDED PDV extType with identification [6]", "3027a00880060001012143f5a1028000830491132654a211a00f300d06032a03042b06a00286008200", ber.ErrUnexpectedElement},
		{"CHARACTER STRING extType with a data-value-descriptor", "302ca00880060001012143f5a1028000830491132654a216a014301206032a03043d0ba00581032a030481008200", ber.ErrUnexpectedElement},
		{"EXTERNAL extType without a direct or indirect reference", "3025a00880060001012143f5a1028000830491132654a20fa00d300b06032a03042804a0020500", ber.ErrInvalidValue},
		{"EMBEDDED PDV extType with a non-minimal presentation-context-id", "3029a00880060001012143f5a1028000830491132654a213a011300f06032a03042b08a004820200018200", ber.ErrInvalidValue},
		{"EMBEDDED PDV extType with an INTEGER segment in its data-value", "302aa00880060001012143f5a1028000830491132654a214a012301006032a03042b09a0028500a203020100", ber.ErrInvalidTag},
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
	// With one value in each [0] and the components of the associated
	// sequence type, the same extTypes parse: an EXTERNAL with a direct
	// reference, or an indirect reference and an octet-aligned encoding; an
	// EMBEDDED PDV with a syntax or syntaxes identification; a CHARACTER
	// STRING with a syntax identification.
	for _, wire := range []string{
		"302aa00880060001012143f5a1028000830491132654a214a012301006032a0304280906032a0304a0020500",
		"3027a00880060001012143f5a1028000830491132654a211a00f300d06032a030428060201018101ab",
		"302aa00880060001012143f5a1028000830491132654a214a012301006032a03042b09a00581032a03048200",
		"302da00880060001012143f5a1028000830491132654a217a015301306032a03042b0ca008a00680012a81012a8200",
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

// Marshal applies the same checks to an opaque PDP-ContextInfo entry (3GPP
// TS 29.002 V19.1.0 §17.7.1) whose private extension, extId 1.2.3.4, holds
// such an extType, and returns the sentinel Parse returns for it in a
// *ber.ElementError. Each entry is also parsed inside an
// AnyTimeInterrogationRes, in ps-PDP-ActiveReachableForPaging.
func TestMarshalOpaqueOpenValueSentinel(t *testing.T) {
	sentinels := []error{
		ber.ErrUnexpectedElement, ber.ErrTruncated, ber.ErrInvalidLength,
		ber.ErrInvalidTag, ber.ErrExtraData, ber.ErrInvalidValue,
	}
	for _, tc := range []struct {
		name, entry, res string
		want             error
	}{
		{"EXTERNAL without a direct or indirect reference",
			"30188001018202f121b10fa00d300b06032a03042804a0020500",
			"3020301ea41ca51a30188001018202f121b10fa00d300b06032a03042804a0020500", ber.ErrInvalidValue},
		{"EMBEDDED PDV without a data-value",
			"30188001018202f121b10fa00d300b06032a03042b04a0028500",
			"3020301ea41ca51a30188001018202f121b10fa00d300b06032a03042b04a0028500", ber.ErrUnexpectedElement},
		{"EMBEDDED PDV with an INTEGER segment in its data-value",
			"301d8001018202f121b114a012301006032a03042b09a0028500a203020100",
			"30253023a421a51f301d8001018202f121b114a012301006032a03042b09a0028500a203020100", ber.ErrInvalidTag},
		{"CHARACTER STRING with a data-value-descriptor",
			"301f8001018202f121b116a014301206032a03043d0ba00581032a030481008200",
			"30273025a423a521301f8001018202f121b116a014301206032a03043d0ba00581032a030481008200", ber.ErrUnexpectedElement},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{PsSubscriberState: &PsSubscriberState{
				PsPDPActiveReachableForPaging: []HexBytes{mustHex(t, tc.entry)},
			}}}
			out, err := in.Marshal()
			if out != nil {
				t.Errorf("Marshal returned %x with an error", out)
			}
			got, perr := ParseAnyTimeInterrogationRes(mustHex(t, tc.res))
			if got != nil {
				t.Errorf("Parse = %+v, want nil", got)
			}
			for _, op := range []struct {
				name string
				err  error
			}{{"Marshal", err}, {"Parse", perr}} {
				if !errors.Is(op.err, tc.want) {
					t.Errorf("%s: err = %v, want %v", op.name, op.err, tc.want)
				}
				for _, s := range sentinels {
					if s != tc.want && errors.Is(op.err, s) {
						t.Errorf("%s: err = %v also matches %v", op.name, op.err, s)
					}
				}
				var ee *ber.ElementError
				if !errors.As(op.err, &ee) {
					t.Errorf("%s: err = %v, want a *ber.ElementError", op.name, op.err)
				}
			}
		})
	}
}

package gsmmap

import (
	"encoding/hex"
	"errors"
	"reflect"
	"strings"
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
		ber.ErrIndefiniteLength, ber.ErrUnsupported, ber.ErrResourceLimit,
		ber.ErrInvalidCodec, ber.ErrInvalidBinding,
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
// such an extType or a restricted string or time with invalid contents, and
// returns the sentinel Parse returns for it in a *ber.ElementError. Each entry is also parsed inside an
// AnyTimeInterrogationRes, in ps-PDP-ActiveReachableForPaging.
func TestMarshalOpaqueOpenValueSentinel(t *testing.T) {
	sentinels := []error{
		ber.ErrUnexpectedElement, ber.ErrTruncated, ber.ErrInvalidLength,
		ber.ErrInvalidTag, ber.ErrExtraData, ber.ErrInvalidValue,
		ber.ErrIndefiniteLength, ber.ErrUnsupported, ber.ErrResourceLimit,
		ber.ErrInvalidCodec, ber.ErrInvalidBinding,
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
		// Restricted string and time contents (X.680 (02/2021) §§41, 47.3;
		// X.690 (02/2021) §§8.23.4, 8.23.8, 8.25), as in
		// TestParseRestrictedOpenValueSentinel's opaque PDP entry.
		{"NumericString outside digits and space",
			"30158001018202f121b10ca00a300806032a0304120141",
			"301d301ba419a51730158001018202f121b10ca00a300806032a0304120141", ber.ErrInvalidValue},
		{"BMPString with a surrogate",
			"30168001018202f121b10da00b300906032a03041e02d800",
			"301e301ca41aa51830168001018202f121b10da00b300906032a03041e02d800", ber.ErrInvalidValue},
		// BMPString payload 4=2+2; entry content 22=18+4, response content 30=26+4.
		{"BMPString U+FFFE",
			"30168001018202f121b10da00b300906032a03041e02fffe",
			"301e301ca41aa51830168001018202f121b10da00b300906032a03041e02fffe", ber.ErrInvalidValue},
		// IA5String payload 5=2+3; entry content 23=18+5, response content 31=26+5.
		{"IA5String designating escape",
			"30178001018202f121b10ea00c300a06032a030416031b2842",
			"301f301da41ba51930178001018202f121b10ea00c300a06032a030416031b2842", ber.ErrInvalidValue},
		// Payload 4=2+2; entry content 22=18+4, response content 30=26+4.
		{"BMPString U+FDD0",
			"30168001018202f121b10da00b300906032a03041e02fdd0",
			"301e301ca41aa51830168001018202f121b10da00b300906032a03041e02fdd0", ber.ErrInvalidValue},
		// Payload 5=2+3; entry content 23=18+5, response content 31=26+5.
		{"UTF8String U+FFFE",
			"30178001018202f121b10ea00c300a06032a03040c03efbfbe",
			"301f301da41ba51930178001018202f121b10ea00c300a06032a03040c03efbfbe", ber.ErrInvalidValue},
		{"UTCTime with an impossible date",
			"301f8001018202f121b116a014301206032a0304170b323630323330303330345a",
			"30273025a423a521301f8001018202f121b116a014301206032a0304170b323630323330303330345a", ber.ErrInvalidValue},
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

// The checked string and time contents in open values follow X.680 (02/2021)
// §41 Tables 8–10, §§41.6, 41.15, 41.16, 46.3, 47.3, X.690 (02/2021)
// §§8.23.3–8.23.10, 8.25 and, for the noncharacters, Unicode 3.2.0 (UAX #27
// D7b). Constructed strings use OCTET STRING segments; a segment may split a
// character's octets or a time spelling. The unchecked controls pin the
// existing behavior without asserting that their contents are valid text.
func TestParseRestrictedOpenValueSentinel(t *testing.T) {
	sentinels := []error{
		ber.ErrUnexpectedElement, ber.ErrTruncated, ber.ErrInvalidLength,
		ber.ErrInvalidTag, ber.ErrExtraData, ber.ErrInvalidValue,
		ber.ErrIndefiniteLength, ber.ErrUnsupported, ber.ErrResourceLimit,
		ber.ErrInvalidCodec, ber.ErrInvalidBinding,
	}
	// For a payload TLV of P octets, the private SEQUENCE content is 5+P
	// (extId is a 5-octet OID TLV), its a0 list is 7+P, and the enclosing
	// extensionContainer is 9+P. The HLR number TLV is 8, so that response
	// content is 19+P. The unknown-addition response content is 10+P.
	// The PDP entry content is 18+P: 7 for its base fields, then 2+2+2+5
	// for b1, a0, private SEQUENCE and OID. Its four enclosing TLVs add 8,
	// giving response content 26+P. Every length below is short form.
	for _, tc := range []struct {
		name, privateExtType, unknownAddition, opaquePDP string
		accept                                           bool
	}{
		// GeneralizedTime admits second 60, a leap second (X.680 (02/2021)
		// §§3.7 h), 46.3 a) 2)); UTCTime seconds stay 00–59 (§47.3 b) 2)).
		// BMPString, UniversalString and UTF8String exclude the 66
		// noncharacters (X.680 §§2, 41.6, 41.15, 41.16; Unicode 3.2.0, UAX #27
		// D7b); their neighbours and U+2FFFD, which is not one, stay valid.
		// Payload 17=2+15; response contents: 36=19+17, 27=10+17, 43=26+17 octets.
		{"GeneralizedTime leap second UTC",
			"302404069113265476f8301aa018301606032a0304180f32303136313233313233353936305a",
			"301b04069113265476f8be11180f32303136313233313233353936305a",
			"302b3029a427a52530238001018202f121b11aa018301606032a0304180f32303136313233313233353936305a", true},
		// Payload 19=2+17; response contents: 38=19+19, 29=10+19, 45=26+19 octets.
		{"GeneralizedTime leap second fraction",
			"302604069113265476f8301ca01a301806032a0304181132303136313233313233353936302e355a",
			"301d04069113265476f8be13181132303136313233313233353936302e355a",
			"302d302ba429a52730258001018202f121b11ca01a301806032a0304181132303136313233313233353936302e355a", true},
		// Payload 16=2+14; response contents: 35=19+16, 26=10+16, 42=26+16 octets.
		{"GeneralizedTime leap second local",
			"302304069113265476f83019a017301506032a0304180e3230313631323331323335393630",
			"301a04069113265476f8be10180e3230313631323331323335393630",
			"302a3028a426a52430228001018202f121b119a017301506032a0304180e3230313631323331323335393630", true},
		// Payload 21=2+19; response contents: 40=19+21, 31=10+21, 47=26+21 octets.
		{"GeneralizedTime leap second time difference",
			"302804069113265476f8301ea01c301a06032a0304181332303137303130313030353936302b30313030",
			"301f04069113265476f8be15181332303137303130313030353936302b30313030",
			"302f302da42ba52930278001018202f121b11ea01c301a06032a0304181332303137303130313030353936302b30313030", true},
		// Segments 15=2+13 and 4=2+2 split seconds 6 | 0 (X.690 §8.23.6).
		// Payload 21=2+19; response contents: 40=19+21, 31=10+21, 47=26+21 octets.
		{"split GeneralizedTime leap second",
			"302804069113265476f8301ea01c301a06032a03043813040d323031363132333132333539360402305a",
			"301f04069113265476f8be153813040d323031363132333132333539360402305a",
			"302f302da42ba52930278001018202f121b11ea01c301a06032a03043813040d323031363132333132333539360402305a", true},
		// Payload 17=2+15; response contents: 36=19+17, 27=10+17, 43=26+17 octets.
		{"GeneralizedTime second 61",
			"302404069113265476f8301aa018301606032a0304180f32303136313233313233353936315a",
			"301b04069113265476f8be11180f32303136313233313233353936315a",
			"302b3029a427a52530238001018202f121b11aa018301606032a0304180f32303136313233313233353936315a", false},
		// Payload 15=2+13; response contents: 34=19+15, 25=10+15, 41=26+15 octets.
		{"UTCTime leap second",
			"302204069113265476f83018a016301406032a0304170d3136313233313233353936305a",
			"301904069113265476f8be0f170d3136313233313233353936305a",
			"30293027a425a52330218001018202f121b118a016301406032a0304170d3136313233313233353936305a", false},
		// Payload 4=2+2; response contents: 23=19+4, 14=10+4, 30=26+4 octets.
		{"BMPString U+FDD0",
			"301704069113265476f8300da00b300906032a03041e02fdd0",
			"300e04069113265476f8be041e02fdd0",
			"301e301ca41aa51830168001018202f121b10da00b300906032a03041e02fdd0", false},
		// Payload 4=2+2; response contents: 23=19+4, 14=10+4, 30=26+4 octets.
		{"BMPString U+FDEF",
			"301704069113265476f8300da00b300906032a03041e02fdef",
			"300e04069113265476f8be041e02fdef",
			"301e301ca41aa51830168001018202f121b10da00b300906032a03041e02fdef", false},
		// Payload 6=2+4; response contents: 25=19+6, 16=10+6, 32=26+6 octets.
		{"UniversalString U+FDD0",
			"301904069113265476f8300fa00d300b06032a03041c040000fdd0",
			"301004069113265476f8be061c040000fdd0",
			"3020301ea41ca51a30188001018202f121b10fa00d300b06032a03041c040000fdd0", false},
		// Payload 6=2+4; response contents: 25=19+6, 16=10+6, 32=26+6 octets.
		{"UniversalString U+1FFFE",
			"301904069113265476f8300fa00d300b06032a03041c040001fffe",
			"301004069113265476f8be061c040001fffe",
			"3020301ea41ca51a30188001018202f121b10fa00d300b06032a03041c040001fffe", false},
		// Payload 6=2+4; response contents: 25=19+6, 16=10+6, 32=26+6 octets.
		{"UniversalString U+10FFFF",
			"301904069113265476f8300fa00d300b06032a03041c040010ffff",
			"301004069113265476f8be061c040010ffff",
			"3020301ea41ca51a30188001018202f121b10fa00d300b06032a03041c040010ffff", false},
		// Payload 5=2+3; response contents: 24=19+5, 15=10+5, 31=26+5 octets.
		{"UTF8String U+FDD0",
			"301804069113265476f8300ea00c300a06032a03040c03efb790",
			"300f04069113265476f8be050c03efb790",
			"301f301da41ba51930178001018202f121b10ea00c300a06032a03040c03efb790", false},
		// Payload 5=2+3; response contents: 24=19+5, 15=10+5, 31=26+5 octets.
		{"UTF8String U+FFFE",
			"301804069113265476f8300ea00c300a06032a03040c03efbfbe",
			"300f04069113265476f8be050c03efbfbe",
			"301f301da41ba51930178001018202f121b10ea00c300a06032a03040c03efbfbe", false},
		// Payload 6=2+4; response contents: 25=19+6, 16=10+6, 32=26+6 octets.
		{"UTF8String U+10FFFF",
			"301904069113265476f8300fa00d300b06032a03040c04f48fbfbf",
			"301004069113265476f8be060c04f48fbfbf",
			"3020301ea41ca51a30188001018202f121b10fa00d300b06032a03040c04f48fbfbf", false},
		// Segments 3=2+1 and 3=2+1 split the two BMP octets (X.690 §8.23.6).
		// Payload 8=2+6; response contents: 27=19+8, 18=10+8, 34=26+8 octets.
		{"split BMPString U+FDD0",
			"301b04069113265476f83011a00f300d06032a03043e060401fd0401d0",
			"301204069113265476f8be083e060401fd0401d0",
			"30223020a41ea51c301a8001018202f121b111a00f300d06032a03043e060401fd0401d0", false},
		// Segments 3=2+1 and 4=2+2 split the UTF-8 octets (X.690 §8.23.6).
		// Payload 9=2+7; response contents: 28=19+9, 19=10+9, 35=26+9 octets.
		{"split UTF8String U+FFFE",
			"301c04069113265476f83012a010300e06032a03042c070401ef0402bfbe",
			"301304069113265476f8be092c070401ef0402bfbe",
			"30233021a41fa51d301b8001018202f121b112a010300e06032a03042c070401ef0402bfbe", false},
		// Payload 4=2+2; response contents: 23=19+4, 14=10+4, 30=26+4 octets.
		{"BMPString U+FDCF",
			"301704069113265476f8300da00b300906032a03041e02fdcf",
			"300e04069113265476f8be041e02fdcf",
			"301e301ca41aa51830168001018202f121b10da00b300906032a03041e02fdcf", true},
		// Payload 4=2+2; response contents: 23=19+4, 14=10+4, 30=26+4 octets.
		{"BMPString U+FDF0",
			"301704069113265476f8300da00b300906032a03041e02fdf0",
			"300e04069113265476f8be041e02fdf0",
			"301e301ca41aa51830168001018202f121b10da00b300906032a03041e02fdf0", true},
		// Payload 6=2+4; response contents: 25=19+6, 16=10+6, 32=26+6 octets.
		{"UniversalString U+1FFFD",
			"301904069113265476f8300fa00d300b06032a03041c040001fffd",
			"301004069113265476f8be061c040001fffd",
			"3020301ea41ca51a30188001018202f121b10fa00d300b06032a03041c040001fffd", true},
		// Payload 5=2+3; response contents: 24=19+5, 15=10+5, 31=26+5 octets.
		{"UTF8String U+FDCF",
			"301804069113265476f8300ea00c300a06032a03040c03efb78f",
			"300f04069113265476f8be050c03efb78f",
			"301f301da41ba51930178001018202f121b10ea00c300a06032a03040c03efb78f", true},
		// Payload 5=2+3; response contents: 24=19+5, 15=10+5, 31=26+5 octets.
		{"UTF8String U+FDF0",
			"301804069113265476f8300ea00c300a06032a03040c03efb7b0",
			"300f04069113265476f8be050c03efb7b0",
			"301f301da41ba51930178001018202f121b10ea00c300a06032a03040c03efb7b0", true},
		// Payload 6=2+4; response contents: 25=19+6, 16=10+6, 32=26+6 octets.
		{"UTF8String U+10FFFD",
			"301904069113265476f8300fa00d300b06032a03040c04f48fbfbd",
			"301004069113265476f8be060c04f48fbfbd",
			"3020301ea41ca51a30188001018202f121b10fa00d300b06032a03040c04f48fbfbd", true},
		// Payload 6=2+4; response contents: 25=19+6, 16=10+6, 32=26+6 octets.
		{"UTF8String U+2FFFD",
			"301904069113265476f8300fa00d300b06032a03040c04f0afbfbd",
			"301004069113265476f8be060c04f0afbfbd",
			"3020301ea41ca51a30188001018202f121b10fa00d300b06032a03040c04f0afbfbd", true},
		// Payload 3=2+1; response contents: 22=19+3, 13=10+3, 29=26+3 octets.
		{"NumericString letter",
			"301604069113265476f8300ca00a300806032a0304120141",
			"300d04069113265476f8be03120141",
			"301d301ba419a51730158001018202f121b10ca00a300806032a0304120141", false},
		// Payload 3=2+1; response contents: 22=19+3, 13=10+3, 29=26+3 octets.
		{"PrintableString at sign",
			"301604069113265476f8300ca00a300806032a0304130140",
			"300d04069113265476f8be03130140",
			"301d301ba419a51730158001018202f121b10ca00a300806032a0304130140", false},
		// Payload 3=2+1; response contents: 22=19+3, 13=10+3, 29=26+3 octets.
		{"IA5String high octet",
			"301604069113265476f8300ca00a300806032a0304160180",
			"300d04069113265476f8be03160180",
			"301d301ba419a51730158001018202f121b10ca00a300806032a0304160180", false},
		// Payload 3=2+1; response contents: 22=19+3, 13=10+3, 29=26+3 octets.
		{"VisibleString control",
			"301604069113265476f8300ca00a300806032a03041a011f",
			"300d04069113265476f8be031a011f",
			"301d301ba419a51730158001018202f121b10ca00a300806032a03041a011f", false},
		// Payload 3=2+1; response contents: 22=19+3, 13=10+3, 29=26+3 octets.
		{"BMPString odd width",
			"301604069113265476f8300ca00a300806032a03041e0100",
			"300d04069113265476f8be031e0100",
			"301d301ba419a51730158001018202f121b10ca00a300806032a03041e0100", false},
		// Payload 4=2+2; response contents: 23=19+4, 14=10+4, 30=26+4 octets.
		{"BMPString surrogate",
			"301704069113265476f8300da00b300906032a03041e02d800",
			"300e04069113265476f8be041e02d800",
			"301e301ca41aa51830168001018202f121b10da00b300906032a03041e02d800", false},
		// Payload 5=2+3; response contents: 24=19+5, 15=10+5, 31=26+5 octets.
		{"UniversalString three octets",
			"301804069113265476f8300ea00c300a06032a03041c03000041",
			"300f04069113265476f8be051c03000041",
			"301f301da41ba51930178001018202f121b10ea00c300a06032a03041c03000041", false},
		// Payload 6=2+4; response contents: 25=19+6, 16=10+6, 32=26+6 octets.
		{"UniversalString surrogate",
			"301904069113265476f8300fa00d300b06032a03041c040000d800",
			"301004069113265476f8be061c040000d800",
			"3020301ea41ca51a30188001018202f121b10fa00d300b06032a03041c040000d800", false},
		// Payload 6=2+4; response contents: 25=19+6, 16=10+6, 32=26+6 octets.
		{"UniversalString above U+10FFFF",
			"301904069113265476f8300fa00d300b06032a03041c0400110000",
			"301004069113265476f8be061c0400110000",
			"3020301ea41ca51a30188001018202f121b10fa00d300b06032a03041c0400110000", false},
		// Payload 12=2+10; response contents: 31=19+12, 22=10+12, 38=26+12 octets.
		{"UTCTime missing zone",
			"301f04069113265476f83015a013301106032a0304170a32363031303230333034",
			"301604069113265476f8be0c170a32363031303230333034",
			"30263024a422a520301e8001018202f121b115a013301106032a0304170a32363031303230333034", false},
		// Payload 13=2+11; response contents: 32=19+13, 23=10+13, 39=26+13 octets.
		{"UTCTime impossible date",
			"302004069113265476f83016a014301206032a0304170b323630323330303330345a",
			"301704069113265476f8be0d170b323630323330303330345a",
			"30273025a423a521301f8001018202f121b116a014301206032a0304170b323630323330303330345a", false},
		// Payload 15=2+13; response contents: 34=19+15, 25=10+15, 41=26+15 octets.
		{"UTCTime second 60",
			"302204069113265476f83018a016301406032a0304170d3236303130323033303436305a",
			"301904069113265476f8be0f170d3236303130323033303436305a",
			"30293027a425a52330218001018202f121b118a016301406032a0304170d3236303130323033303436305a", false},
		// Payload 13=2+11; response contents: 32=19+13, 23=10+13, 39=26+13 octets.
		{"GeneralizedTime empty fraction",
			"302004069113265476f83016a014301206032a0304180b323032363031303230332e",
			"301704069113265476f8be0d180b323032363031303230332e",
			"30273025a423a521301f8001018202f121b116a014301206032a0304180b323032363031303230332e", false},
		// Payload 17=2+15; response contents: 36=19+17, 27=10+17, 43=26+17 octets.
		{"GeneralizedTime impossible date",
			"302404069113265476f8301aa018301606032a0304180f32303236303233303033303430355a",
			"301b04069113265476f8be11180f32303236303233303033303430355a",
			"302b3029a427a52530238001018202f121b11aa018301606032a0304180f32303236303233303033303430355a", false},
		// Two OCTET STRING TLVs occupy 6=3+3 octets, the second holding 'A'.
		// Payload 8=2+6; response contents: 27=19+8, 18=10+8, 34=26+8 octets.
		{"constructed NumericString letter",
			"301b04069113265476f83011a00f300d06032a03043206040131040141",
			"301204069113265476f8be083206040131040141",
			"30223020a41ea51c301a8001018202f121b111a00f300d06032a03043206040131040141", false},
		// Segments d8 and 00 reassemble U+D800; content 6=3+3.
		// Payload 8=2+6; response contents: 27=19+8, 18=10+8, 34=26+8 octets.
		{"split BMPString surrogate",
			"301b04069113265476f83011a00f300d06032a03043e060401d8040100",
			"301204069113265476f8be083e060401d8040100",
			"30223020a41ea51c301a8001018202f121b111a00f300d06032a03043e060401d8040100", false},
		// Segments "26023" and "00304Z" reassemble 2602300304Z, 30 February; content 15=7+8.
		// Payload 17=2+15; response contents: 36=19+17, 27=10+17, 43=26+17 octets.
		{"split UTCTime impossible date",
			"302404069113265476f8301aa018301606032a0304370f04053236303233040630303330345a",
			"301b04069113265476f8be11370f04053236303233040630303330345a",
			"302b3029a427a52530238001018202f121b11aa018301606032a0304370f04053236303233040630303330345a", false},
		// X.680 (02/2021) §41.15; X.690 (02/2021) §§8.23.5.3, 8.23.9–8.23.10.
		// Payload 4=2+2; response contents: 23=19+4, 14=10+4, 30=26+4 octets.
		{"BMPString U+FFFE",
			"301704069113265476f8300da00b300906032a03041e02fffe",
			"300e04069113265476f8be041e02fffe",
			"301e301ca41aa51830168001018202f121b10da00b300906032a03041e02fffe", false},
		// Payload 4=2+2; response contents: 23=19+4, 14=10+4, 30=26+4 octets.
		{"BMPString U+FFFF",
			"301704069113265476f8300da00b300906032a03041e02ffff",
			"300e04069113265476f8be041e02ffff",
			"301e301ca41aa51830168001018202f121b10da00b300906032a03041e02ffff", false},
		// Payload 4=2+2; response contents: 23=19+4, 14=10+4, 30=26+4 octets.
		{"BMPString SO",
			"301704069113265476f8300da00b300906032a03041e02000e",
			"300e04069113265476f8be041e02000e",
			"301e301ca41aa51830168001018202f121b10da00b300906032a03041e02000e", false},
		// Payload 4=2+2; response contents: 23=19+4, 14=10+4, 30=26+4 octets.
		{"BMPString SI",
			"301704069113265476f8300da00b300906032a03041e02000f",
			"300e04069113265476f8be041e02000f",
			"301e301ca41aa51830168001018202f121b10da00b300906032a03041e02000f", false},
		// Payload 4=2+2; response contents: 23=19+4, 14=10+4, 30=26+4 octets.
		{"BMPString SS2",
			"301704069113265476f8300da00b300906032a03041e02008e",
			"300e04069113265476f8be041e02008e",
			"301e301ca41aa51830168001018202f121b10da00b300906032a03041e02008e", false},
		// Payload 4=2+2; response contents: 23=19+4, 14=10+4, 30=26+4 octets.
		{"BMPString SS3",
			"301704069113265476f8300da00b300906032a03041e02008f",
			"300e04069113265476f8be041e02008f",
			"301e301ca41aa51830168001018202f121b10da00b300906032a03041e02008f", false},
		// Payload 8=2+6; response contents: 27=19+8, 18=10+8, 34=26+8 octets.
		{"BMPString designating escape",
			"301b04069113265476f83011a00f300d06032a03041e06001b00280042",
			"301204069113265476f8be081e06001b00280042",
			"30223020a41ea51c301a8001018202f121b111a00f300d06032a03041e06001b00280042", false},
		// Payload 6=2+4; response contents: 25=19+6, 16=10+6, 32=26+6 octets.
		{"UniversalString SI",
			"301904069113265476f8300fa00d300b06032a03041c040000000f",
			"301004069113265476f8be061c040000000f",
			"3020301ea41ca51a30188001018202f121b10fa00d300b06032a03041c040000000f", false},
		// Payload 6=2+4; response contents: 25=19+6, 16=10+6, 32=26+6 octets.
		{"UniversalString SO",
			"301904069113265476f8300fa00d300b06032a03041c040000000e",
			"301004069113265476f8be061c040000000e",
			"3020301ea41ca51a30188001018202f121b10fa00d300b06032a03041c040000000e", false},
		// Payload 6=2+4; response contents: 25=19+6, 16=10+6, 32=26+6 octets.
		{"UniversalString SS2",
			"301904069113265476f8300fa00d300b06032a03041c040000008e",
			"301004069113265476f8be061c040000008e",
			"3020301ea41ca51a30188001018202f121b10fa00d300b06032a03041c040000008e", false},
		// Payload 6=2+4; response contents: 25=19+6, 16=10+6, 32=26+6 octets.
		{"UniversalString SS3",
			"301904069113265476f8300fa00d300b06032a03041c040000008f",
			"301004069113265476f8be061c040000008f",
			"3020301ea41ca51a30188001018202f121b10fa00d300b06032a03041c040000008f", false},
		// Payload 14=2+12; response contents: 33=19+14, 24=10+14, 40=26+14 octets.
		{"UniversalString designating escape",
			"302104069113265476f83017a015301306032a03041c0c0000001b0000002800000042",
			"301804069113265476f8be0e1c0c0000001b0000002800000042",
			"30283026a424a52230208001018202f121b117a015301306032a03041c0c0000001b0000002800000042", false},
		// Payload 5=2+3; response contents: 24=19+5, 15=10+5, 31=26+5 octets.
		{"UTF8String escape sequence",
			"301804069113265476f8300ea00c300a06032a03040c031b2842",
			"300f04069113265476f8be050c031b2842",
			"301f301da41ba51930178001018202f121b10ea00c300a06032a03040c031b2842", false},
		// Payload 3=2+1; response contents: 22=19+3, 13=10+3, 29=26+3 octets.
		{"UTF8String SO",
			"301604069113265476f8300ca00a300806032a03040c010e",
			"300d04069113265476f8be030c010e",
			"301d301ba419a51730158001018202f121b10ca00a300806032a03040c010e", false},
		// Payload 3=2+1; response contents: 22=19+3, 13=10+3, 29=26+3 octets.
		{"UTF8String SI",
			"301604069113265476f8300ca00a300806032a03040c010f",
			"300d04069113265476f8be030c010f",
			"301d301ba419a51730158001018202f121b10ca00a300806032a03040c010f", false},
		// Payload 5=2+3; response contents: 24=19+5, 15=10+5, 31=26+5 octets.
		{"IA5String designating escape",
			"301804069113265476f8300ea00c300a06032a030416031b2842",
			"300f04069113265476f8be0516031b2842",
			"301f301da41ba51930178001018202f121b10ea00c300a06032a030416031b2842", false},
		// Payload 4=2+2; response contents: 23=19+4, 14=10+4, 30=26+4 octets.
		{"IA5String ESC N",
			"301704069113265476f8300da00b300906032a030416021b4e",
			"300e04069113265476f8be0416021b4e",
			"301e301ca41aa51830168001018202f121b10da00b300906032a030416021b4e", false},
		// OCTET STRING segments concatenate before the contents check (X.690 §8.23.6).
		// Payload 8=2+6; response contents: 27=19+8, 18=10+8, 34=26+8 octets.
		{"split BMPString U+FFFE",
			"301b04069113265476f83011a00f300d06032a03043e060401ff0401fe",
			"301204069113265476f8be083e060401ff0401fe",
			"30223020a41ea51c301a8001018202f121b111a00f300d06032a03043e060401ff0401fe", false},
		// OCTET STRING segments concatenate before the contents check (X.690 §8.23.6).
		// Payload 12=2+10; response contents: 31=19+12, 22=10+12, 38=26+12 octets.
		{"split BMPString designating escape",
			"301f04069113265476f83015a013301106032a03043e0a0403001b000403280042",
			"301604069113265476f8be0c3e0a0403001b000403280042",
			"30263024a422a520301e8001018202f121b115a013301106032a03043e0a0403001b000403280042", false},
		// OCTET STRING segments concatenate before the contents check (X.690 §8.23.6).
		// Payload 10=2+8; response contents: 29=19+10, 20=10+10, 36=26+10 octets.
		{"split UniversalString SI",
			"301d04069113265476f83013a011300f06032a03043c08040200000402000f",
			"301404069113265476f8be0a3c08040200000402000f",
			"30243022a420a51e301c8001018202f121b113a011300f06032a03043c08040200000402000f", false},
		// OCTET STRING segments concatenate before the contents check (X.690 §8.23.6).
		// Payload 9=2+7; response contents: 28=19+9, 19=10+9, 35=26+9 octets.
		{"split UTF8String escape sequence",
			"301c04069113265476f83012a010300e06032a03042c0704011b04022842",
			"301304069113265476f8be092c0704011b04022842",
			"30233021a41fa51d301b8001018202f121b112a010300e06032a03042c0704011b04022842", false},
		// OCTET STRING segments concatenate before the contents check (X.690 §8.23.6).
		// Payload 9=2+7; response contents: 28=19+9, 19=10+9, 35=26+9 octets.
		{"split IA5String designating escape",
			"301c04069113265476f83012a010300e06032a0304360704011b04022842",
			"301304069113265476f8be09360704011b04022842",
			"30233021a41fa51d301b8001018202f121b112a010300e06032a0304360704011b04022842", false},
		// Payload 4=2+2; response contents: 23=19+4, 14=10+4, 30=26+4 octets.
		{"BMPString LF",
			"301704069113265476f8300da00b300906032a03041e02000a",
			"300e04069113265476f8be041e02000a",
			"301e301ca41aa51830168001018202f121b10da00b300906032a03041e02000a", true},
		// Payload 4=2+2; response contents: 23=19+4, 14=10+4, 30=26+4 octets.
		{"BMPString U+FFFD",
			"301704069113265476f8300da00b300906032a03041e02fffd",
			"300e04069113265476f8be041e02fffd",
			"301e301ca41aa51830168001018202f121b10da00b300906032a03041e02fffd", true},
		// Payload 3=2+1; response contents: 22=19+3, 13=10+3, 29=26+3 octets.
		{"UTF8String CR",
			"301604069113265476f8300ca00a300806032a03040c010d",
			"300d04069113265476f8be030c010d",
			"301d301ba419a51730158001018202f121b10ca00a300806032a03040c010d", true},
		// Payload 3=2+1; response contents: 22=19+3, 13=10+3, 29=26+3 octets.
		{"IA5String LF",
			"301604069113265476f8300ca00a300806032a030416010a",
			"300d04069113265476f8be0316010a",
			"301d301ba419a51730158001018202f121b10ca00a300806032a030416010a", true},
		// OCTET STRING segments concatenate before the contents check (X.690 §8.23.6).
		// Payload 8=2+6; response contents: 27=19+8, 18=10+8, 34=26+8 octets.
		{"split BMPString LF",
			"301b04069113265476f83011a00f300d06032a03043e0604010004010a",
			"301204069113265476f8be083e0604010004010a",
			"30223020a41ea51c301a8001018202f121b111a00f300d06032a03043e0604010004010a", true},
		// OCTET STRING segments concatenate before the contents check (X.690 §8.23.6).
		// Payload 7=2+5; response contents: 26=19+7, 17=10+7, 33=26+7 octets.
		{"split UTF8String CR",
			"301a04069113265476f83010a00e300c06032a03042c05040004010d",
			"301104069113265476f8be072c05040004010d",
			"3021301fa41da51b30198001018202f121b110a00e300c06032a03042c05040004010d", true},
		// OCTET STRING segments concatenate before the contents check (X.690 §8.23.6).
		// Payload 7=2+5; response contents: 26=19+7, 17=10+7, 33=26+7 octets.
		{"split IA5String LF",
			"301a04069113265476f83010a00e300c06032a03043605040004010a",
			"301104069113265476f8be073605040004010a",
			"3021301fa41da51b30198001018202f121b110a00e300c06032a03043605040004010a", true},
		// Payload 5=2+3; response contents: 24=19+5, 15=10+5, 31=26+5 octets.
		{"valid NumericString",
			"301804069113265476f8300ea00c300a06032a03041203312032",
			"300f04069113265476f8be051203312032",
			"301f301da41ba51930178001018202f121b10ea00c300a06032a03041203312032", true},
		// Payload 5=2+3; response contents: 24=19+5, 15=10+5, 31=26+5 octets.
		{"valid PrintableString",
			"301804069113265476f8300ea00c300a06032a03041303412b3f",
			"300f04069113265476f8be051303412b3f",
			"301f301da41ba51930178001018202f121b10ea00c300a06032a03041303412b3f", true},
		// Payload 5=2+3; response contents: 24=19+5, 15=10+5, 31=26+5 octets.
		{"valid IA5String",
			"301804069113265476f8300ea00c300a06032a0304160300417f",
			"300f04069113265476f8be05160300417f",
			"301f301da41ba51930178001018202f121b10ea00c300a06032a0304160300417f", true},
		// Payload 4=2+2; response contents: 23=19+4, 14=10+4, 30=26+4 octets.
		{"valid VisibleString",
			"301704069113265476f8300da00b300906032a03041a02207e",
			"300e04069113265476f8be041a02207e",
			"301e301ca41aa51830168001018202f121b10da00b300906032a03041a02207e", true},
		// Payload 4=2+2; response contents: 23=19+4, 14=10+4, 30=26+4 octets.
		{"valid BMPString",
			"301704069113265476f8300da00b300906032a03041e020041",
			"300e04069113265476f8be041e020041",
			"301e301ca41aa51830168001018202f121b10da00b300906032a03041e020041", true},
		// Payload 6=2+4; response contents: 25=19+6, 16=10+6, 32=26+6 octets.
		{"valid UniversalString",
			"301904069113265476f8300fa00d300b06032a03041c040001f600",
			"301004069113265476f8be061c040001f600",
			"3020301ea41ca51a30188001018202f121b10fa00d300b06032a03041c040001f600", true},
		// Payload 17=2+15; response contents: 36=19+17, 27=10+17, 43=26+17 octets.
		{"valid UTCTime",
			"302404069113265476f8301aa018301606032a0304170f323630313032303330342b30323030",
			"301b04069113265476f8be11170f323630313032303330342b30323030",
			"302b3029a427a52530238001018202f121b11aa018301606032a0304170f323630313032303330342b30323030", true},
		// Payload 14=2+12; response contents: 33=19+14, 24=10+14, 40=26+14 octets.
		{"valid GeneralizedTime",
			"302104069113265476f83017a015301306032a0304180c323032363031303230332c35",
			"301804069113265476f8be0e180c323032363031303230332c35",
			"30283026a424a52230208001018202f121b117a015301306032a0304180c323032363031303230332c35", true},
		// Two OCTET STRING TLVs occupy 7=3+4 octets.
		// Payload 9=2+7; response contents: 28=19+9, 19=10+9, 35=26+9 octets.
		{"constructed NumericString",
			"301c04069113265476f83012a010300e06032a0304320704013104023220",
			"301304069113265476f8be09320704013104023220",
			"30233021a41fa51d301b8001018202f121b112a010300e06032a0304320704013104023220", true},
		// Each OCTET STRING segment is 3=2+1 octets, so content is 6.
		// Payload 8=2+6; response contents: 27=19+8, 18=10+8, 34=26+8 octets.
		{"split BMPString character",
			"301b04069113265476f83011a00f300d06032a03043e06040100040141",
			"301204069113265476f8be083e06040100040141",
			"30223020a41ea51c301a8001018202f121b111a00f300d06032a03043e06040100040141", true},
		// OCTET STRING segments occupy 15=(2+5)+(2+6) octets.
		// Payload 17=2+15; response contents: 36=19+17, 27=10+17, 43=26+17 octets.
		{"split UTCTime spelling",
			"302404069113265476f8301aa018301606032a0304370f04053236303130040632303330345a",
			"301b04069113265476f8be11370f04053236303130040632303330345a",
			"302b3029a427a52530238001018202f121b11aa018301606032a0304370f04053236303130040632303330345a", true},
		// Payload 4=2+2; response contents: 23=19+4, 14=10+4, 30=26+4 octets.
		{"UTF8String control",
			"301704069113265476f8300da00b300906032a03040c02c3a9",
			"300e04069113265476f8be040c02c3a9",
			"301e301ca41aa51830168001018202f121b10da00b300906032a03040c02c3a9", true},
		// Payload 3=2+1; response contents: 22=19+3, 13=10+3, 29=26+3 octets.
		{"GeneralString unchecked control",
			"301604069113265476f8300ca00a300806032a03041b01ff",
			"300d04069113265476f8be031b01ff",
			"301d301ba419a51730158001018202f121b10ca00a300806032a03041b01ff", true},
		// Payload 3=2+1; response contents: 22=19+3, 13=10+3, 29=26+3 octets.
		{"GraphicString unchecked control",
			"301604069113265476f8300ca00a300806032a03041901ff",
			"300d04069113265476f8be031901ff",
			"301d301ba419a51730158001018202f121b10ca00a300806032a03041901ff", true},
		// Payload 3=2+1; response contents: 22=19+3, 13=10+3, 29=26+3 octets.
		{"TeletexString unchecked control",
			"301604069113265476f8300ca00a300806032a03041401ff",
			"300d04069113265476f8be031401ff",
			"301d301ba419a51730158001018202f121b10ca00a300806032a03041401ff", true},
		// Payload 3=2+1; response contents: 22=19+3, 13=10+3, 29=26+3 octets.
		{"VideotexString unchecked control",
			"301604069113265476f8300ca00a300806032a03041501ff",
			"300d04069113265476f8be031501ff",
			"301d301ba419a51730158001018202f121b10ca00a300806032a03041501ff", true},
		// Payload 3=2+1; response contents: 22=19+3, 13=10+3, 29=26+3 octets.
		{"ObjectDescriptor unchecked control",
			"301604069113265476f8300ca00a300806032a03040701ff",
			"300d04069113265476f8be030701ff",
			"301d301ba419a51730158001018202f121b10ca00a300806032a03040701ff", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, position := range []struct {
				name, wire string
				parse      func([]byte) error
			}{
				{"private extType", tc.privateExtType, func(data []byte) error { _, err := ParseUpdateLocationRes(data); return err }},
				{"unknown addition", tc.unknownAddition, func(data []byte) error { _, err := ParseUpdateLocationRes(data); return err }},
				{"opaque PDP entry", tc.opaquePDP, func(data []byte) error { _, err := ParseAnyTimeInterrogationRes(data); return err }},
			} {
				t.Run(position.name, func(t *testing.T) {
					err := position.parse(mustHex(t, position.wire))
					if tc.accept {
						if err != nil {
							t.Fatalf("Parse = %v, want accepted", err)
						}
						return
					}
					if !errors.Is(err, ber.ErrInvalidValue) {
						t.Fatalf("Parse = %v, want %v", err, ber.ErrInvalidValue)
					}
					for _, sentinel := range sentinels {
						if sentinel != ber.ErrInvalidValue && errors.Is(err, sentinel) {
							t.Errorf("err = %v also matches %v", err, sentinel)
						}
					}
					var element *ber.ElementError
					if !errors.As(err, &element) {
						t.Errorf("err = %v, want a *ber.ElementError", err)
					}
				})
			}
		})
	}
}

// X.690 (02/2021) §8.23.5 prohibits the IA5String high octet. At a
// closed CHOICE, the contents error still names the unclaimed position once
// (go-asn1#273), with its original identifier offset and child indexes.
func TestParseRestrictedContentsPathOnce(t *testing.T) {
	// UpdateLocation base fields occupy 10+8+8=26 octets; pagingArea
	// contains an IA5String TLV of 3 octets, so length is 26+2+3=31.
	const wire = "301f040800010121436587f981069113265476f804069113265476f9ae03160180"
	_, err := ParseUpdateLocation(mustHex(t, wire))
	var element *ber.ElementError
	if !errors.Is(err, ber.ErrInvalidValue) || !errors.As(err, &element) {
		t.Fatalf("Parse = %v, want ErrInvalidValue in *ber.ElementError", err)
	}
	for _, other := range []error{ber.ErrUnexpectedElement, ber.ErrTruncated, ber.ErrInvalidLength, ber.ErrInvalidTag, ber.ErrExtraData} {
		if errors.Is(err, other) {
			t.Errorf("Parse = %v also matches %v", err, other)
		}
	}
	const path = "pagingArea.element[0]"
	if strings.Count(err.Error(), path) != 1 {
		t.Errorf("error = %v, want path %q once", err, path)
	}
	if element.Offset() != 30 || !reflect.DeepEqual(element.ChildIndexes(), []int{3, 0}) {
		t.Errorf("location = %d %v, want 30 [3 0]", element.Offset(), element.ChildIndexes())
	}
}

// ussd.go
//
// Public types for the MAP unstructured supplementary service data (USSD)
// operations:
//
//   - processUnstructuredSS-Request (opCode 59), 3GPP TS 29.002 V19.1.0 §11.9
//   - unstructuredSS-Request (opCode 60), 3GPP TS 29.002 V19.1.0 §11.10
//   - unstructuredSS-Notify (opCode 61), 3GPP TS 29.002 V19.1.0 §11.11
//
// The ASN.1 operations are in 3GPP TS 29.002 V19.1.0 §17.6.4 and the
// USSD-Arg / USSD-Res types in §17.7.4. Marshal / Parse entry points live in
// marshal.go / parse.go; the wire converters in convert_ussd.go.

package gsmmap

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-sms/encoding/cbs"
)

// Sentinel errors of the USSD codec. All are wrapped with context and are
// matchable with errors.Is.
var (
	// ErrUSSDArgNil is returned when a nil USSDArg is marshalled or converted.
	ErrUSSDArgNil = errors.New("ussd: nil USSDArg is not permitted")
	// ErrUSSDResNil is returned when a nil USSDRes is marshalled or converted.
	ErrUSSDResNil = errors.New("ussd: nil USSDRes is not permitted")
	// ErrUSSDUnsupportedDataCodingScheme is returned when a USSD data coding
	// scheme is not one this package can decode or encode text with.
	ErrUSSDUnsupportedDataCodingScheme = errors.New("ussd: unsupported data coding scheme")
	// ErrUSSDTextTooLong is returned when encoded text exceeds
	// maxUSSD-StringLength (160 octets), 3GPP TS 29.002 V19.1.0 §17.7.4.
	ErrUSSDTextTooLong = errors.New("ussd: encoded text exceeds 160 octets")
	// ErrUSSDTextEmpty is returned when the text to encode, or the octets to
	// decode, are empty: USSD-String is SIZE (1..maxUSSD-StringLength).
	ErrUSSDTextEmpty   = errors.New("ussd: empty text")
	ErrUSSDInvalidUTF8 = errors.New("ussd: text is not valid UTF-8")
	// ErrAlertingPatternReserved is returned when marshalling an
	// AlertingPattern that is not one of the eight values 3GPP TS 29.002
	// V19.1.0 §17.7.8 defines.
	ErrAlertingPatternReserved = errors.New("alertingPattern: reserved value")
	// ErrUSSDMSISDNDecodedEmpty is returned when a wire MSISDN is present but
	// carries no digits; presence cannot round-trip through the string API.
	ErrUSSDMSISDNDecodedEmpty = errors.New("ussd: present wire MSISDN decoded to empty digits; presence cannot round-trip through string-based API")
)

// USSDDataCodingScheme is the USSD-DataCodingScheme, 3GPP TS 29.002 V19.1.0
// §7.6.4.36 and §17.7.4. It is coded as the Cell Broadcast Data Coding Scheme
// of 3GPP TS 23.038 V20.0.0 §5, and is one octet on the wire.
type USSDDataCodingScheme uint8

const (
	// USSDDataCodingSchemeGSM7 is the GSM 7 bit default alphabet, language
	// unspecified (3GPP TS 23.038 V20.0.0 §5, coding group 0000, bits 3..0 =
	// 1111).
	USSDDataCodingSchemeGSM7 USSDDataCodingScheme = 0x0F
	// USSDDataCodingSchemeUCS2 is general data coding, uncompressed, no
	// message class, UCS2 (3GPP TS 23.038 V20.0.0 §5, coding group 01xx,
	// bit 5 = 0, bit 4 = 0, bits 3..2 = 10).
	USSDDataCodingSchemeUCS2 USSDDataCodingScheme = 0x48
)

// USSDText is a USSD-String decoded by DecodeText, with the facts its data
// coding scheme gives (3GPP TS 23.038 V20.0.0 §5).
type USSDText struct {
	// Text is the decoded text, without the language indication.
	Text string
	// Language is the language the coding indicates: for 0x10 and 0x11 the
	// two characters that precede the text, as received; for 0x12 its three
	// lower-case letters; for groups 0000 and 0010 the ISO 639-1 code that
	// bits 3..0 name (0x01 is "en"). It is empty for language unspecified,
	// reserved codings and the other groups.
	Language string
	// Charset is the character set the text was decoded with, cbs.CharsetGSM7
	// or cbs.CharsetUCS2, after the §5 rule that a receiving entity assumes
	// the GSM 7 bit default alphabet for any reserved coding.
	Charset cbs.Charset
}

// Decode converts the octets of a USSD-String, coded per d, into text.
// It returns DecodeText's Text, or an empty string on error; use DecodeText
// for the language and effective character set of 3GPP TS 23.038 V20.0.0 §5.
func (d USSDDataCodingScheme) Decode(s []byte) (string, error) {
	text, err := d.DecodeText(s)
	return text.Text, err
}

// DecodeText converts the octets of a USSD-String, coded per d, into text,
// language and character set. Classification and decoding are delegated to
// go-sms encoding/cbs (cbs.DCS.DecodeUSSD), which applies the Cell Broadcast
// Data Coding Scheme of 3GPP TS 23.038 V20.0.0 §5: groups 0000 and 0010 name
// a language, the group 0001 codings 0x10..0x12 precede the text with a
// language indication (0x12 was added in V20.0.0 by CR 0243), and "Any
// reserved codings shall be assumed to be the GSM 7 bit default alphabet (the
// same as codepoint 00001111) by a receiving entity." GSM 7 bit text uses the
// USSD packing of §6.1.2.3.1 (including the removal of a final <CR> on an
// octet boundary) and the default alphabet and its extension table (§6.2.1,
// §6.2.1.1). UCS2 text is mapped per §6.2.3.
//
// Compressed text, 8 bit data, a user data header, the I1 protocol and WAP are
// not supported: their errors match both ErrUSSDUnsupportedDataCodingScheme
// and cbs.ErrUnsupported, even for an empty s. Otherwise an empty s yields
// ErrUSSDTextEmpty. A malformed language indication matches
// cbs.ErrLanguageIndication, and the errors of the GSM 7 bit and UCS2 decoders
// stay matchable (for example ucs2.ErrInvalidLength). On error the USSDText is
// zero.
func (d USSDDataCodingScheme) DecodeText(s []byte) (USSDText, error) {
	coding := cbs.DCS(d)
	text, language, err := coding.DecodeUSSD(s)
	if errors.Is(err, cbs.ErrUnsupported) {
		return USSDText{}, fmt.Errorf("decoding USSD-String (data coding scheme 0x%02X): %w: %w", uint8(d), ErrUSSDUnsupportedDataCodingScheme, err)
	}
	if len(s) == 0 {
		return USSDText{}, fmt.Errorf("decoding USSD-String (data coding scheme 0x%02X): %w", uint8(d), ErrUSSDTextEmpty)
	}
	if err != nil {
		return USSDText{}, fmt.Errorf("decoding USSD-String (data coding scheme 0x%02X): %w", uint8(d), err)
	}
	if language == "" {
		language, _ = coding.Language()
	}
	return USSDText{Text: text, Language: language, Charset: coding.Charset()}, nil
}

// Encode converts text into the octets of a USSD-String, coded per d.
// Classification and encoding are delegated to go-sms encoding/cbs
// (cbs.DCS.EncodeUSSD). Only the GSM 7 bit default alphabet and UCS2 codings
// of 3GPP TS 23.038 V20.0.0 §5 without a language indication are accepted.
// §5 gives a reserved coding a meaning only for a receiving entity, so a
// reserved coding yields an error matching both
// ErrUSSDUnsupportedDataCodingScheme and cbs.ErrReserved; an unsupported coding
// yields one matching both ErrUSSDUnsupportedDataCodingScheme and
// cbs.ErrUnsupported. The language indication codings 0x10..0x12 yield
// ErrUSSDUnsupportedDataCodingScheme, because Encode carries no language;
// cbs.DCS.EncodeUSSD encodes one. Coding errors take precedence over empty
// text.
//
// GSM 7 bit text uses the default alphabet and extension table (§6.2.1,
// §6.2.1.1) with USSD packing (§6.1.2.3.1). UCS2 uses §6.2.3's 16 bits per
// character; characters above U+FFFF are rejected. Invalid UTF-8 with a UCS2
// coding yields ErrUSSDInvalidUTF8. Empty text yields ErrUSSDTextEmpty, and
// a result longer than maxUSSD-StringLength (160 octets, 3GPP TS 29.002
// V19.1.0 §17.7.4; 182 septets per TS 23.038 §6.1.2.3.1) yields
// ErrUSSDTextTooLong. Alphabet errors remain wrapped.
func (d USSDDataCodingScheme) Encode(text string) ([]byte, error) {
	coding := cbs.DCS(d)
	out, err := coding.EncodeUSSD(text, "")
	if errors.Is(err, cbs.ErrReserved) || errors.Is(err, cbs.ErrUnsupported) {
		return nil, fmt.Errorf("encoding USSD-String (data coding scheme 0x%02X): %w: %w", uint8(d), ErrUSSDUnsupportedDataCodingScheme, err)
	}
	if coding.LanguageIndication() != cbs.NoLanguageIndication {
		return nil, fmt.Errorf("data coding scheme 0x%02X (Encode carries no language; cbs.DCS.EncodeUSSD encodes one): %w", uint8(d), ErrUSSDUnsupportedDataCodingScheme)
	}
	if text == "" {
		return nil, fmt.Errorf("encoding USSD-String (data coding scheme 0x%02X): %w", uint8(d), ErrUSSDTextEmpty)
	}
	// []rune would turn invalid UTF-8 into U+FFFD, which UCS2 can code.
	if coding.Charset() == cbs.CharsetUCS2 && !utf8.ValidString(text) {
		return nil, fmt.Errorf("encoding UCS2 USSD-String: %w", ErrUSSDInvalidUTF8)
	}
	if err != nil {
		return nil, fmt.Errorf("encoding USSD-String (data coding scheme 0x%02X): %w", uint8(d), err)
	}
	if len(out) > int(gsm_map.MaxUSSDStringLength) {
		return nil, fmt.Errorf("encoded USSD-String is %d octets, maximum %d: %w", len(out), int(gsm_map.MaxUSSDStringLength), ErrUSSDTextTooLong)
	}
	return out, nil
}

// AlertingPattern is the AlertingPattern of 3GPP TS 29.002 V19.1.0 §17.7.8
// (see also §7.6.3.44): bits 4..3 give the type of pattern (00 level, 01 or 10
// category) and bits 2..1 the type of alerting.
type AlertingPattern uint8

// The eight values 3GPP TS 29.002 V19.1.0 §17.7.8 defines. All other values
// are reserved.
const (
	AlertingLevel0    AlertingPattern = 0x00
	AlertingLevel1    AlertingPattern = 0x01
	AlertingLevel2    AlertingPattern = 0x02
	AlertingCategory1 AlertingPattern = 0x04
	AlertingCategory2 AlertingPattern = 0x05
	AlertingCategory3 AlertingPattern = 0x06
	AlertingCategory4 AlertingPattern = 0x07
	AlertingCategory5 AlertingPattern = 0x08
)

// String returns the ASN.1 value name of p from 3GPP TS 29.002 V19.1.0
// §17.7.8, such as "alertingLevel-0" or "alertingCategory-4", or
// "reserved(0xNN)" for a reserved value.
func (p AlertingPattern) String() string {
	switch p {
	case AlertingLevel0:
		return "alertingLevel-0"
	case AlertingLevel1:
		return "alertingLevel-1"
	case AlertingLevel2:
		return "alertingLevel-2"
	case AlertingCategory1:
		return "alertingCategory-1"
	case AlertingCategory2:
		return "alertingCategory-2"
	case AlertingCategory3:
		return "alertingCategory-3"
	case AlertingCategory4:
		return "alertingCategory-4"
	case AlertingCategory5:
		return "alertingCategory-5"
	default:
		return fmt.Sprintf("reserved(0x%02X)", uint8(p))
	}
}

// USSDArg is USSD-Arg, 3GPP TS 29.002 V19.1.0 §17.7.4. It is the argument of
// processUnstructuredSS-Request (opCode 59, §11.9), unstructuredSS-Request
// (opCode 60, §11.10) and unstructuredSS-Notify (opCode 61, §11.11).
type USSDArg struct {
	// DataCodingScheme is ussd-DataCodingScheme (§7.6.4.36).
	DataCodingScheme USSDDataCodingScheme
	// USSDString is ussd-String (§7.6.4.37), 1..160 octets coded per
	// DataCodingScheme; use DataCodingScheme.Encode / Decode.
	USSDString HexBytes
	// AlertingPattern is alertingPattern (§7.6.3.44), used by
	// unstructuredSS-Request and unstructuredSS-Notify only (§11.10, §11.11).
	// nil means absent.
	AlertingPattern *AlertingPattern
	// MSISDN is the digits of msisdn [0] ISDN-AddressString (§7.6.2.17), used
	// by processUnstructuredSS-Request only (§11.9). "" means absent.
	MSISDN string
	// MSISDNNature is the nature of address of MSISDN (address.Nature*);
	// 0 is unknown.
	MSISDNNature uint8
	// MSISDNPlan is the numbering plan of MSISDN (address.Plan*); 0 is
	// unknown.
	MSISDNPlan uint8
}

// USSDRes is USSD-Res, 3GPP TS 29.002 V19.1.0 §17.7.4. It is the result of
// processUnstructuredSS-Request (opCode 59) and unstructuredSS-Request
// (opCode 60).
type USSDRes struct {
	// DataCodingScheme is ussd-DataCodingScheme (§7.6.4.36).
	DataCodingScheme USSDDataCodingScheme
	// USSDString is ussd-String (§7.6.4.37), 1..160 octets coded per
	// DataCodingScheme.
	USSDString HexBytes
}

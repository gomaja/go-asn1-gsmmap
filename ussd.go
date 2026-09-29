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
	"github.com/gomaja/go-sms/encoding/gsm7"
	"github.com/gomaja/go-sms/encoding/ucs2"
)

// maxUSSDStringLength is maxUSSD-StringLength, 3GPP TS 29.002 V19.1.0 §17.7.4.
const maxUSSDStringLength = 160

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
	ErrUSSDTextEmpty = errors.New("ussd: empty text")
	// ErrAlertingPatternInvalidSize is returned when a wire AlertingPattern is
	// not exactly one octet (SIZE (1), 3GPP TS 29.002 V19.1.0 §17.7.8).
	ErrAlertingPatternInvalidSize = errors.New("alertingPattern: must be exactly 1 octet")
	// ErrAlertingPatternReserved is returned when marshalling an
	// AlertingPattern that is not one of the seven values 3GPP TS 29.002
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

// ussdAlphabet is how a USSDDataCodingScheme codes the USSD-String.
type ussdAlphabet int

const (
	// ussdAlphabetGSM7 is an explicit GSM 7 bit default alphabet coding.
	ussdAlphabetGSM7 ussdAlphabet = iota
	// ussdAlphabetGSM7Reserved is a reserved coding, which a receiving entity
	// assumes to be the GSM 7 bit default alphabet. It is never sent.
	ussdAlphabetGSM7Reserved
	// ussdAlphabetUCS2 is an explicit UCS2 coding.
	ussdAlphabetUCS2
	// ussdAlphabetUnsupported is a coding this package does not handle.
	ussdAlphabetUnsupported
)

// classify sorts d by the coding group table of 3GPP TS 23.038 V20.0.0 §5.
// For ussdAlphabetUnsupported and ussdAlphabetGSM7Reserved the returned
// string names the reason.
func (d USSDDataCodingScheme) classify() (ussdAlphabet, string) {
	group := uint8(d) >> 4
	low := uint8(d) & 0x0F
	switch group {
	case 0x0: // language using the GSM 7 bit default alphabet
		return ussdAlphabetGSM7, ""
	case 0x1:
		switch low {
		case 0x0:
			return ussdAlphabetUnsupported, "GSM 7 bit default alphabet with language indication"
		case 0x1:
			return ussdAlphabetUnsupported, "UCS2 with language indication"
		case 0x2:
			// Added to TS 23.038 in V20.0.0 by CR 0243.
			return ussdAlphabetUnsupported, "UCS2 with three-letter language indication"
		default:
			return ussdAlphabetGSM7Reserved, "reserved coding in group 0001"
		}
	case 0x2:
		if low <= 0x4 { // Czech, Hebrew, Arabic, Russian, Icelandic
			return ussdAlphabetGSM7, ""
		}
		return ussdAlphabetGSM7Reserved, "reserved for other languages in group 0010"
	case 0x3:
		return ussdAlphabetGSM7Reserved, "reserved for other languages in group 0011"
	case 0x4, 0x5, 0x6, 0x7: // general data coding indication
		if d&0x20 != 0 {
			return ussdAlphabetUnsupported, "compressed text (3GPP TS 23.042)"
		}
		switch (uint8(d) >> 2) & 0x03 {
		case 0x0:
			return ussdAlphabetGSM7, ""
		case 0x1:
			return ussdAlphabetUnsupported, "8 bit data"
		case 0x2:
			return ussdAlphabetUCS2, ""
		default:
			return ussdAlphabetGSM7Reserved, "reserved character set in general data coding"
		}
	case 0x8, 0xA, 0xB, 0xC:
		return ussdAlphabetGSM7Reserved, "reserved coding group"
	case 0x9:
		return ussdAlphabetUnsupported, "message with user data header structure"
	case 0xD:
		return ussdAlphabetUnsupported, "I1 protocol message (3GPP TS 24.294)"
	case 0xE:
		return ussdAlphabetUnsupported, "WAP Forum defined"
	default: // 0xF: data coding / message handling
		if d&0x08 != 0 {
			return ussdAlphabetGSM7Reserved, "reserved bit 3 set in group 1111"
		}
		if d&0x04 != 0 {
			return ussdAlphabetUnsupported, "8 bit data"
		}
		return ussdAlphabetGSM7, ""
	}
}

func (d USSDDataCodingScheme) unsupported(reason string) error {
	return fmt.Errorf("data coding scheme 0x%02X (%s): %w", uint8(d), reason, ErrUSSDUnsupportedDataCodingScheme)
}

// Decode converts the octets of a USSD-String, coded per d, into text.
//
// The coding is classified by the Cell Broadcast Data Coding Scheme table of
// 3GPP TS 23.038 V20.0.0 §5. The GSM 7 bit default alphabet codings are
// unpacked with the USSD rules of §6.1.2.3.1 (including the removal of a final
// <CR> on an octet boundary) and mapped with the default alphabet and its
// extension table (§6.2.1, §6.2.1.1). UCS2 codings are mapped per §6.2.3.
//
// A reserved coding is decoded as the GSM 7 bit default alphabet, as §5
// requires of a receiving entity: "Any reserved codings shall be assumed to be
// the GSM 7 bit default alphabet (the same as codepoint 00001111) by a
// receiving entity." Codings with a language indication, compressed text, 8 bit
// data, a user data header, the I1 protocol and WAP are not supported and yield
// ErrUSSDUnsupportedDataCodingScheme. An empty s yields ErrUSSDTextEmpty.
func (d USSDDataCodingScheme) Decode(s []byte) (string, error) {
	alphabet, reason := d.classify()
	if alphabet == ussdAlphabetUnsupported {
		return "", d.unsupported(reason)
	}
	if len(s) == 0 {
		return "", fmt.Errorf("decoding USSD-String (data coding scheme 0x%02X): %w", uint8(d), ErrUSSDTextEmpty)
	}
	if alphabet == ussdAlphabetUCS2 {
		runes, err := ucs2.Decode(s)
		if err != nil {
			return "", fmt.Errorf("decoding UCS2 USSD-String: %w", err)
		}
		return string(runes), nil
	}
	text, err := gsm7.Decode(gsm7.Unpack7BitUSSD(s))
	if err != nil {
		return "", fmt.Errorf("decoding GSM 7 bit USSD-String: %w", err)
	}
	return string(text), nil
}

// Encode converts text into the octets of a USSD-String, coded per d.
//
// Only the explicit GSM 7 bit default alphabet codings and the explicit UCS2
// codings of 3GPP TS 23.038 V20.0.0 §5 are accepted; a reserved coding must not
// be used by a sending entity, and the codings Decode does not support are not
// supported here either (ErrUSSDUnsupportedDataCodingScheme). GSM 7 bit text is
// mapped with the default alphabet and its extension table (§6.2.1, §6.2.1.1)
// and packed with the USSD rules of §6.1.2.3.1. UCS2 text is mapped per §6.2.3;
// UCS2 has 16 bits per character, so a character above U+FFFF, which would need
// a UTF-16 surrogate pair, is rejected. Text that cannot be represented is an
// error, empty text yields ErrUSSDTextEmpty, and a result longer than
// maxUSSD-StringLength (160 octets, 3GPP TS 29.002 V19.1.0 §17.7.4; 182 septets
// per §6.1.2.3.1) yields ErrUSSDTextTooLong.
func (d USSDDataCodingScheme) Encode(text string) ([]byte, error) {
	alphabet, reason := d.classify()
	switch alphabet {
	case ussdAlphabetUnsupported:
		return nil, d.unsupported(reason)
	case ussdAlphabetGSM7Reserved:
		return nil, d.unsupported(reason + "; a sending entity must not use a reserved coding (3GPP TS 23.038 V20.0.0 §5)")
	}
	if text == "" {
		return nil, fmt.Errorf("encoding USSD-String (data coding scheme 0x%02X): %w", uint8(d), ErrUSSDTextEmpty)
	}

	var out []byte
	if alphabet == ussdAlphabetUCS2 {
		if !utf8.ValidString(text) {
			return nil, errors.New("encoding UCS2 USSD-String: text is not valid UTF-8")
		}
		runes := []rune(text)
		for i, r := range runes {
			if r > 0xFFFF {
				return nil, fmt.Errorf("encoding UCS2 USSD-String: character %U at index %d is above U+FFFF and cannot be coded in 16 bits", r, i)
			}
		}
		out = ucs2.Encode(runes)
	} else {
		septets, err := gsm7.Encode([]byte(text))
		if err != nil {
			return nil, fmt.Errorf("encoding GSM 7 bit USSD-String: %w", err)
		}
		out, err = gsm7.Pack7BitUSSD(septets)
		if err != nil {
			return nil, fmt.Errorf("packing GSM 7 bit USSD-String: %w", err)
		}
	}
	if len(out) > maxUSSDStringLength {
		return nil, fmt.Errorf("encoded USSD-String is %d octets, maximum %d: %w", len(out), maxUSSDStringLength, ErrUSSDTextTooLong)
	}
	return out, nil
}

// AlertingPattern is the AlertingPattern of 3GPP TS 29.002 V19.1.0 §17.7.8
// (see also §7.6.3.44): bits 4..3 give the type of pattern (00 level, 01 or 10
// category) and bits 2..1 the type of alerting.
type AlertingPattern uint8

// The seven values 3GPP TS 29.002 V19.1.0 §17.7.8 defines. All other values
// are reserved.
const (
	AlertingLevel0    AlertingPattern = 0x00
	AlertingLevel1    AlertingPattern = 0x01
	AlertingLevel2    AlertingPattern = 0x02
	AlertingCategory1 AlertingPattern = 0x04
	AlertingCategory2 AlertingPattern = 0x05
	AlertingCategory3 AlertingPattern = 0x06
	AlertingCategory4 AlertingPattern = 0x07
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

// NetworkUnstructuredSsContextV2 returns the OBJECT IDENTIFIER arcs of
// networkUnstructuredSsContext-v2, {map-ac networkUnstructuredSs(19)
// version2(2)} = 0.4.0.0.1.0.19.2 (3GPP TS 29.002 V19.1.0 §17.3.2.20), the
// application context of all three USSD operations. Each call returns a new
// slice.
func NetworkUnstructuredSsContextV2() []uint64 {
	return append([]uint64(nil), gsm_map.NetworkUnstructuredSsContextV2()...)
}

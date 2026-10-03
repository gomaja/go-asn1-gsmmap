package gsmmap

import (
	"fmt"
	"math"
	"math/big"

	"github.com/gomaja/go-asn1-gsmmap/address"
	"github.com/gomaja/go-asn1-gsmmap/tbcd"
)

const errEncodingIMSI = "encoding IMSI: %w"

// encodeAddressField encodes digits into an AddressString (3GPP TS 29.002
// V19.1.0 §17.7.8) with the given nature of address (address.Nature*,
// bits 7..5) and numbering plan (address.Plan*, bits 4..1). Zero is
// address.NatureUnknown / address.PlanUnknown, exactly as on the wire, so a
// decoded address encodes back to the same octets.
func encodeAddressField(digits string, nature, plan uint8) ([]byte, error) {
	if nature&^0b01110000 != 0 {
		return nil, fmt.Errorf("nature of address 0x%02X: %w", nature, ErrAddressNatureInvalid)
	}
	if plan&^0b00001111 != 0 {
		return nil, fmt.Errorf("numbering plan 0x%02X: %w", plan, ErrAddressPlanInvalid)
	}
	tbcdBytes, err := tbcd.Encode(digits)
	if err != nil {
		return nil, err
	}
	return address.Encode(address.ExtensionNo, nature, plan, tbcdBytes), nil
}

// decodeAddressField decodes an AddressString byte slice into a phone number string and address components.
func decodeAddressField(encoded []byte) (digits string, nature, plan uint8, err error) {
	// A zero-octet AddressString has no nature/plan octet and violates
	// SIZE (1..9). address.Decode yields nil digits for it, which the TBCD
	// decoder used to reject as a nil input; tbcd.Decode now treats nil as an
	// empty string, so the check lives here.
	if len(encoded) == 0 {
		return "", 0, 0, ErrAddressStringEmpty
	}
	_, nat, pl, rawDigits := address.Decode(encoded)
	digits, err = tbcd.Decode(rawDigits)
	if err != nil {
		return "", 0, 0, err
	}
	return digits, nat, pl, nil
}

// boolToNullPtr converts a Go bool into the ASN.1 NULL pointer convention
// used by go-asn1: nil means "absent", non-nil means "present".
func boolToNullPtr(b bool) *struct{} {
	if !b {
		return nil
	}
	v := struct{}{}
	return &v
}

// nullPtrToBool is the inverse of boolToNullPtr.
func nullPtrToBool(p *struct{}) bool { return p != nil }

// intPtrTo64 narrows a *int public-type field to the *int64 wire-type form.
func intPtrTo64(p *int) *int64 {
	if p == nil {
		return nil
	}
	v := int64(*p)
	return &v
}

// int64PtrTo narrows a *int64 wire-type field to the *int public-type form.
// Rejects values outside [math.MinInt, math.MaxInt] (which collapses to
// [math.MinInt32, math.MaxInt32] on 32-bit platforms and is a no-op on
// 64-bit) rather than silently truncating.
func int64PtrTo(p *int64) (*int, error) {
	if p == nil {
		return nil, nil
	}
	v, err := narrowInt64(*p)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// narrowInt64 narrows an int64 to Go's platform int, rejecting values that
// would truncate on 32-bit builds. On 64-bit platforms the bounds are a
// no-op because int == int64.
func narrowInt64(v int64) (int, error) {
	if v < math.MinInt || v > math.MaxInt {
		return 0, fmt.Errorf("value %d does not fit in Go int on this platform", v)
	}
	return int(v), nil
}

// narrowInt64Range is like narrowInt64 but additionally enforces an
// application-defined inclusive range [lo, hi]. Callers pass a field
// name for inclusion in the error message. Delegates to narrowInt64
// after the range check so callers passing a [lo, hi] that exceeds the
// platform int bounds still get the overflow safeguard.
func narrowInt64Range(v int64, lo, hi int64, field string) (int, error) {
	if v < lo || v > hi {
		return 0, fmt.Errorf("%s out of range %d..%d: %d", field, lo, hi, v)
	}
	return narrowInt64(v)
}

func bigIntFromInt64(v int64) *big.Int {
	return big.NewInt(v)
}

func int64FromBigInt(v *big.Int, field string) (int64, error) {
	if v == nil {
		return 0, fmt.Errorf("%s is nil", field)
	}
	if !v.IsInt64() {
		return 0, fmt.Errorf("%s overflows int64: %s", field, v.String())
	}
	return v.Int64(), nil
}

// validateAPN checks the APN element of a LIPAAllowedAPNList. The codec does
// not yet enforce SEQUENCE OF element SIZE constraints:
// https://github.com/gomaja/go-asn1/issues/79 (TS 29.002 §17.7.1).
func validateAPN(b HexBytes, field string) error {
	if len(b) < 2 || len(b) > 63 {
		return fmt.Errorf("%s: %w (got %d)", field, ErrAPNInvalidSize, len(b))
	}
	return nil
}

// identity is an identity of 3GPP TS 23.003 carried as a TBCD-STRING (3GPP
// TS 29.002 V19.1.0 §17.7.8) and the digit counts the MAP field carrying it
// admits. The octet SIZE the codec checks admits more digit counts than the
// identity has, so the count is checked here, identically on encode and
// decode.
type identity struct {
	min, max int
	err      error
}

var (
	// identityIMSI: 3GPP TS 23.003 V20.1.0 §2.2 "IMSI is composed of three
	// parts: 1) Mobile Country Code (MCC) consisting of three digits. [...]
	// 2) Mobile Network Code (MNC) consisting of two or three digits [...]
	// 3) Mobile Subscriber Identification Number (MSIN)", "Not more than 15
	// digits" (figure 1); §2.3 "The number of digits in IMSI shall not
	// exceed 15." The shortest IMSI is a three-digit MCC, a two-digit MNC
	// and a one-digit MSIN.
	identityIMSI = identity{min: 6, max: 15, err: ErrIMSIInvalidLength}

	// identityIMEI: 3GPP TS 23.003 V20.1.0 §6.2.1 composes the IMEI of the
	// TAC ("Its length is 8 digits"), the SNR ("Its length is 6 digits") and
	// the CD/SD, 15 digits; §6.2.2 composes the IMEISV of the TAC, the SNR
	// and the SVN ("Its length is 2 digits"), 16 digits. A MAP field of type
	// IMEI holds either: 3GPP TS 29.002 V19.1.0 §17.7.8 IMEI "Refers to
	// International Mobile Station Equipment Identity and Software Version
	// Number (SVN) [...] If the SVN is not present the last octet shall
	// contain the digit 0 and a filler. If present the SVN shall be included
	// in the last octet."
	identityIMEI = identity{min: 15, max: 16, err: ErrIMEIInvalidLength}

	// identityIMEISV: the IMEISV parameter (3GPP TS 29.002 V19.1.0
	// §7.6.2.3a, ADD-Info imeisv), 16 digits per 3GPP TS 23.003 V20.1.0
	// §6.2.2.
	identityIMEISV = identity{min: 16, max: 16, err: ErrIMEISVInvalidLength}
)

// encodeIdentityDigits TBCD-encodes the digits of an IMSI, IMEI or IMEISV.
// The TBCD-STRING alphabet of TS 29.002 V19.1.0 §17.7.8 also carries
// * # a b c, but these identities are decimal digit strings (TS 23.003), so
// anything else is rejected with ErrIdentityNotDigits.
func encodeIdentityDigits(id identity, digits string) ([]byte, error) {
	if err := checkIdentityDigits(id, digits); err != nil {
		return nil, err
	}
	return tbcd.Encode(digits)
}

// decodeIdentityDigits is the inverse of encodeIdentityDigits, with the
// same rules.
func decodeIdentityDigits(id identity, raw []byte) (string, error) {
	digits, err := tbcd.Decode(raw)
	if err != nil {
		return "", err
	}
	if err := checkIdentityDigits(id, digits); err != nil {
		return "", err
	}
	return digits, nil
}

// checkIdentityDigits is the single place that checks an identity's
// digits. No digits at all (an absent value, or octets that are all TBCD
// filler, which tbcd.Decode drops) is ErrIdentityEmpty; a character other
// than 0-9 is ErrIdentityNotDigits; a digit count outside the identity's
// rule is id.err.
func checkIdentityDigits(id identity, digits string) error {
	if digits == "" {
		return ErrIdentityEmpty
	}
	for i := 0; i < len(digits); i++ {
		if digits[i] < '0' || digits[i] > '9' {
			return fmt.Errorf("%w: %q at index %d", ErrIdentityNotDigits, digits[i], i)
		}
	}
	if len(digits) < id.min || len(digits) > id.max {
		return fmt.Errorf("%w (got %d digits)", id.err, len(digits))
	}
	return nil
}

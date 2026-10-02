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

// validatePlmnId is the canonical 3-octet PLMN-Id check per TS 23.003,
// shared across all converters that surface a PLMN-Id field.
func validatePlmnId(b HexBytes, field string) error {
	if len(b) != 3 {
		return fmt.Errorf("%s: %w (got %d)", field, ErrPlmnIdInvalidSize, len(b))
	}
	return nil
}

// validateAPN checks the APN OCTET STRING (SIZE 2..63) constraint per
// TS 29.002 MAP-MS-DataTypes.asn:1654.
func validateAPN(b HexBytes, field string) error {
	if len(b) < 2 || len(b) > 63 {
		return fmt.Errorf("%s: %w (got %d)", field, ErrAPNInvalidSize, len(b))
	}
	return nil
}

// validateAPNOIReplacement checks the APN-OI-Replacement OCTET STRING
// (SIZE 9..100) constraint per TS 29.002 MAP-MS-DataTypes.asn:1303.
// Used by GPRSSubscriptionData, PDPContext and APN-Configuration.
func validateAPNOIReplacement(b HexBytes, field string) error {
	if len(b) < 9 || len(b) > 100 {
		return fmt.Errorf("%s: %w (got %d)", field, ErrAPNOIReplacementInvalidSize, len(b))
	}
	return nil
}

// validateFQDN checks the FQDN OCTET STRING (SIZE 9..255) constraint per
// TS 29.002 MAP-MS-DataTypes.asn:1434. Used by PDPContext.SCEFID,
// APN-Configuration and LCSClientExternalID.
func validateFQDN(b HexBytes, field string) error {
	if len(b) < 9 || len(b) > 255 {
		return fmt.Errorf("%s: %w (got %d)", field, ErrFQDNInvalidSize, len(b))
	}
	return nil
}

// validatePDPAddress checks the PDP-Address OCTET STRING (SIZE 1..16)
// constraint per TS 29.002 MAP-MS-DataTypes.asn:1665. Reused by
// PDPContext (PdpAddress, ExtPdpAddress) and APN-Configuration
// (ServedPartyIPIPv4Address, ServedPartyIPIPv6Address).
func validatePDPAddress(b HexBytes, field string) error {
	if len(b) < 1 || len(b) > 16 {
		return fmt.Errorf("%s: %w (got %d)", field, ErrPDPAddressInvalidSize, len(b))
	}
	return nil
}

// encodeIdentityDigits TBCD-encodes the digits of an IMSI, IMEI or IMEISV.
// The TBCD-STRING alphabet of TS 29.002 V19.1.0 §17.7.8 also carries
// * # a b c, but these identities are decimal digit strings (TS 23.003), so
// anything else is rejected with ErrIdentityNotDigits.
func encodeIdentityDigits(digits string) ([]byte, error) {
	if err := checkIdentityDigits(digits); err != nil {
		return nil, err
	}
	return tbcd.Encode(digits)
}

// decodeIdentityDigits is the inverse of encodeIdentityDigits.
//
// This is the single place that rejects an identity with no digits: nil,
// empty, and anything that decodes to "" (for example all-filler octets,
// which tbcd.Decode drops) return ErrIdentityEmpty.
func decodeIdentityDigits(raw []byte) (string, error) {
	digits, err := tbcd.Decode(raw)
	if err != nil {
		return "", err
	}
	if digits == "" {
		return "", ErrIdentityEmpty
	}
	if err := checkIdentityDigits(digits); err != nil {
		return "", err
	}
	return digits, nil
}

func checkIdentityDigits(digits string) error {
	for i := 0; i < len(digits); i++ {
		if digits[i] < '0' || digits[i] > '9' {
			return fmt.Errorf("%w: %q at index %d", ErrIdentityNotDigits, digits[i], i)
		}
	}
	return nil
}

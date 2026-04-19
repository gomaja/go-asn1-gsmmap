package gsmmap

import (
	"fmt"
	"math"

	"github.com/gomaja/go-asn1-gsmmap/address"
	"github.com/gomaja/go-asn1-gsmmap/tbcd"
)

const errEncodingIMSI = "encoding IMSI: %w"

// natureOrDefault returns the given nature if non-zero, otherwise International.
func natureOrDefault(nature uint8) uint8 {
	if nature == 0 {
		return address.NatureInternational
	}
	return nature
}

// planOrDefault returns the given plan if non-zero, otherwise ISDN.
func planOrDefault(plan uint8) uint8 {
	if plan == 0 {
		return address.PlanISDN
	}
	return plan
}

// encodeAddressField encodes a phone number string into an AddressString byte slice.
func encodeAddressField(digits string, nature, plan uint8) ([]byte, error) {
	tbcdBytes, err := tbcd.Encode(digits)
	if err != nil {
		return nil, err
	}
	return address.Encode(address.ExtensionNo, natureOrDefault(nature), planOrDefault(plan), tbcdBytes), nil
}

// decodeAddressField decodes an AddressString byte slice into a phone number string and address components.
func decodeAddressField(encoded []byte) (digits string, nature, plan uint8, err error) {
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
// On 32-bit builds, rejects values outside [math.MinInt32, math.MaxInt32]
// rather than silently truncating.
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

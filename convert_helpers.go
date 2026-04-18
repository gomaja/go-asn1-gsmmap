package gsmmap

import (
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
func int64PtrTo(p *int64) *int {
	if p == nil {
		return nil
	}
	v := int(*p)
	return &v
}

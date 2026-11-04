// Package gsn encodes and decodes the GSN-Address of 3GPP TS 23.003 V20.1.0
// §5.1: one octet holding the address type in its two most significant bits
// and the address length in the other six, followed by the address octets.
// Build sends a 4-octet address as IPv4 (type 0) and a 16-octet address,
// IPv4-mapped ones included, as IPv6 (type 1). Parse renders each address
// in the family it was coded in, so an IPv4-mapped address (RFC 4291
// §2.5.5.2) reads "::ffff:a.b.c.d" (RFC 5952 §5) and builds back to the same
// 17 octets.
//
// Parse ignores any octets after the address its first octet describes, which
// TS 23.003 §5.1 does not define, and Build never sends them. A MAP
// GSN-Address is OCTET STRING (SIZE (5..17)) (3GPP TS 29.002 V19.1.0 §17.7.8),
// so up to 12 such octets fit after an IPv4 address and none after an IPv6
// one. An address received with them is sent back without them.
package gsn

import (
	"errors"
	"fmt"
	"net/netip"
)

// Rejection conditions for GSN address conversion.
var (
	ErrInvalidIP          = errors.New("gsn: invalid IP")
	ErrAddressTooShort    = errors.New("gsn: address too short")
	ErrDataTooShort       = errors.New("gsn: data too short")
	ErrIPv4LengthInvalid  = errors.New("gsn: invalid IPv4 address length")
	ErrIPv6LengthInvalid  = errors.New("gsn: invalid IPv6 address length")
	ErrAddressTypeInvalid = errors.New("gsn: unknown address type")
)

// Build encodes an IPv4/IPv6 address into GSN Address format per 3GPP TS 23.003.
//
// Format:
//
//	| Address Type (2 bits) | Address Length (6 bits) | ...address bytes...
//
// A dotted-quad string is sent as IPv4; an IPv6 string, including the
// IPv4-mapped form ::ffff:a.b.c.d, is sent as IPv6. A zone is refused.
func Build(ipStr string) ([]byte, error) {
	ip, err := netip.ParseAddr(ipStr)
	if err != nil || ip.Zone() != "" {
		return nil, fmt.Errorf("%w: %s", ErrInvalidIP, ipStr)
	}

	var addrType uint8
	var addrBytes []byte

	if ip.Is4() {
		addrType = 0
		b := ip.As4()
		addrBytes = b[:]
	} else {
		addrType = 1
		b := ip.As16()
		addrBytes = b[:]
	}

	addrLen := uint8(len(addrBytes))
	header := (addrType << 6) | (addrLen & 0x3F)

	result := make([]byte, 1+len(addrBytes))
	result[0] = header
	copy(result[1:], addrBytes)

	return result, nil
}

// Parse decodes a GSN Address format byte slice into an IPv4/IPv6 string.
// An IPv6 address that maps an IPv4 address is rendered as ::ffff:a.b.c.d,
// so that Build sends it back in the family it arrived in. Octets after the
// address are ignored.
func Parse(data []byte) (string, error) {
	if len(data) < 1 {
		return "", fmt.Errorf("%w", ErrAddressTooShort)
	}

	header := data[0]
	addrType := (header >> 6) & 0x03
	addrLen := header & 0x3F

	if len(data) < 1+int(addrLen) {
		return "", fmt.Errorf("%w: expected %d bytes, got %d", ErrDataTooShort, 1+int(addrLen), len(data))
	}

	addrBytes := data[1 : 1+addrLen]

	switch addrType {
	case 0:
		if addrLen != 4 {
			return "", fmt.Errorf("%w: %d", ErrIPv4LengthInvalid, addrLen)
		}
		return netip.AddrFrom4([4]byte(addrBytes)).String(), nil
	case 1:
		if addrLen != 16 {
			return "", fmt.Errorf("%w: %d", ErrIPv6LengthInvalid, addrLen)
		}
		return netip.AddrFrom16([16]byte(addrBytes)).String(), nil
	default:
		return "", fmt.Errorf("%w: %d", ErrAddressTypeInvalid, addrType)
	}
}

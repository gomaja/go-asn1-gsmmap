package gsn

import (
	"fmt"
	"net"
)

// Build encodes an IPv4/IPv6 address into GSN Address format per 3GPP TS 23.003.
//
// Format:
//
//	| Address Type (2 bits) | Address Length (6 bits) | ...address bytes...
func Build(ipStr string) ([]byte, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("gsn: invalid IP: %s", ipStr)
	}

	var addrType uint8
	var addrBytes []byte

	if ipv4 := ip.To4(); ipv4 != nil {
		addrType = 0
		addrBytes = ipv4
	} else {
		addrType = 1
		addrBytes = ip.To16()
	}

	addrLen := uint8(len(addrBytes))
	header := (addrType << 6) | (addrLen & 0x3F)

	result := make([]byte, 1+len(addrBytes))
	result[0] = header
	copy(result[1:], addrBytes)

	return result, nil
}

// Parse decodes a GSN Address format byte slice into an IPv4/IPv6 string.
func Parse(data []byte) (string, error) {
	if len(data) < 1 {
		return "", fmt.Errorf("gsn: address too short")
	}

	header := data[0]
	addrType := (header >> 6) & 0x03
	addrLen := header & 0x3F

	if len(data) < 1+int(addrLen) {
		return "", fmt.Errorf("gsn: data too short: expected %d bytes, got %d", 1+int(addrLen), len(data))
	}

	addrBytes := data[1 : 1+addrLen]

	switch addrType {
	case 0:
		if addrLen != 4 {
			return "", fmt.Errorf("gsn: invalid IPv4 address length: %d", addrLen)
		}
		return net.IP(addrBytes).String(), nil
	case 1:
		if addrLen != 16 {
			return "", fmt.Errorf("gsn: invalid IPv6 address length: %d", addrLen)
		}
		return net.IP(addrBytes).String(), nil
	default:
		return "", fmt.Errorf("gsn: unknown address type: %d", addrType)
	}
}

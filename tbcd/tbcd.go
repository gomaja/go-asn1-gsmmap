package tbcd

import (
	"encoding/hex"
	"fmt"
)

// Encode creates a TBCD-encoded byte slice per ETSI TS 129 002.
// TBCD (Telephony Binary Coded Decimal) swaps the nibbles of each byte.
// For odd-length strings, a filler 'F' is appended.
// Input must consist of valid hexadecimal digits (0-9, a-f, A-F).
func Encode(s string) ([]byte, error) {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return nil, fmt.Errorf("tbcd: invalid character in input: %c", r)
		}
	}

	hexString := s
	if len(s)%2 != 0 {
		hexString = s + "f"
	}

	raw, err := hex.DecodeString(hexString)
	if err != nil {
		return nil, fmt.Errorf("tbcd: failed to decode hex string: %w", err)
	}

	return swapNibbles(raw), nil
}

// Decode decodes a TBCD-encoded byte slice into a hexadecimal string.
// It removes trailing 'f' padding if present.
func Decode(raw []byte) (string, error) {
	if raw == nil {
		return "", fmt.Errorf("tbcd: input is nil")
	}

	swapped := swapNibbles(raw)
	s := hex.EncodeToString(swapped)

	if len(s) > 0 && (s[len(s)-1] == 'f' || s[len(s)-1] == 'F') {
		s = s[:len(s)-1]
	}

	return s, nil
}

func swapNibbles(data []byte) []byte {
	swapped := make([]byte, len(data))
	for i, b := range data {
		swapped[i] = ((b & 0x0F) << 4) | ((b & 0xF0) >> 4)
	}
	return swapped
}

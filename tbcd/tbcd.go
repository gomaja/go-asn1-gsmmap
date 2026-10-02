// Package tbcd encodes and decodes the TBCD-STRING type of 3GPP TS 29.002
// (MAP), as used for IMSI, IMEI, IMEISV and the digits of an AddressString.
//
// The governing text is 3GPP TS 29.002 V19.1.0 §17.7.8 (MAP-CommonDataTypes):
//
//	TBCD-STRING ::= OCTET STRING
//	-- This type (Telephony Binary Coded Decimal String) is used to
//	-- represent several digits from 0 through 9, *, #, a, b, c, two
//	-- digits per octet, each digit encoded 0000 to 1001 (0 to 9),
//	-- 1010 (*), 1011 (#), 1100 (a), 1101 (b) or 1110 (c); 1111 used
//	-- as filler when there is an odd number of digits.
//
//	-- bits 8765 of octet n encoding digit 2n
//	-- bits 4321 of octet n encoding digit 2(n-1) +1
//
// TBCD is therefore not hexadecimal: the alphabet is exactly "0123456789*#abc"
// and the nibble value 1111 is a filler, never a digit.
package tbcd

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidCharacter is returned by Encode for a character outside the
	// TBCD alphabet "0123456789*#abc".
	ErrInvalidCharacter = errors.New("tbcd: character outside the TBCD alphabet 0-9 * # a b c")

	// ErrMisplacedFiller is returned by Decode when a filler nibble (1111) is
	// followed by a digit nibble, i.e. the filler is not confined to the end
	// of the string.
	ErrMisplacedFiller = errors.New("tbcd: filler nibble followed by a digit")
)

const (
	filler = 0x0F
	// alphabet maps a nibble value 0..14 to its digit. Value 15 is the filler
	// and has no digit.
	alphabet = "0123456789*#abc"
)

// Encode converts s to a TBCD-STRING. s must consist only of characters from
// the alphabet "0123456789*#abc"; any other character (including A-F, d, e, f
// and non-ASCII characters) yields an error wrapping ErrInvalidCharacter that
// names the character and its byte index in s. An odd number of digits is
// completed with the filler nibble 1111 in the high nibble of the last octet.
// The empty string encodes to an empty, non-nil slice.
func Encode(s string) ([]byte, error) {
	for i, r := range s {
		if r > 0x7F {
			return nil, fmt.Errorf("%w: %q at index %d", ErrInvalidCharacter, r, i)
		}
		if _, ok := nibble(byte(r)); !ok {
			return nil, fmt.Errorf("%w: %q at index %d", ErrInvalidCharacter, r, i)
		}
	}
	out := make([]byte, (len(s)+1)/2)
	for i := 0; i < len(s); i++ {
		n, _ := nibble(s[i])
		if i%2 == 0 {
			out[i/2] = filler<<4 | n
		} else {
			out[i/2] = n<<4 | out[i/2]&0x0F
		}
	}
	return out, nil
}

func nibble(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c == '*':
		return 0xA, true
	case c == '#':
		return 0xB, true
	case c >= 'a' && c <= 'c':
		return c - 'a' + 0xC, true
	}
	return 0, false
}

// Decode converts a TBCD-STRING to its digits. The low nibble of each octet
// is the earlier digit.
//
// Filler rules. The nibble 1111 is a filler. A filler is valid only after the
// last digit; a filler followed by a digit anywhere, including a low nibble
// of 1111 followed by a non-filler high nibble in the same octet, returns an
// error wrapping ErrMisplacedFiller that names the octet index.
//
// Trailing filler octets are tolerated and dropped. The specification only
// has a filler for an odd digit count, but a fixed-size field such as
// GroupId (TBCD-STRING (SIZE (3)), filled with "six TBCD fillers (1111)" when
// a longGroupId is present) must carry filler octets, and some implementations
// pad other fields with 0xFF. Accepting them loses no information, since a
// nibble that follows a filler can only be another filler. The consequence is
// that Decode is not injective: Encode(Decode(x)) equals x with its trailing
// 0xFF octets removed, which is the canonical form.
//
// Decode(nil) and Decode([]byte{}) both return ("", nil). A zero-digit
// string is a well-formed TBCD-STRING; whether an empty value is acceptable
// is a SIZE or presence rule of the containing type and is for the caller to
// enforce. A nil slice carries no more information than an empty one in Go.
func Decode(raw []byte) (string, error) {
	out := make([]byte, 0, len(raw)*2)
	seenFiller := false
	for i, b := range raw {
		for _, n := range [2]byte{b & 0x0F, b >> 4} {
			if n == filler {
				seenFiller = true
				continue
			}
			if seenFiller {
				return "", fmt.Errorf("%w: octet %d (0x%02x)", ErrMisplacedFiller, i, b)
			}
			out = append(out, alphabet[n])
		}
	}
	return string(out), nil
}

package tbcd

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)

// TS 29.002 V19.1.0 §17.7.8: digits 0..9, *, #, a..c and high-nibble filler.
func TestMutationTBCDAlphabet(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  []byte
	}{
		{"0", []byte{0xf0}}, {"9", []byte{0xf9}},
		{"*", []byte{0xfa}}, {"#", []byte{0xfb}},
		{"a", []byte{0xfc}}, {"b", []byte{0xfd}}, {"c", []byte{0xfe}},
		{"09*#abc", []byte{0x90, 0xba, 0xdc, 0xfe}},
	} {
		t.Run(tc.input, func(t *testing.T) {
			got, err := Encode(tc.input)
			if err != nil || !bytes.Equal(got, tc.want) {
				t.Fatalf("Encode = %x, %v; want %x, nil", got, err, tc.want)
			}
			decoded, err := Decode(tc.want)
			if err != nil || decoded != tc.input {
				t.Fatalf("Decode = %q, %v; want %q, nil", decoded, err, tc.input)
			}
		})
	}
	for _, c := range []byte{'/', ':', '`', 'd', 0x7f} {
		t.Run(fmt.Sprintf("invalid_%02x", c), func(t *testing.T) {
			got, err := Encode(string(c))
			if got != nil || !errors.Is(err, ErrInvalidCharacter) {
				t.Fatalf("Encode = %x, %v; want nil, invalid character", got, err)
			}
			want := fmt.Sprintf("%s: %q at index 0", ErrInvalidCharacter, rune(c))
			if err.Error() != want {
				t.Fatalf("error = %q, want %q", err, want)
			}
		})
	}
}

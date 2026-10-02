package tbcd

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		encoded []byte
	}{
		{"even length", "1234567890", []byte{0x21, 0x43, 0x65, 0x87, 0x09}},
		{"odd length with filler", "123451234567890", []byte{0x21, 0x43, 0x15, 0x32, 0x54, 0x76, 0x98, 0xf0}},
		{"single digit", "5", []byte{0xf5}},
		{"empty string", "", []byte{}},
		{"star", "*", []byte{0xfa}},
		{"hash", "#", []byte{0xfb}},
		{"a", "a", []byte{0xfc}},
		{"b", "b", []byte{0xfd}},
		{"c", "c", []byte{0xfe}},
		{"star hash even", "*#", []byte{0xba}},
		{"abc odd", "abc", []byte{0xdc, 0xfe}},
		{"ussd code", "*123#", []byte{0x1a, 0x32, 0xfb}},
		{"ussd code even", "*1234#", []byte{0x1a, 0x32, 0xb4}},
		{"zero", "0", []byte{0xf0}},
		{"nine nine", "99", []byte{0x99}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := Encode(tt.input)
			if err != nil {
				t.Fatalf("Encode(%q) error: %v", tt.input, err)
			}
			if enc == nil || !bytes.Equal(enc, tt.encoded) {
				t.Errorf("Encode(%q) = %x, want %x", tt.input, enc, tt.encoded)
			}
			dec, err := Decode(tt.encoded)
			if err != nil {
				t.Fatalf("Decode(%x) error: %v", tt.encoded, err)
			}
			if dec != tt.input {
				t.Errorf("Decode(%x) = %q, want %q", tt.encoded, dec, tt.input)
			}
		})
	}
}

// Every symbol of the alphabet, both ways, in both nibble positions.
func TestAlphabetEverySymbol(t *testing.T) {
	const alpha = "0123456789*#abc"
	for i := 0; i < len(alpha); i++ {
		nib := byte(i)
		sym := string(alpha[i])
		// Low nibble (first digit of an octet, odd length).
		got, err := Decode([]byte{0xF0 | nib})
		if err != nil || got != sym {
			t.Errorf("Decode(low nibble %x) = %q, %v; want %q", nib, got, err, sym)
		}
		// High nibble (second digit).
		got, err = Decode([]byte{nib<<4 | 0x1})
		if err != nil || got != "1"+sym {
			t.Errorf("Decode(high nibble %x) = %q, %v; want %q", nib, got, err, "1"+sym)
		}
		enc, err := Encode(sym)
		if err != nil || !bytes.Equal(enc, []byte{0xF0 | nib}) {
			t.Errorf("Encode(%q) = %x, %v", sym, enc, err)
		}
		enc, err = Encode("1" + sym)
		if err != nil || !bytes.Equal(enc, []byte{nib<<4 | 0x1}) {
			t.Errorf("Encode(%q) = %x, %v", "1"+sym, enc, err)
		}
	}
}

func TestEncodeInvalidCharacter(t *testing.T) {
	tests := []struct {
		name  string
		input string
		char  string
		index string
	}{
		{"g", "123g456", "'g'", "index 3"},
		{"upper A", "A", "'A'", "index 0"},
		{"upper B", "12B", "'B'", "index 2"},
		{"upper C", "1C", "'C'", "index 1"},
		{"upper D", "D", "'D'", "index 0"},
		{"upper E", "E", "'E'", "index 0"},
		{"upper F", "F", "'F'", "index 0"},
		{"lower d", "12d", "'d'", "index 2"},
		{"lower e", "e", "'e'", "index 0"},
		{"lower f", "1234f", "'f'", "index 4"},
		{"plus", "+123", "'+'", "index 0"},
		{"space", "12 34", "' '", "index 2"},
		{"dash", "12-34", "'-'", "index 2"},
		{"NUL", "1\x00", `'\x00'`, "index 1"},
		{"non-ASCII", "12é", "'é'", "index 2"},
		{"invalid UTF-8", "1\xff", "'�'", "index 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := Encode(tt.input)
			if !errors.Is(err, ErrInvalidCharacter) {
				t.Fatalf("Encode(%q) = %x, %v; want ErrInvalidCharacter", tt.input, enc, err)
			}
			if enc != nil {
				t.Errorf("Encode(%q) returned bytes %x with an error", tt.input, enc)
			}
			if !strings.Contains(err.Error(), tt.char) || !strings.Contains(err.Error(), tt.index) {
				t.Errorf("error %q does not name %s at %s", err, tt.char, tt.index)
			}
		})
	}
}

// Every byte value other than the 15 alphabet characters is rejected.
func TestEncodeRejectsEveryOtherByte(t *testing.T) {
	const alpha = "0123456789*#abc"
	for c := 0; c < 0x100; c++ {
		s := string([]byte{byte(c)})
		_, err := Encode(s)
		if strings.IndexByte(alpha, byte(c)) >= 0 {
			if err != nil {
				t.Errorf("Encode(%q) rejected an alphabet character: %v", s, err)
			}
			continue
		}
		if !errors.Is(err, ErrInvalidCharacter) {
			t.Errorf("Encode(0x%02x) = %v; want ErrInvalidCharacter", c, err)
		}
	}
}

func TestDecodeFiller(t *testing.T) {
	tests := []struct {
		name string
		raw  []byte
		want string
		err  error
	}{
		{"odd count, filler in last high nibble", []byte{0x21, 0xf3}, "123", nil},
		{"all filler octet", []byte{0xff}, "", nil},
		{"two filler octets", []byte{0xff, 0xff}, "", nil},
		{"three filler octets (GroupId padding)", []byte{0xff, 0xff, 0xff}, "", nil},
		{"digits then trailing filler octet", []byte{0x21, 0x43, 0xff}, "1234", nil},
		{"odd digits then trailing filler octets", []byte{0x21, 0xf3, 0xff, 0xff}, "123", nil},
		{"digit and filler then filler octets (GroupId 1 digit)", []byte{0xf1, 0xff, 0xff}, "1", nil},
		{"two digits then filler octet", []byte{0x21, 0xff, 0xff}, "12", nil},

		{"filler octet then digit octet", []byte{0xff, 0x21}, "", ErrMisplacedFiller},
		{"digit, filler octet, digit octet", []byte{0x21, 0xff, 0x43}, "", ErrMisplacedFiller},
		{"low filler then digit in high nibble", []byte{0xaf}, "", ErrMisplacedFiller},
		{"low filler then digit, after digits", []byte{0x21, 0x3f, 0x65}, "", ErrMisplacedFiller},
		{"low filler high zero", []byte{0x0f}, "", ErrMisplacedFiller},
		{"low filler high star", []byte{0xaf}, "", ErrMisplacedFiller},
		{"low filler high c", []byte{0xef}, "", ErrMisplacedFiller},
		{"filler high nibble then digit next octet", []byte{0xf1, 0x21}, "", ErrMisplacedFiller},
		{"filler high nibble, filler octet, digit octet", []byte{0xf1, 0xff, 0x05}, "", ErrMisplacedFiller},
		{"ff ff ff decodes to nothing, not fffff", []byte{0xff, 0xff, 0xff}, "", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decode(tt.raw)
			if tt.err != nil {
				if !errors.Is(err, tt.err) {
					t.Fatalf("Decode(%x) = %q, %v; want %v", tt.raw, got, err, tt.err)
				}
				if got != "" {
					t.Errorf("Decode(%x) returned %q with an error", tt.raw, got)
				}
				if !strings.Contains(err.Error(), "octet") {
					t.Errorf("error %q does not name the octet", err)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Errorf("Decode(%x) = %q, %v; want %q", tt.raw, got, err, tt.want)
			}
		})
	}
}

func TestDecodeNilAndEmpty(t *testing.T) {
	for name, in := range map[string][]byte{"nil": nil, "empty": {}} {
		got, err := Decode(in)
		if err != nil || got != "" {
			t.Errorf("Decode(%s) = %q, %v; want \"\", nil", name, got, err)
		}
	}
	enc, err := Encode("")
	if err != nil || enc == nil || len(enc) != 0 {
		t.Errorf("Encode(\"\") = %#v, %v; want empty non-nil slice", enc, err)
	}
}

// Decode of every one- and two-octet value either fails with
// ErrMisplacedFiller or round-trips through Encode, after the documented
// normalisation of dropping trailing filler octets.
func TestExhaustiveShortRoundTrip(t *testing.T) {
	check := func(raw []byte) {
		dec, err := Decode(raw)
		if err != nil {
			if !errors.Is(err, ErrMisplacedFiller) {
				t.Errorf("Decode(%x): unexpected error %v", raw, err)
			}
			return
		}
		enc, err := Encode(dec)
		if err != nil {
			t.Errorf("Encode(Decode(%x)=%q): %v", raw, dec, err)
			return
		}
		if want := trimFF(raw); !bytes.Equal(enc, want) {
			t.Errorf("Encode(Decode(%x)) = %x, want %x", raw, enc, want)
		}
	}
	for a := 0; a < 256; a++ {
		check([]byte{byte(a)})
		for b := 0; b < 256; b++ {
			check([]byte{byte(a), byte(b)})
		}
	}
}

func TestRoundTripStrings(t *testing.T) {
	for _, s := range []string{
		"", "0", "12", "310150123456789", "*#", "*#0#", "**21*491711#", "#31#", "abcabc", "0123456789*#abc",
		"491711123456789012", "a", "00", "000",
	} {
		enc, err := Encode(s)
		if err != nil {
			t.Fatalf("Encode(%q): %v", s, err)
		}
		if len(enc) != (len(s)+1)/2 {
			t.Errorf("Encode(%q) length %d, want %d", s, len(enc), (len(s)+1)/2)
		}
		dec, err := Decode(enc)
		if err != nil || dec != s {
			t.Errorf("Decode(Encode(%q)) = %q, %v", s, dec, err)
		}
	}
}

func FuzzDecode(f *testing.F) {
	for _, seed := range [][]byte{
		nil, {}, {0xff}, {0xff, 0xff, 0xff}, {0x21, 0xf3}, {0x21, 0x43, 0xff}, {0xaf}, {0x21, 0xff, 0x43},
		{0x1a, 0x32, 0xfb}, {0xfe}, {0x00}, {0x99, 0xf9},
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		dec, err := Decode(raw)
		if err != nil {
			if !errors.Is(err, ErrMisplacedFiller) {
				t.Fatalf("Decode(%x): unexpected error %v", raw, err)
			}
			return
		}
		for i := 0; i < len(dec); i++ {
			if _, ok := nibble(dec[i]); !ok {
				t.Fatalf("Decode(%x) = %q contains a character outside the alphabet", raw, dec)
			}
		}
		enc, err := Encode(dec)
		if err != nil {
			t.Fatalf("Encode(Decode(%x) = %q): %v", raw, dec, err)
		}
		// Documented normalisation: trailing filler octets are dropped.
		if want := trimFF(raw); !bytes.Equal(enc, want) {
			t.Fatalf("Encode(Decode(%x)) = %x, want %x", raw, enc, want)
		}
		again, err := Decode(enc)
		if err != nil || again != dec {
			t.Fatalf("Decode(%x) = %q, %v; want %q", enc, again, err, dec)
		}
	})
}

// trimFF drops trailing 0xFF octets: the canonical form Decode normalises to.
func trimFF(b []byte) []byte {
	for len(b) > 0 && b[len(b)-1] == 0xff {
		b = b[:len(b)-1]
	}
	return b
}

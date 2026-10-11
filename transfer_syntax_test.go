package gsmmap

import (
	"errors"
	"fmt"
	"testing"
)

// checkDefiniteLengths walks the BER of b and reports a length the MAP
// transfer syntax does not allow. 3GPP TS 29.002 V19.1.0 §17.1.1: "When the
// definite form is used for length encoding, a data value of length less
// than 128 octets must have the length encoded in the short form. When the
// long form is employed to code a length, the minimum number of octets shall
// be used to code the length field."
func checkDefiniteLengths(b []byte) error {
	for len(b) > 0 {
		i := 1
		if b[0]&0x1f == 0x1f {
			for i < len(b) && b[i]&0x80 != 0 {
				i++
			}
			i++
		}
		if i >= len(b) {
			return errors.New("truncated identifier or length")
		}
		length := int(b[i])
		i++
		switch {
		case length == 0x80:
			return errors.New("indefinite length")
		case length > 0x80:
			n := length & 0x7f
			if i+n > len(b) || b[i] == 0 {
				return fmt.Errorf("long-form length with a leading zero octet or truncated (%d octets)", n)
			}
			length = 0
			for _, o := range b[i : i+n] {
				length = length<<8 | int(o)
			}
			if length < 128 {
				return fmt.Errorf("length %d in the long form", length)
			}
			i += n
		}
		if i+length > len(b) {
			return errors.New("length beyond the data")
		}
		if b[0]&0x20 != 0 {
			if err := checkDefiniteLengths(b[i : i+length]); err != nil {
				return err
			}
		}
		b = b[i+length:]
	}
	return nil
}

// Every value Parse accepts from the seed corpus marshals with the definite
// lengths 3GPP TS 29.002 V19.1.0 §17.1.1 requires.
func TestMarshalUsesMinimalDefiniteLengths(t *testing.T) {
	checked := 0
	for _, seed := range berSeeds(t) {
		for _, p := range parsers {
			v, err := p.parse(seed)
			if err != nil {
				continue
			}
			out, err := v.Marshal()
			if err != nil {
				continue
			}
			checked++
			if err := checkDefiniteLengths(out); err != nil {
				t.Errorf("%s: %v: %x", p.name, err, out)
			}
		}
	}
	if checked == 0 {
		t.Fatal("no value marshalled")
	}
}

func TestCheckDefiniteLengths(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   []byte
		ok   bool
	}{
		{"short form", []byte{0x30, 0x03, 0x80, 0x01, 0x00}, true},
		{"indefinite", []byte{0x30, 0x80, 0x80, 0x01, 0x00, 0x00, 0x00}, false},
		{"long form below 128", []byte{0x30, 0x81, 0x03, 0x80, 0x01, 0x00}, false},
		{"long form with a leading zero", []byte{0x04, 0x82, 0x00, 0x80}, false},
		{"nested long form below 128", []byte{0x30, 0x04, 0x80, 0x81, 0x01, 0x00}, false},
	} {
		if err := checkDefiniteLengths(tc.in); (err == nil) != tc.ok {
			t.Errorf("%s: err = %v, want ok %t", tc.name, err, tc.ok)
		}
	}
}

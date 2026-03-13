package tbcd

import (
	"bytes"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		encoded []byte
	}{
		{
			name:    "even length",
			input:   "1234567890",
			encoded: []byte{0x21, 0x43, 0x65, 0x87, 0x09},
		},
		{
			name:    "odd length with padding",
			input:   "123451234567890",
			encoded: []byte{0x21, 0x43, 0x15, 0x32, 0x54, 0x76, 0x98, 0xf0},
		},
		{
			name:    "single digit",
			input:   "5",
			encoded: []byte{0xf5},
		},
		{
			name:    "empty string",
			input:   "",
			encoded: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := Encode(tt.input)
			if err != nil {
				t.Fatalf("Encode(%q) error: %v", tt.input, err)
			}
			if !bytes.Equal(enc, tt.encoded) {
				t.Errorf("Encode(%q) = %x, want %x", tt.input, enc, tt.encoded)
			}

			dec, err := Decode(enc)
			if err != nil {
				t.Fatalf("Decode(%x) error: %v", enc, err)
			}
			if dec != tt.input {
				t.Errorf("Decode(%x) = %q, want %q", enc, dec, tt.input)
			}
		})
	}
}

func TestEncodeInvalidInput(t *testing.T) {
	_, err := Encode("123g456")
	if err == nil {
		t.Fatal("expected error for invalid hex digit")
	}
}

func TestDecodeNil(t *testing.T) {
	_, err := Decode(nil)
	if err == nil {
		t.Fatal("expected error for nil input")
	}
}

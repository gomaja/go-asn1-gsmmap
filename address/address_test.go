package address

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	digits := []byte{0x21, 0x43, 0x65}
	encoded := Encode(ExtensionNo, NatureInternational, PlanISDN, digits)

	ext, nature, plan, dec := Decode(encoded)
	if ext != ExtensionNo {
		t.Errorf("extension: got %d, want %d", ext, ExtensionNo)
	}
	if nature != NatureInternational {
		t.Errorf("nature: got %d, want %d", nature, NatureInternational)
	}
	if plan != PlanISDN {
		t.Errorf("plan: got %d, want %d", plan, PlanISDN)
	}
	if !bytes.Equal(dec, digits) {
		t.Errorf("digits: got %x, want %x", dec, digits)
	}
}

func TestDecodeEmpty(t *testing.T) {
	ext, nature, plan, digits := Decode(nil)
	if ext != 0 || nature != 0 || plan != 0 || digits != nil {
		t.Errorf("expected all zeros for empty input")
	}
}

func TestDecodeSingleByte(t *testing.T) {
	_, _, _, digits := Decode([]byte{0x91})
	if len(digits) != 0 {
		t.Errorf("expected empty digits for single-byte input, got %x", digits)
	}
}

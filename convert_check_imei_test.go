package gsmmap

import (
	"errors"
	"testing"
)

func TestCheckIMEIMinimal(t *testing.T) {
	// SEQUENCE: eight IMEI TBCD octets; RequestedEquipmentInfo equipmentStatus, SIZE 2.
	hlrRoundTrip(t, &CheckIMEIArg{IMEI: "123456789012340", RequestedEquipmentInfo: RequestedEquipmentInfo{EquipmentStatus: true}}, ParseCheckIMEIArg, "300e040821436587092143f003020680")
	// All result fields are optional.
	hlrRoundTrip(t, &CheckIMEIRes{}, ParseCheckIMEIRes, "3000")
}

func TestCheckIMEIFull(t *testing.T) {
	// Both requested bits, and a 16-digit IMEISV.
	hlrRoundTrip(t, &CheckIMEIArg{IMEI: "1234567890123456", RequestedEquipmentInfo: RequestedEquipmentInfo{EquipmentStatus: true, BMUEF: true}}, ParseCheckIMEIArg, "300e04082143658709214365030206c0")
	// EquipmentStatus trackingListed, one-bit A and nine-bit B.
	status := EquipmentStatusTrackingListed
	hlrRoundTrip(t, &CheckIMEIRes{EquipmentStatus: &status, BMUEF: &UESBIIu{UESBIIuA: HexBytes{0x80}, UESBIIuABits: 1, UESBIIuB: HexBytes{0xaa, 0x80}, UESBIIuBBits: 9}}, ParseCheckIMEIRes, "300e0a0102300980020780810307aa80")
	// BMUEF can contain neither, only A, or only B.
	hlrRoundTrip(t, &CheckIMEIRes{BMUEF: &UESBIIu{}}, ParseCheckIMEIRes, "30023000")
	hlrRoundTrip(t, &CheckIMEIRes{BMUEF: &UESBIIu{UESBIIuA: HexBytes{0}, UESBIIuABits: 8}}, ParseCheckIMEIRes, "3006300480020000")
	hlrRoundTrip(t, &CheckIMEIRes{BMUEF: &UESBIIu{UESBIIuB: HexBytes{0x80}, UESBIIuBBits: 1}}, ParseCheckIMEIRes, "3006300481020780")
}

func TestRequestedEquipmentInfoReceiverDiscard(t *testing.T) {
	for bit := 2; bit < 8; bit++ {
		// IMEISV and eight bits: equipmentStatus plus one unknown assignment.
		wire := mustHex(t, "300e0408214365870921436503020080")
		wire[len(wire)-1] |= 0x80 >> uint(bit)
		got, err := ParseCheckIMEIArg(wire)
		if err != nil {
			t.Fatal(err)
		}
		wantEqual(t, "known bits", RequestedEquipmentInfo{EquipmentStatus: true}, got.RequestedEquipmentInfo)
		out, err := got.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		wantEqual(t, "canonical", mustHex(t, "300e0408214365870921436503020680"), out)
	}
}

func TestCheckIMEIPadding(t *testing.T) {
	// Received unused bits are ignored; unnamed strings retain their bit count.
	v, err := ParseCheckIMEIRes(mustHex(t, "300a3008800207ff8102077f"))
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "A", HexBytes{0x80}, v.BMUEF.UESBIIuA)
	wantEqual(t, "B", HexBytes{0}, v.BMUEF.UESBIIuB)
	input := &CheckIMEIRes{BMUEF: &UESBIIu{UESBIIuA: HexBytes{0xff}, UESBIIuABits: 1, UESBIIuB: HexBytes{0x7f}, UESBIIuBBits: 1}}
	converted, err := convertCheckIMEIResToWire(input)
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "converter A padding", []byte{0x80}, converted.Bmuef.UesbiIuA.Bytes)
	wantEqual(t, "converter B padding", []byte{0}, converted.Bmuef.UesbiIuB.Bytes)
	wire, err := input.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "zero padding", mustHex(t, "300a30088002078081020700"), wire)
	wantEqual(t, "input A unmodified", HexBytes{0xff}, input.BMUEF.UESBIIuA)
	wantEqual(t, "input B unmodified", HexBytes{0x7f}, input.BMUEF.UESBIIuB)
}

func TestCheckIMEIConstraints(t *testing.T) {
	for _, tc := range []struct {
		v                      *CheckIMEIRes
		wire, path, constraint string
	}{
		{&CheckIMEIRes{EquipmentStatus: new(EquipmentStatus(3))}, "30030a0103", "equipmentStatus", "ENUMERATED {0, 1, 2}"},
		{&CheckIMEIRes{BMUEF: &UESBIIu{UESBIIuA: make(HexBytes, 17), UESBIIuABits: 129}}, "30053003800100", "bmuef.uesbi-IuA", "SIZE (1..128)"},
		{&CheckIMEIRes{BMUEF: &UESBIIu{UESBIIuB: make(HexBytes, 17), UESBIIuBBits: 129}}, "30053003810100", "bmuef.uesbi-IuB", "SIZE (1..128)"},
	} {
		_, err := tc.v.Marshal()
		wantValidationError(t, err, tc.path, tc.constraint)
		_, err = ParseCheckIMEIRes(mustHex(t, tc.wire))
		wantValidationError(t, err, tc.path, tc.constraint)
	}
	// A significant bit beyond RequestedEquipmentInfo SIZE 8 is rejected.
	_, err := ParseCheckIMEIArg(mustHex(t, "300f040821436587092143f00303070080"))
	wantValidationError(t, err, "requestedEquipmentInfo", "SIZE (2..8)")
	for _, bits := range []int{1, 128} {
		data := make(HexBytes, (bits+7)/8)
		v := &CheckIMEIRes{BMUEF: &UESBIIu{UESBIIuA: data, UESBIIuABits: bits, UESBIIuB: data, UESBIIuBBits: bits}}
		out, err := v.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		got, err := ParseCheckIMEIRes(out)
		if err != nil {
			t.Fatal(err)
		}
		wantEqual(t, "boundary", v, got)
	}
	for _, bits := range []int{-1, 129} {
		data := make(HexBytes, 17)
		if _, err := (&CheckIMEIRes{BMUEF: &UESBIIu{UESBIIuA: data, UESBIIuABits: bits}}).Marshal(); err == nil {
			t.Fatalf("A accepted %d bits", bits)
		}
		if _, err := (&CheckIMEIRes{BMUEF: &UESBIIu{UESBIIuB: data, UESBIIuBBits: bits}}).Marshal(); err == nil {
			t.Fatalf("B accepted %d bits", bits)
		}
	}
	if _, err := (&CheckIMEIRes{BMUEF: &UESBIIu{UESBIIuABits: 1}}).Marshal(); err == nil {
		t.Fatal("accepted missing A bytes")
	}
	if _, err := (&CheckIMEIRes{BMUEF: &UESBIIu{UESBIIuBBits: 1}}).Marshal(); err == nil {
		t.Fatal("accepted missing B bytes")
	}
}

func TestCheckIMEISenderRulesFromWire(t *testing.T) {
	for _, tc := range []struct {
		wire string
		want error
	}{
		// BMUEF request with a 15-digit IMEI, rather than an IMEISV (§8.7.1.3).
		{"300e040821436587092143f003020640", ErrCheckIMEIArgBMUEFRequiresIMEISV},
		// Neither defined information bit requested (§8.7.1.3).
		{"300e040821436587092143f003020600", ErrCheckIMEIArgNoEquipmentInfoRequested},
		// Only an unknown bit requested: receiver discards it (§17.7.1).
		{"300e040821436587092143f003020020", ErrCheckIMEIArgNoEquipmentInfoRequested},
	} {
		v, err := ParseCheckIMEIArg(mustHex(t, tc.wire))
		if err != nil {
			t.Fatal(err)
		}
		if _, err = v.Marshal(); !errors.Is(err, tc.want) {
			t.Errorf("got %v, want %v", err, tc.want)
		}
	}
	// BMUEF alone is valid for an IMEISV.
	hlrRoundTrip(t, &CheckIMEIArg{IMEI: "1234567890123456", RequestedEquipmentInfo: RequestedEquipmentInfo{BMUEF: true}}, ParseCheckIMEIArg, "300e0408214365870921436503020640")
}

func TestCheckIMEINamedBitPadding(t *testing.T) {
	// BER permits omitting trailing zero bits from named BIT STRINGs.
	v, err := ParseCheckIMEIArg(mustHex(t, "300e040821436587092143f003020780"))
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "equipment request", RequestedEquipmentInfo{EquipmentStatus: true}, v.RequestedEquipmentInfo)
}

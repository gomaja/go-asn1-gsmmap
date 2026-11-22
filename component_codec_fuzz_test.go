package gsmmap

import (
	"bytes"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	phase1 "github.com/gomaja/go-asn1/telecom/ss7/gsm_map/gsm0902_v3_11_0"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime/ber"
)

var componentFuzzContexts = []OID{
	"", "0.4.0.0.1.0.1.1", "0.4.0.0.1.0.1.2",
	"0.4.0.0.1.0.18.2", "0.4.0.0.1.0.19.2", "0.4.0.0.1.0.23.2", "0.4.0.0.1.0.26.2",
	"0.4.0.0.1.0.20.3", "0.4.0.0.1.0.19.3", "0.4.0.0.1.0.1.4",
}

// FuzzComponentCodec checks that a parsed value marshals, that the output
// parses, and that a second Marshal returns the same octets. The one Marshal
// failure it accepts is go-asn1's documented refusal of an unknown addition
// that needs schema-dependent repair (runtime/ber WithoutBERReplay); see
// TestComponentOpaqueAdditionNeedsRepair.
func FuzzComponentCodec(f *testing.F) {
	seed := func(acn OID, role componentRole, op, code int64, wire string) {
		var index uint8
		for i, name := range componentFuzzContexts {
			if name == acn {
				index = uint8(i)
				break
			}
		}
		data, err := hex.DecodeString(wire)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(index, uint8(role), op, code, data)
	}
	for _, tc := range componentWires() {
		seed(tc.acn, tc.role, tc.op, tc.errorCode, tc.wire)
	}
	for _, tc := range componentRuleCases() {
		seed(tc.acn, tc.role, tc.op, tc.error, tc.bad)
		seed(tc.acn, tc.role, tc.op, tc.error, tc.good)
	}
	// Non-canonical true and an indefinite Phase 1 constructor exercise fresh encoding.
	seed("0.4.0.0.1.0.1.2", componentArgument, 45, 0, "300b8002912181010182029143")
	seed("", componentResult, 9, 0, "308080032143650000")
	seed("", componentArgument, -1, 0, "")
	// Phase 1 passwords: four digits, then values the ASN.1 admits but GSM 02.04 §5.2 does not.
	for _, wire := range []string{"120431323334", "12053132333435", "1200", "130461626364"} {
		seed("", componentResult, 17, 0, wire)
		seed("0.4.0.0.1.0.1.1", componentResult, 18, 0, wire)
	}
	// An unknown addition whose SET needs repair: parsed, refused by Marshal.
	seed("0.4.0.0.1.0.1.2", componentArgument, 45, 0, componentRepairWire)
	seed("0.4.0.0.1.0.20.3", componentArgument, 45, 0, componentRepairWire)
	// Malformed and unexpected-presence wires from the diagnostic/sentinel tests.
	for _, acn := range componentFuzzContexts {
		for _, wire := range []string{"ff", "30810b8002912181010082029143", "300da0040402912181010082029143", "0500"} {
			for role := componentArgument; role <= componentError; role++ {
				seed(acn, role, 45, 13, wire)
			}
		}
	}
	f.Fuzz(func(t *testing.T, edition, role uint8, op, code int64, data []byte) {
		c, err := ComponentCodecFor(componentFuzzContexts[int(edition)%len(componentFuzzContexts)])
		if err != nil {
			t.Fatal(err)
		}
		r := componentRole(role % 3)
		value, err := componentParse(c, r, op, code, data)
		if errors.Is(err, ErrComponentPasswordFormat) {
			// The decoded value is kept for relay: it replays the input and Marshal refuses it.
			password, ok := value.(phase1.Password)
			if !ok {
				t.Fatalf("password format error with %T", value)
			}
			if replayed, err := password.MarshalBER(); err != nil || !bytes.Equal(replayed, data) {
				t.Fatalf("replay %x %v, want %x", replayed, err, data)
			}
			if encoded, err := componentMarshal(c, r, op, code, value); encoded != nil || !errors.Is(err, ErrComponentPasswordFormat) {
				t.Fatalf("marshal accepted a refused password: %x %v", encoded, err)
			}
			return
		}
		if err != nil || value == nil {
			return
		}
		first, err := componentMarshal(c, r, op, code, value)
		if err != nil {
			if errors.Is(err, ber.ErrInvalidValue) && strings.Contains(err.Error(), "opaque element requires a typed value for canonical repair") {
				return
			}
			t.Fatalf("parsed %T cannot marshal: %v", value, err)
		}
		again, err := componentParse(c, r, op, code, first)
		if err != nil || again == nil {
			t.Fatalf("fresh encoding cannot parse: %v", err)
		}
		second, err := componentMarshal(c, r, op, code, again)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(first, second) {
			t.Fatalf("unstable encoding: %x != %x", first, second)
		}
	})
}

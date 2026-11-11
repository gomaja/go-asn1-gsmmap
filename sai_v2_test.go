package gsmmap

import (
	"bytes"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	identity "github.com/gomaja/go-3gpp-identity"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime/ber"
	"github.com/google/go-cmp/cmp"
)

// GSM 09.02 V4.19.1 §14.7.1 wires. An AuthenticationSet has 34 content octets
// (RAND TLV 18, SRES TLV 6, Kc TLV 10), 36 with its own tag and length.
const (
	// A bare IMSI, eight TBCD octets: 001010123456789.
	saiV2ArgHex = "040800010121436587f9"
	// One set; outer content 36 octets.
	saiV2ResHex = "302430220410000102030405060708090a0b0c0d0e0f0404a1a2a3a40408b1b2b3b4b5b6b7b8"
	// Five distinct sets, saiV2DistinctTriplet(1) to (5) in order: 5*36 = 180
	// (0xb4) content octets.
	saiV2FiveHex = "3081b430220410101112131415161718191a1b1c1d1e1f0404a1a1a1a10408b1b1b1b1b1b1b1b130220410202122232425262728292a2b2c2d2e2f0404a2a2a2a20408b2b2b2b2b2b2b2b230220410303132333435363738393a3b3c3d3e3f0404a3a3a3a30408b3b3b3b3b3b3b3b330220410404142434445464748494a4b4c4d4e4f0404a4a4a4a40408b4b4b4b4b4b4b4b430220410505152535455565758595a5b5c5d5e5f0404a5a5a5a50408b5b5b5b5b5b5b5b5"
)

// saiV2Bytes decodes a hex wire literal and fails the test if it is malformed.
func saiV2Bytes(t *testing.T, wire string) []byte {
	t.Helper()
	b, err := hex.DecodeString(wire)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// saiV2Triplet returns the one set of saiV2ResHex: RAND 00 to 0f, SRES a1 to
// a4 and Kc b1 to b8.
func saiV2Triplet() AuthenticationTriplet {
	return AuthenticationTriplet{
		RAND: HexBytes{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		SRES: HexBytes{0xa1, 0xa2, 0xa3, 0xa4},
		Kc:   HexBytes{0xb1, 0xb2, 0xb3, 0xb4, 0xb5, 0xb6, 0xb7, 0xb8},
	}
}

// saiV2DistinctTriplet returns set n of saiV2FiveHex: RAND n0 to nf, SRES
// four octets a0+n and Kc eight octets b0+n.
func saiV2DistinctTriplet(n byte) AuthenticationTriplet {
	t := AuthenticationTriplet{RAND: make(HexBytes, 16), SRES: make(HexBytes, 4), Kc: make(HexBytes, 8)}
	for i := range t.RAND {
		t.RAND[i] = n<<4 | byte(i)
	}
	for i := range t.SRES {
		t.SRES[i] = 0xa0 + n
	}
	for i := range t.Kc {
		t.Kc[i] = 0xb0 + n
	}
	return t
}

// SendAuthenticationInfoArg ::= IMSI (GSM 09.02 V4.19.1 §14.7.1): eight TBCD
// octets decode to fifteen digits and encode back to the same wire.
func TestSAIV2Argument(t *testing.T) {
	want := &SendAuthenticationInfoV2{IMSI: "001010123456789"}
	got, err := ParseSendAuthenticationInfoV2(saiV2Bytes(t, saiV2ArgHex))
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal(diff)
	}
	out, err := want.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if hex.EncodeToString(out) != saiV2ArgHex {
		t.Fatalf("Marshal = %x, want %s", out, saiV2ArgHex)
	}
}

// The sets keep their order in both directions (GSM 09.02 V4.19.1 §14.7.1,
// SEQUENCE OF AuthenticationSet).
func TestSAIV2Result(t *testing.T) {
	five := make([]AuthenticationTriplet, 5)
	for i := range five {
		five[i] = saiV2DistinctTriplet(byte(i + 1))
	}
	for _, tc := range []struct {
		wire     string
		triplets []AuthenticationTriplet
	}{{saiV2ResHex, []AuthenticationTriplet{saiV2Triplet()}}, {saiV2FiveHex, five}} {
		want := &SendAuthenticationInfoV2Res{Triplets: tc.triplets}
		got, err := ParseSendAuthenticationInfoV2Res(saiV2Bytes(t, tc.wire))
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatal(diff)
		}
		out, err := want.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		if hex.EncodeToString(out) != tc.wire {
			t.Fatalf("Marshal = %x, want %s", out, tc.wire)
		}
	}
}

// berSeeds harvests the v2 argument, so FuzzParse and
// TestMarshalUsesMinimalDefiniteLengths reach ParseSendAuthenticationInfoV2.
func TestSAIV2ArgumentSeeded(t *testing.T) {
	want := saiV2Bytes(t, saiV2ArgHex)
	for _, seed := range berSeeds(t) {
		if bytes.Equal(seed, want) {
			return
		}
	}
	t.Fatalf("berSeeds lacks %s", saiV2ArgHex)
}

// Each version's parsers refuse the other's wire with ber.ErrUnexpectedElement.
// The v3 argument is a SEQUENCE and the v3 result a [3] SEQUENCE (TS 29.002
// V19.1.0 §17.7.1). The v2 argument is an IMSI and the v2 result a SEQUENCE OF
// (GSM 09.02 V4.19.1 §14.7.1). The v3 result wire is checked first.
func TestSAIV2VersionSeparation(t *testing.T) {
	arg, err := (&SendAuthenticationInfo{IMSI: "001010123456789", NumberOfRequestedVectors: 1}).Marshal()
	if err != nil {
		t.Fatal(err)
	}
	res, err := (&SendAuthenticationInfoRes{AuthenticationSetList: &AuthenticationSetList{Triplets: []AuthenticationTriplet{saiV2Triplet()}}}).Marshal()
	if err != nil {
		t.Fatal(err)
	}
	// TS 29.002 V19.1.0 §17.7.1: v3 result [3], tripletList [0], one set.
	const v3Res = "a326a02430220410000102030405060708090a0b0c0d0e0f0404a1a2a3a40408b1b2b3b4b5b6b7b8"
	if hex.EncodeToString(res) != v3Res {
		t.Fatalf("v3 Marshal = %x, want %s", res, v3Res)
	}
	parsed, err := ParseSendAuthenticationInfoRes(res)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]AuthenticationTriplet{saiV2Triplet()}, parsed.AuthenticationSetList.Triplets); diff != "" {
		t.Fatal(diff)
	}
	for _, tc := range []struct {
		name  string
		parse func([]byte) (marshaler, error)
		wire  []byte
	}{
		{"v3 argument", asParser(ParseSendAuthenticationInfo), saiV2Bytes(t, saiV2ArgHex)},
		{"v3 result", asParser(ParseSendAuthenticationInfoRes), saiV2Bytes(t, saiV2ResHex)},
		{"v2 argument", asParser(ParseSendAuthenticationInfoV2), arg},
		{"v2 result", asParser(ParseSendAuthenticationInfoV2Res), res},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v, err := tc.parse(tc.wire)
			if v != nil || !errors.Is(err, ber.ErrUnexpectedElement) {
				t.Fatalf("Parse = %v, %v; want nil, ErrUnexpectedElement", v, err)
			}
		})
	}
}

// GSM 09.02 V4.19.1 §§14.7.1, 14.7.8: codec-owned sizes, including
// the one-to-five list constraint. Every literal's enclosing lengths include
// its child TLVs; six 36-octet sets occupy 216 (0xd8) content octets.
func TestSAIV2DecodeSizes(t *testing.T) {
	for _, tc := range []struct {
		name, wire, path, constraint string
		n                            int
	}{
		{"IMSI short", "04020001", "SendAuthenticationInfoArg", "SIZE (3..8)", 2},              // Two IMSI octets.
		{"IMSI long", "0409000101214365870911", "SendAuthenticationInfoArg", "SIZE (3..8)", 9}, // Nine IMSI octets.
		{"empty list", "3000", "AuthenticationSetList", "SIZE (1..5)", 0},                      // Empty SEQUENCE.
		{"six sets", "3081d830220410000102030405060708090a0b0c0d0e0f0404a1a2a3a40408b1b2b3b4b5b6b7b830220410000102030405060708090a0b0c0d0e0f0404a1a2a3a40408b1b2b3b4b5b6b7b830220410000102030405060708090a0b0c0d0e0f0404a1a2a3a40408b1b2b3b4b5b6b7b830220410000102030405060708090a0b0c0d0e0f0404a1a2a3a40408b1b2b3b4b5b6b7b830220410000102030405060708090a0b0c0d0e0f0404a1a2a3a40408b1b2b3b4b5b6b7b830220410000102030405060708090a0b0c0d0e0f0404a1a2a3a40408b1b2b3b4b5b6b7b8", "AuthenticationSetList", "SIZE (1..5)", 6}, // Six sets.
		{"SRES short", "302330210410000102030405060708090a0b0c0d0e0f0403a1a2a30408b1b2b3b4b5b6b7b8", "element[0].sres", "SIZE (4)", 3},   // 35 outer content octets, 33 set content octets.
		{"RAND short", "30233021040f000102030405060708090a0b0c0d0e0404a1a2a3a40408b1b2b3b4b5b6b7b8", "element[0].rand", "SIZE (16)", 15}, // 35 outer, 33 set content octets.
		{"Kc long", "302530230410000102030405060708090a0b0c0d0e0f0404a1a2a3a40409b1b2b3b4b5b6b7b8b9", "element[0].kc", "SIZE (8)", 9},    // 37 outer, 35 set content octets.
	} {
		t.Run(tc.name, func(t *testing.T) {
			parse := asParser(ParseSendAuthenticationInfoV2Res)
			name := "SendAuthenticationInfoV2Res"
			if strings.HasPrefix(tc.name, "IMSI") {
				parse = asParser(ParseSendAuthenticationInfoV2)
				name = "SendAuthenticationInfoV2"
			}
			v, err := parse(saiV2Bytes(t, tc.wire))
			if v != nil || err == nil || !strings.HasPrefix(err.Error(), "decoding "+name+": ") {
				t.Fatalf("Parse = %v, %v; want nil with public decoding context", v, err)
			}
			wantValidationError(t, err, tc.path, tc.constraint, tc.n)
		})
	}
}

// Marshal leaves the sizes to the codec. An empty list breaks SIZE (1..5) and
// a three-octet SRES breaks SIZE (4) (GSM 09.02 V4.19.1 §14.7.1); both errors
// keep the public encoding context.
func TestSAIV2EncodeSizes(t *testing.T) {
	for _, tc := range []struct {
		name, path, constraint string
		triplets               []AuthenticationTriplet
		n                      int
	}{
		{"empty", "AuthenticationSetList", "SIZE (1..5)", nil, 0},
		{"SRES short", "element[0].sres", "SIZE (4)", []AuthenticationTriplet{{RAND: saiV2Triplet().RAND, SRES: HexBytes{1, 2, 3}, Kc: saiV2Triplet().Kc}}, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := (&SendAuthenticationInfoV2Res{Triplets: tc.triplets}).Marshal()
			if out != nil || err == nil || !strings.HasPrefix(err.Error(), "encoding SendAuthenticationInfoV2Res: ") {
				t.Fatalf("Marshal = %x, %v; want nil with public encoding context", out, err)
			}
			wantValidationError(t, err, tc.path, tc.constraint, tc.n)
		})
	}
}

// The Phase 2 source edition enforces GSM 09.02 V4.19.1 §14.1.1,
// including its source-specific error citation.
func TestSAIV2ArgumentProfile(t *testing.T) {
	for _, tc := range []struct{ wire, rule string }{
		{"240c0406000101214365040287f9", "primitive-octet-string"}, // Constructed IMSI, 12 content octets: segments of six and two.
		{"04810800010121436587f9", "short-length"},                 // Eight content octets, non-minimal long-form length.
	} {
		v, err := ParseSendAuthenticationInfoV2(saiV2Bytes(t, tc.wire))
		if v != nil {
			t.Fatalf("Parse = %+v, want nil", v)
		}
		violation := wantValidationError(t, err, "SendAuthenticationInfoArg", "GSM 09.02 V4.19.1 §14.1.1: "+tc.rule)
		wantRule := ber.RuleRef{Document: "GSM 09.02", Edition: "V4.19.1", Clause: "14.1.1"}
		if violation.Kind != ber.ProtocolRule || violation.Rule != wantRule {
			t.Errorf("kind/rule = %v/%+v, want %v/%+v", violation.Kind, violation.Rule, ber.ProtocolRule, wantRule)
		}
		if violation.ObservedValue != "received BER form" {
			t.Errorf("observed value = %q, want received BER form", violation.ObservedValue)
		}
		var elementError *ber.ElementError
		if !errors.As(err, &elementError) {
			t.Errorf("err = %v, want a *ber.ElementError", err)
		}
	}
}

// The IMSI follows the package's digit rules. Parse refuses an all-filler, a
// three-digit and a sixteen-digit IMSI and accepts the five-digit MCC+MNC form.
// Marshal refuses an empty IMSI. Parse's and Marshal's errors keep their
// conversion context.
func TestSAIV2IdentityErrors(t *testing.T) {
	for _, tc := range []struct {
		wire, want string
		err        error
	}{
		{"0403ffffff", "", identity.ErrEmpty},                // Three all-filler octets.
		{"040300f1ff", "", identity.ErrIMSILength},           // Three octets, only three digits.
		{"04080001012143658709", "", identity.ErrIMSILength}, // Eight octets, sixteen digits.
		{"04030001f1", "00101", nil},                         // Three octets, five-digit MCC+MNC.
	} {
		v, err := ParseSendAuthenticationInfoV2(saiV2Bytes(t, tc.wire))
		if tc.err != nil {
			if v != nil || !errors.Is(err, tc.err) || !strings.HasPrefix(err.Error(), "decoding IMSI: ") {
				t.Fatalf("Parse = %+v, %v; want nil, %v with conversion context", v, err, tc.err)
			}
		} else if err != nil || v.IMSI != tc.want {
			t.Fatalf("Parse = %+v, %v; want IMSI %s", v, err, tc.want)
		}
	}
	out, err := (&SendAuthenticationInfoV2{}).Marshal()
	if out != nil || !errors.Is(err, identity.ErrEmpty) || !strings.HasPrefix(err.Error(), "converting SendAuthenticationInfoV2: encoding IMSI: ") {
		t.Fatalf("Marshal = %x, %v; want nil, ErrEmpty with conversion context", out, err)
	}
}

// "An entity supporting a version greater 1 shall not reject unsupported
// extension following "..." of that SEQUENCE", and PLMN specific extensions
// follow the extension marker with PRIVATE tags up to and including 29
// (GSM 09.02 V4.19.1 §14.1.4). Each wire is one
// AuthenticationSet of 34 base octets plus a three-octet addition: set content
// 37, outer content 39 (§14.7.1).
func TestSAIV2ExtensionDiscarded(t *testing.T) {
	for _, tc := range []struct{ name, wire string }{
		{"context [1]", "302730250410000102030405060708090a0b0c0d0e0f0404a1a2a3a40408b1b2b3b4b5b6b7b88101ff"},
		{"PRIVATE [1]", "302730250410000102030405060708090a0b0c0d0e0f0404a1a2a3a40408b1b2b3b4b5b6b7b8c101ff"},
		{"PRIVATE [29]", "302730250410000102030405060708090a0b0c0d0e0f0404a1a2a3a40408b1b2b3b4b5b6b7b8dd01ff"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v, err := ParseSendAuthenticationInfoV2Res(saiV2Bytes(t, tc.wire))
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff([]AuthenticationTriplet{saiV2Triplet()}, v.Triplets); diff != "" {
				t.Fatal(diff)
			}
			out, err := v.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			if hex.EncodeToString(out) != saiV2ResHex {
				t.Fatalf("Marshal = %x, want %s", out, saiV2ResHex)
			}
		})
	}
}

// X.690 (02/2021) §8.1.3.2 b) lets the sender of a constructed encoding use
// the indefinite form, and GSM 09.02 V4.19.1 §14.1.1 restricts only definite
// lengths. Parse accepts it; Marshal sends the definite form.
func TestSAIV2IndefiniteLength(t *testing.T) {
	// The one-set result with an indefinite outer length, ended by 00 00.
	const indefinite = "308030220410000102030405060708090a0b0c0d0e0f0404a1a2a3a40408b1b2b3b4b5b6b7b80000"
	v, err := ParseSendAuthenticationInfoV2Res(saiV2Bytes(t, indefinite))
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]AuthenticationTriplet{saiV2Triplet()}, v.Triplets); diff != "" {
		t.Fatal(diff)
	}
	out, err := v.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if hex.EncodeToString(out) != saiV2ResHex {
		t.Fatalf("Marshal = %x, want %s", out, saiV2ResHex)
	}
}

// A parameterless return result has nothing to parse (GSM 09.02 V4.19.1
// §6.5.2.1); the v2 argument is mandatory (§6.5.2.2).
func TestSAIV2AbsentParameter(t *testing.T) {
	for _, parse := range []func([]byte) (marshaler, error){asParser(ParseSendAuthenticationInfoV2), asParser(ParseSendAuthenticationInfoV2Res)} {
		v, err := parse(nil)
		if v != nil || err == nil {
			t.Fatalf("Parse(nil) = %v, %v; want nil, error", v, err)
		}
	}
}

// The v3 AuthenticationSetList converters keep their nil and unset-CHOICE
// results; both alternatives are exercised through the public codecs.
func TestAuthenticationSetListConverterEdges(t *testing.T) {
	if v, err := convertAuthenticationSetListToWire(nil); v != nil || err != nil {
		t.Fatalf("nil public list: %v, %v", v, err)
	}
	if v, err := convertWireToAuthenticationSetList(nil); v != nil || err != nil {
		t.Fatalf("nil wire list: %v, %v", v, err)
	}
	v, err := convertWireToAuthenticationSetList(&gsm_map.AuthenticationSetList{})
	if v != nil || !errors.Is(err, ErrAuthenticationSetListUnknownAlternative) {
		t.Fatalf("unset CHOICE: %v, %v; want nil, ErrAuthenticationSetListUnknownAlternative", v, err)
	}
}

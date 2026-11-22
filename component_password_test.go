package gsmmap

import (
	"bytes"
	"encoding/hex"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	phase1 "github.com/gomaja/go-asn1/telecom/ss7/gsm_map/gsm0902_v3_11_0"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime/ber"
)

func componentStringTLV(tag byte, s string) []byte {
	return append([]byte{tag, byte(len(s))}, s...)
}

// GSM 02.04 version 3.7.1 §5.2: "The password will consist of four digits in
// the range 0000 to 9999." registerPassword (17) and getPassword (18) return
// a Password in GSM 09.02 V3.11.0, whose ASN.1 admits more than that.
func TestComponentPhase1PasswordFormat(t *testing.T) {
	const numeric, printable = 0x12, 0x13
	cases := []struct {
		name string
		tag  byte
		s    string
		ok   bool
	}{
		{"four digits", numeric, "1234", true},
		{"lowest", numeric, "0000", true},
		{"highest", numeric, "9999", true},
		{"five digits", numeric, "12345", false},
		{"two digits", numeric, "12", false},
		{"empty", numeric, "", false},
		{"inner space", numeric, "12 4", false},
		{"four spaces", numeric, "    ", false},
		{"printable letters", printable, "abcd", false},
		{"printable digits", printable, "1234", false},
		{"printable eight", printable, "abcdefgh", false},
	}
	for _, acn := range []OID{"", "0.4.0.0.1.0.18.1", "0.4.0.0.1.0.1.1"} {
		c, err := ComponentCodecFor(acn)
		if err != nil {
			t.Fatal(err)
		}
		for _, op := range []int64{17, 18} {
			for _, tc := range cases {
				t.Run(string(acn)+"/"+tc.name, func(t *testing.T) {
					wire := componentStringTLV(tc.tag, tc.s)
					value, err := c.ParseResult(op, wire)
					password, isPassword := value.(phase1.Password)
					if !isPassword {
						t.Fatalf("op %d %x: value %T, err %v; want the decoded Password", op, wire, value, err)
					}
					if tc.ok {
						if err != nil {
							t.Fatalf("op %d %x: %v", op, wire, err)
						}
						encoded, err := c.MarshalResult(op, password)
						if err != nil || !bytes.Equal(encoded, wire) {
							t.Fatalf("op %d marshal: %x %v, want %x", op, encoded, err, wire)
						}
						return
					}
					if !errors.Is(err, ErrComponentPasswordFormat) || errors.Is(err, ErrVersionRule) {
						t.Fatalf("op %d %x: err %v, want ErrComponentPasswordFormat", op, wire, err)
					}
					if len(tc.s) > 0 && strings.Contains(err.Error(), tc.s) && strings.TrimSpace(tc.s) != "" {
						t.Fatalf("op %d: error text quotes the password: %v", op, err)
					}
					// The value keeps its received encoding for a caller that relays it.
					replayed, err := password.MarshalBER()
					if err != nil || !bytes.Equal(replayed, wire) {
						t.Fatalf("op %d replay: %x %v, want %x", op, replayed, err, wire)
					}
					encoded, err := c.MarshalResult(op, password)
					if encoded != nil || !errors.Is(err, ErrComponentPasswordFormat) {
						t.Fatalf("op %d marshal: %x %v, want ErrComponentPasswordFormat", op, encoded, err)
					}
				})
			}
		}
	}
}

func TestComponentPhase1PasswordValues(t *testing.T) {
	c, err := ComponentCodecFor("")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		value phase1.Password
		want  string
		err   error
	}{
		{"numeric four", phase1.NewPasswordAlternative1("0420"), "120430343230", nil},
		{"numeric three", phase1.NewPasswordAlternative1("042"), "", ErrComponentPasswordFormat},
		{"numeric letter", phase1.NewPasswordAlternative1("04a0"), "", ErrComponentPasswordFormat},
		{"printable", phase1.NewPasswordAlternative0("0420"), "", ErrComponentPasswordFormat},
		// A nil selected alternative is go-asn1's refusal, not a format finding.
		{"nil alternative", phase1.Password{Choice: phase1.PasswordChoiceAlternative1}, "", ber.ErrInvalidValue},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := c.MarshalResult(18, tc.value)
			if hex.EncodeToString(got) != tc.want || !errors.Is(err, tc.err) {
				t.Fatalf("got %x %v, want %s %v", got, err, tc.want, tc.err)
			}
			if tc.err != ErrComponentPasswordFormat && errors.Is(err, ErrComponentPasswordFormat) {
				t.Fatalf("unexpected format error: %v", err)
			}
		})
	}
}

// A constructed or indefinite-length encoding replays exactly, although a fresh
// encoding would be primitive, so these cases distinguish replay from encoding.
func TestComponentPhase1PasswordReplay(t *testing.T) {
	c, err := ComponentCodecFor("")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, wire, value string }{
		{"constructed", "3209040231320403333435", "12345"},
		{"indefinite", "32800402313204033334350000", "12345"},
		{"constructed empty segments", "32080400040231320400", "12"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wire := componentHex(t, tc.wire)
			value, err := c.ParseResult(18, wire)
			password, ok := value.(phase1.Password)
			if !ok || !errors.Is(err, ErrComponentPasswordFormat) {
				t.Fatalf("%T %v, want a Password and ErrComponentPasswordFormat", value, err)
			}
			if password.Choice != phase1.PasswordChoiceAlternative1 || password.Alternative1 == nil || *password.Alternative1 != tc.value {
				t.Fatalf("decoded %+v, want NumericString %q", password, tc.value)
			}
			replayed, err := password.MarshalBER()
			if err != nil || !bytes.Equal(replayed, wire) {
				t.Fatalf("replay %x %v, want %x", replayed, err, wire)
			}
			fresh, err := password.MarshalBER(ber.WithoutBERReplay())
			if err != nil || bytes.Equal(fresh, wire) {
				t.Fatalf("fresh encoding %x %v should differ from %x", fresh, err, wire)
			}
			if encoded, err := c.MarshalResult(18, password); encoded != nil || !errors.Is(err, ErrComponentPasswordFormat) {
				t.Fatalf("marshal %x %v, want ErrComponentPasswordFormat", encoded, err)
			}
		})
	}
}

// The union refuses only values that satisfy neither operand, and the later
// editions refuse every value that is not four digits in their own ASN.1, so
// those failures are decode errors and not format findings.
func TestComponentPasswordDecodeRefusals(t *testing.T) {
	for _, tc := range []struct {
		acn  OID
		wire []byte
	}{
		{"", componentStringTLV(0x12, "12 345")},
		{"0.4.0.0.1.0.1.2", componentStringTLV(0x12, "12345")},
		{"0.4.0.0.1.0.1.2", componentStringTLV(0x12, "12 4")},
		{"0.4.0.0.1.0.18.2", componentStringTLV(0x12, "123")},
		{"0.4.0.0.1.0.1.3", componentStringTLV(0x12, "12a4")},
		{"0.4.0.0.1.0.1.3", componentStringTLV(0x12, "    ")},
	} {
		c, err := ComponentCodecFor(tc.acn)
		if err != nil {
			t.Fatal(err)
		}
		value, err := c.ParseResult(18, tc.wire)
		if value != nil || err == nil || errors.Is(err, ErrComponentPasswordFormat) {
			t.Errorf("%s %x: %T %v, want a decode error", tc.acn, tc.wire, value, err)
		}
	}
	for _, acn := range []OID{"0.4.0.0.1.0.1.2", "0.4.0.0.1.0.18.2", "0.4.0.0.1.0.1.3"} {
		c, _ := ComponentCodecFor(acn)
		wire := componentStringTLV(0x12, "1234")
		value, err := c.ParseResult(18, wire)
		if err != nil {
			t.Fatalf("%s: %v", acn, err)
		}
		if encoded, err := c.MarshalResult(18, value); err != nil || !bytes.Equal(encoded, wire) {
			t.Fatalf("%s: %x %v", acn, encoded, err)
		}
	}
}

// The check looks only at a whole Phase 1 parameter, so Password must be the
// whole parameter wherever the Phase 1 catalogue reaches it.
func TestComponentPhase1PasswordPositions(t *testing.T) {
	target := reflect.TypeFor[phase1.Password]()
	var whole, nested []string
	add := func(where string, typ *componentType) {
		if typ == nil {
			return
		}
		if typ.valueType == target {
			whole = append(whole, where)
		} else if componentTypeReaches(typ.valueType, target, map[reflect.Type]bool{}) {
			nested = append(nested, where)
		}
	}
	cat := componentCatalogues()[EditionPhase1]
	for _, op := range cat.operations {
		add(op.name+" argument", op.argument)
		add(op.name+" result", op.result)
	}
	for _, e := range cat.errors {
		add(e.name+" parameter", e.parameter)
	}
	if len(nested) != 0 {
		t.Fatalf("Password nested in %v", nested)
	}
	got := strings.Join(slices.Sorted(slices.Values(whole)), ", ")
	if got != "getPassword result, registerPassword result" {
		t.Fatalf("Password positions: %s", got)
	}
}

func componentTypeReaches(t, target reflect.Type, seen map[reflect.Type]bool) bool {
	if t == target {
		return true
	}
	if seen[t] {
		return false
	}
	seen[t] = true
	switch t.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array, reflect.Map:
		return componentTypeReaches(t.Elem(), target, seen)
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if componentTypeReaches(t.Field(i).Type, target, seen) {
				return true
			}
		}
	}
	return false
}

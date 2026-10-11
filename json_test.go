package gsmmap

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"
)

// HexBytes follows encoding/json.Marshaler and Unmarshaler contracts.
func TestHexBytesJSON(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value HexBytes
		json  string
	}{
		{"nil", nil, "null"}, {"empty", HexBytes{}, `""`},
		{"octets", HexBytes{0, 0xff, 0x10}, `"00ff10"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.value.MarshalJSON()
			if err != nil {
				t.Fatal(err)
			}
			wantEqual(t, "JSON", tc.json, string(got))
			decoded := HexBytes{0xaa}
			if err := decoded.UnmarshalJSON([]byte(tc.json)); err != nil {
				t.Fatal(err)
			}
			wantEqual(t, "octets", tc.value, decoded)
		})
	}
	for _, tc := range []struct {
		name, input string
		want        error
	}{
		{"invalid JSON", `{`, &json.SyntaxError{}},
		{"invalid hex", `"gg"`, hex.InvalidByteError('g')},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := HexBytes{0xaa}
			err := got.UnmarshalJSON([]byte(tc.input))
			if tc.name == "invalid JSON" {
				var syntax *json.SyntaxError
				if !errors.As(err, &syntax) || syntax.Offset != 1 {
					t.Fatalf("error = %v, want syntax error at offset 1", err)
				}
			} else if !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
			wantEqual(t, "octets after error", HexBytes{0xaa}, got)
		})
	}
}

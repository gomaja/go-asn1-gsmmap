// ussd_7bit_test.go
//
// 7-bit USSD packing vectors (DCS 0x0F). The octets come from an independent
// reference derived from 3GPP TS 23.038 V20.0.0 §6.1.2.3.1 and hand-checked
// against its bit diagrams; they are not produced by the code under test.

package gsmmap

import (
	"encoding/hex"
	"testing"
)

func TestUSSDGSM7PackingVectors(t *testing.T) {
	cases := []struct {
		name       string
		input      string
		packedHex  string
		wantDecode string
	}{
		// 8n-1 characters: 7 spare bits filled with <CR> (§6.1.2.3.1).
		{"7 chars", "ABCDEFG", "41e19058341e1b", "ABCDEFG"},
		{"6 chars + @ (spare bits are not read as @)", "ABCDEF@", "41e1905834021a", "ABCDEF@"},
		{"6 chars + CR", "ABCDEF\r", "41e1905834361a", "ABCDEF\r"},
		{"8 chars", "ABCDEFGH", "41e19058341e91", "ABCDEFGH"},
		// 8n characters ending in <CR>: the sender adds a second <CR>; the
		// octets do not end on a boundary, so the receiver keeps both and
		// CR CR is identical to CR (§6.1.2.3.1).
		{"8 chars ending CR", "ABCDEFG\r", "41e19058341e1b0d", "ABCDEFG\r\r"},
		{"7 chars + @", "ABCDEFG@", "41e19058341e01", "ABCDEFG@"},
		{"15 chars digits/symbols", "*123*456*789*0#", "aa986ca6a2d56caa1b2ea7828d1a", "*123*456*789*0#"},
		{"15 chars + @", "ABCDEFGHIJKLMN@", "41e19058341e9149e592d974021a", "ABCDEFGHIJKLMN@"},
		{"15 chars ending CR", "ABCDEFGHIJKLMN\r", "41e19058341e9149e592d974361a", "ABCDEFGHIJKLMN\r"},
		{"16 chars", "ABCDEFGHIJKLMNOP", "41e19058341e9149e592d9743ea1", "ABCDEFGHIJKLMNOP"},
		{"16 chars ending CR", "ABCDEFGHIJKLMNO\r", "41e19058341e9149e592d9743e1b0d", "ABCDEFGHIJKLMNO\r\r"},
		{"16 chars ending @", "ABCDEFGHIJKLMNO@", "41e19058341e9149e592d9743e01", "ABCDEFGHIJKLMNO@"},
		{"pound sign", "A£", "c100", "A£"},
		{"two digits", "14", "311a", "14"},
		{"two digits 47", "47", "b41b", "47"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			want, err := hex.DecodeString(tc.packedHex)
			if err != nil {
				t.Fatal(err)
			}
			got, err := USSDDataCodingSchemeGSM7.Encode(tc.input)
			if err != nil {
				t.Fatalf("Encode(%q): %v", tc.input, err)
			}
			if hex.EncodeToString(got) != tc.packedHex {
				t.Errorf("Encode(%q) = %x, want %s", tc.input, got, tc.packedHex)
			}
			text, err := USSDDataCodingSchemeGSM7.Decode(want)
			if err != nil {
				t.Fatalf("Decode(%s): %v", tc.packedHex, err)
			}
			if text != tc.wantDecode {
				t.Errorf("Decode(%s) = %q, want %q", tc.packedHex, text, tc.wantDecode)
			}
		})
	}
}

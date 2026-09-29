// ussd_fuzz_test.go
//
// Fuzz targets for the USSD parsers and for USSDDataCodingScheme.Decode.
// Properties: never panic; whenever a parse and the following Marshal both
// succeed, parsing the marshalled bytes yields the same value. A zero MSISDN
// nature or plan is "unknown" and is preserved exactly.

package gsmmap

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// ussdFuzzSeedArgs are BER USSD-Arg seeds: the golden vectors (see
// ussd_golden_test.go) and malformed inputs.
var ussdFuzzSeedArgs = []string{
	"301c04010f040eaa180da682dd6c31192d36bbdd468007917267415827f2",
	"301104010f04094537bd2c0741934e1d040105",
	"301604010f040e50797a5c06d53665d086f75e6f7c040101",
	"3011040148040c041f04400438043204350442",
	"301204010f0405aa180c36028006a86021436587",
	"300c04010f0407d7327bfc6e9743",
	"300604010f040131",
	"30150401480410041f04400438043204350442002020ac",
	// truncated
	"301c04010f040eaa180da682dd6c31192d36bbdd46",
	"30",
	"",
	// wrong tags
	"311c04010f040eaa180da682dd6c31192d36bbdd468007917267415827f2",
	"3006020101040131",
	// indefinite length
	"308004010f04013100" + "00",
	// zero-octet and 161-octet USSD-String (SIZE is not enforced by go-asn1)
	"300504010f0400",
	"30" + "81a7" + "04010f" + "0481a1" + strings.Repeat("41", 161),
	// wrong-length DCS and AlertingPattern
	"300504" + "00" + "040131",
	"30070402" + "0f0f" + "040131",
	"300a04010f040131" + "04020101",
	// present MSISDN with no digits
	"300904010f04013180" + "0191",
}

func FuzzParseUSSDArg(f *testing.F) {
	for _, s := range ussdFuzzSeedArgs {
		b := ussdMustHex(f, s)
		f.Add(b)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		a, err := ParseUSSDArg(data)
		if err != nil {
			return
		}
		want := *a
		out, err := a.Marshal()
		if err != nil {
			return
		}
		back, err := ParseUSSDArg(out)
		if err != nil {
			t.Fatalf("ParseUSSDArg(Marshal(x)) failed: %v\nx=%+v\nbytes=%x", err, a, out)
		}
		if diff := cmp.Diff(&want, back, cmpopts.EquateEmpty()); diff != "" {
			t.Fatalf("round trip mismatch (-want +got):\n%s\nin=%x out=%x", diff, data, out)
		}
	})
}

func FuzzParseUSSDRes(f *testing.F) {
	for _, s := range ussdFuzzSeedArgs {
		f.Add(ussdMustHex(f, s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		r, err := ParseUSSDRes(data)
		if err != nil {
			return
		}
		out, err := r.Marshal()
		if err != nil {
			return
		}
		back, err := ParseUSSDRes(out)
		if err != nil {
			t.Fatalf("ParseUSSDRes(Marshal(x)) failed: %v\nx=%+v\nbytes=%x", err, r, out)
		}
		if diff := cmp.Diff(r, back, cmpopts.EquateEmpty()); diff != "" {
			t.Fatalf("round trip mismatch (-want +got):\n%s\nin=%x out=%x", diff, data, out)
		}
	})
}

func FuzzUSSDDataCodingSchemeDecode(f *testing.F) {
	seeds := []struct {
		dcs uint8
		hex string
	}{
		{0x0F, "aa180da682dd6c31192d36bbdd46"},
		{0x0F, "41e19058341e1b"},
		{0x0F, "41e19058341e1b0d"},
		{0x0F, "41e19058341e9149e592d9743e1b0d"},
		{0x0F, "50797a5c06d53665d086f75e6f7c"},
		{0x48, "041f04400438043204350442"},
		{0x48, "d83dde00"},
		{0x48, "d83d"},
		{0x48, "004100"},
		{0x00, "31"},
		{0x13, "31"},
		{0x10, "31"},
		{0x44, "31"},
		{0x0F, ""},
	}
	for _, s := range seeds {
		f.Add(s.dcs, ussdMustHex(f, s.hex))
	}
	f.Fuzz(func(t *testing.T, dcs uint8, s []byte) {
		d := USSDDataCodingScheme(dcs)
		t1, err := d.Decode(s)
		if err != nil {
			return
		}
		if len(s) == 0 {
			t.Fatalf("Decode of empty input succeeded for DCS 0x%02X", dcs)
		}
		enc, err := d.Encode(t1)
		if err != nil {
			// Reserved codings are never sent, text can outgrow 160 octets, and
			// UCS2 rejects characters above U+FFFF.
			return
		}
		t2, err := d.Decode(enc)
		if err != nil {
			t.Fatalf("Decode(Encode(%q)) failed: %v", t1, err)
		}
		// 8n GSM 7 bit characters ending in <CR> gain a second <CR> on the
		// wire that the receiver keeps (TS 23.038 §6.1.2.3.1).
		if t2 != t1 && !(strings.HasSuffix(t1, "\r") && t2 == t1+"\r") {
			t.Fatalf("Decode(Encode(x)) = %q, want %q (dcs 0x%02X in=%x enc=%x)", t2, t1, dcs, s, enc)
		}
	})
}

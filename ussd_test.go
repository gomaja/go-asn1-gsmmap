// ussd_test.go
//
// Tests for the USSD types: USSDDataCodingScheme Encode / Decode,
// AlertingPattern, and the USSDArg / USSDRes converters with their
// Marshal() / Parse() entry points.

package gsmmap

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-sms/encoding/gsm7"
	"github.com/gomaja/go-sms/encoding/ucs2"
	"github.com/google/go-cmp/cmp"
)

func TestAlertingPatternString(t *testing.T) {
	cases := []struct {
		in   AlertingPattern
		want string
	}{
		{AlertingLevel0, "alertingLevel-0"},
		{AlertingLevel1, "alertingLevel-1"},
		{AlertingLevel2, "alertingLevel-2"},
		{AlertingCategory1, "alertingCategory-1"},
		{AlertingCategory2, "alertingCategory-2"},
		{AlertingCategory3, "alertingCategory-3"},
		{AlertingCategory4, "alertingCategory-4"},
		{0x03, "reserved(0x03)"},
		{0x08, "reserved(0x08)"},
		{0xFF, "reserved(0xFF)"},
	}
	for _, tc := range cases {
		if got := tc.in.String(); got != tc.want {
			t.Errorf("AlertingPattern(0x%02X).String() = %q, want %q", uint8(tc.in), got, tc.want)
		}
	}
}

func TestAlertingPatternValues(t *testing.T) {
	// 3GPP TS 29.002 V19.1.0 §17.7.8 octet values.
	want := map[AlertingPattern]uint8{
		AlertingLevel0: 0x00, AlertingLevel1: 0x01, AlertingLevel2: 0x02,
		AlertingCategory1: 0x04, AlertingCategory2: 0x05, AlertingCategory3: 0x06, AlertingCategory4: 0x07,
	}
	for p, v := range want {
		if uint8(p) != v {
			t.Errorf("%s = 0x%02X, want 0x%02X", p, uint8(p), v)
		}
	}
	if USSDDataCodingSchemeGSM7 != 0x0F || USSDDataCodingSchemeUCS2 != 0x48 {
		t.Errorf("DCS constants = 0x%02X, 0x%02X; want 0x0F, 0x48", uint8(USSDDataCodingSchemeGSM7), uint8(USSDDataCodingSchemeUCS2))
	}
}

func TestNetworkUnstructuredSsContextV2(t *testing.T) {
	want := []uint64{0, 4, 0, 0, 1, 0, 19, 2}
	got := NetworkUnstructuredSsContextV2()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
	got[7] = 99
	if again := NetworkUnstructuredSsContextV2(); !slices.Equal(again, want) {
		t.Errorf("mutating a returned slice changed the next result: %v", again)
	}
	if !slices.Equal(want, []uint64(gsm_map.NetworkUnstructuredSsContextV2())) {
		t.Error("differs from gsm_map.NetworkUnstructuredSsContextV2()")
	}
}

// ussdSpecClass classifies a data coding scheme straight from the bit string
// of the 3GPP TS 23.038 V20.0.0 §5 table, independently of the
// implementation's arithmetic.
func ussdSpecClass(d uint8) string {
	b := fmt.Sprintf("%08b", d) // b[0] is bit 7 ... b[7] is bit 0
	switch {
	case b[:4] == "0000":
		return "gsm7"
	case b == "00010000", b == "00010001", b == "00010010":
		return "unsupported" // language indication codings
	case b[:4] == "0001":
		return "reserved"
	case b[:4] == "0010", b[:4] == "0011":
		return "gsm7"
	case b[:2] == "01":
		if b[2] == '1' {
			return "unsupported" // compressed
		}
		switch b[4:6] {
		case "00":
			return "gsm7"
		case "01":
			return "unsupported" // 8 bit data
		case "10":
			return "ucs2"
		default:
			return "reserved"
		}
	case b[:4] == "1000", b[:4] == "1010", b[:4] == "1011", b[:4] == "1100":
		return "reserved"
	case b[:4] == "1001", b[:4] == "1101", b[:4] == "1110":
		return "unsupported" // UDH, I1, WAP
	default: // 1111
		if b[5] == '0' {
			return "gsm7"
		}
		return "unsupported" // 8 bit data
	}
}

func TestUSSDDataCodingSchemeClassificationAll256(t *testing.T) {
	gsm7Octets := ussdMustHex(t, "41e19058341e1b") // "ABCDEFG"
	ucs2Octets := ussdMustHex(t, "00410042")       // "AB"
	for v := 0; v < 256; v++ {
		d := USSDDataCodingScheme(v)
		class := ussdSpecClass(uint8(v))
		t.Run(fmt.Sprintf("0x%02X_%s", v, class), func(t *testing.T) {
			switch class {
			case "gsm7", "reserved":
				text, err := d.Decode(gsm7Octets)
				if err != nil || text != "ABCDEFG" {
					t.Errorf("Decode = %q, %v; want ABCDEFG", text, err)
				}
			case "ucs2":
				text, err := d.Decode(ucs2Octets)
				if err != nil || text != "AB" {
					t.Errorf("Decode = %q, %v; want AB", text, err)
				}
			case "unsupported":
				if _, err := d.Decode(gsm7Octets); !errors.Is(err, ErrUSSDUnsupportedDataCodingScheme) {
					t.Errorf("Decode err = %v, want ErrUSSDUnsupportedDataCodingScheme", err)
				}
			}
			out, err := d.Encode("ABCDEFG")
			switch class {
			case "gsm7":
				if err != nil || !bytes.Equal(out, gsm7Octets) {
					t.Errorf("Encode = %x, %v; want %x", out, err, gsm7Octets)
				}
			case "ucs2":
				if want := ussdMustHex(t, "0041004200430044004500460047"); err != nil || !bytes.Equal(out, want) {
					t.Errorf("Encode = %x, %v; want %x", out, err, want)
				}
			default: // reserved and unsupported are never sent
				if !errors.Is(err, ErrUSSDUnsupportedDataCodingScheme) {
					t.Errorf("Encode err = %v, want ErrUSSDUnsupportedDataCodingScheme", err)
				}
			}
		})
	}
}

func TestUSSDDataCodingSchemeUnsupportedReasons(t *testing.T) {
	cases := []struct {
		dcs    USSDDataCodingScheme
		reason string
	}{
		{0x10, "GSM 7 bit default alphabet with language indication"},
		{0x11, "UCS2 with language indication"},
		{0x12, "UCS2 with three-letter language indication"},
		{0x60, "compressed"},
		{0x44, "8 bit data"},
		{0xF4, "8 bit data"},
		{0x90, "user data header"},
		{0xD0, "I1 protocol"},
		{0xE0, "WAP"},
	}
	for _, tc := range cases {
		_, derr := tc.dcs.Decode([]byte{0x41})
		_, eerr := tc.dcs.Encode("A")
		for name, err := range map[string]error{"Decode": derr, "Encode": eerr} {
			if !errors.Is(err, ErrUSSDUnsupportedDataCodingScheme) {
				t.Errorf("0x%02X %s err = %v", uint8(tc.dcs), name, err)
				continue
			}
			if !strings.Contains(err.Error(), tc.reason) || !strings.Contains(err.Error(), fmt.Sprintf("0x%02X", uint8(tc.dcs))) {
				t.Errorf("0x%02X %s err = %q, want reason %q and the DCS in hex", uint8(tc.dcs), name, err, tc.reason)
			}
		}
	}
}

func TestUSSDDataCodingSchemeRoundTrip(t *testing.T) {
	gsm7DCS := []USSDDataCodingScheme{
		USSDDataCodingSchemeGSM7, 0x00, 0x01, 0x0E, 0x20, 0x24, 0x30, 0x3F, 0x40, 0x50, 0x10 | 0x40, 0xF0, 0xF3, 0x0F,
	}
	ucs2DCS := []USSDDataCodingScheme{USSDDataCodingSchemeUCS2, 0x58, 0x4B, 0x5B}
	extended := "€[]{}\\~^|"
	texts := map[string]string{
		"digits":       "*123#",
		"extension":    extended,
		"mixed":        "Balance: 12.50 € (£ ok) {a|b}",
		"cr in middle": "A\rB",
		"single":       "1",
		"eight":        "ABCDEFGH",
		"newline":      "line1\nline2",
	}
	for name, text := range texts {
		for _, d := range gsm7DCS {
			t.Run(fmt.Sprintf("gsm7/0x%02X/%s", uint8(d), name), func(t *testing.T) {
				enc, err := d.Encode(text)
				if err != nil {
					t.Fatalf("Encode: %v", err)
				}
				got, err := d.Decode(enc)
				if err != nil {
					t.Fatalf("Decode: %v", err)
				}
				if got != text {
					t.Errorf("round trip = %q, want %q", got, text)
				}
			})
		}
		for _, d := range ucs2DCS {
			t.Run(fmt.Sprintf("ucs2/0x%02X/%s", uint8(d), name), func(t *testing.T) {
				enc, err := d.Encode(text)
				if err != nil {
					t.Fatalf("Encode: %v", err)
				}
				got, err := d.Decode(enc)
				if err != nil {
					t.Fatalf("Decode: %v", err)
				}
				if got != text {
					t.Errorf("round trip = %q, want %q", got, text)
				}
			})
		}
	}
	// Non-Latin text only round-trips in UCS2.
	for _, text := range []string{"Привет", "你好", "مرحبا", "\U0000FFFF"} {
		enc, err := USSDDataCodingSchemeUCS2.Encode(text)
		if err != nil {
			t.Fatalf("UCS2 Encode(%q): %v", text, err)
		}
		got, err := USSDDataCodingSchemeUCS2.Decode(enc)
		if err != nil || got != text {
			t.Errorf("UCS2 round trip = %q, %v; want %q", got, err, text)
		}
	}
}

func TestUSSDDataCodingSchemeExtensionOctets(t *testing.T) {
	// Each extension-table character is ESC (0x1B) plus one septet
	// (3GPP TS 23.038 V20.0.0 §6.2.1.1), so "€" is 2 septets.
	septets, err := gsm7.Encode([]byte("€[]{}\\~^|"))
	if err != nil {
		t.Fatal(err)
	}
	if want := []byte{0x1B, 0x65, 0x1B, 0x3C, 0x1B, 0x3E, 0x1B, 0x28, 0x1B, 0x29, 0x1B, 0x2F, 0x1B, 0x3D, 0x1B, 0x14, 0x1B, 0x40}; !bytes.Equal(septets, want) {
		t.Fatalf("septets = %x, want %x", septets, want)
	}
	packed, err := gsm7.Pack7BitUSSD(septets)
	if err != nil {
		t.Fatal(err)
	}
	got, err := USSDDataCodingSchemeGSM7.Encode("€[]{}\\~^|")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, packed) {
		t.Errorf("Encode = %x, want %x", got, packed)
	}
}

func TestUSSDDataCodingSchemeMaxLength(t *testing.T) {
	cases := []struct {
		name    string
		dcs     USSDDataCodingScheme
		text    string
		wantLen int
		wantErr error
	}{
		{"gsm7 182 septets = 160 octets", USSDDataCodingSchemeGSM7, strings.Repeat("A", 182), 160, nil},
		{"gsm7 183 septets", USSDDataCodingSchemeGSM7, strings.Repeat("A", 183), 0, ErrUSSDTextTooLong},
		{"gsm7 91 extension chars = 182 septets", USSDDataCodingSchemeGSM7, strings.Repeat("€", 91), 160, nil},
		{"gsm7 92 extension chars", USSDDataCodingSchemeGSM7, strings.Repeat("€", 92), 0, ErrUSSDTextTooLong},
		{"gsm7 many", USSDDataCodingSchemeGSM7, strings.Repeat("A", 1000), 0, ErrUSSDTextTooLong},
		{"ucs2 80 chars = 160 octets", USSDDataCodingSchemeUCS2, strings.Repeat("Ж", 80), 160, nil},
		{"ucs2 81 chars", USSDDataCodingSchemeUCS2, strings.Repeat("Ж", 81), 0, ErrUSSDTextTooLong},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := tc.dcs.Encode(tc.text)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(out) != tc.wantLen {
				t.Errorf("len = %d, want %d", len(out), tc.wantLen)
			}
			back, err := tc.dcs.Decode(out)
			if err != nil || back != tc.text {
				t.Errorf("round trip mismatch: err=%v", err)
			}
		})
	}
}

func TestUSSDDataCodingSchemeEncodeErrors(t *testing.T) {
	t.Run("empty text", func(t *testing.T) {
		for _, d := range []USSDDataCodingScheme{USSDDataCodingSchemeGSM7, USSDDataCodingSchemeUCS2} {
			if _, err := d.Encode(""); !errors.Is(err, ErrUSSDTextEmpty) {
				t.Errorf("0x%02X: err = %v, want ErrUSSDTextEmpty", uint8(d), err)
			}
		}
	})
	t.Run("gsm7 unencodable", func(t *testing.T) {
		_, err := USSDDataCodingSchemeGSM7.Encode("ab日")
		var u gsm7.ErrUnencodable
		if !errors.As(err, &u) || u.Rune != '日' {
			t.Errorf("err = %v, want wrapped gsm7.ErrUnencodable for U+65E5", err)
		}
	})
	t.Run("gsm7 invalid UTF-8", func(t *testing.T) {
		_, err := USSDDataCodingSchemeGSM7.Encode("a\xffb")
		var u gsm7.ErrInvalidUTF8
		if !errors.As(err, &u) {
			t.Errorf("err = %v, want wrapped gsm7.ErrInvalidUTF8", err)
		}
	})
	t.Run("ucs2 invalid UTF-8", func(t *testing.T) {
		if _, err := USSDDataCodingSchemeUCS2.Encode("a\xffb"); err == nil {
			t.Error("want an error, got nil")
		}
	})
	t.Run("ucs2 above U+FFFF", func(t *testing.T) {
		_, err := USSDDataCodingSchemeUCS2.Encode("ab\U0001F600")
		if err == nil || !strings.Contains(err.Error(), "U+1F600") {
			t.Errorf("err = %v, want error naming U+1F600", err)
		}
	})
}

func TestUSSDDataCodingSchemeDecodeErrors(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		for _, d := range []USSDDataCodingScheme{USSDDataCodingSchemeGSM7, USSDDataCodingSchemeUCS2, 0x03, 0x2A, 0x80} {
			if _, err := d.Decode(nil); !errors.Is(err, ErrUSSDTextEmpty) {
				t.Errorf("0x%02X nil: err = %v, want ErrUSSDTextEmpty", uint8(d), err)
			}
			if _, err := d.Decode([]byte{}); !errors.Is(err, ErrUSSDTextEmpty) {
				t.Errorf("0x%02X empty: err = %v, want ErrUSSDTextEmpty", uint8(d), err)
			}
		}
	})
	t.Run("ucs2 odd length", func(t *testing.T) {
		_, err := USSDDataCodingSchemeUCS2.Decode([]byte{0x00, 0x41, 0x00})
		if !errors.Is(err, ucs2.ErrInvalidLength) {
			t.Errorf("err = %v, want ucs2.ErrInvalidLength", err)
		}
	})
	t.Run("ucs2 dangling high surrogate", func(t *testing.T) {
		_, err := USSDDataCodingSchemeUCS2.Decode([]byte{0x00, 0x41, 0xD8, 0x3D})
		var d ucs2.ErrDanglingSurrogate
		if !errors.As(err, &d) {
			t.Errorf("err = %v, want ucs2.ErrDanglingSurrogate", err)
		}
	})
}

func TestUSSDDataCodingSchemeDecodeCRHandling(t *testing.T) {
	// 7 octets ending on an octet boundary with a final <CR>: the receiver
	// removes it (3GPP TS 23.038 V20.0.0 §6.1.2.3.1).
	got, err := USSDDataCodingSchemeGSM7.Decode(ussdMustHex(t, "41e19058341e9149e592d9743e1b0d"))
	if err != nil || got != "ABCDEFGHIJKLMNO\r\r" {
		t.Errorf("15 octets: %q, %v", got, err)
	}
	// 7 octets, last septet CR: removed.
	got, err = USSDDataCodingSchemeGSM7.Decode(ussdMustHex(t, "41e19058341e1b"))
	if err != nil || got != "ABCDEFG" {
		t.Errorf("7 octets: %q, %v", got, err)
	}
}

func TestUSSDDataCodingSchemeDecodeNonBMP(t *testing.T) {
	// go-sms decodes a UTF-16 surrogate pair; Decode is lenient and keeps
	// what it yields.
	got, err := USSDDataCodingSchemeUCS2.Decode([]byte{0xD8, 0x3D, 0xDE, 0x00})
	if err != nil || got != "\U0001F600" {
		t.Errorf("Decode = %q, %v", got, err)
	}
}

func ussdAllPatterns() []AlertingPattern {
	return []AlertingPattern{AlertingLevel0, AlertingLevel1, AlertingLevel2, AlertingCategory1, AlertingCategory2, AlertingCategory3, AlertingCategory4}
}

func ussdText(t *testing.T, d USSDDataCodingScheme, text string) HexBytes {
	t.Helper()
	b, err := d.Encode(text)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestUSSDArgRoundTrip(t *testing.T) {
	type tc struct {
		name string
		in   *USSDArg
	}
	gsm := ussdText(t, USSDDataCodingSchemeGSM7, "*100#")
	cases := []tc{
		{"minimal", &USSDArg{DataCodingScheme: USSDDataCodingSchemeGSM7, USSDString: gsm}},
		{"single octet string", &USSDArg{DataCodingScheme: USSDDataCodingSchemeGSM7, USSDString: []byte{0x31}}},
		{"160 octet gsm7", &USSDArg{DataCodingScheme: USSDDataCodingSchemeGSM7, USSDString: ussdText(t, USSDDataCodingSchemeGSM7, strings.Repeat("Z", 182))}},
		{"160 octet ucs2", &USSDArg{DataCodingScheme: USSDDataCodingSchemeUCS2, USSDString: ussdText(t, USSDDataCodingSchemeUCS2, strings.Repeat("Ж", 80))}},
		{"ucs2 content", &USSDArg{DataCodingScheme: USSDDataCodingSchemeUCS2, USSDString: ussdText(t, USSDDataCodingSchemeUCS2, "Привет")}},
		{"gsm7 extension content", &USSDArg{DataCodingScheme: USSDDataCodingSchemeGSM7, USSDString: ussdText(t, USSDDataCodingSchemeGSM7, "€[]{}\\~^|")}},
		{"other dcs values", &USSDArg{DataCodingScheme: 0x11, USSDString: []byte{0x01, 0x02}}},
	}
	for _, p := range ussdAllPatterns() {
		p := p
		cases = append(cases, tc{"alerting " + p.String(), &USSDArg{DataCodingScheme: USSDDataCodingSchemeGSM7, USSDString: gsm, AlertingPattern: &p}})
	}
	for _, m := range []struct {
		digits        string
		nature, plan  uint8
		defaultedFrom bool
	}{
		{"27761485722", 0x10, 0x01, false},
		{"0612345678", 0x20, 0x08, false},
		{"123", 0x40, 0x09, false},
		{"98765", 0x60, 0x03, false},
		{"41791234567", 0x10, 0x06, false},
		{"12345", 0x00, 0x00, true}, // defaults: international / ISDN
	} {
		name := fmt.Sprintf("msisdn %s nature=0x%02X plan=0x%02X", m.digits, m.nature, m.plan)
		in := &USSDArg{DataCodingScheme: USSDDataCodingSchemeGSM7, USSDString: gsm, MSISDN: m.digits, MSISDNNature: m.nature, MSISDNPlan: m.plan}
		if m.defaultedFrom {
			// Zero nature and plan encode as the package defaults.
			data, err := in.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			got, err := ParseUSSDArg(data)
			if err != nil {
				t.Fatal(err)
			}
			if got.MSISDNNature != 0x10 || got.MSISDNPlan != 0x01 || got.MSISDN != m.digits {
				t.Errorf("%s: got %+v, want international/ISDN defaults", name, got)
			}
			continue
		}
		cases = append(cases, tc{name, in})
	}
	ap := AlertingCategory3
	cases = append(cases, tc{"alerting + msisdn", &USSDArg{
		DataCodingScheme: USSDDataCodingSchemeUCS2, USSDString: ussdText(t, USSDDataCodingSchemeUCS2, "Hi"),
		AlertingPattern: &ap, MSISDN: "27761485722", MSISDNNature: 0x10, MSISDNPlan: 0x01,
	}})

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			data, err := c.in.Marshal()
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			got, err := ParseUSSDArg(data)
			if err != nil {
				t.Fatalf("ParseUSSDArg: %v", err)
			}
			if diff := cmp.Diff(c.in, got); diff != "" {
				t.Errorf("round trip mismatch (-want +got):\n%s", diff)
			}
			// The wire converters round-trip without BER as well.
			w, err := convertUSSDArgToWire(c.in)
			if err != nil {
				t.Fatal(err)
			}
			back, err := convertWireToUSSDArg(w)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(c.in, back); diff != "" {
				t.Errorf("converter round trip mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUSSDArgMarshalDoesNotAliasInput(t *testing.T) {
	in := &USSDArg{DataCodingScheme: USSDDataCodingSchemeGSM7, USSDString: []byte{0x31, 0x32}}
	w, err := convertUSSDArgToWire(in)
	if err != nil {
		t.Fatal(err)
	}
	w.UssdString[0] = 0xFF
	if in.USSDString[0] != 0x31 {
		t.Error("converter output aliases the caller's USSDString")
	}
	out, err := convertWireToUSSDArg(w)
	if err != nil {
		t.Fatal(err)
	}
	out.USSDString[1] = 0xEE
	if w.UssdString[1] != 0x32 {
		t.Error("converter output aliases the wire USSDString")
	}
}

func TestUSSDResRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   *USSDRes
	}{
		{"gsm7", &USSDRes{DataCodingScheme: USSDDataCodingSchemeGSM7, USSDString: ussdText(t, USSDDataCodingSchemeGSM7, "Thank you")}},
		{"ucs2", &USSDRes{DataCodingScheme: USSDDataCodingSchemeUCS2, USSDString: ussdText(t, USSDDataCodingSchemeUCS2, "Спасибо")}},
		{"160 octets", &USSDRes{DataCodingScheme: USSDDataCodingSchemeGSM7, USSDString: ussdText(t, USSDDataCodingSchemeGSM7, strings.Repeat("Q", 182))}},
		{"single octet", &USSDRes{DataCodingScheme: 0x0F, USSDString: []byte{0x31}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			data, err := c.in.Marshal()
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			got, err := ParseUSSDRes(data)
			if err != nil {
				t.Fatalf("ParseUSSDRes: %v", err)
			}
			if diff := cmp.Diff(c.in, got); diff != "" {
				t.Errorf("round trip mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUSSDNilReceivers(t *testing.T) {
	var a *USSDArg
	if _, err := a.Marshal(); !errors.Is(err, ErrUSSDArgNil) {
		t.Errorf("nil USSDArg.Marshal err = %v", err)
	}
	var r *USSDRes
	if _, err := r.Marshal(); !errors.Is(err, ErrUSSDResNil) {
		t.Errorf("nil USSDRes.Marshal err = %v", err)
	}
	if _, err := convertWireToUSSDArg(nil); !errors.Is(err, ErrUSSDArgNil) {
		t.Errorf("convertWireToUSSDArg(nil) err = %v", err)
	}
	if _, err := convertWireToUSSDRes(nil); !errors.Is(err, ErrUSSDResNil) {
		t.Errorf("convertWireToUSSDRes(nil) err = %v", err)
	}
	if _, err := convertUSSDArgToWire(nil); !errors.Is(err, ErrUSSDArgNil) {
		t.Errorf("convertUSSDArgToWire(nil) err = %v", err)
	}
	if _, err := convertUSSDResToWire(nil); !errors.Is(err, ErrUSSDResNil) {
		t.Errorf("convertUSSDResToWire(nil) err = %v", err)
	}
}

func TestUSSDArgMarshalReservedAlertingPattern(t *testing.T) {
	for _, v := range []AlertingPattern{0x03, 0x08, 0x0F, 0x10, 0x80, 0xFF} {
		v := v
		a := &USSDArg{DataCodingScheme: USSDDataCodingSchemeGSM7, USSDString: []byte{0x31}, AlertingPattern: &v}
		_, err := a.Marshal()
		if !errors.Is(err, ErrUSSDReservedAlertingPattern) {
			t.Errorf("0x%02X: err = %v, want ErrUSSDReservedAlertingPattern", uint8(v), err)
		}
	}
}

func TestUSSDArgMarshalBadMSISDN(t *testing.T) {
	a := &USSDArg{DataCodingScheme: USSDDataCodingSchemeGSM7, USSDString: []byte{0x31}, MSISDN: "12x4"}
	if _, err := a.Marshal(); err == nil {
		t.Error("want an error for a non-digit MSISDN, got nil")
	}
}

// ussdWireArg BER-encodes a hand-built wire USSD-Arg.
func ussdWireArg(t *testing.T, w *gsm_map.USSDArg) []byte {
	t.Helper()
	data, err := w.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}
	return data
}

func TestParseUSSDWireLengthErrors(t *testing.T) {
	str := []byte{0x31}
	for _, dcs := range [][]byte{{}, {0x0F, 0x0F}, {0x0F, 0x00, 0x00}} {
		t.Run(fmt.Sprintf("arg dcs %x", dcs), func(t *testing.T) {
			data := ussdWireArg(t, &gsm_map.USSDArg{UssdDataCodingScheme: dcs, UssdString: str})
			if _, err := ParseUSSDArg(data); !errors.Is(err, ErrUSSDInvalidDataCodingSchemeLength) {
				t.Errorf("err = %v, want ErrUSSDInvalidDataCodingSchemeLength", err)
			}
		})
		t.Run(fmt.Sprintf("res dcs %x", dcs), func(t *testing.T) {
			data, err := (&gsm_map.USSDRes{UssdDataCodingScheme: dcs, UssdString: str}).MarshalBER()
			if err != nil {
				t.Fatal(err)
			}
			if _, err := ParseUSSDRes(data); !errors.Is(err, ErrUSSDInvalidDataCodingSchemeLength) {
				t.Errorf("err = %v, want ErrUSSDInvalidDataCodingSchemeLength", err)
			}
		})
	}
	for _, ap := range [][]byte{{}, {0x01, 0x01}, {0x00, 0x00, 0x00}} {
		t.Run(fmt.Sprintf("alerting pattern %x", ap), func(t *testing.T) {
			ap := gsm_map.AlertingPattern(ap)
			data := ussdWireArg(t, &gsm_map.USSDArg{UssdDataCodingScheme: []byte{0x0F}, UssdString: str, AlertingPattern: &ap})
			if _, err := ParseUSSDArg(data); !errors.Is(err, ErrUSSDInvalidAlertingPatternLength) {
				t.Errorf("err = %v, want ErrUSSDInvalidAlertingPatternLength", err)
			}
		})
	}
}

func TestParseUSSDArgLenientAlertingPattern(t *testing.T) {
	// Decode keeps any single octet; Marshal is strict.
	for _, v := range []byte{0x03, 0x08, 0xFF} {
		ap := gsm_map.AlertingPattern{v}
		data := ussdWireArg(t, &gsm_map.USSDArg{UssdDataCodingScheme: []byte{0x0F}, UssdString: []byte{0x31}, AlertingPattern: &ap})
		got, err := ParseUSSDArg(data)
		if err != nil {
			t.Fatalf("0x%02X: %v", v, err)
		}
		if got.AlertingPattern == nil || uint8(*got.AlertingPattern) != v {
			t.Errorf("0x%02X: got %v", v, got.AlertingPattern)
		}
		if _, err := got.Marshal(); !errors.Is(err, ErrUSSDReservedAlertingPattern) {
			t.Errorf("0x%02X: Marshal err = %v", v, err)
		}
	}
}

func TestParseUSSDArgMSISDNDecodedEmpty(t *testing.T) {
	// An address octet with no digits: presence cannot round-trip.
	m := gsm_map.ISDNAddressString{0x91}
	data := ussdWireArg(t, &gsm_map.USSDArg{UssdDataCodingScheme: []byte{0x0F}, UssdString: []byte{0x31}, Msisdn: &m})
	if _, err := ParseUSSDArg(data); !errors.Is(err, ErrUSSDMSISDNDecodedEmpty) {
		t.Errorf("err = %v, want ErrUSSDMSISDNDecodedEmpty", err)
	}
	// A zero-octet address is rejected by the TBCD decoder.
	m = gsm_map.ISDNAddressString{}
	data = ussdWireArg(t, &gsm_map.USSDArg{UssdDataCodingScheme: []byte{0x0F}, UssdString: []byte{0x31}, Msisdn: &m})
	if _, err := ParseUSSDArg(data); err == nil {
		t.Error("zero-octet MSISDN: want an error, got nil")
	}
}

func TestParseUSSDIndefiniteLength(t *testing.T) {
	// BER indefinite length is accepted and re-encoded in definite form.
	data := []byte{0x30, 0x80, 0x04, 0x01, 0x0f, 0x04, 0x01, 0x31, 0x00, 0x00}
	arg, err := ParseUSSDArg(data)
	if err != nil {
		t.Fatal(err)
	}
	want := &USSDArg{DataCodingScheme: USSDDataCodingSchemeGSM7, USSDString: []byte{0x31}}
	if diff := cmp.Diff(want, arg); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
	out, err := arg.Marshal()
	if err != nil || hex.EncodeToString(out) != "300604010f040131" {
		t.Errorf("Marshal = %x, %v", out, err)
	}
	res, err := ParseUSSDRes(data)
	if err != nil || res.DataCodingScheme != 0x0F || !bytes.Equal(res.USSDString, []byte{0x31}) {
		t.Errorf("ParseUSSDRes = %+v, %v", res, err)
	}
}

func TestParseUSSDMalformed(t *testing.T) {
	good := ussdMustHex(t, "301c04010f040eaa180da682dd6c31192d36bbdd468007917267415827f2")
	cases := map[string][]byte{
		"empty":            {},
		"truncated":        good[:len(good)-3],
		"wrong outer tag":  append([]byte{0x31}, good[1:]...),
		"missing string":   {0x30, 0x03, 0x04, 0x01, 0x0f},
		"integer instead":  {0x30, 0x06, 0x02, 0x01, 0x0f, 0x02, 0x01, 0x31},
		"length overflows": {0x30, 0x84, 0xff, 0xff, 0xff, 0xff},
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if got, err := ParseUSSDArg(data); err == nil {
				t.Errorf("ParseUSSDArg(%x) = %+v, want an error", data, got)
			}
			if got, err := ParseUSSDRes(data); err == nil && name != "" {
				// A USSD-Arg with extra components is a valid USSD-Res
				// only if it is the two-field form; none of these are.
				t.Errorf("ParseUSSDRes(%x) = %+v, want an error", data, got)
			}
		})
	}
}

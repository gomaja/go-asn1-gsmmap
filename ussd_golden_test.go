// ussd_golden_test.go
//
// Golden BER vectors for USSD-Arg and USSD-Res, each checked against
// Wireshark 4.6.4 (tshark). The first vector is the parameter of the public
// sample capture gsm_map_with_ussd_string.pcap from the Wireshark
// SampleCaptures wiki (sha256
// b5e755469c67d1aba94901c55114fd3acc088cff70dde0375cad948efba63a21); the
// others were produced by USSDArg.Marshal / USSDRes.Marshal and then wrapped
// in TCAP / SCCP / MTP3 by an external harness and decoded by tshark, with no
// malformed or expert-info entries in any case.

package gsmmap

import (
	"encoding/hex"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func ussdAlerting(p AlertingPattern) *AlertingPattern { return &p }

func TestUSSDArgGoldenVectors(t *testing.T) {
	cases := []struct {
		name string
		hex  string
		want USSDArg
		text string
	}{
		{
			// gsm_map_with_ussd_string.pcap: TCAP Begin, AC 0.4.0.0.1.0.19.2,
			// invoke 1, processUnstructuredSS-Request (59). Wireshark: DCS 0f
			// (Coding Group 0, language unspecified), USSD String
			// "*140*0761241377#", msisdn 917267415827f2 = no extension,
			// international, E.164, 27761485722 (South Africa), no
			// alertingPattern.
			name: "sample capture op 59 invoke",
			hex:  "301c04010f040eaa180da682dd6c31192d36bbdd468007917267415827f2",
			want: USSDArg{
				DataCodingScheme: USSDDataCodingSchemeGSM7,
				USSDString:       ussdMustHex(t, "aa180da682dd6c31192d36bbdd46"),
				MSISDN:           "27761485722",
				MSISDNNature:     0x10,
				MSISDNPlan:       0x01,
			},
			text: "*140*0761241377#",
		},
		{
			// op 60 invoke with alertingPattern. Wireshark 4.6.4: unstructuredSS-Request (60),
			// DCS 0f, USSD String "Enter PIN:", alertingPattern 05
			// (alertingCategory-2), no msisdn.
			name: "op 60 invoke with alerting pattern",
			hex:  "301104010f04094537bd2c0741934e1d040105",
			want: USSDArg{
				DataCodingScheme: USSDDataCodingSchemeGSM7,
				USSDString:       ussdMustHex(t, "4537bd2c0741934e1d"),
				AlertingPattern:  ussdAlerting(AlertingCategory2),
			},
			text: "Enter PIN:",
		},
		{
			// op 61 invoke. Wireshark 4.6.4: unstructuredSS-Notify (61), DCS 0f,
			// USSD String "Price 5€ [ok]" (default alphabet extension table
			// characters decoded), alertingPattern 01 (alertingLevel-1).
			name: "op 61 invoke with extension characters",
			hex:  "301604010f040e50797a5c06d53665d086f75e6f7c040101",
			want: USSDArg{
				DataCodingScheme: USSDDataCodingSchemeGSM7,
				USSDString:       ussdMustHex(t, "50797a5c06d53665d086f75e6f7c"),
				AlertingPattern:  ussdAlerting(AlertingLevel1),
			},
			text: "Price 5€ [ok]",
		},
		{
			// op 61 invoke, UCS2. Wireshark 4.6.4: unstructuredSS-Notify (61),
			// DCS 48 (general data coding, uncompressed, no message class,
			// UCS2 16 bit), USSD String "Привет".
			name: "op 61 invoke UCS2",
			hex:  "3011040148040c041f04400438043204350442",
			want: USSDArg{
				DataCodingScheme: USSDDataCodingSchemeUCS2,
				USSDString:       ussdMustHex(t, "041f04400438043204350442"),
			},
			text: "Привет",
		},
		{
			// op 59 invoke with a non-default MSISDN. Wireshark 4.6.4:
			// processUnstructuredSS-Request (59), DCS 0f, USSD String "*100#",
			// msisdn a86021436587 = no extension, nature National Significant
			// Number (0x2), numbering plan National (0x8), digits 0612345678.
			name: "op 59 invoke MSISDN national / national plan",
			hex:  "301204010f0405aa180c36028006a86021436587",
			want: USSDArg{
				DataCodingScheme: USSDDataCodingSchemeGSM7,
				USSDString:       ussdMustHex(t, "aa180c3602"),
				MSISDN:           "0612345678",
				MSISDNNature:     0x20,
				MSISDNPlan:       0x08,
			},
			text: "*100#",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := ussdMustHex(t, tc.hex)
			got, err := ParseUSSDArg(data)
			if err != nil {
				t.Fatalf("ParseUSSDArg: %v", err)
			}
			if diff := cmp.Diff(&tc.want, got); diff != "" {
				t.Errorf("ParseUSSDArg mismatch (-want +got):\n%s", diff)
			}
			text, err := got.DataCodingScheme.Decode(got.USSDString)
			if err != nil {
				t.Fatalf("Decode: %v", err)
			}
			if text != tc.text {
				t.Errorf("Decode = %q, want %q", text, tc.text)
			}
			out, err := got.Marshal()
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			if hex.EncodeToString(out) != tc.hex {
				t.Errorf("Marshal = %x, want %s", out, tc.hex)
			}
			enc, err := got.DataCodingScheme.Encode(tc.text)
			if err != nil {
				t.Fatalf("Encode: %v", err)
			}
			if diff := cmp.Diff([]byte(tc.want.USSDString), enc); diff != "" {
				t.Errorf("Encode(%q) mismatch (-want +got):\n%s", tc.text, diff)
			}
		})
	}
}

func TestUSSDResGoldenVectors(t *testing.T) {
	cases := []struct {
		name string
		hex  string
		want USSDRes
		text string
	}{
		{
			// op 59 returnResultLast. Wireshark 4.6.4: processUnstructuredSS-Request
			// (59) result, DCS 0f, USSD String "Welcome!".
			name: "op 59 result",
			hex:  "300c04010f0407d7327bfc6e9743",
			want: USSDRes{
				DataCodingScheme: USSDDataCodingSchemeGSM7,
				USSDString:       ussdMustHex(t, "d7327bfc6e9743"),
			},
			text: "Welcome!",
		},
		{
			// op 60 returnResultLast. Wireshark 4.6.4: unstructuredSS-Request (60)
			// result, DCS 0f, USSD String "1".
			name: "op 60 result",
			hex:  "300604010f040131",
			want: USSDRes{
				DataCodingScheme: USSDDataCodingSchemeGSM7,
				USSDString:       ussdMustHex(t, "31"),
			},
			text: "1",
		},
		{
			// op 59 returnResultLast, UCS2. Wireshark 4.6.4: DCS 48 (general
			// data coding, uncompressed, UCS2 16 bit), USSD String "Привет €".
			name: "op 59 result UCS2",
			hex:  "30150401480410041f04400438043204350442002020ac",
			want: USSDRes{
				DataCodingScheme: USSDDataCodingSchemeUCS2,
				USSDString:       ussdMustHex(t, "041f04400438043204350442002020ac"),
			},
			text: "Привет €",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := ussdMustHex(t, tc.hex)
			got, err := ParseUSSDRes(data)
			if err != nil {
				t.Fatalf("ParseUSSDRes: %v", err)
			}
			if diff := cmp.Diff(&tc.want, got); diff != "" {
				t.Errorf("ParseUSSDRes mismatch (-want +got):\n%s", diff)
			}
			text, err := got.DataCodingScheme.Decode(got.USSDString)
			if err != nil {
				t.Fatalf("Decode: %v", err)
			}
			if text != tc.text {
				t.Errorf("Decode = %q, want %q", text, tc.text)
			}
			out, err := got.Marshal()
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			if hex.EncodeToString(out) != tc.hex {
				t.Errorf("Marshal = %x, want %s", out, tc.hex)
			}
		})
	}
}

func ussdMustHex(t testing.TB, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

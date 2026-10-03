// semantic_imei_test.go
//
// 3GPP TS 29.002 V19.1.0 §17.7.8: "IMEI ::= TBCD-STRING (SIZE (8)) -- ... If
// the SVN is not present the last octet shall contain the digit 0 and a
// filler. If present the SVN shall be included in the last octet." A
// 15-digit IMEI is the TAC and SNR (14 digits) and the spare digit 0; the
// Check Digit is not carried (3GPP TS 23.003 V20.1.0 §6.2.1). Marshal sends
// the spare digit as 0; Parse also accepts a peer's Check Digit there, a
// value that then does not marshal again.
package gsmmap

import (
	"errors"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

func TestMarshalIMEISpareDigit(t *testing.T) {
	for _, c := range semIMEICases() {
		for _, tc := range []struct {
			imei string
			want error
		}{
			{"490154203237510", nil},                      // TAC, SNR, spare digit 0
			{"490154203237511", ErrIMEISpareDigitNotZero}, // a check digit in the spare position
			{"490154203237518", ErrIMEISpareDigitNotZero},
			{"490154203237519", ErrIMEISpareDigitNotZero},
			{"4901542032375100", nil}, // TAC, SNR, SVN 00
			{"4901542032375199", nil}, // SVN 99: any SVN digits
		} {
			t.Run(c.name+"/"+tc.imei, func(t *testing.T) {
				data, err := c.marshal(t, tc.imei)
				if !errors.Is(err, tc.want) {
					t.Fatalf("Marshal: err = %v, want %v", err, tc.want)
				}
				if err != nil {
					return
				}
				got, err := c.parse(data)
				if err != nil {
					t.Fatalf("Parse: %v", err)
				}
				if got != tc.imei {
					t.Errorf("IMEI = %q, want %q", got, tc.imei)
				}
			})
		}
	}
}

// Parse accepts a 15-digit IMEI whatever its last digit.
func TestParseIMEICheckDigitTolerated(t *testing.T) {
	for _, imei := range []string{"490154203237510", "490154203237518"} {
		raw := semTBCD(t, imei) // last octet: filler, 15th digit
		w := &gsm_map.AnyTimeInterrogationRes{SubscriberInfo: gsm_map.SubscriberInfo{Imei: &raw}}
		data, err := w.MarshalBER()
		if err != nil {
			t.Fatalf("MarshalBER: %v", err)
		}
		got, err := ParseAnyTimeInterrogationRes(data)
		if err != nil {
			t.Fatalf("%s: ParseAnyTimeInterrogationRes: %v", imei, err)
		}
		if got.SubscriberInfo.IMEI != imei {
			t.Errorf("IMEI = %q, want %q", got.SubscriberInfo.IMEI, imei)
		}
		_, err = got.Marshal()
		if imei[14] == '0' && err != nil {
			t.Errorf("%s: Marshal: %v", imei, err)
		}
		if imei[14] != '0' && !errors.Is(err, ErrIMEISpareDigitNotZero) {
			t.Errorf("%s: Marshal: err = %v, want ErrIMEISpareDigitNotZero", imei, err)
		}
	}
}

// The spare digit rule is the IMEI's; the IMEISV has 16 digits.
func TestIMEISVUnaffectedBySpareDigit(t *testing.T) {
	for _, c := range semIMEISVCases() {
		data, err := c.marshal(t, "4901542032375188")
		if err != nil {
			t.Fatalf("%s: Marshal: %v", c.name, err)
		}
		if got, err := c.parse(data); err != nil || got != "4901542032375188" {
			t.Errorf("%s: Parse = %q, %v", c.name, got, err)
		}
	}
}

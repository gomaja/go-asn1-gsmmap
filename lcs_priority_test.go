package gsmmap

import (
	"errors"
	"fmt"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// 3GPP TS 29.002 V19.1.0 §17.7.13 LCS-Priority ::= OCTET STRING (SIZE (1)):
// "0 = highest priority", "1 = normal priority", "all other values treated
// as 1".

func TestMarshalLCSPriority(t *testing.T) {
	for _, p := range []LCSPriority{{0x00}, {0x01}} {
		a := pslArg()
		a.LcsPriority = p
		data, err := a.Marshal()
		if err != nil {
			t.Fatalf("Marshal %x: %v", p, err)
		}
		got, err := ParseProvideSubscriberLocation(data)
		if err != nil {
			t.Fatalf("Parse %x: %v", p, err)
		}
		wantEqual(t, "LcsPriority", p, got.LcsPriority)
	}
	for _, p := range []LCSPriority{{0x02}, {0x80}, {0xff}} {
		a := pslArg()
		a.LcsPriority = p
		if _, err := a.Marshal(); !errors.Is(err, ErrLCSPriorityInvalid) {
			t.Errorf("Marshal %x: err = %v, want ErrLCSPriorityInvalid", p, err)
		}
	}
}

func TestParseLCSPriorityMapsOtherValuesToNormal(t *testing.T) {
	for _, tc := range []struct{ wire, want byte }{
		{0x00, 0x00},
		{0x01, 0x01},
		{0x02, 0x01}, // first value treated as 1
		{0x7f, 0x01},
		{0xff, 0x01},
	} {
		t.Run(fmt.Sprintf("%02x", tc.wire), func(t *testing.T) {
			w := pslWire(t, pslArg())
			w.LcsPriority = &gsm_map.LCSPriority{tc.wire}
			data := strictBER(t, w)
			got, err := ParseProvideSubscriberLocation(data)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			wantEqual(t, "LcsPriority", LCSPriority{tc.want}, got.LcsPriority)
			if err := checkParseRoundTrip("ProvideSubscriberLocation", asParser(ParseProvideSubscriberLocation), data); err != nil {
				t.Error(err)
			}
		})
	}
}

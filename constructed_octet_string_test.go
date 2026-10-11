package gsmmap

import (
	"encoding/hex"
	"testing"

	"github.com/gomaja/go-asn1-gsmmap/address"
)

// ITU-T X.690 (02/2021) §8.7.3 permits constructed OCTET STRING encodings;
// 3GPP TS 29.002 V19.1.0 §17.7.1 defines AnyTimeInterrogationArg.
func TestAnyTimeInterrogationConstructedOctetStrings(t *testing.T) {
	for _, tc := range []struct {
		name, wire string
	}{
		{"primitive", "3014a00880060001012143f5a1028000830491132654"},
		{"constructed IMSI", "3016a00aa00804060001012143f5a1028000830491132654"},
		{"segmented IMSI", "3018a00ca00a040200010404012143f5a1028000830491132654"},
		{"constructed gsmSCF-Address", "3018a00880060001012143f5a1028000a3080402911304022654"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := hex.DecodeString(tc.wire)
			if err != nil {
				t.Fatal(err)
			}
			ati, err := ParseAnyTimeInterrogation(data)
			if err != nil {
				t.Fatal(err)
			}
			if got := ati.SubscriberIdentity.IMSI; got != "00101012345" {
				t.Errorf("IMSI = %q, want 00101012345", got)
			}
			if ati.GsmSCFAddress != "316245" || ati.GsmSCFAddressNature != address.NatureInternational || ati.GsmSCFAddressPlan != address.PlanISDN {
				t.Errorf("gsmSCF-Address = %q nature 0x%02x plan 0x%02x, want 316245 international ISDN", ati.GsmSCFAddress, ati.GsmSCFAddressNature, ati.GsmSCFAddressPlan)
			}
			if err := checkParseRoundTrip(tc.name, asParser(ParseAnyTimeInterrogation), data); err != nil {
				t.Error(err)
			}
		})
	}
}

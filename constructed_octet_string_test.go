package gsmmap

import (
	"encoding/hex"
	"testing"

	"github.com/gomaja/go-asn1-gsmmap/address"
)

// Parse accepts any BER of ITU-T X.690 (02/2021): constructed OCTET STRINGs
// (§8.7.3), long-form lengths (§8.1.3.5) and the indefinite form (§8.1.3.6).
// Marshal sends the form 3GPP TS 29.002 V19.1.0 §17.1.1 requires: primitive
// strings and short-form lengths below 128 octets. AnyTimeInterrogationArg is
// defined in §17.7.1.
func TestAnyTimeInterrogationConstructedOctetStrings(t *testing.T) {
	const primitive = "3014a00880060001012143f5a1028000830491132654"
	for _, tc := range []struct {
		name, wire string
	}{
		{"primitive", primitive},
		{"constructed IMSI", "3016a00aa00804060001012143f5a1028000830491132654"},
		{"segmented IMSI", "3018a00ca00a040200010404012143f5a1028000830491132654"},
		{"constructed gsmSCF-Address", "3018a00880060001012143f5a1028000a3080402911304022654"},
		{"long-form IMSI length", "3015a0098081060001012143f5a1028000830491132654"},
		{"indefinite constructed IMSI", "3018a00ca08004060001012143f50000a1028000830491132654"},
		{"indefinite argument", "3080a00880060001012143f5a10280008304911326540000"},
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
			out, err := ati.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			if got := hex.EncodeToString(out); got != primitive {
				t.Errorf("Marshal = %s, want %s", got, primitive)
			}
		})
	}
}

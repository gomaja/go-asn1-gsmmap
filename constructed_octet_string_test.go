package gsmmap

import (
	"encoding/hex"
	"errors"
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime/ber"

	"github.com/gomaja/go-asn1-gsmmap/address"
)

// wantMAPProfileError checks that err is go-asn1's located refusal of a
// received form that 3GPP TS 29.002 V19.1.0 §17.1.1 excludes.
func wantMAPProfileError(t *testing.T, err error, path, rule string) {
	t.Helper()
	v := wantValidationError(t, err, path, "TS 29.002 §17.1.1: "+rule)
	wantRule := ber.RuleRef{Document: "3GPP TS 29.002", Edition: "V19.1.0", Clause: "17.1.1"}
	if v.Kind != ber.ProtocolRule || v.Rule != wantRule {
		t.Errorf("kind/rule = %v/%+v, want %v/%+v", v.Kind, v.Rule, ber.ProtocolRule, wantRule)
	}
	if v.ObservedValue != "received BER form" {
		t.Errorf("observed value = %q, want received BER form", v.ObservedValue)
	}
	var ee *ber.ElementError
	if !errors.As(err, &ee) {
		t.Errorf("err = %v, want a *ber.ElementError", err)
	}
}

// Parse enforces 3GPP TS 29.002 V19.1.0 §17.1.1: primitive strings and
// short-form lengths below 128 octets. An indefinite argument is still allowed
// (ITU-T X.690 (02/2021) §8.1.3.6); Marshal sends definite lengths.
// AnyTimeInterrogationArg is defined in §17.7.1.
func TestAnyTimeInterrogationDecodeProfile(t *testing.T) {
	const primitive = "3014a00880060001012143f5a1028000830491132654"
	for _, tc := range []struct {
		name, wire, path, rule string
	}{
		{"primitive", primitive, "", ""},
		{"constructed IMSI", "3016a00aa00804060001012143f5a1028000830491132654", "subscriberIdentity.imsi", "primitive-octet-string"},
		{"segmented IMSI", "3018a00ca00a040200010404012143f5a1028000830491132654", "subscriberIdentity.imsi", "primitive-octet-string"},
		{"constructed gsmSCF-Address", "3018a00880060001012143f5a1028000a3080402911304022654", "gsmSCF-Address", "primitive-octet-string"},
		{"long-form IMSI length", "3015a0098081060001012143f5a1028000830491132654", "subscriberIdentity.imsi", "short-length"},
		{"indefinite constructed IMSI", "3018a00ca08004060001012143f50000a1028000830491132654", "subscriberIdentity.imsi", "primitive-octet-string"},
		{"indefinite argument", "3080a00880060001012143f5a10280008304911326540000", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := hex.DecodeString(tc.wire)
			if err != nil {
				t.Fatal(err)
			}
			ati, err := ParseAnyTimeInterrogation(data)
			if tc.rule != "" {
				if ati != nil {
					t.Errorf("Parse = %+v, want nil", ati)
				}
				wantMAPProfileError(t, err, tc.path, tc.rule)
				return
			}
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

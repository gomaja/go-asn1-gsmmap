package gsmmap

import (
	"bytes"
	"encoding/hex"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"

	"github.com/gomaja/go-asn1-gsmmap/address"
)

// mustHex decodes a hex test vector.
func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// PDP-ContextInfo (3GPP TS 29.002 V19.1.0 §17.7.1) is opaque in
// PsSubscriberState. Parse returns it in DER and Marshal sends it in DER,
// whatever BER it arrived in, so the result meets §17.1.1.
func TestOpaquePDPContextInfoInDER(t *testing.T) {
	const (
		pdp = "3007800101820200f1" // pdp-ContextIdentifier 1, pdp-Type 00f1
		ati = "300f300da40ba509" + pdp
	)
	for _, tc := range []struct{ name, pdp, ati string }{
		{"definite short form", pdp, ati},
		{"long-form length", "300880810101820200f1", "3010300ea40ca50a300880810101820200f1"},
		{"indefinite length", "3080800101820200f10000", "3011300fa40da50b3080800101820200f10000"},
		{"constructed pdp-Type", "3009800101a204040200f1", "3011300fa40da50b3009800101a204040200f1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseAnyTimeInterrogationRes(mustHex(t, tc.ati))
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			wantEqual(t, "PDP-ContextInfo", []HexBytes{mustHex(t, pdp)}, got.SubscriberInfo.PsSubscriberState.PsPDPActiveReachableForPaging)
			out, err := got.Marshal()
			if err != nil {
				t.Fatalf("Marshal of the parsed value: %v", err)
			}
			wantEqual(t, "Marshal of the parsed value", mustHex(t, ati), out)

			in := &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{PsSubscriberState: &PsSubscriberState{
				PsPDPActiveReachableForPaging: []HexBytes{mustHex(t, tc.pdp)},
			}}}
			out, err = in.Marshal()
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			wantEqual(t, "Marshal", mustHex(t, ati), out)
		})
	}
	// An unknown extension addition, [30] here, is kept.
	t.Run("extension addition", func(t *testing.T) {
		const ext = "300a800101820200f19e01ff"
		in := &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{PsSubscriberState: &PsSubscriberState{
			PsPDPActiveReachableForPaging: []HexBytes{mustHex(t, ext)},
		}}}
		out, err := in.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		wantEqual(t, "Marshal", mustHex(t, "30123010a40ea50c"+ext), out)
	})
}

// PagingArea (3GPP TS 29.002 V19.1.0 §17.7.1) entries are opaque
// LocationArea values. Marshal sends them in DER, whatever BER the caller
// gives, so the result meets §17.1.1.
func TestOpaquePagingAreaInDER(t *testing.T) {
	const lac = "81021234" // lac [1] 1234
	for _, tc := range []struct{ name, entry string }{
		{"definite short form", lac},
		{"long-form length", "8181021234"},
		{"constructed lac", "a10404021234"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := &UpdateLocation{
				IMSI:      "001010123456789",
				MscNumber: "31612345678", MscNumberNature: address.NatureInternational, MscNumberPlan: address.PlanISDN,
				VlrNumber: "31612345679", VlrNumberNature: address.NatureInternational, VlrNumberPlan: address.PlanISDN,
				PagingArea: []HexBytes{mustHex(t, tc.entry)},
			}
			out, err := in.Marshal()
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			if !bytes.Contains(out, mustHex(t, "ae04"+lac)) {
				t.Errorf("Marshal = %x, want pagingArea ae04%s", out, lac)
			}
			if err := checkDefiniteLengths(out); err != nil {
				t.Errorf("Marshal = %x: %v", out, err)
			}
			got, err := ParseUpdateLocation(out)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			wantEqual(t, "PagingArea", []HexBytes{mustHex(t, lac)}, got.PagingArea)
		})
	}
}

// https://github.com/gomaja/go-asn1/issues/132: MarshalBER sends a decoded
// value that has not changed in the form it was received in, so
// unmarshalAsDER decodes opaque entries again from their DER encoding. This
// pins today's behaviour; when it fails, MarshalBER no longer replays the
// received form and unmarshalAsDER can decode the entry once.
func TestGoASN1Issue132MarshalBERReplaysReceivedForm(t *testing.T) {
	const longForm = "300880810101820200f1"
	var ctx gsm_map.PDPContextInfo
	if err := ctx.UnmarshalBER(mustHex(t, longForm)); err != nil {
		t.Fatal(err)
	}
	out, err := ctx.MarshalBER()
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "MarshalBER", mustHex(t, longForm), out)
}

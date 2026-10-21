package gsmmap

import (
	"bytes"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime/ber"

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
// PsSubscriberState. Parse and Marshal reject entries excluded by §17.1.1.
// Indefinite lengths remain allowed (ITU-T X.690 (02/2021) §8.1.3.6).
// Parse returns DER; Marshal freshly encodes accepted entries.
func TestOpaquePDPContextInfoInDER(t *testing.T) {
	const (
		pdp = "30078001018202f121" // pdp-ContextIdentifier 1, pdp-Type f121
		ati = "300f300da40ba509" + pdp
	)
	for _, tc := range []struct{ name, pdp, ati, path, rule string }{
		{"definite short form", pdp, ati, "", ""},
		{"long-form length", "3008808101018202f121", "3010300ea40ca50a3008808101018202f121", "pdp-ContextIdentifier", "short-length"},
		{"indefinite length", "30808001018202f1210000", "3011300fa40da50b30808001018202f1210000", "", ""},
		{"constructed pdp-Type", "3009800101a2040402f121", "3011300fa40da50b3009800101a2040402f121", "pdp-Type", "primitive-octet-string"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("Parse", func(t *testing.T) {
				got, err := ParseAnyTimeInterrogationRes(mustHex(t, tc.ati))
				if tc.rule != "" {
					if got != nil {
						t.Errorf("Parse = %+v, want nil", got)
					}
					wantMAPProfileError(t, err, "subscriberInfo.ps-SubscriberState.ps-PDP-ActiveReachableForPaging.element[0]."+tc.path, tc.rule)
					return
				}
				if err != nil {
					t.Fatalf("Parse: %v", err)
				}
				wantEqual(t, "PDP-ContextInfo", []HexBytes{mustHex(t, pdp)}, got.SubscriberInfo.PsSubscriberState.PsPDPActiveReachableForPaging)
				out, err := got.Marshal()
				if err != nil {
					t.Fatalf("Marshal of the parsed value: %v", err)
				}
				wantEqual(t, "Marshal of the parsed value", mustHex(t, ati), out)
			})

			t.Run("Marshal", func(t *testing.T) {
				in := &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{PsSubscriberState: &PsSubscriberState{
					PsPDPActiveReachableForPaging: []HexBytes{mustHex(t, tc.pdp)},
				}}}
				out, err := in.Marshal()
				if tc.rule != "" {
					if out != nil {
						t.Errorf("Marshal = %x, want nil", out)
					}
					wantMAPProfileError(t, err, tc.path, tc.rule)
					return
				}
				if err != nil {
					t.Fatalf("Marshal: %v", err)
				}
				wantEqual(t, "Marshal", mustHex(t, ati), out)
			})
		})
	}
	// An unknown extension addition, [30] here, is kept.
	t.Run("extension addition", func(t *testing.T) {
		const ext = "300a8001018202f1219e01ff"
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
// LocationArea values. Marshal rejects entries excluded by §17.1.1 and
// freshly encodes accepted entries with primitive strings and definite lengths.
func TestOpaquePagingAreaInDER(t *testing.T) {
	const lac = "81021234" // lac [1] 1234
	for _, tc := range []struct{ name, entry, rule string }{
		{"definite short form", lac, ""},
		{"long-form length", "8181021234", "short-length"},
		{"constructed lac", "a10404021234", "primitive-octet-string"},
		{"indefinite constructed lac", "a180040212340000", "primitive-octet-string"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := &UpdateLocation{
				IMSI:      "001010123456789",
				MscNumber: "31612345678", MscNumberNature: address.NatureInternational, MscNumberPlan: address.PlanISDN,
				VlrNumber: "31612345679", VlrNumberNature: address.NatureInternational, VlrNumberPlan: address.PlanISDN,
				PagingArea: []HexBytes{mustHex(t, tc.entry)},
			}
			out, err := in.Marshal()
			if tc.rule != "" {
				if out != nil {
					t.Errorf("Marshal = %x, want nil", out)
				}
				wantMAPProfileError(t, err, "lac", tc.rule)
				return
			}
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

// 3GPP TS 29.002 V19.1.0 §17.1.1 applies to opaque entries too.
func TestOpaquePDPContextUnknownExtension(t *testing.T) {
	in := &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{PsSubscriberState: &PsSubscriberState{
		PsPDPActiveReachableForPaging: []HexBytes{mustHex(t, "30808001018202f1219e01ff0000")},
	}}}
	out, err := in.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	// The entry contains 3 + 4 + 3 octets; enclosing lengths are 12, 14, 16, 18.
	wantEqual(t, "Marshal", mustHex(t, "30123010a40ea50c300a8001018202f1219e01ff"), out)
	if err := checkDefiniteLengths(out); err != nil {
		t.Fatal(err)
	}
}

func TestOpaquePDPContextNonMinimalExtensionLength(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Marshal panicked: %v", r)
		}
	}()
	in := &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{PsSubscriberState: &PsSubscriberState{
		PsPDPActiveReachableForPaging: []HexBytes{mustHex(t, "30808001018202f1219e8101ff0000")},
	}}}
	out, err := in.Marshal()
	if !errors.Is(err, ber.ErrInvalidValue) {
		t.Fatalf("Marshal = %x, %v; want ber.ErrInvalidValue", out, err)
	}
	if out != nil {
		t.Fatalf("Marshal returned %x with an error", out)
	}
}

// ITU-T X.690 (02/2021) §10.3 orders SET components by tag and §11.6 orders
// SET OF components by encoding. A universal SET in an opaque extension that
// follows neither order cannot be repaired without its type, so Marshal and
// Parse reject it. The same SET in SET OF order is kept.
func TestOpaquePDPContextUnorderedSETExtension(t *testing.T) {
	const (
		unordered = "30138001018202f121be0a31080201020201010500" // [30] SET { 2, 1, NULL }
		ordered   = "30138001018202f121be0a31080201010201020500" // [30] SET { 1, 2, NULL }
		envelope  = "301b3019a417a515"                           // the entry has 21 octets
	)
	marshal := func(pdp string) ([]byte, error) {
		in := &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{PsSubscriberState: &PsSubscriberState{
			PsPDPActiveReachableForPaging: []HexBytes{mustHex(t, pdp)},
		}}}
		return in.Marshal()
	}
	out, err := marshal(ordered)
	if err != nil {
		t.Fatalf("Marshal in SET OF order: %v", err)
	}
	wantEqual(t, "Marshal in SET OF order", mustHex(t, envelope+ordered), out)
	got, err := ParseAnyTimeInterrogationRes(out)
	if err != nil {
		t.Fatalf("Parse in SET OF order: %v", err)
	}
	wantEqual(t, "PDP-ContextInfo", []HexBytes{mustHex(t, ordered)}, got.SubscriberInfo.PsSubscriberState.PsPDPActiveReachableForPaging)

	out, err = marshal(unordered)
	if !errors.Is(err, ber.ErrInvalidValue) {
		t.Fatalf("Marshal = %x, %v; want ber.ErrInvalidValue", out, err)
	}
	if out != nil {
		t.Fatalf("Marshal returned %x with an error", out)
	}
	if _, err := ParseAnyTimeInterrogationRes(mustHex(t, envelope+unordered)); !errors.Is(err, ber.ErrInvalidValue) {
		t.Fatalf("Parse: %v, want ber.ErrInvalidValue", err)
	}
}

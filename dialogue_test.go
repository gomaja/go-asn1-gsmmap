package gsmmap

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math/bits"
	"reflect"
	"strings"
	"testing"

	"github.com/gomaja/go-asn1-gsmmap/address"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime/ber"
)

func dialoguePtr[T any](v T) *T { return &v }

func dialogueHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// TS 29.002 V19.1.0 §17.4: IMPLICIT tags replace each info's SEQUENCE tag.
// Lengths below are the sums of the enclosed TLVs, counted in octets.
func TestDialoguePDUWires(t *testing.T) {
	dest := DialogueOpen{DestinationReference: "001010123456789", DestinationReferenceNature: address.NatureInternational, DestinationReferencePlan: address.PlanLandMobile}
	orig := DialogueOpen{OriginationReference: "31612345678", OriginationReferenceNature: address.NatureInternational, OriginationReferencePlan: address.PlanISDN}
	both := dest
	both.OriginationReference, both.OriginationReferenceNature, both.OriginationReferencePlan = orig.OriginationReference, orig.OriginationReferenceNature, orig.OriginationReferencePlan
	for _, tc := range []struct {
		name, wire string
		want       DialoguePDU
	}{
		{"open empty", "a000", DialoguePDU{Open: &DialogueOpen{}}},                                                                                                                                               // Empty MAP-OpenInfo.
		{"open destination", "a00b80099600010121436587f9", DialoguePDU{Open: &dest}},                                                                                                                             // [0], 9 octets: international land-mobile address.
		{"open origination", "a0098107911316325476f8", DialoguePDU{Open: &orig}},                                                                                                                                 // [1], 7 octets: international ISDN address.
		{"open both", "a01480099600010121436587f98107911316325476f8", DialoguePDU{Open: &both}},                                                                                                                  // 11 + 9 octets.
		{"accept", "a100", DialoguePDU{Accept: &DialogueAccept{}}},                                                                                                                                               // Empty MAP-AcceptInfo.
		{"close", "a200", DialoguePDU{Close: &DialogueClose{}}},                                                                                                                                                  // Empty MAP-CloseInfo.
		{"refuse no reason", "a3030a0100", DialoguePDU{Refuse: &DialogueRefuse{Reason: RefuseReasonNoReasonGiven}}},                                                                                              // ENUMERATED 0.
		{"refuse destination", "a3030a0101", DialoguePDU{Refuse: &DialogueRefuse{Reason: RefuseReasonInvalidDestinationReference}}},                                                                              // Issue #138: ENUMERATED 1.
		{"refuse originating", "a3030a0102", DialoguePDU{Refuse: &DialogueRefuse{Reason: RefuseReasonInvalidOriginatingReference}}},                                                                              // ENUMERATED 2.
		{"refuse no reason with AC", "a30c0a0100060704000001001302", DialoguePDU{Refuse: &DialogueRefuse{Reason: RefuseReasonNoReasonGiven, AlternativeApplicationContext: "0.4.0.0.1.0.19.2"}}},                 // 3-octet reason + 9-octet OID.
		{"refuse destination with AC", "a30c0a0101060704000001001302", DialoguePDU{Refuse: &DialogueRefuse{Reason: RefuseReasonInvalidDestinationReference, AlternativeApplicationContext: "0.4.0.0.1.0.19.2"}}}, // Reason 1 + OID.
		{"refuse originating with AC", "a30c0a0102060704000001001302", DialoguePDU{Refuse: &DialogueRefuse{Reason: RefuseReasonInvalidOriginatingReference, AlternativeApplicationContext: "0.4.0.0.1.0.19.2"}}}, // Reason 2 + OID.
		{"user specific", "a4028000", DialoguePDU{UserAbort: &DialogueUserAbort{UserSpecificReason: true}}},                                                                                                      // Issue #138: [0] NULL.
		{"user resource limitation", "a4028100", DialoguePDU{UserAbort: &DialogueUserAbort{UserResourceLimitation: true}}},                                                                                       // [1] NULL.
		{"resource short term", "a403820100", DialoguePDU{UserAbort: &DialogueUserAbort{ResourceUnavailable: dialoguePtr(ResourceUnavailableShortTermResourceLimitation)}}},                                      // [2] ENUMERATED 0.
		{"resource long term", "a403820101", DialoguePDU{UserAbort: &DialogueUserAbort{ResourceUnavailable: dialoguePtr(ResourceUnavailableLongTermResourceLimitation)}}},                                        // [2] ENUMERATED 1.
		{"cancel handover", "a403830100", DialoguePDU{UserAbort: &DialogueUserAbort{ApplicationProcedureCancellation: dialoguePtr(ProcedureCancellationHandoverCancellation)}}},                                  // [3] ENUMERATED 0.
		{"cancel radio", "a403830101", DialoguePDU{UserAbort: &DialogueUserAbort{ApplicationProcedureCancellation: dialoguePtr(ProcedureCancellationRadioChannelRelease)}}},                                      // [3] ENUMERATED 1.
		{"cancel network", "a403830102", DialoguePDU{UserAbort: &DialogueUserAbort{ApplicationProcedureCancellation: dialoguePtr(ProcedureCancellationNetworkPathRelease)}}},                                     // [3] ENUMERATED 2.
		{"cancel call", "a403830103", DialoguePDU{UserAbort: &DialogueUserAbort{ApplicationProcedureCancellation: dialoguePtr(ProcedureCancellationCallRelease)}}},                                               // [3] ENUMERATED 3.
		{"cancel associated", "a403830104", DialoguePDU{UserAbort: &DialogueUserAbort{ApplicationProcedureCancellation: dialoguePtr(ProcedureCancellationAssociatedProcedureFailure)}}},                          // [3] ENUMERATED 4.
		{"cancel tandem", "a403830105", DialoguePDU{UserAbort: &DialogueUserAbort{ApplicationProcedureCancellation: dialoguePtr(ProcedureCancellationTandemDialogueRelease)}}},                                   // [3] ENUMERATED 5.
		{"cancel remote", "a403830106", DialoguePDU{UserAbort: &DialogueUserAbort{ApplicationProcedureCancellation: dialoguePtr(ProcedureCancellationRemoteOperationsFailure)}}},                                 // [3] ENUMERATED 6.
		{"provider abnormal", "a5030a0100", DialoguePDU{ProviderAbort: &DialogueProviderAbort{Reason: ProviderAbortAbnormalDialogue}}},                                                                           // ENUMERATED 0.
		{"provider invalid", "a5030a0101", DialoguePDU{ProviderAbort: &DialogueProviderAbort{Reason: ProviderAbortInvalidPDU}}},                                                                                  // ENUMERATED 1.
		{"wide OID arc", "a3100a0100060b6982808080808080808000", DialoguePDU{Refuse: &DialogueRefuse{AlternativeApplicationContext: "2.25.18446744073709551616"}}},                                               // OID arc 2^64: no uint64 narrowing.
	} {
		t.Run(tc.name, func(t *testing.T) {
			wire := dialogueHex(t, tc.wire)
			got, err := ParseDialoguePDU(wire)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, &tc.want) {
				t.Fatalf("Parse = %+v, want %+v", got, tc.want)
			}
			out, err := got.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(out, wire) {
				t.Errorf("Marshal parsed = %x, want %x", out, wire)
			}
			out, err = tc.want.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(out, wire) {
				t.Errorf("Marshal built = %x, want %x", out, wire)
			}
		})
	}
	if got := MapDialogueAS(); got != "0.4.0.0.1.1.1.1" {
		t.Errorf("MapDialogueAS = %q", got)
	}
}

// Every info is extensible (§17.4); unsupported additions are accepted
// (§17.1.4). The public API drops containers and unknown additions.
func TestDialoguePDUDropsExtensions(t *testing.T) {
	for _, tc := range []struct{ wire, clean string }{
		{"a0023000", "a000"},             // Open: empty ExtensionContainer SEQUENCE.
		{"a1023000", "a100"},             // Accept: container.
		{"a2023000", "a200"},             // Close: container.
		{"a3050a01013000", "a3030a0101"}, // Refuse: reason + container.
		{"a40480003000", "a4028000"},     // UserAbort: choice + container.
		{"a5050a01013000", "a5030a0101"}, // ProviderAbort: reason + container.
		{"a0028700", "a000"},             // Open: unknown extension [7].
		{"a1028700", "a100"},             // Accept: unknown extension.
		{"a2028700", "a200"},             // Close: unknown extension.
		{"a3050a01018700", "a3030a0101"}, // Refuse: unknown extension.
		{"a40480008700", "a4028000"},     // UserAbort: unknown extension.
		{"a5050a01018700", "a5030a0101"}, // ProviderAbort: unknown extension.
		{"a30e0a01013000060704000001001302", "a30c0a0101060704000001001302"}, // Container before alternative AC.
		{"a080800296f10000", "a004800296f1"},                                 // Indefinite Open becomes definite, one destination digit.
	} {
		t.Run(tc.wire, func(t *testing.T) {
			v, err := ParseDialoguePDU(dialogueHex(t, tc.wire))
			if err != nil {
				t.Fatal(err)
			}
			out, err := v.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			if hex.EncodeToString(out) != tc.clean {
				t.Errorf("Marshal = %x, want %s", out, tc.clean)
			}
		})
	}
}

func TestDialoguePDUChoices(t *testing.T) {
	if b, err := (*DialoguePDU)(nil).Marshal(); b != nil || !errors.Is(err, ErrDialoguePDUNoAlternative) {
		t.Errorf("nil Marshal = %x, %v; want %v", b, err, ErrDialoguePDUNoAlternative)
	}
	// TS 29.002 V19.1.0 §17.4: each CHOICE takes exactly one alternative, so
	// every other combination of set fields is refused.
	choiceError := func(set int, none, many error) error {
		switch set {
		case 0:
			return none
		case 1:
			return nil
		default:
			return many
		}
	}
	check := func(p *DialoguePDU, mask int, want error) {
		t.Helper()
		b, err := p.Marshal()
		if want == nil {
			if err != nil {
				t.Errorf("mask %06b: Marshal = %v, want success", mask, err)
			}
			return
		}
		if b != nil || !errors.Is(err, want) || !strings.HasPrefix(err.Error(), "converting DialoguePDU: ") {
			t.Errorf("mask %06b: Marshal = %x, %v; want %v", mask, b, err, want)
		}
	}
	for mask := 0; mask < 1<<6; mask++ {
		p := &DialoguePDU{}
		if mask&1 != 0 {
			p.Open = &DialogueOpen{}
		}
		if mask&2 != 0 {
			p.Accept = &DialogueAccept{}
		}
		if mask&4 != 0 {
			p.Close = &DialogueClose{}
		}
		if mask&8 != 0 {
			p.Refuse = &DialogueRefuse{}
		}
		if mask&16 != 0 {
			p.UserAbort = &DialogueUserAbort{UserSpecificReason: true}
		}
		if mask&32 != 0 {
			p.ProviderAbort = &DialogueProviderAbort{}
		}
		check(p, mask, choiceError(bits.OnesCount(uint(mask)), ErrDialoguePDUNoAlternative, ErrDialoguePDUMultipleAlternatives))
	}
	for mask := 0; mask < 1<<4; mask++ {
		u := &DialogueUserAbort{UserSpecificReason: mask&1 != 0, UserResourceLimitation: mask&2 != 0}
		if mask&4 != 0 {
			u.ResourceUnavailable = dialoguePtr(ResourceUnavailableShortTermResourceLimitation)
		}
		if mask&8 != 0 {
			u.ApplicationProcedureCancellation = dialoguePtr(ProcedureCancellationHandoverCancellation)
		}
		check(&DialoguePDU{UserAbort: u}, mask, choiceError(bits.OnesCount(uint(mask)), ErrDialogueUserAbortNoAlternative, ErrDialogueUserAbortMultipleAlternatives))
	}
	// Unknown wire tags are refused by the codec first. Converter defaults
	// remain defensive against unsupported generated choices (§17.4).
	for _, choice := range []int{0, 99} {
		if p, err := convertWireToDialoguePDU(&gsm_map.MAPDialoguePDU{Choice: choice}); p != nil || !errors.Is(err, ErrDialoguePDUUnknownAlternative) {
			t.Errorf("PDU default = %v, %v", p, err)
		}
		w := gsm_map.NewMAPDialoguePDUMapUserAbort(gsm_map.MAPUserAbortInfo{MapUserAbortChoice: gsm_map.MAPUserAbortChoice{Choice: choice}})
		if p, err := convertWireToDialoguePDU(&w); p != nil || !errors.Is(err, ErrDialogueUserAbortUnknownAlternative) {
			t.Errorf("UserAbort default = %v, %v", p, err)
		}
	}
}

func TestDialogueOpenReferences(t *testing.T) {
	for _, tc := range []struct {
		wire string
		want error
	}{
		{"a003800196", ErrDialogueOpenDestinationReferenceDecodedEmpty},   // Destination: only nature/plan.
		{"a004800296ff", ErrDialogueOpenDestinationReferenceDecodedEmpty}, // Destination: all filler.
		{"a003810191", ErrDialogueOpenOriginationReferenceDecodedEmpty},   // Origination: only nature/plan.
		{"a004810291ff", ErrDialogueOpenOriginationReferenceDecodedEmpty}, // Origination: all filler.
	} {
		if p, err := ParseDialoguePDU(dialogueHex(t, tc.wire)); p != nil || !errors.Is(err, tc.want) || strings.HasPrefix(err.Error(), "decoding MAPDialoguePDU:") {
			t.Errorf("Parse = %v, %v; want unwrapped %v", p, err, tc.want)
		}
	}
	// As with every optional address, empty digits omit the field, ignoring
	// nature/plan, even if those values would be invalid for a present field.
	p := &DialoguePDU{Open: &DialogueOpen{DestinationReferenceNature: 0xff, DestinationReferencePlan: 0xff, OriginationReferenceNature: 0xff, OriginationReferencePlan: 0xff}}
	if b, err := p.Marshal(); err != nil || !bytes.Equal(b, dialogueHex(t, "a000")) {
		t.Fatalf("empty references: %x, %v", b, err)
	}
	for _, destination := range []bool{true, false} {
		for _, tc := range []struct {
			digits       string
			nature, plan uint8
			want         error
		}{
			{"1", address.NatureReserved, 1, ErrAddressNatureInvalid},
			{"1", address.NatureInternational, 0xff, ErrAddressPlanInvalid},
			{"z", address.NatureInternational, 1, nil},
		} {
			o := &DialogueOpen{}
			if destination {
				o.DestinationReference, o.DestinationReferenceNature, o.DestinationReferencePlan = tc.digits, tc.nature, tc.plan
			} else {
				o.OriginationReference, o.OriginationReferenceNature, o.OriginationReferencePlan = tc.digits, tc.nature, tc.plan
			}
			if _, err := (&DialoguePDU{Open: o}).Marshal(); err == nil || (tc.want != nil && !errors.Is(err, tc.want)) {
				t.Errorf("bad reference: %v; want %v", err, tc.want)
			}
		}
	}
	for _, wire := range []string{"a004800296f1", "a004810291f1"} { // One digit in each reference.
		b := dialogueHex(t, wire)
		p, err := ParseDialoguePDU(b)
		if err != nil {
			t.Fatal(err)
		}
		out, err := p.Marshal()
		if err != nil || !bytes.Equal(out, b) {
			t.Errorf("one digit = %x, %v", out, err)
		}
	}
	// Invalid TBCD filler position, independently in each reference.
	for _, wire := range []string{"a0048002961f", "a0048102911f"} {
		if p, err := ParseDialoguePDU(dialogueHex(t, wire)); err == nil || p != nil {
			t.Errorf("invalid TBCD = %v, %v", p, err)
		}
	}
}

func TestDialoguePDUOID(t *testing.T) {
	for _, text := range []OID{"garbage", "0..4", "3.0", "0.40", "0", "00.4", "0.04", "+0.4", "0.4 ", " 0.4", "0.4."} {
		p := &DialoguePDU{Refuse: &DialogueRefuse{AlternativeApplicationContext: text}}
		if b, err := p.Marshal(); b != nil || !errors.Is(err, ErrInvalidOID) || !errors.Is(err, runtime.ErrInvalidOID) {
			t.Errorf("OID %q = %x, %v; want both OID sentinels", text, b, err)
		}
	}
	p := &DialoguePDU{Refuse: &DialogueRefuse{AlternativeApplicationContext: OID("2.25." + strings.Repeat("1", runtime.MaxOIDArcDigits+1))}}
	if _, err := p.Marshal(); !errors.Is(err, runtime.ErrOIDTextLimit) {
		t.Errorf("OID text limit: %v", err)
	}
	// Valid wire OIDs have no arc-width limit. Decimal conversion does:
	// this 3501-bit arc needs more than MaxOIDArcDigits decimal digits.
	oid := runtime.MustObjectIdentifierFromEncoded("\x69" + strings.Repeat("\x81", 500) + "\x00")
	w := gsm_map.NewMAPDialoguePDUMapRefuse(gsm_map.MAPRefuseInfo{AlternativeApplicationContext: &oid})
	b, err := w.MarshalBER()
	if err != nil {
		t.Fatal(err)
	}
	if p, err := ParseDialoguePDU(b); p != nil || !errors.Is(err, runtime.ErrOIDTextLimit) {
		t.Errorf("OID formatting limit: %v, %v", p, err)
	}
}

func TestDialoguePDUEnumConstraints(t *testing.T) {
	for _, tc := range []struct {
		name, high, negative, path, constraint string
		build                                  func(int64) *DialoguePDU
		max                                    int64
	}{
		{"refuse", "a3030a0103", "a3030a01ff", "map-refuse.reason", "ENUMERATED {0, 1, 2}", func(v int64) *DialoguePDU {
			return &DialoguePDU{Refuse: &DialogueRefuse{Reason: DialogueRefuseReason(v)}}
		}, 2},
		{"resource", "a403820102", "a4038201ff", "map-userAbort.map-UserAbortChoice.resourceUnavailable", "ENUMERATED {0, 1}", func(v int64) *DialoguePDU {
			return &DialoguePDU{UserAbort: &DialogueUserAbort{ResourceUnavailable: dialoguePtr(ResourceUnavailableReason(v))}}
		}, 1},
		{"cancellation", "a403830107", "a4038301ff", "map-userAbort.map-UserAbortChoice.applicationProcedureCancellation", "ENUMERATED {0, 1, 2, 3, 4, 5, 6}", func(v int64) *DialoguePDU {
			return &DialoguePDU{UserAbort: &DialogueUserAbort{ApplicationProcedureCancellation: dialoguePtr(ProcedureCancellationReason(v))}}
		}, 6},
		{"provider", "a5030a0102", "a5030a01ff", "map-providerAbort.map-ProviderAbortReason", "ENUMERATED {0, 1}", func(v int64) *DialoguePDU {
			return &DialoguePDU{ProviderAbort: &DialogueProviderAbort{Reason: ProviderAbortReason(v)}}
		}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, wire := range []string{tc.high, tc.negative} { // First unlisted and negative ENUMERATED values.
				p, err := ParseDialoguePDU(dialogueHex(t, wire))
				if p != nil {
					t.Error("non-nil failed Parse")
				}
				wantValidationError(t, err, tc.path, tc.constraint)
			}
			for _, v := range []int64{-1, tc.max + 1} {
				b, err := tc.build(v).Marshal()
				if b != nil {
					t.Error("non-nil failed Marshal")
				}
				wantValidationError(t, err, tc.path, tc.constraint)
				if !strings.HasPrefix(err.Error(), "encoding MAPDialoguePDU: ") {
					t.Errorf("encode wrapper = %v", err)
				}
			}
		})
	}
}

func TestDialoguePDUCodecRules(t *testing.T) {
	for _, tc := range []struct {
		wire string
		want error
	}{
		{"a00000", ber.ErrExtraData},           // Empty Open followed by a trailing octet.
		{"a600", ber.ErrUnexpectedElement},     // Unknown PDU alternative [6].
		{"8000", ber.ErrInvalidTag},            // Open in primitive form.
		{"a4028400", ber.ErrUnexpectedElement}, // Unknown UserAbort alternative [4].
		// §17.4: reason, map-UserAbortChoice and map-ProviderAbortReason are
		// mandatory, so no decoder may default them to zero.
		{"a300", ber.ErrUnexpectedElement},           // Refuse without a reason.
		{"a3023000", ber.ErrUnexpectedElement},       // Refuse with only an extension container.
		{"a30506032a0304", ber.ErrUnexpectedElement}, // Refuse with only an alternativeApplicationContext, 1.2.3.4.
		{"a400", ber.ErrUnexpectedElement},           // UserAbort without a choice.
		{"a4023000", ber.ErrUnexpectedElement},       // UserAbort with only an extension container.
		{"a500", ber.ErrUnexpectedElement},           // ProviderAbort without a reason.
		{"a5023000", ber.ErrUnexpectedElement},       // ProviderAbort with only an extension container.
	} {
		p, err := ParseDialoguePDU(dialogueHex(t, tc.wire))
		if p != nil || !errors.Is(err, tc.want) || !strings.HasPrefix(err.Error(), "decoding MAPDialoguePDU: ") {
			t.Errorf("Parse %s = %v, %v; want %v", tc.wire, p, err, tc.want)
		}
	}
	for _, tc := range []struct{ wire, rule string }{
		{"a006a004040296f1", "primitive-octet-string"}, // Constructed destination containing an OCTET STRING.
		{"a00580810296f1", "short-length"},             // Destination length 2 in long form.
	} {
		_, err := ParseDialoguePDU(dialogueHex(t, tc.wire))
		wantMAPProfileError(t, err, "map-open.destinationReference", tc.rule)
	}
	// AddressString SIZE (1..20), §17.7.8: 1 nature/plan + 20 TBCD octets.
	oversized := bytes.Repeat([]byte{0x11}, 21)
	oversized[0] = 0x96
	for _, destination := range []bool{true, false} {
		w := gsm_map.MAPOpenInfo{}
		o := &DialogueOpen{}
		path := "map-open.originationReference"
		if destination {
			w.DestinationReference = &oversized
			o.DestinationReference = strings.Repeat("1", 40)
			path = "map-open.destinationReference"
		} else {
			w.OriginationReference = &oversized
			o.OriginationReference = strings.Repeat("1", 40)
		}
		wp := gsm_map.NewMAPDialoguePDUMapOpen(w)
		_, err := ParseDialoguePDU(tolerantBER(t, &wp))
		wantValidationError(t, err, path, "SIZE (1..20)", 21)
		_, err = (&DialoguePDU{Open: o}).Marshal()
		wantValidationError(t, err, path, "SIZE (1..20)", 21)
	}
}

// Keep the dialogue in the seed-based minimal-length gate, including its
// shortest PDUs, which have only two octets.
func TestDialoguePDUSeeds(t *testing.T) {
	seen := map[string]bool{}
	checked := 0
	for _, b := range berSeeds(t) {
		seen[hex.EncodeToString(b)] = true
		p, err := ParseDialoguePDU(b)
		if err != nil {
			continue
		}
		out, err := p.Marshal()
		if err != nil {
			continue
		}
		checked++
		if err := checkDefiniteLengths(out); err != nil {
			t.Errorf("%x: %v", out, err)
		}
	}
	for _, s := range []string{"a000", "a100", "a200", "a3030a0101", "a4028000"} {
		if !seen[s] {
			t.Errorf("missing seed %s", s)
		}
	}
	if checked == 0 {
		t.Fatal("no DialoguePDU reached the minimal-length check")
	}
	t.Logf("%d DialoguePDU seeds checked", checked)
}

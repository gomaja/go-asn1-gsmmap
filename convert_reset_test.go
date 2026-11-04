package gsmmap

import (
	"errors"
	identity "github.com/gomaja/go-3gpp-identity"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"testing"
)

func TestResetMinimal(t *testing.T) {
	// SEQUENCE: SendingNode-Number hlr-Number, nature/plan 0x91 and digits 12.
	hlrRoundTrip(t, &ResetArg{SendingNodeNumber: SendingNodeNumber{HLRNumber: "12", HLRNumberNature: 0x10, HLRNumberPlan: 1}}, ParseResetArg, "300404029121")
	// SendingNode-Number css-Number uses context tag [1].
	hlrRoundTrip(t, &ResetArg{SendingNodeNumber: SendingNodeNumber{CSSNumber: "12", CSSNumberNature: 0x10, CSSNumberPlan: 1}}, ParseResetArg, "300481029121")
}

func TestResetFull(t *testing.T) {
	// HLR-List contains two HLR-Id IMSIs, three octets each.
	hlrRoundTrip(t, &ResetArg{SendingNodeNumber: SendingNodeNumber{HLRNumber: "12"}, HLRList: []string{"001010", "123456"}}, ParseResetArg, "301004028021300a04030001010403214365")
	// [1] Reset-ID list and [2] subscription data with optional IMSI.
	hlrRoundTrip(t, &ResetArg{SendingNodeNumber: SendingNodeNumber{CSSNumber: "12"}, ResetIdList: ResetIdList{{1}, {2, 3, 4, 5}}, SubscriptionData: &InsertSubscriberDataArg{IMSI: "001010123456789"}}, ParseResetArg, "301b81028021a109040101040402030405a20a800800010121436587f9")
	// [3] subscription data deletion reuses DeleteSubscriberDataArg.
	hlrRoundTrip(t, &ResetArg{SendingNodeNumber: SendingNodeNumber{HLRNumber: "12"}, ResetIdList: ResetIdList{{1}}, SubscriptionDataDeletion: &DeleteSubscriberDataArg{IMSI: "001010123456789"}}, ParseResetArg, "301504028021a103040101a30a800800010121436587f9")
}

func TestResetConstraints(t *testing.T) {
	for _, tc := range []struct {
		v                      *ResetArg
		wire, path, constraint string
	}{
		{&ResetArg{SendingNodeNumber: SendingNodeNumber{HLRNumber: "12"}, HLRList: []string{}}, "3006040280213000", "hlr-List", "SIZE (1..50)"},
		{&ResetArg{SendingNodeNumber: SendingNodeNumber{HLRNumber: "12"}, ResetIdList: ResetIdList{}}, "300604028021a100", "reset-Id-List", "SIZE (1..50)"},
		{&ResetArg{SendingNodeNumber: SendingNodeNumber{HLRNumber: "12"}, ResetIdList: ResetIdList{{}}}, "300804028021a1020400", "reset-Id-List.element[0]", "SIZE (1..4)"},
	} {
		_, err := tc.v.Marshal()
		wantValidationError(t, err, tc.path, tc.constraint)
		_, err = ParseResetArg(mustHex(t, tc.wire))
		wantValidationError(t, err, tc.path, tc.constraint)
	}
}

func TestResetSenderRulesFromWire(t *testing.T) {
	for _, tc := range []struct {
		wire string
		want error
	}{
		// CSS sending number plus HLR-List violates the HSS/HLR restart condition.
		{"300b8102802130050403000101", ErrResetArgHLRListWithCSSNumber},
		// HLR-List and Reset-ID LIST may not coexist.
		{"30100402802130050403000101a103040101", ErrResetArgHLRListWithResetIdList},
		// Shared insertion and deletion are mutually exclusive.
		{"301704028021a103040101a200a30a800800010121436587f9", ErrResetArgSubscriptionDataConflict},
		// Each shared update requires Reset-ID LIST.
		{"300604028021a200", ErrResetArgSubscriptionDataWithoutResetIdList},
		{"301004028021a30a800800010121436587f9", ErrResetArgSubscriptionDataDeletionWithoutResetIdList},
	} {
		v, err := ParseResetArg(mustHex(t, tc.wire))
		if err != nil {
			t.Fatal(err)
		}
		if _, err = v.Marshal(); !errors.Is(err, tc.want) {
			t.Errorf("got %v, want %v", err, tc.want)
		}
	}
}

func TestResetNestedErrors(t *testing.T) {
	for _, tc := range []struct {
		v    *ResetArg
		want error
	}{
		{&ResetArg{SendingNodeNumber: SendingNodeNumber{HLRNumber: "12"}, ResetIdList: ResetIdList{{1}}, SubscriptionData: &InsertSubscriberDataArg{IMSI: "x"}}, identity.ErrNotDigits},
		{&ResetArg{SendingNodeNumber: SendingNodeNumber{HLRNumber: "12"}, ResetIdList: ResetIdList{{1}}, SubscriptionDataDeletion: &DeleteSubscriberDataArg{}}, identity.ErrEmpty},
	} {
		if _, err := tc.v.Marshal(); !errors.Is(err, tc.want) {
			t.Errorf("got %v, want %v", err, tc.want)
		}
	}
	for _, tc := range []struct {
		v    *gsm_map.ResetArg
		want error
	}{
		{&gsm_map.ResetArg{SendingNodenumber: gsm_map.SendingNodeNumber{Choice: 99}}, ErrSendingNodeNumberUnknownAlternative},
		{&gsm_map.ResetArg{SendingNodenumber: gsm_map.NewSendingNodeNumberHlrNumber([]byte{0x80, 0x21}), SubscriptionData: &gsm_map.InsertSubscriberDataArg{Imsi: new([]byte{0xff, 0xff, 0xff})}}, identity.ErrEmpty},
		{&gsm_map.ResetArg{SendingNodenumber: gsm_map.NewSendingNodeNumberHlrNumber([]byte{0x80, 0x21}), SubscriptionDataDeletion: &gsm_map.DeleteSubscriberDataArg{Imsi: []byte{0xff, 0xff, 0xff}}}, identity.ErrEmpty},
	} {
		if _, err := convertWireToResetArg(tc.v); !errors.Is(err, tc.want) {
			t.Errorf("got %v, want %v", err, tc.want)
		}
	}
}

// The nested arguments keep their contents, and their sender rules still apply
// to what Parse returned.
func TestResetNestedContents(t *testing.T) {
	node := SendingNodeNumber{HLRNumber: "12"}
	// HLR number; Reset-ID 01; [2] IMSI and category 0a.
	hlrRoundTrip(t, &ResetArg{SendingNodeNumber: node, ResetIdList: ResetIdList{{1}}, SubscriptionData: &InsertSubscriberDataArg{IMSI: "001010123456789", Category: HexBytes{0x0a}}},
		ParseResetArg, "301804028021a103040101a20d800800010121436587f982010a")
	// HLR number; Reset-ID 01; [3] IMSI, ss-List 0x21 and O-CSI withdrawn.
	hlrRoundTrip(t, &ResetArg{SendingNodeNumber: node, ResetIdList: ResetIdList{{1}}, SubscriptionDataDeletion: &DeleteSubscriberDataArg{IMSI: "001010123456789", SsList: []SsCode{0x21}, SpecificCSIWithdraw: &SpecificCSIWithdraw{OCSI: true}}},
		ParseResetArg, "301e04028021a103040101a313800800010121436587f9a2030401218f020080")
	// The same deletion with no CSI withdrawn parses, and Marshal refuses it.
	got, err := ParseResetArg(mustHex(t, "301904028021a103040101a30e800800010121436587f98f020000"))
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "nested SpecificCSIWithdraw", &SpecificCSIWithdraw{}, got.SubscriptionDataDeletion.SpecificCSIWithdraw)
	if _, err := got.Marshal(); !errors.Is(err, ErrSpecificCSIWithdrawEmpty) {
		t.Errorf("Marshal: err = %v, want %v", err, ErrSpecificCSIWithdrawEmpty)
	}
}

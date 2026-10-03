// extssinfo_test.go
//
// Tests for Ext-SS-Info CHOICE and its 5 alternatives (ISD PR D).
// Round-trip per alternative + boundary/validation per nested type.
package gsmmap

import (
	"errors"
	"fmt"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/google/go-cmp/cmp"
)

// --- SSSubscriptionOption CHOICE ---

func TestSSSubscriptionOptionRoundTrip(t *testing.T) {
	cli := CliRestrictionTemporaryDefaultRestricted
	over := OverrideEnabled
	cases := []struct {
		name string
		in   *SSSubscriptionOption
	}{
		{"cliRestriction", &SSSubscriptionOption{CliRestriction: &cli}},
		{"overrideCategory", &SSSubscriptionOption{Override: &over}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertSSSubscriptionOptionToWire(tc.in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			got, err := convertWireToSSSubscriptionOption(wire)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if diff := cmp.Diff(tc.in, got); diff != "" {
				t.Errorf("round-trip (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSSSubscriptionOptionValidation(t *testing.T) {
	cli := CliRestrictionPermanent
	over := OverrideDisabled
	t.Run("noAlt", func(t *testing.T) {
		_, err := convertSSSubscriptionOptionToWire(&SSSubscriptionOption{})
		if !errors.Is(err, ErrSSSubscriptionOptionChoiceNoAlternative) {
			t.Errorf("want ErrSSSubscriptionOptionChoiceNoAlternative, got %v", err)
		}
	})
	t.Run("multipleAlts", func(t *testing.T) {
		_, err := convertSSSubscriptionOptionToWire(&SSSubscriptionOption{CliRestriction: &cli, Override: &over})
		if !errors.Is(err, ErrSSSubscriptionOptionChoiceMultipleAlternatives) {
			t.Errorf("want ErrSSSubscriptionOptionChoiceMultipleAlternatives, got %v", err)
		}
	})
	t.Run("invalidCliRestriction", func(t *testing.T) {
		bad := CliRestrictionOption(99)
		_, err := convertSSSubscriptionOptionToWire(&SSSubscriptionOption{CliRestriction: &bad})
		if !errors.Is(err, ErrCliRestrictionOptionInvalidValue) {
			t.Errorf("want ErrCliRestrictionOptionInvalidValue, got %v", err)
		}
	})
	t.Run("invalidOverrideCategory", func(t *testing.T) {
		bad := OverrideCategory(99)
		_, err := convertSSSubscriptionOptionToWire(&SSSubscriptionOption{Override: &bad})
		if !errors.Is(err, ErrOverrideCategoryInvalidValue) {
			t.Errorf("want ErrOverrideCategoryInvalidValue, got %v", err)
		}
	})
}

// --- ExtForwInfo / ExtForwFeature ---

func TestExtForwInfoRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   *ExtForwInfo
	}{
		{
			name: "minimal",
			in: &ExtForwInfo{
				SsCode: 0x21, // CFU
				ForwardingFeatureList: []ExtForwFeature{
					{SsStatus: HexBytes{0x05}},
				},
			},
		},
		{
			name: "fullFeature",
			in: &ExtForwInfo{
				SsCode: 0x29, // CFNRy
				ForwardingFeatureList: []ExtForwFeature{
					{
						BasicService: &ExtBasicServiceCode{
							ExtTeleservice: HexBytes{0x11},
						},
						SsStatus:          HexBytes{0x05},
						ForwardedToNumber: "31611111111",
						ForwardedToNature: 16, ForwardedToPlan: 1,
						ForwardedToSubaddress: HexBytes{0xa0, 0x01, 0x02},
						ForwardingOptions:     HexBytes{0xc0},
						NoReplyConditionTime:  intPtrV(20),
					},
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertExtForwInfoToWire(tc.in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			got, err := convertWireToExtForwInfo(wire)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if diff := cmp.Diff(tc.in, got); diff != "" {
				t.Errorf("round-trip (-want +got):\n%s", diff)
			}
		})
	}
}

func TestExtForwInfoValidation(t *testing.T) {
	t.Run("emptyFeatureList", func(t *testing.T) {
		_, err := strictWire(convertExtForwInfoToWire(&ExtForwInfo{SsCode: 0x21}))
		if !matchesConstraint(err, "forwardingFeatureList", "SIZE (1..32)") {
			t.Errorf("want BER constraint error, got %v", err)
		}
	})
	t.Run("featureListTooLong", func(t *testing.T) {
		big := make([]ExtForwFeature, 32+1)
		for i := range big {
			big[i] = ExtForwFeature{SsStatus: HexBytes{0x05}}
		}
		_, err := strictWire(convertExtForwInfoToWire(&ExtForwInfo{SsCode: 0x21, ForwardingFeatureList: big}))
		if !matchesConstraint(err, "forwardingFeatureList", "SIZE (1..32)") {
			t.Errorf("want BER constraint error, got %v", err)
		}
	})
	t.Run("ssStatusEmpty", func(t *testing.T) {
		_, err := strictWire(convertExtForwInfoToWire(&ExtForwInfo{
			SsCode:                0x21,
			ForwardingFeatureList: []ExtForwFeature{{SsStatus: HexBytes{}}},
		}))
		if !matchesConstraint(err, "ss-Status", "SIZE (1..5)") {
			t.Errorf("want BER constraint error, got %v", err)
		}
	})
	t.Run("ssStatusTooLong", func(t *testing.T) {
		_, err := strictWire(convertExtForwInfoToWire(&ExtForwInfo{
			SsCode:                0x21,
			ForwardingFeatureList: []ExtForwFeature{{SsStatus: HexBytes{1, 2, 3, 4, 5, 6}}},
		}))
		if !matchesConstraint(err, "ss-Status", "SIZE (1..5)") {
			t.Errorf("want BER constraint error, got %v", err)
		}
	})
	t.Run("forwardingOptionsTooLong", func(t *testing.T) {
		_, err := strictWire(convertExtForwInfoToWire(&ExtForwInfo{
			SsCode: 0x21,
			ForwardingFeatureList: []ExtForwFeature{{
				SsStatus:          HexBytes{0x05},
				ForwardingOptions: HexBytes{1, 2, 3, 4, 5, 6}, // 6 octets, max is 5
			}},
		}))
		if !matchesConstraint(err, "forwardingOptions", "SIZE (1..5)") {
			t.Errorf("want BER constraint error, got %v", err)
		}
	})
	t.Run("noRepCondTimeOutOfRange", func(t *testing.T) {
		bad := 200
		_, err := strictWire(convertExtForwInfoToWire(&ExtForwInfo{
			SsCode: 0x21,
			ForwardingFeatureList: []ExtForwFeature{{
				SsStatus:             HexBytes{0x05},
				NoReplyConditionTime: &bad,
			}},
		}))
		if !errors.Is(err, ErrNoReplyConditionTimeOutOfRange) {
			t.Errorf("want ErrNoReplyConditionTimeOutOfRange, got %v", err)
		}
	})
}

// LongForwardedToNumber must round-trip its TON/NPI when it's the only
// number present. Encoder uses the shared ForwardedToNature/Plan; the
// decoder must populate them from the long number's wire octets when
// ForwardedToNumber is absent.
func TestExtForwFeatureLongForwardedToNumberRoundTripPreservesNaturePlan(t *testing.T) {
	in := &ExtForwInfo{
		SsCode: 0x29,
		ForwardingFeatureList: []ExtForwFeature{{
			SsStatus:              HexBytes{0x05},
			LongForwardedToNumber: "31611111111",
			ForwardedToNature:     0x20, // National — non-default to detect loss
			ForwardedToPlan:       9,    // Private — non-default to detect loss
		}},
	}
	wire, err := convertExtForwInfoToWire(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, err := convertWireToExtForwInfo(wire)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("non-default TON/NPI on LongForwardedToNumber must survive round-trip (-want +got):\n%s", diff)
	}
}

// SIZE(1..21) + SIZE(1..5) discipline: a non-nil empty slice is rejected
// at both encode and decode (PR #29 pattern).
func TestExtForwFeatureNilVsEmptyDiscipline(t *testing.T) {
	t.Run("encoderRejectsEmptySubaddress", func(t *testing.T) {
		_, err := strictWire(convertExtForwInfoToWire(&ExtForwInfo{
			SsCode: 0x21,
			ForwardingFeatureList: []ExtForwFeature{{
				SsStatus:              HexBytes{0x05},
				ForwardedToSubaddress: HexBytes{}, // non-nil, empty
			}},
		}))
		if !matchesConstraint(err, "forwardedToSubaddress", "SIZE (1..21)") {
			t.Errorf("want BER constraint error, got %v", err)
		}
	})
	t.Run("encoderRejectsOversizeSubaddress", func(t *testing.T) {
		_, err := strictWire(convertExtForwInfoToWire(&ExtForwInfo{
			SsCode: 0x21,
			ForwardingFeatureList: []ExtForwFeature{{
				SsStatus:              HexBytes{0x05},
				ForwardedToSubaddress: make(HexBytes, 22), // 22 octets, max is 21
			}},
		}))
		if !matchesConstraint(err, "forwardedToSubaddress", "SIZE (1..21)") {
			t.Errorf("want BER constraint error, got %v", err)
		}
	})
	t.Run("encoderRejectsEmptyForwOptions", func(t *testing.T) {
		_, err := strictWire(convertExtForwInfoToWire(&ExtForwInfo{
			SsCode: 0x21,
			ForwardingFeatureList: []ExtForwFeature{{
				SsStatus:          HexBytes{0x05},
				ForwardingOptions: HexBytes{}, // non-nil, empty
			}},
		}))
		if !matchesConstraint(err, "forwardingOptions", "SIZE (1..5)") {
			t.Errorf("want BER constraint error, got %v", err)
		}
	})
}

// Lenient NoReplyConditionTime decode mapping: 1..4 → 5; 31..100 → 30.
func TestExtForwFeatureLenientNoRepCondTime(t *testing.T) {
	cases := []struct {
		name string
		wire int64
		want int
	}{
		{"five", 5, 5},
		{"thirty", 30, 30},
		{"low2MapsToFive", 2, 5},
		{"low4MapsToFive", 4, 5},
		{"high31MapsToThirty", 31, 30},
		{"high99MapsToThirty", 99, 30},
		{"reject0", 0, -1},     // out of spec → error
		{"reject101", 101, -1}, // out of spec → error
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Build a valid feature, then force the wire field.
			in := &ExtForwInfo{
				SsCode: 0x21,
				ForwardingFeatureList: []ExtForwFeature{{
					SsStatus:             HexBytes{0x05},
					NoReplyConditionTime: intPtrV(20),
				}},
			}
			wire, err := strictWire(convertExtForwInfoToWire(in))
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			v := gsm_map.ExtNoRepCondTime(tc.wire)
			wire.ForwardingFeatureList.Values[0].NoReplyConditionTime = &v
			if tc.want < 0 {
				wantConstraintError(t, strictDecodeWire(wire), "forwardingFeatureList.element[0].noReplyConditionTime", "(1..100)")
				return
			}
			got, err := convertWireToExtForwInfo(wire)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			gotV := got.ForwardingFeatureList[0].NoReplyConditionTime
			if gotV == nil || *gotV != tc.want {
				t.Errorf("wire=%d: got %v, want %d", tc.wire, gotV, tc.want)
			}
		})
	}
}

// noRepCondTimeISD returns an InsertSubscriberDataArg provisioning call
// forwarding on no reply (SS-Code cfnry) with NoReplyConditionTime v.
func noRepCondTimeISD(v int) *InsertSubscriberDataArg {
	return &InsertSubscriberDataArg{ProvisionedSS: []ExtSSInfo{{ForwardingInfo: &ExtForwInfo{
		SsCode: 0x2A,
		ForwardingFeatureList: []ExtForwFeature{{
			SsStatus:             HexBytes{0x05},
			ForwardedToNumber:    "31611111111",
			ForwardedToNature:    16,
			ForwardedToPlan:      1,
			NoReplyConditionTime: &v,
		}},
	}}}}
}

// 3GPP TS 29.002 V19.1.0 §17.7.1 Ext-NoRepCondTime ::= INTEGER (1..100):
// "Only values 5-30 are used. Values in the ranges 1-4 and 31-100 are
// reserved for future use". A sender sends only 5..30.
func TestMarshalNoReplyConditionTimeSenderRange(t *testing.T) {
	for _, tc := range []struct {
		v    int
		want error
	}{
		{5, nil},
		{20, nil},
		{30, nil},
		{4, ErrNoReplyConditionTimeOutOfRange},
		{31, ErrNoReplyConditionTimeOutOfRange},
		{1, ErrNoReplyConditionTimeOutOfRange},
		{100, ErrNoReplyConditionTimeOutOfRange},
		{0, ErrNoReplyConditionTimeOutOfRange},
		{101, ErrNoReplyConditionTimeOutOfRange},
		{-1, ErrNoReplyConditionTimeOutOfRange},
	} {
		t.Run(fmt.Sprint(tc.v), func(t *testing.T) {
			data, err := noRepCondTimeISD(tc.v).Marshal()
			if !errors.Is(err, tc.want) {
				t.Fatalf("Marshal: err = %v, want %v", err, tc.want)
			}
			if err != nil {
				return
			}
			got, err := ParseInsertSubscriberData(data)
			if err != nil {
				t.Fatalf("ParseInsertSubscriberData: %v", err)
			}
			gotV := got.ProvisionedSS[0].ForwardingInfo.ForwardingFeatureList[0].NoReplyConditionTime
			if gotV == nil || *gotV != tc.v {
				t.Errorf("NoReplyConditionTime = %v, want %d", gotV, tc.v)
			}
		})
	}
}

// "If received: values 1-4 shall be mapped on to value 5", "values 31-100
// shall be mapped on to value 30" (§17.7.1 Ext-NoRepCondTime). The mapped
// value is one a sender uses, so the parsed message marshals again.
func TestParseNoReplyConditionTimeMapsReserved(t *testing.T) {
	for _, tc := range []struct {
		wire int64
		want int
	}{
		{1, 5},
		{4, 5},
		{5, 5},
		{30, 30},
		{31, 30},
		{100, 30},
	} {
		t.Run(fmt.Sprint(tc.wire), func(t *testing.T) {
			w, err := convertInsertSubscriberDataArgToWire(noRepCondTimeISD(20))
			if err != nil {
				t.Fatalf("convertInsertSubscriberDataArgToWire: %v", err)
			}
			v := gsm_map.ExtNoRepCondTime(tc.wire)
			w.ProvisionedSS.Values[0].ForwardingInfo.ForwardingFeatureList.Values[0].NoReplyConditionTime = &v
			data, err := w.MarshalBER()
			if err != nil {
				t.Fatalf("MarshalBER: %v", err)
			}
			got, err := ParseInsertSubscriberData(data)
			if err != nil {
				t.Fatalf("ParseInsertSubscriberData: %v", err)
			}
			gotV := got.ProvisionedSS[0].ForwardingInfo.ForwardingFeatureList[0].NoReplyConditionTime
			if gotV == nil || *gotV != tc.want {
				t.Errorf("wire %d: NoReplyConditionTime = %v, want %d", tc.wire, gotV, tc.want)
			}
			if err := checkParseRoundTrip("InsertSubscriberData", asParser(ParseInsertSubscriberData), data); err != nil {
				t.Error(err)
			}
		})
	}
}

// --- ExtCallBarInfo ---

func TestExtCallBarInfoRoundTrip(t *testing.T) {
	in := &ExtCallBarInfo{
		SsCode: 0x91, // BAOC
		CallBarringFeatureList: []ExtCallBarringFeature{
			{
				BasicService: &ExtBasicServiceCode{ExtBearerService: HexBytes{0x11}},
				SsStatus:     HexBytes{0x05},
			},
			{SsStatus: HexBytes{0x01, 0x02}},
		},
	}
	wire, err := convertExtCallBarInfoToWire(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, err := convertWireToExtCallBarInfo(wire)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("round-trip (-want +got):\n%s", diff)
	}
}

func TestExtCallBarInfoValidation(t *testing.T) {
	t.Run("emptyList", func(t *testing.T) {
		_, err := strictWire(convertExtCallBarInfoToWire(&ExtCallBarInfo{SsCode: 0x91}))
		if !matchesConstraint(err, "callBarringFeatureList", "SIZE (1..32)") {
			t.Errorf("want BER constraint error, got %v", err)
		}
	})
}

// --- CUGInfo / CUGSubscription / CUGFeature ---

func TestCUGInfoRoundTrip(t *testing.T) {
	cugIdx := 100
	cases := []struct {
		name string
		in   *CUGInfo
	}{
		{
			name: "subOnly",
			in: &CUGInfo{
				CugSubscriptionList: []CUGSubscription{{
					CugIndex:        7,
					CugInterlock:    HexBytes{0x12, 0x34, 0x56, 0x78},
					IntraCUGOptions: IntraCUGNoRestrictions,
				}},
			},
		},
		{
			name: "subPlusFeature",
			in: &CUGInfo{
				CugSubscriptionList: []CUGSubscription{
					{
						CugIndex:        0,
						CugInterlock:    HexBytes{0, 0, 0, 0},
						IntraCUGOptions: IntraCUGICCallBarred,
						BasicServiceGroupList: []ExtBasicServiceCode{
							{ExtBearerService: HexBytes{0x10}},
						},
					},
					{
						CugIndex:        32767,
						CugInterlock:    HexBytes{0xff, 0xff, 0xff, 0xff},
						IntraCUGOptions: IntraCUGOGCallBarred,
					},
				},
				CugFeatureList: []CUGFeature{{
					BasicService:         &ExtBasicServiceCode{ExtTeleservice: HexBytes{0x11}},
					PreferentialCUGIndex: &cugIdx,
					InterCUGRestrictions: 0x03,
				}},
			},
		},
		{
			name: "emptySubList", // spec allows SIZE(0..10)
			in:   &CUGInfo{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertCUGInfoToWire(tc.in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			got, err := convertWireToCUGInfo(wire)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if diff := cmp.Diff(tc.in, got); diff != "" {
				t.Errorf("round-trip (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCUGInfoValidation(t *testing.T) {
	t.Run("subListTooLong", func(t *testing.T) {
		big := make([]CUGSubscription, 10+1)
		for i := range big {
			big[i] = CUGSubscription{
				CugInterlock:    HexBytes{0, 0, 0, 0},
				IntraCUGOptions: IntraCUGNoRestrictions,
			}
		}
		_, err := strictWire(convertCUGInfoToWire(&CUGInfo{CugSubscriptionList: big}))
		if !matchesConstraint(err, "cug-SubscriptionList", "SIZE (0..10)") {
			t.Errorf("want BER constraint error, got %v", err)
		}
	})
	t.Run("cugIndexOutOfRange", func(t *testing.T) {
		_, err := strictWire(convertCUGInfoToWire(&CUGInfo{
			CugSubscriptionList: []CUGSubscription{{
				CugIndex:        32768, // > 32767
				CugInterlock:    HexBytes{0, 0, 0, 0},
				IntraCUGOptions: IntraCUGNoRestrictions,
			}},
		}))
		if !matchesConstraint(err, "cug-Index", "(0..32767)") {
			t.Errorf("want BER constraint error, got %v", err)
		}
	})
	t.Run("interlockWrongLength", func(t *testing.T) {
		_, err := strictWire(convertCUGInfoToWire(&CUGInfo{
			CugSubscriptionList: []CUGSubscription{{
				CugInterlock:    HexBytes{0, 0, 0}, // 3 octets, need 4
				IntraCUGOptions: IntraCUGNoRestrictions,
			}},
		}))
		if !matchesConstraint(err, "cug-Interlock", "SIZE (4)") {
			t.Errorf("want BER constraint error, got %v", err)
		}
	})
	t.Run("invalidIntraOpts", func(t *testing.T) {
		_, err := strictWire(convertCUGInfoToWire(&CUGInfo{
			CugSubscriptionList: []CUGSubscription{{
				CugInterlock:    HexBytes{0, 0, 0, 0},
				IntraCUGOptions: IntraCUGOptions(99),
			}},
		}))
		if !errors.Is(err, ErrIntraCUGOptionsInvalidValue) {
			t.Errorf("want ErrIntraCUGOptionsInvalidValue, got %v", err)
		}
	})
	t.Run("featureListEmpty", func(t *testing.T) {
		_, err := strictWire(convertCUGInfoToWire(&CUGInfo{
			CugSubscriptionList: []CUGSubscription{{CugInterlock: HexBytes{0, 0, 0, 0}, IntraCUGOptions: IntraCUGNoRestrictions}},
			CugFeatureList:      []CUGFeature{}, // non-nil, empty
		}))
		if !matchesConstraint(err, "cug-FeatureList", "SIZE (1..32)") {
			t.Errorf("want BER constraint error, got %v", err)
		}
	})
}

// --- ExtSSData ---

func TestExtSSDataRoundTrip(t *testing.T) {
	cli := CliRestrictionTemporaryDefaultAllowed
	in := &ExtSSData{
		SsCode:               0x09,
		SsStatus:             HexBytes{0x05},
		SsSubscriptionOption: &SSSubscriptionOption{CliRestriction: &cli},
		BasicServiceGroupList: []ExtBasicServiceCode{
			{ExtTeleservice: HexBytes{0x11}},
		},
	}
	wire, err := convertExtSSDataToWire(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, err := convertWireToExtSSData(wire)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("round-trip (-want +got):\n%s", diff)
	}
}

// --- EMLPPInfo ---

func TestEMLPPInfoRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   *EMLPPInfo
	}{
		{"defaults", &EMLPPInfo{MaximumEntitledPriority: 6, DefaultPriority: 0}},
		{"midrange", &EMLPPInfo{MaximumEntitledPriority: 5, DefaultPriority: 4}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertEMLPPInfoToWire(tc.in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			got, err := convertWireToEMLPPInfo(wire)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if diff := cmp.Diff(tc.in, got); diff != "" {
				t.Errorf("round-trip (-want +got):\n%s", diff)
			}
		})
	}
}

func TestEMLPPInfoValidation(t *testing.T) {
	t.Run("encoderRejectsSpare", func(t *testing.T) {
		_, err := convertEMLPPInfoToWire(&EMLPPInfo{MaximumEntitledPriority: 7, DefaultPriority: 4})
		if !errors.Is(err, ErrEMLPPPriorityOutOfRange) {
			t.Errorf("want ErrEMLPPPriorityOutOfRange, got %v", err)
		}
	})
}

// EMLPP lenient decode: wire 7..15 → 4 per spec.
func TestEMLPPInfoLenientDecode(t *testing.T) {
	cases := []struct {
		name string
		wire int64
		want int
		fail bool
	}{
		{"valid0", 0, 0, false},
		{"validA", 6, 6, false},
		{"spare7MapsTo4", 7, 4, false},
		{"spare15MapsTo4", 15, 4, false},
		{"reject16", 16, 0, true},
		{"rejectNeg", -1, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := &gsm_map.EMLPPInfo{
				MaximumentitledPriority: gsm_map.EMLPPPriority(tc.wire),
				DefaultPriority:         gsm_map.EMLPPPriority(0),
			}
			got, err := convertWireToEMLPPInfo(w)
			if tc.fail {
				wantConstraintError(t, strictDecodeWire(w), "maximumentitledPriority", "(0..15)")
				return
			}
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if got.MaximumEntitledPriority != tc.want {
				t.Errorf("wire=%d: got %d, want %d", tc.wire, got.MaximumEntitledPriority, tc.want)
			}
		})
	}
}

// --- Ext-SS-Info CHOICE orchestrator ---

func TestExtSSInfoCHOICERoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   *ExtSSInfo
	}{
		{"forwardingInfo", &ExtSSInfo{
			ForwardingInfo: &ExtForwInfo{
				SsCode:                0x21,
				ForwardingFeatureList: []ExtForwFeature{{SsStatus: HexBytes{0x05}}},
			},
		}},
		{"callBarringInfo", &ExtSSInfo{
			CallBarringInfo: &ExtCallBarInfo{
				SsCode:                 0x91,
				CallBarringFeatureList: []ExtCallBarringFeature{{SsStatus: HexBytes{0x05}}},
			},
		}},
		{"cugInfo", &ExtSSInfo{
			CugInfo: &CUGInfo{
				CugSubscriptionList: []CUGSubscription{{
					CugIndex:        7,
					CugInterlock:    HexBytes{1, 2, 3, 4},
					IntraCUGOptions: IntraCUGNoRestrictions,
				}},
			},
		}},
		{"ssData", &ExtSSInfo{
			SsData: &ExtSSData{
				SsCode:   0x09,
				SsStatus: HexBytes{0x05},
			},
		}},
		{"emlppInfo", &ExtSSInfo{
			EmlppInfo: &EMLPPInfo{MaximumEntitledPriority: 6, DefaultPriority: 0},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertExtSSInfoToWire(tc.in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			got, err := convertWireToExtSSInfo(wire)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if diff := cmp.Diff(tc.in, got); diff != "" {
				t.Errorf("round-trip (-want +got):\n%s", diff)
			}
		})
	}
}

func TestExtSSInfoCHOICEValidation(t *testing.T) {
	t.Run("noAlt", func(t *testing.T) {
		_, err := convertExtSSInfoToWire(&ExtSSInfo{})
		if !errors.Is(err, ErrExtSSInfoChoiceNoAlternative) {
			t.Errorf("want ErrExtSSInfoChoiceNoAlternative, got %v", err)
		}
	})
	t.Run("multipleAlts", func(t *testing.T) {
		_, err := convertExtSSInfoToWire(&ExtSSInfo{
			ForwardingInfo: &ExtForwInfo{
				SsCode:                0x21,
				ForwardingFeatureList: []ExtForwFeature{{SsStatus: HexBytes{0x05}}},
			},
			EmlppInfo: &EMLPPInfo{MaximumEntitledPriority: 6, DefaultPriority: 0},
		})
		if !errors.Is(err, ErrExtSSInfoChoiceMultipleAlternatives) {
			t.Errorf("want ErrExtSSInfoChoiceMultipleAlternatives, got %v", err)
		}
	})
}

// intPtrV is a local int-pointer helper used only by tests in this file.
func intPtrV(v int) *int { return &v }

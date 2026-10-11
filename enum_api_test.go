package gsmmap

import (
	"reflect"
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

func TestEnumFieldTypes(t *testing.T) {
	for _, tc := range []struct {
		name string
		got  any
		want any
	}{
		{"SubscriberStateInfo.NotReachableReason", SubscriberStateInfo{}.NotReachableReason, (*gsm_map.NotReachableReason)(nil)},
		{"PsSubscriberState.NetDetNotReachable", PsSubscriberState{}.NetDetNotReachable, (*gsm_map.NotReachableReason)(nil)},
		{"SubscriberInfo.DaylightSavingTime", SubscriberInfo{}.DaylightSavingTime, (*gsm_map.DaylightSavingTime)(nil)},
		{"VlrCapability.IstSupportIndicator", VlrCapability{}.IstSupportIndicator, (*gsm_map.ISTSupportIndicator)(nil)},
		{"Sri.IstSupportIndicator", Sri{}.IstSupportIndicator, (*gsm_map.ISTSupportIndicator)(nil)},
		{"ExternalSignalInfo.ProtocolID", ExternalSignalInfo{}.ProtocolID, gsm_map.ProtocolId(0)},
		{"ExtExternalSignalInfo.ExtProtocolID", ExtExternalSignalInfo{}.ExtProtocolID, gsm_map.ExtProtocolId(0)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got, want := reflect.TypeOf(tc.got), reflect.TypeOf(tc.want); got != want {
				t.Errorf("type = %v, want %v", got, want)
			}
		})
	}
}

func TestOctetConstantsMatchASN1(t *testing.T) {
	for _, tc := range []struct {
		name string
		got  byte
		want string
	}{
		{"AlertingLevel0", byte(AlertingLevel0), gsm_map.AlertingLevel0},
		{"AlertingLevel1", byte(AlertingLevel1), gsm_map.AlertingLevel1},
		{"AlertingLevel2", byte(AlertingLevel2), gsm_map.AlertingLevel2},
		{"AlertingCategory1", byte(AlertingCategory1), gsm_map.AlertingCategory1},
		{"AlertingCategory2", byte(AlertingCategory2), gsm_map.AlertingCategory2},
		{"AlertingCategory3", byte(AlertingCategory3), gsm_map.AlertingCategory3},
		{"AlertingCategory4", byte(AlertingCategory4), gsm_map.AlertingCategory4},
		{"AlertingCategory5", byte(AlertingCategory5), gsm_map.AlertingCategory5},
		{"SsCodeECT", byte(SsCodeECT), gsm_map.Ect},
		{"SsCodeMultiPTY", byte(SsCodeMultiPTY), gsm_map.MultiPTY},
		{"SsCodeCD", byte(SsCodeCD), gsm_map.Cd},
		{"SsCodeCCBS", byte(SsCodeCCBS), gsm_map.CcbsB},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.want) != 1 {
				t.Fatalf("ASN.1 constant has %d octets, want 1", len(tc.want))
			}
			if tc.got != tc.want[0] {
				t.Errorf("constant = 0x%02x, want ASN.1 value 0x%02x", tc.got, tc.want[0])
			}
		})
	}
}

func TestSubscriberEnumsRoundTrip(t *testing.T) {
	for _, reason := range []NotReachableReason{
		NotReachableReasonMsPurged,
		NotReachableReasonImsiDetached,
		NotReachableReasonRestrictedArea,
		NotReachableReasonNotRegistered,
	} {
		for _, dst := range []DaylightSavingTime{
			DaylightSavingTimeNoAdjustment,
			DaylightSavingTimePlusOneHourAdjustment,
			DaylightSavingTimePlusTwoHoursAdjustment,
		} {
			t.Run(reason.String()+"/"+dst.String(), func(t *testing.T) {
				in := &AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{
					SubscriberState:    &SubscriberStateInfo{State: StateNetDetNotReachable, NotReachableReason: &reason},
					PsSubscriberState:  &PsSubscriberState{NetDetNotReachable: &reason},
					EpsSubscriberState: &PsSubscriberState{NetDetNotReachable: &reason},
					DaylightSavingTime: &dst,
				}}
				data, err := in.Marshal()
				if err != nil {
					t.Fatal(err)
				}
				got, err := ParseAnyTimeInterrogationRes(data)
				if err != nil {
					t.Fatal(err)
				}
				wantEqual(t, "SubscriberInfo", in, got)
			})
		}
	}
}

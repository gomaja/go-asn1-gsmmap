package gsmmap

import (
	"errors"
	"fmt"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// CAMEL phase 1 and CauseValue 0..127 are valid: TS 29.002 V19.1.0 §17.7.1.
func TestCAMELSenderBoundaries(t *testing.T) {
	phase := 1
	in := camelISD()
	in.VlrCamelSubscriptionInfo.OCSI.CamelCapabilityHandling = &phase
	in.VlrCamelSubscriptionInfo.OBcsmCamelTDPCriteriaList[0].OCauseValueCriteria = []int{0, 127}
	in.VlrCamelSubscriptionInfo.TBcsmCamelTDPCriteriaList = []TBcsmCamelTDPCriteria{{TBcsmTriggerDetectionPoint: TBcsmTriggerTBusy, TCauseValueCriteria: []int{0, 127}}}
	data, err := in.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	out, err := ParseInsertSubscriberData(data)
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "CAMEL boundaries", in, out)
}

// Present empty lists fail their SIZE before conversion: TS 29.002 V19.1.0 §17.7.1.
func TestCAMELStrictEmptyLists(t *testing.T) {
	cases := []struct {
		name, path, constraint string
		change                 func(*gsm_map.InsertSubscriberDataArg)
	}{
		{"destination-number", "vlrCamelSubscriptionInfo.o-BcsmCamelTDP-CriteriaList.element[0].destinationNumberCriteria.destinationNumberList", "SIZE (1..10)", func(w *gsm_map.InsertSubscriberDataArg) {
			w.VlrCamelSubscriptionInfo.OBcsmCamelTDPCriteriaList.Values[0].DestinationNumberCriteria = &gsm_map.DestinationNumberCriteria{MatchType: MatchTypeEnabling, DestinationNumberList: &gsm_map.DestinationNumberList{}}
		}},
		{"destination-length", "vlrCamelSubscriptionInfo.o-BcsmCamelTDP-CriteriaList.element[0].destinationNumberCriteria.destinationNumberLengthList", "SIZE (1..3)", func(w *gsm_map.InsertSubscriberDataArg) {
			w.VlrCamelSubscriptionInfo.OBcsmCamelTDPCriteriaList.Values[0].DestinationNumberCriteria = &gsm_map.DestinationNumberCriteria{MatchType: MatchTypeEnabling, DestinationNumberLengthList: &gsm_map.DestinationNumberLengthList{}}
		}},
		{"originating-basic-service", "vlrCamelSubscriptionInfo.o-BcsmCamelTDP-CriteriaList.element[0].basicServiceCriteria", "SIZE (1..5)", func(w *gsm_map.InsertSubscriberDataArg) {
			w.VlrCamelSubscriptionInfo.OBcsmCamelTDPCriteriaList.Values[0].BasicServiceCriteria = &gsm_map.BasicServiceCriteria{}
		}},
		{"terminating-basic-service", "vlrCamelSubscriptionInfo.t-BCSM-CAMEL-TDP-CriteriaList.element[0].basicServiceCriteria", "SIZE (1..5)", func(w *gsm_map.InsertSubscriberDataArg) {
			w.VlrCamelSubscriptionInfo.TBCSMCAMELTDPCriteriaList = &gsm_map.TBCSMCAMELTDPCriteriaList{Values: []gsm_map.TBCSMCAMELTDPCriteria{{TBCSMTriggerDetectionPoint: TBcsmTriggerTBusy, BasicServiceCriteria: &gsm_map.BasicServiceCriteria{}}}}
		}},
		{"dialled-number", "vlrCamelSubscriptionInfo.d-CSI.dp-AnalysedInfoCriteriaList", "SIZE (1..10)", func(w *gsm_map.InsertSubscriberDataArg) {
			w.VlrCamelSubscriptionInfo.DCSI.DpAnalysedInfoCriteriaList = &gsm_map.DPAnalysedInfoCriteriaList{}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := marshalledISDWire(t, camelISD())
			tc.change(w)
			_, err := ParseInsertSubscriberData(tolerantBER(t, w))
			wantConstraintError(t, err, tc.path, tc.constraint)
		})
	}
}

// DestinationNumberLengthList alone supplies criteria: TS 29.002 V19.1.0 §17.7.1.
func TestCAMELDestinationLengthOnly(t *testing.T) {
	in := camelISD()
	in.VlrCamelSubscriptionInfo.OBcsmCamelTDPCriteriaList[0].DestinationNumberCriteria = &DestinationNumberCriteria{MatchType: MatchTypeEnabling, DestinationNumberLengthList: []int{1, 15}}
	data, err := in.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	out, err := ParseInsertSubscriberData(data)
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "length-only destination criteria", in, out)
}

// DefaultCallHandling receiver exceptions: TS 29.002 V19.1.0 §17.7.1.
func TestCAMELDefaultCallMapping(t *testing.T) {
	cases := []struct{ wire, want DefaultCallHandling }{{0, 0}, {1, 1}, {2, 0}, {31, 0}, {32, 1}}
	for _, tc := range cases {
		t.Run(fmt.Sprint(tc.wire), func(t *testing.T) {
			w := marshalledISDWire(t, camelISD())
			for _, p := range wireDefaultCallHandlingFields(w.VlrCamelSubscriptionInfo) {
				*p = tc.wire
			}
			out := parseISD(t, w)
			for name, p := range defaultCallHandlingFields(out.VlrCamelSubscriptionInfo) {
				wantEqual(t, name, tc.want, *p)
			}
		})
	}
}

// SS-CSI sent to the VLR omits CCBS: TS 29.002 V19.1.0 §17.7.1 SS-EventList.
func TestCAMELVLRSSCSI(t *testing.T) {
	in := &InsertSubscriberDataArg{VlrCamelSubscriptionInfo: &VlrCamelSubscriptionInfo{SsCSI: &SSCSI{SsEventList: []SsCode{SsCodeECT}, GsmSCFAddress: testGsmSCF, GsmSCFAddressNature: 0x10, GsmSCFAddressPlan: 1}}}
	data, err := in.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	out, err := ParseInsertSubscriberData(data)
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "VLR SS-CSI", in, out)
	in.VlrCamelSubscriptionInfo.SsCSI.SsEventList = []SsCode{SsCodeCCBS}
	_, err = in.Marshal()
	if !errors.Is(err, ErrSSCSICCBSToVLR) {
		t.Fatalf("CCBS: want %v, got %v", ErrSSCSICCBSToVLR, err)
	}
}

// DefaultSMS-Handling receiver exceptions: TS 29.002 V19.1.0 §17.7.1.
func TestCAMELDefaultSMSMapping(t *testing.T) {
	cases := []struct{ wire, want DefaultSMSHandling }{{0, 0}, {1, 1}, {2, 0}, {31, 0}, {32, 1}}
	for _, tc := range cases {
		t.Run(fmt.Sprint(tc.wire), func(t *testing.T) {
			w := marshalledISDWire(t, camelISD())
			w.VlrCamelSubscriptionInfo.MoSmsCSI.SmsCAMELTDPDataList.Values[0].DefaultSMSHandling = tc.wire
			out := parseISD(t, w)
			wantEqual(t, "default SMS handling", tc.want, out.VlrCamelSubscriptionInfo.MoSmsCSI.SmsCAMELTDPDataList[0].DefaultSMSHandling)
		})
	}
}

// semantic_camel_test.go
//
// CAMEL subscription rules of 3GPP TS 29.002 V19.1.0 §17.7.1 that the BER
// codec cannot express: the CAMEL phase 4 exclusion of sms-SUBMIT-REPORT
// from TPDU-TypeCriterion, and the presence of the SMS-CSI and D-CSI
// components that the ASN.1 marks OPTIONAL. Every case goes through the
// public Marshal and Parse entry points.
package gsmmap

import (
	"errors"
	"testing"

	"github.com/gomaja/go-asn1/runtime/ber"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/google/go-cmp/cmp"

	"github.com/gomaja/go-asn1-gsmmap/address"
)

func semWantEqual[T any](t *testing.T, name string, want, got T) {
	t.Helper()
	if d := cmp.Diff(want, got); d != "" {
		t.Errorf("%s mismatch (-want +got):\n%s", name, d)
	}
}

// semCamelPhase is the CamelCapabilityHandling of every CSI built here.
const semCamelPhase = 4

func semPhase() *int {
	v := semCamelPhase
	return &v
}

func semSMSTDPData(tdp SMSTriggerDetectionPoint) SMSCAMELTDPData {
	return SMSCAMELTDPData{
		SmsTriggerDetectionPoint: tdp,
		ServiceKey:               1,
		GsmSCFAddress:            "31611111111",
		GsmSCFNature:             address.NatureInternational,
		GsmSCFPlan:               address.PlanISDN,
		DefaultSMSHandling:       DefaultSMSHandlingContinueTransaction,
	}
}

func semDPCriterium() DPAnalysedInfoCriterium {
	return DPAnalysedInfoCriterium{
		DialledNumber:       "31622222222",
		DialledNumberNature: address.NatureInternational,
		DialledNumberPlan:   address.PlanISDN,
		ServiceKey:          1,
		GsmSCFAddress:       "31611111111",
		GsmSCFAddressNature: address.NatureInternational,
		GsmSCFAddressPlan:   address.PlanISDN,
		DefaultCallHandling: DefaultCallHandlingContinueCall,
	}
}

// semMarshalParseISD marshals a, parses the result and returns it.
func semMarshalParseISD(t *testing.T, a *InsertSubscriberDataArg) *InsertSubscriberDataArg {
	t.Helper()
	data, err := a.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseInsertSubscriberData(data)
	if err != nil {
		t.Fatalf("ParseInsertSubscriberData: %v", err)
	}
	return got
}

// semParseISDWire encodes w with the strict codec and parses it.
func semParseISDWire(t *testing.T, w *gsm_map.InsertSubscriberDataArg) (*InsertSubscriberDataArg, error) {
	t.Helper()
	data, err := w.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}
	return ParseInsertSubscriberData(data)
}

// --- MT-SMS-TPDU-Type in CAMEL phase 4 ---

// 3GPP TS 29.002 V19.1.0 §17.7.1 MT-SMS-TPDU-Type: "In CAMEL phase 4,
// sms-SUBMIT-REPORT shall not be used and a received TPDU-TypeCriterion
// sequence containing sms-SUBMIT-REPORT shall be wholly ignored."
// MT-smsCAMELTDP-Criteria only exist from CAMEL phase 4 on (§8.8.1 VLR and
// SGSN CAMEL Subscription Info), so the rule applies to every criterion.
func TestMarshalRejectsSubmitReportTPDUType(t *testing.T) {
	submitReport := gsm_map.MTSMSTPDUTypeSmsSUBMITREPORT
	for _, tc := range []struct {
		name string
		tpdu []MTSMSTPDUType
	}{
		{"alone", []MTSMSTPDUType{submitReport}},
		{"after sms-DELIVER", []MTSMSTPDUType{MTSMSTPDUTypeSmsDELIVER, submitReport}},
		{"fifth of five", []MTSMSTPDUType{MTSMSTPDUTypeSmsDELIVER, MTSMSTPDUTypeSmsSTATUSREPORT, MTSMSTPDUTypeSmsDELIVER, MTSMSTPDUTypeSmsSTATUSREPORT, submitReport}},
	} {
		criteria := []MTSmsCAMELTDPCriteria{{SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsDeliveryRequest, TpduTypeCriterion: tc.tpdu}}
		for where, a := range map[string]*InsertSubscriberDataArg{
			"VLR":  {VlrCamelSubscriptionInfo: &VlrCamelSubscriptionInfo{MtSmsCAMELTDPCriteriaList: criteria}},
			"SGSN": {SgsnCAMELSubscriptionInfo: &SGSNCAMELSubscriptionInfo{MtSmsCAMELTDPCriteriaList: criteria}},
		} {
			t.Run(where+"/"+tc.name, func(t *testing.T) {
				if _, err := a.Marshal(); !errors.Is(err, ErrCamelInvalidMTSMSTPDUType) {
					t.Errorf("Marshal: err = %v, want ErrCamelInvalidMTSMSTPDUType", err)
				}
			})
		}
	}
}

func TestMTSMSTPDUTypeListedPhase4ValuesRoundTrip(t *testing.T) {
	criteria := []MTSmsCAMELTDPCriteria{{
		SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsDeliveryRequest,
		TpduTypeCriterion:        []MTSMSTPDUType{MTSMSTPDUTypeSmsDELIVER, MTSMSTPDUTypeSmsSTATUSREPORT},
	}}
	in := &InsertSubscriberDataArg{
		VlrCamelSubscriptionInfo:  &VlrCamelSubscriptionInfo{MtSmsCAMELTDPCriteriaList: criteria},
		SgsnCAMELSubscriptionInfo: &SGSNCAMELSubscriptionInfo{MtSmsCAMELTDPCriteriaList: criteria},
	}
	got := semMarshalParseISD(t, in)
	semWantEqual(t, "ISD", in, got)
}

func TestParseIgnoresTPDUCriterionWithSubmitReport(t *testing.T) {
	listed := []MTSMSTPDUType{MTSMSTPDUTypeSmsDELIVER, MTSMSTPDUTypeSmsSTATUSREPORT}
	criteria := []MTSmsCAMELTDPCriteria{
		{SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsDeliveryRequest, TpduTypeCriterion: listed},
		{SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsDeliveryRequest, TpduTypeCriterion: listed},
		{SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsDeliveryRequest, TpduTypeCriterion: listed},
	}
	w, err := convertInsertSubscriberDataArgToWire(&InsertSubscriberDataArg{
		VlrCamelSubscriptionInfo:  &VlrCamelSubscriptionInfo{MtSmsCAMELTDPCriteriaList: criteria},
		SgsnCAMELSubscriptionInfo: &SGSNCAMELSubscriptionInfo{MtSmsCAMELTDPCriteriaList: criteria},
	})
	if err != nil {
		t.Fatalf("convertInsertSubscriberDataArgToWire: %v", err)
	}
	submitReport := gsm_map.MTSMSTPDUTypeSmsSUBMITREPORT
	for _, l := range []*gsm_map.MTSmsCAMELTDPCriteriaList{
		w.VlrCamelSubscriptionInfo.MtSmsCAMELTDPCriteriaList,
		w.SgsnCAMELSubscriptionInfo.MtSmsCAMELTDPCriteriaList,
	} {
		// sms-SUBMIT-REPORT among listed values, and alone.
		l.Values[1].TpduTypeCriterion = &gsm_map.TPDUTypeCriterion{Values: []gsm_map.MTSMSTPDUType{MTSMSTPDUTypeSmsDELIVER, submitReport}}
		l.Values[2].TpduTypeCriterion = &gsm_map.TPDUTypeCriterion{Values: []gsm_map.MTSMSTPDUType{submitReport}}
	}
	got, err := semParseISDWire(t, w)
	if err != nil {
		t.Fatalf("ParseInsertSubscriberData: %v", err)
	}
	want := []MTSmsCAMELTDPCriteria{
		{SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsDeliveryRequest, TpduTypeCriterion: listed},
		{SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsDeliveryRequest},
		{SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsDeliveryRequest},
	}
	semWantEqual(t, "VLR", want, got.VlrCamelSubscriptionInfo.MtSmsCAMELTDPCriteriaList)
	semWantEqual(t, "SGSN", want, got.SgsnCAMELSubscriptionInfo.MtSmsCAMELTDPCriteriaList)
}

// --- SMS-CSI presence ---

// 3GPP TS 29.002 V19.1.0 §17.7.1 SMS-CSI: "SMS-CAMEL-TDP-Data and
// camelCapabilityHandling shall be present in the SMS-CSI sequence." The
// encoder reports a missing list with the sentinel the decoder uses, so
// errors.Is agrees in both directions.
func TestMarshalSMSCSIWithoutTDPData(t *testing.T) {
	for _, tc := range []struct {
		name string
		list []SMSCAMELTDPData
	}{
		{"nil list", nil},
		{"empty list", []SMSCAMELTDPData{}},
	} {
		for where, build := range map[string]func(*SMSCSI) *InsertSubscriberDataArg{
			"VLR mo-sms-CSI": func(c *SMSCSI) *InsertSubscriberDataArg {
				return &InsertSubscriberDataArg{VlrCamelSubscriptionInfo: &VlrCamelSubscriptionInfo{MoSmsCSI: c}}
			},
			"VLR mt-sms-CSI": func(c *SMSCSI) *InsertSubscriberDataArg {
				return &InsertSubscriberDataArg{VlrCamelSubscriptionInfo: &VlrCamelSubscriptionInfo{MtSmsCSI: c}}
			},
			"SGSN mo-sms-CSI": func(c *SMSCSI) *InsertSubscriberDataArg {
				return &InsertSubscriberDataArg{SgsnCAMELSubscriptionInfo: &SGSNCAMELSubscriptionInfo{MoSmsCSI: c}}
			},
			"SGSN mt-sms-CSI": func(c *SMSCSI) *InsertSubscriberDataArg {
				return &InsertSubscriberDataArg{SgsnCAMELSubscriptionInfo: &SGSNCAMELSubscriptionInfo{MtSmsCSI: c}}
			},
		} {
			t.Run(where+"/"+tc.name, func(t *testing.T) {
				a := build(&SMSCSI{SmsCAMELTDPDataList: tc.list, CamelCapabilityHandling: semPhase()})
				if _, err := a.Marshal(); !errors.Is(err, ErrCamelSMSCSIMissingTDPData) {
					t.Errorf("Marshal: err = %v, want ErrCamelSMSCSIMissingTDPData", err)
				}
			})
		}
	}
}

// The upper bound stays the codec's SIZE (1..10) check.
func TestMarshalSMSCSITDPDataListSize(t *testing.T) {
	list := make([]SMSCAMELTDPData, 10)
	for i := range list {
		list[i] = semSMSTDPData(SMSTriggerDetectionPointSmsCollectedInfo)
	}
	in := &InsertSubscriberDataArg{VlrCamelSubscriptionInfo: &VlrCamelSubscriptionInfo{
		MoSmsCSI: &SMSCSI{SmsCAMELTDPDataList: list, CamelCapabilityHandling: semPhase()},
	}}
	got := semMarshalParseISD(t, in)
	semWantEqual(t, "ISD", in, got)

	in.VlrCamelSubscriptionInfo.MoSmsCSI.SmsCAMELTDPDataList = append(list, semSMSTDPData(SMSTriggerDetectionPointSmsCollectedInfo))
	_, err := in.Marshal()
	var ce *ber.ConstraintError
	if !errors.As(err, &ce) || ce.Path != "sms-CAMEL-TDP-DataList" || ce.Constraint != "SIZE (1..10)" {
		t.Errorf("11 entries: err = %v, want the sms-CAMEL-TDP-DataList SIZE (1..10) violation", err)
	}
}

// --- D-CSI presence ---

// 3GPP TS 29.002 V19.1.0 §17.7.1 D-CSI: "DP-AnalysedInfoCriteria and
// camelCapabilityHandling shall be present in the D-CSI sequence."
func TestMarshalDCSIPresence(t *testing.T) {
	for _, tc := range []struct {
		name string
		dcsi DCSI
		want error
	}{
		{"no criteria", DCSI{CamelCapabilityHandling: semPhase()}, ErrCamelDCSIMissingCriteriaList},
		{"empty criteria", DCSI{DPAnalysedInfoCriteriaList: []DPAnalysedInfoCriterium{}, CamelCapabilityHandling: semPhase()}, ErrCamelDCSIMissingCriteriaList},
		{"no camelCapabilityHandling", DCSI{DPAnalysedInfoCriteriaList: []DPAnalysedInfoCriterium{semDPCriterium()}}, ErrCamelDCSIMissingCapabilityHandling},
		{"neither", DCSI{}, ErrCamelDCSIMissingCriteriaList},
	} {
		t.Run("VLR/"+tc.name, func(t *testing.T) {
			d := tc.dcsi
			a := &InsertSubscriberDataArg{VlrCamelSubscriptionInfo: &VlrCamelSubscriptionInfo{DCSI: &d}}
			if _, err := a.Marshal(); !errors.Is(err, tc.want) {
				t.Errorf("Marshal: err = %v, want %v", err, tc.want)
			}
		})
		t.Run("GMSC/"+tc.name, func(t *testing.T) {
			d := tc.dcsi
			r := &SriResp{
				IMSI:                "204080012345678",
				ExtendedRoutingInfo: &ExtendedRoutingInfo{CamelRoutingInfo: &CamelRoutingInfo{GmscCamelSubscriptionInfo: GmscCamelSubscriptionInfo{DCSI: &d}}},
			}
			if _, err := r.Marshal(); !errors.Is(err, tc.want) {
				t.Errorf("Marshal: err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestDCSIRoundTrip(t *testing.T) {
	list := make([]DPAnalysedInfoCriterium, 10)
	for i := range list {
		list[i] = semDPCriterium()
		list[i].ServiceKey = int64(i)
	}
	for _, n := range []int{1, 10} {
		in := &InsertSubscriberDataArg{VlrCamelSubscriptionInfo: &VlrCamelSubscriptionInfo{
			DCSI: &DCSI{DPAnalysedInfoCriteriaList: list[:n], CamelCapabilityHandling: semPhase()},
		}}
		got := semMarshalParseISD(t, in)
		semWantEqual(t, "ISD", in, got)
	}
}

func TestParseDCSIPresence(t *testing.T) {
	dp, err := convertDPAnalysedInfoCriteriumToWire(&DPAnalysedInfoCriterium{
		DialledNumber: "31622222222", ServiceKey: 1, GsmSCFAddress: "31611111111",
		DefaultCallHandling: DefaultCallHandlingContinueCall,
	})
	if err != nil {
		t.Fatalf("convertDPAnalysedInfoCriteriumToWire: %v", err)
	}
	cch := gsm_map.CamelCapabilityHandling(semCamelPhase)
	list := &gsm_map.DPAnalysedInfoCriteriaList{Values: []gsm_map.DPAnalysedInfoCriterium{dp}}
	for _, tc := range []struct {
		name string
		dcsi gsm_map.DCSI
		want error
	}{
		{"no criteria", gsm_map.DCSI{CamelCapabilityHandling: &cch}, ErrCamelDCSIMissingCriteriaList},
		{"no camelCapabilityHandling", gsm_map.DCSI{DpAnalysedInfoCriteriaList: list}, ErrCamelDCSIMissingCapabilityHandling},
		{"neither", gsm_map.DCSI{}, ErrCamelDCSIMissingCriteriaList},
	} {
		t.Run("VLR/"+tc.name, func(t *testing.T) {
			d := tc.dcsi
			_, err := semParseISDWire(t, &gsm_map.InsertSubscriberDataArg{VlrCamelSubscriptionInfo: &gsm_map.VlrCamelSubscriptionInfo{DCSI: &d}})
			if !errors.Is(err, tc.want) {
				t.Errorf("ParseInsertSubscriberData: err = %v, want %v", err, tc.want)
			}
		})
		t.Run("GMSC/"+tc.name, func(t *testing.T) {
			d := tc.dcsi
			cri := gsm_map.NewExtendedRoutingInfoCamelRoutingInfo(gsm_map.CamelRoutingInfo{
				GmscCamelSubscriptionInfo: gsm_map.GmscCamelSubscriptionInfo{DCsi: &d},
			})
			imsi := gsm_map.IMSI{0x02, 0x04, 0x08, 0x10, 0x32, 0x54, 0x76, 0xf8}
			data, err := (&gsm_map.SendRoutingInfoRes{Imsi: &imsi, ExtendedRoutingInfo: &cri}).MarshalBER()
			if err != nil {
				t.Fatalf("MarshalBER: %v", err)
			}
			if _, err := ParseSriResp(data); !errors.Is(err, tc.want) {
				t.Errorf("ParseSriResp: err = %v, want %v", err, tc.want)
			}
		})
	}
}

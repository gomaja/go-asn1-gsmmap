// semantic_segment_test.go
//
// D-CSI, SMS-CSI and GPRS-CSI may be segmented across the
// InsertSubscriberData operations of one dialogue, and 3GPP TS 29.002
// V19.1.0 §17.7.1 ties the presence of their TDP lists and
// camelCapabilityHandling to the segment: the first segment carries both,
// a subsequent D-CSI segment "shall not contain camelCapabilityHandling, but
// may contain dp-AnalysedInfoCriteriaList". A single message may hold either
// kind of segment, so both kinds marshal, reach the wire exactly as given
// and parse back unchanged.
package gsmmap

import (
	"bytes"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

type semSegment struct {
	name string
	// list and cch report whether the segment carries the TDP list and
	// camelCapabilityHandling.
	list, cch bool
}

var semSegments = []semSegment{
	{"first segment", true, true},
	{"subsequent segment", true, false},
	{"subsequent segment without list", false, false},
}

func semSegmentCases(s semSegment) map[string]*InsertSubscriberDataArg {
	var cch *int
	if s.cch {
		cch = semPhase()
	}
	var dp []DPAnalysedInfoCriterium
	var mo, mt []SMSCAMELTDPData
	var gprs GPRSCamelTDPDataList
	if s.list {
		dp = []DPAnalysedInfoCriterium{semDPCriterium()}
		mo = []SMSCAMELTDPData{semSMSTDPData(SMSTriggerDetectionPointSmsCollectedInfo)}
		mt = []SMSCAMELTDPData{semSMSTDPData(SMSTriggerDetectionPointSmsDeliveryRequest)}
		gprs = GPRSCamelTDPDataList{{
			GprsTriggerDetectionPoint: GPRSTDPAttach,
			ServiceKey:                1,
			GsmSCFAddress:             "31611111111",
		}}
	}
	return map[string]*InsertSubscriberDataArg{
		"VLR D-CSI": {VlrCamelSubscriptionInfo: &VlrCamelSubscriptionInfo{
			DCSI: &DCSI{DPAnalysedInfoCriteriaList: dp, CamelCapabilityHandling: cch},
		}},
		"VLR MO-SMS-CSI": {VlrCamelSubscriptionInfo: &VlrCamelSubscriptionInfo{
			MoSmsCSI: &SMSCSI{SmsCAMELTDPDataList: mo, CamelCapabilityHandling: cch},
		}},
		"VLR MT-SMS-CSI": {VlrCamelSubscriptionInfo: &VlrCamelSubscriptionInfo{
			MtSmsCSI: &SMSCSI{SmsCAMELTDPDataList: mt, CamelCapabilityHandling: cch},
		}},
		"SGSN MO-SMS-CSI": {SgsnCAMELSubscriptionInfo: &SGSNCAMELSubscriptionInfo{
			MoSmsCSI: &SMSCSI{SmsCAMELTDPDataList: mo, CamelCapabilityHandling: cch},
		}},
		"SGSN MT-SMS-CSI": {SgsnCAMELSubscriptionInfo: &SGSNCAMELSubscriptionInfo{
			MtSmsCSI: &SMSCSI{SmsCAMELTDPDataList: mt, CamelCapabilityHandling: cch},
		}},
		"SGSN GPRS-CSI": {SgsnCAMELSubscriptionInfo: &SGSNCAMELSubscriptionInfo{
			GprsCSI: &GPRSCSI{GprsCamelTDPDataList: gprs, CamelCapabilityHandling: cch},
		}},
	}
}

// semWireCSIFields returns, for the single CSI of a semSegmentCases message,
// whether its wire TDP list and camelCapabilityHandling are present.
func semWireCSIFields(w *gsm_map.InsertSubscriberDataArg) (list, cch bool) {
	if v := w.VlrCamelSubscriptionInfo; v != nil {
		switch {
		case v.DCSI != nil:
			return v.DCSI.DpAnalysedInfoCriteriaList != nil, v.DCSI.CamelCapabilityHandling != nil
		case v.MoSmsCSI != nil:
			return v.MoSmsCSI.SmsCAMELTDPDataList != nil, v.MoSmsCSI.CamelCapabilityHandling != nil
		case v.MtSmsCSI != nil:
			return v.MtSmsCSI.SmsCAMELTDPDataList != nil, v.MtSmsCSI.CamelCapabilityHandling != nil
		}
	}
	if s := w.SgsnCAMELSubscriptionInfo; s != nil {
		switch {
		case s.MoSmsCSI != nil:
			return s.MoSmsCSI.SmsCAMELTDPDataList != nil, s.MoSmsCSI.CamelCapabilityHandling != nil
		case s.MtSmsCSI != nil:
			return s.MtSmsCSI.SmsCAMELTDPDataList != nil, s.MtSmsCSI.CamelCapabilityHandling != nil
		case s.GprsCSI != nil:
			return s.GprsCSI.GprsCamelTDPDataList != nil, s.GprsCSI.CamelCapabilityHandling != nil
		}
	}
	panic("no CSI in message")
}

func TestSegmentedCSIRoundTrip(t *testing.T) {
	for _, seg := range semSegments {
		for name, in := range semSegmentCases(seg) {
			t.Run(name+"/"+seg.name, func(t *testing.T) {
				data, err := in.Marshal()
				if err != nil {
					t.Fatalf("Marshal: %v", err)
				}
				// On the wire each component is present exactly when set.
				var w gsm_map.InsertSubscriberDataArg
				if err := w.UnmarshalBER(data); err != nil {
					t.Fatalf("UnmarshalBER: %v", err)
				}
				if list, cch := semWireCSIFields(&w); list != seg.list || cch != seg.cch {
					t.Errorf("wire list present = %t, camelCapabilityHandling present = %t; want %t, %t", list, cch, seg.list, seg.cch)
				}
				got, err := ParseInsertSubscriberData(data)
				if err != nil {
					t.Fatalf("ParseInsertSubscriberData: %v", err)
				}
				semWantEqual(t, "ISD", in, got)
				again, err := got.Marshal()
				if err != nil {
					t.Fatalf("Marshal of the parsed value: %v", err)
				}
				if !bytes.Equal(again, data) {
					t.Errorf("re-marshalled %x, want %x", again, data)
				}
			})
		}
	}
}

// The GMSC CAMEL subscription info of a SendRoutingInfo result carries a
// D-CSI as well.
func TestSegmentedGmscDCSIRoundTrip(t *testing.T) {
	for _, seg := range semSegments {
		t.Run(seg.name, func(t *testing.T) {
			d := &DCSI{}
			if seg.list {
				d.DPAnalysedInfoCriteriaList = []DPAnalysedInfoCriterium{semDPCriterium()}
			}
			if seg.cch {
				d.CamelCapabilityHandling = semPhase()
			}
			in := &SriResp{
				IMSI:                "204080012345678",
				ExtendedRoutingInfo: &ExtendedRoutingInfo{CamelRoutingInfo: &CamelRoutingInfo{GmscCamelSubscriptionInfo: GmscCamelSubscriptionInfo{DCSI: d}}},
			}
			data, err := in.Marshal()
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			got, err := ParseSriResp(data)
			if err != nil {
				t.Fatalf("ParseSriResp: %v", err)
			}
			semWantEqual(t, "D-CSI", d, got.ExtendedRoutingInfo.CamelRoutingInfo.GmscCamelSubscriptionInfo.DCSI)
		})
	}
}

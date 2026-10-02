// receiver_exception_test.go
//
// Receiver exception handling of 3GPP TS 29.002 V19.1.0 §17.7: values the
// specification tells a receiver to map, discard or ignore. Each case
// converts a valid public message to its wire struct, overwrites one field
// with the value under test, encodes the struct with go-asn1's strict
// MarshalBER and decodes the bytes with the public Parse* entry point, so a
// passing case proves that the codec admits the value and that Parse applies
// the rule. The encoder cases check that Marshal still refuses the values a
// sender must not send.
package gsmmap

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/gomaja/go-asn1/runtime/ber"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/google/go-cmp/cmp"

	"github.com/gomaja/go-asn1-gsmmap/address"
)

// --- helpers ---

type berMarshaler interface {
	MarshalBER(opts ...ber.EncodeOption) ([]byte, error)
}

// strictBER encodes w with go-asn1's strict encoder.
func strictBER(t *testing.T, w berMarshaler) []byte {
	t.Helper()
	data, err := w.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}
	return data
}

// tolerantBER encodes w although a field violates its ASN.1 constraint, to
// show that the strict decoder rejects the first value beyond a range the
// receiver maps.
func tolerantBER(t *testing.T, w berMarshaler) []byte {
	t.Helper()
	var log ber.ViolationLog
	data, err := w.MarshalBER(ber.WithConstraintTolerance(&log))
	if err != nil {
		t.Fatalf("MarshalBER with constraint tolerance: %v", err)
	}
	return data
}

const testGsmSCF = "31611111111"

func oTDPData(tdp OBcsmTriggerDetectionPoint, serviceKey int64) OBcsmCamelTDPData {
	return OBcsmCamelTDPData{
		OBcsmTriggerDetectionPoint: tdp,
		ServiceKey:                 serviceKey,
		GsmSCFAddress:              testGsmSCF,
		GsmSCFAddressNature:        address.NatureInternational,
		GsmSCFAddressPlan:          address.PlanISDN,
		DefaultCallHandling:        DefaultCallHandlingContinueCall,
	}
}

func tTDPData(tdp TBcsmTriggerDetectionPoint, serviceKey int64) TBcsmCamelTDPData {
	return TBcsmCamelTDPData{
		TBcsmTriggerDetectionPoint: tdp,
		ServiceKey:                 serviceKey,
		GsmSCFAddress:              testGsmSCF,
		GsmSCFAddressNature:        address.NatureInternational,
		GsmSCFAddressPlan:          address.PlanISDN,
		DefaultCallHandling:        DefaultCallHandlingContinueCall,
	}
}

func dpCriterium(serviceKey int64) DPAnalysedInfoCriterium {
	return DPAnalysedInfoCriterium{
		DialledNumber:       "31622222222",
		DialledNumberNature: address.NatureInternational,
		DialledNumberPlan:   address.PlanISDN,
		ServiceKey:          serviceKey,
		GsmSCFAddress:       testGsmSCF,
		GsmSCFAddressNature: address.NatureInternational,
		GsmSCFAddressPlan:   address.PlanISDN,
		DefaultCallHandling: DefaultCallHandlingContinueCall,
	}
}

func smsTDPData(tdp SMSTriggerDetectionPoint, serviceKey int64) SMSCAMELTDPData {
	return SMSCAMELTDPData{
		SmsTriggerDetectionPoint: tdp,
		ServiceKey:               serviceKey,
		GsmSCFAddress:            testGsmSCF,
		GsmSCFNature:             address.NatureInternational,
		GsmSCFPlan:               address.PlanISDN,
		DefaultSMSHandling:       DefaultSMSHandlingContinueTransaction,
	}
}

func smsCSI(tdp SMSTriggerDetectionPoint, serviceKey int64) *SMSCSI {
	cch := 4
	return &SMSCSI{
		SmsCAMELTDPDataList:     []SMSCAMELTDPData{smsTDPData(tdp, serviceKey)},
		CamelCapabilityHandling: &cch,
	}
}

func mtSMSCriteria() []MTSmsCAMELTDPCriteria {
	return []MTSmsCAMELTDPCriteria{{SmsTriggerDetectionPoint: mtSMSTriggerDetectionPoint}}
}

// camelISD returns an InsertSubscriberDataArg carrying every VLR and SGSN
// CSI and criteria list whose receiver rules these tests exercise. Entries
// are told apart by their service keys.
func camelISD() *InsertSubscriberDataArg {
	cch := 4
	return &InsertSubscriberDataArg{
		VlrCamelSubscriptionInfo: &VlrCamelSubscriptionInfo{
			OCSI: &OCSI{
				OBcsmCamelTDPDataList:   []OBcsmCamelTDPData{oTDPData(OBcsmTriggerCollectedInfo, 1)},
				CamelCapabilityHandling: &cch,
			},
			OBcsmCamelTDPCriteriaList: []OBcsmCamelTDPCriteria{{OBcsmTriggerDetectionPoint: OBcsmTriggerCollectedInfo}},
			MoSmsCSI:                  smsCSI(moSMSTriggerDetectionPoint, 2),
			VtCSI: &TCSI{
				TBcsmCamelTDPDataList:   []TBcsmCamelTDPData{tTDPData(TBcsmTriggerTermAttemptAuthorized, 3)},
				CamelCapabilityHandling: &cch,
			},
			DCSI: &DCSI{
				DPAnalysedInfoCriteriaList: []DPAnalysedInfoCriterium{dpCriterium(4)},
				CamelCapabilityHandling:    &cch,
			},
			MtSmsCSI:                  smsCSI(mtSMSTriggerDetectionPoint, 5),
			MtSmsCAMELTDPCriteriaList: mtSMSCriteria(),
		},
		SgsnCAMELSubscriptionInfo: &SGSNCAMELSubscriptionInfo{
			GprsCSI: &GPRSCSI{
				GprsCamelTDPDataList:    GPRSCamelTDPDataList{makeGPRSCamelTDPData()},
				CamelCapabilityHandling: &cch,
			},
			MoSmsCSI:                  smsCSI(moSMSTriggerDetectionPoint, 6),
			MtSmsCSI:                  smsCSI(mtSMSTriggerDetectionPoint, 7),
			MtSmsCAMELTDPCriteriaList: mtSMSCriteria(),
		},
	}
}

func isdWire(t *testing.T, a *InsertSubscriberDataArg) *gsm_map.InsertSubscriberDataArg {
	t.Helper()
	w, err := convertInsertSubscriberDataArgToWire(a)
	if err != nil {
		t.Fatalf("convertInsertSubscriberDataArgToWire: %v", err)
	}
	return w
}

func parseISD(t *testing.T, w *gsm_map.InsertSubscriberDataArg) *InsertSubscriberDataArg {
	t.Helper()
	got, err := ParseInsertSubscriberData(strictBER(t, w))
	if err != nil {
		t.Fatalf("ParseInsertSubscriberData: %v", err)
	}
	return got
}

// gmscCamel returns the GMSC CAMEL subscription info of a
// SendRoutingInfo result, with every CSI and criteria list these tests
// exercise.
func gmscCamel() GmscCamelSubscriptionInfo {
	cch := 4
	return GmscCamelSubscriptionInfo{
		TCSI: &TCSI{
			TBcsmCamelTDPDataList:   []TBcsmCamelTDPData{tTDPData(TBcsmTriggerTermAttemptAuthorized, 1)},
			CamelCapabilityHandling: &cch,
		},
		OCSI: &OCSI{
			OBcsmCamelTDPDataList:   []OBcsmCamelTDPData{oTDPData(OBcsmTriggerCollectedInfo, 2)},
			CamelCapabilityHandling: &cch,
		},
		OBcsmCamelTDPCriteriaList: []OBcsmCamelTDPCriteria{{OBcsmTriggerDetectionPoint: OBcsmTriggerCollectedInfo}},
		DCSI: &DCSI{
			DPAnalysedInfoCriteriaList: []DPAnalysedInfoCriterium{dpCriterium(3)},
			CamelCapabilityHandling:    &cch,
		},
	}
}

func sriRespWire(t *testing.T, g GmscCamelSubscriptionInfo) *gsm_map.SendRoutingInfoRes {
	t.Helper()
	w, err := convertSriRespToRes(&SriResp{
		IMSI:                "204080012345678",
		ExtendedRoutingInfo: &ExtendedRoutingInfo{CamelRoutingInfo: &CamelRoutingInfo{GmscCamelSubscriptionInfo: g}},
	})
	if err != nil {
		t.Fatalf("convertSriRespToRes: %v", err)
	}
	return w
}

func parseSriRespGmsc(t *testing.T, w *gsm_map.SendRoutingInfoRes) GmscCamelSubscriptionInfo {
	t.Helper()
	got, err := ParseSriResp(strictBER(t, w))
	if err != nil {
		t.Fatalf("ParseSriResp: %v", err)
	}
	return got.ExtendedRoutingInfo.CamelRoutingInfo.GmscCamelSubscriptionInfo
}

func wantEqual[T any](t *testing.T, name string, want, got T) {
	t.Helper()
	if d := cmp.Diff(want, got); d != "" {
		t.Errorf("%s mismatch (-want +got):\n%s", name, d)
	}
}

// --- CamelCapabilityHandling, §17.7.1 ---

// capabilityFields returns the CamelCapabilityHandling of every CSI in a
// camelISD, by CSI name.
func capabilityFields(a *InsertSubscriberDataArg) map[string]**int {
	v, s := a.VlrCamelSubscriptionInfo, a.SgsnCAMELSubscriptionInfo
	return map[string]**int{
		"O-CSI":       &v.OCSI.CamelCapabilityHandling,
		"VT-CSI":      &v.VtCSI.CamelCapabilityHandling,
		"D-CSI":       &v.DCSI.CamelCapabilityHandling,
		"MO-SMS-CSI":  &v.MoSmsCSI.CamelCapabilityHandling,
		"MT-SMS-CSI":  &v.MtSmsCSI.CamelCapabilityHandling,
		"GPRS-CSI":    &s.GprsCSI.CamelCapabilityHandling,
		"SGSN MO-SMS": &s.MoSmsCSI.CamelCapabilityHandling,
		"SGSN MT-SMS": &s.MtSmsCSI.CamelCapabilityHandling,
	}
}

// wireCapabilityFields is capabilityFields for the wire struct.
func wireCapabilityFields(w *gsm_map.InsertSubscriberDataArg) map[string]**gsm_map.CamelCapabilityHandling {
	v, s := w.VlrCamelSubscriptionInfo, w.SgsnCAMELSubscriptionInfo
	return map[string]**gsm_map.CamelCapabilityHandling{
		"O-CSI":       &v.OCSI.CamelCapabilityHandling,
		"VT-CSI":      &v.VtCSI.CamelCapabilityHandling,
		"D-CSI":       &v.DCSI.CamelCapabilityHandling,
		"MO-SMS-CSI":  &v.MoSmsCSI.CamelCapabilityHandling,
		"MT-SMS-CSI":  &v.MtSmsCSI.CamelCapabilityHandling,
		"GPRS-CSI":    &s.GprsCSI.CamelCapabilityHandling,
		"SGSN MO-SMS": &s.MoSmsCSI.CamelCapabilityHandling,
		"SGSN MT-SMS": &s.MtSmsCSI.CamelCapabilityHandling,
	}
}

// "reception of values greater than 4 shall be treated as CAMEL phase 4."
func TestParseCamelCapabilityHandlingAbovePhase4(t *testing.T) {
	cases := []struct {
		wire gsm_map.CamelCapabilityHandling
		want int
	}{
		{1, 1},  // lowest phase
		{4, 4},  // last defined phase
		{5, 4},  // lower bound of the mapped range
		{16, 4}, // upper bound of the mapped range, INTEGER (1..16)
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("VLR and SGSN, wire %d", tc.wire), func(t *testing.T) {
			w := isdWire(t, camelISD())
			for _, p := range wireCapabilityFields(w) {
				v := tc.wire
				*p = &v
			}
			for name, p := range capabilityFields(parseISD(t, w)) {
				if *p == nil || **p != tc.want {
					t.Errorf("%s: got %v, want %d", name, *p, tc.want)
				}
			}
		})
		t.Run(fmt.Sprintf("GMSC, wire %d", tc.wire), func(t *testing.T) {
			w := sriRespWire(t, gmscCamel())
			g := &w.ExtendedRoutingInfo.CamelRoutingInfo.GmscCamelSubscriptionInfo
			for _, p := range []**gsm_map.CamelCapabilityHandling{
				&g.TCSI.CamelCapabilityHandling, &g.OCSI.CamelCapabilityHandling, &g.DCsi.CamelCapabilityHandling,
			} {
				v := tc.wire
				*p = &v
			}
			got := parseSriRespGmsc(t, w)
			for name, p := range map[string]*int{
				"T-CSI": got.TCSI.CamelCapabilityHandling,
				"O-CSI": got.OCSI.CamelCapabilityHandling,
				"D-CSI": got.DCSI.CamelCapabilityHandling,
			} {
				if p == nil || *p != tc.want {
					t.Errorf("%s: got %v, want %d", name, p, tc.want)
				}
			}
		})
	}
	t.Run("wire 17 violates INTEGER (1..16)", func(t *testing.T) {
		w := isdWire(t, camelISD())
		v := gsm_map.CamelCapabilityHandling(17)
		w.VlrCamelSubscriptionInfo.OCSI.CamelCapabilityHandling = &v
		_, err := ParseInsertSubscriberData(tolerantBER(t, w))
		wantConstraintError(t, err, "vlrcamelsubscriptioninfo.ocsi.camelCapabilityHandling", "(1..16)")
	})
}

// A sender uses only the defined phases 1 to 4.
func TestMarshalRejectsUndefinedCamelCapabilityHandling(t *testing.T) {
	for name := range capabilityFields(camelISD()) {
		for _, v := range []int{0, 5, 16} {
			a := camelISD()
			*capabilityFields(a)[name] = &v
			if _, err := a.Marshal(); !errors.Is(err, ErrCamelCapabilityHandlingOutOfRange) {
				t.Errorf("%s = %d: got %v, want ErrCamelCapabilityHandlingOutOfRange", name, v, err)
			}
		}
	}
}

// --- DefaultCallHandling, §17.7.1 ---

// defaultCallHandlingFields returns the DefaultCallHandling of the first
// O-BCSM, T-BCSM and D-CSI entry of a camelISD, by container name.
func defaultCallHandlingFields(v *VlrCamelSubscriptionInfo) map[string]*DefaultCallHandling {
	return map[string]*DefaultCallHandling{
		"O-BCSM": &v.OCSI.OBcsmCamelTDPDataList[0].DefaultCallHandling,
		"T-BCSM": &v.VtCSI.TBcsmCamelTDPDataList[0].DefaultCallHandling,
		"D-CSI":  &v.DCSI.DPAnalysedInfoCriteriaList[0].DefaultCallHandling,
	}
}

// wireDefaultCallHandlingFields is defaultCallHandlingFields for the wire
// struct.
func wireDefaultCallHandlingFields(w *gsm_map.VlrCamelSubscriptionInfo) map[string]*DefaultCallHandling {
	return map[string]*DefaultCallHandling{
		"O-BCSM": &w.OCSI.OBcsmCamelTDPDataList.Values[0].DefaultCallHandling,
		"T-BCSM": &w.VtCSI.TBcsmCamelTDPDataList.Values[0].DefaultCallHandling,
		"D-CSI":  &w.DCSI.DpAnalysedInfoCriteriaList.Values[0].DefaultCallHandling,
	}
}

// "reception of values in range 2-31 shall be treated as "continueCall"";
// "reception of values greater than 31 shall be treated as "releaseCall"".
func TestParseDefaultCallHandlingMapping(t *testing.T) {
	cases := []struct {
		wire, want DefaultCallHandling
	}{
		{0, DefaultCallHandlingContinueCall},
		{1, DefaultCallHandlingReleaseCall}, // last defined value
		{2, DefaultCallHandlingContinueCall},
		{31, DefaultCallHandlingContinueCall},
		{32, DefaultCallHandlingReleaseCall},
		{math.MaxInt64, DefaultCallHandlingReleaseCall},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("wire %d", tc.wire), func(t *testing.T) {
			w := isdWire(t, camelISD())
			for _, p := range wireDefaultCallHandlingFields(w.VlrCamelSubscriptionInfo) {
				*p = tc.wire
			}
			for name, p := range defaultCallHandlingFields(parseISD(t, w).VlrCamelSubscriptionInfo) {
				if *p != tc.want {
					t.Errorf("%s: got %d, want %d", name, *p, tc.want)
				}
			}
		})
	}
	// Negative values lie outside both ranges of the clause.
	for name := range defaultCallHandlingFields(camelISD().VlrCamelSubscriptionInfo) {
		for _, v := range []DefaultCallHandling{-1, math.MinInt64} {
			w := isdWire(t, camelISD())
			*wireDefaultCallHandlingFields(w.VlrCamelSubscriptionInfo)[name] = v
			if _, err := ParseInsertSubscriberData(strictBER(t, w)); !errors.Is(err, ErrCamelInvalidDefaultCallHandling) {
				t.Errorf("%s wire %d: got %v, want ErrCamelInvalidDefaultCallHandling", name, v, err)
			}
		}
	}
}

func TestMarshalRejectsUndefinedDefaultCallHandling(t *testing.T) {
	for name := range defaultCallHandlingFields(camelISD().VlrCamelSubscriptionInfo) {
		for _, d := range []DefaultCallHandling{-1, 2, 32} {
			a := camelISD()
			*defaultCallHandlingFields(a.VlrCamelSubscriptionInfo)[name] = d
			if _, err := a.Marshal(); !errors.Is(err, ErrCamelInvalidDefaultCallHandling) {
				t.Errorf("%s = %d: got %v, want ErrCamelInvalidDefaultCallHandling", name, d, err)
			}
		}
	}
}

// --- O-BcsmTriggerDetectionPoint, §17.7.1 ---

// "For O-BcsmCamelTDPData sequences containing this parameter with any
// other value than the ones listed the receiver shall ignore the whole
// O-BcsmCamelTDPData sequence." and the same for O-BcsmCamelTDP-Criteria.
// The listed values are collectedInfo (2) and routeSelectFailure (4).
func TestParseOBcsmTriggerDetectionPointIgnoresSequence(t *testing.T) {
	kept := []OBcsmCamelTDPData{
		oTDPData(OBcsmTriggerCollectedInfo, 10),
		oTDPData(OBcsmTriggerRouteSelectFailure, 12),
	}
	forwarded := CallTypeCriteriaForwarded
	keptCriteria := []OBcsmCamelTDPCriteria{
		{OBcsmTriggerDetectionPoint: OBcsmTriggerCollectedInfo, CallTypeCriteria: &forwarded},
		{OBcsmTriggerDetectionPoint: OBcsmTriggerRouteSelectFailure},
	}
	// ignoredData and ignoredCriteria carry the unlisted TDP and a field the
	// decoder would reject, which proves that it skips the whole sequence.
	ignoredData := func(w []gsm_map.OBcsmCamelTDPData, tdp OBcsmTriggerDetectionPoint) []gsm_map.OBcsmCamelTDPData {
		e := w[0]
		e.OBcsmTriggerDetectionPoint, e.ServiceKey, e.DefaultCallHandling = tdp, 11, -1
		return slices.Insert(w, 1, e)
	}
	ignoredCriteria := func(w []gsm_map.OBcsmCamelTDPCriteria, tdp OBcsmTriggerDetectionPoint) []gsm_map.OBcsmCamelTDPCriteria {
		bad := CallTypeCriteria(99)
		return slices.Insert(w, 1, gsm_map.OBcsmCamelTDPCriteria{OBcsmTriggerDetectionPoint: tdp, CallTypeCriteria: &bad})
	}
	for _, tdp := range []OBcsmTriggerDetectionPoint{1, 3, 5, -1, math.MaxInt64} {
		t.Run(fmt.Sprintf("VLR, wire %d", tdp), func(t *testing.T) {
			in := camelISD()
			in.VlrCamelSubscriptionInfo.OCSI.OBcsmCamelTDPDataList = kept
			in.VlrCamelSubscriptionInfo.OBcsmCamelTDPCriteriaList = keptCriteria
			w := isdWire(t, in)
			vlr := w.VlrCamelSubscriptionInfo
			vlr.OCSI.OBcsmCamelTDPDataList.Values = ignoredData(vlr.OCSI.OBcsmCamelTDPDataList.Values, tdp)
			vlr.OBcsmCamelTDPCriteriaList.Values = ignoredCriteria(vlr.OBcsmCamelTDPCriteriaList.Values, tdp)
			got := parseISD(t, w).VlrCamelSubscriptionInfo
			wantEqual(t, "O-CSI", kept, got.OCSI.OBcsmCamelTDPDataList)
			wantEqual(t, "O-BcsmCamelTDP-CriteriaList", keptCriteria, got.OBcsmCamelTDPCriteriaList)
		})
		t.Run(fmt.Sprintf("GMSC, wire %d", tdp), func(t *testing.T) {
			g := gmscCamel()
			g.OCSI.OBcsmCamelTDPDataList = kept
			g.OBcsmCamelTDPCriteriaList = keptCriteria
			w := sriRespWire(t, g)
			wg := &w.ExtendedRoutingInfo.CamelRoutingInfo.GmscCamelSubscriptionInfo
			wg.OCSI.OBcsmCamelTDPDataList.Values = ignoredData(wg.OCSI.OBcsmCamelTDPDataList.Values, tdp)
			wg.OBcsmCamelTDPCriteriaList.Values = ignoredCriteria(wg.OBcsmCamelTDPCriteriaList.Values, tdp)
			got := parseSriRespGmsc(t, w)
			wantEqual(t, "O-CSI", kept, got.OCSI.OBcsmCamelTDPDataList)
			wantEqual(t, "O-BcsmCamelTDP-CriteriaList", keptCriteria, got.OBcsmCamelTDPCriteriaList)
		})
	}
	// With every entry ignored the O-CSI and the criteria list are absent;
	// the other CSIs survive.
	t.Run("VLR, all ignored", func(t *testing.T) {
		w := isdWire(t, camelISD())
		vlr := w.VlrCamelSubscriptionInfo
		vlr.OCSI.OBcsmCamelTDPDataList.Values[0].OBcsmTriggerDetectionPoint = 3
		vlr.OBcsmCamelTDPCriteriaList.Values[0].OBcsmTriggerDetectionPoint = 3
		got := parseISD(t, w).VlrCamelSubscriptionInfo
		if got.OCSI != nil || got.OBcsmCamelTDPCriteriaList != nil {
			t.Errorf("O-CSI %+v, criteria %+v: want both absent", got.OCSI, got.OBcsmCamelTDPCriteriaList)
		}
		if got.VtCSI == nil || got.DCSI == nil || got.MoSmsCSI == nil {
			t.Errorf("sibling CSIs lost: %+v", got)
		}
	})
	t.Run("GMSC, all ignored", func(t *testing.T) {
		w := sriRespWire(t, gmscCamel())
		wg := &w.ExtendedRoutingInfo.CamelRoutingInfo.GmscCamelSubscriptionInfo
		wg.OCSI.OBcsmCamelTDPDataList.Values[0].OBcsmTriggerDetectionPoint = 3
		wg.OBcsmCamelTDPCriteriaList.Values[0].OBcsmTriggerDetectionPoint = 3
		got := parseSriRespGmsc(t, w)
		if got.OCSI != nil || got.OBcsmCamelTDPCriteriaList != nil {
			t.Errorf("O-CSI %+v, criteria %+v: want both absent", got.OCSI, got.OBcsmCamelTDPCriteriaList)
		}
		if got.TCSI == nil || got.DCSI == nil {
			t.Errorf("sibling CSIs lost: %+v", got)
		}
	})
}

// --- T-BcsmTriggerDetectionPoint, §17.7.1 ---

// "For T-BcsmCamelTDPData sequences containing this parameter with any
// other value than the ones listed above, the receiver shall ignore the
// whole T-BcsmCamelTDPData sequence." The listed values are
// termAttemptAuthorized (12), tBusy (13) and tNoAnswer (14).
func TestParseTBcsmTriggerDetectionPointIgnoresSequence(t *testing.T) {
	kept := []TBcsmCamelTDPData{
		tTDPData(TBcsmTriggerTermAttemptAuthorized, 20),
		tTDPData(TBcsmTriggerTNoAnswer, 22),
	}
	ignored := func(w []gsm_map.TBcsmCamelTDPData, tdp TBcsmTriggerDetectionPoint) []gsm_map.TBcsmCamelTDPData {
		e := w[0]
		e.TBcsmTriggerDetectionPoint, e.ServiceKey, e.DefaultCallHandling = tdp, 21, -1
		return slices.Insert(w, 1, e)
	}
	for _, tdp := range []TBcsmTriggerDetectionPoint{11, 15, -1, math.MaxInt64} {
		t.Run(fmt.Sprintf("VT-CSI, wire %d", tdp), func(t *testing.T) {
			in := camelISD()
			in.VlrCamelSubscriptionInfo.VtCSI.TBcsmCamelTDPDataList = kept
			w := isdWire(t, in)
			vt := w.VlrCamelSubscriptionInfo.VtCSI
			vt.TBcsmCamelTDPDataList.Values = ignored(vt.TBcsmCamelTDPDataList.Values, tdp)
			wantEqual(t, "VT-CSI", kept, parseISD(t, w).VlrCamelSubscriptionInfo.VtCSI.TBcsmCamelTDPDataList)
		})
		t.Run(fmt.Sprintf("GMSC T-CSI, wire %d", tdp), func(t *testing.T) {
			g := gmscCamel()
			g.TCSI.TBcsmCamelTDPDataList = kept
			w := sriRespWire(t, g)
			tc := w.ExtendedRoutingInfo.CamelRoutingInfo.GmscCamelSubscriptionInfo.TCSI
			tc.TBcsmCamelTDPDataList.Values = ignored(tc.TBcsmCamelTDPDataList.Values, tdp)
			wantEqual(t, "T-CSI", kept, parseSriRespGmsc(t, w).TCSI.TBcsmCamelTDPDataList)
		})
	}
	t.Run("VT-CSI, all ignored", func(t *testing.T) {
		w := isdWire(t, camelISD())
		w.VlrCamelSubscriptionInfo.VtCSI.TBcsmCamelTDPDataList.Values[0].TBcsmTriggerDetectionPoint = 15
		got := parseISD(t, w).VlrCamelSubscriptionInfo
		if got.VtCSI != nil {
			t.Errorf("VT-CSI = %+v, want absent", got.VtCSI)
		}
		if got.OCSI == nil || got.DCSI == nil {
			t.Errorf("sibling CSIs lost: %+v", got)
		}
	})
	t.Run("GMSC T-CSI, all ignored", func(t *testing.T) {
		w := sriRespWire(t, gmscCamel())
		w.ExtendedRoutingInfo.CamelRoutingInfo.GmscCamelSubscriptionInfo.TCSI.TBcsmCamelTDPDataList.Values[0].TBcsmTriggerDetectionPoint = 15
		got := parseSriRespGmsc(t, w)
		if got.TCSI != nil {
			t.Errorf("T-CSI = %+v, want absent", got.TCSI)
		}
		if got.OCSI == nil || got.DCSI == nil {
			t.Errorf("sibling CSIs lost: %+v", got)
		}
	})
}

// --- GPRS-TriggerDetectionPoint, §17.7.1 ---

func gprsTDPData(tdp GPRSTriggerDetectionPoint, serviceKey int64) GPRSCamelTDPData {
	return GPRSCamelTDPData{
		GprsTriggerDetectionPoint: tdp,
		ServiceKey:                serviceKey,
		GsmSCFAddress:             testGsmSCF,
		GsmSCFAddressNature:       address.NatureInternational,
		GsmSCFAddressPlan:         address.PlanISDN,
		DefaultSessionHandling:    DefaultGPRSContinueTransaction,
	}
}

// "For GPRS-CamelTDPData sequences containing this parameter with any other
// value than the ones listed the receiver shall ignore the whole
// GPRS-CamelTDPData sequence." The listed values are attach (1),
// attachChangeOfPosition (2), pdp-ContextEstablishment (11),
// pdp-ContextEstablishmentAcknowledgement (12) and
// pdp-ContextChangeOfPosition (14).
func TestParseGPRSTriggerDetectionPointIgnoresSequence(t *testing.T) {
	kept := GPRSCamelTDPDataList{
		gprsTDPData(GPRSTDPAttach, 30),
		gprsTDPData(GPRSTDPPdpContextChangeOfPosition, 32),
	}
	for _, tdp := range []GPRSTriggerDetectionPoint{0, 3, 10, 13, 15, -1, math.MaxInt64} {
		t.Run(fmt.Sprintf("wire %d", tdp), func(t *testing.T) {
			in := camelISD()
			in.SgsnCAMELSubscriptionInfo.GprsCSI.GprsCamelTDPDataList = kept
			w := isdWire(t, in)
			l := w.SgsnCAMELSubscriptionInfo.GprsCSI.GprsCamelTDPDataList
			// The ignored entry carries a DefaultSessionHandling the decoder
			// would reject, which proves that it skips the whole sequence.
			e := l.Values[0]
			e.GprsTriggerDetectionPoint, e.ServiceKey, e.DefaultSessionHandling = tdp, 31, -1
			l.Values = slices.Insert(l.Values, 1, e)
			got := parseISD(t, w).SgsnCAMELSubscriptionInfo.GprsCSI
			if got == nil {
				t.Fatal("GPRS-CSI absent, want the listed entries")
			}
			wantEqual(t, "GPRS-CamelTDPDataList", kept, got.GprsCamelTDPDataList)
		})
	}
	// With every entry ignored the GPRS-CSI is absent; the other SGSN CSIs
	// survive.
	t.Run("all ignored", func(t *testing.T) {
		w := isdWire(t, camelISD())
		w.SgsnCAMELSubscriptionInfo.GprsCSI.GprsCamelTDPDataList.Values[0].GprsTriggerDetectionPoint = 13
		got := parseISD(t, w).SgsnCAMELSubscriptionInfo
		if got.GprsCSI != nil {
			t.Errorf("GPRS-CSI = %+v, want absent", got.GprsCSI)
		}
		if got.MoSmsCSI == nil || got.MtSmsCSI == nil {
			t.Errorf("sibling CSIs lost: %+v", got)
		}
	})
}

func TestMarshalRejectsUnlistedGPRSTriggerDetectionPoint(t *testing.T) {
	for _, tdp := range []GPRSTriggerDetectionPoint{0, 3, 13, 15, -1} {
		a := camelISD()
		a.SgsnCAMELSubscriptionInfo.GprsCSI.GprsCamelTDPDataList[0].GprsTriggerDetectionPoint = tdp
		if _, err := a.Marshal(); !errors.Is(err, ErrGPRSTriggerDetectionPointInvalid) {
			t.Errorf("GprsTriggerDetectionPoint %d: got %v, want ErrGPRSTriggerDetectionPointInvalid", tdp, err)
		}
	}
}

// --- SMS-TriggerDetectionPoint, §17.7.1 ---

// mo-sms-CSI keeps only sms-CollectedInfo (1); mt-sms-CSI and
// MT-smsCAMELTDP-Criteria keep only sms-DeliveryRequest (2). The receiver
// ignores the whole sequence carrying any other value.
func TestParseSMSTriggerDetectionPointIgnoresSequence(t *testing.T) {
	collected, delivery := SMSTriggerDetectionPointSmsCollectedInfo, SMSTriggerDetectionPointSmsDeliveryRequest
	paths := []struct {
		name        string
		csi         func(w *gsm_map.InsertSubscriberDataArg) *gsm_map.SMSCSI
		got         func(a *InsertSubscriberDataArg) *SMSCSI
		keep, wrong SMSTriggerDetectionPoint
	}{
		{"VLR MO-SMS-CSI",
			func(w *gsm_map.InsertSubscriberDataArg) *gsm_map.SMSCSI { return w.VlrCamelSubscriptionInfo.MoSmsCSI },
			func(a *InsertSubscriberDataArg) *SMSCSI { return a.VlrCamelSubscriptionInfo.MoSmsCSI },
			collected, delivery},
		{"VLR MT-SMS-CSI",
			func(w *gsm_map.InsertSubscriberDataArg) *gsm_map.SMSCSI { return w.VlrCamelSubscriptionInfo.MtSmsCSI },
			func(a *InsertSubscriberDataArg) *SMSCSI { return a.VlrCamelSubscriptionInfo.MtSmsCSI },
			delivery, collected},
		{"SGSN MO-SMS-CSI",
			func(w *gsm_map.InsertSubscriberDataArg) *gsm_map.SMSCSI { return w.SgsnCAMELSubscriptionInfo.MoSmsCSI },
			func(a *InsertSubscriberDataArg) *SMSCSI { return a.SgsnCAMELSubscriptionInfo.MoSmsCSI },
			collected, delivery},
		{"SGSN MT-SMS-CSI",
			func(w *gsm_map.InsertSubscriberDataArg) *gsm_map.SMSCSI { return w.SgsnCAMELSubscriptionInfo.MtSmsCSI },
			func(a *InsertSubscriberDataArg) *SMSCSI { return a.SgsnCAMELSubscriptionInfo.MtSmsCSI },
			delivery, collected},
	}
	for _, p := range paths {
		// The other listed TDP, then unlisted values around and beyond 1..2.
		for _, tdp := range []SMSTriggerDetectionPoint{p.wrong, 0, 3, -1, math.MaxInt64} {
			t.Run(fmt.Sprintf("%s, wire %d", p.name, tdp), func(t *testing.T) {
				w := isdWire(t, camelISD())
				csi := p.csi(w)
				// The ignored entry carries a DefaultSMSHandling the decoder
				// would reject, which proves that it skips the whole sequence.
				e := csi.SmsCAMELTDPDataList.Values[0]
				e.SmsTriggerDetectionPoint, e.ServiceKey, e.DefaultSMSHandling = tdp, 99, -1
				csi.SmsCAMELTDPDataList.Values = slices.Insert(csi.SmsCAMELTDPDataList.Values, 0, e)
				got := p.got(parseISD(t, w))
				if got == nil || len(got.SmsCAMELTDPDataList) != 1 || got.SmsCAMELTDPDataList[0].SmsTriggerDetectionPoint != p.keep {
					t.Errorf("got %+v, want only the %d entry", got, p.keep)
				}
			})
		}
		t.Run(p.name+", all ignored", func(t *testing.T) {
			w := isdWire(t, camelISD())
			p.csi(w).SmsCAMELTDPDataList.Values[0].SmsTriggerDetectionPoint = p.wrong
			got := parseISD(t, w)
			if csi := p.got(got); csi != nil {
				t.Errorf("SMS-CSI = %+v, want absent", csi)
			}
			if got.VlrCamelSubscriptionInfo.OCSI == nil || got.SgsnCAMELSubscriptionInfo.GprsCSI == nil {
				t.Errorf("sibling CSIs lost: %+v", got)
			}
		})
	}

	criteria := map[string]struct {
		list func(w *gsm_map.InsertSubscriberDataArg) *gsm_map.MTSmsCAMELTDPCriteriaList
		got  func(a *InsertSubscriberDataArg) []MTSmsCAMELTDPCriteria
	}{
		"VLR MT-smsCAMELTDP-Criteria": {
			func(w *gsm_map.InsertSubscriberDataArg) *gsm_map.MTSmsCAMELTDPCriteriaList {
				return w.VlrCamelSubscriptionInfo.MtSmsCAMELTDPCriteriaList
			},
			func(a *InsertSubscriberDataArg) []MTSmsCAMELTDPCriteria {
				return a.VlrCamelSubscriptionInfo.MtSmsCAMELTDPCriteriaList
			},
		},
		"SGSN MT-smsCAMELTDP-Criteria": {
			func(w *gsm_map.InsertSubscriberDataArg) *gsm_map.MTSmsCAMELTDPCriteriaList {
				return w.SgsnCAMELSubscriptionInfo.MtSmsCAMELTDPCriteriaList
			},
			func(a *InsertSubscriberDataArg) []MTSmsCAMELTDPCriteria {
				return a.SgsnCAMELSubscriptionInfo.MtSmsCAMELTDPCriteriaList
			},
		},
	}
	for name, c := range criteria {
		for _, tdp := range []SMSTriggerDetectionPoint{collected, 0, 3, -1, math.MaxInt64} {
			t.Run(fmt.Sprintf("%s, wire %d", name, tdp), func(t *testing.T) {
				w := isdWire(t, camelISD())
				l := c.list(w)
				l.Values = append(l.Values, gsm_map.MTSmsCAMELTDPCriteria{
					SmsTriggerDetectionPoint: tdp,
					TpduTypeCriterion:        &gsm_map.TPDUTypeCriterion{Values: []gsm_map.MTSMSTPDUType{99}},
				})
				wantEqual(t, name, mtSMSCriteria(), c.got(parseISD(t, w)))
			})
		}
		t.Run(name+", all ignored", func(t *testing.T) {
			w := isdWire(t, camelISD())
			c.list(w).Values[0].SmsTriggerDetectionPoint = collected
			got := parseISD(t, w)
			if l := c.got(got); l != nil {
				t.Errorf("criteria = %+v, want absent", l)
			}
			if got.VlrCamelSubscriptionInfo.MtSmsCSI == nil || got.SgsnCAMELSubscriptionInfo.MtSmsCSI == nil {
				t.Errorf("sibling CSIs lost: %+v", got)
			}
		})
	}
}

// A sender puts sms-CollectedInfo only in an MO-SMS-CSI and
// sms-DeliveryRequest only in an MT-SMS-CSI or MT-smsCAMELTDP-Criteria.
func TestMarshalRejectsMisplacedSMSTriggerDetectionPoint(t *testing.T) {
	// smsTDPFields returns the TDP of the first entry of every SMS-CSI and
	// MT-smsCAMELTDP-Criteria list in a camelISD, which holds the only
	// value the encoder accepts there.
	smsTDPFields := func(a *InsertSubscriberDataArg) map[string]*SMSTriggerDetectionPoint {
		v, s := a.VlrCamelSubscriptionInfo, a.SgsnCAMELSubscriptionInfo
		return map[string]*SMSTriggerDetectionPoint{
			"VLR MO-SMS-CSI":               &v.MoSmsCSI.SmsCAMELTDPDataList[0].SmsTriggerDetectionPoint,
			"VLR MT-SMS-CSI":               &v.MtSmsCSI.SmsCAMELTDPDataList[0].SmsTriggerDetectionPoint,
			"VLR MT-smsCAMELTDP-Criteria":  &v.MtSmsCAMELTDPCriteriaList[0].SmsTriggerDetectionPoint,
			"SGSN MO-SMS-CSI":              &s.MoSmsCSI.SmsCAMELTDPDataList[0].SmsTriggerDetectionPoint,
			"SGSN MT-SMS-CSI":              &s.MtSmsCSI.SmsCAMELTDPDataList[0].SmsTriggerDetectionPoint,
			"SGSN MT-smsCAMELTDP-Criteria": &s.MtSmsCAMELTDPCriteriaList[0].SmsTriggerDetectionPoint,
		}
	}
	for name := range smsTDPFields(camelISD()) {
		for _, tdp := range []SMSTriggerDetectionPoint{
			0, SMSTriggerDetectionPointSmsCollectedInfo, SMSTriggerDetectionPointSmsDeliveryRequest, 3,
		} {
			a := camelISD()
			f := smsTDPFields(a)[name]
			if tdp == *f {
				continue // the permitted value
			}
			*f = tdp
			if _, err := a.Marshal(); !errors.Is(err, ErrCamelInvalidSMSTriggerDetectionPoint) {
				t.Errorf("%s = %d: got %v, want ErrCamelInvalidSMSTriggerDetectionPoint", name, tdp, err)
			}
		}
	}
}

// --- MT-SMS-TPDU-Type, §17.7.1 ---

// "For TPDU-TypeCriterion sequences containing this parameter with any
// other value than the ones listed above the receiver shall ignore the
// whole TPDU-TypeCriterion sequence." The listed values are sms-DELIVER (0),
// sms-SUBMIT-REPORT (1) and sms-STATUS-REPORT (2).
func TestParseMTSMSTPDUTypeIgnoresCriterion(t *testing.T) {
	listed := []MTSMSTPDUType{MTSMSTPDUTypeSmsDELIVER, MTSMSTPDUTypeSmsSTATUSREPORT}
	for _, tpdu := range []MTSMSTPDUType{3, -1, math.MaxInt64} {
		t.Run(fmt.Sprintf("wire %d", tpdu), func(t *testing.T) {
			in := camelISD()
			in.VlrCamelSubscriptionInfo.MtSmsCAMELTDPCriteriaList = []MTSmsCAMELTDPCriteria{
				{SmsTriggerDetectionPoint: mtSMSTriggerDetectionPoint, TpduTypeCriterion: listed},
				{SmsTriggerDetectionPoint: mtSMSTriggerDetectionPoint, TpduTypeCriterion: listed},
				{SmsTriggerDetectionPoint: mtSMSTriggerDetectionPoint, TpduTypeCriterion: listed},
			}
			in.SgsnCAMELSubscriptionInfo.MtSmsCAMELTDPCriteriaList = in.VlrCamelSubscriptionInfo.MtSmsCAMELTDPCriteriaList
			w := isdWire(t, in)
			for _, l := range []*gsm_map.MTSmsCAMELTDPCriteriaList{
				w.VlrCamelSubscriptionInfo.MtSmsCAMELTDPCriteriaList,
				w.SgsnCAMELSubscriptionInfo.MtSmsCAMELTDPCriteriaList,
			} {
				// One unlisted type among listed ones, and only unlisted ones.
				l.Values[1].TpduTypeCriterion = &gsm_map.TPDUTypeCriterion{Values: []gsm_map.MTSMSTPDUType{MTSMSTPDUTypeSmsDELIVER, tpdu}}
				l.Values[2].TpduTypeCriterion = &gsm_map.TPDUTypeCriterion{Values: []gsm_map.MTSMSTPDUType{tpdu}}
			}
			want := []MTSmsCAMELTDPCriteria{
				{SmsTriggerDetectionPoint: mtSMSTriggerDetectionPoint, TpduTypeCriterion: listed},
				{SmsTriggerDetectionPoint: mtSMSTriggerDetectionPoint},
				{SmsTriggerDetectionPoint: mtSMSTriggerDetectionPoint},
			}
			got := parseISD(t, w)
			wantEqual(t, "VLR", want, got.VlrCamelSubscriptionInfo.MtSmsCAMELTDPCriteriaList)
			wantEqual(t, "SGSN", want, got.SgsnCAMELSubscriptionInfo.MtSmsCAMELTDPCriteriaList)
		})
	}
}

func TestMarshalRejectsUnlistedMTSMSTPDUType(t *testing.T) {
	for _, tpdu := range []MTSMSTPDUType{3, -1} {
		a := camelISD()
		a.VlrCamelSubscriptionInfo.MtSmsCAMELTDPCriteriaList[0].TpduTypeCriterion = []MTSMSTPDUType{tpdu}
		if _, err := a.Marshal(); !errors.Is(err, ErrCamelInvalidMTSMSTPDUType) {
			t.Errorf("wire %d: got %v, want ErrCamelInvalidMTSMSTPDUType", tpdu, err)
		}
	}
}

// --- GMLC-Restriction and NotificationToMSUser, §17.7.1 ---

const (
	ssCodeCallSessionUnrelated SsCode = 0xB3 // callSessionUnrelated, §17.7.5
	ssCodeServiceType          SsCode = 0xB5 // serviceType, §17.7.5
)

// lcsISD returns an InsertSubscriberDataArg whose LCS privacy classes carry
// every GMLC-Restriction and NotificationToMSUser container.
func lcsISD() *InsertSubscriberDataArg {
	gmlc, notify := GMLCRestrictionHomeCountry, NotifyAndVerifyLocationAllowedIfNoResponse
	return &InsertSubscriberDataArg{
		LcsInformation: &LCSInformation{
			LcsPrivacyExceptionList: LCSPrivacyExceptionList{
				{
					SsCode:               ssCodeCallSessionUnrelated,
					SsStatus:             HexBytes{0x05},
					NotificationToMSUser: &notify,
					ExternalClientList: ExternalClientList{{
						ClientIdentity:       makeLCSClientExternalID(),
						GmlcRestriction:      &gmlc,
						NotificationToMSUser: &notify,
					}},
				},
				{
					SsCode:   ssCodeServiceType,
					SsStatus: HexBytes{0x05},
					ServiceTypeList: ServiceTypeList{{
						ServiceTypeIdentity:  7,
						GmlcRestriction:      &gmlc,
						NotificationToMSUser: &notify,
					}},
				},
			},
		},
	}
}

// "At reception of any other value than the ones listed the receiver shall
// ignore GMLC-Restriction." (listed: gmlc-List (0), home-Country (1)) and
// "At reception of any other value than the ones listed the receiver shall
// ignore NotificationToMSUser." (listed: 0 to 3, locationNotAllowed (3)
// after the extension marker).
func TestParseGMLCRestrictionAndNotificationToMSUserIgnored(t *testing.T) {
	type lcsWire struct {
		class   *gsm_map.LCSPrivacyClass
		client  *gsm_map.ExternalClient
		service *gsm_map.ServiceType
	}
	wireOf := func(w *gsm_map.InsertSubscriberDataArg) lcsWire {
		l := w.LcsInformation.LcsPrivacyExceptionList.Values
		return lcsWire{&l[0], &l[0].ExternalClientList.Values[0], &l[1].ServiceTypeList.Values[0]}
	}
	// Each case compares the whole decoded LcsInformation with lcsISD's, so
	// the siblings of an ignored parameter are checked too.
	for _, tc := range []struct {
		wire GMLCRestriction
		keep bool
	}{{0, true}, {1, true}, {2, false}, {-1, false}, {math.MaxInt64, false}} {
		t.Run(fmt.Sprintf("GMLC-Restriction wire %d", tc.wire), func(t *testing.T) {
			w := isdWire(t, lcsISD())
			lw := wireOf(w)
			v := tc.wire
			lw.client.GmlcRestriction, lw.service.GmlcRestriction = &v, &v
			want := lcsISD()
			l := want.LcsInformation.LcsPrivacyExceptionList
			l[0].ExternalClientList[0].GmlcRestriction, l[1].ServiceTypeList[0].GmlcRestriction = nil, nil
			if tc.keep {
				l[0].ExternalClientList[0].GmlcRestriction, l[1].ServiceTypeList[0].GmlcRestriction = &v, &v
			}
			wantEqual(t, "LcsInformation", want.LcsInformation, parseISD(t, w).LcsInformation)
		})
	}
	for _, tc := range []struct {
		wire NotificationToMSUser
		keep bool
	}{{0, true}, {3, true}, {4, false}, {-1, false}, {math.MaxInt64, false}} {
		t.Run(fmt.Sprintf("NotificationToMSUser wire %d", tc.wire), func(t *testing.T) {
			w := isdWire(t, lcsISD())
			lw := wireOf(w)
			v := tc.wire
			lw.class.NotificationToMSUser, lw.client.NotificationToMSUser, lw.service.NotificationToMSUser = &v, &v, &v
			want := lcsISD()
			l := want.LcsInformation.LcsPrivacyExceptionList
			var p *NotificationToMSUser
			if tc.keep {
				p = &v
			}
			l[0].NotificationToMSUser, l[0].ExternalClientList[0].NotificationToMSUser, l[1].ServiceTypeList[0].NotificationToMSUser = p, p, p
			wantEqual(t, "LcsInformation", want.LcsInformation, parseISD(t, w).LcsInformation)
		})
	}
	t.Run("both ignored", func(t *testing.T) {
		w := isdWire(t, lcsISD())
		lw := wireOf(w)
		g, n := GMLCRestriction(2), NotificationToMSUser(4)
		lw.class.NotificationToMSUser = &n
		lw.client.GmlcRestriction, lw.client.NotificationToMSUser = &g, &n
		lw.service.GmlcRestriction, lw.service.NotificationToMSUser = &g, &n
		want := lcsISD()
		l := want.LcsInformation.LcsPrivacyExceptionList
		l[0].NotificationToMSUser = nil
		l[0].ExternalClientList[0].GmlcRestriction, l[0].ExternalClientList[0].NotificationToMSUser = nil, nil
		l[1].ServiceTypeList[0].GmlcRestriction, l[1].ServiceTypeList[0].NotificationToMSUser = nil, nil
		wantEqual(t, "LcsInformation", want.LcsInformation, parseISD(t, w).LcsInformation)
	})
}

func TestMarshalRejectsUnlistedGMLCRestrictionAndNotificationToMSUser(t *testing.T) {
	for _, g := range []GMLCRestriction{2, -1} {
		for name, set := range map[string]func(a *InsertSubscriberDataArg){
			"ExternalClient": func(a *InsertSubscriberDataArg) {
				a.LcsInformation.LcsPrivacyExceptionList[0].ExternalClientList[0].GmlcRestriction = &g
			},
			"ServiceType": func(a *InsertSubscriberDataArg) {
				a.LcsInformation.LcsPrivacyExceptionList[1].ServiceTypeList[0].GmlcRestriction = &g
			},
		} {
			a := lcsISD()
			set(a)
			if _, err := a.Marshal(); !errors.Is(err, ErrGMLCRestrictionInvalid) {
				t.Errorf("%s GmlcRestriction %d: got %v, want ErrGMLCRestrictionInvalid", name, g, err)
			}
		}
	}
	for _, n := range []NotificationToMSUser{4, -1} {
		for name, set := range map[string]func(a *InsertSubscriberDataArg){
			"LCSPrivacyClass": func(a *InsertSubscriberDataArg) {
				a.LcsInformation.LcsPrivacyExceptionList[0].NotificationToMSUser = &n
			},
			"ExternalClient": func(a *InsertSubscriberDataArg) {
				a.LcsInformation.LcsPrivacyExceptionList[0].ExternalClientList[0].NotificationToMSUser = &n
			},
			"ServiceType": func(a *InsertSubscriberDataArg) {
				a.LcsInformation.LcsPrivacyExceptionList[1].ServiceTypeList[0].NotificationToMSUser = &n
			},
		} {
			a := lcsISD()
			set(a)
			if _, err := a.Marshal(); !errors.Is(err, ErrNotificationToMSUserInvalid) {
				t.Errorf("%s NotificationToMSUser %d: got %v, want ErrNotificationToMSUserInvalid", name, n, err)
			}
		}
	}
}

// --- SupportedCCBS-Phase, §17.7.3 ---

// "If received values 2-127 shall be mapped on to value 1."
func TestParseSupportedCCBSPhaseMapsReservedToOne(t *testing.T) {
	for _, v := range []gsm_map.SupportedCCBSPhase{1, 2, 127} {
		w := newSriArg()
		w.SupportedCCBSPhase = &v
		got, err := ParseSri(strictBER(t, w))
		if err != nil {
			t.Fatalf("wire %d: ParseSri: %v", v, err)
		}
		if got.SupportedCCBSPhase == nil || *got.SupportedCCBSPhase != 1 {
			t.Errorf("wire %d: SupportedCCBSPhase = %v, want 1", v, got.SupportedCCBSPhase)
		}
	}
	for _, v := range []gsm_map.SupportedCCBSPhase{0, 128} {
		w := newSriArg()
		w.SupportedCCBSPhase = &v
		_, err := ParseSri(tolerantBER(t, w))
		wantConstraintError(t, err, "supportedCCBS-Phase", "(1..127)")
	}
}

// "Only value 1 is used. Values in the ranges 2-127 are reserved for future
// use."
func TestMarshalRejectsReservedSupportedCCBSPhase(t *testing.T) {
	s, err := convertArgToSri(newSriArg())
	if err != nil {
		t.Fatalf("convertArgToSri: %v", err)
	}
	for _, v := range []int{0, 2, 127} {
		s.SupportedCCBSPhase = &v
		if _, err := s.Marshal(); !errors.Is(err, ErrSriInvalidSupportedCCBSPhase) {
			t.Errorf("SupportedCCBSPhase %d: got %v, want ErrSriInvalidSupportedCCBSPhase", v, err)
		}
	}
	one := 1
	s.SupportedCCBSPhase = &one
	if _, err := s.Marshal(); err != nil {
		t.Errorf("SupportedCCBSPhase 1: %v", err)
	}
}

// --- SM-RP-MTI, §17.7.6 ---

func sriSmWire(t *testing.T) *gsm_map.RoutingInfoForSMArg {
	t.Helper()
	w, err := convertSriSmToArg(&SriSm{MSISDN: "31612345678", SmRpPri: true, ServiceCentreAddress: "31201111111"})
	if err != nil {
		t.Fatalf("convertSriSmToArg: %v", err)
	}
	return w
}

// "0 SMS Deliver; 1 SMS Status Report; other values are reserved for future
// use and shall be discarded if received".
func TestParseSmRpMtiDiscardsReserved(t *testing.T) {
	for _, tc := range []struct {
		wire gsm_map.SMRPMTI
		want *int
	}{
		{0, intPtr(0)},
		{1, intPtr(1)}, // last defined value
		{2, nil},       // first reserved value
		{10, nil},      // last reserved value, INTEGER (0..10)
	} {
		w := sriSmWire(t)
		v := tc.wire
		w.SmRPMTI = &v
		got, err := ParseSriSm(strictBER(t, w))
		if err != nil {
			t.Fatalf("wire %d: ParseSriSm: %v", tc.wire, err)
		}
		wantEqual(t, fmt.Sprintf("wire %d SmRpMti", tc.wire), tc.want, got.SmRpMti)
	}
	for _, v := range []gsm_map.SMRPMTI{-1, 11} {
		w := sriSmWire(t)
		w.SmRPMTI = &v
		_, err := ParseSriSm(tolerantBER(t, w))
		wantConstraintError(t, err, "sm-RP-MTI", "(0..10)")
	}
}

func TestMarshalRejectsReservedSmRpMti(t *testing.T) {
	for _, v := range []int{-1, 2, 10} {
		s := &SriSm{MSISDN: "31612345678", SmRpPri: true, ServiceCentreAddress: "31201111111", SmRpMti: &v}
		if _, err := s.Marshal(); !errors.Is(err, ErrSriSmInvalidSmRpMti) {
			t.Errorf("SmRpMti %d: got %v, want ErrSriSmInvalidSmRpMti", v, err)
		}
	}
}

// --- LCS, §17.7.13 ---

func slrWithTerminationCause(c TerminationCause) *SubscriberLocationReportArg {
	a := minimalSLRArg()
	a.DeferredmtLrData = &DeferredmtLrData{
		DeferredLocationEventType: DeferredLocationEventType{MsAvailable: true},
		TerminationCause:          &c,
	}
	return a
}

// "an unrecognized value shall be treated the same as value 1
// (errorundefined)".
func TestParseTerminationCauseUnrecognized(t *testing.T) {
	for _, tc := range []struct {
		wire, want TerminationCause
	}{
		{TerminationNormal, TerminationNormal},
		{TerminationNetworkTermination, TerminationNetworkTermination}, // last listed value, 9
		{10, TerminationErrorundefined},
		{math.MaxInt64, TerminationErrorundefined},
		{-1, TerminationErrorundefined},
		{math.MinInt64, TerminationErrorundefined},
	} {
		w, err := convertSubscriberLocationReportArgToWire(slrWithTerminationCause(TerminationNormal))
		if err != nil {
			t.Fatalf("convertSubscriberLocationReportArgToWire: %v", err)
		}
		v := tc.wire
		w.DeferredmtLrData.TerminationCause = &v
		got, err := ParseSubscriberLocationReport(strictBER(t, w))
		if err != nil {
			t.Fatalf("wire %d: ParseSubscriberLocationReport: %v", tc.wire, err)
		}
		if c := got.DeferredmtLrData.TerminationCause; c == nil || *c != tc.want {
			t.Errorf("wire %d: TerminationCause = %v, want %v", tc.wire, c, tc.want)
		}
	}
}

func TestMarshalRejectsUnrecognizedTerminationCause(t *testing.T) {
	for _, c := range []TerminationCause{-1, 10} {
		if _, err := slrWithTerminationCause(c).Marshal(); !errors.Is(err, ErrTerminationCauseInvalid) {
			t.Errorf("TerminationCause %d: got %v, want ErrTerminationCauseInvalid", c, err)
		}
	}
}

func pslArg() *ProvideSubscriberLocationArg {
	return &ProvideSubscriberLocationArg{
		LocationType:    LocationType{LocationEstimateType: LocationEstimateCurrentLocation},
		MlcNumber:       "31612345678",
		MlcNumberNature: address.NatureInternational,
		MlcNumberPlan:   address.PlanISDN,
		LcsQoS:          &LCSQoS{ResponseTime: &ResponseTime{ResponseTimeCategory: ResponseTimeLowdelay}},
		LcsPrivacyCheck: &LCSPrivacyCheck{CallSessionUnrelated: PrivacyCheckAllowedWithoutNotification},
	}
}

func pslWire(t *testing.T, a *ProvideSubscriberLocationArg) *gsm_map.ProvideSubscriberLocationArg {
	t.Helper()
	w, err := convertProvideSubscriberLocationArgToWire(a)
	if err != nil {
		t.Fatalf("convertProvideSubscriberLocationArgToWire: %v", err)
	}
	return w
}

// "a ProvideSubscriberLocation-Arg containing an unrecognized
// PrivacyCheckRelatedAction shall be rejected by the receiver with a return
// error cause of unexpected data value". The return error is the
// application's, so Parse hands it the value.
func TestParsePreservesUnrecognizedPrivacyCheckRelatedAction(t *testing.T) {
	for _, v := range []PrivacyCheckRelatedAction{
		PrivacyCheckNotAllowed, // last listed value, 4
		5, -1, math.MaxInt64,
	} {
		w := pslWire(t, pslArg())
		related := v
		w.LcsPrivacyCheck.CallSessionUnrelated, w.LcsPrivacyCheck.CallSessionRelated = v, &related
		got, err := ParseProvideSubscriberLocation(strictBER(t, w))
		if err != nil {
			t.Fatalf("wire %d: ParseProvideSubscriberLocation: %v", v, err)
		}
		wantEqual(t, fmt.Sprintf("wire %d", v), &LCSPrivacyCheck{CallSessionUnrelated: v, CallSessionRelated: &related}, got.LcsPrivacyCheck)
	}
}

func TestMarshalRejectsUnrecognizedPrivacyCheckRelatedAction(t *testing.T) {
	for _, v := range []PrivacyCheckRelatedAction{-1, 5} {
		a := pslArg()
		a.LcsPrivacyCheck.CallSessionUnrelated = v
		if _, err := a.Marshal(); !errors.Is(err, ErrPrivacyCheckRelatedActionInvalid) {
			t.Errorf("CallSessionUnrelated %d: got %v, want ErrPrivacyCheckRelatedActionInvalid", v, err)
		}
		a = pslArg()
		a.LcsPrivacyCheck.CallSessionRelated = &v
		if _, err := a.Marshal(); !errors.Is(err, ErrPrivacyCheckRelatedActionInvalid) {
			t.Errorf("CallSessionRelated %d: got %v, want ErrPrivacyCheckRelatedActionInvalid", v, err)
		}
	}
}

// "an unrecognized value shall be treated the same as value 0
// (bestEffort)". LCSQoS does not surface lcs-qos-class, which "may only be
// included in MO-LR request sent by the UE to the network" and so never in a
// ProvideSubscriberLocation-Arg; Parse decodes every class, recognized or
// not, to the same LCSQoS.
func TestParseLCSQoSClassAnyValue(t *testing.T) {
	want := pslArg().LcsQoS
	for _, v := range []gsm_map.LCSQoSClass{gsm_map.LCSQoSClassBestEffort, gsm_map.LCSQoSClassAssured, 2, -1, math.MaxInt64} {
		w := pslWire(t, pslArg())
		class := v
		w.LcsQoS.LcsQosClass = &class
		got, err := ParseProvideSubscriberLocation(strictBER(t, w))
		if err != nil {
			t.Fatalf("lcs-qos-class %d: ParseProvideSubscriberLocation: %v", v, err)
		}
		wantEqual(t, fmt.Sprintf("lcs-qos-class %d", v), want, got.LcsQoS)
	}
}

// --- DomainType, §17.7.1 ---

// "reception of values > 1 shall be mapped to 'cs-Domain'".
func TestParseRequestedDomainMapsAbovePsToCs(t *testing.T) {
	ps := PsDomain
	ri := RequestedInfo{LocationInformation: true, RequestedDomain: &ps}
	for _, tc := range []struct {
		wire, want DomainType
	}{
		{CsDomain, CsDomain},
		{PsDomain, PsDomain}, // last listed value
		{2, CsDomain},        // first mapped value
		{math.MaxInt64, CsDomain},
	} {
		t.Run(fmt.Sprintf("ATI, wire %d", tc.wire), func(t *testing.T) {
			w, err := convertATIToArg(&AnyTimeInterrogation{
				SubscriberIdentity: SubscriberIdentity{IMSI: "204080012345678"},
				RequestedInfo:      ri,
				GsmSCFAddress:      testGsmSCF,
			})
			if err != nil {
				t.Fatalf("convertATIToArg: %v", err)
			}
			v := tc.wire
			w.RequestedInfo.RequestedDomain = &v
			got, err := ParseAnyTimeInterrogation(strictBER(t, w))
			if err != nil {
				t.Fatalf("ParseAnyTimeInterrogation: %v", err)
			}
			wantEqual(t, "RequestedDomain", &tc.want, got.RequestedInfo.RequestedDomain)
		})
		t.Run(fmt.Sprintf("PSI, wire %d", tc.wire), func(t *testing.T) {
			w, err := convertProvideSubscriberInfoToArg(&ProvideSubscriberInfo{IMSI: "204080012345678", RequestedInfo: ri})
			if err != nil {
				t.Fatalf("convertProvideSubscriberInfoToArg: %v", err)
			}
			v := tc.wire
			w.RequestedInfo.RequestedDomain = &v
			got, err := ParseProvideSubscriberInfo(strictBER(t, w))
			if err != nil {
				t.Fatalf("ParseProvideSubscriberInfo: %v", err)
			}
			wantEqual(t, "RequestedDomain", &tc.want, got.RequestedInfo.RequestedDomain)
		})
	}
}

func TestMarshalRejectsUnlistedRequestedDomain(t *testing.T) {
	for _, d := range []DomainType{-1, 2} {
		ri := RequestedInfo{LocationInformation: true, RequestedDomain: &d}
		ati := &AnyTimeInterrogation{
			SubscriberIdentity: SubscriberIdentity{IMSI: "204080012345678"},
			RequestedInfo:      ri,
			GsmSCFAddress:      testGsmSCF,
		}
		if _, err := ati.Marshal(); !errors.Is(err, ErrRequestedDomainInvalid) {
			t.Errorf("ATI RequestedDomain %d: got %v, want ErrRequestedDomainInvalid", d, err)
		}
		psi := &ProvideSubscriberInfo{IMSI: "204080012345678", RequestedInfo: ri}
		if _, err := psi.Marshal(); !errors.Is(err, ErrRequestedDomainInvalid) {
			t.Errorf("PSI RequestedDomain %d: got %v, want ErrRequestedDomainInvalid", d, err)
		}
	}
}

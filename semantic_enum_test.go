// semantic_enum_test.go
//
// Extensible ENUMERATED values on Parse. 3GPP TS 29.002 V19.1.0 §17.1.4:
// "An entity supporting a version greater than 1 shall not reject an
// unsupported extension following "..." of that SEQUENCE or ENUMERATED data
// type." Parse keeps a value the type does not list, or applies the
// receiver rule the type's exception handling gives; Marshal still sends
// only the listed values. Each case builds a valid message, overwrites the
// value on the wire, encodes it with go-asn1 and parses it.
package gsmmap

import (
	"encoding/hex"
	"errors"
	"math"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"

	"github.com/gomaja/go-asn1-gsmmap/address"
)

// unknownEnumValues are values outside every root and extension of the
// types below, on both sides of the listed ranges.
var semUnknownEnumValues = []int64{-1, math.MinInt32}

// CancellationType ::= ENUMERATED { updateProcedure(0),
// subscriptionWithdraw(1), ..., initialAttachProcedure (2)} (§17.7.1).
func TestParseCancellationTypeUnknownValue(t *testing.T) {
	// The CancelLocation-Arg of the report: CancellationType 3.
	data, _ := hex.DecodeString("a30804032143650a0103")
	got, err := ParseCancelLocation(data)
	if err != nil {
		t.Fatalf("ParseCancelLocation: %v", err)
	}
	if got.CancellationType == nil || *got.CancellationType != 3 {
		t.Fatalf("CancellationType = %v, want 3", got.CancellationType)
	}
	if _, err := got.Marshal(); !errors.Is(err, ErrCancelLocInvalidCancellationType) {
		t.Errorf("Marshal: err = %v, want ErrCancelLocInvalidCancellationType", err)
	}

	for _, v := range append(semUnknownEnumValues, math.MaxInt64) {
		ct := gsm_map.CancellationType(v)
		w := &gsm_map.CancelLocationArg{Identity: gsm_map.NewIdentityImsi(gsm_map.IMSI{0x21, 0x43, 0x65}), CancellationType: &ct}
		data, err := w.MarshalBER()
		if err != nil {
			t.Fatalf("MarshalBER: %v", err)
		}
		got, err := ParseCancelLocation(data)
		if err != nil {
			t.Errorf("CancellationType %d: ParseCancelLocation: %v", v, err)
			continue
		}
		if got.CancellationType == nil || int64(*got.CancellationType) != v {
			t.Errorf("CancellationType = %v, want %d", got.CancellationType, v)
		}
	}
}

// TypeOfUpdate ::= ENUMERATED { sgsn-change (0), mme-change (1), ...}
// (§17.7.1).
func TestParseTypeOfUpdateUnknownValue(t *testing.T) {
	for _, v := range append(semUnknownEnumValues, 2) {
		ct, tou := gsm_map.CancellationTypeUpdateProcedure, gsm_map.TypeOfUpdate(v)
		w := &gsm_map.CancelLocationArg{
			Identity:         gsm_map.NewIdentityImsi(gsm_map.IMSI{0x21, 0x43, 0x65}),
			CancellationType: &ct,
			TypeOfUpdate:     &tou,
		}
		data, err := w.MarshalBER()
		if err != nil {
			t.Fatalf("MarshalBER: %v", err)
		}
		got, err := ParseCancelLocation(data)
		if err != nil {
			t.Errorf("TypeOfUpdate %d: ParseCancelLocation: %v", v, err)
			continue
		}
		if got.TypeOfUpdate == nil || int64(*got.TypeOfUpdate) != v {
			t.Errorf("TypeOfUpdate = %v, want %d", got.TypeOfUpdate, v)
		}
		if _, err := got.Marshal(); !errors.Is(err, ErrCancelLocInvalidTypeOfUpdate) {
			t.Errorf("Marshal: err = %v, want ErrCancelLocInvalidTypeOfUpdate", err)
		}
	}
}

// T-BcsmTriggerDetectionPoint ::= ENUMERATED { termAttemptAuthorized (12),
// ..., tBusy (13), tNoAnswer (14)} (§17.7.1). Its exception handling covers
// T-BcsmCamelTDPData only; T-BCSM-CAMEL-TDP-Criteria keeps the value.
func TestParseTBcsmCriteriaUnknownTriggerDetectionPoint(t *testing.T) {
	criteria := []TBcsmCamelTDPCriteria{{TBcsmTriggerDetectionPoint: TBcsmTriggerTermAttemptAuthorized}}
	for _, v := range append(semUnknownEnumValues, 15) {
		tdp := TBcsmTriggerDetectionPoint(v)

		w, err := convertInsertSubscriberDataArgToWire(&InsertSubscriberDataArg{
			VlrCamelSubscriptionInfo: &VlrCamelSubscriptionInfo{TBcsmCamelTDPCriteriaList: criteria},
		})
		if err != nil {
			t.Fatalf("convertInsertSubscriberDataArgToWire: %v", err)
		}
		w.VlrCamelSubscriptionInfo.TBCSMCAMELTDPCriteriaList.Values[0].TBCSMTriggerDetectionPoint = tdp
		got, err := semParseISDWire(t, w)
		if err != nil {
			t.Errorf("VLR %d: ParseInsertSubscriberData: %v", v, err)
		} else if g := got.VlrCamelSubscriptionInfo.TBcsmCamelTDPCriteriaList; len(g) != 1 || g[0].TBcsmTriggerDetectionPoint != tdp {
			t.Errorf("VLR %d: criteria = %+v", v, g)
		}

		res, err := convertSriRespToRes(&SriResp{
			IMSI:                "204080012345678",
			ExtendedRoutingInfo: &ExtendedRoutingInfo{CamelRoutingInfo: &CamelRoutingInfo{GmscCamelSubscriptionInfo: GmscCamelSubscriptionInfo{TBcsmCamelTDPCriteriaList: criteria}}},
		})
		if err != nil {
			t.Fatalf("convertSriRespToRes: %v", err)
		}
		res.ExtendedRoutingInfo.CamelRoutingInfo.GmscCamelSubscriptionInfo.TBCSMCAMELTDPCriteriaList.Values[0].TBCSMTriggerDetectionPoint = tdp
		data, err := res.MarshalBER()
		if err != nil {
			t.Fatalf("MarshalBER: %v", err)
		}
		sri, err := ParseSriResp(data)
		if err != nil {
			t.Errorf("GMSC %d: ParseSriResp: %v", v, err)
		} else if g := sri.ExtendedRoutingInfo.CamelRoutingInfo.GmscCamelSubscriptionInfo.TBcsmCamelTDPCriteriaList; len(g) != 1 || g[0].TBcsmTriggerDetectionPoint != tdp {
			t.Errorf("GMSC %d: criteria = %+v", v, g)
		}
	}
}

// LCSClientInternalID ::= ENUMERATED { broadcastService (0),
// o-andM-HPLMN (1), o-andM-VPLMN (2), anonymousLocation (3),
// targetMSsubscribedService (4), ... } (§17.7.8).
func TestParseLCSClientInternalIDUnknownValue(t *testing.T) {
	for _, v := range append(semUnknownEnumValues, 5) {
		id := LCSClientInternalID(v)

		w, err := convertInsertSubscriberDataArgToWire(&InsertSubscriberDataArg{LcsInformation: &LCSInformation{
			LcsPrivacyExceptionList: LCSPrivacyExceptionList{{
				SsCode:         0xB4, // plmnoperator, TS 29.002 §17.7.5
				SsStatus:       HexBytes{0x05},
				PlmnClientList: PLMNClientList{LCSClientBroadcastService},
			}},
		}})
		if err != nil {
			t.Fatalf("convertInsertSubscriberDataArgToWire: %v", err)
		}
		w.LcsInformation.LcsPrivacyExceptionList.Values[0].PlmnClientList.Values[0] = id
		got, err := semParseISDWire(t, w)
		if err != nil {
			t.Errorf("PLMNClientList %d: ParseInsertSubscriberData: %v", v, err)
		} else if g := got.LcsInformation.LcsPrivacyExceptionList[0].PlmnClientList; len(g) != 1 || g[0] != id {
			t.Errorf("PLMNClientList %d: got %v", v, g)
		}

		a := semPSLArg()
		listed := LCSClientBroadcastService
		a.LcsClientID = &LCSClientID{LcsClientType: LCSClientTypeEmergencyServices, LcsClientInternalID: &listed}
		pw, err := convertProvideSubscriberLocationArgToWire(a)
		if err != nil {
			t.Fatalf("convertProvideSubscriberLocationArgToWire: %v", err)
		}
		pw.LcsClientID.LcsClientInternalID = &id
		data, err := pw.MarshalBER()
		if err != nil {
			t.Fatalf("MarshalBER: %v", err)
		}
		psl, err := ParseProvideSubscriberLocation(data)
		if err != nil {
			t.Errorf("LCS-ClientID %d: ParseProvideSubscriberLocation: %v", v, err)
		} else if g := psl.LcsClientID.LcsClientInternalID; g == nil || *g != id {
			t.Errorf("LCS-ClientID %d: got %v", v, g)
		}
		if err == nil {
			if _, err := psl.Marshal(); !errors.Is(err, ErrLCSClientInternalIDInvalid) {
				t.Errorf("Marshal: err = %v, want ErrLCSClientInternalIDInvalid", err)
			}
		}
	}
}

// UnavailabilityCause ::= ENUMERATED { bearerServiceNotProvisioned(1), ...,
// cug-Reject(6), ...} (§17.7.3): "Reception of other values than the ones
// listed shall result in the service being unavailable for that call", a
// rule for the application, which reads the value.
func TestParseUnavailabilityCauseUnknownValue(t *testing.T) {
	for _, v := range append(semUnknownEnumValues, 0, 7) {
		res, err := convertSriRespToRes(&SriResp{IMSI: "204080012345678"})
		if err != nil {
			t.Fatalf("convertSriRespToRes: %v", err)
		}
		uc := gsm_map.UnavailabilityCause(v)
		res.UnavailabilityCause = &uc
		data, err := res.MarshalBER()
		if err != nil {
			t.Fatalf("MarshalBER: %v", err)
		}
		got, err := ParseSriResp(data)
		if err != nil {
			t.Errorf("UnavailabilityCause %d: ParseSriResp: %v", v, err)
			continue
		}
		if got.UnavailabilityCause == nil || int64(*got.UnavailabilityCause) != v {
			t.Errorf("UnavailabilityCause = %v, want %d", got.UnavailabilityCause, v)
		}
	}
}

// IST-SupportIndicator ::= ENUMERATED { basicISTSupported(0),
// istCommandSupported(1), ...}: "reception of values > 1 shall be mapped to
// ' istCommandSupported '" (§17.7.1). A negative value is outside that rule
// and is kept.
func TestParseISTSupportIndicatorNegative(t *testing.T) {
	for _, v := range semUnknownEnumValues {
		ist := gsm_map.ISTSupportIndicator(v)

		sw, err := convertSriToArg(&Sri{MSISDN: "31612345678", GmscOrGsmSCFAddress: "31600000001"})
		if err != nil {
			t.Fatalf("convertSriToArg: %v", err)
		}
		sw.IstSupportIndicator = &ist
		data, err := sw.MarshalBER()
		if err != nil {
			t.Fatalf("MarshalBER: %v", err)
		}
		sri, err := ParseSri(data)
		if err != nil {
			t.Errorf("SendRoutingInfo %d: %v", v, err)
		} else if sri.IstSupportIndicator == nil || int64(*sri.IstSupportIndicator) != v {
			t.Errorf("SendRoutingInfo IstSupportIndicator = %v, want %d", sri.IstSupportIndicator, v)
		}

		uw, err := convertUpdateLocationToArg(&UpdateLocation{
			IMSI: "001010123456789", MSCNumber: "31612345678", VLRNumber: "31612345678",
			VlrCapability: &VlrCapability{},
		})
		if err != nil {
			t.Fatalf("convertUpdateLocationToArg: %v", err)
		}
		uw.VlrCapability.IstSupportIndicator = &ist
		data, err = uw.MarshalBER()
		if err != nil {
			t.Fatalf("MarshalBER: %v", err)
		}
		ul, err := ParseUpdateLocation(data)
		if err != nil {
			t.Errorf("UpdateLocation %d: %v", v, err)
		} else if ul.VlrCapability.IstSupportIndicator == nil || int64(*ul.VlrCapability.IstSupportIndicator) != v {
			t.Errorf("UpdateLocation IstSupportIndicator = %v, want %d", ul.VlrCapability.IstSupportIndicator, v)
		}
	}
}

// RequestingNodeType ::= ENUMERATED { vlr (0), sgsn (1), ..., s-cscf (2),
// bsf (3), gan-aaa-server (4), wlan-aaa-server (5), mme(16), mme-sgsn(17)}:
// "received values in the range (6-15) shall be treated as "vlr"" and
// "received values greater than 17 shall be treated as "sgsn"" (§17.7.1).
// A negative value is outside both rules and is kept.
func TestParseRequestingNodeTypeNegative(t *testing.T) {
	for _, v := range semUnknownEnumValues {
		w := &gsm_map.SendAuthenticationInfoArg{Imsi: gsm_map.IMSI{0x21, 0x43, 0x65}, NumberOfRequestedVectors: 1}
		rnt := gsm_map.RequestingNodeType(v)
		w.RequestingNodeType = &rnt
		data, err := w.MarshalBER()
		if err != nil {
			t.Fatalf("MarshalBER: %v", err)
		}
		got, err := ParseSendAuthenticationInfo(data)
		if err != nil {
			t.Errorf("RequestingNodeType %d: %v", v, err)
			continue
		}
		if got.RequestingNodeType == nil || int64(*got.RequestingNodeType) != v {
			t.Errorf("RequestingNodeType = %v, want %d", got.RequestingNodeType, v)
		}
	}
}

// DefaultCallHandling, DefaultSMS-Handling and DefaultGPRS-Handling map
// 2..31 and values above 31 (§17.7.1); a negative value is outside both
// rules and is kept.
func TestParseDefaultHandlingNegative(t *testing.T) {
	cch := semCamelPhase
	in := &InsertSubscriberDataArg{
		VlrCamelSubscriptionInfo: &VlrCamelSubscriptionInfo{
			OCSI: &OCSI{
				OBcsmCamelTDPDataList: []OBcsmCamelTDPData{{
					OBcsmTriggerDetectionPoint: OBcsmTriggerCollectedInfo,
					ServiceKey:                 1,
					GsmSCFAddress:              "31611111111",
					GsmSCFAddressNature:        address.NatureInternational,
					GsmSCFAddressPlan:          address.PlanISDN,
				}},
				CamelCapabilityHandling: &cch,
			},
			MoSmsCSI: &SMSCSI{
				SmsCAMELTDPDataList:     []SMSCAMELTDPData{semSMSTDPData(SMSTriggerDetectionPointSmsCollectedInfo)},
				CamelCapabilityHandling: &cch,
			},
		},
		SgsnCAMELSubscriptionInfo: &SGSNCAMELSubscriptionInfo{GprsCSI: &GPRSCSI{
			GprsCamelTDPDataList: GPRSCamelTDPDataList{{
				GprsTriggerDetectionPoint: GPRSTDPAttach,
				ServiceKey:                1,
				GsmSCFAddress:             "31611111111",
			}},
			CamelCapabilityHandling: &cch,
		}},
	}
	for _, tc := range []struct {
		name string
		set  func(w *gsm_map.InsertSubscriberDataArg, v int64)
		get  func(a *InsertSubscriberDataArg) int64
	}{
		{"DefaultCallHandling",
			func(w *gsm_map.InsertSubscriberDataArg, v int64) {
				w.VlrCamelSubscriptionInfo.OCSI.OBcsmCamelTDPDataList.Values[0].DefaultCallHandling = gsm_map.DefaultCallHandling(v)
			},
			func(a *InsertSubscriberDataArg) int64 {
				return int64(a.VlrCamelSubscriptionInfo.OCSI.OBcsmCamelTDPDataList[0].DefaultCallHandling)
			}},
		{"DefaultSMS-Handling",
			func(w *gsm_map.InsertSubscriberDataArg, v int64) {
				w.VlrCamelSubscriptionInfo.MoSmsCSI.SmsCAMELTDPDataList.Values[0].DefaultSMSHandling = gsm_map.DefaultSMSHandling(v)
			},
			func(a *InsertSubscriberDataArg) int64 {
				return int64(a.VlrCamelSubscriptionInfo.MoSmsCSI.SmsCAMELTDPDataList[0].DefaultSMSHandling)
			}},
		{"DefaultGPRS-Handling",
			func(w *gsm_map.InsertSubscriberDataArg, v int64) {
				w.SgsnCAMELSubscriptionInfo.GprsCSI.GprsCamelTDPDataList.Values[0].DefaultSessionHandling = gsm_map.DefaultGPRSHandling(v)
			},
			func(a *InsertSubscriberDataArg) int64 {
				return int64(a.SgsnCAMELSubscriptionInfo.GprsCSI.GprsCamelTDPDataList[0].DefaultSessionHandling)
			}},
	} {
		for _, v := range semUnknownEnumValues {
			w, err := convertInsertSubscriberDataArgToWire(in)
			if err != nil {
				t.Fatalf("convertInsertSubscriberDataArgToWire: %v", err)
			}
			tc.set(w, v)
			got, err := semParseISDWire(t, w)
			if err != nil {
				t.Errorf("%s %d: ParseInsertSubscriberData: %v", tc.name, v, err)
				continue
			}
			if g := tc.get(got); g != v {
				t.Errorf("%s = %d, want %d", tc.name, g, v)
			}
		}
	}
}

package gsmmap

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// Non-extensible ENUMERATED membership is checked on encode and decode
// per 3GPP TS 29.002 V19.1.0 §§17.7.1, 17.7.6 and 17.7.8.

// unlistedEnums are wire values outside every enumeration below.
var unlistedEnums = []int64{-1, 9, math.MaxInt64, math.MinInt64}

// --- SubscriberStatus, §17.7.1 ---

func TestSubscriberStatusMembership(t *testing.T) {
	for _, v := range []SubscriberStatus{SubscriberStatusServiceGranted, SubscriberStatusOperatorDeterminedBarring} {
		s := v
		data, err := (&InsertSubscriberDataArg{SubscriberStatus: &s}).Marshal()
		if err != nil {
			t.Fatalf("Marshal SubscriberStatus %d: %v", v, err)
		}
		got, err := ParseInsertSubscriberData(data)
		if err != nil {
			t.Fatalf("Parse SubscriberStatus %d: %v", v, err)
		}
		wantEqual(t, fmt.Sprintf("SubscriberStatus %d", v), &s, got.SubscriberStatus)
	}
	for _, v := range append(unlistedEnums, 2) {
		s := SubscriberStatus(v)
		if _, err := (&InsertSubscriberDataArg{SubscriberStatus: &s}).Marshal(); !matchesConstraint(err, "subscriberStatus", "ENUMERATED {0, 1}") {
			t.Errorf("Marshal SubscriberStatus %d: err = %v, want codec constraint", v, err)
		}
		w := isdWire(t, &InsertSubscriberDataArg{})
		w.SubscriberStatus = &s
		if got, err := ParseInsertSubscriberData(tolerantBER(t, w)); !matchesConstraint(err, "subscriberStatus", "ENUMERATED {0, 1}") || got != nil {
			t.Errorf("Parse SubscriberStatus %d: got %v, %v; want codec constraint", v, got, err)
		}
	}
}

// --- RegionalSubscriptionResponse, §17.7.1 ---

func TestRegionalSubscriptionResponseMembership(t *testing.T) {
	for _, v := range []RegionalSubscriptionResponse{RegionalSubscriptionResponseNetworkNodeAreaRestricted, RegionalSubscriptionResponseRegionalSubscNotSupported} {
		r := v
		data, err := (&InsertSubscriberDataRes{RegionalSubscriptionResponse: &r}).Marshal()
		if err != nil {
			t.Fatalf("Marshal RegionalSubscriptionResponse %d: %v", v, err)
		}
		got, err := ParseInsertSubscriberDataRes(data)
		if err != nil {
			t.Fatalf("Parse RegionalSubscriptionResponse %d: %v", v, err)
		}
		wantEqual(t, fmt.Sprintf("RegionalSubscriptionResponse %d", v), &r, got.RegionalSubscriptionResponse)
	}
	for _, v := range append(unlistedEnums, 4) {
		r := RegionalSubscriptionResponse(v)
		if _, err := (&InsertSubscriberDataRes{RegionalSubscriptionResponse: &r}).Marshal(); !matchesConstraint(err, "regionalSubscriptionResponse", "ENUMERATED {0, 1, 2, 3}") {
			t.Errorf("Marshal RegionalSubscriptionResponse %d: err = %v, want codec constraint", v, err)
		}
		w := &gsm_map.InsertSubscriberDataRes{RegionalSubscriptionResponse: &r}
		if got, err := ParseInsertSubscriberDataRes(tolerantBER(t, w)); !matchesConstraint(err, "regionalSubscriptionResponse", "ENUMERATED {0, 1, 2, 3}") || got != nil {
			t.Errorf("Parse RegionalSubscriptionResponse %d: got %v, %v; want codec constraint", v, got, err)
		}
	}
}

// --- NetworkResource, §17.7.8 ---

// systemFailure carries NetworkResource as its legacy CHOICE alternative and
// inside ExtensibleSystemFailureParam. MAP errors are decoded only.
func TestNetworkResourceMembership(t *testing.T) {
	// The probe: networkResource 9 as the legacy alternative.
	data, err := hex.DecodeString("0a0109")
	if err != nil {
		t.Fatal(err)
	}
	if got, err := ParseReturnErrorParameter(MapErrorSystemFailure, data); !matchesConstraint(err, "networkResource", "ENUMERATED {0, 1, 2, 3, 4, 5, 6, 7}") || got != nil {
		t.Errorf("0a0109: got %v, %v; want codec constraint", got, err)
	}
	encode := func(t *testing.T, w gsm_map.SystemFailureParam) []byte {
		t.Helper()
		return strictBER(t, &w)
	}
	for _, v := range []gsm_map.NetworkResource{gsm_map.NetworkResourcePlmn, gsm_map.NetworkResourceRss} {
		got, err := ParseReturnErrorParameter(MapErrorSystemFailure, encode(t, gsm_map.NewSystemFailureParamNetworkResource(v)))
		if err != nil {
			t.Fatalf("legacy %d: %v", v, err)
		}
		r := v
		wantEqual(t, "legacy", &SystemFailureParam{NetworkResource: &r}, got.(*SystemFailureParam))

		got, err = ParseReturnErrorParameter(MapErrorSystemFailure, encode(t, gsm_map.NewSystemFailureParamExtensibleSystemFailureParam(gsm_map.ExtensibleSystemFailureParam{NetworkResource: &r})))
		if err != nil {
			t.Fatalf("extensible %d: %v", v, err)
		}
		wantEqual(t, "extensible", &SystemFailureParam{ExtensibleSystemFailureParam: &ExtensibleSystemFailureParam{NetworkResource: &r}}, got.(*SystemFailureParam))
	}
	for _, v := range append(unlistedEnums, 8) {
		r := gsm_map.NetworkResource(v)
		legacy := gsm_map.NewSystemFailureParamNetworkResource(r)
		extensible := gsm_map.NewSystemFailureParamExtensibleSystemFailureParam(gsm_map.ExtensibleSystemFailureParam{NetworkResource: &r})
		if got, err := ParseReturnErrorParameter(MapErrorSystemFailure, tolerantBER(t, &legacy)); !matchesConstraint(err, "networkResource", "ENUMERATED {0, 1, 2, 3, 4, 5, 6, 7}") || got != nil {
			t.Errorf("legacy %d: got %v, %v; want codec constraint", v, got, err)
		}
		if got, err := ParseReturnErrorParameter(MapErrorSystemFailure, tolerantBER(t, &extensible)); !matchesConstraint(err, "extensibleSystemFailureParam.networkResource", "ENUMERATED {0, 1, 2, 3, 4, 5, 6, 7}") || got != nil {
			t.Errorf("extensible %d: got %v, %v; want codec constraint", v, got, err)
		}
	}
}

// --- ProtocolId, §17.7.8 ---

// sriWithSignalInfo is a SendRoutingInfo-Arg carrying networkSignalInfo and
// networkSignalInfo2, and sriRespWithBearer a SendRoutingInfo-Res carrying
// gsmBearerCapability: the three ExternalSignalInfo fields.
func sriWithSignalInfo(id, id2 ProtocolID) *Sri {
	return &Sri{
		MSISDN:              "31612345678",
		InterrogationType:   InterrogationBasicCall,
		GmscOrGsmSCFAddress: "31201111111",
		NetworkSignalInfo:   &ExternalSignalInfo{ProtocolID: id, SignalInfo: HexBytes{0xde, 0xad}},
		NetworkSignalInfo2:  &ExternalSignalInfo{ProtocolID: id2, SignalInfo: HexBytes{0xbe, 0xef}},
	}
}

func sriRespWithBearer(id ProtocolID) *SriResp {
	return &SriResp{
		IMSI:                "204080012345678",
		GsmBearerCapability: &ExternalSignalInfo{ProtocolID: id, SignalInfo: HexBytes{0xde, 0xad}},
	}
}

func TestProtocolIDMembership(t *testing.T) {
	for _, id := range []ProtocolID{ProtocolIDGsm0408, ProtocolIDGsm0806, ProtocolIDEts3001021} {
		data, err := sriWithSignalInfo(id, id).Marshal()
		if err != nil {
			t.Fatalf("Sri ProtocolID %d: Marshal: %v", id, err)
		}
		got, err := ParseSri(data)
		if err != nil {
			t.Fatalf("Sri ProtocolID %d: Parse: %v", id, err)
		}
		wantEqual(t, "Sri", sriWithSignalInfo(id, id), got)

		data, err = sriRespWithBearer(id).Marshal()
		if err != nil {
			t.Fatalf("SriResp ProtocolID %d: Marshal: %v", id, err)
		}
		resp, err := ParseSriResp(data)
		if err != nil {
			t.Fatalf("SriResp ProtocolID %d: Parse: %v", id, err)
		}
		wantEqual(t, "SriResp", sriRespWithBearer(id), resp)
	}
	for _, v := range append(unlistedEnums, 0, 5) {
		id := ProtocolID(v)
		for name, s := range map[string]*Sri{"networkSignalInfo": sriWithSignalInfo(id, 1), "networkSignalInfo2": sriWithSignalInfo(1, id)} {
			if _, err := s.Marshal(); !matchesConstraint(err, name+".protocolId", "ENUMERATED {1, 2, 3, 4}") {
				t.Errorf("Marshal %s ProtocolID %d: err = %v, want codec constraint", name, v, err)
			}
		}
		if _, err := sriRespWithBearer(id).Marshal(); !matchesConstraint(err, "gsm-BearerCapability.protocolId", "ENUMERATED {1, 2, 3, 4}") {
			t.Errorf("Marshal gsmBearerCapability ProtocolID %d: err = %v, want codec constraint", v, err)
		}
	}
	for _, v := range append(unlistedEnums, 0, 5) {
		w, err := convertSriToArg(sriWithSignalInfo(1, 2))
		if err != nil {
			t.Fatal(err)
		}
		w.NetworkSignalInfo.ProtocolId = gsm_map.ProtocolId(v)
		if got, err := ParseSri(tolerantBER(t, w)); !matchesConstraint(err, "networkSignalInfo.protocolId", "ENUMERATED {1, 2, 3, 4}") || got != nil {
			t.Errorf("Parse networkSignalInfo ProtocolId %d: got %v, %v; want codec constraint", v, got, err)
		}
		w, err = convertSriToArg(sriWithSignalInfo(1, 2))
		if err != nil {
			t.Fatal(err)
		}
		w.NetworkSignalInfo2.ProtocolId = gsm_map.ProtocolId(v)
		if got, err := ParseSri(tolerantBER(t, w)); !matchesConstraint(err, "networkSignalInfo2.protocolId", "ENUMERATED {1, 2, 3, 4}") || got != nil {
			t.Errorf("Parse networkSignalInfo2 ProtocolId %d: got %v, %v; want codec constraint", v, got, err)
		}
		res, err := convertSriRespToRes(sriRespWithBearer(1))
		if err != nil {
			t.Fatal(err)
		}
		res.GsmBearerCapability.ProtocolId = gsm_map.ProtocolId(v)
		if got, err := ParseSriResp(tolerantBER(t, res)); !matchesConstraint(err, "gsm-BearerCapability.protocolId", "ENUMERATED {1, 2, 3, 4}") || got != nil {
			t.Errorf("Parse gsmBearerCapability ProtocolId %d: got %v, %v; want codec constraint", v, got, err)
		}
	}
}

// gsm-BSSMAP (3) is listed, but "Value 3 is reserved and must not be used":
// a sender rule. Marshal refuses it; Parse keeps a received 3.
func TestProtocolIDReservedValue(t *testing.T) {
	if _, err := sriWithSignalInfo(3, 1).Marshal(); !errors.Is(err, ErrProtocolIDReserved) {
		t.Errorf("Marshal networkSignalInfo 3: err = %v, want ErrProtocolIDReserved", err)
	}
	if _, err := sriWithSignalInfo(1, 3).Marshal(); !errors.Is(err, ErrProtocolIDReserved) {
		t.Errorf("Marshal networkSignalInfo2 3: err = %v, want ErrProtocolIDReserved", err)
	}
	if _, err := sriRespWithBearer(3).Marshal(); !errors.Is(err, ErrProtocolIDReserved) {
		t.Errorf("Marshal gsmBearerCapability 3: err = %v, want ErrProtocolIDReserved", err)
	}

	w, err := convertSriToArg(sriWithSignalInfo(1, 2))
	if err != nil {
		t.Fatal(err)
	}
	w.NetworkSignalInfo.ProtocolId = gsm_map.ProtocolIdGsmBSSMAP
	w.NetworkSignalInfo2.ProtocolId = gsm_map.ProtocolIdGsmBSSMAP
	got, err := ParseSri(strictBER(t, w))
	if err != nil {
		t.Fatalf("ParseSri: %v", err)
	}
	wantEqual(t, "Sri", sriWithSignalInfo(3, 3), got)

	res, err := convertSriRespToRes(sriRespWithBearer(1))
	if err != nil {
		t.Fatal(err)
	}
	res.GsmBearerCapability.ProtocolId = gsm_map.ProtocolIdGsmBSSMAP
	resp, err := ParseSriResp(strictBER(t, res))
	if err != nil {
		t.Fatalf("ParseSriResp: %v", err)
	}
	wantEqual(t, "SriResp", sriRespWithBearer(3), resp)
	if err := checkParseRoundTrip("Sri", asParser(ParseSri), strictBER(t, w)); err != nil {
		t.Error(err)
	}
}

// --- SM-DeliveryOutcome in MO-ForwardSM-Arg, §17.7.6 ---

func TestMoFsmSmDeliveryOutcomeMembership(t *testing.T) {
	for _, v := range []SmDeliveryOutcome{gsm_map.SMDeliveryOutcomeMemoryCapacityExceeded, gsm_map.SMDeliveryOutcomeSuccessfulTransfer} {
		m := knownMoFsm(t)
		o := v
		m.SmDeliveryOutcome = &o
		data, err := m.Marshal()
		if err != nil {
			t.Fatalf("Marshal SmDeliveryOutcome %d: %v", v, err)
		}
		got, err := ParseMoFsm(data)
		if err != nil {
			t.Fatalf("Parse SmDeliveryOutcome %d: %v", v, err)
		}
		wantEqual(t, "SmDeliveryOutcome", &o, got.SmDeliveryOutcome)
	}
	for _, v := range append(unlistedEnums, 3) {
		m := knownMoFsm(t)
		o := SmDeliveryOutcome(v)
		m.SmDeliveryOutcome = &o
		if _, err := m.Marshal(); !matchesConstraint(err, "sm-DeliveryOutcome", "ENUMERATED {0, 1, 2}") {
			t.Errorf("Marshal SmDeliveryOutcome %d: err = %v, want codec constraint", v, err)
		}
		w, err := convertMoFsmToArg(knownMoFsm(t))
		if err != nil {
			t.Fatal(err)
		}
		w.SmDeliveryOutcome = &o
		if got, err := ParseMoFsm(tolerantBER(t, w)); !matchesConstraint(err, "sm-DeliveryOutcome", "ENUMERATED {0, 1, 2}") || got != nil {
			t.Errorf("Parse SmDeliveryOutcome %d: got %v, %v; want codec constraint", v, got, err)
		}
	}
}

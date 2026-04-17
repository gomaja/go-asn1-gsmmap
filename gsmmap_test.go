package gsmmap

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gomaja/go-asn1-gsmmap/address"
)

func TestSriSmRoundTrip(t *testing.T) {
	sriSm := &SriSm{
		MSISDN:               "123456789",
		SmRpPri:              true,
		ServiceCentreAddress: "12345",
	}

	data, err := sriSm.Marshal()
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	parsed, err := ParseSriSm(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if sriSm.MSISDN != parsed.MSISDN {
		t.Errorf("MSISDN: got %s, want %s", parsed.MSISDN, sriSm.MSISDN)
	}
	if sriSm.SmRpPri != parsed.SmRpPri {
		t.Errorf("SmRpPri: got %v, want %v", parsed.SmRpPri, sriSm.SmRpPri)
	}
	if sriSm.ServiceCentreAddress != parsed.ServiceCentreAddress {
		t.Errorf("ServiceCentreAddress: got %s, want %s", parsed.ServiceCentreAddress, sriSm.ServiceCentreAddress)
	}
}

func TestSriSmRespRoundTrip(t *testing.T) {
	resp := &SriSmResp{
		IMSI: "123456789012345",
		LocationInfoWithLMSI: LocationInfoWithLMSI{
			NetworkNodeNumber: "12345",
		},
	}

	data, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	parsed, err := ParseSriSmResp(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if resp.IMSI != parsed.IMSI {
		t.Errorf("IMSI: got %s, want %s", parsed.IMSI, resp.IMSI)
	}
	if resp.LocationInfoWithLMSI.NetworkNodeNumber != parsed.LocationInfoWithLMSI.NetworkNodeNumber {
		t.Errorf("NetworkNodeNumber: got %s, want %s",
			parsed.LocationInfoWithLMSI.NetworkNodeNumber,
			resp.LocationInfoWithLMSI.NetworkNodeNumber)
	}
}

func TestUpdateLocationRoundTrip(t *testing.T) {
	ul := &UpdateLocation{
		IMSI:      "607036003958556",
		MSCNumber: "628160360000",
		VLRNumber: "628160360000",
		VlrCapability: &VlrCapability{
			SupportedCamelPhases: &SupportedCamelPhases{
				Phase1: true,
				Phase2: true,
			},
		},
	}

	data, err := ul.Marshal()
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	parsed, err := ParseUpdateLocation(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if ul.IMSI != parsed.IMSI {
		t.Errorf("IMSI: got %s, want %s", parsed.IMSI, ul.IMSI)
	}
	if ul.MSCNumber != parsed.MSCNumber {
		t.Errorf("MSCNumber: got %s, want %s", parsed.MSCNumber, ul.MSCNumber)
	}
	if ul.VLRNumber != parsed.VLRNumber {
		t.Errorf("VLRNumber: got %s, want %s", parsed.VLRNumber, ul.VLRNumber)
	}

	if parsed.VlrCapability == nil {
		t.Fatal("VlrCapability is nil")
	}
	if parsed.VlrCapability.SupportedCamelPhases == nil {
		t.Fatal("SupportedCamelPhases is nil")
	}
	if !parsed.VlrCapability.SupportedCamelPhases.Phase1 {
		t.Error("Phase1 should be true")
	}
	if !parsed.VlrCapability.SupportedCamelPhases.Phase2 {
		t.Error("Phase2 should be true")
	}
	if parsed.VlrCapability.SupportedCamelPhases.Phase3 {
		t.Error("Phase3 should be false")
	}
}

func TestUpdateLocationWithLCSRoundTrip(t *testing.T) {
	ul := &UpdateLocation{
		IMSI:      "234507097995732",
		MSCNumber: "996772589400",
		VLRNumber: "996772589400",
		VlrCapability: &VlrCapability{
			SupportedCamelPhases: &SupportedCamelPhases{
				Phase1: true,
				Phase2: true,
				Phase3: true,
				Phase4: true,
			},
			SupportedLCSCapabilitySets: &SupportedLCSCapabilitySets{
				LcsCapabilitySet1: true,
				LcsCapabilitySet2: true,
			},
		},
	}

	data, err := ul.Marshal()
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	parsed, err := ParseUpdateLocation(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if parsed.VlrCapability == nil || parsed.VlrCapability.SupportedLCSCapabilitySets == nil {
		t.Fatal("LCS capability sets missing")
	}
	if !parsed.VlrCapability.SupportedLCSCapabilitySets.LcsCapabilitySet1 {
		t.Error("LcsCapabilitySet1 should be true")
	}
	if !parsed.VlrCapability.SupportedLCSCapabilitySets.LcsCapabilitySet2 {
		t.Error("LcsCapabilitySet2 should be true")
	}
}

func TestUpdateLocationResRoundTrip(t *testing.T) {
	res := &UpdateLocationRes{
		HLRNumber: "62816036",
	}

	data, err := res.Marshal()
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	parsed, err := ParseUpdateLocationRes(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if res.HLRNumber != parsed.HLRNumber {
		t.Errorf("HLRNumber: got %s, want %s", parsed.HLRNumber, res.HLRNumber)
	}
}

func TestUpdateGprsLocationRoundTrip(t *testing.T) {
	ul := &UpdateGprsLocation{
		IMSI:        "1234567890123456",
		SGSNNumber:  "628160360000",
		SGSNAddress: "192.168.1.1",
		SGSNCapability: &SGSNCapability{
			GprsEnhancementsSupportIndicator: true,
		},
	}

	data, err := ul.Marshal()
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	parsed, err := ParseUpdateGprsLocation(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if ul.IMSI != parsed.IMSI {
		t.Errorf("IMSI: got %s, want %s", parsed.IMSI, ul.IMSI)
	}
	if ul.SGSNNumber != parsed.SGSNNumber {
		t.Errorf("SGSNNumber: got %s, want %s", parsed.SGSNNumber, ul.SGSNNumber)
	}
	if ul.SGSNAddress != parsed.SGSNAddress {
		t.Errorf("SGSNAddress: got %s, want %s", parsed.SGSNAddress, ul.SGSNAddress)
	}
	if parsed.SGSNCapability == nil {
		t.Fatal("SGSNCapability is nil")
	}
	if !parsed.SGSNCapability.GprsEnhancementsSupportIndicator {
		t.Error("GprsEnhancementsSupportIndicator should be true")
	}
}

func TestUpdateGprsLocationWithLCSRoundTrip(t *testing.T) {
	ul := &UpdateGprsLocation{
		IMSI:        "1234567890123456",
		SGSNNumber:  "628160360000",
		SGSNAddress: "192.168.1.1",
		SGSNCapability: &SGSNCapability{
			GprsEnhancementsSupportIndicator: true,
			SupportedLCSCapabilitySets: &SupportedLCSCapabilitySets{
				LcsCapabilitySet1: true,
				LcsCapabilitySet2: true,
			},
		},
	}

	data, err := ul.Marshal()
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	parsed, err := ParseUpdateGprsLocation(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if parsed.SGSNCapability == nil || parsed.SGSNCapability.SupportedLCSCapabilitySets == nil {
		t.Fatal("LCS capability sets missing")
	}
	if !parsed.SGSNCapability.SupportedLCSCapabilitySets.LcsCapabilitySet1 {
		t.Error("LcsCapabilitySet1 should be true")
	}
	if !parsed.SGSNCapability.SupportedLCSCapabilitySets.LcsCapabilitySet2 {
		t.Error("LcsCapabilitySet2 should be true")
	}
}

func TestUpdateGprsLocationResRoundTrip(t *testing.T) {
	res := &UpdateGprsLocationRes{
		HLRNumber: "62816036",
	}

	data, err := res.Marshal()
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	parsed, err := ParseUpdateGprsLocationRes(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if res.HLRNumber != parsed.HLRNumber {
		t.Errorf("HLRNumber: got %s, want %s", parsed.HLRNumber, res.HLRNumber)
	}
}

func TestATIMsisdnRoundTrip(t *testing.T) {
	ati := &AnyTimeInterrogation{
		SubscriberIdentity: SubscriberIdentity{
			MSISDN: "881018742052329",
		},
		RequestedInfo: RequestedInfo{
			IMEI: true,
		},
		GsmSCFAddress: "881009060400008",
	}

	data, err := ati.Marshal()
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	parsed, err := ParseAnyTimeInterrogation(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if ati.SubscriberIdentity.MSISDN != parsed.SubscriberIdentity.MSISDN {
		t.Errorf("MSISDN: got %s, want %s", parsed.SubscriberIdentity.MSISDN, ati.SubscriberIdentity.MSISDN)
	}
	if !parsed.RequestedInfo.IMEI {
		t.Error("IMEI should be true")
	}
	if ati.GsmSCFAddress != parsed.GsmSCFAddress {
		t.Errorf("GsmSCFAddress: got %s, want %s", parsed.GsmSCFAddress, ati.GsmSCFAddress)
	}
}

func TestATIFullFeaturesRoundTrip(t *testing.T) {
	csDomain := CsDomain
	ati := &AnyTimeInterrogation{
		SubscriberIdentity: SubscriberIdentity{
			MSISDN: "881018684015514",
		},
		RequestedInfo: RequestedInfo{
			LocationInformation: true,
			SubscriberState:     true,
			CurrentLocation:     true,
			RequestedDomain:     &csDomain,
			IMEI:                true,
			MsClassmark:         true,
			MnpRequestedInfo:    true,
		},
		GsmSCFAddress: "881009060400004",
	}

	data, err := ati.Marshal()
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	parsed, err := ParseAnyTimeInterrogation(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if !parsed.RequestedInfo.LocationInformation {
		t.Error("LocationInformation should be true")
	}
	if !parsed.RequestedInfo.SubscriberState {
		t.Error("SubscriberState should be true")
	}
	if !parsed.RequestedInfo.CurrentLocation {
		t.Error("CurrentLocation should be true")
	}
	if parsed.RequestedInfo.RequestedDomain == nil {
		t.Fatal("RequestedDomain is nil")
	}
	if *parsed.RequestedInfo.RequestedDomain != CsDomain {
		t.Errorf("RequestedDomain: got %d, want %d", *parsed.RequestedInfo.RequestedDomain, CsDomain)
	}
	if !parsed.RequestedInfo.IMEI {
		t.Error("IMEI should be true")
	}
	if !parsed.RequestedInfo.MsClassmark {
		t.Error("MsClassmark should be true")
	}
	if !parsed.RequestedInfo.MnpRequestedInfo {
		t.Error("MnpRequestedInfo should be true")
	}
}

func TestATIValidationErrors(t *testing.T) {
	// Neither IMSI nor MSISDN set
	ati := &AnyTimeInterrogation{
		RequestedInfo: RequestedInfo{IMEI: true},
		GsmSCFAddress: "12345",
	}
	_, err := ati.Marshal()
	if err == nil {
		t.Error("expected error for empty SubscriberIdentity")
	}

	// Both IMSI and MSISDN set
	ati = &AnyTimeInterrogation{
		SubscriberIdentity: SubscriberIdentity{
			IMSI:   "123456789",
			MSISDN: "987654321",
		},
		RequestedInfo: RequestedInfo{IMEI: true},
		GsmSCFAddress: "12345",
	}
	_, err = ati.Marshal()
	if err == nil {
		t.Error("expected error for ambiguous SubscriberIdentity")
	}
}

func TestParseInvalidData(t *testing.T) {
	badData := []byte{0x00, 0x01, 0x02}

	if _, err := ParseSriSm(badData); err == nil {
		t.Error("expected error for invalid SriSm data")
	}
	if _, err := ParseSriSmResp(badData); err == nil {
		t.Error("expected error for invalid SriSmResp data")
	}
	if _, err := ParseUpdateLocation(badData); err == nil {
		t.Error("expected error for invalid UpdateLocation data")
	}
	if _, err := ParseUpdateGprsLocation(badData); err == nil {
		t.Error("expected error for invalid UpdateGprsLocation data")
	}
	if _, err := ParseAnyTimeInterrogation(badData); err == nil {
		t.Error("expected error for invalid ATI data")
	}
}

func TestSriSmParseKnownBytes(t *testing.T) {
	// Known DER-encoded SriSm from go-gsmmap test suite
	hexStr := "301380069122608538188101ff8206912260909899"
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		t.Fatalf("hex decode error: %v", err)
	}

	parsed, err := ParseSriSm(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if parsed.MSISDN == "" {
		t.Error("MSISDN should not be empty")
	}
	if parsed.ServiceCentreAddress == "" {
		t.Error("ServiceCentreAddress should not be empty")
	}

	// Round-trip: marshal and parse again, verify semantic equality
	remarshaled, err := parsed.Marshal()
	if err != nil {
		t.Fatalf("Re-marshal error: %v", err)
	}

	reparsed, err := ParseSriSm(remarshaled)
	if err != nil {
		t.Fatalf("Re-parse error: %v", err)
	}

	if parsed.MSISDN != reparsed.MSISDN {
		t.Errorf("MSISDN mismatch after round-trip: %s vs %s", parsed.MSISDN, reparsed.MSISDN)
	}
	if parsed.SmRpPri != reparsed.SmRpPri {
		t.Errorf("SmRpPri mismatch after round-trip")
	}
	if parsed.ServiceCentreAddress != reparsed.ServiceCentreAddress {
		t.Errorf("ServiceCentreAddress mismatch after round-trip: %s vs %s",
			parsed.ServiceCentreAddress, reparsed.ServiceCentreAddress)
	}
}

func TestSriSmRespParseKnownBytes(t *testing.T) {
	hexStr := "3015040882131068584836f3a0098107917394950862f6"
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		t.Fatalf("hex decode error: %v", err)
	}

	parsed, err := ParseSriSmResp(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if parsed.IMSI == "" {
		t.Error("IMSI should not be empty")
	}
	if parsed.LocationInfoWithLMSI.NetworkNodeNumber == "" {
		t.Error("NetworkNodeNumber should not be empty")
	}
}

func TestUpdateLocationResParseKnownBytes(t *testing.T) {
	hexStr := "300704059126180663"
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		t.Fatalf("hex decode error: %v", err)
	}

	parsed, err := ParseUpdateLocationRes(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if parsed.HLRNumber != "62816036" {
		t.Errorf("HLRNumber: got %s, want 62816036", parsed.HLRNumber)
	}
}

func TestUpdateGprsLocationParseKnownBytes(t *testing.T) {
	// IPv4 with LCS capabilities
	hexStr := "3022040862006630020000f20407911487390120f3040504d5378647a006830085020640"
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		t.Fatalf("hex decode error: %v", err)
	}

	parsed, err := ParseUpdateGprsLocation(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if parsed.IMSI != "260066032000002" {
		t.Errorf("IMSI: got %s, want 260066032000002", parsed.IMSI)
	}
	if parsed.SGSNAddress != "213.55.134.71" {
		t.Errorf("SGSNAddress: got %s, want 213.55.134.71", parsed.SGSNAddress)
	}
}

func TestATIResSubscriberStateRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		state SubscriberState
	}{
		{"AssumedIdle", StateAssumedIdle},
		{"CamelBusy", StateCamelBusy},
		{"NotProvidedFromVLR", StateNotProvidedFromVLR},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &AnyTimeInterrogationRes{
				SubscriberInfo: SubscriberInfo{
					SubscriberState: &SubscriberStateInfo{
						State: tt.state,
					},
				},
			}

			data, err := res.Marshal()
			if err != nil {
				t.Fatalf("Marshal error: %v", err)
			}

			parsed, err := ParseAnyTimeInterrogationRes(data)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			if parsed.SubscriberInfo.SubscriberState == nil {
				t.Fatal("SubscriberState is nil")
			}
			if parsed.SubscriberInfo.SubscriberState.State != tt.state {
				t.Errorf("State: got %d, want %d", parsed.SubscriberInfo.SubscriberState.State, tt.state)
			}
		})
	}
}

func TestATIResNetDetNotReachableRoundTrip(t *testing.T) {
	reason := ReasonImsiDetached
	res := &AnyTimeInterrogationRes{
		SubscriberInfo: SubscriberInfo{
			SubscriberState: &SubscriberStateInfo{
				State:              StateNetDetNotReachable,
				NotReachableReason: &reason,
			},
		},
	}

	data, err := res.Marshal()
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	parsed, err := ParseAnyTimeInterrogationRes(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	ss := parsed.SubscriberInfo.SubscriberState
	if ss == nil {
		t.Fatal("SubscriberState is nil")
	}
	if ss.State != StateNetDetNotReachable {
		t.Errorf("State: got %d, want %d", ss.State, StateNetDetNotReachable)
	}
	if ss.NotReachableReason == nil {
		t.Fatal("NotReachableReason is nil")
	}
	if *ss.NotReachableReason != ReasonImsiDetached {
		t.Errorf("NotReachableReason: got %d, want %d", *ss.NotReachableReason, ReasonImsiDetached)
	}
}

func TestATIResCSLocationRoundTrip(t *testing.T) {
	age := 120
	res := &AnyTimeInterrogationRes{
		SubscriberInfo: SubscriberInfo{
			LocationInformation: &CSLocationInformation{
				AgeOfLocationInformation: &age,
				VlrNumber:                "628160360000",
				MscNumber:                "628160360001",
				GeographicalInformation: &GeographicalInfo{
					ShapeType:       ShapeEllipsoidPointUncertainty,
					Latitude:        22.632522583007812,
					Longitude:       113.02974700927734,
					UncertaintyCode: func() *uint8 { v := uint8(0); return &v }(),
				},
				CurrentLocationRetrieved: true,
				SAIPresent:               true,
			},
		},
	}

	data, err := res.Marshal()
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	parsed, err := ParseAnyTimeInterrogationRes(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	loc := parsed.SubscriberInfo.LocationInformation
	if loc == nil {
		t.Fatal("LocationInformation is nil")
	}
	if loc.AgeOfLocationInformation == nil || *loc.AgeOfLocationInformation != 120 {
		t.Errorf("AgeOfLocationInformation: got %v, want 120", loc.AgeOfLocationInformation)
	}
	if loc.VlrNumber != "628160360000" {
		t.Errorf("VlrNumber: got %s, want 628160360000", loc.VlrNumber)
	}
	if loc.MscNumber != "628160360001" {
		t.Errorf("MscNumber: got %s, want 628160360001", loc.MscNumber)
	}
	if loc.GeographicalInformation == nil {
		t.Fatal("GeographicalInformation is nil")
	}
	if loc.GeographicalInformation.ShapeType != ShapeEllipsoidPointUncertainty {
		t.Errorf("GeographicalInformation.ShapeType: got %d, want %d", loc.GeographicalInformation.ShapeType, ShapeEllipsoidPointUncertainty)
	}
	if math.Abs(loc.GeographicalInformation.Latitude-22.632522583007812) > 0.0001 {
		t.Errorf("GeographicalInformation.Latitude: got %f, want ~22.6325", loc.GeographicalInformation.Latitude)
	}
	if math.Abs(loc.GeographicalInformation.Longitude-113.02974700927734) > 0.0001 {
		t.Errorf("GeographicalInformation.Longitude: got %f, want ~113.0297", loc.GeographicalInformation.Longitude)
	}
	if !loc.CurrentLocationRetrieved {
		t.Error("CurrentLocationRetrieved should be true")
	}
	if !loc.SAIPresent {
		t.Error("SAIPresent should be true")
	}
}

func TestATIResCSLocationCellIdRoundTrip(t *testing.T) {
	cellId := []byte{0x62, 0xf2, 0x20, 0x01, 0x23, 0x45, 0x67}
	res := &AnyTimeInterrogationRes{
		SubscriberInfo: SubscriberInfo{
			LocationInformation: &CSLocationInformation{
				CellGlobalId: cellId,
			},
		},
	}

	data, err := res.Marshal()
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	parsed, err := ParseAnyTimeInterrogationRes(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	loc := parsed.SubscriberInfo.LocationInformation
	if loc == nil {
		t.Fatal("LocationInformation is nil")
	}
	if !bytes.Equal(loc.CellGlobalId, cellId) {
		t.Errorf("CellGlobalId: got %x, want %x", loc.CellGlobalId, cellId)
	}
}

func TestATIResCSLocationLAIRoundTrip(t *testing.T) {
	lai := []byte{0x62, 0xf2, 0x20, 0x01, 0x23}
	res := &AnyTimeInterrogationRes{
		SubscriberInfo: SubscriberInfo{
			LocationInformation: &CSLocationInformation{
				LAI: lai,
			},
		},
	}

	data, err := res.Marshal()
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	parsed, err := ParseAnyTimeInterrogationRes(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	loc := parsed.SubscriberInfo.LocationInformation
	if loc == nil {
		t.Fatal("LocationInformation is nil")
	}
	if !bytes.Equal(loc.LAI, lai) {
		t.Errorf("LAI: got %x, want %x", loc.LAI, lai)
	}
}

func TestATIResEPSLocationRoundTrip(t *testing.T) {
	age := 30
	ecgi := []byte{0x62, 0xf2, 0x20, 0x01, 0x23, 0x45, 0x67}
	tai := []byte{0x62, 0xf2, 0x20, 0x01, 0x23}
	res := &AnyTimeInterrogationRes{
		SubscriberInfo: SubscriberInfo{
			LocationInformationEPS: &EPSLocationInformation{
				AgeOfLocationInformation: &age,
				EUtranCellGlobalIdentity: ecgi,
				TrackingAreaIdentity:     tai,
				CurrentLocationRetrieved: true,
			},
		},
	}

	data, err := res.Marshal()
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	parsed, err := ParseAnyTimeInterrogationRes(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	eps := parsed.SubscriberInfo.LocationInformationEPS
	if eps == nil {
		t.Fatal("LocationInformationEPS is nil")
	}
	if eps.AgeOfLocationInformation == nil || *eps.AgeOfLocationInformation != 30 {
		t.Errorf("AgeOfLocationInformation: got %v, want 30", eps.AgeOfLocationInformation)
	}
	if !bytes.Equal(eps.EUtranCellGlobalIdentity, ecgi) {
		t.Errorf("E-UTRAN CGI: got %x, want %x", eps.EUtranCellGlobalIdentity, ecgi)
	}
	if !bytes.Equal(eps.TrackingAreaIdentity, tai) {
		t.Errorf("TAI: got %x, want %x", eps.TrackingAreaIdentity, tai)
	}
	if !eps.CurrentLocationRetrieved {
		t.Error("CurrentLocationRetrieved should be true")
	}
}

func TestATIResGPRSLocationRoundTrip(t *testing.T) {
	age := 60
	res := &AnyTimeInterrogationRes{
		SubscriberInfo: SubscriberInfo{
			LocationInformationGPRS: &GPRSLocationInformation{
				AgeOfLocationInformation: &age,
				SgsnNumber:               "628160360000",
				RouteingAreaIdentity:     []byte{0x62, 0xf2, 0x20, 0x01, 0x23, 0x45},
				CurrentLocationRetrieved: true,
			},
		},
	}

	data, err := res.Marshal()
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	parsed, err := ParseAnyTimeInterrogationRes(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	gprs := parsed.SubscriberInfo.LocationInformationGPRS
	if gprs == nil {
		t.Fatal("LocationInformationGPRS is nil")
	}
	if gprs.AgeOfLocationInformation == nil || *gprs.AgeOfLocationInformation != 60 {
		t.Errorf("AgeOfLocationInformation: got %v, want 60", gprs.AgeOfLocationInformation)
	}
	if gprs.SgsnNumber != "628160360000" {
		t.Errorf("SgsnNumber: got %s, want 628160360000", gprs.SgsnNumber)
	}
	expectedRAI := []byte{0x62, 0xf2, 0x20, 0x01, 0x23, 0x45}
	if !bytes.Equal(gprs.RouteingAreaIdentity, expectedRAI) {
		t.Errorf("RouteingAreaIdentity: got %x, want %x", gprs.RouteingAreaIdentity, expectedRAI)
	}
	if !gprs.CurrentLocationRetrieved {
		t.Error("CurrentLocationRetrieved should be true")
	}
}

func TestATIResIMEIRoundTrip(t *testing.T) {
	res := &AnyTimeInterrogationRes{
		SubscriberInfo: SubscriberInfo{
			IMEI: "353456789012345",
		},
	}

	data, err := res.Marshal()
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	parsed, err := ParseAnyTimeInterrogationRes(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if parsed.SubscriberInfo.IMEI != "353456789012345" {
		t.Errorf("IMEI: got %s, want 353456789012345", parsed.SubscriberInfo.IMEI)
	}
}

func TestATIResFullRoundTrip(t *testing.T) {
	age := 90
	reason := ReasonRestrictedArea
	dst := 1
	res := &AnyTimeInterrogationRes{
		SubscriberInfo: SubscriberInfo{
			LocationInformation: &CSLocationInformation{
				AgeOfLocationInformation: &age,
				VlrNumber:                "628160360000",
				CurrentLocationRetrieved: true,
			},
			SubscriberState: &SubscriberStateInfo{
				State:              StateNetDetNotReachable,
				NotReachableReason: &reason,
			},
			IMEI:               "353456789012345",
			MsClassmark2:       []byte{0x33, 0x19, 0x83},
			TimeZone:           []byte{0x08},
			DaylightSavingTime: &dst,
		},
	}

	data, err := res.Marshal()
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	parsed, err := ParseAnyTimeInterrogationRes(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	si := parsed.SubscriberInfo

	// Location
	if si.LocationInformation == nil {
		t.Fatal("LocationInformation is nil")
	}
	if si.LocationInformation.VlrNumber != "628160360000" {
		t.Errorf("VlrNumber: got %s, want 628160360000", si.LocationInformation.VlrNumber)
	}

	// Subscriber state
	if si.SubscriberState == nil {
		t.Fatal("SubscriberState is nil")
	}
	if si.SubscriberState.State != StateNetDetNotReachable {
		t.Errorf("State: got %d, want %d", si.SubscriberState.State, StateNetDetNotReachable)
	}

	// IMEI
	if si.IMEI != "353456789012345" {
		t.Errorf("IMEI: got %s, want 353456789012345", si.IMEI)
	}

	// MsClassmark2
	expectedMsClassmark2 := []byte{0x33, 0x19, 0x83}
	if !bytes.Equal(si.MsClassmark2, expectedMsClassmark2) {
		t.Errorf("MsClassmark2: got %x, want %x", si.MsClassmark2, expectedMsClassmark2)
	}

	// TimeZone
	if len(si.TimeZone) != 1 || si.TimeZone[0] != 0x08 {
		t.Errorf("TimeZone: got %x, want [08]", si.TimeZone)
	}

	// DaylightSavingTime
	if si.DaylightSavingTime == nil || *si.DaylightSavingTime != 1 {
		t.Errorf("DaylightSavingTime: got %v, want 1", si.DaylightSavingTime)
	}
}

func TestATIResEmptyRoundTrip(t *testing.T) {
	// Minimal response with no optional fields
	res := &AnyTimeInterrogationRes{}

	data, err := res.Marshal()
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	parsed, err := ParseAnyTimeInterrogationRes(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if parsed.SubscriberInfo.LocationInformation != nil {
		t.Error("LocationInformation should be nil")
	}
	if parsed.SubscriberInfo.SubscriberState != nil {
		t.Error("SubscriberState should be nil")
	}
	if parsed.SubscriberInfo.IMEI != "" {
		t.Error("IMEI should be empty")
	}
}

func TestParseATIResInvalidData(t *testing.T) {
	if _, err := ParseAnyTimeInterrogationRes([]byte{0x00, 0x01, 0x02}); err == nil {
		t.Error("expected error for invalid ATI response data")
	}
}

// --- ATI (opCode 71) expanded coverage tests ---

func TestATIResPSSubscriberStateRoundTrip(t *testing.T) {
	reason := ReasonRestrictedArea
	tests := []struct {
		name string
		in   *PsSubscriberState
	}{
		{"ps-Detached", &PsSubscriberState{PsDetached: true}},
		{"ps-AttachedReachableForPaging", &PsSubscriberState{PsAttachedReachableForPaging: true}},
		{"ps-AttachedNotReachableForPaging", &PsSubscriberState{PsAttachedNotReachableForPaging: true}},
		{"notProvidedFromSGSNorMME", &PsSubscriberState{NotProvidedFromSGSNorMME: true}},
		{"netDetNotReachable", &PsSubscriberState{NetDetNotReachable: &reason}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &AnyTimeInterrogationRes{
				SubscriberInfo: SubscriberInfo{PsSubscriberState: tt.in},
			}
			data, err := res.Marshal()
			if err != nil {
				t.Fatalf("Marshal error: %v", err)
			}
			parsed, err := ParseAnyTimeInterrogationRes(data)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			ps := parsed.SubscriberInfo.PsSubscriberState
			if ps == nil {
				t.Fatal("PsSubscriberState is nil after round-trip")
			}
			if diff := cmp.Diff(tt.in, ps); diff != "" {
				t.Errorf("PsSubscriberState mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestATIResMNPInfoResRoundTrip(t *testing.T) {
	nps := MnpOwnNumberPortedOut
	in := &MnpInfoRes{
		RouteingNumber:          HexBytes{0x31, 0x33, 0x37},
		IMSI:                    "310150123456789",
		MSISDN:                  "31612345678",
		MSISDNNature:            address.NatureInternational,
		MSISDNPlan:              address.PlanISDN,
		NumberPortabilityStatus: &nps,
	}
	res := &AnyTimeInterrogationRes{
		SubscriberInfo: SubscriberInfo{MnpInfoRes: in},
	}
	data, err := res.Marshal()
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	parsed, err := ParseAnyTimeInterrogationRes(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	got := parsed.SubscriberInfo.MnpInfoRes
	if got == nil {
		t.Fatal("MnpInfoRes is nil after round-trip")
	}
	if !bytes.Equal(got.RouteingNumber, in.RouteingNumber) {
		t.Errorf("RouteingNumber: got %x, want %x", got.RouteingNumber, in.RouteingNumber)
	}
	if got.IMSI != in.IMSI {
		t.Errorf("IMSI: got %s, want %s", got.IMSI, in.IMSI)
	}
	if got.MSISDN != in.MSISDN {
		t.Errorf("MSISDN: got %s, want %s", got.MSISDN, in.MSISDN)
	}
	if got.NumberPortabilityStatus == nil || *got.NumberPortabilityStatus != nps {
		t.Errorf("NumberPortabilityStatus: got %v, want %v", got.NumberPortabilityStatus, nps)
	}
}

func TestATIResImsVoiceSupportRoundTrip(t *testing.T) {
	values := []ImsVoiceOverPSSessionsIndication{
		IMSVoiceOverPSNotSupported,
		IMSVoiceOverPSSupported,
		IMSVoiceOverPSUnknown,
	}
	for _, v := range values {
		v := v
		t.Run(map[ImsVoiceOverPSSessionsIndication]string{
			IMSVoiceOverPSNotSupported: "NotSupported",
			IMSVoiceOverPSSupported:    "Supported",
			IMSVoiceOverPSUnknown:      "Unknown",
		}[v], func(t *testing.T) {
			res := &AnyTimeInterrogationRes{
				SubscriberInfo: SubscriberInfo{
					ImsVoiceOverPSSessionsIndication: &v,
				},
			}
			data, err := res.Marshal()
			if err != nil {
				t.Fatalf("Marshal error: %v", err)
			}
			parsed, err := ParseAnyTimeInterrogationRes(data)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			got := parsed.SubscriberInfo.ImsVoiceOverPSSessionsIndication
			if got == nil || *got != v {
				t.Errorf("IMSVoiceOverPSSessionsIndication: got %v, want %v", got, v)
			}
		})
	}
}

func TestATIResLastActivityRoundTrip(t *testing.T) {
	ratType := UsedRatEUTRAN
	// Time is an opaque octet string per 3GPP TS 23.032.
	lastTime := HexBytes{0x31, 0x32, 0x31, 0x35, 0x31, 0x36, 0x32, 0x30, 0x34, 0x34, 0x35, 0x36, 0x5a}
	res := &AnyTimeInterrogationRes{
		SubscriberInfo: SubscriberInfo{
			LastUEActivityTime: lastTime,
			LastRATType:        &ratType,
		},
	}
	data, err := res.Marshal()
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	parsed, err := ParseAnyTimeInterrogationRes(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	si := parsed.SubscriberInfo
	if !bytes.Equal(si.LastUEActivityTime, lastTime) {
		t.Errorf("LastUEActivityTime: got %x, want %x", si.LastUEActivityTime, lastTime)
	}
	if si.LastRATType == nil || *si.LastRATType != ratType {
		t.Errorf("LastRATType: got %v, want %v", si.LastRATType, ratType)
	}
}

func TestATIResLocationInformation5GSRoundTrip(t *testing.T) {
	age := 31
	rat := UsedRatEUTRAN
	in := &LocationInformation5GS{
		NrCellGlobalIdentity:     HexBytes{0x62, 0xf2, 0x20, 0x01, 0x23, 0x45, 0x67, 0x89},
		EUtranCellGlobalIdentity: HexBytes{0x62, 0xf2, 0x20, 0x01, 0x23, 0x45, 0x67},
		GeographicalInformation: &GeographicalInfo{
			ShapeType:       ShapeEllipsoidPointUncertainty,
			Latitude:        22.632522583007812,
			Longitude:       113.02974700927734,
			UncertaintyCode: func() *uint8 { v := uint8(0); return &v }(),
		},
		GeodeticInformation:      HexBytes{0x80, 0x31, 0x11, 0x11, 0x22, 0x22, 0x33, 0x33, 0x44, 0x44},
		AmfAddress:               HexBytes("amf.example.com"),
		TrackingAreaIdentity:     HexBytes{0x62, 0xf2, 0x20, 0x01, 0x23},
		CurrentLocationRetrieved: true,
		AgeOfLocationInformation: &age,
		VplmnID:                  HexBytes{0x62, 0xf2, 0x20},
		LocalTimeZone:            HexBytes{0x08},
		RatType:                  &rat,
		NrTrackingAreaIdentity:   HexBytes{0x62, 0xf2, 0x20, 0x01, 0x23, 0x45},
	}
	res := &AnyTimeInterrogationRes{
		SubscriberInfo: SubscriberInfo{LocationInformation5GS: in},
	}
	data, err := res.Marshal()
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	parsed, err := ParseAnyTimeInterrogationRes(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	got := parsed.SubscriberInfo.LocationInformation5GS
	if got == nil {
		t.Fatal("LocationInformation5GS is nil after round-trip")
	}
	if !bytes.Equal(got.NrCellGlobalIdentity, in.NrCellGlobalIdentity) {
		t.Errorf("NrCellGlobalIdentity: got %x, want %x", got.NrCellGlobalIdentity, in.NrCellGlobalIdentity)
	}
	if !bytes.Equal(got.EUtranCellGlobalIdentity, in.EUtranCellGlobalIdentity) {
		t.Errorf("EUtranCellGlobalIdentity: got %x, want %x", got.EUtranCellGlobalIdentity, in.EUtranCellGlobalIdentity)
	}
	if got.GeographicalInformation == nil {
		t.Fatal("GeographicalInformation is nil")
	}
	if math.Abs(got.GeographicalInformation.Latitude-in.GeographicalInformation.Latitude) > 0.0001 {
		t.Errorf("Latitude: got %f, want %f", got.GeographicalInformation.Latitude, in.GeographicalInformation.Latitude)
	}
	if !bytes.Equal(got.GeodeticInformation, in.GeodeticInformation) {
		t.Errorf("GeodeticInformation: got %x, want %x", got.GeodeticInformation, in.GeodeticInformation)
	}
	if !bytes.Equal(got.AmfAddress, in.AmfAddress) {
		t.Errorf("AmfAddress: got %x, want %x", got.AmfAddress, in.AmfAddress)
	}
	if !bytes.Equal(got.TrackingAreaIdentity, in.TrackingAreaIdentity) {
		t.Errorf("TrackingAreaIdentity: got %x, want %x", got.TrackingAreaIdentity, in.TrackingAreaIdentity)
	}
	if !got.CurrentLocationRetrieved {
		t.Error("CurrentLocationRetrieved should be true")
	}
	if got.AgeOfLocationInformation == nil || *got.AgeOfLocationInformation != age {
		t.Errorf("AgeOfLocationInformation: got %v, want %v", got.AgeOfLocationInformation, age)
	}
	if !bytes.Equal(got.VplmnID, in.VplmnID) {
		t.Errorf("VplmnID: got %x, want %x", got.VplmnID, in.VplmnID)
	}
	if !bytes.Equal(got.LocalTimeZone, in.LocalTimeZone) {
		t.Errorf("LocalTimeZone: got %x, want %x", got.LocalTimeZone, in.LocalTimeZone)
	}
	if got.RatType == nil || *got.RatType != rat {
		t.Errorf("RatType: got %v, want %v", got.RatType, rat)
	}
	if !bytes.Equal(got.NrTrackingAreaIdentity, in.NrTrackingAreaIdentity) {
		t.Errorf("NrTrackingAreaIdentity: got %x, want %x", got.NrTrackingAreaIdentity, in.NrTrackingAreaIdentity)
	}
}

func TestATIResUserCSGInformationRoundTrip(t *testing.T) {
	// 27-bit CSG-Id packed into 4 octets per BIT STRING semantics.
	csg := &UserCSGInformation{
		CsgID:      HexBytes{0x11, 0x22, 0x33, 0x40},
		CsgIDBits:  27,
		AccessMode: HexBytes{0x00},
		CMI:        HexBytes{0x01},
	}

	// CS location carries UserCSGInformation on tag [11].
	csLoc := &CSLocationInformation{
		VlrNumber:          "31612345678",
		SelectedLSAId:      HexBytes{0xAA, 0xBB, 0xCC},
		UserCSGInformation: csg,
	}
	gprsLoc := &GPRSLocationInformation{
		SgsnNumber:          "31612345678",
		SelectedLSAIdentity: HexBytes{0xDD, 0xEE, 0xFF},
		UserCSGInformation:  csg,
	}

	res := &AnyTimeInterrogationRes{
		SubscriberInfo: SubscriberInfo{
			LocationInformation:     csLoc,
			LocationInformationGPRS: gprsLoc,
		},
	}
	data, err := res.Marshal()
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	parsed, err := ParseAnyTimeInterrogationRes(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	gotCS := parsed.SubscriberInfo.LocationInformation
	if gotCS == nil || gotCS.UserCSGInformation == nil {
		t.Fatal("CS LocationInformation.UserCSGInformation is nil")
	}
	if !bytes.Equal(gotCS.UserCSGInformation.CsgID, csg.CsgID) ||
		gotCS.UserCSGInformation.CsgIDBits != csg.CsgIDBits {
		t.Errorf("CS CsgID: got %x/%d, want %x/%d",
			gotCS.UserCSGInformation.CsgID, gotCS.UserCSGInformation.CsgIDBits,
			csg.CsgID, csg.CsgIDBits)
	}
	if !bytes.Equal(gotCS.SelectedLSAId, HexBytes{0xAA, 0xBB, 0xCC}) {
		t.Errorf("CS SelectedLSAId: got %x", gotCS.SelectedLSAId)
	}

	gotGPRS := parsed.SubscriberInfo.LocationInformationGPRS
	if gotGPRS == nil || gotGPRS.UserCSGInformation == nil {
		t.Fatal("GPRS LocationInformationGPRS.UserCSGInformation is nil")
	}
	if !bytes.Equal(gotGPRS.SelectedLSAIdentity, HexBytes{0xDD, 0xEE, 0xFF}) {
		t.Errorf("GPRS SelectedLSAIdentity: got %x", gotGPRS.SelectedLSAIdentity)
	}
	if gotGPRS.UserCSGInformation.CsgIDBits != csg.CsgIDBits {
		t.Errorf("GPRS UserCSGInformation.CsgIDBits: got %d, want %d",
			gotGPRS.UserCSGInformation.CsgIDBits, csg.CsgIDBits)
	}
}

func TestATIResFull5GSRoundTrip(t *testing.T) {
	age := 31
	nps := MnpForeignNumberPortedIn
	imsVoice := IMSVoiceOverPSSupported
	rat := UsedRatEUTRAN

	in := &AnyTimeInterrogationRes{
		SubscriberInfo: SubscriberInfo{
			LocationInformation: &CSLocationInformation{
				VlrNumber:                "31612345678",
				CurrentLocationRetrieved: true,
			},
			PsSubscriberState: &PsSubscriberState{
				PsAttachedReachableForPaging: true,
			},
			MnpInfoRes: &MnpInfoRes{
				RouteingNumber:          HexBytes{0x31, 0x33, 0x37},
				IMSI:                    "310150123456789",
				MSISDN:                  "31612345678",
				NumberPortabilityStatus: &nps,
			},
			ImsVoiceOverPSSessionsIndication: &imsVoice,
			LastUEActivityTime:               HexBytes{0x31, 0x32, 0x31, 0x35, 0x31, 0x36},
			LastRATType:                      &rat,
			LocationInformation5GS: &LocationInformation5GS{
				NrCellGlobalIdentity:     HexBytes{0x62, 0xf2, 0x20, 0x01, 0x23, 0x45, 0x67, 0x89},
				CurrentLocationRetrieved: true,
				AgeOfLocationInformation: &age,
				VplmnID:                  HexBytes{0x62, 0xf2, 0x20},
				RatType:                  &rat,
			},
		},
	}

	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	parsed, err := ParseAnyTimeInterrogationRes(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	si := parsed.SubscriberInfo
	if si.LocationInformation == nil || si.LocationInformation.VlrNumber != "31612345678" {
		t.Errorf("LocationInformation.VlrNumber mismatch: %+v", si.LocationInformation)
	}
	if si.PsSubscriberState == nil || !si.PsSubscriberState.PsAttachedReachableForPaging {
		t.Errorf("PsSubscriberState.PsAttachedReachableForPaging not set: %+v", si.PsSubscriberState)
	}
	if si.MnpInfoRes == nil || si.MnpInfoRes.IMSI != "310150123456789" {
		t.Errorf("MnpInfoRes.IMSI mismatch: %+v", si.MnpInfoRes)
	}
	if si.ImsVoiceOverPSSessionsIndication == nil || *si.ImsVoiceOverPSSessionsIndication != imsVoice {
		t.Errorf("IMSVoiceOverPSSessionsIndication mismatch: %v", si.ImsVoiceOverPSSessionsIndication)
	}
	if si.LastRATType == nil || *si.LastRATType != rat {
		t.Errorf("LastRATType mismatch: %v", si.LastRATType)
	}
	if si.LocationInformation5GS == nil {
		t.Fatal("LocationInformation5GS is nil")
	}
	if !si.LocationInformation5GS.CurrentLocationRetrieved {
		t.Error("LocationInformation5GS.CurrentLocationRetrieved should be true")
	}
}

func TestPsSubscriberStateChoiceValidation(t *testing.T) {
	reason := ReasonImsiDetached
	cases := []struct {
		name    string
		in      *PsSubscriberState
		wantErr error
	}{
		{
			name:    "empty",
			in:      &PsSubscriberState{},
			wantErr: ErrAtiPsSubscriberStateNoAlternative,
		},
		{
			name: "multiple",
			in: &PsSubscriberState{
				PsDetached:                   true,
				PsAttachedReachableForPaging: true,
			},
			wantErr: ErrAtiPsSubscriberStateMultipleAlternatives,
		},
		{
			name: "multiple-with-net-det",
			in: &PsSubscriberState{
				PsDetached:         true,
				NetDetNotReachable: &reason,
			},
			wantErr: ErrAtiPsSubscriberStateMultipleAlternatives,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := &AnyTimeInterrogationRes{
				SubscriberInfo: SubscriberInfo{PsSubscriberState: tc.in},
			}
			_, err := res.Marshal()
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got err=%v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestMarshalInvalidInputs(t *testing.T) {
	// Invalid MSISDN (non-hex)
	sriSm := &SriSm{
		MSISDN:               "invalid!",
		SmRpPri:              true,
		ServiceCentreAddress: "12345",
	}
	if _, err := sriSm.Marshal(); err == nil {
		t.Error("expected error for invalid MSISDN")
	}

	// Invalid IP in UpdateGprsLocation
	ugprs := &UpdateGprsLocation{
		IMSI:        "1234567890",
		SGSNNumber:  "12345",
		SGSNAddress: "not-an-ip",
	}
	if _, err := ugprs.Marshal(); err == nil {
		t.Error("expected error for invalid SGSNAddress")
	}
}

func TestSriSmFullStressRoundTrip(t *testing.T) {
	smRpMti := 1
	smDni := SmDeliveryOnlyMCCMNCRequested

	in := &SriSm{
		MSISDN:               "31612345678",
		SmRpPri:              true,
		ServiceCentreAddress: "31201111111",

		GprsSupportIndicator:    true,
		SmRpMti:                 &smRpMti,
		SmRpSmea:                HexBytes{0x91, 0x13, 0x26, 0x09, 0x10},
		SmDeliveryNotIntended:   &smDni,
		IpSmGwGuidanceIndicator: true,
		IMSI:                    "204080012345678",
		SingleAttemptDelivery:   true,
		T4TriggerIndicator:      true,
		CorrelationID: &SriSmCorrelationID{
			HlrID:   HexBytes{0xAA, 0xBB},
			SipUriA: HexBytes{0xCC, 0xDD},
			SipUriB: HexBytes{0xEE, 0xFF},
		},
		SmsfSupportIndicator: true,
	}

	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseSriSm(data)
	if err != nil {
		t.Fatalf("ParseSriSm: %v", err)
	}

	// Natures/plans normalize to International/ISDN when zero.
	in.MSISDNNature, in.MSISDNPlan = address.NatureInternational, address.PlanISDN
	in.SCANature, in.SCAPlan = address.NatureInternational, address.PlanISDN

	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("round-trip diff (-want +got):\n%s", diff)
	}
}

func TestSriSmRespFullStressRoundTrip(t *testing.T) {
	in := &SriSmResp{
		IMSI: "204080012345678",
		LocationInfoWithLMSI: LocationInfoWithLMSI{
			NetworkNodeNumber: "31612345678",
			LMSI:              HexBytes{0x01, 0x02, 0x03, 0x04},
			GprsNodeIndicator: true,
			AdditionalNumber: &AdditionalNumber{
				MscNumber: "31201111111",
			},
			NetworkNodeDiameterAddress: &NetworkNodeDiameterAddress{
				DiameterName:  HexBytes("msc.example.com"),
				DiameterRealm: HexBytes("example.com"),
			},
			AdditionalNetworkNodeDiameterAddress: &NetworkNodeDiameterAddress{
				DiameterName:  HexBytes("sgsn.example.com"),
				DiameterRealm: HexBytes("example.com"),
			},
			ThirdNumber: &AdditionalNumber{
				SgsnNumber: "31699999999",
			},
			ThirdNetworkNodeDiameterAddress: &NetworkNodeDiameterAddress{
				DiameterName:  HexBytes("third.example.com"),
				DiameterRealm: HexBytes("example.com"),
			},
			ImsNodeIndicator: true,
			Smsf3gppNumber:   "31688888888",
			Smsf3gppDiameterAddress: &NetworkNodeDiameterAddress{
				DiameterName:  HexBytes("smsf3gpp.example.com"),
				DiameterRealm: HexBytes("example.com"),
			},
			SmsfNon3gppNumber: "31677777777",
			SmsfNon3gppDiameterAddress: &NetworkNodeDiameterAddress{
				DiameterName:  HexBytes("smsfnon.example.com"),
				DiameterRealm: HexBytes("example.com"),
			},
			Smsf3gppAddressIndicator:    true,
			SmsfNon3gppAddressIndicator: true,
		},
		IpSmGwGuidance: &IpSmGwGuidance{
			MinimumDeliveryTimeValue:     60,
			RecommendedDeliveryTimeValue: 120,
		},
	}

	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseSriSmResp(data)
	if err != nil {
		t.Fatalf("ParseSriSmResp: %v", err)
	}

	// Normalize natures/plans.
	in.LocationInfoWithLMSI.NetworkNodeNumberNature = address.NatureInternational
	in.LocationInfoWithLMSI.NetworkNodeNumberPlan = address.PlanISDN
	in.LocationInfoWithLMSI.AdditionalNumber.MscNumberNature = address.NatureInternational
	in.LocationInfoWithLMSI.AdditionalNumber.MscNumberPlan = address.PlanISDN
	in.LocationInfoWithLMSI.ThirdNumber.SgsnNumberNature = address.NatureInternational
	in.LocationInfoWithLMSI.ThirdNumber.SgsnNumberPlan = address.PlanISDN
	in.LocationInfoWithLMSI.Smsf3gppNumberNature = address.NatureInternational
	in.LocationInfoWithLMSI.Smsf3gppNumberPlan = address.PlanISDN
	in.LocationInfoWithLMSI.SmsfNon3gppNumberNature = address.NatureInternational
	in.LocationInfoWithLMSI.SmsfNon3gppNumberPlan = address.PlanISDN

	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("round-trip diff (-want +got):\n%s", diff)
	}
}

func TestAdditionalNumberRoundTrip(t *testing.T) {
	t.Run("MscNumber", func(t *testing.T) {
		in := &SriSmResp{
			IMSI: "204080012345678",
			LocationInfoWithLMSI: LocationInfoWithLMSI{
				NetworkNodeNumber: "31612345678",
				AdditionalNumber: &AdditionalNumber{
					MscNumber: "31201111111",
				},
			},
		}

		data, err := in.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		got, err := ParseSriSmResp(data)
		if err != nil {
			t.Fatalf("ParseSriSmResp: %v", err)
		}

		if got.LocationInfoWithLMSI.AdditionalNumber == nil {
			t.Fatal("AdditionalNumber is nil")
		}
		if got.LocationInfoWithLMSI.AdditionalNumber.MscNumber != "31201111111" {
			t.Errorf("MscNumber: got %s, want 31201111111", got.LocationInfoWithLMSI.AdditionalNumber.MscNumber)
		}
		if got.LocationInfoWithLMSI.AdditionalNumber.SgsnNumber != "" {
			t.Errorf("SgsnNumber should be empty, got %s", got.LocationInfoWithLMSI.AdditionalNumber.SgsnNumber)
		}
	})

	t.Run("SgsnNumber", func(t *testing.T) {
		in := &SriSmResp{
			IMSI: "204080012345678",
			LocationInfoWithLMSI: LocationInfoWithLMSI{
				NetworkNodeNumber: "31612345678",
				AdditionalNumber: &AdditionalNumber{
					SgsnNumber: "31699999999",
				},
			},
		}

		data, err := in.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		got, err := ParseSriSmResp(data)
		if err != nil {
			t.Fatalf("ParseSriSmResp: %v", err)
		}

		if got.LocationInfoWithLMSI.AdditionalNumber == nil {
			t.Fatal("AdditionalNumber is nil")
		}
		if got.LocationInfoWithLMSI.AdditionalNumber.SgsnNumber != "31699999999" {
			t.Errorf("SgsnNumber: got %s, want 31699999999", got.LocationInfoWithLMSI.AdditionalNumber.SgsnNumber)
		}
		if got.LocationInfoWithLMSI.AdditionalNumber.MscNumber != "" {
			t.Errorf("MscNumber should be empty, got %s", got.LocationInfoWithLMSI.AdditionalNumber.MscNumber)
		}
	})
}

func TestMtFsmFullStressRoundTrip(t *testing.T) {
	// Parse a known valid MT-FSM to get a valid TPDU.
	knownHex := "3077800832140080803138f684069169318488880463040b916971101174f40000422182612464805bd2e2b1252d467ff6de6c47efd96eb6a1d056cb0d69b49a10269c098537586e96931965b260d15613da72c29b91261bde72c6a1ad2623d682b5996d58331271375a0d1733eee4bd98ec768bd966b41c0d"
	knownBytes, err := hex.DecodeString(knownHex)
	if err != nil {
		t.Fatalf("hex decode: %v", err)
	}
	base, err := ParseMtFsm(knownBytes)
	if err != nil {
		t.Fatalf("ParseMtFsm: %v", err)
	}

	timer := 120
	in := &MtFsm{
		IMSI:                   base.IMSI,
		ServiceCentreAddressOA: base.ServiceCentreAddressOA,
		SCAOANature:            base.SCAOANature,
		SCAOAPlan:              base.SCAOAPlan,
		TPDU:                   base.TPDU,
		MoreMessagesToSend:     true,

		SmDeliveryTimer:        &timer,
		SmDeliveryStartTime:    HexBytes{0x01, 0x02, 0x03, 0x04},
		SmsOverIPOnlyIndicator: true,
		CorrelationID: &SriSmCorrelationID{
			HlrID:   HexBytes{0xAA, 0xBB},
			SipUriA: HexBytes{0xCC, 0xDD},
			SipUriB: HexBytes{0xEE, 0xFF},
		},
		MaximumRetransmissionTime: HexBytes{0x05, 0x06, 0x07, 0x08},
		SmsGmscAddress:            "31612345678",
		SmsGmscDiameterAddress: &NetworkNodeDiameterAddress{
			DiameterName:  HexBytes("gmsc.example.com"),
			DiameterRealm: HexBytes("example.com"),
		},
	}

	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseMtFsm(data)
	if err != nil {
		t.Fatalf("ParseMtFsm: %v", err)
	}

	// Normalize default natures/plans.
	in.SmsGmscAddressNature = address.NatureInternational
	in.SmsGmscAddressPlan = address.PlanISDN

	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("round-trip diff (-want +got):\n%s", diff)
	}
}

func TestMtFsmRespRoundTrip(t *testing.T) {
	t.Run("WithSmRpUI", func(t *testing.T) {
		in := &MtFsmResp{
			SmRpUI: HexBytes{0x01, 0x02, 0x03},
		}
		data, err := in.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		got, err := ParseMtFsmResp(data)
		if err != nil {
			t.Fatalf("ParseMtFsmResp: %v", err)
		}
		if !bytes.Equal(in.SmRpUI, got.SmRpUI) {
			t.Errorf("SmRpUI: got %x, want %x", got.SmRpUI, in.SmRpUI)
		}
	})

	t.Run("Empty", func(t *testing.T) {
		in := &MtFsmResp{}
		data, err := in.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		got, err := ParseMtFsmResp(data)
		if err != nil {
			t.Fatalf("ParseMtFsmResp: %v", err)
		}
		if got.SmRpUI != nil {
			t.Errorf("SmRpUI should be nil, got %x", got.SmRpUI)
		}
	})
}

func TestMtFsmDeliveryTimerValidation(t *testing.T) {
	// Parse a known valid MT-FSM to get a valid TPDU.
	knownHex := "3077800832140080803138f684069169318488880463040b916971101174f40000422182612464805bd2e2b1252d467ff6de6c47efd96eb6a1d056cb0d69b49a10269c098537586e96931965b260d15613da72c29b91261bde72c6a1ad2623d682b5996d58331271375a0d1733eee4bd98ec768bd966b41c0d"
	knownBytes, err := hex.DecodeString(knownHex)
	if err != nil {
		t.Fatalf("hex decode: %v", err)
	}
	base, err := ParseMtFsm(knownBytes)
	if err != nil {
		t.Fatalf("ParseMtFsm: %v", err)
	}

	tests := []struct {
		name  string
		timer int
	}{
		{"TooLow", 29},
		{"TooHigh", 601},
		{"Zero", 0},
		{"Negative", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MtFsm{
				IMSI:                   base.IMSI,
				ServiceCentreAddressOA: base.ServiceCentreAddressOA,
				SCAOANature:            base.SCAOANature,
				SCAOAPlan:              base.SCAOAPlan,
				TPDU:                   base.TPDU,
				SmDeliveryTimer:        &tt.timer,
			}
			_, err := m.Marshal()
			if err == nil {
				t.Fatal("expected error for invalid SmDeliveryTimer")
			}
			if !errors.Is(err, ErrMtFsmInvalidDeliveryTimer) {
				t.Errorf("expected ErrMtFsmInvalidDeliveryTimer, got: %v", err)
			}
		})
	}
}

func TestAdditionalNumberChoiceValidation(t *testing.T) {
	t.Run("BothSet", func(t *testing.T) {
		in := &SriSmResp{
			IMSI: "204080012345678",
			LocationInfoWithLMSI: LocationInfoWithLMSI{
				NetworkNodeNumber: "31612345678",
				AdditionalNumber: &AdditionalNumber{
					MscNumber:  "31201111111",
					SgsnNumber: "31699999999",
				},
			},
		}
		_, err := in.Marshal()
		if err == nil {
			t.Fatal("expected error for both-set AdditionalNumber CHOICE")
		}
		if !errors.Is(err, ErrSriChoiceMultipleAlternatives) {
			t.Errorf("expected ErrSriChoiceMultipleAlternatives, got: %v", err)
		}
	})

	t.Run("NoneSet", func(t *testing.T) {
		in := &SriSmResp{
			IMSI: "204080012345678",
			LocationInfoWithLMSI: LocationInfoWithLMSI{
				NetworkNodeNumber: "31612345678",
				AdditionalNumber:  &AdditionalNumber{},
			},
		}
		_, err := in.Marshal()
		if err == nil {
			t.Fatal("expected error for empty AdditionalNumber CHOICE")
		}
		if !errors.Is(err, ErrSriChoiceNoAlternative) {
			t.Errorf("expected ErrSriChoiceNoAlternative, got: %v", err)
		}
	})
}

func TestMoFsmFullStressRoundTrip(t *testing.T) {
	// Parse a known valid MO-FSM to get a valid TPDU.
	knownHex := "302d84069122609098998206912260539128041b01510a912260716622000011d972180d4a82eee13928cc7ebbcb20"
	knownBytes, err := hex.DecodeString(knownHex)
	if err != nil {
		t.Fatalf("hex decode: %v", err)
	}
	base, err := ParseMoFsm(knownBytes)
	if err != nil {
		t.Fatalf("ParseMoFsm: %v", err)
	}

	outcome := SmDeliverySuccessfulTransfer
	in := &MoFsm{
		ServiceCentreAddressDA: base.ServiceCentreAddressDA,
		SCADANature:            base.SCADANature,
		SCADAPlan:              base.SCADAPlan,
		MSISDN:                 base.MSISDN,
		MSISDNNature:           base.MSISDNNature,
		MSISDNPlan:             base.MSISDNPlan,
		TPDU:                   base.TPDU,

		IMSI: "310260123456789",
		CorrelationID: &SriSmCorrelationID{
			HlrID:   HexBytes{0xAA, 0xBB},
			SipUriA: HexBytes{0xCC, 0xDD},
			SipUriB: HexBytes{0xEE, 0xFF},
		},
		SmDeliveryOutcome: &outcome,
	}

	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseMoFsm(data)
	if err != nil {
		t.Fatalf("ParseMoFsm: %v", err)
	}

	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("round-trip diff (-want +got):\n%s", diff)
	}
}

func TestMoFsmSmRpDaVariants(t *testing.T) {
	knownHex := "302d84069122609098998206912260539128041b01510a912260716622000011d972180d4a82eee13928cc7ebbcb20"
	knownBytes, err := hex.DecodeString(knownHex)
	if err != nil {
		t.Fatalf("hex decode: %v", err)
	}
	base, err := ParseMoFsm(knownBytes)
	if err != nil {
		t.Fatalf("ParseMoFsm: %v", err)
	}

	t.Run("IMSI", func(t *testing.T) {
		in := &MoFsm{
			SmRpDa:       &SmRpDa{IMSI: "310260123456789"},
			MSISDN:       base.MSISDN,
			MSISDNNature: base.MSISDNNature,
			MSISDNPlan:   base.MSISDNPlan,
			TPDU:         base.TPDU,
		}
		data, err := in.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		got, err := ParseMoFsm(data)
		if err != nil {
			t.Fatalf("ParseMoFsm: %v", err)
		}
		if got.SmRpDa == nil {
			t.Fatal("expected SmRpDa to be set")
		}
		if got.SmRpDa.IMSI != "310260123456789" {
			t.Errorf("IMSI: got %q, want %q", got.SmRpDa.IMSI, "310260123456789")
		}
	})

	t.Run("LMSI", func(t *testing.T) {
		in := &MoFsm{
			SmRpDa:       &SmRpDa{LMSI: HexBytes{0x01, 0x02, 0x03, 0x04}},
			MSISDN:       base.MSISDN,
			MSISDNNature: base.MSISDNNature,
			MSISDNPlan:   base.MSISDNPlan,
			TPDU:         base.TPDU,
		}
		data, err := in.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		got, err := ParseMoFsm(data)
		if err != nil {
			t.Fatalf("ParseMoFsm: %v", err)
		}
		if got.SmRpDa == nil {
			t.Fatal("expected SmRpDa to be set")
		}
		if !bytes.Equal(got.SmRpDa.LMSI, HexBytes{0x01, 0x02, 0x03, 0x04}) {
			t.Errorf("LMSI: got %x, want 01020304", got.SmRpDa.LMSI)
		}
	})

	t.Run("NoSmRpDa", func(t *testing.T) {
		in := &MoFsm{
			SmRpDa:       &SmRpDa{NoSmRpDa: true},
			MSISDN:       base.MSISDN,
			MSISDNNature: base.MSISDNNature,
			MSISDNPlan:   base.MSISDNPlan,
			TPDU:         base.TPDU,
		}
		data, err := in.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		got, err := ParseMoFsm(data)
		if err != nil {
			t.Fatalf("ParseMoFsm: %v", err)
		}
		if got.SmRpDa == nil {
			t.Fatal("expected SmRpDa to be set")
		}
		if !got.SmRpDa.NoSmRpDa {
			t.Error("expected NoSmRpDa to be true")
		}
	})

	t.Run("ServiceCentreAddressDA_via_SmRpDa", func(t *testing.T) {
		in := &MoFsm{
			SmRpDa:       &SmRpDa{ServiceCentreAddressDA: "31612345678"},
			MSISDN:       base.MSISDN,
			MSISDNNature: base.MSISDNNature,
			MSISDNPlan:   base.MSISDNPlan,
			TPDU:         base.TPDU,
		}
		data, err := in.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		got, err := ParseMoFsm(data)
		if err != nil {
			t.Fatalf("ParseMoFsm: %v", err)
		}
		if got.ServiceCentreAddressDA != "31612345678" {
			t.Errorf("ServiceCentreAddressDA: got %q, want %q", got.ServiceCentreAddressDA, "31612345678")
		}
		if got.SmRpDa != nil {
			t.Error("SmRpDa should be nil for serviceCentreAddressDA variant")
		}
	})
}

func TestMoFsmSmRpOaVariants(t *testing.T) {
	knownHex := "302d84069122609098998206912260539128041b01510a912260716622000011d972180d4a82eee13928cc7ebbcb20"
	knownBytes, err := hex.DecodeString(knownHex)
	if err != nil {
		t.Fatalf("hex decode: %v", err)
	}
	base, err := ParseMoFsm(knownBytes)
	if err != nil {
		t.Fatalf("ParseMoFsm: %v", err)
	}

	t.Run("ServiceCentreAddressOA", func(t *testing.T) {
		in := &MoFsm{
			ServiceCentreAddressDA: base.ServiceCentreAddressDA,
			SCADANature:            base.SCADANature,
			SCADAPlan:              base.SCADAPlan,
			SmRpOa:                 &SmRpOa{ServiceCentreAddressOA: "31699887766"},
			TPDU:                   base.TPDU,
		}
		data, err := in.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		got, err := ParseMoFsm(data)
		if err != nil {
			t.Fatalf("ParseMoFsm: %v", err)
		}
		if got.SmRpOa == nil {
			t.Fatal("expected SmRpOa to be set")
		}
		if got.SmRpOa.ServiceCentreAddressOA != "31699887766" {
			t.Errorf("ServiceCentreAddressOA: got %q, want %q", got.SmRpOa.ServiceCentreAddressOA, "31699887766")
		}
		if got.MSISDN != "" {
			t.Errorf("MSISDN should be empty, got %q", got.MSISDN)
		}
	})

	t.Run("NoSmRpOa", func(t *testing.T) {
		in := &MoFsm{
			ServiceCentreAddressDA: base.ServiceCentreAddressDA,
			SCADANature:            base.SCADANature,
			SCADAPlan:              base.SCADAPlan,
			SmRpOa:                 &SmRpOa{NoSmRpOa: true},
			TPDU:                   base.TPDU,
		}
		data, err := in.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		got, err := ParseMoFsm(data)
		if err != nil {
			t.Fatalf("ParseMoFsm: %v", err)
		}
		if got.SmRpOa == nil {
			t.Fatal("expected SmRpOa to be set")
		}
		if !got.SmRpOa.NoSmRpOa {
			t.Error("expected NoSmRpOa to be true")
		}
	})

	t.Run("MSISDN_via_SmRpOa", func(t *testing.T) {
		in := &MoFsm{
			ServiceCentreAddressDA: base.ServiceCentreAddressDA,
			SCADANature:            base.SCADANature,
			SCADAPlan:              base.SCADAPlan,
			SmRpOa:                 &SmRpOa{MSISDN: "31612345678"},
			TPDU:                   base.TPDU,
		}
		data, err := in.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		got, err := ParseMoFsm(data)
		if err != nil {
			t.Fatalf("ParseMoFsm: %v", err)
		}
		if got.MSISDN != "31612345678" {
			t.Errorf("MSISDN: got %q, want %q", got.MSISDN, "31612345678")
		}
		if got.SmRpOa != nil {
			t.Error("SmRpOa should be nil for msisdn variant")
		}
	})
}

func TestMoFsmRespRoundTrip(t *testing.T) {
	t.Run("WithSmRpUI", func(t *testing.T) {
		in := &MoFsmResp{SmRpUI: HexBytes{0x01, 0x02, 0x03}}
		data, err := in.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		got, err := ParseMoFsmResp(data)
		if err != nil {
			t.Fatalf("ParseMoFsmResp: %v", err)
		}
		if !bytes.Equal(in.SmRpUI, got.SmRpUI) {
			t.Errorf("SmRpUI: got %x, want %x", got.SmRpUI, in.SmRpUI)
		}
	})

	t.Run("Empty", func(t *testing.T) {
		in := &MoFsmResp{}
		data, err := in.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		got, err := ParseMoFsmResp(data)
		if err != nil {
			t.Fatalf("ParseMoFsmResp: %v", err)
		}
		if got.SmRpUI != nil {
			t.Errorf("SmRpUI should be nil, got %x", got.SmRpUI)
		}
	})
}

func TestMoFsmChoiceValidation(t *testing.T) {
	knownHex := "302d84069122609098998206912260539128041b01510a912260716622000011d972180d4a82eee13928cc7ebbcb20"
	knownBytes, err := hex.DecodeString(knownHex)
	if err != nil {
		t.Fatalf("hex decode: %v", err)
	}
	base, err := ParseMoFsm(knownBytes)
	if err != nil {
		t.Fatalf("ParseMoFsm: %v", err)
	}

	t.Run("EmptySmRpDa", func(t *testing.T) {
		in := &MoFsm{
			SmRpDa:       &SmRpDa{},
			MSISDN:       base.MSISDN,
			MSISDNNature: base.MSISDNNature,
			MSISDNPlan:   base.MSISDNPlan,
			TPDU:         base.TPDU,
		}
		_, err := in.Marshal()
		if err == nil {
			t.Fatal("expected error for empty SmRpDa CHOICE")
		}
		if !errors.Is(err, ErrMoFsmSmRpDaNoAlternative) {
			t.Errorf("expected ErrMoFsmSmRpDaNoAlternative, got: %v", err)
		}
	})

	t.Run("MultipleSmRpDa", func(t *testing.T) {
		in := &MoFsm{
			SmRpDa:       &SmRpDa{IMSI: "310260123456789", NoSmRpDa: true},
			MSISDN:       base.MSISDN,
			MSISDNNature: base.MSISDNNature,
			MSISDNPlan:   base.MSISDNPlan,
			TPDU:         base.TPDU,
		}
		_, err := in.Marshal()
		if err == nil {
			t.Fatal("expected error for multiple SmRpDa CHOICE alternatives")
		}
		if !errors.Is(err, ErrMoFsmSmRpDaMultipleAlternatives) {
			t.Errorf("expected ErrMoFsmSmRpDaMultipleAlternatives, got: %v", err)
		}
	})

	t.Run("EmptySmRpOa", func(t *testing.T) {
		in := &MoFsm{
			ServiceCentreAddressDA: base.ServiceCentreAddressDA,
			SCADANature:            base.SCADANature,
			SCADAPlan:              base.SCADAPlan,
			SmRpOa:                 &SmRpOa{},
			TPDU:                   base.TPDU,
		}
		_, err := in.Marshal()
		if err == nil {
			t.Fatal("expected error for empty SmRpOa CHOICE")
		}
		if !errors.Is(err, ErrMoFsmSmRpOaNoAlternative) {
			t.Errorf("expected ErrMoFsmSmRpOaNoAlternative, got: %v", err)
		}
	})

	t.Run("MultipleSmRpOa", func(t *testing.T) {
		in := &MoFsm{
			ServiceCentreAddressDA: base.ServiceCentreAddressDA,
			SCADANature:            base.SCADANature,
			SCADAPlan:              base.SCADAPlan,
			SmRpOa:                 &SmRpOa{MSISDN: "31612345678", NoSmRpOa: true},
			TPDU:                   base.TPDU,
		}
		_, err := in.Marshal()
		if err == nil {
			t.Fatal("expected error for multiple SmRpOa CHOICE alternatives")
		}
		if !errors.Is(err, ErrMoFsmSmRpOaMultipleAlternatives) {
			t.Errorf("expected ErrMoFsmSmRpOaMultipleAlternatives, got: %v", err)
		}
	})
}

func TestUpdateLocationFullStressRoundTrip(t *testing.T) {
	istVal := 1 // istCommandSupported
	in := &UpdateLocation{
		IMSI:      "310260123456789",
		MSCNumber: "31612345678",
		VLRNumber: "31699887766",

		VlrCapability: &VlrCapability{
			SupportedCamelPhases: &SupportedCamelPhases{
				Phase1: true,
				Phase2: true,
				Phase3: true,
				Phase4: true,
			},
			SupportedLCSCapabilitySets: &SupportedLCSCapabilitySets{
				LcsCapabilitySet1: true,
				LcsCapabilitySet2: true,
				LcsCapabilitySet3: true,
				LcsCapabilitySet4: true,
				LcsCapabilitySet5: true,
			},
			SolsaSupportIndicator: true,
			IstSupportIndicator:   &istVal,
			SuperChargerSupportedInServingNetworkEntity: &SuperChargerInfo{
				SendSubscriberData: true,
			},
			LongFTNSupported: true,
			OfferedCamel4CSIs: &OfferedCamel4CSIs{
				OCSI:            true,
				DCSI:            true,
				VTCSI:           true,
				TCSI:            true,
				MTSMSCSI:        true,
				MGCSI:           true,
				PsiEnhancements: true,
			},
			SupportedRATTypesIndicator: &SupportedRATTypes{
				UTRAN:          true,
				GERAN:          true,
				GAN:            true,
				IHSPAEvolution: true,
				EUTRAN:         true,
			},
			LongGroupIDSupported:         true,
			MtRoamingForwardingSupported: true,
			MsisdnLessOperationSupported: true,
			ResetIdsSupported:            true,
		},

		LMSI:                        HexBytes{0x01, 0x02, 0x03, 0x04},
		InformPreviousNetworkEntity: true,
		CsLCSNotSupportedByUE:       true,
		VGmlcAddress:                "192.168.1.1",
		AddInfo: &AddInfo{
			IMEISV:                   "3534567890123456",
			SkipSubscriberDataUpdate: true,
		},
		SkipSubscriberDataUpdate: true,
		RestorationIndicator:     true,
		EplmnList: []HexBytes{
			{0x13, 0x00, 0x26},
			{0x62, 0xf2, 0x20},
		},
		MmeDiameterAddress: &NetworkNodeDiameterAddress{
			DiameterName:  HexBytes("mme.example.com"),
			DiameterRealm: HexBytes("example.com"),
		},
	}

	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseUpdateLocation(data)
	if err != nil {
		t.Fatalf("ParseUpdateLocation: %v", err)
	}

	// Normalize natures/plans to defaults.
	in.MSCNature = address.NatureInternational
	in.MSCPlan = address.PlanISDN
	in.VLRNature = address.NatureInternational
	in.VLRPlan = address.PlanISDN

	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("round-trip diff (-want +got):\n%s", diff)
	}
}

func TestUpdateLocationResFullRoundTrip(t *testing.T) {
	in := &UpdateLocationRes{
		HLRNumber:            "31612345678",
		AddCapability:        true,
		PagingAreaCapability: true,
	}

	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseUpdateLocationRes(data)
	if err != nil {
		t.Fatalf("ParseUpdateLocationRes: %v", err)
	}

	// Normalize nature/plan.
	in.HLRNumberNature = address.NatureInternational
	in.HLRNumberPlan = address.PlanISDN

	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("round-trip diff (-want +got):\n%s", diff)
	}
}

func TestSuperChargerInfoRoundTrip(t *testing.T) {
	t.Run("SendSubscriberData", func(t *testing.T) {
		in := &UpdateLocation{
			IMSI:      "310260123456789",
			MSCNumber: "31612345678",
			VLRNumber: "31699887766",
			VlrCapability: &VlrCapability{
				SuperChargerSupportedInServingNetworkEntity: &SuperChargerInfo{
					SendSubscriberData: true,
				},
			},
		}
		data, err := in.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		got, err := ParseUpdateLocation(data)
		if err != nil {
			t.Fatalf("ParseUpdateLocation: %v", err)
		}
		if got.VlrCapability == nil || got.VlrCapability.SuperChargerSupportedInServingNetworkEntity == nil {
			t.Fatal("SuperChargerInfo is nil")
		}
		sc := got.VlrCapability.SuperChargerSupportedInServingNetworkEntity
		if !sc.SendSubscriberData {
			t.Error("SendSubscriberData should be true")
		}
		if len(sc.SubscriberDataStored) > 0 {
			t.Errorf("SubscriberDataStored should be nil, got %x", sc.SubscriberDataStored)
		}
	})

	t.Run("SubscriberDataStored", func(t *testing.T) {
		in := &UpdateLocation{
			IMSI:      "310260123456789",
			MSCNumber: "31612345678",
			VLRNumber: "31699887766",
			VlrCapability: &VlrCapability{
				SuperChargerSupportedInServingNetworkEntity: &SuperChargerInfo{
					SubscriberDataStored: HexBytes{0x01, 0x02, 0x03},
				},
			},
		}
		data, err := in.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		got, err := ParseUpdateLocation(data)
		if err != nil {
			t.Fatalf("ParseUpdateLocation: %v", err)
		}
		if got.VlrCapability == nil || got.VlrCapability.SuperChargerSupportedInServingNetworkEntity == nil {
			t.Fatal("SuperChargerInfo is nil")
		}
		sc := got.VlrCapability.SuperChargerSupportedInServingNetworkEntity
		if sc.SendSubscriberData {
			t.Error("SendSubscriberData should be false")
		}
		if !bytes.Equal(sc.SubscriberDataStored, HexBytes{0x01, 0x02, 0x03}) {
			t.Errorf("SubscriberDataStored: got %x, want 010203", sc.SubscriberDataStored)
		}
	})

	t.Run("BothSet", func(t *testing.T) {
		in := &UpdateLocation{
			IMSI:      "310260123456789",
			MSCNumber: "31612345678",
			VLRNumber: "31699887766",
			VlrCapability: &VlrCapability{
				SuperChargerSupportedInServingNetworkEntity: &SuperChargerInfo{
					SendSubscriberData:   true,
					SubscriberDataStored: HexBytes{0x01},
				},
			},
		}
		_, err := in.Marshal()
		if err == nil {
			t.Fatal("expected error for both-set SuperChargerInfo CHOICE")
		}
		if !errors.Is(err, ErrSuperChargerInfoMultipleAlternatives) {
			t.Errorf("expected ErrSuperChargerInfoMultipleAlternatives, got: %v", err)
		}
	})

	t.Run("NoneSet", func(t *testing.T) {
		in := &UpdateLocation{
			IMSI:      "310260123456789",
			MSCNumber: "31612345678",
			VLRNumber: "31699887766",
			VlrCapability: &VlrCapability{
				SuperChargerSupportedInServingNetworkEntity: &SuperChargerInfo{},
			},
		}
		_, err := in.Marshal()
		if err == nil {
			t.Fatal("expected error for empty SuperChargerInfo CHOICE")
		}
		if !errors.Is(err, ErrSuperChargerInfoNoAlternative) {
			t.Errorf("expected ErrSuperChargerInfoNoAlternative, got: %v", err)
		}
	})
}

func TestSupportedRATTypesRoundTrip(t *testing.T) {
	// Exhaustive 32-case coverage for 5 bits.
	for mask := 0; mask < 32; mask++ {
		rats := &SupportedRATTypes{
			UTRAN:          mask&1 != 0,
			GERAN:          mask&2 != 0,
			GAN:            mask&4 != 0,
			IHSPAEvolution: mask&8 != 0,
			EUTRAN:         mask&16 != 0,
		}

		// Skip the all-zero case since SupportedRATTypesIndicator would be nil on parse.
		if mask == 0 {
			continue
		}

		in := &UpdateLocation{
			IMSI:      "310260123456789",
			MSCNumber: "31612345678",
			VLRNumber: "31699887766",
			VlrCapability: &VlrCapability{
				SupportedRATTypesIndicator: rats,
			},
		}
		data, err := in.Marshal()
		if err != nil {
			t.Fatalf("mask=%d Marshal: %v", mask, err)
		}
		got, err := ParseUpdateLocation(data)
		if err != nil {
			t.Fatalf("mask=%d ParseUpdateLocation: %v", mask, err)
		}
		if got.VlrCapability == nil || got.VlrCapability.SupportedRATTypesIndicator == nil {
			t.Fatalf("mask=%d SupportedRATTypesIndicator is nil", mask)
		}
		r := got.VlrCapability.SupportedRATTypesIndicator
		if r.UTRAN != rats.UTRAN {
			t.Errorf("mask=%d UTRAN: got %v, want %v", mask, r.UTRAN, rats.UTRAN)
		}
		if r.GERAN != rats.GERAN {
			t.Errorf("mask=%d GERAN: got %v, want %v", mask, r.GERAN, rats.GERAN)
		}
		if r.GAN != rats.GAN {
			t.Errorf("mask=%d GAN: got %v, want %v", mask, r.GAN, rats.GAN)
		}
		if r.IHSPAEvolution != rats.IHSPAEvolution {
			t.Errorf("mask=%d IHSPAEvolution: got %v, want %v", mask, r.IHSPAEvolution, rats.IHSPAEvolution)
		}
		if r.EUTRAN != rats.EUTRAN {
			t.Errorf("mask=%d EUTRAN: got %v, want %v", mask, r.EUTRAN, rats.EUTRAN)
		}
	}
}

func TestUpdateGprsLocationFullStressRoundTrip(t *testing.T) {
	truthy := true
	usedRat := UsedRatEUTRAN
	ueSrvcc := UeSrvccSupported
	smsReg := SmsRegistrationRequired
	ctxID := 31

	in := &UpdateGprsLocation{
		IMSI:        "310260311111111",
		SGSNNumber:  "31631000001",
		SGSNAddress: "192.168.31.1",

		SGSNCapability: &SGSNCapability{
			SolsaSupportIndicator: true,
			SuperChargerSupportedInServingNetworkEntity: &SuperChargerInfo{
				SendSubscriberData: true,
			},
			GprsEnhancementsSupportIndicator: true,
			SupportedCamelPhases: &SupportedCamelPhases{
				Phase1: true, Phase2: true, Phase3: true, Phase4: true,
			},
			SupportedLCSCapabilitySets: &SupportedLCSCapabilitySets{
				LcsCapabilitySet1: true, LcsCapabilitySet2: true,
				LcsCapabilitySet3: true, LcsCapabilitySet4: true, LcsCapabilitySet5: true,
			},
			OfferedCamel4CSIs: &OfferedCamel4CSIs{
				OCSI: true, DCSI: true, VTCSI: true, TCSI: true,
				MTSMSCSI: true, MGCSI: true, PsiEnhancements: true,
			},
			SmsCallBarringSupportIndicator: true,
			SupportedRATTypesIndicator: &SupportedRATTypes{
				UTRAN: true, GERAN: true, GAN: true, IHSPAEvolution: true, EUTRAN: true,
			},
			SupportedFeatures:                                  HexBytes{0xA0},
			SupportedFeaturesBits:                              4,
			TAdsDataRetrieval:                                  true,
			HomogeneousSupportOfIMSVoiceOverPSSessions:         &truthy,
			CancellationTypeInitialAttach:                      true,
			MsisdnLessOperationSupported:                       true,
			UpdateofHomogeneousSupportOfIMSVoiceOverPSSessions: true,
			ResetIdsSupported:                                  true,
			ExtSupportedFeatures:                               HexBytes{0x80},
			ExtSupportedFeaturesBits:                           2,
		},

		InformPreviousNetworkEntity: true,
		PsLCSNotSupportedByUE:       true,
		VGmlcAddress:                "192.168.31.77",
		AddInfo: &AddInfo{
			IMEISV:                   "3534567890123456",
			SkipSubscriberDataUpdate: true,
		},
		EpsInfo: &EpsInfo{
			PdnGwUpdate: &PdnGwUpdate{
				APN: HexBytes{0x03, 'a', 'p', 'n'},
				PdnGwIdentity: &PdnGwIdentity{
					IPv4Address: HexBytes{0xC0, 0xA8, 0x1F, 0x01},
					Name:        HexBytes("pgw.example.com"),
				},
				ContextID: &ctxID,
			},
		},
		ServingNodeTypeIndicator:      true,
		SkipSubscriberDataUpdate:      true,
		UsedRatType:                   &usedRat,
		GprsSubscriptionDataNotNeeded: true,
		NodeTypeIndicator:             true,
		AreaRestricted:                true,
		UeReachableIndicator:          true,
		EpsSubscriptionDataNotNeeded:  true,
		UeSrvccCapability:             &ueSrvcc,
		EplmnList: []HexBytes{
			{0x13, 0x00, 0x26},
			{0x62, 0xf2, 0x20},
		},
		MmeNumberForMTSMS:              "31699900099",
		SmsRegisterRequest:             &smsReg,
		SmsOnly:                        true,
		SgsnName:                       HexBytes("sgsn.example.com"),
		SgsnRealm:                      HexBytes("example.com"),
		LgdSupportIndicator:            true,
		RemovalofMMERegistrationforSMS: true,
		AdjacentPLMNList: []HexBytes{
			{0x21, 0x43, 0x65},
		},
	}

	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseUpdateGprsLocation(data)
	if err != nil {
		t.Fatalf("ParseUpdateGprsLocation: %v", err)
	}

	// Normalize natures/plans to defaults.
	in.SGSNNature = address.NatureInternational
	in.SGSNPlan = address.PlanISDN
	in.MmeNumberForMTSMSNature = address.NatureInternational
	in.MmeNumberForMTSMSPlan = address.PlanISDN

	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("round-trip diff (-want +got):\n%s", diff)
	}
}

func TestUpdateGprsLocationEpsInfoIsr(t *testing.T) {
	in := &UpdateGprsLocation{
		IMSI:        "310260311111111",
		SGSNNumber:  "31631000001",
		SGSNAddress: "192.168.31.1",
		EpsInfo: &EpsInfo{
			IsrInformation:     HexBytes{0xC0},
			IsrInformationBits: 3,
		},
	}

	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseUpdateGprsLocation(data)
	if err != nil {
		t.Fatalf("ParseUpdateGprsLocation: %v", err)
	}
	if got.EpsInfo == nil {
		t.Fatal("EpsInfo is nil")
	}
	if got.EpsInfo.PdnGwUpdate != nil {
		t.Error("PdnGwUpdate should be nil in IsrInformation alternative")
	}
	if got.EpsInfo.IsrInformationBits != in.EpsInfo.IsrInformationBits {
		t.Errorf("IsrInformationBits: got %d want %d", got.EpsInfo.IsrInformationBits, in.EpsInfo.IsrInformationBits)
	}
	if !bytes.Equal(got.EpsInfo.IsrInformation, in.EpsInfo.IsrInformation) {
		t.Errorf("IsrInformation: got %x want %x", got.EpsInfo.IsrInformation, in.EpsInfo.IsrInformation)
	}
}

func TestUpdateGprsLocationResFullRoundTrip(t *testing.T) {
	in := &UpdateGprsLocationRes{
		HLRNumber:                  "31612345678",
		AddCapability:              true,
		SgsnMmeSeparationSupported: true,
		MmeRegisteredforSMS:        true,
	}

	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseUpdateGprsLocationRes(data)
	if err != nil {
		t.Fatalf("ParseUpdateGprsLocationRes: %v", err)
	}

	in.HLRNumberNature = address.NatureInternational
	in.HLRNumberPlan = address.PlanISDN

	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("round-trip diff (-want +got):\n%s", diff)
	}
}

func TestEpsInfoChoiceValidation(t *testing.T) {
	t.Run("NoneSet", func(t *testing.T) {
		in := &UpdateGprsLocation{
			IMSI:        "310260311111111",
			SGSNNumber:  "31631000001",
			SGSNAddress: "192.168.31.1",
			EpsInfo:     &EpsInfo{},
		}
		_, err := in.Marshal()
		if err == nil {
			t.Fatal("expected error for empty EpsInfo CHOICE")
		}
		if !errors.Is(err, ErrSriChoiceNoAlternative) {
			t.Errorf("expected ErrSriChoiceNoAlternative, got: %v", err)
		}
	})

	t.Run("BothSet", func(t *testing.T) {
		ctxID := 1
		in := &UpdateGprsLocation{
			IMSI:        "310260311111111",
			SGSNNumber:  "31631000001",
			SGSNAddress: "192.168.31.1",
			EpsInfo: &EpsInfo{
				PdnGwUpdate: &PdnGwUpdate{
					ContextID: &ctxID,
				},
				IsrInformation:     HexBytes{0x80},
				IsrInformationBits: 1,
			},
		}
		_, err := in.Marshal()
		if err == nil {
			t.Fatal("expected error for both-set EpsInfo CHOICE")
		}
		if !errors.Is(err, ErrSriChoiceMultipleAlternatives) {
			t.Errorf("expected ErrSriChoiceMultipleAlternatives, got: %v", err)
		}
	})
}

// --- InformServiceCentre (opCode 63) tests ---

func TestInformServiceCentreMinimalRoundTrip(t *testing.T) {
	in := &InformServiceCentre{}

	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseInformServiceCentre(data)
	if err != nil {
		t.Fatalf("ParseInformServiceCentre: %v", err)
	}

	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("round-trip diff (-want +got):\n%s", diff)
	}
}

func TestInformServiceCentreFullStressRoundTrip(t *testing.T) {
	abs := 42
	addAbs := 100
	smsf3gpp := 200
	smsfNon3gpp := 255

	in := &InformServiceCentre{
		StoredMSISDN: "31612345678",
		MwStatus: &MwStatusFlags{
			SCAddressNotIncluded: true,
			MnrfSet:              true,
			McefSet:              true,
			MnrgSet:              true,
			Mnr5gSet:             true,
			Mnr5gn3gSet:          true,
		},
		AbsentSubscriberDiagnosticSM:            &abs,
		AdditionalAbsentSubscriberDiagnosticSM:  &addAbs,
		Smsf3gppAbsentSubscriberDiagnosticSM:    &smsf3gpp,
		SmsfNon3gppAbsentSubscriberDiagnosticSM: &smsfNon3gpp,
	}

	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseInformServiceCentre(data)
	if err != nil {
		t.Fatalf("ParseInformServiceCentre: %v", err)
	}

	// Normalize defaults.
	in.StoredMSISDNNature = address.NatureInternational
	in.StoredMSISDNPlan = address.PlanISDN

	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("round-trip diff (-want +got):\n%s", diff)
	}
}

func TestMwStatusBitStringRoundTrip(t *testing.T) {
	// Exhaustive 64-case coverage (6 bits = 2^6 combinations).
	for combo := 0; combo < 64; combo++ {
		m := &MwStatusFlags{
			SCAddressNotIncluded: combo&(1<<0) != 0,
			MnrfSet:              combo&(1<<1) != 0,
			McefSet:              combo&(1<<2) != 0,
			MnrgSet:              combo&(1<<3) != 0,
			Mnr5gSet:             combo&(1<<4) != 0,
			Mnr5gn3gSet:          combo&(1<<5) != 0,
		}

		bs := convertMwStatusToBitString(m)
		if bs.BitLength != 6 {
			t.Errorf("combo=%06b: BitLength=%d want 6", combo, bs.BitLength)
		}

		got := convertBitStringToMwStatus(bs)
		if diff := cmp.Diff(m, got); diff != "" {
			t.Errorf("combo=%06b: round-trip diff (-want +got):\n%s", combo, diff)
		}

		// Also verify a full round-trip through the full InformServiceCentre
		// Marshal/Parse pipeline.
		in := &InformServiceCentre{MwStatus: m}
		data, err := in.Marshal()
		if err != nil {
			t.Fatalf("combo=%06b: Marshal: %v", combo, err)
		}
		parsed, err := ParseInformServiceCentre(data)
		if err != nil {
			t.Fatalf("combo=%06b: Parse: %v", combo, err)
		}
		if parsed.MwStatus == nil {
			t.Fatalf("combo=%06b: parsed MwStatus is nil", combo)
		}
		if diff := cmp.Diff(m, parsed.MwStatus); diff != "" {
			t.Errorf("combo=%06b: pipeline diff (-want +got):\n%s", combo, diff)
		}
	}
}

func TestInformServiceCentreValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		mutator func(i *InformServiceCentre)
	}{
		{
			name: "AbsentSubscriberDiagnosticSM_negative",
			mutator: func(i *InformServiceCentre) {
				v := -1
				i.AbsentSubscriberDiagnosticSM = &v
			},
		},
		{
			name: "AbsentSubscriberDiagnosticSM_overflow",
			mutator: func(i *InformServiceCentre) {
				v := 256
				i.AbsentSubscriberDiagnosticSM = &v
			},
		},
		{
			name: "AdditionalAbsentSubscriberDiagnosticSM_overflow",
			mutator: func(i *InformServiceCentre) {
				v := 1000
				i.AdditionalAbsentSubscriberDiagnosticSM = &v
			},
		},
		{
			name: "Smsf3gppAbsentSubscriberDiagnosticSM_negative",
			mutator: func(i *InformServiceCentre) {
				v := -100
				i.Smsf3gppAbsentSubscriberDiagnosticSM = &v
			},
		},
		{
			name: "SmsfNon3gppAbsentSubscriberDiagnosticSM_overflow",
			mutator: func(i *InformServiceCentre) {
				v := math.MaxInt32
				i.SmsfNon3gppAbsentSubscriberDiagnosticSM = &v
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			in := &InformServiceCentre{}
			tc.mutator(in)

			_, err := in.Marshal()
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if !errors.Is(err, ErrIscInvalidAbsentSubscriberDiagnosticSM) {
				t.Errorf("expected ErrIscInvalidAbsentSubscriberDiagnosticSM, got: %v", err)
			}
		})
	}
}

func TestAlertServiceCentreMandatoryRoundTrip(t *testing.T) {
	in := &AlertServiceCentre{
		MSISDN:               "31612345678",
		ServiceCentreAddress: "31611111111",
	}

	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseAlertServiceCentre(data)
	if err != nil {
		t.Fatalf("ParseAlertServiceCentre: %v", err)
	}

	// Normalize default natures/plans.
	in.MSISDNNature = address.NatureInternational
	in.MSISDNPlan = address.PlanISDN
	in.SCANature = address.NatureInternational
	in.SCAPlan = address.PlanISDN

	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("round-trip diff (-want +got):\n%s", diff)
	}
}

func TestAlertServiceCentreFullStressRoundTrip(t *testing.T) {
	event := SmsGmscAlertMsUnderNewServingNode

	in := &AlertServiceCentre{
		MSISDN:               "31612345678",
		ServiceCentreAddress: "31611111111",
		IMSI:                 "204080012345678",
		CorrelationID: &SriSmCorrelationID{
			HlrID:   HexBytes{0xAA, 0xBB},
			SipUriA: HexBytes{0xCC, 0xDD},
			SipUriB: HexBytes{0xEE, 0xFF},
		},
		MaximumUeAvailabilityTime: HexBytes{0x01, 0x02, 0x03, 0x04},
		SmsGmscAlertEvent:         &event,
		SmsGmscDiameterAddress: &NetworkNodeDiameterAddress{
			DiameterName:  HexBytes("gmsc.example.com"),
			DiameterRealm: HexBytes("example.com"),
		},
		NewSGSNNumber: "31622222222",
		NewSGSNDiameterAddress: &NetworkNodeDiameterAddress{
			DiameterName:  HexBytes("sgsn.example.com"),
			DiameterRealm: HexBytes("example.com"),
		},
		NewMMENumber: "31633333333",
		NewMMEDiameterAddress: &NetworkNodeDiameterAddress{
			DiameterName:  HexBytes("mme.example.com"),
			DiameterRealm: HexBytes("example.com"),
		},
		NewMSCNumber: "31644444444",
	}

	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseAlertServiceCentre(data)
	if err != nil {
		t.Fatalf("ParseAlertServiceCentre: %v", err)
	}

	// Normalize default natures/plans for all address fields.
	in.MSISDNNature = address.NatureInternational
	in.MSISDNPlan = address.PlanISDN
	in.SCANature = address.NatureInternational
	in.SCAPlan = address.PlanISDN
	in.NewSGSNNumberNature = address.NatureInternational
	in.NewSGSNNumberPlan = address.PlanISDN
	in.NewMMENumberNature = address.NatureInternational
	in.NewMMENumberPlan = address.PlanISDN
	in.NewMSCNumberNature = address.NatureInternational
	in.NewMSCNumberPlan = address.PlanISDN

	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("round-trip diff (-want +got):\n%s", diff)
	}
}

func TestAlertServiceCentreValidationErrors(t *testing.T) {
	t.Run("MissingMSISDN", func(t *testing.T) {
		in := &AlertServiceCentre{
			ServiceCentreAddress: "31611111111",
		}
		_, err := in.Marshal()
		if err == nil {
			t.Fatal("expected error for missing MSISDN")
		}
		if !errors.Is(err, ErrAscMissingMSISDN) {
			t.Errorf("expected ErrAscMissingMSISDN, got: %v", err)
		}
	})

	t.Run("MissingServiceCentreAddress", func(t *testing.T) {
		in := &AlertServiceCentre{
			MSISDN: "31612345678",
		}
		_, err := in.Marshal()
		if err == nil {
			t.Fatal("expected error for missing ServiceCentreAddress")
		}
		if !errors.Is(err, ErrAscMissingServiceCentreAddress) {
			t.Errorf("expected ErrAscMissingServiceCentreAddress, got: %v", err)
		}
	})

	t.Run("InvalidSmsGmscAlertEvent", func(t *testing.T) {
		invalid := SmsGmscAlertEvent(42)
		in := &AlertServiceCentre{
			MSISDN:               "31612345678",
			ServiceCentreAddress: "31611111111",
			SmsGmscAlertEvent:    &invalid,
		}
		_, err := in.Marshal()
		if err == nil {
			t.Fatal("expected error for invalid SmsGmscAlertEvent")
		}
		if !errors.Is(err, ErrAscInvalidSmsGmscAlertEvent) {
			t.Errorf("expected ErrAscInvalidSmsGmscAlertEvent, got: %v", err)
		}
	})
}

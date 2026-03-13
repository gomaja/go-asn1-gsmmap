package gsmmap

import (
	"encoding/hex"
	"testing"
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
				GeographicalInformation:  []byte{0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70, 0x80},
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
	if len(loc.GeographicalInformation) != 8 {
		t.Errorf("GeographicalInformation length: got %d, want 8", len(loc.GeographicalInformation))
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
	if len(loc.CellGlobalId) != len(cellId) {
		t.Fatalf("CellGlobalId length: got %d, want %d", len(loc.CellGlobalId), len(cellId))
	}
	for i := range cellId {
		if loc.CellGlobalId[i] != cellId[i] {
			t.Errorf("CellGlobalId[%d]: got %02x, want %02x", i, loc.CellGlobalId[i], cellId[i])
		}
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
	if len(loc.LAI) != len(lai) {
		t.Fatalf("LAI length: got %d, want %d", len(loc.LAI), len(lai))
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
	if len(eps.EUtranCellGlobalIdentity) != 7 {
		t.Errorf("E-UTRAN CGI length: got %d, want 7", len(eps.EUtranCellGlobalIdentity))
	}
	if len(eps.TrackingAreaIdentity) != 5 {
		t.Errorf("TAI length: got %d, want 5", len(eps.TrackingAreaIdentity))
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
	if len(si.MsClassmark2) != 3 {
		t.Errorf("MsClassmark2 length: got %d, want 3", len(si.MsClassmark2))
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

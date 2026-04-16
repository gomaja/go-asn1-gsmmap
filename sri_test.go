// sri_test.go
package gsmmap

import (
	"bytes"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gomaja/go-asn1-gsmmap/address"
)

func TestSriTypesCompile(t *testing.T) {
	var _ Sri
	var _ SriResp
	var _ InterrogationType
	var _ ForwardingReason
	var _ NumberPortabilityStatus
	var _ UnavailabilityCause
	var _ SuppressMTSSFlags
	var _ AllowedServicesFlags
	var _ CugCheckInfo
	var _ ExtBasicServiceCode
	var _ ExternalSignalInfo
	var _ ExtExternalSignalInfo
	var _ SriCamelInfo
	var _ ExtendedRoutingInfo
	var _ RoutingInfo
	var _ ForwardingData
	var _ CamelRoutingInfo
	var _ GmscCamelSubscriptionInfo
	var _ CcbsIndicators
	var _ NaeaPreferredCI
	var _ OfferedCamel4CSIs
	var _ SsCode
}

func TestSriSentinelErrorsExist(t *testing.T) {
	errs := []error{
		ErrSriMissingMSISDN,
		ErrSriMissingGmsc,
		ErrSriInvalidInterrogationType,
		ErrSriInvalidNumberOfForwarding,
		ErrSriInvalidOrCapability,
		ErrSriInvalidCallReferenceNumber,
		ErrSriChoiceMultipleAlternatives,
		ErrSriChoiceNoAlternative,
	}
	for _, e := range errs {
		if e == nil {
			t.Errorf("sentinel error is nil")
		}
	}
}

func TestAllowedServicesBitStringRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   AllowedServicesFlags
	}{
		{"empty", AllowedServicesFlags{}},
		{"first", AllowedServicesFlags{FirstServiceAllowed: true}},
		{"second", AllowedServicesFlags{SecondServiceAllowed: true}},
		{"both", AllowedServicesFlags{FirstServiceAllowed: true, SecondServiceAllowed: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bs := convertAllowedServicesToBitString(&tc.in)
			got := convertBitStringToAllowedServices(bs)
			if got.FirstServiceAllowed != tc.in.FirstServiceAllowed ||
				got.SecondServiceAllowed != tc.in.SecondServiceAllowed {
				t.Errorf("got %+v, want %+v (via %+v)", got, tc.in, bs)
			}
		})
	}
}

func TestSuppressMTSSBitStringRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   SuppressMTSSFlags
	}{
		{"empty", SuppressMTSSFlags{}},
		{"cug", SuppressMTSSFlags{SuppressCUG: true}},
		{"ccbs", SuppressMTSSFlags{SuppressCCBS: true}},
		{"both", SuppressMTSSFlags{SuppressCUG: true, SuppressCCBS: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bs := convertSuppressMTSSToBitString(&tc.in)
			got := convertBitStringToSuppressMTSS(bs)
			if got.SuppressCUG != tc.in.SuppressCUG || got.SuppressCCBS != tc.in.SuppressCCBS {
				t.Errorf("got %+v, want %+v (via %+v)", got, tc.in, bs)
			}
		})
	}
}

func TestOfferedCamel4CSIsBitStringRoundTrip(t *testing.T) {
	for input := 0; input < 128; input++ {
		in := &OfferedCamel4CSIs{
			OCSI:            input&(1<<6) != 0,
			DCSI:            input&(1<<5) != 0,
			VTCSI:           input&(1<<4) != 0,
			TCSI:            input&(1<<3) != 0,
			MTSMSCSI:        input&(1<<2) != 0,
			MGCSI:           input&(1<<1) != 0,
			PsiEnhancements: input&(1<<0) != 0,
		}
		bs := convertOfferedCamel4CSIsToBitString(in)
		got := convertBitStringToOfferedCamel4CSIs(bs)
		if *got != *in {
			t.Errorf("input=%07b: got %+v, want %+v (via %+v)", input, got, in, bs)
		}
	}
}

func TestForwardingDataRoundTrip(t *testing.T) {
	in := &ForwardingData{
		ForwardedToNumber:       "972501234567",
		ForwardedToNumberNature: 4,
		ForwardedToNumberPlan:   1,
		ForwardingOptions:       HexBytes{0x05},
	}
	wire, err := convertForwardingDataToWire(in)
	if err != nil {
		t.Fatalf("to wire: %v", err)
	}
	got, err := convertWireToForwardingData(wire)
	if err != nil {
		t.Fatalf("from wire: %v", err)
	}
	if got.ForwardedToNumber != in.ForwardedToNumber {
		t.Errorf("ForwardedToNumber: got %q want %q", got.ForwardedToNumber, in.ForwardedToNumber)
	}
	if !bytes.Equal(got.ForwardingOptions, in.ForwardingOptions) {
		t.Errorf("ForwardingOptions: got %x want %x", got.ForwardingOptions, in.ForwardingOptions)
	}
}

func TestCcbsIndicatorsRoundTrip(t *testing.T) {
	in := &CcbsIndicators{CcbsPossible: true, KeepCCBSCallIndicator: true}
	wire := convertCcbsIndicatorsToWire(in)
	got := convertWireToCcbsIndicators(wire)
	if *got != *in {
		t.Errorf("got %+v want %+v", got, in)
	}
}

func TestCugCheckInfoRoundTrip(t *testing.T) {
	in := &CugCheckInfo{CugInterlock: HexBytes{0x01, 0x02, 0x03, 0x04}, CugOutgoingAccess: true}
	wire := convertCugCheckInfoToWire(in)
	got := convertWireToCugCheckInfo(wire)
	if !bytes.Equal(got.CugInterlock, in.CugInterlock) || got.CugOutgoingAccess != in.CugOutgoingAccess {
		t.Errorf("got %+v want %+v", got, in)
	}
}

func TestExtBasicServiceCodeRoundTrip(t *testing.T) {
	bearer := &ExtBasicServiceCode{ExtBearerService: HexBytes{0x10}}
	wire, err := convertExtBasicServiceCodeToWire(bearer)
	if err != nil {
		t.Fatal(err)
	}
	got, err := convertWireToExtBasicServiceCode(wire)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.ExtBearerService, bearer.ExtBearerService) || len(got.ExtTeleservice) != 0 {
		t.Errorf("bearer round-trip failed: got %+v want %+v", got, bearer)
	}

	tele := &ExtBasicServiceCode{ExtTeleservice: HexBytes{0x21}}
	wire2, err := convertExtBasicServiceCodeToWire(tele)
	if err != nil {
		t.Fatal(err)
	}
	got2, err := convertWireToExtBasicServiceCode(wire2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got2.ExtTeleservice, tele.ExtTeleservice) {
		t.Errorf("teleservice round-trip failed: got %+v want %+v", got2, tele)
	}
}

func TestExtBasicServiceCodeChoiceValidation(t *testing.T) {
	if _, err := convertExtBasicServiceCodeToWire(&ExtBasicServiceCode{}); err == nil {
		t.Errorf("expected ErrSriChoiceNoAlternative for empty ExtBasicServiceCode")
	}
	both := &ExtBasicServiceCode{ExtBearerService: HexBytes{0x10}, ExtTeleservice: HexBytes{0x21}}
	if _, err := convertExtBasicServiceCodeToWire(both); err == nil {
		t.Errorf("expected ErrSriChoiceMultipleAlternatives for both set")
	}
}

func TestRoutingInfoRoundTrip(t *testing.T) {
	rn := &RoutingInfo{RoamingNumber: "972501111111"}
	wire, err := convertRoutingInfoToWire(rn)
	if err != nil {
		t.Fatal(err)
	}
	got, err := convertWireToRoutingInfo(wire)
	if err != nil {
		t.Fatal(err)
	}
	if got.RoamingNumber != rn.RoamingNumber {
		t.Errorf("RoamingNumber round-trip: got %q want %q", got.RoamingNumber, rn.RoamingNumber)
	}

	fd := &RoutingInfo{ForwardingData: &ForwardingData{ForwardedToNumber: "972502222222"}}
	wire2, err := convertRoutingInfoToWire(fd)
	if err != nil {
		t.Fatal(err)
	}
	got2, err := convertWireToRoutingInfo(wire2)
	if err != nil {
		t.Fatal(err)
	}
	if got2.ForwardingData == nil || got2.ForwardingData.ForwardedToNumber != "972502222222" {
		t.Errorf("ForwardingData round-trip failed: got %+v", got2)
	}
}

func TestExtendedRoutingInfoChoiceValidation(t *testing.T) {
	if _, err := convertExtendedRoutingInfoToWire(&ExtendedRoutingInfo{}); err == nil {
		t.Errorf("expected ErrSriChoiceNoAlternative for empty ExtendedRoutingInfo")
	}
}

func TestExternalSignalInfoRoundTrip(t *testing.T) {
	in := &ExternalSignalInfo{ProtocolID: 0, SignalInfo: HexBytes{0xDE, 0xAD}}
	w := convertExternalSignalInfoToWire(in)
	got := convertWireToExternalSignalInfo(w)
	if got.ProtocolID != in.ProtocolID || !bytes.Equal(got.SignalInfo, in.SignalInfo) {
		t.Errorf("got %+v want %+v", got, in)
	}
}

func TestSriCamelInfoRoundTrip(t *testing.T) {
	in := &SriCamelInfo{
		SupportedCamelPhases: SupportedCamelPhases{Phase1: true, Phase2: true, Phase3: true, Phase4: true},
		SuppressTCSI:         true,
	}
	w := convertSriCamelInfoToWire(in)
	got := convertWireToSriCamelInfo(w)
	if got.SuppressTCSI != in.SuppressTCSI {
		t.Errorf("SuppressTCSI: got %v want %v", got.SuppressTCSI, in.SuppressTCSI)
	}
	if got.SupportedCamelPhases != in.SupportedCamelPhases {
		t.Errorf("SupportedCamelPhases: got %+v want %+v", got.SupportedCamelPhases, in.SupportedCamelPhases)
	}
}

func TestSriMandatoryRoundTrip(t *testing.T) {
	in := &Sri{
		MSISDN:              "972501234567",
		InterrogationType:   InterrogationBasicCall,
		GmscOrGsmSCFAddress: "972531111111",
	}
	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseSri(data)
	if err != nil {
		t.Fatalf("ParseSri: %v", err)
	}
	if got.MSISDN != in.MSISDN {
		t.Errorf("MSISDN: got %q want %q", got.MSISDN, in.MSISDN)
	}
	if got.InterrogationType != in.InterrogationType {
		t.Errorf("InterrogationType: got %v want %v", got.InterrogationType, in.InterrogationType)
	}
	if got.GmscOrGsmSCFAddress != in.GmscOrGsmSCFAddress {
		t.Errorf("GmscOrGsmSCFAddress: got %q want %q", got.GmscOrGsmSCFAddress, in.GmscOrGsmSCFAddress)
	}
}

func TestSriValidationErrors(t *testing.T) {
	cases := []struct {
		name string
		in   *Sri
		err  error
	}{
		{"missing msisdn", &Sri{GmscOrGsmSCFAddress: "1"}, ErrSriMissingMSISDN},
		{"missing gmsc", &Sri{MSISDN: "1"}, ErrSriMissingGmsc},
		{"bad interrogation", &Sri{MSISDN: "1", GmscOrGsmSCFAddress: "1", InterrogationType: 7}, ErrSriInvalidInterrogationType},
		{"bad numberOfForwarding", &Sri{MSISDN: "1", GmscOrGsmSCFAddress: "1", NumberOfForwarding: intPtr(9)}, ErrSriInvalidNumberOfForwarding},
		{"bad orCapability", &Sri{MSISDN: "1", GmscOrGsmSCFAddress: "1", OrCapability: intPtr(200)}, ErrSriInvalidOrCapability},
		{"bad callref", &Sri{MSISDN: "1", GmscOrGsmSCFAddress: "1", CallReferenceNumber: make(HexBytes, 9)}, ErrSriInvalidCallReferenceNumber},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.in.Marshal()
			if !errors.Is(err, tc.err) {
				t.Errorf("got %v, want %v", err, tc.err)
			}
		})
	}
}

func TestSriRespMandatoryRoundTrip(t *testing.T) {
	mnp := MnpForeignNumberPortedIn
	in := &SriResp{
		IMSI:                    "425010123456789",
		ExtendedRoutingInfo:     &ExtendedRoutingInfo{RoutingInfo: &RoutingInfo{RoamingNumber: "972501111111"}},
		NumberPortabilityStatus: &mnp,
	}
	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseSriResp(data)
	if err != nil {
		t.Fatalf("ParseSriResp: %v", err)
	}
	if got.IMSI != in.IMSI {
		t.Errorf("IMSI: got %q want %q", got.IMSI, in.IMSI)
	}
	if got.ExtendedRoutingInfo == nil || got.ExtendedRoutingInfo.RoutingInfo == nil ||
		got.ExtendedRoutingInfo.RoutingInfo.RoamingNumber != "972501111111" {
		t.Errorf("RoamingNumber round-trip failed: %+v", got.ExtendedRoutingInfo)
	}
	if got.NumberPortabilityStatus == nil || *got.NumberPortabilityStatus != MnpForeignNumberPortedIn {
		t.Errorf("NumberPortabilityStatus: got %v want %v", got.NumberPortabilityStatus, in.NumberPortabilityStatus)
	}
}

func TestSriRespForwardingDataRoundTrip(t *testing.T) {
	in := &SriResp{
		IMSI: "425010123456789",
		ExtendedRoutingInfo: &ExtendedRoutingInfo{
			RoutingInfo: &RoutingInfo{
				ForwardingData: &ForwardingData{ForwardedToNumber: "972502222222"},
			},
		},
	}
	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseSriResp(data)
	if err != nil {
		t.Fatalf("ParseSriResp: %v", err)
	}
	if got.ExtendedRoutingInfo.RoutingInfo.ForwardingData == nil ||
		got.ExtendedRoutingInfo.RoutingInfo.ForwardingData.ForwardedToNumber != "972502222222" {
		t.Errorf("ForwardingData round-trip failed: %+v", got.ExtendedRoutingInfo.RoutingInfo)
	}
}

func intPtr(v int) *int { return &v }

func TestSriRespFullStressRoundTrip(t *testing.T) {
	mnp := MnpOwnNumberPortedOut
	ua := UnavailCallBarred
	timer := 5
	camel4 := &OfferedCamel4CSIs{OCSI: true, TCSI: true, PsiEnhancements: true}

	in := &SriResp{
		IMSI: "425010123456789",
		ExtendedRoutingInfo: &ExtendedRoutingInfo{
			RoutingInfo: &RoutingInfo{
				ForwardingData: &ForwardingData{
					ForwardedToNumber: "972502222222",
					ForwardingOptions: HexBytes{0x05},
				},
			},
		},
		CugCheckInfo:                    &CugCheckInfo{CugInterlock: HexBytes{0x01, 0x02, 0x03, 0x04}},
		CugSubscriptionFlag:             true,
		SsList:                          []SsCode{0x11, 0x22, 0x33},
		BasicService:                    &ExtBasicServiceCode{ExtTeleservice: HexBytes{0x11}},
		BasicService2:                   &ExtBasicServiceCode{ExtBearerService: HexBytes{0x21}},
		ForwardingInterrogationRequired: true,
		VmscAddress:                     "972533333333",
		CcbsIndicators:                  &CcbsIndicators{CcbsPossible: true, KeepCCBSCallIndicator: true},
		MSISDN:                          "972501234567",
		NumberPortabilityStatus:         &mnp,
		IstAlertTimer:                   &timer,
		SupportedCamelPhasesInVMSC:      &SupportedCamelPhases{Phase1: true, Phase2: true, Phase3: true, Phase4: true},
		OfferedCamel4CSIsInVMSC:         camel4,
		RoutingInfo2:                    &RoutingInfo{RoamingNumber: "972544444444"},
		SsList2:                         []SsCode{0x44},
		AllowedServices:                 &AllowedServicesFlags{FirstServiceAllowed: true, SecondServiceAllowed: true},
		UnavailabilityCause:             &ua,
		ReleaseResourcesSupported:       true,
		GsmBearerCapability:             &ExternalSignalInfo{ProtocolID: 0, SignalInfo: HexBytes{0xDE, 0xAD}},
	}

	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseSriResp(data)
	if err != nil {
		t.Fatalf("ParseSriResp: %v", err)
	}

	// Normalize nature/plan defaults in input (address.NatureInternational, address.PlanISDN)
	in.VmscNature, in.VmscPlan = address.NatureInternational, address.PlanISDN
	in.MSISDNNature, in.MSISDNPlan = address.NatureInternational, address.PlanISDN
	if in.ExtendedRoutingInfo.RoutingInfo.ForwardingData != nil {
		in.ExtendedRoutingInfo.RoutingInfo.ForwardingData.ForwardedToNumberNature = address.NatureInternational
		in.ExtendedRoutingInfo.RoutingInfo.ForwardingData.ForwardedToNumberPlan = address.PlanISDN
	}
	if in.RoutingInfo2 != nil {
		in.RoutingInfo2.RoamingNumberNature, in.RoutingInfo2.RoamingNumberPlan = address.NatureInternational, address.PlanISDN
	}

	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("round-trip diff (-want +got):\n%s", diff)
	}
}

func TestSriFullStressRoundTrip(t *testing.T) {
	nof := 3
	orCap := 2
	ccbs := 1
	istSup := 1
	emlpp := 4
	fr := ForwardingBusy

	in := &Sri{
		MSISDN:              "972501234567",
		InterrogationType:   InterrogationForwarding,
		GmscOrGsmSCFAddress: "972531111111",

		CugCheckInfo:       &CugCheckInfo{CugInterlock: HexBytes{0x01, 0x02, 0x03, 0x04}, CugOutgoingAccess: true},
		NumberOfForwarding: &nof,
		OrInterrogation:    true,
		OrCapability:       &orCap,
		CallReferenceNumber: HexBytes{0xAA, 0xBB, 0xCC, 0xDD},
		ForwardingReason:   &fr,
		BasicServiceGroup:  &ExtBasicServiceCode{ExtTeleservice: HexBytes{0x11}},
		BasicServiceGroup2: &ExtBasicServiceCode{ExtBearerService: HexBytes{0x21}},
		NetworkSignalInfo:  &ExternalSignalInfo{ProtocolID: 0, SignalInfo: HexBytes{0xDE, 0xAD}},
		NetworkSignalInfo2: &ExternalSignalInfo{ProtocolID: 1, SignalInfo: HexBytes{0xBE, 0xEF}},
		CamelInfo: &SriCamelInfo{
			SupportedCamelPhases: SupportedCamelPhases{Phase1: true, Phase2: true, Phase3: true, Phase4: true},
			SuppressTCSI:         true,
		},
		SuppressionOfAnnouncement:       true,
		AlertingPattern:                 HexBytes{0x05},
		CcbsCall:                        true,
		SupportedCCBSPhase:              &ccbs,
		AdditionalSignalInfo:            &ExtExternalSignalInfo{ExtProtocolID: 1, SignalInfo: HexBytes{0xCA, 0xFE}},
		IstSupportIndicator:             &istSup,
		PrePagingSupported:              true,
		CallDiversionTreatmentIndicator: HexBytes{0x01},
		LongFTNSupported:                true,
		SuppressVTCSI:                   true,
		SuppressIncomingCallBarring:     true,
		GsmSCFInitiatedCall:             true,
		SuppressMTSS:                    &SuppressMTSSFlags{SuppressCUG: true, SuppressCCBS: true},
		MtRoamingRetrySupported:         true,
		CallPriority:                    &emlpp,
	}

	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseSri(data)
	if err != nil {
		t.Fatalf("ParseSri: %v", err)
	}

	// Natures/plans normalize to International(0x10)/ISDN(1) when zero.
	in.MSISDNNature, in.MSISDNPlan = address.NatureInternational, address.PlanISDN
	in.GmscNature, in.GmscPlan = address.NatureInternational, address.PlanISDN

	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("round-trip diff (-want +got):\n%s", diff)
	}
}

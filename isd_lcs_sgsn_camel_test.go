package gsmmap

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// ----------------------------------------------------------------------------
// LCSClientExternalID + ExternalClient
// ----------------------------------------------------------------------------

func makeLCSClientExternalID() LCSClientExternalID {
	return LCSClientExternalID{
		ExternalAddress:       "31611111111",
		ExternalAddressNature: 0x10, // International
		ExternalAddressPlan:   0x01, // ISDN
	}
}

func makeExternalClient() ExternalClient {
	gmlc := GMLCRestrictionGmlcList
	notify := NotifyLocationAllowed
	return ExternalClient{
		ClientIdentity:       makeLCSClientExternalID(),
		GmlcRestriction:      &gmlc,
		NotificationToMSUser: &notify,
	}
}

func TestLCSClientExternalID_RoundTrip(t *testing.T) {
	in := makeLCSClientExternalID()
	w, err := convertLCSClientExternalIDToWire(&in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToLCSClientExternalID(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(&in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, *out)
	}
}

func TestExternalClient_RoundTrip(t *testing.T) {
	in := makeExternalClient()
	w, err := convertExternalClientToWire(&in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToExternalClient(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(&in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, *out)
	}
}

func TestExternalClient_GMLCRestrictionInvalid(t *testing.T) {
	in := makeExternalClient()
	bad := GMLCRestriction(99)
	in.GmlcRestriction = &bad
	_, err := convertExternalClientToWire(&in)
	if !errors.Is(err, ErrGMLCRestrictionInvalid) {
		t.Fatalf("want ErrGMLCRestrictionInvalid, got %v", err)
	}
}

func TestExternalClient_NotificationInvalid(t *testing.T) {
	in := makeExternalClient()
	bad := NotificationToMSUser(99)
	in.NotificationToMSUser = &bad
	_, err := convertExternalClientToWire(&in)
	if !errors.Is(err, ErrNotificationToMSUserInvalid) {
		t.Fatalf("want ErrNotificationToMSUserInvalid, got %v", err)
	}
}

// ----------------------------------------------------------------------------
// Lists: ExternalClientList, ExtExternalClientList, PLMNClientList,
//        ServiceTypeList, MOLRList, GMLCList
// ----------------------------------------------------------------------------

func TestExternalClientList_AllowsEmpty(t *testing.T) {
	// SIZE 0..5 — empty list IS valid for this list (the only one)
	in := ExternalClientList{}
	w, err := convertExternalClientListToWire(in)
	if err != nil {
		t.Fatalf("toWire empty: %v", err)
	}
	out, err := convertWireToExternalClientList(w)
	if err != nil {
		t.Fatalf("fromWire empty: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch")
	}
}

func TestExternalClientList_OverMax(t *testing.T) {
	too := make(ExternalClientList, 6)
	for i := range too {
		too[i] = makeExternalClient()
	}
	_, err := convertExternalClientListToWire(too)
	if !errors.Is(err, ErrExternalClientListSize) {
		t.Fatalf("want ErrExternalClientListSize, got %v", err)
	}
}

func TestExtExternalClientList_BoundsRejected(t *testing.T) {
	_, err := convertExtExternalClientListToWire(ExtExternalClientList{})
	if !errors.Is(err, ErrExtExternalClientListSize) {
		t.Fatalf("empty: want size err, got %v", err)
	}
	too := make(ExtExternalClientList, 36)
	for i := range too {
		too[i] = makeExternalClient()
	}
	_, err = convertExtExternalClientListToWire(too)
	if !errors.Is(err, ErrExtExternalClientListSize) {
		t.Fatalf("over-max: want size err, got %v", err)
	}
}

func TestPLMNClientList_RoundTrip(t *testing.T) {
	in := PLMNClientList{LCSClientBroadcastService, LCSClientOAndMHPLMN}
	w, err := convertPLMNClientListToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToPLMNClientList(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch")
	}
}

func TestPLMNClientList_BoundsRejected(t *testing.T) {
	_, err := convertPLMNClientListToWire(PLMNClientList{})
	if !errors.Is(err, ErrPLMNClientListSize) {
		t.Fatalf("empty: want size err, got %v", err)
	}
	too := make(PLMNClientList, 6)
	_, err = convertPLMNClientListToWire(too)
	if !errors.Is(err, ErrPLMNClientListSize) {
		t.Fatalf("over-max: want size err, got %v", err)
	}
}

func TestPLMNClientList_InvalidValue(t *testing.T) {
	in := PLMNClientList{LCSClientInternalID(99)}
	_, err := convertPLMNClientListToWire(in)
	if !errors.Is(err, ErrLCSClientInternalIDInvalid) {
		t.Fatalf("want ErrLCSClientInternalIDInvalid, got %v", err)
	}
}

func TestServiceType_RoundTrip(t *testing.T) {
	gmlc := GMLCRestrictionHomeCountry
	in := ServiceType{
		ServiceTypeIdentity: 5,
		GmlcRestriction:     &gmlc,
	}
	w, err := convertServiceTypeToWire(&in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToServiceType(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(&in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, *out)
	}
}

func TestServiceTypeList_BoundsRejected(t *testing.T) {
	_, err := convertServiceTypeListToWire(ServiceTypeList{})
	if !errors.Is(err, ErrServiceTypeListSize) {
		t.Fatalf("empty: want size err, got %v", err)
	}
	too := make(ServiceTypeList, 33)
	for i := range too {
		too[i] = ServiceType{ServiceTypeIdentity: int64(i)}
	}
	_, err = convertServiceTypeListToWire(too)
	if !errors.Is(err, ErrServiceTypeListSize) {
		t.Fatalf("over-max: want size err, got %v", err)
	}
}

// ----------------------------------------------------------------------------
// LCSPrivacyClass / LCSPrivacyExceptionList
// ----------------------------------------------------------------------------

func makeLCSPrivacyClass() LCSPrivacyClass {
	notify := NotifyLocationAllowed
	return LCSPrivacyClass{
		SsCode:               SsCode(0x21),
		SsStatus:             HexBytes{0x01},
		NotificationToMSUser: &notify,
		ExternalClientList:   ExternalClientList{makeExternalClient()},
		PlmnClientList:       PLMNClientList{LCSClientBroadcastService},
		ServiceTypeList:      ServiceTypeList{{ServiceTypeIdentity: 1}},
	}
}

func TestLCSPrivacyClass_RoundTrip(t *testing.T) {
	in := makeLCSPrivacyClass()
	w, err := convertLCSPrivacyClassToWire(&in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToLCSPrivacyClass(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(&in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, *out)
	}
}

func TestLCSPrivacyClass_SsStatusInvalid(t *testing.T) {
	in := makeLCSPrivacyClass()
	in.SsStatus = HexBytes{}
	_, err := convertLCSPrivacyClassToWire(&in)
	if !errors.Is(err, ErrExtSSStatusInvalidSize) {
		t.Fatalf("want ErrExtSSStatusInvalidSize, got %v", err)
	}
}

func TestLCSPrivacyExceptionList_BoundsRejected(t *testing.T) {
	_, err := convertLCSPrivacyExceptionListToWire(LCSPrivacyExceptionList{})
	if !errors.Is(err, ErrLCSPrivacyExceptionListSize) {
		t.Fatalf("empty: want size err, got %v", err)
	}
	too := make(LCSPrivacyExceptionList, 5)
	for i := range too {
		too[i] = makeLCSPrivacyClass()
	}
	_, err = convertLCSPrivacyExceptionListToWire(too)
	if !errors.Is(err, ErrLCSPrivacyExceptionListSize) {
		t.Fatalf("over-max: want size err, got %v", err)
	}
}

// ----------------------------------------------------------------------------
// MOLRClass / MOLRList
// ----------------------------------------------------------------------------

func TestMOLRClass_RoundTrip(t *testing.T) {
	in := MOLRClass{SsCode: SsCode(0x42), SsStatus: HexBytes{0x01}}
	w, err := convertMOLRClassToWire(&in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToMOLRClass(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(&in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, *out)
	}
}

func TestMOLRList_BoundsRejected(t *testing.T) {
	_, err := convertMOLRListToWire(MOLRList{})
	if !errors.Is(err, ErrMOLRListSize) {
		t.Fatalf("empty: want size err, got %v", err)
	}
	too := make(MOLRList, 4)
	for i := range too {
		too[i] = MOLRClass{SsCode: SsCode(byte(i)), SsStatus: HexBytes{0x01}}
	}
	_, err = convertMOLRListToWire(too)
	if !errors.Is(err, ErrMOLRListSize) {
		t.Fatalf("over-max: want size err, got %v", err)
	}
}

// ----------------------------------------------------------------------------
// GMLCList
// ----------------------------------------------------------------------------

func TestGMLCList_RoundTrip(t *testing.T) {
	in := GMLCList{
		{Address: "31611111111", Nature: 0x10, Plan: 0x01},
		{Address: "31622222222", Nature: 0x10, Plan: 0x01},
	}
	w, err := convertGMLCListToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToGMLCList(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestGMLCList_BoundsRejected(t *testing.T) {
	_, err := convertGMLCListToWire(GMLCList{})
	if !errors.Is(err, ErrGMLCListSize) {
		t.Fatalf("empty: want size err, got %v", err)
	}
	too := make(GMLCList, 6)
	for i := range too {
		too[i] = GMLCAddress{Address: "31611111111", Nature: 0x10, Plan: 0x01}
	}
	_, err = convertGMLCListToWire(too)
	if !errors.Is(err, ErrGMLCListSize) {
		t.Fatalf("over-max: want size err, got %v", err)
	}
}

// ----------------------------------------------------------------------------
// LCSInformation full + minimal round-trip
// ----------------------------------------------------------------------------

func TestLCSInformation_FullRoundTrip(t *testing.T) {
	in := &LCSInformation{
		GmlcList:                GMLCList{{Address: "31611111111", Nature: 0x10, Plan: 0x01}},
		LcsPrivacyExceptionList: LCSPrivacyExceptionList{makeLCSPrivacyClass()},
		MolrList:                MOLRList{{SsCode: SsCode(0x42), SsStatus: HexBytes{0x01}}},
	}
	w, err := convertLCSInformationToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToLCSInformation(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestLCSInformation_MinimalRoundTrip(t *testing.T) {
	in := &LCSInformation{}
	w, err := convertLCSInformationToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToLCSInformation(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

// ============================================================================
// SGSN-CAMEL Subscription Tests
// ============================================================================

func makeGPRSCamelTDPData() GPRSCamelTDPData {
	return GPRSCamelTDPData{
		GprsTriggerDetectionPoint: GPRSTDPAttach,
		ServiceKey:                42,
		GsmSCFAddress:             "31611111111",
		GsmSCFAddressNature:       0x10,
		GsmSCFAddressPlan:         0x01,
		DefaultSessionHandling:    DefaultGPRSContinueTransaction,
	}
}

func TestGPRSCamelTDPData_RoundTrip(t *testing.T) {
	in := makeGPRSCamelTDPData()
	w, err := convertGPRSCamelTDPDataToWire(&in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToGPRSCamelTDPData(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(&in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, *out)
	}
}

func TestGPRSCamelTDPData_DefaultGPRSHandlingLenientRemap(t *testing.T) {
	// Spec exception clause: decoder remaps DefaultSessionHandling >1 to
	// releaseTransaction. Verify by hand-crafting a wire frame.
	addr, _ := encodeAddressField("31611111111", 0x10, 0x01)
	w := &gsm_map.GPRSCamelTDPData{
		GprsTriggerDetectionPoint: gsm_map.GPRSTriggerDetectionPointAttach,
		ServiceKey:                gsm_map.ServiceKey(1),
		GsmSCFAddress:             gsm_map.ISDNAddressString(addr),
		DefaultSessionHandling:    gsm_map.DefaultGPRSHandling(99), // out-of-range
	}
	out, err := convertWireToGPRSCamelTDPData(w)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.DefaultSessionHandling != DefaultGPRSReleaseTransaction {
		t.Fatalf("lenient remap: want DefaultGPRSReleaseTransaction, got %v", out.DefaultSessionHandling)
	}
}

func TestGPRSCamelTDPData_ServiceKeyRange(t *testing.T) {
	for _, sk := range []int64{-1, 2147483648, 1 << 40} {
		in := makeGPRSCamelTDPData()
		in.ServiceKey = sk
		_, err := convertGPRSCamelTDPDataToWire(&in)
		if !errors.Is(err, ErrCamelInvalidServiceKey) {
			t.Fatalf("encode sk=%d: want ErrCamelInvalidServiceKey, got %v", sk, err)
		}
	}
}

func TestGPRSCamelTDPData_EmptyAddressRejected(t *testing.T) {
	in := makeGPRSCamelTDPData()
	in.GsmSCFAddress = ""
	_, err := convertGPRSCamelTDPDataToWire(&in)
	if err == nil {
		t.Fatalf("encode empty GsmSCFAddress: want error, got nil")
	}
}

func TestMGCSI_ServiceKeyRange(t *testing.T) {
	for _, sk := range []int64{-1, 2147483648, 1 << 40} {
		in := makeMGCSI()
		in.ServiceKey = sk
		_, err := convertMGCSIToWire(in)
		if !errors.Is(err, ErrCamelInvalidServiceKey) {
			t.Fatalf("encode sk=%d: want ErrCamelInvalidServiceKey, got %v", sk, err)
		}
	}
}

func TestLCSPrivacyClass_SsCodeStrictSize(t *testing.T) {
	w := &gsm_map.LCSPrivacyClass{
		SsCode:   gsm_map.SSCode{0x21, 0x42}, // 2 octets — should be 1
		SsStatus: gsm_map.ExtSSStatus{0x01},
	}
	_, err := convertWireToLCSPrivacyClass(w)
	if !errors.Is(err, ErrLCSPrivacyClassSsCodeInvalidSize) {
		t.Fatalf("want ErrLCSPrivacyClassSsCodeInvalidSize, got %v", err)
	}
}

func TestMOLRClass_SsCodeStrictSize(t *testing.T) {
	w := &gsm_map.MOLRClass{
		SsCode:   gsm_map.SSCode{0x21, 0x42},
		SsStatus: gsm_map.ExtSSStatus{0x01},
	}
	_, err := convertWireToMOLRClass(w)
	if !errors.Is(err, ErrMOLRClassSsCodeInvalidSize) {
		t.Fatalf("want ErrMOLRClassSsCodeInvalidSize, got %v", err)
	}
}

func TestGMLCAddress_EmptyRejected(t *testing.T) {
	in := GMLCList{{Address: "", Nature: 0x10, Plan: 0x01}}
	_, err := convertGMLCListToWire(in)
	if !errors.Is(err, ErrGMLCAddressEmpty) {
		t.Fatalf("encode empty: want ErrGMLCAddressEmpty, got %v", err)
	}
}

func TestSGSNCAMELSubscriptionInfo_MtSmsCAMELTDPCriteriaListSize(t *testing.T) {
	tdp := MTSmsCAMELTDPCriteria{
		SmsTriggerDetectionPoint: SMSTriggerDetectionPoint(1),
	}
	in := &SGSNCAMELSubscriptionInfo{
		MtSmsCAMELTDPCriteriaList: []MTSmsCAMELTDPCriteria{},
	}
	_, err := convertSGSNCAMELSubscriptionInfoToWire(in)
	if !errors.Is(err, ErrSGSNMtSmsCAMELTDPCriteriaListSize) {
		t.Fatalf("empty: want ErrSGSNMtSmsCAMELTDPCriteriaListSize, got %v", err)
	}
	too := make([]MTSmsCAMELTDPCriteria, 11)
	for i := range too {
		too[i] = tdp
	}
	in.MtSmsCAMELTDPCriteriaList = too
	_, err = convertSGSNCAMELSubscriptionInfoToWire(in)
	if !errors.Is(err, ErrSGSNMtSmsCAMELTDPCriteriaListSize) {
		t.Fatalf("over-max: want ErrSGSNMtSmsCAMELTDPCriteriaListSize, got %v", err)
	}
}

func TestGPRSCamelTDPDataList_BoundsRejected(t *testing.T) {
	_, err := convertGPRSCamelTDPDataListToWire(GPRSCamelTDPDataList{})
	if !errors.Is(err, ErrGPRSCamelTDPDataListSize) {
		t.Fatalf("empty: want size err, got %v", err)
	}
	too := make(GPRSCamelTDPDataList, 11)
	for i := range too {
		too[i] = makeGPRSCamelTDPData()
	}
	_, err = convertGPRSCamelTDPDataListToWire(too)
	if !errors.Is(err, ErrGPRSCamelTDPDataListSize) {
		t.Fatalf("over-max: want size err, got %v", err)
	}
}

// ----------------------------------------------------------------------------
// GPRSCSI
// ----------------------------------------------------------------------------

func TestGPRSCSI_RoundTrip(t *testing.T) {
	phase := 4
	in := &GPRSCSI{
		GprsCamelTDPDataList:    GPRSCamelTDPDataList{makeGPRSCamelTDPData()},
		CamelCapabilityHandling: &phase,
		NotificationToCSE:       true,
		CsiActive:               true,
	}
	w, err := convertGPRSCSIToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToGPRSCSI(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestGPRSCSI_RequiresBothListAndPhase(t *testing.T) {
	// Spec clause 8.8: both must be present together
	phase := 2
	cases := []*GPRSCSI{
		{GprsCamelTDPDataList: GPRSCamelTDPDataList{makeGPRSCamelTDPData()}}, // missing phase
		{CamelCapabilityHandling: &phase},                                    // missing list
	}
	for i, c := range cases {
		_, err := convertGPRSCSIToWire(c)
		if !errors.Is(err, ErrGPRSCSIRequiresTDPListAndPhase) {
			t.Fatalf("case %d: want ErrGPRSCSIRequiresTDPListAndPhase, got %v", i, err)
		}
	}
}

func TestGPRSCSI_PhaseOutOfRange(t *testing.T) {
	for _, p := range []int{0, 5, 100} {
		phase := p
		in := &GPRSCSI{
			GprsCamelTDPDataList:    GPRSCamelTDPDataList{makeGPRSCamelTDPData()},
			CamelCapabilityHandling: &phase,
		}
		_, err := convertGPRSCSIToWire(in)
		if !errors.Is(err, ErrCamelCapabilityHandlingOutOfRange) {
			t.Fatalf("phase=%d: want ErrCamelCapabilityHandlingOutOfRange, got %v", p, err)
		}
	}
}

// ----------------------------------------------------------------------------
// MGCSI
// ----------------------------------------------------------------------------

func makeMGCSI() *MGCSI {
	return &MGCSI{
		MobilityTriggers:    []HexBytes{{0x01}, {0x02}, {0x03}},
		ServiceKey:          7,
		GsmSCFAddress:       "31633333333",
		GsmSCFAddressNature: 0x10,
		GsmSCFAddressPlan:   0x01,
		NotificationToCSE:   true,
	}
}

func TestMGCSI_RoundTrip(t *testing.T) {
	in := makeMGCSI()
	w, err := convertMGCSIToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToMGCSI(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestMGCSI_MobilityTriggersBoundsRejected(t *testing.T) {
	in := makeMGCSI()
	in.MobilityTriggers = []HexBytes{}
	_, err := convertMGCSIToWire(in)
	if !errors.Is(err, ErrMobilityTriggersSize) {
		t.Fatalf("empty: want ErrMobilityTriggersSize, got %v", err)
	}
	in.MobilityTriggers = make([]HexBytes, 11)
	for i := range in.MobilityTriggers {
		in.MobilityTriggers[i] = HexBytes{byte(i)}
	}
	_, err = convertMGCSIToWire(in)
	if !errors.Is(err, ErrMobilityTriggersSize) {
		t.Fatalf("over-max: want ErrMobilityTriggersSize, got %v", err)
	}
}

func TestMGCSI_MMCodeWrongSize(t *testing.T) {
	in := makeMGCSI()
	in.MobilityTriggers[0] = HexBytes{0x01, 0x02} // not 1 octet
	_, err := convertMGCSIToWire(in)
	if !errors.Is(err, ErrMMCodeInvalidSize) {
		t.Fatalf("want ErrMMCodeInvalidSize, got %v", err)
	}
}

// ----------------------------------------------------------------------------
// SGSNCAMELSubscriptionInfo full round-trip
// ----------------------------------------------------------------------------

func TestSGSNCAMELSubscriptionInfo_FullRoundTrip(t *testing.T) {
	phase := 4
	in := &SGSNCAMELSubscriptionInfo{
		GprsCSI: &GPRSCSI{
			GprsCamelTDPDataList:    GPRSCamelTDPDataList{makeGPRSCamelTDPData()},
			CamelCapabilityHandling: &phase,
		},
		MgCsi: makeMGCSI(),
	}
	w, err := convertSGSNCAMELSubscriptionInfoToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToSGSNCAMELSubscriptionInfo(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestSGSNCAMELSubscriptionInfo_MinimalRoundTrip(t *testing.T) {
	in := &SGSNCAMELSubscriptionInfo{}
	w, err := convertSGSNCAMELSubscriptionInfoToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToSGSNCAMELSubscriptionInfo(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

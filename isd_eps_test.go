package gsmmap

import (
	"errors"
	"reflect"
	"testing"
)

// ----------------------------------------------------------------------------
// AllocationRetentionPriority
// ----------------------------------------------------------------------------

func ptrBool(b bool) *bool { return &b }

func TestAllocationRetentionPriority_RoundTrip(t *testing.T) {
	in := &AllocationRetentionPriority{
		PriorityLevel:           5,
		PreEmptionCapability:    ptrBool(true),
		PreEmptionVulnerability: ptrBool(false),
	}
	w := convertAllocationRetentionPriorityToWire(in)
	out := convertWireToAllocationRetentionPriority(w)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

// ----------------------------------------------------------------------------
// EPSQoSSubscribed
// ----------------------------------------------------------------------------

func TestEPSQoSSubscribed_RoundTrip(t *testing.T) {
	in := &EPSQoSSubscribed{
		QosClassIdentifier: 5,
		AllocationRetentionPriority: AllocationRetentionPriority{
			PriorityLevel: 7,
		},
	}
	w, err := convertEPSQoSSubscribedToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToEPSQoSSubscribed(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestEPSQoSSubscribed_QCIOutOfRange(t *testing.T) {
	for _, qci := range []int{0, 10, 100} {
		in := &EPSQoSSubscribed{QosClassIdentifier: qci}
		_, err := convertEPSQoSSubscribedToWire(in)
		if !errors.Is(err, ErrQoSClassIdentifierOutOfRange) {
			t.Fatalf("qci=%d: want ErrQoSClassIdentifierOutOfRange, got %v", qci, err)
		}
	}
}

// ----------------------------------------------------------------------------
// PDNGWIdentity
// ----------------------------------------------------------------------------

func TestPDNGWIdentity_RoundTrip(t *testing.T) {
	in := &PDNGWIdentity{
		PdnGwIpv4Address: HexBytes{0x0a, 0x01, 0x02, 0x03},
		PdnGwIpv6Address: HexBytes{0x20, 0x01, 0x0d, 0xb8},
		PdnGwName:        HexBytes("pgw.example.com"),
	}
	w, err := convertPDNGWIdentityToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToPDNGWIdentity(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch")
	}
}

func TestPDNGWIdentity_InvalidName(t *testing.T) {
	in := &PDNGWIdentity{PdnGwName: HexBytes("short")}
	_, err := convertPDNGWIdentityToWire(in)
	if !errors.Is(err, ErrFQDNInvalidSize) {
		t.Fatalf("want ErrFQDNInvalidSize, got %v", err)
	}
}

// ----------------------------------------------------------------------------
// SpecificAPNInfoList
// ----------------------------------------------------------------------------

func makeSpecificAPNInfo() SpecificAPNInfo {
	return SpecificAPNInfo{
		Apn: HexBytes{'a', 'p', 'n', '.', 'e', 'x'},
		PdnGwIdentity: PDNGWIdentity{
			PdnGwIpv4Address: HexBytes{0x0a, 0x01, 0x02, 0x03},
		},
	}
}

func TestSpecificAPNInfoList_RoundTrip(t *testing.T) {
	in := SpecificAPNInfoList{makeSpecificAPNInfo(), makeSpecificAPNInfo()}
	w, err := convertSpecificAPNInfoListToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToSpecificAPNInfoList(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch")
	}
}

func TestSpecificAPNInfoList_BoundsRejected(t *testing.T) {
	_, err := convertSpecificAPNInfoListToWire(SpecificAPNInfoList{})
	if !errors.Is(err, ErrSpecificAPNInfoListSize) {
		t.Fatalf("empty: want size err, got %v", err)
	}
	too := make(SpecificAPNInfoList, 51)
	for i := range too {
		too[i] = makeSpecificAPNInfo()
	}
	_, err = convertSpecificAPNInfoListToWire(too)
	if !errors.Is(err, ErrSpecificAPNInfoListSize) {
		t.Fatalf("over-max: want size err, got %v", err)
	}
}

// ----------------------------------------------------------------------------
// WLANOffloadability
// ----------------------------------------------------------------------------

func TestWLANOffloadability_RoundTrip(t *testing.T) {
	eutran := WLANOffloadabilityAllowed
	utran := WLANOffloadabilityNotAllowed
	in := &WLANOffloadability{
		WlanOffloadabilityEUTRAN: &eutran,
		WlanOffloadabilityUTRAN:  &utran,
	}
	w, err := convertWLANOffloadabilityToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToWLANOffloadability(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch")
	}
}

func TestWLANOffloadability_OutOfRange(t *testing.T) {
	v := WLANOffloadabilityIndication(99)
	in := &WLANOffloadability{WlanOffloadabilityEUTRAN: &v}
	_, err := convertWLANOffloadabilityToWire(in)
	if !errors.Is(err, ErrWLANOffloadabilityIndicationInvalid) {
		t.Fatalf("want ErrWLANOffloadabilityIndicationInvalid, got %v", err)
	}
}

// ----------------------------------------------------------------------------
// APNConfiguration (full + minimal round-trips, validation cases)
// ----------------------------------------------------------------------------

func makeAPNConfiguration() APNConfiguration {
	siptoP := SIPTOAboveRanAllowed
	siptoLN := SIPTOAtLocalNetworkAllowed
	lipa := LIPAOnly
	nidd := NIDDSCEFBasedDataDelivery
	pgwAlloc := PDNGWAllocationDynamic
	pdnCC := PDNConnectionMaintain
	eutranWLAN := WLANOffloadabilityAllowed
	return APNConfiguration{
		ContextId: 1,
		PdnType:   HexBytes{0x01},
		ServedPartyIPIPv4Address: HexBytes{0x0a, 0x01, 0x02, 0x03},
		Apn:       HexBytes("internet.apn"),
		EpsQosSubscribed: EPSQoSSubscribed{
			QosClassIdentifier: 5,
			AllocationRetentionPriority: AllocationRetentionPriority{
				PriorityLevel:           7,
				PreEmptionCapability:    ptrBool(true),
				PreEmptionVulnerability: ptrBool(false),
			},
		},
		PdnGwIdentity:           &PDNGWIdentity{PdnGwIpv4Address: HexBytes{0x0a, 0x01, 0x02, 0x04}},
		PdnGwAllocationType:     &pgwAlloc,
		VplmnAddressAllowed:     true,
		ChargingCharacteristics: HexBytes{0x08, 0x00},
		Ambr: &AMBR{
			MaxRequestedBandwidthUL: 1_000_000,
			MaxRequestedBandwidthDL: 5_000_000,
		},
		SpecificAPNInfoList:         SpecificAPNInfoList{makeSpecificAPNInfo()},
		ServedPartyIPIPv6Address:    HexBytes{0x20, 0x01, 0x0d, 0xb8},
		ApnOiReplacement:            HexBytes("apn-oi-9b"),
		SiptoPermission:             &siptoP,
		LipaPermission:              &lipa,
		RestorationPriority:         HexBytes{0x05},
		SiptoLocalNetworkPermission: &siptoLN,
		WlanOffloadability:          &WLANOffloadability{WlanOffloadabilityEUTRAN: &eutranWLAN},
		NonIPPDNTypeIndicator:       true,
		NIDDMechanism:               &nidd,
		SCEFID:                      HexBytes("scef.example.com"),
		PdnConnectionContinuity:     &pdnCC,
	}
}

func TestAPNConfiguration_FullRoundTrip(t *testing.T) {
	in := makeAPNConfiguration()
	w, err := convertAPNConfigurationToWire(&in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToAPNConfiguration(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(&in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, *out)
	}
}

func TestAPNConfiguration_MinimalRoundTrip(t *testing.T) {
	in := APNConfiguration{
		ContextId: 1,
		PdnType:   HexBytes{0x01},
		Apn:       HexBytes{'a', 'p'},
		EpsQosSubscribed: EPSQoSSubscribed{
			QosClassIdentifier:          1,
			AllocationRetentionPriority: AllocationRetentionPriority{PriorityLevel: 1},
		},
	}
	w, err := convertAPNConfigurationToWire(&in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToAPNConfiguration(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(&in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, *out)
	}
}

func TestAPNConfiguration_FieldSizeViolations(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*APNConfiguration)
		want error
	}{
		{"ContextId out of range", func(a *APNConfiguration) { a.ContextId = 51 }, ErrPDPContextIdOutOfRange},
		{"PdnType wrong size", func(a *APNConfiguration) { a.PdnType = HexBytes{0x01, 0x02} }, ErrPDNTypeInvalidSize},
		{"Apn too short", func(a *APNConfiguration) { a.Apn = HexBytes{'a'} }, ErrAPNInvalidSize},
		{"ChargingCharacteristics wrong", func(a *APNConfiguration) { a.ChargingCharacteristics = HexBytes{0x01} }, ErrPDPChargingCharsInvalidSize},
		{"ApnOiReplacement too short", func(a *APNConfiguration) { a.ApnOiReplacement = HexBytes("short") }, ErrAPNOIReplacementInvalidSize},
		{"RestorationPriority wrong", func(a *APNConfiguration) { a.RestorationPriority = HexBytes{0x01, 0x02} }, ErrRestorationPriorityInvalidSize},
		{"SCEFID too short", func(a *APNConfiguration) { a.SCEFID = HexBytes("short") }, ErrFQDNInvalidSize},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := makeAPNConfiguration()
			tc.mut(&in)
			_, err := convertAPNConfigurationToWire(&in)
			if !errors.Is(err, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, err)
			}
		})
	}
}

func TestAPNConfiguration_EnumOutOfRange(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*APNConfiguration)
		want error
	}{
		{"PdnGwAllocationType", func(a *APNConfiguration) {
			v := PDNGWAllocationType(99)
			a.PdnGwAllocationType = &v
		}, ErrPDNGWAllocationTypeInvalid},
		{"PdnConnectionContinuity", func(a *APNConfiguration) {
			v := PDNConnectionContinuity(99)
			a.PdnConnectionContinuity = &v
		}, ErrPDNConnectionContinuityInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := makeAPNConfiguration()
			tc.mut(&in)
			_, err := convertAPNConfigurationToWire(&in)
			if !errors.Is(err, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, err)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// EPSDataList / APNConfigurationProfile / EPSSubscriptionData
// ----------------------------------------------------------------------------

func TestEPSDataList_BoundsRejected(t *testing.T) {
	_, err := convertEPSDataListToWire(EPSDataList{})
	if !errors.Is(err, ErrEPSDataListSize) {
		t.Fatalf("empty: want size err, got %v", err)
	}
	too := make(EPSDataList, 51)
	for i := range too {
		c := makeAPNConfiguration()
		c.ContextId = (i % 50) + 1
		too[i] = c
	}
	_, err = convertEPSDataListToWire(too)
	if !errors.Is(err, ErrEPSDataListSize) {
		t.Fatalf("over-max: want size err, got %v", err)
	}
}

func TestAPNConfigurationProfile_RoundTrip(t *testing.T) {
	additional := 7
	in := &APNConfigurationProfile{
		DefaultContext:           1,
		CompleteDataListIncluded: true,
		EpsDataList:              EPSDataList{makeAPNConfiguration()},
		AdditionalDefaultContext: &additional,
	}
	w, err := convertAPNConfigurationProfileToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToAPNConfigurationProfile(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestAPNConfigurationProfile_MissingList(t *testing.T) {
	in := &APNConfigurationProfile{DefaultContext: 1} // EpsDataList nil
	_, err := convertAPNConfigurationProfileToWire(in)
	if !errors.Is(err, ErrAPNConfigurationProfileMissingList) {
		t.Fatalf("want ErrAPNConfigurationProfileMissingList, got %v", err)
	}
}

func TestAPNConfigurationProfile_DefaultContextOutOfRange(t *testing.T) {
	in := &APNConfigurationProfile{DefaultContext: 0, EpsDataList: EPSDataList{makeAPNConfiguration()}}
	_, err := convertAPNConfigurationProfileToWire(in)
	if !errors.Is(err, ErrPDPContextIdOutOfRange) {
		t.Fatalf("want ErrPDPContextIdOutOfRange, got %v", err)
	}
}

func TestEPSSubscriptionData_FullRoundTrip(t *testing.T) {
	rfsp := 100
	in := &EPSSubscriptionData{
		ApnOiReplacement: HexBytes("apn-oi-9b"),
		RfspId:           &rfsp,
		Ambr: &AMBR{
			MaxRequestedBandwidthUL: 1_000_000,
			MaxRequestedBandwidthDL: 5_000_000,
		},
		ApnConfigurationProfile: &APNConfigurationProfile{
			DefaultContext: 1,
			EpsDataList:    EPSDataList{makeAPNConfiguration()},
		},
		StnSr:       "12345",
		StnSrNature: 0x10, // International (pre-shifted)
		StnSrPlan:   0x01, // ISDN
		MpsCSPriority:    true,
		MpsEPSPriority:   true,
		SubscribedVsrvcc: true,
	}
	w, err := convertEPSSubscriptionDataToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToEPSSubscriptionData(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestEPSSubscriptionData_MinimalRoundTrip(t *testing.T) {
	in := &EPSSubscriptionData{}
	w, err := convertEPSSubscriptionDataToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToEPSSubscriptionData(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestEPSSubscriptionData_RFSPOutOfRange(t *testing.T) {
	for _, v := range []int{0, 257, 1000} {
		in := &EPSSubscriptionData{RfspId: &v}
		_, err := convertEPSSubscriptionDataToWire(in)
		if !errors.Is(err, ErrRFSPIDOutOfRange) {
			t.Fatalf("rfsp=%d: want ErrRFSPIDOutOfRange, got %v", v, err)
		}
	}
}

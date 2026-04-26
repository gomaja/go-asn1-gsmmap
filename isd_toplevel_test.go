package gsmmap

import (
	"errors"
	"reflect"
	"testing"
)

// ----------------------------------------------------------------------------
// InsertSubscriberDataArg round-trips
// ----------------------------------------------------------------------------

func TestInsertSubscriberDataArg_MinimalRoundTrip(t *testing.T) {
	in := &InsertSubscriberDataArg{
		IMSI: HexBytes{0x12, 0x34, 0x56, 0x78, 0x90}, // opaque BCD
	}
	encoded, err := in.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out, err := ParseInsertSubscriberData(encoded)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestInsertSubscriberDataArg_MSISDNRoundTrip(t *testing.T) {
	in := &InsertSubscriberDataArg{
		IMSI:         HexBytes{0x12, 0x34},
		MSISDN:       "31611111111",
		MSISDNNature: 0x10, // International
		MSISDNPlan:   0x01, // ISDN
	}
	encoded, err := in.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out, err := ParseInsertSubscriberData(encoded)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestInsertSubscriberDataArg_NULLFlagsRoundTrip(t *testing.T) {
	// Exercise the dozen-or-so OPTIONAL NULL bool flags to confirm the
	// encode/decode pair via boolToNullPtr/nullPtrToBool round-trips.
	in := &InsertSubscriberDataArg{
		IMSI:                                      HexBytes{0x12},
		RoamingRestrictionDueToUnsupportedFeature: true,
		LmuIndicator:                              true,
		UeReachabilityRequestIndicator:            true,
		VplmnLIPAAllowed:                          true,
		PsAndSMSOnlyServiceProvision:              true,
		SmsInSGSNAllowed:                          true,
		CsToPsSRVCCAllowedIndicator:               true,
		PcscfRestorationRequest:                   true,
		UserPlaneIntegrityProtectionIndicator:     true,
		IabOperationAllowedIndicator:              true,
	}
	encoded, err := in.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out, err := ParseInsertSubscriberData(encoded)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestInsertSubscriberDataArg_OdbAndZoneCodeRoundTrip(t *testing.T) {
	in := &InsertSubscriberDataArg{
		IMSI: HexBytes{0x12},
		OdbData: &ODBData{
			OdbGeneralData: &ODBGeneralData{AllOGCallsBarred: true},
		},
		RegionalSubscriptionData: ZoneCodeList{
			ZoneCode(HexBytes{0x12, 0x34}),
			ZoneCode(HexBytes{0x56, 0x78}),
		},
	}
	encoded, err := in.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out, err := ParseInsertSubscriberData(encoded)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestInsertSubscriberDataArg_GprsAndLsaRoundTrip(t *testing.T) {
	pdp := PDPContext{
		PdpContextId:  1,
		PdpType:       HexBytes{0xf1, 0x21},
		QosSubscribed: HexBytes{0x09, 0x00, 0x00},
		Apn:           HexBytes{'a', 'p', 'n'},
	}
	in := &InsertSubscriberDataArg{
		IMSI: HexBytes{0x12},
		GprsSubscriptionData: &GPRSSubscriptionData{
			GprsDataList: GPRSDataList{pdp},
		},
		LsaInformation: &LSAInformation{
			LsaDataList: LSADataList{{
				LsaIdentity:   HexBytes{0x01, 0x02, 0x03},
				LsaAttributes: HexBytes{0xff},
			}},
		},
	}
	encoded, err := in.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out, err := ParseInsertSubscriberData(encoded)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestInsertSubscriberDataArg_EpsAndCsgRoundTrip(t *testing.T) {
	apn := APNConfiguration{
		ContextId: 1,
		PdnType:   HexBytes{0x01},
		Apn:       HexBytes{'a', 'p'},
		EpsQosSubscribed: EPSQoSSubscribed{
			QosClassIdentifier: 5,
			AllocationRetentionPriority: AllocationRetentionPriority{
				PriorityLevel: 7,
			},
		},
	}
	in := &InsertSubscriberDataArg{
		IMSI: HexBytes{0x12},
		EpsSubscriptionData: &EPSSubscriptionData{
			ApnConfigurationProfile: &APNConfigurationProfile{
				DefaultContext: 1,
				EpsDataList:    EPSDataList{apn},
			},
		},
		CsgSubscriptionDataList: CSGSubscriptionDataList{{
			CsgId:          HexBytes{0x12, 0x34, 0x56, 0x60},
			CsgIdBitLength: 27,
		}},
	}
	encoded, err := in.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out, err := ParseInsertSubscriberData(encoded)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestInsertSubscriberDataArg_LcsAndSgsnCamelRoundTrip(t *testing.T) {
	phase := 4
	in := &InsertSubscriberDataArg{
		IMSI: HexBytes{0x12},
		LcsInformation: &LCSInformation{
			GmlcList: GMLCList{{Address: "31622222222", Nature: 0x10, Plan: 0x01}},
		},
		SgsnCAMELSubscriptionInfo: &SGSNCAMELSubscriptionInfo{
			GprsCSI: &GPRSCSI{
				GprsCamelTDPDataList: GPRSCamelTDPDataList{{
					GprsTriggerDetectionPoint: GPRSTDPAttach,
					ServiceKey:                42,
					GsmSCFAddress:             "31633333333",
					GsmSCFAddressNature:       0x10,
					GsmSCFAddressPlan:         0x01,
					DefaultSessionHandling:    DefaultGPRSContinueTransaction,
				}},
				CamelCapabilityHandling: &phase,
			},
		},
	}
	encoded, err := in.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out, err := ParseInsertSubscriberData(encoded)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestInsertSubscriberDataArg_PR_E1aSubtypesRoundTrip(t *testing.T) {
	in := &InsertSubscriberDataArg{
		IMSI: HexBytes{0x12},
		McSSInfo: &MCSSInfo{
			SsCode:   SsCode(0x21),
			SsStatus: HexBytes{0x01},
			NbrSB:    7,
			NbrUser:  3,
		},
		AdjacentAccessRestrictionDataList: AdjacentAccessRestrictionDataList{{
			PlmnId: HexBytes{0x62, 0xf2, 0x10},
			AccessRestrictionData: AccessRestrictionData{
				UtranNotAllowed: true,
			},
		}},
		ImsiGroupIdList: IMSIGroupIdList{{
			GroupServiceID: 0xDEADBEEF,
			PlmnId:         HexBytes{0x62, 0xf2, 0x10},
			LocalGroupID:   HexBytes{0x01, 0x02, 0x03},
		}},
		EDRXCycleLengthList: EDRXCycleLengthList{{
			RatType:              UsedRatEUTRAN,
			EDRXCycleLengthValue: HexBytes{0x09},
		}},
		ResetIdList: ResetIdList{HexBytes{0x01, 0x02}},
	}
	encoded, err := in.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out, err := ParseInsertSubscriberData(encoded)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestInsertSubscriberDataArg_ServiceListsRoundTrip(t *testing.T) {
	in := &InsertSubscriberDataArg{
		IMSI:              HexBytes{0x12},
		BearerServiceList: []HexBytes{{0x10}, {0x20, 0x30}},
		TeleserviceList:   []HexBytes{{0x11}, {0x21}},
	}
	encoded, err := in.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out, err := ParseInsertSubscriberData(encoded)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

// ----------------------------------------------------------------------------
// InsertSubscriberDataRes round-trips
// ----------------------------------------------------------------------------

func TestInsertSubscriberDataRes_MinimalRoundTrip(t *testing.T) {
	in := &InsertSubscriberDataRes{}
	encoded, err := in.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out, err := ParseInsertSubscriberDataRes(encoded)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestInsertSubscriberDataRes_FullRoundTrip(t *testing.T) {
	region := RegionalSubscriptionResponse(0)
	in := &InsertSubscriberDataRes{
		TeleserviceList:              []HexBytes{{0x11}, {0x21}},
		BearerServiceList:            []HexBytes{{0x10}},
		SsList:                       []SsCode{SsCode(0x21), SsCode(0x42)},
		OdbGeneralData:               &ODBGeneralData{AllOGCallsBarred: true},
		RegionalSubscriptionResponse: &region,
		SupportedCamelPhases:         &SupportedCamelPhases{Phase1: true, Phase4: true},
	}
	encoded, err := in.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out, err := ParseInsertSubscriberDataRes(encoded)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

// ----------------------------------------------------------------------------
// Negative tests
// ----------------------------------------------------------------------------

func TestInsertSubscriberDataArg_NilRejected(t *testing.T) {
	var a *InsertSubscriberDataArg
	if _, err := a.Marshal(); err == nil {
		t.Fatalf("Marshal(nil): want error, got nil")
	}
}

func TestInsertSubscriberDataRes_NilRejected(t *testing.T) {
	var r *InsertSubscriberDataRes
	if _, err := r.Marshal(); err == nil {
		t.Fatalf("Marshal(nil) Res: want error, got nil")
	}
}

func TestInsertSubscriberDataArg_ListCardinalityRejected(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*InsertSubscriberDataArg)
		want error
	}{
		{"BearerServiceList over-max", func(a *InsertSubscriberDataArg) {
			too := make([]HexBytes, 51)
			for i := range too {
				too[i] = HexBytes{0x10}
			}
			a.BearerServiceList = too
		}, ErrIsdBearerServiceListSize},
		{"BearerServiceList empty", func(a *InsertSubscriberDataArg) {
			a.BearerServiceList = []HexBytes{}
		}, ErrIsdBearerServiceListSize},
		{"TeleserviceList over-max", func(a *InsertSubscriberDataArg) {
			too := make([]HexBytes, 21)
			for i := range too {
				too[i] = HexBytes{0x11}
			}
			a.TeleserviceList = too
		}, ErrIsdTeleserviceListSize},
		{"TeleserviceList empty", func(a *InsertSubscriberDataArg) {
			a.TeleserviceList = []HexBytes{}
		}, ErrIsdTeleserviceListSize},
		{"ProvisionedSS empty", func(a *InsertSubscriberDataArg) {
			a.ProvisionedSS = []ExtSSInfo{}
		}, ErrIsdProvisionedSSListSize},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := &InsertSubscriberDataArg{IMSI: HexBytes{0x12}}
			tc.mut(in)
			_, err := in.Marshal()
			if !errors.Is(err, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, err)
			}
		})
	}
}

func TestInsertSubscriberDataArg_MmeNameFQDNValidation(t *testing.T) {
	in := &InsertSubscriberDataArg{
		IMSI:    HexBytes{0x12},
		MmeName: HexBytes("short"), // < 9 octets
	}
	_, err := in.Marshal()
	if !errors.Is(err, ErrFQDNInvalidSize) {
		t.Fatalf("want ErrFQDNInvalidSize, got %v", err)
	}
}

func TestInsertSubscriberDataArg_NilSentinel(t *testing.T) {
	var a *InsertSubscriberDataArg
	_, err := a.Marshal()
	if !errors.Is(err, ErrIsdArgNil) {
		t.Fatalf("Marshal(nil) Arg: want ErrIsdArgNil, got %v", err)
	}
	var r *InsertSubscriberDataRes
	_, err = r.Marshal()
	if !errors.Is(err, ErrIsdResNil) {
		t.Fatalf("Marshal(nil) Res: want ErrIsdResNil, got %v", err)
	}
}

func TestInsertSubscriberDataArg_BadFieldSizes(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*InsertSubscriberDataArg)
		want error
	}{
		{"Category wrong size", func(a *InsertSubscriberDataArg) { a.Category = HexBytes{0x01, 0x02} }, ErrIsdCategoryInvalidSize},
		{"ChargingChars wrong size", func(a *InsertSubscriberDataArg) { a.ChargingCharacteristics = HexBytes{0x01} }, ErrIsdChargingCharsInvalidSize},
		{"CsAllocationRetentionPriority wrong", func(a *InsertSubscriberDataArg) { a.CsAllocationRetentionPriority = HexBytes{0x01, 0x02} }, ErrIsdCsAllocRetentionInvalidSize},
		{"AgeIndicator over-max", func(a *InsertSubscriberDataArg) { a.SuperChargerSupportedInHLR = make(HexBytes, 7) }, ErrIsdAgeIndicatorInvalidSize},
		{"BearerServiceList entry too long", func(a *InsertSubscriberDataArg) { a.BearerServiceList = []HexBytes{make(HexBytes, 6)} }, ErrIsdBearerServiceCodeSize},
		{"TeleserviceList entry empty", func(a *InsertSubscriberDataArg) { a.TeleserviceList = []HexBytes{{}} }, ErrIsdTeleserviceCodeSize},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := &InsertSubscriberDataArg{IMSI: HexBytes{0x12}}
			tc.mut(in)
			_, err := in.Marshal()
			if !errors.Is(err, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, err)
			}
		})
	}
}

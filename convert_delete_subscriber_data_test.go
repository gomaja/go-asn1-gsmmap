package gsmmap

import (
	"errors"
	identity "github.com/gomaja/go-3gpp-identity"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"reflect"
	"testing"
)

func hlrRoundTrip[T any](t *testing.T, in *T, parse func([]byte) (*T, error), wire string) {
	t.Helper()
	expected := mustHex(t, wire)
	out, err := any(in).(marshaler).Marshal()
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "wire", expected, out)
	got, err := parse(expected)
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "parsed", in, got)
	out, err = any(got).(marshaler).Marshal()
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "reencoded", expected, out)
}

func TestDeleteSubscriberDataMinimal(t *testing.T) {
	// SEQUENCE: [0] IMSI, eight TBCD octets.
	hlrRoundTrip(t, &DeleteSubscriberDataArg{IMSI: "001010123456789"}, ParseDeleteSubscriberDataArg, "300a800800010121436587f9")
	// Empty optional result parameter.
	hlrRoundTrip(t, &DeleteSubscriberDataRes{}, ParseDeleteSubscriberDataRes, "3000")
}

func fullDeleteSubscriberData() *DeleteSubscriberDataArg {
	return &DeleteSubscriberDataArg{IMSI: "001010123456789",
		BasicServiceList: []ExtBasicServiceCode{{ExtBearerService: HexBytes{0x11}}, {ExtTeleservice: HexBytes{0x21}}},
		SsList:           []SsCode{0x21}, RegionalSubscriptionIdentifier: ZoneCode{0xff, 0xff},
		GprsSubscriptionDataWithdraw:                   &GPRSSubscriptionDataWithdraw{ContextIDList: []int{1, 50}},
		EpsSubscriptionDataWithdraw:                    &EPSSubscriptionDataWithdraw{ContextIDList: []int{2}},
		LsaInformationWithdraw:                         &LSAInformationWithdraw{LSAIdentityList: []HexBytes{{1, 2, 3}}},
		SpecificCSIWithdraw:                            &SpecificCSIWithdraw{OCSI: true, SsCSI: true, TifCSI: true, DCSI: true, VtCSI: true, MoSmsCSI: true, MCSI: true, GprsCSI: true, MtSmsCSI: true, MgCSI: true},
		RoamingRestrictionDueToUnsupportedFeature:      true,
		VbsGroupIndication:                             true,
		VgcsGroupIndication:                            true,
		CamelSubscriptionInfoWithdraw:                  true,
		RoamingRestrictedInSgsnDueToUnsupportedFeature: true,
		GmlcListWithdraw:                               true,
		IstInformationWithdraw:                         true,
		ChargingCharacteristicsWithdraw:                true,
		StnSrWithdraw:                                  true,
		ApnOiReplacementWithdraw:                       true,
		CsgSubscriptionDeleted:                         true,
		SubscribedPeriodicTAURAUTimerWithdraw:          true,
		SubscribedPeriodicLAUTimerWithdraw:             true,
		SubscribedVsrvccWithdraw:                       true,
		VplmnCsgSubscriptionDeleted:                    true,
		AdditionalMSISDNWithdraw:                       true,
		CsToPsSRVCCWithdraw:                            true,
		ImsiGroupIdListWithdraw:                        true,
		UserPlaneIntegrityProtectionWithdraw:           true,
		DlBufferingSuggestedPacketCountWithdraw:        true,
		UeUsageTypeWithdraw:                            true,
		ResetIdsWithdraw:                               true,
		IabOperationWithdraw:                           true,
	}
}

func TestDeleteSubscriberDataFull(t *testing.T) {
	// All surfaced fields: two basic services, SS, opaque Zone Code, three list
	// alternatives, 11-bit SpecificCSI (bits 0..7,9,10), and all NULL markers.
	hlrRoundTrip(t, fullDeleteSubscriberData(), ParseDeleteSubscriberDataArg, "306a800800010121436587f9a106820111830121a20304012184008502ffff870088008900aa0830060201010201328b00ac07300504030102038d008e008f0305ff6090009100b205300302010293009400960097009500980099009a009b009c009d009e009f1f009f2000")

	for _, code := range []RegionalSubscriptionResponse{RegionalSubscriptionResponseNetworkNodeAreaRestricted, RegionalSubscriptionResponseRegionalSubscNotSupported} {
		wire := "3003800100"
		if code == RegionalSubscriptionResponseRegionalSubscNotSupported {
			wire = "3003800103"
		}
		hlrRoundTrip(t, &DeleteSubscriberDataRes{RegionalSubscriptionResponse: &code}, ParseDeleteSubscriberDataRes, wire)
	}
	// Each all-data CHOICE alternative is a NULL inside an explicit tag.
	hlrRoundTrip(t, &DeleteSubscriberDataArg{IMSI: "001010123456789", GprsSubscriptionDataWithdraw: &GPRSSubscriptionDataWithdraw{AllGPRSData: true}, EpsSubscriptionDataWithdraw: &EPSSubscriptionDataWithdraw{AllEPSData: true}, LsaInformationWithdraw: &LSAInformationWithdraw{AllLSAData: true}}, ParseDeleteSubscriberDataArg, "3016800800010121436587f9aa020500ac020500b2020500")
}

func TestSpecificCSIWithdrawBits(t *testing.T) {
	// Each permitted element must occupy its assigned position; ASN.1 SIZE
	// requires at least 8 bits even for bit 0 alone.
	fields := []string{"OCSI", "SsCSI", "TifCSI", "DCSI", "VtCSI", "MoSmsCSI", "MCSI", "GprsCSI", "MtSmsCSI", "MgCSI"}
	for i, field := range fields {
		bit := i
		if i >= 8 {
			bit++
		}
		v := &SpecificCSIWithdraw{}
		reflect.ValueOf(v).Elem().FieldByName(field).SetBool(true)
		w, err := convertSpecificCSIWithdrawToWire(v)
		if err != nil {
			t.Fatal(err)
		}
		wantEqual(t, "length", max(8, bit+1), w.BitLength)
		expected := make([]byte, (max(8, bit+1)+7)/8)
		expected[bit/8] = 0x80 >> uint(bit%8)
		wantEqual(t, "bits", expected, w.Bytes)
		wantEqual(t, "decoded", v, convertWireToSpecificCSIWithdraw(*w))
	}
}

func TestSpecificCSIWithdrawReceiverIgnores(t *testing.T) {
	for bit := 11; bit < 32; bit++ {
		for _, known := range []bool{false, true} {
			// Raw 32-bit SpecificCSI: an ignored assignment, with or without O-CSI.
			body := mustHex(t, "800800010121436587f98f050000000000")
			if known {
				body[13] = 0x80
			}
			body[13+bit/8] |= 0x80 >> uint(bit%8)
			got, err := ParseDeleteSubscriberDataArg(ussdTLV(0x30, body))
			if err != nil {
				t.Fatal(err)
			}
			wantEqual(t, "specificCSI", &SpecificCSIWithdraw{OCSI: known}, got.SpecificCSIWithdraw)
			out, err := got.Marshal()
			if !known {
				if !errors.Is(err, ErrSpecificCSIWithdrawEmpty) {
					t.Fatalf("unknown-only bits: %v", err)
				}
				continue
			}
			if err != nil {
				t.Fatal(err)
			}
			// Ignored bits disappear and the SIZE-minimal 8-bit form is sent.
			wantEqual(t, "canonical wire", mustHex(t, "300e800800010121436587f98f020080"), out)
		}
	}
}

func TestHLRCodecSentinels(t *testing.T) {
	for _, tc := range []struct {
		name string
		msg  marshaler
		want error
	}{
		{"nil deletion arg", (*DeleteSubscriberDataArg)(nil), ErrDeleteSubscriberDataArgNil},
		{"nil deletion result", (*DeleteSubscriberDataRes)(nil), ErrDeleteSubscriberDataResNil},
		{"nil reset", (*ResetArg)(nil), ErrResetArgNil},
		{"nil check arg", (*CheckIMEIArg)(nil), ErrCheckIMEIArgNil},
		{"nil check result", (*CheckIMEIRes)(nil), ErrCheckIMEIResNil},
		{"nil ready arg", (*ReadyForSMArg)(nil), ErrReadyForSMArgNil},
		{"nil ready result", (*ReadyForSMRes)(nil), ErrReadyForSMResNil},
		{"GPRS none", &DeleteSubscriberDataArg{IMSI: "001010", GprsSubscriptionDataWithdraw: &GPRSSubscriptionDataWithdraw{}}, ErrGPRSSubscriptionDataWithdrawNoAlternative},
		{"GPRS both", &DeleteSubscriberDataArg{IMSI: "001010", GprsSubscriptionDataWithdraw: &GPRSSubscriptionDataWithdraw{AllGPRSData: true, ContextIDList: []int{}}}, ErrGPRSSubscriptionDataWithdrawMultipleAlternatives},
		{"EPS none", &DeleteSubscriberDataArg{IMSI: "001010", EpsSubscriptionDataWithdraw: &EPSSubscriptionDataWithdraw{}}, ErrEPSSubscriptionDataWithdrawNoAlternative},
		{"EPS both", &DeleteSubscriberDataArg{IMSI: "001010", EpsSubscriptionDataWithdraw: &EPSSubscriptionDataWithdraw{AllEPSData: true, ContextIDList: []int{}}}, ErrEPSSubscriptionDataWithdrawMultipleAlternatives},
		{"LSA none", &DeleteSubscriberDataArg{IMSI: "001010", LsaInformationWithdraw: &LSAInformationWithdraw{}}, ErrLSAInformationWithdrawNoAlternative},
		{"LSA both", &DeleteSubscriberDataArg{IMSI: "001010", LsaInformationWithdraw: &LSAInformationWithdraw{AllLSAData: true, LSAIdentityList: []HexBytes{}}}, ErrLSAInformationWithdrawMultipleAlternatives},
		{"node none", &ResetArg{}, ErrSendingNodeNumberNoAlternative},
		{"node both", &ResetArg{SendingNodeNumber: SendingNodeNumber{HLRNumber: "12", CSSNumber: "34"}}, ErrSendingNodeNumberMultipleAlternatives},
		{"basic none", &DeleteSubscriberDataArg{IMSI: "001010", BasicServiceList: []ExtBasicServiceCode{{}}}, ErrExtBasicServiceCodeNoAlternative},
		{"basic both", &DeleteSubscriberDataArg{IMSI: "001010", BasicServiceList: []ExtBasicServiceCode{{ExtBearerService: HexBytes{0x11}, ExtTeleservice: HexBytes{0x21}}}}, ErrExtBasicServiceCodeMultipleAlternatives},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.msg.Marshal()
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
	for _, tc := range []struct {
		want error
		run  func() error
	}{
		{ErrGPRSSubscriptionDataWithdrawUnknownAlternative, func() error {
			_, err := convertWireToGPRSSubscriptionDataWithdraw(&gsm_map.GPRSSubscriptionDataWithdraw{Choice: 99})
			return err
		}},
		{ErrEPSSubscriptionDataWithdrawUnknownAlternative, func() error {
			_, err := convertWireToEPSSubscriptionDataWithdraw(&gsm_map.EPSSubscriptionDataWithdraw{Choice: 99})
			return err
		}},
		{ErrLSAInformationWithdrawUnknownAlternative, func() error {
			_, err := convertWireToLSAInformationWithdraw(&gsm_map.LSAInformationWithdraw{Choice: 99})
			return err
		}},
		{ErrSendingNodeNumberUnknownAlternative, func() error {
			_, err := convertWireToSendingNodeNumber(&gsm_map.SendingNodeNumber{Choice: 99})
			return err
		}},
	} {
		if err := tc.run(); !errors.Is(err, tc.want) {
			t.Errorf("got %v, want %v", err, tc.want)
		}
	}
}

func hlrSemLenientCases() []semLenientCase {
	reset := func() *gsm_map.ResetArg {
		return &gsm_map.ResetArg{SendingNodenumber: gsm_map.NewSendingNodeNumberHlrNumber([]byte{0x80, 0x21})}
	}
	deletion := func() *gsm_map.DeleteSubscriberDataArg {
		return &gsm_map.DeleteSubscriberDataArg{Imsi: []byte{0, 1, 1}}
	}
	return []semLenientCase{
		{ErrSpecificCSIWithdrawEmpty, asParser(ParseDeleteSubscriberDataArg), func(*testing.T) semWire {
			w := deletion()
			b := semMust(gsm_map.NewSpecificCSIWithdraw())
			w.SpecificCSIWithdraw = &b
			return w
		}},
		{ErrSpecificCSIWithdrawTCSINotApplicable, asParser(ParseDeleteSubscriberDataArg), func(*testing.T) semWire {
			w := deletion()
			b := semMust(gsm_map.NewSpecificCSIWithdraw(gsm_map.SpecificCSIWithdrawTCsi))
			w.SpecificCSIWithdraw = &b
			return w
		}},
		{ErrDeleteSubscriberDataResRegionalSubscriptionResponseInvalid, asParser(ParseDeleteSubscriberDataRes), func(*testing.T) semWire {
			c := gsm_map.RegionalSubscriptionResponseTooManyZoneCodes
			return &gsm_map.DeleteSubscriberDataRes{RegionalSubscriptionResponse: &c}
		}},
		{ErrResetArgHLRListWithCSSNumber, asParser(ParseResetArg), func(*testing.T) semWire {
			w := reset()
			w.SendingNodenumber = gsm_map.NewSendingNodeNumberCssNumber([]byte{0x80, 0x21})
			w.HlrList = &gsm_map.HLRList{Values: []gsm_map.HLRId{{0, 1, 1}}}
			return w
		}},
		{ErrResetArgHLRListWithResetIdList, asParser(ParseResetArg), func(*testing.T) semWire {
			w := reset()
			w.HlrList = &gsm_map.HLRList{Values: []gsm_map.HLRId{{0, 1, 1}}}
			w.ResetIdList = &gsm_map.ResetIdList{Values: []gsm_map.ResetId{{1}}}
			return w
		}},
		{ErrResetArgSubscriptionDataConflict, asParser(ParseResetArg), func(*testing.T) semWire {
			w := reset()
			w.ResetIdList = &gsm_map.ResetIdList{Values: []gsm_map.ResetId{{1}}}
			w.SubscriptionData = &gsm_map.InsertSubscriberDataArg{}
			w.SubscriptionDataDeletion = deletion()
			return w
		}},
		{ErrResetArgSubscriptionDataWithoutResetIdList, asParser(ParseResetArg), func(*testing.T) semWire {
			w := reset()
			w.SubscriptionData = &gsm_map.InsertSubscriberDataArg{}
			return w
		}},
		{ErrResetArgSubscriptionDataDeletionWithoutResetIdList, asParser(ParseResetArg), func(*testing.T) semWire { w := reset(); w.SubscriptionDataDeletion = deletion(); return w }},
		{ErrCheckIMEIArgBMUEFRequiresIMEISV, asParser(ParseCheckIMEIArg), func(*testing.T) semWire {
			return &gsm_map.CheckIMEIArg{Imei: []byte{0x21, 0x43, 0x65, 0x87, 9, 0x21, 0x43, 0xf0}, RequestedEquipmentInfo: semMust(gsm_map.NewRequestedEquipmentInfo(1))}
		}},
		{ErrCheckIMEIArgNoEquipmentInfoRequested, asParser(ParseCheckIMEIArg), func(*testing.T) semWire {
			return &gsm_map.CheckIMEIArg{Imei: []byte{0x21, 0x43, 0x65, 0x87, 9, 0x21, 0x43, 0xf0}, RequestedEquipmentInfo: semMust(gsm_map.NewRequestedEquipmentInfo())}
		}},
	}
}

func TestDeleteSubscriberDataSenderRules(t *testing.T) {
	for _, tc := range []struct {
		basic ExtBasicServiceCode
		want  error
	}{
		{ExtBasicServiceCode{ExtBearerService: HexBytes{0x50}}, ErrCompoundBearerServiceCodeNotAllowed},
		{ExtBasicServiceCode{ExtBearerService: HexBytes{0x58}}, ErrCompoundBearerServiceCodeNotAllowed},
		{ExtBasicServiceCode{ExtBearerService: HexBytes{0x60}}, ErrCompoundBearerServiceCodeNotAllowed},
		{ExtBasicServiceCode{ExtBearerService: HexBytes{0x68}}, ErrCompoundBearerServiceCodeNotAllowed},
		{ExtBasicServiceCode{ExtTeleservice: HexBytes{0x70}}, ErrCompoundTeleserviceCodeNotAllowed},
		{ExtBasicServiceCode{ExtTeleservice: HexBytes{0x80}}, ErrCompoundTeleserviceCodeNotAllowed},
		{ExtBasicServiceCode{ExtBearerService: HexBytes{0x11, 0}}, ErrExtBearerServiceCodeReservedOctets},
		{ExtBasicServiceCode{ExtTeleservice: HexBytes{0x21, 0}}, ErrExtTeleserviceCodeReservedOctets},
		{ExtBasicServiceCode{ExtBearerService: HexBytes{0x81}}, ErrBearerServiceCodeUnusedBit},
	} {
		// The incoming assigned or unsupported value is preserved; support is
		// contextual under §8.8.2.3 and belongs to the caller.
		tag := byte(0x82)
		value := tc.basic.ExtBearerService
		if tc.basic.ExtTeleservice != nil {
			tag = 0x83
			value = tc.basic.ExtTeleservice
		}
		body := append(mustHex(t, "800800010121436587f9"), ussdTLV(0xa1, ussdTLV(tag, value))...)
		v, err := ParseDeleteSubscriberDataArg(ussdTLV(0x30, body))
		if err != nil {
			t.Fatal(err)
		}
		wantEqual(t, "basic code preserved", []ExtBasicServiceCode{tc.basic}, v.BasicServiceList)
		if _, err = v.Marshal(); !errors.Is(err, tc.want) {
			t.Errorf("got %v, want %v", err, tc.want)
		}
	}
	// Reserved SS-Code is a sender restriction (§17.7.5).
	v, err := ParseDeleteSubscriberDataArg(mustHex(t, "300f800800010121436587f9a203040100"))
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "SS retained", []SsCode{0}, v.SsList)
	if _, err = v.Marshal(); !errors.Is(err, ErrSSCodeReserved) {
		t.Fatal(err)
	}
	for _, wire := range []string{
		// TooManyZoneCodes(1) and ZoneCodesConflict(2) are ASN.1 values but
		// cannot be sent for this operation (§8.8.2.3).
		"3003800101", "3003800102",
	} {
		v, err := ParseDeleteSubscriberDataRes(mustHex(t, wire))
		if err != nil {
			t.Fatal(err)
		}
		if _, err = v.Marshal(); !errors.Is(err, ErrDeleteSubscriberDataResRegionalSubscriptionResponseInvalid) {
			t.Fatal(err)
		}
	}
	// An unknown service code is kept even when it has no local meaning.
	hlrRoundTrip(t, &DeleteSubscriberDataArg{IMSI: "001010123456789", BasicServiceList: []ExtBasicServiceCode{{ExtTeleservice: HexBytes{0xee}}}}, ParseDeleteSubscriberDataArg, "300f800800010121436587f9a1038301ee")
}

func TestDeleteSubscriberDataConstraints(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		change                 func(*DeleteSubscriberDataArg)
		tail, path, constraint string
	}{
		{"basic list", func(v *DeleteSubscriberDataArg) { v.BasicServiceList = []ExtBasicServiceCode{} }, "a100", "basicServiceList", "SIZE (1..70)"},
		{"SS list", func(v *DeleteSubscriberDataArg) { v.SsList = []SsCode{} }, "a200", "ss-List", "SIZE (1..30)"},
		{"zone", func(v *DeleteSubscriberDataArg) { v.RegionalSubscriptionIdentifier = ZoneCode{1} }, "850101", "regionalSubscriptionIdentifier", "SIZE (2)"},
		{"GPRS list", func(v *DeleteSubscriberDataArg) {
			v.GprsSubscriptionDataWithdraw = &GPRSSubscriptionDataWithdraw{ContextIDList: []int{}}
		}, "aa023000", "gprsSubscriptionDataWithdraw.contextIdList", "SIZE (1..50)"},
		{"EPS list", func(v *DeleteSubscriberDataArg) {
			v.EpsSubscriptionDataWithdraw = &EPSSubscriptionDataWithdraw{ContextIDList: []int{}}
		}, "b2023000", "epsSubscriptionDataWithdraw.contextIdList", "SIZE (1..50)"},
		{"GPRS context", func(v *DeleteSubscriberDataArg) {
			v.GprsSubscriptionDataWithdraw = &GPRSSubscriptionDataWithdraw{ContextIDList: []int{0}}
		}, "aa053003020100", "gprsSubscriptionDataWithdraw.contextIdList.element[0]", "(1..50)"},
		{"EPS context", func(v *DeleteSubscriberDataArg) {
			v.EpsSubscriptionDataWithdraw = &EPSSubscriptionDataWithdraw{ContextIDList: []int{51}}
		}, "b2053003020133", "epsSubscriptionDataWithdraw.contextIdList.element[0]", "(1..50)"},
		{"LSA list", func(v *DeleteSubscriberDataArg) {
			v.LsaInformationWithdraw = &LSAInformationWithdraw{LSAIdentityList: []HexBytes{}}
		}, "ac023000", "lsaInformationWithdraw.lsaIdentityList", "SIZE (1..20)"},
		{"LSA identity", func(v *DeleteSubscriberDataArg) {
			v.LsaInformationWithdraw = &LSAInformationWithdraw{LSAIdentityList: []HexBytes{{1, 2}}}
		}, "ac06300404020102", "lsaInformationWithdraw.lsaIdentityList.element[0]", "SIZE (3)"},
		{"basic leaf", func(v *DeleteSubscriberDataArg) {
			v.BasicServiceList = []ExtBasicServiceCode{{ExtTeleservice: HexBytes{1, 2, 3, 4, 5, 6}}}
		}, "a1088306010203040506", "basicServiceList.element[0].ext-Teleservice", "SIZE (1..5)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := &DeleteSubscriberDataArg{IMSI: "001010123456789"}
			tc.change(v)
			_, err := v.Marshal()
			wantValidationError(t, err, tc.path, tc.constraint)
			body := append(mustHex(t, "800800010121436587f9"), mustHex(t, tc.tail)...)
			_, err = ParseDeleteSubscriberDataArg(ussdTLV(0x30, body))
			wantValidationError(t, err, tc.path, tc.constraint)
		})
	}
	// A significant bit beyond the SpecificCSI-Withdraw SIZE 32 is rejected.
	_, err := ParseDeleteSubscriberDataArg(mustHex(t, "3012800800010121436587f98f06070000000080"))
	wantValidationError(t, err, "specificCSI-Withdraw", "SIZE (8..32)")
	// The response ENUMERATED is non-extensible.
	_, err = ParseDeleteSubscriberDataRes(mustHex(t, "3003800104"))
	wantValidationError(t, err, "regionalSubscriptionResponse", "ENUMERATED {0, 1, 2, 3}")
}

func TestHLRIdentityValidation(t *testing.T) {
	for _, tc := range []struct {
		v     marshaler
		parse func([]byte) (marshaler, error)
		wire  string
	}{
		{&DeleteSubscriberDataArg{IMSI: ""}, asParser(ParseDeleteSubscriberDataArg), "30058003ffffff"},
		{&CheckIMEIArg{IMEI: "", RequestedEquipmentInfo: RequestedEquipmentInfo{EquipmentStatus: true}}, asParser(ParseCheckIMEIArg), "300e0408ffffffffffffffff03020680"},
		{&ReadyForSMArg{IMSI: ""}, asParser(ParseReadyForSMArg), "30088003ffffff0a0100"},
		{&ResetArg{SendingNodeNumber: SendingNodeNumber{HLRNumber: "12"}, HLRList: []string{""}}, asParser(ParseResetArg), "300b0402802130050403ffffff"},
	} {
		if _, err := tc.v.Marshal(); !errors.Is(err, identity.ErrEmpty) {
			t.Fatalf("Marshal: %v", err)
		}
		if _, err := tc.parse(mustHex(t, tc.wire)); !errors.Is(err, identity.ErrEmpty) {
			t.Fatalf("Parse: %v", err)
		}
	}
}

func TestDeleteSubscriberDataNestedDecodeErrors(t *testing.T) {
	for _, tc := range []struct {
		v    *gsm_map.DeleteSubscriberDataArg
		want error
	}{
		{&gsm_map.DeleteSubscriberDataArg{Imsi: semAddrIMSI, BasicServiceList: &gsm_map.BasicServiceList{Values: []gsm_map.ExtBasicServiceCode{{Choice: 99}}}}, ErrExtBasicServiceCodeUnknownAlternative},
		{&gsm_map.DeleteSubscriberDataArg{Imsi: semAddrIMSI, GprsSubscriptionDataWithdraw: &gsm_map.GPRSSubscriptionDataWithdraw{Choice: 99}}, ErrGPRSSubscriptionDataWithdrawUnknownAlternative},
		{&gsm_map.DeleteSubscriberDataArg{Imsi: semAddrIMSI, EpsSubscriptionDataWithdraw: &gsm_map.EPSSubscriptionDataWithdraw{Choice: 99}}, ErrEPSSubscriptionDataWithdrawUnknownAlternative},
		{&gsm_map.DeleteSubscriberDataArg{Imsi: semAddrIMSI, LsaInformationWithdraw: &gsm_map.LSAInformationWithdraw{Choice: 99}}, ErrLSAInformationWithdrawUnknownAlternative},
	} {
		if _, err := convertWireToDeleteSubscriberDataArg(tc.v); !errors.Is(err, tc.want) {
			t.Errorf("got %v, want %v", err, tc.want)
		}
	}
}

func TestDeleteSubscriberDataResultEnumCodecOwnership(t *testing.T) {
	// Non-extensible ENUMERATED membership belongs to the ASN.1 codec;
	// the operation-specific sender restriction covers values 1 and 2 only.
	_, err := (&DeleteSubscriberDataRes{RegionalSubscriptionResponse: new(RegionalSubscriptionResponse(4))}).Marshal()
	wantValidationError(t, err, "regionalSubscriptionResponse", "ENUMERATED {0, 1, 2, 3}")
}

// 3GPP TS 29.002 V19.1.0 §17.7.1: "Bit 8 and bits 11-13 are only applicable for
// the NoteSubscriberDataModified operation". Parse keeps bit 8, with no discard
// rule to apply, and Marshal refuses it.
func TestSpecificCSIWithdrawTCSI(t *testing.T) {
	// IMSI; [15] with 9 bits: bit 8 only, then bits 0 and 8.
	for wire, want := range map[string]*SpecificCSIWithdraw{
		"300f800800010121436587f98f03070080": {TCSI: true},
		"300f800800010121436587f98f03078080": {OCSI: true, TCSI: true},
	} {
		got, err := ParseDeleteSubscriberDataArg(mustHex(t, wire))
		if err != nil {
			t.Fatal(err)
		}
		wantEqual(t, "SpecificCSIWithdraw", want, got.SpecificCSIWithdraw)
		if _, err := got.Marshal(); !errors.Is(err, ErrSpecificCSIWithdrawTCSINotApplicable) {
			t.Errorf("Marshal: err = %v, want %v", err, ErrSpecificCSIWithdrawTCSINotApplicable)
		}
	}
}

// Each NULL marker maps to its own tag (3GPP TS 29.002 V19.1.0 §17.7.1). Every
// wire is the IMSI [0] and one marker.
func TestDeleteSubscriberDataMarkers(t *testing.T) {
	for _, c := range []struct{ field, wire string }{
		{"RoamingRestrictionDueToUnsupportedFeature", "300c800800010121436587f98400"},
		{"VbsGroupIndication", "300c800800010121436587f98700"},
		{"VgcsGroupIndication", "300c800800010121436587f98800"},
		{"CamelSubscriptionInfoWithdraw", "300c800800010121436587f98900"},
		{"RoamingRestrictedInSgsnDueToUnsupportedFeature", "300c800800010121436587f98b00"},
		{"GmlcListWithdraw", "300c800800010121436587f98d00"},
		{"IstInformationWithdraw", "300c800800010121436587f98e00"},
		{"ChargingCharacteristicsWithdraw", "300c800800010121436587f99000"},
		{"StnSrWithdraw", "300c800800010121436587f99100"},
		{"ApnOiReplacementWithdraw", "300c800800010121436587f99300"},
		{"CsgSubscriptionDeleted", "300c800800010121436587f99400"},
		{"SubscribedVsrvccWithdraw", "300c800800010121436587f99500"},
		{"SubscribedPeriodicTAURAUTimerWithdraw", "300c800800010121436587f99600"},
		{"SubscribedPeriodicLAUTimerWithdraw", "300c800800010121436587f99700"},
		{"VplmnCsgSubscriptionDeleted", "300c800800010121436587f99800"},
		{"AdditionalMSISDNWithdraw", "300c800800010121436587f99900"},
		{"CsToPsSRVCCWithdraw", "300c800800010121436587f99a00"},
		{"ImsiGroupIdListWithdraw", "300c800800010121436587f99b00"},
		{"UserPlaneIntegrityProtectionWithdraw", "300c800800010121436587f99c00"},
		{"DlBufferingSuggestedPacketCountWithdraw", "300c800800010121436587f99d00"},
		{"UeUsageTypeWithdraw", "300c800800010121436587f99e00"},
		{"ResetIdsWithdraw", "300d800800010121436587f99f1f00"},
		{"IabOperationWithdraw", "300d800800010121436587f99f2000"},
	} {
		t.Run(c.field, func(t *testing.T) {
			in := &DeleteSubscriberDataArg{IMSI: "001010123456789"}
			reflect.ValueOf(in).Elem().FieldByName(c.field).SetBool(true)
			hlrRoundTrip(t, in, ParseDeleteSubscriberDataArg, c.wire)
		})
	}
}

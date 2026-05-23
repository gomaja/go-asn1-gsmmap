// convert_report_sm_delivery_status_test.go
//
// Tests for ReportSMDeliveryStatus (opCode 47): the Arg/Res converters
// and their Marshal()/Parse() entry points. Round-trip, BER round-trip,
// and targeted negative cases.
package gsmmap

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

func minimalReportSMDeliveryStatus() *ReportSMDeliveryStatus {
	return &ReportSMDeliveryStatus{
		MSISDN:               "31612345678",
		MSISDNNature:         0x10,
		MSISDNPlan:           0x01,
		ServiceCentreAddress: "31600000000",
		SCANature:            0x10,
		SCAPlan:              0x01,
		SmDeliveryOutcome:    SmDeliveryAbsentSubscriber,
	}
}

func TestReportSMDeliveryStatusRoundTrip(t *testing.T) {
	diag := 3 // purgedMS
	addDiag := 0
	ipDiag := 2
	smsf3 := 1
	smsfN := 5
	addOutcome := SmDeliverySuccessfulTransfer
	ipOutcome := SmDeliveryMemoryCapacityExceeded
	smsf3Outcome := SmDeliveryAbsentSubscriber
	smsfNOutcome := SmDeliverySuccessfulTransfer

	cases := []struct {
		name string
		in   *ReportSMDeliveryStatus
	}{
		{"minimal absent-subscriber", minimalReportSMDeliveryStatus()},
		{"successful transfer (clears waiting)", &ReportSMDeliveryStatus{
			MSISDN:               "31612345678",
			MSISDNNature:         0x10,
			MSISDNPlan:           0x01,
			ServiceCentreAddress: "31600000000",
			SCANature:            0x10,
			SCAPlan:              0x01,
			SmDeliveryOutcome:    SmDeliverySuccessfulTransfer,
		}},
		{"with diagnostic + GPRS + IMSI", &ReportSMDeliveryStatus{
			MSISDN:                                 "31612345678",
			MSISDNNature:                           0x10,
			MSISDNPlan:                             0x01,
			ServiceCentreAddress:                   "31600000000",
			SCANature:                              0x10,
			SCAPlan:                                0x01,
			SmDeliveryOutcome:                      SmDeliveryAbsentSubscriber,
			AbsentSubscriberDiagnosticSM:           &diag,
			GprsSupportIndicator:                   true,
			DeliveryOutcomeIndicator:               true,
			AdditionalSMDeliveryOutcome:            &addOutcome,
			AdditionalAbsentSubscriberDiagnosticSM: &addDiag,
			IMSI:                                   "204080000000001",
			SingleAttemptDelivery:                  true,
		}},
		{"full (all variants populated)", &ReportSMDeliveryStatus{
			MSISDN:                                 "31612345678",
			MSISDNNature:                           0x10,
			MSISDNPlan:                             0x01,
			ServiceCentreAddress:                   "31600000000",
			SCANature:                              0x10,
			SCAPlan:                                0x01,
			SmDeliveryOutcome:                      SmDeliveryAbsentSubscriber,
			AbsentSubscriberDiagnosticSM:           &diag,
			GprsSupportIndicator:                   true,
			DeliveryOutcomeIndicator:               true,
			AdditionalSMDeliveryOutcome:            &addOutcome,
			AdditionalAbsentSubscriberDiagnosticSM: &addDiag,
			IMSI:                                   "204080000000001",
			SingleAttemptDelivery:                  true,
			CorrelationID: &SriSmCorrelationID{
				HlrID:   HexBytes{0xAA, 0xBB},
				SipUriA: HexBytes{0xCC, 0xDD},
				SipUriB: HexBytes{0xEE, 0xFF},
			},
			IpSmGwIndicator:                         true,
			IpSmGwSmDeliveryOutcome:                 &ipOutcome,
			IpSmGwAbsentSubscriberDiagnosticSM:      &ipDiag,
			Smsf3gppDeliveryOutcomeIndicator:        true,
			Smsf3gppDeliveryOutcome:                 &smsf3Outcome,
			Smsf3gppAbsentSubscriberDiagnosticSM:    &smsf3,
			SmsfNon3gppDeliveryOutcomeIndicator:     true,
			SmsfNon3gppDeliveryOutcome:              &smsfNOutcome,
			SmsfNon3gppAbsentSubscriberDiagnosticSM: &smsfN,
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertReportSMDeliveryStatusToArg(tc.in)
			if err != nil {
				t.Fatalf("toArg: %v", err)
			}
			got, err := convertArgToReportSMDeliveryStatus(wire)
			if err != nil {
				t.Fatalf("toStruct: %v", err)
			}
			if !reflect.DeepEqual(tc.in, got) {
				t.Errorf("round-trip mismatch:\n in=%+v\ngot=%+v", tc.in, got)
			}
		})
	}
}

func TestReportSMDeliveryStatusBERRoundTrip(t *testing.T) {
	in := minimalReportSMDeliveryStatus()
	diag := 2
	in.AbsentSubscriberDiagnosticSM = &diag
	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseReportSMDeliveryStatus(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !reflect.DeepEqual(in, got) {
		t.Errorf("BER round-trip mismatch:\n in=%+v\ngot=%+v", in, got)
	}
}

func TestReportSMDeliveryStatusEncodeNegative(t *testing.T) {
	bad := 999
	cases := []struct {
		name string
		mut  func(r *ReportSMDeliveryStatus)
		want error
	}{
		{"nil arg", nil, ErrReportSMDeliveryStatusNil},
		{"empty MSISDN", func(r *ReportSMDeliveryStatus) { r.MSISDN = "" }, ErrReportSMDeliveryStatusMSISDNEmpty},
		{"empty ServiceCentreAddress", func(r *ReportSMDeliveryStatus) { r.ServiceCentreAddress = "" }, ErrReportSMDeliveryStatusSCAEmpty},
		{"outcome out of range", func(r *ReportSMDeliveryStatus) { r.SmDeliveryOutcome = SmDeliveryOutcome(9) }, ErrReportSMDeliveryStatusOutcomeInvalid},
		{"diagnostic out of range", func(r *ReportSMDeliveryStatus) { r.AbsentSubscriberDiagnosticSM = &bad }, ErrAbsentSubscriberDiagnosticSMOutOfRange},
		{"IMSI too short", func(r *ReportSMDeliveryStatus) { r.IMSI = "1234" }, ErrReportSMDeliveryStatusIMSIInvalidSize},
		{"IMSI too long", func(r *ReportSMDeliveryStatus) { r.IMSI = "1234567890123456" }, ErrReportSMDeliveryStatusIMSIInvalidSize},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var in *ReportSMDeliveryStatus
			if tc.mut != nil {
				in = minimalReportSMDeliveryStatus()
				tc.mut(in)
			}
			_, err := convertReportSMDeliveryStatusToArg(in)
			if !errors.Is(err, tc.want) {
				t.Errorf("want %v, got %v", tc.want, err)
			}
		})
	}
}

func TestReportSMDeliveryStatusDecodeNegative(t *testing.T) {
	t.Run("nil wire", func(t *testing.T) {
		_, err := convertArgToReportSMDeliveryStatus(nil)
		if !errors.Is(err, ErrReportSMDeliveryStatusNil) {
			t.Errorf("want ErrReportSMDeliveryStatusNil, got %v", err)
		}
	})
	// validWireArg is a minimal decodable wire arg (valid MSISDN + SCA +
	// in-range outcome); each subtest mutates one field to exercise a
	// specific decode-side rejection.
	validWireArg := func() *gsm_map.ReportSMDeliveryStatusArg {
		return &gsm_map.ReportSMDeliveryStatusArg{
			Msisdn:               gsm_map.ISDNAddressString{0x91, 0x13, 0x16, 0x32, 0x54, 0x76, 0x98},
			ServiceCentreAddress: gsm_map.AddressString{0x91, 0x13, 0x06, 0x00, 0x00, 0x00},
			SmDeliveryOutcome:    gsm_map.SMDeliveryOutcomeAbsentSubscriber,
		}
	}
	emptyAddr := func() []byte { return []byte{0x91} } // header only, no TBCD digits
	diag999 := gsm_map.AbsentSubscriberDiagnosticSM(999)
	imsiShort := gsm_map.IMSI{0x21, 0xf3} // 3 digits after TBCD decode (< 5)

	cases := []struct {
		name string
		mut  func(w *gsm_map.ReportSMDeliveryStatusArg)
		want error
	}{
		{"outcome out of range", func(w *gsm_map.ReportSMDeliveryStatusArg) { w.SmDeliveryOutcome = gsm_map.SMDeliveryOutcome(7) }, ErrReportSMDeliveryStatusOutcomeInvalid},
		{"MSISDN present but empty", func(w *gsm_map.ReportSMDeliveryStatusArg) { w.Msisdn = emptyAddr() }, ErrReportSMDeliveryStatusMSISDNDecodedEmpty},
		{"SCA present but empty", func(w *gsm_map.ReportSMDeliveryStatusArg) { w.ServiceCentreAddress = emptyAddr() }, ErrReportSMDeliveryStatusSCADecodedEmpty},
		{"diagnostic out of range on wire", func(w *gsm_map.ReportSMDeliveryStatusArg) { w.AbsentSubscriberDiagnosticSM = &diag999 }, ErrAbsentSubscriberDiagnosticSMOutOfRange},
		{"IMSI invalid size on wire", func(w *gsm_map.ReportSMDeliveryStatusArg) { v := imsiShort; w.Imsi = &v }, ErrReportSMDeliveryStatusIMSIInvalidSize},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := validWireArg()
			tc.mut(w)
			_, err := convertArgToReportSMDeliveryStatus(w)
			if !errors.Is(err, tc.want) {
				t.Errorf("want %v, got %v", tc.want, err)
			}
		})
	}

	t.Run("Res StoredMSISDN present but empty", func(t *testing.T) {
		ea := gsm_map.ISDNAddressString(emptyAddr())
		w := &gsm_map.ReportSMDeliveryStatusRes{StoredMSISDN: &ea}
		_, err := convertResToReportSMDeliveryStatusRes(w)
		if !errors.Is(err, ErrReportSMDeliveryStatusResStoredMSISDNEmpty) {
			t.Errorf("want ErrReportSMDeliveryStatusResStoredMSISDNEmpty, got %v", err)
		}
	})
}

// ============================================================================
// ReportSMDeliveryStatus-Res
// ============================================================================

func TestReportSMDeliveryStatusResRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   *ReportSMDeliveryStatusRes
	}{
		{"empty", &ReportSMDeliveryStatusRes{}},
		{"with StoredMSISDN", &ReportSMDeliveryStatusRes{
			StoredMSISDN:       "31698765432",
			StoredMSISDNNature: 0x10,
			StoredMSISDNPlan:   0x01,
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertReportSMDeliveryStatusResToRes(tc.in)
			if err != nil {
				t.Fatalf("toRes: %v", err)
			}
			got, err := convertResToReportSMDeliveryStatusRes(wire)
			if err != nil {
				t.Fatalf("toStruct: %v", err)
			}
			if !reflect.DeepEqual(tc.in, got) {
				t.Errorf("round-trip mismatch:\n in=%+v\ngot=%+v", tc.in, got)
			}
		})
	}
}

func TestReportSMDeliveryStatusResBERRoundTrip(t *testing.T) {
	in := &ReportSMDeliveryStatusRes{
		StoredMSISDN:       "31698765432",
		StoredMSISDNNature: 0x10,
		StoredMSISDNPlan:   0x01,
	}
	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := ParseReportSMDeliveryStatusRes(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !reflect.DeepEqual(in, got) {
		t.Errorf("BER round-trip mismatch:\n in=%+v\ngot=%+v", in, got)
	}
}

func TestReportSMDeliveryStatusResNilNegative(t *testing.T) {
	if _, err := convertReportSMDeliveryStatusResToRes(nil); !errors.Is(err, ErrReportSMDeliveryStatusResNil) {
		t.Errorf("encode nil: want ErrReportSMDeliveryStatusResNil, got %v", err)
	}
	if _, err := convertResToReportSMDeliveryStatusRes(nil); !errors.Is(err, ErrReportSMDeliveryStatusResNil) {
		t.Errorf("decode nil: want ErrReportSMDeliveryStatusResNil, got %v", err)
	}
}

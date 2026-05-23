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
		Msisdn:               "31612345678",
		MsisdnNature:         0x10,
		MsisdnPlan:           0x01,
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
			Msisdn:               "31612345678",
			MsisdnNature:         0x10,
			MsisdnPlan:           0x01,
			ServiceCentreAddress: "31600000000",
			SCANature:            0x10,
			SCAPlan:              0x01,
			SmDeliveryOutcome:    SmDeliverySuccessfulTransfer,
		}},
		{"with diagnostic + GPRS + IMSI", &ReportSMDeliveryStatus{
			Msisdn:                                 "31612345678",
			MsisdnNature:                           0x10,
			MsisdnPlan:                             0x01,
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
			Msisdn:                                 "31612345678",
			MsisdnNature:                           0x10,
			MsisdnPlan:                             0x01,
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
		{"empty Msisdn", func(r *ReportSMDeliveryStatus) { r.Msisdn = "" }, ErrReportSMDeliveryStatusMsisdnEmpty},
		{"empty ServiceCentreAddress", func(r *ReportSMDeliveryStatus) { r.ServiceCentreAddress = "" }, ErrReportSMDeliveryStatusSCAEmpty},
		{"outcome out of range", func(r *ReportSMDeliveryStatus) { r.SmDeliveryOutcome = SmDeliveryOutcome(9) }, ErrReportSMDeliveryStatusOutcomeInvalid},
		{"diagnostic out of range", func(r *ReportSMDeliveryStatus) { r.AbsentSubscriberDiagnosticSM = &bad }, ErrAbsentSubscriberDiagnosticSMOutOfRange},
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
	t.Run("outcome out of range on wire", func(t *testing.T) {
		w := &gsm_map.ReportSMDeliveryStatusArg{
			Msisdn:               gsm_map.ISDNAddressString{0x91, 0x13, 0x16, 0x32, 0x54, 0x76, 0x98},
			ServiceCentreAddress: gsm_map.AddressString{0x91, 0x13, 0x06, 0x00, 0x00, 0x00},
			SmDeliveryOutcome:    gsm_map.SMDeliveryOutcome(7),
		}
		_, err := convertArgToReportSMDeliveryStatus(w)
		if !errors.Is(err, ErrReportSMDeliveryStatusOutcomeInvalid) {
			t.Errorf("want ErrReportSMDeliveryStatusOutcomeInvalid, got %v", err)
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

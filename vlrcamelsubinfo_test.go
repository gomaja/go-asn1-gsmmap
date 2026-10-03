// vlrcamelsubinfo_test.go
//
// Tests for VlrCamelSubscriptionInfo and its novel sub-types (ISD PR C).
// OCSI/TCSI/DCSI/OBcsm/TBcsm criteria are covered by the existing camel
// tests; this file focuses on SSCSI, MCSI, SMSCSI, SMSCAMELTDPData, and
// MTSmsCAMELTDPCriteria, plus the orchestrating VlrCamelSubscriptionInfo
// converter.
package gsmmap

import (
	"errors"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/google/go-cmp/cmp"
)

// gsmMapDefaultSMSHandling is a test helper for injecting spec-invalid
// DefaultSMSHandling values onto the wire struct, so the decoder's
// lenient exception-handling rules can be exercised. Takes int64 so
// callers can build values that exceed platform int on 32-bit builds.
func gsmMapDefaultSMSHandling(v int64) gsm_map.DefaultSMSHandling {
	return gsm_map.DefaultSMSHandling(v)
}

// --- SSCSI ---

func TestSSCSIRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   *SSCSI
	}{
		{
			name: "minimal",
			in: &SSCSI{
				SsEventList:         []SsCode{0x31}, // ect
				GsmSCFAddress:       "31611111111",
				GsmSCFAddressNature: 16, GsmSCFAddressPlan: 1,
			},
		},
		{
			name: "multipleEvents",
			in: &SSCSI{
				SsEventList:         []SsCode{0x31, 0x51, 0x24, 0x44}, // ect, multiPTY, cd, ccbs
				GsmSCFAddress:       "31622222222",
				GsmSCFAddressNature: 16, GsmSCFAddressPlan: 1,
				NotificationToCSE: true,
				CsiActive:         true,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertSSCSIToWire(tc.in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			got, err := convertWireToSSCSI(wire)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if diff := cmp.Diff(tc.in, got); diff != "" {
				t.Errorf("round-trip mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSSCSIValidation(t *testing.T) {
	t.Run("emptyEventList", func(t *testing.T) {
		_, err := strictWire(convertSSCSIToWire(&SSCSI{GsmSCFAddress: "111"}))
		if !matchesConstraint(err, "ss-EventList", "SIZE (1..10)") {
			t.Errorf("want BER constraint error, got %v", err)
		}
	})
	t.Run("tooManyEvents", func(t *testing.T) {
		big := make([]SsCode, 11)
		_, err := strictWire(convertSSCSIToWire(&SSCSI{SsEventList: big, GsmSCFAddress: "1"}))
		if !matchesConstraint(err, "ss-EventList", "SIZE (1..10)") {
			t.Errorf("want BER constraint error, got %v", err)
		}
	})
	t.Run("missingGsmSCF", func(t *testing.T) {
		_, err := strictWire(convertSSCSIToWire(&SSCSI{SsEventList: []SsCode{0x31}}))
		if !errors.Is(err, ErrCamelMissingGsmSCFAddress) {
			t.Errorf("want ErrCamelMissingGsmSCFAddress, got %v", err)
		}
	})
}

// --- MCSI ---

func TestMCSIRoundTrip(t *testing.T) {
	in := &MCSI{
		MobilityTriggers:    []byte{0x00, 0x01, 0x02}, // LU-same-VLR, LU-other-VLR, IMSI-Attach
		ServiceKey:          42,
		GsmSCFAddress:       "31633333333",
		GsmSCFAddressNature: 16, GsmSCFAddressPlan: 1,
		NotificationToCSE: true,
	}
	wire, err := convertMCSIToWire(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, err := convertWireToMCSI(wire)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("round-trip mismatch (-want +got):\n%s", diff)
	}
}

func TestMCSIValidation(t *testing.T) {
	t.Run("emptyTriggers", func(t *testing.T) {
		_, err := strictWire(convertMCSIToWire(&MCSI{GsmSCFAddress: "1"}))
		if !matchesConstraint(err, "mobilityTriggers", "SIZE (1..10)") {
			t.Errorf("want BER constraint error, got %v", err)
		}
	})
	t.Run("tooManyTriggers", func(t *testing.T) {
		big := make([]byte, 11)
		_, err := strictWire(convertMCSIToWire(&MCSI{MobilityTriggers: big, GsmSCFAddress: "1"}))
		if !matchesConstraint(err, "mobilityTriggers", "SIZE (1..10)") {
			t.Errorf("want BER constraint error, got %v", err)
		}
	})
	t.Run("serviceKeyOutOfRange", func(t *testing.T) {
		_, err := strictWire(convertMCSIToWire(&MCSI{
			MobilityTriggers: []byte{0x00},
			ServiceKey:       -1,
			GsmSCFAddress:    "1",
		}))
		if !matchesConstraint(err, "serviceKey", "(0..2147483647)") {
			t.Errorf("want BER constraint error, got %v", err)
		}
	})
	t.Run("missingGsmSCF", func(t *testing.T) {
		_, err := strictWire(convertMCSIToWire(&MCSI{MobilityTriggers: []byte{0x00}}))
		if !errors.Is(err, ErrCamelMissingGsmSCFAddress) {
			t.Errorf("want ErrCamelMissingGsmSCFAddress, got %v", err)
		}
	})
}

// --- SMSCAMELTDPData + SMSCSI ---

func TestSMSCSIRoundTrip(t *testing.T) {
	cch := 3
	for _, tdp := range []SMSTriggerDetectionPoint{moSMSTriggerDetectionPoint, mtSMSTriggerDetectionPoint} {
		in := &SMSCSI{
			SmsCAMELTDPDataList: []SMSCAMELTDPData{
				{
					SmsTriggerDetectionPoint: tdp,
					ServiceKey:               100,
					GsmSCFAddress:            "31644444444",
					GsmSCFAddressNature:      16, GsmSCFAddressPlan: 1,
					DefaultSMSHandling: DefaultSMSHandlingContinueTransaction,
				},
			},
			CamelCapabilityHandling: &cch,
		}
		wire, err := convertSMSCSIToWire(in, tdp)
		if err != nil {
			t.Fatalf("tdp %d: encode: %v", tdp, err)
		}
		got, err := convertWireToSMSCSI(wire, tdp)
		if err != nil {
			t.Fatalf("tdp %d: decode: %v", tdp, err)
		}
		if diff := cmp.Diff(in, got); diff != "" {
			t.Errorf("tdp %d: round-trip mismatch (-want +got):\n%s", tdp, diff)
		}
	}
}

func TestSMSCSIValidation(t *testing.T) {
	cch := 2
	t.Run("oversizeTDPList", func(t *testing.T) {
		// Eleven entries repeat the one TDP an MO-SMS-CSI lists; the
		// one-instance rule (3GPP TS 29.002 V19.1.0 §17.7.1) rejects the
		// second before the codec's SIZE (1..10).
		big := make([]SMSCAMELTDPData, 11)
		for i := range big {
			big[i] = SMSCAMELTDPData{
				SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsCollectedInfo,
				GsmSCFAddress:            "1",
				DefaultSMSHandling:       DefaultSMSHandlingContinueTransaction,
			}
		}
		_, err := strictWire(convertSMSCSIToWire(&SMSCSI{
			SmsCAMELTDPDataList:     big,
			CamelCapabilityHandling: &cch,
		}, moSMSTriggerDetectionPoint))
		if !errors.Is(err, ErrCamelDuplicateTriggerDetectionPoint) {
			t.Errorf("want ErrCamelDuplicateTriggerDetectionPoint, got %v", err)
		}
	})
	t.Run("invalidTriggerDetectionPoint", func(t *testing.T) {
		_, err := convertSMSCAMELTDPDataToWire(&SMSCAMELTDPData{
			SmsTriggerDetectionPoint: SMSTriggerDetectionPoint(99),
			GsmSCFAddress:            "1",
			DefaultSMSHandling:       DefaultSMSHandlingContinueTransaction,
		}, moSMSTriggerDetectionPoint)
		if !errors.Is(err, ErrCamelInvalidSMSTriggerDetectionPoint) {
			t.Errorf("want ErrCamelInvalidSMSTriggerDetectionPoint, got %v", err)
		}
	})
	t.Run("invalidDefaultSMSHandling", func(t *testing.T) {
		_, err := convertSMSCAMELTDPDataToWire(&SMSCAMELTDPData{
			SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsCollectedInfo,
			GsmSCFAddress:            "1",
			DefaultSMSHandling:       DefaultSMSHandling(-1),
		}, moSMSTriggerDetectionPoint)
		if !errors.Is(err, ErrCamelInvalidDefaultSMSHandling) {
			t.Errorf("want ErrCamelInvalidDefaultSMSHandling, got %v", err)
		}
	})
}

// --- MTSmsCAMELTDPCriteria ---

func TestMTSmsCAMELTDPCriteriaRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   *MTSmsCAMELTDPCriteria
	}{
		{
			name: "noTpduCriterion",
			in:   &MTSmsCAMELTDPCriteria{SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsDeliveryRequest},
		},
		{
			name: "withTpduCriterion",
			in: &MTSmsCAMELTDPCriteria{
				SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsDeliveryRequest,
				TpduTypeCriterion:        []MTSMSTPDUType{MTSMSTPDUTypeSmsDELIVER, MTSMSTPDUTypeSmsSTATUSREPORT},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertMTSmsCAMELTDPCriteriaToWire(tc.in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			got, err := convertWireToMTSmsCAMELTDPCriteria(&wire)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if diff := cmp.Diff(tc.in, got); diff != "" {
				t.Errorf("round-trip mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMTSmsCAMELTDPCriteriaValidation(t *testing.T) {
	t.Run("invalidTriggerDetectionPoint", func(t *testing.T) {
		_, err := convertMTSmsCAMELTDPCriteriaToWire(&MTSmsCAMELTDPCriteria{
			SmsTriggerDetectionPoint: SMSTriggerDetectionPoint(99),
		})
		if !errors.Is(err, ErrCamelInvalidSMSTriggerDetectionPoint) {
			t.Errorf("want ErrCamelInvalidSMSTriggerDetectionPoint, got %v", err)
		}
	})
	t.Run("tpduListTooLong", func(t *testing.T) {
		big := make([]MTSMSTPDUType, 6)
		_, err := strictWire(convertMTSmsCAMELTDPCriteriaToWire(&MTSmsCAMELTDPCriteria{
			SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsDeliveryRequest,
			TpduTypeCriterion:        big,
		}))
		if !matchesConstraint(err, "tpdu-TypeCriterion", "SIZE (1..5)") {
			t.Errorf("want BER constraint error, got %v", err)
		}
	})
	t.Run("invalidTpduType", func(t *testing.T) {
		_, err := convertMTSmsCAMELTDPCriteriaToWire(&MTSmsCAMELTDPCriteria{
			SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsDeliveryRequest,
			TpduTypeCriterion:        []MTSMSTPDUType{99},
		})
		if !errors.Is(err, ErrCamelInvalidMTSMSTPDUType) {
			t.Errorf("want ErrCamelInvalidMTSMSTPDUType, got %v", err)
		}
	})
}

// --- Lenient DefaultSMSHandling decode ---

func TestSMSCAMELTDPDataLenientDefaultSMSHandling(t *testing.T) {
	// Per TS 29.002 §8.8.1: values 2..31 → continueTransaction;
	// values > 31 → releaseTransaction. Apply the mapping in int64 space
	// so wire values exceeding platform int still follow the rule on
	// 32-bit builds (locked in by the >MaxInt32 case below).
	cases := []struct {
		name string
		wire int64
		want DefaultSMSHandling
	}{
		{"continue", 0, DefaultSMSHandlingContinueTransaction},
		{"release", 1, DefaultSMSHandlingReleaseTransaction},
		{"reserved2Maps", 2, DefaultSMSHandlingContinueTransaction},
		{"reserved31Maps", 31, DefaultSMSHandlingContinueTransaction},
		{"reserved32Maps", 32, DefaultSMSHandlingReleaseTransaction},
		{"reserved200Maps", 200, DefaultSMSHandlingReleaseTransaction},
		// Values larger than platform int on 32-bit builds must still
		// map to releaseTransaction per spec, not error on the narrow.
		{"hugeMapsToRelease", 1 << 33, DefaultSMSHandlingReleaseTransaction},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Encode a valid entry, then replace the wire DefaultSMSHandling
			// before decoding.
			in := &SMSCAMELTDPData{
				SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsCollectedInfo,
				ServiceKey:               1,
				GsmSCFAddress:            "1",
				DefaultSMSHandling:       DefaultSMSHandlingContinueTransaction,
			}
			w, err := convertSMSCAMELTDPDataToWire(in, moSMSTriggerDetectionPoint)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			// Force wire value
			w.DefaultSMSHandling = gsmMapDefaultSMSHandling(tc.wire)
			got, err := convertWireToSMSCAMELTDPData(&w, moSMSTriggerDetectionPoint)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if got.DefaultSMSHandling != tc.want {
				t.Errorf("wire=%d: got %d, want %d", tc.wire, got.DefaultSMSHandling, tc.want)
			}
		})
	}
}

// --- VlrCamelSubscriptionInfo orchestration ---

func TestVlrCamelSubscriptionInfoFullStressRoundTrip(t *testing.T) {
	cch := 4
	in := &VlrCamelSubscriptionInfo{
		OCSI: &OCSI{
			OBcsmCamelTDPDataList: []OBcsmCamelTDPData{{
				OBcsmTriggerDetectionPoint: OBcsmTriggerCollectedInfo,
				ServiceKey:                 1,
				GsmSCFAddress:              "31611111111",
				GsmSCFAddressNature:        16, GsmSCFAddressPlan: 1,
				DefaultCallHandling: DefaultCallHandlingContinueCall,
			}},
			CamelCapabilityHandling: &cch,
		},
		SsCSI: &SSCSI{
			SsEventList:         []SsCode{0x31},
			GsmSCFAddress:       "31622222222",
			GsmSCFAddressNature: 16, GsmSCFAddressPlan: 1,
		},
		TifCSI: true,
		MCSI: &MCSI{
			MobilityTriggers:    []byte{0x00, 0x02},
			ServiceKey:          7,
			GsmSCFAddress:       "31633333333",
			GsmSCFAddressNature: 16, GsmSCFAddressPlan: 1,
		},
		MoSmsCSI: &SMSCSI{
			SmsCAMELTDPDataList: []SMSCAMELTDPData{{
				SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsCollectedInfo,
				ServiceKey:               11,
				GsmSCFAddress:            "31644444444",
				GsmSCFAddressNature:      16, GsmSCFAddressPlan: 1,
				DefaultSMSHandling: DefaultSMSHandlingContinueTransaction,
			}},
			CamelCapabilityHandling: &cch,
		},
		VtCSI: &TCSI{
			TBcsmCamelTDPDataList: []TBcsmCamelTDPData{{
				TBcsmTriggerDetectionPoint: TBcsmTriggerTermAttemptAuthorized,
				ServiceKey:                 5,
				GsmSCFAddress:              "31655555555",
				GsmSCFAddressNature:        16, GsmSCFAddressPlan: 1,
				DefaultCallHandling: DefaultCallHandlingContinueCall,
			}},
			CamelCapabilityHandling: &cch,
		},
		MtSmsCSI: &SMSCSI{
			SmsCAMELTDPDataList: []SMSCAMELTDPData{{
				SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsDeliveryRequest,
				ServiceKey:               13,
				GsmSCFAddress:            "31666666666",
				GsmSCFAddressNature:      16, GsmSCFAddressPlan: 1,
				DefaultSMSHandling: DefaultSMSHandlingReleaseTransaction,
			}},
			CamelCapabilityHandling: &cch,
		},
		MtSmsCAMELTDPCriteriaList: []MTSmsCAMELTDPCriteria{{
			SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsDeliveryRequest,
			TpduTypeCriterion:        []MTSMSTPDUType{MTSMSTPDUTypeSmsDELIVER},
		}},
	}

	wire, err := convertVlrCamelSubscriptionInfoToWire(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, err := convertWireToVlrCamelSubscriptionInfo(wire)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("round-trip mismatch (-want +got):\n%s", diff)
	}
}

func TestVlrCamelSubscriptionInfoMinimalRoundTrip(t *testing.T) {
	// Every field optional per spec; an empty struct must round-trip
	// to an empty struct without errors.
	in := &VlrCamelSubscriptionInfo{}
	wire, err := convertVlrCamelSubscriptionInfoToWire(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, err := convertWireToVlrCamelSubscriptionInfo(wire)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("round-trip mismatch (-want +got):\n%s", diff)
	}
}

func TestVlrCamelSubscriptionInfoCriteriaListBoundaries(t *testing.T) {
	for n := 6; n <= 11; n++ {
		criteria := make([]MTSmsCAMELTDPCriteria, n)
		for i := range criteria {
			criteria[i] = MTSmsCAMELTDPCriteria{SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsDeliveryRequest}
		}
		_, err := strictWire(convertVlrCamelSubscriptionInfoToWire(&VlrCamelSubscriptionInfo{MtSmsCAMELTDPCriteriaList: criteria}))
		if n <= 10 {
			if err != nil {
				t.Errorf("%d criteria should be accepted: %v", n, err)
			}
		} else {
			wantConstraintError(t, err, "mt-smsCAMELTDP-CriteriaList", "SIZE (1..10)")
		}
	}

	// PR #29 pattern: a non-nil but empty optional list violates SIZE(1..N)
	// and must be rejected at both encode and decode.
	t.Run("OBcsmCriteriaListEmptyRejected", func(t *testing.T) {
		_, err := strictWire(convertVlrCamelSubscriptionInfoToWire(&VlrCamelSubscriptionInfo{
			OBcsmCamelTDPCriteriaList: []OBcsmCamelTDPCriteria{},
		}))
		if !matchesConstraint(err, "o-BcsmCamelTDP-CriteriaList", "SIZE (1..10)") {
			t.Errorf("want BER constraint error, got %v", err)
		}
	})
	t.Run("TBcsmCriteriaListEmptyRejected", func(t *testing.T) {
		_, err := strictWire(convertVlrCamelSubscriptionInfoToWire(&VlrCamelSubscriptionInfo{
			TBcsmCamelTDPCriteriaList: []TBcsmCamelTDPCriteria{},
		}))
		if !matchesConstraint(err, "t-BCSM-CAMEL-TDP-CriteriaList", "SIZE (1..10)") {
			t.Errorf("want BER constraint error, got %v", err)
		}
	})
	t.Run("MtSmsCAMELTDPCriteriaListEmptyRejected", func(t *testing.T) {
		_, err := strictWire(convertVlrCamelSubscriptionInfoToWire(&VlrCamelSubscriptionInfo{
			MtSmsCAMELTDPCriteriaList: []MTSmsCAMELTDPCriteria{},
		}))
		if !matchesConstraint(err, "mt-smsCAMELTDP-CriteriaList", "SIZE (1..10)") {
			t.Errorf("want BER constraint error, got %v", err)
		}
	})
	t.Run("TpduTypeCriterionEmptyRejected", func(t *testing.T) {
		_, err := strictWire(convertMTSmsCAMELTDPCriteriaToWire(&MTSmsCAMELTDPCriteria{
			SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsDeliveryRequest,
			TpduTypeCriterion:        []MTSMSTPDUType{}, // non-nil, empty
		}))
		if !matchesConstraint(err, "tpdu-TypeCriterion", "SIZE (1..5)") {
			t.Errorf("want BER constraint error, got %v", err)
		}
	})
}

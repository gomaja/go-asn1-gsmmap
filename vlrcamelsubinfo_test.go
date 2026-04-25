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

	"github.com/google/go-cmp/cmp"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
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
				SsEventList:   []SsCode{0x31}, // ect
				GsmSCFAddress: "31611111111",
				GsmSCFNature:  16, GsmSCFPlan: 1,
			},
		},
		{
			name: "multipleEvents",
			in: &SSCSI{
				SsEventList:   []SsCode{0x31, 0x51, 0x24, 0x44}, // ect, multiPTY, cd, ccbs
				GsmSCFAddress: "31622222222",
				GsmSCFNature:  16, GsmSCFPlan: 1,
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
		_, err := convertSSCSIToWire(&SSCSI{GsmSCFAddress: "111"})
		if !errors.Is(err, ErrCamelInvalidSSEventListSize) {
			t.Errorf("want ErrCamelInvalidSSEventListSize, got %v", err)
		}
	})
	t.Run("tooManyEvents", func(t *testing.T) {
		big := make([]SsCode, 11)
		_, err := convertSSCSIToWire(&SSCSI{SsEventList: big, GsmSCFAddress: "1"})
		if !errors.Is(err, ErrCamelInvalidSSEventListSize) {
			t.Errorf("want ErrCamelInvalidSSEventListSize, got %v", err)
		}
	})
	t.Run("missingGsmSCF", func(t *testing.T) {
		_, err := convertSSCSIToWire(&SSCSI{SsEventList: []SsCode{0x31}})
		if !errors.Is(err, ErrCamelMissingGsmSCFAddress) {
			t.Errorf("want ErrCamelMissingGsmSCFAddress, got %v", err)
		}
	})
}

// --- MCSI ---

func TestMCSIRoundTrip(t *testing.T) {
	in := &MCSI{
		MobilityTriggers: []byte{0x00, 0x01, 0x02}, // LU-same-VLR, LU-other-VLR, IMSI-Attach
		ServiceKey:       42,
		GsmSCFAddress:    "31633333333",
		GsmSCFNature:     16, GsmSCFPlan: 1,
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
		_, err := convertMCSIToWire(&MCSI{GsmSCFAddress: "1"})
		if !errors.Is(err, ErrCamelInvalidMobilityTriggersSize) {
			t.Errorf("want ErrCamelInvalidMobilityTriggersSize, got %v", err)
		}
	})
	t.Run("tooManyTriggers", func(t *testing.T) {
		big := make([]byte, 11)
		_, err := convertMCSIToWire(&MCSI{MobilityTriggers: big, GsmSCFAddress: "1"})
		if !errors.Is(err, ErrCamelInvalidMobilityTriggersSize) {
			t.Errorf("want ErrCamelInvalidMobilityTriggersSize, got %v", err)
		}
	})
	t.Run("serviceKeyOutOfRange", func(t *testing.T) {
		_, err := convertMCSIToWire(&MCSI{
			MobilityTriggers: []byte{0x00},
			ServiceKey:       -1,
			GsmSCFAddress:    "1",
		})
		if !errors.Is(err, ErrCamelInvalidServiceKey) {
			t.Errorf("want ErrCamelInvalidServiceKey, got %v", err)
		}
	})
	t.Run("missingGsmSCF", func(t *testing.T) {
		_, err := convertMCSIToWire(&MCSI{MobilityTriggers: []byte{0x00}})
		if !errors.Is(err, ErrCamelMissingGsmSCFAddress) {
			t.Errorf("want ErrCamelMissingGsmSCFAddress, got %v", err)
		}
	})
}

// --- SMSCAMELTDPData + SMSCSI ---

func TestSMSCSIRoundTrip(t *testing.T) {
	cch := 3
	in := &SMSCSI{
		SmsCAMELTDPDataList: []SMSCAMELTDPData{
			{
				SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsCollectedInfo,
				ServiceKey:               100,
				GsmSCFAddress:            "31644444444",
				GsmSCFNature:             16, GsmSCFPlan: 1,
				DefaultSMSHandling: DefaultSMSHandlingContinueTransaction,
			},
			{
				SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsDeliveryRequest,
				ServiceKey:               200,
				GsmSCFAddress:            "31655555555",
				GsmSCFNature:             16, GsmSCFPlan: 1,
				DefaultSMSHandling: DefaultSMSHandlingReleaseTransaction,
			},
		},
		CamelCapabilityHandling: &cch,
	}
	wire, err := convertSMSCSIToWire(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, err := convertWireToSMSCSI(wire)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if diff := cmp.Diff(in, got); diff != "" {
		t.Errorf("round-trip mismatch (-want +got):\n%s", diff)
	}
}

func TestSMSCSIValidation(t *testing.T) {
	cch := 2
	t.Run("missingTDPList", func(t *testing.T) {
		// Empty TDP list: per spec §8.8.1 the field is required, so the
		// "missing" sentinel applies (not the size sentinel).
		_, err := convertSMSCSIToWire(&SMSCSI{CamelCapabilityHandling: &cch})
		if !errors.Is(err, ErrCamelSMSCSIMissingTDPData) {
			t.Errorf("want ErrCamelSMSCSIMissingTDPData, got %v", err)
		}
	})
	t.Run("oversizeTDPList", func(t *testing.T) {
		// 11 entries: above SIZE(1..10), so the size sentinel applies.
		big := make([]SMSCAMELTDPData, 11)
		for i := range big {
			big[i] = SMSCAMELTDPData{
				SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsCollectedInfo,
				GsmSCFAddress:            "1",
				DefaultSMSHandling:       DefaultSMSHandlingContinueTransaction,
			}
		}
		_, err := convertSMSCSIToWire(&SMSCSI{
			SmsCAMELTDPDataList:     big,
			CamelCapabilityHandling: &cch,
		})
		if !errors.Is(err, ErrCamelInvalidSMSTDPDataListSize) {
			t.Errorf("want ErrCamelInvalidSMSTDPDataListSize, got %v", err)
		}
	})
	t.Run("missingCapabilityHandling", func(t *testing.T) {
		_, err := convertSMSCSIToWire(&SMSCSI{
			SmsCAMELTDPDataList: []SMSCAMELTDPData{{
				SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsCollectedInfo,
				GsmSCFAddress:            "1",
				DefaultSMSHandling:       DefaultSMSHandlingContinueTransaction,
			}},
		})
		if !errors.Is(err, ErrCamelSMSCSIMissingCapabilityHandling) {
			t.Errorf("want ErrCamelSMSCSIMissingCapabilityHandling, got %v", err)
		}
	})
	t.Run("invalidTriggerDetectionPoint", func(t *testing.T) {
		_, err := convertSMSCAMELTDPDataToWire(&SMSCAMELTDPData{
			SmsTriggerDetectionPoint: SMSTriggerDetectionPoint(99),
			GsmSCFAddress:            "1",
			DefaultSMSHandling:       DefaultSMSHandlingContinueTransaction,
		})
		if !errors.Is(err, ErrCamelInvalidSMSTriggerDetectionPoint) {
			t.Errorf("want ErrCamelInvalidSMSTriggerDetectionPoint, got %v", err)
		}
	})
	t.Run("invalidDefaultSMSHandling", func(t *testing.T) {
		_, err := convertSMSCAMELTDPDataToWire(&SMSCAMELTDPData{
			SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsCollectedInfo,
			GsmSCFAddress:            "1",
			DefaultSMSHandling:       DefaultSMSHandling(-1),
		})
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
		_, err := convertMTSmsCAMELTDPCriteriaToWire(&MTSmsCAMELTDPCriteria{
			SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsDeliveryRequest,
			TpduTypeCriterion:        big,
		})
		if !errors.Is(err, ErrCamelInvalidTPDUTypeCriterionSize) {
			t.Errorf("want ErrCamelInvalidTPDUTypeCriterionSize, got %v", err)
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
			w, err := convertSMSCAMELTDPDataToWire(in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			// Force wire value
			w.DefaultSMSHandling = gsmMapDefaultSMSHandling(tc.wire)
			got, err := convertWireToSMSCAMELTDPData(&w)
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
			SsEventList:   []SsCode{0x31},
			GsmSCFAddress: "31622222222",
			GsmSCFNature:  16, GsmSCFPlan: 1,
		},
		TifCSI: true,
		MCSI: &MCSI{
			MobilityTriggers: []byte{0x00, 0x02},
			ServiceKey:       7,
			GsmSCFAddress:    "31633333333",
			GsmSCFNature:     16, GsmSCFPlan: 1,
		},
		MoSmsCSI: &SMSCSI{
			SmsCAMELTDPDataList: []SMSCAMELTDPData{{
				SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsCollectedInfo,
				ServiceKey:               11,
				GsmSCFAddress:            "31644444444",
				GsmSCFNature:             16, GsmSCFPlan: 1,
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
				GsmSCFNature:             16, GsmSCFPlan: 1,
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
	t.Run("MtSmsCAMELTDPCriteriaListTooLong", func(t *testing.T) {
		big := make([]MTSmsCAMELTDPCriteria, 6)
		for i := range big {
			big[i] = MTSmsCAMELTDPCriteria{SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsDeliveryRequest}
		}
		_, err := convertVlrCamelSubscriptionInfoToWire(&VlrCamelSubscriptionInfo{MtSmsCAMELTDPCriteriaList: big})
		if !errors.Is(err, ErrCamelInvalidMTSmsCAMELCriteriaSize) {
			t.Errorf("want ErrCamelInvalidMTSmsCAMELCriteriaSize, got %v", err)
		}
	})

	// PR #29 pattern: a non-nil but empty optional list violates SIZE(1..N)
	// and must be rejected at both encode and decode.
	t.Run("OBcsmCriteriaListEmptyRejected", func(t *testing.T) {
		_, err := convertVlrCamelSubscriptionInfoToWire(&VlrCamelSubscriptionInfo{
			OBcsmCamelTDPCriteriaList: []OBcsmCamelTDPCriteria{},
		})
		if !errors.Is(err, ErrCamelInvalidCriteriaListSize) {
			t.Errorf("want ErrCamelInvalidCriteriaListSize, got %v", err)
		}
	})
	t.Run("TBcsmCriteriaListEmptyRejected", func(t *testing.T) {
		_, err := convertVlrCamelSubscriptionInfoToWire(&VlrCamelSubscriptionInfo{
			TBcsmCamelTDPCriteriaList: []TBcsmCamelTDPCriteria{},
		})
		if !errors.Is(err, ErrCamelInvalidCriteriaListSize) {
			t.Errorf("want ErrCamelInvalidCriteriaListSize, got %v", err)
		}
	})
	t.Run("MtSmsCAMELTDPCriteriaListEmptyRejected", func(t *testing.T) {
		_, err := convertVlrCamelSubscriptionInfoToWire(&VlrCamelSubscriptionInfo{
			MtSmsCAMELTDPCriteriaList: []MTSmsCAMELTDPCriteria{},
		})
		if !errors.Is(err, ErrCamelInvalidMTSmsCAMELCriteriaSize) {
			t.Errorf("want ErrCamelInvalidMTSmsCAMELCriteriaSize, got %v", err)
		}
	})
	t.Run("TpduTypeCriterionEmptyRejected", func(t *testing.T) {
		_, err := convertMTSmsCAMELTDPCriteriaToWire(&MTSmsCAMELTDPCriteria{
			SmsTriggerDetectionPoint: SMSTriggerDetectionPointSmsDeliveryRequest,
			TpduTypeCriterion:        []MTSMSTPDUType{}, // non-nil, empty
		})
		if !errors.Is(err, ErrCamelInvalidTPDUTypeCriterionSize) {
			t.Errorf("want ErrCamelInvalidTPDUTypeCriterionSize, got %v", err)
		}
	})
}

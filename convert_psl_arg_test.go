// convert_psl_arg_test.go
//
// Tests for the top-level ProvideSubscriberLocationArg (opCode 83)
// converter and the Marshal/Parse entry points wired in marshal.go /
// parse.go.
package gsmmap

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// Minimal PSL-Arg: just the two mandatory fields (LocationType,
// MlcNumber). Round-trip + Marshal/Parse.
func TestProvideSubscriberLocationArgMinimalRoundTrip(t *testing.T) {
	in := &ProvideSubscriberLocationArg{
		LocationType: LocationType{
			LocationEstimateType: LocationEstimateCurrentLocation,
		},
		MlcNumber:       "31612345678",
		MlcNumberNature: 0x10, // International
		MlcNumberPlan:   0x01, // ISDN
	}
	wire, err := convertProvideSubscriberLocationArgToWire(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	out, err := convertWireToProvideSubscriberLocationArg(wire)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("round-trip mismatch:\n in=%+v\nout=%+v", in, out)
	}

	// Marshal/Parse over BER should also round-trip.
	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	parsed, err := ParseProvideSubscriberLocation(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !reflect.DeepEqual(in, parsed) {
		t.Errorf("Marshal/Parse mismatch:\n in=%+v\nout=%+v", in, parsed)
	}
}

// Full population: every optional field set. Locks the assembly of all
// the sub-converters from D1/D2/D3.
func TestProvideSubscriberLocationArgFullPopulationRoundTrip(t *testing.T) {
	internalID := LCSClientBroadcastService
	occ := OccurrenceMultipleTimeEvent
	intv := IntervalTime(60)
	tech := RANTechnologyUmts
	serviceType := int64(42)

	in := &ProvideSubscriberLocationArg{
		LocationType: LocationType{
			LocationEstimateType: LocationEstimateActivateDeferredLocation,
			DeferredLocationEventType: &DeferredLocationEventType{
				MsAvailable: true, PeriodicLDR: true,
			},
		},
		MlcNumber:       "31611111111",
		MlcNumberNature: 0x10,
		MlcNumberPlan:   0x01,

		LcsClientID: &LCSClientID{
			LcsClientType:       LCSClientTypeEmergencyServices,
			LcsClientInternalID: &internalID,
		},
		PrivacyOverride: true,
		IMSI:            "001010123456789",
		MSISDN:          "31622222222",
		MSISDNNature:    0x10,
		MSISDNPlan:      0x01,
		LMSI:            HexBytes{0x01, 0x02, 0x03, 0x04},
		IMEI:            "490154203237518",
		LcsPriority:     LCSPriority{0x00},
		LcsQoS: &LCSQoS{
			HorizontalAccuracy:        HexBytes{0x10},
			VerticalCoordinateRequest: true,
			VerticalAccuracy:          HexBytes{0x20},
			ResponseTime: &ResponseTime{
				ResponseTimeCategory: ResponseTimeDelaytolerant,
			},
			VelocityRequest: true,
		},
		SupportedGADShapes: &SupportedGADShapes{
			EllipsoidPoint: true, Polygon: true,
		},
		LcsReferenceNumber: LCSReferenceNumber{0x42},
		LcsServiceTypeID:   &serviceType,
		LcsCodeword: &LCSCodeword{
			DataCodingScheme:  0x0f,
			LcsCodewordString: HexBytes{0x01, 0x02, 0x03},
		},
		LcsPrivacyCheck: &LCSPrivacyCheck{
			CallSessionUnrelated: PrivacyCheckAllowedWithNotification,
		},
		AreaEventInfo: &AreaEventInfo{
			AreaDefinition: AreaDefinition{
				AreaList: AreaList{
					{AreaType: AreaTypeCountryCode, AreaIdentification: HexBytes{0x01, 0x02}},
				},
			},
			OccurrenceInfo: &occ,
			IntervalTime:   &intv,
		},
		HGmlcAddress:              "192.168.1.1",
		MoLrShortCircuitIndicator: true,
		PeriodicLDRInfo: &PeriodicLDRInfo{
			ReportingAmount:   10,
			ReportingInterval: 60,
		},
		ReportingPLMNList: &ReportingPLMNList{
			PlmnListPrioritized: true,
			PlmnList: PLMNList{
				{
					PlmnId:        HexBytes{0x32, 0xf4, 0x10},
					RanTechnology: &tech,
				},
			},
		},
	}

	wire, err := convertProvideSubscriberLocationArgToWire(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	out, err := convertWireToProvideSubscriberLocationArg(wire)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("round-trip mismatch:\n in=%+v\nout=%+v", in, out)
	}

	// BER round-trip via Marshal/Parse.
	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	parsed, err := ParseProvideSubscriberLocation(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !reflect.DeepEqual(in, parsed) {
		t.Errorf("Marshal/Parse mismatch:\n in=%+v\nout=%+v", in, parsed)
	}
}

// =============================================================================
// Validation: mandatory fields
// =============================================================================

func TestProvideSubscriberLocationArgNilRejected(t *testing.T) {
	_, err := convertProvideSubscriberLocationArgToWire(nil)
	if !errors.Is(err, ErrPSLArgNil) {
		t.Errorf("encode nil: want ErrPSLArgNil, got %v", err)
	}
	_, err = convertWireToProvideSubscriberLocationArg(nil)
	if !errors.Is(err, ErrPSLArgNil) {
		t.Errorf("decode nil: want ErrPSLArgNil, got %v", err)
	}
}

func TestProvideSubscriberLocationArgEmptyMlcNumberRejected(t *testing.T) {
	_, err := convertProvideSubscriberLocationArgToWire(&ProvideSubscriberLocationArg{
		LocationType: LocationType{LocationEstimateType: LocationEstimateCurrentLocation},
	})
	if !errors.Is(err, ErrPSLArgMlcNumberEmpty) {
		t.Errorf("encode empty MlcNumber: want ErrPSLArgMlcNumberEmpty, got %v", err)
	}
}

// =============================================================================
// Validation: optional field sizes / ranges
// =============================================================================

// IMSI must be 5..15 BCD digits per TS 29.002 (TBCD-STRING SIZE 3..8
// octets per ITU E.212). Reject under-min and over-max on encode; the
// decode path applies the same check after tbcd.Decode.
func TestProvideSubscriberLocationArgIMSIDigitCountValidation(t *testing.T) {
	base := func() *ProvideSubscriberLocationArg {
		return &ProvideSubscriberLocationArg{
			LocationType: LocationType{LocationEstimateType: LocationEstimateCurrentLocation},
			MlcNumber:    "31612345678", MlcNumberNature: 0x10, MlcNumberPlan: 0x01,
		}
	}
	a := base()
	a.IMSI = "1234" // 4 digits — under the 5-digit minimum
	if _, err := convertProvideSubscriberLocationArgToWire(a); !errors.Is(err, ErrPSLArgIMSIInvalidSize) {
		t.Errorf("IMSI=4 digits: want ErrPSLArgIMSIInvalidSize, got %v", err)
	}
	a = base()
	a.IMSI = "1234567890123456" // 16 digits — over the 15-digit maximum
	if _, err := convertProvideSubscriberLocationArgToWire(a); !errors.Is(err, ErrPSLArgIMSIInvalidSize) {
		t.Errorf("IMSI=16 digits: want ErrPSLArgIMSIInvalidSize, got %v", err)
	}
}

// IMEI must be exactly 15 BCD digits per 3GPP TS 23.003.
func TestProvideSubscriberLocationArgIMEIDigitCountValidation(t *testing.T) {
	base := func() *ProvideSubscriberLocationArg {
		return &ProvideSubscriberLocationArg{
			LocationType: LocationType{LocationEstimateType: LocationEstimateCurrentLocation},
			MlcNumber:    "31612345678", MlcNumberNature: 0x10, MlcNumberPlan: 0x01,
		}
	}
	a := base()
	a.IMEI = "12345678901234" // 14 digits
	if _, err := convertProvideSubscriberLocationArgToWire(a); !errors.Is(err, ErrPSLArgIMEIInvalidSize) {
		t.Errorf("IMEI=14 digits: want ErrPSLArgIMEIInvalidSize, got %v", err)
	}
	a = base()
	a.IMEI = "1234567890123456" // 16 digits
	if _, err := convertProvideSubscriberLocationArgToWire(a); !errors.Is(err, ErrPSLArgIMEIInvalidSize) {
		t.Errorf("IMEI=16 digits: want ErrPSLArgIMEIInvalidSize, got %v", err)
	}
}

func TestProvideSubscriberLocationArgLMSISizeValidation(t *testing.T) {
	a := &ProvideSubscriberLocationArg{
		LocationType: LocationType{LocationEstimateType: LocationEstimateCurrentLocation},
		MlcNumber:    "31612345678", MlcNumberNature: 0x10, MlcNumberPlan: 0x01,
		LMSI: HexBytes{0x01, 0x02, 0x03}, // 3 octets — must be exactly 4
	}
	if _, err := convertProvideSubscriberLocationArgToWire(a); !errors.Is(err, ErrPSLArgLMSIInvalidSize) {
		t.Errorf("LMSI=3 octets: want ErrPSLArgLMSIInvalidSize, got %v", err)
	}
}

func TestProvideSubscriberLocationArgLcsServiceTypeIDOutOfRange(t *testing.T) {
	bad := int64(128)
	a := &ProvideSubscriberLocationArg{
		LocationType:     LocationType{LocationEstimateType: LocationEstimateCurrentLocation},
		MlcNumber:        "31612345678", MlcNumberNature: 0x10, MlcNumberPlan: 0x01,
		LcsServiceTypeID: &bad,
	}
	if _, err := convertProvideSubscriberLocationArgToWire(a); !errors.Is(err, ErrPSLArgLcsServiceTypeIDOutOfRange) {
		t.Errorf("LcsServiceTypeID=128: want ErrPSLArgLcsServiceTypeIDOutOfRange, got %v", err)
	}

	negative := int64(-1)
	a.LcsServiceTypeID = &negative
	if _, err := convertProvideSubscriberLocationArgToWire(a); !errors.Is(err, ErrPSLArgLcsServiceTypeIDOutOfRange) {
		t.Errorf("LcsServiceTypeID=-1: want ErrPSLArgLcsServiceTypeIDOutOfRange, got %v", err)
	}
}

func TestProvideSubscriberLocationArgLcsPrioritySizeValidation(t *testing.T) {
	a := &ProvideSubscriberLocationArg{
		LocationType: LocationType{LocationEstimateType: LocationEstimateCurrentLocation},
		MlcNumber:    "31612345678", MlcNumberNature: 0x10, MlcNumberPlan: 0x01,
		LcsPriority:  LCSPriority{0x01, 0x02}, // 2 octets — must be 1
	}
	if _, err := convertProvideSubscriberLocationArgToWire(a); !errors.Is(err, ErrLCSPriorityInvalidSize) {
		t.Errorf("LcsPriority=2 octets: want ErrLCSPriorityInvalidSize, got %v", err)
	}
}

func TestProvideSubscriberLocationArgLcsReferenceNumberSizeValidation(t *testing.T) {
	a := &ProvideSubscriberLocationArg{
		LocationType:       LocationType{LocationEstimateType: LocationEstimateCurrentLocation},
		MlcNumber:          "31612345678", MlcNumberNature: 0x10, MlcNumberPlan: 0x01,
		LcsReferenceNumber: LCSReferenceNumber{0x01, 0x02}, // 2 octets — must be 1
	}
	if _, err := convertProvideSubscriberLocationArgToWire(a); !errors.Is(err, ErrLCSReferenceNumberInvalidSize) {
		t.Errorf("LcsReferenceNumber=2 octets: want ErrLCSReferenceNumberInvalidSize, got %v", err)
	}
}

// Round-trip fidelity: optional string-based fields decoded from the
// wire that yield empty digits cannot round-trip through the public
// API. Reject at decode rather than silently dropping the field on
// the next encode.
func TestProvideSubscriberLocationArgMSISDNDecodedEmptyRejected(t *testing.T) {
	// AddressString header (extension=1, nature=international(1),
	// plan=ISDN(1)) with no TBCD digits.
	emptyAddr := gsm_map.ISDNAddressString{0x91}
	mlc := gsm_map.ISDNAddressString{0x91, 0x13, 0x16, 0x32, 0x54, 0x76, 0x98} // 31612345678
	w := &gsm_map.ProvideSubscriberLocationArg{
		LocationType: gsm_map.LocationType{LocationEstimateType: gsm_map.LocationEstimateTypeCurrentLocation},
		MlcNumber:    mlc,
		Msisdn:       &emptyAddr,
	}
	_, err := convertWireToProvideSubscriberLocationArg(w)
	if !errors.Is(err, ErrPSLArgMSISDNDecodedEmpty) {
		t.Errorf("MSISDN empty digits: want ErrPSLArgMSISDNDecodedEmpty, got %v", err)
	}
}

// Mandatory MlcNumber that decodes to empty digits is rejected — there
// is no way to round-trip a present-but-empty MlcNumber through the
// string-based public API.
func TestProvideSubscriberLocationArgMlcNumberDecodedEmptyRejected(t *testing.T) {
	// AddressString header (extension=1, nature=international(1),
	// plan=ISDN(1)) with no TBCD digits.
	emptyAddr := gsm_map.ISDNAddressString{0x91}
	w := &gsm_map.ProvideSubscriberLocationArg{
		LocationType: gsm_map.LocationType{LocationEstimateType: gsm_map.LocationEstimateTypeCurrentLocation},
		MlcNumber:    emptyAddr,
	}
	_, err := convertWireToProvideSubscriberLocationArg(w)
	if !errors.Is(err, ErrPSLArgMlcNumberDecodedEmpty) {
		t.Errorf("MlcNumber empty digits: want ErrPSLArgMlcNumberDecodedEmpty, got %v", err)
	}
}

// IMSI / IMEI present-but-empty wire fields are rejected on decode for
// round-trip fidelity (parallel to the MSISDN/MlcNumber tests above).
// A zero-length wire octet string decodes to "" after tbcd.Decode,
// which would silently collapse to "absent" on re-encode without this
// guard.
func TestProvideSubscriberLocationArgIMSIDecodedEmptyRejected(t *testing.T) {
	mlc := gsm_map.ISDNAddressString{0x91, 0x13, 0x16, 0x32, 0x54, 0x76, 0x98}
	emptyImsi := gsm_map.IMSI{} // zero octets → "" after Decode
	w := &gsm_map.ProvideSubscriberLocationArg{
		LocationType: gsm_map.LocationType{LocationEstimateType: gsm_map.LocationEstimateTypeCurrentLocation},
		MlcNumber:    mlc,
		Imsi:         &emptyImsi,
	}
	_, err := convertWireToProvideSubscriberLocationArg(w)
	if !errors.Is(err, ErrPSLArgIMSIDecodedEmpty) {
		t.Errorf("IMSI zero-length: want ErrPSLArgIMSIDecodedEmpty, got %v", err)
	}
}

func TestProvideSubscriberLocationArgIMEIDecodedEmptyRejected(t *testing.T) {
	mlc := gsm_map.ISDNAddressString{0x91, 0x13, 0x16, 0x32, 0x54, 0x76, 0x98}
	emptyImei := gsm_map.IMEI{} // zero octets → "" after Decode
	w := &gsm_map.ProvideSubscriberLocationArg{
		LocationType: gsm_map.LocationType{LocationEstimateType: gsm_map.LocationEstimateTypeCurrentLocation},
		MlcNumber:    mlc,
		Imei:         &emptyImei,
	}
	_, err := convertWireToProvideSubscriberLocationArg(w)
	if !errors.Is(err, ErrPSLArgIMEIDecodedEmpty) {
		t.Errorf("IMEI zero-length: want ErrPSLArgIMEIDecodedEmpty, got %v", err)
	}
}

// Decode-side size / range validation: every encode-path size/range
// check has a symmetric guard in convertWireToProvideSubscriberLocationArg.
// These tests construct wire args with out-of-range values directly and
// assert the matching sentinel via errors.Is.
func TestProvideSubscriberLocationArgDecodeSizeRangeValidation(t *testing.T) {
	mlc := gsm_map.ISDNAddressString{0x91, 0x13, 0x16, 0x32, 0x54, 0x76, 0x98}
	mkBase := func() *gsm_map.ProvideSubscriberLocationArg {
		return &gsm_map.ProvideSubscriberLocationArg{
			LocationType: gsm_map.LocationType{LocationEstimateType: gsm_map.LocationEstimateTypeCurrentLocation},
			MlcNumber:    mlc,
		}
	}

	// IMSI 16 digits (over 15-digit max). 8 octets = 16 nibbles, no
	// trailing filler → 16-digit decode.
	t.Run("IMSI over-max", func(t *testing.T) {
		w := mkBase()
		imsi := gsm_map.IMSI{0x21, 0x43, 0x65, 0x87, 0x09, 0x21, 0x43, 0x65}
		w.Imsi = &imsi
		_, err := convertWireToProvideSubscriberLocationArg(w)
		if !errors.Is(err, ErrPSLArgIMSIInvalidSize) {
			t.Errorf("want ErrPSLArgIMSIInvalidSize, got %v", err)
		}
	})

	// IMEI 14 digits (one short of fixed 15). Even digit count, no
	// filler nibble required per 3GPP TBCD encoding.
	t.Run("IMEI wrong digit count", func(t *testing.T) {
		w := mkBase()
		imei := gsm_map.IMEI{0x21, 0x43, 0x65, 0x87, 0x09, 0x21, 0x43} // 14 digits, no filler
		w.Imei = &imei
		_, err := convertWireToProvideSubscriberLocationArg(w)
		if !errors.Is(err, ErrPSLArgIMEIInvalidSize) {
			t.Errorf("want ErrPSLArgIMEIInvalidSize, got %v", err)
		}
	})

	// LMSI must be exactly 4 octets.
	t.Run("LMSI wrong size", func(t *testing.T) {
		w := mkBase()
		lmsi := gsm_map.LMSI{0x01, 0x02, 0x03} // 3 octets
		w.Lmsi = &lmsi
		_, err := convertWireToProvideSubscriberLocationArg(w)
		if !errors.Is(err, ErrPSLArgLMSIInvalidSize) {
			t.Errorf("want ErrPSLArgLMSIInvalidSize, got %v", err)
		}
	})

	// LcsPriority must be exactly 1 octet.
	t.Run("LcsPriority wrong size", func(t *testing.T) {
		w := mkBase()
		pri := gsm_map.LCSPriority{0x01, 0x02} // 2 octets
		w.LcsPriority = &pri
		_, err := convertWireToProvideSubscriberLocationArg(w)
		if !errors.Is(err, ErrLCSPriorityInvalidSize) {
			t.Errorf("want ErrLCSPriorityInvalidSize, got %v", err)
		}
	})

	// LcsReferenceNumber must be exactly 1 octet.
	t.Run("LcsReferenceNumber wrong size", func(t *testing.T) {
		w := mkBase()
		ref := gsm_map.LCSReferenceNumber{0x01, 0x02} // 2 octets
		w.LcsReferenceNumber = &ref
		_, err := convertWireToProvideSubscriberLocationArg(w)
		if !errors.Is(err, ErrLCSReferenceNumberInvalidSize) {
			t.Errorf("want ErrLCSReferenceNumberInvalidSize, got %v", err)
		}
	})

	// LcsServiceTypeID must be 0..127 (exception path: wire INTEGER 128).
	t.Run("LcsServiceTypeID over-max", func(t *testing.T) {
		w := mkBase()
		sid := gsm_map.LCSServiceTypeID(128)
		w.LcsServiceTypeID = &sid
		_, err := convertWireToProvideSubscriberLocationArg(w)
		if !errors.Is(err, ErrPSLArgLcsServiceTypeIDOutOfRange) {
			t.Errorf("want ErrPSLArgLcsServiceTypeIDOutOfRange, got %v", err)
		}
	})
}

// Sub-converter errors must propagate via fmt.Errorf %w wrapping.
func TestProvideSubscriberLocationArgSubConverterErrorsPropagate(t *testing.T) {
	// Out-of-range LocationEstimateType: should surface
	// ErrLocationEstimateTypeInvalid (raised by convertLocationTypeToWire).
	a := &ProvideSubscriberLocationArg{
		LocationType: LocationType{LocationEstimateType: 99},
		MlcNumber:    "31612345678", MlcNumberNature: 0x10, MlcNumberPlan: 0x01,
	}
	_, err := convertProvideSubscriberLocationArgToWire(a)
	if !errors.Is(err, ErrLocationEstimateTypeInvalid) {
		t.Errorf("LocationEstimateType=99: want ErrLocationEstimateTypeInvalid, got %v", err)
	}
}

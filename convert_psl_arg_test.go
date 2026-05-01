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

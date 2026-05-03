// convert_psl_res_test.go
//
// Tests for ProvideSubscriberLocationRes (opCode 83) and the
// ServingNodeAddress / CellIdOrSai CHOICE codecs.
package gsmmap

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// =============================================================================
// ServingNodeAddress CHOICE codec
// =============================================================================

func TestServingNodeAddressMscNumberRoundTrip(t *testing.T) {
	in := &ServingNodeAddress{
		MscNumber:       "31611111111",
		MscNumberNature: 0x10,
		MscNumberPlan:   0x01,
	}
	wire, err := convertServingNodeAddressToWire(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if wire.Choice != gsm_map.ServingNodeAddressChoiceMscNumber {
		t.Errorf("Choice: want MscNumber, got %d", wire.Choice)
	}
	out, err := convertWireToServingNodeAddress(wire)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("round-trip mismatch:\n in=%+v\nout=%+v", in, out)
	}
}

func TestServingNodeAddressSgsnNumberRoundTrip(t *testing.T) {
	in := &ServingNodeAddress{
		SgsnNumber:       "31622222222",
		SgsnNumberNature: 0x10,
		SgsnNumberPlan:   0x01,
	}
	wire, err := convertServingNodeAddressToWire(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if wire.Choice != gsm_map.ServingNodeAddressChoiceSgsnNumber {
		t.Errorf("Choice: want SgsnNumber, got %d", wire.Choice)
	}
	out, err := convertWireToServingNodeAddress(wire)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("round-trip mismatch:\n in=%+v\nout=%+v", in, out)
	}
}

func TestServingNodeAddressMmeNumberRoundTrip(t *testing.T) {
	in := &ServingNodeAddress{
		MmeNumber: HexBytes("mme1.example.com"), // 16 octets — within 9..255
	}
	wire, err := convertServingNodeAddressToWire(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if wire.Choice != gsm_map.ServingNodeAddressChoiceMmeNumber {
		t.Errorf("Choice: want MmeNumber, got %d", wire.Choice)
	}
	out, err := convertWireToServingNodeAddress(wire)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("round-trip mismatch:\n in=%+v\nout=%+v", in, out)
	}
}

func TestServingNodeAddressNoAlternativeRejected(t *testing.T) {
	_, err := convertServingNodeAddressToWire(&ServingNodeAddress{})
	if !errors.Is(err, ErrServingNodeAddressNoAlt) {
		t.Errorf("encode empty: want ErrServingNodeAddressNoAlt, got %v", err)
	}
}

func TestServingNodeAddressMultipleAlternativesRejected(t *testing.T) {
	_, err := convertServingNodeAddressToWire(&ServingNodeAddress{
		MscNumber:  "31611111111",
		SgsnNumber: "31622222222",
	})
	if !errors.Is(err, ErrServingNodeAddressMultipleAlts) {
		t.Errorf("encode 2 alts: want ErrServingNodeAddressMultipleAlts, got %v", err)
	}
}

func TestServingNodeAddressMmeNumberSizeValidation(t *testing.T) {
	short := &ServingNodeAddress{MmeNumber: HexBytes("short")} // 5 octets — under min 9
	_, err := convertServingNodeAddressToWire(short)
	if !errors.Is(err, ErrServingNodeAddressMmeNumberSize) {
		t.Errorf("encode 5 octets: want ErrServingNodeAddressMmeNumberSize, got %v", err)
	}
}

// Decoder must reject present-but-empty wire MscNumber/SgsnNumber for
// round-trip fidelity, parallel to the PSL-Arg DecodedEmpty pattern.
func TestServingNodeAddressMscNumberDecodedEmptyRejected(t *testing.T) {
	emptyAddr := gsm_map.ISDNAddressString{0x91} // header-only AddressString
	w := &gsm_map.ServingNodeAddress{
		Choice:    gsm_map.ServingNodeAddressChoiceMscNumber,
		MscNumber: &emptyAddr,
	}
	_, err := convertWireToServingNodeAddress(w)
	if !errors.Is(err, ErrServingNodeAddressMscNumberDecodedEmpty) {
		t.Errorf("MscNumber empty digits: want ErrServingNodeAddressMscNumberDecodedEmpty, got %v", err)
	}
}

func TestServingNodeAddressSgsnNumberDecodedEmptyRejected(t *testing.T) {
	emptyAddr := gsm_map.ISDNAddressString{0x91}
	w := &gsm_map.ServingNodeAddress{
		Choice:     gsm_map.ServingNodeAddressChoiceSgsnNumber,
		SgsnNumber: &emptyAddr,
	}
	_, err := convertWireToServingNodeAddress(w)
	if !errors.Is(err, ErrServingNodeAddressSgsnNumberDecodedEmpty) {
		t.Errorf("SgsnNumber empty digits: want ErrServingNodeAddressSgsnNumberDecodedEmpty, got %v", err)
	}
}

// Decoder must reject malformed CellIdOrSai CHOICEs (selected
// alternative but nil payload, or unknown choice value) instead of
// silently coercing to "absent". Caught by 3 reviewers (CodeRabbit,
// Codex, cubic) on PR #47.
func TestProvideSubscriberLocationResCellIdOrSaiInvalidChoice(t *testing.T) {
	t.Run("CGI choice but nil payload", func(t *testing.T) {
		w := &gsm_map.ProvideSubscriberLocationRes{
			LocationEstimate: gsm_map.ExtGeographicalInformation{0x10, 0x20, 0x30, 0x40},
			CellIdOrSai: &gsm_map.CellGlobalIdOrServiceAreaIdOrLAI{
				Choice: gsm_map.CellGlobalIdOrServiceAreaIdOrLAIChoiceCellGlobalIdOrServiceAreaIdFixedLength,
			},
		}
		_, err := convertWireToProvideSubscriberLocationRes(w)
		if !errors.Is(err, ErrPSLResCellIdOrSaiInvalidChoice) {
			t.Errorf("want ErrPSLResCellIdOrSaiInvalidChoice, got %v", err)
		}
	})
	t.Run("LAI choice but nil payload", func(t *testing.T) {
		w := &gsm_map.ProvideSubscriberLocationRes{
			LocationEstimate: gsm_map.ExtGeographicalInformation{0x10, 0x20, 0x30, 0x40},
			CellIdOrSai: &gsm_map.CellGlobalIdOrServiceAreaIdOrLAI{
				Choice: gsm_map.CellGlobalIdOrServiceAreaIdOrLAIChoiceLaiFixedLength,
			},
		}
		_, err := convertWireToProvideSubscriberLocationRes(w)
		if !errors.Is(err, ErrPSLResCellIdOrSaiInvalidChoice) {
			t.Errorf("want ErrPSLResCellIdOrSaiInvalidChoice, got %v", err)
		}
	})
	t.Run("unknown choice value", func(t *testing.T) {
		w := &gsm_map.ProvideSubscriberLocationRes{
			LocationEstimate: gsm_map.ExtGeographicalInformation{0x10, 0x20, 0x30, 0x40},
			CellIdOrSai:      &gsm_map.CellGlobalIdOrServiceAreaIdOrLAI{Choice: 99},
		}
		_, err := convertWireToProvideSubscriberLocationRes(w)
		if !errors.Is(err, ErrPSLResCellIdOrSaiInvalidChoice) {
			t.Errorf("want ErrPSLResCellIdOrSaiInvalidChoice, got %v", err)
		}
	})
}

func TestServingNodeAddressNilPassesThrough(t *testing.T) {
	wire, err := convertServingNodeAddressToWire(nil)
	if err != nil || wire != nil {
		t.Errorf("encode nil: want (nil,nil), got (%v,%v)", wire, err)
	}
	out, err := convertWireToServingNodeAddress(nil)
	if err != nil || out != nil {
		t.Errorf("decode nil: want (nil,nil), got (%v,%v)", out, err)
	}
}

// =============================================================================
// PSL-Res top-level
// =============================================================================

func TestProvideSubscriberLocationResMinimalRoundTrip(t *testing.T) {
	// Minimal: only the mandatory LocationEstimate.
	in := &ProvideSubscriberLocationRes{
		LocationEstimate: ExtGeographicalInformation{0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70, 0x80},
	}
	wire, err := convertProvideSubscriberLocationResToWire(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	out, err := convertWireToProvideSubscriberLocationRes(wire)
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
	parsed, err := ParseProvideSubscriberLocationRes(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !reflect.DeepEqual(in, parsed) {
		t.Errorf("Marshal/Parse mismatch:\n in=%+v\nout=%+v", in, parsed)
	}
}

func TestProvideSubscriberLocationResFullPopulationRoundTrip(t *testing.T) {
	age := int64(5)
	acc := AccuracyFulfilmentRequestedAccuracyFulfilled
	baro := UtranBaroPressureMeas(101325)
	in := &ProvideSubscriberLocationRes{
		LocationEstimate:               ExtGeographicalInformation{0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70, 0x80},
		AgeOfLocationEstimate:          &age,
		AddLocationEstimate:            AddGeographicalInformation{0x01, 0x02, 0x03, 0x04, 0x05},
		DeferredmtLrResponseIndicator:  true,
		GeranPositioningData:           PositioningDataInformation{0x01, 0x02, 0x03, 0x04},
		UtranPositioningData:           UtranPositioningDataInfo{0x01, 0x02, 0x03, 0x04, 0x05},
		CellGlobalId:                   HexBytes{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
		SaiPresent:                     true,
		AccuracyFulfilmentIndicator:    &acc,
		VelocityEstimate:               VelocityEstimate{0x01, 0x02, 0x03, 0x04},
		MoLrShortCircuitIndicator:      true,
		GeranGANSSpositioningData:      GeranGANSSpositioningData{0x01, 0x02, 0x03},
		UtranGANSSpositioningData:      UtranGANSSpositioningData{0x01, 0x02, 0x03},
		TargetServingNodeForHandover:   &ServingNodeAddress{MmeNumber: HexBytes("mme.example.com")},
		UtranAdditionalPositioningData: UtranAdditionalPositioningData{0x01, 0x02},
		UtranBaroPressureMeas:          &baro,
		UtranCivicAddress:              UtranCivicAddress("123 Main St"),
	}
	wire, err := convertProvideSubscriberLocationResToWire(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	out, err := convertWireToProvideSubscriberLocationRes(wire)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("round-trip mismatch:\n in=%+v\nout=%+v", in, out)
	}

	// BER round-trip.
	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	parsed, err := ParseProvideSubscriberLocationRes(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !reflect.DeepEqual(in, parsed) {
		t.Errorf("Marshal/Parse mismatch:\n in=%+v\nout=%+v", in, parsed)
	}
}

func TestProvideSubscriberLocationResLAIRoundTrip(t *testing.T) {
	// Use the LAI alternative of the CellIdOrSai CHOICE.
	in := &ProvideSubscriberLocationRes{
		LocationEstimate: ExtGeographicalInformation{0x10, 0x20, 0x30, 0x40},
		LAI:              HexBytes{0x32, 0xf4, 0x10, 0x12, 0x34}, // 5 octets
	}
	wire, err := convertProvideSubscriberLocationResToWire(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	out, err := convertWireToProvideSubscriberLocationRes(wire)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("LAI round-trip mismatch:\n in=%+v\nout=%+v", in, out)
	}
}

// =============================================================================
// Validation: mandatory + CHOICE constraints
// =============================================================================

func TestProvideSubscriberLocationResNilRejected(t *testing.T) {
	_, err := convertProvideSubscriberLocationResToWire(nil)
	if !errors.Is(err, ErrPSLResNil) {
		t.Errorf("encode nil: want ErrPSLResNil, got %v", err)
	}
	_, err = convertWireToProvideSubscriberLocationRes(nil)
	if !errors.Is(err, ErrPSLResNil) {
		t.Errorf("decode nil: want ErrPSLResNil, got %v", err)
	}
}

func TestProvideSubscriberLocationResMissingLocationEstimateRejected(t *testing.T) {
	_, err := convertProvideSubscriberLocationResToWire(&ProvideSubscriberLocationRes{})
	if !errors.Is(err, ErrPSLResLocationEstimateMissing) {
		t.Errorf("encode empty LocationEstimate: want ErrPSLResLocationEstimateMissing, got %v", err)
	}
}

func TestProvideSubscriberLocationResCellIdOrSaiMutex(t *testing.T) {
	in := &ProvideSubscriberLocationRes{
		LocationEstimate: ExtGeographicalInformation{0x10, 0x20, 0x30, 0x40},
		CellGlobalId:     HexBytes{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
		LAI:              HexBytes{0x32, 0xf4, 0x10, 0x12, 0x34},
	}
	_, err := convertProvideSubscriberLocationResToWire(in)
	if !errors.Is(err, ErrPSLResCellGlobalIdAndLAIMutex) {
		t.Errorf("encode both CGI+LAI: want ErrPSLResCellGlobalIdAndLAIMutex, got %v", err)
	}
}

func TestProvideSubscriberLocationResCellGlobalIdSizeValidation(t *testing.T) {
	in := &ProvideSubscriberLocationRes{
		LocationEstimate: ExtGeographicalInformation{0x10, 0x20, 0x30, 0x40},
		CellGlobalId:     HexBytes{0x01, 0x02, 0x03}, // 3 octets — must be 7
	}
	_, err := convertProvideSubscriberLocationResToWire(in)
	if !errors.Is(err, ErrPSLResCellGlobalIdSize) {
		t.Errorf("encode CGI=3: want ErrPSLResCellGlobalIdSize, got %v", err)
	}
}

func TestProvideSubscriberLocationResLAISizeValidation(t *testing.T) {
	in := &ProvideSubscriberLocationRes{
		LocationEstimate: ExtGeographicalInformation{0x10, 0x20, 0x30, 0x40},
		LAI:              HexBytes{0x01, 0x02, 0x03}, // 3 octets — must be 5
	}
	_, err := convertProvideSubscriberLocationResToWire(in)
	if !errors.Is(err, ErrPSLResLAIInvalidSize) {
		t.Errorf("encode LAI=3: want ErrPSLResLAIInvalidSize, got %v", err)
	}
}

func TestProvideSubscriberLocationResUtranBaroPressureRangeValidation(t *testing.T) {
	low := UtranBaroPressureMeas(29999) // under 30000 minimum
	in := &ProvideSubscriberLocationRes{
		LocationEstimate:      ExtGeographicalInformation{0x10, 0x20, 0x30, 0x40},
		UtranBaroPressureMeas: &low,
	}
	_, err := convertProvideSubscriberLocationResToWire(in)
	if !errors.Is(err, ErrUtranBaroPressureMeasOutOfRange) {
		t.Errorf("encode baro=29999: want ErrUtranBaroPressureMeasOutOfRange, got %v", err)
	}
}

func TestProvideSubscriberLocationResAccuracyFulfilmentEncoderStrict(t *testing.T) {
	bad := AccuracyFulfilmentIndicator(99)
	in := &ProvideSubscriberLocationRes{
		LocationEstimate:            ExtGeographicalInformation{0x10, 0x20, 0x30, 0x40},
		AccuracyFulfilmentIndicator: &bad,
	}
	_, err := convertProvideSubscriberLocationResToWire(in)
	if !errors.Is(err, ErrAccuracyFulfilmentIndicatorInvalid) {
		t.Errorf("encode acc=99: want ErrAccuracyFulfilmentIndicatorInvalid, got %v", err)
	}
}

// AccuracyFulfilmentIndicator is extensible — decoder must preserve
// unknown values per Postel.
func TestProvideSubscriberLocationResAccuracyFulfilmentDecoderLenient(t *testing.T) {
	bad := gsm_map.AccuracyFulfilmentIndicator(99)
	w := &gsm_map.ProvideSubscriberLocationRes{
		LocationEstimate:            gsm_map.ExtGeographicalInformation{0x10, 0x20, 0x30, 0x40},
		AccuracyFulfilmentIndicator: &bad,
	}
	out, err := convertWireToProvideSubscriberLocationRes(w)
	if err != nil {
		t.Fatalf("decode acc=99: unexpected error %v", err)
	}
	if out.AccuracyFulfilmentIndicator == nil || int64(*out.AccuracyFulfilmentIndicator) != 99 {
		t.Errorf("decoder leniency: want acc=99 preserved, got %v", out.AccuracyFulfilmentIndicator)
	}
}

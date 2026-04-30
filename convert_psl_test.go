// convert_psl_test.go
//
// Tests for the PSL leaf SEQUENCE converters and BIT STRING surrogate
// codecs. Round-trip + targeted negative cases.
package gsmmap

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gomaja/go-asn1/runtime"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// ============================================================================
// DeferredLocationEventType BIT STRING
// ============================================================================

func TestDeferredLocationEventTypeRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   *DeferredLocationEventType
		bits int
	}{
		{"none", &DeferredLocationEventType{}, 1},
		{"first only", &DeferredLocationEventType{MsAvailable: true}, 1},
		{"second only", &DeferredLocationEventType{EnteringIntoArea: true}, 2},
		{"third only", &DeferredLocationEventType{LeavingFromArea: true}, 3},
		{"fifth only", &DeferredLocationEventType{PeriodicLDR: true}, 5},
		{"all set", &DeferredLocationEventType{
			MsAvailable: true, EnteringIntoArea: true, LeavingFromArea: true,
			BeingInsideArea: true, PeriodicLDR: true,
		}, 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bs := convertDeferredLocationEventTypeToBitString(tc.in)
			if bs.BitLength != tc.bits {
				t.Errorf("BitLength: want %d, got %d", tc.bits, bs.BitLength)
			}
			out, err := convertBitStringToDeferredLocationEventType(bs)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !reflect.DeepEqual(tc.in, out) {
				t.Errorf("round-trip mismatch: in=%+v out=%+v", tc.in, out)
			}
		})
	}
}

func TestDeferredLocationEventTypeOversizedRejected(t *testing.T) {
	bs := runtime.BitString{Bytes: []byte{0x00, 0x00, 0x00}, BitLength: 17}
	_, err := convertBitStringToDeferredLocationEventType(bs)
	if !errors.Is(err, ErrDeferredLocationEventTypeSize) {
		t.Errorf("want ErrDeferredLocationEventTypeSize, got %v", err)
	}
}

func TestDeferredLocationEventTypeZeroBitsRejected(t *testing.T) {
	bs := runtime.BitString{Bytes: []byte{}, BitLength: 0}
	_, err := convertBitStringToDeferredLocationEventType(bs)
	if !errors.Is(err, ErrDeferredLocationEventTypeSize) {
		t.Errorf("want ErrDeferredLocationEventTypeSize, got %v", err)
	}
}

// ============================================================================
// SupportedGADShapes BIT STRING
// ============================================================================

func TestSupportedGADShapesRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   *SupportedGADShapes
	}{
		{"none", &SupportedGADShapes{}},
		{"point only", &SupportedGADShapes{EllipsoidPoint: true}},
		{"arc only", &SupportedGADShapes{EllipsoidArc: true}},
		{"all set", &SupportedGADShapes{
			EllipsoidPoint: true, EllipsoidPointWithUncertaintyCircle: true,
			EllipsoidPointWithUncertaintyEllipse: true, Polygon: true,
			EllipsoidPointWithAltitude:                        true,
			EllipsoidPointWithAltitudeAndUncertaintyEllipsoid: true,
			EllipsoidArc: true,
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bs := convertSupportedGADShapesToBitString(tc.in)
			if bs.BitLength != 7 {
				t.Errorf("BitLength: want 7 (spec minimum), got %d", bs.BitLength)
			}
			out, err := convertBitStringToSupportedGADShapes(bs)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !reflect.DeepEqual(tc.in, out) {
				t.Errorf("round-trip mismatch: in=%+v out=%+v", tc.in, out)
			}
		})
	}
}

func TestSupportedGADShapesUndersizedRejected(t *testing.T) {
	bs := runtime.BitString{Bytes: []byte{0x80}, BitLength: 6}
	_, err := convertBitStringToSupportedGADShapes(bs)
	if !errors.Is(err, ErrSupportedGADShapesSize) {
		t.Errorf("want ErrSupportedGADShapesSize, got %v", err)
	}
}

func TestSupportedGADShapesOversizedRejected(t *testing.T) {
	bs := runtime.BitString{Bytes: []byte{0xff, 0xff, 0xff}, BitLength: 17}
	_, err := convertBitStringToSupportedGADShapes(bs)
	if !errors.Is(err, ErrSupportedGADShapesSize) {
		t.Errorf("want ErrSupportedGADShapesSize, got %v", err)
	}
}

// ============================================================================
// LocationType
// ============================================================================

func TestLocationTypeRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   *LocationType
	}{
		{"current location only", &LocationType{LocationEstimateType: LocationEstimateCurrentLocation}},
		{"with deferred event", &LocationType{
			LocationEstimateType: LocationEstimateActivateDeferredLocation,
			DeferredLocationEventType: &DeferredLocationEventType{
				MsAvailable: true, PeriodicLDR: true,
			},
		}},
		{"notification verification only", &LocationType{LocationEstimateType: LocationEstimateNotificationVerificationOnly}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertLocationTypeToWire(tc.in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			out, err := convertWireToLocationType(wire)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !reflect.DeepEqual(tc.in, out) {
				t.Errorf("round-trip mismatch: in=%+v out=%+v", tc.in, out)
			}
		})
	}
}

func TestLocationTypeNilPassesThrough(t *testing.T) {
	wire, err := convertLocationTypeToWire(nil)
	if err != nil || wire != nil {
		t.Errorf("nil → nil expected, got wire=%v err=%v", wire, err)
	}
	out, err := convertWireToLocationType(nil)
	if err != nil || out != nil {
		t.Errorf("nil → nil expected, got out=%v err=%v", out, err)
	}
}

func TestLocationTypeOutOfRangeEnumRejected(t *testing.T) {
	_, err := convertLocationTypeToWire(&LocationType{LocationEstimateType: 99})
	if !errors.Is(err, ErrLocationEstimateTypeInvalid) {
		t.Errorf("want ErrLocationEstimateTypeInvalid, got %v", err)
	}
}

// ============================================================================
// LCSCodeword
// ============================================================================

func TestLCSCodewordRoundTrip(t *testing.T) {
	in := &LCSCodeword{
		DataCodingScheme:  0x0f,
		LcsCodewordString: HexBytes{0x01, 0x02, 0x03, 0x04, 0x05},
	}
	wire, err := convertLCSCodewordToWire(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	out, err := convertWireToLCSCodeword(wire)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("round-trip mismatch: in=%+v out=%+v", in, out)
	}
}

func TestLCSCodewordEmptyStringRejected(t *testing.T) {
	_, err := convertLCSCodewordToWire(&LCSCodeword{
		DataCodingScheme:  0x0f,
		LcsCodewordString: HexBytes{},
	})
	if !errors.Is(err, ErrLCSCodewordStringSize) {
		t.Errorf("want ErrLCSCodewordStringSize, got %v", err)
	}
}

func TestLCSCodewordOversizedStringRejected(t *testing.T) {
	tooBig := make(HexBytes, LCSCodewordStringMaxLen+1)
	_, err := convertLCSCodewordToWire(&LCSCodeword{
		DataCodingScheme:  0x0f,
		LcsCodewordString: tooBig,
	})
	if !errors.Is(err, ErrLCSCodewordStringSize) {
		t.Errorf("want ErrLCSCodewordStringSize, got %v", err)
	}
}

func TestLCSCodewordWireDataCodingSchemeMustBeOneOctet(t *testing.T) {
	w := &gsm_map.LCSCodeword{
		DataCodingScheme:  gsm_map.USSDDataCodingScheme{0x0f, 0x10}, // too long
		LcsCodewordString: gsm_map.LCSCodewordString{0x01},
	}
	_, err := convertWireToLCSCodeword(w)
	if err == nil {
		t.Error("want error for >1 octet DataCodingScheme")
	}
}

// ============================================================================
// LCSPrivacyCheck
// ============================================================================

func TestLCSPrivacyCheckRoundTrip(t *testing.T) {
	related := PrivacyCheckAllowedWithNotification
	cases := []struct {
		name string
		in   *LCSPrivacyCheck
	}{
		{"unrelated only", &LCSPrivacyCheck{
			CallSessionUnrelated: PrivacyCheckAllowedWithoutNotification,
		}},
		{"both set", &LCSPrivacyCheck{
			CallSessionUnrelated: PrivacyCheckRestrictedIfNoResponse,
			CallSessionRelated:   &related,
		}},
		{"max enum value", &LCSPrivacyCheck{
			CallSessionUnrelated: PrivacyCheckNotAllowed,
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertLCSPrivacyCheckToWire(tc.in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			out, err := convertWireToLCSPrivacyCheck(wire)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !reflect.DeepEqual(tc.in, out) {
				t.Errorf("round-trip mismatch: in=%+v out=%+v", tc.in, out)
			}
		})
	}
}

func TestLCSPrivacyCheckOutOfRangeRejected(t *testing.T) {
	_, err := convertLCSPrivacyCheckToWire(&LCSPrivacyCheck{
		CallSessionUnrelated: 99,
	})
	if !errors.Is(err, ErrPrivacyCheckRelatedActionInvalid) {
		t.Errorf("want ErrPrivacyCheckRelatedActionInvalid, got %v", err)
	}

	related := PrivacyCheckRelatedAction(7)
	_, err = convertLCSPrivacyCheckToWire(&LCSPrivacyCheck{
		CallSessionUnrelated: PrivacyCheckAllowedWithoutNotification,
		CallSessionRelated:   &related,
	})
	if !errors.Is(err, ErrPrivacyCheckRelatedActionInvalid) {
		t.Errorf("want ErrPrivacyCheckRelatedActionInvalid for related, got %v", err)
	}
}

// ============================================================================
// ResponseTime
// ============================================================================

func TestResponseTimeRoundTrip(t *testing.T) {
	cases := []ResponseTimeCategory{ResponseTimeLowdelay, ResponseTimeDelaytolerant}
	for _, cat := range cases {
		in := &ResponseTime{ResponseTimeCategory: cat}
		wire, err := convertResponseTimeToWire(in)
		if err != nil {
			t.Fatalf("encode %v: %v", cat, err)
		}
		out, err := convertWireToResponseTime(wire)
		if err != nil {
			t.Fatalf("decode %v: %v", cat, err)
		}
		if !reflect.DeepEqual(in, out) {
			t.Errorf("round-trip mismatch for %v: in=%+v out=%+v", cat, in, out)
		}
	}
}

func TestResponseTimeEncoderRejectsUnknownValue(t *testing.T) {
	_, err := convertResponseTimeToWire(&ResponseTime{ResponseTimeCategory: 5})
	if !errors.Is(err, ErrResponseTimeCategoryInvalid) {
		t.Errorf("want ErrResponseTimeCategoryInvalid, got %v", err)
	}
}

func TestResponseTimeDecoderAppliesSpecExceptionClause(t *testing.T) {
	// Per TS 29.002 MAP-LCS-DataTypes.asn:270-271, an unrecognized value
	// shall be treated the same as delaytolerant(1) on decode.
	w := &gsm_map.ResponseTime{ResponseTimeCategory: 5}
	out, err := convertWireToResponseTime(w)
	if err != nil {
		t.Fatalf("decode unexpected error: %v", err)
	}
	if out.ResponseTimeCategory != ResponseTimeDelaytolerant {
		t.Errorf("spec exception: want delaytolerant(1), got %d", out.ResponseTimeCategory)
	}
}

// ============================================================================
// LCSQoS
// ============================================================================

func TestLCSQoSRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   *LCSQoS
	}{
		{"empty", &LCSQoS{}},
		{"horizontal accuracy only", &LCSQoS{HorizontalAccuracy: HexBytes{0x42}}},
		{"vertical coord request only", &LCSQoS{VerticalCoordinateRequest: true}},
		{"velocity request only", &LCSQoS{VelocityRequest: true}},
		{"with response time", &LCSQoS{
			HorizontalAccuracy:        HexBytes{0x10},
			VerticalCoordinateRequest: true,
			VerticalAccuracy:          HexBytes{0x20},
			ResponseTime:              &ResponseTime{ResponseTimeCategory: ResponseTimeDelaytolerant},
			VelocityRequest:           true,
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertLCSQoSToWire(tc.in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			out, err := convertWireToLCSQoS(wire)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !reflect.DeepEqual(tc.in, out) {
				t.Errorf("round-trip mismatch: in=%+v out=%+v", tc.in, out)
			}
		})
	}
}

func TestLCSQoSHorizontalAccuracyMustBeOneOctet(t *testing.T) {
	_, err := convertLCSQoSToWire(&LCSQoS{HorizontalAccuracy: HexBytes{0x01, 0x02}})
	if !errors.Is(err, ErrHorizontalAccuracyInvalidSize) {
		t.Errorf("want ErrHorizontalAccuracyInvalidSize on encode, got %v", err)
	}
	w := &gsm_map.LCSQoS{HorizontalAccuracy: &gsm_map.HorizontalAccuracy{0x01, 0x02}}
	_, err = convertWireToLCSQoS(w)
	if !errors.Is(err, ErrHorizontalAccuracyInvalidSize) {
		t.Errorf("want ErrHorizontalAccuracyInvalidSize on decode, got %v", err)
	}
}

func TestLCSQoSVerticalAccuracyMustBeOneOctet(t *testing.T) {
	_, err := convertLCSQoSToWire(&LCSQoS{VerticalAccuracy: HexBytes{0x01, 0x02}})
	if !errors.Is(err, ErrVerticalAccuracyInvalidSize) {
		t.Errorf("want ErrVerticalAccuracyInvalidSize on encode, got %v", err)
	}
	w := &gsm_map.LCSQoS{VerticalAccuracy: &gsm_map.VerticalAccuracy{0x01, 0x02}}
	_, err = convertWireToLCSQoS(w)
	if !errors.Is(err, ErrVerticalAccuracyInvalidSize) {
		t.Errorf("want ErrVerticalAccuracyInvalidSize on decode, got %v", err)
	}
}

func TestLCSQoSNilPassesThrough(t *testing.T) {
	wire, err := convertLCSQoSToWire(nil)
	if err != nil || wire != nil {
		t.Errorf("nil → nil expected, got wire=%v err=%v", wire, err)
	}
	out, err := convertWireToLCSQoS(nil)
	if err != nil || out != nil {
		t.Errorf("nil → nil expected, got out=%v err=%v", out, err)
	}
}

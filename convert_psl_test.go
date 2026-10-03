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
			out := convertBitStringToDeferredLocationEventType(bs)
			if !reflect.DeepEqual(tc.in, out) {
				t.Errorf("round-trip mismatch: in=%+v out=%+v", tc.in, out)
			}
		})
	}
}

// Decode-only tests from hardcoded wire bytes lock the bit positions
// independently of the encoder, catching symmetric encode/decode bit-
// order bugs that would slip through round-trip-only assertions.
func TestDeferredLocationEventTypeBitMappingFromWire(t *testing.T) {
	cases := []struct {
		name string
		bs   runtime.BitString
		want *DeferredLocationEventType
	}{
		{"bit 0 → MsAvailable", runtime.BitString{Bytes: []byte{0x80}, BitLength: 1}, &DeferredLocationEventType{MsAvailable: true}},
		{"bit 1 → EnteringIntoArea", runtime.BitString{Bytes: []byte{0x40}, BitLength: 2}, &DeferredLocationEventType{EnteringIntoArea: true}},
		{"bit 2 → LeavingFromArea", runtime.BitString{Bytes: []byte{0x20}, BitLength: 3}, &DeferredLocationEventType{LeavingFromArea: true}},
		{"bit 3 → BeingInsideArea", runtime.BitString{Bytes: []byte{0x10}, BitLength: 4}, &DeferredLocationEventType{BeingInsideArea: true}},
		{"bit 4 → PeriodicLDR", runtime.BitString{Bytes: []byte{0x08}, BitLength: 5}, &DeferredLocationEventType{PeriodicLDR: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := convertBitStringToDeferredLocationEventType(tc.bs)
			if !reflect.DeepEqual(tc.want, got) {
				t.Errorf("bit mapping: want %+v, got %+v", tc.want, got)
			}
		})
	}
}

func TestDeferredLocationEventTypeOversizedRejected(t *testing.T) {
	// The public flags only produce valid bit lengths; mutate the converted
	// wire value to exercise the BER bound.
	bs := runtime.BitString{Bytes: []byte{0x00, 0x00, 0x00}, BitLength: 17}
	wire, err := convertLocationTypeToWire(&LocationType{LocationEstimateType: LocationEstimateCurrentLocation})
	if err != nil {
		t.Fatal(err)
	}
	wire.DeferredLocationEventType = &bs
	_, err = strictWire(wire, nil)
	wantConstraintError(t, err, "deferredLocationEventType", "SIZE (1..16)")
}

func TestDeferredLocationEventTypeZeroBitsRejected(t *testing.T) {
	bs := runtime.BitString{Bytes: []byte{}, BitLength: 0}
	wire, err := convertLocationTypeToWire(&LocationType{LocationEstimateType: LocationEstimateCurrentLocation})
	if err != nil {
		t.Fatal(err)
	}
	wire.DeferredLocationEventType = &bs
	_, err = strictWire(wire, nil)
	wantConstraintError(t, err, "deferredLocationEventType", "SIZE (1..16)")
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
			out := convertBitStringToSupportedGADShapes(bs)
			if !reflect.DeepEqual(tc.in, out) {
				t.Errorf("round-trip mismatch: in=%+v out=%+v", tc.in, out)
			}
		})
	}
}

// Decode-only tests from hardcoded wire bytes lock the bit positions.
func TestSupportedGADShapesBitMappingFromWire(t *testing.T) {
	cases := []struct {
		name string
		bs   runtime.BitString
		want *SupportedGADShapes
	}{
		{"bit 0 → EllipsoidPoint", runtime.BitString{Bytes: []byte{0x80}, BitLength: 7}, &SupportedGADShapes{EllipsoidPoint: true}},
		{"bit 1 → EllipsoidPointWithUncertaintyCircle", runtime.BitString{Bytes: []byte{0x40}, BitLength: 7}, &SupportedGADShapes{EllipsoidPointWithUncertaintyCircle: true}},
		{"bit 2 → EllipsoidPointWithUncertaintyEllipse", runtime.BitString{Bytes: []byte{0x20}, BitLength: 7}, &SupportedGADShapes{EllipsoidPointWithUncertaintyEllipse: true}},
		{"bit 3 → Polygon", runtime.BitString{Bytes: []byte{0x10}, BitLength: 7}, &SupportedGADShapes{Polygon: true}},
		{"bit 4 → EllipsoidPointWithAltitude", runtime.BitString{Bytes: []byte{0x08}, BitLength: 7}, &SupportedGADShapes{EllipsoidPointWithAltitude: true}},
		{"bit 5 → EllipsoidPointWithAltitudeAndUncertaintyEllipsoid", runtime.BitString{Bytes: []byte{0x04}, BitLength: 7}, &SupportedGADShapes{EllipsoidPointWithAltitudeAndUncertaintyEllipsoid: true}},
		{"bit 6 → EllipsoidArc", runtime.BitString{Bytes: []byte{0x02}, BitLength: 7}, &SupportedGADShapes{EllipsoidArc: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := convertBitStringToSupportedGADShapes(tc.bs)
			if !reflect.DeepEqual(tc.want, got) {
				t.Errorf("bit mapping: want %+v, got %+v", tc.want, got)
			}
		})
	}
}

func TestSupportedGADShapesUndersizedRejected(t *testing.T) {
	bs := runtime.BitString{Bytes: []byte{0x80}, BitLength: 6}
	wire, err := convertProvideSubscriberLocationArgToWire(&ProvideSubscriberLocationArg{
		LocationType: LocationType{LocationEstimateType: LocationEstimateCurrentLocation}, MlcNumber: "1",
	})
	if err != nil {
		t.Fatal(err)
	}
	wire.SupportedGADShapes = &bs
	_, err = strictWire(wire, nil)
	wantConstraintError(t, err, "supportedGADShapes", "SIZE (7..16)")
}

func TestSupportedGADShapesOversizedRejected(t *testing.T) {
	bs := runtime.BitString{Bytes: []byte{0xff, 0xff, 0xff}, BitLength: 17}
	wire, err := convertProvideSubscriberLocationArgToWire(&ProvideSubscriberLocationArg{
		LocationType: LocationType{LocationEstimateType: LocationEstimateCurrentLocation}, MlcNumber: "1",
	})
	if err != nil {
		t.Fatal(err)
	}
	wire.SupportedGADShapes = &bs
	_, err = strictWire(wire, nil)
	wantConstraintError(t, err, "supportedGADShapes", "SIZE (7..16)")
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
			out := convertWireToLocationType(wire)
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
	out := convertWireToLocationType(nil)
	if out != nil {
		t.Errorf("nil → nil expected, got out=%v", out)
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
	wire := convertLCSCodewordToWire(in)
	out := convertWireToLCSCodeword(wire)
	if !reflect.DeepEqual(in, out) {
		t.Errorf("round-trip mismatch: in=%+v out=%+v", in, out)
	}
}

func TestLCSCodewordEmptyStringRejected(t *testing.T) {
	_, err := strictWire(convertLCSCodewordToWire(&LCSCodeword{
		DataCodingScheme:  0x0f,
		LcsCodewordString: HexBytes{},
	}), nil)
	if !matchesConstraint(err, "lcsCodewordString", "SIZE (1..20)") {
		t.Errorf("want BER constraint error, got %v", err)
	}
}

func TestLCSCodewordOversizedStringRejected(t *testing.T) {
	tooBig := make(HexBytes, 20+1)
	_, err := strictWire(convertLCSCodewordToWire(&LCSCodeword{
		DataCodingScheme:  0x0f,
		LcsCodewordString: tooBig,
	}), nil)
	if !matchesConstraint(err, "lcsCodewordString", "SIZE (1..20)") {
		t.Errorf("want BER constraint error, got %v", err)
	}
}

func TestLCSCodewordWireDataCodingSchemeMustBeOneOctet(t *testing.T) {
	w := &gsm_map.LCSCodeword{
		DataCodingScheme:  gsm_map.USSDDataCodingScheme{0x0f, 0x10}, // too long
		LcsCodewordString: gsm_map.LCSCodewordString{0x01},
	}
	err := strictDecodeWire(w)
	if !matchesConstraint(err, "dataCodingScheme", "SIZE (1)") {
		t.Errorf("want BER constraint error, got %v", err)
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
			out := convertWireToLCSPrivacyCheck(wire)
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
		out := convertWireToResponseTime(wire)
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
	// Per 3GPP TS 29.002 V19.1.0 §17.7.13, an unrecognized value
	// shall be treated the same as delaytolerant(1) on decode.
	w := &gsm_map.ResponseTime{ResponseTimeCategory: 5}
	out := convertWireToResponseTime(w)
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
	_, err := strictWire(convertLCSQoSToWire(&LCSQoS{HorizontalAccuracy: HexBytes{0x01, 0x02}}))
	if !matchesConstraint(err, "horizontal-accuracy", "SIZE (1)") {
		t.Errorf("want BER constraint error on encode, got %v", err)
	}
	w := &gsm_map.LCSQoS{HorizontalAccuracy: &gsm_map.HorizontalAccuracy{0x01, 0x02}}
	err = strictDecodeWire(w)
	if !matchesConstraint(err, "horizontal-accuracy", "SIZE (1)") {
		t.Errorf("want BER constraint error on decode, got %v", err)
	}
}

func TestLCSQoSVerticalAccuracyMustBeOneOctet(t *testing.T) {
	_, err := strictWire(convertLCSQoSToWire(&LCSQoS{VerticalAccuracy: HexBytes{0x01, 0x02}}))
	if !matchesConstraint(err, "vertical-accuracy", "SIZE (1)") {
		t.Errorf("want BER constraint error on encode, got %v", err)
	}
	w := &gsm_map.LCSQoS{VerticalAccuracy: &gsm_map.VerticalAccuracy{0x01, 0x02}}
	err = strictDecodeWire(w)
	if !matchesConstraint(err, "vertical-accuracy", "SIZE (1)") {
		t.Errorf("want BER constraint error on decode, got %v", err)
	}
}

// Spec mandates bit 8 = 0 on the Horizontal-Accuracy / Vertical-Accuracy
// uncertainty code octet (3GPP TS 29.002 V19.1.0 §17.7.13).
// MSB-set values must be rejected on both encode and decode.
func TestLCSQoSHorizontalAccuracyReservedBitRejected(t *testing.T) {
	_, err := convertLCSQoSToWire(&LCSQoS{HorizontalAccuracy: HexBytes{0x80}})
	if !errors.Is(err, ErrHorizontalAccuracyReservedBit) {
		t.Errorf("want ErrHorizontalAccuracyReservedBit on encode, got %v", err)
	}
	_, err = convertLCSQoSToWire(&LCSQoS{HorizontalAccuracy: HexBytes{0xff}})
	if !errors.Is(err, ErrHorizontalAccuracyReservedBit) {
		t.Errorf("want ErrHorizontalAccuracyReservedBit on encode (0xff), got %v", err)
	}
	w := &gsm_map.LCSQoS{HorizontalAccuracy: &gsm_map.HorizontalAccuracy{0x80}}
	_, err = convertWireToLCSQoS(w)
	if !errors.Is(err, ErrHorizontalAccuracyReservedBit) {
		t.Errorf("want ErrHorizontalAccuracyReservedBit on decode, got %v", err)
	}
}

func TestLCSQoSVerticalAccuracyReservedBitRejected(t *testing.T) {
	_, err := convertLCSQoSToWire(&LCSQoS{VerticalAccuracy: HexBytes{0x80}})
	if !errors.Is(err, ErrVerticalAccuracyReservedBit) {
		t.Errorf("want ErrVerticalAccuracyReservedBit on encode, got %v", err)
	}
	w := &gsm_map.LCSQoS{VerticalAccuracy: &gsm_map.VerticalAccuracy{0xc0}}
	_, err = convertWireToLCSQoS(w)
	if !errors.Is(err, ErrVerticalAccuracyReservedBit) {
		t.Errorf("want ErrVerticalAccuracyReservedBit on decode, got %v", err)
	}
}

// Empty (non-nil) HexBytes slices are treated as absent on encode, matching
// the package-wide convention for optional slice fields. The wire output
// must therefore omit both fields, not surface a 0-byte length error.
func TestLCSQoSEmptyAccuracyTreatedAsAbsent(t *testing.T) {
	wire, err := convertLCSQoSToWire(&LCSQoS{
		HorizontalAccuracy: HexBytes{},
		VerticalAccuracy:   HexBytes{},
	})
	if err != nil {
		t.Fatalf("encode empty HexBytes: %v", err)
	}
	if wire.HorizontalAccuracy != nil {
		t.Error("empty HexBytes should produce nil HorizontalAccuracy on the wire")
	}
	if wire.VerticalAccuracy != nil {
		t.Error("empty HexBytes should produce nil VerticalAccuracy on the wire")
	}
}

// Boundary-case acceptance: 0x7F is valid (bit 8 = 0, all 7 low bits set).
func TestLCSQoSAccuracyBoundaryAccepted(t *testing.T) {
	wire, err := convertLCSQoSToWire(&LCSQoS{
		HorizontalAccuracy: HexBytes{0x7f},
		VerticalAccuracy:   HexBytes{0x7f},
	})
	if err != nil {
		t.Fatalf("encode 0x7f: %v", err)
	}
	if _, err := convertWireToLCSQoS(wire); err != nil {
		t.Fatalf("decode 0x7f: %v", err)
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

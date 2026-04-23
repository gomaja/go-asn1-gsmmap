// geo_test.go
//
// Tests for geographical-information encoder validation per 3GPP TS 23.032.
// The Encode() path must reject caller errors (out-of-range lat/lon,
// out-of-range altitude, missing required fields) at the boundary rather
// than silently clamping or zero-filling.
package gsmmap

import (
	"bytes"
	"math"
	"testing"
)

// Encode must reject latitudes outside (-90, 90). The spec's encoding
// cannot represent exactly ±90 without silent quantization, so those
// boundary values are also rejected.
func TestGeoEncode_RejectsOutOfRangeLatitude(t *testing.T) {
	for _, lat := range []float64{-90.01, 90.01, 200, -500, 90, -90} {
		gi := &GeographicalInfo{ShapeType: ShapeEllipsoidPoint, Latitude: lat, Longitude: 0}
		if _, err := gi.Encode(); err == nil {
			t.Errorf("Latitude=%v: expected error, got nil", lat)
		}
	}
}

// Encode must reject longitudes outside [-180, 180). lon=-180 is
// representable exactly (two's-complement 0x800000) and is accepted;
// lon=+180 cannot be represented and is rejected.
func TestGeoEncode_RejectsOutOfRangeLongitude(t *testing.T) {
	for _, lon := range []float64{-180.01, 180.01, 500, -500, 180} {
		gi := &GeographicalInfo{ShapeType: ShapeEllipsoidPoint, Latitude: 0, Longitude: lon}
		if _, err := gi.Encode(); err == nil {
			t.Errorf("Longitude=%v: expected error, got nil", lon)
		}
	}
}

// Encode must accept valid lat/lon values inside the open/half-open
// ranges. lon=-180 is the only exactly-representable boundary value.
func TestGeoEncode_AcceptsBoundaryLatLon(t *testing.T) {
	cases := []struct {
		lat, lon float64
	}{
		{0, 0},
		{89.9999, 179.9999},
		{-89.9999, -180},
	}
	for _, tc := range cases {
		gi := &GeographicalInfo{ShapeType: ShapeEllipsoidPoint, Latitude: tc.lat, Longitude: tc.lon}
		if _, err := gi.Encode(); err != nil {
			t.Errorf("Encode(lat=%v, lon=%v): unexpected error: %v", tc.lat, tc.lon, err)
		}
	}
}

// Encode must reject values that pass the caller-facing range check but
// still round up to the next quantum in encodeLatLon. The ULP just below
// 90 and +180 lands on 0x800000 after math.Round, so emitting it would
// silently quantize to 0x7FFFFF.
func TestGeoEncode_RejectsULPJustBelowBoundary(t *testing.T) {
	// The largest float64 strictly less than 90 is 90 - 2^-46 ≈
	// 89.99999999999999, which passes the < 90 guard and then rounds up
	// to 0x800000 in encodeLatLon.
	latULP := math.Nextafter(90, 0)
	if latULP >= 90 {
		t.Fatalf("math.Nextafter(90,0) was not strictly below 90: %v", latULP)
	}
	gi := &GeographicalInfo{ShapeType: ShapeEllipsoidPoint, Latitude: latULP, Longitude: 0}
	if _, err := gi.Encode(); err == nil {
		t.Errorf("Latitude=%v (ULP below 90): expected error, got nil", latULP)
	}

	// Same for latitude near the negative boundary.
	latNegULP := math.Nextafter(-90, 0)
	if latNegULP <= -90 {
		t.Fatalf("math.Nextafter(-90,0) was not strictly above -90: %v", latNegULP)
	}
	gi = &GeographicalInfo{ShapeType: ShapeEllipsoidPoint, Latitude: latNegULP, Longitude: 0}
	if _, err := gi.Encode(); err == nil {
		t.Errorf("Latitude=%v (ULP above -90): expected error, got nil", latNegULP)
	}

	// Same for longitude near +180.
	lonULP := math.Nextafter(180, 0)
	if lonULP >= 180 {
		t.Fatalf("math.Nextafter(180,0) was not strictly below 180: %v", lonULP)
	}
	gi = &GeographicalInfo{ShapeType: ShapeEllipsoidPoint, Latitude: 0, Longitude: lonULP}
	if _, err := gi.Encode(); err == nil {
		t.Errorf("Longitude=%v (ULP below 180): expected error, got nil", lonULP)
	}

	// Longitude just above -180: math.Round lands on -0x800000, which is
	// the legitimate quantum for lon=-180 exactly. Accepting this would
	// silently collapse the caller's input onto -180 on the wire.
	lonNegULP := math.Nextafter(-180, 0)
	if lonNegULP <= -180 {
		t.Fatalf("math.Nextafter(-180,0) was not strictly above -180: %v", lonNegULP)
	}
	gi = &GeographicalInfo{ShapeType: ShapeEllipsoidPoint, Latitude: 0, Longitude: lonNegULP}
	if _, err := gi.Encode(); err == nil {
		t.Errorf("Longitude=%v (ULP above -180): expected error, got nil", lonNegULP)
	}
}

// lon=-180 must round-trip exactly — the two's-complement encoding
// makes 0x800000 an exact representation of -180°.
func TestGeoEncode_LonNegative180RoundTrips(t *testing.T) {
	gi := &GeographicalInfo{ShapeType: ShapeEllipsoidPoint, Latitude: 0, Longitude: -180}
	data, err := gi.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got, err := DecodeGeographicalInfo(data)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got.Longitude != -180 {
		t.Errorf("Longitude round-trip: got %v, want -180", got.Longitude)
	}
}

// Encode must reject altitudes outside the encodable range [-32767, 32767].
// Previously only the lower bound was checked, and the int16 sign-flip for
// MinInt16 (-32768) silently produced a zero-magnitude value.
func TestGeoEncode_AltitudeRangeCheck(t *testing.T) {
	cases := []struct {
		name    string
		alt     int16
		wantErr bool
	}{
		{"zero", 0, false},
		{"max valid", 32767, false},
		{"min valid", -32767, false},
		{"MinInt16 rejected", math.MinInt16, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			alt := tc.alt
			gi := &GeographicalInfo{
				ShapeType: ShapeEllipsoidPointAltitude,
				Latitude:  0, Longitude: 0,
				Altitude:             &alt,
				UncertaintySemiMajor: u8(1),
				UncertaintySemiMinor: u8(1),
				AngleMajorAxis:       u8(0),
				UncertaintyAltitude:  u8(1),
				Confidence:           u8(50),
			}
			_, err := gi.Encode()
			if tc.wantErr && err == nil {
				t.Errorf("alt=%d: expected error, got nil", tc.alt)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("alt=%d: unexpected error: %v", tc.alt, err)
			}
		})
	}
}

// Round-trip altitude 32767 and -32767 must survive encode+decode with
// the original value intact.
func TestGeoEncode_AltitudeBoundaryRoundTrip(t *testing.T) {
	for _, alt := range []int16{32767, -32767, 0, 1, -1, 100, -100} {
		a := alt
		gi := &GeographicalInfo{
			ShapeType: ShapeEllipsoidPointAltitude,
			Latitude:  0, Longitude: 0,
			Altitude:             &a,
			UncertaintySemiMajor: u8(1),
			UncertaintySemiMinor: u8(1),
			AngleMajorAxis:       u8(0),
			UncertaintyAltitude:  u8(1),
			Confidence:           u8(50),
		}
		data, err := gi.Encode()
		if err != nil {
			t.Fatalf("alt=%d: Encode error: %v", alt, err)
		}
		got, err := DecodeGeographicalInfo(data)
		if err != nil {
			t.Fatalf("alt=%d: Decode error: %v", alt, err)
		}
		if got.Altitude == nil || *got.Altitude != alt {
			t.Errorf("alt=%d: round-trip got %v", alt, got.Altitude)
		}
	}
}

// Encode for shapes with required fields must reject missing fields
// rather than silently writing zeros. TS 23.032 specifies which octets
// are present per shape; a nil required field is a caller bug.
func TestGeoEncode_RejectsMissingRequiredFields(t *testing.T) {
	cases := []struct {
		name string
		gi   *GeographicalInfo
	}{
		{
			"EllipsoidPointUncertainty missing UncertaintyCode",
			&GeographicalInfo{ShapeType: ShapeEllipsoidPointUncertainty, Latitude: 0, Longitude: 0},
		},
		{
			"EllipsoidPointUncertaintyEllipse missing SemiMajor",
			&GeographicalInfo{
				ShapeType: ShapeEllipsoidPointUncertaintyEllipse,
				Latitude:  0, Longitude: 0,
				UncertaintySemiMinor: u8(1), AngleMajorAxis: u8(0), Confidence: u8(50),
			},
		},
		{
			"EllipsoidPointAltitude missing Altitude",
			&GeographicalInfo{
				ShapeType: ShapeEllipsoidPointAltitude,
				Latitude:  0, Longitude: 0,
				UncertaintySemiMajor: u8(1), UncertaintySemiMinor: u8(1),
				AngleMajorAxis: u8(0), UncertaintyAltitude: u8(1), Confidence: u8(50),
			},
		},
		{
			"EllipsoidArc missing InnerRadius",
			&GeographicalInfo{
				ShapeType: ShapeEllipsoidArc,
				Latitude:  0, Longitude: 0,
				UncertaintyRadius: u8(1), OffsetAngle: u8(0),
				IncludedAngle: u8(90), Confidence: u8(50),
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := tc.gi.Encode(); err == nil {
				t.Errorf("%s: expected error, got nil", tc.name)
			}
		})
	}
}

// Full valid shapes must still round-trip.
func TestGeoEncode_ValidShapesRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		gi   *GeographicalInfo
	}{
		{
			"EllipsoidPoint",
			&GeographicalInfo{ShapeType: ShapeEllipsoidPoint, Latitude: 22.63, Longitude: 113.03},
		},
		{
			"EllipsoidPointUncertainty",
			&GeographicalInfo{
				ShapeType: ShapeEllipsoidPointUncertainty,
				Latitude:  22.63, Longitude: 113.03,
				UncertaintyCode: u8(10),
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := tc.gi.Encode()
			if err != nil {
				t.Fatalf("Encode: %v", err)
			}
			got, err := DecodeGeographicalInfo(data)
			if err != nil {
				t.Fatalf("Decode: %v", err)
			}
			if got.ShapeType != tc.gi.ShapeType {
				t.Errorf("ShapeType: got %d, want %d", got.ShapeType, tc.gi.ShapeType)
			}
			if math.Abs(got.Latitude-tc.gi.Latitude) > 0.001 {
				t.Errorf("Latitude: got %v, want %v", got.Latitude, tc.gi.Latitude)
			}
			if math.Abs(got.Longitude-tc.gi.Longitude) > 0.001 {
				t.Errorf("Longitude: got %v, want %v", got.Longitude, tc.gi.Longitude)
			}
		})
	}
}

// Confirm the 7-bit fields encode bit-for-bit at their maximum valid
// value (0x7F) without any masking-induced corruption.
func TestGeoEncode_SevenBitFieldsExactBytes(t *testing.T) {
	gi := &GeographicalInfo{
		ShapeType: ShapeEllipsoidPointUncertainty,
		Latitude:  0, Longitude: 0,
		UncertaintyCode: u8(0x7F),
	}
	data, err := gi.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	want := []byte{0x10, 0, 0, 0, 0, 0, 0, 0x7F}
	if !bytes.Equal(data, want) {
		t.Errorf("Encode bytes: got %x, want %x", data, want)
	}
}

// Encode must reject 7-bit shape-specific fields whose value exceeds 127.
// Previously these were silently masked with & 0x7F, which landed an
// unrelated value on the wire (e.g. 200 -> 72). The encoder now refuses
// the input so the caller sees the error instead of silent corruption.
func TestGeoEncode_Rejects7BitFieldsOverflow(t *testing.T) {
	cases := []struct {
		name string
		gi   *GeographicalInfo
	}{
		{
			"UncertaintyCode=200",
			&GeographicalInfo{ShapeType: ShapeEllipsoidPointUncertainty, UncertaintyCode: u8(200)},
		},
		{
			"UncertaintyCode=128",
			&GeographicalInfo{ShapeType: ShapeEllipsoidPointUncertainty, UncertaintyCode: u8(128)},
		},
		{
			"Ellipse: SemiMajor=200",
			&GeographicalInfo{
				ShapeType:            ShapeEllipsoidPointUncertaintyEllipse,
				UncertaintySemiMajor: u8(200), UncertaintySemiMinor: u8(1),
				AngleMajorAxis: u8(0), Confidence: u8(50),
			},
		},
		{
			"Ellipse: Confidence=255",
			&GeographicalInfo{
				ShapeType:            ShapeEllipsoidPointUncertaintyEllipse,
				UncertaintySemiMajor: u8(1), UncertaintySemiMinor: u8(1),
				AngleMajorAxis: u8(0), Confidence: u8(255),
			},
		},
		{
			"Altitude shape: UncertaintyAltitude=200",
			func() *GeographicalInfo {
				alt := int16(100)
				return &GeographicalInfo{
					ShapeType: ShapeEllipsoidPointAltitude,
					Altitude:  &alt, UncertaintySemiMajor: u8(1), UncertaintySemiMinor: u8(1),
					AngleMajorAxis: u8(0), UncertaintyAltitude: u8(200), Confidence: u8(50),
				}
			}(),
		},
		{
			"Arc: UncertaintyRadius=200",
			&GeographicalInfo{
				ShapeType: ShapeEllipsoidArc,
				InnerRadius:   u16(1000),
				UncertaintyRadius: u8(200), OffsetAngle: u8(0),
				IncludedAngle: u8(90), Confidence: u8(50),
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := tc.gi.Encode(); err == nil {
				t.Errorf("%s: expected error, got nil", tc.name)
			}
		})
	}
}

// u8 returns a pointer to the given uint8 literal.
func u8(v uint8) *uint8 { return &v }

// u16 returns a pointer to the given uint16 literal.
func u16(v uint16) *uint16 { return &v }

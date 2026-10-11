// geo_test.go
//
// Tests for geographical-information encoder validation per 3GPP TS 23.032.
// The Encode() path must reject caller errors (out-of-range lat/lon,
// out-of-range altitude, missing required fields) at the boundary rather
// than accepting invalid input or zero-filling required fields.
package gsmmap

import (
	"bytes"
	"math"
	"testing"
)

// TS 23.032 V19.0.0 §6.1: the maximum latitude bin includes ±90.
// Encode must reject latitudes outside [-90, 90].
func TestGeoEncode_RejectsOutOfRangeLatitude(t *testing.T) {
	for _, lat := range []float64{-90.01, 90.01, 200, -500} {
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

// Encode accepts latitude in [-90, 90] and longitude in [-180, 180).
// TS 23.032 V19.0.0 §6.1 extends the highest latitude bin to the poles.
func TestGeoEncode_AcceptsBoundaryLatLon(t *testing.T) {
	cases := []struct {
		lat, lon float64
	}{
		{0, 0},
		{90, 0},
		{-90, 0},
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

// TS 23.032 V19.0.0 §6.1: valid inputs immediately inside the range
// bounds belong to the boundary bins and must be accepted.
func TestGeoEncode_AcceptsULPInsideBoundary(t *testing.T) {
	for _, tc := range []struct {
		lat, lon float64
		want     []byte
	}{
		{math.Nextafter(90, 0), 0, []byte{0, 0x7f, 0xff, 0xff, 0, 0, 0}},
		{math.Nextafter(-90, 0), 0, []byte{0, 0xff, 0xff, 0xff, 0, 0, 0}},
		{0, math.Nextafter(180, 0), []byte{0, 0, 0, 0, 0x7f, 0xff, 0xff}},
		{0, math.Nextafter(-180, 0), []byte{0, 0, 0, 0, 0x80, 0, 0}},
	} {
		gi := &GeographicalInfo{ShapeType: ShapeEllipsoidPoint, Latitude: tc.lat, Longitude: tc.lon}
		got, err := gi.Encode()
		if err != nil {
			t.Fatalf("Encode(%v, %v): %v", tc.lat, tc.lon, err)
		}
		if !bytes.Equal(got, tc.want) {
			t.Errorf("Encode(%v, %v) = %x, want %x", tc.lat, tc.lon, got, tc.want)
		}
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
				ShapeType:         ShapeEllipsoidArc,
				InnerRadius:       u16(1000),
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

// Encode must reject angular fields whose octet value exceeds the
// TS 23.032 spec-defined semantic bound of 179:
//   - AngleMajorAxis: stored = degrees directly, valid 0..179°.
//   - OffsetAngle / IncludedAngle (arc): stored × 2 = degrees; encoded
//     octet valid 0..179 (representing 0..358° in 2° steps). Values
//     180..255 are spec-invalid.
func TestGeoEncode_RejectsAngularFieldsOutOfRange(t *testing.T) {
	alt := int16(100)
	cases := []struct {
		name string
		gi   *GeographicalInfo
	}{
		{
			"Ellipse: AngleMajorAxis=180",
			&GeographicalInfo{
				ShapeType:            ShapeEllipsoidPointUncertaintyEllipse,
				UncertaintySemiMajor: u8(1), UncertaintySemiMinor: u8(1),
				AngleMajorAxis: u8(180), Confidence: u8(50),
			},
		},
		{
			"Ellipse: AngleMajorAxis=255",
			&GeographicalInfo{
				ShapeType:            ShapeEllipsoidPointUncertaintyEllipse,
				UncertaintySemiMajor: u8(1), UncertaintySemiMinor: u8(1),
				AngleMajorAxis: u8(255), Confidence: u8(50),
			},
		},
		{
			"Altitude: AngleMajorAxis=200",
			&GeographicalInfo{
				ShapeType: ShapeEllipsoidPointAltitude,
				Altitude:  &alt, UncertaintySemiMajor: u8(1), UncertaintySemiMinor: u8(1),
				AngleMajorAxis: u8(200), UncertaintyAltitude: u8(1), Confidence: u8(50),
			},
		},
		{
			"Arc: OffsetAngle=180",
			&GeographicalInfo{
				ShapeType:   ShapeEllipsoidArc,
				InnerRadius: u16(1000), UncertaintyRadius: u8(1),
				OffsetAngle: u8(180), IncludedAngle: u8(90), Confidence: u8(50),
			},
		},
		{
			"Arc: IncludedAngle=200",
			&GeographicalInfo{
				ShapeType:   ShapeEllipsoidArc,
				InnerRadius: u16(1000), UncertaintyRadius: u8(1),
				OffsetAngle: u8(0), IncludedAngle: u8(200), Confidence: u8(50),
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

// Encode must accept angular fields at the upper valid boundary (179)
// and also at the lower boundary (0).
func TestGeoEncode_AcceptsAngularFieldsAtBoundary(t *testing.T) {
	alt := int16(100)
	cases := []struct {
		name string
		gi   *GeographicalInfo
	}{
		{
			"Ellipse: AngleMajorAxis=0",
			&GeographicalInfo{
				ShapeType:            ShapeEllipsoidPointUncertaintyEllipse,
				UncertaintySemiMajor: u8(1), UncertaintySemiMinor: u8(1),
				AngleMajorAxis: u8(0), Confidence: u8(50),
			},
		},
		{
			"Ellipse: AngleMajorAxis=179",
			&GeographicalInfo{
				ShapeType:            ShapeEllipsoidPointUncertaintyEllipse,
				UncertaintySemiMajor: u8(1), UncertaintySemiMinor: u8(1),
				AngleMajorAxis: u8(179), Confidence: u8(50),
			},
		},
		{
			"Altitude: AngleMajorAxis=179",
			&GeographicalInfo{
				ShapeType: ShapeEllipsoidPointAltitude,
				Altitude:  &alt, UncertaintySemiMajor: u8(1), UncertaintySemiMinor: u8(1),
				AngleMajorAxis: u8(179), UncertaintyAltitude: u8(1), Confidence: u8(50),
			},
		},
		{
			"Arc: OffsetAngle=179, IncludedAngle=179",
			&GeographicalInfo{
				ShapeType:   ShapeEllipsoidArc,
				InnerRadius: u16(1000), UncertaintyRadius: u8(1),
				OffsetAngle: u8(179), IncludedAngle: u8(179), Confidence: u8(50),
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := tc.gi.Encode(); err != nil {
				t.Errorf("%s: unexpected error: %v", tc.name, err)
			}
		})
	}
}

// Decode deliberately accepts spec-invalid angular octets (the receiver
// needs to observe what a peer sent), but re-encoding such a decoded
// value must fail at the boundary check rather than silently re-emit
// non-spec-compliant bytes. This locks in the Postel's-law asymmetry
// established by PR #25 / PR #26.
func TestGeoDecodeThenEncode_SpecInvalidAngleFailsOnReEncode(t *testing.T) {
	// Hand-craft EllipsoidPointUncertaintyEllipse wire bytes with
	// AngleMajorAxis=200 (invalid; spec caps at 179). Decode must succeed;
	// re-encoding must fail.
	wire := []byte{
		0x30,             // ShapeType=3 (UncertaintyEllipse)<<4
		0x00, 0x00, 0x00, // lat = 0
		0x00, 0x00, 0x00, // lon = 0
		0x01, // SemiMajor = 1
		0x01, // SemiMinor = 1
		200,  // AngleMajorAxis = 200 (SPEC-INVALID)
		0x32, // Confidence = 50
	}
	gi, err := DecodeGeographicalInfo(wire)
	if err != nil {
		t.Fatalf("Decode of spec-invalid AngleMajorAxis=200: unexpected error: %v", err)
	}
	if gi.AngleMajorAxis == nil || *gi.AngleMajorAxis != 200 {
		t.Fatalf("AngleMajorAxis after decode: got %v, want 200", gi.AngleMajorAxis)
	}
	if _, err := gi.Encode(); err == nil {
		t.Error("Encode of spec-invalid AngleMajorAxis=200: expected error, got nil")
	}
}

// u8 returns a pointer to the given uint8 literal.
func u8(v uint8) *uint8 { return &v }

// u16 returns a pointer to the given uint16 literal.
func u16(v uint16) *uint16 { return &v }

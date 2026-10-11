// geo_test.go
//
// Tests for geographical-information encoder validation per 3GPP TS 23.032.
// The Encode() path must reject caller errors (out-of-range lat/lon,
// out-of-range altitude, missing required fields) at the boundary rather
// than accepting invalid input or zero-filling required fields.
package gsmmap

import (
	"bytes"
	"errors"
	"fmt"
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
				ShapeType: ShapeEllipsoidPointAltitudeUncertaintyEllipsoid,
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
			ShapeType: ShapeEllipsoidPointAltitudeUncertaintyEllipsoid,
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
				ShapeType: ShapeEllipsoidPointAltitudeUncertaintyEllipsoid,
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
					ShapeType: ShapeEllipsoidPointAltitudeUncertaintyEllipsoid,
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
				ShapeType: ShapeEllipsoidPointAltitudeUncertaintyEllipsoid,
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
				ShapeType: ShapeEllipsoidPointAltitudeUncertaintyEllipsoid,
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

// TS 23.032 V19.0.0 §§6.2, 7.3.2: C=10, x=0.1 and seven-bit K.
func TestGeoUncertainty(t *testing.T) {
	for _, tc := range []struct {
		k      byte
		meters float64
	}{
		{0, 0}, {1, 1.0000000000000009}, {2, 2.100000000000002},
	} {
		t.Run(fmt.Sprint(tc.k), func(t *testing.T) {
			got, err := DecodeGeographicalInfo([]byte{0x10, 0, 0, 0, 0, 0, 0, tc.k})
			if err != nil {
				t.Fatal(err)
			}
			if got.UncertaintyMeters == nil || got.UncertaintyCode == nil {
				t.Fatal("missing uncertainty")
			}
			wantEqual(t, "code", tc.k, *got.UncertaintyCode)
			wantEqual(t, "metres", tc.meters, *got.UncertaintyMeters)
		})
	}
}

// TS 23.032 V19.0.0 §§7.3.1, 7.3.7: point and arc layouts.
func TestGeoDecodeLengths(t *testing.T) {
	for _, tc := range []struct {
		name    string
		data    []byte
		want    error
		message string
	}{
		{"empty", nil, ErrGeoInformationTooShort, "geographical information too short"},
		{"point header", []byte{0}, ErrGeoPointInvalidLength, "ellipsoid point requires 7 octets, got 1"},
		{"arc truncated", []byte{0xa0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, ErrGeoArcInvalidLength, "ellipsoid arc requires 13 octets, got 12"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DecodeGeographicalInfo(tc.data)
			if got != nil || !errors.Is(err, tc.want) {
				t.Fatalf("Decode = %v, %v; want nil, %v", got, err, tc.want)
			}
			wantEqual(t, "error", fmt.Sprintf("%s: %s", tc.want, tc.message), err.Error())
		})
	}
	got, err := DecodeGeographicalInfo([]byte{0xa0, 0, 0, 0, 0, 0, 0, 0x12, 0x34, 3, 4, 5, 99})
	if err != nil {
		t.Fatal(err)
	}
	if got.InnerRadius == nil || got.UncertaintyRadius == nil || got.OffsetAngle == nil || got.IncludedAngle == nil || got.Confidence == nil {
		t.Fatal("missing arc fields")
	}
	wantEqual(t, "shape", ShapeEllipsoidArc, got.ShapeType)
	wantEqual(t, "radius", uint16(0x1234), *got.InnerRadius)
	wantEqual(t, "uncertainty", uint8(3), *got.UncertaintyRadius)
	wantEqual(t, "offset", uint8(4), *got.OffsetAngle)
	wantEqual(t, "included", uint8(5), *got.IncludedAngle)
	wantEqual(t, "confidence", uint8(99), *got.Confidence)
}

// TS 23.032 V19.0.0 §6.1: 23-bit latitude magnitude and signed 24-bit longitude.
// The maximum latitude bin includes the poles; Decode returns its lower edge.
func TestGeoCoordinates(t *testing.T) {
	for _, tc := range []struct {
		name     string
		lat, lon float64
		want     HexBytes
		wantErr  error
	}{
		{"south pole", -90, 0, HexBytes{0, 0xff, 0xff, 0xff, 0, 0, 0}, nil},
		{"north pole", 90, 0, HexBytes{0, 0x7f, 0xff, 0xff, 0, 0, 0}, nil},
		{"west boundary", 0, -180, HexBytes{0, 0, 0, 0, 0x80, 0, 0}, nil},
		{"east boundary", 0, 180, nil, ErrGeoLongitudeOutOfRange},
		{"maximum quantum", 8388607.0 * 90 / 8388608, 8388607.0 * 360 / 16777216, HexBytes{0, 0x7f, 0xff, 0xff, 0x7f, 0xff, 0xff}, nil},
		{"southern hemisphere", -45, -90, HexBytes{0, 0xc0, 0, 0, 0xc0, 0, 0}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := &GeographicalInfo{ShapeType: ShapeEllipsoidPoint, Latitude: tc.lat, Longitude: tc.lon}
			data, err := in.Encode()
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Encode = %x, %v; want %v", data, err, tc.wantErr)
			}
			wantEqual(t, "octets", tc.want, HexBytes(data))
			if err != nil {
				return
			}
			got, err := DecodeGeographicalInfo(tc.want)
			if err != nil {
				t.Fatal(err)
			}
			wantLat := tc.lat
			if tc.lat == 90 || tc.lat == -90 {
				wantLat = math.Copysign(8388607.0*90/8388608, tc.lat)
			}
			wantEqual(t, "latitude", wantLat, got.Latitude)
			wantEqual(t, "longitude", tc.lon, got.Longitude)
		})
	}
}

// TS 23.032 V19.0.0 §§6.3, 7.3.6: zero altitude has an unset sign bit.
func TestGeoZeroAltitude(t *testing.T) {
	alt := int16(0)
	zero := uint8(0)
	in := &GeographicalInfo{ShapeType: ShapeEllipsoidPointAltitudeUncertaintyEllipsoid, Altitude: &alt,
		UncertaintySemiMajor: &zero, UncertaintySemiMinor: &zero, AngleMajorAxis: &zero,
		UncertaintyAltitude: &zero, Confidence: &zero}
	got, err := in.Encode()
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "octets", HexBytes{0x90, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, HexBytes(got))
}

// TS 23.032 V19.0.0 §§6.5, 7.3.7: confidence occupies seven bits.
func TestGeoArcConfidence(t *testing.T) {
	for _, confidence := range []uint8{99, 128} {
		t.Run(fmt.Sprint(confidence), func(t *testing.T) {
			radius := uint16(0x1234)
			uncertainty, offset, included := uint8(3), uint8(4), uint8(5)
			in := &GeographicalInfo{ShapeType: ShapeEllipsoidArc, InnerRadius: &radius,
				UncertaintyRadius: &uncertainty, OffsetAngle: &offset, IncludedAngle: &included, Confidence: &confidence}
			got, err := in.Encode()
			if confidence == 128 {
				if got != nil || !errors.Is(err, ErrGeoCodeOutOfRange) {
					t.Fatalf("Encode = %x, %v; want confidence range error", got, err)
				}
				wantEqual(t, "error", fmt.Sprintf("%s: Confidence out of range [0, 127]: 128", ErrGeoCodeOutOfRange), err.Error())
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			wantEqual(t, "octets", HexBytes{0xa0, 0, 0, 0, 0, 0, 0, 0x12, 0x34, 3, 4, 5, 99}, HexBytes(got))
		})
	}
}

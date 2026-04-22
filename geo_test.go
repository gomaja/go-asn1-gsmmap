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

// Encode must reject latitudes outside [-90, 90].
func TestGeoEncode_RejectsOutOfRangeLatitude(t *testing.T) {
	for _, lat := range []float64{-90.01, 90.01, 200, -500} {
		gi := &GeographicalInfo{ShapeType: ShapeEllipsoidPoint, Latitude: lat, Longitude: 0}
		if _, err := gi.Encode(); err == nil {
			t.Errorf("Latitude=%v: expected error, got nil", lat)
		}
	}
}

// Encode must reject longitudes outside [-180, 180].
func TestGeoEncode_RejectsOutOfRangeLongitude(t *testing.T) {
	for _, lon := range []float64{-180.01, 180.01, 500, -500} {
		gi := &GeographicalInfo{ShapeType: ShapeEllipsoidPoint, Latitude: 0, Longitude: lon}
		if _, err := gi.Encode(); err == nil {
			t.Errorf("Longitude=%v: expected error, got nil", lon)
		}
	}
}

// Encode must accept valid lat/lon at the boundary values.
func TestGeoEncode_AcceptsBoundaryLatLon(t *testing.T) {
	cases := []struct {
		lat, lon float64
	}{
		{0, 0},
		{90, 180},
		{-90, -180},
		{89.9999, 179.9999},
	}
	for _, tc := range cases {
		gi := &GeographicalInfo{ShapeType: ShapeEllipsoidPoint, Latitude: tc.lat, Longitude: tc.lon}
		if _, err := gi.Encode(); err != nil {
			t.Errorf("Encode(lat=%v, lon=%v): unexpected error: %v", tc.lat, tc.lon, err)
		}
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

// Confirm the 7-bit fields encode/decode bit-for-bit.
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

// u8 returns a pointer to the given uint8 literal.
func u8(v uint8) *uint8 { return &v }

package gsmmap_test

import (
	"bytes"
	"errors"
	"math"
	"testing"

	gsmmap "github.com/gomaja/go-asn1-gsmmap"
)

// TS 23.032 V19.0.0 §6.1: latitude uses absolute-value bins (the last
// includes the pole); signed longitude bins include their lower edge.
func TestGeoPointCoordinateBins(t *testing.T) {
	const latStep = 90.0 / (1 << 23)
	const lonStep = 360.0 / (1 << 24)
	for _, tc := range []struct {
		name     string
		lat, lon float64
		want     []byte
	}{
		{"origin", 0, 0, []byte{0, 0, 0, 0, 0, 0, 0}},
		// -0 is the southern bin 0, which decoding 00800000 returns.
		{"negative zero", math.Copysign(0, -1), math.Copysign(0, -1), []byte{0, 0x80, 0, 0, 0, 0, 0}},
		{"smallest positive longitude", 0, math.SmallestNonzeroFloat64, []byte{0, 0, 0, 0, 0, 0, 0}},
		{"smallest negative longitude", 0, -math.SmallestNonzeroFloat64, []byte{0, 0, 0, 0, 0xff, 0xff, 0xff}},
		{"latitude inside first bin", 0.75 * latStep, 0, []byte{0, 0, 0, 0, 0, 0, 0}},
		{"south inside first bin", -0.75 * latStep, 0, []byte{0, 0x80, 0, 0, 0, 0, 0}},
		{"longitude inside first positive bin", 0, 0.75 * lonStep, []byte{0, 0, 0, 0, 0, 0, 0}},
		{"longitude inside first negative bin", 0, -0.25 * lonStep, []byte{0, 0, 0, 0, 0xff, 0xff, 0xff}},
		{"latitude below bin edge", math.Nextafter(latStep, 0), 0, []byte{0, 0, 0, 0, 0, 0, 0}},
		{"latitude on bin edge", latStep, 0, []byte{0, 0, 0, 1, 0, 0, 0}},
		{"latitude above bin edge", math.Nextafter(latStep, 90), 0, []byte{0, 0, 0, 1, 0, 0, 0}},
		{"longitude below positive bin edge", 0, math.Nextafter(lonStep, 0), []byte{0, 0, 0, 0, 0, 0, 0}},
		{"longitude on positive bin edge", 0, lonStep, []byte{0, 0, 0, 0, 0, 0, 1}},
		{"longitude above positive bin edge", 0, math.Nextafter(lonStep, 180), []byte{0, 0, 0, 0, 0, 0, 1}},
		{"longitude below negative bin edge", 0, math.Nextafter(-lonStep, -180), []byte{0, 0, 0, 0, 0xff, 0xff, 0xfe}},
		{"longitude on negative bin edge", 0, -lonStep, []byte{0, 0, 0, 0, 0xff, 0xff, 0xff}},
		{"longitude above negative bin edge", 0, math.Nextafter(-lonStep, 0), []byte{0, 0, 0, 0, 0xff, 0xff, 0xff}},
		{"positive mid-bin", 45 + 0.75*latStep, 90 + 0.75*lonStep, []byte{0, 0x40, 0, 0, 0x40, 0, 0}},
		{"negative mid-bin", -45 - 0.75*latStep, -90 - 0.25*lonStep, []byte{0, 0xc0, 0, 0, 0xbf, 0xff, 0xff}},
		{"north pole", 90, 0, []byte{0, 0x7f, 0xff, 0xff, 0, 0, 0}},
		{"south pole", -90, 0, []byte{0, 0xff, 0xff, 0xff, 0, 0, 0}},
		{"inside north pole", math.Nextafter(90, 0), 0, []byte{0, 0x7f, 0xff, 0xff, 0, 0, 0}},
		{"inside south pole", math.Nextafter(-90, 0), 0, []byte{0, 0xff, 0xff, 0xff, 0, 0, 0}},
		{"inside east boundary", 0, math.Nextafter(180, 0), []byte{0, 0, 0, 0, 0x7f, 0xff, 0xff}},
		{"west boundary", 0, -180, []byte{0, 0, 0, 0, 0x80, 0, 0}},
		{"inside west boundary", 0, math.Nextafter(-180, 0), []byte{0, 0, 0, 0, 0x80, 0, 0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := (&gsmmap.GeographicalInfo{ShapeType: gsmmap.ShapeEllipsoidPoint, Latitude: tc.lat, Longitude: tc.lon}).Encode()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, tc.want) {
				t.Fatalf("Encode(%v, %v) = %x, want %x", tc.lat, tc.lon, got, tc.want)
			}
		})
	}
}

func TestGeoPointCoordinateRangeControls(t *testing.T) {
	for _, tc := range []struct {
		name     string
		lat, lon float64
		want     error
	}{
		{"below south pole", math.Nextafter(-90, math.Inf(-1)), 0, gsmmap.ErrGeoLatitudeOutOfRange},
		{"above north pole", math.Nextafter(90, math.Inf(1)), 0, gsmmap.ErrGeoLatitudeOutOfRange},
		{"below west boundary", 0, math.Nextafter(-180, math.Inf(-1)), gsmmap.ErrGeoLongitudeOutOfRange},
		{"east boundary", 0, 180, gsmmap.ErrGeoLongitudeOutOfRange},
		{"latitude NaN", math.NaN(), 0, gsmmap.ErrGeoLatitudeNotFinite},
		{"longitude NaN", 0, math.NaN(), gsmmap.ErrGeoLongitudeNotFinite},
		{"latitude positive infinity", math.Inf(1), 0, gsmmap.ErrGeoLatitudeNotFinite},
		{"latitude negative infinity", math.Inf(-1), 0, gsmmap.ErrGeoLatitudeNotFinite},
		{"longitude positive infinity", 0, math.Inf(1), gsmmap.ErrGeoLongitudeNotFinite},
		{"longitude negative infinity", 0, math.Inf(-1), gsmmap.ErrGeoLongitudeNotFinite},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := (&gsmmap.GeographicalInfo{ShapeType: gsmmap.ShapeEllipsoidPoint, Latitude: tc.lat, Longitude: tc.lon}).Encode()
			if got != nil || !errors.Is(err, tc.want) {
				t.Fatalf("Encode = %x, %v; want nil, %v", got, err, tc.want)
			}
		})
	}
}

// TS 23.032 V19.0.0 §6.1: the decoder continues to return each bin's
// lower magnitude/longitude edge, including the extended polar bin.
func TestGeoPointReceiverBinEdges(t *testing.T) {
	for _, tc := range []struct {
		name     string
		data     []byte
		lat, lon float64
	}{
		{"zero", []byte{0, 0, 0, 0, 0, 0, 0}, 0, 0},
		{"first positive bins", []byte{0, 0, 0, 1, 0, 0, 1}, 90.0 / (1 << 23), 360.0 / (1 << 24)},
		{"first negative bins", []byte{0, 0x80, 0, 1, 0xff, 0xff, 0xff}, -90.0 / (1 << 23), -360.0 / (1 << 24)},
		{"nontrivial bins", []byte{0, 0x12, 0x34, 0x56, 0xed, 0xcb, 0xaa}, 0x123456 * 90.0 / (1 << 23), -0x123456 * 360.0 / (1 << 24)},
		{"north polar bin", []byte{0, 0x7f, 0xff, 0xff, 0x7f, 0xff, 0xff}, 0x7fffff * 90.0 / (1 << 23), 0x7fffff * 360.0 / (1 << 24)},
		{"south polar bin", []byte{0, 0xff, 0xff, 0xff, 0x80, 0, 0}, -0x7fffff * 90.0 / (1 << 23), -180},
	} {
		t.Run(tc.name, func(t *testing.T) {
			point, err := gsmmap.DecodeGeographicalInfo(tc.data)
			if err != nil {
				t.Fatal(err)
			}
			if point.Latitude != tc.lat || point.Longitude != tc.lon {
				t.Fatalf("Decode = (%v, %v), want (%v, %v)", point.Latitude, point.Longitude, tc.lat, tc.lon)
			}
			got, err := point.Encode()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, tc.data) {
				t.Fatalf("re-encoded bin edges = %x, want %x", got, tc.data)
			}
		})
	}
}

// TS 23.032 V19.0.0 §§7.3.1, 7.3.2, 7.3.3, 7.3.6, 7.3.7: supported shapes
// share the same point-coordinate coding in octets 2 through 7.
func TestGeoPointBinsAcrossShapes(t *testing.T) {
	zero, radius, altitude := uint8(0), uint16(0), int16(0)
	for _, shape := range []gsmmap.ShapeType{
		gsmmap.ShapeEllipsoidPoint, gsmmap.ShapeEllipsoidPointUncertainty,
		gsmmap.ShapeEllipsoidPointUncertaintyEllipse, gsmmap.ShapeEllipsoidPointAltitudeUncertaintyEllipsoid,
		gsmmap.ShapeEllipsoidArc,
	} {
		in := &gsmmap.GeographicalInfo{
			ShapeType: shape, Latitude: -90, Longitude: -0.25 * 360 / (1 << 24),
			UncertaintyCode: &zero, UncertaintySemiMajor: &zero, UncertaintySemiMinor: &zero,
			AngleMajorAxis: &zero, Confidence: &zero, Altitude: &altitude,
			UncertaintyAltitude: &zero, InnerRadius: &radius, UncertaintyRadius: &zero,
			OffsetAngle: &zero, IncludedAngle: &zero,
		}
		got, err := in.Encode()
		if err != nil {
			t.Fatalf("shape %d: %v", shape, err)
		}
		if got[0] != byte(shape)<<4 || !bytes.Equal(got[1:7], []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
			t.Errorf("shape %d: encoded coordinates = %x", shape, got)
		}
	}
}

func FuzzGeoPointCoordinateBins(f *testing.F) {
	for _, point := range [][2]float64{
		{0, 0}, {90, 0}, {-90, -180}, {45.000008046627045, 90.00001609325409},
		{-45.000008046627045, -90.00000536441803},
		{math.Nextafter(90, 0), math.Nextafter(180, 0)},
		{math.Nextafter(-90, 0), math.Nextafter(-180, 0)},
		{0, -math.SmallestNonzeroFloat64}, {math.NaN(), 0}, {0, math.Inf(1)}, {91, 181},
	} {
		f.Add(point[0], point[1])
	}
	f.Fuzz(func(t *testing.T, lat, lon float64) {
		data, err := (&gsmmap.GeographicalInfo{ShapeType: gsmmap.ShapeEllipsoidPoint, Latitude: lat, Longitude: lon}).Encode()
		if math.IsNaN(lat) || math.IsInf(lat, 0) || math.IsNaN(lon) || math.IsInf(lon, 0) || lat < -90 || lat > 90 || lon < -180 || lon >= 180 {
			if err == nil || data != nil {
				t.Fatalf("invalid coordinates (%v, %v) encoded as %x, %v", lat, lon, data, err)
			}
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		point, err := gsmmap.DecodeGeographicalInfo(data)
		if err != nil {
			t.Fatal(err)
		}
		// Compare the represented interval, rather than duplicating the encoder's
		// quantization calculation. Latitude's final bin includes its upper edge.
		magnitude := math.Abs(point.Latitude)
		upperLatitude := magnitude + 90.0/(1<<23)
		if magnitude > math.Abs(lat) || math.Abs(lat) > upperLatitude || (math.Abs(lat) == upperLatitude && math.Abs(lat) != 90) {
			t.Fatalf("latitude %v outside decoded magnitude bin [%v, %v)", lat, magnitude, upperLatitude)
		}
		if (data[1]&0x80 != 0) != (lat < 0) {
			t.Fatalf("latitude %v has incorrect sign in %x", lat, data)
		}
		upperLongitude := point.Longitude + 360.0/(1<<24)
		if point.Longitude > lon || lon >= upperLongitude {
			t.Fatalf("longitude %v outside decoded bin [%v, %v)", lon, point.Longitude, upperLongitude)
		}
	})
}

// Every latitude bin, north and south, decodes to a value that encodes to
// the same octets; the southern bin 0 decodes to -0 (TS 23.032 V19.0.0 §6.1).
func TestGeoPointSignRoundTrip(t *testing.T) {
	for _, wire := range [][]byte{
		{0, 0x00, 0, 0, 0, 0, 0},
		{0, 0x80, 0, 0, 0, 0, 0},
		{0, 0x80, 0, 1, 0, 0, 0},
		{0, 0xff, 0xff, 0xff, 0, 0, 0},
	} {
		gi, err := gsmmap.DecodeGeographicalInfo(wire)
		if err != nil {
			t.Fatal(err)
		}
		got, err := gi.Encode()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, wire) {
			t.Errorf("%x decodes to %v and encodes to %x", wire, gi.Latitude, got)
		}
	}
}

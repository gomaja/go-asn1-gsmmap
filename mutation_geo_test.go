package gsmmap

import (
	"errors"
	"fmt"
	"testing"
)

// TS 23.032 V19.0.0 §§6.2, 7.3.2: C=10, x=0.1 and seven-bit K.
func TestMutationGeoUncertainty(t *testing.T) {
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
func TestMutationGeoDecodeLengths(t *testing.T) {
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
// Encode's documented range rejects the poles included in the maximum latitude bin.
func TestMutationGeoCoordinates(t *testing.T) {
	for _, tc := range []struct {
		name     string
		lat, lon float64
		want     HexBytes
		wantErr  error
	}{
		{"south pole", -90, 0, nil, ErrGeoLatitudeOutOfRange},
		{"north pole", 90, 0, nil, ErrGeoLatitudeOutOfRange},
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
			wantEqual(t, "latitude", tc.lat, got.Latitude)
			wantEqual(t, "longitude", tc.lon, got.Longitude)
		})
	}
}

// TS 23.032 V19.0.0 §§6.3, 7.3.6: zero altitude has an unset sign bit.
func TestMutationGeoZeroAltitude(t *testing.T) {
	alt := int16(0)
	zero := uint8(0)
	in := &GeographicalInfo{ShapeType: ShapeEllipsoidPointAltitude, Altitude: &alt,
		UncertaintySemiMajor: &zero, UncertaintySemiMinor: &zero, AngleMajorAxis: &zero,
		UncertaintyAltitude: &zero, Confidence: &zero}
	got, err := in.Encode()
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, "octets", HexBytes{0x90, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, HexBytes(got))
}

// TS 23.032 V19.0.0 §§6.5, 7.3.7: confidence occupies seven bits.
func TestMutationGeoArcConfidence(t *testing.T) {
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

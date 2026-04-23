package gsmmap

import (
	"fmt"
	"math"
)

// ShapeType represents the type of geographical shape per 3GPP TS 23.032.
type ShapeType int

const (
	ShapeEllipsoidPoint                   ShapeType = 0
	ShapeEllipsoidPointUncertainty        ShapeType = 1
	ShapeEllipsoidPointUncertaintyEllipse ShapeType = 3
	ShapeEllipsoidPointAltitude           ShapeType = 9
	ShapeEllipsoidArc                     ShapeType = 10
)

// GeographicalInfo represents decoded geographical information per 3GPP TS 23.032.
//
// Field encoding notes:
//   - 7-bit fields (UncertaintyCode, UncertaintySemiMajor/Minor,
//     UncertaintyAltitude, UncertaintyRadius, Confidence) carry values
//     0..127 in the low 7 bits of their octet; bit 7 is reserved.
//   - AngleMajorAxis carries the orientation of the semi-major axis in
//     degrees, integer 0..179 (1° steps).
//   - OffsetAngle and IncludedAngle for the Ellipsoid Arc shape carry
//     the octet value directly: stored N represents N × 2 degrees, so
//     the valid octet range is 0..179 (representing 0..358° in 2° steps).
type GeographicalInfo struct {
	ShapeType            ShapeType `json:"ShapeType"`
	Latitude             float64   `json:"Latitude"`
	Longitude            float64   `json:"Longitude"`
	UncertaintyCode      *uint8    `json:"UncertaintyCode,omitempty"`
	UncertaintyMeters    *float64  `json:"UncertaintyMeters,omitempty"`
	UncertaintySemiMajor *uint8    `json:"UncertaintySemiMajor,omitempty"`
	UncertaintySemiMinor *uint8    `json:"UncertaintySemiMinor,omitempty"`
	AngleMajorAxis       *uint8    `json:"AngleMajorAxis,omitempty"`
	Confidence           *uint8    `json:"Confidence,omitempty"`
	Altitude             *int16    `json:"Altitude,omitempty"`
	UncertaintyAltitude  *uint8    `json:"UncertaintyAltitude,omitempty"`
	InnerRadius          *uint16   `json:"InnerRadius,omitempty"`
	UncertaintyRadius    *uint8    `json:"UncertaintyRadius,omitempty"`
	OffsetAngle          *uint8    `json:"OffsetAngle,omitempty"`
	IncludedAngle        *uint8    `json:"IncludedAngle,omitempty"`
}

// uncertaintyToMeters converts an uncertainty code K to meters.
// Formula per 3GPP TS 23.032: C(K) = 10 * (1.1^K - 1)
func uncertaintyToMeters(k uint8) float64 {
	return 10 * (math.Pow(1.1, float64(k)) - 1)
}

// check7Bit validates that a shape-specific 7-bit field fits within the
// 0..127 range defined by TS 23.032. The wire octet reserves bit 7 as
// spare/zero, so values above 127 cannot be encoded without truncation.
func check7Bit(name string, v uint8) error {
	if v > 127 {
		return fmt.Errorf("%s out of range [0, 127]: %d", name, v)
	}
	return nil
}

// checkAngle validates an angular octet against the TS 23.032 semantic
// bound of 0..179. Used for AngleMajorAxis (stored = degrees) and for
// the ellipsoid arc's OffsetAngle / IncludedAngle (stored × 2 = degrees,
// so the octet itself is capped at 179).
func checkAngle(name string, v uint8) error {
	if v > 179 {
		return fmt.Errorf("%s out of range [0, 179]: %d", name, v)
	}
	return nil
}

// DecodeGeographicalInfo decodes raw Ext-GeographicalInformation octets per 3GPP TS 23.032.
//
// Octet layout:
//
//	Octet 1:       [ShapeType:4 bits][spare/shape-specific:4 bits]
//	Octets 2-4:    Latitude  (1 sign bit + 23 magnitude bits)
//	Octets 5-7:    Longitude (24-bit two's complement)
//	Octets 8+:     Shape-specific fields
func DecodeGeographicalInfo(data []byte) (*GeographicalInfo, error) {
	if len(data) < 1 {
		return nil, fmt.Errorf("geographical information too short")
	}

	shapeType := ShapeType(data[0] >> 4)
	gi := &GeographicalInfo{ShapeType: shapeType}

	switch shapeType {
	case ShapeEllipsoidPoint:
		if len(data) < 7 {
			return nil, fmt.Errorf("ellipsoid point requires 7 octets, got %d", len(data))
		}
		gi.Latitude, gi.Longitude = decodeLatLon(data[1:7])

	case ShapeEllipsoidPointUncertainty:
		if len(data) < 8 {
			return nil, fmt.Errorf("ellipsoid point with uncertainty requires 8 octets, got %d", len(data))
		}
		gi.Latitude, gi.Longitude = decodeLatLon(data[1:7])
		uc := data[7] & 0x7F
		gi.UncertaintyCode = &uc
		meters := uncertaintyToMeters(uc)
		gi.UncertaintyMeters = &meters

	case ShapeEllipsoidPointUncertaintyEllipse:
		if len(data) < 11 {
			return nil, fmt.Errorf("ellipsoid point with uncertainty ellipse requires 11 octets, got %d", len(data))
		}
		gi.Latitude, gi.Longitude = decodeLatLon(data[1:7])
		semiMajor := data[7] & 0x7F
		semiMinor := data[8] & 0x7F
		angle := data[9]
		confidence := data[10] & 0x7F
		gi.UncertaintySemiMajor = &semiMajor
		gi.UncertaintySemiMinor = &semiMinor
		gi.AngleMajorAxis = &angle
		gi.Confidence = &confidence

	case ShapeEllipsoidPointAltitude:
		if len(data) < 14 {
			return nil, fmt.Errorf("ellipsoid point with altitude requires 14 octets, got %d", len(data))
		}
		gi.Latitude, gi.Longitude = decodeLatLon(data[1:7])
		altSign := data[7] >> 7
		altVal := int16(data[7]&0x7F)<<8 | int16(data[8])
		if altSign == 1 {
			altVal = -altVal
		}
		gi.Altitude = &altVal
		semiMajor := data[9] & 0x7F
		semiMinor := data[10] & 0x7F
		angle := data[11]
		uncAlt := data[12] & 0x7F
		confidence := data[13] & 0x7F
		gi.UncertaintySemiMajor = &semiMajor
		gi.UncertaintySemiMinor = &semiMinor
		gi.AngleMajorAxis = &angle
		gi.UncertaintyAltitude = &uncAlt
		gi.Confidence = &confidence

	case ShapeEllipsoidArc:
		if len(data) < 13 {
			return nil, fmt.Errorf("ellipsoid arc requires 13 octets, got %d", len(data))
		}
		gi.Latitude, gi.Longitude = decodeLatLon(data[1:7])
		innerRadius := uint16(data[7])<<8 | uint16(data[8])
		uncRadius := data[9] & 0x7F
		offsetAngle := data[10]
		includedAngle := data[11]
		confidence := data[12] & 0x7F
		gi.InnerRadius = &innerRadius
		gi.UncertaintyRadius = &uncRadius
		gi.OffsetAngle = &offsetAngle
		gi.IncludedAngle = &includedAngle
		gi.Confidence = &confidence

	default:
		return nil, fmt.Errorf("unsupported shape type: %d", shapeType)
	}

	return gi, nil
}

// Encode encodes a GeographicalInfo back to raw octets per 3GPP TS 23.032.
//
// Accepted ranges:
//   - Latitude must lie in the open interval (-90, 90). Exact ±90 is
//     rejected because the 23-bit wire encoding cannot represent it
//     without quantization.
//   - Longitude must lie in the half-open interval [-180, 180). The
//     lower bound -180 is exactly representable (two's-complement
//     0x800000); the upper bound +180 is not and is rejected.
//   - Values whose float64 ULP rounds onto a quantum they don't
//     represent are rejected: ULPs just inside ±90 and +180 round up
//     to the next quantum, and ULPs just above -180 round down onto
//     -180's own quantum. All would otherwise silently quantize.
//
// Rejects non-finite coordinates, missing required shape fields,
// shape-specific 7-bit fields whose value exceeds 127, and angular
// octets (AngleMajorAxis, OffsetAngle, IncludedAngle) exceeding the
// TS 23.032 bound of 179. Accepted inputs are rounded to the nearest
// representable quantum (roughly 1.07e-5° for latitude and 2.14e-5°
// for longitude), so the decoded float64 may differ from the caller's
// by up to half a quantum; the ULP-rejection rules above ensure it
// never crosses onto a different boundary.
//
// DecodeGeographicalInfo deliberately accepts spec-invalid wire input
// (e.g. an AngleMajorAxis octet of 200) without error so that a
// receiver can inspect what a peer actually sent. Round-tripping such
// a decoded value back through Encode will therefore fail at the
// boundary check above — that is intentional: the library will not
// re-emit bytes that violate TS 23.032.
func (gi *GeographicalInfo) Encode() ([]byte, error) {
	if math.IsNaN(gi.Latitude) || math.IsInf(gi.Latitude, 0) {
		return nil, fmt.Errorf("latitude is not a finite number: %v", gi.Latitude)
	}
	if math.IsNaN(gi.Longitude) || math.IsInf(gi.Longitude, 0) {
		return nil, fmt.Errorf("longitude is not a finite number: %v", gi.Longitude)
	}
	if gi.Latitude <= -90 || gi.Latitude >= 90 {
		return nil, fmt.Errorf("latitude out of range (-90, 90): %v", gi.Latitude)
	}
	if gi.Longitude < -180 || gi.Longitude >= 180 {
		return nil, fmt.Errorf("longitude out of range [-180, 180): %v", gi.Longitude)
	}
	latBytes, lonBytes, err := encodeLatLon(gi.Latitude, gi.Longitude)
	if err != nil {
		return nil, err
	}

	switch gi.ShapeType {
	case ShapeEllipsoidPoint:
		data := make([]byte, 7)
		data[0] = byte(gi.ShapeType) << 4
		copy(data[1:4], latBytes[:])
		copy(data[4:7], lonBytes[:])
		return data, nil

	case ShapeEllipsoidPointUncertainty:
		if gi.UncertaintyCode == nil {
			return nil, fmt.Errorf("EllipsoidPointUncertainty requires UncertaintyCode")
		}
		if err := check7Bit("UncertaintyCode", *gi.UncertaintyCode); err != nil {
			return nil, err
		}
		data := make([]byte, 8)
		data[0] = byte(gi.ShapeType) << 4
		copy(data[1:4], latBytes[:])
		copy(data[4:7], lonBytes[:])
		data[7] = *gi.UncertaintyCode
		return data, nil

	case ShapeEllipsoidPointUncertaintyEllipse:
		if gi.UncertaintySemiMajor == nil || gi.UncertaintySemiMinor == nil ||
			gi.AngleMajorAxis == nil || gi.Confidence == nil {
			return nil, fmt.Errorf("EllipsoidPointUncertaintyEllipse requires UncertaintySemiMajor, UncertaintySemiMinor, AngleMajorAxis, Confidence")
		}
		if err := check7Bit("UncertaintySemiMajor", *gi.UncertaintySemiMajor); err != nil {
			return nil, err
		}
		if err := check7Bit("UncertaintySemiMinor", *gi.UncertaintySemiMinor); err != nil {
			return nil, err
		}
		if err := checkAngle("AngleMajorAxis", *gi.AngleMajorAxis); err != nil {
			return nil, err
		}
		if err := check7Bit("Confidence", *gi.Confidence); err != nil {
			return nil, err
		}
		data := make([]byte, 11)
		data[0] = byte(gi.ShapeType) << 4
		copy(data[1:4], latBytes[:])
		copy(data[4:7], lonBytes[:])
		data[7] = *gi.UncertaintySemiMajor
		data[8] = *gi.UncertaintySemiMinor
		data[9] = *gi.AngleMajorAxis
		data[10] = *gi.Confidence
		return data, nil

	case ShapeEllipsoidPointAltitude:
		if gi.Altitude == nil || gi.UncertaintySemiMajor == nil ||
			gi.UncertaintySemiMinor == nil || gi.AngleMajorAxis == nil ||
			gi.UncertaintyAltitude == nil || gi.Confidence == nil {
			return nil, fmt.Errorf("EllipsoidPointAltitude requires Altitude, UncertaintySemiMajor, UncertaintySemiMinor, AngleMajorAxis, UncertaintyAltitude, Confidence")
		}
		alt := *gi.Altitude
		if alt < -32767 {
			return nil, fmt.Errorf("altitude out of range [-32767, 32767]: %d", alt)
		}
		if err := check7Bit("UncertaintySemiMajor", *gi.UncertaintySemiMajor); err != nil {
			return nil, err
		}
		if err := check7Bit("UncertaintySemiMinor", *gi.UncertaintySemiMinor); err != nil {
			return nil, err
		}
		if err := checkAngle("AngleMajorAxis", *gi.AngleMajorAxis); err != nil {
			return nil, err
		}
		if err := check7Bit("UncertaintyAltitude", *gi.UncertaintyAltitude); err != nil {
			return nil, err
		}
		if err := check7Bit("Confidence", *gi.Confidence); err != nil {
			return nil, err
		}
		data := make([]byte, 14)
		data[0] = byte(gi.ShapeType) << 4
		copy(data[1:4], latBytes[:])
		copy(data[4:7], lonBytes[:])
		if alt < 0 {
			data[7] = 0x80 | byte((-alt)>>8)
			data[8] = byte(-alt)
		} else {
			data[7] = byte(alt >> 8)
			data[8] = byte(alt)
		}
		data[9] = *gi.UncertaintySemiMajor
		data[10] = *gi.UncertaintySemiMinor
		data[11] = *gi.AngleMajorAxis
		data[12] = *gi.UncertaintyAltitude
		data[13] = *gi.Confidence
		return data, nil

	case ShapeEllipsoidArc:
		if gi.InnerRadius == nil || gi.UncertaintyRadius == nil ||
			gi.OffsetAngle == nil || gi.IncludedAngle == nil || gi.Confidence == nil {
			return nil, fmt.Errorf("EllipsoidArc requires InnerRadius, UncertaintyRadius, OffsetAngle, IncludedAngle, Confidence")
		}
		if err := check7Bit("UncertaintyRadius", *gi.UncertaintyRadius); err != nil {
			return nil, err
		}
		if err := checkAngle("OffsetAngle", *gi.OffsetAngle); err != nil {
			return nil, err
		}
		if err := checkAngle("IncludedAngle", *gi.IncludedAngle); err != nil {
			return nil, err
		}
		if err := check7Bit("Confidence", *gi.Confidence); err != nil {
			return nil, err
		}
		data := make([]byte, 13)
		data[0] = byte(gi.ShapeType) << 4
		copy(data[1:4], latBytes[:])
		copy(data[4:7], lonBytes[:])
		data[7] = byte(*gi.InnerRadius >> 8)
		data[8] = byte(*gi.InnerRadius)
		data[9] = *gi.UncertaintyRadius
		data[10] = *gi.OffsetAngle
		data[11] = *gi.IncludedAngle
		data[12] = *gi.Confidence
		return data, nil

	default:
		return nil, fmt.Errorf("unsupported shape type: %d", gi.ShapeType)
	}
}

// decodeLatLon decodes latitude and longitude from 6 octets per 3GPP TS 23.032.
//
//	Octets 1-3: Latitude  — bit 7 of octet 1 = sign (0=North, 1=South),
//	            bits 6-0 of octet 1 + octets 2-3 = 23-bit magnitude
//	            Lat(degrees) = N * 90 / 2^23
//	Octets 4-6: Longitude — 24-bit two's complement
//	            Lon(degrees) = N * 360 / 2^24
func decodeLatLon(data []byte) (lat, lon float64) {
	sign := (data[0] >> 7) & 0x01
	latN := uint32(data[0]&0x7F)<<16 | uint32(data[1])<<8 | uint32(data[2])
	lat = float64(latN) * 90.0 / float64(1<<23)
	if sign == 1 {
		lat = -lat
	}

	lonN := int32(data[3])<<16 | int32(data[4])<<8 | int32(data[5])
	// Sign-extend from 24 bits
	if lonN&0x800000 != 0 {
		lonN |= -1 << 24
	}
	lon = float64(lonN) * 360.0 / float64(1<<24)

	return lat, lon
}

// encodeLatLon encodes latitude and longitude to 6 octets per 3GPP TS 23.032.
// Returns [3]byte for latitude (sign in bit 7, 23-bit magnitude) and
// [3]byte for longitude (24-bit two's complement).
//
// Returns an error when the rounded quantum would force a value inside
// the caller-facing open range to share bytes with a boundary it didn't
// ask for. That happens for float64 ULPs just inside the ±90 and +180
// open bounds (they round up to 0x800000, colliding with the rejected
// boundary) and symmetrically for ULPs just above −180 (they round down
// to −0x800000, colliding with the legitimate -180 encoding).
func encodeLatLon(lat, lon float64) (latBytes [3]byte, lonBytes [3]byte, err error) {
	var sign byte
	rawLat := lat
	if lat < 0 {
		sign = 1
		lat = -lat
	}
	latN := uint32(math.Round(lat / 90.0 * float64(1<<23)))
	if latN > 0x7FFFFF {
		return latBytes, lonBytes, fmt.Errorf("latitude %v rounds past the ±90 boundary of the 23-bit encoding", rawLat)
	}
	latBytes[0] = sign<<7 | byte((latN>>16)&0x7F)
	latBytes[1] = byte(latN >> 8)
	latBytes[2] = byte(latN)

	lonN := int32(math.Round(lon / 360.0 * float64(1<<24)))
	if lonN > 0x7FFFFF {
		return latBytes, lonBytes, fmt.Errorf("longitude %v rounds past the +180 boundary of the 24-bit encoding", lon)
	}
	// -0x800000 is the legitimate quantum for lon=-180 exactly. Any other
	// caller value that rounds onto it (ULPs in (-180, -179.99998927°])
	// would silently collapse onto -180 on the wire.
	if lonN == -0x800000 && lon != -180 {
		return latBytes, lonBytes, fmt.Errorf("longitude %v rounds onto the -180 quantum (would silently collapse onto -180)", lon)
	}
	lonBytes[0] = byte(lonN >> 16)
	lonBytes[1] = byte(lonN >> 8)
	lonBytes[2] = byte(lonN)

	return latBytes, lonBytes, nil
}

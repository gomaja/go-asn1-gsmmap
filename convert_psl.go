// convert_psl.go
//
// Converters for ProvideSubscriberLocation (opCode 83) leaf SEQUENCE
// types and BIT STRING surrogates. PR D1 of the staged PSL
// implementation. Container converters (LCSClientID,
// AreaEventInfo, PeriodicLDRInfo, ReportingPLMNList) and the
// top-level ProvideSubscriberLocationArg/Res live in subsequent PRs.
//
// Each converter pair:
//   convertXToWire(*X) (*gsm_map.X, error)   — public type → wire
//   convertWireToX(*gsm_map.X) (*X, error)   — wire → public type
//
// Validation (size/range/enum) lives in the converters and surfaces the
// sentinels defined in gsmmap.go.

package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/runtime"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// ============================================================================
// BIT STRING surrogates
// ============================================================================

// DeferredLocationEventType (BIT STRING SIZE 1..16, 5 named bits) per
// TS 29.002 MAP-LCS-DataTypes.asn:165.
//
// Encode rule: BitLength is the position of the highest set bit + 1
// (minimum 1 to satisfy the SIZE 1..16 lower bound).
func convertDeferredLocationEventTypeToBitString(d *DeferredLocationEventType) runtime.BitString {
	var b byte
	bitLen := 1
	if d.MsAvailable {
		b |= 0x80
	}
	if d.EnteringIntoArea {
		b |= 0x40
		if bitLen < 2 {
			bitLen = 2
		}
	}
	if d.LeavingFromArea {
		b |= 0x20
		if bitLen < 3 {
			bitLen = 3
		}
	}
	if d.BeingInsideArea {
		b |= 0x10
		if bitLen < 4 {
			bitLen = 4
		}
	}
	if d.PeriodicLDR {
		b |= 0x08
		if bitLen < 5 {
			bitLen = 5
		}
	}
	return runtime.BitString{Bytes: []byte{b}, BitLength: bitLen}
}

// convertBitStringToDeferredLocationEventType validates the wire size
// (1..16 bits) and decodes the 5 named bits. Bits past the 5 named bits
// are tolerated and ignored on decode (forward-compat).
func convertBitStringToDeferredLocationEventType(bs runtime.BitString) (*DeferredLocationEventType, error) {
	if bs.BitLength < 1 || bs.BitLength > 16 {
		return nil, fmt.Errorf("DeferredLocationEventType: %w (got %d bits)", ErrDeferredLocationEventTypeSize, bs.BitLength)
	}
	d := &DeferredLocationEventType{}
	if bs.BitLength > 0 {
		d.MsAvailable = bs.Has(0)
	}
	if bs.BitLength > 1 {
		d.EnteringIntoArea = bs.Has(1)
	}
	if bs.BitLength > 2 {
		d.LeavingFromArea = bs.Has(2)
	}
	if bs.BitLength > 3 {
		d.BeingInsideArea = bs.Has(3)
	}
	if bs.BitLength > 4 {
		d.PeriodicLDR = bs.Has(4)
	}
	return d, nil
}

// SupportedGADShapes (BIT STRING SIZE 7..16, 7 named bits) per TS 29.002
// MAP-LCS-DataTypes.asn:280.
//
// Encode rule: always emit 7 bits to satisfy the SIZE 7..16 lower bound,
// even when no flag is set.
func convertSupportedGADShapesToBitString(g *SupportedGADShapes) runtime.BitString {
	var b byte
	if g.EllipsoidPoint {
		b |= 0x80
	}
	if g.EllipsoidPointWithUncertaintyCircle {
		b |= 0x40
	}
	if g.EllipsoidPointWithUncertaintyEllipse {
		b |= 0x20
	}
	if g.Polygon {
		b |= 0x10
	}
	if g.EllipsoidPointWithAltitude {
		b |= 0x08
	}
	if g.EllipsoidPointWithAltitudeAndUncertaintyEllipsoid {
		b |= 0x04
	}
	if g.EllipsoidArc {
		b |= 0x02
	}
	return runtime.BitString{Bytes: []byte{b}, BitLength: 7}
}

// convertBitStringToSupportedGADShapes validates the wire size (7..16
// bits) and decodes the 7 named bits. Bits past the 7 named bits are
// tolerated and ignored on decode.
func convertBitStringToSupportedGADShapes(bs runtime.BitString) (*SupportedGADShapes, error) {
	if bs.BitLength < 7 || bs.BitLength > 16 {
		return nil, fmt.Errorf("SupportedGADShapes: %w (got %d bits)", ErrSupportedGADShapesSize, bs.BitLength)
	}
	g := &SupportedGADShapes{}
	g.EllipsoidPoint = bs.Has(0)
	g.EllipsoidPointWithUncertaintyCircle = bs.Has(1)
	g.EllipsoidPointWithUncertaintyEllipse = bs.Has(2)
	g.Polygon = bs.Has(3)
	g.EllipsoidPointWithAltitude = bs.Has(4)
	g.EllipsoidPointWithAltitudeAndUncertaintyEllipsoid = bs.Has(5)
	g.EllipsoidArc = bs.Has(6)
	return g, nil
}

// ============================================================================
// LocationType — TS 29.002 MAP-LCS-DataTypes.asn:148
// ============================================================================

func convertLocationTypeToWire(l *LocationType) (*gsm_map.LocationType, error) {
	if l == nil {
		return nil, nil
	}
	out := &gsm_map.LocationType{
		LocationEstimateType: l.LocationEstimateType,
	}
	if int64(l.LocationEstimateType) < 0 || int64(l.LocationEstimateType) > 5 {
		return nil, fmt.Errorf("LocationType.LocationEstimateType=%d: %w", l.LocationEstimateType, ErrLocationEstimateTypeInvalid)
	}
	if l.DeferredLocationEventType != nil {
		bs := convertDeferredLocationEventTypeToBitString(l.DeferredLocationEventType)
		out.DeferredLocationEventType = &bs
	}
	return out, nil
}

func convertWireToLocationType(w *gsm_map.LocationType) (*LocationType, error) {
	if w == nil {
		return nil, nil
	}
	out := &LocationType{
		LocationEstimateType: w.LocationEstimateType,
	}
	if w.DeferredLocationEventType != nil {
		d, err := convertBitStringToDeferredLocationEventType(*w.DeferredLocationEventType)
		if err != nil {
			return nil, fmt.Errorf("LocationType.DeferredLocationEventType: %w", err)
		}
		out.DeferredLocationEventType = d
	}
	return out, nil
}

// ============================================================================
// LCSCodeword — TS 29.002 MAP-LCS-DataTypes.asn:293
// ============================================================================

func convertLCSCodewordToWire(c *LCSCodeword) (*gsm_map.LCSCodeword, error) {
	if c == nil {
		return nil, nil
	}
	if len(c.LcsCodewordString) < 1 || len(c.LcsCodewordString) > LCSCodewordStringMaxLen {
		return nil, fmt.Errorf("LCSCodeword.LcsCodewordString len=%d: %w", len(c.LcsCodewordString), ErrLCSCodewordStringSize)
	}
	out := &gsm_map.LCSCodeword{
		DataCodingScheme:  gsm_map.USSDDataCodingScheme{c.DataCodingScheme},
		LcsCodewordString: gsm_map.LCSCodewordString(c.LcsCodewordString),
	}
	return out, nil
}

func convertWireToLCSCodeword(w *gsm_map.LCSCodeword) (*LCSCodeword, error) {
	if w == nil {
		return nil, nil
	}
	if len(w.DataCodingScheme) != 1 {
		return nil, fmt.Errorf("LCSCodeword.DataCodingScheme len=%d: %w", len(w.DataCodingScheme), ErrUSSDDataCodingSchemeInvalidSize)
	}
	if len(w.LcsCodewordString) < 1 || len(w.LcsCodewordString) > LCSCodewordStringMaxLen {
		return nil, fmt.Errorf("LCSCodeword.LcsCodewordString len=%d: %w", len(w.LcsCodewordString), ErrLCSCodewordStringSize)
	}
	return &LCSCodeword{
		DataCodingScheme:  w.DataCodingScheme[0],
		LcsCodewordString: HexBytes(w.LcsCodewordString),
	}, nil
}

// ============================================================================
// LCSPrivacyCheck — TS 29.002 MAP-LCS-DataTypes.asn:302
// ============================================================================

func convertLCSPrivacyCheckToWire(p *LCSPrivacyCheck) (*gsm_map.LCSPrivacyCheck, error) {
	if p == nil {
		return nil, nil
	}
	if int64(p.CallSessionUnrelated) < 0 || int64(p.CallSessionUnrelated) > 4 {
		return nil, fmt.Errorf("LCSPrivacyCheck.CallSessionUnrelated=%d: %w", p.CallSessionUnrelated, ErrPrivacyCheckRelatedActionInvalid)
	}
	out := &gsm_map.LCSPrivacyCheck{
		CallSessionUnrelated: p.CallSessionUnrelated,
	}
	if p.CallSessionRelated != nil {
		v := *p.CallSessionRelated
		if int64(v) < 0 || int64(v) > 4 {
			return nil, fmt.Errorf("LCSPrivacyCheck.CallSessionRelated=%d: %w", v, ErrPrivacyCheckRelatedActionInvalid)
		}
		out.CallSessionRelated = &v
	}
	return out, nil
}

func convertWireToLCSPrivacyCheck(w *gsm_map.LCSPrivacyCheck) (*LCSPrivacyCheck, error) {
	if w == nil {
		return nil, nil
	}
	// PrivacyCheckRelatedAction is NOT extensible (TS 29.002 MAP-LCS-DataTypes.asn:307);
	// validate symmetrically with the encoder.
	if int64(w.CallSessionUnrelated) < 0 || int64(w.CallSessionUnrelated) > 4 {
		return nil, fmt.Errorf("LCSPrivacyCheck.CallSessionUnrelated=%d: %w", w.CallSessionUnrelated, ErrPrivacyCheckRelatedActionInvalid)
	}
	out := &LCSPrivacyCheck{
		CallSessionUnrelated: w.CallSessionUnrelated,
	}
	if w.CallSessionRelated != nil {
		v := *w.CallSessionRelated
		if int64(v) < 0 || int64(v) > 4 {
			return nil, fmt.Errorf("LCSPrivacyCheck.CallSessionRelated=%d: %w", v, ErrPrivacyCheckRelatedActionInvalid)
		}
		out.CallSessionRelated = &v
	}
	return out, nil
}

// ============================================================================
// ResponseTime — TS 29.002 MAP-LCS-DataTypes.asn:261
// ============================================================================
//
// ResponseTimeCategory is an extensible ENUMERATED with a spec exception
// clause: unrecognized values shall be treated as delaytolerant(1) on
// decode. Decoder applies the exception clause; encoder is strict
// (lowdelay or delaytolerant only).

func convertResponseTimeToWire(r *ResponseTime) (*gsm_map.ResponseTime, error) {
	if r == nil {
		return nil, nil
	}
	if r.ResponseTimeCategory != ResponseTimeLowdelay && r.ResponseTimeCategory != ResponseTimeDelaytolerant {
		return nil, fmt.Errorf("ResponseTime.ResponseTimeCategory=%d: %w", r.ResponseTimeCategory, ErrResponseTimeCategoryInvalid)
	}
	return &gsm_map.ResponseTime{
		ResponseTimeCategory: r.ResponseTimeCategory,
	}, nil
}

func convertWireToResponseTime(w *gsm_map.ResponseTime) (*ResponseTime, error) {
	if w == nil {
		return nil, nil
	}
	cat := w.ResponseTimeCategory
	// Per TS 29.002 MAP-LCS-DataTypes.asn:270-271, an unrecognized value
	// shall be treated the same as delaytolerant(1).
	if cat != ResponseTimeLowdelay && cat != ResponseTimeDelaytolerant {
		cat = ResponseTimeDelaytolerant
	}
	return &ResponseTime{
		ResponseTimeCategory: cat,
	}, nil
}

// ============================================================================
// LCSQoS — TS 29.002 MAP-LCS-DataTypes.asn:237
// ============================================================================

func convertLCSQoSToWire(q *LCSQoS) (*gsm_map.LCSQoS, error) {
	if q == nil {
		return nil, nil
	}
	out := &gsm_map.LCSQoS{}

	if q.HorizontalAccuracy != nil {
		if len(q.HorizontalAccuracy) != 1 {
			return nil, fmt.Errorf("LCSQoS.HorizontalAccuracy len=%d: %w", len(q.HorizontalAccuracy), ErrHorizontalAccuracyInvalidSize)
		}
		// Spec mandates bit 8 = 0 (TS 29.002 MAP-LCS-DataTypes.asn:250):
		// only the low 7 bits encode the uncertainty code per TS 23.032.
		if q.HorizontalAccuracy[0]&0x80 != 0 {
			return nil, fmt.Errorf("LCSQoS.HorizontalAccuracy=0x%02x: %w", q.HorizontalAccuracy[0], ErrHorizontalAccuracyReservedBit)
		}
		v := gsm_map.HorizontalAccuracy(q.HorizontalAccuracy)
		out.HorizontalAccuracy = &v
	}
	out.VerticalCoordinateRequest = boolToNullPtr(q.VerticalCoordinateRequest)
	if q.VerticalAccuracy != nil {
		if len(q.VerticalAccuracy) != 1 {
			return nil, fmt.Errorf("LCSQoS.VerticalAccuracy len=%d: %w", len(q.VerticalAccuracy), ErrVerticalAccuracyInvalidSize)
		}
		// Spec mandates bit 8 = 0 (TS 29.002 MAP-LCS-DataTypes.asn:256):
		// only the low 7 bits encode the vertical uncertainty code per TS 23.032.
		if q.VerticalAccuracy[0]&0x80 != 0 {
			return nil, fmt.Errorf("LCSQoS.VerticalAccuracy=0x%02x: %w", q.VerticalAccuracy[0], ErrVerticalAccuracyReservedBit)
		}
		v := gsm_map.VerticalAccuracy(q.VerticalAccuracy)
		out.VerticalAccuracy = &v
	}
	if q.ResponseTime != nil {
		rt, err := convertResponseTimeToWire(q.ResponseTime)
		if err != nil {
			return nil, fmt.Errorf("LCSQoS.ResponseTime: %w", err)
		}
		out.ResponseTime = rt
	}
	out.VelocityRequest = boolToNullPtr(q.VelocityRequest)
	return out, nil
}

func convertWireToLCSQoS(w *gsm_map.LCSQoS) (*LCSQoS, error) {
	if w == nil {
		return nil, nil
	}
	out := &LCSQoS{}
	if w.HorizontalAccuracy != nil {
		if len(*w.HorizontalAccuracy) != 1 {
			return nil, fmt.Errorf("LCSQoS.HorizontalAccuracy len=%d: %w", len(*w.HorizontalAccuracy), ErrHorizontalAccuracyInvalidSize)
		}
		if (*w.HorizontalAccuracy)[0]&0x80 != 0 {
			return nil, fmt.Errorf("LCSQoS.HorizontalAccuracy=0x%02x: %w", (*w.HorizontalAccuracy)[0], ErrHorizontalAccuracyReservedBit)
		}
		out.HorizontalAccuracy = HexBytes(*w.HorizontalAccuracy)
	}
	out.VerticalCoordinateRequest = nullPtrToBool(w.VerticalCoordinateRequest)
	if w.VerticalAccuracy != nil {
		if len(*w.VerticalAccuracy) != 1 {
			return nil, fmt.Errorf("LCSQoS.VerticalAccuracy len=%d: %w", len(*w.VerticalAccuracy), ErrVerticalAccuracyInvalidSize)
		}
		if (*w.VerticalAccuracy)[0]&0x80 != 0 {
			return nil, fmt.Errorf("LCSQoS.VerticalAccuracy=0x%02x: %w", (*w.VerticalAccuracy)[0], ErrVerticalAccuracyReservedBit)
		}
		out.VerticalAccuracy = HexBytes(*w.VerticalAccuracy)
	}
	if w.ResponseTime != nil {
		rt, err := convertWireToResponseTime(w.ResponseTime)
		if err != nil {
			return nil, fmt.Errorf("LCSQoS.ResponseTime: %w", err)
		}
		out.ResponseTime = rt
	}
	out.VelocityRequest = nullPtrToBool(w.VelocityRequest)
	return out, nil
}
